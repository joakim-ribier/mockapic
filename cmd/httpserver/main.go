package main

import (
	"fmt"
	"log"
	"os"

	"github.com/joakim-ribier/gmocky-v2/internal"
	"github.com/joakim-ribier/gmocky-v2/internal/server"
	"github.com/joakim-ribier/go-utils/pkg/slicesutil"
	"github.com/joakim-ribier/go-utils/pkg/stringsutil"
)

func main() {
	args := slicesutil.ToMap(os.Args[1:])
	if arg, ok := args[string("--home")]; ok {
		internal.GMOCKY_HOME = arg
	}
	if arg, ok := args[string("--port")]; ok {
		internal.GMOCKY_PORT = arg
	}

	httpServer := server.NewHTTPServer(
		stringsutil.OrElse(internal.GMOCKY_PORT, "3333"),
		internal.NewMock(internal.GMOCKY_HOME))

	fmt.Print(internal.LOGO)
	fmt.Printf("Server running on port %s....\n", httpServer.Port)

	if err := httpServer.Listen(); err != nil {
		log.Fatal("could not open httpServer", err)
	}
}
