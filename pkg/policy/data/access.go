package data

type DataAccessDriver interface {
	// since we are only concerned with
	Delete(string, string) error
}

type DriverFactory struct {
}
