package repository

import (
	"context"
	"database/sql"
	"time"
	"videodownloader/internal/entities"
)

type ClipRepository interface {
	Create(ctx context.Context, clip *entities.Clip) error
	GetByID(ctx context.Context, id int) (*entities.Clip, error)
	List(ctx context.Context, limit, offset int) ([]*entities.Clip, error)
	Count(ctx context.Context) (int, error)
	UpdateStatus(ctx context.Context, id int, status entities.ClipStatus) error
	Update(ctx context.Context, clip *entities.Clip) error
}

type SQLiteClipRepository struct {
	db *sql.DB
}

func NewSQLiteClipRepository(db *sql.DB) *SQLiteClipRepository {
	return &SQLiteClipRepository{db: db}
}

func (r *SQLiteClipRepository) Create(ctx context.Context, clip *entities.Clip) error {
	query := `
		INSERT INTO clips (created_at, title, start_time, end_time, duration_seconds, format, size, file_path, original_url, status)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	
	result, err := r.db.ExecContext(
		ctx, query,
		time.Now(),
		clip.Title,
		clip.StartTime,
		clip.EndTime,
		clip.DurationSeconds,
		clip.Format,
		clip.Size,
		clip.FilePath,
		clip.OriginalURL,
		clip.Status,
	)
	if err != nil {
		return err
	}
	
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	clip.ID = int(id)
	return nil
}

func (r *SQLiteClipRepository) GetByID(ctx context.Context, id int) (*entities.Clip, error) {
	query := `
		SELECT id, created_at, title, start_time, end_time, duration_seconds, format, size, file_path, original_url, status
		FROM clips
		WHERE id = ?`
	
	clip := &entities.Clip{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
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
	)
	
	if err == sql.ErrNoRows {
		return nil, nil
	}
	
	return clip, err
}

func (r *SQLiteClipRepository) List(ctx context.Context, limit, offset int) ([]*entities.Clip, error) {
	query := `
		SELECT id, created_at, title, start_time, end_time, duration_seconds, format, size, file_path, original_url, status
		FROM clips
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?`
	
	rows, err := r.db.QueryContext(ctx, query, limit, offset)
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
		)
		if err != nil {
			return nil, err
		}
		clips = append(clips, clip)
	}
	
	return clips, rows.Err()
}

func (r *SQLiteClipRepository) Count(ctx context.Context) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM clips`
	err := r.db.QueryRowContext(ctx, query).Scan(&count)
	return count, err
}

func (r *SQLiteClipRepository) UpdateStatus(ctx context.Context, id int, status entities.ClipStatus) error {
	query := `UPDATE clips SET status = ? WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, status, id)
	return err
}

func (r *SQLiteClipRepository) Update(ctx context.Context, clip *entities.Clip) error {
	query := `
		UPDATE clips 
		SET title = ?, size = ?, file_path = ?, status = ?
		WHERE id = ?`
	
	_, err := r.db.ExecContext(
		ctx, query,
		clip.Title,
		clip.Size,
		clip.FilePath,
		clip.Status,
		clip.ID,
	)
	return err
}
