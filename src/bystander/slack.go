package bystander

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type slackAlerter struct {
	webHook    string
	webAddress string
}

func newSlackAlerter(webHook, webAddress string) *slackAlerter {
	return &slackAlerter{
		webHook:    webHook,
		webAddress: webAddress,
	}
}

func escapeChars(s string) string {
	s = strings.Replace(s, "&", "&amp;", -1)
	s = strings.Replace(s, "<", "&lt;", -1)
	s = strings.Replace(s, ">", "&gt;", -1)
	return s
}

func (s *slackAlerter) alert(id, checkName string, ok bool, details map[string]string) {
	fields := []SlackField{}
	for k, v := range details {
		if v == "" {
			v = "<empty>"
		}
		fields = append(fields, SlackField{
			Title: k,
			Value: escapeChars(v),
			Short: false,
		})
	}

	titleLink := fmt.Sprintf("%s#%s", s.webAddress, id)

	var fallback string
	var color string
	if ok {
		fallback = fmt.Sprintf("%s is ok; visit %s for more information", checkName, s.webAddress)
		color = "#00CC00"
	} else {
		fallback = fmt.Sprintf("%s is not ok; visit %s for more information", checkName, s.webAddress)
		color = "#CC0000"
	}

	notification := &SlackNotification{
		Attachments: []SlackAttachment{
			{
				Fallback:   fallback,
				Color:      color,
				Title:      checkName,
				TitleLink:  titleLink,
				MarkDownIn: []string{"pretext", "text"},
				Fields:     fields,
			},
		},
	}

	err := sendNotification(s.webHook, notification)
	if err != nil {
		panic(err)
	}
}

func sendNotification(webhook string, notification *SlackNotification) error {
	buffer := bytes.NewBuffer(nil)
	if err := json.NewEncoder(buffer).Encode(notification); err != nil {
		return err
	}

	res, err := http.Post(webhook, "application/json", buffer)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		buffer.Reset()
		if _, err := io.Copy(buffer, res.Body); err != nil {
			return err
		}

		return &SlackError{
			Code: res.StatusCode,
			Body: buffer.String(),
		}
	}
	return nil
}

type SlackError struct {
	Code int
	Body string
}

func (e SlackError) Error() string {
	return fmt.Sprintf("slack webhook returned %d: %s", e.Code, e.Body)
}

type SlackNotification struct {
	Text        string            `json:"text"`
	Channel     string            `json:"channel,omitempty"`
	Username    string            `json:"username,omitempty"`
	IconEmoji   string            `json:"icon_emoji,omitempty"`
	IconURL     string            `json:"icon_url,omitempty"`
	Attachments []SlackAttachment `json:"attachments,omitempty"`
}

type SlackAttachment struct {
	Fallback   string       `json:"fallback"`
	Pretext    string       `json:"pretext,omitempty"`
	Text       string       `json:"text,omitempty"`
	Color      string       `json:"color,omitempty"`
	Title      string       `json:"title,omitmepty"`
	TitleLink  string       `json:"title_link,omitempty"`
	Fields     []SlackField `json:"fields"`
	MarkDownIn []string     `json:"mrkdwn_in"`
	Footer     string       `json:"footer"`
}

type SlackField struct {
	Title string      `json:"title"`
	Value interface{} `json:"value"`
	Short bool        `json:"short,omitempty"`
}
