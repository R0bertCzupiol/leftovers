package elb

import (
	"fmt"
	"runtime"
	"strings"

	awselb "github.com/aws/aws-sdk-go/service/elb"
	"github.com/genevieve/leftovers/common"
)

//go:generate faux --interface loadBalancersClient --output fakes/load_balancers_client.go
type loadBalancersClient interface {
	DescribeLoadBalancers(*awselb.DescribeLoadBalancersInput) (*awselb.DescribeLoadBalancersOutput, error)
	DeleteLoadBalancer(*awselb.DeleteLoadBalancerInput) (*awselb.DeleteLoadBalancerOutput, error)

	DescribeTags(*awselb.DescribeTagsInput) (*awselb.DescribeTagsOutput, error)
}

type LoadBalancers struct {
	client loadBalancersClient
	logger logger
}

func NewLoadBalancers(client loadBalancersClient, logger logger) LoadBalancers {
	return LoadBalancers{
		client: client,
		logger: logger,
	}
}

func (l LoadBalancers) List(filter string) ([]common.Deletable, error) {
	loadBalancers, err := l.client.DescribeLoadBalancers(&awselb.DescribeLoadBalancersInput{})
	if err != nil {
		return nil, fmt.Errorf("Describe ELB Load Balancers: %s", err)
	}

	var resources []common.Deletable
	for _, lb := range loadBalancers.LoadBalancerDescriptions {
		r := NewLoadBalancer(l.client, lb.LoadBalancerName)

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

		proceed := l.logger.PromptWithDetails(r.Type(), r.Name())
		if !proceed {
			continue
		}

		resources = append(resources, r)
	}

	return resources, nil
}

func (l LoadBalancers) Type() string {
	return "elb-load-balancer"
}
