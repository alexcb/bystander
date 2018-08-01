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

// Run runs the check
func (s *DockerCheck) Run() (bool, map[string]string) {
	configArgs, err := shlex.Split(s.command)
	if err != nil {
		panic(err)
	}
	args := []string{"run", "--network", "host", "--rm"}

	for k, v := range s.env {
		args = append(args, "-e", fmt.Sprintf("%s=%s", k, v))
	}

	args = append(args, s.image)
	args = append(args, configArgs...)

	cmd := exec.Command("docker", args...)
	stdoutStderr, err := cmd.CombinedOutput()

	details := map[string]string{
		"output": string(stdoutStderr),
		"err":    fmt.Sprintf("%v", err),
	}

	ok := err == nil
	return ok, details
}
