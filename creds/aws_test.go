package creds

import (
	"bytes"
	"encoding/xml"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/private/protocol/xml/xmlutil"
	"github.com/aws/aws-sdk-go/service/sts"
)

var dummyConfig = &StsCredentialGetterConfig{
	Region:  "us-east-1",
	Arn:     "testing:arn:blah:blah",
	Session: "testSession",
}

var dummyCreds = &sts.Credentials{
	AccessKeyId:     aws.String("IAMADUMMYIAMACCESSKEY"),
	SecretAccessKey: aws.String("MadeUpSecretAccessKey"),
	Expiration:      aws.Time(time.Now().Add(3 * time.Second)),
	SessionToken:    aws.String("MadeupToken"),
}
var dummyAssumedUser = &sts.AssumedRoleUser{
	Arn:           aws.String("made:up:arn:1233453337:something:"),
	AssumedRoleId: aws.String("madeUpRole"),
}

var dummyOutput = &sts.AssumeRoleOutput{
	AssumedRoleUser:  dummyAssumedUser,
	Credentials:      dummyCreds,
	PackedPolicySize: aws.Int64(99),
}

type XMLResponse struct {
	AssumeRoleResponse *XMLResult
}

type XMLResult struct {
	AssumeRoleResult *sts.AssumeRoleOutput
}

var marshalMe = &XMLResponse{AssumeRoleResponse: &XMLResult{AssumeRoleResult: dummyOutput}}

func TestStsCredentialGetter(t *testing.T) {
	sess := session.New()
	sess.Handlers.Send.Clear()
	sess.Handlers.Send.PushFront(func(r *request.Request) {
		arn := r.Params.(*sts.AssumeRoleInput).RoleArn
		sess := r.Params.(*sts.AssumeRoleInput).RoleSessionName
		if *arn != dummyConfig.Arn {
			t.Errorf("Assume role ARN expected: %s, got %s", dummyConfig.Arn, *arn)
		}
		if *sess != dummyConfig.Session {
			t.Errorf("Assume role session name expected: %s, got %s", dummyConfig.Session, *sess)
		}
		var buf bytes.Buffer
		err := xmlutil.BuildXML(marshalMe, xml.NewEncoder(&buf))
		if err != nil {
			panic(err)
		}
		r.HTTPResponse = &http.Response{
			StatusCode: 200,
			Body:       ioutil.NopCloser(bytes.NewReader(buf.Bytes())),
			// this for UnmarshalMetaHandler
			Header: http.Header{"X-Amzn-Requestid": []string{"12345254232"}},
		}

	})
	s := &StsCredentialGetter{
		Config:  dummyConfig,
		session: sess,
	}
	err := s.updateStsCredentials()
	if err != nil {
		t.Fatalf("Error refreshing credentials", err)
	}
	creds, err := s.GetCreds()
	if err != nil {
		t.Fatalf("Error running GetCreds()", err)
	}
	if creds.AccessKeyID != *dummyCreds.AccessKeyId {
		t.Errorf("GetCreds() Access Key expected: %s, got %s", *dummyCreds.AccessKeyId, creds.AccessKeyID)
	}

	// Credentials should not be re-requested until the Expiration time has been reached.
	// Expiration is set to 3 seconds in the future
	dummyCreds.AccessKeyId = aws.String("IAMADUMMYIAMACCESSKEY-NEW")
	creds, err = s.GetCreds()
	if err != nil {
		t.Fatalf("Error running GetCreds()", err)
	}

	if creds.AccessKeyID == *dummyCreds.AccessKeyId {
		t.Error("GetCreds renewed credentials before expiration.")
	}

	// Sleep 4 seconds and request creds again. This time they should renew.
	time.Sleep(4 * time.Second)
	creds, err = s.GetCreds()
	if err != nil {
		t.Fatalf("Error running GetCreds()", err)
	}

	if creds.AccessKeyID != *dummyCreds.AccessKeyId {
		t.Error("GetCreds failed to renew credentials after expiration.")
	}

}
