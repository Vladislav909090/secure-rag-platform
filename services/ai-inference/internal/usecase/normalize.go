package usecase

import "math"

func normalizeVectors(vectors [][]float32) {
	for i := range vectors {
		normalizeVector(vectors[i])
	}
}

func normalizeVector(vector []float32) {
	if len(vector) == 0 {
		return
	}

	var sum float64
	for _, value := range vector {
		sum += float64(value * value)
	}
	if sum == 0 {
		return
	}

	norm := float32(math.Sqrt(sum))
	for i := range vector {
		vector[i] = vector[i] / norm
	}
}
