# Soccer Team API

REST API для fantasy soccer manager на Go.

Текущее состояние проекта:

- регистрация и логин по JWT access token;
- автосоздание одной команды на пользователя;
- генерация стартового состава из 20 игроков;
- просмотр и редактирование своей команды и игроков;
- трансферный рынок: выставление, снятие с продажи, покупка игрока;
- локализация ошибок на английский и грузинский языки через `Accept-Language`.

## Технологии

- Go
- Chi
- Bun + pgdriver
- PostgreSQL
- Goose
- JWT (`golang-jwt/jwt/v5`)
- bcrypt
- `log/slog`

## Что реализовано

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

## Postman collection

В репозитории есть готовая коллекция:

- `postman/Soccer Team API.postman_collection.json`

Коллекция покрывает:

- все текущие endpoints;
- основной happy path;
- базовые error-сценарии.

Рекомендуемый порядок запуска happy path в Postman:

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

После этого можно проверить error-сценарии из папок `Auth`, `Me` и `Market`.

Коллекция использует collection variables и сама сохраняет:

- `sellerAccessToken`
- `buyerAccessToken`
- `sellerPlayerId`
- `sellerTeamCountryId`
- `sellerPlayerCountryId`
- `listingId`

## Требования

Для запуска проекта в полном Docker-сценарии нужны только:

- Docker
- Docker Compose

Для локальной разработки без Docker для API дополнительно нужен:

- Go

`goose` на хост-машине не нужен. Миграции запускаются в отдельном контейнере `migrate`.
Для уменьшения размера Docker image `goose` в этом контейнере собран только с поддержкой PostgreSQL.

## Быстрый старт

### 1. Подготовить переменные окружения

Скопируйте шаблон:

```bash
cp .env.example .env
```

По умолчанию в `.env.example` уже есть рабочие локальные значения для Docker Compose.

### 2. Собрать и запустить проект

```bash
make up
```

Эта команда делает все нужное для проверки проекта:

- поднимает PostgreSQL;
- собирает Docker image для API;
- запускает контейнер `migrate` и применяет все migrations;
- запускает контейнер `api`.

После старта API будет доступен по адресу:

```text
http://localhost:8080
```

### 3. Проверить логи при необходимости

```bash
make logs
```

## Основные команды

```bash
make help            # список команд
make up              # собрать images и поднять postgres + migrate + api
make down            # остановить контейнеры
make logs            # смотреть логи postgres, migrate и api
make migrate-up      # применить миграции в контейнере migrate
make migrate-down    # откатить последнюю миграцию в контейнере migrate
make migrate-status  # статус миграций в контейнере migrate
make migrate-create  # создать новый SQL migration-файл локально
make run             # запустить API локально без Docker
make test            # запустить тесты
```

## Режимы запуска

### Полностью в Docker

Это рекомендуемый сценарий:

```bash
cp .env.example .env
make up
```

### Локальный запуск API поверх Docker PostgreSQL

Этот сценарий удобен для разработки, если API хочется запускать локально, а PostgreSQL оставить в Docker:

```bash
cp .env.example .env
docker compose up -d postgres
make migrate-up
make run
```

Если до этого уже был запущен полный Docker-сценарий, сначала остановите контейнер `api`, чтобы не занять порт `8080` дважды:

```bash
docker compose stop api
```

## Конфигурация

При локальном запуске приложение читает конфиг из `.env`.
В полном Docker-сценарии значения пробрасываются в контейнеры через `compose.yaml`.

### Основные переменные

| Переменная | Назначение | Значение по умолчанию |
|---|---|---|
| `APP_ENV` | окружение приложения: `local`, `test`, `prod` | `local` |
| `LOG_LEVEL` | уровень логирования | `info` |
| `HTTP_HOST` | хост HTTP-сервера | `0.0.0.0` |
| `HTTP_PORT` | порт HTTP-сервера | `8080` |
| `DB_HOST` | хост PostgreSQL | `localhost` |
| `DB_PORT` | порт PostgreSQL | `5432` |
| `DB_NAME` | имя БД | `soccer_team_api` |
| `DB_USER` | пользователь БД | `postgres` |
| `DB_PASSWORD` | пароль БД | `postgres` |
| `DB_SSLMODE` | режим SSL для PostgreSQL | `disable` |
| `JWT_ACCESS_SECRET` | секрет для access token | `change-me-access-secret` |
| `JWT_ACCESS_TTL` | TTL access token | `15m` |
| `AUTH_PASSWORD_MIN_LEN` | минимальная длина пароля в символах | `5` |
| `AUTH_PASSWORD_MAX_LEN` | максимальная длина пароля в байтах | `72` |
| `DEFAULT_LOCALE` | fallback locale | `en` |

Для полного Docker-сценария `compose.yaml` автоматически переопределяет `DB_HOST=postgres` и `DB_PORT=5432` внутри контейнеров `api` и `migrate`.

Дополнительно настраиваются HTTP и DB timeout'ы:

- `HTTP_READ_TIMEOUT`
- `HTTP_WRITE_TIMEOUT`
- `HTTP_IDLE_TIMEOUT`
- `HTTP_SHUTDOWN_TIMEOUT`
- `DB_CONNECT_TIMEOUT`

## Бизнес-логика (согласно Тестовому Заданию)

### Signup

После успешной регистрации автоматически создаются:

- 1 команда;
- 20 игроков;
- стартовый бюджет команды: `5_000_000`;
- стартовая стоимость каждого игрока: `1_000_000`.

Состав стартовой команды:

- 3 goalkeepers
- 6 defenders
- 6 midfielders
- 5 attackers

### Team / Player management

Пользователь может менять только свои данные:

- у команды: `name`, `country_id`;
- у игрока: `first_name`, `last_name`, `country_id`.

При просмотре команды API также возвращает:

- `total_players_market_value` - сумму `market_value` всех игроков команды.

Нельзя менять через API:

- `team_id`
- `position`
- `age`
- `market_value`

### Transfer market

Поддерживается следующий flow:

- владелец команды выставляет игрока на рынок с `asking_price`;
- рынок показывает только активные листинги;
- владелец может снять игрока с продажи;
- покупка выполняется в транзакции;
- бюджет покупателя уменьшается;
- бюджет продавца увеличивается;
- игрок переходит в новую команду;
- `market_value` игрока увеличивается на случайный процент от `10%` до `100%`;
- листинг закрывается со статусом `sold`;
- в таблицу `transfers` записывается история сделки.

## Локализация

Ошибки локализуются по заголовку `Accept-Language`.

Поддерживаемые языки:

- `en`
- `ka`

Если язык не поддерживается или не передан, используется fallback `en`.

Примеры:

```http
Accept-Language: en
Accept-Language: ka
Accept-Language: ka-GE,ka;q=0.9,en;q=0.8
```

Поле `code` в error response не локализуется и остается постоянным.

## Формат ошибок

Все error responses возвращаются в виде:

```json
{
  "code": "invalid_request",
  "message": "Invalid request body."
}
```

Примеры кодов:

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

## Краткий API overview

### Public endpoints

| Метод | Путь | Описание |
|---|---|---|
| `GET` | `/health` | health check |
| `POST` | `/auth/signup` | регистрация |
| `POST` | `/auth/login` | логин |
| `GET` | `/market/players` | список активных листингов |

### Protected endpoints

Для всех protected endpoints нужен заголовок:

```http
Authorization: Bearer <access_token>
```

| Метод | Путь | Описание |
|---|---|---|
| `GET` | `/me/team` | получить свою команду |
| `PATCH` | `/me/team` | обновить свою команду |
| `GET` | `/me/players` | список игроков своей команды |
| `GET` | `/me/players/{playerId}` | получить игрока своей команды |
| `PATCH` | `/me/players/{playerId}` | обновить игрока своей команды |
| `POST` | `/market/players/{playerId}/list` | выставить своего игрока на рынок |
| `DELETE` | `/market/players/{playerId}/list` | снять своего игрока с рынка |
| `POST` | `/market/listings/{listingId}/buy` | купить игрока |


## Рекомендуемый happy path для ручной проверки

1. Зарегистрировать пользователя A через `/auth/signup`.
2. Получить его токен.
3. Проверить `/me/team` и `/me/players`.
4. Выставить одного игрока A на рынок.
5. Зарегистрировать пользователя B.
6. Под токеном B получить `/market/players`.
7. Купить листинг через `/market/listings/{listingId}/buy`.
8. Снова проверить `/me/team` и `/me/players` у обеих команд.
9. Убедиться, что после покупки:
   - изменились `budget` у покупателя и продавца;
   - игрок перешел в новую команду;
   - пересчиталось поле `total_players_market_value`.

## Тесты

Запуск всех тестов:

```bash
make test
```

Сейчас проект покрыт unit и handler tests для:

- auth
- me endpoints
- transfer market
- i18n
- config validation

## Что было сознательно упрощено для тестового задания

В проекте есть несколько решений, которые подходят для тестового задания, но в production-сценарии обычно были бы расширены.

### Аутентификация

Сейчас используется только JWT access token (без refresh token flow).

Для реального проекта можно добавить:

- refresh token;
- отзыв токенов;
- rotation refresh token;
- logout / session management;
- хранение и аудит активных сессий.

В рамках тестового задания этого не добавлял, чтобы не усложнять auth flow и не размывать основной scope.

### HTTP contracts и DTO

В проекте DTO используются не везде: в ряде случаев HTTP handlers работают с domain/model структурами напрямую.

Для реального проекта обычно делают более строгий слой DTO:

- отдельные request/response модели;
- явное отделение domain от HTTP contract;
- независимую эволюцию API без привязки к внутренним моделям.

В рамках тестового задания оставил более прямой вариант, чтобы не перегружать проект лишним boilerplate.

### Логирование и observability

Сейчас в проекте есть настраиваемый logger, но нет полного audit/debug logging по всем действиям сервисного слоя.

Для production-сценария обычно добавляют:

- structured logs на ключевых бизнес-операциях;
- correlation/request IDs;
- аудит чувствительных действий;
- метрики и tracing.

Здесь ограничился базовым логированием и аккуратным конфигурированием логгера, потому что основная цель тестового задания — показать API, транзакционную логику и структуру проекта.

### Общий принцип

Во всех этих местах я сознательно выбирал баланс между инженерной аккуратностью и объемом тестового задания:

- реализовать полный рабочий функционал;
- сохранить читаемую архитектуру;
- не уводить проект в production-overengineering там, где это не требуется условиями.

## Структура проекта

```text
cmd/api                 # точка входа HTTP API
docker                  # scripts для Docker-контейнеров
internal/api            # handlers, router, response helpers
internal/api/middleware # auth middleware
internal/auth           # JWT provider и auth context helpers
internal/config         # загрузка и валидация конфига
internal/db             # bun bootstrap и tx manager
internal/domain         # bun/domain модели
internal/i18n           # localizer и embedded catalogs
internal/repository     # доступ к БД
internal/service        # бизнес-логика
migrations              # SQL миграции
postman                 # Postman collection для ручной проверки API
compose.yaml            # Docker Compose для postgres, migrate и api
Dockerfile              # сборка API и контейнера миграций
Makefile                # команды для локальной разработки и Docker flow
```
