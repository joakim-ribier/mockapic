package internal

import (
	"fmt"
	"io/fs"
	"os"
	"reflect"
	"time"

	"github.com/google/uuid"
	"github.com/joakim-ribier/gmocky-v2/pkg"
	"github.com/joakim-ribier/go-utils/pkg/genericsutil"
	"github.com/joakim-ribier/go-utils/pkg/iosutil"
	"github.com/joakim-ribier/go-utils/pkg/jsonsutil"
	"github.com/joakim-ribier/go-utils/pkg/slicesutil"
)

type MockedRequest struct {
	UUID      string
	CreatedAt string
	// payload from request
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
	UUID        string
	CreatedAt   string
	Name        string
	Status      int
	ContentType string
}

type Mocker interface {
	Get(mockId string) (*MockedRequest, error)
	List() ([]MockedRequestLight, error)
	New(body []byte) (*string, error)
	Clean(maxLimit int) (int, error)
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

	values := slicesutil.SortT[MockedRequestLight, string](
		slicesutil.TransformT[fs.DirEntry, MockedRequestLight](entries, func(e fs.DirEntry) (*MockedRequestLight, error) {
			var mockId string = ""
			if len(e.Name()) > 5 {
				mockId = e.Name()[:len(e.Name())-5]
			}
			return get[MockedRequestLight](m.workingDirectory, mockId)
		}), func(mrl1, mrl2 MockedRequestLight) (string, string) { return mrl2.CreatedAt, mrl1.CreatedAt })

	return genericsutil.OrElse(
		values, func() bool { return len(values) > 0 }, []MockedRequestLight{}), nil
}

// New adds a new mocked request and returns the new UUID
func (m Mock) New(body []byte) (*string, error) {
	var mock *MockedRequest
	data, err := jsonsutil.Unmarshal[MockedRequest](body)
	if err != nil {
		return nil, err
	}
	mock = &data

	if _, is := pkg.HTTP_CODES[mock.Status]; !is {
		return nil, fmt.Errorf("status {%d} does not exist", mock.Status)
	}

	if !slicesutil.Exist(pkg.CONTENT_TYPES, mock.ContentType) {
		return nil, fmt.Errorf("content type {%s} does not exist", mock.ContentType)
	}

	if !slicesutil.Exist(pkg.CHARSET, mock.Charset) {
		return nil, fmt.Errorf("charset {%s} does not exist", mock.Charset)
	}

	mock.UUID = uuid.NewString()
	mock.CreatedAt = time.Now().Format("2006-01-02 15:04:05")

	body, err = jsonsutil.Marshal(mock)
	if err != nil {
		return nil, err
	}

	err = iosutil.Write(body, m.workingDirectory+"/"+mock.UUID+".json")
	if err != nil {
		return nil, err
	}

	return &mock.UUID, nil
}

// Clean removes the x (nb mocked request - max limit) last requests
func (m Mock) Clean(maxLimit int) (int, error) {
	nb := 0
	if maxLimit < 1 {
		return nb, nil
	}
	mockedRequests, err := m.List()
	if err != nil {
		return nb, err
	}
	nbToDelete := len(mockedRequests) - maxLimit
	if nbToDelete < 1 {
		return nb, nil
	}
	for _, mockedRequest := range mockedRequests[len(mockedRequests)-nbToDelete:] {
		if err := os.Remove(m.workingDirectory + "/" + mockedRequest.UUID + ".json"); err == nil {
			nb = nb + 1
		}
	}
	return nb, nil
}
