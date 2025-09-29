-- +migrate Up
CREATE TABLE IF NOT EXISTS chapters (
    id SMALLINT PRIMARY KEY,
    name_en TEXT NOT NULL,
    name_hi TEXT,
    summary_en TEXT,
    summary_hi TEXT,
    verse_count SMALLINT DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS verses (
    id UUID PRIMARY KEY,
    chapter_id SMALLINT NOT NULL REFERENCES chapters(id) ON DELETE CASCADE,
    verse_number SMALLINT NOT NULL,
    sanskrit TEXT,
    transliteration TEXT,
    english TEXT,
    hindi TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (chapter_id, verse_number)
);

CREATE INDEX IF NOT EXISTS verses_chapter_id_idx ON verses(chapter_id);
CREATE INDEX IF NOT EXISTS verses_chapter_verse_idx ON verses(chapter_id, verse_number);

-- +migrate Down
DROP TABLE IF EXISTS verses;
DROP TABLE IF EXISTS chapters;
