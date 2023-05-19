package bus_test

import (
	"pikpak-bot/bus"
	"testing"
	"time"
)

func TestEventBus(t *testing.T) {
	eb := bus.NewEventBus()
	eb.Start()
	topic := "topic"
	eventType := "type"

	msgNum := 1000000

	var recvValues []int
	eb.SubscribeWithFn(topic, func(event bus.Event) {
		switch event.EventType {
		case eventType:
			recvValues = append(recvValues, event.Inner.(int))
		}
	})
	go func() {
		for i := 0; i < msgNum; i++ {
			eb.Publish(topic, bus.Event{
				EventType: eventType,
				Inner:     i,
			})
		}
	}()

	for {
		if len(recvValues) != msgNum {
			time.Sleep(time.Second)
		} else {
			break
		}
	}
}
