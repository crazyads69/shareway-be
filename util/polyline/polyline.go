package polyline

import "database/sql/driver"

type Polyline string

func (p *Polyline) Scan(value interface{}) error {
	if value == nil {
		*p = ""
		return nil
	}
	*p = Polyline(value.([]byte))
	return nil
}

func (p Polyline) Value() (driver.Value, error) {
	return string(p), nil
}
