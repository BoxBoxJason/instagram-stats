---
title: Export your data
weight: 20
---

instagram-stats works on the export Instagram gives you of your own data. Request it from the Instagram app or website, in **Accounts Center → Your information and permissions → Download your information**:

1. Choose to download some of your information, and select **Messages**.
2. Choose **Download to device**, and set the format to **JSON** (not HTML).
3. Pick the date range you want statistics for.
4. Wait for Instagram to tell you the export is ready, then download and extract the archive.

{{< callout type="info" >}}
The names of these menus change from time to time on Instagram's side. What matters is a **JSON** export that includes your **messages**.
{{< /callout >}}

## Where the conversations are

Each conversation is a folder of the export, under a `messages/inbox` directory (its exact location in the archive depends on when the export was made). A conversation folder holds one or more `message_<n>.json` files, one per batch of messages:

```text
messages/
  inbox/
    alice_123456789/
      message_1.json
      message_2.json
    friends_group_987654321/
      message_1.json
```

The folder of the conversation you want statistics for is the one you give to instagram-stats, see [Usage]({{< relref "usage" >}}).
