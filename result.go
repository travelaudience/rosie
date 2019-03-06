package rosie

// Result ...
type Result struct {
	taskName string
	key      string
	err      error
	value    interface{}
}

// Err if returns non-nil error (after task being completed) indicates that task did not finish successfully.
func (r Result) Err() error {
	return r.err
}

// Value ...
func (r Result) Value() interface{} {
	if res, ok := r.value.(Result); ok {
		return res.Value()
	}

	return r.value
}
