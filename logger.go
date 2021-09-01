package statsig

import (
	"strconv"
	"time"
)

const (
	maxEvents           = 500
	gateExposureEvent   = "statsig::gate_exposure"
	configExposureEvent = "statsig::config_exposure"
)

type logEventInput struct {
	Events          []Event  `json:"events"`
	StatsigMetadata metadata `json:"statsigMetadata"`
}

type logEventResponse struct{}

type logger struct {
	events    []Event
	transport *transport
	tick      *time.Ticker
}

func newLogger(transport *transport) *logger {
	log := &logger{
		events:    make([]Event, 0),
		transport: transport,
		tick:      time.NewTicker(time.Minute),
	}

	go log.backgroundFlush()

	return log
}

func (l *logger) backgroundFlush() {
	for range l.tick.C {
		l.Flush(false)
	}
}

func (l *logger) Log(evt Event) {
	evt.User.PrivateAttributes = nil
	l.events = append(l.events, evt)
	if len(l.events) >= maxEvents {
		l.Flush(false)
	}
}

func (l *logger) LogGateExposure(
	user User,
	gateName string,
	value bool,
	ruleID string,
) {
	evt := &Event{
		User:      user,
		EventName: gateExposureEvent,
		Metadata: map[string]string{
			"gate":      gateName,
			"gateValue": strconv.FormatBool(value),
			"ruleID":    ruleID,
		},
	}
	l.Log(*evt)
}

func (l *logger) LogConfigExposure(
	user User,
	configName string,
	ruleID string,
) {
	evt := &Event{
		User:      user,
		EventName: configExposureEvent,
		Metadata: map[string]string{
			"config": configName,
			"ruleID": ruleID,
		},
	}
	l.Log(*evt)
}

func (l *logger) Flush(closing bool) {
	if closing {
		l.tick.Stop()
	}
	if len(l.events) == 0 {
		return
	}

	if closing {
		l.sendEvents(l.events)
	} else {
		go l.sendEvents(l.events)
	}

	l.events = make([]Event, 0)
}

func (l *logger) sendEvents(events []Event) {
	input := &logEventInput{
		Events:          events,
		StatsigMetadata: l.transport.metadata,
	}
	var res logEventResponse
	l.transport.retryablePostRequest("/log_event", input, &res, maxRetries)
}
