package worker

import (
	"github.com/robertlestak/sigc/internal/keys"
)

func Start() {
	go keys.Expirer()
	select {}
}
