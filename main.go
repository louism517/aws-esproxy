package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/louism517/aws-esproxy/creds"
	"github.com/louism517/aws-esproxy/esproxy"
	"github.com/namsral/flag"
)

func main() {

	var esUrl, region, arn, session, port string
	var debug bool

	flag.StringVar(&esUrl, "esurl", "", "URL of AWS Elasticsearch endpoint")
	flag.StringVar(&region, "region", "us-east-1", "AWS region to use when assuming STS role.")
	flag.StringVar(&arn, "arn", "", "ARN of role to be assumed. If ommitted then the default credential chain is used instead.")
	flag.StringVar(&session, "session", "", "Session name to be used when assuming STS role.")
	flag.StringVar(&port, "port", "8080", "Port to serve proxy on.")
	flag.BoolVar(&debug, "debug", false, "Run in Debug mode.")

	flag.Parse()

	if esUrl == "" {
		fmt.Println("Please supply an endpoint URL (through ESURL environment variable or --esurl flag):")
		flag.PrintDefaults()
		os.Exit(1)
	}
	u, err := url.Parse(esUrl)
	fmt.Println(u)
	if err != nil {
		log.Fatalf("Unable to parse URL: %s\n", err)
	}

	if arn != "" {
		c, err := creds.NewStsCredentialGetter(&creds.StsCredentialGetterConfig{
			Region:  region,
			Arn:     arn,
			Session: session,
		})
		if err != nil {
			log.Fatal(err)
		}

		http.ListenAndServe(fmt.Sprintf(":%s", port), esproxy.NewReverseAWSProxy(u, c, debug))
	} else {
		// Use the default credential chain by default
		c, err := creds.NewChainCredentialGetter()
		if err != nil {
			log.Fatal(err)
		}

		http.ListenAndServe(fmt.Sprintf(":%s", port), esproxy.NewReverseAWSProxy(u, c, debug))
	}

}
