package crypto

var algorithms map[string]Algorithm

// GetAlgorithm returns algorithm by @name, or nil if no algorithm exists with that name
func GetAlgorithm(name string) Algorithm {
	if algo, exists := algorithms[name]; exists {
		return algo
	} else {
		return nil
	}
}

// RegisterAlgorithm inserts algorithm @algo with name @name so it can be retrieved later with GetAlgorithm
func RegisterAlgorithm(name string, algo Algorithm) {
	if algorithms == nil {
		algorithms = make(map[string]Algorithm)
	}
	algorithms[name] = algo
}
