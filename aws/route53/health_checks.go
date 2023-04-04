package route53

import (
	"fmt"
	"runtime"
	"strings"

	awsroute53 "github.com/aws/aws-sdk-go/service/route53"
	"github.com/genevieve/leftovers/common"
)

//go:generate faux --interface healthChecksClient --output fakes/health_checks_client.go
type healthChecksClient interface {
	ListHealthChecks(*awsroute53.ListHealthChecksInput) (*awsroute53.ListHealthChecksOutput, error)
	DeleteHealthCheck(*awsroute53.DeleteHealthCheckInput) (*awsroute53.DeleteHealthCheckOutput, error)
}

type HealthChecks struct {
	client healthChecksClient
	logger logger
}

func NewHealthChecks(client healthChecksClient, logger logger) HealthChecks {
	return HealthChecks{
		client: client,
		logger: logger,
	}
}

func (h HealthChecks) List(filter string) ([]common.Deletable, error) {
	checks, err := h.client.ListHealthChecks(&awsroute53.ListHealthChecksInput{})
	if err != nil {
		return nil, fmt.Errorf("List Route53 Health Checks: %s", err)
	}

	var resources []common.Deletable
	for _, check := range checks.HealthChecks {
		r := NewHealthCheck(h.client, check.Id)
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

		proceed := h.logger.PromptWithDetails(r.Type(), r.Name())
		if !proceed {
			continue
		}

		resources = append(resources, r)
	}

	return resources, nil
}

func (h HealthChecks) Type() string {
	return "route53-health-check"
}
