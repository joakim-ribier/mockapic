package internal

import (
	"fmt"
	"io/fs"
	"os"
	"reflect"

	"github.com/google/uuid"
	"github.com/joakim-ribier/gmocky-v2/pkg"
	"github.com/joakim-ribier/go-utils/pkg/genericsutil"
	"github.com/joakim-ribier/go-utils/pkg/iosutil"
	"github.com/joakim-ribier/go-utils/pkg/jsonsutil"
	"github.com/joakim-ribier/go-utils/pkg/slicesutil"
)

type MockedRequest struct {
	Name        string
	Status      int
	ContentType string
	Charset     string
	Headers     map[string]string
	Body        string
}

// Equals returns true if the two structs are equal
func (m MockedRequest) Equals(arg MockedRequest) bool {
	return m.Status == arg.Status &&
		m.ContentType == arg.ContentType &&
		m.Charset == arg.Charset &&
		m.Body == arg.Body &&
		reflect.DeepEqual(m.Headers, arg.Headers)
}

type MockedRequestLight struct {
	Name        string
	UUID        string
	Status      int
	ContentType string
}

type Mocker interface {
	Get(mockId string) (*MockedRequest, error)
	List() ([]MockedRequestLight, error)
	New(body []byte) (*string, error)
}

type Mock struct {
	workingDirectory string
}

func NewMock(workingDirectory string) Mock {
	return Mock{workingDirectory: workingDirectory}
}

// Get finds the mocked request {mockId} on the storage
func (m Mock) Get(mockId string) (*MockedRequest, error) {
	return get[MockedRequest](m.workingDirectory, mockId)
}

func get[T any](workingDirectory, mockId string) (*T, error) {
	bytes, err := iosutil.Load(workingDirectory + "/" + mockId + ".json")
	if err != nil {
		return nil, err
	}

	mock, err := jsonsutil.Unmarshal[T](bytes)
	if err != nil {
		return nil, err
	}
	return &mock, nil
}

// List gets all mocked request on the storage
func (m Mock) List() ([]MockedRequestLight, error) {
	entries, err := os.ReadDir(m.workingDirectory + "/")
	if err != nil {
		return nil, err
	}
	values := slicesutil.TransformT[fs.DirEntry, MockedRequestLight](entries, func(e fs.DirEntry) (*MockedRequestLight, error) {
		var mockId string = ""
		if len(e.Name()) > 5 {
			mockId = e.Name()[:len(e.Name())-5]
		}
		mock, err := get[MockedRequestLight](m.workingDirectory, mockId)
		if mock != nil {
			mock.UUID = mockId
		}
		return mock, err
	})
	return genericsutil.OrElse(
		values, func() bool { return len(values) > 0 }, []MockedRequestLight{}), nil
}

// New adds a new mocked request and returns the new UUID
func (m Mock) New(body []byte) (*string, error) {
	mock, err := jsonsutil.Unmarshal[MockedRequest](body)
	if err != nil {
		return nil, err
	}

	if _, is := pkg.HTTP_CODES[mock.Status]; !is {
		return nil, fmt.Errorf("status {%d} does not exist", mock.Status)
	}

	if !slicesutil.Exist(pkg.CONTENT_TYPES, mock.ContentType) {
		return nil, fmt.Errorf("content type {%s} does not exist", mock.ContentType)
	}

	if !slicesutil.Exist(pkg.CHARSET, mock.Charset) {
		return nil, fmt.Errorf("charset {%s} does not exist", mock.Charset)
	}

	newUUID := uuid.NewString()
	err = iosutil.Write(body, m.workingDirectory+"/"+newUUID+".json")
	if err != nil {
		return nil, err
	}

	return &newUUID, nil
}
