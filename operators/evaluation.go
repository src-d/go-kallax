package operators

import (
	"gopkg.in/mgo.v2/bson"
)

// Mod Performs a modulo operation on the value of a field and selects documents
// with a specified result.
func Mod(field Field, divisor, remainder float64) bson.M {
	return bson.M{field.String(): bson.M{"$mod": []float64{divisor, remainder}}}
}

// RegEx Selects documents where values match a specified regular expression.
func RegEx(field Field, regexp, options string) bson.M {
	return bson.M{field.String(): bson.M{"$regex": bson.RegEx{regexp, options}}}
}

// Text performs a text search on the content of the fields indexed with a text
// index.
func Text(field Field, search, lang string) bson.M {
	return bson.M{
		field.String(): bson.M{
			"$text": bson.M{"$search": search, "$language": lang},
		},
	}
}

// Where Matches documents that satisfy a JavaScript expression.
func Where(field Field, code string, scope interface{}) bson.M {
	return bson.M{
		field.String(): bson.M{
			"$where": bson.JavaScript{Code: code, Scope: scope},
		},
	}
}
