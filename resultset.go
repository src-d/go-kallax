package kallax

import (
	"errors"

	"gopkg.in/mgo.v2"
)

var (
	//ErrResultSetClosed is throwed when you are working over a closed ResultSet
	ErrResultSetClosed = errors.New("closed resultset")
	//ErrNotFound document not found
	ErrNotFound = errors.New("document not found")
	// ErrStop if is used on a callback of a ResultSet.ForEach function the loop
	// is stopped
	ErrStop = errors.New("document not found")
)

// ResultSet contains the result of an executed query command.
type ResultSet struct {
	IsClosed bool
	session  *mgo.Session
	mgoQuery *mgo.Query
	mgoIter  *mgo.Iter
}

// Count returns the total number of documents in the ResultSet. Count DON'T
// close the ResultSet after be called.
func (r *ResultSet) Count() (int, error) {
	return r.mgoQuery.Count()
}

// All returns all the documents in the ResultSet and close it. Dont use it
// with large results.
func (r *ResultSet) All(result interface{}) error {
	defer r.Close()
	return r.mgoQuery.All(result)
}

// One return a document from the ResultSet and close it, the following calls
// to One returns ErrResultSetClosed error. If a document is not returned the
// error ErrNotFound is retuned.
func (r *ResultSet) One(doc interface{}) error {
	defer r.Close()
	found, err := r.Next(doc)
	if err != nil {
		return err
	}

	if !found {
		return ErrNotFound
	}

	return nil
}

// Next return a document from the ResultSet, can be called multiple times.
func (r *ResultSet) Next(doc interface{}) (bool, error) {
	if r.mgoIter == nil {
		r.mgoIter = r.mgoQuery.Iter()
	}

	returned := r.mgoIter.Next(doc)
	if !returned {
		r.Close()
	}

	return returned, r.mgoIter.Err()
}

// Close close the ResultSet closing the internal iter.
func (r *ResultSet) Close() error {
	if r.IsClosed {
		return ErrResultSetClosed
	}

	defer func() {
		r.session.Close()
		r.IsClosed = true
	}()

	if r.mgoIter == nil {
		return nil
	}

	return r.mgoIter.Close()
}
