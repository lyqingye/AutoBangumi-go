package bus

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
}

func NewEventBus() *EventBus {
	eb := EventBus{
		pendingEvents: make(chan PendingEvent, 4096),
		handlers:      make(map[string][]EventHandler),
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
	pending := PendingEvent{
		Topic: topic,
		Inner: event,
	}
	eb.pendingEvents <- pending
}

func (eb *EventBus) Start() {
	go eb.runLoop()
}

func (eb *EventBus) runLoop() {
	for ev := range eb.pendingEvents {
		eb.dispatch(ev.Topic, ev.Inner)
	}
	panic("event bus channel close")
}

func (eb *EventBus) dispatch(topic string, event Event) {
	for _, handler := range eb.handlers[topic] {
		handler.HandleEvent(event)
	}
}
