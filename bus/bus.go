package bus

import (
	"autobangumi-go/utils"

	"github.com/rs/zerolog"
)

type EventHandler interface {
	HandleEvent(event Event)
}

type FnEventHandler struct {
	inner func(event Event)
}

func (handler *FnEventHandler) HandleEvent(event Event) {
	handler.inner(event)
}

type Event struct {
	EventType string
	Inner     interface{}
}

type PendingEvent struct {
	Topic string
	Inner Event
}

type EventBus struct {
	pendingEvents chan PendingEvent
	handlers      map[string][]EventHandler
	logger        zerolog.Logger
}

func NewEventBus() *EventBus {
	eb := EventBus{
		pendingEvents: make(chan PendingEvent, 4096),
		handlers:      make(map[string][]EventHandler),
		logger:        utils.GetLogger("event-bus"),
	}
	return &eb
}

func (eb *EventBus) Subscribe(topic string, handler EventHandler) {
	eb.handlers[topic] = append(eb.handlers[topic], handler)
}

func (eb *EventBus) SubscribeWithFn(topic string, fn func(event Event)) {
	eb.handlers[topic] = append(eb.handlers[topic], &FnEventHandler{inner: fn})
}

func (eb *EventBus) Publish(topic string, event Event) {
	eb.logger.Trace().Str("topic", topic).Str("event type", event.EventType).Msg("publising event")
	pending := PendingEvent{
		Topic: topic,
		Inner: event,
	}
	eb.pendingEvents <- pending
	eb.logger.Trace().Str("topic", topic).Str("event type", event.EventType).Msg("published event")
}

func (eb *EventBus) Start() {
	go eb.runLoop()
}

func (eb *EventBus) runLoop() {
	for ev := range eb.pendingEvents {
		eb.logger.Trace().Str("topic", ev.Topic).Str("event type", ev.Inner.EventType).Msg("dispatch event")
		go eb.dispatch(ev.Topic, ev.Inner)
	}
	panic("event bus channel close")
}

func (eb *EventBus) dispatch(topic string, event Event) {
	if handlers, found := eb.handlers[topic]; found && len(handlers) > 0 {
		for _, handler := range handlers {
			eb.logger.Trace().Str("topic", topic).Str("event type", event.EventType).Msg("handle event")
			handler.HandleEvent(event)
		}
	} else {
		eb.logger.Warn().Str("topic", topic).Str("event type", event.EventType).Msg("non handlers, this event will be discard")
	}
}
