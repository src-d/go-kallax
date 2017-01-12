package tests

import (
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"gopkg.in/src-d/storable.v1"
	"gopkg.in/src-d/storable.v1/operators"
)

type EventsFixtureStore struct {
	storable.Store
}

func NewEventsFixtureStore(db *mgo.Database) *EventsFixtureStore {
	return &EventsFixtureStore{*storable.NewStore(db, "event")}
}

// New returns a new instance of EventsFixture.
func (s *EventsFixtureStore) New() (doc *EventsFixture) {
	doc = newEventsFixture()
	if doc != nil {
		doc.SetIsNew(true)
		doc.SetId(bson.NewObjectId())
	}
	return
}

// Query return a new instance of EventsFixtureQuery.
func (s *EventsFixtureStore) Query() *EventsFixtureQuery {
	return &EventsFixtureQuery{*storable.NewBaseQuery()}
}

// Find performs a find on the collection using the given query.
func (s *EventsFixtureStore) Find(query *EventsFixtureQuery) (*EventsFixtureResultSet, error) {
	resultSet, err := s.Store.Find(query)
	if err != nil {
		return nil, err
	}

	return &EventsFixtureResultSet{ResultSet: *resultSet}, nil
}

// MustFind like Find but panics on error
func (s *EventsFixtureStore) MustFind(query *EventsFixtureQuery) *EventsFixtureResultSet {
	resultSet := s.Store.MustFind(query)
	return &EventsFixtureResultSet{ResultSet: *resultSet}
}

// FindOne performs a find on the collection using the given query returning
// the first document from the resultset.
func (s *EventsFixtureStore) FindOne(query *EventsFixtureQuery) (*EventsFixture, error) {
	resultSet, err := s.Find(query)
	if err != nil {
		return nil, err
	}

	return resultSet.One()
}

// MustFindOne like FindOne but panics on error
func (s *EventsFixtureStore) MustFindOne(query *EventsFixtureQuery) *EventsFixture {
	doc, err := s.FindOne(query)
	if err != nil {
		panic(err)
	}

	return doc
}

// Insert insert the given document on the collection, trigger BeforeInsert and
// AfterInsert if any. Throws ErrNonNewDocument if doc is a non-new document.
func (s *EventsFixtureStore) Insert(doc *EventsFixture) error {
	if err := s.BeforeInsert(doc); err != nil {
		return err
	}

	err := s.Store.Insert(doc)
	if err != nil {
		return err
	}

	return s.AfterInsert(doc)
}

// Update update the given document on the collection, trigger BeforeUpdate and
// AfterUpdate if any. Throws ErrNewDocument if doc is a new document.
func (s *EventsFixtureStore) Update(doc *EventsFixture) error {
	if err := s.BeforeUpdate(doc); err != nil {
		return err
	}

	err := s.Store.Update(doc)
	if err != nil {
		return err
	}

	return s.AfterUpdate(doc)
}

// Save insert or update the given document on the collection using Upsert,
// trigger BeforeUpdate and AfterUpdate if the document is non-new and
// BeforeInsert and AfterInset if is new.
func (s *EventsFixtureStore) Save(doc *EventsFixture) (updated bool, err error) {
	switch doc.IsNew() {
	case true:
		if err := s.BeforeInsert(doc); err != nil {
			return false, err
		}
	case false:
		if err := s.BeforeUpdate(doc); err != nil {
			return false, err
		}
	}

	updated, err = s.Store.Save(doc)
	if err != nil {
		return false, err
	}

	switch updated {
	case false:
		if err := s.AfterInsert(doc); err != nil {
			return false, err
		}
	case true:
		if err := s.AfterUpdate(doc); err != nil {
			return false, err
		}
	}

	return
}

type EventsFixtureQuery struct {
	storable.BaseQuery
}

// FindById add a new criteria to the query searching by _id
func (q *EventsFixtureQuery) FindById(ids ...bson.ObjectId) *EventsFixtureQuery {
	var vs []interface{}
	for _, id := range ids {
		vs = append(vs, id)
	}
	q.AddCriteria(operators.In(storable.IdField, vs...))

	return q
}

type EventsFixtureResultSet struct {
	storable.ResultSet
	last    *EventsFixture
	lastErr error
}

// All returns all documents on the resultset and close the resultset
func (r *EventsFixtureResultSet) All() ([]*EventsFixture, error) {
	var result []*EventsFixture
	err := r.ResultSet.All(&result)

	return result, err
}

// One returns the first document on the resultset and close the resultset
func (r *EventsFixtureResultSet) One() (*EventsFixture, error) {
	var result *EventsFixture
	err := r.ResultSet.One(&result)

	return result, err
}

// Next prepares the next result document for reading with the Get method.
func (r *EventsFixtureResultSet) Next() (returned bool) {
	r.last = nil
	returned, r.lastErr = r.ResultSet.Next(&r.last)

	return
}

// Get returns the document retrieved with the Next method.
func (r *EventsFixtureResultSet) Get() (*EventsFixture, error) {
	return r.last, r.lastErr
}

// ForEach iterates the resultset calling to the given function.
func (r *EventsFixtureResultSet) ForEach(f func(*EventsFixture) error) error {
	for {
		var result *EventsFixture
		found, err := r.ResultSet.Next(&result)
		if err != nil {
			return err
		}

		if !found {
			break
		}

		err = f(result)
		if err == storable.ErrStop {
			break
		}

		if err != nil {
			return err
		}
	}

	return nil
}

type EventsSaveFixtureStore struct {
	storable.Store
}

func NewEventsSaveFixtureStore(db *mgo.Database) *EventsSaveFixtureStore {
	return &EventsSaveFixtureStore{*storable.NewStore(db, "event")}
}

// New returns a new instance of EventsSaveFixture.
func (s *EventsSaveFixtureStore) New() (doc *EventsSaveFixture) {
	doc = newEventsSaveFixture()
	if doc != nil {
		doc.SetIsNew(true)
		doc.SetId(bson.NewObjectId())
	}
	return
}

// Query return a new instance of EventsSaveFixtureQuery.
func (s *EventsSaveFixtureStore) Query() *EventsSaveFixtureQuery {
	return &EventsSaveFixtureQuery{*storable.NewBaseQuery()}
}

// Find performs a find on the collection using the given query.
func (s *EventsSaveFixtureStore) Find(query *EventsSaveFixtureQuery) (*EventsSaveFixtureResultSet, error) {
	resultSet, err := s.Store.Find(query)
	if err != nil {
		return nil, err
	}

	return &EventsSaveFixtureResultSet{ResultSet: *resultSet}, nil
}

// MustFind like Find but panics on error
func (s *EventsSaveFixtureStore) MustFind(query *EventsSaveFixtureQuery) *EventsSaveFixtureResultSet {
	resultSet := s.Store.MustFind(query)
	return &EventsSaveFixtureResultSet{ResultSet: *resultSet}
}

// FindOne performs a find on the collection using the given query returning
// the first document from the resultset.
func (s *EventsSaveFixtureStore) FindOne(query *EventsSaveFixtureQuery) (*EventsSaveFixture, error) {
	resultSet, err := s.Find(query)
	if err != nil {
		return nil, err
	}

	return resultSet.One()
}

// MustFindOne like FindOne but panics on error
func (s *EventsSaveFixtureStore) MustFindOne(query *EventsSaveFixtureQuery) *EventsSaveFixture {
	doc, err := s.FindOne(query)
	if err != nil {
		panic(err)
	}

	return doc
}

// Insert insert the given document on the collection, trigger BeforeInsert and
// AfterInsert if any. Throws ErrNonNewDocument if doc is a non-new document.
func (s *EventsSaveFixtureStore) Insert(doc *EventsSaveFixture) error {
	if err := s.BeforeSave(doc); err != nil {
		return err
	}

	err := s.Store.Insert(doc)
	if err != nil {
		return err
	}

	return s.AfterSave(doc)
}

// Update update the given document on the collection, trigger BeforeUpdate and
// AfterUpdate if any. Throws ErrNewDocument if doc is a new document.
func (s *EventsSaveFixtureStore) Update(doc *EventsSaveFixture) error {
	if err := s.BeforeSave(doc); err != nil {
		return err
	}

	err := s.Store.Update(doc)
	if err != nil {
		return err
	}

	return s.AfterSave(doc)
}

// Save insert or update the given document on the collection using Upsert,
// trigger BeforeUpdate and AfterUpdate if the document is non-new and
// BeforeInsert and AfterInset if is new.
func (s *EventsSaveFixtureStore) Save(doc *EventsSaveFixture) (updated bool, err error) {
	if err := s.BeforeSave(doc); err != nil {
		return false, err
	}

	updated, err = s.Store.Save(doc)
	if err != nil {
		return false, err
	}

	if err := s.AfterSave(doc); err != nil {
		return false, err
	}
	return
}

type EventsSaveFixtureQuery struct {
	storable.BaseQuery
}

// FindById add a new criteria to the query searching by _id
func (q *EventsSaveFixtureQuery) FindById(ids ...bson.ObjectId) *EventsSaveFixtureQuery {
	var vs []interface{}
	for _, id := range ids {
		vs = append(vs, id)
	}
	q.AddCriteria(operators.In(storable.IdField, vs...))

	return q
}

type EventsSaveFixtureResultSet struct {
	storable.ResultSet
	last    *EventsSaveFixture
	lastErr error
}

// All returns all documents on the resultset and close the resultset
func (r *EventsSaveFixtureResultSet) All() ([]*EventsSaveFixture, error) {
	var result []*EventsSaveFixture
	err := r.ResultSet.All(&result)

	return result, err
}

// One returns the first document on the resultset and close the resultset
func (r *EventsSaveFixtureResultSet) One() (*EventsSaveFixture, error) {
	var result *EventsSaveFixture
	err := r.ResultSet.One(&result)

	return result, err
}

// Next prepares the next result document for reading with the Get method.
func (r *EventsSaveFixtureResultSet) Next() (returned bool) {
	r.last = nil
	returned, r.lastErr = r.ResultSet.Next(&r.last)

	return
}

// Get returns the document retrieved with the Next method.
func (r *EventsSaveFixtureResultSet) Get() (*EventsSaveFixture, error) {
	return r.last, r.lastErr
}

// ForEach iterates the resultset calling to the given function.
func (r *EventsSaveFixtureResultSet) ForEach(f func(*EventsSaveFixture) error) error {
	for {
		var result *EventsSaveFixture
		found, err := r.ResultSet.Next(&result)
		if err != nil {
			return err
		}

		if !found {
			break
		}

		err = f(result)
		if err == storable.ErrStop {
			break
		}

		if err != nil {
			return err
		}
	}

	return nil
}

type MultiKeySortFixtureStore struct {
	storable.Store
}

func NewMultiKeySortFixtureStore(db *mgo.Database) *MultiKeySortFixtureStore {
	return &MultiKeySortFixtureStore{*storable.NewStore(db, "query")}
}

// New returns a new instance of MultiKeySortFixture.
func (s *MultiKeySortFixtureStore) New() (doc *MultiKeySortFixture) {
	doc = &MultiKeySortFixture{}
	if doc != nil {
		doc.SetIsNew(true)
		doc.SetId(bson.NewObjectId())
	}
	return
}

// Query return a new instance of MultiKeySortFixtureQuery.
func (s *MultiKeySortFixtureStore) Query() *MultiKeySortFixtureQuery {
	return &MultiKeySortFixtureQuery{*storable.NewBaseQuery()}
}

// Find performs a find on the collection using the given query.
func (s *MultiKeySortFixtureStore) Find(query *MultiKeySortFixtureQuery) (*MultiKeySortFixtureResultSet, error) {
	resultSet, err := s.Store.Find(query)
	if err != nil {
		return nil, err
	}

	return &MultiKeySortFixtureResultSet{ResultSet: *resultSet}, nil
}

// MustFind like Find but panics on error
func (s *MultiKeySortFixtureStore) MustFind(query *MultiKeySortFixtureQuery) *MultiKeySortFixtureResultSet {
	resultSet := s.Store.MustFind(query)
	return &MultiKeySortFixtureResultSet{ResultSet: *resultSet}
}

// FindOne performs a find on the collection using the given query returning
// the first document from the resultset.
func (s *MultiKeySortFixtureStore) FindOne(query *MultiKeySortFixtureQuery) (*MultiKeySortFixture, error) {
	resultSet, err := s.Find(query)
	if err != nil {
		return nil, err
	}

	return resultSet.One()
}

// MustFindOne like FindOne but panics on error
func (s *MultiKeySortFixtureStore) MustFindOne(query *MultiKeySortFixtureQuery) *MultiKeySortFixture {
	doc, err := s.FindOne(query)
	if err != nil {
		panic(err)
	}

	return doc
}

// Insert insert the given document on the collection, trigger BeforeInsert and
// AfterInsert if any. Throws ErrNonNewDocument if doc is a non-new document.
func (s *MultiKeySortFixtureStore) Insert(doc *MultiKeySortFixture) error {

	err := s.Store.Insert(doc)
	if err != nil {
		return err
	}

	return nil
}

// Update update the given document on the collection, trigger BeforeUpdate and
// AfterUpdate if any. Throws ErrNewDocument if doc is a new document.
func (s *MultiKeySortFixtureStore) Update(doc *MultiKeySortFixture) error {

	err := s.Store.Update(doc)
	if err != nil {
		return err
	}

	return nil
}

// Save insert or update the given document on the collection using Upsert,
// trigger BeforeUpdate and AfterUpdate if the document is non-new and
// BeforeInsert and AfterInset if is new.
func (s *MultiKeySortFixtureStore) Save(doc *MultiKeySortFixture) (updated bool, err error) {
	updated, err = s.Store.Save(doc)
	if err != nil {
		return false, err
	}

	return
}

type MultiKeySortFixtureQuery struct {
	storable.BaseQuery
}

// FindById add a new criteria to the query searching by _id
func (q *MultiKeySortFixtureQuery) FindById(ids ...bson.ObjectId) *MultiKeySortFixtureQuery {
	var vs []interface{}
	for _, id := range ids {
		vs = append(vs, id)
	}
	q.AddCriteria(operators.In(storable.IdField, vs...))

	return q
}

type MultiKeySortFixtureResultSet struct {
	storable.ResultSet
	last    *MultiKeySortFixture
	lastErr error
}

// All returns all documents on the resultset and close the resultset
func (r *MultiKeySortFixtureResultSet) All() ([]*MultiKeySortFixture, error) {
	var result []*MultiKeySortFixture
	err := r.ResultSet.All(&result)

	return result, err
}

// One returns the first document on the resultset and close the resultset
func (r *MultiKeySortFixtureResultSet) One() (*MultiKeySortFixture, error) {
	var result *MultiKeySortFixture
	err := r.ResultSet.One(&result)

	return result, err
}

// Next prepares the next result document for reading with the Get method.
func (r *MultiKeySortFixtureResultSet) Next() (returned bool) {
	r.last = nil
	returned, r.lastErr = r.ResultSet.Next(&r.last)

	return
}

// Get returns the document retrieved with the Next method.
func (r *MultiKeySortFixtureResultSet) Get() (*MultiKeySortFixture, error) {
	return r.last, r.lastErr
}

// ForEach iterates the resultset calling to the given function.
func (r *MultiKeySortFixtureResultSet) ForEach(f func(*MultiKeySortFixture) error) error {
	for {
		var result *MultiKeySortFixture
		found, err := r.ResultSet.Next(&result)
		if err != nil {
			return err
		}

		if !found {
			break
		}

		err = f(result)
		if err == storable.ErrStop {
			break
		}

		if err != nil {
			return err
		}
	}

	return nil
}

type QueryFixtureStore struct {
	storable.Store
}

func NewQueryFixtureStore(db *mgo.Database) *QueryFixtureStore {
	return &QueryFixtureStore{*storable.NewStore(db, "query")}
}

// New returns a new instance of QueryFixture.
func (s *QueryFixtureStore) New(f string) (doc *QueryFixture) {
	doc = newQueryFixture(f)
	if doc != nil {
		doc.SetIsNew(true)
		doc.SetId(bson.NewObjectId())
	}
	return
}

// Query return a new instance of QueryFixtureQuery.
func (s *QueryFixtureStore) Query() *QueryFixtureQuery {
	return &QueryFixtureQuery{*storable.NewBaseQuery()}
}

// Find performs a find on the collection using the given query.
func (s *QueryFixtureStore) Find(query *QueryFixtureQuery) (*QueryFixtureResultSet, error) {
	resultSet, err := s.Store.Find(query)
	if err != nil {
		return nil, err
	}

	return &QueryFixtureResultSet{ResultSet: *resultSet}, nil
}

// MustFind like Find but panics on error
func (s *QueryFixtureStore) MustFind(query *QueryFixtureQuery) *QueryFixtureResultSet {
	resultSet := s.Store.MustFind(query)
	return &QueryFixtureResultSet{ResultSet: *resultSet}
}

// FindOne performs a find on the collection using the given query returning
// the first document from the resultset.
func (s *QueryFixtureStore) FindOne(query *QueryFixtureQuery) (*QueryFixture, error) {
	resultSet, err := s.Find(query)
	if err != nil {
		return nil, err
	}

	return resultSet.One()
}

// MustFindOne like FindOne but panics on error
func (s *QueryFixtureStore) MustFindOne(query *QueryFixtureQuery) *QueryFixture {
	doc, err := s.FindOne(query)
	if err != nil {
		panic(err)
	}

	return doc
}

// Insert insert the given document on the collection, trigger BeforeInsert and
// AfterInsert if any. Throws ErrNonNewDocument if doc is a non-new document.
func (s *QueryFixtureStore) Insert(doc *QueryFixture) error {

	err := s.Store.Insert(doc)
	if err != nil {
		return err
	}

	return nil
}

// Update update the given document on the collection, trigger BeforeUpdate and
// AfterUpdate if any. Throws ErrNewDocument if doc is a new document.
func (s *QueryFixtureStore) Update(doc *QueryFixture) error {

	err := s.Store.Update(doc)
	if err != nil {
		return err
	}

	return nil
}

// Save insert or update the given document on the collection using Upsert,
// trigger BeforeUpdate and AfterUpdate if the document is non-new and
// BeforeInsert and AfterInset if is new.
func (s *QueryFixtureStore) Save(doc *QueryFixture) (updated bool, err error) {
	updated, err = s.Store.Save(doc)
	if err != nil {
		return false, err
	}

	return
}

type QueryFixtureQuery struct {
	storable.BaseQuery
}

// FindById add a new criteria to the query searching by _id
func (q *QueryFixtureQuery) FindById(ids ...bson.ObjectId) *QueryFixtureQuery {
	var vs []interface{}
	for _, id := range ids {
		vs = append(vs, id)
	}
	q.AddCriteria(operators.In(storable.IdField, vs...))

	return q
}

type QueryFixtureResultSet struct {
	storable.ResultSet
	last    *QueryFixture
	lastErr error
}

// All returns all documents on the resultset and close the resultset
func (r *QueryFixtureResultSet) All() ([]*QueryFixture, error) {
	var result []*QueryFixture
	err := r.ResultSet.All(&result)

	return result, err
}

// One returns the first document on the resultset and close the resultset
func (r *QueryFixtureResultSet) One() (*QueryFixture, error) {
	var result *QueryFixture
	err := r.ResultSet.One(&result)

	return result, err
}

// Next prepares the next result document for reading with the Get method.
func (r *QueryFixtureResultSet) Next() (returned bool) {
	r.last = nil
	returned, r.lastErr = r.ResultSet.Next(&r.last)

	return
}

// Get returns the document retrieved with the Next method.
func (r *QueryFixtureResultSet) Get() (*QueryFixture, error) {
	return r.last, r.lastErr
}

// ForEach iterates the resultset calling to the given function.
func (r *QueryFixtureResultSet) ForEach(f func(*QueryFixture) error) error {
	for {
		var result *QueryFixture
		found, err := r.ResultSet.Next(&result)
		if err != nil {
			return err
		}

		if !found {
			break
		}

		err = f(result)
		if err == storable.ErrStop {
			break
		}

		if err != nil {
			return err
		}
	}

	return nil
}

type ResultSetFixtureStore struct {
	storable.Store
}

func NewResultSetFixtureStore(db *mgo.Database) *ResultSetFixtureStore {
	return &ResultSetFixtureStore{*storable.NewStore(db, "resultset")}
}

// New returns a new instance of ResultSetFixture.
func (s *ResultSetFixtureStore) New(f string) (doc *ResultSetFixture) {
	doc = newResultSetFixture(f)
	if doc != nil {
		doc.SetIsNew(true)
		doc.SetId(bson.NewObjectId())
	}
	return
}

// Query return a new instance of ResultSetFixtureQuery.
func (s *ResultSetFixtureStore) Query() *ResultSetFixtureQuery {
	return &ResultSetFixtureQuery{*storable.NewBaseQuery()}
}

// Find performs a find on the collection using the given query.
func (s *ResultSetFixtureStore) Find(query *ResultSetFixtureQuery) (*ResultSetFixtureResultSet, error) {
	resultSet, err := s.Store.Find(query)
	if err != nil {
		return nil, err
	}

	return &ResultSetFixtureResultSet{ResultSet: *resultSet}, nil
}

// MustFind like Find but panics on error
func (s *ResultSetFixtureStore) MustFind(query *ResultSetFixtureQuery) *ResultSetFixtureResultSet {
	resultSet := s.Store.MustFind(query)
	return &ResultSetFixtureResultSet{ResultSet: *resultSet}
}

// FindOne performs a find on the collection using the given query returning
// the first document from the resultset.
func (s *ResultSetFixtureStore) FindOne(query *ResultSetFixtureQuery) (*ResultSetFixture, error) {
	resultSet, err := s.Find(query)
	if err != nil {
		return nil, err
	}

	return resultSet.One()
}

// MustFindOne like FindOne but panics on error
func (s *ResultSetFixtureStore) MustFindOne(query *ResultSetFixtureQuery) *ResultSetFixture {
	doc, err := s.FindOne(query)
	if err != nil {
		panic(err)
	}

	return doc
}

// Insert insert the given document on the collection, trigger BeforeInsert and
// AfterInsert if any. Throws ErrNonNewDocument if doc is a non-new document.
func (s *ResultSetFixtureStore) Insert(doc *ResultSetFixture) error {

	err := s.Store.Insert(doc)
	if err != nil {
		return err
	}

	return nil
}

// Update update the given document on the collection, trigger BeforeUpdate and
// AfterUpdate if any. Throws ErrNewDocument if doc is a new document.
func (s *ResultSetFixtureStore) Update(doc *ResultSetFixture) error {

	err := s.Store.Update(doc)
	if err != nil {
		return err
	}

	return nil
}

// Save insert or update the given document on the collection using Upsert,
// trigger BeforeUpdate and AfterUpdate if the document is non-new and
// BeforeInsert and AfterInset if is new.
func (s *ResultSetFixtureStore) Save(doc *ResultSetFixture) (updated bool, err error) {
	updated, err = s.Store.Save(doc)
	if err != nil {
		return false, err
	}

	return
}

type ResultSetFixtureQuery struct {
	storable.BaseQuery
}

// FindById add a new criteria to the query searching by _id
func (q *ResultSetFixtureQuery) FindById(ids ...bson.ObjectId) *ResultSetFixtureQuery {
	var vs []interface{}
	for _, id := range ids {
		vs = append(vs, id)
	}
	q.AddCriteria(operators.In(storable.IdField, vs...))

	return q
}

type ResultSetFixtureResultSet struct {
	storable.ResultSet
	last    *ResultSetFixture
	lastErr error
}

// All returns all documents on the resultset and close the resultset
func (r *ResultSetFixtureResultSet) All() ([]*ResultSetFixture, error) {
	var result []*ResultSetFixture
	err := r.ResultSet.All(&result)

	return result, err
}

// One returns the first document on the resultset and close the resultset
func (r *ResultSetFixtureResultSet) One() (*ResultSetFixture, error) {
	var result *ResultSetFixture
	err := r.ResultSet.One(&result)

	return result, err
}

// Next prepares the next result document for reading with the Get method.
func (r *ResultSetFixtureResultSet) Next() (returned bool) {
	r.last = nil
	returned, r.lastErr = r.ResultSet.Next(&r.last)

	return
}

// Get returns the document retrieved with the Next method.
func (r *ResultSetFixtureResultSet) Get() (*ResultSetFixture, error) {
	return r.last, r.lastErr
}

// ForEach iterates the resultset calling to the given function.
func (r *ResultSetFixtureResultSet) ForEach(f func(*ResultSetFixture) error) error {
	for {
		var result *ResultSetFixture
		found, err := r.ResultSet.Next(&result)
		if err != nil {
			return err
		}

		if !found {
			break
		}

		err = f(result)
		if err == storable.ErrStop {
			break
		}

		if err != nil {
			return err
		}
	}

	return nil
}

type ResultSetInitFixtureStore struct {
	storable.Store
}

func NewResultSetInitFixtureStore(db *mgo.Database) *ResultSetInitFixtureStore {
	return &ResultSetInitFixtureStore{*storable.NewStore(db, "resultset")}
}

// New returns a new instance of ResultSetInitFixture.
func (s *ResultSetInitFixtureStore) New() (doc *ResultSetInitFixture) {
	doc = &ResultSetInitFixture{}
	if doc != nil {
		doc.SetIsNew(true)
		doc.SetId(bson.NewObjectId())
	}
	return
}

// Query return a new instance of ResultSetInitFixtureQuery.
func (s *ResultSetInitFixtureStore) Query() *ResultSetInitFixtureQuery {
	return &ResultSetInitFixtureQuery{*storable.NewBaseQuery()}
}

// Find performs a find on the collection using the given query.
func (s *ResultSetInitFixtureStore) Find(query *ResultSetInitFixtureQuery) (*ResultSetInitFixtureResultSet, error) {
	resultSet, err := s.Store.Find(query)
	if err != nil {
		return nil, err
	}

	return &ResultSetInitFixtureResultSet{ResultSet: *resultSet}, nil
}

// MustFind like Find but panics on error
func (s *ResultSetInitFixtureStore) MustFind(query *ResultSetInitFixtureQuery) *ResultSetInitFixtureResultSet {
	resultSet := s.Store.MustFind(query)
	return &ResultSetInitFixtureResultSet{ResultSet: *resultSet}
}

// FindOne performs a find on the collection using the given query returning
// the first document from the resultset.
func (s *ResultSetInitFixtureStore) FindOne(query *ResultSetInitFixtureQuery) (*ResultSetInitFixture, error) {
	resultSet, err := s.Find(query)
	if err != nil {
		return nil, err
	}

	return resultSet.One()
}

// MustFindOne like FindOne but panics on error
func (s *ResultSetInitFixtureStore) MustFindOne(query *ResultSetInitFixtureQuery) *ResultSetInitFixture {
	doc, err := s.FindOne(query)
	if err != nil {
		panic(err)
	}

	return doc
}

// Insert insert the given document on the collection, trigger BeforeInsert and
// AfterInsert if any. Throws ErrNonNewDocument if doc is a non-new document.
func (s *ResultSetInitFixtureStore) Insert(doc *ResultSetInitFixture) error {

	err := s.Store.Insert(doc)
	if err != nil {
		return err
	}

	return nil
}

// Update update the given document on the collection, trigger BeforeUpdate and
// AfterUpdate if any. Throws ErrNewDocument if doc is a new document.
func (s *ResultSetInitFixtureStore) Update(doc *ResultSetInitFixture) error {

	err := s.Store.Update(doc)
	if err != nil {
		return err
	}

	return nil
}

// Save insert or update the given document on the collection using Upsert,
// trigger BeforeUpdate and AfterUpdate if the document is non-new and
// BeforeInsert and AfterInset if is new.
func (s *ResultSetInitFixtureStore) Save(doc *ResultSetInitFixture) (updated bool, err error) {
	updated, err = s.Store.Save(doc)
	if err != nil {
		return false, err
	}

	return
}

type ResultSetInitFixtureQuery struct {
	storable.BaseQuery
}

// FindById add a new criteria to the query searching by _id
func (q *ResultSetInitFixtureQuery) FindById(ids ...bson.ObjectId) *ResultSetInitFixtureQuery {
	var vs []interface{}
	for _, id := range ids {
		vs = append(vs, id)
	}
	q.AddCriteria(operators.In(storable.IdField, vs...))

	return q
}

type ResultSetInitFixtureResultSet struct {
	storable.ResultSet
	last    *ResultSetInitFixture
	lastErr error
}

// All returns all documents on the resultset and close the resultset
func (r *ResultSetInitFixtureResultSet) All() ([]*ResultSetInitFixture, error) {
	var result []*ResultSetInitFixture
	err := r.ResultSet.All(&result)
	if err != nil {
		return result, err
	}

	for _, r := range result {
		if err := r.Init(r); err != nil {
			return result, err
		}
	}

	return result, err
}

// One returns the first document on the resultset and close the resultset
func (r *ResultSetInitFixtureResultSet) One() (*ResultSetInitFixture, error) {
	var result *ResultSetInitFixture
	err := r.ResultSet.One(&result)
	if err != nil {
		return result, err
	}

	err = result.Init(result)

	return result, err
}

// Next prepares the next result document for reading with the Get method.
func (r *ResultSetInitFixtureResultSet) Next() (returned bool) {
	r.last = nil
	returned, r.lastErr = r.ResultSet.Next(&r.last)
	if r.lastErr != nil {
		return
	}
	if returned {
		r.lastErr = r.last.Init(r.last)
	}

	return
}

// Get returns the document retrieved with the Next method.
func (r *ResultSetInitFixtureResultSet) Get() (*ResultSetInitFixture, error) {
	return r.last, r.lastErr
}

// ForEach iterates the resultset calling to the given function.
func (r *ResultSetInitFixtureResultSet) ForEach(f func(*ResultSetInitFixture) error) error {
	for {
		var result *ResultSetInitFixture
		found, err := r.ResultSet.Next(&result)
		if err != nil {
			return err
		}

		if !found {
			break
		}

		if err := result.Init(result); err != nil {
			return err
		}

		err = f(result)
		if err == storable.ErrStop {
			break
		}

		if err != nil {
			return err
		}
	}

	return nil
}

type SchemaFixtureStore struct {
	storable.Store
}

func NewSchemaFixtureStore(db *mgo.Database) *SchemaFixtureStore {
	return &SchemaFixtureStore{*storable.NewStore(db, "schema")}
}

// New returns a new instance of SchemaFixture.
func (s *SchemaFixtureStore) New() (doc *SchemaFixture) {
	doc = &SchemaFixture{}
	if doc != nil {
		doc.SetIsNew(true)
		doc.SetId(bson.NewObjectId())
	}
	return
}

// Query return a new instance of SchemaFixtureQuery.
func (s *SchemaFixtureStore) Query() *SchemaFixtureQuery {
	return &SchemaFixtureQuery{*storable.NewBaseQuery()}
}

// Find performs a find on the collection using the given query.
func (s *SchemaFixtureStore) Find(query *SchemaFixtureQuery) (*SchemaFixtureResultSet, error) {
	resultSet, err := s.Store.Find(query)
	if err != nil {
		return nil, err
	}

	return &SchemaFixtureResultSet{ResultSet: *resultSet}, nil
}

// MustFind like Find but panics on error
func (s *SchemaFixtureStore) MustFind(query *SchemaFixtureQuery) *SchemaFixtureResultSet {
	resultSet := s.Store.MustFind(query)
	return &SchemaFixtureResultSet{ResultSet: *resultSet}
}

// FindOne performs a find on the collection using the given query returning
// the first document from the resultset.
func (s *SchemaFixtureStore) FindOne(query *SchemaFixtureQuery) (*SchemaFixture, error) {
	resultSet, err := s.Find(query)
	if err != nil {
		return nil, err
	}

	return resultSet.One()
}

// MustFindOne like FindOne but panics on error
func (s *SchemaFixtureStore) MustFindOne(query *SchemaFixtureQuery) *SchemaFixture {
	doc, err := s.FindOne(query)
	if err != nil {
		panic(err)
	}

	return doc
}

// Insert insert the given document on the collection, trigger BeforeInsert and
// AfterInsert if any. Throws ErrNonNewDocument if doc is a non-new document.
func (s *SchemaFixtureStore) Insert(doc *SchemaFixture) error {

	err := s.Store.Insert(doc)
	if err != nil {
		return err
	}

	return nil
}

// Update update the given document on the collection, trigger BeforeUpdate and
// AfterUpdate if any. Throws ErrNewDocument if doc is a new document.
func (s *SchemaFixtureStore) Update(doc *SchemaFixture) error {

	err := s.Store.Update(doc)
	if err != nil {
		return err
	}

	return nil
}

// Save insert or update the given document on the collection using Upsert,
// trigger BeforeUpdate and AfterUpdate if the document is non-new and
// BeforeInsert and AfterInset if is new.
func (s *SchemaFixtureStore) Save(doc *SchemaFixture) (updated bool, err error) {
	updated, err = s.Store.Save(doc)
	if err != nil {
		return false, err
	}

	return
}

type SchemaFixtureQuery struct {
	storable.BaseQuery
}

// FindById add a new criteria to the query searching by _id
func (q *SchemaFixtureQuery) FindById(ids ...bson.ObjectId) *SchemaFixtureQuery {
	var vs []interface{}
	for _, id := range ids {
		vs = append(vs, id)
	}
	q.AddCriteria(operators.In(storable.IdField, vs...))

	return q
}

type SchemaFixtureResultSet struct {
	storable.ResultSet
	last    *SchemaFixture
	lastErr error
}

// All returns all documents on the resultset and close the resultset
func (r *SchemaFixtureResultSet) All() ([]*SchemaFixture, error) {
	var result []*SchemaFixture
	err := r.ResultSet.All(&result)

	return result, err
}

// One returns the first document on the resultset and close the resultset
func (r *SchemaFixtureResultSet) One() (*SchemaFixture, error) {
	var result *SchemaFixture
	err := r.ResultSet.One(&result)

	return result, err
}

// Next prepares the next result document for reading with the Get method.
func (r *SchemaFixtureResultSet) Next() (returned bool) {
	r.last = nil
	returned, r.lastErr = r.ResultSet.Next(&r.last)

	return
}

// Get returns the document retrieved with the Next method.
func (r *SchemaFixtureResultSet) Get() (*SchemaFixture, error) {
	return r.last, r.lastErr
}

// ForEach iterates the resultset calling to the given function.
func (r *SchemaFixtureResultSet) ForEach(f func(*SchemaFixture) error) error {
	for {
		var result *SchemaFixture
		found, err := r.ResultSet.Next(&result)
		if err != nil {
			return err
		}

		if !found {
			break
		}

		err = f(result)
		if err == storable.ErrStop {
			break
		}

		if err != nil {
			return err
		}
	}

	return nil
}

type StoreFixtureStore struct {
	storable.Store
}

func NewStoreFixtureStore(db *mgo.Database) *StoreFixtureStore {
	return &StoreFixtureStore{*storable.NewStore(db, "store")}
}

// New returns a new instance of StoreFixture.
func (s *StoreFixtureStore) New() (doc *StoreFixture) {
	doc = &StoreFixture{}
	if doc != nil {
		doc.SetIsNew(true)
		doc.SetId(bson.NewObjectId())
	}
	return
}

// Query return a new instance of StoreFixtureQuery.
func (s *StoreFixtureStore) Query() *StoreFixtureQuery {
	return &StoreFixtureQuery{*storable.NewBaseQuery()}
}

// Find performs a find on the collection using the given query.
func (s *StoreFixtureStore) Find(query *StoreFixtureQuery) (*StoreFixtureResultSet, error) {
	resultSet, err := s.Store.Find(query)
	if err != nil {
		return nil, err
	}

	return &StoreFixtureResultSet{ResultSet: *resultSet}, nil
}

// MustFind like Find but panics on error
func (s *StoreFixtureStore) MustFind(query *StoreFixtureQuery) *StoreFixtureResultSet {
	resultSet := s.Store.MustFind(query)
	return &StoreFixtureResultSet{ResultSet: *resultSet}
}

// FindOne performs a find on the collection using the given query returning
// the first document from the resultset.
func (s *StoreFixtureStore) FindOne(query *StoreFixtureQuery) (*StoreFixture, error) {
	resultSet, err := s.Find(query)
	if err != nil {
		return nil, err
	}

	return resultSet.One()
}

// MustFindOne like FindOne but panics on error
func (s *StoreFixtureStore) MustFindOne(query *StoreFixtureQuery) *StoreFixture {
	doc, err := s.FindOne(query)
	if err != nil {
		panic(err)
	}

	return doc
}

// Insert insert the given document on the collection, trigger BeforeInsert and
// AfterInsert if any. Throws ErrNonNewDocument if doc is a non-new document.
func (s *StoreFixtureStore) Insert(doc *StoreFixture) error {

	err := s.Store.Insert(doc)
	if err != nil {
		return err
	}

	return nil
}

// Update update the given document on the collection, trigger BeforeUpdate and
// AfterUpdate if any. Throws ErrNewDocument if doc is a new document.
func (s *StoreFixtureStore) Update(doc *StoreFixture) error {

	err := s.Store.Update(doc)
	if err != nil {
		return err
	}

	return nil
}

// Save insert or update the given document on the collection using Upsert,
// trigger BeforeUpdate and AfterUpdate if the document is non-new and
// BeforeInsert and AfterInset if is new.
func (s *StoreFixtureStore) Save(doc *StoreFixture) (updated bool, err error) {
	updated, err = s.Store.Save(doc)
	if err != nil {
		return false, err
	}

	return
}

type StoreFixtureQuery struct {
	storable.BaseQuery
}

// FindById add a new criteria to the query searching by _id
func (q *StoreFixtureQuery) FindById(ids ...bson.ObjectId) *StoreFixtureQuery {
	var vs []interface{}
	for _, id := range ids {
		vs = append(vs, id)
	}
	q.AddCriteria(operators.In(storable.IdField, vs...))

	return q
}

type StoreFixtureResultSet struct {
	storable.ResultSet
	last    *StoreFixture
	lastErr error
}

// All returns all documents on the resultset and close the resultset
func (r *StoreFixtureResultSet) All() ([]*StoreFixture, error) {
	var result []*StoreFixture
	err := r.ResultSet.All(&result)

	return result, err
}

// One returns the first document on the resultset and close the resultset
func (r *StoreFixtureResultSet) One() (*StoreFixture, error) {
	var result *StoreFixture
	err := r.ResultSet.One(&result)

	return result, err
}

// Next prepares the next result document for reading with the Get method.
func (r *StoreFixtureResultSet) Next() (returned bool) {
	r.last = nil
	returned, r.lastErr = r.ResultSet.Next(&r.last)

	return
}

// Get returns the document retrieved with the Next method.
func (r *StoreFixtureResultSet) Get() (*StoreFixture, error) {
	return r.last, r.lastErr
}

// ForEach iterates the resultset calling to the given function.
func (r *StoreFixtureResultSet) ForEach(f func(*StoreFixture) error) error {
	for {
		var result *StoreFixture
		found, err := r.ResultSet.Next(&result)
		if err != nil {
			return err
		}

		if !found {
			break
		}

		err = f(result)
		if err == storable.ErrStop {
			break
		}

		if err != nil {
			return err
		}
	}

	return nil
}

type StoreWithConstructFixtureStore struct {
	storable.Store
}

func NewStoreWithConstructFixtureStore(db *mgo.Database) *StoreWithConstructFixtureStore {
	return &StoreWithConstructFixtureStore{*storable.NewStore(db, "store_construct")}
}

// New returns a new instance of StoreWithConstructFixture.
func (s *StoreWithConstructFixtureStore) New(f string) (doc *StoreWithConstructFixture) {
	doc = newStoreWithConstructFixture(f)
	if doc != nil {
		doc.SetIsNew(true)
		doc.SetId(bson.NewObjectId())
	}
	return
}

// Query return a new instance of StoreWithConstructFixtureQuery.
func (s *StoreWithConstructFixtureStore) Query() *StoreWithConstructFixtureQuery {
	return &StoreWithConstructFixtureQuery{*storable.NewBaseQuery()}
}

// Find performs a find on the collection using the given query.
func (s *StoreWithConstructFixtureStore) Find(query *StoreWithConstructFixtureQuery) (*StoreWithConstructFixtureResultSet, error) {
	resultSet, err := s.Store.Find(query)
	if err != nil {
		return nil, err
	}

	return &StoreWithConstructFixtureResultSet{ResultSet: *resultSet}, nil
}

// MustFind like Find but panics on error
func (s *StoreWithConstructFixtureStore) MustFind(query *StoreWithConstructFixtureQuery) *StoreWithConstructFixtureResultSet {
	resultSet := s.Store.MustFind(query)
	return &StoreWithConstructFixtureResultSet{ResultSet: *resultSet}
}

// FindOne performs a find on the collection using the given query returning
// the first document from the resultset.
func (s *StoreWithConstructFixtureStore) FindOne(query *StoreWithConstructFixtureQuery) (*StoreWithConstructFixture, error) {
	resultSet, err := s.Find(query)
	if err != nil {
		return nil, err
	}

	return resultSet.One()
}

// MustFindOne like FindOne but panics on error
func (s *StoreWithConstructFixtureStore) MustFindOne(query *StoreWithConstructFixtureQuery) *StoreWithConstructFixture {
	doc, err := s.FindOne(query)
	if err != nil {
		panic(err)
	}

	return doc
}

// Insert insert the given document on the collection, trigger BeforeInsert and
// AfterInsert if any. Throws ErrNonNewDocument if doc is a non-new document.
func (s *StoreWithConstructFixtureStore) Insert(doc *StoreWithConstructFixture) error {

	err := s.Store.Insert(doc)
	if err != nil {
		return err
	}

	return nil
}

// Update update the given document on the collection, trigger BeforeUpdate and
// AfterUpdate if any. Throws ErrNewDocument if doc is a new document.
func (s *StoreWithConstructFixtureStore) Update(doc *StoreWithConstructFixture) error {

	err := s.Store.Update(doc)
	if err != nil {
		return err
	}

	return nil
}

// Save insert or update the given document on the collection using Upsert,
// trigger BeforeUpdate and AfterUpdate if the document is non-new and
// BeforeInsert and AfterInset if is new.
func (s *StoreWithConstructFixtureStore) Save(doc *StoreWithConstructFixture) (updated bool, err error) {
	updated, err = s.Store.Save(doc)
	if err != nil {
		return false, err
	}

	return
}

type StoreWithConstructFixtureQuery struct {
	storable.BaseQuery
}

// FindById add a new criteria to the query searching by _id
func (q *StoreWithConstructFixtureQuery) FindById(ids ...bson.ObjectId) *StoreWithConstructFixtureQuery {
	var vs []interface{}
	for _, id := range ids {
		vs = append(vs, id)
	}
	q.AddCriteria(operators.In(storable.IdField, vs...))

	return q
}

type StoreWithConstructFixtureResultSet struct {
	storable.ResultSet
	last    *StoreWithConstructFixture
	lastErr error
}

// All returns all documents on the resultset and close the resultset
func (r *StoreWithConstructFixtureResultSet) All() ([]*StoreWithConstructFixture, error) {
	var result []*StoreWithConstructFixture
	err := r.ResultSet.All(&result)

	return result, err
}

// One returns the first document on the resultset and close the resultset
func (r *StoreWithConstructFixtureResultSet) One() (*StoreWithConstructFixture, error) {
	var result *StoreWithConstructFixture
	err := r.ResultSet.One(&result)

	return result, err
}

// Next prepares the next result document for reading with the Get method.
func (r *StoreWithConstructFixtureResultSet) Next() (returned bool) {
	r.last = nil
	returned, r.lastErr = r.ResultSet.Next(&r.last)

	return
}

// Get returns the document retrieved with the Next method.
func (r *StoreWithConstructFixtureResultSet) Get() (*StoreWithConstructFixture, error) {
	return r.last, r.lastErr
}

// ForEach iterates the resultset calling to the given function.
func (r *StoreWithConstructFixtureResultSet) ForEach(f func(*StoreWithConstructFixture) error) error {
	for {
		var result *StoreWithConstructFixture
		found, err := r.ResultSet.Next(&result)
		if err != nil {
			return err
		}

		if !found {
			break
		}

		err = f(result)
		if err == storable.ErrStop {
			break
		}

		if err != nil {
			return err
		}
	}

	return nil
}

type StoreWithNewFixtureStore struct {
	storable.Store
}

func NewStoreWithNewFixtureStore(db *mgo.Database) *StoreWithNewFixtureStore {
	return &StoreWithNewFixtureStore{*storable.NewStore(db, "store_new")}
}

// Query return a new instance of StoreWithNewFixtureQuery.
func (s *StoreWithNewFixtureStore) Query() *StoreWithNewFixtureQuery {
	return &StoreWithNewFixtureQuery{*storable.NewBaseQuery()}
}

// Find performs a find on the collection using the given query.
func (s *StoreWithNewFixtureStore) Find(query *StoreWithNewFixtureQuery) (*StoreWithNewFixtureResultSet, error) {
	resultSet, err := s.Store.Find(query)
	if err != nil {
		return nil, err
	}

	return &StoreWithNewFixtureResultSet{ResultSet: *resultSet}, nil
}

// MustFind like Find but panics on error
func (s *StoreWithNewFixtureStore) MustFind(query *StoreWithNewFixtureQuery) *StoreWithNewFixtureResultSet {
	resultSet := s.Store.MustFind(query)
	return &StoreWithNewFixtureResultSet{ResultSet: *resultSet}
}

// FindOne performs a find on the collection using the given query returning
// the first document from the resultset.
func (s *StoreWithNewFixtureStore) FindOne(query *StoreWithNewFixtureQuery) (*StoreWithNewFixture, error) {
	resultSet, err := s.Find(query)
	if err != nil {
		return nil, err
	}

	return resultSet.One()
}

// MustFindOne like FindOne but panics on error
func (s *StoreWithNewFixtureStore) MustFindOne(query *StoreWithNewFixtureQuery) *StoreWithNewFixture {
	doc, err := s.FindOne(query)
	if err != nil {
		panic(err)
	}

	return doc
}

// Insert insert the given document on the collection, trigger BeforeInsert and
// AfterInsert if any. Throws ErrNonNewDocument if doc is a non-new document.
func (s *StoreWithNewFixtureStore) Insert(doc *StoreWithNewFixture) error {

	err := s.Store.Insert(doc)
	if err != nil {
		return err
	}

	return nil
}

// Update update the given document on the collection, trigger BeforeUpdate and
// AfterUpdate if any. Throws ErrNewDocument if doc is a new document.
func (s *StoreWithNewFixtureStore) Update(doc *StoreWithNewFixture) error {

	err := s.Store.Update(doc)
	if err != nil {
		return err
	}

	return nil
}

// Save insert or update the given document on the collection using Upsert,
// trigger BeforeUpdate and AfterUpdate if the document is non-new and
// BeforeInsert and AfterInset if is new.
func (s *StoreWithNewFixtureStore) Save(doc *StoreWithNewFixture) (updated bool, err error) {
	updated, err = s.Store.Save(doc)
	if err != nil {
		return false, err
	}

	return
}

type StoreWithNewFixtureQuery struct {
	storable.BaseQuery
}

// FindById add a new criteria to the query searching by _id
func (q *StoreWithNewFixtureQuery) FindById(ids ...bson.ObjectId) *StoreWithNewFixtureQuery {
	var vs []interface{}
	for _, id := range ids {
		vs = append(vs, id)
	}
	q.AddCriteria(operators.In(storable.IdField, vs...))

	return q
}

type StoreWithNewFixtureResultSet struct {
	storable.ResultSet
	last    *StoreWithNewFixture
	lastErr error
}

// All returns all documents on the resultset and close the resultset
func (r *StoreWithNewFixtureResultSet) All() ([]*StoreWithNewFixture, error) {
	var result []*StoreWithNewFixture
	err := r.ResultSet.All(&result)

	return result, err
}

// One returns the first document on the resultset and close the resultset
func (r *StoreWithNewFixtureResultSet) One() (*StoreWithNewFixture, error) {
	var result *StoreWithNewFixture
	err := r.ResultSet.One(&result)

	return result, err
}

// Next prepares the next result document for reading with the Get method.
func (r *StoreWithNewFixtureResultSet) Next() (returned bool) {
	r.last = nil
	returned, r.lastErr = r.ResultSet.Next(&r.last)

	return
}

// Get returns the document retrieved with the Next method.
func (r *StoreWithNewFixtureResultSet) Get() (*StoreWithNewFixture, error) {
	return r.last, r.lastErr
}

// ForEach iterates the resultset calling to the given function.
func (r *StoreWithNewFixtureResultSet) ForEach(f func(*StoreWithNewFixture) error) error {
	for {
		var result *StoreWithNewFixture
		found, err := r.ResultSet.Next(&result)
		if err != nil {
			return err
		}

		if !found {
			break
		}

		err = f(result)
		if err == storable.ErrStop {
			break
		}

		if err != nil {
			return err
		}
	}

	return nil
}

type schema struct {
	EventsFixture             *schemaEventsFixture
	EventsSaveFixture         *schemaEventsSaveFixture
	MultiKeySortFixture       *schemaMultiKeySortFixture
	QueryFixture              *schemaQueryFixture
	ResultSetFixture          *schemaResultSetFixture
	ResultSetInitFixture      *schemaResultSetInitFixture
	SchemaFixture             *schemaSchemaFixture
	StoreFixture              *schemaStoreFixture
	StoreWithConstructFixture *schemaStoreWithConstructFixture
	StoreWithNewFixture       *schemaStoreWithNewFixture
}

type schemaEventsFixture struct {
	Checks storable.Map
}

type schemaEventsSaveFixture struct {
	Checks storable.Map
}

type schemaMultiKeySortFixture struct {
	Name  storable.Field
	Start storable.Field
	End   storable.Field
}

type schemaQueryFixture struct {
	Foo storable.Field
}

type schemaResultSetFixture struct {
	Foo storable.Field
}

type schemaResultSetInitFixture struct {
	Foo storable.Field
}

type schemaSchemaFixture struct {
	String         storable.Field
	Int            storable.Field
	Nested         *schemaSchemaFixtureNested
	Inline         *schemaSchemaFixtureInline
	MapOfString    storable.Map
	MapOfInterface storable.Map
	MapOfSomeType  *schemaSchemaFixtureMapOfSomeType
}

type schemaStoreFixture struct {
	Foo storable.Field
}

type schemaStoreWithConstructFixture struct {
	Foo storable.Field
}

type schemaStoreWithNewFixture struct {
	Foo storable.Field
	Bar storable.Field
}

type schemaSchemaFixtureNested struct {
	String         storable.Field
	Int            storable.Field
	Nested         *schemaSchemaFixtureNestedNested
	Inline         *schemaSchemaFixtureNestedInline
	MapOfString    storable.Map
	MapOfInterface storable.Map
	MapOfSomeType  *schemaSchemaFixtureNestedMapOfSomeType
}

type schemaSchemaFixtureInline struct {
	Inline storable.Field
}

type schemaSchemaFixtureMapOfSomeType struct {
	Foo storable.Map
}

type schemaSchemaFixtureNestedNested struct {
}

type schemaSchemaFixtureNestedInline struct {
	Inline storable.Field
}

type schemaSchemaFixtureNestedMapOfSomeType struct {
	Foo storable.Map
}

var Schema = schema{
	EventsFixture: &schemaEventsFixture{
		Checks: storable.NewMap("checks.[map]", "bool"),
	},
	EventsSaveFixture: &schemaEventsSaveFixture{
		Checks: storable.NewMap("checks.[map]", "bool"),
	},
	MultiKeySortFixture: &schemaMultiKeySortFixture{
		Name:  storable.NewField("name", "string"),
		Start: storable.NewField("start", "time.Time"),
		End:   storable.NewField("end", "time.Time"),
	},
	QueryFixture: &schemaQueryFixture{
		Foo: storable.NewField("foo", "string"),
	},
	ResultSetFixture: &schemaResultSetFixture{
		Foo: storable.NewField("foo", "string"),
	},
	ResultSetInitFixture: &schemaResultSetInitFixture{
		Foo: storable.NewField("foo", "string"),
	},
	SchemaFixture: &schemaSchemaFixture{
		String: storable.NewField("string", "string"),
		Int:    storable.NewField("foo", "int"),
		Nested: &schemaSchemaFixtureNested{
			String: storable.NewField("nested.string", "string"),
			Int:    storable.NewField("nested.foo", "int"),
			Nested: &schemaSchemaFixtureNestedNested{},
			Inline: &schemaSchemaFixtureNestedInline{
				Inline: storable.NewField("nested.inline", "string"),
			},
			MapOfString:    storable.NewMap("nested.mapofstring.[map]", "string"),
			MapOfInterface: storable.NewMap("nested.mapofinterface.[map]", "interface{}"),
			MapOfSomeType: &schemaSchemaFixtureNestedMapOfSomeType{
				Foo: storable.NewMap("nested.mapofsometype.[map].foo", "string"),
			},
		},
		Inline: &schemaSchemaFixtureInline{
			Inline: storable.NewField("inline", "string"),
		},
		MapOfString:    storable.NewMap("mapofstring.[map]", "string"),
		MapOfInterface: storable.NewMap("mapofinterface.[map]", "interface{}"),
		MapOfSomeType: &schemaSchemaFixtureMapOfSomeType{
			Foo: storable.NewMap("mapofsometype.[map].foo", "string"),
		},
	},
	StoreFixture: &schemaStoreFixture{
		Foo: storable.NewField("foo", "string"),
	},
	StoreWithConstructFixture: &schemaStoreWithConstructFixture{
		Foo: storable.NewField("foo", "string"),
	},
	StoreWithNewFixture: &schemaStoreWithNewFixture{
		Foo: storable.NewField("foo", "string"),
		Bar: storable.NewField("bar", "string"),
	},
}
