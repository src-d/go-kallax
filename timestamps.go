package kallax

import "time"

// Timestamps contains the dates of the last time the model was created
// or deleted. Because this is such a common functionality in models, it is
// provided by default by the library. It is intended to be embedded in the
// model.
//
//	type MyModel struct {
//		kallax.Model
//		kallax.Timestamps
//		Foo string
//	}
type Timestamps struct {
	// CreatedAt is the time where the object was created.
	CreatedAt time.Time
	// UpdatedAt is the time where the object was updated.
	UpdatedAt time.Time
}

// BeforeSave updates the last time the model was updated every single time the
// model is saved, and the last time the model was created only if the model
// has no date of creation yet.
func (t *Timestamps) BeforeSave() error {
	if t.CreatedAt.IsZero() {
		t.CreatedAt = time.Now()
	}

	t.UpdatedAt = time.Now()
	return nil
}
