package elbv2

import (
	"fmt"
	"runtime"
	"strings"

	awselbv2 "github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/genevieve/leftovers/common"
)

//go:generate faux --interface targetGroupsClient --output fakes/target_groups_client.go
type targetGroupsClient interface {
	DescribeTargetGroups(*awselbv2.DescribeTargetGroupsInput) (*awselbv2.DescribeTargetGroupsOutput, error)
	DeleteTargetGroup(*awselbv2.DeleteTargetGroupInput) (*awselbv2.DeleteTargetGroupOutput, error)
}

type TargetGroups struct {
	client targetGroupsClient
	logger logger
}

func NewTargetGroups(client targetGroupsClient, logger logger) TargetGroups {
	return TargetGroups{
		client: client,
		logger: logger,
	}
}

func (t TargetGroups) List(filter string) ([]common.Deletable, error) {
	targetGroups, err := t.client.DescribeTargetGroups(&awselbv2.DescribeTargetGroupsInput{})
	if err != nil {
		return nil, fmt.Errorf("Describe ELBV2 Target Groups: %s", err)
	}

	var resources []common.Deletable
	for _, g := range targetGroups.TargetGroups {
		r := NewTargetGroup(t.client, g.TargetGroupName, g.TargetGroupArn)

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

		proceed := t.logger.PromptWithDetails(r.Type(), r.Name())
		if !proceed {
			continue
		}

		resources = append(resources, r)
	}

	return resources, nil
}

func (t TargetGroups) Type() string {
	return "elbv2-target-group"
}
