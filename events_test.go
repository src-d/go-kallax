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
	}

	after struct {
		model
		evented
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
	return nil
}

func (b *before) BeforeUpdate() error {
	b.setup()
	b.events["BeforeUpdate"]++
	return errors.New("foo")
}

func (b *before) BeforeSave() error {
	b.setup()
	b.events["BeforeSave"]++
	return nil
}

func (b *after) AfterInsert() error {
	b.setup()
	b.events["AfterInsert"]++
	return nil
}

func (b *after) AfterUpdate() error {
	b.setup()
	b.events["AfterUpdate"]++
	return nil
}

func (b *after) AfterSave() error {
	b.setup()
	b.events["AfterSave"]++
	return errors.New("foo")
}

func TestApplyBeforeEvents(t *testing.T) {
	r := require.New(t)

	var before before
	r.Nil(ApplyBeforeEvents(&before))
	before.setPersisted()
	r.NotNil(ApplyBeforeEvents(&before))

	r.Equal(1, before.events["BeforeInsert"])
	r.Equal(1, before.events["BeforeUpdate"])
	r.Equal(2, before.events["BeforeSave"])
}

func TestApplyAfterEvents(t *testing.T) {
	r := require.New(t)

	var after after
	r.NotNil(ApplyAfterEvents(&after, false))
	r.NotNil(ApplyAfterEvents(&after, true))

	r.Equal(1, after.events["AfterInsert"])
	r.Equal(1, after.events["AfterUpdate"])
	r.Equal(2, after.events["AfterSave"])
}
