package kallax

// AfterLoader must be implemented by those entities listening the AfterLoad event
type AfterLoader interface {
	// AfterLoad runs after the entitie has been loaded from DB
	AfterLoad()
}

// AfterLoader must be implemented by those entities listening the BeforeSave event
type BeforeSaver interface {
	// BeforeSave runs before the entitie has being Saved in the DB.
	//  It will return an error if the Saveing process should be interrupted
	BeforeSave() error
}

// AfterLoader must be implemented by those entities listening the AfterSave event
type AfterSaver interface {
	// AfterSave runs after the entitie has being Saved in the DB.
	AfterSave()
}

// AfterLoader must be implemented by those entities listening the BeforeUpdate event
type BeforeUpdater interface {
	// BeforeUpdate runs before the entitie has been updated in the DB
	//  It will return an error if the update process should be interrupted
	BeforeUpdate() error
}

// AfterLoader must be implemented by those entities listening the AfterUpdate event
type AfterUpdater interface {
	// AfterUpdate runs after the entitie has been updated in the DB
	AfterUpdate()
}

// AfterLoader must be implemented by those entities listening the BeforeInsert event
type BeforeInserter interface {
	// BeforeInsert runs before the entitie has been inserted in the DB
	//  It will return an error if the insertion process should be interrupted
	BeforeInsert() error
}

// AfterLoader must be implemented by those entities listening the AfterInsert event
type AfterInserter interface {
	// AfterInsert runs after the entitie has been inserted in the DB
	AfterInsert()
}

// AfterLoader must be implemented by those entities listening the BeforeDelete event
type BeforeDeleter interface {
	// BeforeDelete runs before the entitie has been deleted from DB
	//  It will return an error if the deletion process should be interrupted
	BeforeDelete() error
}

// AfterLoader must be implemented by those entities listening the AfterDelete event
type AfterDeleter interface {
	// AfterDelete runs after the entitie has been deleted from DB
	AfterDelete()
}
