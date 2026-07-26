package closet

import (
	"context"

	"github.com/lyimoexiao/akari/pkg/pagination"
)

// ClosetItem is the public representation of a closet entry with texture info.
type ClosetItem struct {
	TextureTID uint   `json:"texture_tid"`
	ItemName   string `json:"item_name"`
	CreatedAt  string `json:"created_at"`
	Texture    *struct {
		Name     string `json:"name"`
		Type     string `json:"type"`
		Hash     string `json:"hash"`
		URL      string `json:"url"`
		Size     int64  `json:"size"`
		Public   bool   `json:"public"`
		Likes    int64  `json:"likes"`
		Uploader uint   `json:"uploader"`
	} `json:"texture,omitempty"`
}

// ListResult is the paginated closet list.
type ListResult = pagination.Paged[ClosetItem]

// Repository defines data access for the user closet.
type Repository interface {
	// Add adds a texture to the user's closet.
	Add(ctx context.Context, userID, textureTID uint, itemName string) error

	// Remove removes a texture from the user's closet.
	Remove(ctx context.Context, userID, textureTID uint) error

	// Rename renames a closet item.
	Rename(ctx context.Context, userID, textureTID uint, newName string) error

	// List returns the user's closet items with texture info.
	List(ctx context.Context, userID uint, query ListQuery) ([]ClosetItem, int64, error)

	// Exists checks if a texture is in the user's closet.
	Exists(ctx context.Context, userID, textureTID uint) (bool, error)

	// AllTextureIDs returns all texture TIDs in the user's closet.
	AllTextureIDs(ctx context.Context, userID uint) ([]uint, error)
}

// ListQuery for closet listing.
type ListQuery struct {
	Pagination pagination.Paging
	Type       string // "skin", "cape", or empty (all)
	Search     string
}

// TextureRepository defines the texture lookups needed by closet.
type TextureRepository interface {
	Exists(ctx context.Context, tid uint) (bool, error)
	IsPublicOrUploader(ctx context.Context, tid, userID uint) (bool, error)
	GetUploader(ctx context.Context, tid uint) (uint, error)
}

// TextureDeleter defines the ability to soft-delete a texture by ID.
type TextureDeleter interface {
	DeleteByID(ctx context.Context, tid, callerID uint, force bool) error
}

// ProfileTextureCleaner defines the ability to clear a texture from
// a user's Yggdrasil profile (skin/cape) when it's removed from the closet.
type ProfileTextureCleaner interface {
	ClearTextureFromProfile(ctx context.Context, userID, textureTID uint) error
}

// ScoreOperator defines the score operations needed by closet.
type ScoreOperator interface {
	HasEnough(ctx context.Context, userID uint, amount int64) (bool, error)
	Deduct(ctx context.Context, userID uint, amount int64, reason string) (int64, error)
	Award(ctx context.Context, userID uint, amount int64, reason string) error
}

type Dependencies struct {
	Repository          Repository
	TextureRepo         TextureRepository
	TextureDeleter      TextureDeleter
	ProfileCleaner      ProfileTextureCleaner
	ScoreOps            ScoreOperator
	CostPerItem         int64
	ReturnScoreOnRemove bool
	AwardPerLike        int64
}
