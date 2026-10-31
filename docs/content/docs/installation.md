---
title: Installation
description: "Install instagram-stats: Python 3, its only dependency matplotlib, and how to get the project code on your machine."
weight: 10
---

instagram-stats is a Python program. It needs:

- Python 3
- [matplotlib](https://matplotlib.org/), the only third party dependency

Install matplotlib, preferably in a virtual environment:

```bash
python3 -m venv .venv
. .venv/bin/activate
pip install matplotlib
```

Then get the code:

```bash
git clone https://github.com/boxboxjason/instagram-stats.git
```

The program is the `instagram-stats` directory itself: Python runs it through its `__main__.py` file, see [Usage]({{< relref "usage" >}}).
