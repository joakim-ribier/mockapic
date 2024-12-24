package internal

import (
	"os"

	"github.com/joakim-ribier/go-utils/pkg/stringsutil"
)

const LOGO = `
    __  ___              __                  _
   /  |/  /____   _____ / /__ ____ _ ____   (_)_____
  / /|_/ // __ \ / ___// //_// __ '// __ \ / // ___/
 / /  / // /_/ // /__ / ,<  / /_/ // /_/ // // /__   _  _  _
/_/  /_/ \____/ \___//_/|_| \__,_// .___//_/ \___/  (_)(_)(_)
                                 /_/
                    https://github.com/joakim-ribier/mockapic
`

var MOCKAPIC_REQ_MAX_LIMIT = stringsutil.Int(os.Getenv("MOCKAPIC_REQ_MAX_LIMIT"), -1)

var MOCKAPIC_SSL = stringsutil.Bool(os.Getenv("MOCKAPIC_SSL"))
var MOCKAPIC_CERT_DIRECTORY = os.Getenv("MOCKAPIC_CERT")
var MOCKAPIC_CERT_FILENAME = "mockapic.crt"
var MOCKAPIC_PEM_FILENAME = "mockapic.key"
