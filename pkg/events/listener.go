package events

type Listener interface {
	Listen(chan Event)
}
