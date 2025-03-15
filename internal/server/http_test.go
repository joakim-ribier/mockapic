package server

import (
	"errors"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/joakim-ribier/go-utils/pkg/httpsutil"
	"github.com/joakim-ribier/go-utils/pkg/iosutil"
	"github.com/joakim-ribier/go-utils/pkg/logsutil"
	"github.com/joakim-ribier/go-utils/pkg/stringsutil"
	"github.com/joakim-ribier/mockapic/internal"
	"github.com/joakim-ribier/mockapic/pkg"
)

type MockerTest struct {
	mockResponse       *internal.MockedRequest
	mockResponseLights []internal.MockedRequestLight
	clean              bool
}

func (m *MockerTest) Get(mockId string) (*internal.MockedRequest, error) {
	if m.mockResponse != nil && m.mockResponse.Id == mockId {
		return m.mockResponse, nil
	}
	return nil, errors.New("mockId does not exist")
}

func (m *MockerTest) List() ([]internal.MockedRequestLight, error) {
	if m.mockResponseLights != nil {
		return m.mockResponseLights, nil
	}
	return nil, errors.New("error to list mocked responses")
}

func (m *MockerTest) New(reqParams map[string][]string, body []byte) (*internal.MockedRequest, error) {
	if len(reqParams["status"]) == 0 || len(reqParams["contentType"]) == 0 || len(reqParams["charset"]) == 0 {
		return nil, errors.New("error to add new mocked response")
	}

	mockedRequest := &internal.MockedRequest{
		MockedRequestLight: internal.MockedRequestLight{
			Id: "{id}",
			MockedRequestHeader: internal.MockedRequestHeader{
				Status:      stringsutil.Int(reqParams["status"][0], -1),
				ContentType: reqParams["contentType"][0],
				Charset:     reqParams["charset"][0],
				Headers:     map[string]string{},
			},
		},
		Body64: body,
	}

	if len(reqParams["path"]) > 0 {
		mockedRequest.Path = reqParams["path"][0]
	}

	m.mockResponse = mockedRequest

	return mockedRequest, nil
}

func (m *MockerTest) Clean(maxLimit int) (int, error) {
	m.clean = true
	return 0, nil
}

var workingDirectory string
var logger *logsutil.Logger

func TestMain(m *testing.M) {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}
	workingDirectory = dir
	logger, err = logsutil.NewLogger(workingDirectory+"/application-test.log", "mockapic-test")
	if err != nil {
		panic(err)
	}

	exitVal := m.Run()

	os.Exit(exitVal)
}

// TestListenEndpoints calls HTTPServer.Listen(),
// checking for a valid return value.
func TestListenEndpoints(t *testing.T) {
	mockedRequest := &internal.MockedRequest{
		MockedRequestLight: internal.MockedRequestLight{
			Id: "{id}",
			MockedRequestHeader: internal.MockedRequestHeader{
				Status:      200,
				ContentType: "text/plain",
				Charset:     "UTF-8",
				Headers:     map[string]string{},
			},
		},
		Body: "Hello World",
	}

	mockerTest := &MockerTest{
		mockResponse:       mockedRequest,
		mockResponseLights: []internal.MockedRequestLight{mockedRequest.MockedRequestLight}}

	baseURL := "http://localhost:3334"
	httpServer := NewHTTPServer("3334", NewSSL(false, "", "", ""), workingDirectory, -1, mockerTest, *logger, "test")

	go func() {
		err := httpServer.Listen()
		if err != nil {
			t.Errorf("Error: %v", err)
		}
	}()
	time.Sleep(100 * time.Millisecond)

	assertReq := func(method, url, body string, expectedCodeResult int) {
		req, _ := httpsutil.NewHttpRequest(baseURL+url, body)
		resp, _ := req.Method(method).Call()
		if resp.StatusCode != expectedCodeResult {
			t.Fatalf(`result: [%s] %s => {%v} but expected {%v}`, method, url, resp.StatusCode, expectedCodeResult)
		}
	}

	assertReq(http.MethodGet, "", "", http.StatusOK)
	assertReq(http.MethodGet, "/static/content-types", "", http.StatusOK)
	assertReq(http.MethodGet, "/static/charsets", "", http.StatusOK)
	assertReq(http.MethodGet, "/static/status-codes", "", http.StatusOK)
	assertReq(http.MethodGet, "/v1/{id}", "", http.StatusOK)
	assertReq(http.MethodPost, "/v1/{id}", "", http.StatusOK)
	assertReq(http.MethodGet, "/v1/{wrong-id}", "", http.StatusNotFound)
	assertReq(http.MethodGet, "/v1/raw/{id}", "", http.StatusOK)
	assertReq(http.MethodGet, "/v1/list", "", http.StatusOK)
	assertReq(http.MethodPost, "/v1/new?status=200&contentType=text/plain&charset=UTF-8", "Hello World", http.StatusCreated)
}

// TestListen calls HTTPServer.Listen(),
// checking for a valid return value.
func TestListenSSL(t *testing.T) {
	httpServer := NewHTTPServer("3333", NewSSL(true, "cert", "", ""), workingDirectory, -1, &MockerTest{}, *logger, "test")

	go func() {
		err := httpServer.Listen()
		if err != nil {
			t.Errorf("Error: %v", err)
		}
	}()
	time.Sleep(100 * time.Millisecond)

	req, _ := httpsutil.NewHttpRequest("https://localhost:3333/", "")
	resp, _ := req.InsecureSkipVerify().Call()

	if resp.StatusCode != 200 {
		t.Fatalf(`result: {%v} but expected {%v}`, resp, 200)
	}

	// testing '404' if bad endpoint is called
	req, _ = httpsutil.NewHttpRequest("https://localhost:3333/", "")
	resp, _ = req.Method("POST").InsecureSkipVerify().Call()
	if resp.StatusCode != 404 {
		t.Fatalf(`result: {%v} but expected {%v}`, resp.StatusCode, 200)
	}
}

// ##
// #### ~/ endpoint
// ##

// TestRootEndpoint calls HTTPServer.home(http.ResponseWriter, *http.Request),
// checking for a valid return value.
func TestRootEndpoint(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "http://localhost:3333/", nil)
	w := httptest.NewRecorder()

	NewHTTPServer("{port}", NewSSL(false, "", "", ""), workingDirectory, -1, &MockerTest{}, *logger, "test").home(w, req)

	_, body := geResultResponse(w, t)
	if !strings.Contains(string(body), internal.LOGO) {
		t.Fatalf(`result: {%v} but expected {%v}`, string(body), internal.LOGO)
	}
}

// ##
// #### ~/static/content-types endpoint
// ##

// TestGetContentTypesEndpoint calls HTTPServer.getContentTypes(http.ResponseWriter, *http.Request),
// checking for a valid return value.
func TestGetContentTypesEndpoint(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "http://localhost:3333/static/content-types", nil)
	w := httptest.NewRecorder()

	NewHTTPServer("{port}", NewSSL(false, "", "", ""), workingDirectory, -1, &MockerTest{}, *logger, "test").getContentTypes(w, req)

	_, body := geResultResponse(w, t)
	if !strings.Contains(string(body), "application/json") {
		t.Fatalf(`result: {%v} but expected {%v}`, string(body), pkg.CONTENT_TYPES)
	}
}

// ##
// #### ~/static/charsets endpoint
// ##

// TestGetCharsetsEndpoint calls HTTPServer.getCharsets(http.ResponseWriter, *http.Request),
// checking for a valid return value.
func TestGetCharsetsEndpoint(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "http://localhost:3333/static/charsets", nil)
	w := httptest.NewRecorder()

	NewHTTPServer("{port}", NewSSL(false, "", "", ""), workingDirectory, -1, &MockerTest{}, *logger, "test").getCharsets(w, req)

	_, body := geResultResponse(w, t)
	if !strings.Contains(string(body), "ISO-8859-1") {
		t.Fatalf(`result: {%v} but expected {%v}`, string(body), pkg.CHARSET)
	}
}

// ##
// #### ~/static/status-codes endpoint
// ##

// TestGetStatusCodesEndpoint calls HTTPServer.getStatusCodes(http.ResponseWriter, *http.Request),
// checking for a valid return value.
func TestGetStatusCodesEndpoint(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "http://localhost:3333/static/content-types", nil)
	w := httptest.NewRecorder()

	NewHTTPServer("{port}", NewSSL(false, "", "", ""), workingDirectory, -1, &MockerTest{}, *logger, "test").getStatusCodes(w, req)

	_, body := geResultResponse(w, t)
	if !strings.Contains(string(body), "Method Not Allowed") {
		t.Fatalf(`result: {%v} but expected {%v}`, string(body), pkg.HTTP_CODES)
	}
}

// ##
// #### ~/v1/{id} endpoint
// ##

// TestGetMockedRequestEndpointIdNotFound calls HTTPServer.getMockedRequest(http.ResponseWriter, *http.Request),
// checking for a valid return value.
func TestGetMockedRequestEndpointIdNotFound(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "http://localhost:3333/v1/{id-not-found}", nil)
	w := httptest.NewRecorder()

	NewHTTPServer("{port}", NewSSL(false, "", "", ""), workingDirectory, -1, &MockerTest{}, *logger, "test").getMockedRequest(w, req)

	res, _ := geResultResponse(w, t)
	if res.Status != "404 Not Found" {
		t.Fatalf(`result: {%v} but expected {%v}`, res, "404")
	}
}

// TestGetMockedRequestEndpoint calls HTTPServer.getMockedRequest(http.ResponseWriter, *http.Request),
// checking for a valid return value.
func TestGetMockedRequestEndpointWithId(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "http://localhost:3333/v1/{id}", nil)
	w := httptest.NewRecorder()

	mocker := &MockerTest{
		mockResponse: &internal.MockedRequest{
			MockedRequestLight: internal.MockedRequestLight{
				Id: "{id}",
				MockedRequestHeader: internal.MockedRequestHeader{
					Status:      200,
					ContentType: "text/plain",
					Charset:     "UTF-8",
					Headers:     map[string]string{},
				},
			},
			Body: "Hello World",
		},
	}

	NewHTTPServer("{port}", NewSSL(false, "", "", ""), workingDirectory, -1, mocker, *logger, "test").getMockedRequest(w, req)

	res, _ := geResultResponse(w, t)
	if res.Status != "200 OK" {
		t.Fatalf(`result: {%v} but expected {%v}`, res, mocker)
	}
}

// TestGetMockedRequestEndpointWithPath calls HTTPServer.getMockedRequest(http.ResponseWriter, *http.Request),
// checking for a valid return value.
func TestGetMockedRequestEndpointWithPath(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/my-path?id={param}", nil)
	w := httptest.NewRecorder()

	mocker := &MockerTest{
		mockResponse: &internal.MockedRequest{
			MockedRequestLight: internal.MockedRequestLight{
				Id: "{id}",
				MockedRequestHeader: internal.MockedRequestHeader{
					Status:      200,
					ContentType: "text/plain",
					Charset:     "UTF-8",
					Path:        "/my-path",
					Headers:     map[string]string{},
				},
			},
			Body: "Hello World",
		},
	}

	s := NewHTTPServer("{port}", NewSSL(false, "", "", ""), workingDirectory, -1, mocker, *logger, "test")
	s.PathToMockId["/my-path"] = "{id}"

	s.getMockedRequest(w, req)

	res, _ := geResultResponse(w, t)
	if res.Status != "200 OK" {
		t.Fatalf(`result: {%v} but expected {%v}`, res, mocker)
	}
}

func TestGetMockedRequestEndpointWithStatusCode(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "http://localhost:3333/v1/418", nil)
	w := httptest.NewRecorder()

	mocker := &MockerTest{}

	NewHTTPServer("{port}", NewSSL(false, "", "", ""), workingDirectory, -1, mocker, *logger, "test").getMockedRequest(w, req)

	res, _ := geResultResponse(w, t)
	if res.Status != "418 I'm a teapot" {
		t.Fatalf(`result: {%v} but expected {%v}`, res, mocker)
	}
}

// TestGetMockedRequestEndpointWithBadRequestURI calls HTTPServer.getMockedRequest(http.ResponseWriter, *http.Request),
// checking for a valid return value.
func TestGetMockedRequestEndpointWithBadRequestURI(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "http://localhost:3333/v1/{id}", nil)
	w := httptest.NewRecorder()

	mocker := &MockerTest{}

	req.RequestURI = "{bad request uri}"
	NewHTTPServer("{port}", NewSSL(false, "", "", ""), workingDirectory, -1, mocker, *logger, "test").getMockedRequest(w, req)

	res, _ := geResultResponse(w, t)
	if res.Status != "409 Conflict" {
		t.Fatalf(`result: {%v} but expected {%v}`, res, mocker)
	}
}

// ##
// #### ~/v1/raw/{id} endpoint
// ##

// TestGetMockedRequestRawEndpoint calls HTTPServer.getMockedRequestRaw(http.ResponseWriter, *http.Request),
// checking for a valid return value.
func TestGetMockedRequestRawEndpoint(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "http://localhost:3333/v1/raw/{id}", nil)
	w := httptest.NewRecorder()

	mocker := &MockerTest{
		mockResponse: &internal.MockedRequest{
			MockedRequestLight: internal.MockedRequestLight{
				Id: "{id}",
				MockedRequestHeader: internal.MockedRequestHeader{
					Status:      200,
					ContentType: "text/plain",
					Charset:     "UTF-8",
					Headers:     map[string]string{"x-language": "golang"},
				},
			},
			Body64: []byte("Hello World"),
		},
	}

	NewHTTPServer("{port}", NewSSL(false, "", "", ""), workingDirectory, -1, mocker, *logger, "test").getMockedRequestRaw(w, req)

	res, data := geResultResponse(w, t)

	if res.Status != "200 OK" ||
		string(data) != `{"id":"{id}","status":200,"contentType":"text/plain","charset":"UTF-8","headers":{"x-language":"golang"},"body":"Hello World","body64":"SGVsbG8gV29ybGQ="}` {

		t.Fatalf(`result: {%v} but expected {%v}`, res, mocker.mockResponse)
	}
}

// ##
// #### ~/v1/list endpoint
// ##

// TestListEndpointWithError calls HTTPServer.list(http.ResponseWriter, *http.Request),
// checking for a valid return value.
func TestListEndpointWithError(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "http://localhost:3333/v1/list", nil)
	w := httptest.NewRecorder()

	NewHTTPServer("{port}", NewSSL(false, "", "", ""), workingDirectory, -1, &MockerTest{}, *logger, "test").list(w, req)

	res, body := geResultResponse(w, t)
	if res.Status != "500 Internal Server Error" || string(body) != `{"message": "error to list mocked responses"}` {
		t.Fatalf(`result: {%v} but expected {%v}`, res, "409")
	}
}

// TestListEndpoint calls HTTPServer.list(http.ResponseWriter, *http.Request),
// checking for a valid return value.
func TestListEndpoint(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "http://localhost:3333/v1/list", nil)
	w := httptest.NewRecorder()

	mocker := &MockerTest{
		mockResponseLights: []internal.MockedRequestLight{
			{
				MockedRequestHeader: internal.MockedRequestHeader{
					Status:      200,
					ContentType: "text/plain",
					Charset:     "UTF-8",
					Headers:     map[string]string{"x-language": "golang"},
				},
				Id: "{id-200}",
			},
			{
				MockedRequestHeader: internal.MockedRequestHeader{
					Status:      404,
					ContentType: "application/json",
					Charset:     "UTF-8",
					Headers:     map[string]string{"x-language": "golang"},
				},
				Id: "{id-404}",
			},
		},
	}
	NewHTTPServer("{port}", NewSSL(false, "", "", ""), workingDirectory, -1, mocker, *logger, "test").list(w, req)

	res, body := geResultResponse(w, t)

	expected := `[{"id":"{id-200}","status":200,"contentType":"text/plain","charset":"UTF-8","headers":{"x-language":"golang"},"_links":{"raw":"http://localhost:3333/v1/raw/{id-200}","self":"http://localhost:3333/v1/{id-200}"}},{"id":"{id-404}","status":404,"contentType":"application/json","charset":"UTF-8","headers":{"x-language":"golang"},"_links":{"raw":"http://localhost:3333/v1/raw/{id-404}","self":"http://localhost:3333/v1/{id-404}"}}]`
	if res.Status != "200 OK" || string(body) != expected {
		t.Fatalf(`result: {%v} but expected {%v}`, res, mocker)
	}
}

// TestListEndpointReturnsEmptyNilInsteadOfNull calls HTTPServer.list(http.ResponseWriter, *http.Request),
// checking for a valid return value.
func TestListEndpointReturnsEmptyNilInsteadOfNull(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "http://localhost:3333/v1/list", nil)
	w := httptest.NewRecorder()

	mocker := &MockerTest{
		mockResponseLights: []internal.MockedRequestLight{},
	}
	NewHTTPServer("{port}", NewSSL(false, "", "", ""), workingDirectory, -1, mocker, *logger, "test").list(w, req)

	res, body := geResultResponse(w, t)

	if res.Status != "200 OK" || string(body) != "[]" {
		t.Fatalf(`result: {%v} but expected {%v}`, string(body), "[]")
	}
}

// ##
// #### ~/v1/new endpoint
// ##

// TestAddNewEndpoint calls HTTPServer.addNewMock(http.ResponseWriter, *http.Request),
// checking for a valid return value.
func TestAddNewEndpoint(t *testing.T) {
	err := iosutil.Write([]byte(``), workingDirectory+"/remote-addr.json")
	if err != nil {
		t.Errorf("Error: %v", err)
	}

	URL := "http://localhost:3333/v1/new?status=200&contentType=text/plain&charset=UTF-8&path=/my-path"
	req := httptest.NewRequest(http.MethodPost, URL, strings.NewReader("Hello World"))
	w := httptest.NewRecorder()

	mocker := &MockerTest{mockResponse: nil, clean: false}
	s := NewHTTPServer("{port}", NewSSL(false, "", "", ""), workingDirectory, 100, mocker, *logger, "test")

	s.addNewMock(w, req)

	res, body := geResultResponse(w, t)

	expected := internal.MockedRequest{
		MockedRequestLight: internal.MockedRequestLight{
			MockedRequestHeader: internal.MockedRequestHeader{
				Status:      200,
				ContentType: "text/plain",
				Charset:     "UTF-8",
				Path:        "/my-path",
				Headers:     map[string]string{},
			},
		},
		Body64: []byte("Hello World"),
	}

	if res.Status != "201 Created" ||
		string(body) != `{"_links":{"path":"http://localhost:3333/v1/my-path","raw":"http://localhost:3333/v1/raw/{id}","self":"http://localhost:3333/v1/{id}"},"id":"{id}"}` ||
		!mocker.mockResponse.Equals(expected) ||
		!mocker.clean ||
		len(s.getRemoteAddr()) != 1 ||
		s.PathToMockId["/v1/my-path"] != "{id}" {
		t.Fatalf(`result: {%v} but expected {%v}`, res, expected)
	}
}

// TestAddNewEndpointWithBadRequest calls HTTPServer.addNewMock(http.ResponseWriter, *http.Request),
// checking for a valid return value.
func TestAddNewEndpointWithBadRequest(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/v1/new", strings.NewReader("bad body..."))
	w := httptest.NewRecorder()

	mocker := &MockerTest{mockResponse: nil}
	NewHTTPServer("{port}", NewSSL(false, "", "", ""), workingDirectory, -1, mocker, *logger, "test").addNewMock(w, req)

	res, body := geResultResponse(w, t)

	if res.Status != "500 Internal Server Error" ||
		string(body) != `{"message": "error to add new mocked response"}` {
		t.Fatalf(`result: {%v} but expected {%v}`, res, "409")
	}
}

// TestFindRemoteAddr calls HTTPServer.findRemoteAddr(string),
// checking for a valid return value.
func TestFindRemoteAddr(t *testing.T) {
	httpServer := NewHTTPServer("{port}", NewSSL(false, "", "", ""), workingDirectory, -1, &MockerTest{}, *logger, "test")

	if r := httpServer.findRemoteAddr("127.0.0.1:3333"); r != "127.0.0.1" {
		t.Fatalf(`result: {%v} but expected {%v}`, r, "127.0.0.1")
	}
	if r := httpServer.findRemoteAddr("[::1]:3333"); r != "[::1]" {
		t.Fatalf(`result: {%v} but expected {%v}`, r, "[::1]")
	}
	if r := httpServer.findRemoteAddr(""); r != "[::1]" {
		t.Fatalf(`result: {%v} but expected {%v}`, r, "[::1]")
	}
	if r := httpServer.findRemoteAddr("127.0.0.1"); r != "127.0.0.1" {
		t.Fatalf(`result: {%v} but expected {%v}`, r, "127.0.0.1")
	}
}

// TestGetRemoteAddr calls HTTPServer.getRemoteAddr(),
// checking for a valid return value.
func TestGetRemoteAddr(t *testing.T) {
	err := iosutil.Write([]byte(`{"127.0.0.1":6,"192.168.0.1":10}`), workingDirectory+"/remote-addr.json")
	if err != nil {
		t.Errorf("Error: %v", err)
	}

	httpServer := NewHTTPServer("{port}", NewSSL(false, "", "", ""), workingDirectory, -1, &MockerTest{}, *logger, "test")

	r := httpServer.getRemoteAddr()

	if len(r) != 2 || r["127.0.0.1"] != 6 || r["192.168.0.1"] != 10 {
		t.Fatalf(`result: {%v} but expected {%v}`, r, "empty result")
	}
}

// TestGetRemoteAddrWithBadContent calls HTTPServer.getRemoteAddr(),
// checking for a valid return value.
func TestGetRemoteAddrWithBadContent(t *testing.T) {
	err := iosutil.Write([]byte("bad-content"), workingDirectory+"/remote-addr.json")
	if err != nil {
		t.Errorf("Error: %v", err)
	}

	httpServer := NewHTTPServer("{port}", NewSSL(false, "", "", ""), workingDirectory, -1, &MockerTest{}, *logger, "test")

	if r := httpServer.getRemoteAddr(); len(r) != 0 {
		t.Fatalf(`result: {%v} but expected {%v}`, r, "empty result")
	}
}

// TestCountRemoteAddr calls HTTPServer.countRemoteAddr(),
// checking for a valid return value.
func TestCountRemoteAddr(t *testing.T) {
	err := iosutil.Write([]byte(`{"192.1.34.1":6}`), workingDirectory+"/remote-addr.json")
	if err != nil {
		t.Errorf("Error: %v", err)
	}

	httpServer := NewHTTPServer("{port}", NewSSL(false, "", "", ""), workingDirectory, -1, &MockerTest{}, *logger, "test")

	httpServer.countRemoteAddr("92.1.34.1")
	httpServer.countRemoteAddr("192.1.34.1")

	if r := httpServer.getRemoteAddr(); len(r) != 2 || r["92.1.34.1"] != 1 || r["192.1.34.1"] != 7 {
		t.Fatalf(`result: {%v} but expected {%v}`, r, "[92.1.34.1:1,192.1.34.1:7]")
	}
}

func geResultResponse(w *httptest.ResponseRecorder, t *testing.T) (http.Response, []byte) {
	res := w.Result()
	defer res.Body.Close()

	data, err := io.ReadAll(res.Body)
	if err != nil {
		t.Errorf("Error: %v", err)
	}

	return *res, data
}
