package sdc

import (
	"encoding/json"
	"io"
	"io/ioutil"
)

func parseJsonFromReader(r io.Reader, obj interface{}) error {
	contents, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}

	err = json.Unmarshal(contents, obj)
	return nil
}
