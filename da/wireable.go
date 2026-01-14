package da

// Wireable receives the concrete container for wiring dependencies.
// Components implement this interface with their specific container type
// to receive the container during the wiring phase.
//
// This is used with the concrete container pattern where developers define
// their own container struct with explicit types, as opposed to the
// reflection-based Container type.
type Wireable[C any] interface {
	Wire(c *C) error
}
