# aria2-hoook

简单地配置 aria2 的 hook，在触发事件后运行一些命令。

## 配置示例

程序默认读取 `./config.yaml`

```yaml
url: "https://domain.example:6800/jsonrpc"
token: "your_token"

onDownloadStart:
    - "touch ${Name}"
    - "send-notification.sh ${Name} ${TotalLength} ${CompletedLength}"
onDownloadPause:
    # 异步执行
    - "ASYNC: sleep 10"
    - "sleep 10"
onDownloadStop:
onDownloadComplete:
onDownloadError:
    - "send-notification.sh ${Name} ${ErrCode} ${ErrMsg}"
onBtDownloadComplete:
```

使用 yaml 进行配置，onDownloadPause, onDownloadStop 等是事件名称，值是字符串列表。每个元素为一条命令，会在触发对应事件后使用 sh 解释和执行配置命令列表。

### 异步执行

默认情况下命令会阻塞地执行每一条命令，直到所有命令执行完成。在命令前加上 `ASYNC:` 前缀，命令将异步地运行。

### 占位符

在命令中可以使用占位符来引用下载任务的相关信息。
| 占位符 | 说明 |
| --- | --- |
｜`${Gid}` | 下载任务的 GID |
｜`${Name}` | 下载任务的文件名 |
｜`${TotalLength}` | 下载任务的总大小 |
｜`${CompletedLength}` | 下载任务已经下载的大小 |
｜`${ErrCode}` | 下载任务的错误码 (如果有) |
｜`${ErrMsg}` | 下载任务的错误信息 (如果有) |
