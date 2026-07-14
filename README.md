# Rate Notifier Backend

Backend для отслеживания курсов валют и пользовательских целей.

## О проекте

Rate Notifier - серверная часть для Flutter Web клиента.
Проект хранит пользователей, курсы валют, цели и уведомления в PostgreSQL.

Приложение работает в двух процессах:

- `api` - HTTP API
- `worker` - фоновое обновление курсов и проверка целей

## Технологии

- Go 1.24
- PostgreSQL 16
- Docker Compose
- GitHub Actions

## Возможности

- регистрация и вход по email и паролю
- JWT-аутентификация
- получение текущего пользователя
- получение курсов USD, EUR и CNY
- CRUD для целей
- внутренние уведомления
- отметка уведомления прочитанным
- фоновое обновление курсов

## Архитектура

Проект сделан как монолит с двумя процессами.

- `api` обрабатывает HTTP-запросы, auth, цели и уведомления
- `worker` получает курсы из внешнего API, сохраняет их в БД и проверяет цели
- оба процесса используют одну базу PostgreSQL
- миграции запускаются автоматически при старте API

## Структура проекта

- `cmd/api/main.go` - точка входа API
- `cmd/worker/main.go` - точка входа worker
- `internal/app` - HTTP-обработчики и запуск приложения
- `internal/auth` - хеширование паролей и JWT
- `internal/config` - загрузка конфигурации из env
- `internal/db` - подключение к PostgreSQL
- `internal/httpx` - JSON-ответы и ошибки
- `internal/logger` - настройка логгера
- `internal/middleware` - auth, CORS, logging, recovery
- `internal/migrations` - SQL-миграции
- `internal/models` - модели данных
- `internal/rateprovider` - клиент внешнего API курсов
- `internal/storage` - SQL-запросы к БД

## Структура БД

### `users`

- `id`
- `email`
- `password_hash`
- `created_at`

### `rates`

- `currency`
- `value`
- `previous_value`
- `updated_at`

### `targets`

- `id`
- `user_id`
- `currency`
- `target_value`
- `condition`
- `is_active`
- `triggered_at`
- `created_at`
- `updated_at`

### `notifications`

- `id`
- `user_id`
- `target_id`
- `currency`
- `target_value`
- `actual_value`
- `condition`
- `is_read`
- `created_at`

## Переменные окружения

Пример в `.env.example`.

Обязательные:

- `DATABASE_URL` - URL подключения к PostgreSQL
- `JWT_SECRET` - секрет для подписи JWT
- `RATE_PROVIDER_URL` - адрес внешнего API курсов
- `RATE_FETCH_INTERVAL` - интервал обновления курсов

Дополнительные:

- `APP_PORT` - порт API
- `JWT_TTL` - время жизни JWT
- `RATE_PROVIDER_TIMEOUT` - timeout запроса к внешнему API
- `CORS_ALLOWED_ORIGIN` - разрешённый origin для CORS

## Запуск через Docker

```bash
cp .env.example .env
docker compose up --build
```

API будет доступно по адресу:

- `http://localhost:8080`

## Ручная проверка

Проверка health:

```bash
curl http://localhost:8080/health
```

## Тесты

```bash
go test ./...
```

## CI

GitHub Actions запускает:

- `go vet ./...`
- `go test ./...`
- `go build ./cmd/api`
- `go build ./cmd/worker`
- `docker build -t rate-notifier .`
