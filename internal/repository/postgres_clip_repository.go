package repository

import (
	"context"
	"fmt"
	"log"
	"youclips/internal/entities"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresClipRepository struct {
	db *pgxpool.Pool
}

func NewPostgresClipRepository(db *pgxpool.Pool) *PostgresClipRepository {
	return &PostgresClipRepository{db: db}
}

func (r *PostgresClipRepository) Create(ctx context.Context, clip *entities.Clip) error {
	query := `
		INSERT INTO clips (created_at, title, start_time, end_time, duration_seconds, format, quality, size, expected_size, file_path, original_url, status, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING id`
	
	err := r.db.QueryRow(
		ctx, query,
		clip.CreatedAt,
		clip.Title,
		clip.StartTime,
		clip.EndTime,
		clip.DurationSeconds,
		clip.Format,
		clip.Quality,
		clip.Size,
		clip.ExpectedSize,
		clip.FilePath,
		clip.OriginalURL,
		clip.Status,
		clip.ExpiresAt,
	).Scan(&clip.ID)
	
	return err
}

func (r *PostgresClipRepository) GetByID(ctx context.Context, id int) (*entities.Clip, error) {
	query := `
		SELECT id, created_at, title, start_time, end_time, duration_seconds, format, quality, size, expected_size, file_path, original_url, status, expires_at
		FROM clips
		WHERE id = $1`
	
	clip := &entities.Clip{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&clip.ID,
		&clip.CreatedAt,
		&clip.Title,
		&clip.StartTime,
		&clip.EndTime,
		&clip.DurationSeconds,
		&clip.Format,
		&clip.Quality,
		&clip.Size,
		&clip.ExpectedSize,
		&clip.FilePath,
		&clip.OriginalURL,
		&clip.Status,
		&clip.ExpiresAt,
	)
	
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	
	return clip, err
}

func (r *PostgresClipRepository) List(ctx context.Context, limit, offset int) ([]*entities.Clip, error) {
	query := `
		SELECT id, created_at, title, start_time, end_time, duration_seconds, format, quality, size, expected_size, file_path, original_url, status, expires_at
		FROM clips
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2`
	
	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var clips []*entities.Clip
	for rows.Next() {
		clip := &entities.Clip{}
		err := rows.Scan(
			&clip.ID,
			&clip.CreatedAt,
			&clip.Title,
			&clip.StartTime,
			&clip.EndTime,
			&clip.DurationSeconds,
			&clip.Format,
			&clip.Quality,
			&clip.Size,
			&clip.ExpectedSize,
			&clip.FilePath,
			&clip.OriginalURL,
			&clip.Status,
			&clip.ExpiresAt,
		)
		if err != nil {
			return nil, err
		}
		clips = append(clips, clip)
	}
	
	return clips, rows.Err()
}

func (r *PostgresClipRepository) Count(ctx context.Context) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM clips`
	err := r.db.QueryRow(ctx, query).Scan(&count)
	return count, err
}

func (r *PostgresClipRepository) UpdateStatus(ctx context.Context, id int, status entities.ClipStatus) error {
	query := `UPDATE clips SET status = $1 WHERE id = $2`
	_, err := r.db.Exec(ctx, query, status, id)
	return err
}

func (r *PostgresClipRepository) Update(ctx context.Context, clip *entities.Clip) error {
	query := `
		UPDATE clips 
		SET title = $1, size = $2, expected_size = $3, file_path = $4, status = $5, expires_at = $6
		WHERE id = $7`
	
	_, err := r.db.Exec(
		ctx, query,
		clip.Title,
		clip.Size,
		clip.ExpectedSize,
		clip.FilePath,
		clip.Status,
		clip.ExpiresAt,
		clip.ID,
	)
	return err
}

func (r *PostgresClipRepository) Delete(ctx context.Context, id int) (bool, error) {
	query := `DELETE FROM clips WHERE id = $1`
	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return false, err
	}
	return result.RowsAffected() > 0, nil
}

func (r *PostgresClipRepository) FindExpiredClips(ctx context.Context) ([]*entities.Clip, error) {
	query := `
		SELECT id, created_at, title, start_time, end_time, duration_seconds, format, size, file_path, original_url, status, expires_at
		FROM clips
		WHERE status = $1 AND expires_at IS NOT NULL AND expires_at <= NOW()`
	
	rows, err := r.db.Query(ctx, query, entities.StatusCompleted)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var clips []*entities.Clip
	for rows.Next() {
		clip := &entities.Clip{}
		err := rows.Scan(
			&clip.ID,
			&clip.CreatedAt,
			&clip.Title,
			&clip.StartTime,
			&clip.EndTime,
			&clip.DurationSeconds,
			&clip.Format,
			&clip.Size,
			&clip.FilePath,
			&clip.OriginalURL,
			&clip.Status,
			&clip.ExpiresAt,
		)
		if err != nil {
			return nil, err
		}
		clips = append(clips, clip)
	}
	
	return clips, rows.Err()
}

func InitPostgresDatabase(ctx context.Context, pool *pgxpool.Pool) error {
	// Create clips table if it doesn't exist
	clipsSchema := `
	CREATE TABLE IF NOT EXISTS clips (
		id SERIAL PRIMARY KEY,
		created_at TIMESTAMP NOT NULL,
		title TEXT NOT NULL DEFAULT '',
		start_time INTEGER NOT NULL,
		end_time INTEGER NOT NULL,
		duration_seconds INTEGER NOT NULL,
		format TEXT NOT NULL,
		quality TEXT NOT NULL DEFAULT '720p',
		size INTEGER NOT NULL DEFAULT 0,
		expected_size INTEGER NOT NULL DEFAULT 0,
		file_path TEXT NOT NULL DEFAULT '',
		original_url TEXT NOT NULL,
		status TEXT NOT NULL,
		expires_at TIMESTAMP
	);
	`
	if _, err := pool.Exec(ctx, clipsSchema); err != nil {
		return fmt.Errorf("failed to create clips table: %w", err)
	}

	// Create indexes for clips table
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_clips_status ON clips(status);",
		"CREATE INDEX IF NOT EXISTS idx_clips_created_at ON clips(created_at DESC);",
		"CREATE INDEX IF NOT EXISTS idx_clips_expires_at ON clips(expires_at);",
	}

	for _, indexSQL := range indexes {
		if _, err := pool.Exec(ctx, indexSQL); err != nil {
			return fmt.Errorf("failed to create index: %w", err)
		}
	}

	// Create video_metadata table
	metadataSchema := `
	CREATE TABLE IF NOT EXISTS video_metadata (
		id SERIAL PRIMARY KEY,
		url TEXT NOT NULL UNIQUE,
		title TEXT NOT NULL,
		duration INTEGER NOT NULL,
		duration_string TEXT,
		channel TEXT,
		channel_url TEXT,
		uploader TEXT,
		uploader_id TEXT,
		upload_date TEXT,
		thumbnail TEXT,
		categories TEXT,
		ext TEXT,
		filesize_approx INTEGER,
		formats TEXT,
		webpage_url TEXT,
		created_at TIMESTAMP NOT NULL
	);
	`
	if _, err := pool.Exec(ctx, metadataSchema); err != nil {
		return fmt.Errorf("failed to create video_metadata table: %w", err)
	}

	// Create index for video_metadata
	if _, err := pool.Exec(ctx, "CREATE INDEX IF NOT EXISTS idx_video_metadata_url ON video_metadata(url);"); err != nil {
		return fmt.Errorf("failed to create video_metadata index: %w", err)
	}

	log.Println("PostgreSQL database initialized successfully")
	return nil
}
