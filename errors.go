package sdc

import (
	"fmt"
	"net/http"
)

type SDCError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e SDCError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func NewSDCError(code, message string) SDCError {
	return SDCError{Code: code, Message: message}
}

func checkResponseErrors(res *http.Response) error {
	switch res.StatusCode {
	case 400:
		fallthrough
	case 401:
		fallthrough
	case 403:
		fallthrough
	case 404:
		fallthrough
	case 405:
		fallthrough
	case 406:
		fallthrough
	case 409:
		fallthrough
	case 413:
		fallthrough
	case 415:
		fallthrough
	case 420:
		fallthrough
	case 449:
		fallthrough
	case 500:
		fallthrough
	case 502:
		fallthrough
	case 503:
		sdcError := SDCError{}
		err := parseJsonFromReader(res.Body, &sdcError)
		if err != nil {
			return err
		}
		return sdcError
	}
	return nil
}
