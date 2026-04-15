# Soccer Team API

A REST API for a fantasy soccer manager written in Go.

Current project status:

- user signup and login with JWT access tokens;
- automatic creation of one team per user;
- generation of an initial squad of 20 players;
- viewing and updating the current user's team and players;
- transfer market: list player, cancel listing, buy player;
- error localization in English and Georgian via `Accept-Language`.

## Technologies

- Go
- Chi
- Bun + pgdriver
- PostgreSQL
- Goose
- JWT (`golang-jwt/jwt/v5`)
- bcrypt
- `log/slog`

## Implemented Endpoints

- `POST /auth/signup`
- `POST /auth/login`
- `GET /me/team`
- `PATCH /me/team`
- `GET /me/players`
- `GET /me/players/{playerId}`
- `PATCH /me/players/{playerId}`
- `GET /market/players`
- `POST /market/players/{playerId}/list`
- `DELETE /market/players/{playerId}/list`
- `POST /market/listings/{listingId}/buy`
- `GET /health`

## Postman Collection

The repository includes a ready-to-use collection:

- `postman/Soccer Team API.postman_collection.json`

The collection covers:

- all current endpoints;
- the main happy path;
- basic error scenarios.

Recommended happy path order in Postman:

1. `Health Check`
2. `Signup Seller`
3. `Get Seller Team`
4. `Get Seller Players`
5. `Get Seller Player`
6. `Update Seller Team`
7. `Update Seller Player`
8. `Create Listing as Seller`
9. `List Public Market`
10. `Signup Buyer`
11. `Get Buyer Team`
12. `Get Buyer Players`
13. `Buy Listing as Buyer`

After that, you can run the error scenarios from the `Auth`, `Me`, and `Market` folders.

The collection uses collection variables and automatically stores:

- `sellerAccessToken`
- `buyerAccessToken`
- `sellerPlayerId`
- `sellerTeamCountryId`
- `sellerPlayerCountryId`
- `listingId`

## Requirements

To run the project in the full Docker scenario you only need:

- Docker
- Docker Compose

For local development without Docker for the API, you additionally need:

- Go

`goose` is not required on the host machine. Migrations run inside a dedicated `migrate` container.
To reduce Docker image size, `goose` in that container is built with PostgreSQL support only.

## Quick Start

### 1. Prepare environment variables

Copy the template:

```bash
cp .env.example .env
```

By default, `.env.example` already contains working local values for Docker Compose.

### 2. Build and start the project

```bash
make up
```

This command does everything needed to verify the project:

- starts PostgreSQL;
- builds the Docker image for the API;
- runs the `migrate` container and applies all migrations;
- starts the `api` container.

After startup, the API is available at:

```text
http://localhost:8080
```

### 3. Check logs if needed

```bash
make logs
```

## Main Commands

```bash
make help            # list available commands
make up              # build images and start postgres + migrate + api
make down            # stop containers
make logs            # follow postgres, migrate, and api logs
make migrate-up      # apply migrations inside the migrate container
make migrate-down    # roll back the last migration inside the migrate container
make migrate-status  # show migration status inside the migrate container
make migrate-create  # create a new SQL migration file locally
make run             # run the API locally without Docker
make test            # run tests
```

## Run Modes

### Full Docker Mode

This is the recommended scenario:

```bash
cp .env.example .env
make up
```

### Run the API Locally on Top of Docker PostgreSQL

This mode is convenient for development if you want to run the API locally but keep PostgreSQL in Docker:

```bash
cp .env.example .env
docker compose up -d postgres
make migrate-up
make run
```

If the full Docker setup was already started before that, stop the `api` container first so that port `8080` is not occupied twice:

```bash
docker compose stop api
```

## Configuration

For local runs, the application reads configuration from `.env`.
In the full Docker scenario, values are passed into containers through `compose.yaml`.

### Main Variables

| Variable | Purpose | Default value |
|---|---|---|
| `APP_ENV` | application environment: `local`, `test`, `prod` | `local` |
| `LOG_LEVEL` | log level | `info` |
| `HTTP_HOST` | HTTP server host | `0.0.0.0` |
| `HTTP_PORT` | HTTP server port | `8080` |
| `DB_HOST` | PostgreSQL host | `localhost` |
| `DB_PORT` | PostgreSQL port | `5432` |
| `DB_NAME` | database name | `soccer_team_api` |
| `DB_USER` | database user | `postgres` |
| `DB_PASSWORD` | database password | `postgres` |
| `DB_SSLMODE` | PostgreSQL SSL mode | `disable` |
| `JWT_ACCESS_SECRET` | access token secret | `change-me-access-secret` |
| `JWT_ACCESS_TTL` | access token TTL | `15m` |
| `AUTH_PASSWORD_MIN_LEN` | minimum password length in characters | `5` |
| `AUTH_PASSWORD_MAX_LEN` | maximum password length in bytes | `72` |
| `DEFAULT_LOCALE` | fallback locale | `en` |

In the full Docker scenario, `compose.yaml` automatically overrides `DB_HOST=postgres` and `DB_PORT=5432` inside the `api` and `migrate` containers.

Additional HTTP and DB timeout variables:

- `HTTP_READ_TIMEOUT`
- `HTTP_WRITE_TIMEOUT`
- `HTTP_IDLE_TIMEOUT`
- `HTTP_SHUTDOWN_TIMEOUT`
- `DB_CONNECT_TIMEOUT`

## Business Logic

### Signup

After a successful signup, the system automatically creates:

- 1 team;
- 20 players;
- initial team budget: `5_000_000`;
- initial market value of each player: `1_000_000`.

Initial squad composition:

- 3 goalkeepers
- 6 defenders
- 6 midfielders
- 5 attackers

### Team / Player Management

A user can update only their own data:

- team: `name`, `country_id`;
- player: `first_name`, `last_name`, `country_id`.

When a team is fetched, the API also returns:

- `total_players_market_value` - the sum of `market_value` for all players on the team.

The following fields cannot be changed through the API:

- `team_id`
- `position`
- `age`
- `market_value`

### Transfer Market

The supported flow is:

- a team owner lists a player on the market with an `asking_price`;
- the market shows only active listings;
- the owner can cancel the listing;
- the purchase is executed in a transaction;
- the buyer's budget decreases;
- the seller's budget increases;
- the player moves to the new team;
- the player's `market_value` increases by a random percentage from `10%` to `100%`;
- the listing is closed with status `sold`;
- a transfer record is written to the `transfers` table.

## Localization

Errors are localized using the `Accept-Language` header.

Supported languages:

- `en`
- `ka`

If the language is not supported or not provided, the fallback is `en`.

Examples:

```http
Accept-Language: en
Accept-Language: ka
Accept-Language: ka-GE,ka;q=0.9,en;q=0.8
```

The `code` field in error responses is not localized and remains stable.

## Error Format

All error responses use this format:

```json
{
  "code": "invalid_request",
  "message": "Invalid request body."
}
```

Example error codes:

- `invalid_request`
- `invalid_email`
- `invalid_password`
- `invalid_credentials`
- `invalid_access_token`
- `invalid_team_name`
- `invalid_player`
- `invalid_country`
- `invalid_player_id`
- `invalid_listing_id`
- `invalid_asking_price`
- `player_not_owned`
- `listing_already_active`
- `cannot_buy_own_player`
- `insufficient_budget`
- `team_not_found`
- `player_not_found`
- `listing_not_found`
- `internal_error`

## API Overview

### Public Endpoints

| Method | Path | Description |
|---|---|---|
| `GET` | `/health` | health check |
| `POST` | `/auth/signup` | signup |
| `POST` | `/auth/login` | login |
| `GET` | `/market/players` | list active listings |

### Protected Endpoints

All protected endpoints require this header:

```http
Authorization: Bearer <access_token>
```

| Method | Path | Description |
|---|---|---|
| `GET` | `/me/team` | get current team |
| `PATCH` | `/me/team` | update current team |
| `GET` | `/me/players` | list current team players |
| `GET` | `/me/players/{playerId}` | get a player from the current team |
| `PATCH` | `/me/players/{playerId}` | update a player from the current team |
| `POST` | `/market/players/{playerId}/list` | list your player on the market |
| `DELETE` | `/market/players/{playerId}/list` | cancel your player's listing |
| `POST` | `/market/listings/{listingId}/buy` | buy a player |

## Recommended Manual Happy Path

1. Register user A via `/auth/signup`.
2. Get the access token.
3. Check `/me/team` and `/me/players`.
4. List one of user A's players on the market.
5. Register user B.
6. Under user B's token, call `/market/players`.
7. Buy the listing via `/market/listings/{listingId}/buy`.
8. Check `/me/team` and `/me/players` again for both teams.
9. Verify that after the purchase:
   - the buyer and seller budgets changed;
   - the player moved to the new team;
   - `total_players_market_value` was recalculated.

## Tests

Run all tests:

```bash
make test
```

The project currently includes unit and handler tests for:

- auth
- me endpoints
- transfer market
- i18n
- config validation

## What Was Intentionally Simplified for the Take-Home Task

The project contains a few decisions that are reasonable for a take-home assignment, but would typically be expanded in a production system.

### Authentication

The current implementation uses only JWT access tokens, without a refresh token flow.

For a real project, it would make sense to add:

- refresh tokens;
- token revocation;
- refresh token rotation;
- logout / session management;
- storage and audit of active sessions.

I intentionally did not add that here to keep the auth flow focused and avoid expanding the scope beyond the assignment.

### HTTP Contracts and DTOs

DTOs are not used everywhere: in some places HTTP handlers work directly with domain/model structures.

In a real project, a stricter DTO layer would usually be introduced:

- separate request/response models;
- explicit separation between domain and HTTP contracts;
- independent API evolution without coupling to internal models.

For this take-home task, I kept the implementation more direct to avoid unnecessary boilerplate.

### Logging and Observability

The project includes a configurable logger, but it does not have full audit/debug logging for every service-layer action.

In a production setup, this would usually be extended with:

- structured logs for key business operations;
- correlation/request IDs;
- audit logging for sensitive actions;
- metrics and tracing.

Here I kept logging intentionally basic because the main goal of the assignment is to demonstrate API behavior, transaction handling, and project structure.

### Overall Principle

In all of these areas, I deliberately aimed for a balance between engineering quality and the scope of the take-home assignment:

- implement the full working functionality;
- keep the architecture readable;
- avoid production over-engineering where it is not required by the assignment.

## Project Structure

```text
cmd/api                 # HTTP API entry point
docker                  # scripts for Docker containers
internal/api            # handlers, router, response helpers
internal/api/middleware # auth middleware
internal/auth           # JWT provider and auth context helpers
internal/config         # config loading and validation
internal/db             # bun bootstrap and tx manager
internal/domain         # bun/domain models
internal/i18n           # localizer and embedded catalogs
internal/repository     # database access
internal/service        # business logic
migrations              # SQL migrations
postman                 # Postman collection for manual API verification
compose.yaml            # Docker Compose for postgres, migrate, and api
Dockerfile              # build for the API and migration container
Makefile                # commands for local development and Docker flow
```
