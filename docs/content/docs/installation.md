---
title: Installation
description: "Install instagram-stats, a single Go program with no runtime dependency, with go install or from its source code."
weight: 10
---

instagram-stats is a single program, written in Go: once built, it needs nothing else to run.

## With go install

With [Go](https://go.dev/dl/) 1.26 or later installed:

```bash
go install github.com/boxboxjason/instagram-stats@latest
```

The `instagram-stats` program is installed in the `bin` folder of your Go path, `~/go/bin` by default: add that folder to your `PATH` to run it from anywhere.

## From the source code

```bash
git clone https://github.com/boxboxjason/instagram-stats.git
cd instagram-stats
go build
```

This builds the `instagram-stats` program in the project folder. `go test ./...` runs the tests.

Then see [Usage]({{< relref "usage" >}}).
