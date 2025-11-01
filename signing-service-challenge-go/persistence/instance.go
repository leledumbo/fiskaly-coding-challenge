package persistence

var instance Storage

func GetInstance() Storage {
	return instance
}

func init() {
	instance = NewInMemoryDB()
}
