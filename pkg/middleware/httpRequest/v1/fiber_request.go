package httpRequestV1

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

func SendRequest(uri, method string, pathParams []string, body []byte, headers map[string]string, queryParam map[string]any, timeout int) (any, error) {
	// Construct the final URL with path parameters
	finalUrl := strings.TrimRight(uri, "/")
	if len(pathParams) > 0 {
		finalUrl += "/" + strings.Join(pathParams, "/")
	}

	// Construct query parameters
	params := url.Values{}
	for qkey, qvalue := range queryParam {
		params.Add(qkey, fmt.Sprint(qvalue))
	}

	if len(params) > 0 {
		finalUrl += "?" + params.Encode()
	}

	// Debug: Print the final URL
	fmt.Println("Final URL:", finalUrl)
	reqBody := bytes.NewBuffer(body)

	// Send the HTTP request using a helper function
	req, reqErr := http.NewRequest(method, finalUrl, reqBody)
	if reqErr != nil {
		return nil, reqErr
	}

	// Set default content-type if not provided
	if _, exists := headers["Content-Type"]; !exists {
		req.Header.Set("Content-Type", "application/json")
	}

	// Add user-defined headers
	for hkey, hvalue := range headers {
		req.Header.Set(hkey, hvalue)
	}

	client := &http.Client{Timeout: time.Duration(timeout) * time.Second}

	resp, respErr := client.Do(req)
	if respErr != nil {
		return nil, respErr
	}
	defer resp.Body.Close()

	// Read the response body
	respBody, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return nil, readErr
	}

	// Parse the response as JSON Object
	var jsonResponseObject map[string]any
	if err := json.Unmarshal(respBody, &jsonResponseObject); err == nil {
		return jsonResponseObject, nil
	}

	// If paarsing as JSON Object fails, parse as JSON Array
	var jsonResponseArray []any
	if err := json.Unmarshal(respBody, &jsonResponseArray); err == nil {
		return jsonResponseArray, nil
	}

	// If both parsing attempts fail, return the raw response body as a string
	//return nil, fmt.Errorf("response is neither a JSON object nor a JSON array: %s", string(body))
	// Return respBody so the caller can still use the raw response
	return string(respBody), fmt.Errorf("response is neither a JSON object nor a JSON array: %s", string(respBody))
}
