package bystander

import (
	"fmt"
	"os/exec"

	shlex "github.com/flynn-archive/go-shlex"
)

// DockerCheckConfig defines a check that runs docker
type DockerCheckConfig struct {
	CheckConfig
	image   string
	command string
	env     map[string]string
}

func parseDockerCheck(c map[interface{}]interface{}) Check {
	image, ok := c["image"]
	if !ok {
		panic("missing docker image")
	}

	command, ok := c["command"]
	if !ok {
		command = ""
	}

	env := map[string]string{}
	if _, ok := c["env"]; ok {
		for _, x := range c["env"].([]interface{}) {
			xx := x.(map[interface{}]interface{})
			for k, v := range xx {
				kk := k.(string)
				if _, ok := env[kk]; ok {
					panic("key defined twice")
				}
				env[kk] = v.(string)
			}
		}
	}

	return &DockerCheckConfig{
		image:   image.(string),
		command: command.(string),
		env:     env,
	}

}

// Run runs the check
func (s *DockerCheckConfig) Run() (bool, map[string]string) {
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
