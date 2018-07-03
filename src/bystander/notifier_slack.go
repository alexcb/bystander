package bystander

// SlackNotifierConfig defines a slack notifier
type SlackNotifierConfig struct {
	NotifierConfig
	webhook string

	alerter *slackAlerter
}

func parseSlackNotifier(c map[interface{}]interface{}) Notifier {
	webhook, ok := c["webhook"]
	if !ok {
		panic("webhook missing")
	}

	return &SlackNotifierConfig{
		webhook: webhook.(string),
	}
}

func (s *SlackNotifierConfig) Init(webAddr string) error {
	s.alerter = newSlackAlerter(s.webhook, webAddr)
	return nil
}

// Notify runs the notifier
func (s *SlackNotifierConfig) Notify(id, checkName string, ok bool, details map[string]string) error {
	s.alerter.alert(id, checkName, ok, details)
	return nil
}
