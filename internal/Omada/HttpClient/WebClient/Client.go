package WebClient

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"sync"
	"time"

	"omada_exporter_go/internal"
	"omada_exporter_go/internal/Log"
	"omada_exporter_go/internal/Omada/HttpClient/ApiClient"
	"omada_exporter_go/internal/Omada/HttpClient/ClientInstrumentation"
	"omada_exporter_go/internal/Omada/HttpClient/Utils"
)

const (
	path_login       = "/{omadaID}/api/v2/login"
	path_logout      = "/{omadaID}/api/v2/logout"
	path_loginStatus = "/{omadaID}/api/v2/loginStatus"

	threshold_loginCheck int64 = 10
)

type WebClient struct {
	BaseURL            string
	OmadaID            string
	username           string
	password           string
	SiteID             string
	SiteName           string
	Client             *http.Client
	Token              string
	lastLoginTime      int64
	lastLoginCheckTime int64
}

func (w *WebClient) fillInOmadaIDs(placeholders map[string]string) map[string]string {
	if placeholders == nil {
		placeholders = make(map[string]string)
	}
	placeholders["omadaID"] = w.OmadaID
	placeholders["siteID"] = w.SiteID
	return placeholders
}

var (
	instance *WebClient
	once     sync.Once
)

func newClient(baseURL string, username string, password string, siteName string) *WebClient {
	jar, _ := cookiejar.New(nil)
	customTransport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	instrumentedTransport := &ClientInstrumentation.InstrumentedRoundTripper{
		RoundTripper: customTransport,
		ClientType:   ClientInstrumentation.WebApiClientType,
	}

	openApiClient := ApiClient.GetInstance()

	clientObject := &WebClient{
		BaseURL:  baseURL,
		OmadaID:  openApiClient.OmadaID,
		username: username,
		password: password,
		SiteName: siteName,
		SiteID:   openApiClient.SiteID,

		Client: &http.Client{
			Jar:       jar,
			Transport: instrumentedTransport,
		},
	}
	// A failed login is not fatal: every request retries it, so report the reason once
	// and let the caller keep the client.
	loginErr := clientObject.Login()
	if !clientObject.isLoggedIn() {
		Log.Error(loginErr, "Failed to log in to Omada controller WebUI")
		return clientObject
	}

	Log.Info("New WebAPI client created. URL: %s, Site: %s", baseURL, siteName)
	return clientObject
}

func GetInstance() *WebClient {
	once.Do(func() {
		conf := internal.GetConfig().Omada
		instance = newClient(conf.OmadaURL, conf.Username, conf.Password, conf.SiteName)
	})
	return instance
}

func (c *WebClient) Login() error {
	endpoint := Utils.FillInEndpointPlaceholders(path_login, c.fillInOmadaIDs(nil))
	if endpoint == "" {
		fmt.Println("Endpoint cannot be empty")
		return Log.Error(nil, "Endpoint cannot be empty")
	}

	url, err := Utils.CreateURL(c.BaseURL, endpoint, nil)
	if err != nil {
		return Log.Error(err, "Error creating URL")
	}
	bodyBytes, err := json.Marshal(map[string]string{"username": c.username, "password": c.password})
	if err != nil {
		return Log.Error(err, "Error marshalling request body")
	}

	response, err := c.Client.Post(url, "application/json", bytes.NewReader(bodyBytes))
	if err != nil {
		return Log.Error(err, "Error making POST request to login endpoint")
	}
	if response.StatusCode != http.StatusOK {
		return Log.Error(nil, "Received non-OK status code from login endpoint: %d", response.StatusCode)
	}

	defer response.Body.Close()
	var loginResponse Response[Login]
	if err := json.NewDecoder(response.Body).Decode(&loginResponse); err != nil {
		return Log.Error(err, "Error decoding login response")
	}
	if loginResponse.ErrorCode != 0 {
		return Log.Error(nil, "API error: %s (code %d)", loginResponse.Message, loginResponse.ErrorCode)
	}

	Log.Info("Logged in successfully to Omada controller WebUI. URL: %s, Site: %s", c.BaseURL, c.SiteName)
	c.lastLoginTime = time.Now().Unix()
	c.Token = loginResponse.Result.Token
	return nil
}

func (c *WebClient) isLoggedIn() bool {
	if c.lastLoginCheckTime > 0 && ((time.Now().Unix() - c.lastLoginCheckTime) <= threshold_loginCheck) {
		Log.Debug("Last logged in check performed in less than %d seconds, returning true", threshold_loginCheck)
		return true
	}
	c.lastLoginCheckTime = time.Now().Unix()

	endpoint := Utils.FillInEndpointPlaceholders(path_loginStatus, c.fillInOmadaIDs(nil))
	if endpoint == "" {
		Log.Error(nil, "Endpoint cannot be empty")
		return false
	}

	url, err := Utils.CreateURL(c.BaseURL, endpoint, Utils.AddTimestampParam(nil))
	if err != nil {
		Log.Error(err, "Error creating URL")
		return false
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		Log.Error(err, "Error creating GET request")
		return false
	}

	if err := c.setAuthorizationHeader(req); err != nil {
		Log.Error(err, "Error setting authorization header")
		return false
	}

	response, err := c.Client.Do(req)
	if err != nil {
		Log.Error(err, "Error making GET request to login status endpoint")
		return false
	}
	defer response.Body.Close()
	var result Response[IsLoggedIn]
	if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
		Log.Error(err, "Error decoding response for login status")
		return false
	}
	if result.ErrorCode != 0 {
		Log.Error(nil, "API error: %s (code %d)", result.Message, result.ErrorCode)
		return false
	}

	return result.Result.Login
}

func (c *WebClient) setAuthorizationHeader(req *http.Request) error {
	req.Header.Set("Csrf-Token", c.Token)
	return nil
}

func (c *WebClient) getHttpClient() (*http.Client, error) {
	if !c.isLoggedIn() {
		Log.Warn("WebAPI client is not logged in, re-logging in")
		err := c.Login()
		if err != nil {
			return nil, Log.Error(err, "Failed to re-login to Omada controller WebUI")
		}
	}

	return c.Client, nil
}
