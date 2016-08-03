package esproxy

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/louism517/aws-esproxy/creds"
	"github.com/smartystreets/go-aws-auth"
)

func NewReverseAWSProxy(target *url.URL, creds creds.AwsCredentialGetter, debug bool) *httputil.ReverseProxy {
	targetQuery := target.RawQuery
	director := func(req *http.Request) {
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
		req.URL.Path = singleJoiningSlash(target.Path, req.URL.Path)
		req.Host = target.Host
		if targetQuery == "" || req.URL.RawQuery == "" {
			req.URL.RawQuery = targetQuery + req.URL.RawQuery
		} else {
			req.URL.RawQuery = targetQuery + "&" + req.URL.RawQuery
		}
		awsCreds, err := creds.GetCreds()
		if err != nil {
			log.Println(err)
		}
		awsauth.Sign4(req, awsauth.Credentials{
			AccessKeyID:     awsCreds.AccessKeyID,
			SecretAccessKey: awsCreds.SecretAccessKey,
			SecurityToken:   awsCreds.SessionToken,
		})

		if debug == true {
			d, _ := httputil.DumpRequestOut(req, true)
			fmt.Printf("%s\n", d)
		}
	}

	return &httputil.ReverseProxy{Director: director}
}

func singleJoiningSlash(a, b string) string {
	aslash := strings.HasSuffix(a, "/")
	bslash := strings.HasPrefix(b, "/")
	switch {
	case aslash && bslash:
		return a + b[1:]
	case !aslash && !bslash:
		return a + "/" + b
	}
	return a + b
}
