package iam

import (
	"fmt"
	"runtime"
	"strings"

	awsiam "github.com/aws/aws-sdk-go/service/iam"
	"github.com/genevieve/leftovers/common"
)

//go:generate faux --interface serverCertificatesClient --output fakes/server_certificates_client.go
type serverCertificatesClient interface {
	ListServerCertificates(*awsiam.ListServerCertificatesInput) (*awsiam.ListServerCertificatesOutput, error)
	DeleteServerCertificate(*awsiam.DeleteServerCertificateInput) (*awsiam.DeleteServerCertificateOutput, error)
}

type ServerCertificates struct {
	client serverCertificatesClient
	logger logger
}

func NewServerCertificates(client serverCertificatesClient, logger logger) ServerCertificates {
	return ServerCertificates{
		client: client,
		logger: logger,
	}
}

func (s ServerCertificates) List(filter string) ([]common.Deletable, error) {
	certificates, err := s.client.ListServerCertificates(&awsiam.ListServerCertificatesInput{})
	if err != nil {
		return nil, fmt.Errorf("List IAM Server Certificates: %s", err)
	}

	var resources []common.Deletable
	for _, c := range certificates.ServerCertificateMetadataList {
		r := NewServerCertificate(s.client, c.ServerCertificateName)

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

		proceed := s.logger.PromptWithDetails(r.Type(), r.Name())
		if !proceed {
			continue
		}

		resources = append(resources, r)
	}

	return resources, nil
}

func (s ServerCertificates) Type() string {
	return "iam-server-certificate"
}
