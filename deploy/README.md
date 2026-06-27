# Деплой: paytrack.mikhailgusev.dev

## Архитектура деплоя

```
                    77.222.43.150
                         |
                    ┌────┴────┐
                    │  Caddy  │  (:80, :443)
                    └────┬────┘
                         │
          ┌──────────────┼──────────────┐
          │                             │
paytrack.mikhailgusev.dev    api.paytrack.mikhailgusev.dev
          │                             │
    ┌─────┴─────┐               ┌───────┴───────┐
    │  Frontend │               │    Backend    │
    │  (nginx)  │               │   (Go/Gin)   │
    │   :80     │               │    :8080     │
    └───────────┘               └───────────────┘
```

Caddy автоматически получает и продлевает Let's Encrypt сертификаты.

---

## DNS (SpaceWeb)

| Поддомен | Тип | Значение |
|----------|-----|----------|
| `paytrack` | A | 77.222.43.150 |
| `api.paytrack` | A | 77.222.43.150 |

---

## Запуск на сервере

```bash
ssh root@77.222.43.150
git clone <repo-url> /opt/paytrack
cd /opt/paytrack

# Запуск (Caddy + Backend + Frontend):
docker compose -f docker-compose.yml -f deploy/docker-compose.caddy.yml up -d --build
```

Caddy пробует получить сертификат при первом запросе. Убедитесь, что:
- Порты 80 и 443 открыты на файрволе
- DNS записи уже применились (обычно ~5 минут)

---

## Проверка

```bash
# API health:
curl https://api.paytrack.mikhailgusev.dev/health
# {"status":"ok"}

# API readiness (с подключением к БД):
curl https://api.paytrack.mikhailgusev.dev/ready
# {"status":"ready","db_open_conns":1,"db_in_use":0,"db_idle":1}

# Фронтенд:
# https://paytrack.mikhailgusev.dev
```

---

## Обновление

```bash
cd /opt/paytrack
git pull
docker compose -f docker-compose.yml -f deploy/docker-compose.caddy.yml up -d --build
```

---

## Сброс базы данных (пересоздание seed)

```bash
docker compose down
docker volume rm paytrack_sqlite-data
docker compose -f docker-compose.yml -f deploy/docker-compose.caddy.yml up -d --build
```

---

## Локальная разработка (без HTTPS/домена)

```bash
docker compose up --build
# Фронт: http://localhost:3000 (nginx проксирует /api → backend:8080)
# API: http://localhost:8080
```

Или без Docker:
```bash
# Backend:
cd backend && go run ./cmd/server

# Frontend:
cd frontend && npm install && npm run dev
# http://localhost:5173 (Vite проксирует /api → :8080)
```

---

## Переключение на PostgreSQL

```bash
# docker-compose.yml:
#   DB_DRIVER: "postgres"
#   DB_DSN: "host=db user=paytrack password=... dbname=paytrack sslmode=disable"
```

Код поддерживает оба драйвера — просто меняется env-переменная.

---

## Настройки окружения (backend)

| Переменная | По умолчанию | Описание |
|-----------|-------------|----------|
| `PORT` | 8080 | Порт HTTP-сервера |
| `DB_DRIVER` | sqlite | `sqlite` или `postgres` |
| `DB_DSN` | ./data/payments.db | Путь к БД или DSN Postgres |
| `ALLOWED_ORIGINS` | http://localhost:5173,... | CORS origins (через запятую) |
| `LOG_LEVEL` | info | debug/info/warn/error |
| `LOG_FORMAT` | text | text или json |
| `AI_PROVIDER` | (пусто) | openai / ollama / (пусто = regex) |
| `AI_API_KEY` | (пусто) | API ключ для LLM |
| `AI_MODEL` | gpt-4o-mini | Модель LLM |
