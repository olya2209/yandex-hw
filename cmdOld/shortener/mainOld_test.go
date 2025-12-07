package cmdOld

import (
	"bytes"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// cleaner before test
func setup() {
	urlsMap = make(map[string]string)
	rand.Seed(time.Now().UnixNano())
}

const (
	localHost = "localhost:8080"
)

type want struct {
	contentType string
	statusCode  int
	lenShortUrl int
	bodyText    string
}

type testCase struct {
	name    string
	request string
	url     string
	metod   string
	want    want
}

func testRequest(t *testing.T, ts *httptest.Server, method,
	path, url string) (*http.Response, string) {
	body := []byte(url)
	req, err := http.NewRequest(method, ts.URL+path, bytes.NewBuffer(body))
	req.Host = localHost
	require.NoError(t, err)

	resp, err := ts.Client().Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	return resp, string(respBody)
}

func TestRouter(t *testing.T) {
	ts := httptest.NewServer(registrationRoutes())
	defer ts.Close()
	testTable := []testCase{
		{
			name:    "success post request",
			request: "/",
			url:     "https://test.com",
			metod:   http.MethodPost,
			want: want{
				contentType: "text/plain",
				statusCode:  http.StatusCreated,
				lenShortUrl: 8,
			},
		},
		{
			name:    "nil body",
			request: "/",
			metod:   http.MethodPost,
			want: want{
				statusCode: http.StatusBadRequest,
				bodyText:   "Ошибка чтения тела запроса\n",
			},
		},
		{
			name:    "incorrect route for post request",
			request: "/test.com",
			metod:   http.MethodPost,
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name:    "incorrect route for get request - id have incorrect symbols",
			request: "/test.com",
			metod:   http.MethodGet,
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name:    "incorrect route for get request - id length is not equal 8 symbols",
			request: "/1234",
			metod:   http.MethodGet,
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name:    "incorrect route method - delete",
			request: "/12345678",
			metod:   http.MethodDelete,
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name:    "incorrect route method - connect",
			request: "/12345678",
			metod:   http.MethodConnect,
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name:    "incorrect route method - head",
			request: "/12345678",
			metod:   http.MethodHead,
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name:    "incorrect route method - options",
			request: "/12345678",
			metod:   http.MethodOptions,
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name:    "incorrect route method - patch",
			request: "/12345678",
			metod:   http.MethodPatch,
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name:    "incorrect route method - trace",
			request: "/12345678",
			metod:   http.MethodTrace,
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name:    "incorrect route method - put",
			request: "/12345678",
			metod:   http.MethodPut,
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},

		{
			name:    "incorrect route method - delete",
			request: "/",
			metod:   http.MethodDelete,
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name:    "incorrect route method - connect",
			request: "/",
			metod:   http.MethodConnect,
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name:    "incorrect route method - head",
			request: "/",
			metod:   http.MethodHead,
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name:    "incorrect route method - options",
			request: "/",
			metod:   http.MethodOptions,
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name:    "incorrect route method - patch",
			request: "/",
			metod:   http.MethodPatch,
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name:    "incorrect route method - trace",
			request: "/",
			metod:   http.MethodTrace,
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},
		{
			name:    "incorrect route method - put",
			request: "/",
			metod:   http.MethodPut,
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},
	}
	for _, tt := range testTable {
		resp, _ := testRequest(t, ts, tt.metod, tt.request, tt.url)
		assert.Equal(t, tt.want.statusCode, resp.StatusCode)
		if tt.want.statusCode == http.StatusBadRequest && tt.want.bodyText == "" {
			//	assert.Equal(t, http.StatusText(tt.want.statusCode), strings.TrimSpace(resBody))
		}
	}
}

func TestHandleCreateShortURLRedirect_Success(t *testing.T) {
	tests := []testCase{
		{
			name:    "success post request",
			request: "/",
			url:     "https://test.com",
			want: want{
				contentType: "text/plain",
				statusCode:  http.StatusCreated,
				lenShortUrl: 8,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorderPost := httptest.NewRecorder()

			body := []byte(tt.url)

			requestPost := httptest.NewRequest(http.MethodPost, tt.request, bytes.NewBuffer(body))
			requestPost.Host = localHost

			h := http.HandlerFunc(handleCreateShortURL)
			h(recorderPost, requestPost)

			resultPost := recorderPost.Result()

			assert.Equal(t, tt.want.statusCode, resultPost.StatusCode)
			assert.Equal(t, tt.want.contentType, resultPost.Header.Get("Content-Type"))

			shortUrl, err := ioutil.ReadAll(resultPost.Body)
			require.NoError(t, err)
			err = resultPost.Body.Close()
			require.NoError(t, err)

			key := shortUrl[len(shortUrl)-tt.want.lenShortUrl:]
			value, ok := GetValue(string(key))
			if !ok {
				t.Errorf("Не найдено значение короткого url в мапе по ключу %d", key)
			}

			assert.Equal(t, tt.url, value)
			assert.Equal(t, tt.want.lenShortUrl, len(key))

			//Проверяем get
			recorderGet := httptest.NewRecorder()

			requestG := "/" + string(key)
			requestGet := httptest.NewRequest(http.MethodGet, requestG, nil)
			requestGet.Host = localHost

			hGet := http.HandlerFunc(handleRedirect)
			hGet(recorderGet, requestGet)

			resultGet := recorderGet.Result()
			assert.Equal(t, http.StatusTemporaryRedirect, resultGet.StatusCode)
			require.Equal(t, value, resultGet.Header.Get("Location"))
		})
	}
}

func TestHandleCreateShortURL_Fail(t *testing.T) {
	tests := []testCase{
		{
			name:    "nil body",
			request: "/",
			metod:   http.MethodPost,
			want: want{
				statusCode: http.StatusBadRequest,
				bodyText:   "Ошибка чтения тела запроса\n",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()

			body := []byte(tt.url)
			request := httptest.NewRequest(http.MethodPost, tt.request, bytes.NewBuffer(body))
			request.Host = localHost

			h := http.HandlerFunc(handleCreateShortURL)

			h(recorder, request)

			result := recorder.Result()

			resBody, err := ioutil.ReadAll(result.Body)
			require.NoError(t, err)
			err = result.Body.Close()
			require.NoError(t, err)

			assert.Equal(t, tt.want.statusCode, result.StatusCode)
			assert.Equal(t, tt.want.bodyText, string(resBody))
		})
	}
}

// TODO dont work - must be 400, have 404

//func TestRegistrationRoutes_BadRoute(t *testing.T) {
//	registrationRoutes()
//	tests := []testCase{
//		{
//			name:    "incorrect route for post request",
//			request: "/test.com",
//			metod:   http.MethodPost,
//			want: want{
//				statusCode: http.StatusBadRequest,
//			},
//		},
//		{
//			name:    "incorrect route for get request - id have incorrect symbols",
//			request: "/test.com",
//			metod:   http.MethodGet,
//			want: want{
//				statusCode: http.StatusBadRequest,
//			},
//		},
//		{
//			name:    "incorrect route for get request - id length is not equal 8 symbols",
//			request: "/1234",
//			metod:   http.MethodGet,
//			want: want{
//				statusCode: http.StatusBadRequest,
//			},
//		},
//		{
//			name:    "incorrect route method - delete",
//			request: "/12345678",
//			metod:   http.MethodDelete,
//			want: want{
//				statusCode: http.StatusBadRequest,
//			},
//		},
//		{
//			name:    "incorrect route method - connect",
//			request: "/12345678",
//			metod:   http.MethodConnect,
//			want: want{
//				statusCode: http.StatusBadRequest,
//			},
//		},
//		{
//			name:    "incorrect route method - head",
//			request: "/12345678",
//			metod:   http.MethodHead,
//			want: want{
//				statusCode: http.StatusBadRequest,
//			},
//		},
//		{
//			name:    "incorrect route method - options",
//			request: "/12345678",
//			metod:   http.MethodOptions,
//			want: want{
//				statusCode: http.StatusBadRequest,
//			},
//		},
//		{
//			name:    "incorrect route method - patch",
//			request: "/12345678",
//			metod:   http.MethodPatch,
//			want: want{
//				statusCode: http.StatusBadRequest,
//			},
//		},
//		{
//			name:    "incorrect route method - trace",
//			request: "/12345678",
//			metod:   http.MethodTrace,
//			want: want{
//				statusCode: http.StatusBadRequest,
//			},
//		},
//		{
//			name:    "incorrect route method - put",
//			request: "/12345678",
//			metod:   http.MethodPut,
//			want: want{
//				statusCode: http.StatusBadRequest,
//			},
//		},
//
//		{
//			name:    "incorrect route method - delete",
//			request: "/",
//			metod:   http.MethodDelete,
//			want: want{
//				statusCode: http.StatusBadRequest,
//			},
//		},
//		{
//			name:    "incorrect route method - connect",
//			request: "/",
//			metod:   http.MethodConnect,
//			want: want{
//				statusCode: http.StatusBadRequest,
//			},
//		},
//		{
//			name:    "incorrect route method - head",
//			request: "/",
//			metod:   http.MethodHead,
//			want: want{
//				statusCode: http.StatusBadRequest,
//			},
//		},
//		{
//			name:    "incorrect route method - options",
//			request: "/",
//			metod:   http.MethodOptions,
//			want: want{
//				statusCode: http.StatusBadRequest,
//			},
//		},
//		{
//			name:    "incorrect route method - patch",
//			request: "/",
//			metod:   http.MethodPatch,
//			want: want{
//				statusCode: http.StatusBadRequest,
//			},
//		},
//		{
//			name:    "incorrect route method - trace",
//			request: "/",
//			metod:   http.MethodTrace,
//			want: want{
//				statusCode: http.StatusBadRequest,
//			},
//		},
//		{
//			name:    "incorrect route method - put",
//			request: "/",
//			metod:   http.MethodPut,
//			want: want{
//				statusCode: http.StatusBadRequest,
//			},
//		},
//	}
//
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			recorder := httptest.NewRecorder()
//			body := []byte(tt.request)
//			request := httptest.NewRequest(tt.metod, tt.request, bytes.NewBuffer(body))
//			request.Host = localHost
//
//			http.DefaultServeMux.ServeHTTP(recorder, request)
//
//			resBody, _ := ioutil.ReadAll(recorder.Body)
//
//			assert.Equal(t, tt.want.statusCode, recorder.Code)
//			assert.Equal(t, http.StatusText(tt.want.statusCode), strings.TrimSpace(string(resBody)))
//		})
//	}
//}
