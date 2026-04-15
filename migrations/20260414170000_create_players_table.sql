-- +goose Up
-- +goose StatementBegin
CREATE TABLE players (
    id BIGSERIAL PRIMARY KEY,
    team_id BIGINT NOT NULL,
    first_name TEXT NOT NULL,
    last_name TEXT NOT NULL,
    country_id BIGINT NOT NULL,
    age INT NOT NULL,
    position TEXT NOT NULL,
    market_value BIGINT NOT NULL DEFAULT 1000000,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT players_team_id_fk
        FOREIGN KEY (team_id)
        REFERENCES teams (id)
        ON DELETE CASCADE,
    CONSTRAINT players_country_id_fk
        FOREIGN KEY (country_id)
        REFERENCES countries (id),
    CONSTRAINT players_age_range_chk CHECK (age BETWEEN 18 AND 40),
    CONSTRAINT players_market_value_non_negative_chk CHECK (market_value >= 0),
    CONSTRAINT players_position_chk CHECK (
        position IN ('goalkeeper', 'defender', 'midfielder', 'attacker')
    )
);

CREATE INDEX players_team_id_idx ON players (team_id);
CREATE INDEX players_team_id_position_idx ON players (team_id, position);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS players_team_id_position_idx;
DROP INDEX IF EXISTS players_team_id_idx;
DROP TABLE IF EXISTS players;
-- +goose StatementEnd
