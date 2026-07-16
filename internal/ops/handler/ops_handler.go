package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	authmiddleware "bedrock/internal/auth/middleware"
	"bedrock/internal/ops/model"
	"bedrock/internal/ops/service"
	"bedrock/internal/pkg"
	rbacmw "bedrock/internal/rbac/middleware"
	rbacservice "bedrock/internal/rbac/service"
)

type OpsHandler struct {
	processes  *service.ProcessService
	toolchains *service.ToolchainService
	perm       *rbacservice.PermissionService
}

func NewOpsHandler(processes *service.ProcessService, toolchains *service.ToolchainService, perm *rbacservice.PermissionService) *OpsHandler {
	return &OpsHandler{processes: processes, toolchains: toolchains, perm: perm}
}

func (h *OpsHandler) RegisterRoutes(rg *gin.RouterGroup, authMW gin.HandlerFunc) {
	g := rg.Group("/ops", authMW)

	g.GET("/processes", rbacmw.RequirePermission(h.perm, "ops.processes:view"), h.ListProcesses)
	g.POST("/processes/:pid/kill", rbacmw.RequirePermission(h.perm, "ops.processes:execute"), h.KillProcess)

	g.GET("/toolchains", rbacmw.RequirePermission(h.perm, "ops.toolchains:view"), h.ListToolchains)
	g.POST("/toolchains", rbacmw.RequirePermission(h.perm, "ops.toolchains:create"), h.CreateToolchain)
	g.PUT("/toolchains/:id", rbacmw.RequirePermission(h.perm, "ops.toolchains:update"), h.UpdateToolchain)
	g.DELETE("/toolchains/:id", rbacmw.RequirePermission(h.perm, "ops.toolchains:delete"), h.DeleteToolchain)
	g.POST("/toolchains/:id/detect", rbacmw.RequirePermission(h.perm, "ops.toolchains:execute"), h.DetectToolchain)
	g.POST("/toolchains/:id/install", rbacmw.RequirePermission(h.perm, "ops.toolchains:execute"), h.EnqueueInstall)
	g.POST("/toolchains/:id/upgrade", rbacmw.RequirePermission(h.perm, "ops.toolchains:execute"), h.EnqueueUpgrade)
	g.POST("/toolchains/:id/uninstall", rbacmw.RequirePermission(h.perm, "ops.toolchains:execute"), h.EnqueueUninstall)
	g.POST("/toolchains/:id/switch", rbacmw.RequirePermission(h.perm, "ops.toolchains:execute"), h.EnqueueSwitch)

	g.GET("/install-sources", rbacmw.RequirePermission(h.perm, "ops.toolchains:view"), h.ListSources)
	g.POST("/install-sources", rbacmw.RequirePermission(h.perm, "ops.toolchains:create"), h.CreateSource)
	g.PUT("/install-sources/:id", rbacmw.RequirePermission(h.perm, "ops.toolchains:update"), h.UpdateSource)
	g.DELETE("/install-sources/:id", rbacmw.RequirePermission(h.perm, "ops.toolchains:delete"), h.DeleteSource)
	g.POST("/install-sources/:id/ping", rbacmw.RequirePermission(h.perm, "ops.toolchains:execute"), h.PingSource)

	g.GET("/install-jobs", rbacmw.RequirePermission(h.perm, "ops.toolchains:view"), h.ListJobs)
	g.GET("/install-jobs/:id", rbacmw.RequirePermission(h.perm, "ops.toolchains:view"), h.GetJob)
	g.GET("/install-jobs/:id/logs", rbacmw.RequirePermission(h.perm, "ops.toolchains:view"), h.JobLogs)
	g.POST("/install-jobs/:id/retry", rbacmw.RequirePermission(h.perm, "ops.toolchains:execute"), h.RetryJob)
}

func (h *OpsHandler) ListProcesses(c *gin.Context) {
	opts := model.ProcessListOptions{
		Keyword: c.Query("keyword"), Sort: c.Query("sort"), Order: c.Query("order"),
	}
	if raw := c.Query("pid"); raw != "" {
		pid, err := strconv.ParseInt(raw, 10, 32)
		if err != nil {
			pkg.Error(c, http.StatusBadRequest, "无效 PID")
			return
		}
		value := int32(pid)
		opts.PID = &value
	}
	if raw := c.Query("port"); raw != "" {
		port, err := strconv.ParseUint(raw, 10, 32)
		if err != nil {
			pkg.Error(c, http.StatusBadRequest, "无效端口")
			return
		}
		value := uint32(port)
		opts.Port = &value
	}
	items, err := h.processes.ListProcesses(opts)
	if err != nil {
		pkg.Error(c, http.StatusInternalServerError, "查询进程失败")
		return
	}
	pkg.Success(c, gin.H{"items": items})
}

func (h *OpsHandler) KillProcess(c *gin.Context) {
	pid, err := strconv.ParseInt(c.Param("pid"), 10, 32)
	if err != nil {
		pkg.Error(c, http.StatusBadRequest, "无效 PID")
		return
	}
	name, err := h.processes.KillProcess(int32(pid))
	if errors.Is(err, service.ErrKillSelf) || errors.Is(err, service.ErrDangerousProcess) {
		pkg.Error(c, http.StatusForbidden, err.Error())
		return
	}
	if err != nil {
		pkg.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	pkg.Success(c, gin.H{"pid": pid, "name": name, "status": "terminated"})
}

func (h *OpsHandler) ListToolchains(c *gin.Context) {
	items, err := h.toolchains.ListToolchains()
	if err != nil {
		pkg.Error(c, http.StatusInternalServerError, "查询工具链失败")
		return
	}
	pkg.Success(c, gin.H{"items": items})
}

func (h *OpsHandler) CreateToolchain(c *gin.Context) {
	var input service.ToolchainInput
	if err := c.ShouldBindJSON(&input); err != nil {
		pkg.Error(c, http.StatusBadRequest, "无效工具链定义")
		return
	}
	item, err := h.toolchains.CreateCustom(input, authmiddleware.GetUserID(c))
	if err != nil {
		pkg.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	pkg.Created(c, item)
}

func (h *OpsHandler) UpdateToolchain(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var input service.ToolchainInput
	if err := c.ShouldBindJSON(&input); err != nil {
		pkg.Error(c, http.StatusBadRequest, "无效工具链定义")
		return
	}
	item, err := h.toolchains.UpdateCustom(id, input)
	if err != nil {
		writeServiceError(c, err)
		return
	}
	pkg.Success(c, item)
}

func (h *OpsHandler) DeleteToolchain(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	if err := h.toolchains.DeleteCustom(id); err != nil {
		writeServiceError(c, err)
		return
	}
	pkg.Success(c, gin.H{"id": id})
}

func (h *OpsHandler) DetectToolchain(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	result, err := h.toolchains.Detect(id)
	if err != nil {
		writeServiceError(c, err)
		return
	}
	pkg.Success(c, result)
}

func (h *OpsHandler) EnqueueInstall(c *gin.Context)   { h.enqueue(c, "install") }
func (h *OpsHandler) EnqueueUpgrade(c *gin.Context)   { h.enqueue(c, "upgrade") }
func (h *OpsHandler) EnqueueUninstall(c *gin.Context) { h.enqueue(c, "uninstall") }
func (h *OpsHandler) EnqueueSwitch(c *gin.Context)    { h.enqueue(c, "switch") }

func (h *OpsHandler) enqueue(c *gin.Context, operation string) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var input service.JobInput
	if err := c.ShouldBindJSON(&input); err != nil {
		pkg.Error(c, http.StatusBadRequest, "无效任务参数")
		return
	}
	job, err := h.toolchains.Enqueue(id, operation, input, authmiddleware.GetUserID(c))
	if err != nil {
		writeServiceError(c, err)
		return
	}
	c.JSON(http.StatusAccepted, pkg.Response{Code: 0, Message: "accepted", Data: job})
}

func (h *OpsHandler) ListSources(c *gin.Context) {
	items, err := h.toolchains.ListSources()
	if err != nil {
		pkg.Error(c, http.StatusInternalServerError, "查询安装源失败")
		return
	}
	pkg.Success(c, gin.H{"items": items})
}

func (h *OpsHandler) CreateSource(c *gin.Context) {
	var input service.SourceInput
	if err := c.ShouldBindJSON(&input); err != nil {
		pkg.Error(c, http.StatusBadRequest, "无效安装源")
		return
	}
	item, err := h.toolchains.CreateSource(input)
	if err != nil {
		pkg.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	pkg.Created(c, item)
}

func (h *OpsHandler) UpdateSource(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var input service.SourceInput
	if err := c.ShouldBindJSON(&input); err != nil {
		pkg.Error(c, http.StatusBadRequest, "无效安装源")
		return
	}
	item, err := h.toolchains.UpdateSource(id, input)
	if err != nil {
		writeServiceError(c, err)
		return
	}
	pkg.Success(c, item)
}

func (h *OpsHandler) DeleteSource(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	if err := h.toolchains.DeleteSource(id); err != nil {
		writeServiceError(c, err)
		return
	}
	pkg.Success(c, gin.H{"id": id})
}

func (h *OpsHandler) PingSource(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	okResult, detail, err := h.toolchains.PingSource(id)
	if err != nil {
		writeServiceError(c, err)
		return
	}
	pkg.Success(c, gin.H{"ok": okResult, "detail": detail})
}

func (h *OpsHandler) ListJobs(c *gin.Context) {
	page := pkg.ParsePage(c)
	items, total, err := h.toolchains.ListJobs(page.Page, page.PageSize, c.Query("status"))
	if err != nil {
		pkg.Error(c, http.StatusInternalServerError, "查询安装任务失败")
		return
	}
	pkg.PageSuccess(c, items, total, page)
}

func (h *OpsHandler) GetJob(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	item, err := h.toolchains.GetJob(id)
	if err != nil {
		writeServiceError(c, err)
		return
	}
	pkg.Success(c, item)
}

func (h *OpsHandler) JobLogs(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	logs, err := h.toolchains.JobLogs(id)
	if err != nil {
		writeServiceError(c, err)
		return
	}
	c.Data(http.StatusOK, "text/plain; charset=utf-8", []byte(logs))
}

func (h *OpsHandler) RetryJob(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	job, err := h.toolchains.Retry(id, authmiddleware.GetUserID(c))
	if err != nil {
		writeServiceError(c, err)
		return
	}
	c.JSON(http.StatusAccepted, pkg.Response{Code: 0, Message: "accepted", Data: job})
}

func parseID(c *gin.Context) (uint, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		pkg.Error(c, http.StatusBadRequest, "无效 ID")
		return 0, false
	}
	return uint(id), true
}

func writeServiceError(c *gin.Context, err error) {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		pkg.Error(c, http.StatusNotFound, "资源不存在")
		return
	}
	if errors.Is(err, service.ErrBuiltinImmutable) || errors.Is(err, service.ErrMissingTemplate) ||
		errors.Is(err, service.ErrInvalidOperation) {
		pkg.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	pkg.Error(c, http.StatusBadRequest, err.Error())
}
