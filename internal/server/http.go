package server

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"

	"github.com/google/uuid"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/joakim-ribier/gmocky-v2/internal"
	"github.com/joakim-ribier/gmocky-v2/pkg"
	"github.com/joakim-ribier/go-utils/pkg/jsonsutil"
)

// HTTPServer represents a http server struct
type HTTPServer struct {
	Server *http.Server
	Port   string
	mocker internal.Mocker
}

// NewHTTPServer creates and initializes a {HTTPServer} struct
func NewHTTPServer(port string, mocker internal.Mocker) *HTTPServer {
	return &HTTPServer{
		Server: &http.Server{Addr: ":" + port},
		Port:   port,
		mocker: mocker,
	}
}

// Listen creates the http server and dispatches the incoming requests
func (s HTTPServer) Listen() error {
	handleFunc := func(method, pattern string, handle func(w http.ResponseWriter, r *http.Request)) {
		http.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
			if r.Method != method {
				w.WriteHeader(404)
				return
			}
			handle(w, r)
		})
	}

	handleFunc("GET", "/", s.home)

	handleFunc("GET", "/static/content-types", s.getContentTypes)
	handleFunc("GET", "/static/charsets", s.getCharsets)
	handleFunc("GET", "/static/status-codes", s.getStatusCodes)

	handleFunc("GET", "/v1/", s.findMock)
	handleFunc("GET", "/v1/list", s.list)
	handleFunc("POST", "/v1/new", s.addNewMock)

	return s.Server.ListenAndServe()
}

func (s HTTPServer) home(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
	w.Header().Set("Content-Type", "text/plain")

	t := table.NewWriter()
	t.SetTitle("List available APIs\n")

	t.SetStyle(table.StyleDefault)
	t.Style().Options.DrawBorder = false
	t.Style().Options.SeparateColumns = true
	t.Style().Options.SeparateFooter = true
	t.Style().Options.SeparateHeader = false
	t.Style().Options.SeparateRows = false

	t.AppendHeader(table.Row{"Method\n", "Endpoint\n", "Description\n"})

	t.SetColumnConfigs([]table.ColumnConfig{
		{Number: 1, AutoMerge: true},
		{Number: 2, AutoMerge: true},
		{Number: 3, AutoMerge: true},
	})

	t.AppendSeparator()
	t.AppendRows([]table.Row{
		{"GET", "/", "Get info"},
	})
	t.AppendSeparator()
	t.AppendRows([]table.Row{
		{"GET", "/static/content-types", "Get allowed content types"},
		{"GET", "/static/charsets", "Get allowed charsets"},
		{"GET", "/static/status-codes", "Get allowed status codes"},
	})
	t.AppendSeparator()
	t.AppendRows([]table.Row{
		{"GET", "/v1/{uuid}", "Get a mocked request"},
		{"GET", "/v1/list", "Get the list of all mocked requests"},
		{"POST", "/v1/add", "Create a new mocked request"},
	})

	w.Write([]byte(fmt.Sprintf(
		"%s\n\n%s",
		internal.LOGO,
		t.Render())))
}

func (s HTTPServer) getContentTypes(w http.ResponseWriter, r *http.Request) {
	data, err := jsonsutil.Marshal(pkg.CONTENT_TYPES)
	if err != nil {
		writeError(w, err)
		return
	}

	w.WriteHeader(200)
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

func (s HTTPServer) getCharsets(w http.ResponseWriter, r *http.Request) {
	data, err := jsonsutil.Marshal(pkg.CHARSET)
	if err != nil {
		writeError(w, err)
		return
	}

	w.WriteHeader(200)
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

func (s HTTPServer) getStatusCodes(w http.ResponseWriter, r *http.Request) {
	data, err := jsonsutil.Marshal(pkg.HTTP_CODES)
	if err != nil {
		writeError(w, err)
		return
	}

	w.WriteHeader(200)
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

func (s HTTPServer) findMock(w http.ResponseWriter, r *http.Request) {
	url, err := url.Parse(r.RequestURI)
	if err != nil {
		writeError(w, err)
		return
	}

	mockId := path.Base(url.Path)
	if err := uuid.Validate(mockId); err != nil {
		writeError(w, err)
		return

	}

	mock, err := s.mocker.Get(mockId)
	if err != nil {
		fmt.Printf("%v\n", err)
		w.WriteHeader(404)
		return
	}

	fmt.Printf("mock request: %s\n", mockId)
	NewResponse(w, "60s").Write(*mock, r.URL.Query().Get("delay"))
}

func (s HTTPServer) addNewMock(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeError(w, err)
		return
	}

	newUUID, err := s.mocker.New(body)
	if err != nil {
		writeError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write([]byte(fmt.Sprintf(`{"uuid": "%s"}`, *newUUID)))
}

func (s HTTPServer) list(w http.ResponseWriter, r *http.Request) {
	all, err := s.mocker.List()
	if err != nil {
		writeError(w, err)
		return
	}

	data, err := jsonsutil.Marshal(all)
	if err != nil {
		writeError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write(data)
}

func writeError(w http.ResponseWriter, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(409)
	w.Write([]byte(fmt.Sprintf(`{"message": "%s"}`, err.Error())))
}
