package pagination

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
)

// Cursor - a generic cursor for all entities
type Cursor struct {
	UniqueID    string `json:"uniqueId"`
	OrderingKey string `json:"orderingKey"`
}

func (c *Cursor) Encode() (string, error) {
	bc, err := json.Marshal(c)
	if err != nil {
		return "", fmt.Errorf("failed to marshal cursor: %v: with error: %v", c, err)
	}
	return base64.URLEncoding.EncodeToString(bc), nil
}

func (c *Cursor) Decode(encoded string) error {
	bc, err := base64.URLEncoding.DecodeString(encoded)
	if err != nil {
		return fmt.Errorf("failed to decode cursor: %v: with error: %v", c, err)
	}

	err = json.Unmarshal(bc, c)
	if err != nil {
		return fmt.Errorf("failed to unmarshal byte array :%v to cursor with error: %v", string(bc), err)
	}
	return nil
}
