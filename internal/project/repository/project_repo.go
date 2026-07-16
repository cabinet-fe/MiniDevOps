package repository

import (
	"strings"

	"bedrock/internal/project/model"

	"gorm.io/gorm"
)

type ProjectRepository struct {
	db *gorm.DB
}

func NewProjectRepository(db *gorm.DB) *ProjectRepository {
	return &ProjectRepository{db: db}
}

func (r *ProjectRepository) CreateProject(project *model.ProductProject) error {
	return r.db.Create(project).Error
}

func (r *ProjectRepository) CreateProjectWithOwner(project *model.ProductProject) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(project).Error; err != nil {
			return err
		}
		return tx.Create(&model.ProjectMember{
			ProjectID: project.ID,
			UserID:    project.OwnerID,
			Role:      model.ProjectRoleOwner,
		}).Error
	})
}

func (r *ProjectRepository) FindProject(id uint) (*model.ProductProject, error) {
	var project model.ProductProject
	if err := r.db.First(&project, id).Error; err != nil {
		return nil, err
	}
	return &project, nil
}

func (r *ProjectRepository) FindProjectBySlug(slug string) (*model.ProductProject, error) {
	var project model.ProductProject
	if err := r.db.Where("slug = ?", slug).First(&project).Error; err != nil {
		return nil, err
	}
	return &project, nil
}

func (r *ProjectRepository) ListProjects(page, pageSize uint, keyword, status string, userID uint, all bool) ([]model.ProductProject, int64, error) {
	q := r.db.Model(&model.ProductProject{})
	if !all {
		q = q.Joins("JOIN project_members ON project_members.project_id = product_projects.id").
			Where("project_members.user_id = ?", userID)
	}
	if keyword = strings.TrimSpace(keyword); keyword != "" {
		like := "%" + keyword + "%"
		q = q.Where("product_projects.name LIKE ? OR product_projects.slug LIKE ? OR product_projects.tags LIKE ?", like, like, like)
	}
	if status = strings.TrimSpace(status); status != "" {
		q = q.Where("product_projects.status = ?", status)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var projects []model.ProductProject
	err := q.Order("product_projects.updated_at DESC, product_projects.id DESC").
		Offset(int((page - 1) * pageSize)).Limit(int(pageSize)).Find(&projects).Error
	return projects, total, err
}

func (r *ProjectRepository) UpdateProject(project *model.ProductProject) error {
	return r.db.Save(project).Error
}

func (r *ProjectRepository) DeleteProject(id uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("project_id = ?", id).Delete(&model.ProjectMember{}).Error; err != nil {
			return err
		}
		requirements := tx.Model(&model.Requirement{}).Select("id").Where("project_id = ?", id)
		if err := tx.Where("requirement_id IN (?)", requirements).Delete(&model.RequirementComment{}).Error; err != nil {
			return err
		}
		if err := tx.Where("requirement_id IN (?)", requirements).Delete(&model.RequirementAttachment{}).Error; err != nil {
			return err
		}
		if err := tx.Where("project_id = ?", id).Delete(&model.Requirement{}).Error; err != nil {
			return err
		}
		if err := tx.Where("project_id = ?", id).Delete(&model.ApiDocNode{}).Error; err != nil {
			return err
		}
		return tx.Delete(&model.ProductProject{}, id).Error
	})
}

func (r *ProjectRepository) CreateMember(member *model.ProjectMember) error {
	return r.db.Create(member).Error
}

func (r *ProjectRepository) FindMember(projectID, userID uint) (*model.ProjectMember, error) {
	var member model.ProjectMember
	if err := r.db.Where("project_id = ? AND user_id = ?", projectID, userID).First(&member).Error; err != nil {
		return nil, err
	}
	return &member, nil
}

func (r *ProjectRepository) ListMemberRoles(projectIDs []uint, userID uint) (map[uint]string, error) {
	if len(projectIDs) == 0 {
		return map[uint]string{}, nil
	}
	var members []model.ProjectMember
	if err := r.db.Select("project_id", "role").
		Where("user_id = ? AND project_id IN ?", userID, projectIDs).
		Find(&members).Error; err != nil {
		return nil, err
	}
	roles := make(map[uint]string, len(members))
	for _, member := range members {
		roles[member.ProjectID] = member.Role
	}
	return roles, nil
}

func (r *ProjectRepository) HasProjectMembership(userID uint) (bool, error) {
	var count int64
	if err := r.db.Model(&model.ProjectMember{}).Where("user_id = ?", userID).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *ProjectRepository) ListMembers(projectID uint) ([]model.ProjectMember, error) {
	var members []model.ProjectMember
	err := r.db.Where("project_id = ?", projectID).
		Order("CASE role WHEN 'owner' THEN 0 WHEN 'admin' THEN 1 WHEN 'member' THEN 2 ELSE 3 END, id ASC").
		Find(&members).Error
	return members, err
}

func (r *ProjectRepository) UpdateMember(member *model.ProjectMember) error {
	return r.db.Save(member).Error
}

func (r *ProjectRepository) DeleteMember(projectID, userID uint) error {
	return r.db.Where("project_id = ? AND user_id = ?", projectID, userID).Delete(&model.ProjectMember{}).Error
}

func (r *ProjectRepository) TransferOwner(projectID, previousOwnerID, nextOwnerID uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.ProjectMember{}).
			Where("project_id = ? AND user_id = ?", projectID, previousOwnerID).
			Update("role", model.ProjectRoleAdmin).Error; err != nil {
			return err
		}
		if err := tx.Model(&model.ProjectMember{}).
			Where("project_id = ? AND user_id = ?", projectID, nextOwnerID).
			Update("role", model.ProjectRoleOwner).Error; err != nil {
			return err
		}
		return tx.Model(&model.ProductProject{}).Where("id = ?", projectID).Update("owner_id", nextOwnerID).Error
	})
}

func (r *ProjectRepository) CreateRequirement(requirement *model.Requirement) error {
	return r.db.Create(requirement).Error
}

func (r *ProjectRepository) FindRequirement(id uint) (*model.Requirement, error) {
	var requirement model.Requirement
	if err := r.db.First(&requirement, id).Error; err != nil {
		return nil, err
	}
	return &requirement, nil
}

func (r *ProjectRepository) ListRequirements(projectID, page, pageSize uint, keyword, status, priority, assignee string, sort string) ([]model.Requirement, int64, error) {
	q := r.db.Model(&model.Requirement{}).Where("project_id = ?", projectID)
	if keyword = strings.TrimSpace(keyword); keyword != "" {
		like := "%" + keyword + "%"
		q = q.Where("title LIKE ? OR tags LIKE ?", like, like)
	}
	if status = strings.TrimSpace(status); status != "" {
		q = q.Where("status = ?", status)
	}
	if priority = strings.TrimSpace(priority); priority != "" {
		q = q.Where("priority = ?", priority)
	}
	if assignee = strings.TrimSpace(assignee); assignee != "" {
		q = q.Where("assignee_id = ?", assignee)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	order := "updated_at DESC, id DESC"
	switch sort {
	case "created_at@asc":
		order = "created_at ASC, id ASC"
	case "created_at@desc":
		order = "created_at DESC, id DESC"
	case "title@asc":
		order = "title ASC, id ASC"
	case "title@desc":
		order = "title DESC, id DESC"
	case "priority@asc":
		order = "priority ASC, id ASC"
	case "priority@desc":
		order = "priority DESC, id DESC"
	}
	var requirements []model.Requirement
	err := q.Order(order).Offset(int((page - 1) * pageSize)).Limit(int(pageSize)).Find(&requirements).Error
	return requirements, total, err
}

func (r *ProjectRepository) UpdateRequirement(requirement *model.Requirement) error {
	return r.db.Save(requirement).Error
}

func (r *ProjectRepository) DeleteRequirement(id uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("requirement_id = ?", id).Delete(&model.RequirementComment{}).Error; err != nil {
			return err
		}
		if err := tx.Where("requirement_id = ?", id).Delete(&model.RequirementAttachment{}).Error; err != nil {
			return err
		}
		return tx.Delete(&model.Requirement{}, id).Error
	})
}

func (r *ProjectRepository) CreateComment(comment *model.RequirementComment) error {
	return r.db.Create(comment).Error
}

func (r *ProjectRepository) FindComment(id uint) (*model.RequirementComment, error) {
	var comment model.RequirementComment
	if err := r.db.First(&comment, id).Error; err != nil {
		return nil, err
	}
	return &comment, nil
}

func (r *ProjectRepository) ListComments(requirementID uint) ([]model.RequirementComment, error) {
	var comments []model.RequirementComment
	err := r.db.Where("requirement_id = ?", requirementID).Order("created_at ASC, id ASC").Find(&comments).Error
	return comments, err
}

func (r *ProjectRepository) UpdateComment(comment *model.RequirementComment) error {
	return r.db.Save(comment).Error
}

func (r *ProjectRepository) DeleteComment(id uint) error {
	return r.db.Delete(&model.RequirementComment{}, id).Error
}

func (r *ProjectRepository) CreateAttachment(attachment *model.RequirementAttachment) error {
	return r.db.Create(attachment).Error
}

func (r *ProjectRepository) FindAttachment(id uint) (*model.RequirementAttachment, error) {
	var attachment model.RequirementAttachment
	if err := r.db.First(&attachment, id).Error; err != nil {
		return nil, err
	}
	return &attachment, nil
}

func (r *ProjectRepository) ListAttachments(requirementID uint) ([]model.RequirementAttachment, error) {
	var attachments []model.RequirementAttachment
	err := r.db.Where("requirement_id = ?", requirementID).Order("id ASC").Find(&attachments).Error
	return attachments, err
}

func (r *ProjectRepository) ListAttachmentsByProject(projectID uint) ([]model.RequirementAttachment, error) {
	var attachments []model.RequirementAttachment
	err := r.db.Model(&model.RequirementAttachment{}).
		Joins("JOIN requirements ON requirements.id = requirement_attachments.requirement_id").
		Where("requirements.project_id = ?", projectID).
		Order("requirement_attachments.id ASC").
		Find(&attachments).Error
	return attachments, err
}

func (r *ProjectRepository) DeleteAttachment(id uint) error {
	return r.db.Delete(&model.RequirementAttachment{}, id).Error
}

func (r *ProjectRepository) CreateDocNode(node *model.ApiDocNode) error {
	return r.db.Create(node).Error
}

func (r *ProjectRepository) FindDocNode(id uint) (*model.ApiDocNode, error) {
	var node model.ApiDocNode
	if err := r.db.First(&node, id).Error; err != nil {
		return nil, err
	}
	return &node, nil
}

func (r *ProjectRepository) ListDocNodes(projectID uint) ([]model.ApiDocNode, error) {
	var nodes []model.ApiDocNode
	err := r.db.Where("project_id = ?", projectID).Order("sort_order ASC, id ASC").Find(&nodes).Error
	return nodes, err
}

func (r *ProjectRepository) UpdateDocNode(node *model.ApiDocNode) error {
	return r.db.Save(node).Error
}

func (r *ProjectRepository) DeleteDocNodes(ids []uint) error {
	if len(ids) == 0 {
		return nil
	}
	return r.db.Where("id IN ?", ids).Delete(&model.ApiDocNode{}).Error
}

func (r *ProjectRepository) PublishDocNode(id uint, expectedVersion int) (bool, error) {
	result := r.db.Model(&model.ApiDocNode{}).
		Where("id = ? AND content_version = ? AND kind = ?", id, expectedVersion, model.DocNodeDocument).
		Updates(map[string]interface{}{
			"published_content":   gorm.Expr("draft_content"),
			"draft_content":       "",
			"content_version":     gorm.Expr("content_version + ?", 1),
			"draft_base_version":  0,
			"draft_updated_at":    nil,
			"draft_source_run_id": nil,
		})
	return result.RowsAffected == 1, result.Error
}

func (r *ProjectRepository) RequirementStatusExists(value string) (bool, error) {
	var count int64
	err := r.db.Table("dict_items").
		Joins("JOIN dictionaries ON dictionaries.id = dict_items.dictionary_id").
		Where("dictionaries.code = ? AND dict_items.value = ? AND dict_items.enabled = ?", "requirement_status", value, true).
		Count(&count).Error
	return count > 0, err
}

func (r *ProjectRepository) ListRequirementStatuses() ([]model.RequirementStatusOption, error) {
	var statuses []model.RequirementStatusOption
	err := r.db.Table("dict_items").
		Select("dict_items.label, dict_items.value, dict_items.sort_order, dict_items.enabled").
		Joins("JOIN dictionaries ON dictionaries.id = dict_items.dictionary_id").
		Where("dictionaries.code = ? AND dict_items.enabled = ?", "requirement_status", true).
		Order("dict_items.sort_order ASC, dict_items.id ASC").
		Scan(&statuses).Error
	return statuses, err
}
