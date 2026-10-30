# Instagram Messages Statistics

Creates statistics from your Instagram conversations: it reads the messages of your Instagram data export (JSON format), and builds a set of graphs and `.txt` rankings (messages, reactions, reels, photos and videos, vocabulary, activity per day, month and hour of the week).

It can analyse one to one conversations, group conversations, and even several conversations mixed together. Your data stays on your machine.

## Documentation

📖 **<https://boxboxjason.github.io/instagram-stats/>**

The documentation covers [exporting your data from Instagram](https://boxboxjason.github.io/instagram-stats/docs/export/), [usage](https://boxboxjason.github.io/instagram-stats/docs/usage/), every [configuration option](https://boxboxjason.github.io/instagram-stats/docs/configuration/) and the [generated files](https://boxboxjason.github.io/instagram-stats/docs/outputs/). It is built from the [`docs`](./docs) directory of this repository.

## Quick start

1. Download your messages from Instagram, in **JSON** format.
2. Install Python 3 and matplotlib (`pip install matplotlib`), and clone this repository.
3. Create the result folder, `Results` in the project by default, then run:

   ```bash
   python3 instagram-stats <conversation folder> [<config.json>]
   ```

The graphs and rankings are written to the result folder.

## Documentation development

The site is built with [Hugo](https://gohugo.io/) (extended edition) and the shared [hugo-theme](https://github.com/boxboxjason/hugo-theme): `hugo server --source docs` serves it locally. Change it in the same pull request as the code it describes. `python3 -I docs/check_docs.py` (also run by CI) fails when a configuration option or an output file is not documented.
