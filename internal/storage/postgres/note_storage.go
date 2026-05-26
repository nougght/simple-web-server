package postgres

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"log"
	"simple-server/internal/config"
	"simple-server/internal/model"

	"github.com/golang-migrate/migrate/v4"
	pg "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

type NoteStorage struct {
	config *config.PostgresConfig
	db     *sqlx.DB
}

func RunMigrations(db *sqlx.DB) error {
	sourceDriver, err := iofs.New(migrationsFS, "migrations")
	if err != nil {
		return fmt.Errorf("failed to create iofs driver: %w", err)
	}

	dbDriver, err := pg.WithInstance(db.DB, &pg.Config{})
	if err != nil {
		return fmt.Errorf("failed to create postgres driver: %w", err)
	}

	m, err := migrate.NewWithInstance("iofs", sourceDriver, "postgres", dbDriver)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}

	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			log.Println("no changes to migrate")
			return nil
		}
		return fmt.Errorf("migration failed: %w", err)
	}

	return nil
}

func NewNoteStorage(cfg *config.PostgresConfig) (*NoteStorage, error) {
	db, err := sqlx.Connect("postgres", cfg.ConnectionString())
	if err != nil {
		return nil, fmt.Errorf("postgres connection failed: %w", err)
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("database ping failed: %w", err)
	}

	if err := RunMigrations(db); err != nil {
		return nil, fmt.Errorf("migrations failed: %w", err)
	}
	log.Println("successful migrations")

	return &NoteStorage{
		config: cfg,
		db:     db,
	}, nil
}

// список заметок с фильтрами
func (s *NoteStorage) GetNotes(ctx context.Context, filters map[string]interface{}) ([]model.Note, error) {
	notes := []model.Note{}
	query := "SELECT * FROM notes WHERE 1=1"

	// добавляем фильтры если они есть
	for key, value := range filters {
		query += fmt.Sprintf(" AND %s = '%v'", key, value)
	}
	if err := s.db.SelectContext(ctx, &notes, query); err != nil {
		return nil, fmt.Errorf("select failed: %w", err)
	}
	return notes, nil
}

// получение заметки по ID
func (s *NoteStorage) GetNoteByID(ctx context.Context, ID uuid.UUID) (*model.Note, error) {
	var note model.Note
	query := "SELECT * FROM notes WHERE notes.id = $1"
	if err := s.db.GetContext(ctx, &note, query, ID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, model.ErrNotFound
		}
		return nil, fmt.Errorf("select failed: %w", err)
	}
	return &note, nil
}

// добавление заметки
func (s *NoteStorage) AddNote(ctx context.Context, note *model.Note) (*model.Note, error) {
	// insert с возвратом сгенерированного ID
	query := `INSERT INTO notes(header, body)
			  VALUES ($1, $2)
			  RETURNING id`

	err := s.db.QueryRowxContext(ctx, query, note.Header, note.Body).Scan(&note.ID)
	if err != nil {
		return nil, fmt.Errorf("insert failed: %w", err)
	}
	return note, nil
}

// обновление заметки
func (s *NoteStorage) UpdateNote(ctx context.Context, note *model.Note) error {
	query := `UPDATE notes
			  SET header = $1, body = $2
			  WHERE notes.id = $3`

	res, err := s.db.ExecContext(ctx, query, note.Header, note.Body, note.ID)
	if err != nil {
		return fmt.Errorf("update failed: %w", err)
	}

	// если заметки с таким ID нет - возвращаем ошибку
	rowsCount, _ := res.RowsAffected()
	if rowsCount == 0 {
		return model.ErrNotFound
	}
	return nil
}

// удаление заметки
func (s *NoteStorage) DeleteNote(ctx context.Context, noteID uuid.UUID) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM notes WHERE notes.id = $1`, noteID)
	if err != nil {
		return fmt.Errorf("delete failed: %w", err)
	}
	return nil
}

// проверка существования заметки (не используется)
func (s *NoteStorage) NoteExists(ctx context.Context, noteID uuid.UUID) (bool, error) {
	var exists bool
	err := s.db.GetContext(ctx, &exists, "SELECT EXISTS(SELECT 1 FROM notes WHERE notes.id = $1)", noteID)
	if err != nil {
		return false, err
	}
	return exists, nil
}
