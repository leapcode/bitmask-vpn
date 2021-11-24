MOTD (message of the day)
=========================

This is a stub until a more sophisticated motd mechanism can be implemented in
the future, with better platform integration.

Providers can opt-in to the motd feature (only riseup is using it at the moment).

If motd is enabled for a given provider, the client will attempt to fetch
the motd.json file from a well-known URL, and will display the first valid
message on the splash screen.

The structure of the `motd.json` file is like follows:

```
{
    "motd": [{
        "begin":    "01 Nov 21 00:00 -0700",
        "end":      "31 Jan 22 00:00 -0700",
        "type":     "daily",
        "platform": "all",
        "urgency":  "normal",
        "text": [
          { "lang": "en",
            "str": "This is a <a href='https://leap.se'>test!</a>"},
          { "lang": "es",
            "str": "Esto es una <a href='https://leap.se'>pruebita!</a>"}
        ]}
    ]
}
```

Valid values are: 

* Begin, End are date strings, in the format: "01 Jan 21 00:00:00 -0700".
* Type: "once" for a one-shot message, "daily" for a message that is displayed daily during the specified duration.
* Platform: one of "windows", "osx", "snap", "linux", or "all".
* Urgency: either "normal" or "critical".

The text message can contain links.

You can use the `motd-cli` tool to parse and validate the json:

```
❯ ./motd-cli
file: motd-example.json
count: 1

Message 1 ✓
-----------
Type: daily ✓
Platform: all ✓
Urgency: normal ✓
Languages: 2 ✓
```

Use `motd-cli -url https://example.com/motd.json` to validate a remote file.

Notes: I'm considering adding an explicit layer of verification of the motd
payload. Please comment on
[#554](https://0xacab.org/leap/bitmask-vpn/-/issues/554) if you have an opinion
on this.
