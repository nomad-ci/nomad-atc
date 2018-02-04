package cloud

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/SHyx0rmZ/go-bitbucket/bitbucket"
	"io"
	"net/http"
	"reflect"
	"strings"
)

type Client struct {
	httpClient *http.Client
	endpoint   string
	auth       bitbucket.Auth
}

func NewClient(httpClient *http.Client, endpoint string) (*Client, error) {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	if endpoint == "" {
		endpoint = "https://api.bitbucket.org/"
	}
	return &Client{
		httpClient: httpClient,
		endpoint:   endpoint,
	}, nil
}

func (c *Client) SetBasicAuth(auth *bitbucket.BasicAuth) {
	c.auth = auth
}

func (c *Client) CurrentUser() (string, error) {
	var u user

	err := c.request("/2.0/user", &u)
	if err != nil {
		return "", err
	}

	return u.Name, nil
}

func (c *Client) Users() ([]bitbucket.User, error) {
	/* Bitbucket cloud does not allow access to a list of all users. */
	return []bitbucket.User{}, nil
}

func (c *Client) Projects() ([]bitbucket.Project, error) {
	return nil, errors.New("not implemented")
}

func (c *Client) CreateRepository(path string) (bitbucket.Repository, error) {
	return nil, errors.New("not yet implemented")
}

func (c *Client) Repository(path string) (bitbucket.Repository, error) {
	var r repository

	if strings.Contains(path, "..") {
		return nil, errors.New("no recursive paths allowed")
	}

	err := c.request("/2.0/repositories/"+path, &r)
	if err != nil {
		return nil, err
	}

	return &r, nil
}

func (c *Client) Repositories() ([]bitbucket.Repository, error) {
	repositories := make([]repository, 0, 0)

	err := c.pagedRequest("/2.0/repositories?role=member", &repositories)
	if err != nil {
		return nil, err
	}

	bitbucketRepositories := make([]bitbucket.Repository, len(repositories))
	for index := range repositories {
		bitbucketRepositories[index] = &repositories[index]
	}

	return bitbucketRepositories, nil
}

func (c *Client) Teams() ([]team, error) {
	teams := make([]team, 0, 0)

	err := c.pagedRequest("/2.0/teams", &teams)
	if err != nil {
		return nil, err
	}

	return teams, nil
}

func (c *Client) TeamsWithRole(role string) ([]team, error) {
	teams := make([]team, 0, 0)

	err := c.pagedRequest("/2.0/teams?role="+role, &teams)
	if err != nil {
		return nil, err
	}

	return teams, nil
}

func (c *Client) getUrl(apiResource string) (url string) {
	return strings.TrimRight(c.endpoint, "/") + apiResource
}

func (c *Client) do(method string, url string, body io.Reader) (*http.Response, error) {
	request, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	if basicAuth, ok := c.auth.(*bitbucket.BasicAuth); ok {
		request.SetBasicAuth(basicAuth.Username, basicAuth.Password)
	}

	if method == "POST" {
		request.Header.Set("Content-Type", "application/json")
	}

	response, err := c.httpClient.Do(request)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (c *Client) request(apiResource string, v interface{}) error {
	response, err := c.do("GET", c.getUrl(apiResource), strings.NewReader(""))
	if err != nil {
		return err
	}

	defer response.Body.Close()

	if response.StatusCode != 200 {
		return errors.New(response.Status)
	}

	decoder := json.NewDecoder(response.Body)

	err = decoder.Decode(v)
	if err != nil {
		return err
	}

	if ca, ok := v.(clientAware); ok {
		ca.SetClient(c)
	}

	return nil
}

func (c *Client) requestPost(apiResource string, v interface{}, data interface{}) error {
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return err
	}

	response, err := c.do("POST", c.getUrl(apiResource), bytes.NewBuffer(jsonBytes))
	if err != nil {
		return err
	}

	defer response.Body.Close()

	if response.StatusCode != 201 {
		return errors.New(response.Status)
	}

	decoder := json.NewDecoder(response.Body)

	err = decoder.Decode(v)
	if err != nil {
		return err
	}

	if ca, ok := v.(clientAware); ok {
		ca.SetClient(c)
	}

	return nil
}

type clientAware interface {
	SetClient(c *Client)
}

func (c *Client) SetHTTPClient(hc *http.Client) {
	c.httpClient = hc
}

type PagedResult struct {
	PageLength  bool              `json:"isLastPage"`
	Values      []json.RawMessage `json:"values,omitempty"`
	NextPageURL string            `json:"next,omitempty"`
}

func (c *Client) pagedRequest(apiResource string, v interface{}) error {
	resultValue := reflect.ValueOf(v)

	if resultValue.Kind() != reflect.Ptr || resultValue.IsNil() {
		return errors.New("invalid return type")
	}

	resultList := reflect.ValueOf(v).Elem()
	resultElemType := resultList.Type().Elem()

	url := c.getUrl(apiResource)

	for {
		var results PagedResult

		response, err := c.do("GET", url, strings.NewReader(""))
		if err != nil {
			return err
		}

		if response.StatusCode != 200 {
			response.Body.Close()
			return errors.New(response.Status)
		}

		decoder := json.NewDecoder(response.Body)

		err = decoder.Decode(&results)
		response.Body.Close()
		if err != nil {
			return err
		}

		for _, jsonBytes := range results.Values {
			newResult := reflect.New(resultElemType).Elem()

			err = json.Unmarshal(jsonBytes, newResult.Addr().Interface())
			if err != nil {
				return err
			}

			if ca, ok := newResult.Addr().Interface().(clientAware); ok {
				ca.SetClient(c)
			}

			resultList.Set(reflect.Append(resultList, newResult))
		}

		if results.NextPageURL == "" {
			break
		}

		url = results.NextPageURL
	}

	return nil
}
