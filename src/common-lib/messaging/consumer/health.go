package consumer

// Health : Returns current health of the Kafka connection with Service
type Health struct {
	ConnectionState bool
	Address         []string
	Topics          []string
	Group           string
}

func newHealth() *Health {
	return &Health{
		ConnectionState: false,
	}
}
