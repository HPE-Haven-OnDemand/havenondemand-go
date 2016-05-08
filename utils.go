package hod

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
)

// Creates a new file upload http request with optional extra params
func NewFileUploadRequest(uri string, params map[string]string, paramName, path string) (*http.Request, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile(paramName, filepath.Base(path))
	if err != nil {
		return nil, err
	}
	_, err = io.Copy(part, file)

	for key, val := range params {
		_ = writer.WriteField(key, val)
	}
	err = writer.Close()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", uri, body)
	if err != nil {
		return nil, err
	}

	// Don't forget to set the content type, this will contain the boundary.
	req.Header.Set("Content-Type", writer.FormDataContentType())

	return req, nil
}

// b is a byte array containing JSON
// Returns the value of a given field
// Note that this doesn't work with nested fields
func GetJsonField(b []byte, field string) (string, error) {
	var f interface{}
	err := json.Unmarshal(b, &f)
	if err != nil {
		return "", err
	}
	m := f.(map[string]interface{})
	if val, ok := m["error"]; ok {
		if val2, ok2 := m["reason"]; ok2 {
			return "", errors.New(fmt.Sprintf("error: %d reason: %s", val, val2))
		} else {
			return "", errors.New(fmt.Sprintf("error: %d", val))
		}
	} else {
		// no error
		if v, ok := m[field].(string); ok {
			return v, nil
		} else {
			return "", errors.New(fmt.Sprintf("Unable to extract field %s", field))
		}
	}
}
