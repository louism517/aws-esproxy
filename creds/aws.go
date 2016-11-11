package creds

import (
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/ec2rolecreds"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
)

type AwsCredentialGetter interface {
	GetCreds() (credentials.Value, error)
}

type StsCredentialGetterConfig struct {
	Region  string
	Arn     string
	Session string
}

type StsCredentialGetter struct {
	Credentials *sts.Credentials
	Config      *StsCredentialGetterConfig
	session     *session.Session
}

func NewStsCredentialGetter(config *StsCredentialGetterConfig) (*StsCredentialGetter, error) {
	c := &StsCredentialGetter{Config: config}
	c.session = session.New(&aws.Config{Region: aws.String(c.Config.Region)})
	return c, c.updateStsCredentials()
}

func (c *StsCredentialGetter) updateStsCredentials() error {
	svc := sts.New(c.session)
	params := &sts.AssumeRoleInput{
		RoleArn:         aws.String(c.Config.Arn),
		RoleSessionName: aws.String(c.Config.Session),
		DurationSeconds: aws.Int64(3600),
	}
	resp, err := svc.AssumeRole(params)
	if err != nil {
		return err
	}
	c.Credentials = resp.Credentials
	return nil
}

func (c *StsCredentialGetter) GetCreds() (credentials.Value, error) {
	if time.Now().Unix()-c.Credentials.Expiration.Unix() > 0 {
		c.updateStsCredentials()
	}
	ak := *c.Credentials.AccessKeyId
	sak := *c.Credentials.SecretAccessKey
	st := *c.Credentials.SessionToken
	return credentials.Value{
		AccessKeyID:     ak,
		SecretAccessKey: sak,
		SessionToken:    st,
	}, nil

}

type ChainCredentialGetter struct {
	Credentials *credentials.Credentials
}

func NewChainCredentialGetter() (*ChainCredentialGetter, error) {
	creds := credentials.NewChainCredentials(
		[]credentials.Provider{
			&credentials.EnvProvider{},
			&credentials.SharedCredentialsProvider{},
			&ec2rolecreds.EC2RoleProvider{},
		})

	c := &ChainCredentialGetter{Credentials: creds}
	return c, nil
}

func (c ChainCredentialGetter) GetCreds() (credentials.Value, error) {
	return c.Credentials.Get()
}
