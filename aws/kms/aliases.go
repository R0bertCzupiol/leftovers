package kms

import (
	"fmt"
	"runtime"
	"strings"

	awskms "github.com/aws/aws-sdk-go/service/kms"
	"github.com/genevieve/leftovers/common"
)

//go:generate faux --interface aliasesClient --output fakes/aliases_client.go
type aliasesClient interface {
	ListAliases(*awskms.ListAliasesInput) (*awskms.ListAliasesOutput, error)
	DeleteAlias(*awskms.DeleteAliasInput) (*awskms.DeleteAliasOutput, error)
}

type Aliases struct {
	client aliasesClient
	logger logger
}

func NewAliases(client aliasesClient, logger logger) Aliases {
	return Aliases{
		client: client,
		logger: logger,
	}
}

func (a Aliases) List(filter string) ([]common.Deletable, error) {
	aliases, err := a.client.ListAliases(&awskms.ListAliasesInput{})
	if err != nil {
		return nil, fmt.Errorf("Listing KMS Aliases: %s", err)
	}

	var resources []common.Deletable
	for _, alias := range aliases.Aliases {
		r := NewAlias(a.client, alias.AliasName)

		if !strings.Contains(r.Name(), filter) {
			continue
		}

		var check = false
		for _, element := range common.CriticalFilter {
			if strings.Contains(r.Name(), element) {
				check = true
				_, file, _, _ := runtime.Caller(1)
				if common.Debug {
					println(file + " skipped value by CriticalFilter: " + r.Name())
				}
			}
		}
		if check {
			continue
		}

		proceed := a.logger.PromptWithDetails(r.Type(), r.Name())
		if !proceed {
			continue
		}

		resources = append(resources, r)
	}

	return resources, nil
}

func (a Aliases) Type() string {
	return "kms-alias"
}
