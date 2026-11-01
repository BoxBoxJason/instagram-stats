---
title: Outputs
description: "The graphs and text rankings instagram-stats writes for a conversation: message counts, activity over time and words used."
weight: 50
---

Everything is written to the [result folder]({{< relref "configuration#options" >}}). Running instagram-stats again overwrites the files of the previous run.

## Graphs

The graphs are 1600×900 PNG images, on a dark background, named after their title.

| File | Content |
| ---- | ------- |
| `Number of Messages.png` | Pie chart of the number of messages sent by each participant. |
| `Number of Reactions.png` | Pie chart of the number of reactions each participant added to messages. |
| `Number of Reels sent.png` | Pie chart of the number of reels, posts and profiles each participant shared. |
| `Number of photos and videos sent.png` | Pie chart of the number of messages with photos or videos each participant sent. |
| `Vocabulary.png` | Pie chart of the number of different words each participant used. Its total is the number of different words of the whole conversation. |
| `Number of words found.png` | Pie chart of how many of the messages of each participant contain the [searched words](#words-found), once per word. |
| `Activity by time and day.png` | Heatmap of the average number of messages sent on each day of the week, at each hour: the busiest periods of an average week. |
| `Number of messages per day.png` | Bar graph of the number of messages sent each day, from the first message to the last one. |
| `Number of messages per month.png` | Bar graph of the number of messages sent each month. |

The slices of a pie chart go from the biggest to the smallest, and its legend gives the count and share of each participant. A participant keeps the same color in every pie chart. Past 20 participants, the smallest counts are grouped in a single "others" slice: the [text rankings](#text-rankings) give every participant.

A pie chart whose counts are all 0 (for example no reels at all) is not written, and a message is logged instead.

{{< callout type="info" >}}
The graphs draw the names of the participants without their emoji, which their font cannot draw. Letters styled with symbols, like 𝕹𝖆𝖒𝖊, are drawn as plain letters. The text rankings keep the names as they are.
{{< /callout >}}

## Text rankings

Unless `outputTextRanking` is `false`, the per participant counts are also written as rankings, easier to read than a pie chart when a group has many participants:

| File | Content |
| ---- | ------- |
| `Number of messages.txt` | Participants ranked by number of messages. |
| `Number of reactions.txt` | Participants ranked by number of reactions. |
| `Number of reels.txt` | Participants ranked by number of reels shared. |
| `Number of photos and videos.txt` | Participants ranked by number of photos and videos. |
| `Vocabulary richness.txt` | Participants ranked by number of different words. |
| `Number of words found.txt` | Participants ranked by messages containing the [searched words](#words-found), once per word. |

Each line is a position of the ranking, with the participants sharing it:

```text
Number of messages ranking (Total 1520):
1: 812 messages: Alice
2: 354 messages: Bob, Carol
```

## What is counted

- **Messages** are the messages that are not a share, a photo or a video: text, voice messages, calls, stickers...
- **Reels** are the shared reels, posts and profiles. **Photos and videos** count the messages holding photos or videos, once per message.
- The **vocabulary** is made of the words of the text messages, regardless of case. A word is a run of letters, digits and apostrophes. The texts Instagram writes itself, such as "Alice sent an attachment." or a call log, are left out.
- The **activity** graphs count every kind of message, dated in the time zone of the computer running instagram-stats.
- The "Liked a message" notifications are not counted at all.

## Words found

The `Number of words found` graph and ranking are only produced when [`wordsToSearchFor`]({{< relref "configuration#options" >}}) is not empty. A message counts once for each searched word it contains, however many times it contains it.
