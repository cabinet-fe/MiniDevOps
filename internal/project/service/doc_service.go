package service

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"path"
	"strings"
	"time"

	projectmodel "bedrock/internal/project/model"
	storagemodel "bedrock/internal/storage/model"

	"gorm.io/gorm"
)

const (
	maxZIPEntries = 1000
	maxZIPRatio   = 100
)

type DocNodeInput struct {
	ParentID     *uint   `json:"parent_id"`
	Kind         string  `json:"kind"`
	Name         string  `json:"name"`
	SortOrder    int     `json:"sort_order"`
	RepositoryID *uint   `json:"repository_id"`
	DraftContent *string `json:"draft_content"`
}

type DocMoveInput struct {
	ParentID  *uint `json:"parent_id"`
	SortOrder int   `json:"sort_order"`
}

type DocDiff struct {
	NodeID         uint `json:"node_id"`
	ContentVersion int  `json:"content_version"`
	HasDraft       bool `json:"has_draft"`
	PublishedLines int  `json:"published_lines"`
	DraftLines     int  `json:"draft_lines"`
	AddedLines     int  `json:"added_lines"`
	RemovedLines   int  `json:"removed_lines"`
}

func (s *ProjectService) ListDocTree(actor AccessContext, projectID uint) ([]projectmodel.ApiDocNode, error) {
	if _, err := s.acl.Require(projectID, actor, "project.docs:view", capDocView); err != nil {
		return nil, err
	}
	nodes, err := s.repo.ListDocNodes(projectID)
	if err != nil {
		return nil, err
	}
	return buildDocTree(nodes), nil
}

func (s *ProjectService) GetDocNode(actor AccessContext, id uint) (*projectmodel.ApiDocNode, error) {
	node, err := s.repo.FindDocNode(id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, NewNotFound("文档节点不存在")
	}
	if err != nil {
		return nil, err
	}
	if _, err := s.acl.Require(node.ProjectID, actor, "project.docs:view", capDocView); err != nil {
		return nil, err
	}
	return node, nil
}

// CheckDocProject validates nested document routes without imposing a separate
// global :view grant on callers that only hold a write permission.
func (s *ProjectService) CheckDocProject(actor AccessContext, projectID, nodeID uint, globalPermission string, write bool) error {
	node, err := s.repo.FindDocNode(nodeID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return NewNotFound("文档节点不存在")
	}
	if err != nil {
		return err
	}
	if node.ProjectID != projectID {
		return NewNotFound("文档节点不存在")
	}
	capability := capDocView
	if write {
		capability = capDocEdit
	}
	_, err = s.acl.Require(projectID, actor, globalPermission, capability)
	return err
}

func (s *ProjectService) CreateDocNode(actor AccessContext, projectID uint, input DocNodeInput) (*projectmodel.ApiDocNode, error) {
	if _, err := s.acl.Require(projectID, actor, "project.docs:create", capDocEdit); err != nil {
		return nil, err
	}
	if err := s.requireActiveProject(projectID); err != nil {
		return nil, err
	}
	kind := strings.ToLower(strings.TrimSpace(input.Kind))
	if kind != projectmodel.DocNodeDirectory && kind != projectmodel.DocNodeDocument {
		return nil, errors.New("节点类型必须为 dir 或 doc")
	}
	name := safeDocName(input.Name)
	if name == "" {
		return nil, errors.New("文档节点名称不能为空")
	}
	if err := s.validateDocParent(projectID, input.ParentID); err != nil {
		return nil, err
	}
	node := &projectmodel.ApiDocNode{
		ProjectID: projectID, ParentID: input.ParentID, Kind: kind, Name: name, SortOrder: input.SortOrder,
		RepositoryID: input.RepositoryID, CreatedBy: actor.UserID, UpdatedBy: actor.UserID,
	}
	if kind == projectmodel.DocNodeDocument && input.DraftContent != nil {
		now := time.Now().UTC()
		node.DraftContent = *input.DraftContent
		node.DraftBaseVersion = 0
		node.DraftUpdatedAt = &now
	}
	if err := s.repo.CreateDocNode(node); err != nil {
		return nil, err
	}
	return node, nil
}

func (s *ProjectService) UpdateDocNode(actor AccessContext, id uint, input DocNodeInput) (*projectmodel.ApiDocNode, error) {
	node, err := s.repo.FindDocNode(id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, NewNotFound("文档节点不存在")
	}
	if err != nil {
		return nil, err
	}
	if _, err := s.acl.Require(node.ProjectID, actor, "project.docs:update", capDocEdit); err != nil {
		return nil, err
	}
	if err := s.requireActiveProject(node.ProjectID); err != nil {
		return nil, err
	}
	if strings.TrimSpace(input.Name) != "" {
		name := safeDocName(input.Name)
		if name == "" {
			return nil, errors.New("文档节点名称不能为空")
		}
		node.Name = name
	}
	if input.RepositoryID != nil {
		node.RepositoryID = input.RepositoryID
	}
	if input.DraftContent != nil {
		if node.Kind != projectmodel.DocNodeDocument {
			return nil, errors.New("目录不能编辑 Markdown 内容")
		}
		s.writeDraft(node, *input.DraftContent, actor.UserID)
	}
	node.UpdatedBy = actor.UserID
	if err := s.repo.UpdateDocNode(node); err != nil {
		return nil, err
	}
	return node, nil
}

func (s *ProjectService) MoveDocNode(actor AccessContext, id uint, input DocMoveInput) (*projectmodel.ApiDocNode, error) {
	node, err := s.repo.FindDocNode(id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, NewNotFound("文档节点不存在")
	}
	if err != nil {
		return nil, err
	}
	if _, err := s.acl.Require(node.ProjectID, actor, "project.docs:update", capDocEdit); err != nil {
		return nil, err
	}
	if err := s.requireActiveProject(node.ProjectID); err != nil {
		return nil, err
	}
	if err := s.validateDocParent(node.ProjectID, input.ParentID); err != nil {
		return nil, err
	}
	if input.ParentID != nil && *input.ParentID == node.ID {
		return nil, errors.New("节点不能移动到自身")
	}
	nodes, err := s.repo.ListDocNodes(node.ProjectID)
	if err != nil {
		return nil, err
	}
	descendants := docSubtreeIDs(nodes, node.ID)
	if input.ParentID != nil {
		for _, descendantID := range descendants {
			if descendantID == *input.ParentID {
				return nil, errors.New("节点不能移动到自己的子节点")
			}
		}
	}
	node.ParentID = input.ParentID
	node.SortOrder = input.SortOrder
	node.UpdatedBy = actor.UserID
	if err := s.repo.UpdateDocNode(node); err != nil {
		return nil, err
	}
	return node, nil
}

func (s *ProjectService) DeleteDocNode(actor AccessContext, id uint) error {
	node, err := s.repo.FindDocNode(id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return NewNotFound("文档节点不存在")
	}
	if err != nil {
		return err
	}
	if _, err := s.acl.Require(node.ProjectID, actor, "project.docs:delete", capDocAdmin); err != nil {
		return err
	}
	nodes, err := s.repo.ListDocNodes(node.ProjectID)
	if err != nil {
		return err
	}
	return s.repo.DeleteDocNodes(docSubtreeIDs(nodes, id))
}

func (s *ProjectService) UploadMarkdown(actor AccessContext, projectID uint, parentID *uint, filename, contentType string, source io.Reader, size int64) (*projectmodel.ApiDocNode, error) {
	if _, err := s.acl.Require(projectID, actor, "project.docs:create", capDocEdit); err != nil {
		return nil, err
	}
	if err := s.requireActiveProject(projectID); err != nil {
		return nil, err
	}
	if !strings.EqualFold(path.Ext(filename), ".md") {
		return nil, errors.New("仅支持上传 .md 文件")
	}
	if err := s.validateDocParent(projectID, parentID); err != nil {
		return nil, err
	}
	object, err := s.storage.Put(storagemodel.KindDocImport, contentType, source, size, actor.UserID)
	if err != nil {
		return nil, err
	}
	defer s.storage.Delete(object.ID)
	file, _, err := s.storage.Open(object.ID)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	content, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}
	return s.createImportedDocument(projectID, parentID, safeDocName(path.Base(filename)), string(content), actor.UserID)
}

func (s *ProjectService) ImportZIP(actor AccessContext, projectID uint, parentID *uint, filename, contentType string, source io.Reader, size int64) ([]projectmodel.ApiDocNode, error) {
	if _, err := s.acl.Require(projectID, actor, "project.docs:create", capDocEdit); err != nil {
		return nil, err
	}
	if err := s.requireActiveProject(projectID); err != nil {
		return nil, err
	}
	if !strings.EqualFold(path.Ext(filename), ".zip") {
		return nil, errors.New("仅支持上传 .zip 文件")
	}
	if err := s.validateDocParent(projectID, parentID); err != nil {
		return nil, err
	}
	object, err := s.storage.Put(storagemodel.KindDocImport, contentType, source, size, actor.UserID)
	if err != nil {
		return nil, err
	}
	defer s.storage.Delete(object.ID)
	file, _, err := s.storage.Open(object.ID)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	reader, err := zip.NewReader(file, object.Size)
	if err != nil {
		return nil, errors.New("无效 ZIP 文件")
	}
	if len(reader.File) > maxZIPEntries {
		return nil, errors.New("ZIP 条目数超过限制")
	}
	var totalUncompressed uint64
	for _, entry := range reader.File {
		if err := validateZIPEntry(entry); err != nil {
			return nil, err
		}
		totalUncompressed += entry.UncompressedSize64
		if totalUncompressed > uint64(s.storage.MaxBytes(storagemodel.KindDocImport)) {
			return nil, errors.New("ZIP 解压后内容超过限制")
		}
	}

	nodes, err := s.repo.ListDocNodes(projectID)
	if err != nil {
		return nil, err
	}
	index := newDocImportIndex(nodes, parentID)
	imported := make([]projectmodel.ApiDocNode, 0)
	for _, entry := range reader.File {
		clean, err := cleanZIPPath(entry.Name)
		if err != nil {
			return nil, err
		}
		if entry.FileInfo().IsDir() || !strings.EqualFold(path.Ext(clean), ".md") {
			continue
		}
		parts := strings.Split(clean, "/")
		documentName := safeDocName(parts[len(parts)-1])
		currentParent := parentID
		currentKey := importParentPath(nodes, parentID)
		for _, directory := range parts[:len(parts)-1] {
			key := currentKey + "/" + directory
			if existing, ok := index[key]; ok {
				if existing.Kind != projectmodel.DocNodeDirectory {
					return nil, fmt.Errorf("导入路径与文档节点冲突: %s", directory)
				}
				id := existing.ID
				currentParent = &id
				currentKey = key
				continue
			}
			directoryNode := &projectmodel.ApiDocNode{
				ProjectID: projectID, ParentID: currentParent, Kind: projectmodel.DocNodeDirectory, Name: directory,
				CreatedBy: actor.UserID, UpdatedBy: actor.UserID,
			}
			if err := s.repo.CreateDocNode(directoryNode); err != nil {
				return nil, err
			}
			index[key] = *directoryNode
			id := directoryNode.ID
			currentParent = &id
			currentKey = key
		}
		content, err := readZIPEntry(entry, s.storage.MaxBytes(storagemodel.KindDocImport))
		if err != nil {
			return nil, err
		}
		documentKey := currentKey + "/" + documentName
		if existing, ok := index[documentKey]; ok {
			if existing.Kind != projectmodel.DocNodeDocument {
				return nil, fmt.Errorf("导入路径与目录节点冲突: %s", documentName)
			}
			node := existing
			s.writeDraft(&node, string(content), actor.UserID)
			if err := s.repo.UpdateDocNode(&node); err != nil {
				return nil, err
			}
			index[documentKey] = node
			imported = append(imported, node)
			continue
		}
		node, err := s.createImportedDocument(projectID, currentParent, documentName, string(content), actor.UserID)
		if err != nil {
			return nil, err
		}
		index[documentKey] = *node
		imported = append(imported, *node)
	}
	return imported, nil
}

func (s *ProjectService) PublishDocNode(actor AccessContext, id uint, expectedVersion int) (*projectmodel.ApiDocNode, error) {
	node, err := s.repo.FindDocNode(id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, NewNotFound("文档节点不存在")
	}
	if err != nil {
		return nil, err
	}
	if _, err := s.acl.Require(node.ProjectID, actor, "project.docs:update", capDocEdit); err != nil {
		return nil, err
	}
	if node.Kind != projectmodel.DocNodeDocument {
		return nil, errors.New("目录不能发布")
	}
	if node.DraftUpdatedAt == nil {
		return nil, errors.New("没有待发布草稿")
	}
	if expectedVersion != node.ContentVersion {
		return nil, NewConflict("文档版本冲突，请刷新后重试")
	}
	published, err := s.repo.PublishDocNode(id, expectedVersion)
	if err != nil {
		return nil, err
	}
	if !published {
		return nil, NewConflict("文档版本冲突，请刷新后重试")
	}
	return s.repo.FindDocNode(id)
}

func (s *ProjectService) GetDocDiff(actor AccessContext, id uint) (*DocDiff, error) {
	node, err := s.GetDocNode(actor, id)
	if err != nil {
		return nil, err
	}
	if node.Kind != projectmodel.DocNodeDocument {
		return nil, errors.New("目录没有 Markdown 差异")
	}
	draft := node.DraftContent
	return &DocDiff{
		NodeID: node.ID, ContentVersion: node.ContentVersion, HasDraft: node.DraftUpdatedAt != nil,
		PublishedLines: lineCount(node.PublishedContent), DraftLines: lineCount(draft),
		AddedLines:   changedLineCount(node.PublishedContent, draft),
		RemovedLines: changedLineCount(draft, node.PublishedContent),
	}, nil
}

func (s *ProjectService) GenerateDocs(actor AccessContext, projectID uint) error {
	if _, err := s.acl.Require(projectID, actor, "project.docs:execute", capDocEdit); err != nil {
		return err
	}
	return ErrAIDomainUnavailable
}

func (s *ProjectService) validateDocParent(projectID uint, parentID *uint) error {
	if parentID == nil {
		return nil
	}
	parent, err := s.repo.FindDocNode(*parentID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return NewNotFound("父节点不存在")
	}
	if err != nil {
		return err
	}
	if parent.ProjectID != projectID {
		return errors.New("父节点不属于当前项目")
	}
	if parent.Kind != projectmodel.DocNodeDirectory {
		return errors.New("父节点必须为目录")
	}
	return nil
}

func (s *ProjectService) writeDraft(node *projectmodel.ApiDocNode, content string, userID uint) {
	now := time.Now().UTC()
	node.DraftContent = content
	node.DraftBaseVersion = node.ContentVersion
	node.DraftUpdatedAt = &now
	node.DraftSourceRunID = nil
	node.UpdatedBy = userID
}

func (s *ProjectService) createImportedDocument(projectID uint, parentID *uint, name, content string, userID uint) (*projectmodel.ApiDocNode, error) {
	if name == "" {
		return nil, errors.New("无效 Markdown 文件名")
	}
	now := time.Now().UTC()
	node := &projectmodel.ApiDocNode{
		ProjectID: projectID, ParentID: parentID, Kind: projectmodel.DocNodeDocument, Name: name,
		DraftContent: content, DraftBaseVersion: 0, DraftUpdatedAt: &now, CreatedBy: userID, UpdatedBy: userID,
	}
	if err := s.repo.CreateDocNode(node); err != nil {
		return nil, err
	}
	return node, nil
}

func buildDocTree(nodes []projectmodel.ApiDocNode) []projectmodel.ApiDocNode {
	byID := make(map[uint]*projectmodel.ApiDocNode, len(nodes))
	for i := range nodes {
		nodes[i].Children = nil
		byID[nodes[i].ID] = &nodes[i]
	}
	roots := make([]projectmodel.ApiDocNode, 0)
	for _, node := range nodes {
		current := byID[node.ID]
		if current.ParentID == nil {
			roots = append(roots, *current)
			continue
		}
		parent, ok := byID[*current.ParentID]
		if !ok {
			roots = append(roots, *current)
			continue
		}
		parent.Children = append(parent.Children, *current)
	}
	var materialize func(projectmodel.ApiDocNode) projectmodel.ApiDocNode
	materialize = func(node projectmodel.ApiDocNode) projectmodel.ApiDocNode {
		if current, ok := byID[node.ID]; ok {
			node = *current
		}
		for i := range node.Children {
			node.Children[i] = materialize(node.Children[i])
		}
		return node
	}
	for i := range roots {
		roots[i] = materialize(roots[i])
	}
	return roots
}

func docSubtreeIDs(nodes []projectmodel.ApiDocNode, rootID uint) []uint {
	children := make(map[uint][]uint)
	for _, node := range nodes {
		if node.ParentID != nil {
			children[*node.ParentID] = append(children[*node.ParentID], node.ID)
		}
	}
	ids := []uint{rootID}
	for index := 0; index < len(ids); index++ {
		ids = append(ids, children[ids[index]]...)
	}
	return ids
}

func safeDocName(value string) string {
	name := strings.TrimSpace(path.Base(strings.ReplaceAll(value, "\\", "/")))
	if name == "" || name == "." || name == ".." || strings.Contains(name, "\x00") {
		return ""
	}
	return name
}

func validateZIPEntry(entry *zip.File) error {
	clean, err := cleanZIPPath(entry.Name)
	if err != nil {
		return err
	}
	_ = clean
	if entry.UncompressedSize64 > 0 && entry.CompressedSize64 == 0 {
		return errors.New("ZIP 条目压缩比异常")
	}
	if entry.CompressedSize64 > 0 && entry.UncompressedSize64/entry.CompressedSize64 > maxZIPRatio {
		return errors.New("ZIP 条目压缩比超过限制")
	}
	return nil
}

func cleanZIPPath(value string) (string, error) {
	if value == "" || strings.Contains(value, "\x00") || strings.HasPrefix(value, "/") || strings.HasPrefix(value, "\\") {
		return "", errors.New("ZIP 包含非法路径")
	}
	clean := path.Clean(strings.ReplaceAll(value, "\\", "/"))
	if clean == "." || clean == ".." || strings.HasPrefix(clean, "../") {
		return "", errors.New("ZIP 包含路径穿越")
	}
	return clean, nil
}

func readZIPEntry(entry *zip.File, maxBytes int64) ([]byte, error) {
	reader, err := entry.Open()
	if err != nil {
		return nil, err
	}
	defer reader.Close()
	data, err := io.ReadAll(io.LimitReader(reader, maxBytes+1))
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > maxBytes {
		return nil, errors.New("ZIP 条目超过限制")
	}
	return data, nil
}

func newDocImportIndex(nodes []projectmodel.ApiDocNode, rootParent *uint) map[string]projectmodel.ApiDocNode {
	index := map[string]projectmodel.ApiDocNode{}
	byID := make(map[uint]projectmodel.ApiDocNode, len(nodes))
	for _, node := range nodes {
		byID[node.ID] = node
	}
	var nodePath func(uint) string
	nodePath = func(id uint) string {
		node, ok := byID[id]
		if !ok {
			return "root"
		}
		parentPath := "root"
		if node.ParentID != nil {
			parentPath = nodePath(*node.ParentID)
		}
		return parentPath + "/" + node.Name
	}
	for _, node := range nodes {
		index[nodePath(node.ID)] = node
	}
	_ = rootParent
	return index
}

func importParentPath(nodes []projectmodel.ApiDocNode, parentID *uint) string {
	if parentID == nil {
		return "root"
	}
	byID := make(map[uint]projectmodel.ApiDocNode, len(nodes))
	for _, node := range nodes {
		byID[node.ID] = node
	}
	var nodePath func(uint) string
	nodePath = func(id uint) string {
		node, ok := byID[id]
		if !ok {
			return "root"
		}
		parentPath := "root"
		if node.ParentID != nil {
			parentPath = nodePath(*node.ParentID)
		}
		return parentPath + "/" + node.Name
	}
	return nodePath(*parentID)
}

func lineCount(content string) int {
	if content == "" {
		return 0
	}
	return len(strings.Split(content, "\n"))
}

func changedLineCount(from, to string) int {
	fromSet := make(map[string]int)
	for _, line := range strings.Split(from, "\n") {
		if line != "" || from != "" {
			fromSet[line]++
		}
	}
	changed := 0
	for _, line := range strings.Split(to, "\n") {
		if line == "" && to == "" {
			continue
		}
		if fromSet[line] > 0 {
			fromSet[line]--
		} else {
			changed++
		}
	}
	return changed
}
