package internal

import (
	"fmt"
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
	r, err := NewMock(workingDirectory, *logger).Get("file-does-not-exist.json")
	if err == nil {
		t.Fatalf(`result: {%v} but expected error`, r)
	}
}

// TestGetWithBadRequest calls Mocker.Get,
// checking for a valid return value.
func TestGetWithBadRequest(t *testing.T) {
	newUUID := uuid.NewString()
	file := workingDirectory + "/" + newUUID + ".json"
	defer os.Remove(file)

	mockedRequest := "bad request"

	err := iosutil.Write([]byte(mockedRequest), file)
	if err != nil {
		t.Fatalf(err.Error())
	}

	r, err := NewMock(workingDirectory, *logger).Get(newUUID)
	if err == nil {
		t.Fatalf(`result: {%v} but expected error`, r)
	}
}

// TestGet calls Mocker.Get,
// checking for a valid return value.
func TestGet(t *testing.T) {
	mockedRequest := createMockedRequest()
	defer os.Remove(workingDirectory + "/" + mockedRequest.UUID + ".json")

	r, err := NewMock(workingDirectory, *logger).Get(mockedRequest.UUID)
	if err != nil {
		t.Fatalf(err.Error())
	}

	if !r.Equals(mockedRequest) {
		t.Fatalf(`result: {%v} but expected {%v}`, r, mockedRequest)
	}
}

// TestListWithBadWorkingDir calls Mocker.List,
// checking for a valid return value.
func TestListWithBadWorkingDir(t *testing.T) {
	r, err := NewMock("wrong-directory", *logger).List()
	if err == nil {
		t.Fatalf(`result: {%v} but expected error`, r)
	}
}

// TestList calls Mocker.List,
// checking for a valid return value.
func TestList(t *testing.T) {
	newUUID1 := createMockedRequest().UUID
	newUUID2 := createMockedRequest().UUID

	r, err := NewMock(workingDirectory, *logger).List()
	if err != nil {
		t.Fatalf(err.Error())
	}

	if !slicesutil.ExistT[MockedRequestLight](r, func(ml MockedRequestLight) bool {
		return ml.UUID == newUUID1 || ml.UUID == newUUID2
	}) {
		t.Fatalf(`result: {%v} but expected {%v}`, r, []string{newUUID1, newUUID2})
	}
}

// TestClean calls Mocker.Clean(int),
// checking for a valid return value.
func TestClean(t *testing.T) {
	createMockedRequest()
	createMockedRequest()

	nbBefore, _ := NewMock(workingDirectory, *logger).List()
	nbClean, _ := NewMock(workingDirectory, *logger).Clean(1)
	nbAfter, _ := NewMock(workingDirectory, *logger).List()

	if !(len(nbBefore) > 1 && nbClean > 0 && len(nbAfter) == 1) {
		t.Fatalf(`result: {%v} but expected {%v}`, nbAfter, []string{})
	}

	// test if the max limit is < 0
	r, err := NewMock(workingDirectory, *logger).Clean(-1)
	if r != 0 || err != nil {
		t.Fatalf(`result: {%v} but expected {%v}`, r, 0)
	}

	// test if the max limit is > to the total nb mocked request
	r, err = NewMock(workingDirectory, *logger).Clean(100)
	if r != 0 || err != nil {
		t.Fatalf(`result: {%v} but expected {%v}`, r, 0)
	}

	// test if Mocker.List returns an error
	r, err = NewMock("wrong-directory", *logger).Clean(100)
	if !strings.Contains(err.Error(), "wrong-directory/: no such file or directory") {
		t.Fatalf(`result: {%v} but expected {%v}`, r, err)
	}
}

// TestNewWithBadRequest calls Mocker.New,
// checking for a valid return value.
func TestNewWithBadRequest(t *testing.T) {
	reqParams := map[string][]string{}
	reqBody := `{wrong body}`

	newUUID, err := NewMock(workingDirectory, *logger).New(reqParams, []byte(reqBody))
	if err == nil {
		t.Fatalf(`result: {%v} but expected error`, newUUID)
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

	r, err := NewMock("wrong-directory", *logger).New(reqParams, []byte(reqBody))
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
	}
	reqBody := "Hello World"

	newUUID, err := NewMock(workingDirectory, *logger).New(reqParams, []byte(reqBody))
	if err != nil {
		t.Fatalf(err.Error())
	}

	mock, err := NewMock(workingDirectory, *logger).Get(*newUUID)
	if err != nil {
		t.Fatalf(err.Error())
	}

	expected := &MockedRequest{
		Status:      200,
		ContentType: "text/plain",
		Charset:     "UTF-8",
		Body:        []byte("Hello World"),
		Headers:     map[string]string{"x-language": "golang", "x-domain": "github.com"},
	}
	if !mock.Equals(*expected) {
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

	_, err := NewMock(workingDirectory, *logger).New(reqParams, []byte(reqBody))
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

	_, err := NewMock(workingDirectory, *logger).New(reqParams, []byte(reqBody))
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

	_, err := NewMock(workingDirectory, *logger).New(reqParams, []byte(reqBody))
	if err.Error() != "content type {} does not exist" {
		t.Fatalf(`result: {%v} but expected {%v}`, err.Error(), "content type does not exist")
	}
}

func createMockedRequest() MockedRequest {
	newUUID := uuid.NewString()

	mockedRequest := MockedRequest{
		UUID:        newUUID,
		Status:      200,
		ContentType: "text/plain",
		Charset:     "UTF-8",
		Body:        []byte(fmt.Sprintf("%s-body", newUUID)),
		Headers:     map[string]string{"x-language": "golang", "x-domain": "github.com"},
		CreatedAt:   time.Now().Format("2006-01-02 15:04:05"),
	}

	bytes, err := jsonsutil.Marshal(mockedRequest)
	if err != nil {
		log.Fatal(err)
	}

	err = iosutil.Write(bytes, workingDirectory+"/"+newUUID+".json")
	if err != nil {
		log.Fatal(err)
	}

	return mockedRequest
}
