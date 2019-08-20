Go-Notifier
===========

`go-notifier` is a gui application intended for showing periodic notification received from another service.

Example
-------

```bash

# start event generator
$> python3 server.py &

# start notification listener
$> go-notifier localhost:9999 icon.ico
```

![Ubuntu screenshot](screenshot-ubuntu.png)

![Windows 7 screenshot](screenshot-windows7.png)

Message format
--------------

Current socket transport implementation uses `\n`-separated single-line json message (see `server.py`):
```
{"tooltip":"tip","title":"title","info":"info"}\n
```

```json
{
  "tooltip": "program title",
  "title": "popup header",
  "info": "popup text"
}
```

Command-line arguments
----------------------

```sh
go-notifier [address] [icon]
--address value  tcp address for socket server (default: "localhost:9998")
--icon value     notification icon path (default: "icon.ico")

```