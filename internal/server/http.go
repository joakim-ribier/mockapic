package server

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strconv"

	"github.com/google/uuid"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/joakim-ribier/go-utils/pkg/iosutil"
	"github.com/joakim-ribier/go-utils/pkg/jsonsutil"
	"github.com/joakim-ribier/go-utils/pkg/logsutil"
	"github.com/joakim-ribier/go-utils/pkg/stringsutil"
	"github.com/joakim-ribier/mockapic/internal"
	"github.com/joakim-ribier/mockapic/pkg"
)

// HTTPServer represents a http server struct
type HTTPServer struct {
	Port             string
	SSLEnabled       bool
	certDirectory    string
	workingDirectory string
	mocker           internal.Mocker

	logger logsutil.Logger
}

// NewHTTPServer creates and initializes a {HTTPServer} struct
func NewHTTPServer(
	port string, ssl bool, certDirectory, workingDirectory string, mocker internal.Mocker, logger logsutil.Logger) *HTTPServer {

	return &HTTPServer{
		Port:             port,
		mocker:           mocker,
		SSLEnabled:       ssl,
		certDirectory:    certDirectory,
		workingDirectory: workingDirectory,
		logger:           logger.Namespace("server"),
	}
}

// Listen creates the http server and dispatches the incoming requests
func (s HTTPServer) Listen() error {
	server := http.NewServeMux()

	handleFunc := func(method, pattern string, handle func(w http.ResponseWriter, r *http.Request)) {
		server.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
			remoteAddr := s.findRemoteAddr(r.RemoteAddr)
			s.logger.Info("request", "uri", r.RequestURI, "method", r.Method, "remoteAddr", remoteAddr)

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

	if s.SSLEnabled {
		return http.ListenAndServeTLS(
			":"+s.Port,
			s.certDirectory+"/"+internal.MOCKAPIC_CERT_FILENAME,
			s.certDirectory+"/"+internal.MOCKAPIC_PEM_FILENAME,
			server,
		)
	} else {
		return http.ListenAndServe(":"+s.Port, server)
	}
}

func (s HTTPServer) home(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
	w.Header().Set("Content-Type", "text/plain")

	buildAPITable := func() string {
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

		return t.Render()
	}

	buildStatsTable := func() string {
		maxLimit := "unlimited"
		if internal.MOCKAPIC_REQ_MAX_LIMIT > 0 {
			maxLimit = strconv.Itoa(internal.MOCKAPIC_REQ_MAX_LIMIT)
		}

		nb := "N/A"
		lastUUID := "N/A"
		lastCreatedAt := "N/A"
		mockedRequests, _ := s.mocker.List()
		if mockedRequests != nil {
			nb = strconv.Itoa(len(mockedRequests))
			if len(mockedRequests) > 0 {
				lastUUID = mockedRequests[0].UUID
				lastCreatedAt = mockedRequests[0].CreatedAt
			}
		}

		t := table.NewWriter()

		t.SetStyle(table.StyleDefault)
		t.Style().Options.DrawBorder = false
		t.Style().Options.SeparateColumns = true
		t.Style().Options.SeparateFooter = true
		t.Style().Options.SeparateHeader = false
		t.Style().Options.SeparateRows = false

		t.AppendHeader(table.Row{"Name\n", "Value\n"})

		t.AppendSeparator()
		t.AppendRows([]table.Row{
			{"Requests max number authorized", maxLimit},
		})
		t.AppendSeparator()
		t.AppendRows([]table.Row{
			{"Remote addr total number", len(s.getRemoteAddr())},
			{"Requests total number\n", nb},
			{"Last UUID", lastUUID},
			{"Last createdAt", lastCreatedAt},
		})

		return t.Render()
	}

	w.Write([]byte(fmt.Sprintf(
		"%s\n\n%s\n\n\n%s",
		internal.LOGO,
		buildStatsTable(),
		buildAPITable())))
}

func (s HTTPServer) getContentTypes(w http.ResponseWriter, r *http.Request) {
	data, err := jsonsutil.Marshal(pkg.CONTENT_TYPES)
	if err != nil {
		s.logger.Error(err, "error to get content types", "uri", r.RequestURI)
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
		s.logger.Error(err, "error to get charsets", "uri", r.RequestURI)
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
		s.logger.Error(err, "error to get status HTTP codes", "uri", r.RequestURI)
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
		s.logger.Error(err, "error to parse URI", "uri", r.RequestURI)
		writeError(w, err)
		return
	}

	mockId := path.Base(url.Path)
	if err := uuid.Validate(mockId); err != nil {
		s.logger.Error(err, "error to parse UUID", "uri", r.RequestURI)
		writeError(w, err)
		return

	}

	mock, err := s.mocker.Get(mockId)
	if err != nil {
		s.logger.Error(err, "error to get mock", "uri", r.RequestURI, "uuid", mockId)
		w.WriteHeader(404)
		return
	}

	fmt.Printf("mock request: %s\n", mockId)
	NewResponse(w, "60s").Write(*mock, r.URL.Query().Get("delay"))
}

func (s HTTPServer) addNewMock(w http.ResponseWriter, r *http.Request) {

	countRemoteAddr := func() {
		remoteAddrHistory := s.getRemoteAddr()

		remoteAddr := s.findRemoteAddr(r.RemoteAddr)
		if count, is := remoteAddrHistory[remoteAddr]; is {
			remoteAddrHistory[remoteAddr] = count + 1
		} else {
			remoteAddrHistory[remoteAddr] = 1
		}

		data, err := jsonsutil.Marshal(remoteAddrHistory)
		if err == nil {
			iosutil.Write(data, s.workingDirectory+"/remote-addr.json")
		}
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		s.logger.Error(err, "error to read body", "uri", r.RequestURI)
		writeError(w, err)
		return
	}

	newUUID, err := s.mocker.New(r.URL.Query(), body)
	if err != nil {
		s.logger.Error(err, "error to create new mock", "uri", r.RequestURI, "body", body)
		writeError(w, err)
		return
	}

	if internal.MOCKAPIC_REQ_MAX_LIMIT > 0 {
		s.mocker.Clean(internal.MOCKAPIC_REQ_MAX_LIMIT)
	}

	countRemoteAddr()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write([]byte(fmt.Sprintf(`{"uuid": "%s", "_links": {"self": "%s"}}`,
		*newUUID, s.getProtocol(r)+"://"+r.Host+"/v1/"+*newUUID)))
}

func (s HTTPServer) getProtocol(r *http.Request) string {
	protocol := "https"
	if r.TLS == nil {
		protocol = "http"
	}
	return protocol
}

func (s HTTPServer) findRemoteAddr(data string) string {
	ipPort := stringsutil.Split(data, ":", "")
	if len(ipPort) == 0 {
		return "[::1]"
	}
	if len(ipPort) == 1 || len(ipPort) == 2 {
		return ipPort[0]
	}
	return data[:len(data)-(len(ipPort[len(ipPort)-1])+1)]
}

func (s HTTPServer) getRemoteAddr() map[string]int {
	loaded, err := iosutil.Load(s.workingDirectory + "/remote-addr.json")
	if err != nil {
		s.logger.Error(err, "error to load remote addresses", "file", s.workingDirectory+"/remote-addr.json")
		return map[string]int{}
	}

	data, err := jsonsutil.Unmarshal[map[string]int](loaded)
	if err != nil {
		s.logger.Error(err, "error to unmarshal data", "file", s.workingDirectory+"/remote-addr.json", "body", data)
		return map[string]int{}
	}

	return data
}

func (s HTTPServer) list(w http.ResponseWriter, r *http.Request) {
	all, err := s.mocker.List()
	if err != nil {
		s.logger.Error(err, "error to get mocked list", "uri", r.RequestURI)
		writeError(w, err)
		return
	}

	data, err := jsonsutil.Marshal(all)
	if err != nil {
		s.logger.Error(err, "error to marshal list", "uri", r.RequestURI)
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
