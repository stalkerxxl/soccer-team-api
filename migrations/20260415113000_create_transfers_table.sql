-- +goose Up
-- +goose StatementBegin
CREATE TABLE transfers (
    id BIGSERIAL PRIMARY KEY,
    player_id BIGINT NOT NULL,
    seller_team_id BIGINT NOT NULL,
    buyer_team_id BIGINT NOT NULL,
    price BIGINT NOT NULL,
    old_market_value BIGINT NOT NULL,
    new_market_value BIGINT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT transfers_player_id_fk
        FOREIGN KEY (player_id)
        REFERENCES players (id),
    CONSTRAINT transfers_seller_team_id_fk
        FOREIGN KEY (seller_team_id)
        REFERENCES teams (id),
    CONSTRAINT transfers_buyer_team_id_fk
        FOREIGN KEY (buyer_team_id)
        REFERENCES teams (id),
    CONSTRAINT transfers_price_positive_chk CHECK (price > 0),
    CONSTRAINT transfers_old_market_value_non_negative_chk CHECK (old_market_value >= 0),
    CONSTRAINT transfers_new_market_value_non_negative_chk CHECK (new_market_value >= 0)
);

CREATE INDEX transfers_buyer_team_id_created_at_idx
    ON transfers (buyer_team_id, created_at DESC);

CREATE INDEX transfers_seller_team_id_created_at_idx
    ON transfers (seller_team_id, created_at DESC);

CREATE INDEX transfers_player_id_created_at_idx
    ON transfers (player_id, created_at DESC);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS transfers_player_id_created_at_idx;
DROP INDEX IF EXISTS transfers_seller_team_id_created_at_idx;
DROP INDEX IF EXISTS transfers_buyer_team_id_created_at_idx;
DROP TABLE IF EXISTS transfers;
-- +goose StatementEnd
