package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

var ErrNotFound = errors.New("not found")

// Chapter represents chapter metadata.
type Chapter struct {
	ID         int    `json:"id"`
	NameEN     string `json:"name_en"`
	NameHI     string `json:"name_hi,omitempty"`
	SummaryEN  string `json:"summary_en,omitempty"`
	SummaryHI  string `json:"summary_hi,omitempty"`
	VerseCount int    `json:"verse_count"`
}

// Verse represents a Bhagavad Gita verse and translations.
type Verse struct {
	ID              string `json:"id"`
	ChapterID       int    `json:"chapter_id"`
	VerseNumber     int    `json:"verse_number"`
	Sanskrit        string `json:"sanskrit"`
	Transliteration string `json:"transliteration"`
	English         string `json:"english"`
	Hindi           string `json:"hindi"`
}

func (s *Store) Ping(ctx context.Context) error {
	return s.db.PingContext(ctx)
}

func (s *Store) ListChapters(ctx context.Context) ([]Chapter, error) {
	rows, err := s.db.QueryContext(
		ctx,
		`SELECT id, name_en, COALESCE(name_hi, ''), COALESCE(summary_en, ''), COALESCE(summary_hi, ''), verse_count
            FROM chapters
            ORDER BY id`,
	)
	if err != nil {
		return nil, fmt.Errorf("list chapters: %w", err)
	}
	defer rows.Close()

	var chapters []Chapter
	for rows.Next() {
		var ch Chapter
		if err := rows.Scan(&ch.ID, &ch.NameEN, &ch.NameHI, &ch.SummaryEN, &ch.SummaryHI, &ch.VerseCount); err != nil {
			return nil, fmt.Errorf("scan chapter: %w", err)
		}
		chapters = append(chapters, ch)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate chapters: %w", err)
	}

	return chapters, nil
}

func (s *Store) GetChapterWithVerses(ctx context.Context, chapterID int) (Chapter, []Verse, error) {
	var ch Chapter
	row := s.db.QueryRowContext(
		ctx,
		`SELECT id, name_en, COALESCE(name_hi, ''), COALESCE(summary_en, ''), COALESCE(summary_hi, ''), verse_count
			FROM chapters
			WHERE id = $1`,
		chapterID,
	)

	if err := row.Scan(&ch.ID, &ch.NameEN, &ch.NameHI, &ch.SummaryEN, &ch.SummaryHI, &ch.VerseCount); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Chapter{}, nil, ErrNotFound
		}
		return Chapter{}, nil, fmt.Errorf("get chapter: %w", err)
	}

	verses, err := s.listVersesByChapter(ctx, chapterID)
	if err != nil {
		return Chapter{}, nil, err
	}

	return ch, verses, nil
}

func (s *Store) GetVerse(ctx context.Context, chapterID, verseNumber int) (Verse, error) {
	row := s.db.QueryRowContext(
		ctx,
		`SELECT id, chapter_id, verse_number, COALESCE(sanskrit, ''), COALESCE(transliteration, ''), COALESCE(english, ''), COALESCE(hindi, '')
			FROM verses
			WHERE chapter_id = $1 AND verse_number = $2`,
		chapterID,
		verseNumber,
	)

	var verse Verse
	if err := row.Scan(&verse.ID, &verse.ChapterID, &verse.VerseNumber, &verse.Sanskrit, &verse.Transliteration, &verse.English, &verse.Hindi); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Verse{}, ErrNotFound
		}
		return Verse{}, fmt.Errorf("get verse: %w", err)
	}

	return verse, nil
}

// SearchVerses performs a simple keyword search on selected language columns.
func (s *Store) SearchVerses(ctx context.Context, query, lang string, limit int) ([]Verse, error) {
	if limit <= 0 {
		limit = 20
	} else if limit > 100 {
		limit = 100
	}

	column := "english"
	if lang == "hi" {
		column = "hindi"
	}

	pattern := fmt.Sprintf("%%%s%%", query)

	stmt := fmt.Sprintf(`SELECT id, chapter_id, verse_number, COALESCE(sanskrit, ''), COALESCE(transliteration, ''), COALESCE(english, ''), COALESCE(hindi, '')
		FROM verses
		WHERE %s ILIKE $1
		ORDER BY chapter_id, verse_number
		LIMIT $2`, column)

	rows, err := s.db.QueryContext(ctx, stmt, pattern, limit)
	if err != nil {
		return nil, fmt.Errorf("search verses: %w", err)
	}
	defer rows.Close()

	var verses []Verse
	for rows.Next() {
		var verse Verse
		if err := rows.Scan(&verse.ID, &verse.ChapterID, &verse.VerseNumber, &verse.Sanskrit, &verse.Transliteration, &verse.English, &verse.Hindi); err != nil {
			return nil, fmt.Errorf("scan verse: %w", err)
		}
		verses = append(verses, verse)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate search results: %w", err)
	}

	return verses, nil
}

func (s *Store) RandomVerse(ctx context.Context) (Verse, error) {
	row := s.db.QueryRowContext(
		ctx,
		`SELECT id, chapter_id, verse_number, COALESCE(sanskrit, ''), COALESCE(transliteration, ''), COALESCE(english, ''), COALESCE(hindi, '')
			FROM verses
			ORDER BY random()
			LIMIT 1`,
	)

	var verse Verse
	if err := row.Scan(&verse.ID, &verse.ChapterID, &verse.VerseNumber, &verse.Sanskrit, &verse.Transliteration, &verse.English, &verse.Hindi); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Verse{}, ErrNotFound
		}
		return Verse{}, fmt.Errorf("random verse: %w", err)
	}

	return verse, nil
}

func (s *Store) listVersesByChapter(ctx context.Context, chapterID int) ([]Verse, error) {
	rows, err := s.db.QueryContext(
		ctx,
		`SELECT id, chapter_id, verse_number, COALESCE(sanskrit, ''), COALESCE(transliteration, ''), COALESCE(english, ''), COALESCE(hindi, '')
			FROM verses
			WHERE chapter_id = $1
			ORDER BY verse_number`,
		chapterID,
	)
	if err != nil {
		return nil, fmt.Errorf("list verses: %w", err)
	}
	defer rows.Close()

	var verses []Verse
	for rows.Next() {
		var verse Verse
		if err := rows.Scan(&verse.ID, &verse.ChapterID, &verse.VerseNumber, &verse.Sanskrit, &verse.Transliteration, &verse.English, &verse.Hindi); err != nil {
			return nil, fmt.Errorf("scan verse: %w", err)
		}
		verses = append(verses, verse)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate verses: %w", err)
	}

	return verses, nil
}
