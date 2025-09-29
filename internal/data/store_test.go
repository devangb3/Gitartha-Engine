package data_test

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/devang/Gitartha-Engine/internal/data"
)

func setupMockStore(t *testing.T) (*sql.DB, sqlmock.Sqlmock, *data.Store) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	store := data.NewStore(db)
	return db, mock, store
}

func TestListChapters(t *testing.T) {
	_, mock, store := setupMockStore(t)

	expectedRows := sqlmock.NewRows([]string{"id", "name_en", "name_hi", "summary_en", "summary_hi", "verse_count"}).
		AddRow(1, "Chapter 1", "अध्याय 1", "Summary EN", "सारांश", 47)

	mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT id, name_en, COALESCE(name_hi, ''), COALESCE(summary_en, ''), COALESCE(summary_hi, ''), verse_count
            FROM chapters
            ORDER BY id`,
	)).WillReturnRows(expectedRows)

	chapters, err := store.ListChapters(context.Background())
	if err != nil {
		t.Fatalf("ListChapters returned error: %v", err)
	}

	if len(chapters) != 1 {
		t.Fatalf("expected 1 chapter, got %d", len(chapters))
	}

	if chapters[0].NameEN != "Chapter 1" {
		t.Errorf("unexpected chapter name: %s", chapters[0].NameEN)
	}
}

func TestGetVerse_NotFound(t *testing.T) {
	_, mock, store := setupMockStore(t)

	mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT id, chapter_id, verse_number, COALESCE(sanskrit, ''), COALESCE(transliteration, ''), COALESCE(english, ''), COALESCE(hindi, '')
			FROM verses
			WHERE chapter_id = $1 AND verse_number = $2`,
	)).WithArgs(1, 1).WillReturnError(sql.ErrNoRows)

	_, err := store.GetVerse(context.Background(), 1, 1)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(err, data.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestRandomVerse(t *testing.T) {
	_, mock, store := setupMockStore(t)

	expectedRows := sqlmock.NewRows([]string{"id", "chapter_id", "verse_number", "sanskrit", "transliteration", "english", "hindi"}).
		AddRow("uuid", 1, 1, "sanskrit", "trans", "english", "hindi")

	mock.ExpectQuery(regexp.QuoteMeta(
		`SELECT id, chapter_id, verse_number, COALESCE(sanskrit, ''), COALESCE(transliteration, ''), COALESCE(english, ''), COALESCE(hindi, '')
			FROM verses
			ORDER BY random()
			LIMIT 1`,
	)).WillReturnRows(expectedRows)

	verse, err := store.RandomVerse(context.Background())
	if err != nil {
		t.Fatalf("RandomVerse returned error: %v", err)
	}

	if verse.VerseNumber != 1 {
		t.Errorf("expected verse number 1, got %d", verse.VerseNumber)
	}
}
