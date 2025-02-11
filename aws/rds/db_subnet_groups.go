package rds

import (
	"fmt"
	"runtime"
	"strings"

	awsrds "github.com/aws/aws-sdk-go/service/rds"
	"github.com/genevieve/leftovers/common"
)

//go:generate faux --interface dbSubnetGroupsClient --output fakes/db_subnet_groups_client.go
type dbSubnetGroupsClient interface {
	DescribeDBSubnetGroups(*awsrds.DescribeDBSubnetGroupsInput) (*awsrds.DescribeDBSubnetGroupsOutput, error)
	DeleteDBSubnetGroup(*awsrds.DeleteDBSubnetGroupInput) (*awsrds.DeleteDBSubnetGroupOutput, error)
}

type DBSubnetGroups struct {
	client dbSubnetGroupsClient
	logger logger
}

func NewDBSubnetGroups(client dbSubnetGroupsClient, logger logger) DBSubnetGroups {
	return DBSubnetGroups{
		client: client,
		logger: logger,
	}
}

func (d DBSubnetGroups) List(filter string) ([]common.Deletable, error) {
	dbSubnetGroups, err := d.client.DescribeDBSubnetGroups(&awsrds.DescribeDBSubnetGroupsInput{})
	if err != nil {
		return nil, fmt.Errorf("Describing RDS DB Subnet Groups: %s", err)
	}

	var resources []common.Deletable
	for _, db := range dbSubnetGroups.DBSubnetGroups {
		r := NewDBSubnetGroup(d.client, db.DBSubnetGroupName)

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

		proceed := d.logger.PromptWithDetails(r.Type(), r.Name())
		if !proceed {
			continue
		}

		resources = append(resources, r)
	}

	return resources, nil
}

func (d DBSubnetGroups) Type() string {
	return "rds-db-subnet-group"
}
