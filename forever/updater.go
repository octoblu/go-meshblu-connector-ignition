package forever

import (
	"net/http"

	"github.com/inconshreveable/go-update"
)

func doUpdate() error {
	resp, err := http.Get("https://github.com/octoblu/go-meshblu-connector-ignition/releases/download/v8.2.3/meshblu-connector-ignition-darwin-amd64")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	err = update.Apply(resp.Body, update.Options{})
	if err != nil {
		if rerr := update.RollbackError(err); rerr != nil {
			return rerr
		}
		return err
	}
	return nil
}
