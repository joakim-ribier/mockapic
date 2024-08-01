package internal

import (
	"os"

	"github.com/joakim-ribier/go-utils/pkg/stringsutil"
)

const LOGO = `
       ______        __  ___ ____   ______ __ ____  __      __   _____  ______ ____  _    __ ______ ____
      / ____/       /  |/  // __ \ / ____// //_/\ \/ /    _/_/  / ___/ / ____// __ \| |  / // ____// __ \
     / / __ ______ / /|_/ // / / // /    / ,<    \  /   _/_/    \__ \ / __/  / /_/ /| | / // __/  / /_/ /
    / /_/ //_____// /  / // /_/ // /___ / /| |   / /  _/_/     ___/ // /___ / _, _/ | |/ // /___ / _, _ /_  _  _
    \____/       /_/  /_/ \____/ \____//_/ |_|  /_/  /_/      /____//_____//_/ |_|  |___//_____//_/ |_| (_)(_)(_)
                                                                    https://github.com/joakim-ribier/gmocky-v2
`

var GMOCKY_HOME = os.Getenv("GMOCKY_HOME")
var GMOCKY_REQ_MAX_LIMIT = stringsutil.Int(os.Getenv("GMOCKY_REQ_MAX_LIMIT"), -1)

var GMOCKY_PORT = os.Getenv("GMOCKY_PORT")

var GMOCKY_SSL = stringsutil.Bool(os.Getenv("GMOCKY_SSL"))
var GMOCKY_CERT_DIRECTORY = os.Getenv("GMOCKY_CERT")
var GMOCKY_CERT_FILENAME = "gmocky.crt"
var GMOCKY_PEM_FILENAME = "gmocky.key"
