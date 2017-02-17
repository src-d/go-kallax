package fixtures

import "database/sql/driver"

type AliasArray [3]string
type AliasSlice []string
type AliasString string
type AliasInt int
type AliasArrAliasSlice []AliasSlice
type AliasArrAliasString []AliasString
type AliasDummyParam QueryDummy

type QueryDummy struct {
	name string
}

type InterfaceImplementation struct {
	ScannerValuer
	Str string
}

type ScannerValuer struct{}

func (i ScannerValuer) Value() (driver.Value, error) {
	return nil, nil
}

func (i ScannerValuer) Scan(src interface{}) error {
	return nil
}
