package eks

import (
	"fmt"
	"runtime"
	"strings"

	awseks "github.com/aws/aws-sdk-go/service/eks"
	"github.com/genevieve/leftovers/common"
)

//go:generate faux --interface clustersClient --output fakes/clusters_client.go
type clustersClient interface {
	ListClusters(*awseks.ListClustersInput) (*awseks.ListClustersOutput, error)
	DeleteCluster(*awseks.DeleteClusterInput) (*awseks.DeleteClusterOutput, error)
}

//go:generate faux --interface logger --output fakes/logger.go
type logger interface {
	Printf(m string, a ...interface{})
	PromptWithDetails(resourceType, resourceName string) (proceed bool)
}

type Clusters struct {
	client clustersClient
	logger logger
}

func NewClusters(client clustersClient, logger logger) Clusters {
	return Clusters{
		client: client,
		logger: logger,
	}
}

func (c Clusters) List(filter string) ([]common.Deletable, error) {
	clusters, err := c.client.ListClusters(&awseks.ListClustersInput{})
	if err != nil {
		return nil, fmt.Errorf("List EKS Clusters: %s", err)
	}

	var resources []common.Deletable
	for _, cluster := range clusters.Clusters {
		r := NewCluster(c.client, cluster)

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

		proceed := c.logger.PromptWithDetails(r.Type(), r.Name())
		if !proceed {
			continue
		}

		resources = append(resources, r)
	}

	return resources, nil
}

func (c Clusters) Type() string {
	return "eks-cluster"
}
