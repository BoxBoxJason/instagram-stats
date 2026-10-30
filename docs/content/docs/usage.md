---
title: Usage
weight: 30
---

Run the project directory with Python, giving it the folder of the conversation to analyse:

```bash
python3 instagram-stats <source folder> [<config file>]
```

| Argument | Mandatory | Description |
| -------- | --------- | ----------- |
| `<source folder>` | Yes | Folder holding the `.json` message files to analyse. The files must be directly in that folder. |
| `<config file>` | No | Path to a `.json` [configuration file]({{< relref "configuration" >}}). Defaults to `Config/config.json` in the project. When the file does not exist, cannot be read or does not end with `.json`, the default configuration is used. |

For example:

```bash
python3 instagram-stats ~/instagram-export/messages/inbox/alice_123456789
```

The graphs and rankings are written to the result folder, `Results` in the project by default (see [Configuration]({{< relref "configuration#options" >}})).

{{< callout type="warning" >}}
The result folder must already exist and be writable: instagram-stats does not create it, and stops with an error when it is missing.
{{< /callout >}}

## Analysing several conversations together

instagram-stats reads every `.json` file directly in the source folder, so it can mix conversations: copy their message files into a single folder. As every conversation names its files `message_1.json`, `message_2.json`, ..., rename them on the way so they do not overwrite each other:

```bash
mkdir mixed
cp inbox/alice_123456789/message_1.json mixed/alice_1.json
cp inbox/friends_group_987654321/message_1.json mixed/friends_1.json
python3 instagram-stats mixed
```

Files that are not `.json`, or that cannot be read, are skipped with an error in the logs.

## Logs

Progress is printed to the console, and also written to `logging.log` in the project directory.
