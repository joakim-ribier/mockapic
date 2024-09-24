package internal

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"reflect"
	"time"

	"github.com/google/uuid"
	"github.com/joakim-ribier/go-utils/pkg/iosutil"
	"github.com/joakim-ribier/go-utils/pkg/jsonsutil"
	"github.com/joakim-ribier/go-utils/pkg/logsutil"
	"github.com/joakim-ribier/go-utils/pkg/slicesutil"
	"github.com/joakim-ribier/go-utils/pkg/stringsutil"
	"github.com/joakim-ribier/mockapic/pkg"
)

type MockedRequestHeader struct {
	Status      int               `json:"status,omitempty"`
	ContentType string            `json:"contentType,omitempty"`
	Charset     string            `json:"charset,omitempty"`
	Headers     map[string]string `json:"headers,omitempty"`
}

type MockedRequestLight struct {
	Id        string `json:"id,omitempty"`
	CreatedAt string `json:"createdAt,omitempty"`
	MockedRequestHeader
}

type MockedRequest struct {
	MockedRequestLight
	Body   string `json:"body,omitempty"`
	Body64 []byte `json:"body64,omitempty"`
}

type PredefinedMockedRequest struct {
	MockedRequest
}

func (m PredefinedMockedRequest) toMockedRequest() *MockedRequest {
	body := m.Body64
	if len(m.Body) > 0 {
		body = []byte(m.Body)
	}
	return &MockedRequest{
		MockedRequestLight: MockedRequestLight{
			MockedRequestHeader: m.MockedRequestHeader,
			Id:                  m.Id},
		Body64: body,
	}
}

// Equals returns true if the two requests are equal
func (m MockedRequest) Equals(arg MockedRequest) bool {
	return m.Status == arg.Status &&
		m.ContentType == arg.ContentType &&
		m.Charset == arg.Charset &&
		m.Body == arg.Body &&
		bytes.Equal(m.Body64, arg.Body64) &&
		reflect.DeepEqual(m.Headers, arg.Headers)
}

type Mocker interface {
	Get(mockId string) (*MockedRequest, error)
	List() ([]MockedRequestLight, error)
	New(params map[string][]string, body []byte) (*string, error)
	Clean(maxLimit int) (int, error)
}

type Mock struct {
	workingDirectory         string
	logger                   logsutil.Logger
	predefinedMockedRequests []PredefinedMockedRequest
}

func NewMock(workingDirectory string, predefinedMockedRequests []PredefinedMockedRequest, logger logsutil.Logger) Mock {
	return Mock{
		workingDirectory:         workingDirectory,
		logger:                   logger.Namespace("mock"),
		predefinedMockedRequests: predefinedMockedRequests}
}

// Get finds the mocked request by {mockId} value on the storage or in the predefined requests.
func (m Mock) Get(mockId string) (*MockedRequest, error) {
	mock, err := get[MockedRequest](m.workingDirectory, mockId, m.logger)
	if mock != nil {
		return mock, nil
	}

	if mock := slicesutil.FindT[PredefinedMockedRequest](
		m.predefinedMockedRequests, func(mr PredefinedMockedRequest) bool { return mr.Id == mockId }); mock != nil {
		return mock.toMockedRequest(), nil
	}

	return nil, err
}

func get[T any](workingDirectory, mockId string, logger logsutil.Logger) (*T, error) {
	bytes, err := iosutil.Load(workingDirectory + "/" + mockId + ".json")
	if err != nil {
		logger.Error(err, "error to load data", "mockId", mockId, "workingDirectory", workingDirectory)
		return nil, err
	}

	mock, err := jsonsutil.Unmarshal[T](bytes)
	if err != nil {
		logger.Error(err, "error to unmarshal data", "mockId", mockId, "workingDirectory", workingDirectory, "data", bytes)
		return nil, err
	}
	return &mock, nil
}

// List gets all mocked requests on the storage and the predefined requests.
func (m Mock) List() ([]MockedRequestLight, error) {
	fileEntries, err := os.ReadDir(m.workingDirectory + "/")
	if err != nil {
		m.logger.Error(err, "error to read directory", "workingDirectory", m.workingDirectory)
		return nil, err
	}

	mockedRequestsLight := slicesutil.TransformT[fs.DirEntry, MockedRequestLight](fileEntries, func(e fs.DirEntry) (*MockedRequestLight, error) {
		var mockId string = ""
		if len(e.Name()) > 5 {
			mockId = e.Name()[:len(e.Name())-5]
		}
		return get[MockedRequestLight](m.workingDirectory, mockId, m.logger)
	})

	if len(m.predefinedMockedRequests) > 0 {
		mockedRequestsLight = append(mockedRequestsLight, slicesutil.TransformT[PredefinedMockedRequest, MockedRequestLight](
			m.predefinedMockedRequests, func(lmr PredefinedMockedRequest) (*MockedRequestLight, error) {
				return &lmr.MockedRequestLight, nil
			})...)
	}

	return slicesutil.SortT[MockedRequestLight, string](
		mockedRequestsLight, func(mrl1, mrl2 MockedRequestLight) (string, string) {
			return mrl2.CreatedAt, mrl1.CreatedAt
		}), nil
}

// New creates a new mocked request and returns the new identifier.
func (m Mock) New(reqParams map[string][]string, reqBody []byte) (*string, error) {
	mock := &MockedRequest{
		MockedRequestLight: MockedRequestLight{
			Id:                  uuid.NewString(),
			CreatedAt:           time.Now().Format("2006-01-02 15:04:05"),
			MockedRequestHeader: MockedRequestHeader{Headers: map[string]string{}},
		},
		Body64: reqBody,
	}

	getReqParam := func(values []string) string {
		if len(values) == 0 {
			return ""
		}
		return values[0]
	}

	for name, values := range reqParams {
		switch name {
		case "contentType":
			mock.ContentType = getReqParam(values)
		case "charset":
			mock.Charset = getReqParam(values)
		case "status":
			mock.Status = stringsutil.Int(getReqParam(values), -1)
		default:
			if len(values) > 0 {
				mock.Headers[name] = values[0]
			}
		}
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

	bytes, err := jsonsutil.Marshal(mock)
	if err != nil {
		m.logger.Error(err, "error to nmarshal data", "mock", mock)
		return nil, err
	}

	err = iosutil.Write(bytes, m.workingDirectory+"/"+mock.Id+".json")
	if err != nil {
		m.logger.Error(err, "error to write data", "mock", mock, "workingDirectory", m.workingDirectory)
		return nil, err
	}

	return &mock.Id, nil
}

// Clean removes the x (nb mocked request - max limit) last requests.
func (m Mock) Clean(maxLimit int) (int, error) {
	nb := 0
	if maxLimit < 1 {
		return nb, nil
	}

	mockedRequests, err := m.List()
	if err != nil {
		m.logger.Error(err, "error to list requests", "workingDirectory", m.workingDirectory)
		return nb, err
	}

	nbToDelete := len(mockedRequests) - maxLimit
	if nbToDelete < 1 {
		return nb, nil
	}

	for _, mockedRequest := range mockedRequests[len(mockedRequests)-nbToDelete:] {
		if err := os.Remove(m.workingDirectory + "/" + mockedRequest.Id + ".json"); err == nil {
			nb = nb + 1
		}
	}
	return nb, nil
}
