package internal

type Context struct {
	pubsub   *PubSub
	registry *Registry
}

func NewContext() Context {
	return Context{
		pubsub:   NewPubSub(),
		registry: NewRegistry(),
	}
}

func (c *Context) PubSub() *PubSub {
	return c.pubsub
}

func (c *Context) Registry() *Registry {
	return c.registry
}
