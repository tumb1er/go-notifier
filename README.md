Go-Notifier
===========

`go-notifier` is a gui application intended for showing periodic notification received from another service.

Example:

```bash

# start event generator
$> python3 server.py &

# start notification listener
$> go-notifier localhost:9999 icon.ico
```

![Ubuntu screenshot](screenshot-ubuntu.png)

![Windows 7 screenshot](screenshot-windows7.png)