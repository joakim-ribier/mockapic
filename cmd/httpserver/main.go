package main

import (
	"fmt"
	"log"
	"os"

	"github.com/joakim-ribier/gmocky-v2/internal"
	"github.com/joakim-ribier/gmocky-v2/internal/server"
	"github.com/joakim-ribier/go-utils/pkg/genericsutil"
	"github.com/joakim-ribier/go-utils/pkg/slicesutil"
	"github.com/joakim-ribier/go-utils/pkg/stringsutil"
)

func main() {
	args := slicesutil.ToMap(os.Args[1:])
	if arg, ok := args["--home"]; ok {
		internal.GMOCKY_HOME = arg
	}
	if arg, ok := args["--req_max"]; ok {
		internal.GMOCKY_REQ_MAX_LIMIT = stringsutil.Int(arg, -1)
	}
	if arg, ok := args["--port"]; ok {
		internal.GMOCKY_PORT = arg
	}
	if arg, ok := args["--cert"]; ok {
		internal.GMOCKY_CERT_DIRECTORY = arg
	}
	if arg, ok := args["--ssl"]; ok {
		internal.GMOCKY_SSL = stringsutil.Bool(arg)
		if internal.GMOCKY_SSL && internal.GMOCKY_CERT_DIRECTORY == "" {
			internal.GMOCKY_CERT_FILENAME = "example.crt"
			internal.GMOCKY_PEM_FILENAME = "example.key"
			internal.GMOCKY_CERT_DIRECTORY = "../../cert"
		}
	}

	httpServer := server.NewHTTPServer(
		stringsutil.OrElse(internal.GMOCKY_PORT, "3333"),
		internal.GMOCKY_SSL,
		internal.GMOCKY_CERT_DIRECTORY,
		internal.GMOCKY_HOME,
		internal.NewMock(internal.GMOCKY_HOME+"/requests"))

	fmt.Print(internal.LOGO)
	fmt.Printf("Server running on port %s[:%s]....\n",
		genericsutil.When(internal.GMOCKY_SSL, func(arg bool) bool { return arg }, "https", "http"),
		httpServer.Port)

	if err := httpServer.Listen(); err != nil {
		log.Fatal("could not open httpServer", err)
	}
}
