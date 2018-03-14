package yoti

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
)

type httpResponse struct {
	Success    bool
	StatusCode int
	Content    string
}

type httpRequester func(uri string, headers map[string]string, httpRequestMethod string, contentBytes []byte) (result *httpResponse, err error)

func doRequest(uri string, headers map[string]string, httpRequestMethod string, contentBytes []byte) (result *httpResponse, err error) {
	client := &http.Client{}

	supportedHTTPMethods := map[string]bool{"GET": true, "POST": true, "PUT": true, "PATCH": true}

	if !supportedHTTPMethods[httpRequestMethod] {
		err = fmt.Errorf("HTTP Method: '%s' is unsupported", httpRequestMethod)
	}

	var req *http.Request
	if req, err = http.NewRequest(
		httpRequestMethod,
		uri,
		bytes.NewBuffer(contentBytes)); err != nil {
		return
	}

	for key, value := range headers {
		req.Header.Add(key, value)
	}

	var resp *http.Response
	resp, err = client.Do(req)

	if err != nil {
		return
	}

	defer resp.Body.Close()

	var responseBody []byte
	if responseBody, err = ioutil.ReadAll(resp.Body); err != nil {
		return
	}

	result = &httpResponse{
		Success:    resp.StatusCode < 300,
		StatusCode: resp.StatusCode,
		Content:    string(responseBody)}

	return
}
