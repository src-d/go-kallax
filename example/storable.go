package example

import (
	"time"

	"github.com/src-d/go-kallax"
	"github.com/src-d/go-kallax/operators"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type ProductStore struct {
	kallax.Store
}

func NewProductStore(db *mgo.Database) *ProductStore {
	return &ProductStore{*kallax.NewStore(db, "products")}
}

// New returns a new instance of Product.
func (s *ProductStore) New(name string, price Price, createdAt time.Time) (doc *Product, err error) {
	doc, err = newProduct(name, price, createdAt)
	if doc != nil {
		doc.SetIsNew(true)
		doc.SetId(bson.NewObjectId())
	}
	return
}

// Query return a new instance of ProductQuery.
func (s *ProductStore) Query() *ProductQuery {
	return &ProductQuery{*kallax.NewBaseQuery()}
}

// Find performs a find on the collection using the given query.
func (s *ProductStore) Find(query *ProductQuery) (*ProductResultSet, error) {
	resultSet, err := s.Store.Find(query)
	if err != nil {
		return nil, err
	}

	return &ProductResultSet{ResultSet: *resultSet}, nil
}

// MustFind like Find but panics on error
func (s *ProductStore) MustFind(query *ProductQuery) *ProductResultSet {
	resultSet := s.Store.MustFind(query)
	return &ProductResultSet{ResultSet: *resultSet}
}

// FindOne performs a find on the collection using the given query returning
// the first document from the resultset.
func (s *ProductStore) FindOne(query *ProductQuery) (*Product, error) {
	resultSet, err := s.Find(query)
	if err != nil {
		return nil, err
	}

	return resultSet.One()
}

// MustFindOne like FindOne but panics on error
func (s *ProductStore) MustFindOne(query *ProductQuery) *Product {
	doc, err := s.FindOne(query)
	if err != nil {
		panic(err)
	}

	return doc
}

// Insert insert the given document on the collection, trigger BeforeInsert and
// AfterInsert if any. Throws ErrNonNewDocument if doc is a non-new document.
func (s *ProductStore) Insert(doc *Product) error {

	err := s.Store.Insert(doc)
	if err != nil {
		return err
	}

	return nil
}

// Update update the given document on the collection, trigger BeforeUpdate and
// AfterUpdate if any. Throws ErrNewDocument if doc is a new document.
func (s *ProductStore) Update(doc *Product) error {

	err := s.Store.Update(doc)
	if err != nil {
		return err
	}

	return nil
}

// Save insert or update the given document on the collection using Upsert,
// trigger BeforeUpdate and AfterUpdate if the document is non-new and
// BeforeInsert and AfterInset if is new.
func (s *ProductStore) Save(doc *Product) (updated bool, err error) {
	updated, err = s.Store.Save(doc)
	if err != nil {
		return false, err
	}

	return
}

type ProductQuery struct {
	kallax.BaseQuery
}

// FindById add a new criteria to the query searching by _id
func (q *ProductQuery) FindById(ids ...bson.ObjectId) *ProductQuery {
	var vs []interface{}
	for _, id := range ids {
		vs = append(vs, id)
	}
	q.AddCriteria(operators.In(kallax.IdField, vs...))

	return q
}

type ProductResultSet struct {
	kallax.ResultSet
	last    *Product
	lastErr error
}

// All returns all documents on the resultset and close the resultset
func (r *ProductResultSet) All() ([]*Product, error) {
	var result []*Product
	err := r.ResultSet.All(&result)

	return result, err
}

// One returns the first document on the resultset and close the resultset
func (r *ProductResultSet) One() (*Product, error) {
	var result *Product
	err := r.ResultSet.One(&result)

	return result, err
}

// Next prepares the next result document for reading with the Get method.
func (r *ProductResultSet) Next() (returned bool) {
	r.last = nil
	returned, r.lastErr = r.ResultSet.Next(&r.last)

	return
}

// Get returns the document retrieved with the Next method.
func (r *ProductResultSet) Get() (*Product, error) {
	return r.last, r.lastErr
}

// ForEach iterates the resultset calling to the given function.
func (r *ProductResultSet) ForEach(f func(*Product) error) error {
	for {
		var result *Product
		found, err := r.ResultSet.Next(&result)
		if err != nil {
			return err
		}

		if !found {
			break
		}

		err = f(result)
		if err == kallax.ErrStop {
			break
		}

		if err != nil {
			return err
		}
	}

	return nil
}

type schema struct {
	Product *schemaProduct
}

type schemaProduct struct {
	Status    kallax.Field
	CreatedAt kallax.Field
	UpdatedAt kallax.Field
	Name      kallax.Field
	Price     *schemaProductPrice
	Discount  kallax.Field
	Url       kallax.Field
	Tags      kallax.Field
}

type schemaProductPrice struct {
	Amount   kallax.Field
	Discount kallax.Field
}

var Schema = schema{
	Product: &schemaProduct{
		Status:    kallax.NewField("status", "int"),
		CreatedAt: kallax.NewField("createdat", "time.Time"),
		UpdatedAt: kallax.NewField("updatedat", "time.Time"),
		Name:      kallax.NewField("name", "string"),
		Price: &schemaProductPrice{
			Amount:   kallax.NewField("price.amount", "float64"),
			Discount: kallax.NewField("price.discount", "float64"),
		},
		Discount: kallax.NewField("discount", "float64"),
		Url:      kallax.NewField("url", "string"),
		Tags:     kallax.NewField("tags", "string"),
	},
}
