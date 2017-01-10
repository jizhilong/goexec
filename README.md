# goexec-get access to your docker container in a web browser
## how to use it?

```
go get github.com/jizhilong/goexec
goexec -w
# open http://<hostip>:8000/?container=<containerid> in your browser
```

## how goexec works.
goexec is a command line tool based on [gotty](https://github.com/yudai/gotty), it works by converting docker-daemon's exec websocket to gotty-protocol.
