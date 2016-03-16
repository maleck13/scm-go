package routes_test

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/maleck13/scm-go/Godeps/_workspace/src/github.com/fheng/scm-go/test"
	"github.com/maleck13/scm-go/Godeps/_workspace/src/github.com/stretchr/testify/assert"
)

func TestPingOk(t *testing.T) {
	server := test.SetUpIntegrationServer()
	defer server.Close()
	url := fmt.Sprintf("%s/sys/info/ping", server.URL)
	res, err := http.Get(url)
	assert.NoError(t, err, "did not expect an error")
	content, err := ioutil.ReadAll(res.Body)
	assert.NoError(t, err, "did not expect an error")
	check := string(content)
	assert.EqualValues(t, "OK", check, "expected an ok from ping")
}
