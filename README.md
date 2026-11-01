# Instagram Messages Statistics

Creates statistics from your Instagram conversations: it reads the messages of your Instagram data export (JSON format), and builds a set of graphs and `.txt` rankings (messages, reactions, reels, photos and videos, vocabulary, activity per day, month and hour of the week).

It can analyse one to one conversations, group conversations, and even all your conversations together. Your data stays on your machine.

## Documentation

📖 **<https://boxboxjason.github.io/instagram-stats/>**

The documentation covers [exporting your data from Instagram](https://boxboxjason.github.io/instagram-stats/docs/export/), [usage](https://boxboxjason.github.io/instagram-stats/docs/usage/), every [configuration option](https://boxboxjason.github.io/instagram-stats/docs/configuration/) and the [generated files](https://boxboxjason.github.io/instagram-stats/docs/outputs/). It is built from the [`docs`](./docs) directory of this repository.

## Quick start

1. Download your messages from Instagram, in **JSON** format.
2. Install [Go](https://go.dev/dl/) 1.26 or later, then instagram-stats:

   ```bash
   go install github.com/boxboxjason/instagram-stats@latest
   ```

3. Run it on a conversation folder of the export, or on its `inbox` folder for all your conversations:

   ```bash
   instagram-stats [-config <config.json>] <conversation folder>...
   ```

The graphs and rankings are written to the result folder, `Results` in the current folder by default.

## Development

```bash
go build        # builds the instagram-stats program
go test ./...   # runs the tests
```

The code is organised in packages under `internal`:

| Package | Role |
| ------- | ---- |
| `export` | Types of the JSON files of the Instagram export, repairing the broken encoding of their strings. |
| `config` | The JSON configuration file. |
| `stats` | What each participant sent, and when. |
| `report` | The PNG graphs and the text rankings. |

## Documentation development

The site is built with [Hugo](https://gohugo.io/) (extended edition) and the shared [hugo-theme](https://github.com/boxboxjason/hugo-theme): `hugo server --source docs` serves it locally. Change it in the same pull request as the code it describes. `go test -run TestDocumentation .` (also run by CI) fails when a configuration option or an output file is not documented.
