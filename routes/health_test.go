package routes_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/maleck13/scm-go/test"
)

func TestHealth(t *testing.T) {
	server := test.SetUpIntegrationServer()
	defer server.Close()
	healthUrl := fmt.Sprintf("%s/health", server.URL)

	request, err := http.NewRequest("GET", healthUrl, nil)

	res, err := http.DefaultClient.Do(request)

	handleFail(t, err)

	if res.StatusCode != 200 {
		t.Errorf("Success expected: %d", res.StatusCode)
	}

	con, err := ioutil.ReadAll(res.Body)
	handleFail(t, err)
	var jMap map[string]string
	json.Unmarshal(con, &jMap)

	i, ok := jMap["http"]
	if !ok {
		handleFail(t, errors.New("no health key returned"))
	}
	if "ok" != i {
		handleFail(t, errors.New("value for health should be ok"))
	}

}

func handleFail(t *testing.T, err error) {
	if err != nil {
		t.Error(err) //Something is wrong while sending request
	}
}
