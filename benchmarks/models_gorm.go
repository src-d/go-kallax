package benchmark

type GORMPerson struct {
	ID   int64 `gorm:"primary_key"`
	Name string
	Pets []*GORMPet `gorm:"ForeignKey:PersonID"`
}

func (GORMPerson) TableName() string {
	return "people"
}

type GORMPet struct {
	ID       int64 `gorm:"primary_key"`
	PersonID int64
	Name     string
	Kind     string
}

func (GORMPet) TableName() string {
	return "pets"
}
