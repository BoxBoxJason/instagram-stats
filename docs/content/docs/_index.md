---
title: Documentation
description: "Documentation for instagram-stats, a Go program that builds graphs and text rankings from your Instagram conversations."
weight: 1
---

instagram-stats creates statistics from your Instagram conversations. It reads the messages of your Instagram data export (JSON format), and builds a set of graphs and `.txt` rankings.

It works on one to one conversations, group conversations, and even several conversations mixed together.

{{< callout type="info" >}}
Your data stays on your machine: instagram-stats only reads local files and writes its results to a local folder. It never connects to Instagram or to any other service.
{{< /callout >}}

{{< cards >}}
  {{< card link="installation" title="Installation" subtitle="Install it with Go, or build it from source." >}}
  {{< card link="export" title="Export your data" subtitle="Download your messages from Instagram in JSON format." >}}
  {{< card link="usage" title="Usage" subtitle="Run it on a conversation, or several." >}}
  {{< card link="configuration" title="Configuration" subtitle="Choose the output folder, words to search for, text rankings." >}}
  {{< card link="outputs" title="Outputs" subtitle="Every graph and ranking file it produces." >}}
{{< /cards >}}
