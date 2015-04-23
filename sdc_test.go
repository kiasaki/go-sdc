package sdc

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"
)

var (
	mux *http.ServeMux

	client *Client

	server *httptest.Server
)

func setup() {
	mux = http.NewServeMux()
	server = httptest.NewServer(mux)

	client = DefaultClient()
	client.Key = "_testdata/id_rsa"
	url, _ := url.Parse(server.URL)
	client.Url = url.String()
}

func teardown() {
	server.Close()
}

func TestClient_Do(t *testing.T) {
	setup()
	defer teardown()

	type foo struct {
		A string
	}

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if m := "GET"; m != r.Method {
			t.Errorf("request method want: %v, got: %v", r.Method, m)
		}
		fmt.Fprint(w, `{"A":"a"}`)
	})

	body := new(foo)
	_, err := client.Get("/", &body)
	if err != nil {
		t.Fatalf("Get(): %v", err)
	}

	expected := &foo{"a"}
	if !reflect.DeepEqual(body, expected) {
		t.Errorf("response body want: %v, got: %v", body, expected)
	}
}

func TestClient_DoHttpError(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Bad Request", 400)
	})

	_, err := client.Get("/", nil)

	if err == nil {
		t.Error("expected HTTP 400 error.")
	}

	switch err.(type) {
	case SDCError:
	default:
		t.Error("expected error to be of type SDCError")
	}
}

func TestClient_InvalidJSON(t *testing.T) {
	c := DefaultClient()
	c.Url = ""

	type T struct {
		A map[int]interface{}
	}
	_, err := c.Post("/", &T{}, nil)

	if err == nil {
		t.Error("expected error to be returned.")
	}
	if err, ok := err.(*json.UnsupportedTypeError); !ok {
		t.Errorf("want: JSON error, got: %#v.", err)
	}
}

func TestClient_BadURL(t *testing.T) {
	c := DefaultClient()
	c.Url = ""
	_, err := c.Get(":", nil)

	if err == nil {
		t.Errorf("expected error to be returned")
	}
	if err, ok := err.(*url.Error); !ok || err.Op != "parse" {
		t.Errorf("want: URL parse error, got: %+v", err)
	}
}
