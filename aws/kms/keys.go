package kms

import (
	"fmt"
	"runtime"
	"strings"

	awskms "github.com/aws/aws-sdk-go/service/kms"
	"github.com/genevieve/leftovers/common"
)

//go:generate faux --interface keysClient --output fakes/keys_client.go
type keysClient interface {
	ListKeys(*awskms.ListKeysInput) (*awskms.ListKeysOutput, error)
	DescribeKey(*awskms.DescribeKeyInput) (*awskms.DescribeKeyOutput, error)
	ListResourceTags(*awskms.ListResourceTagsInput) (*awskms.ListResourceTagsOutput, error)
	DisableKey(*awskms.DisableKeyInput) (*awskms.DisableKeyOutput, error)
	ScheduleKeyDeletion(*awskms.ScheduleKeyDeletionInput) (*awskms.ScheduleKeyDeletionOutput, error)
}

type Keys struct {
	client keysClient
	logger logger
}

func NewKeys(client keysClient, logger logger) Keys {
	return Keys{
		client: client,
		logger: logger,
		//TODO: Add	resourceTags
	}
}

func (k Keys) List(filter string) ([]common.Deletable, error) {
	keys, err := k.client.ListKeys(&awskms.ListKeysInput{})
	if err != nil {
		return nil, fmt.Errorf("Listing KMS Keys: %s", err)
	}

	var resources []common.Deletable
	for _, key := range keys.Keys {
		metadata, _ := k.client.DescribeKey(&awskms.DescribeKeyInput{KeyId: key.KeyId})
		if metadata == nil || *metadata.KeyMetadata.KeyState != awskms.KeyStateEnabled {
			continue
		}

		tags, _ := k.client.ListResourceTags(&awskms.ListResourceTagsInput{KeyId: key.KeyId})

		r := NewKey(k.client, key.KeyId, metadata.KeyMetadata, tags.Tags)

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

		proceed := k.logger.PromptWithDetails(r.Type(), r.Name())
		if !proceed {
			continue
		}

		resources = append(resources, r)
	}

	return resources, nil
}

func (k Keys) Type() string {
	return "kms-key"
}
