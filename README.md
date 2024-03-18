# aria2-hoook

aria2-hook provides a simple way to add invocation for aria2's download event.

## Configuration

The structure of the configuration can be referenced from this file: [config.go](https://github.com/HumXC/aria2-hook/blob/main/config/config.go#L15).

aria2-hook has three ways for reading configuration:

-   `Command line options`
-   `Environment variables`
-   `Configuration files`

If multiple configuration methods are used simultaneously, the final configuration will be overridden according to the priority order:

[ `Command line options` > `Environment variables` > `Configuration files` ].

### Command line options

Run `aria2-hook --help`

### Environment variables

As simple as this:

```bash
URL="https://domain.example:6800/jsonrpc" \
TOKEN="your_token" \
ON_DOWNLOAD_START="touch ${Name}.Start" \ # <- This command actually serves no real purpose; it's only used for testing.
aria2-hook
```

### Configuration files

```yaml
url: "https://domain.example:6800/jsonrpc"
token: "your_token"

onDownloadStart:
    - "touch {{Name}}"
    - "send-notification.sh {{Name}} {{TotalLength}} {{CompletedLength}}"
onDownloadPause:
    # async
    - "ASYNC: sleep 10"
    - "sleep 10"
onDownloadStop:
onDownloadComplete:
onDownloadError:
    - "send-notification.sh {{Name}} {{ErrCode}} {{ErrMsg}}"
onBtDownloadComplete:
```

### Multi-line command

When configuring using both command line and environment variables, if you need to execute multiple commands within one event, you need to separate multiple commands with '#' symbol. As an example of configuring with environment variables:

The ON_DOWNLOAD_START is interpreted as a string list within the program. In the environment variables, '#' is used as a delimiter, for example:

`ON_DOWNLOAD_START="touch ${Name}.Start#echo ${Name}.Start#"`

`ON_DOWNLOAD_START` is actually interpreted as three commands:

Split(ON_DOWNLOAD_START, "#")

-   `"touch ${Name}.Start"`
-   `"echo ${Name}.Start"`
-   `""` (empty string)

## Async

If a command starts with the `ASYNC:` keyword, it will be called asynchronously. If the `ASYNC:` keyword is not used, all commands under an event will be called in a blocking manner. If the previous command has not finished, the next command will not be called.

## Placeholders

You can use placeholders in the command to reference information related to the download task.
| Placeholder | Description |
| --- | --- |
| `{{Gid}}` | GID of the download task |
| `{{Name}}` | Filename of the download task |
| `{{TotalLength}}` | Total size of the download task |
| `{{CompletedLength}}` | Size already downloaded for the task |
| `{{ErrCode}}` | Error code of the download task (if any) |
| `{{ErrMsg}}` | Error message of the download task (if any) |

## Use docker

I use aria2-hook to set up message sending with [ntfy.sh](https://ntfy.sh). When a download event occurs, I receive messages from ntfy.

```yaml
version: "3.8"
services:
    aria2-hook:
        image: humxc/aria2-hook
        container_name: aria2-hook
        restart: always
        environment:
            TZ: ${TZ}
            TOKEN: ${ARIA2_RPC_SECRET}
            URL: http://aria2:6800/jsonrpc
            ON_DOWNLOAD_START: >
                curl
                -H "Title: Aria2 download start"
                -d "{{Name}} - {{CompletedLength}}/{{TotalLength}}"
                ${NTFY_URL}
            ON_DOWNLOAD_STOP: >
                curl
                -H "Title: Aria2 download stop"
                -d "{{Name}} - {{CompletedLength}}/{{TotalLength}}"
                ${NTFY_URL}
            ON_DOWNLOAD_PAUSE: >
                curl
                -H "Title: Aria2 download pause"
                -d "{{Name}} - {{CompletedLength}}/{{TotalLength}}"
                ${NTFY_URL}
            ON_DOWNLOAD_COMPLETE: >
                curl
                -H "Title: Aria2 download complete"
                -d "{{Name}} - {{CompletedLength}}"
                ${NTFY_URL}
            ON_BT_DOWNLOAD_COMPLETE: >
                curl
                -H "Title: Aria2 download complete"
                -d "{{Name}} - {{CompletedLength}}"
                ${NTFY_URL}
            ON_DOWNLOAD_ERROR: >
                curl
                -H "Title: Aria2 download error"
                -d "{{Name}}: {{ErrMsg}}"
                ${NTFY_URL}
```
