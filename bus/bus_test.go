package bus_test

import (
	"autobangumi-go/bus"
	"testing"
	"time"
)

func TestEventBus(t *testing.T) {
	eb := bus.NewEventBus()
	eb.Start()
	topic := "topic"
	eventType := "type"

	msgNum := 100

	var recvValues []int
	eb.SubscribeWithFn(topic, func(event bus.Event) {
		switch event.EventType {
		case eventType:
			time.Sleep(time.Second)
			recvValues = append(recvValues, event.Inner.(int))
		}
	})
	go func() {
		for i := 0; i < msgNum; i++ {
			go func() {
				time.Sleep(time.Second)
				eb.Publish(topic, bus.Event{
					EventType: eventType,
					Inner:     i,
				})
			}()
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
