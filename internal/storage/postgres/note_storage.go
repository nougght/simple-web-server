package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"simple-server/internal/config"
	"simple-server/internal/model"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type NoteStorage struct {
	config *config.PostgresConfig
	db     *sqlx.DB
}

func NewNoteStorage(cfg *config.PostgresConfig) (*NoteStorage, error) {
	db, err := sqlx.Connect("postgres", cfg.ConnectionString())
	if err != nil {
		return nil, fmt.Errorf("postgres connection failed: %w", err)
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("database ping failed: %w", err)
	}
	return &NoteStorage{
		config: cfg,
		db:     db,
	}, nil
}

// список всех заметок
func (s *NoteStorage) GetNotes(ctx context.Context) ([]model.Note, error) {
	notes := []model.Note{}
	query := "SELECT * FROM notes"
	if err := s.db.SelectContext(ctx, &notes, query); err != nil {
		return nil, fmt.Errorf("select failed: %w", err)
	}
	return notes, nil
}

// получение заметок по заголовку
func (s *NoteStorage) GetNotesByHeader(ctx context.Context, header string) ([]model.Note, error) {
	notes := []model.Note{}
	query := "SELECT * FROM notes WHERE notes.header = $1"
	if err := s.db.SelectContext(ctx, &notes, query, header); err != nil {
		return nil, fmt.Errorf("select failed: %w", err)
	}
	return notes, nil
}

// получение заметки по id
func (s *NoteStorage) GetNoteById(ctx context.Context, id uuid.UUID) (*model.Note, error) {
	var note model.Note
	query := "SELECT * FROM notes WHERE notes.note_id = $1"
	if err := s.db.GetContext(ctx, &note, query, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, model.ErrNotFound
		}
		return nil, fmt.Errorf("select failed: %w", err)
	}
	return &note, nil
}

// добавление заметки
func (s *NoteStorage) AddNote(ctx context.Context, note *model.Note) (*model.Note, error) {
	// insert с возвратом сгенерированного id
	query := `INSERT INTO notes(header, body)
			  VALUES ($1, $2)
			  RETURNING note_id`

	err := s.db.QueryRowxContext(ctx, query, note.Header, note.Body).Scan(&note.NoteId)
	if err != nil {
		return nil, fmt.Errorf("insert failed: %w", err)
	}
	return note, nil
}

// обновление заметки
func (s *NoteStorage) UpdateNote(ctx context.Context, note *model.Note) error {
	query := `UPDATE notes
			  SET header = $1, body = $2
			  WHERE notes.note_id = $3`

	res, err := s.db.ExecContext(ctx, query, note.Header, note.Body, note.NoteId)
	if err != nil {
		return fmt.Errorf("update failed: %w", err)
	}

	// если заметки с таким id нет - возвращаем ошибку
	rowsCount, _ := res.RowsAffected()
	if rowsCount == 0 {
		return model.ErrNotFound
	}
	return nil
}

// удаление заметки
func (s *NoteStorage) DeleteNote(ctx context.Context, noteId uuid.UUID) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM notes WHERE notes.note_id = $1`, noteId)
	if err != nil {
		return fmt.Errorf("delete failed: %w", err)
	}
	return nil
}

// проверка существования заметки (не используется)
func (s *NoteStorage) NoteExists(ctx context.Context, id uuid.UUID) (bool, error) {
	var exists bool
	err := s.db.GetContext(ctx, &exists, "SELECT EXISTS(SELECT 1 FROM notes WHERE notes.note_id = $1)", id)
	if err != nil {
		return false, err
	}
	return exists, nil
}
