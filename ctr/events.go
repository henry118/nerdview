package ctr

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/containerd/containerd/v2/core/events"
)

type EventMsg struct {
	Namespace string
	Topic     string
	Timestamp time.Time
}

type EventErrMsg struct {
	Err error
}

func WaitForEvent(c *Client) tea.Cmd {
	return func() tea.Msg {
		select {
		case env, ok := <-c.EventCh():
			if !ok {
				return EventErrMsg{Err: fmt.Errorf("event channel closed")}
			}
			return eventFromEnvelope(env)
		case err, ok := <-c.ErrCh():
			if !ok {
				return EventErrMsg{Err: fmt.Errorf("error channel closed")}
			}
			if err != nil {
				return EventErrMsg{Err: err}
			}
			return nil
		}
	}
}

func eventFromEnvelope(env *events.Envelope) EventMsg {
	return EventMsg{
		Namespace: env.Namespace,
		Topic:     env.Topic,
		Timestamp: env.Timestamp,
	}
}
