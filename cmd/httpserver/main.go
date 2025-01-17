package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/joakim-ribier/go-utils/pkg/genericsutil"
	"github.com/joakim-ribier/go-utils/pkg/iosutil"
	"github.com/joakim-ribier/go-utils/pkg/jsonsutil"
	"github.com/joakim-ribier/go-utils/pkg/logsutil"
	"github.com/joakim-ribier/go-utils/pkg/stringsutil"
	"github.com/joakim-ribier/mockapic/internal"
	"github.com/joakim-ribier/mockapic/internal/server"
)

func main() {
	reqMaxLimit := flag.Int("req_max", stringsutil.Int(os.Getenv("MOCKAPIC_REQ_MAX_LIMIT"), -1), "define the nb requests max limit")
	port := flag.String("port", stringsutil.OrElse(os.Getenv("MOCKAPIC_PORT"), "3333"), "define the server [port]")
	workingDir := flag.String("home", stringsutil.OrElse(os.Getenv("MOCKAPIC_HOME"), "."), "define the [working home] directory")

	ssl := flag.Bool("ssl", stringsutil.Bool(stringsutil.OrElse(os.Getenv("MOCKAPIC_SSL"), "false")), "enable [ssl] mode")
	certificatesDir := flag.String("cert", os.Getenv("MOCKAPIC_CERT"), "define the [certificate] directory that contains the *crt and the *key files if [ssl] mode enabled")
	crtFilePath := flag.String("crt", os.Getenv("MOCKAPIC_CRT_FILE_PATH"), "define the [*crt] file path if [ssl] mode enabled")
	keyFilePath := flag.String("key", os.Getenv("MOCKAPIC_KEY_FILE_PATH"), "define the [*key] file path if [ssl] mode enabled")

	flag.Parse()
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)

	logger, err := logsutil.NewLogger(*workingDir+"/application.log", "mockapic")
	if err != nil {
		log.Fatalf("%v", err)
	}

	if _, err := os.Open(*workingDir); err != nil {
		log.Fatalf("'--home' parameter {%s} must be a valid directory.\n%v", *workingDir, err)
	}
	requestsDir := *workingDir + "/requests"
	requestsPredefinedFile := *workingDir + "/mockapic.json"
	*certificatesDir = genericsutil.When[bool, string](*ssl, func(b bool) bool { return *ssl && *certificatesDir == "" }, *workingDir, *certificatesDir)

	logger.Info(internal.LOGO,
		"home", workingDir,
		"port", port,
		"ssl", ssl,
		"cert", certificatesDir,
		"crt", crtFilePath,
		"key", keyFilePath,
		"req_max", reqMaxLimit,
	)

	err = os.MkdirAll(requestsDir, os.ModePerm)
	if err != nil {
		log.Fatalf("%v", err)
	}

	predefinedMockedRequests := []internal.PredefinedMockedRequest{}
	data, err := iosutil.Load(requestsPredefinedFile)
	if err != nil {
		logger.Error(err, fmt.Sprintf("file {%s} not found", requestsPredefinedFile))
	} else {
		predefinedMockedRequests, err = jsonsutil.Unmarshal[[]internal.PredefinedMockedRequest](data)
		if err != nil {
			logger.Error(err, fmt.Sprintf("file {%s} cannot be parsed", requestsPredefinedFile))
		}
	}

	mock := internal.NewMock(requestsDir, predefinedMockedRequests, *logger)

	httpServer := server.NewHTTPServer(
		stringsutil.OrElse(*port, "3333"),
		server.NewSSL(*ssl, *certificatesDir, *crtFilePath, *keyFilePath),
		*workingDir,
		*reqMaxLimit,
		mock,
		*logger,
		version())

	fmt.Print(internal.LOGO)

	// load existing requests...
	var pathToMockId map[string]string
	if values, err := mock.List(); err == nil && len(values) > 0 {
		fmt.Printf("\nLoad %d request%s!", len(values), genericsutil.When(values, func(v []internal.MockedRequestLight) bool { return len(v) > 1 }, "s", ""))
		pathToMockId = make(map[string]string, len(values))
		for _, b := range values {
			pathToMockId["/v1"+b.Path] = b.Id
		}
		httpServer.PathToMockId = pathToMockId
	}

	fmt.Printf("\nServer running on port %s[:%s]....\n",
		genericsutil.When(*ssl, func(arg bool) bool { return arg }, "https", "http"),
		httpServer.Port)

	if err := httpServer.Listen(); err != nil {
		log.Fatal("could not open httpServer", err)
	}
}

// version reads the version number from the auto generated file {generated-version.txt}
func version() string {
	version, err := os.ReadFile("generated-version.txt")
	if err != nil {
		return "unknown"
	}
	return string(version)
}
