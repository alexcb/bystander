package bystander

import (
	"fmt"
	"os/exec"

	shlex "github.com/flynn-archive/go-shlex"
)

// DockerCheck defines a check that runs docker
type DockerCheck struct {
	CheckCommon
	image         string
	imagePublic   string
	command       string
	commandPublic string
	env           map[string]string
	envPublic     map[string]string
	volumes       map[string]string
}

// Command returns the command
func (s *DockerCheck) constructCommand(image, command string, env, volumes map[string]string) []string {
	configArgs, err := shlex.Split(command)
	if err != nil {
		panic(err)
	}
	args := []string{"docker", "run", "--network", "host", "--rm"}

	for k, v := range env {
		args = append(args, "-e", fmt.Sprintf("%s=%s", k, v))
	}

	for k, v := range volumes {
		args = append(args, "-v", fmt.Sprintf("%s:%s", k, v))
	}

	args = append(args, image)
	args = append(args, configArgs...)

	return args
}

// Command returns the command
func (s *DockerCheck) Command() []string {
	return s.constructCommand(s.image, s.command, s.env, s.volumes)
}

// CommandPublic returns a public version of the command without any secrets
func (s *DockerCheck) CommandPublic() []string {
	return s.constructCommand(s.imagePublic, s.commandPublic, s.envPublic, s.volumes)
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
