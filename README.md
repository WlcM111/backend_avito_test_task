# PR Reviewer Assignment Service

Сервис для автоматического назначения ревьюверов на Pull Request'ы (PR) внутри команды.  

---

## 1. Функционал:

Внутри есть три основные сущности:

- **User**
  - `user_id` (строка, уникальный идентификатор)
  - `username` (имя пользователя)
  - `team_name` (название команды)
  - `is_active` (флаг активности; неактивные не назначаются ревьюверами)

- **Team**
  - `team_name` (уникальное имя команды)
  - `members` — список пользователей (`User`)

- **PullRequest**
  - `pull_request_id`
  - `pull_request_name`
  - `author_id`
  - `status` ∈ {`OPEN`, `MERGED`}
  - `assigned_reviewers` — массив `user_id` (0..2 ревьюверов)

### Основная бизнес-логика:

1. При создании PR:
   - Автоматически назначаются **до двух** активных ревьюверов из **команды автора**, исключая самого автора.
   - Если доступных кандидатов < 2 — назначается доступное количество (0 или 1).
   - Пользователи с `is_active = false` **не назначаются**.

2. Переназначение ревьювера:
   - Заменяет конкретного ревьювера на случайного активного участника **из его команды**.
   - Уже назначенные на этот PR ревьюверы не могут быть переназначены повторно в этот же PR (без дублей).
   - Если кандидатов нет — возвращается ошибка `NO_CANDIDATE`.

3. После `MERGED`:
   - менять список ревьюверов **нельзя** (`PR_MERGED`).

4. Merge (`/pullRequest/merge`):
   - Идемпотентен:
     - первый вызов переводит PR в `MERGED` и записывает `mergedAt`,
     - повторные вызовы просто возвращают актуальное состояние без ошибки.

5. Статистика:
   - `/stats/assignments` возвращает количество назначений по ревьюверам.

---

## 2. Тех. стек

- **Go** (см. версию в `go.mod`)
- **PostgreSQL** (используется как основная БД)
- **Docker** + **docker-compose** — запуск сервиса и базы
- **chi** — HTTP-роутер
- **log/slog** — логирование
- **SQL-миграции** (`internal/storage/migrations.go` + `migrations/*.sql`)
- **E2E-тесты** — `test/e2e/e2e_test.go`

---

## 3. Структура проекта

```text
.
├── cmd/
│   └── server/
│       └── main.go            # точка входа, сборка и запуск сервиса
├── internal/
│   ├── config/                # конфиг (HTTP, DB, ENV)
│   ├── domain/                # доменные модели, ошибки, интерфейсы репозиториев
│   ├── logging/               # инициализация slog-логгера
│   ├── random/                # источник случайности (для выбора ревьюверов)
│   ├── storage/               # запуск SQL-миграций
│   ├── server/                # обёртка над http.Server (start/shutdown)
│   ├── service/               # бизнес-логика (Team, User, PullRequest, Stats)
│   ├── http/                  # HTTP-слой: роутер, хендлеры, DTO, middleware
│   └── repository/
│       └── postgres/          # реализация репозиториев на PostgreSQL
├── migrations/
│   └── 001_init.sql           # создание таблиц teams, users, pull_requests, pr_reviewers
├── test/
│   └── e2e/
│       └── e2e_test.go        # E2E-тесты, гоняющие API end-to-end
├── Dockerfile                 # сборка Docker-образа сервиса
├── docker-compose.yml         # Postgres + сервис (app)
├── openapi.yaml               # спецификация API
├── Makefile                   # удобные команды для сборки/запуска/тестов
└── README.md                  # этот файл
```

# Вариант 1 — запуск через Docker + docker-compose

## 1.1 Клонировать репозиторий

```bash       
git clone <URL_ВАШЕГО_РЕПОЗИТОРИЯ> pr-reviewer-service
cd pr-reviewer-service
```

## 1.2 Собрать и запустить сервис и PostgreSQL
```bash
docker-compose up --build
```

## 1.3 Проверить, что сервис работает
```bash
curl http://localhost:8080/health
```


# Вариант 2 — запуск локально (Go + PostgreSQL в Docker)

## 2.1 Клонировать репозиторий
```bash
git clone <URL_ВАШЕГО_РЕПОЗИТОРИЯ> pr-reviewer-service
cd pr-reviewer-service
```

## 2.2 Поднять PostgreSQL в отдельном контейнере
```bash
docker run -d \
  --name pr-reviewer-postgres-local \
  -e POSTGRES_USER=postgres \
  -e POSTGRES_PASSWORD=postgres \
  -e POSTGRES_DB=postgres \
  -p 5432:5432 \
  postgres:16-alpine
```

## 2.3 Настроить переменные окружения для сервиса
```bash
export DB_DSN="postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable"
export HTTP_PORT="8080"
export ENV="dev"
```

## 2.4 Запустить сервис локально
```bash
go run ./cmd/server
```

## 2.5 Проверить, что сервис работает
```bash
curl http://localhost:8080/health
```

## 2.6 Запустить все тесты (включая e2e)
```bash
go test ./...
```

## 2.7 Запустить только e2e-тесты

```bash
go test ./test/e2e -v      
```

## Линтер

Используется `golangci-lint` с набором линтеров:

- `govet`, `staticcheck` — поиск потенциальных багов
- `errcheck` — контроль обработки ошибок
- `gofmt`, `goimports` — форматирование и импорты
- `revive` — стиль и нейминг
- `gocyclo` — контроль сложности функций
- `dupl` — поиск дублирующегося кода

## Запуск Линтера:
```bash
golangci-lint run ./...
```

## Нагрузочное тестирование

Инструмент: [k6](https://k6.io)

Сценарий:
- 10 виртуальных пользователей
- 30 секунд
- Операции:
  - `/health`
  - `/team/add`
  - `/pullRequest/create`
  - `/pullRequest/merge`

Команда:
```bash
k6 run load/k6_pr_service.js
```

## Swagger / OpenAPI

Для сервиса описана OpenAPI-схема в файле `openapi.yml`.  
Документация доступна через Swagger UI в отдельном контейнере на порту 8081 на localhost.

### Как запустить

Требуется установленный Docker и docker compose.

```bash
docker compose up --build
# или
docker-compose up --build
```