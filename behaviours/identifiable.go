package behaviours

import "github.com/src-d/go-kallax/common"

//Identifiable must be implemented by those values that can be identified by an ID
type Identifiable interface {
	// GetID returns the ID.
	GetID() common.ID
	// SetID overrides the ID.
	SetID(id common.ID)
	// Identify sets the ID if it does not exist and returns the ID of the object
	Identify() common.ID
}

// Identificator modelates an object that knows about its ID and generates it if neccessary
type Identificator struct {
	ID common.ID
}

// GetID returns the ID.
func (i *Identificator) GetID() common.ID {
	return i.ID
}

// SetID overrides the ID.
//  The ID should not be modified once it has been set and stored in the DB
func (i *Identificator) SetID(id common.ID) {
	i.ID = id
}

// Identify sets the ID if it does not exist and returns the ID of the object
func (i *Identificator) Identify() common.ID {
	if i.ID.IsEmpty() {
		i.ID = common.NewID()
	}

	return i.ID
}

// BeforeInsert runs all actions that must be performed before the insetion¡
//  - Identify
func (i *Identificator) BeforeInsert() error {
	i.Identify()
	return nil
}
