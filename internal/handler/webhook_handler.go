package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"buildflow/internal/model"
	"buildflow/internal/pkg"
	"buildflow/internal/repository"
	"buildflow/internal/service"
)

type WebhookHandler struct {
	projectService *service.ProjectService
	buildService   *service.BuildService
	envRepo        *repository.EnvironmentRepository
	scheduler      BuildScheduler
}

func NewWebhookHandler(ps *service.ProjectService, bs *service.BuildService, envRepo *repository.EnvironmentRepository, scheduler BuildScheduler) *WebhookHandler {
	return &WebhookHandler{
		projectService: ps,
		buildService:   bs,
		envRepo:        envRepo,
		scheduler:      scheduler,
	}
}

// GitHub push payload: refs/heads/branch_name
type githubPushPayload struct {
	Ref    string `json:"ref"`
	After  string `json:"after"`
	Pusher struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	} `json:"pusher"`
	Repository struct {
		CloneURL string `json:"clone_url"`
	} `json:"repository"`
	HeadCommit struct {
		ID      string `json:"id"`
		Message string `json:"message"`
	} `json:"head_commit"`
}

// GitLab push payload: refs/heads/branch_name
type gitlabPushPayload struct {
	Ref          string `json:"ref"`
	UserUsername string `json:"user_username"`
	Project      struct {
		GitHTTPURL string `json:"git_http_url"`
	} `json:"project"`
	CheckoutSha string `json:"checkout_sha"`
	Commit      *struct {
		Message string `json:"message"`
	} `json:"commit"`
}

type bitbucketPushPayload struct {
	Push struct {
		Changes []struct {
			New *struct {
				Name   string `json:"name"`
				Target struct {
					Hash    string `json:"hash"`
					Message string `json:"message"`
				} `json:"target"`
			} `json:"new"`
		} `json:"changes"`
	} `json:"push"`
}

type webhookEvent struct {
	Ref           string
	CommitHash    string
	CommitMessage string
}

func extractBranchFromRef(ref string) string {
	// refs/heads/branch_name or refs/tags/tag_name
	if strings.HasPrefix(ref, "refs/heads/") {
		return strings.TrimPrefix(ref, "refs/heads/")
	}
	return ref
}

// POST /api/v1/webhook/:projectId/:secret - verify secret, parse GitHub/GitLab payload, trigger builds
func (h *WebhookHandler) Handle(c *gin.Context) {
	projectID, err := strconv.ParseUint(c.Param("projectId"), 10, 32)
	if err != nil {
		pkg.Error(c, http.StatusBadRequest, "参数错误")
		return
	}
	secret := c.Param("secret")
	if secret == "" {
		pkg.Error(c, http.StatusBadRequest, "webhook secret 必填")
		return
	}
	project, err := h.projectService.GetByID(uint(projectID))
	if err != nil {
		pkg.Error(c, http.StatusNotFound, "项目不存在")
		return
	}
	if project.WebhookSecret != secret {
		pkg.Error(c, http.StatusUnauthorized, "无效的 webhook secret")
		return
	}
	body, err := c.GetRawData()
	if err != nil {
		pkg.Error(c, http.StatusBadRequest, "无法读取请求体")
		return
	}
	var ref string
	var commitHash, commitMessage string
	event, err := parseWebhookPayload(c, project, body)
	if err != nil {
		pkg.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	ref = event.Ref
	commitHash = event.CommitHash
	commitMessage = event.CommitMessage
	branch := extractBranchFromRef(ref)
	var filterEnvID uint
	if q := strings.TrimSpace(c.Query("environment_id")); q != "" {
		parsed, err := strconv.ParseUint(q, 10, 32)
		if err != nil {
			pkg.Error(c, http.StatusBadRequest, "environment_id 无效")
			return
		}
		filterEnvID = uint(parsed)
	}
	envs, err := h.envRepo.ListByProjectID(uint(projectID))
	if err != nil {
		pkg.Error(c, http.StatusInternalServerError, "查询环境失败")
		return
	}
	var triggered int
	for _, env := range envs {
		if filterEnvID != 0 && env.ID != filterEnvID {
			continue
		}
		if env.Branch == branch {
			build, err := h.buildService.TriggerBuild(uint(projectID), env.ID, 0, "webhook", "", commitHash, commitMessage)
			if err != nil {
				continue
			}
			if h.scheduler != nil {
				h.scheduler.Submit(build.ID)
			}
			triggered++
		}
	}
	out := gin.H{"triggered": triggered, "branch": branch}
	if filterEnvID != 0 {
		out["environment_id"] = filterEnvID
	}
	pkg.Success(c, out)
}

func parseWebhookPayload(c *gin.Context, project *model.Project, body []byte) (*webhookEvent, error) {
	platform := detectWebhookPlatform(c, project)
	switch platform {
	case "github":
		return parseGitHubWebhook(c, body)
	case "gitlab":
		return parseGitLabWebhook(c, body)
	case "gitea":
		return parseGiteaWebhook(c, body)
	case "gitee":
		return parseGiteeWebhook(c, body)
	case "bitbucket":
		return parseBitbucketWebhook(c, body)
	case "generic":
		return parseGenericWebhook(project, body)
	default:
		return nil, fmt.Errorf("无法识别 webhook 平台，请检查请求头或项目配置")
	}
}

func detectWebhookPlatform(c *gin.Context, project *model.Project) string {
	switch {
	case c.GetHeader("X-Gitea-Event") != "":
		return "gitea"
	case c.GetHeader("X-Gitee-Event") != "":
		return "gitee"
	case c.GetHeader("X-GitHub-Event") != "":
		return "github"
	case c.GetHeader("X-Gitlab-Event") != "":
		return "gitlab"
	case c.GetHeader("X-Event-Key") != "":
		return "bitbucket"
	case project.WebhookType != "" && project.WebhookType != "auto":
		return project.WebhookType
	case project.WebhookRefPath != "":
		return "generic"
	default:
		return ""
	}
}

// parseGiteeWebhook 码云 Push / Tag Push：请求体与 GitHub push 类似（含 ref、after、head_commit）。
// 请求头 X-Gitee-Event 一般为 Push Hook 或 Tag Push Hook，见 https://help.gitee.com  WebHook 说明。
func parseGiteeWebhook(c *gin.Context, body []byte) (*webhookEvent, error) {
	if eventType := c.GetHeader("X-Gitee-Event"); eventType != "" {
		et := strings.ToLower(strings.TrimSpace(eventType))
		if et != "push hook" && et != "tag push hook" {
			return nil, fmt.Errorf("仅支持 Push Hook / Tag Push Hook 事件")
		}
	}
	var payload githubPushPayload
	if err := json.Unmarshal(body, &payload); err != nil || payload.Ref == "" {
		return nil, fmt.Errorf("无法解析 Gitee push payload")
	}
	commitHash := payload.HeadCommit.ID
	if commitHash == "" {
		commitHash = payload.After
	}
	return &webhookEvent{
		Ref:           payload.Ref,
		CommitHash:    commitHash,
		CommitMessage: payload.HeadCommit.Message,
	}, nil
}

func parseGitHubWebhook(c *gin.Context, body []byte) (*webhookEvent, error) {
	if eventType := c.GetHeader("X-GitHub-Event"); eventType != "" && eventType != "push" {
		return nil, fmt.Errorf("仅支持 push 事件")
	}
	var payload githubPushPayload
	if err := json.Unmarshal(body, &payload); err != nil || payload.Ref == "" {
		return nil, fmt.Errorf("无法解析 GitHub push payload")
	}
	commitHash := payload.HeadCommit.ID
	if commitHash == "" {
		commitHash = payload.After
	}
	return &webhookEvent{
		Ref:           payload.Ref,
		CommitHash:    commitHash,
		CommitMessage: payload.HeadCommit.Message,
	}, nil
}

func parseGitLabWebhook(c *gin.Context, body []byte) (*webhookEvent, error) {
	if eventType := c.GetHeader("X-Gitlab-Event"); eventType != "" && !strings.EqualFold(eventType, "Push Hook") {
		return nil, fmt.Errorf("仅支持 push 事件")
	}
	var payload gitlabPushPayload
	if err := json.Unmarshal(body, &payload); err != nil || payload.Ref == "" {
		return nil, fmt.Errorf("无法解析 GitLab push payload")
	}
	message := ""
	if payload.Commit != nil {
		message = payload.Commit.Message
	}
	return &webhookEvent{
		Ref:           payload.Ref,
		CommitHash:    payload.CheckoutSha,
		CommitMessage: message,
	}, nil
}

func parseGiteaWebhook(c *gin.Context, body []byte) (*webhookEvent, error) {
	if eventType := c.GetHeader("X-Gitea-Event"); eventType != "" && !strings.EqualFold(eventType, "push") {
		return nil, fmt.Errorf("仅支持 push 事件")
	}
	var payload githubPushPayload
	if err := json.Unmarshal(body, &payload); err != nil || payload.Ref == "" {
		return nil, fmt.Errorf("无法解析 Gitea push payload")
	}
	commitHash := payload.HeadCommit.ID
	if commitHash == "" {
		commitHash = payload.After
	}
	return &webhookEvent{
		Ref:           payload.Ref,
		CommitHash:    commitHash,
		CommitMessage: payload.HeadCommit.Message,
	}, nil
}

func parseBitbucketWebhook(c *gin.Context, body []byte) (*webhookEvent, error) {
	if eventType := c.GetHeader("X-Event-Key"); eventType != "" && eventType != "repo:push" {
		return nil, fmt.Errorf("仅支持 repo:push 事件")
	}
	var payload bitbucketPushPayload
	if err := json.Unmarshal(body, &payload); err != nil || len(payload.Push.Changes) == 0 {
		return nil, fmt.Errorf("无法解析 Bitbucket push payload")
	}
	change := payload.Push.Changes[0]
	if change.New == nil {
		return nil, fmt.Errorf("bitbucket payload 缺少分支信息")
	}
	return &webhookEvent{
		Ref:           "refs/heads/" + change.New.Name,
		CommitHash:    change.New.Target.Hash,
		CommitMessage: change.New.Target.Message,
	}, nil
}

func parseGenericWebhook(project *model.Project, body []byte) (*webhookEvent, error) {
	if project.WebhookRefPath == "" {
		return nil, fmt.Errorf("通用 Webhook 未配置 ref JSONPath")
	}
	var payload any
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, fmt.Errorf("无法解析通用 JSON payload")
	}
	ref, err := extractJSONString(payload, project.WebhookRefPath)
	if err != nil {
		return nil, fmt.Errorf("读取 ref 失败: %w", err)
	}
	commitHash := ""
	if project.WebhookCommitPath != "" {
		commitHash, _ = extractJSONString(payload, project.WebhookCommitPath)
	}
	commitMessage := ""
	if project.WebhookMessagePath != "" {
		commitMessage, _ = extractJSONString(payload, project.WebhookMessagePath)
	}
	return &webhookEvent{
		Ref:           ref,
		CommitHash:    commitHash,
		CommitMessage: commitMessage,
	}, nil
}

func extractJSONString(payload any, path string) (string, error) {
	value, err := extractJSONValue(payload, path)
	if err != nil {
		return "", err
	}
	switch typed := value.(type) {
	case string:
		return typed, nil
	case float64:
		return fmt.Sprintf("%.0f", typed), nil
	case bool:
		if typed {
			return "true", nil
		}
		return "false", nil
	default:
		return "", fmt.Errorf("JSONPath %s 对应值不是字符串", path)
	}
}

func extractJSONValue(payload any, path string) (any, error) {
	trimmed := strings.TrimSpace(path)
	trimmed = strings.TrimPrefix(trimmed, "$.")
	trimmed = strings.TrimPrefix(trimmed, "$")
	if trimmed == "" {
		return nil, fmt.Errorf("JSONPath 不能为空")
	}
	current := payload
	for _, token := range splitJSONPath(trimmed) {
		name := token
		index := -1
		if open := strings.Index(token, "["); open >= 0 && strings.HasSuffix(token, "]") {
			name = token[:open]
			var parsed int
			if _, err := fmt.Sscanf(token[open:], "[%d]", &parsed); err != nil {
				return nil, fmt.Errorf("不支持的 JSONPath 片段 %s", token)
			}
			index = parsed
		}
		if name != "" {
			obj, ok := current.(map[string]any)
			if !ok {
				return nil, fmt.Errorf("JSONPath 片段 %s 不是对象", token)
			}
			value, exists := obj[name]
			if !exists {
				return nil, fmt.Errorf("JSONPath 片段 %s 不存在", token)
			}
			current = value
		}
		if index >= 0 {
			items, ok := current.([]any)
			if !ok {
				return nil, fmt.Errorf("JSONPath 片段 %s 不是数组", token)
			}
			if index >= len(items) {
				return nil, fmt.Errorf("JSONPath 数组索引越界: %s", token)
			}
			current = items[index]
		}
	}
	return current, nil
}

func splitJSONPath(path string) []string {
	parts := strings.Split(path, ".")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		if part != "" {
			result = append(result, part)
		}
	}
	return result
}
