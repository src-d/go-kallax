package kallax

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

type (
	evented struct {
		events map[string]int
	}

	before struct {
		model
		evented
		errorBeforeSave   bool
		errorBeforeInsert bool
		errorBeforeUpdate bool
	}

	after struct {
		model
		evented
		errorAfterSave   bool
		errorAfterInsert bool
		errorAfterUpdate bool
	}
)

func (e *evented) setup() {
	if e.events == nil {
		e.events = make(map[string]int)
	}
}

func (b *before) BeforeInsert() error {
	b.setup()
	b.events["BeforeInsert"]++
	if b.errorBeforeInsert {
		return errors.New("foo")
	}
	return nil
}

func (b *before) BeforeUpdate() error {
	b.setup()
	b.events["BeforeUpdate"]++
	if b.errorBeforeUpdate {
		return errors.New("foo")
	}
	return nil
}

func (b *before) BeforeSave() error {
	b.setup()
	b.events["BeforeSave"]++
	if b.errorBeforeSave {
		return errors.New("foo")
	}
	return nil
}

func (b *after) AfterInsert() error {
	b.setup()
	b.events["AfterInsert"]++
	if b.errorAfterInsert {
		return errors.New("foo")
	}
	return nil
}

func (b *after) AfterUpdate() error {
	b.setup()
	b.events["AfterUpdate"]++
	if b.errorAfterUpdate {
		return errors.New("foo")
	}
	return nil
}

func (b *after) AfterSave() error {
	b.setup()
	b.events["AfterSave"]++
	if b.errorAfterSave {
		return errors.New("foo")
	}
	return nil
}

func TestApplyBeforeEvents(t *testing.T) {
	r := require.New(t)

	var before before
	r.Nil(ApplyBeforeEvents(&before))
	before.setPersisted()
	r.Nil(ApplyBeforeEvents(&before))

	r.Equal(1, before.events["BeforeInsert"])
	r.Equal(1, before.events["BeforeUpdate"])
	r.Equal(2, before.events["BeforeSave"])

	before.errorBeforeUpdate = true
	r.NotNil(ApplyBeforeEvents(&before))

	before.errorBeforeInsert = true
	before.errorBeforeUpdate = false
	before.persisted = false
	r.NotNil(ApplyBeforeEvents(&before))

	before.errorBeforeSave = true
	before.errorBeforeInsert = false
	r.NotNil(ApplyBeforeEvents(&before))
}

func TestApplyAfterEvents(t *testing.T) {
	r := require.New(t)

	var after after
	r.Nil(ApplyAfterEvents(&after, false))
	r.Nil(ApplyAfterEvents(&after, true))

	r.Equal(1, after.events["AfterInsert"])
	r.Equal(1, after.events["AfterUpdate"])
	r.Equal(2, after.events["AfterSave"])

	after.errorAfterUpdate = true
	r.NotNil(ApplyAfterEvents(&after, true))

	after.errorAfterInsert = true
	after.errorAfterUpdate = false
	r.NotNil(ApplyAfterEvents(&after, false))

	after.errorAfterSave = true
	after.errorAfterInsert = false
	r.NotNil(ApplyAfterEvents(&after, false))
}
