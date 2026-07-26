package skinlib

import (
	"bytes"
	"context"
	"fmt"

	"github.com/lyimoexiao/akari/pkg/pagination"
)

// Service implements texture/skin library business logic.
type Service struct {
	repo          Repository
	storage       FileStorage
	scoreOps      ScoreOperator
	closetAdder   ClosetAdder
	closetCleaner ClosetCleaner
	baseURL       string
	awardUpload   int64
}

func NewService(deps Dependencies) *Service {
	return &Service{
		repo:          deps.Repository,
		storage:       deps.Storage,
		scoreOps:      deps.ScoreOps,
		closetAdder:   deps.ClosetAdder,
		closetCleaner: deps.ClosetCleaner,
		baseURL:       deps.BaseURL,
		awardUpload:   deps.AwardUpload,
	}
}

// Upload saves a texture file and creates a DB record.
func (s *Service) Upload(ctx context.Context, uploader uint, name, textureType string, reader interface{ Read(p []byte) (int, error) }) (*TextureItem, error) {
	if textureType != "steve" && textureType != "alex" && textureType != "cape" {
		return nil, ErrInvalidType
	}

	// Validate and sanitize the PNG before saving
	sanitized, err := validateAndSanitizePNG(reader)
	if err != nil {
		return nil, err
	}

	// Save sanitized data to storage (storage handles file-level dedup by hash)
	hash, size, err := s.storage.Save(ctx, bytes.NewReader(sanitized))
	if err != nil {
		return nil, fmt.Errorf("save texture file: %w", err)
	}

	// Same user uploading identical content — notify user
	existing, err := s.repo.FindByHashAndUploader(ctx, hash, uploader)
	if err == nil && existing != nil {
		return nil, ErrAlreadyUploaded
	}

	record := &TextureRecord{
		Name:     name,
		Type:     textureType,
		Hash:     hash,
		URL:      fmt.Sprintf("%s/api/v1/raw/%s", s.baseURL, hash),
		Size:     size,
		Uploader: uploader,
		Public:   true,
		Likes:    0,
	}

	if err := s.repo.Create(ctx, record); err != nil {
		_ = s.storage.Delete(ctx, hash)
		return nil, fmt.Errorf("create texture record: %w", err)
	}

	// Award score
	if s.awardUpload > 0 && s.scoreOps != nil {
		_ = s.scoreOps.Award(ctx, uploader, s.awardUpload, "texture_upload")
	}

	// Transfer closet binding: if the user already has a different texture with the
	// same hash in their closet (uploaded by someone else), remove the old entry so
	// the auto-add below replaces it with the user's own newly uploaded texture.
	if s.closetAdder != nil {
		existingByHash, _ := s.repo.FindByHash(ctx, hash)
		if existingByHash != nil && existingByHash.TID != record.TID {
			// Found another texture with the same hash — remove it from
			// the uploader's closet so only the new entry remains.
			_ = s.closetAdder.Remove(ctx, uploader, existingByHash.TID)
		}
		_ = s.closetAdder.Add(ctx, uploader, record.TID, name)
	}

	return s.toItem(record), nil
}

// List returns a paginated texture list.
func (s *Service) List(ctx context.Context, query ListQuery) (*ListResult, error) {
	query.Pagination.Normalise()

	records, total, err := s.repo.List(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list textures: %w", err)
	}

	items := make([]TextureItem, len(records))
	for i, rec := range records {
		items[i] = *s.toItem(&rec)
	}

	return pagination.NewPaged(items, total, query.Pagination), nil
}

// GetByID returns a single texture by TID. Returns error if not found.
func (s *Service) GetByID(ctx context.Context, tid uint) (*TextureRecord, error) {
	record, err := s.repo.FindByID(ctx, tid)
	if err != nil {
		return nil, err
	}
	return record, nil
}

// Update updates a texture's name and public status.
// callerID must match the uploader, or set force=true for admin override.
// When setting public=false, removes this texture from all other users' closets.
func (s *Service) Update(ctx context.Context, tid, callerID uint, name string, public bool, force bool) error {
	record, err := s.repo.FindByID(ctx, tid)
	if err != nil {
		return err
	}
	if record.Uploader != callerID && !force {
		return ErrForbidden
	}
	if err := s.repo.Update(ctx, tid, name, public); err != nil {
		return fmt.Errorf("update texture: %w", err)
	}

	// When setting to private, remove from all other users' closets
	if !public && s.closetCleaner != nil {
		_ = s.closetCleaner.RemoveByTextureExceptUploader(ctx, tid, callerID)
	}

	return nil
}

// Delete soft-deletes a texture.
// callerID must match uploader, or force=true for admin.
func (s *Service) Delete(ctx context.Context, tid, callerID uint, force bool) error {
	record, err := s.repo.FindByID(ctx, tid)
	if err != nil {
		return err
	}
	if record.Uploader != callerID && !force {
		return ErrForbidden
	}
	if err := s.repo.Delete(ctx, tid); err != nil {
		return fmt.Errorf("delete texture: %w", err)
	}
	return nil
}

func (s *Service) toItem(r *TextureRecord) *TextureItem {
	return &TextureItem{
		TID:      r.TID,
		Name:     r.Name,
		Type:     r.Type,
		Hash:     r.Hash,
		URL:      r.URL,
		Size:     r.Size,
		Uploader: r.Uploader,
		Public:   r.Public,
		Likes:    r.Likes,
		UploadAt: r.UploadAt,
	}
}
