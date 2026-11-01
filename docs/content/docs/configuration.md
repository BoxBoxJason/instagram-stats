---
title: Configuration
description: "Every option of the instagram-stats JSON configuration file, with its default value and an example configuration."
weight: 40
---

The configuration is a JSON file, given with the `-config` [flag]({{< relref "usage" >}}). Without it, every option has its default value. Every option of the file is optional, and an unknown option is an error, to catch typos.

## Options

| Option | Type | Default | Description |
| ------ | ---- | ------- | ----------- |
| `resultDir` | string | `Results` | Folder the graphs and rankings are written to. A relative path is relative to the current folder. It is created when missing. |
| `wordsToSearchFor` | list of strings | `[]` | Words or phrases to count in the messages. When the list is not empty, a [words found]({{< relref "outputs#words-found" >}}) graph and ranking are added. |
| `outputTextRanking` | boolean | `true` | Also write the `.txt` [rankings]({{< relref "outputs#text-rankings" >}}), easier to read than pie charts in large group conversations. |

{{< callout type="info" >}}
`wordsToSearchFor` is matched anywhere in the text of each message, regardless of case: `lol` also counts `LOL` and `lollipop`. A message counts once per word of the list it contains.
{{< /callout >}}

## Example

```json
{
  "resultDir": "/home/me/instagram-results",
  "wordsToSearchFor": ["lol", "see you"],
  "outputTextRanking": true
}
```
