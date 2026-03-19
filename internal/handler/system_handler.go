package handler

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"buildflow/internal/config"
	"buildflow/internal/pkg"
	"buildflow/internal/service"
)

type SystemHandler struct {
	auditService *service.AuditService
}

func NewSystemHandler(as *service.AuditService) *SystemHandler {
	return &SystemHandler{auditService: as}
}

// GET /api/v1/system/audit-logs - query audit logs with filters
func (h *SystemHandler) AuditLogs(c *gin.Context) {
	page, pageSize := pkg.GetPage(c)
	filters := &service.AuditListFilters{
		Page:     page,
		PageSize: pageSize,
	}
	if action := c.Query("action"); action != "" {
		filters.Action = action
	}
	if resourceType := c.Query("resource_type"); resourceType != "" {
		filters.ResourceType = resourceType
	}
	if userIDStr := c.Query("user_id"); userIDStr != "" {
		var userID uint
		if _, err := fmt.Sscanf(userIDStr, "%d", &userID); err == nil {
			filters.UserID = &userID
		}
	}
	if fromStr := c.Query("from"); fromStr != "" {
		if t, err := time.Parse("2006-01-02", fromStr); err == nil {
			filters.From = t
		}
	}
	if toStr := c.Query("to"); toStr != "" {
		if t, err := time.Parse("2006-01-02", toStr); err == nil {
			filters.To = t.Add(24*time.Hour - time.Nanosecond)
		}
	}
	logs, total, err := h.auditService.List(filters)
	if err != nil {
		pkg.Error(c, http.StatusInternalServerError, "查询失败")
		return
	}
	pkg.Paginated(c, logs, total, page, pageSize)
}

// POST /api/v1/system/backup - backup SQLite + config as tar.gz download
func (h *SystemHandler) Backup(c *gin.Context) {
	if config.C == nil {
		pkg.Error(c, http.StatusInternalServerError, "配置未加载")
		return
	}
	dbPath := config.C.Database.Path
	configPath := "config.yaml"
	for _, p := range []string{"config.yaml", "config/config.yaml", "./config.yaml"} {
		if _, err := os.Stat(p); err == nil {
			configPath = p
			break
		}
	}
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=buildflow-backup-%s.tar.gz", time.Now().Format("20060102-150405")))
	c.Header("Content-Type", "application/gzip")
	gw := gzip.NewWriter(c.Writer)
	defer gw.Close()
	tw := tar.NewWriter(gw)
	defer tw.Close()
	// Add db file
	if st, err := os.Stat(dbPath); err == nil && !st.IsDir() {
		f, err := os.Open(dbPath)
		if err == nil {
			defer f.Close()
			hdr := &tar.Header{
				Name: filepath.Base(dbPath),
				Mode: 0644,
				Size: st.Size(),
			}
			if err := tw.WriteHeader(hdr); err == nil {
				io.Copy(tw, f)
			}
		}
	}
	// Add config file
	if st, err := os.Stat(configPath); err == nil && !st.IsDir() {
		f, err := os.Open(configPath)
		if err == nil {
			defer f.Close()
			hdr := &tar.Header{
				Name: "config.yaml",
				Mode: 0644,
				Size: st.Size(),
			}
			if err := tw.WriteHeader(hdr); err == nil {
				io.Copy(tw, f)
			}
		}
	}
}

// POST /api/v1/system/restore - upload tar.gz and restore
func (h *SystemHandler) Restore(c *gin.Context) {
	if config.C == nil {
		pkg.Error(c, http.StatusInternalServerError, "配置未加载")
		return
	}
	file, _, err := c.Request.FormFile("file")
	if err != nil {
		pkg.Error(c, http.StatusBadRequest, "请上传备份文件")
		return
	}
	defer file.Close()
	gr, err := gzip.NewReader(file)
	if err != nil {
		pkg.Error(c, http.StatusBadRequest, "无效的 gzip 文件")
		return
	}
	defer gr.Close()
	tr := tar.NewReader(gr)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			pkg.Error(c, http.StatusBadRequest, "解析备份文件失败")
			return
		}
		if hdr.Typeflag != tar.TypeReg {
			continue
		}
		if hdr.Name == "config.yaml" || filepath.Base(hdr.Name) == "config.yaml" {
			configPath := "config.yaml"
			f, err := os.Create(configPath)
			if err != nil {
				pkg.Error(c, http.StatusInternalServerError, "无法写入配置文件")
				return
			}
			if _, err := io.Copy(f, tr); err != nil {
				f.Close()
				pkg.Error(c, http.StatusInternalServerError, "写入配置失败")
				return
			}
			f.Close()
		} else if filepath.Ext(hdr.Name) == ".sqlite" || hdr.Name == "db.sqlite" {
			dbPath := config.C.Database.Path
			dir := filepath.Dir(dbPath)
			if err := os.MkdirAll(dir, 0755); err != nil {
				pkg.Error(c, http.StatusInternalServerError, "无法创建数据目录")
				return
			}
			f, err := os.Create(dbPath)
			if err != nil {
				pkg.Error(c, http.StatusInternalServerError, "无法写入数据库文件")
				return
			}
			if _, err := io.Copy(f, tr); err != nil {
				f.Close()
				pkg.Error(c, http.StatusInternalServerError, "写入数据库失败")
				return
			}
			f.Close()
		}
	}
	pkg.Success(c, gin.H{"message": "恢复完成，请重启服务"})
}

// WorkspaceInfo represents disk usage for a project's workspace and cache.
type WorkspaceInfo struct {
	ProjectID     uint   `json:"project_id"`
	ProjectName   string `json:"project_name"`
	WorkspaceSize int64  `json:"workspace_size"`
	CacheSize     int64  `json:"cache_size"`
}

// GET /api/v1/system/workspaces - list workspace and cache disk usage per project
func (h *SystemHandler) ListWorkspaces(c *gin.Context) {
	if config.C == nil {
		pkg.Error(c, http.StatusInternalServerError, "配置未加载")
		return
	}

	workspaceDir := config.C.Build.WorkspaceDir
	cacheDir := config.C.Build.CacheDir

	// Collect project directories from workspace
	projectMap := make(map[uint]*WorkspaceInfo)

	// Scan workspace directory
	if entries, err := os.ReadDir(workspaceDir); err == nil {
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			var pid uint
			if _, err := fmt.Sscanf(entry.Name(), "project-%d", &pid); err != nil || pid == 0 {
				continue
			}
			info := &WorkspaceInfo{ProjectID: pid}
			info.WorkspaceSize = dirSize(filepath.Join(workspaceDir, entry.Name()))
			projectMap[pid] = info
		}
	}

	// Scan cache directory
	if cacheDir != "" {
		if entries, err := os.ReadDir(cacheDir); err == nil {
			for _, entry := range entries {
				if !entry.IsDir() {
					continue
				}
				var pid uint
				if _, err := fmt.Sscanf(entry.Name(), "project-%d", &pid); err != nil || pid == 0 {
					continue
				}
				if _, ok := projectMap[pid]; !ok {
					projectMap[pid] = &WorkspaceInfo{ProjectID: pid}
				}
				projectMap[pid].CacheSize = dirSize(filepath.Join(cacheDir, entry.Name()))
			}
		}
	}

	var result []WorkspaceInfo
	for _, info := range projectMap {
		result = append(result, *info)
	}

	pkg.Success(c, result)
}

// DELETE /api/v1/system/workspaces/:projectId - clean project workspace
func (h *SystemHandler) CleanWorkspace(c *gin.Context) {
	if config.C == nil {
		pkg.Error(c, http.StatusInternalServerError, "配置未加载")
		return
	}
	projectID, err := strconv.ParseUint(c.Param("projectId"), 10, 32)
	if err != nil {
		pkg.Error(c, http.StatusBadRequest, "参数错误")
		return
	}
	workspaceDir := filepath.Join(config.C.Build.WorkspaceDir, fmt.Sprintf("project-%d", projectID))
	if err := os.RemoveAll(workspaceDir); err != nil {
		pkg.Error(c, http.StatusInternalServerError, "清理工作空间失败: "+err.Error())
		return
	}
	pkg.Success(c, gin.H{"message": "工作空间已清理"})
}

// DELETE /api/v1/system/caches/:projectId - clean project build cache
func (h *SystemHandler) CleanCache(c *gin.Context) {
	if config.C == nil {
		pkg.Error(c, http.StatusInternalServerError, "配置未加载")
		return
	}
	projectID, err := strconv.ParseUint(c.Param("projectId"), 10, 32)
	if err != nil {
		pkg.Error(c, http.StatusBadRequest, "参数错误")
		return
	}
	cacheDir := filepath.Join(config.C.Build.CacheDir, fmt.Sprintf("project-%d", projectID))
	if err := os.RemoveAll(cacheDir); err != nil {
		pkg.Error(c, http.StatusInternalServerError, "清理缓存失败: "+err.Error())
		return
	}
	pkg.Success(c, gin.H{"message": "构建缓存已清理"})
}

// dirSize calculates the total size of a directory in bytes.
func dirSize(path string) int64 {
	var size int64
	filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	return size
}
