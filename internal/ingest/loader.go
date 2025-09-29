package ingest

import (
	"bufio"
	"context"
	"database/sql"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strconv"
	"time"

	"github.com/google/uuid"
)

// Loader ingests Bhagavad Gita verses from a CSV file.
type Loader struct {
	db *sql.DB
}

// New creates a new Loader using the provided DB handle.
func New(db *sql.DB) *Loader {
	return &Loader{db: db}
}

// LoadCSV reads the CSV file at path and populates chapters/verses tables.
func (l *Loader) LoadCSV(ctx context.Context, csvPath string) error {
	f, err := os.Open(csvPath)
	if err != nil {
		return fmt.Errorf("open csv: %w", err)
	}
	defer f.Close()

	reader := csv.NewReader(bufio.NewReader(f))
	reader.TrimLeadingSpace = true

	header, err := reader.Read()
	if err != nil {
		return fmt.Errorf("read header: %w", err)
	}

	columnIndex := map[string]int{}
	for idx, col := range header {
		columnIndex[col] = idx
	}

	tx, err := l.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	insertChapterStmt := `INSERT INTO chapters (id, name_en, name_hi, summary_en, summary_hi, verse_count, created_at, updated_at)
        VALUES ($1, $2, '', '', '', 0, NOW(), NOW())
        ON CONFLICT (id) DO NOTHING`

	insertVerseStmt := `INSERT INTO verses (id, chapter_id, verse_number, sanskrit, transliteration, english, hindi, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW())
        ON CONFLICT (chapter_id, verse_number) DO UPDATE
        SET sanskrit = EXCLUDED.sanskrit,
            transliteration = EXCLUDED.transliteration,
            english = EXCLUDED.english,
            hindi = EXCLUDED.hindi,
            updated_at = NOW()`

	chapterVerseCounts := make(map[int]int)

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("read record: %w", err)
		}

		chapterNum, err := parseInt(record[columnIndex["chapter"]])
		if err != nil {
			return fmt.Errorf("parse chapter: %w", err)
		}

		verseNum, err := parseInt(record[columnIndex["verse"]])
		if err != nil {
			return fmt.Errorf("parse verse: %w", err)
		}

		sanskrit := record[columnIndex["sanskrit"]]
		hindi := record[columnIndex["hindi"]]
		english := record[columnIndex["english"]]
		transliteration := record[columnIndex["transliteration"]]

		if _, err := tx.ExecContext(ctx, insertChapterStmt, chapterNum, fmt.Sprintf("Chapter %d", chapterNum)); err != nil {
			return fmt.Errorf("insert chapter %d: %w", chapterNum, err)
		}

		verseID := uuid.New().String()
		if _, err := tx.ExecContext(ctx, insertVerseStmt, verseID, chapterNum, verseNum, sanskrit, transliteration, english, hindi); err != nil {
			return fmt.Errorf("insert verse %d.%d: %w", chapterNum, verseNum, err)
		}

		chapterVerseCounts[chapterNum]++
	}

	updateChapterStmt := `UPDATE chapters SET verse_count = $1, updated_at = $2 WHERE id = $3`

	now := time.Now()
	for chapterID, count := range chapterVerseCounts {
		if _, err := tx.ExecContext(ctx, updateChapterStmt, count, now, chapterID); err != nil {
			return fmt.Errorf("update chapter %d: %w", chapterID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}

	return nil
}

func parseInt(value string) (int, error) {
	n, err := strconv.Atoi(value)
	if err != nil {
		return 0, err
	}
	return n, nil
}
