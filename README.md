# bot-auto-api-tokens

Telegram-бот поддержки с локальными Ollama-моделями.

Бот читает настройки из JSON-конфига, принимает сообщения из Telegram через long polling, обрабатывает их в одной очереди и отправляет в Ollama последние `N` сообщений переписки вместе с системным промптом и документацией.

## Возможности

- ответы пользователям через локальную Ollama-модель;
- последовательная обработка входящих сообщений через очередь;
- настраиваемый лимит истории переписки;
- системный промпт из конфига;
- дополнительный контекст из файлов документации;
- web search через Ollama Web Search API;
- настройка thinking level для моделей GPT-OSS;
- ограничение доступа по Telegram user ID;
- нулевые runtime-зависимости кроме Go, Telegram Bot API, Ollama и опционального Ollama API key для web search.

## Требования

Для запуска через Docker Compose:

- Docker;
- Docker Compose;
- Telegram bot token.
- Ollama API key, если включен web search.

Для локального запуска без Docker:

- Go 1.25.6;
- запущенная Ollama;
- скачанная модель, например:

```bash
ollama pull gpt-oss:20b
```

## Конфигурация

Скопируйте `.env`:

```bash
cp .env.example .env
```

В `.env` задается Telegram-токен, активный конфиг модели и размер очереди:

```dotenv
TELEGRAM_BOT_TOKEN=<telegram-token>
OLLAMA_API_KEY=<ollama-api-key>
BOT_CONFIG=./configs/gpt-oss-20b.json
QUEUE_SIZE=100
```

Готовые конфиги моделей лежат в `configs/`:

- `configs/gpt-oss-20b.json` - дефолт: `gpt-oss:20b`, `think: low`, web search включен;
- `configs/llama3.2.json`
- `configs/qwen2.5-7b.json`
- `configs/mistral.json`

Чтобы сменить модель, поменяйте только `BOT_CONFIG` в `.env`:

```dotenv
BOT_CONFIG=./configs/qwen2.5-7b.json
```

Пример конфига модели:

```json
{
  "telegram": {
    "token": "${TELEGRAM_BOT_TOKEN}",
    "poll_timeout_seconds": 30,
    "allowed_user_ids": []
  },
  "ollama": {
    "base_url": "${OLLAMA_BASE_URL}",
    "model": "gpt-oss:20b",
    "think": "low",
    "temperature": 0.2,
    "options": {
      "num_ctx": 8192
    },
    "web_search": {
      "enabled": true,
      "api_key": "${OLLAMA_API_KEY}",
      "base_url": "https://ollama.com/api",
      "max_results": 5
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

`ollama.web_search`:

- при `enabled: true` бот перед запросом к модели делает web search по вопросу пользователя;
- результаты добавляются отдельным `system`-сообщением в контекст;
- нужен `OLLAMA_API_KEY`.

## Запуск

### Docker Compose

Запуск бота и Ollama:

```bash
cp .env.example .env
# отредактируйте .env
docker compose up --build
```

`ollama-pull` читает `ollama.model` из выбранного JSON-конфига и скачивает нужную модель перед запуском бота.

Ollama доступна на хосте по адресу:

```text
http://localhost:11434
```

Данные Ollama сохраняются в volume `ollama-data`, поэтому модель не будет скачиваться заново при каждом запуске.

### Локально без Docker

```bash
set -a
. ./.env
set +a
export OLLAMA_BASE_URL=http://localhost:11434
go run ./cmd/bot -config "$BOT_CONFIG"
```

Размер очереди можно изменить флагом:

```bash
go run ./cmd/bot -config "$BOT_CONFIG" -queue-size 200
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
- `.env` задает `BOT_CONFIG`, `TELEGRAM_BOT_TOKEN`, `OLLAMA_API_KEY` и `QUEUE_SIZE`.
- `BOT_CONFIG` задает путь к JSON-конфигу из `configs/`, который монтируется в контейнеры как `/app/config.json`.

## Лицензия

MIT. См. `LICENSE`.
