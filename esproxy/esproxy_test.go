package esproxy

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"

	"github.com/louism517/aws-esproxy/creds"
)

func TestReverseAWSProxy(t *testing.T) {
	os.Setenv("AWS_ACCESS_KEY", "ADUMMYACCESSSKEY")
	os.Setenv("AWS_SECRET_ACCESS_KEY", "ADUMMYSECRETACCESSSKEY1234")

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := r.Header["Authorization"]; !ok {
			t.Fatalf("Proxied request, expected Authorization header.")
		}
	}))
	defer ts.Close()

	c, err := creds.NewChainCredentialGetter()
	if err != nil {
		t.Fatalf("Failed to create NewChainCredentialGetter")
	}
	u, _ := url.Parse(ts.URL)
	p := NewReverseAWSProxy(u, c, false)
	go http.ListenAndServe("localhost:63443", p)
	req, _ := http.NewRequest("GET", "http://localhost:63443", nil)
	http.DefaultClient.Do(req)

}
