package sql

import (
	"database/sql/driver"
	"encoding"
	"fmt"
	"time"
)

// DateTimeUTC ensures time is always in UTC
type DateTimeUTC struct {
	time.Time
}

var _ encoding.TextMarshaler = (*DateTimeUTC)(nil)

// Scan implements the sql.Scanner interface
func (t *DateTimeUTC) Scan(value interface{}) error {
	if value == nil {
		*t = DateTimeUTC{Time: time.Time{}}
		return nil
	}

	// Convert to time.Time
	switch v := value.(type) {
	case time.Time:
		*t = DateTimeUTC{Time: v.UTC()} // Ensure UTC
	default:
		return fmt.Errorf("cannot convert %T to UTCDateTime", value)
	}
	return nil
}

// Value implements the driver.Valuer interface (for inserting data)
func (t DateTimeUTC) Value() (driver.Value, error) {
	return t.UTC(), nil
}
