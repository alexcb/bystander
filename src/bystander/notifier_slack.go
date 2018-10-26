package bystander

// SlackNotifierConfig defines a slack notifier
type SlackNotifierConfig struct {
	NotifierConfig
	webhook   string
	okColor   string
	failColor string

	alerter *slackAlerter
}

func parseSlackNotifier(c map[interface{}]interface{}) Notifier {
	webhook, ok := c["webhook"]
	if !ok {
		panic("webhook missing")
	}

	okColor, ok := c["ok_color"]
	if !ok {
		okColor = "#00CC00"
	}

	failColor, ok := c["fail_color"]
	if !ok {
		failColor = "#CC0000"
	}

	return &SlackNotifierConfig{
		webhook:   webhook.(string),
		okColor:   okColor.(string),
		failColor: failColor.(string),
	}
}

func (s *SlackNotifierConfig) Init(webAddr, serverID string) error {
	s.alerter = newSlackAlerter(s.webhook, webAddr, serverID, s.okColor, s.failColor)
	return nil
}

// Notify runs the notifier
func (s *SlackNotifierConfig) Notify(id, checkName string, ok bool, details map[string]string) error {
	s.alerter.alert(id, checkName, ok, details)
	return nil
}
