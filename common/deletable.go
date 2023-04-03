package common

type Deletable interface {
	Delete() error
	Name() string
	Type() string
}

var CriticalFilter = [5]string{"prod", "live", "qa", "stage", "master"}

var Debug bool = false
