package closet

import (
	"context"
	"fmt"

	"github.com/lyimoexiao/akari/pkg/pagination"
)

type Service struct {
	repo                Repository
	textureRepo         TextureRepository
	textureDeleter      TextureDeleter
	profileCleaner      ProfileTextureCleaner
	scoreOps            ScoreOperator
	costPerItem         int64
	returnScoreOnRemove bool
	awardPerLike        int64
}

func NewService(deps Dependencies) *Service {
	return &Service{
		repo:                deps.Repository,
		textureRepo:         deps.TextureRepo,
		textureDeleter:      deps.TextureDeleter,
		profileCleaner:      deps.ProfileCleaner,
		scoreOps:            deps.ScoreOps,
		costPerItem:         deps.CostPerItem,
		returnScoreOnRemove: deps.ReturnScoreOnRemove,
		awardPerLike:        deps.AwardPerLike,
	}
}

// Add adds a texture to the user's closet.
// Only public textures or the user's own uploads can be added.
// Deducts score if cost_per_item > 0.
func (s *Service) Add(ctx context.Context, userID, textureTID uint, itemName string) error {
	// Check texture is accessible (public or own upload)
	accessible, err := s.textureRepo.IsPublicOrUploader(ctx, textureTID, userID)
	if err != nil {
		return fmt.Errorf("check texture access: %w", err)
	}
	if !accessible {
		return ErrClosetItemNotFound
	}

	// Check already in closet
	inCloset, err := s.repo.Exists(ctx, userID, textureTID)
	if err != nil {
		return fmt.Errorf("check closet: %w", err)
	}
	if inCloset {
		return ErrAlreadyInCloset
	}

	// Deduct score if needed
	if s.costPerItem > 0 && s.scoreOps != nil {
		enough, err := s.scoreOps.HasEnough(ctx, userID, s.costPerItem)
		if err != nil {
			return fmt.Errorf("check score: %w", err)
		}
		if !enough {
			return ErrNotEnoughScore
		}
		if _, err := s.scoreOps.Deduct(ctx, userID, s.costPerItem, "closet_add"); err != nil {
			return fmt.Errorf("deduct score: %w", err)
		}
	}

	if err := s.repo.Add(ctx, userID, textureTID, itemName); err != nil {
		return fmt.Errorf("add to closet: %w", err)
	}

	return nil
}

// Remove removes a texture from the user's closet.
// If the user is the uploader of the texture, it is also soft-deleted from the skin library.
// If the texture was added from someone else's upload, only the closet entry is removed.
// Returns score if return_score_on_remove is true.
func (s *Service) Remove(ctx context.Context, userID, textureTID uint) error {
	if err := s.repo.Remove(ctx, userID, textureTID); err != nil {
		return err
	}

	// Only soft-delete from skinlib if the user is the uploader
	if s.textureDeleter != nil {
		uploader, err := s.textureRepo.GetUploader(ctx, textureTID)
		if err == nil && uploader == userID {
			_ = s.textureDeleter.DeleteByID(ctx, textureTID, userID, false)
		}
	}

	// Clear profile skin/cape if this texture was set as active skin/cape
	if s.profileCleaner != nil {
		_ = s.profileCleaner.ClearTextureFromProfile(ctx, userID, textureTID)
	}

	if s.returnScoreOnRemove && s.costPerItem > 0 && s.scoreOps != nil {
		_ = s.scoreOps.Award(ctx, userID, s.costPerItem, "closet_remove_refund")
	}

	return nil
}

// Rename renames a closet item.
// Only the texture uploader is allowed to rename.
func (s *Service) Rename(ctx context.Context, userID, textureTID uint, newName string) error {
	uploader, err := s.textureRepo.GetUploader(ctx, textureTID)
	if err != nil {
		return fmt.Errorf("get texture uploader: %w", err)
	}
	if uploader != userID {
		return ErrNotTextureUploader
	}

	if err := s.repo.Rename(ctx, userID, textureTID, newName); err != nil {
		return err
	}
	return nil
}

// List returns the user's closet items.
func (s *Service) List(ctx context.Context, userID uint, query ListQuery) (*ListResult, error) {
	query.Pagination.Normalise()

	items, total, err := s.repo.List(ctx, userID, query)
	if err != nil {
		return nil, fmt.Errorf("list closet: %w", err)
	}

	return pagination.NewPaged(items, total, query.Pagination), nil
}

// AllTextureIDs returns all texture TIDs in the user's closet.
func (s *Service) AllTextureIDs(ctx context.Context, userID uint) ([]uint, error) {
	return s.repo.AllTextureIDs(ctx, userID)
}
