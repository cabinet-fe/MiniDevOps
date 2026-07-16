package service

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"bedrock/internal/storage/model"
	"bedrock/internal/storage/repository"

	"gorm.io/gorm"
)

const (
	DefaultAttachmentMaxBytes int64 = 20 * 1024 * 1024
	DefaultDocImportMaxBytes  int64 = 100 * 1024 * 1024
	DefaultSkillZIPMaxBytes   int64 = 50 * 1024 * 1024
)

var (
	ErrTooLarge    = errors.New("文件超过允许大小")
	ErrUnavailable = errors.New("文件不可用")
)

type Limits struct {
	AttachmentMaxBytes int64
	DocImportMaxBytes  int64
	SkillZIPMaxBytes   int64
}

func (l Limits) normalized() Limits {
	if l.AttachmentMaxBytes <= 0 {
		l.AttachmentMaxBytes = DefaultAttachmentMaxBytes
	}
	if l.DocImportMaxBytes <= 0 {
		l.DocImportMaxBytes = DefaultDocImportMaxBytes
	}
	if l.SkillZIPMaxBytes <= 0 {
		l.SkillZIPMaxBytes = DefaultSkillZIPMaxBytes
	}
	return l
}

// StorageService owns every filesystem path used for persisted uploads.
type StorageService struct {
	repo   *repository.StorageRepository
	root   string
	limits Limits
}

func NewStorageService(repo *repository.StorageRepository, root string, limits Limits) (*StorageService, error) {
	root = strings.TrimSpace(root)
	if root == "" {
		return nil, errors.New("storage root 不能为空")
	}
	absoluteRoot, err := filepath.Abs(root)
	if err != nil {
		return nil, fmt.Errorf("resolve storage root: %w", err)
	}
	if err := os.MkdirAll(absoluteRoot, 0o755); err != nil {
		return nil, fmt.Errorf("create storage root: %w", err)
	}
	return &StorageService{repo: repo, root: absoluteRoot, limits: limits.normalized()}, nil
}

func (s *StorageService) MaxBytes(kind string) int64 {
	switch kind {
	case model.KindAttachment:
		return s.limits.AttachmentMaxBytes
	case model.KindDocImport:
		return s.limits.DocImportMaxBytes
	case model.KindSkillZIP:
		return s.limits.SkillZIPMaxBytes
	default:
		return s.limits.DocImportMaxBytes
	}
}

// Put persists one reference to the content. It hashes while copying, so a
// spoofed Content-Length cannot bypass the configured limit.
func (s *StorageService) Put(kind, contentType string, source io.Reader, declaredSize int64, createdBy uint) (*model.StorageObject, error) {
	maxBytes := s.MaxBytes(kind)
	if declaredSize > maxBytes {
		return nil, ErrTooLarge
	}
	if source == nil {
		return nil, errors.New("上传内容不能为空")
	}

	tmpDir := filepath.Join(s.root, "tmp")
	if err := os.MkdirAll(tmpDir, 0o755); err != nil {
		return nil, err
	}
	tmp, err := os.CreateTemp(tmpDir, "upload-*")
	if err != nil {
		return nil, err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)
	defer tmp.Close()

	hasher := sha256.New()
	written, err := io.Copy(io.MultiWriter(tmp, hasher), io.LimitReader(source, maxBytes+1))
	if err != nil {
		return nil, fmt.Errorf("保存上传文件: %w", err)
	}
	if written > maxBytes {
		return nil, ErrTooLarge
	}
	if err := tmp.Close(); err != nil {
		return nil, err
	}

	digest := fmt.Sprintf("%x", hasher.Sum(nil))
	if existing, err := s.repo.FindBySHA256(digest); err == nil {
		if err := s.repo.IncrementRef(existing.ID); err != nil {
			return nil, err
		}
		existing.RefCount++
		return existing, nil
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	relativePath := filepath.Join("objects", digest[:2], digest)
	finalPath, err := s.absolutePath(relativePath)
	if err != nil {
		return nil, err
	}
	if deleted, err := s.repo.FindBySHA256IncludingDeleted(digest); err == nil {
		if _, statErr := os.Stat(finalPath); statErr == nil {
			if err := s.repo.Restore(deleted.ID); err != nil {
				return nil, err
			}
			deleted.RefCount = 1
			deleted.PurgeAfter = nil
			deleted.DeletedAt = gorm.DeletedAt{}
			return deleted, nil
		}
		if err := s.repo.Purge(deleted.ID); err != nil {
			return nil, err
		}
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if err := os.MkdirAll(filepath.Dir(finalPath), 0o755); err != nil {
		return nil, err
	}
	if err := os.Rename(tmpName, finalPath); err != nil && !errors.Is(err, os.ErrExist) {
		return nil, fmt.Errorf("finalize storage object: %w", err)
	}

	object := &model.StorageObject{
		Kind:        kind,
		SHA256:      digest,
		Size:        written,
		ContentType: strings.TrimSpace(contentType),
		Path:        filepath.ToSlash(relativePath),
		RefCount:    1,
		CreatedBy:   createdBy,
	}
	if err := s.repo.Create(object); err != nil {
		// A concurrent upload may have won the SHA unique index. Reuse it.
		if existing, findErr := s.repo.FindBySHA256(digest); findErr == nil {
			if incErr := s.repo.IncrementRef(existing.ID); incErr != nil {
				return nil, incErr
			}
			existing.RefCount++
			return existing, nil
		}
		return nil, err
	}
	return object, nil
}

func (s *StorageService) Open(id uint) (*os.File, *model.StorageObject, error) {
	object, err := s.repo.FindByID(id)
	if err != nil {
		return nil, nil, err
	}
	if object.RefCount <= 0 {
		return nil, nil, ErrUnavailable
	}
	path, err := s.absolutePath(object.Path)
	if err != nil {
		return nil, nil, err
	}
	file, err := os.Open(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil, ErrUnavailable
		}
		return nil, nil, err
	}
	return file, object, nil
}

// Delete releases one reference. The object remains until its purge window so
// deduplicated content is never unlinked while another reference still exists.
func (s *StorageService) Delete(id uint) error {
	purgeAfter := time.Now().UTC().Add(24 * time.Hour)
	_, err := s.repo.DecrementRef(id, &purgeAfter)
	return err
}

func (s *StorageService) absolutePath(relative string) (string, error) {
	clean := filepath.Clean(filepath.FromSlash(relative))
	if clean == "." || filepath.IsAbs(clean) || clean == ".." || strings.HasPrefix(clean, ".."+string(os.PathSeparator)) {
		return "", errors.New("非法存储路径")
	}
	path := filepath.Join(s.root, clean)
	relativeToRoot, err := filepath.Rel(s.root, path)
	if err != nil || relativeToRoot == ".." || strings.HasPrefix(relativeToRoot, ".."+string(os.PathSeparator)) {
		return "", errors.New("非法存储路径")
	}
	return path, nil
}
