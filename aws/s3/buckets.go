package s3

import (
	"fmt"
	"runtime"
	"strings"

	awss3 "github.com/aws/aws-sdk-go/service/s3"
	"github.com/genevieve/leftovers/common"
)

//go:generate faux --interface bucketsClient --output fakes/buckets_client.go
type bucketsClient interface {
	ListBuckets(*awss3.ListBucketsInput) (*awss3.ListBucketsOutput, error)
	DeleteBucket(*awss3.DeleteBucketInput) (*awss3.DeleteBucketOutput, error)

	ListObjectVersions(*awss3.ListObjectVersionsInput) (*awss3.ListObjectVersionsOutput, error)
	DeleteObjects(*awss3.DeleteObjectsInput) (*awss3.DeleteObjectsOutput, error)
}

type Buckets struct {
	client  bucketsClient
	logger  logger
	manager bucketManager
}

func NewBuckets(client bucketsClient, logger logger, manager bucketManager) Buckets {
	return Buckets{
		client:  client,
		logger:  logger,
		manager: manager,
	}
}

func (b Buckets) List(filter string) ([]common.Deletable, error) {
	buckets, err := b.client.ListBuckets(&awss3.ListBucketsInput{})
	if err != nil {
		return nil, fmt.Errorf("Listing S3 Buckets: %s", err)
	}

	var resources []common.Deletable
	for _, bucket := range buckets.Buckets {
		r := NewBucket(b.client, bucket.Name)

		if !strings.Contains(r.Name(), filter) {
			continue
		}

		if !b.manager.IsInRegion(r.Name()) {
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

		proceed := b.logger.PromptWithDetails(r.Type(), r.Name())
		if !proceed {
			continue
		}

		resources = append(resources, r)
	}

	return resources, nil
}

func (b Buckets) Type() string {
	return "s3-bucket"
}
