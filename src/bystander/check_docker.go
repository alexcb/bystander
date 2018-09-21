package bystander

import (
	"fmt"
	"os/exec"

	shlex "github.com/flynn-archive/go-shlex"
)

// DockerCheck defines a check that runs docker
type DockerCheck struct {
	CheckCommon
	image   string
	command string
	env     map[string]string
}

// Command returns the command
func (s *DockerCheck) Command() []string {
	configArgs, err := shlex.Split(s.command)
	if err != nil {
		panic(err)
	}
	args := []string{"docker", "run", "--network", "host", "--rm"}

	for k, v := range s.env {
		args = append(args, "-e", fmt.Sprintf("%s=%s", k, v))
	}

	args = append(args, s.image)
	args = append(args, configArgs...)

	return args
}

// Run runs the check
func (s *DockerCheck) Run() (bool, map[string]string) {
	args := s.Command()
	cmd := exec.Command(args[0], args[1:]...)
	stdoutStderr, err := cmd.CombinedOutput()

	details := map[string]string{
		"output": string(stdoutStderr),
		"err":    fmt.Sprintf("%v", err),
	}

	ok := err == nil
	return ok, details
}
