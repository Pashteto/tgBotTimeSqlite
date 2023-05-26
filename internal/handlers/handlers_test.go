package handlers

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/caarlos0/env/v6"

	"github.com/Pashteto/tgBotTimeSqlite/repos"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/Pashteto/tgBotTimeSqlite/config"
	filedb "github.com/Pashteto/tgBotTimeSqlite/filed_history"
)

// Test of GetHandler
func TestHandlersWithDBStore_GetHandler(t *testing.T) {
	type fields struct {
		rdb repos.SetterGetter //redis.Client

		code           int
		headerLocation string
		contentType    string
		id             string
		method         string
		conf           *config.Config
	}
	repoMock := new(repositoryMock)
	repoMock.On("Ping", mock.MatchedBy(func(_ context.Context) bool { return true })).Return(nil)
	repoMock.On("Get", mock.MatchedBy(func(_ context.Context) bool { return true }), "this_id_is_a_wrong_id").Return("", nil)
	repoMock.On("Get", mock.MatchedBy(func(_ context.Context) bool { return true }), "this_id_is_a_correct_id").Return("http://example.com", nil)

	var conf config.Config

	err := env.Parse(&conf)

	/*err := config.ReadFile(&conf)
	 */

	if err != nil {
		t.Errorf("Unable to read config file conf.json:\t%v", err)
		return
	}

	tests := []struct {
		name   string
		fields fields
	}{

		// tests list
		{
			name: "Test 1: Get Handler with wrong id",
			fields: fields{
				rdb:            repoMock,
				code:           400,
				id:             "this_id_is_a_wrong_id",
				contentType:    "text/plain; charset=utf-8",
				headerLocation: "",
				method:         "GET",
				conf:           &conf,
			},
		},
		{
			name: "Test 2: Get Handler with correct id",
			fields: fields{
				rdb:            repoMock,
				code:           307,
				contentType:    "text/html; charset=utf-8",
				id:             "this_id_is_a_correct_id",
				headerLocation: "http://example.com",

				method: "GET",
				conf:   &conf,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			hObj := &HandlersWithDBStore{
				Rdb:  tt.fields.rdb,
				Conf: &conf,
			}
			request := httptest.NewRequest(tt.fields.method, "/"+tt.fields.id, nil)
			w := httptest.NewRecorder()
			h := http.HandlerFunc(hObj.GetHandler)
			h.ServeHTTP(w, request)
			res := w.Result()
			defer res.Body.Close()
			// test StatusCode
			if res.StatusCode != tt.fields.code {
				t.Errorf("Expected status code %d, got %d", tt.fields.code, w.Code)
			}
			// test Header Location
			if res.Header.Get("Location") != tt.fields.headerLocation {
				t.Errorf("Expected Location %s, got %s", tt.fields.headerLocation, res.Header.Get("Location"))
			}
			// test Header Content-Type
			assert.Equal(t, tt.fields.contentType, res.Header.Get("Content-Type"),
				"Expected Content-Type %s, got %s", tt.fields.contentType, res.Header.Get("Content-Type"))
		})
	}
}

// Test of PostHandler
func TestHandlersWithDBStore_PostHandler(t *testing.T) {
	type fields struct {
		rdb repos.SetterGetter //redis.Client

		code        int
		postAddress string
		response    string
		method      string
		conf        *config.Config
	}

	repoMock := new(repositoryMock)
	repoMock.On("Ping", mock.MatchedBy(func(_ context.Context) bool { return true })).Return(nil)
	repoMock.On("ListAllKeys",
		mock.MatchedBy(func(_ context.Context) bool { return true })).Return([]string{}, nil)

	repoMock.On("Set",
		mock.MatchedBy(func(_ context.Context) bool { return true }),
		mock.MatchedBy(func(_ string) bool { return true }),
		"http://example.com",
		mock.MatchedBy(func(_ time.Duration) bool { return true })).Return(nil)

	repoMock.On("Set", mock.MatchedBy(func(_ context.Context) bool { return true }),
		mock.MatchedBy(func(_ string) bool { return true }), "",
		mock.MatchedBy(func(_ time.Duration) bool { return true })).Return(nil)

	repoMock.On("Get",
		mock.MatchedBy(func(_ context.Context) bool { return true }),
		mock.MatchedBy(func(_ string) bool { return true })).Return("", nil)

	var conf config.Config
	err := env.Parse(&conf)

	if err != nil {
		t.Errorf("Unable to read config file conf.json:\t%v", err)
		return
	}
	err = filedb.CreateDirFileDBExists(conf)
	if err != nil {
		t.Errorf("Test tmp file creating error:\t%v", err)
		return
	}

	tests := []struct {
		name   string
		fields fields
	}{
		// tests list
		{
			name: "Test 1: Post Handler correct response",
			fields: fields{
				rdb: repoMock,

				code:        201,
				postAddress: "http://example.com",
				response:    "http://localhost:8080/81",
				method:      "POST",
				conf:        &conf,
			},
		},
		{
			name: "Test 2: Post Handler empty body",
			fields: fields{
				rdb:         repoMock,
				code:        400,
				postAddress: "",
				response:    "No URL recieved\n",
				method:      "POST",
				conf:        &conf,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hObj := &HandlersWithDBStore{
				Rdb:  tt.fields.rdb,
				Conf: &conf,
			}

			endpoint := "http://localhost:8080/"
			data := tt.fields.postAddress

			request, err := http.NewRequest(tt.fields.method, endpoint, bytes.NewBufferString(data))
			if err != nil {
				t.Fatal(err)
			}
			w := httptest.NewRecorder()
			h := http.HandlerFunc(hObj.PostHandler)
			h.ServeHTTP(w, request)
			res := w.Result()

			// test StatusCode
			assert.Equal(t, tt.fields.code, res.StatusCode,
				"Expected status code %d, got %d", tt.fields.code, w.Code)

			// reading body
			defer res.Body.Close()
			resBody, err := ioutil.ReadAll(res.Body)
			if err != nil {
				t.Fatal(err)
			}
			if string(resBody) != tt.fields.response {
				t.Errorf("Expected body %s, got %s", tt.fields.response, w.Body.String())
			}
		})
	}
}

// Test of PostHandlerJSON
func TestHandlersWithDBStore_PostHandlerJSON(t *testing.T) {
	type fields struct {
		rdb         repos.SetterGetter //redis.Client
		code        int
		postAddress string
		response    string
		method      string
		conf        *config.Config
	}
	repoMock := new(repositoryMock)
	repoMock.On("Ping", mock.MatchedBy(func(_ context.Context) bool { return true })).Return(nil)
	repoMock.On("ListAllKeys", mock.MatchedBy(func(_ context.Context) bool { return true })).Return([]string{}, nil)

	repoMock.On("Set",
		mock.MatchedBy(func(_ context.Context) bool { return true }),
		mock.MatchedBy(func(_ string) bool { return true }),
		"http://example.com",
		mock.MatchedBy(func(_ time.Duration) bool { return true })).Return(nil)

	repoMock.On("Set", mock.MatchedBy(func(_ context.Context) bool { return true }),
		mock.MatchedBy(func(_ string) bool { return true }), "",
		mock.MatchedBy(func(_ time.Duration) bool { return true })).Return(nil)

	repoMock.On("Get",
		mock.MatchedBy(func(_ context.Context) bool { return true }),
		mock.MatchedBy(func(_ string) bool { return true })).Return("", nil)

	var conf config.Config
	err := env.Parse(&conf)

	if err != nil {
		t.Errorf("Unable to read config file conf.json:\t%v", err)
		return
	}
	err = filedb.CreateDirFileDBExists(conf)
	if err != nil {
		t.Errorf("Test tmp file creating error:\t%v", err)
		return
	}
	tests := []struct {
		name   string
		fields fields
	}{
		// tests list
		{
			name: "Test 1: Post Handler correct response",
			fields: fields{
				rdb:         repoMock,
				code:        201,
				postAddress: "{\"url\":\"http://example.com\"}",
				response:    "{\"result\":\"http://localhost:8080/887\"}",
				method:      "POST",
				conf:        &conf,
			},
		},
		{
			name: "Test 2: Post Handler empty JSON",
			fields: fields{
				rdb:         repoMock,
				code:        400,
				postAddress: "",
				response:    "unable to unmarshal JSON\n",
				method:      "POST",
				conf:        &conf,
			},
		},
		{
			name: "Test 3: Post Handler JSON w/o URL",
			fields: fields{
				rdb:         repoMock,
				code:        400,
				postAddress: "{\"url\":\"\"}",
				response:    "No URL recieved\n",
				method:      "POST",
				conf:        &conf,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hObj := &HandlersWithDBStore{
				Rdb:  tt.fields.rdb,
				Conf: &conf,
			}

			endpoint := "http://localhost:8080/"

			data := tt.fields.postAddress
			request, err := http.NewRequest(tt.fields.method, endpoint, bytes.NewBufferString(string(data)))
			if err != nil {
				t.Fatal(err)
			}
			w := httptest.NewRecorder()
			h := http.HandlerFunc(hObj.PostHandlerJSON)
			h.ServeHTTP(w, request)
			res := w.Result()

			// test StatusCode
			assert.Equal(t, tt.fields.code, res.StatusCode,
				"Expected status code %d, got %d", tt.fields.code, w.Code)

			// reading body
			defer res.Body.Close()
			resBody, err := ioutil.ReadAll(res.Body)
			if err != nil {
				t.Fatal(err)
			}
			if string(resBody) != tt.fields.response {
				t.Errorf("Expected body %s, got %s", tt.fields.response, w.Body.String())
			}
		})
	}

}

// repository represent the MOCK of repository model
type repositoryMock struct {
	mock.Mock
}

// Set attaches the MOCK repository and set the data
func (r *repositoryMock) Set(ctx context.Context, key string, value interface{}, exp time.Duration) error {
	args := r.Called(ctx, key, value, exp)
	return args.Error(0)
}

// Get attaches the MOCK repository and get the data
func (r *repositoryMock) Get(ctx context.Context, key string) (string, error) {
	args := r.Called(ctx, key)
	return args.String(0), args.Error(1)
}

func (r *repositoryMock) Ping(ctx context.Context) error {
	args := r.Called(ctx)
	return args.Error(0)
}
func (r *repositoryMock) ListAllKeys(ctx context.Context) ([]string, error) {
	args := r.Called(ctx)
	//	var sdfv []string
	return []string{}, args.Error(1)
}

func (r *repositoryMock) FlushAllKeys(ctx context.Context) error {
	args := r.Called(ctx)
	return args.Error(0)
}
