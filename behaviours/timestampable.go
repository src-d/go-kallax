package behaviours

import "time"

// Timestampable must be implemented by those object that can be timestamped
type Timestampable interface {
	Timestamp()
}

// Timestamp modelates an object that knows about when was created and updated
type TimestampDates struct {
	// CreatedAt is the time where the object was created
	CreatedAt time.Time
	// UpdatedAt is the time where the object was updated
	UpdatedAt time.Time
	now       func() time.Time
}

// Timestamp updates the UpdatedAt and creates a new CreatedAt if it does not exist
func (t *TimestampDates) Timestamp() {
	if t.now == nil {
		t.now = time.Now
	}

	if t.CreatedAt.IsZero() {
		t.CreatedAt = t.now()
	}

	t.UpdatedAt = t.now()
}

// BeforePersist runs all actions that must be performed before the persist
//  - Timestamp
func (t *TimestampDates) BeforePersist() error {
	t.Timestamp()
	return nil
}
