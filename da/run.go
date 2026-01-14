package da

import (
	"os"
	"os/signal"
	"reflect"
	"syscall"
)

// Wire calls Wire(c) on all Wireable[C] components in the container.
// Components are processed in order specified by `da:"order=N"` tags.
func Wire[C any](c *C) error {
	v := reflect.ValueOf(c)
	components := traverse(v)

	for _, comp := range components {
		obj := comp.value.Interface()
		if wirer, ok := obj.(Wireable[C]); ok {
			if err := wirer.Wire(c); err != nil {
				return err
			}
		}
	}
	return nil
}

// Start calls Start() on all Startable components in the container.
// Components are processed in order specified by `da:"order=N"` tags.
func Start[C any](c *C) error {
	v := reflect.ValueOf(c)
	components := traverse(v)

	for _, comp := range components {
		obj := comp.value.Interface()
		if starter, ok := obj.(Startable); ok {
			if err := starter.Start(); err != nil {
				return err
			}
		}
	}
	return nil
}

// Stop calls Stop() on all Stoppable components in the container.
// Components are processed in reverse order of `da:"order=N"` tags.
// Continues on error and returns the first error encountered.
func Stop[C any](c *C) error {
	v := reflect.ValueOf(c)
	components := traverse(v)

	// reverse order for shutdown
	var firstErr error
	for i := len(components) - 1; i >= 0; i-- {
		obj := components[i].value.Interface()
		if stopper, ok := obj.(Stoppable); ok {
			if err := stopper.Stop(); err != nil && firstErr == nil {
				firstErr = err
			}
		}
	}
	return firstErr
}

// Run is a convenience function that: Wire -> Start -> wait for signal -> Stop.
// Blocks until SIGINT or SIGTERM is received.
func Run[C any](c *C) error {
	if err := Wire(c); err != nil {
		return err
	}
	if err := Start(c); err != nil {
		Stop(c)
		return err
	}
	WaitForSignal()
	return Stop(c)
}

// WaitForSignal blocks until SIGINT or SIGTERM is received.
func WaitForSignal() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch
}
