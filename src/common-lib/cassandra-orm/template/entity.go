package template

// Entity describes fake entity
type Entity struct{}

// AcquireID acquire model's ID if it not present/unique
func (m *Entity) AcquireID() error {
	return nil
}
