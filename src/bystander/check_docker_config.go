package bystander

// DockerCheckConfig defines a check that runs docker
type DockerCheckConfig struct {
	CheckCommonConfig
	image   string
	command string
	env     map[string]string
}

func parseDockerCheck(c map[interface{}]interface{}) CheckConfig {
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
func (s *DockerCheckConfig) Init(vars map[string]string) (Check, error) {
	c := &DockerCheck{}
	initCheckCommon(c, s, vars)

	c.image = subVar(s.image, c.Common().tags, false)
	c.imagePublic = subVar(s.image, c.Common().tagsPublic, true)

	c.command = subVar(s.command, c.Common().tags, false)
	c.commandPublic = subVar(s.command, c.Common().tagsPublic, true)

	c.env = subVars(s.env, c.Common().tags, false)
	c.envPublic = subVars(s.env, c.Common().tagsPublic, true)

	return c, nil
}
