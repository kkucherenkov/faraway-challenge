package pow

type POW interface {
	Check() bool              // Check - checks that hash has leading <zerosCount> zeros
	Compute() (string, error) // Compute - bruteforce calculation correct POW
}
