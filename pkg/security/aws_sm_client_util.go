package security

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/kameshsampath/balloon-popper/pkg/logger"
	"os"
)

var client *secretsmanager.Client

// InitAndGetAWSSecretManagerClient initializes the AWS Secret Manager Client
// Acts as Singleton method
func InitAndGetAWSSecretManagerClient() (*secretsmanager.Client, error) {
	if client == nil {
		var awsRegion string
		var err error
		if v, ok := os.LookupEnv("AWS_REGION"); ok {
			awsRegion = v
		} else {
			awsRegion = "us-west-2"
		}
		// Load AWS configuration
		cfg, err := config.LoadDefaultConfig(context.Background(),
			config.WithRegion(awsRegion),
		)
		if err != nil {
			return nil, err
		}
		// Initialize the Secrets Manager client
		client = secretsmanager.NewFromConfig(cfg)
		logger.Get().Infof("Using AWS Region: %s", awsRegion)
	}

	return client, nil
}
