# goexec-get access to your docker container in a web browser
## how to use it?
the first way is to build and run on your host.

```
go get github.com/jizhilong/goexec
goexec -w
# open http://<hostip>:8000/?container=<containerid> in your browser
```

another way is to run goexec inside a container.

```
docker run -td --net host jizhilong/goexec goexec -w -p 8000
# open http://<hostip>:8000/?container=<containerid> in your browser
```

## how goexec works.
goexec is a command line tool based on [gotty](https://github.com/yudai/gotty), it works by converting docker-daemon's exec websocket to gotty-protocol.
