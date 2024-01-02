package internal

import (
	"fmt"
	"log/slog"
	"sync"
)

const (
	Subscribe = iota
	Unsubscribe
	Publish
)

type (
	Topics map[string]Topic
	Topic  map[string]chan string
)

type PubSub struct {
	sync.RWMutex
	topics Topics
}

func NewPubSub() *PubSub {
	return &PubSub{
		topics: make(Topics),
	}
}

func (p *PubSub) Subscribe(id, topic string) (<-chan string, error) {
	slog.Debug("Subscribe", "id", id, "topic", topic)

	p.Lock()
	defer p.Unlock()

	if p.topics[topic] == nil {
		p.topics[topic] = make(Topic)
	}
	if p.topics[topic][id] != nil {
		return nil, fmt.Errorf("`%v` is already subscribed to `%v`", id, topic)
	}

	in := make(chan string)
	p.topics[topic][id] = in
	return in, nil
}

func (p *PubSub) Unsubscribe(id, topic string) error {
	slog.Debug("Unsubscribe", "id", id, "topic", topic)

	p.Lock()
	defer p.Unlock()

	if p.topics[topic] == nil {
		return fmt.Errorf("`%v` is not subscribed to `%v`", id, topic)
	}

	if p.topics[topic][id] == nil {
		return fmt.Errorf("`%v` is not subscribed to `%v`", id, topic)
	}

	close(p.topics[topic][id])
	p.topics[topic][id] = nil
	return nil
}

func (p *PubSub) UnsubscribeAll(id string) {
	slog.Debug("UnsubscribeAll", "id", id)

	p.Lock()
	defer p.Unlock()

	for topic, idMap := range p.topics {
		channel := idMap[id]
		if channel != nil {
			close(channel)
		}
		p.topics[topic][id] = nil
	}
}

func (p *PubSub) Publish(topic, message string) {
	slog.Debug("Publish", "topic", topic, "msg", message)

	p.RLock()
	defer p.RUnlock()

	if p.topics[topic] == nil {
		return
	}

	for _, in := range p.topics[topic] {
		in <- message
	}
}
