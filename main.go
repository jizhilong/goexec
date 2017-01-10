package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/codegangsta/cli"
	"github.com/yudai/gotty/app"
	"github.com/yudai/gotty/utils"

	"github.com/jizhilong/goexec/docker-exec-gotty-backend"
)

var version = "undefined"

func main() {
	cmd := cli.NewApp()
	cmd.Name = "goexec"
	cmd.Version = version
	cmd.Usage = "Access terminal of docker containers as web application"
	cmd.HideHelp = true
	cli.AppHelpTemplate = helpTemplate

	appOptions := &app.Options{}
	if err := utils.ApplyDefaultValues(appOptions); err != nil {
		exit(err, 1)
	}
	backendOptions := &dockerExec.Options{}
	if err := utils.ApplyDefaultValues(backendOptions); err != nil {
		exit(err, 1)
	}

	cliFlags, flagMappings, err := utils.GenerateFlags(appOptions, backendOptions)
	if err != nil {
		exit(err, 3)
	}

	cmd.Flags = cliFlags

	cmd.Action = func(c *cli.Context) {
		utils.ApplyFlags(cliFlags, flagMappings, c, appOptions, backendOptions)

		appOptions.EnableBasicAuth = c.IsSet("credential")
		appOptions.EnableTLSClientAuth = c.IsSet("tls-ca-crt")

		if err := app.CheckConfig(appOptions); err != nil {
			exit(err, 6)
		}

		manager := dockerExec.NewContextManager(backendOptions)
		app, err := app.New(manager, appOptions)
		if err != nil {
			exit(err, 3)
		}

		registerSignals(app)

		err = app.Run()
		if err != nil {
			exit(err, 4)
		}
	}
	cmd.Run(os.Args)
}

func exit(err error, code int) {
	if err != nil {
		fmt.Println(err)
	}
	os.Exit(code)
}

func registerSignals(app *app.App) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(
		sigChan,
		syscall.SIGINT,
		syscall.SIGTERM,
	)

	go func() {
		for {
			s := <-sigChan
			switch s {
			case syscall.SIGINT, syscall.SIGTERM:
				if app.Exit() {
					fmt.Println("Send ^C to force exit.")
				} else {
					os.Exit(5)
				}
			}
		}
	}()
}
