package producer

// Health : Returns current health of the Kafka connection with Service
type Health struct {
	ConnectionState bool
	Address         []string
}

func newHealth() *Health {
	return &Health{
		ConnectionState: false,
	}
}
