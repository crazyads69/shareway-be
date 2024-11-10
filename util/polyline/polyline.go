package polyline

import (
	"database/sql/driver"
	"fmt"
)

type Polyline string

func (p *Polyline) Scan(value interface{}) error {
	if value == nil {
		*p = ""
		return nil
	}

	switch v := value.(type) {
	case []byte:
		*p = Polyline(v)
	case string:
		*p = Polyline(v)
	default:
		return fmt.Errorf("unsupported type for Polyline: %T", value)
	}
	return nil
}

func (p Polyline) Value() (driver.Value, error) {
	return string(p), nil
}
