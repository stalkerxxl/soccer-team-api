-- +goose Up
-- +goose StatementBegin
CREATE TABLE countries (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    iso_code TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT countries_name_uq UNIQUE (name),
    CONSTRAINT countries_iso_code_uq UNIQUE (iso_code),
    CONSTRAINT countries_iso_code_format_chk CHECK (
        iso_code = UPPER(iso_code)
        AND char_length(iso_code) = 2
    )
);

INSERT INTO countries (id, name, iso_code)
VALUES
    (1, 'United Kingdom', 'GB'),
    (2, 'Georgia', 'GE'),
    (3, 'Germany', 'DE'),
    (4, 'Spain', 'ES'),
    (5, 'France', 'FR'),
    (6, 'Italy', 'IT'),
    (7, 'Portugal', 'PT'),
    (8, 'Netherlands', 'NL'),
    (9, 'Brazil', 'BR'),
    (10, 'Argentina', 'AR');

-- Advance the sequence after explicit seed ids so future inserts continue from the current max id.
SELECT setval(
    pg_get_serial_sequence('countries', 'id'),
    (SELECT MAX(id) FROM countries)
);

CREATE TABLE teams (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    name TEXT NOT NULL DEFAULT 'Your team',
    country_id BIGINT NOT NULL DEFAULT 1,
    budget BIGINT NOT NULL DEFAULT 5000000,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT teams_user_id_uq UNIQUE (user_id),
    CONSTRAINT teams_user_id_fk
        FOREIGN KEY (user_id)
        REFERENCES users (id)
        ON DELETE CASCADE,
    CONSTRAINT teams_country_id_fk
        FOREIGN KEY (country_id)
        REFERENCES countries (id),
    CONSTRAINT teams_budget_non_negative_chk CHECK (budget >= 0)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS teams;
DROP TABLE IF EXISTS countries;
-- +goose StatementEnd
