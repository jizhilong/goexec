package dockerExec

import (
	"errors"
	"io"
	"log"
	"net/url"

	"github.com/fsouza/go-dockerclient"
	"github.com/yudai/gotty/backends"
)

var Command []string = []string{"env", "TERM=xterm-256color", "sh", "-c", "if command -v bash > /dev/null;then exec bash;else exec sh;fi"}

type Options struct {
}

type DockerExecClientContextManager struct {
	docker  *docker.Client
	options *Options
}

type DockerExecClientContext struct {
	docker       *docker.Client
	containerId  string
	exec         *docker.Exec
	stdinReader  io.ReadCloser  // read by docker client
	stdinWriter  io.WriteCloser // write by our code when proxying inputs from ws
	stdoutReader io.ReadCloser  // read by our code, will be proxied to ws
	stdoutWriter io.WriteCloser // write by docker client
}

func NewContextManager(options *Options) *DockerExecClientContextManager {
	dockerClient, err := docker.NewClientFromEnv()
	if err != nil {
		return nil
	} else {
		return &DockerExecClientContextManager{docker: dockerClient, options: options}
	}
}

func (mgr *DockerExecClientContextManager) New(params url.Values) (context backends.ClientContext, err error) {
	var exec *docker.Exec
	if len(params["container"]) == 0 {
		return nil, errors.New("no container specified")
	} else if len(params["container"]) != 1 {
		return nil, errors.New("multiple containers specified")
	}
	containerId := params["container"][0]

	opts := docker.CreateExecOptions{
		Container:    containerId,
		AttachStdin:  true,
		AttachStdout: true,
		AttachStderr: true,
		Tty:          true,
		Cmd:          Command,
	}

	if exec, err = mgr.docker.CreateExec(opts); err != nil {
		return
	}
	stdinPipeReader, stdinPipeWriter := io.Pipe()
	stdoutPipeReader, stdoutPipeWriter := io.Pipe()

	context = &DockerExecClientContext{
		docker:       mgr.docker,
		containerId:  containerId,
		exec:         exec,
		stdinReader:  stdinPipeReader,
		stdinWriter:  stdinPipeWriter,
		stdoutReader: stdoutPipeReader,
		stdoutWriter: stdoutPipeWriter,
	}
	log.Printf("succedd to create docker exec")
	return
}

func (context *DockerExecClientContext) WindowTitle() (title string, err error) {
	return context.containerId, nil
}

func (context *DockerExecClientContext) Start(exitCh chan bool) {
	go func() {
		defer func() { exitCh <- true }()
		if err := context.docker.StartExec(context.exec.ID, docker.StartExecOptions{
			Detach:       false,
			OutputStream: context.stdoutWriter,
			ErrorStream:  context.stdoutWriter,
			InputStream:  context.stdinReader,
			RawTerminal:  false,
		}); err != nil {
			log.Printf("failed to start docker exec %v: %v", context.exec.ID, err)
		} else {
			log.Printf("docker exec %v finished", context.exec.ID)
		}
	}()
}

func (context *DockerExecClientContext) InputWriter() io.Writer {
	return context.stdinWriter
}

func (context *DockerExecClientContext) OutputReader() io.Reader {
	return context.stdoutReader
}

func (context *DockerExecClientContext) ResizeTerminal(width, height uint16) error {
	log.Printf("width: %v, height: %v", width, height)
	if width >= 0 && height >= 0 {
		if err := context.docker.ResizeExecTTY(context.exec.ID, int(height), int(width)); err != nil {
			return err
		}
	} else {
		return errors.New("invalid new tty size")
	}
	return nil
}

func (context *DockerExecClientContext) TearDown() error {
	exitKeySeq := []byte{4, 4}
	context.stdinWriter.Write(exitKeySeq)
	context.stdinReader.Close()
	context.stdoutWriter.Close()
	return nil
}
