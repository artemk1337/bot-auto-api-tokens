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
    "base_url": "http://localhost:11434",
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

```bash
export TELEGRAM_BOT_TOKEN=<telegram-token>
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

## Лицензия

MIT. См. `LICENSE`.
