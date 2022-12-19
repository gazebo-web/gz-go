package gztest

// Important note: functions in this module should NOT include
// references to parent package 'ign', to avoid circular dependencies.
// These functions should be independent.

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

var router http.Handler

// FileDesc describes a file to be created. It is used by
// func CreateTmpFolderWithContents and sendMultipartPOST.
// Fields:
// path: is the file path to be sent in the multipart form.
// contents: is the string contents to write in the file. Note: if contents
// value is ":dir" then a Directory will be created instead of a File. This is only
// valid when used with CreateTmpFolderWithContents func.
type FileDesc struct {
	Path     string
	Contents string
}

// SetupTest - Setup helper function
func SetupTest(_router http.Handler) {
	router = _router
}

// SendMultipartPOST executes a multipart POST request with the given form
// fields and multipart files, and returns the received http status code,
// the response body, and a success flag.
func SendMultipartPOST(testName string, t *testing.T, uri string, jwt *string,
	params map[string]string, files []FileDesc) (respCode int,
	bslice *[]byte, ok bool) {

	return SendMultipartMethod(testName, t, "POST", uri, jwt, params, files)
}

// SendMultipartMethod executes a multipart POST/PUT/PATCH request with the given form
// fields and multipart files, and returns the received http status code,
// the response body, and a success flag.
func SendMultipartMethod(testName string, t *testing.T, method, uri string, jwt *string,
	params map[string]string, files []FileDesc) (respCode int,
	bslice *[]byte, ok bool) {

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	for _, fd := range files {
		// Remove base path
		part, err := writer.CreateFormFile("file", fd.Path)
		assert.NoError(t, err, "Could not create FormFile. TestName:[%s]. fd.Path:[%s]", testName, fd.Path)
		_, err = io.WriteString(part, fd.Contents)
	}

	for key, val := range params {
		_ = writer.WriteField(key, val)
	}
	assert.NoError(t, writer.Close(), "Could not close multipart form writer. TestName:[%s]", testName)

	req, err := http.NewRequest(method, uri, body)
	assert.NoError(t, err, "Could not create POST request. TestName:[%s]", testName)
	// Adds the "Content-Type: multipart/form-data" header.
	req.Header.Add("Content-Type", writer.FormDataContentType())
	if jwt != nil {
		// Add the authorization token
		req.Header.Set("Authorization", "Bearer "+*jwt)
	}

	// Process the request
	respRec := httptest.NewRecorder()
	router.ServeHTTP(respRec, req)
	// Process results
	respCode = respRec.Code

	var b []byte
	var er error
	b, er = ioutil.ReadAll(respRec.Body)
	assert.NoError(t, er, "Failed to read the server response. TestName:[%s]", testName)

	bslice = &b
	ok = true
	return
}

// CreateTmpFolderWithContents creates a tmp folder with the given files and
// returns the path to the created folder. See type fileDesc above.
func CreateTmpFolderWithContents(folderName string, files []FileDesc) (string, error) {
	baseDir, err := ioutil.TempDir("", folderName)
	if err != nil {
		return "", err
	}

	for _, fd := range files {
		fullpath := filepath.Join(baseDir, fd.Path)
		dir := filepath.Dir(fullpath)
		if dir != baseDir {
			if err := os.MkdirAll(dir, os.ModePerm); err != nil {
				return "", err
			}
		}

		if fd.Contents == ":dir" {
			// folder
			if err := os.MkdirAll(fullpath, os.ModePerm); err != nil {
				return "", err
			}
		} else {
			// normal file with given contents
			f, err := os.Create(fullpath)
			defer f.Close()
			if err != nil {
				log.Println("Unable to create [" + fullpath + "]")
				return "", err
			}
			if _, err := f.WriteString(fd.Contents); err != nil {
				log.Println("Unable to write contents to [" + fullpath + "]")
				return "", err
			}
			f.Sync()
		}
	}
	return baseDir, nil
}

// AssertRoute is a helper function that checks for a valid route
// \param[in] method One of "GET", "PATCH", "PUT", "POST", "DELETE", "OPTIONS"
// \param[in] route The URL string
// \param[in] code The expected result HTTP code
// \param[in] t Testing pointer
// \return[out] *[]byte A pointer to a bytes slice containing the response body.
// \return[out] bool A flag indicating if the operation was ok.
func AssertRoute(method, route string, code int, t *testing.T) (*[]byte, bool) {
	return AssertRouteWithBody(method, route, nil, code, t)
}

// AssertRouteWithBody is a helper function that checks for a valid route
// \return[out] *[]byte A pointer to a bytes slice containing the response body.
// \return[out] bool A flag indicating if the operation was ok.
func AssertRouteWithBody(method, route string, body *bytes.Buffer, code int, t *testing.T) (*[]byte, bool) {
	jwt := os.Getenv("IGN_TEST_JWT")
	return AssertRouteMultipleArgs(method, route, body, code, &jwt,
		"application/json", t)
}

// AssertRouteMultipleArgs is a helper function that checks for a valid route.
// \param[in] method One of "GET", "PATCH", "PUT", "POST", "DELETE"
// \param[in] route The URL string
// \param[in] body The body to send in the request, or nil
// \param[in] code The expected response HTTP code
// \param[in] signedToken JWT token as base64 string, or nil.
// \param[in] contentType The expected response content type
// \param[in] t Test pointer
// \return[out] *[]byte A pointer to a bytes slice containing the response body.
// \return[out] bool A flag indicating if the operation was ok.
func AssertRouteMultipleArgs(method string, route string, body *bytes.Buffer, code int, signedToken *string, contentType string, t *testing.T) (*[]byte, bool) {
	args := RequestArgs{
		Method:      method,
		Route:       route,
		Body:        body,
		SignedToken: signedToken,
	}
	ar := AssertRouteMultipleArgsStruct(args, code, contentType, t)
	return ar.BodyAsBytes, ar.Ok
}

// RequestArgs - input arguments struct used by AssertRouteMultipleArgsStruct func.
type RequestArgs struct {
	Method      string
	Route       string
	Body        *bytes.Buffer
	SignedToken *string
	Headers     map[string]string
}

// AssertResponse - response of AssertRouteMultipleArgsStruct func.
type AssertResponse struct {
	Ok           bool
	RespRecorder *httptest.ResponseRecorder
	BodyAsBytes  *[]byte
}

// AssertRouteMultipleArgsStruct is a convenient helper function that accepts input
// arguments as a struct, allowing us to keep extending the number of args without changing
// the func signature and the corresponding invocation code everywhere.
// \param[in] args RequestArgs struct
// \param[in] expCode The expected response HTTP code
// \param[in] contentType The expected response content type
// \param[in] t Test pointer
// \return[out] *AssertResponse A pointer to a AssertResponse struct with the response.
func AssertRouteMultipleArgsStruct(args RequestArgs, expCode int, contentType string, t *testing.T) *AssertResponse {
	var b []byte
	var ar AssertResponse

	var buff bytes.Buffer
	if args.Body != nil {
		buff = *args.Body
	}
	// Create a new http request
	req, err := http.NewRequest(args.Method, args.Route, &buff)
	assert.NoError(t, err, "Request failed!")

	// Add the authorization token
	if args.SignedToken != nil {
		req.Header.Set("Authorization", "Bearer "+*args.SignedToken)
	}
	for key, val := range args.Headers {
		req.Header.Set(key, val)
	}

	// Process the request
	respRec := httptest.NewRecorder()
	router.ServeHTTP(respRec, req)
	ar.RespRecorder = respRec

	// Read the result
	var er error
	b, er = ioutil.ReadAll(respRec.Body)
	assert.NoError(t, er, "Failed to read the server response")
	ar.BodyAsBytes = &b

	// Make sure the error code is correct
	assert.Equal(t, expCode, respRec.Code, "Server error: returned %d instead of %d. Route:%s",
		respRec.Code, expCode, args.Route)

	gotCT := respRec.Header().Get("Content-Type")
	assert.Equal(t, contentType, gotCT, "Expected Content-Type[%s] != [%s]. Route:%s",
		contentType, gotCT, args.Route)
	ar.Ok = true
	return &ar
}

// GetIndentedTraceToTest returns a formatted and indented string containing the
// file and line number of each stack frame leading from the current test to
// the assert call that failed.
func GetIndentedTraceToTest() string {
	return fmt.Sprintf("\n\r\tTrace:\n\r%s", indentStrings("\t\t", assert.CallerInfo()))
}

func indentStrings(ind string, lines []string) string {
	res := ""
	for _, s := range lines {
		res += ind + s + "\n\r"
	}
	return res
}

// AssertBackendErrorCode is a function that tries to unmarshal a backend's
// ErrMsg and compares to given error code
func AssertBackendErrorCode(testName string, bslice *[]byte, errCode int, t *testing.T) {
	var errMsg interface{}
	assert.NoError(t, json.Unmarshal(*bslice, &errMsg),
		"Unable to unmarshal bytes slice. Testname:[%s]. Body:[%s]", testName,
		string(*bslice))

	em := errMsg.(map[string]interface{})
	gotCode := em["errcode"].(float64)
	assert.Equal(t, errCode, int(gotCode),
		"errcode [%d] is different than expected code [%d]", int(gotCode), errCode)
	assert.NotEmpty(t, em["errid"], "ErrMsg 'errid' is empty but it should not")
}
