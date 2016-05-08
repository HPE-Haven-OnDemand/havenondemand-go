package hod

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

const (
	HOD_API_URL        = "http://api.havenondemand.com/1/api"
	HOD_JOB_RESULT_URL = "http://api.havenondemand.com/1/job/result"
	HOD_JOB_STATUS_URL = "http://api.havenondemand.com/1/job/status"
)

// HODClient manages communication with the Haven On Demand (HOD) API.
type HODClient struct {
	HttpClient *http.Client
	ApiKey     string
	ApiVersion string
}

// NewHodClient returns a new HOD API client.
// apiKey is the api key used to authenticate against the HOD API
// If httpClient is nil we use the default client, this is useful to use a custom http.Client
func NewHODClient(apiKey, apiVersion string, httpClient *http.Client) *HODClient {
	if httpClient == nil {
		httpClient = &http.Client{}
	}
	c := &HODClient{
		ApiKey:     apiKey,
		ApiVersion: apiVersion,
	}
	c.HttpClient = httpClient
	return c
}

// Perform a GET call to the API
// op is the HOD API (e.g. querytextindex or ocrdocuemnt)
// params is a k/v map (e.g. foo=bar, foo2=bar2)
// if async is true the call to the api is async (see the Future type below)
// it returns the JSON response or an error is anything goes wrong
func (c *HODClient) Get(op string, params url.Values, async bool) (string, error) {
	return c.do(op, "GET", params, nil, async)
}

// Defines which data goes in the POST request
// File is the name of the file whose contents will be included
// Data will be included as-is in the POST body
// You should define File or Data but not both
type PostData struct {
	File string
	Data string
}

// Perform a POST call to the API
// op is the HOD API (e.g. querytextindex or ocrdocuemnt)
// postData defines which data goes in the POST request
// if async is true the call to the api is async (see the Future type below)
// it returns the JSON response or an error is anything goes wrong
func (c *HODClient) Post(op string, postData *PostData, async bool) (string, error){
	return c.do(op, "POST", url.Values{}, postData, async)
}

// do is a helper function that sends a GET or POST request to the HOD API
// it returns the JSON response or an error is anything goes wrong
func (c *HODClient) do(op, method string, params url.Values, postData *PostData, async bool) (string, error){
	var mode string
	if async {
		mode = "async"
	} else {
		mode = "sync"
	}
	u, err := url.Parse(fmt.Sprintf("%s/%s/%s/%s", HOD_API_URL, mode, op, c.ApiVersion))
	if err != nil {
		return "", err
	}

	var req *http.Request
	var err2 error
	if method == "POST" {
		if postData.Data != "" {
			req, err2 = http.NewRequest("POST", u.String(), bytes.NewBuffer([]byte(postData.Data)))
			if err2 != nil {
				return "", err
			}

		} else if postData.File != "" {
			params := map[string]string{
				"apikey": c.ApiKey,
			}
			req, err2 = NewFileUploadRequest(u.String(), params, "file", postData.File)
			if err2 != nil {
				return "", err
			}
		} else {
			return "", errors.New("Need to specify either data of file")
		}
	} else if method == "GET" {
		params.Add("apikey", c.ApiKey)
		u.RawQuery = params.Encode()
		req, err2 = http.NewRequest("GET", u.String(), nil)
	} else {
		return "", errors.New(fmt.Sprintf("Unsupported method %s", method))
	}

	rsp, err := c.HttpClient.Do(req)
	defer rsp.Body.Close()
	if err != nil {
		return  "", err
	}
	contents, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		return "", err
	}
	if async {
		// get JobId
		jobId, err := GetJsonField(contents, "jobID")
		if err != nil {
			return "", err
		}
		f := NewFuture(jobId, c.ApiKey, nil)
		res, err := f.Result()
		if err != nil {
			return "", nil
		}
		return res, nil
	} else {
		return string(contents), nil
	}
}

// Future encapsulates the result of an async operation
type Future struct {
	JobId string
	HttpClient *http.Client
	ApiKey string
}

// NewHodClient returns a new Future
// If httpClient is nil we use the default client, this is useful to use a custom http.Client
func NewFuture(jobId, apiKey string, httpClient *http.Client) *Future {
	if httpClient == nil {
		httpClient = &http.Client{}
	}
	f := &Future{JobId: jobId, ApiKey: apiKey}
	f.HttpClient = httpClient
	return f
}

// Result polls the API until the job is completed (either succeeds or fails)
// It uses exponential backoff to not overwhelm the API
// It returns the JSON result or an error if anything goes wrong
func (f *Future) Result() (string, error) {
	b := &Backoff{
	    //These are the defaults
	    Min:    100 * time.Millisecond,
	    Max:    5 * time.Second,
	    Factor: 2,
	    Jitter: false,
	}
	u, err := url.Parse(fmt.Sprintf("%s/%s", HOD_JOB_STATUS_URL, f.JobId))
	if err != nil {
		return "", err
	}
	params := url.Values{}
	params.Add("apikey", f.ApiKey)
	u.RawQuery = params.Encode()
	for {
		rsp, err := f.HttpClient.Get(u.String())
		defer rsp.Body.Close()
		if err != nil {
			return "", err
		}
		contents, err := ioutil.ReadAll(rsp.Body)
		if err != nil {
			return "", err
		}
		status, err := GetJsonField(contents, "status")
		if err != nil {
			return "", err
		}
		switch status {
		case "finished":
			return string(contents), nil
		case "queued":
		// keep trying
		case "in progress":
		// keep trying
		case "failed":
			return string(contents), nil
		default:
			return "", errors.New(fmt.Sprintf("Unknown status %s", status))
		}
		d := b.Duration()
        	time.Sleep(d)
	}
}
