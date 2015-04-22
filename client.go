// SDC implements a client for the Joyent Smart Datacenter API.
// https://apidocs.joyent.com/cloudapi/
//
// Included in the package is an incomplete implementation of the
// CLI Utilities.
// https://apidocs.joyent.com/cloudapi/#appendix-d-cloudapi-cli-commands
package manta

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"time"
)

const SDC_VERSION = "~7.0"
const JOYENT_SDC_URL = "https://us-west-1.api.joyentcloud.com"

// Client is a Manta client. Client is not safe for concurrent use.
type Client struct {
	Url     string
	Account string
	User    string
	KeyId   string
	Key     string
	signer  Signer
}

func mustHomedir() string {
	user, err := user.Current()
	if err != nil {
		log.Fatal("manta: could not determine home directory: %v", err)
	}
	return user.HomeDir
}

// NewClient returns a Client instance configured from the
// default SDC environment variables or the parameters passed in.
//
// Here `sdcKey` represents a path on your system
func NewClient(sdcUrl, sdcAccount, sdcUser, sdcKeyId, sdcKey string) *Client {
	c := &Client{}

	c.Url = JOYENT_SDC_URL
	c.Account = os.Getenv("SDC_ACCOUNT")
	c.User = os.Getenv("SDC_USER")
	c.KeyId = os.Getenv("SDC_KEY_ID")
	c.Key = filepath.Join(mustHomedir(), ".ssh", "id_rsa")

	// override default api url
	if url := os.Getenv("SDC_URL"); url != "" {
		c.Url = url
	}
	// override default key
	if key := os.Getenv("SDC_KEY"); key != "" {
		c.Key = key
	}
	// overrides based on parameters given
	if sdcUrl != "" {
		c.Url = sdcUrl
	}
	if sdcAccount != "" {
		c.Account = sdcAccount
	}
	if sdcUser != "" {
		c.User = sdcUser
	}
	if sdcKeyId != "" {
		c.KeyId = sdcKeyId
	}
	if sdcKey != "" {
		c.Key = sdcKey
	}

	// user is account is non provided (not everybody uses RBAC)
	if c.User == "" {
		c.User = c.Account
	}

	return c
}

// NewRequest is similar to http.NewRequest except it appends path to
// the API endpoint this client is configured for.
func (c *Client) NewRequest(method, path string, r io.Reader) (*http.Request, error) {
	url := fmt.Sprintf("%s%s", c.Url, path)
	return http.NewRequest(method, url, r)
}

// SignRequest signs the 'date' field of req.
func (c *Client) SignRequest(req *http.Request) error {
	if c.signer == nil {
		var err error
		c.signer, err = loadPrivateKey(c.Key)
		if err != nil {
			return fmt.Errorf("could not load private key %q: %v", c.Key, err)
		}
	}
	return signRequest(req, fmt.Sprintf("/%s/keys/%s", c.User, c.KeyId), c.signer)
}

// Get executes a GET request and returns the response.
func (c *Client) Get(path string) (*http.Response, error) {
	return c.Do("GET", path, nil)
}

// Post executes a POST request and returns the response.
func (c *Client) Post(path string, data interface{}) (*http.Response, error) {
	return c.Do("POST", path, data)
}

// Put executes a PUT request and returns the response.
func (c *Client) Put(path string, data interface{}) (*http.Response, error) {
	return c.Do("PUT", path, data)
}

// Delete executes a DELETE request and returns the response.
func (c *Client) Delete(path string, data interface{}) (*http.Response, error) {
	return c.Do("DELETE", path, data)
}

// Do executes a method request and returns the response.
func (c *Client) Do(method, path string, data interface{}) (*http.Response, error) {
	var r io.Reader

	// encode json body
	if data != nil {
		dataBytes, err := json.Marshal(data)
		if err != nil {
			return nil, err
		}
		r = strings.NewReader(string(dataBytes))
	}

	// create & sign request
	req, err := c.NewRequest(method, path, r)
	if err != nil {
		return nil, err
	}
	if err := c.SignRequest(req); err != nil {
		return nil, err
	}

	// add json header
	if data != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	// add api version header
	req.Header.Set("Api-Version", SDC_VERSION)

	return http.DefaultClient.Do(req)
}

func signRequest(req *http.Request, keyid string, priv Signer) error {
	now := time.Now().UTC().Format(time.RFC1123)
	req.Header.Set("date", now)
	signed, err := priv.Sign([]byte(fmt.Sprintf("date: %s", now)))
	if err != nil {
		return fmt.Errorf("could not sign request: %v", err)
	}
	sig := base64.StdEncoding.EncodeToString(signed)
	authz := fmt.Sprintf("Signature keyId=%q,algorithm=%q %s", keyid, "rsa-sha256", sig)
	req.Header.Set("Authorization", authz)
	return nil
}

// loadPrivateKey loads an parses a PEM encoded private key file.
func loadPrivateKey(path string) (Signer, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return parsePrivateKey(data)
}

// parsePublicKey parses a PEM encoded private key.
func parsePrivateKey(pemBytes []byte) (Signer, error) {
	block, _ := pem.Decode(pemBytes)
	if block == nil {
		return nil, errors.New("ssh: no key found")
	}

	var rawkey interface{}
	switch block.Type {
	case "RSA PRIVATE KEY":
		rsa, err := x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return nil, err
		}
		rawkey = rsa
	default:
		return nil, fmt.Errorf("ssh: unsupported key type %q", block.Type)
	}
	return newSignerFromKey(rawkey)
}

// A Signer is can create signatures that verify against a public key.
type Signer interface {
	// Sign returns raw signature for the given data. This method
	// will apply the hash specified for the keytype to the data.
	Sign(data []byte) ([]byte, error)
}

func newSignerFromKey(k interface{}) (Signer, error) {
	var sshKey Signer
	switch t := k.(type) {
	case *rsa.PrivateKey:
		sshKey = &rsaPrivateKey{t}
	default:
		return nil, fmt.Errorf("ssh: unsupported key type %T", k)
	}
	return sshKey, nil
}

type rsaPublicKey rsa.PublicKey

type rsaPrivateKey struct {
	*rsa.PrivateKey
}

// Sign signs data with rsa-sha256
func (r *rsaPrivateKey) Sign(data []byte) ([]byte, error) {
	h := sha256.New()
	h.Write(data)
	d := h.Sum(nil)
	return rsa.SignPKCS1v15(rand.Reader, r.PrivateKey, crypto.SHA256, d)
}
