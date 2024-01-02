package internal

import (
	"errors"
	"log/slog"
	"sync"
)

const (
	Subscribe = iota
	Unsubscribe
	Publish
)

var (
	AlreadySubscribed = errors.New("Already subscribed")
	NotSubscribed     = errors.New("Not subscribed")
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
	logger := slog.With("id", id, "topic", topic)

	logger.Debug("Subscribe")

	p.Lock()
	defer p.Unlock()

	if p.topics[topic] == nil {
		logger.Debug("Topic is uninitialized, creating map")
		p.topics[topic] = make(Topic)
	}

	if p.topics[topic][id] != nil {
		logger.Debug("Already subscribed")
		return nil, AlreadySubscribed
	}

	in := make(chan string, 1)
	p.topics[topic][id] = in
	return in, nil
}

func (p *PubSub) Unsubscribe(id, topic string) error {
	logger := slog.With("id", id, "topic", topic)

	logger.Debug("Unsubscribe")

	p.Lock()
	defer p.Unlock()

	if p.topics[topic] == nil || p.topics[topic][id] == nil {
		logger.Debug("Not subscribed")
		return NotSubscribed
	}

	close(p.topics[topic][id])
	delete(p.topics[topic], id)
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
		delete(p.topics[topic], id)
	}
}

func (p *PubSub) Publish(topic, message string) {
	logger := slog.With("topic", topic)

	logger.Debug("Publish", "msg", message)

	p.RLock()
	defer p.RUnlock()

	if p.topics[topic] == nil {
		logger.Debug("Topic is uninitialized")
		return
	}

	for id, in := range p.topics[topic] {
		logger.Debug("Publishing to client", "id", id)
		in <- message
	}
}
