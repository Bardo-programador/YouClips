package repository

import (
	"context"
	"database/sql"
	"log"
	"time"
	"youclips/internal/entities"
)

type ClipRepository interface {
	Create(ctx context.Context, clip *entities.Clip) error
	GetByID(ctx context.Context, id int) (*entities.Clip, error)
	List(ctx context.Context, limit, offset int) ([]*entities.Clip, error)
	Count(ctx context.Context) (int, error)
	UpdateStatus(ctx context.Context, id int, status entities.ClipStatus) error
	Update(ctx context.Context, clip *entities.Clip) error
	Delete(ctx context.Context, id int) (bool, error)
	FindExpiredClips(ctx context.Context) ([]*entities.Clip, error)
}

type SQLiteClipRepository struct {
	db *sql.DB
}

func NewSQLiteClipRepository(db *sql.DB) *SQLiteClipRepository {
	return &SQLiteClipRepository{db: db}
}

func (r *SQLiteClipRepository) Create(ctx context.Context, clip *entities.Clip) error {
	query := `
		INSERT INTO clips (created_at, title, start_time, end_time, duration_seconds, format, quality, size, expected_size, file_path, original_url, status, expires_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	
	result, err := r.db.ExecContext(
		ctx, query,
		time.Now(),
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
		SELECT id, created_at, title, start_time, end_time, duration_seconds, format, quality, size, expected_size, file_path, original_url, status, expires_at
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
		&clip.Quality,
		&clip.Size,
		&clip.ExpectedSize,
		&clip.FilePath,
		&clip.OriginalURL,
		&clip.Status,
		&clip.ExpiresAt,
	)
	
	if err == sql.ErrNoRows {
		return nil, nil
	}
	
	return clip, err
}

func (r *SQLiteClipRepository) List(ctx context.Context, limit, offset int) ([]*entities.Clip, error) {
	query := `
		SELECT id, created_at, title, start_time, end_time, duration_seconds, format, quality, size, expected_size, file_path, original_url, status, expires_at
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
		SET title = ?, size = ?, expected_size = ?, file_path = ?, status = ?, expires_at = ?
		WHERE id = ?`
	
	_, err := r.db.ExecContext(
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

func (r *SQLiteClipRepository) Delete(ctx context.Context, id int) (bool, error) {
	query := `DELETE FROM clips WHERE id = ?`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return false, err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, err
	}
	return rowsAffected > 0, nil
}

func (r *SQLiteClipRepository) FindExpiredClips(ctx context.Context) ([]*entities.Clip, error) {
	query := `
		SELECT id, created_at, title, start_time, end_time, duration_seconds, format, size, file_path, original_url, status, expires_at
		FROM clips
		WHERE status = ? AND expires_at IS NOT NULL AND expires_at <= ?`
	
	rows, err := r.db.QueryContext(ctx, query, entities.StatusCompleted, time.Now())
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

func InitSQLiteDatabase(db *sql.DB) error {
	// Check if clips table exists
	var tableName string
	err := db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='clips'").Scan(&tableName)
	tableExists := err == nil

	if !tableExists {
		// Create table with all columns including expires_at, quality, and expected_size
		// Note: progress is NOT stored - it's calculated on-demand
		schema := `
		CREATE TABLE clips (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			created_at DATETIME NOT NULL,
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
			expires_at DATETIME
		);
		CREATE INDEX idx_clips_status ON clips(status);
		CREATE INDEX idx_clips_created_at ON clips(created_at DESC);
		CREATE INDEX idx_clips_expires_at ON clips(expires_at);
		`
		if _, err := db.Exec(schema); err != nil {
			return err
		}
	} else {
		
		// Check if expires_at column exists
		var expiresAtExists bool
		err = db.QueryRow("SELECT COUNT(*) FROM pragma_table_info('clips') WHERE name='expires_at'").Scan(&expiresAtExists)
		if err == nil && !expiresAtExists {
			// Add expires_at column
			if _, err := db.Exec("ALTER TABLE clips ADD COLUMN expires_at DATETIME"); err != nil {
				return err
			}
			log.Println("Added expires_at column to clips table")
		}
		
		// Check if quality column exists
		var qualityExists bool
		err = db.QueryRow("SELECT COUNT(*) FROM pragma_table_info('clips') WHERE name='quality'").Scan(&qualityExists)
		if err == nil && !qualityExists {
			// Add quality column
			if _, err := db.Exec("ALTER TABLE clips ADD COLUMN quality TEXT NOT NULL DEFAULT '720p'"); err != nil {
				log.Println("Warning: could not add quality column:", err)
			} else {
				log.Println("Added quality column to clips table")
			}
		}
		
		// Check if expected_size column exists
		var expectedSizeExists bool
		err = db.QueryRow("SELECT COUNT(*) FROM pragma_table_info('clips') WHERE name='expected_size'").Scan(&expectedSizeExists)
		if err == nil && !expectedSizeExists {
			// Add expected_size column
			if _, err := db.Exec("ALTER TABLE clips ADD COLUMN expected_size INTEGER NOT NULL DEFAULT 0"); err != nil {
				log.Println("Warning: could not add expected_size column:", err)
			} else {
				log.Println("Added expected_size column to clips table")
			}
		}
		
		// Remove progress column if it exists (progress is now calculated on-demand)
		var progressExists bool
		err = db.QueryRow("SELECT COUNT(*) FROM pragma_table_info('clips') WHERE name='progress'").Scan(&progressExists)
		if err == nil && progressExists {
			// SQLite doesn't support DROP COLUMN in older versions, so we need to recreate the table
			log.Println("Removing progress column from clips table (recreating table)...")
			
			// Create new table without progress column
			recreateSchema := `
			CREATE TABLE clips_new (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				created_at DATETIME NOT NULL,
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
				expires_at DATETIME
			);
			INSERT INTO clips_new SELECT id, created_at, title, start_time, end_time, duration_seconds, format, quality, size, expected_size, file_path, original_url, status, expires_at FROM clips;
			DROP TABLE clips;
			ALTER TABLE clips_new RENAME TO clips;
			CREATE INDEX idx_clips_status ON clips(status);
			CREATE INDEX idx_clips_created_at ON clips(created_at DESC);
			CREATE INDEX idx_clips_expires_at ON clips(expires_at);
			`
			
			if _, err := db.Exec(recreateSchema); err != nil {
				log.Println("Warning: could not remove progress column:", err)
			} else {
				log.Println("Successfully removed progress column from clips table")
			}
		}
		
		// Ensure indexes exist
		db.Exec("CREATE INDEX IF NOT EXISTS idx_clips_status ON clips(status)")
		db.Exec("CREATE INDEX IF NOT EXISTS idx_clips_created_at ON clips(created_at DESC)")
		db.Exec("CREATE INDEX IF NOT EXISTS idx_clips_expires_at ON clips(expires_at)")
	}

	// Create video_metadata table
	metadataSchema := `
	CREATE TABLE IF NOT EXISTS video_metadata (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
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
		created_at DATETIME NOT NULL
	);
	CREATE INDEX IF NOT EXISTS idx_video_metadata_url ON video_metadata(url);
	`
	if _, err := db.Exec(metadataSchema); err != nil {
		return err
	}

	// Check and add new columns to existing video_metadata table if they don't exist
	newColumns := []struct {
		name       string
		definition string
	}{
		{"duration_string", "TEXT"},
		{"channel", "TEXT"},
		{"channel_url", "TEXT"},
		{"uploader", "TEXT"},
		{"uploader_id", "TEXT"},
		{"upload_date", "TEXT"},
		{"thumbnail", "TEXT"},
		{"categories", "TEXT"},
		{"ext", "TEXT"},
		{"filesize_approx", "INTEGER"},
		{"formats", "TEXT"},
		{"webpage_url", "TEXT"},
		{"max_quality", "TEXT"},
	}

	for _, col := range newColumns {
		var colExists bool
		err := db.QueryRow("SELECT COUNT(*) FROM pragma_table_info('video_metadata') WHERE name=?", col.name).Scan(&colExists)
		if err == nil && !colExists {
			alterSQL := "ALTER TABLE video_metadata ADD COLUMN " + col.name + " " + col.definition
			if _, err := db.Exec(alterSQL); err != nil {
				log.Printf("Warning: could not add %s column to video_metadata: %v", col.name, err)
			} else {
				log.Printf("Added %s column to video_metadata table", col.name)
			}
		}
	}

	return nil
}
