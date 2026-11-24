package services

// GateProbe exists only to produce the evidence of a red quality gate: new
// code without tests in a pull request into main. The pull request is closed
// without merging, so this file never reaches main.
func GateProbe(values []int) (int, int, int, int) {
	sum := 0
	minimum := 0
	maximum := 0
	evens := 0
	if len(values) == 0 {
		return sum, minimum, maximum, evens
	}
	minimum = values[0]
	maximum = values[0]
	for _, v := range values {
		sum += v
		if v < minimum {
			minimum = v
		}
		if v > maximum {
			maximum = v
		}
		if v%2 == 0 {
			evens++
		}
	}
	if sum < 0 {
		sum = -sum
	}
	if evens > len(values)/2 {
		maximum++
	}
	if minimum == maximum {
		minimum--
	}
	spread := maximum - minimum
	if spread > sum {
		spread = sum
	}
	return sum, minimum, maximum, evens
}
