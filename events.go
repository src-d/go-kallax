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
