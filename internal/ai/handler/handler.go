package handler

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"bedrock/internal/ai/service"
	authmiddleware "bedrock/internal/auth/middleware"
	"bedrock/internal/pkg"
	rbacmw "bedrock/internal/rbac/middleware"
	rbacservice "bedrock/internal/rbac/service"
	storageservice "bedrock/internal/storage/service"
)

type Handler struct {
	cli    *service.CLIService
	agents *service.AgentService
	skills *service.SkillService
	pats   *service.PATService
	perm   *rbacservice.PermissionService
}

func NewHandler(
	cli *service.CLIService,
	agents *service.AgentService,
	skills *service.SkillService,
	pats *service.PATService,
	perm *rbacservice.PermissionService,
) *Handler {
	return &Handler{cli: cli, agents: agents, skills: skills, pats: pats, perm: perm}
}

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup, authMW gin.HandlerFunc) {
	ai := rg.Group("/ai", authMW)
	ai.GET("/clis", rbacmw.RequirePermission(h.perm, "ai.clis:view"), h.ListCLIs)
	ai.POST("/clis/:key/detect", rbacmw.RequirePermission(h.perm, "ai.clis:execute"), h.DetectCLI)
	ai.POST("/clis/:key/install", rbacmw.RequirePermission(h.perm, "ai.clis:execute"), h.InstallCLI)
	ai.POST("/clis/:key/upgrade", rbacmw.RequirePermission(h.perm, "ai.clis:execute"), h.UpgradeCLI)
	ai.POST("/clis/:key/uninstall", rbacmw.RequirePermission(h.perm, "ai.clis:execute"), h.UninstallCLI)
	ai.GET("/cli-sources", rbacmw.RequirePermission(h.perm, "ai.clis:view"), h.ListCLISources)
	ai.POST("/cli-sources", rbacmw.RequirePermission(h.perm, "ai.clis:create"), h.CreateCLISource)
	ai.PUT("/cli-sources/:id", rbacmw.RequirePermission(h.perm, "ai.clis:update"), h.UpdateCLISource)
	ai.DELETE("/cli-sources/:id", rbacmw.RequirePermission(h.perm, "ai.clis:delete"), h.DeleteCLISource)

	ai.GET("/agents", rbacmw.RequirePermission(h.perm, "ai.agents:view"), h.ListAgents)
	ai.POST("/agents", rbacmw.RequirePermission(h.perm, "ai.agents:create"), h.CreateAgent)
	ai.GET("/agents/:id", rbacmw.RequirePermission(h.perm, "ai.agents:view"), h.GetAgent)
	ai.PUT("/agents/:id", rbacmw.RequirePermission(h.perm, "ai.agents:update"), h.UpdateAgent)
	ai.DELETE("/agents/:id", rbacmw.RequirePermission(h.perm, "ai.agents:delete"), h.DeleteAgent)
	ai.GET("/agents/:id/triggers", rbacmw.RequirePermission(h.perm, "ai.agents:view"), h.ListTriggers)
	ai.POST("/agents/:id/triggers", rbacmw.RequirePermission(h.perm, "ai.agents:update"), h.CreateTrigger)
	ai.PUT("/agents/:id/triggers/:tid", rbacmw.RequirePermission(h.perm, "ai.agents:update"), h.UpdateTrigger)
	ai.DELETE("/agents/:id/triggers/:tid", rbacmw.RequirePermission(h.perm, "ai.agents:update"), h.DeleteTrigger)
	ai.POST("/agents/:id/runs", rbacmw.RequirePermission(h.perm, "ai.agents:execute"), h.ManualRun)
	// API trigger also accepts PAT scope agents:run (checked in middleware/handler).
	ai.POST("/agents/:id/api-runs", h.APIRun)

	ai.GET("/runs", rbacmw.RequirePermission(h.perm, "ai.runs:view"), h.ListRuns)
	ai.GET("/runs/:id", rbacmw.RequirePermission(h.perm, "ai.runs:view"), h.GetRun)
	ai.GET("/runs/:id/artifact", rbacmw.RequirePermission(h.perm, "ai.runs:view"), h.DownloadRunArtifact)
	ai.POST("/runs/:id/cancel", rbacmw.RequirePermission(h.perm, "ai.agents:execute"), h.CancelRun)

	skills := rg.Group("/skills", authMW)
	skills.GET("", rbacmw.RequirePermission(h.perm, "ai.skills:view"), h.ListSkills)
	skills.POST("", rbacmw.RequirePermission(h.perm, "ai.skills:create"), h.CreateSkill)
	skills.GET("/:id", rbacmw.RequirePermission(h.perm, "ai.skills:view"), h.GetSkill)
	skills.PUT("/:id", rbacmw.RequirePermission(h.perm, "ai.skills:update"), h.OverwriteSkill)
	skills.DELETE("/:id", rbacmw.RequirePermission(h.perm, "ai.skills:delete"), h.DeleteSkill)
	skills.GET("/:id/package", h.DownloadSkill)

	tokens := rg.Group("/tokens", authMW)
	tokens.GET("", h.ListPATs)
	tokens.POST("", h.CreatePAT)
	tokens.DELETE("/:id", h.DeletePAT)
}

func (h *Handler) ListCLIs(c *gin.Context) {
	items, err := h.cli.ListCLIs()
	if err != nil {
		pkg.Error(c, http.StatusInternalServerError, "查询 CLI 失败")
		return
	}
	pkg.Success(c, gin.H{"items": items, "risk_notice": "AI CLI 与构建脚本均以 Bedrock 进程同一操作系统用户直接执行，无 OS/容器沙箱隔离。"})
}

func (h *Handler) DetectCLI(c *gin.Context) {
	result, err := h.cli.Detect(c.Param("key"))
	if err != nil {
		writeErr(c, err)
		return
	}
	pkg.Success(c, result)
}

func (h *Handler) InstallCLI(c *gin.Context) {
	h.executeCLI(c, "install")
}
func (h *Handler) UpgradeCLI(c *gin.Context) {
	h.executeCLI(c, "upgrade")
}
func (h *Handler) UninstallCLI(c *gin.Context) {
	h.executeCLI(c, "uninstall")
}

func (h *Handler) executeCLI(c *gin.Context, op string) {
	var input service.ExecuteInput
	_ = c.ShouldBindJSON(&input)
	result, err := h.cli.Execute(c.Request.Context(), c.Param("key"), op, input, authmiddleware.GetUserID(c))
	if err != nil {
		writeErr(c, err)
		return
	}
	pkg.Success(c, result)
}

func (h *Handler) ListCLISources(c *gin.Context) {
	items, err := h.cli.ListSources(c.Query("cli_key"))
	if err != nil {
		pkg.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	pkg.Success(c, gin.H{"items": items})
}

func (h *Handler) CreateCLISource(c *gin.Context) {
	var input service.SourceInput
	if err := c.ShouldBindJSON(&input); err != nil {
		pkg.Error(c, http.StatusBadRequest, "无效请求")
		return
	}
	item, err := h.cli.CreateSource(input)
	if err != nil {
		writeErr(c, err)
		return
	}
	pkg.Created(c, item)
}

func (h *Handler) UpdateCLISource(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		pkg.Error(c, http.StatusBadRequest, "无效 ID")
		return
	}
	var input service.SourceInput
	if err := c.ShouldBindJSON(&input); err != nil {
		pkg.Error(c, http.StatusBadRequest, "无效请求")
		return
	}
	item, err := h.cli.UpdateSource(uint(id), input)
	if err != nil {
		writeErr(c, err)
		return
	}
	pkg.Success(c, item)
}

func (h *Handler) DeleteCLISource(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		pkg.Error(c, http.StatusBadRequest, "无效 ID")
		return
	}
	if err := h.cli.DeleteSource(uint(id)); err != nil {
		writeErr(c, err)
		return
	}
	pkg.Success(c, gin.H{"deleted": true})
}

func (h *Handler) ListAgents(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	items, total, err := h.agents.ListAgents(page, pageSize)
	if err != nil {
		pkg.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	pkg.Paginated(c, items, total, page, pageSize)
}

func (h *Handler) CreateAgent(c *gin.Context) {
	var input service.AgentInput
	if err := c.ShouldBindJSON(&input); err != nil {
		pkg.Error(c, http.StatusBadRequest, "无效请求")
		return
	}
	item, err := h.agents.CreateAgent(authmiddleware.GetUserID(c), input)
	if err != nil {
		writeErr(c, err)
		return
	}
	pkg.Created(c, item)
}

func (h *Handler) GetAgent(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	item, err := h.agents.GetAgent(uint(id))
	if err != nil {
		writeErr(c, err)
		return
	}
	pkg.Success(c, item)
}

func (h *Handler) UpdateAgent(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var input service.AgentInput
	if err := c.ShouldBindJSON(&input); err != nil {
		pkg.Error(c, http.StatusBadRequest, "无效请求")
		return
	}
	item, err := h.agents.UpdateAgent(uint(id), authmiddleware.GetUserID(c), input)
	if err != nil {
		writeErr(c, err)
		return
	}
	pkg.Success(c, item)
}

func (h *Handler) DeleteAgent(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := h.agents.DeleteAgent(uint(id), authmiddleware.GetUserID(c)); err != nil {
		writeErr(c, err)
		return
	}
	pkg.Success(c, gin.H{"deleted": true})
}

func (h *Handler) ListTriggers(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	items, err := h.agents.ListTriggers(uint(id))
	if err != nil {
		writeErr(c, err)
		return
	}
	pkg.Success(c, gin.H{"items": items})
}

func (h *Handler) CreateTrigger(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var input service.TriggerInput
	if err := c.ShouldBindJSON(&input); err != nil {
		pkg.Error(c, http.StatusBadRequest, "无效请求")
		return
	}
	item, err := h.agents.CreateTrigger(uint(id), authmiddleware.GetUserID(c), input)
	if err != nil {
		writeErr(c, err)
		return
	}
	pkg.Created(c, item)
}

func (h *Handler) UpdateTrigger(c *gin.Context) {
	tid, _ := strconv.ParseUint(c.Param("tid"), 10, 64)
	var input service.TriggerInput
	if err := c.ShouldBindJSON(&input); err != nil {
		pkg.Error(c, http.StatusBadRequest, "无效请求")
		return
	}
	item, err := h.agents.UpdateTrigger(uint(tid), authmiddleware.GetUserID(c), input)
	if err != nil {
		writeErr(c, err)
		return
	}
	pkg.Success(c, item)
}

func (h *Handler) DeleteTrigger(c *gin.Context) {
	tid, _ := strconv.ParseUint(c.Param("tid"), 10, 64)
	if err := h.agents.DeleteTrigger(uint(tid), authmiddleware.GetUserID(c)); err != nil {
		writeErr(c, err)
		return
	}
	pkg.Success(c, gin.H{"deleted": true})
}

func (h *Handler) ManualRun(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	run, err := h.agents.ManualRun(uint(id), authmiddleware.GetUserID(c))
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusAccepted, pkg.Response{Code: 0, Message: "accepted", Data: run})
}

func (h *Handler) APIRun(c *gin.Context) {
	// JWT needs ai.agents:execute; PAT needs agents:run scope.
	if authmiddleware.IsPAT(c) {
		if err := authmiddleware.RequirePATScope(c, "agents:run"); err != nil {
			pkg.Error(c, http.StatusForbidden, "token scope insufficient")
			return
		}
	} else if err := h.perm.CheckAccess(authmiddleware.GetUserID(c), authmiddleware.IsSuperAdmin(c), "ai.agents:execute"); err != nil {
		pkg.Error(c, http.StatusForbidden, "forbidden")
		return
	}
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	run, err := h.agents.APIRun(uint(id), authmiddleware.GetUserID(c))
	if err != nil {
		writeErr(c, err)
		return
	}
	c.JSON(http.StatusAccepted, pkg.Response{Code: 0, Message: "accepted", Data: run})
}

func (h *Handler) ListRuns(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	var agentID uint
	if raw := c.Query("agent_id"); raw != "" {
		v, _ := strconv.ParseUint(raw, 10, 64)
		agentID = uint(v)
	}
	items, total, err := h.agents.ListRuns(page, pageSize, agentID, c.Query("status"))
	if err != nil {
		pkg.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	pkg.Paginated(c, items, total, page, pageSize)
}

func (h *Handler) GetRun(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	run, err := h.agents.GetRun(uint(id))
	if err != nil {
		writeErr(c, err)
		return
	}
	pkg.Success(c, run)
}

func (h *Handler) DownloadRunArtifact(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	path, filename, err := h.agents.ArtifactPath(uint(id))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			pkg.Error(c, http.StatusNotFound, "资源不存在")
			return
		}
		pkg.Error(c, http.StatusNotFound, err.Error())
		return
	}
	c.FileAttachment(path, filename)
}

func (h *Handler) CancelRun(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := h.agents.CancelRun(uint(id)); err != nil {
		writeErr(c, err)
		return
	}
	pkg.Success(c, gin.H{"cancelled": true})
}

func (h *Handler) ListSkills(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	items, total, err := h.skills.List(page, pageSize, authmiddleware.GetUserID(c), authmiddleware.IsSuperAdmin(c))
	if err != nil {
		pkg.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	pkg.Paginated(c, items, total, page, pageSize)
}

func (h *Handler) GetSkill(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	item, err := h.skills.Get(uint(id), authmiddleware.GetUserID(c), authmiddleware.IsSuperAdmin(c))
	if err != nil {
		writeSkillErr(c, err)
		return
	}
	pkg.Success(c, item)
}

func (h *Handler) CreateSkill(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		pkg.Error(c, http.StatusBadRequest, "缺少 file")
		return
	}
	defer file.Close()
	item, err := h.skills.Create(service.SkillUploadInput{
		Name: c.PostForm("name"), Description: c.PostForm("description"),
		Visibility: defaultStr(c.PostForm("visibility"), "private"),
		Filename: header.Filename, ContentType: header.Header.Get("Content-Type"),
		Size: header.Size, Source: file, UserID: authmiddleware.GetUserID(c),
		IsSuperAdmin: authmiddleware.IsSuperAdmin(c),
	})
	if err != nil {
		writeSkillErr(c, err)
		return
	}
	pkg.Created(c, item)
}

func (h *Handler) OverwriteSkill(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		pkg.Error(c, http.StatusBadRequest, "缺少 file")
		return
	}
	defer file.Close()
	item, err := h.skills.Overwrite(uint(id), service.SkillUploadInput{
		Name: c.PostForm("name"), Description: c.PostForm("description"),
		Visibility: c.PostForm("visibility"),
		Filename: header.Filename, ContentType: header.Header.Get("Content-Type"),
		Size: header.Size, Source: file, UserID: authmiddleware.GetUserID(c),
		IsSuperAdmin: authmiddleware.IsSuperAdmin(c),
	})
	if err != nil {
		writeSkillErr(c, err)
		return
	}
	pkg.Success(c, item)
}

func (h *Handler) DeleteSkill(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := h.skills.Delete(uint(id), authmiddleware.GetUserID(c), authmiddleware.IsSuperAdmin(c)); err != nil {
		writeSkillErr(c, err)
		return
	}
	pkg.Success(c, gin.H{"deleted": true})
}

func (h *Handler) DownloadSkill(c *gin.Context) {
	if authmiddleware.IsPAT(c) {
		if err := authmiddleware.RequirePATScope(c, "skills:read"); err != nil {
			pkg.Error(c, http.StatusForbidden, "token scope insufficient")
			return
		}
	} else if err := h.perm.CheckAccess(authmiddleware.GetUserID(c), authmiddleware.IsSuperAdmin(c), "ai.skills:download"); err != nil {
		pkg.Error(c, http.StatusForbidden, "forbidden")
		return
	}
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	skill, rc, filename, err := h.skills.OpenPackage(uint(id), authmiddleware.GetUserID(c), authmiddleware.IsSuperAdmin(c))
	if err != nil {
		writeSkillErr(c, err)
		return
	}
	defer rc.Close()
	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Header("X-Skill-Digest", skill.PackageDigest)
	c.DataFromReader(http.StatusOK, skill.SizeBytes, "application/zip", rc, nil)
}

func (h *Handler) ListPATs(c *gin.Context) {
	page := pkg.ParsePage(c)
	items, total, err := h.pats.List(authmiddleware.GetUserID(c), page.Page, page.PageSize)
	if err != nil {
		pkg.Error(c, http.StatusInternalServerError, err.Error())
		return
	}
	pkg.Paginated(c, items, total, page.Page, page.PageSize)
}

func (h *Handler) CreatePAT(c *gin.Context) {
	var input service.CreatePATInput
	if err := c.ShouldBindJSON(&input); err != nil {
		pkg.Error(c, http.StatusBadRequest, "无效请求")
		return
	}
	result, err := h.pats.Create(authmiddleware.GetUserID(c), input)
	if err != nil {
		writeErr(c, err)
		return
	}
	pkg.Created(c, result)
}

func (h *Handler) DeletePAT(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := h.pats.Delete(authmiddleware.GetUserID(c), uint(id)); err != nil {
		writeErr(c, err)
		return
	}
	pkg.Success(c, gin.H{"deleted": true})
}

func writeErr(c *gin.Context, err error) {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		pkg.Error(c, http.StatusNotFound, "资源不存在")
		return
	}
	if errors.Is(err, service.ErrPATInvalid) {
		pkg.Error(c, http.StatusUnauthorized, err.Error())
		return
	}
	if errors.Is(err, service.ErrPATWrongScope) || errors.Is(err, service.ErrPATBadScope) {
		pkg.Error(c, http.StatusForbidden, err.Error())
		return
	}
	pkg.Error(c, http.StatusBadRequest, err.Error())
}

func writeSkillErr(c *gin.Context, err error) {
	if errors.Is(err, service.ErrMissingSkillMD) {
		pkg.Error(c, http.StatusUnprocessableEntity, err.Error())
		return
	}
	if errors.Is(err, service.ErrSkillForbidden) {
		pkg.Error(c, http.StatusForbidden, err.Error())
		return
	}
	if errors.Is(err, service.ErrSkillNotFound) {
		pkg.Error(c, http.StatusNotFound, err.Error())
		return
	}
	if errors.Is(err, storageservice.ErrTooLarge) {
		pkg.Error(c, http.StatusRequestEntityTooLarge, err.Error())
		return
	}
	writeErr(c, err)
}

func defaultStr(v, def string) string {
	if strings.TrimSpace(v) == "" {
		return def
	}
	return v
}
