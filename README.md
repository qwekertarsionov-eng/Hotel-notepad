# Guest Accounting — infrastructure

Система учёта гостей: бэкенд на Go + PostgreSQL, разворачивается через Docker Compose.

## Стек

- Go 1.27, pgx, PostgreSQL 15
- Docker Compose

## Быстрый старт

```bash
cp .env.example .env
# при необходимости отредактируйте .env (секреты не коммитить)

docker compose up --build
```

Бэкенд будет доступен на `http://localhost:8080`, PostgreSQL — на `localhost:5432`.

Остановка:

```bash
docker compose down
```

## Переменные окружения

Задаются в `.env` (пример — `.env.example`, не коммитится):

| Переменная | Описание | Пример |
|------------|----------|--------|
| `DATABASE_URL` | Строка подключения к PostgreSQL | `postgres://guest:guest@db:5432/guest_accounting?sslmode=disable` |
| `JWT_SECRET` | Секрет для подписи JWT | `change-me-secret` |
| `HTTP_ADDR` | Адрес HTTP-сервера | `:8080` |

В `compose.yaml` `DATABASE_URL` переопределяется на сервис `db` (`postgres://guest:guest@db:5432/guest_accounting?sslmode=disable`).

## Миграции

SQL-миграции лежат в `migrations/` (`001_users.up.sql`, `002_guests.up.sql` и т.д.).

Бэкенд применяет миграции автоматически при старте: после подключения к БД вызывается `postgres.Up(db)`, который применяет все `*.up.sql` в алфавитном порядке, каждая в своей транзакции. Дополнительно применять миграции вручную не требуется.

Откат (вне контейнера, локально):

```bash
go run ./internal/repository/postgres # (доступен пакет postgres с функцией Down)
```

или прямо через `psql`:

```bash
psql "$DATABASE_URL" -f migrations/002_guests.down.sql
psql "$DATABASE_URL" -f migrations/001_users.down.sql
```

## API эндпоинты

| Метод | Путь | Описание | Auth |
|-------|------|----------|------|
| POST | `/api/v1/auth/register` | Регистрация пользователя | — |
| POST | `/api/v1/auth/login` | Логин, выдача JWT | — |
| GET | `/api/v1/guests` | Список гостей | Bearer JWT |
| POST | `/api/v1/guests` | Создать гостя | Bearer JWT |
| GET | `/api/v1/guests/search?q=` | Поиск гостей | Bearer JWT |
| GET | `/api/v1/guests/{id}` | Гостя по ID | Bearer JWT |
| PUT | `/api/v1/guests/{id}` | Обновить гостя | Bearer JWT |
| DELETE | `/api/v1/guests/{id}` | Удалить гостя | Bearer JWT |

## Тесты

```bash
go test ./...
```

Некоторые интеграционные тесты требуют запущенной PostgreSQL и переменной `JWT_SECRET` — они пропускаются (`t.Skip`), если окружение недоступно.

## Docker

Если Docker доступен:

```bash
docker compose up --build
```

Backend отвечает на `http://localhost:8080`, PostgreSQL принимает подключения на `localhost:5432`, миграции применяются при старте бэкенда.

В этом окружении Docker не найден (`docker` не установлен), поэтому запуск через `docker compose` не проверялся — см. раздел "Риски" ниже.

## Безопасность

- `.env` с реальными секретами не коммитится (см. `.gitignore`). В репозитории только `.env.example`.
- Секреты не логируются и не попадают в тела ответов (тесты `TestTST_SEC_*` проверяют отсутствие `JWT_SECRET` в ответах).

## Риски

- `cmd/server/main.go` мог отсутствовать — создан минимальный entry point (загрузка конфига, подключение к БД, применение миграции `postgres.Up`, HTTP-роутер).
- Docker не доступен в текущем окружении — проверка `docker compose up --build` невозможна, валидация только статическая (`go build`, `go test`).
- `compose.yaml` использует `env_file: - .env` — перед первым запуском нужно скопировать `.env.example` в `.env`.

## Ветки и оформление

Коммерческий код хранится в закрытых GitLab-репозиториях; публичного портфолио нет.
