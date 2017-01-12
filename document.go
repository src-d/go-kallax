package kallax

import (
	"gopkg.in/mgo.v2/bson"
)

type DocumentBase interface {
	GetId() bson.ObjectId
	SetId(bson.ObjectId)
	IsNew() bool
	SetIsNew(isNew bool)
}

type Document struct {
	Id bson.ObjectId `bson:"_id" json:"_id"`

	//Tracks if the document has been saved or recovered from the db or not.
	isNew bool
}

// SetId sets the document id.
func (d *Document) SetId(id bson.ObjectId) {
	d.Id = id
}

// GetId returns the document id.
func (d *Document) GetId() bson.ObjectId {
	return d.Id
}

// SetIsNew configures is this document is new in the store or not, dont mess
// with this if you dont want have duplicate records on your database.
func (d *Document) SetIsNew(isNew bool) {
	d.isNew = isNew
}

// IsNew returns if this document is new or not.
func (d *Document) IsNew() bool {
	return d.isNew
}
