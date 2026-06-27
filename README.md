# PayTrack — Система учёта оплат и документов

Мини-система учёта оплат, проектов и закрывающих документов для digital-агентства.

## Стек

| Слой | Технология |
|------|-----------|
| Backend | **Go** + Gin + GORM |
| Frontend | **React** + TypeScript + Vite |
| База данных | **SQLite** (файл, без установки сервера) |
| Деплой | **Docker Compose** |

### Почему Go + React, а не Laravel + Vue?

**Go** — один из основных языков моего стека. Выбор обусловлен несколькими факторами:

1. **Микросервисная готовность.** Go идеально ложится на микросервисную архитектуру. Текущий монолит спроектирован так, чтобы его можно было декомпозировать на независимые сервисы:
   - **projects-service** — управление проектами и клиентами (юрлицами);
   - **payments-service** — оплаты, импорт банковских выписок;
   - **documents-service** — акты и документооборот.

   Это даёт **отказоустойчивость** (падение одного сервиса не роняет систему), **независимое масштабирование** (payments-service под нагрузкой при массовом импорте выписок), **автономные деплои** (обновление документов не затрагивает оплаты). Трейдофф — более сложная отладка (distributed tracing, межсервисная коммуникация).

2. **Единый бинарь.** Go компилируется в один исполняемый файл без runtime-зависимостей. Docker-образ — ~15MB (alpine). Нет JVM, нет интерпретатора, нет node_modules на сервере.

3. **Встроенная конкурентность.** Горутины и каналы позволяют элегантно обрабатывать параллельный импорт операций, retry на блокировках, graceful shutdown — без внешних библиотек для async.

4. **Строгая типизация.** Ошибки ловятся на этапе компиляции, а не в рантайме. Рефакторинг безопасен.

5. **Чистая архитектура из коробки.** Интерфейсы Go (implicit implementation) естественным образом дают инверсию зависимостей. Domain-слой не импортирует GORM, Gin или любой фреймворк — только стандартную библиотеку.

**React + TypeScript** — строгая типизация на фронтенде, FSD-архитектура (Feature-Sliced Design), переиспользуемые ui-компоненты.

**SQLite → PostgreSQL** — для прототипа SQLite убирает необходимость отдельного сервера БД. В продакшене переключение на Postgres — одна env-переменная (`DB_DRIVER=postgres`), миграции те же (GORM AutoMigrate).

---

## Быстрый старт (рекомендуется)

### Вариант 1 — Docker Compose (одна команда)

```bash
git clone <repo>
cd payments-dashboard
docker compose up --build
```

Откройте: **http://localhost:3000**

База данных заполняется автоматически при первом запуске (seed).

---

### Вариант 2 — Локальный запуск

#### Требования
- Go 1.22+
- Node.js 20+
- gcc (для CGO/sqlite3): на macOS — `xcode-select --install`, на Ubuntu — `apt install build-essential`

#### Backend

```bash
cd backend
mkdir -p data
go run ./cmd/server
# Сервер запустится на :8080
# База данных и seed создадутся автоматически в ./data/payments.db
```

#### Frontend

```bash
cd frontend
npm install
npm run dev
# Откройте http://localhost:5173
```

Vite автоматически проксирует `/api/*` → `localhost:8080`.

---

## Архитектура

### Сущности

```
Client (юридическое лицо / плательщик)
  ├── id, name, inn, ogrn, bank_account, contact_person
  └── → Projects (1:N)

Project (проект)
  ├── id, name, client_id, status (active/completed/paused)
  └── → Payments (1:N)

Payment (оплата)
  ├── id, project_id, legal_entity_id
  ├── payment_date, amount, payment_purpose
  ├── service_stage (Разработка/Дизайн/SEO/Реклама/Контент/Сопровождение)
  ├── invoice_number, contract_number
  └── → Act (1:1)

Act (закрывающий документ)
  ├── id, payment_id
  ├── is_sent, sent_at
  ├── is_signed, signed_at
  ├── status (вычисляется, не хранится)
  └── manager_comment
```

### Логика статусов акта

| Условие | Статус |
|---------|--------|
| Акт не существует или не отправлен, оплата < 30 дней | `not_sent` |
| Акт не существует или не отправлен, оплата > 30 дней | `needs_attention` |
| Отправлен, не подписан, < 14 дней с отправки | `waiting_signature` |
| Отправлен, не подписан, > 14 дней с отправки | `needs_attention` |
| Отправлен и подписан | `closed` |

Статус **вычисляется на лету** (в Go при каждом запросе), не хранится в БД.
Это сделано намеренно: логика может меняться без миграций.

### Структура кода

```
backend/
  cmd/server/
    main.go       # точка входа, роутер, БД
    seed.go       # начальные данные
  internal/
    models/       # структуры GORM + DTO + бизнес-логика статусов
    repository/   # все запросы к БД
    handlers/     # HTTP handlers (Gin)

frontend/src/
  api/            # fetch-клиент к API
  types/          # TypeScript-типы + утилиты (форматирование, статусы)
  pages/
    Dashboard.tsx  # сводка + таблица проектов
    ProjectsPage.tsx
    PaymentsPage.tsx  # фильтры + таблица + модалка акта
  components/
    ActModal.tsx   # редактирование статуса акта
```

### API

| Метод | URL | Описание |
|-------|-----|----------|
| GET | `/api/v1/dashboard/summary` | Сводка: суммы, счётчики |
| GET | `/api/v1/clients` | Список юрлиц |
| GET | `/api/v1/projects` | Проекты с агрегатами |
| GET | `/api/v1/payments` | Оплаты с фильтрами |
| GET | `/api/v1/payments/:id` | Оплата по ID |
| PUT | `/api/v1/payments/:id/act` | Создать/обновить акт |

Параметры фильтрации (GET /payments):
- `project_id`, `legal_entity_id`, `act_status`, `service_stage`
- `search` — поиск по назначению платежа и имени клиента
- `date_from`, `date_to`

---

## Деплой на свой сервер (paytrack.mikhailgusev.dev)

### DNS (уже настроено)

| Поддомен | Тип | Значение |
|----------|-----|----------|
| `paytrack` | A | 77.222.43.150 |
| `api.paytrack` | A | 77.222.43.150 |

### Запуск на сервере

```bash
ssh root@77.222.43.150
git clone <repo> /opt/paytrack
cd /opt/paytrack

# Запуск с Caddy (автоматический HTTPS через Let's Encrypt):
docker compose -f docker-compose.yml -f deploy/docker-compose.caddy.yml up -d --build
```

Caddy автоматически:
- получит сертификаты Let's Encrypt для обоих поддоменов
- `paytrack.mikhailgusev.dev` → фронтенд (React SPA)
- `api.paytrack.mikhailgusev.dev` → бэкенд (Go API)

### Проверка

```bash
curl https://api.paytrack.mikhailgusev.dev/health
# {"status":"ok"}

# Фронтенд:
# https://paytrack.mikhailgusev.dev
```

### Обновление

```bash
cd /opt/paytrack
git pull
docker compose -f docker-compose.yml -f deploy/docker-compose.caddy.yml up -d --build
```

### Локальная разработка (без домена)

```bash
docker compose up --build
# Фронт: http://localhost:3000
# API: http://localhost:8080
```

---

## Что не реализовано (намеренно)

- Авторизация/роли
- Загрузка PDF-документов
- Интеграция с банком/1С
- Генерация актов

### Парсинг банковской выписки — реализован write-путь

Endpoint `POST /api/v1/import/bank-statement` принимает уже распарсенные
операции и создаёт оплаты. Ценная часть — не парсинг PDF (он за рамками задания),
а **надёжная запись**:

- **Идемпотентность**: у каждой операции детерминированный `ExternalID`
  (sha256 от даты + суммы + ИНН + назначения). Повторный импорт той же выписки
  **не создаёт дублей** — они считаются `skipped`.
- **Retry на блокировках**: запись обёрнута в транзакцию и повторяется при
  транзиентных конфликтах блокировок (SQLite `database is locked`, Postgres
  `40001`).
- **Матчинг**: ИНН плательщика → клиент (с блокировкой строки), назначение
  платежа → этап услуги. Нераспознанные считаются `unmatched` для ручной разборки.

Проверить:
```bash
curl -X POST http://localhost:8080/api/v1/import/bank-statement \
  -H 'Content-Type: application/json' \
  -d @backend/sample-bank-statement.json
# Повторный вызов вернёт {"imported":0,"skipped":2,"unmatched":1}
```

---

## Конкурентность и блокировки строк

`PUT /payments/:id/act` — единственное место с реальным race condition:
два менеджера одновременно отмечают акт подписанным. Защита:

1. вся операция read-modify-write в одной транзакции;
2. блокировка строки `SELECT … FOR UPDATE` (`clause.Locking`) — на Postgres/MySQL
   реальный lock, на SQLite вырождается в сериализованную запись (корректно);
3. **retry** всей транзакции при транзиентном конфликте блокировки с линейным
   backoff (до 5 попыток).

Покрыто тестом `TestRepository_UpsertAct_Concurrent`: два параллельных upsert,
проверка что создалась **ровно одна** строка акта.

---

## Тесты

Table-driven тесты:

```bash
cd backend
go test ./... -v
```

Покрыто:
- `models/models_test.go` — логика расчёта статуса акта (7 кейсов, включая граничные)
- `repository/repository_test.go` — upsert акта, идемпотентность, **конкурентный доступ**, детектор lock-конфликтов
- `services/import_test.go` — идемпотентность импорта, матчинг этапов, детерминированность ExternalID

---

## API-документация (Swagger)

После запуска бэкенда:

**http://localhost:8080/swagger**

OpenAPI 3 спецификация — `backend/cmd/server/docs/openapi.yaml`, встроена в бинарь
через `go:embed` (не нужно раздавать файлы отдельно).

---

## Адаптив

- Десктоп: фиксированный сайдбар слева.
- Мобильный (<768px): верхний бар с бургер-меню, выезжающий drawer с backdrop,
  фильтры в колонку, карточки сводки в 1–2 столбца, таблицы со скроллом.

---

## Логирование (slog)

Структурированное логирование на стандартном `log/slog` (без внешних зависимостей).
Настройка через env: `LOG_LEVEL` (debug/info/warn/error), `LOG_FORMAT` (text/json).

Логируется: старт/остановка сервера, конфигурация пула, подключение к БД и
каждая попытка ping, каждый HTTP-запрос (метод, путь, статус, длительность, IP),
ошибки хендлеров, retry на блокировках строк, результат импорта выписки.

Пример (LOG_FORMAT=json):
```json
{"time":"...","level":"INFO","msg":"http_request","method":"PUT","path":"/api/v1/payments/1/act","status":200,"duration_ms":4,"client_ip":"::1"}
{"time":"...","level":"WARN","msg":"lock conflict, retrying transaction","attempt":1,"max":5}
```

## Health-check и пул соединений к БД

- `GET /health` — liveness (процесс жив).
- `GET /ready` — readiness: пингует БД и возвращает статистику пула
  (`db_open_conns`, `db_in_use`, `db_idle`). 503 если БД недоступна.
- При старте подключение к БД идёт с **retry/backoff** (до 10 попыток) — БД может
  подниматься позже приложения (docker-compose `depends_on: service_healthy`).
- **Пул соединений** настраивается через env: `DB_MAX_OPEN_CONNS`,
  `DB_MAX_IDLE_CONNS`, `DB_CONN_MAX_LIFETIME_SEC`. Для SQLite по умолчанию
  `MaxOpenConns=1` (SQLite сериализует запись; включён WAL + busy_timeout).
  Для Postgres — пул из 20 соединений.

## Переключение на PostgreSQL

Сборка по умолчанию — только SQLite (ноль лишних зависимостей). Postgres — за
build-флагом, чтобы дефолтная сборка не тянула pgx:

```bash
go get gorm.io/driver/postgres
go build -tags postgres ./cmd/server
DB_DRIVER=postgres DB_DSN="host=... user=... dbname=... sslmode=disable" ./server
```

## Адаптив (точно под задачу, без излишеств)

- **Десктоп**: таблицы оплат и проектов как обычно.
- **Мобильный (<768px)**: каждая запись — **плитка/карточка** (сумма и статус
  акта сверху, далее юрлицо/дата/этап/счёт, назначение и комментарий, кнопка
  «Статус акта»). Никаких горизонтальных скроллов по широкой таблице.
- **Модалка акта**: на телефоне выезжает снизу на всю ширину (bottom sheet).
- Сайдбар → бургер-меню с выезжающим drawer и backdrop.
- Фильтры на телефоне — в колонку, на всю ширину.
- Графиков нет (не требуются).

## Подключение фронта на внешнем хостинге к своему бэку

Фронт читает адрес API из `VITE_API_URL` (см. `frontend/.env.example`).
Бэк ограничивает CORS через `ALLOWED_ORIGINS`. Пошаговые инструкции для
Cloudflare Tunnel и DuckDNS+Caddy (бесплатный HTTPS без покупки домена) —
в `deploy/README.md`.

---

## AI-извлечение операций из выписки (опционально, без токенов)

Задание просит показать, «как мог бы быть устроен парсинг выписки». Реализовано
через слой `internal/ai` с интерфейсом `Extractor` и двумя реализациями:

- **regex** — детерминированный парсер по умолчанию. **Не требует ключей,
  токенов и интернета.** Работает сразу из коробки.
- **LLM** (OpenAI-совместимый) — включается только если задать `AI_PROVIDER`.
  Качественные промпты с JSON-контрактом и few-shot примером лежат в
  `internal/ai/prompts.go`. Поддерживает OpenAI, любой совместимый proxy и Ollama
  (`AI_PROVIDER=ollama`).

Эндпоинт `POST /api/v1/import/bank-statement/raw` принимает сырой текст,
извлекает операции выбранным экстрактором и идемпотентно импортирует. В ответе
поле `extractor` показывает, что сработало.

```bash
# По умолчанию (regex, без ключей):
curl -X POST http://localhost:8080/api/v1/import/bank-statement/raw \
  -H 'Content-Type: application/json' \
  -d @backend/sample-statement-raw.json
# -> {"imported":2,"skipped":0,"unmatched":0,"extractor":"regex","extracted_count":2}

# Чтобы включить LLM (пример, ключи НЕ входят в проект):
#   AI_PROVIDER=openai AI_API_KEY=sk-... AI_MODEL=gpt-4o-mini ./server
# Локально через Ollama (без ключей):
#   AI_PROVIDER=ollama AI_MODEL=llama3.1 ./server   # нужен запущенный ollama
```

Архитектурно LLM-путь отделён от записи: экстрактор только превращает текст в
структуру, а идемпотентный импорт (ExternalID + транзакция + retry) — общий.

---

## HTTP-таймауты

Сервер настроен с `ReadTimeout` / `ReadHeaderTimeout` / `WriteTimeout` /
`IdleTimeout` (env: `HTTP_*_TIMEOUT_SEC`) — защита от медленных/зависших
клиентов (slowloris) и утечки соединений. Плюс graceful shutdown по SIGTERM.

---

## Деплой: docker DNS (основной путь, без Caddy/доменов)

В `docker-compose.yml` фронт и бэк в одной сети `paytrack`. Nginx фронта
проксирует `/api` на `http://backend:8080` по **docker DNS** — публичный домен
для бэка не нужен, фронт обращается к бэку по имени сервиса.

```bash
docker compose up --build   # фронт :3000, бэк :8080, всё связано через docker DNS
```

Раздельный хостинг (фронт на Vercel + бэк на своём сервере) остаётся возможен:
задай `VITE_API_URL` на фронте и `ALLOWED_ORIGINS` на бэке. Готовые конфиги
для бесплатного HTTPS (Cloudflare Tunnel / DuckDNS+Caddy) — в `deploy/`,
но для связки «всё в одном compose» они не нужны.
