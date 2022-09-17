package uptime

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/hashicorp/go-retryablehttp"
)

var (
	Version    string
	CommitHash string
)

const (
	defaultBaseUrl = "https://uptime.com/api/v1/"
)

type UptimeClient struct {
	client     *retryablehttp.Client
	baseUrl    string
	subaccount string
	token      string
}

type ClientCredentials struct {
	Token    string
	Email    string
	Password string
}

func NewClient(ctx context.Context, baseUrl, subaccount, token, email, password string) (*UptimeClient, error) {
	if baseUrl == "" {
		baseUrl = defaultBaseUrl
	}

	if token == "" { // use token
		if email != "" && password != "" { // use login credentials
			t, err := credentialToToken(baseUrl, email, password)

			if err != nil {
				return nil, err
			}

			token = t
		} else {
			return nil, fmt.Errorf("no credentials provided")
		}
	}

	client := UptimeClient{
		client:     retryablehttp.NewClient(),
		baseUrl:    baseUrl,
		subaccount: subaccount,
		token:      token,
	}

	return &client, nil
}

func (c *UptimeClient) addRequestHeaders(req *retryablehttp.Request) {
	req.Header.Set("Authorization", fmt.Sprintf("Token %s", c.token))

	if c.subaccount != "" {
		req.Header.Set("X-Subaccount", c.subaccount)
	}

	req.Header.Set("User-Agent", "go-uptime")

	req.Header.Set("Accept", "application/json")

	if req.Method == http.MethodPost || req.Method == http.MethodPut || req.Method == http.MethodDelete {
		req.Header.Set("Content-type", "application/json")
	}
}

func (c *UptimeClient) do(ctx context.Context, req *retryablehttp.Request, body []byte) ([]byte, error) {
	if body != nil {
		req.Body = ioutil.NopCloser(bytes.NewReader(body))
	}

	c.addRequestHeaders(req)

	response, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %v", err)
	}

	if response.StatusCode == http.StatusUnauthorized || response.StatusCode == http.StatusForbidden {
		return nil, fmt.Errorf("token authentication failed")
	}

	defer response.Body.Close()

	responseBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	if response.StatusCode >= 400 {
		errorMessage := fmt.Errorf("error sending %s request to %s: %s.", req.Method, req.URL.Path, response.Status)

		if len(responseBody) != 0 {
			errorMessage = fmt.Errorf("%s response body: %s", errorMessage, responseBody)
		}

		return nil, errorMessage
	}

	return responseBody, nil
}

func (c *UptimeClient) getRaw(ctx context.Context, path string, params map[string]string) ([]byte, error) {
	request, err := retryablehttp.NewRequestWithContext(ctx, http.MethodGet, c.baseUrl+path, nil)
	if err != nil {
		return nil, err
	}

	if params != nil {
		query := url.Values{}
		for k, v := range params {
			query.Add(k, v)
		}
		request.URL.RawQuery = query.Encode()
	}

	body, err := c.do(ctx, request, nil)
	return body, err
}

func (c *UptimeClient) get(ctx context.Context, path string, resource interface{}, params map[string]string) error {
	body, err := c.getRaw(ctx, path, params)
	if err != nil {
		return err
	}

	return json.Unmarshal(body, resource)
}

func (c *UptimeClient) post(ctx context.Context, path string, requestBody interface{}) ([]byte, error) {
	payload, err := json.Marshal(requestBody)
	if err != nil {
		return nil, err
	}

	request, err := retryablehttp.NewRequestWithContext(ctx, http.MethodPost, c.baseUrl+path, nil)
	if err != nil {
		return nil, err
	}

	body, err := c.do(ctx, request, payload)

	return body, err
}

func (c *UptimeClient) put(ctx context.Context, path string, requestBody interface{}) error {
	payload, err := json.Marshal(requestBody)
	if err != nil {
		return err
	}

	request, err := retryablehttp.NewRequestWithContext(ctx, http.MethodPut, c.baseUrl+path, nil)
	if err != nil {
		return err
	}

	_, err = c.do(ctx, request, payload)

	return err
}

func (c *UptimeClient) delete(ctx context.Context, path string, requestBody interface{}) error {
	var (
		payload []byte
		err     error
	)

	if requestBody != nil {
		payload, err = json.Marshal(requestBody)
		if err != nil {
			return err
		}
	}

	request, err := retryablehttp.NewRequestWithContext(ctx, http.MethodDelete, c.baseUrl+path, nil)
	if err != nil {
		return err
	}

	_, err = c.do(ctx, request, payload)

	return err
}

// get an access token using login credentials
func credentialToToken(baseUrl, email, password string) (string, error) {
	c := retryablehttp.NewClient()

	resp, err := c.PostForm(baseUrl+"auth/login/", url.Values{
		"email":    {email},
		"password": {password},
	})

	if err != nil {
		return "", fmt.Errorf("credentials to token failed: %v", err)
	}

	var res map[string]interface{}

	json.NewDecoder(resp.Body).Decode(&res)

	token := fmt.Sprintf("%v", res["access_token"]) // convert to string

	return token, nil
}
