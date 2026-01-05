package utils

import (
  "context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)



func AccountAndRegion(ctx context.Context, cfg aws.Config) (string, string, error) {
	stsClient := sts.NewFromConfig(cfg)

	id, err := stsClient.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		return "", "", err
	}

	return *id.Account, cfg.Region, nil
}
