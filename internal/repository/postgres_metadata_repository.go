package repository

import (
	"context"
	"time"
	"youclips/internal/entities"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresMetadataRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresMetadataRepository(pool *pgxpool.Pool) *PostgresMetadataRepository {
	return &PostgresMetadataRepository{pool: pool}
}

func (r *PostgresMetadataRepository) GetByURL(ctx context.Context, url string) (*entities.VideoMetadata, error) {
	query := `
		SELECT id, url, title, duration, duration_string, channel, channel_url, 
		       uploader, uploader_id, upload_date, thumbnail, categories, ext, 
		       filesize_approx, formats, webpage_url, created_at
		FROM video_metadata
		WHERE url = $1`
	
	metadata := &entities.VideoMetadata{}
	err := r.pool.QueryRow(ctx, query, url).Scan(
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
	
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	
	return metadata, err
}

func (r *PostgresMetadataRepository) Create(ctx context.Context, metadata *entities.VideoMetadata) error {
	query := `
		INSERT INTO video_metadata (
			url, title, duration, duration_string, channel, channel_url, 
			uploader, uploader_id, upload_date, thumbnail, categories, ext, 
			filesize_approx, formats, webpage_url, created_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
		RETURNING id`
	
	err := r.pool.QueryRow(
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
	).Scan(&metadata.ID)
	
	return err
}
