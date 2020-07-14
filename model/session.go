package model

import "fmt"

type SessionList []Session

type Session struct {
	UserName string
	Count    int
}

func (sl SessionList) GetStatText() string {
	text := "Statistics:"
	for _, s := range sl {
		text += fmt.Sprintf("\n%s: %d messages", s.UserName, s.Count)
	}
	return text
}

func (sl SessionList) GetUsers() string {
	text := ""
	for _, s := range sl {
		if text != "" {
			text += "|"
		}
		text += s.UserName
	}
	return text
}
