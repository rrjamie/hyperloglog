// Package hyperloglog provides an implementation of the probabilistic
// distinct value counter as described in ...TODO...
package hyperloglog

import (
	"errors"
	"math"
)

// CountLeadingZeros returns the number of leading binary zeroes in a
// unsigned integer.
func CountLeadingZeros(i uint64) byte {
	zeros := byte(0)

	// Handles case of all 0-bits
	if i == 0 {
		return 64
	}

	for j := i; (j != 0) && (j&0x8000000000000000 == 0); j = j << 1 {
		zeros++
	}

	return zeros
}

// SplitWord64 splits an uint64 in to two uint64s (head, tail) along
// at the nth bit.
//
// head is composed of the first (most significant) n-bits, then shifted
// to the right so they use only the last (least significant) bits.
// Eg, n=8, the maximum value of head will be 2**8 - 1 = 255.
//
// tail is composed of the remaining bits (64-n least significant bits),
// then shifted to the left so the most significant bit of the remainging
// bits becomes the most-significant bit of the uint64.
func SplitWord64(word uint64, n uint) (uint64, uint64) {
	// First portion, the head, is shifted to the right
	head := word >> (64 - n)

	// Second portion, the tail, is shifted to the left
	tail := word << n

	return head, tail
}

//
func HarmonicMean(arr []byte) float64 {
	sum := 0.0

	for _, i := range arr {
		sum += math.Pow(2, -float64(i))
	}

	return 1.0 / sum
}

// HLLConstant returns the normalization constant for a given
// register count.
// Only valid for registerCount >= 128
func HLLConstant(registerCount uint) float64 {
	return 0.7213 / (1.0 + 1.079/float64(registerCount))
}

// HLLEstimator implements the HyperLogLog algorithm.
// It implements the Estimator interface.
type HLLEstimator struct {
	registerWidth uint
	registers     []byte
}

// NewHLL returns a fully initialized HLLEstimator.
// The registerWidth specifies the size (in bits) of the index in to
// the registers. More bits gives better accuracy, but the space grows
// like O(2^n).
//
// registerWidth must be between [8..20]
func NewHLL(registerWidth uint) *HLLEstimator {
	// We limit the sizes to be larger than 7 (because below that
	// we do not have the proper normalization constant)
	// We limit the sizes to be at most 20 due to size limitations.
	if registerWidth < 8 || registerWidth > 20 {
		panic(errors.New("NewHLL: registrWidth must be in [8..20]"))
	}

	registerCount := 1 << registerWidth

	return &HLLEstimator{
		registerWidth: registerWidth,
		registers:     make([]byte, registerCount)}
}

// Add adds the value to the estimator. The value must be hashed
// before being added.
func (e *HLLEstimator) Add(key uint64) {
	register, remainder := SplitWord64(key, e.registerWidth)

	// A seen register is at least 1, each leading zero on the remainder
	// adds to the count.
	count := CountLeadingZeros(remainder) + 1

	// Because remainder is actually a (64 - registerWidth) number,
	// we have to be careful and cap the highest possible count, as
	// the last registerWidth-bits of the remainder are always zero,
	// we could overestimate the cardinality.
	//
	// Since this only happens when the remainder is all zeros,
	// if remainder is 0, we just set it to the expected
	// maximum. The odds of this are very low.
	if remainder == 0 {
		count = byte((64 - e.registerWidth) + 1)
	}

	// Update the count
	currentCount := e.registers[register]

	if count > currentCount {
		e.registers[register] = count
	}
}

// Count returns an estimate of the cardinality of the distinct values
// it has seen so far.
func (e *HLLEstimator) Count() uint64 {
	registerCount := uint(1) << e.registerWidth

	// HLL Estimate
	estimate := HLLConstant(registerCount) * float64(registerCount) * float64(registerCount)
	estimate = estimate * HarmonicMean(e.registers)

	// If we have a low enough estimate, we fall back to linear counting which has
	// better accuracy at low cardinalities.
	if estimate < (5.0 / 2.0 * float64(registerCount)) {
		zeros := e.countZeros()

		// Linear counting does not work if we have filled all the registers.
		if zeros != 0 {
			return e.LinearCount()
		}
	}

	return uint64(math.Floor(estimate))
}

// countZeros returns the number of zero registers
func (e *HLLEstimator) countZeros() uint {
	zeros := uint(0)
	for _, i := range e.registers {
		if i == 0 {
			zeros++
		}
	}

	return zeros
}

// LinearCount estimates the candinality of the set using linear counting
func (e *HLLEstimator) LinearCount() uint64 {
	zeros := e.countZeros()

	registerCount := float64(uint(1) << e.registerWidth)

	return uint64(math.Floor(registerCount * math.Log(registerCount/float64(zeros))))
}
