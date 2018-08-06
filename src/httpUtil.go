/*
Package main is the main implementation of the Bitly serverless app and retrieves statistics on the various bitlinks associated with the account of the authenticated user
*/
package main

// The imports
import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

// httpRequest executes a request to a URL and returns the response body as a JSON object
func httpRequest(URL string, authToken string) (map[string]interface{}, error) {

	req, err := http.NewRequest("GET", URL, nil)
	if err != nil {
		return nil, fmt.Errorf("error while creating HTTP request: %s", err.Error())
	}
	req.Header.Add("authorization", fmt.Sprintf("Bearer %s", authToken))

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
