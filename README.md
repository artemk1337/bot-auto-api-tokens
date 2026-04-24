# bot-auto-api-tokens

Telegram-бот поддержки с локальными Ollama-моделями.

Бот читает настройки из JSON-конфига, принимает сообщения из Telegram через long polling, обрабатывает их в одной очереди и отправляет в Ollama последние `N` сообщений переписки вместе с системным промптом и документацией.

## Возможности

- ответы пользователям через локальную Ollama-модель;
- последовательная обработка входящих сообщений через очередь;
- настраиваемый лимит истории переписки;
- системный промпт из конфига;
- дополнительный контекст из файлов документации;
- ограничение доступа по Telegram user ID;
- нулевые runtime-зависимости кроме Go, Telegram Bot API и Ollama.

## Требования

Для запуска через Docker Compose:

- Docker;
- Docker Compose;
- Telegram bot token.

Для локального запуска без Docker:

- Go 1.25.6;
- запущенная Ollama;
- скачанная модель, например:

```bash
ollama pull llama3.2
```

## Конфигурация

Скопируйте пример:

```bash
cp config.example.json config.json
```

Минимальный конфиг:

```json
{
  "telegram": {
    "token": "${TELEGRAM_BOT_TOKEN}",
    "poll_timeout_seconds": 30,
    "allowed_user_ids": []
  },
  "ollama": {
    "base_url": "${OLLAMA_BASE_URL}",
    "model": "llama3.2",
    "temperature": 0.2,
    "options": {
      "num_ctx": 8192
    }
  },
  "bot": {
    "history_limit": 10,
    "system_prompt": "Ты помощник поддержки. Отвечай кратко и по делу.",
    "documentation_files": [
      "docs/support.example.md"
    ]
  }
}
```

`allowed_user_ids`:

- пустой список разрешает всех пользователей;
- непустой список разрешает только указанные Telegram user ID.

`documentation_files`:

- файлы читаются при старте;
- содержимое добавляется в `system`-сообщение перед запросом к модели;
- если файл не найден, бот не стартует, чтобы не отвечать без ожидаемого контекста.

## Запуск

### Docker Compose

Запуск бота и Ollama:

```bash
export TELEGRAM_BOT_TOKEN=<telegram-token>
docker compose up --build
```

Модель указывается один раз в `config.json`:

```json
{
  "ollama": {
    "model": "qwen2.5:7b"
  }
}
```

`ollama-pull` читает это же поле и скачивает нужную модель перед запуском бота.

Если хотите использовать не `config.example.json`, укажите путь:

```bash
export TELEGRAM_BOT_TOKEN=<telegram-token>
export BOT_CONFIG=./config.json
docker compose up --build
```

Ollama доступна на хосте по адресу:

```text
http://localhost:11434
```

Данные Ollama сохраняются в volume `ollama-data`, поэтому модель не будет скачиваться заново при каждом запуске.

### Локально без Docker

```bash
export TELEGRAM_BOT_TOKEN=<telegram-token>
export OLLAMA_BASE_URL=http://localhost:11434
go run ./cmd/bot -config config.json
```

Размер очереди можно изменить флагом:

```bash
go run ./cmd/bot -config config.json -queue-size 200
```

## Тесты

```bash
go test ./...
```

## Архитектура

- `cmd/bot` - точка входа и wiring зависимостей;
- `internal/config` - загрузка и валидация JSON-конфига;
- `internal/telegram` - минимальный клиент Telegram Bot API;
- `internal/ollama` - клиент Ollama `/api/chat`;
- `internal/bot` - очередь, история переписки, сбор контекста и обработка сообщений.

## Docker

- `Dockerfile` собирает статический бинарник бота.
- `docker-compose.yml` поднимает `ollama`, читает модель из конфига, скачивает ее через `ollama-pull`, затем запускает `bot`.
- `BOT_CONFIG` задает путь к JSON-конфигу, который монтируется в контейнеры как `/app/config.json`.

## Лицензия

MIT. См. `LICENSE`.
