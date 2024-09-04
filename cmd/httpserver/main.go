package main

import (
	"fmt"
	"log"
	"os"

	"github.com/joakim-ribier/go-utils/pkg/genericsutil"
	"github.com/joakim-ribier/go-utils/pkg/logsutil"
	"github.com/joakim-ribier/go-utils/pkg/slicesutil"
	"github.com/joakim-ribier/go-utils/pkg/stringsutil"
	"github.com/joakim-ribier/mockapic/internal"
	"github.com/joakim-ribier/mockapic/internal/server"
)

func main() {
	args := slicesutil.ToMap(os.Args[1:])
	if arg, ok := args["--home"]; ok {
		internal.MOCKAPIC_HOME = arg
		if _, err := os.Open(internal.MOCKAPIC_HOME); err != nil {
			log.Fatalf("%v", err)
		}
	}
	if arg, ok := args["--req_max"]; ok {
		internal.MOCKAPIC_REQ_MAX_LIMIT = stringsutil.Int(arg, -1)
	}
	if arg, ok := args["--port"]; ok {
		internal.MOCKAPIC_PORT = arg
	}
	if arg, ok := args["--cert"]; ok {
		internal.MOCKAPIC_CERT_DIRECTORY = arg
	}
	if arg, ok := args["--ssl"]; ok {
		internal.MOCKAPIC_SSL = stringsutil.Bool(arg)
		if internal.MOCKAPIC_SSL && internal.MOCKAPIC_CERT_DIRECTORY == "" {
			internal.MOCKAPIC_CERT_FILENAME = "example.crt"
			internal.MOCKAPIC_PEM_FILENAME = "example.key"
			internal.MOCKAPIC_CERT_DIRECTORY = "../../cert"
		}
	}

	logger, err := logsutil.NewLogger(internal.MOCKAPIC_HOME+"/application.log", "mockapic")
	if err != nil {
		log.Fatalf("%v", err)
	}

	logger.Info(internal.LOGO,
		"home", internal.MOCKAPIC_HOME,
		"port", internal.MOCKAPIC_PORT,
		"ssl", internal.MOCKAPIC_SSL,
		"req_max", internal.MOCKAPIC_REQ_MAX_LIMIT,
	)

	httpServer := server.NewHTTPServer(
		stringsutil.OrElse(internal.MOCKAPIC_PORT, "3333"),
		internal.MOCKAPIC_SSL,
		internal.MOCKAPIC_CERT_DIRECTORY,
		internal.MOCKAPIC_HOME,
		internal.NewMock(internal.MOCKAPIC_HOME+"/requests", *logger),
		*logger)

	fmt.Print(internal.LOGO)
	fmt.Printf("\nServer running on port %s[:%s]....\n",
		genericsutil.When(internal.MOCKAPIC_SSL, func(arg bool) bool { return arg }, "https", "http"),
		httpServer.Port)

	if err := httpServer.Listen(); err != nil {
		log.Fatal("could not open httpServer", err)
	}
}
