package main

import (
	"errors"

	"github.com/genevieve/leftovers/app"
	"github.com/genevieve/leftovers/aws"
)

type leftovers interface {
	Delete(filter string) error
	DeleteByType(filter, rType string) error
	List(filter string)
	ListByType(filter, rType string)
	Types()
}

func GetLeftovers(logger *app.Logger, o app.Options) (leftovers, error) {
	var (
		l   leftovers
		err error
	)

	switch o.IAAS {
	case app.AWS:
		l, err = aws.NewLeftovers(logger, o.AWSAccessKeyID, o.AWSSecretAccessKey, o.AWSSessionToken, o.AWSRegion)
	default:
		err = errors.New("Missing or unsupported BBL_IAAS.")
	}

	if err != nil {
		return nil, err
	}
	return l, nil
}
