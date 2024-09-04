package server

import (
	"net/http"
	"time"

	"github.com/joakim-ribier/go-utils/pkg/genericsutil"
	"github.com/joakim-ribier/go-utils/pkg/stringsutil"
	"github.com/joakim-ribier/mockapic/internal"
)

// Response represents a {http.ResponseWriter} from the HTTP request
type Response struct {
	ResponseWriter http.ResponseWriter
	DelayMax       time.Duration
}

// NewResponse creates and initializes a {Response} struct
func NewResponse(responseWriter http.ResponseWriter, delayMax string) Response {
	duration, _ := time.ParseDuration(delayMax)

	return Response{
		ResponseWriter: responseWriter,
		DelayMax:       duration,
	}
}

// Write writes the http response using the provided {mock} value
// and delays the response {delay} parameter is setted
func (r Response) Write(mock internal.MockedRequest, delay string) {
	var duration time.Duration = 0
	if parse, err := time.ParseDuration(delay); err == nil {
		duration = genericsutil.OrElse(
			parse, func() bool { return parse <= r.DelayMax }, r.DelayMax)
	}

	if duration > 0 {
		time.Sleep(duration)
	}

	r.
		writeContentType(mock).
		writeHeaders(mock).
		writeBody(mock)
}

func (r Response) writeContentType(mock internal.MockedRequest) Response {
	if contentType := mock.ContentType; contentType != "" {
		r.ResponseWriter.Header().
			Set("Content-Type", contentType+"; charset="+stringsutil.OrElse(mock.Charset, "utf-8"))
	}
	return r
}

func (r Response) writeHeaders(mock internal.MockedRequest) Response {
	for key, value := range mock.Headers {
		r.ResponseWriter.Header().Set(key, value)
	}
	r.ResponseWriter.WriteHeader(mock.Status)
	return r
}

func (r Response) writeBody(mock internal.MockedRequest) Response {
	if len(mock.Body) > 0 {
		r.ResponseWriter.Write([]byte(mock.Body))
	}
	return r
}
