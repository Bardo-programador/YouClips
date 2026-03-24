package repository

import (
	"context"
	"database/sql"
	"time"
	"youclips/internal/entities"
)

type MetadataRepository interface {
	GetByURL(ctx context.Context, url string) (*entities.VideoMetadata, error)
	Create(ctx context.Context, metadata *entities.VideoMetadata) error
}

type SQLiteMetadataRepository struct {
	db *sql.DB
}

func NewSQLiteMetadataRepository(db *sql.DB) *SQLiteMetadataRepository {
	return &SQLiteMetadataRepository{db: db}
}

func (r *SQLiteMetadataRepository) GetByURL(ctx context.Context, url string) (*entities.VideoMetadata, error) {
	query := `
		SELECT id, url, title, duration, duration_string, channel, channel_url, 
		       uploader, uploader_id, upload_date, thumbnail, categories, ext, 
		       filesize_approx, formats, webpage_url, created_at
		FROM video_metadata
		WHERE url = ?`
	
	metadata := &entities.VideoMetadata{}
	err := r.db.QueryRowContext(ctx, query, url).Scan(
		&metadata.ID,
		&metadata.URL,
		&metadata.Title,
		&metadata.Duration,
		&metadata.DurationString,
		&metadata.Channel,
		&metadata.ChannelURL,
		&metadata.Uploader,
		&metadata.UploaderID,
		&metadata.UploadDate,
		&metadata.Thumbnail,
		&metadata.Categories,
		&metadata.Ext,
		&metadata.FilesizeApprox,
		&metadata.Formats,
		&metadata.WebpageURL,
		&metadata.CreatedAt,
	)
	
	if err == sql.ErrNoRows {
		return nil, nil
	}
	
	return metadata, err
}

func (r *SQLiteMetadataRepository) Create(ctx context.Context, metadata *entities.VideoMetadata) error {
	query := `
		INSERT INTO video_metadata (
			url, title, duration, duration_string, channel, channel_url, 
			uploader, uploader_id, upload_date, thumbnail, categories, ext, 
			filesize_approx, formats, webpage_url, created_at
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	
	result, err := r.db.ExecContext(
		ctx, query,
		metadata.URL,
		metadata.Title,
		metadata.Duration,
		metadata.DurationString,
		metadata.Channel,
		metadata.ChannelURL,
		metadata.Uploader,
		metadata.UploaderID,
		metadata.UploadDate,
		metadata.Thumbnail,
		metadata.Categories,
		metadata.Ext,
		metadata.FilesizeApprox,
		metadata.Formats,
		metadata.WebpageURL,
		time.Now(),
	)
	if err != nil {
		return err
	}
	
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	metadata.ID = int(id)
	return nil
}
