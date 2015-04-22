package sdc

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/user"
	"path/filepath"
	"testing"
)

func TestParsePrivateKey(t *testing.T) {
	data, err := ioutil.ReadFile("_testdata/id_rsa")
	if err != nil {
		t.Fatal(err)
	}
	_, err = parsePrivateKey(data)
	if err != nil {
		t.Fatal(err)
	}
}

func TestLoadPrivateKey(t *testing.T) {
	if _, err := loadPrivateKey("_testdata/id_rsa"); err != nil {
		t.Error(err)
	}
}

func TestPrivateKeySign(t *testing.T) {
	priv, err := loadPrivateKey("_testdata/id_rsa")
	if err != nil {
		t.Fatal(err)
	}
	sig, err := priv.Sign([]byte("date: Thu, 05 Jan 2012 21:31:40 GMT"))
	if err != nil {
		t.Fatal(err)
	}
	const want = "ATp0r26dbMIxOopqw0OfABDT7CKMIoENumuruOtarj8n/97Q3htHFYpH8yOSQk3Z5zh8UxUym6FYTb5+A0Nz3NRsXJibnYi7brE/4tx5But9kkFGzG+xpUmimN4c3TMN7OFH//+r8hBf7BT9/GmHDUVZT2JzWGLZES2xDOUuMtA="
	if got := base64.StdEncoding.EncodeToString(sig); got != want {
		t.Fatalf("want: %q, got %q", want, got)
	}
}

func TestSignRequest(t *testing.T) {
	priv, err := loadPrivateKey("_testdata/id_rsa")
	if err != nil {
		t.Fatal(err)
	}
	req, err := http.NewRequest("GET", "http://example.com/", nil)
	if err != nil {
		t.Fatal(err)
	}
	signRequest(req, "Test", priv)
	date := req.Header.Get("date")
	sig, err := priv.Sign([]byte("date: " + date))
	if err != nil {
		t.Fatal(err)
	}
	want := fmt.Sprintf("Signature keyId=%q,algorithm=%q %s", "Test", "rsa-sha256", base64.StdEncoding.EncodeToString(sig))
	if got := req.Header.Get("Authorization"); got != want {
		t.Fatalf("want: %q, got: %q", want, got)
	}
}

func TestClientNewRequest(t *testing.T) {
	client := Client{
		User:  "test",
		KeyId: "q",
		Key:   "_testdata/id_rsa",
		Url:   "http://example.com",
	}
	req, err := client.NewRequest("GET", "/test/public", nil)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := req.URL.Host, "example.com"; got != want {
		t.Errorf("want: %q, got: %q", want, got)
	}
	if got, want := req.URL.Path, "/test/public"; got != want {
		t.Errorf("want: %q, got: %q", want, got)
	}
}

func TestClientSignRequest(t *testing.T) {
	client := Client{
		User:  "test",
		KeyId: "q",
		Key:   "_testdata/id_rsa",
		Url:   "http://example.com",
	}
	req, err := client.NewRequest("GET", "/test/public", nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := client.SignRequest(req); err != nil {
		t.Fatal(err)
	}
}

func TestNewClient_Defaults(t *testing.T) {
	clearEnv()

	c := NewClient("", "", "", "", "")

	user, err := user.Current()
	if err != nil {
		t.Fatal(err)
	}
	wantedKey := filepath.Join(user.HomeDir, ".ssh", "id_rsa")

	checkClient(t, c, JOYENT_SDC_URL, "", "", "", wantedKey)
}

func TestNewClient_Environment(t *testing.T) {
	wantedUrl := "http://api.example.com"
	wantedAccount := "bert"
	wantedUser := "bert-jr"
	wantedKeyId := "f0:12:a2:8b:31:..."
	wantedKey := "/home/bert/.ssh/joyent"

	os.Setenv("SDC_URL", wantedUrl)
	os.Setenv("SDC_ACCOUNT", wantedAccount)
	os.Setenv("SDC_USER", wantedUser)
	os.Setenv("SDC_KEY_ID", wantedKeyId)
	os.Setenv("SDC_KEY", wantedKey)

	c := NewClient("", "", "", "", "")
	checkClient(t, c, wantedUrl, wantedAccount, wantedUser, wantedKeyId, wantedKey)
}

func TestNewClient_Params(t *testing.T) {
	clearEnv()

	wantedUrl := "http://api.example.com"
	wantedAccount := "bert"
	wantedUser := "bert-jr"
	wantedKeyId := "f0:12:a2:8b:31:..."
	wantedKey := "/home/bert/.ssh/joyent"

	c := NewClient(wantedUrl, wantedAccount, wantedUser, wantedKeyId, wantedKey)
	checkClient(t, c, wantedUrl, wantedAccount, wantedUser, wantedKeyId, wantedKey)
}

func TestNewClient_AccountIsUser(t *testing.T) {
	clearEnv()

	account := "bert"
	c := NewClient("", account, "", "", "")

	if c.User != account {
		t.Errorf("want: %q, got: %q", account, c.User)
	}
}

func checkClient(t *testing.T, c *Client,
	wantedUrl, wantedAccount, wantedUser, wantedKeyId, wantedKey string) {

	if c.Url != wantedUrl {
		t.Errorf("want: %q, got: %q", wantedUrl, c.Url)
	}
	if c.Account != wantedAccount {
		t.Errorf("want: %q, got: %q", wantedAccount, c.Account)
	}
	if c.User != wantedUser {
		t.Errorf("want: %q, got: %q", wantedUser, c.User)
	}
	if c.KeyId != wantedKeyId {
		t.Errorf("want: %q, got: %q", wantedKeyId, c.KeyId)
	}
	if c.Key != wantedKey {
		t.Errorf("want: %q, got: %q", wantedKey, c.Key)
	}
}

func clearEnv() {
	os.Setenv("SDC_URL", "")
	os.Setenv("SDC_ACCOUNT", "")
	os.Setenv("SDC_USER", "")
	os.Setenv("SDC_KEY_ID", "")
	os.Setenv("SDC_KEY", "")
}
