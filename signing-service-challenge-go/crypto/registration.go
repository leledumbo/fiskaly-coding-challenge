package crypto

var algorithms map[string]Algorithm

func GetAlgorithm(name string) Algorithm {
	if algo, exists := algorithms[name]; exists {
		return algo
	} else {
		return nil
	}
}

func RegisterAlgorithm(name string, algo Algorithm) {
	if algorithms == nil {
		algorithms = make(map[string]Algorithm)
	}
	algorithms[name] = algo
}
