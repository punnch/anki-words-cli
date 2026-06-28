# Anki Words TUI

Terminal app for turning English vocabulary into Anki notes with OpenAI.

The flow is:

1. Start Anki Desktop with AnkiConnect enabled.
2. Open the terminal UI.
3. Enter comma-separated words.
4. The app asks OpenAI for structured card data.
5. The generated JSON is validated locally.
6. Valid notes are inserted into Anki through AnkiConnect.

This project is now local-only. There is no Telegram integration.

## Requirements

- Anki Desktop running on the host machine
- AnkiConnect Anki add-on installed and enabled
- OpenAI API key
- Go 1.23+ for local development and `make aw-run`
- Docker with Docker Compose only if running the app container

## AnkiConnect

Recommended AnkiConnect config:

```json
{
  "apiKey": null,
  "apiLogPath": null,
  "webBindAddress": "127.0.0.1",
  "webBindPort": 8765,
  "webCorsOriginList": ["http://localhost"],
  "ignoreOriginList": []
}
```

Restart Anki after changing the add-on config.

Check AnkiConnect from the host:

```bash
curl -X POST http://127.0.0.1:8765 \
  -H 'Content-Type: application/json' \
  -d '{"action":"version","version":6}'
```

Expected response:

```json
{"result": 6, "error": null}
```

## Configuration

Copy the example file and fill in secrets:

```bash
cp .env.example .env
```

Important variables:

| Variable | Required | Default | Description |
| --- | --- | --- | --- |
| `OPENAI_API_KEY` | yes | none | OpenAI API key used to generate cards. |
| `OPENAI_BASE_URL` | no | `https://api.openai.com/v1` | OpenAI-compatible API base URL. |
| `OPENAI_MODEL` | no | `gpt-5.4-mini` | Model used for card generation. |
| `ANKICONNECT_URL` | yes | none | Local AnkiConnect endpoint. Use `http://127.0.0.1:8765`. |
| `SETTINGS_FILE` | no | `./out/settings/settings.json` | JSON file used to persist local preferences. |
| `SENTENCE_HIGHLIGHT_COLOR` | no | `#00557f` | HTML color used for the target word highlight. |
| `LOGGER_LEVEL` | no | `DEBUG` | Zap logger level. |
| `LOGGER_FOLDER` | no | `./out/logs` | Log directory. |

The app fails fast on startup if required settings are missing.

## Run Locally

Start Anki Desktop, then open the TUI:

```bash
make aw-run
```

## Run In Docker

The TUI can also run in Docker, but it must run in the foreground because it is interactive:

```bash
make aw-deploy
```

The app container uses host networking so it can reach local AnkiConnect at `127.0.0.1:8765`.

## TUI Menu

The terminal UI supports:

```text
1. Create cards
2. List decks
3. Change deck
4. List models
5. Change model
6. Show generation preset
7. Edit generation preset
8. Quit
```

Preferences are stored in the local settings file and reused on the next run:

- active deck
- Anki note model
- generation preset

On first run, the app initializes preferences from the first deck and first note model returned by AnkiConnect.
You can change both from the TUI afterward.

## Anki Note Model

The configured Anki note model must exist in Anki.
The app reads the selected model's fields from Anki and asks OpenAI to fill exactly those fields.
This supports simple models such as `Front`/`Back` and richer vocabulary models such as `Word`/`Sentence`/`Translation`.

If the selected model has a `Sentence` field, the generated sentence must contain the exact target word highlight:

```html
<span style="color:#00557f;"><b>WORD</b></span>
```

## Data

Preferences are stored as JSON:

```json
{
  "users": {
    "1": {
      "user_id": "1",
      "active_deck": "Default",
      "generation_template": "...",
      "preferred_model": "Basic"
    }
  }
}
```

This app uses a single local profile ID internally.

Preferences live under:

```text
out/settings/settings.json
```

Logs live under:

```text
out/logs
```

## Development

Run tests:

```bash
make test
```

Format Go code:

```bash
make fmt
```

Run vet:

```bash
make vet
```

Build the Docker image:

```bash
docker compose build anki-words
```

## Project Layout

- `cmd/ankiwords/main.go` - application entrypoint
- `internal/app` - terminal UI and workflow orchestration
- `internal/anki` - AnkiConnect HTTP client
- `internal/config` - environment loading and validation
- `internal/model` - shared data structures and sentence formatting
- `internal/openai` - prompt building and OpenAI API client
- `internal/repository` - local settings persistence
- `internal/validation` - generated card validation

## Operational Notes

- Keep `.env` private. It contains your OpenAI API key.
- Back up `out/settings/settings.json` if the saved preferences matter.
- Confirm Anki is running before listing decks, selecting a deck, or creating cards.
- Confirm the configured OpenAI model is available for your API key before creating cards.
