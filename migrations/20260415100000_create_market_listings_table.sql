-- +goose Up
-- +goose StatementBegin
CREATE TABLE market_listings (
    id BIGSERIAL PRIMARY KEY,
    player_id BIGINT NOT NULL,
    seller_team_id BIGINT NOT NULL,
    asking_price BIGINT NOT NULL,
    status TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT market_listings_player_id_fk
        FOREIGN KEY (player_id)
        REFERENCES players (id)
        ON DELETE CASCADE,
    CONSTRAINT market_listings_seller_team_id_fk
        FOREIGN KEY (seller_team_id)
        REFERENCES teams (id)
        ON DELETE CASCADE,
    CONSTRAINT market_listings_asking_price_positive_chk CHECK (asking_price > 0),
    CONSTRAINT market_listings_status_chk CHECK (
        status IN ('active', 'sold', 'cancelled')
    )
);

-- Allow historical sold/cancelled rows while still guaranteeing at most one active listing per player.
CREATE UNIQUE INDEX market_listings_player_id_active_uq
    ON market_listings (player_id)
    WHERE status = 'active';

CREATE INDEX market_listings_status_created_at_idx
    ON market_listings (status, created_at DESC);

CREATE INDEX market_listings_seller_team_id_status_idx
    ON market_listings (seller_team_id, status);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS market_listings_seller_team_id_status_idx;
DROP INDEX IF EXISTS market_listings_status_created_at_idx;
DROP INDEX IF EXISTS market_listings_player_id_active_uq;
DROP TABLE IF EXISTS market_listings;
-- +goose StatementEnd
