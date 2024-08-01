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
	"github.com/joakim-ribier/go-utils/pkg/slicesutil"
)

var workingDirectory string

func TestMain(m *testing.M) {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}
	workingDirectory = dir

	exitVal := m.Run()

	os.Exit(exitVal)
}

// TestGetWithBadFilename calls Mocker.Get,
// checking for a valid return value.
func TestGetWithBadFilename(t *testing.T) {
	r, err := NewMock(workingDirectory).Get("file-does-not-exist.json")
	if err == nil {
		t.Fatalf(`result: {%v} but expected error`, r)
	}
}

// TestGet calls Mocker.Get,
// checking for a valid return value.
func TestGetWithBadContent(t *testing.T) {
	newUUID := uuid.NewString()
	file := workingDirectory + "/" + newUUID + ".json"
	defer os.Remove(file)

	mockedRequest := "BAD BODY"

	err := iosutil.Write([]byte(mockedRequest), file)
	if err != nil {
		log.Fatal(err)
	}

	r, err := NewMock(workingDirectory).Get(newUUID)
	if err == nil {
		t.Fatalf(`result: {%v} but expected error`, r)
	}
}

// TestGet calls Mocker.Get,
// checking for a valid return value.
func TestGet(t *testing.T) {
	newUUID, mockedRequest := createMockedRequest()
	defer os.Remove(workingDirectory + "/" + newUUID + ".json")

	r, _ := NewMock(workingDirectory).Get(newUUID)
	if r.Body != newUUID {
		t.Fatalf(`result: {%v} but expected {%s}`, r, mockedRequest)
	}
}

// TestListWithBadWorkingDir calls Mocker.List,
// checking for a valid return value.
func TestListWithBadWorkingDir(t *testing.T) {
	r, err := NewMock("wrong-directory").List()
	if err == nil {
		t.Fatalf(`result: {%v} but expected error`, r)
	}
}

// TestList calls Mocker.List,
// checking for a valid return value.
func TestList(t *testing.T) {
	newUUID1, _ := createMockedRequest()
	newUUID2, _ := createMockedRequest()

	r, _ := NewMock(workingDirectory).List()

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

	nbBefore, _ := NewMock(workingDirectory).List()
	nbClean, _ := NewMock(workingDirectory).Clean(1)
	nbAfter, _ := NewMock(workingDirectory).List()

	if !(len(nbBefore) > 1 && nbClean > 0 && len(nbAfter) == 1) {
		t.Fatalf(`result: {%v} but expected {%v}`, nbAfter, []string{})
	}

	// test if the max limit is < 0
	r, err := NewMock(workingDirectory).Clean(-1)
	if r != 0 || err != nil {
		t.Fatalf(`result: {%v} but expected {%v}`, r, 0)
	}

	// test if the max limit is > to the total nb mocked request
	r, err = NewMock(workingDirectory).Clean(100)
	if r != 0 || err != nil {
		t.Fatalf(`result: {%v} but expected {%v}`, r, 0)
	}

	// test if Mocker.List returns an error
	r, err = NewMock("wrong-directory").Clean(100)
	if !strings.Contains(err.Error(), "wrong-directory/: no such file or directory") {
		t.Fatalf(`result: {%v} but expected {%v}`, r, err)
	}
}

// TestNewWithBadBody calls Mocker.New,
// checking for a valid return value.
func TestNewWithBadBody(t *testing.T) {
	body := `{wrong body}`

	newUUID, err := NewMock(workingDirectory).New([]byte(body))
	if err == nil {
		t.Fatalf(`result: {%v} but expected error`, newUUID)
	}
}

// TestNewWithBadWorkingDir calls Mocker.New,
// checking for a valid return value.
func TestNewWithBadWorkingDir(t *testing.T) {
	r, err := NewMock("wrong-directory").New([]byte(`{
    	"status": 200,
    	"contentType": "text/plain",
    	"charset": "UTF-8",
    	"body": "Hello World"
	}`))
	if err == nil {
		t.Fatalf(`result: {%v} but expected error`, r)
	}
}

// TestNew calls Mocker.New,
// checking for a valid return value.
func TestNew(t *testing.T) {
	body := `{
    	"status": 200,
    	"contentType": "text/plain",
    	"charset": "UTF-8",
    	"body": "Hello World",
		"headers": {"x-language": "golang"}
	}`

	newUUID, _ := NewMock(workingDirectory).New([]byte(body))
	mock, _ := NewMock(workingDirectory).Get(*newUUID)

	expected := &MockedRequest{
		Status:      200,
		ContentType: "text/plain",
		Charset:     "UTF-8",
		Body:        "Hello World",
		Headers:     map[string]string{"x-language": "golang"},
	}
	if !mock.Equals(*expected) {
		t.Fatalf(`result: \n%v\n but expected \n%v\n`, mock, expected)
	}
}

// TestNewWithBadStatus calls Mocker.New,
// checking for a valid return value.
func TestNewWithBadStatus(t *testing.T) {
	body := `{
    	"status": -1,
    	"contentType": "text/plain",
    	"charset": "UTF-8",
    	"body": "Hello World"
	}`

	_, err := NewMock(workingDirectory).New([]byte(body))
	if err.Error() != "status {-1} does not exist" {
		t.Fatalf(`result: {%v} but expected {%v}`, err.Error(), "status does not exist")
	}
}

// TestNewWithBadCharset calls Mocker.New,
// checking for a valid return value.
func TestNewWithBadCharset(t *testing.T) {
	body := `{
    	"status": 200,
    	"contentType": "text/plain",
    	"charset": "wrong-charset",
    	"body": "Hello World"
	}`

	_, err := NewMock(workingDirectory).New([]byte(body))
	if err.Error() != "charset {wrong-charset} does not exist" {
		t.Fatalf(`result: {%v} but expected {%v}`, err.Error(), "charset does not exist")
	}
}

// TestNewWithBadContentType calls Mocker.New,
// checking for a valid return value.
func TestNewWithBadContentType(t *testing.T) {
	body := `{
    	"status": 200,
    	"contentType": "wrong-content-type",
    	"charset": "UTF-8",
    	"body": "Hello World"
	}`

	_, err := NewMock(workingDirectory).New([]byte(body))
	if err.Error() != "content type {wrong-content-type} does not exist" {
		t.Fatalf(`result: {%v} but expected {%v}`, err.Error(), "content type does not exist")
	}
}

func createMockedRequest() (string, string) {
	newUUID := uuid.NewString()
	mockedRequest := fmt.Sprintf(`{
	    "uuid": "%s",
    	"status": 200,
    	"contentType": "text/plain",
    	"charset": "UTF-8",
    	"body": "%s",
		"createdAt:": "%s"
	}`, newUUID, newUUID, time.Now().Format("2006-01-02 15:04:05"))

	err := iosutil.Write([]byte(mockedRequest), workingDirectory+"/"+newUUID+".json")
	if err != nil {
		log.Fatal(err)
	}

	return newUUID, mockedRequest
}
