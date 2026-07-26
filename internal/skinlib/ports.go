package skinlib

import (
	"context"
	"io"

	"github.com/lyimoexiao/akari/pkg/pagination"
)

// TextureItem is the public representation of a texture.
type TextureItem struct {
	TID      uint   `json:"tid"`
	Name     string `json:"name"`
	Type     string `json:"type"` // "steve", "alex", "cape"
	Hash     string `json:"hash"`
	URL      string `json:"url"`
	Size     int64  `json:"size"`
	Uploader uint   `json:"uploader"`
	Public   bool   `json:"public"`
	Likes    int64  `json:"likes"`
	UploadAt string `json:"upload_at"`
}

// TextureDetail is a texture with owner info.
type TextureDetail struct {
	TextureItem
	UploaderName string `json:"uploader_name,omitempty"`
}

// ListResult is the paginated result for skin library queries.
type ListResult = pagination.Paged[TextureItem]

// ListQuery defines filtering for skin library listing.
type ListQuery struct {
	Pagination pagination.Paging
	Type       string // "skin" (steve+alex) or "cape" or empty (all)
	Search     string // keyword search in name
	Uploader   uint   // filter by uploader (0 = no filter)
	PublicOnly bool   // only show public textures
	Order      string // "latest" (default), "likes"
}

// Repository defines data access for textures.
type Repository interface {
	// Create inserts a new texture record.
	Create(ctx context.Context, texture *TextureRecord) error

	// FindByID returns a texture by TID.
	FindByID(ctx context.Context, tid uint) (*TextureRecord, error)

	// FindByHash returns a texture by hash.
	FindByHash(ctx context.Context, hash string) (*TextureRecord, error)

	// FindByHashAndUploader returns a non-deleted texture by hash and uploader.
	FindByHashAndUploader(ctx context.Context, hash string, uploader uint) (*TextureRecord, error)

	// List returns paginated texture list.
	List(ctx context.Context, query ListQuery) ([]TextureRecord, int64, error)

	// Update updates texture fields (name, public).
	Update(ctx context.Context, tid uint, name string, public bool) error

	// Delete soft-deletes a texture.
	Delete(ctx context.Context, tid uint) error

	// IncrementLikes atomically adds 1 to likes count.
	IncrementLikes(ctx context.Context, tid uint) error

	// DecrementLikes atomically subtracts 1 from likes count.
	DecrementLikes(ctx context.Context, tid uint) error

	// CountByUploader returns the number of textures uploaded by a user.
	CountByUploader(ctx context.Context, uploader uint) (int64, error)
}

// TextureRecord is the internal representation with all DB fields.
type TextureRecord struct {
	TID      uint
	Name     string
	Type     string
	Hash     string
	URL      string
	Size     int64
	Uploader uint
	Public   bool
	Likes    int64
	UploadAt string
}

// FileStorage defines operations for texture file storage.
type FileStorage interface {
	// Save stores a texture file and returns its SHA256 hash and size.
	Save(ctx context.Context, reader io.Reader) (hash string, size int64, err error)

	// Open opens a texture file for reading by hash.
	Open(ctx context.Context, hash string) (io.ReadCloser, error)

	// Delete removes a texture file by hash.
	Delete(ctx context.Context, hash string) error

	// Exists checks if a texture file exists.
	Exists(ctx context.Context, hash string) (bool, error)
}

// ScoreOperator defines score operations needed by skinlib.
type ScoreOperator interface {
	Award(ctx context.Context, userID uint, amount int64, reason string) error
	HasEnough(ctx context.Context, userID uint, amount int64) (bool, error)
	Deduct(ctx context.Context, userID uint, amount int64, reason string) (int64, error)
}

// ClosetAdder defines the interface for adding/removing textures in the user's closet,
// used after upload so the uploader immediately sees it in their closet.
type ClosetAdder interface {
	Add(ctx context.Context, userID, textureTID uint, itemName string) error
	Remove(ctx context.Context, userID, textureTID uint) error
}

// ClosetCleaner defines the interface for removing closet entries
// when a texture is set to private — removes all entries except the uploader's.
type ClosetCleaner interface {
	RemoveByTextureExceptUploader(ctx context.Context, textureTID, uploaderID uint) error
}

// Dependencies groups dependencies for Service construction.
type Dependencies struct {
	Repository    Repository
	Storage       FileStorage
	ScoreOps      ScoreOperator
	ClosetAdder   ClosetAdder
	ClosetCleaner ClosetCleaner
	BaseURL       string
	AwardUpload   int64
}
