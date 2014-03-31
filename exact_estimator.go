package hyperloglog

// ExactEstimator generates an exact estimate using an in-memory map.
type ExactEstimator struct {
	values map[uint64]bool
}

// NewExactEstimator creates a new ExactEstimator
func NewExactEstimator() *ExactEstimator {
	return &ExactEstimator{values: make(map[uint64]bool)}
}

// Add adds the value to the estimator. The value must be hashed
// before being added.
func (e *ExactEstimator) Add(key uint64) {
	e.values[key] = true
}

// Count returns the exact cardinality of the distinct values
// it has seen so far.
func (e *ExactEstimator) Count() uint64 {
	return uint64(len(e.values))
}
