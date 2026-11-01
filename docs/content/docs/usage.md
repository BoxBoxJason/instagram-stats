---
title: Usage
description: "Run instagram-stats on an Instagram conversation folder, analyse several conversations together, and read its logs."
weight: 30
---

Run instagram-stats on the folder of the conversation to analyse:

```bash
instagram-stats [flags] <source>...
```

A source is one of:

- a conversation folder of the [export]({{< relref "export" >}}), such as `messages/inbox/alice_123456789`;
- a folder of conversations, such as `messages/inbox`: every conversation in it is analysed;
- a `message_<n>.json` file.

The `.json` files of a folder are read in its subfolders too. The other files, like the photos and videos of the export, and the `.json` files that are not conversations, are skipped.

| Flag | Description |
| ---- | ----------- |
| `-config <file>` | Path to a `.json` [configuration file]({{< relref "configuration" >}}). Without it, the default configuration is used. |
| `-verbose` | Log every file read and written. |
| `-log-file <file>` | Also write the logs to this file, after its current content, as JSON lines. |

For example:

```bash
instagram-stats ~/instagram-export/your_instagram_activity/messages/inbox/alice_123456789
```

The graphs and rankings are written to the result folder, `Results` in the current folder by default (see [Configuration]({{< relref "configuration#options" >}})). It is created when missing.

## Analysing several conversations together

Several sources are analysed together, as a single conversation. Give them one after the other:

```bash
instagram-stats inbox/alice_123456789 inbox/friends_group_987654321
```

Or give the folder holding them to analyse all your conversations at once:

```bash
instagram-stats ~/instagram-export/your_instagram_activity/messages/inbox
```

## Logs

Progress is logged to the console (standard error). Files that cannot be read are skipped with a warning in the logs. Add `-log-file` to also keep the logs in a file, one JSON object per line, easy to process with tools like `jq`:

```json
{"level":"info","ts":"2026-11-01T23:37:34.042+0100","msg":"read conversations","files":109,"messages":121110,"participants":118}
```
