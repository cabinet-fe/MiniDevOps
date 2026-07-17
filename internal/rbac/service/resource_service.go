package service

import (
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"bedrock/internal/rbac"
	"bedrock/internal/rbac/model"
	"bedrock/internal/rbac/repository"
)

type ResourceService struct {
	resources *repository.ResourceRepository
}

func NewResourceService(resources *repository.ResourceRepository) *ResourceService {
	return &ResourceService{resources: resources}
}

type CreateResourceInput struct {
	Path     string `json:"path"`
	Type     string `json:"type"`
	ParentID *uint  `json:"parent_id"`
	Enabled  *bool  `json:"enabled"`
	SortKey  int    `json:"sort_key"`
	Title    string `json:"title"`
	Route    string `json:"route"`
}

func (s *ResourceService) Create(in CreateResourceInput) (*model.RbacResource, error) {
	path := strings.TrimSpace(in.Path)
	typ := strings.TrimSpace(in.Type)
	if path == "" || typ == "" {
		return nil, errors.New("path 与 type 不能为空")
	}
	if !validResourceType(typ) {
		return nil, errors.New("type 必须为 menu|page|action|card")
	}
	enabled := true
	if in.Enabled != nil {
		enabled = *in.Enabled
	}
	res := &model.RbacResource{
		Path:     path,
		Type:     typ,
		ParentID: in.ParentID,
		Enabled:  enabled,
		SortKey:  in.SortKey,
	}
	if err := s.resources.Create(res); err != nil {
		return nil, fmt.Errorf("创建资源失败: %w", err)
	}
	if typ == model.ResourceTypeMenu {
		title := strings.TrimSpace(in.Title)
		if title == "" {
			title = path
		}
		meta := &model.MenuMetadata{
			ResourceID: res.ID,
			Title:      title,
			Route:      strings.TrimSpace(in.Route),
		}
		if err := s.resources.UpsertMenuMetadata(meta); err != nil {
			return nil, err
		}
	}
	return s.resources.FindByID(res.ID)
}

func (s *ResourceService) Get(id uint) (*model.RbacResource, error) {
	return s.resources.FindByID(id)
}

// ListResourcesFilter filters the RBAC resource tree.
// Matching nodes keep their ancestors so the response remains a valid tree.
type ListResourcesFilter struct {
	Keyword string
	Type    string
	Enabled *bool
}

func (f ListResourcesFilter) active() bool {
	return strings.TrimSpace(f.Keyword) != "" || strings.TrimSpace(f.Type) != "" || f.Enabled != nil
}

func (s *ResourceService) ListTree(filter ListResourcesFilter) ([]model.RbacResource, error) {
	items, err := s.resources.ListAll()
	if err != nil {
		return nil, err
	}
	if filter.active() {
		if typ := strings.TrimSpace(filter.Type); typ != "" && !validResourceType(typ) {
			return nil, errors.New("type 必须为 menu|page|action|card")
		}
		items = filterResourcesWithAncestors(items, filter)
	}
	return buildResourceTree(items), nil
}

func (s *ResourceService) ListMenusTree() ([]model.RbacResource, error) {
	items, err := s.resources.ListMenus()
	if err != nil {
		return nil, err
	}
	return buildResourceTree(items), nil
}

type UpdateResourceInput struct {
	Enabled *bool  `json:"enabled"`
	SortKey *int   `json:"sort_key"`
	Title   string `json:"title"`
	Route   string `json:"route"`
}

func (s *ResourceService) Update(id uint, in UpdateResourceInput) (*model.RbacResource, error) {
	res, err := s.resources.FindByID(id)
	if err != nil {
		return nil, err
	}
	if in.Enabled != nil {
		res.Enabled = *in.Enabled
	}
	if in.SortKey != nil {
		res.SortKey = *in.SortKey
	}
	if err := s.resources.Update(res); err != nil {
		return nil, err
	}
	if res.Type == model.ResourceTypeMenu {
		meta := res.MenuMetadata
		if meta == nil {
			meta = &model.MenuMetadata{ResourceID: res.ID}
		}
		if t := strings.TrimSpace(in.Title); t != "" {
			meta.Title = t
		}
		if in.Route != "" || meta.Route != "" {
			if in.Route != "" || strings.TrimSpace(in.Title) != "" {
				// allow clearing route only when explicitly sent as empty with title update;
				// keep existing if both empty in request beyond title.
			}
			meta.Route = strings.TrimSpace(in.Route)
		}
		if meta.Title == "" {
			meta.Title = res.Path
		}
		if err := s.resources.UpsertMenuMetadata(meta); err != nil {
			return nil, err
		}
	}
	return s.resources.FindByID(id)
}

func (s *ResourceService) Delete(id uint) error {
	n, err := s.resources.CountChildren(id)
	if err != nil {
		return err
	}
	if n > 0 {
		return errors.New("请先删除子资源")
	}
	return s.resources.Delete(id)
}

// UpdateMenuIcon stores Base64 icon; rejects when raw decoded size > 32KB.
func (s *ResourceService) UpdateMenuIcon(resourceID uint, iconBase64, iconMime string) (*model.RbacResource, error) {
	res, err := s.resources.FindByID(resourceID)
	if err != nil {
		return nil, err
	}
	if res.Type != model.ResourceTypeMenu {
		return nil, errors.New("仅菜单资源可设置图标")
	}
	if res.ParentID != nil {
		return nil, errors.New("仅一级菜单可上传图标")
	}
	raw, mime, err := decodeIconPayload(iconBase64, iconMime)
	if err != nil {
		return nil, err
	}
	if len(raw) > rbac.MaxMenuIconBytes {
		return nil, fmt.Errorf("图标原始体积不得超过 32KB（当前 %d 字节）", len(raw))
	}
	stored := base64.StdEncoding.EncodeToString(raw)
	meta := &model.MenuMetadata{
		ResourceID: resourceID,
		Title:      res.Path,
		IconBase64: stored,
		IconMime:   mime,
	}
	if res.MenuMetadata != nil {
		meta.Title = res.MenuMetadata.Title
		meta.Route = res.MenuMetadata.Route
		meta.ID = res.MenuMetadata.ID
	}
	if err := s.resources.UpsertMenuMetadata(meta); err != nil {
		return nil, err
	}
	return s.resources.FindByID(resourceID)
}

func decodeIconPayload(iconBase64, iconMime string) ([]byte, string, error) {
	payload := strings.TrimSpace(iconBase64)
	mime := strings.TrimSpace(iconMime)
	if payload == "" {
		return nil, "", errors.New("图标不能为空")
	}
	if strings.HasPrefix(payload, "data:") {
		parts := strings.SplitN(payload, ",", 2)
		if len(parts) != 2 {
			return nil, "", errors.New("无效的 data URL")
		}
		header := parts[0]
		payload = parts[1]
		if mime == "" {
			// data:image/png;base64
			header = strings.TrimPrefix(header, "data:")
			header = strings.TrimSuffix(header, ";base64")
			mime = header
		}
	}
	raw, err := base64.StdEncoding.DecodeString(payload)
	if err != nil {
		// try raw URL encoding variant
		raw, err = base64.RawStdEncoding.DecodeString(payload)
		if err != nil {
			return nil, "", errors.New("图标 Base64 解码失败")
		}
	}
	if mime == "" {
		mime = "image/png"
	}
	return raw, mime, nil
}

func validResourceType(t string) bool {
	switch t {
	case model.ResourceTypeMenu, model.ResourceTypePage, model.ResourceTypeAction, model.ResourceTypeCard:
		return true
	default:
		return false
	}
}

func resourceMatchesFilter(item model.RbacResource, filter ListResourcesFilter) bool {
	if typ := strings.TrimSpace(filter.Type); typ != "" && item.Type != typ {
		return false
	}
	if filter.Enabled != nil && item.Enabled != *filter.Enabled {
		return false
	}
	keyword := strings.TrimSpace(filter.Keyword)
	if keyword == "" {
		return true
	}
	kw := strings.ToLower(keyword)
	if strings.Contains(strings.ToLower(item.Path), kw) {
		return true
	}
	if item.MenuMetadata != nil {
		if strings.Contains(strings.ToLower(item.MenuMetadata.Title), kw) {
			return true
		}
		if strings.Contains(strings.ToLower(item.MenuMetadata.Route), kw) {
			return true
		}
	}
	return false
}

func filterResourcesWithAncestors(items []model.RbacResource, filter ListResourcesFilter) []model.RbacResource {
	byID := make(map[uint]model.RbacResource, len(items))
	for _, item := range items {
		byID[item.ID] = item
	}
	keep := make(map[uint]struct{})
	for _, item := range items {
		if !resourceMatchesFilter(item, filter) {
			continue
		}
		cur := item.ID
		for {
			if _, ok := keep[cur]; ok {
				break
			}
			keep[cur] = struct{}{}
			node, ok := byID[cur]
			if !ok || node.ParentID == nil {
				break
			}
			cur = *node.ParentID
		}
	}
	out := make([]model.RbacResource, 0, len(keep))
	for _, item := range items {
		if _, ok := keep[item.ID]; ok {
			out = append(out, item)
		}
	}
	return out
}

func buildResourceTree(items []model.RbacResource) []model.RbacResource {
	byID := make(map[uint]*model.RbacResource, len(items))
	for i := range items {
		cp := items[i]
		cp.Children = nil
		byID[cp.ID] = &cp
	}
	childrenOf := map[uint][]uint{}
	var rootIDs []uint
	for _, item := range items {
		if item.ParentID == nil {
			rootIDs = append(rootIDs, item.ID)
			continue
		}
		if _, ok := byID[*item.ParentID]; !ok {
			rootIDs = append(rootIDs, item.ID)
			continue
		}
		childrenOf[*item.ParentID] = append(childrenOf[*item.ParentID], item.ID)
	}
	var attach func(id uint) model.RbacResource
	attach = func(id uint) model.RbacResource {
		n := *byID[id]
		for _, cid := range childrenOf[id] {
			n.Children = append(n.Children, attach(cid))
		}
		return n
	}
	out := make([]model.RbacResource, 0, len(rootIDs))
	for _, id := range rootIDs {
		out = append(out, attach(id))
	}
	return out
}
