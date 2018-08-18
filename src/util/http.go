// Package util implements utility methods
package util

// The imports
import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

// HTTPRequest executes a request to a URL and returns the response body as a JSON object
func HTTPRequest(URL string, header http.Header) (map[string]interface{}, error) {

	req, err := http.NewRequest("GET", URL, nil)
	if err != nil {
		return nil, fmt.Errorf("error while creating HTTP request: %s", err.Error())
	}

	if header != nil {
		req.Header = header
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error while performing HTTP request: %s", err.Error())
	}

	defer res.Body.Close()

	if res.StatusCode != 200 {
		return nil, fmt.Errorf("the HTTP request returned a non-OK response: %v", res.StatusCode)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var data map[string]interface{}

	if err := json.Unmarshal(body, &data); err != nil {
		return nil, fmt.Errorf("error while unmarshaling HTTP response to JSON: %s", err.Error())
	}

	return data, nil
}
