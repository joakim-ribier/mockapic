package server

import (
	"net/http"
	"testing"

	"github.com/joakim-ribier/go-utils/pkg/slicesutil"
	"github.com/joakim-ribier/go-utils/pkg/timesutil"
	"github.com/joakim-ribier/mockapic/internal"
)

type ResponseWriterTest struct {
	statusCode int
	headers    map[string][]string
	body       string
}

func (r *ResponseWriterTest) Header() http.Header {
	return r.headers
}

func (r *ResponseWriterTest) Write(body []byte) (int, error) {
	r.body = string(body)
	return -1, nil
}

func (r *ResponseWriterTest) WriteHeader(d int) {
	r.statusCode = d
}

// TestWrite calls Response.Write(internal.Mock, string),
// checking for a valid return value.
func TestWrite(t *testing.T) {
	mocked := internal.MockedRequest{
		MockedRequestLight: internal.MockedRequestLight{
			MockedRequestHeader: internal.MockedRequestHeader{
				Status:      200,
				ContentType: "text/plain",
				Charset:     "UTF-8",
				Headers:     map[string]string{"x-language": "golang"},
			},
		},
		Body64: []byte("Hello World"),
	}

	r := NewResponse(&ResponseWriterTest{
		headers: make(map[string][]string),
	}, "60s")

	withTime, _ := timesutil.WithExecutionTime(func() (*internal.MockedRequest, error) {
		r.Write(mocked, "")
		return &mocked, nil
	})

	value := r.ResponseWriter.(*ResponseWriterTest)
	if value.statusCode != 200 ||
		value.body != "Hello World" ||
		!slicesutil.ContainAll(value.headers["Content-Type"], []string{"text/plain; charset=UTF-8"}) ||
		!slicesutil.ContainAll(value.headers["X-Language"], []string{"golang"}) ||
		withTime.TimeInMillis > 100 {

		t.Fatalf(`result: {%v} but expected {%v}`, value, mocked)
	}
}

// TestWriteWithMaxDelay calls Response.Write(internal.Mock, string),
// checking for a valid return value.
func TestWriteWithMaxDelay(t *testing.T) {
	mocked := internal.MockedRequest{
		MockedRequestLight: internal.MockedRequestLight{
			MockedRequestHeader: internal.MockedRequestHeader{
				Status: 500,
			},
		},
	}

	r := NewResponse(&ResponseWriterTest{
		headers: make(map[string][]string),
	}, "1000ms")

	withTime, _ := timesutil.WithExecutionTime(func() (*internal.MockedRequest, error) {
		r.Write(mocked, "30s")
		return &mocked, nil
	})

	if !(withTime.TimeInMillis > 950 && withTime.TimeInMillis < 1050) {
		t.Fatalf(`result: {%v} but expected {%v}`, withTime.TimeInMillis, "1s max")
	}
}
