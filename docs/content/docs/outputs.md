---
title: Outputs
description: "The graphs and text rankings instagram-stats writes for a conversation: message counts, activity over time and words used."
weight: 50
---

Everything is written to the [result folder]({{< relref "configuration#options" >}}). Running instagram-stats again overwrites the files of the previous run.

## Graphs

The graphs are PNG images, named after their title.

| File | Content |
| ---- | ------- |
| `Number of Messages.png` | Pie chart of the number of messages sent by each participant. |
| `Number of Reactions.png` | Pie chart of the number of reactions each participant added to messages. |
| `Number of Reels sent.png` | Pie chart of the number of reels each participant shared. |
| `Number of photos and videos sent.png` | Pie chart of the number of photos and videos each participant sent. |
| `Vocabulary.png` | Pie chart of the number of different words each participant used. |
| `Number of words found.png` | Pie chart of how many times each participant used the [searched words](#words-found). |
| `Activity by time and day.png` | Heatmap of the messages by hour of the day and day of the week: the busiest periods of an average week. |
| `Number of messages per day.png` | Bar graph of the number of messages sent each day. |
| `Number of messages per month.png` | Bar graph of the number of messages sent each month. |

A pie chart whose counts are all 0 (for example no reels at all) is not written, and an error is logged instead.

## Text rankings

Unless `outputTextRanking` is `false`, the per participant counts are also written as rankings, easier to read than a pie chart when a group has many participants:

| File | Content |
| ---- | ------- |
| `Number of messages.txt` | Participants ranked by number of messages. |
| `Number of reactions.txt` | Participants ranked by number of reactions. |
| `Number of reels.txt` | Participants ranked by number of reels shared. |
| `Number of photos and videos.txt` | Participants ranked by number of photos and videos. |
| `Vocabulary richness.txt` | Participants ranked by number of different words. |
| `Number of words found.txt` | Participants ranked by uses of the [searched words](#words-found). |

## Words found

The `Number of words found` graph and ranking are only produced when [`wordsToSearchFor`]({{< relref "configuration#options" >}}) is not empty.
