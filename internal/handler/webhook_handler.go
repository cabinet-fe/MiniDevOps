package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

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
	Ref string `json:"ref"`
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
	CheckoutSha string   `json:"checkout_sha"`
	Commit      *struct {
		Message string `json:"message"`
	} `json:"commit"`
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

	// Try GitHub format first
	var gh githubPushPayload
	if err := json.Unmarshal(body, &gh); err == nil && gh.Ref != "" {
		ref = gh.Ref
		commitHash = gh.HeadCommit.ID
		commitMessage = gh.HeadCommit.Message
	} else {
		// Try GitLab format
		var gl gitlabPushPayload
		if err := json.Unmarshal(body, &gl); err == nil && gl.Ref != "" {
			ref = gl.Ref
			commitHash = gl.CheckoutSha
			if gl.Commit != nil {
				commitMessage = gl.Commit.Message
			}
		}
	}
	if ref == "" {
		pkg.Error(c, http.StatusBadRequest, "无法解析 push 事件，请检查 payload 格式")
		return
	}
	branch := extractBranchFromRef(ref)
	envs, err := h.envRepo.ListByProjectID(uint(projectID))
	if err != nil {
		pkg.Error(c, http.StatusInternalServerError, "查询环境失败")
		return
	}
	var triggered int
	for _, env := range envs {
		if env.Branch == branch {
			build, err := h.buildService.TriggerBuild(uint(projectID), env.ID, 0, "webhook", commitHash, commitMessage)
			if err != nil {
				continue
			}
			if h.scheduler != nil {
				h.scheduler.Submit(build.ID)
			}
			triggered++
		}
	}
	pkg.Success(c, gin.H{"triggered": triggered, "branch": branch})
}
