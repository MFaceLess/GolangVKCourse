package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const (
	unknownConnectionError = "unknownConnectionError"
)

type TestCase struct {
	AccessToken string
	Request     SearchRequest
	Result      *SearchResponse
	IsError     bool
	TextError   string
}

func MockServer(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("AccessToken")

	if token == "timeout_error_testing" {
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
		return
	}

	if token == unknownConnectionError {
		hijacker, ok := w.(http.Hijacker)
		if !ok {
			http.Error(w, "Cannot hijack", http.StatusInternalServerError)
			return
		}
		conn, _, err := hijacker.Hijack()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		conn.Close()
	}

	if token == "bad AccessToken" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	if token == "SearchServerInternalError" {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if token == "SearchServerBadRequestUnpackJSON" {
		w.WriteHeader(http.StatusBadRequest)
		_, err := w.Write([]byte("Invalid JSON"))
		if err != nil {
			return
		}
		return
	}

	if token == "ErrorBadOrderField" {
		checkError := SearchErrorResponse{Error: ErrorBadOrderField}
		w.WriteHeader(http.StatusBadRequest)
		err := json.NewEncoder(w).Encode(checkError)
		if err != nil {
			return
		}
		return
	}

	if token == "UnknownBadRequest" {
		checkError := SearchErrorResponse{Error: "UnknownError"}
		w.WriteHeader(http.StatusBadRequest)
		err := json.NewEncoder(w).Encode(checkError)
		if err != nil {
			return
		}
		return
	}

	if token == "Invalid format result" {
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte("Invalid Result JSON"))
		if err != nil {
			return
		}
		return
	}

	if token == "has next page" {
		users := []User{
			{
				ID:     1,
				Name:   "Misha",
				Age:    24,
				About:  "Trying to get better",
				Gender: "male",
			},
			{
				ID:     2,
				Name:   "Max",
				Age:    38,
				About:  "Yep",
				Gender: "male",
			},
		}
		w.WriteHeader(http.StatusOK)
		err := json.NewEncoder(w).Encode(users)
		if err != nil {
			return
		}
		return
	}

	if token == "hasn't next page" {
		users := []User{
			{
				ID:     1,
				Name:   "Misha",
				Age:    24,
				About:  "Trying to get better",
				Gender: "male",
			},
			{
				ID:     2,
				Name:   "Max",
				Age:    38,
				About:  "Yep",
				Gender: "male",
			},
		}
		w.WriteHeader(http.StatusOK)
		err := json.NewEncoder(w).Encode(users)
		if err != nil {
			return
		}
		return
	}

	w.WriteHeader(http.StatusOK)
}

func TestFindUsersInvalidParameters(t *testing.T) {
	cases := []TestCase{
		{
			AccessToken: "token",
			Request: SearchRequest{
				Limit:      -1,
				Offset:     20,
				Query:      "on",
				OrderField: "Age",
				OrderBy:    OrderByAsIs,
			},
			Result:    nil,
			IsError:   true,
			TextError: "limit must be > 0",
		},
		{
			AccessToken: "token",
			Request: SearchRequest{
				Limit:      0,
				Offset:     -1,
				Query:      "on",
				OrderField: "Age",
				OrderBy:    OrderByAsIs,
			},
			Result:    nil,
			IsError:   true,
			TextError: "offset must be > 0",
		},
	}

	ts := httptest.NewServer(http.HandlerFunc(MockServer))
	defer ts.Close()

	for caseNum, item := range cases {
		client := SearchClient{AccessToken: item.AccessToken, URL: ts.URL}

		result, err := client.FindUsers(item.Request)

		if err != nil {
			assert.Equal(t, item.TextError, err.Error(), "[%d] wrong type of error: expected %#v, got %#v", caseNum, item.TextError, err.Error())
		}

		if item.IsError {
			assert.Error(t, err, "[%d] expected error, got nil", caseNum)
		} else {
			assert.NoError(t, err, "[%d] unexpected error: %#v", caseNum, err)
		}

		assert.Equal(t, item.Result, result, "[%d] wrong result, expected %#v, got %#v", caseNum, item.Result, result)
	}
}

func TestBrokenServer(t *testing.T) {
	cases := []TestCase{
		{
			AccessToken: "timeout_error_testing",
			Request: SearchRequest{
				Limit:      5,
				Offset:     0,
				Query:      "on",
				OrderField: "Age",
				OrderBy:    OrderByAsIs,
			},
			Result:    nil,
			IsError:   true,
			TextError: "timeout for limit=6&offset=0&order_by=0&order_field=Age&query=on",
		},
		{
			AccessToken: unknownConnectionError,
			Request: SearchRequest{
				Limit:      5,
				Offset:     0,
				Query:      "on",
				OrderField: "Age",
				OrderBy:    OrderByAsIs,
			},
			Result:  nil,
			IsError: true,
		},
	}

	ts := httptest.NewServer(http.HandlerFunc(MockServer))
	defer ts.Close()

	for caseNum, item := range cases {
		client := SearchClient{AccessToken: item.AccessToken, URL: ts.URL}

		result, err := client.FindUsers(item.Request)

		if err != nil && item.AccessToken != unknownConnectionError {
			assert.Equal(t, item.TextError, err.Error(), "[%d] wrong type of error: expected %#v, got %#v", caseNum, item.TextError, err.Error())
		}

		if item.IsError {
			assert.Error(t, err, "[%d] expected error, got nil", caseNum)
		} else {
			assert.NoError(t, err, "[%d] unexpected error: %#v", caseNum, err)
		}

		assert.Equal(t, item.Result, result, "[%d] wrong result, expected %#v, got %#v", caseNum, item.Result, result)
	}
}

func TestErrorStatusCodes(t *testing.T) {
	cases := []TestCase{
		{
			AccessToken: "bad AccessToken",
			Request: SearchRequest{
				Limit:      5,
				Offset:     0,
				Query:      "on",
				OrderField: "Age",
				OrderBy:    OrderByAsIs,
			},
			Result:    nil,
			IsError:   true,
			TextError: "bad AccessToken",
		},
		{
			AccessToken: "SearchServerInternalError",
			Request: SearchRequest{
				Limit:      5,
				Offset:     0,
				Query:      "on",
				OrderField: "Age",
				OrderBy:    OrderByAsIs,
			},
			Result:    nil,
			IsError:   true,
			TextError: "SearchServer fatal error",
		},
		{
			AccessToken: "SearchServerBadRequestUnpackJSON",
			Request: SearchRequest{
				Limit:      5,
				Offset:     0,
				Query:      "on",
				OrderField: "Age",
				OrderBy:    OrderByAsIs,
			},
			Result:    nil,
			IsError:   true,
			TextError: "cant unpack error json: invalid character 'I' looking for beginning of value",
		},
		{
			AccessToken: "ErrorBadOrderField",
			Request: SearchRequest{
				Limit:      5,
				Offset:     0,
				Query:      "on",
				OrderField: "Incorrect",
				OrderBy:    OrderByAsIs,
			},
			Result:    nil,
			IsError:   true,
			TextError: "OrderFeld Incorrect invalid",
		},
		{
			AccessToken: "UnknownBadRequest",
			Request: SearchRequest{
				Limit:      5,
				Offset:     0,
				Query:      "on",
				OrderField: "Incorrect",
				OrderBy:    OrderByAsIs,
			},
			Result:    nil,
			IsError:   true,
			TextError: "unknown bad request error: UnknownError",
		},
	}

	ts := httptest.NewServer(http.HandlerFunc(MockServer))
	defer ts.Close()

	for caseNum, item := range cases {
		client := SearchClient{AccessToken: item.AccessToken, URL: ts.URL}

		result, err := client.FindUsers(item.Request)
		if err != nil && item.AccessToken != unknownConnectionError {
			assert.Equal(t, item.TextError, err.Error(), "[%d] wrong type of error: expected %#v, got %#v", caseNum, item.TextError, err.Error())
		}

		if item.IsError {
			assert.Error(t, err, "[%d] expected error, got nil", caseNum)
		} else {
			assert.NoError(t, err, "[%d] unexpected error: %#v", caseNum, err)
		}

		assert.Equal(t, item.Result, result, "[%d] wrong result, expected %#v, got %#v", caseNum, item.Result, result)
	}
}

func TestDifferentResults(t *testing.T) {
	cases := []TestCase{
		{
			AccessToken: "Invalid format result",
			Request: SearchRequest{
				Limit:      5,
				Offset:     0,
				Query:      "on",
				OrderField: "Age",
				OrderBy:    OrderByAsIs,
			},
			Result:    nil,
			IsError:   true,
			TextError: "cant unpack result json: invalid character 'I' looking for beginning of value",
		},
		{
			AccessToken: "has next page",
			Request: SearchRequest{
				Limit:      1,
				Offset:     0,
				Query:      "on",
				OrderField: "Age",
				OrderBy:    OrderByAsIs,
			},
			Result: &SearchResponse{
				Users: []User{
					{
						ID:     1,
						Name:   "Misha",
						Age:    24,
						About:  "Trying to get better",
						Gender: "male",
					},
				},
				NextPage: true,
			},
			IsError: false,
		},
		{
			AccessToken: "hasn't next page",
			Request: SearchRequest{
				Limit:      1000,
				Offset:     0,
				Query:      "on",
				OrderField: "Age",
				OrderBy:    OrderByAsIs,
			},
			Result: &SearchResponse{
				Users: []User{
					{
						ID:     1,
						Name:   "Misha",
						Age:    24,
						About:  "Trying to get better",
						Gender: "male",
					},
					{
						ID:     2,
						Name:   "Max",
						Age:    38,
						About:  "Yep",
						Gender: "male",
					},
				},
				NextPage: false,
			},
			IsError: false,
		},
	}

	ts := httptest.NewServer(http.HandlerFunc(MockServer))
	defer ts.Close()

	for caseNum, item := range cases {
		client := SearchClient{AccessToken: item.AccessToken, URL: ts.URL}

		result, err := client.FindUsers(item.Request)
		if err != nil && item.AccessToken != unknownConnectionError {
			assert.Equal(t, item.TextError, err.Error(), "[%d] wrong type of error: expected %#v, got %#v", caseNum, item.TextError, err.Error())
		}

		if item.IsError {
			assert.Error(t, err, "[%d] expected error, got nil", caseNum)
		} else {
			assert.NoError(t, err, "[%d] unexpected error: %#v", caseNum, err)
		}

		assert.Equal(t, item.Result, result, "[%d] wrong result, expected %#v, got %#v", caseNum, item.Result, result)
	}
}
