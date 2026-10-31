---
title: Configuration
description: "Every option of the instagram-stats JSON configuration file, with its default value and an example configuration."
weight: 40
---

The configuration is a JSON file, `Config/config.json` in the project unless [another one is given]({{< relref "usage" >}}) on the command line. Every option is optional.

## Options

| Option | Type | Default | Description |
| ------ | ---- | ------- | ----------- |
| `resultDir` | string | `Results` in the project | Folder the graphs and rankings are written to. It must already exist and be writable. |
| `wordsToSearchFor` | list of strings | `[]` | Words or phrases to count in the messages. When the list is not empty, a [words found]({{< relref "outputs#words-found" >}}) graph and ranking are added. |
| `outputTextRanking` | boolean | `true` | Also write the `.txt` [rankings]({{< relref "outputs#text-rankings" >}}), easier to read than pie charts in large group conversations. |

{{< callout type="info" >}}
`wordsToSearchFor` is matched against the lowercased text of each message, anywhere in it: write the words in lowercase. A message counts once per word of the list it contains.
{{< /callout >}}

## Example

```json
{
  "resultDir": "/home/me/instagram-results",
  "wordsToSearchFor": ["lol", "see you"],
  "outputTextRanking": true
}
```
