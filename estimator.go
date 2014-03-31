package hyperloglog

// Estimator is the interface for a multiset cardinality estimator
type Estimator interface {
	// Add adds the value to the estimator. The value must be hashed
	// before being added.
	Add(uint64)

	// Count returns the exact cardinality of the distinct values
	// it has seen so far.
	Count() uint64
}
