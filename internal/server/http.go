package server

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strconv"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/joakim-ribier/go-utils/pkg/iosutil"
	"github.com/joakim-ribier/go-utils/pkg/jsonsutil"
	"github.com/joakim-ribier/go-utils/pkg/logsutil"
	"github.com/joakim-ribier/go-utils/pkg/slicesutil"
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

	PathToMockId map[string]string
	logger       logsutil.Logger
}

type MockedRequestLightWithLinks struct {
	internal.MockedRequestLight
	Links map[string]string `json:"_links,omitempty"`
}

// NewHTTPServer creates and initializes a {HTTPServer} struct
func NewHTTPServer(
	port string,
	ssl bool,
	certDirectory, workingDirectory string,
	mocker internal.Mocker,
	logger logsutil.Logger) *HTTPServer {

	return &HTTPServer{
		Port:             port,
		mocker:           mocker,
		SSLEnabled:       ssl,
		certDirectory:    certDirectory,
		workingDirectory: workingDirectory,
		logger:           logger.Namespace("server"),
		PathToMockId:     map[string]string{},
	}
}

// Listen creates the http server and dispatches the incoming requests
func (s HTTPServer) Listen() error {
	server := http.NewServeMux()

	handleFuncToMethods := func(methods []string, pattern string, handle func(w http.ResponseWriter, r *http.Request)) {
		server.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
			remoteAddr := s.findRemoteAddr(r.RemoteAddr)
			s.logger.Info("request", "uri", r.RequestURI, "method", r.Method, "remoteAddr", remoteAddr)
			fmt.Printf("%s [%s] %s\n", remoteAddr, r.Method, r.RequestURI)

			if !slicesutil.Exist(methods, r.Method) {
				w.WriteHeader(404)
				return
			}
			handle(w, r)
		})
	}

	handleFunc := func(method string, pattern string, handle func(w http.ResponseWriter, r *http.Request)) {
		handleFuncToMethods([]string{method}, pattern, handle)
	}

	handleFunc(http.MethodGet, "/", s.home)

	handleFunc(http.MethodGet, "/static/content-types", s.getContentTypes)
	handleFunc(http.MethodGet, "/static/charsets", s.getCharsets)
	handleFunc(http.MethodGet, "/static/status-codes", s.getStatusCodes)

	handleFuncToMethods(
		[]string{http.MethodGet, http.MethodPost}, "/v1/", s.getMockedRequest)
	handleFunc(http.MethodGet, "/v1/raw/", s.getMockedRequestRaw)
	handleFunc(http.MethodGet, "/v1/list", s.list)
	handleFunc(http.MethodPost, "/v1/new", s.addNewMock)

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
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(200)

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
			{"GET", "/v1/{id}", "Get a mocked request"},
			{"GET", "/v1/raw/{id}", "Get a raw mocked request"},
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
		lastId := "N/A"
		lastCreatedAt := "N/A"
		mockedRequests, _ := s.mocker.List()
		if mockedRequests != nil {
			nb = strconv.Itoa(len(mockedRequests))
			if len(mockedRequests) > 0 {
				lastId = mockedRequests[0].Id
				lastCreatedAt = stringsutil.OrElse(mockedRequests[0].CreatedAt, "undefined")
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
			{"Last Id", lastId},
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
	s.writeResponse(w, r, pkg.CONTENT_TYPES, http.StatusOK)
}

func (s HTTPServer) getCharsets(w http.ResponseWriter, r *http.Request) {
	s.writeResponse(w, r, pkg.CHARSET, http.StatusOK)
}

func (s HTTPServer) getStatusCodes(w http.ResponseWriter, r *http.Request) {
	s.writeResponse(w, r, pkg.HTTP_CODES, http.StatusOK)
}

func (s HTTPServer) findMockedRequest(r *http.Request) (*internal.MockedRequest, int, error) {
	var mockId string

	decodedURI, _ := url.QueryUnescape(r.URL.Path)
	if id, ok := s.PathToMockId[decodedURI]; ok {
		mockId = id
	} else {
		url, err := url.ParseRequestURI(r.RequestURI)
		if err != nil {
			s.logger.Error(err, "error to parse URI", "uri", r.RequestURI)
			return nil, 409, err
		}
		mockId = path.Base(url.Path)
	}

	mock, err := s.mocker.Get(mockId)
	if err != nil {
		s.logger.Error(err, "error to get mock", "uri", r.RequestURI)
		return nil, 404, err
	}

	return mock, -1, nil
}

func (s HTTPServer) getMockedRequest(w http.ResponseWriter, r *http.Request) {
	mock, statusCode, err := s.findMockedRequest(r)
	if err != nil {
		writeError(w, err, statusCode)
		return
	}

	NewResponse(w, "60s").Write(*mock, r.URL.Query().Get("delay"))
}

func (s HTTPServer) getMockedRequestRaw(w http.ResponseWriter, r *http.Request) {
	mock, statusCode, err := s.findMockedRequest(r)
	if err != nil {
		writeError(w, err, statusCode)
		return
	}

	if slicesutil.Exist(pkg.IS_DISPLAY_CONTENT, mock.ContentType) {
		mock.Body = string(mock.Body64)
	}

	s.writeResponse(w, r, mock, http.StatusOK)
}

func (s HTTPServer) addNewMock(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		s.logger.Error(err, "error to read body", "uri", r.RequestURI)
		writeError(w, err, 500)
		return
	}

	mock, err := s.mocker.New(r.URL.Query(), body)
	if err != nil {
		s.logger.Error(err, "error to create new mock", "uri", r.RequestURI, "body", body)
		writeError(w, err, 500)
		return
	}

	if mock.Path != "" {
		s.PathToMockId["/v1"+mock.Path] = mock.Id
	}

	if internal.MOCKAPIC_REQ_MAX_LIMIT > 0 {
		s.mocker.Clean(internal.MOCKAPIC_REQ_MAX_LIMIT)
	}

	s.countRemoteAddr(r.RemoteAddr)

	s.writeResponse(w, r, map[string]interface{}{"id": mock.Id, "_links": s.getLinks(r, mock.MockedRequestLight)}, http.StatusCreated)
}

func (s HTTPServer) countRemoteAddr(requestRemoteAddr string) {
	remoteAddrHistory := s.getRemoteAddr()

	remoteAddr := s.findRemoteAddr(requestRemoteAddr)
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

func (s HTTPServer) getLinks(r *http.Request, mock internal.MockedRequestLight) map[string]string {
	values := map[string]string{
		"self": s.getProtocol(r) + "://" + r.Host + "/v1/" + mock.Id,
		"raw":  s.getProtocol(r) + "://" + r.Host + "/v1/raw/" + mock.Id,
	}

	if len(mock.Path) > 0 {
		values["path"] = s.getProtocol(r) + "://" + r.Host + "/v1" + mock.Path
	}

	return values
}

func (s HTTPServer) list(w http.ResponseWriter, r *http.Request) {
	mockedRequestLights, err := s.mocker.List()
	if err != nil {
		s.logger.Error(err, "error to get mocked list", "uri", r.RequestURI)
		writeError(w, err, 500)
		return
	}

	all := slicesutil.TransformT[internal.MockedRequestLight, MockedRequestLightWithLinks](mockedRequestLights, func(mrl internal.MockedRequestLight) (*MockedRequestLightWithLinks, error) {
		return &MockedRequestLightWithLinks{
			MockedRequestLight: mrl,
			Links:              s.getLinks(r, mrl),
		}, nil
	})

	if len(all) == 0 {
		all = []MockedRequestLightWithLinks{}
	}

	s.writeResponse(w, r, all, http.StatusOK)
}

func (s HTTPServer) writeResponse(w http.ResponseWriter, r *http.Request, data any, statusCode int) {
	bytes, err := jsonsutil.Marshal(data)
	if err != nil {
		s.logger.Error(err, "error to marshal data", "uri", r.RequestURI, "data", data)
		writeError(w, err, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	w.Write(bytes)
}

func writeError(w http.ResponseWriter, err error, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if statusCode != http.StatusNotFound {
		w.Write([]byte(fmt.Sprintf(`{"message": "%s"}`, err.Error())))
	}
}
