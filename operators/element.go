package operators

import (
	"gopkg.in/mgo.v2/bson"
)

type BSONType int

const (
	Double       BSONType = 1
	String                = 2
	Object                = 3
	Array                 = 4
	Binary                = 5
	Undefined             = 6
	ObjectId              = 7
	Boolean               = 8
	Date                  = 9
	Null                  = 10
	RegExp                = 11
	JavaScript            = 13
	Symbol                = 14
	JavaScriptWS          = 15
	Int32                 = 16
	Timestamp             = 17
	Int64                 = 18
	MinKey                = -1
	MaxKey                = 127
)

// Exists Matches documents that have the specified field.
func Exists(field Field, exists bool) bson.M {
	return bson.M{field.String(): bson.M{"$exists": exists}}
}

// Type Selects documents if a field is of the specified type.
func Type(field Field, t BSONType) bson.M {
	return bson.M{field.String(): bson.M{"$type": t}}
}
