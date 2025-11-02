package persistence

var instance Storage

// Return the Storage instance
func GetInstance() Storage {
	return instance
}

func init() {
	// If you need to change the Storage implementation, change this
	instance = NewInMemoryDB()
}
