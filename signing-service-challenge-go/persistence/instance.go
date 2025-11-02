package persistence

var instance Storage

// Return the Storage instance
func GetInstance() Storage {
	return instance
}

// Replace the Storage instance
func SetInstance(newInstance Storage) {
	instance = newInstance
}

func init() {
	// If you need to change the Storage implementation, change this
	instance = NewInMemoryDB()
}
