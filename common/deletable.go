package common

type Deletable interface {
	Delete() error
	Name() string
	Type() string
}

var CriticalFilter = [6]string{"prod", "live", "qa", "ci.k8s.sgr-labs.com.", "k8s.sgr-cloud.sh.", "dev.k8s.sgr-cloud.sh."}

var Debug bool = false
