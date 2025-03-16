package internal

import (
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/joakim-ribier/go-utils/pkg/iosutil"
	"github.com/joakim-ribier/go-utils/pkg/jsonsutil"
	"github.com/joakim-ribier/go-utils/pkg/logsutil"
	"github.com/joakim-ribier/go-utils/pkg/slicesutil"
)

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

// TestGetWithBadFilename calls Mocker.Get,
// checking for a valid return value.
func TestGetWithBadFilename(t *testing.T) {
	r, err := NewMock(workingDirectory, nil, *logger).Get("file-does-not-exist.json")
	if err == nil {
		t.Fatalf(`result: {%v} but expected error`, r)
	}
}

// TestGetWithBadRequest calls Mocker.Get,
// checking for a valid return value.
func TestGetWithBadRequest(t *testing.T) {
	file := workingDirectory + "/{id}.json"
	defer os.Remove(file)

	mockedRequest := "bad request"

	err := iosutil.Write([]byte(mockedRequest), file)
	if err != nil {
		t.Fatal(err.Error())
	}

	r, err := NewMock(workingDirectory, nil, *logger).Get("{id}")
	if err == nil {
		t.Fatalf(`result: {%v} but expected error`, r)
	}
}

// TestGet calls Mocker.Get,
// checking for a valid return value.
func TestGet(t *testing.T) {
	mockedRequest := createMockedRequest()
	defer os.Remove(workingDirectory + "/" + mockedRequest.Id + ".json")

	r, err := NewMock(workingDirectory, nil, *logger).Get(mockedRequest.Id)
	if err != nil {
		t.Fatal(err.Error())
	}

	if !r.Equals(mockedRequest) {
		t.Fatalf(`result: {%v} but expected {%v}`, r, mockedRequest)
	}
}

// TestGetFromLoadedMockedRequest calls Mocker.Get,
// checking for a valid return value.
func TestGetFromLoadedMockedRequest(t *testing.T) {
	mockedRequest := &MockedRequest{
		MockedRequestLight: MockedRequestLight{
			Id:        "my-own-mocked-request",
			CreatedAt: time.Now().Format("2006-01-02 15:04:05"),
			MockedRequestHeader: MockedRequestHeader{
				Status:      200,
				ContentType: "text/plain",
				Charset:     "UTF-8",
				Headers:     map[string]string{"x-language": "golang", "x-domain": "github.com"},
			},
		},
		Body: "Hello World",
	}

	r, err := NewMock(workingDirectory, []PredefinedMockedRequest{{MockedRequest: *mockedRequest}}, *logger).Get(mockedRequest.Id)
	if err != nil {
		t.Fatal(err.Error())
	}

	mockedRequest.Body64 = []byte("Hello World")
	mockedRequest.Body = ""

	if !r.Equals(*mockedRequest) {
		t.Fatalf(`result: {%v} but expected {%v}`, r, mockedRequest)
	}
}

// TestListWithBadWorkingDir calls Mocker.List,
// checking for a valid return value.
func TestListWithBadWorkingDir(t *testing.T) {
	r, err := NewMock("wrong-directory", nil, *logger).List()
	if err == nil {
		t.Fatalf(`result: {%v} but expected error`, r)
	}
}

// TestList calls Mocker.List,
// checking for a valid return value.
func TestList(t *testing.T) {
	id1 := createMockedRequest().Id
	id2 := createMockedRequest().Id

	r, err := NewMock(workingDirectory, nil, *logger).List()
	if err != nil {
		t.Fatal(err.Error())
	}

	if !slicesutil.ExistT[MockedRequestLight](r, func(ml MockedRequestLight) bool {
		return ml.Id == id1 || ml.Id == id2
	}) {
		t.Fatalf(`result: {%v} but expected {%v}`, r, []string{id1, id2})
	}
}

// TestListWithPredefinedMockedRequests calls Mocker.List,
// checking for a valid return value.
func TestListWithPredefinedMockedRequests(t *testing.T) {
	mockedRequest := MockedRequest{
		MockedRequestLight: MockedRequestLight{
			Id:        uuid.NewString(),
			CreatedAt: time.Now().Format("2006-01-02 15:04:05"),
			MockedRequestHeader: MockedRequestHeader{
				Status:      200,
				ContentType: "text/plain",
				Charset:     "UTF-8",
				Headers:     map[string]string{"x-language": "golang", "x-domain": "github.com"},
			},
		},
		Body64: []byte("Hello World"),
	}

	r, err := NewMock(workingDirectory, []PredefinedMockedRequest{{MockedRequest: mockedRequest}}, *logger).List()
	if err != nil {
		t.Fatal(err.Error())
	}

	if !slicesutil.ExistT[MockedRequestLight](r, func(ml MockedRequestLight) bool { return ml.Id == mockedRequest.Id }) {
		t.Fatalf(`result: {%v} but expected {%v}`, r, []string{mockedRequest.Id})
	}
}

// TestClean calls Mocker.Clean(int),
// checking for a valid return value.
func TestClean(t *testing.T) {
	createMockedRequest()
	createMockedRequest()

	nbBefore, _ := NewMock(workingDirectory, nil, *logger).List()
	nbClean, _ := NewMock(workingDirectory, nil, *logger).Clean(1)
	nbAfter, _ := NewMock(workingDirectory, nil, *logger).List()

	if !(len(nbBefore) > 1 && nbClean > 0 && len(nbAfter) == 1) {
		t.Fatalf(`result: {%v} but expected {%v}`, nbAfter, []string{})
	}

	// test if the max limit is < 0
	r, err := NewMock(workingDirectory, nil, *logger).Clean(-1)
	if r != 0 || err != nil {
		t.Fatalf(`result: {%v} but expected {%v}`, r, 0)
	}

	// test if the max limit is > to the total nb mocked request
	r, err = NewMock(workingDirectory, nil, *logger).Clean(100)
	if r != 0 || err != nil {
		t.Fatalf(`result: {%v} but expected {%v}`, r, 0)
	}

	// test if Mocker.List returns an error
	r, err = NewMock("wrong-directory", nil, *logger).Clean(100)
	if !strings.Contains(err.Error(), "wrong-directory/: no such file or directory") {
		t.Fatalf(`result: {%v} but expected {%v}`, r, err)
	}
}

// TestNewWithBadRequest calls Mocker.New,
// checking for a valid return value.
func TestNewWithBadRequest(t *testing.T) {
	reqParams := map[string][]string{}
	reqBody := `{wrong body}`

	id, err := NewMock(workingDirectory, nil, *logger).New(reqParams, []byte(reqBody))
	if err == nil {
		t.Fatalf(`result: {%v} but expected error`, id)
	}
}

// TestNewWithBadWorkingDir calls Mocker.New,
// checking for a valid return value.
func TestNewWithBadWorkingDir(t *testing.T) {
	reqParams := map[string][]string{
		"status":      {"200"},
		"contentType": {"text/plain"},
		"charset":     {"UTF-8"},
	}
	reqBody := "Hello World"

	r, err := NewMock("wrong-directory", nil, *logger).New(reqParams, []byte(reqBody))
	if err == nil {
		t.Fatalf(`result: {%v} but expected error`, r)
	}
}

// TestNew calls Mocker.New,
// checking for a valid return value.
func TestNew(t *testing.T) {
	reqParams := map[string][]string{
		"status":      {"200"},
		"contentType": {"text/plain"},
		"charset":     {"UTF-8"},
		"x-language":  {"golang"},
		"x-domain":    {"github.com"},
		"path":        {"/my-path"},
	}
	reqBody := "Hello World"

	newMocked, err := NewMock(workingDirectory, nil, *logger).New(reqParams, []byte(reqBody))
	if err != nil {
		t.Fatal(err.Error())
	}

	mock, err := NewMock(workingDirectory, nil, *logger).Get(newMocked.Id)
	if err != nil {
		t.Fatal(err.Error())
	}

	expected := MockedRequest{
		MockedRequestLight: MockedRequestLight{
			MockedRequestHeader: MockedRequestHeader{
				Status:      200,
				ContentType: "text/plain",
				Charset:     "UTF-8",
				Headers:     map[string]string{"x-language": "golang", "x-domain": "github.com"},
				Path:        "/my-path",
			},
		},
		Body64: []byte("Hello World"),
	}
	if !mock.Equals(expected) {
		t.Fatalf(`result: \n%v\n but expected \n%v\n`, mock, expected)
	}
}

// TestNewWithBadStatus calls Mocker.New,
// checking for a valid return value.
func TestNewWithBadStatus(t *testing.T) {
	reqParams := map[string][]string{
		"status":      {"-1"},
		"contentType": {"text/plain"},
		"charset":     {"UTF-8"},
		"x-language":  {"golang"},
	}
	reqBody := "Hello World"

	_, err := NewMock(workingDirectory, nil, *logger).New(reqParams, []byte(reqBody))
	if err.Error() != "status {-1} does not exist" {
		t.Fatalf(`result: {%v} but expected {%v}`, err.Error(), "status does not exist")
	}
}

// TestNewWithBadCharset calls Mocker.New,
// checking for a valid return value.
func TestNewWithBadCharset(t *testing.T) {
	reqParams := map[string][]string{
		"status":      {"200"},
		"contentType": {"text/plain"},
		"charset":     {"wrong-charset"},
	}
	reqBody := "Hello World"

	_, err := NewMock(workingDirectory, nil, *logger).New(reqParams, []byte(reqBody))
	if err.Error() != "charset {wrong-charset} does not exist" {
		t.Fatalf(`result: {%v} but expected {%v}`, err.Error(), "charset does not exist")
	}
}

// TestNewWithBadContentType calls Mocker.New,
// checking for a valid return value.
func TestNewWithBadContentType(t *testing.T) {
	reqParams := map[string][]string{
		"status":      {"200"},
		"contentType": {},
		"charset":     {"UTF-8"},
	}
	reqBody := "Hello World"

	_, err := NewMock(workingDirectory, nil, *logger).New(reqParams, []byte(reqBody))
	if err.Error() != "content type {} does not exist" {
		t.Fatalf(`result: {%v} but expected {%v}`, err.Error(), "content type does not exist")
	}
}

// TestNewMockedRequestFromHttpCode calls NewMockedRequestFromHttpCode,
// checking for a valid return value.
func TestNewMockedRequestFromHttpCode(t *testing.T) {
	r := NewMockedRequestFromHttpCode(418, "I'm a teapot")

	expected := MockedRequest{
		MockedRequestLight: MockedRequestLight{
			Id:        "418",
			CreatedAt: time.Now().Format("2006-01-02 15:04:05"),
			MockedRequestHeader: MockedRequestHeader{
				Status:      418,
				ContentType: "text/plain",
				Charset:     "UTF-8",
			},
		},
		Body64: []byte("I'm a teapot"),
	}

	if !r.Equals(expected) {
		t.Fatalf(`result: {%v} but expected {%v}`, r, expected)
	}
}

func createMockedRequest() MockedRequest {
	mockedRequest := MockedRequest{
		MockedRequestLight: MockedRequestLight{
			Id:        uuid.NewString(),
			CreatedAt: time.Now().Format("2006-01-02 15:04:05"),
			MockedRequestHeader: MockedRequestHeader{
				Status:      200,
				ContentType: "text/plain",
				Charset:     "UTF-8",
				Headers:     map[string]string{"x-language": "golang", "x-domain": "github.com"},
			},
		},
		Body64: []byte("Hello World"),
	}

	bytes, err := jsonsutil.Marshal(mockedRequest)
	if err != nil {
		log.Fatal(err)
	}

	err = iosutil.Write(bytes, workingDirectory+"/"+mockedRequest.Id+".json")
	if err != nil {
		log.Fatal(err)
	}

	return mockedRequest
}
