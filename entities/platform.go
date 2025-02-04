package entities

import (
	"database/sql/driver"
	"fmt"
)

type Platform string

const (
	PlatformTwitter Platform = "twitter"
)

// Scan for Platform
func (p *Platform) Scan(value interface{}) error {
	if value == nil {
		*p = ""
		return nil
	}
	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("failed to scan Platform: invalid type %T", value)
	}
	*p = Platform(str)
	return nil
}

// Value for Platform
func (p Platform) Value() (driver.Value, error) {
	return string(p), nil
}
