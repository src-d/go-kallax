package example

import (
	"errors"
	"log"
	"time"

	"github.com/src-d/go-kallax"
)

//go:generate kallax gen

type Product struct {
	kallax.Document `bson:",inline" collection:"products"`

	Status    Status
	CreatedAt time.Time
	UpdatedAt time.Time
	Name      string
	Price     Price

	Discount float64
	Url      string
	Tags     []string
}

func newProduct(name string, price Price, createdAt time.Time) (*Product, error) {
	if len(name) == 0 {
		return nil, errors.New("name should not be empty.")
	}
	return &Product{
		Name:      name,
		Price:     price,
		CreatedAt: createdAt,
		Status:    Draft,
	}, nil
}

func (p *Product) BeforeInsert() error {
	p.CreatedAt = time.Now()
	return nil
}

func (p *Product) BeforeSave() error {
	p.UpdatedAt = time.Now()
	return nil
}

type Status int

const (
	Draft Status = iota
	Published
)

func (s *Status) BeforeInsert() error {
	log.Println("Status before insert:", s)
	return nil
}

func (s Status) AfterInsert() error {
	log.Println("Status after insert:", s)
	return nil
}

type Price struct {
	Amount   float64
	Discount float64
}
