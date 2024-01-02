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

func (c *Context) GetPubSub() *PubSub {
	return c.pubsub
}

func (c *Context) GetRegistry() *Registry {
	return c.registry
}
