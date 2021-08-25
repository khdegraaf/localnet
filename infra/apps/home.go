package apps

import (
	"os"

	"github.com/ridge/must"
)

var home string

func init() {
	home = must.String(os.UserHomeDir()) + "/.localnet"
}
