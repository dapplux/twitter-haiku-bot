package entities

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

type Author struct {
	ID       string
	Username string
}

// Scan implements the sql.Scanner interface for PostgreSQL JSONB
func (a *Author) Scan(value interface{}) error {
	if value == nil {
		*a = Author{}
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to scan Author: invalid type %T", value)
	}

	return json.Unmarshal(bytes, a)
}

// Value implements the driver.Valuer interface to store as JSONB
func (a Author) Value() (driver.Value, error) {
	return json.Marshal(a)
}
