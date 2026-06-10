package postgres

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"simple-server/internal/model"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

type NoteStorage struct {
	db *sqlx.DB
}

func NewNoteStorage(db *sqlx.DB) (*NoteStorage, error) {
	return &NoteStorage{
		db: db,
	}, nil
}

// список заметок с фильтрами
func (s *NoteStorage) GetNotes(ctx context.Context, filters model.NotesFilters) ([]model.Note, error) {
	notes := []model.Note{}
	query := "SELECT * FROM notes WHERE 1=1"

	args := make([]interface{}, 0)
	if filters.Header != nil {
		query += " AND header = $1"
		args = append(args, *filters.Header)
	}

	if err := s.db.SelectContext(ctx, &notes, query, args...); err != nil {
		return nil, fmt.Errorf("select failed: %w", err)
	}
	return notes, nil
}

// получение заметки по ID
func (s *NoteStorage) GetNoteByID(ctx context.Context, id uuid.UUID) (*model.Note, error) {
	var note model.Note
	query := "SELECT * FROM notes WHERE notes.id = $1"
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
