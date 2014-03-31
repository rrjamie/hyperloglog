package hyperloglog

import (
	"encoding/binary"
	"fmt"
	"hash/fnv"
	"math"
	"math/rand"
	"testing"
)

func GenerateUInt64(n uint64) uint64 {
	hash := fnv.New64a()
	binary.Write(hash, binary.BigEndian, n)
	return hash.Sum64()
}

func TestCountLeadingZeros(t *testing.T) {
	test := func(i uint64, expected byte) {
		actual := CountLeadingZeros(i)
		if expected != actual {
			t.Errorf("Expected %d zeros, got %d. \nActual:   %064b",
				expected, actual, i)
		}
	}

	test(0, 64)
	test(1, 63)
	test(1<<60, 3)
	test(1<<31, 32)
	test(1<<63, 0)

}

func TestSplitWord64(t *testing.T) {
	test := func(word uint64, width uint, expectedA, expectedB uint64) {
		actualA, actualB := SplitWord64(word, width)
		if (expectedA != actualA) || (expectedB != actualB) {
			t.Errorf("Head: Expected %d, got %d.\nTail: Expected %d, got %d.",
				expectedA, actualA, expectedB, actualB)
		}
	}

	test(1<<63, 10, 1<<9, 0)
	test(1<<63+1<<31, 32, 1<<31, 1<<63)
}

func TestExactEstimator(t *testing.T) {
	e := NewExactEstimator()

	for i := 0; i < 5; i++ {
		for j := 0; j < 1000; j++ {
			e.Add(uint64(j))
		}
	}

	count := e.Count()

	if count != 1000 {
		t.Errorf("Expected exact count to be 1000, got %d", count)
	}
}

func BenchmarkHLL(b *testing.B) {
	rand.Seed(1)

	hll := NewHLL(16)
	for i := 0; i < b.N; i++ {
		hll.Add(GenerateUInt64(uint64(i)))
	}
}

func ExampleHLLEstimator() {
	hll := NewHLL(8)

	hll.Add(GenerateUInt64(1))

	fmt.Printf("%d\n", hll.Count())
	// Output: 1
}

func TestLargeCardinality(t *testing.T) {
	rand.Seed(1)

	registerWidth := uint(10)

	hll := NewHLL(registerWidth)
	exact := NewExactEstimator()

	for i := 0; i < 100000; i++ {
		r := GenerateUInt64(uint64(rand.Int63()))
		hll.Add(r)
		exact.Add(r)
	}

	expectedError := 2 * (1.04 / math.Sqrt(math.Pow(2.0, float64(registerWidth))))
	setSize := exact.Count()
	lowerBound := uint64(math.Floor((1 - expectedError) * float64(setSize)))
	upperBound := uint64(math.Floor((1 + expectedError) * float64(setSize)))

	count := hll.Count()

	if count < lowerBound || count > upperBound {
		t.Errorf("Count is outside expected range: %d < %d < %d", lowerBound, count, upperBound)
	}
}

func TestAllZeroesHash(t *testing.T) {
	hll := NewHLL(8)

	hll.Add(0)

	count := hll.Count()

	if count != 1 {
		t.Errorf("Expected %d, got %d", 1, count)
	}

	if hll.registers[0] != (64 - 8 + 1) {
		t.Errorf("Expected register[0]=57, got %d", hll.registers[0])
	}
}

func TestHLLSmallWidthPanic(t *testing.T) {
	assertWidthPanics := func(width uint) {
		// Capture Panic
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("Expected NewHLL(%d) to panic.", width)
			}
		}()

		NewHLL(width)
	}

	assertWidthPanics(4)
	assertWidthPanics(7)
	NewHLL(8)
	NewHLL(20)
	assertWidthPanics(21)
}
