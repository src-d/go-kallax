package kallax

// BeforeInserter will do some operations before being inserted.
type BeforeInserter interface {
	// BeforeInsert will do some operations before being inserted. If an error is
	// returned, it will prevent the insert from happening.
	BeforeInsert() error
}

// BeforeUpdater will do some operations before being updated.
type BeforeUpdater interface {
	// BeforeUpdate will do some operations before being updated. If an error is
	// returned, it will prevent the update from happening.
	BeforeUpdate() error
}

// BeforeSaver will do some operations before being updated or inserted.
type BeforeSaver interface {
	// BeforeSave will do some operations before being updated or inserted. If an
	// error is returned, it will prevent the update or insert from happening.
	BeforeSave() error
}

// BeforeDeleter will do some operations before being deleted.
type BeforeDeleter interface {
	// BeforeDelete will do some operations before being deleted. If an error is
	// returned, it will prevent the delete from happening.
	BeforeDelete() error
}

// AfterInserter will do some operations after being inserted.
type AfterInserter interface {
	// AfterInsert will do some operations after being inserted. If an error is
	// returned, it will cause the insert to be rolled back.
	AfterInsert() error
}

// AfterUpdater will do some operations after being updated.
type AfterUpdater interface {
	// AfterUpdate will do some operations after being updated. If an error is
	// returned, it will cause the update to be rolled back.
	AfterUpdate() error
}

// AfterSaver will do some operations after being inserted or updated.
type AfterSaver interface {
	// AfterSave will do some operations after being inserted or updated. If an
	// error is returned, it will cause the insert or update to be rolled back.
	AfterSave() error
}

// AfterDeleter will do some operations after being deleted.
type AfterDeleter interface {
	// AfterDelete will do some operations after being deleted. If an error is
	// returned, it will cause the delete to be rolled back.
	AfterDelete() error
}

// ApplyBeforeEvents calls all the update, insert or save before events of the
// record. Save events are always called before the insert or update event.
func ApplyBeforeEvents(r Record) error {
	if rec, ok := r.(BeforeSaver); ok {
		if err := rec.BeforeSave(); err != nil {
			return err
		}
	}

	if rec, ok := r.(BeforeInserter); ok && !r.IsPersisted() {
		return rec.BeforeInsert()
	}

	if rec, ok := r.(BeforeUpdater); ok && r.IsPersisted() {
		return rec.BeforeUpdate()
	}

	return nil
}

// ApplyAfterEvents calls all the update, insert or save after events of the
// record. Save events are always called after the insert or update event.
func ApplyAfterEvents(r Record, wasPersisted bool) error {
	if rec, ok := r.(AfterInserter); ok && !wasPersisted {
		if err := rec.AfterInsert(); err != nil {
			return err
		}
	}

	if rec, ok := r.(AfterUpdater); ok && wasPersisted {
		if err := rec.AfterUpdate(); err != nil {
			return err
		}
	}

	if rec, ok := r.(AfterSaver); ok {
		return rec.AfterSave()
	}

	return nil
}
