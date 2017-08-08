package guard

// Randomizer generates random numbers.
type Randomizer interface {
	// Float64 returns random floating point number in [0.0,1.0).
	Float64() float64
}
