package container

import "github.com/numbereddev/zero-daemon/internal/events"

func (c *Container) Emitter() events.ReadBus[[]byte] {
	return c.eventBus
}

func (c *Container) SetState(action string) {
	c.stateMx.Lock()
	c.state = action
	c.stateMx.Unlock()
}

func (c *Container) State() string {
	c.stateMx.Lock()
	defer c.stateMx.Unlock()
	return c.state
}
