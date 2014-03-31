// benchmaker prints out accuracy reports for range of register and set sizes
package main

import (
	"encoding/binary"
	"fmt"
	"github.com/rrjamie/hyperloglog"
	"hash/fnv"
	"math"
	"math/rand"
	"os"
	"text/tabwriter"
)

func GenerateUInt64(n uint64) uint64 {
	hash := fnv.New64a()
	binary.Write(hash, binary.BigEndian, n)
	return hash.Sum64()
}

type Result struct {
	width uint
	size  uint
	error float64
}

func calcEstimatorAccuracy(registerWidth uint, size uint, seed int64) (float64, uint64) {
	rand.Seed(seed)

	data := make([]uint64, size)
	for i := uint(0); i < size; i++ {
		data[i] = GenerateUInt64(uint64(rand.Int63()))
	}

	exact := hyperloglog.NewExactEstimator()
	hll := hyperloglog.NewHLL(registerWidth)
	for i := range data {
		exact.Add(data[i])
		hll.Add(data[i])
	}

	exactCount := float64(exact.Count())
	hllCount := float64(hll.Count())

	error := math.Abs(exactCount-hllCount) / exactCount

	return error, exact.Count()
}

func CalculateAccuracy() {
	results := make([]Result, 0)

	SAMPLES := 10

	for width := uint(8); width < 20; width++ {
		for i := uint(1); i < 10000000; i = i * 10 {
			totalError := 0.0

			for j := 0; j < SAMPLES; j++ {
				error, _ := calcEstimatorAccuracy(width, i, int64(i))
				totalError += error
			}

			results = append(results, Result{width, i, totalError / float64(SAMPLES)})
		}
	}

	// Ideally, this would be right-aligned, but I cannot seem to make it work.
	w := tabwriter.NewWriter(os.Stdout, 20, 8, 1, '\t', 0)

	fmt.Fprintf(w, "Register Width\tSet Size\tError\n")
	for _, r := range results {
		fmt.Fprintf(w, "%d\t%d\t%0.2f%%\n", r.width, r.size, r.error*100.0)
	}

	w.Flush()
}

func main() {
	CalculateAccuracy()
}
