package main

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	errInvalidResponseWriterObj = errors.New("write error")
	errCantMarhalJSON           = errors.New("json: error calling MarshalJSON for type main.InvalidUserXMLData: write error")
)

type ServerTestCase struct {
	AccessToken string
	Request     SearchRequest
	Response    []User
	StatusCode  int
}

type mockResponseWriter struct {
	headerWritten bool
	err           error
}

func (m *mockResponseWriter) Header() http.Header {
	return http.Header{}
}

func (m *mockResponseWriter) Write(bytes []byte) (int, error) {
	return 0, m.err
}

func (m *mockResponseWriter) WriteHeader(statusCode int) {
	m.headerWritten = true
}

func TestInvalidTokenAndQueryParams(t *testing.T) {
	cases := []ServerTestCase{
		{
			AccessToken: "",
			Request: SearchRequest{
				Limit:      5,
				Offset:     5,
				Query:      "",
				OrderField: "",
				OrderBy:    OrderByAsIs,
			},
			Response:   []User{},
			StatusCode: http.StatusUnauthorized,
		},
		{
			AccessToken: "valid_token",
			Request: SearchRequest{
				Limit:      5,
				Offset:     5,
				Query:      "",
				OrderField: "InvalidOrderField",
				OrderBy:    OrderByAsIs,
			},
			Response:   []User{},
			StatusCode: http.StatusBadRequest,
		},
		{
			AccessToken: "valid_token",
			Request: SearchRequest{
				Limit:      5,
				Offset:     5,
				Query:      "",
				OrderField: "",
				OrderBy:    1000,
			},
			Response:   []User{},
			StatusCode: http.StatusBadRequest,
		},
	}

	for caseNum, item := range cases {
		addr, err := url.Parse("http://example.com/api/")
		if err != nil {
			continue
		}
		params := url.Values{}

		params.Add("limit", strconv.Itoa(item.Request.Limit))
		params.Add("offset", strconv.Itoa(item.Request.Offset))
		params.Add("query", item.Request.Query)
		params.Add("order_field", item.Request.OrderField)
		params.Add("order_by", strconv.Itoa(item.Request.OrderBy))

		addr.RawQuery = params.Encode()

		req := httptest.NewRequest(http.MethodGet, addr.String(), nil)
		req.Header.Set("AccessToken", item.AccessToken)
		w := httptest.NewRecorder()

		SearchServer(w, req)

		assert.Equal(t, item.StatusCode, w.Code, "[%d] wrong StatusCode: got %d, expected %d", caseNum, w.Code, item.StatusCode)

		resp := w.Result()
		body, err := io.ReadAll(resp.Body)

		assert.NoError(t, err, "[%d] failed to read body: %v", caseNum, err)

		if w.Code != http.StatusOK {
			continue
		}

		data := []User{}
		err = json.Unmarshal(body, &data)

		assert.NoError(t, err, "[%d] failed to Unmarshal usersData", caseNum)

		assert.Equal(t, item.Response, data, "[%d] wrong result, expected %#v, got %#v", caseNum, item.Response, data)
	}
}

func TestInvalidLimit(t *testing.T) {
	addr, err := url.Parse("http://example.com/api/")
	if err != nil {
		return
	}
	params := url.Values{}

	params.Add("limit", "incorrectLimit")
	params.Add("offset", "5")
	params.Add("query", "")
	params.Add("order_field", "")
	params.Add("order_by", "1")

	addr.RawQuery = params.Encode()

	req := httptest.NewRequest(http.MethodGet, addr.String(), nil)
	req.Header.Set("AccessToken", "validToken")
	w := httptest.NewRecorder()

	SearchServer(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code, "wrong StatusCode: got %d, expected %d", w.Code, http.StatusBadRequest)
}

func TestInvalidOffset(t *testing.T) {
	addr, err := url.Parse("http://example.com/api/")
	if err != nil {
		return
	}
	params := url.Values{}

	params.Add("limit", "5")
	params.Add("offset", "incorrectLimit")
	params.Add("query", "")
	params.Add("order_field", "")
	params.Add("order_by", "1")

	addr.RawQuery = params.Encode()

	req := httptest.NewRequest(http.MethodGet, addr.String(), nil)
	req.Header.Set("AccessToken", "validToken")
	w := httptest.NewRecorder()

	SearchServer(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code, "wrong StatusCode: got %d, expected %d", w.Code, http.StatusBadRequest)
}

func TestInvalidDataBase(t *testing.T) {
	DataFileName = "not_exists_database.xml"
	defer func() { DataFileName = "dataset.xml" }()

	addr, err := url.Parse("http://example.com/api/")
	if err != nil {
		return
	}
	params := url.Values{}

	params.Add("limit", "5")
	params.Add("offset", "5")
	params.Add("query", "")
	params.Add("order_field", "")
	params.Add("order_by", "1")

	addr.RawQuery = params.Encode()

	req := httptest.NewRequest(http.MethodGet, addr.String(), nil)
	req.Header.Set("AccessToken", "validToken")
	w := httptest.NewRecorder()

	SearchServer(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code, "wrong StatusCode: got %d, expected %d", w.Code, http.StatusInternalServerError)
}

func TestInvalidFileStructure(t *testing.T) {
	DataFileName = "incorrectData.xml"
	defer func() { DataFileName = "dataset.xml" }()

	addr, err := url.Parse("http://example.com/api/")
	if err != nil {
		return
	}
	params := url.Values{}

	params.Add("limit", "5")
	params.Add("offset", "5")
	params.Add("query", "")
	params.Add("order_field", "")
	params.Add("order_by", "1")

	addr.RawQuery = params.Encode()

	req := httptest.NewRequest(http.MethodGet, addr.String(), nil)
	req.Header.Set("AccessToken", "validToken")
	w := httptest.NewRecorder()

	SearchServer(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code, "wrong StatusCode: got %d, expected %d", w.Code, http.StatusInternalServerError)
}

func TestDifferentValidCases(t *testing.T) {
	cases := []ServerTestCase{
		{
			AccessToken: "valid_token",
			Request: SearchRequest{
				Limit:      5,
				Offset:     5,
				Query:      "",
				OrderField: "",
				OrderBy:    OrderByAsIs,
			},
			Response: []User{
				{ID: 5, Name: "Beulah Stark", Age: 30, About: "Enim cillum eu cillum velit labore. In sint esse nulla occaecat voluptate pariatur aliqua aliqua non officia nulla aliqua. Fugiat nostrud irure officia minim cupidatat laborum ad incididunt dolore. Fugiat nostrud eiusmod ex ea nulla commodo. Reprehenderit sint qui anim non ad id adipisicing qui officia Lorem.\n", Gender: "female"},
				{ID: 6, Name: "Jennings Mays", Age: 39, About: "Veniam consectetur non non aliquip exercitation quis qui. Aliquip duis ut ad commodo consequat ipsum cupidatat id anim voluptate deserunt enim laboris. Sunt nostrud voluptate do est tempor esse anim pariatur. Ea do amet Lorem in mollit ipsum irure Lorem exercitation. Exercitation deserunt adipisicing nulla aute ex amet sint tempor incididunt magna. Quis et consectetur dolor nulla reprehenderit culpa laboris voluptate ut mollit. Qui ipsum nisi ullamco sit exercitation nisi magna fugiat anim consectetur officia.\n", Gender: "male"},
				{ID: 7, Name: "Leann Travis", Age: 34, About: "Lorem magna dolore et velit ut officia. Cupidatat deserunt elit mollit amet nulla voluptate sit. Quis aute aliquip officia deserunt sint sint nisi. Laboris sit et ea dolore consequat laboris non. Consequat do enim excepteur qui mollit consectetur eiusmod laborum ut duis mollit dolor est. Excepteur amet duis enim laborum aliqua nulla ea minim.\n", Gender: "female"},
				{ID: 8, Name: "Glenn Jordan", Age: 29, About: "Duis reprehenderit sit velit exercitation non aliqua magna quis ad excepteur anim. Eu cillum cupidatat sit magna cillum irure occaecat sunt officia officia deserunt irure. Cupidatat dolor cupidatat ipsum minim consequat Lorem adipisicing. Labore fugiat cupidatat nostrud voluptate ea eu pariatur non. Ipsum quis occaecat irure amet esse eu fugiat deserunt incididunt Lorem esse duis occaecat mollit.\n", Gender: "male"},
				{ID: 9, Name: "Rose Carney", Age: 36, About: "Voluptate ipsum ad consequat elit ipsum tempor irure consectetur amet. Et veniam sunt in sunt ipsum non elit ullamco est est eu. Exercitation ipsum do deserunt do eu adipisicing id deserunt duis nulla ullamco eu. Ad duis voluptate amet quis commodo nostrud occaecat minim occaecat commodo. Irure sint incididunt est cupidatat laborum in duis enim nulla duis ut in ut. Cupidatat ex incididunt do ullamco do laboris eiusmod quis nostrud excepteur quis ea.\n", Gender: "female"},
			},
			StatusCode: http.StatusOK,
		},
		{
			AccessToken: "valid_token",
			Request:     SearchRequest{Limit: 2, Offset: 20, Query: "on", OrderField: "Age", OrderBy: OrderByAsIs},
			Response: []User{
				{ID: 21, Name: "Johns Whitney", Age: 26, About: "Elit sunt exercitation incididunt est ea quis do ad magna. Commodo laboris nisi aliqua eu incididunt eu irure. Labore ullamco quis deserunt non cupidatat sint aute in incididunt deserunt elit velit. Duis est mollit veniam aliquip. Nulla sunt veniam anim et sint dolore.\n", Gender: "male"},
				{ID: 22, Name: "Beth Wynn", Age: 31, About: "Proident non nisi dolore id non. Aliquip ex anim cupidatat dolore amet veniam tempor non adipisicing. Aliqua adipisicing eu esse quis reprehenderit est irure cillum duis dolor ex. Laborum do aute commodo amet. Fugiat aute in excepteur ut aliqua sint fugiat do nostrud voluptate duis do deserunt. Elit esse ipsum duis ipsum.\n", Gender: "female"},
			},
			StatusCode: http.StatusOK,
		},
		{
			AccessToken: "valid_token",
			Request:     SearchRequest{Limit: 3, OrderField: "Age", OrderBy: OrderByDesc},
			Response: []User{
				{ID: 32, Name: "Christy Knapp", Age: 40, About: "Incididunt culpa dolore laborum cupidatat consequat. Aliquip cupidatat pariatur sit consectetur laboris labore anim labore. Est sint ut ipsum dolor ipsum nisi tempor in tempor aliqua. Aliquip labore cillum est consequat anim officia non reprehenderit ex duis elit. Amet aliqua eu ad velit incididunt ad ut magna. Culpa dolore qui anim consequat commodo aute.\n", Gender: "female"},
				{ID: 13, Name: "Whitley Davidson", Age: 40, About: "Consectetur dolore anim veniam aliqua deserunt officia eu. Et ullamco commodo ad officia duis ex incididunt proident consequat nostrud proident quis tempor. Sunt magna ad excepteur eu sint aliqua eiusmod deserunt proident. Do labore est dolore voluptate ullamco est dolore excepteur magna duis quis. Quis laborum deserunt ipsum velit occaecat est laborum enim aute. Officia dolore sit voluptate quis mollit veniam. Laborum nisi ullamco nisi sit nulla cillum et id nisi.\n", Gender: "male"},
				{ID: 6, Name: "Jennings Mays", Age: 39, About: "Veniam consectetur non non aliquip exercitation quis qui. Aliquip duis ut ad commodo consequat ipsum cupidatat id anim voluptate deserunt enim laboris. Sunt nostrud voluptate do est tempor esse anim pariatur. Ea do amet Lorem in mollit ipsum irure Lorem exercitation. Exercitation deserunt adipisicing nulla aute ex amet sint tempor incididunt magna. Quis et consectetur dolor nulla reprehenderit culpa laboris voluptate ut mollit. Qui ipsum nisi ullamco sit exercitation nisi magna fugiat anim consectetur officia.\n", Gender: "male"},
			},
			StatusCode: http.StatusOK,
		},
		{
			AccessToken: "valid_token",
			Request:     SearchRequest{Limit: 2, Offset: 2, OrderField: "Id", OrderBy: OrderByAsc},
			Response: []User{
				{ID: 2, Name: "Brooks Aguilar", Age: 25, About: "Velit ullamco est aliqua voluptate nisi do. Voluptate magna anim qui cillum aliqua sint veniam reprehenderit consectetur enim. Laborum dolore ut eiusmod ipsum ad anim est do tempor culpa ad do tempor. Nulla id aliqua dolore dolore adipisicing.\n", Gender: "male"},
				{ID: 3, Name: "Everett Dillard", Age: 27, About: "Sint eu id sint irure officia amet cillum. Amet consectetur enim mollit culpa laborum ipsum adipisicing est laboris. Adipisicing fugiat esse dolore aliquip quis laborum aliquip dolore. Pariatur do elit eu nostrud occaecat.\n", Gender: "male"},
			},
			StatusCode: http.StatusOK,
		},
		{
			AccessToken: "valid_token",
			Request:     SearchRequest{Limit: 2, OrderBy: OrderByAsc},
			Response: []User{
				{ID: 15, Name: "Allison Valdez", Age: 21, About: "Labore excepteur voluptate velit occaecat est nisi minim. Laborum ea et irure nostrud enim sit incididunt reprehenderit id est nostrud eu. Ullamco sint nisi voluptate cillum nostrud aliquip et minim. Enim duis esse do aute qui officia ipsum ut occaecat deserunt. Pariatur pariatur nisi do ad dolore reprehenderit et et enim esse dolor qui. Excepteur ullamco adipisicing qui adipisicing tempor minim aliquip.\n", Gender: "male"},
				{ID: 16, Name: "Annie Osborn", Age: 35, About: "Consequat fugiat veniam commodo nisi nostrud culpa pariatur. Aliquip velit adipisicing dolor et nostrud. Eu nostrud officia velit eiusmod ullamco duis eiusmod ad non do quis.\n", Gender: "female"},
			},
			StatusCode: http.StatusOK,
		},
		{
			AccessToken: "valid_token",
			Request:     SearchRequest{Limit: 40, Offset: 33},
			Response: []User{
				{ID: 33, Name: "Twila Snow", Age: 36, About: "Sint non sunt adipisicing sit laborum cillum magna nisi exercitation. Dolore officia esse dolore officia ea adipisicing amet ea nostrud elit cupidatat laboris. Proident culpa ullamco aute incididunt aute. Laboris et nulla incididunt consequat pariatur enim dolor incididunt adipisicing enim fugiat tempor ullamco. Amet est ullamco officia consectetur cupidatat non sunt laborum nisi in ex. Quis labore quis ipsum est nisi ex officia reprehenderit ad adipisicing fugiat. Labore fugiat ea dolore exercitation sint duis aliqua.\n", Gender: "female"},
				{ID: 34, Name: "Kane Sharp", Age: 34, About: "Lorem proident sint minim anim commodo cillum. Eiusmod velit culpa commodo anim consectetur consectetur sint sint labore. Mollit consequat consectetur magna nulla veniam commodo eu ut et. Ut adipisicing qui ex consectetur officia sint ut fugiat ex velit cupidatat fugiat nisi non. Dolor minim mollit aliquip veniam nostrud. Magna eu aliqua Lorem aliquip.\n", Gender: "male"},
			},
			StatusCode: http.StatusOK,
		},
		{
			AccessToken: "valid_token",
			Request:     SearchRequest{Limit: 40, Offset: 35},
			Response:    []User{},
			StatusCode:  http.StatusOK,
		},
		{
			AccessToken: "valid_token",
			Request:     SearchRequest{Limit: 40, Offset: 34},
			Response: []User{
				{ID: 34, Name: "Kane Sharp", Age: 34, About: "Lorem proident sint minim anim commodo cillum. Eiusmod velit culpa commodo anim consectetur consectetur sint sint labore. Mollit consequat consectetur magna nulla veniam commodo eu ut et. Ut adipisicing qui ex consectetur officia sint ut fugiat ex velit cupidatat fugiat nisi non. Dolor minim mollit aliquip veniam nostrud. Magna eu aliqua Lorem aliquip.\n", Gender: "male"},
			},
			StatusCode: http.StatusOK,
		},
	}

	for caseNum, item := range cases {
		addr, err := url.Parse("http://example.com/api/")
		if err != nil {
			continue
		}
		params := url.Values{}

		params.Add("limit", strconv.Itoa(item.Request.Limit))
		params.Add("offset", strconv.Itoa(item.Request.Offset))
		params.Add("query", item.Request.Query)
		params.Add("order_field", item.Request.OrderField)
		params.Add("order_by", strconv.Itoa(item.Request.OrderBy))

		addr.RawQuery = params.Encode()

		req := httptest.NewRequest(http.MethodGet, addr.String(), nil)
		req.Header.Set("AccessToken", item.AccessToken)
		w := httptest.NewRecorder()

		SearchServer(w, req)

		assert.Equal(t, item.StatusCode, w.Code, "[%d] wrong StatusCode: got %d, expected %d", caseNum, w.Code, item.StatusCode)

		resp := w.Result()
		body, err := io.ReadAll(resp.Body)

		assert.NoError(t, err, "[%d] failed to read body: %v", caseNum, err)

		if w.Code != http.StatusOK {
			continue
		}

		data := []User{}
		err = json.Unmarshal(body, &data)

		assert.NoError(t, err, "[%d] failed to Unmarshal usersData", caseNum)

		assert.Equal(t, item.Response, data, "[%d] wrong result, expected %#v, got %#v", caseNum, item.Response, data)
	}
}

func TestNotLimitCases(t *testing.T) {
	addr, err := url.Parse("http://example.com/api/")
	if err != nil {
		return
	}
	params := url.Values{}

	params.Add("offset", "33")

	addr.RawQuery = params.Encode()

	req := httptest.NewRequest(http.MethodGet, addr.String(), nil)
	req.Header.Set("AccessToken", "validToken")
	w := httptest.NewRecorder()

	SearchServer(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "wrong StatusCode: got %d, expected %d", w.Code, http.StatusOK)

	resp := w.Result()
	body, err := io.ReadAll(resp.Body)

	assert.NoError(t, err, "failed to read body: %v", err)

	data := []User{}
	err = json.Unmarshal(body, &data)

	assert.NoError(t, err, "failed to Unmarshal usersData")

	expectedResult := []User{
		{ID: 33, Name: "Twila Snow", Age: 36, About: "Sint non sunt adipisicing sit laborum cillum magna nisi exercitation. Dolore officia esse dolore officia ea adipisicing amet ea nostrud elit cupidatat laboris. Proident culpa ullamco aute incididunt aute. Laboris et nulla incididunt consequat pariatur enim dolor incididunt adipisicing enim fugiat tempor ullamco. Amet est ullamco officia consectetur cupidatat non sunt laborum nisi in ex. Quis labore quis ipsum est nisi ex officia reprehenderit ad adipisicing fugiat. Labore fugiat ea dolore exercitation sint duis aliqua.\n", Gender: "female"},
		{ID: 34, Name: "Kane Sharp", Age: 34, About: "Lorem proident sint minim anim commodo cillum. Eiusmod velit culpa commodo anim consectetur consectetur sint sint labore. Mollit consequat consectetur magna nulla veniam commodo eu ut et. Ut adipisicing qui ex consectetur officia sint ut fugiat ex velit cupidatat fugiat nisi non. Dolor minim mollit aliquip veniam nostrud. Magna eu aliqua Lorem aliquip.\n", Gender: "male"},
	}

	assert.Equal(t, expectedResult, data, "wrong result, expected %#v, got %#v", expectedResult, data)
}

func TestInvalidResponseWriterObj(t *testing.T) {
	cases := []ServerTestCase{
		{
			Request: SearchRequest{OrderField: "Invalid"},
		},
		{
			Request: SearchRequest{OrderBy: -10},
		},
		{
			Request: SearchRequest{Limit: -1000},
		},
		{
			Request: SearchRequest{Offset: -1000},
		},
	}

	for caseNum, item := range cases {
		mockWriter := &mockResponseWriter{err: errInvalidResponseWriterObj}

		addr, err := url.Parse("http://example.com/api/")
		if err != nil {
			return
		}

		params := url.Values{}
		if item.Request.OrderField == "Invalid" {
			params.Add("order_field", item.Request.OrderField)
		}
		if item.Request.OrderBy == -10 {
			params.Add("order_by", strconv.Itoa(item.Request.OrderBy))
		}
		if item.Request.Limit == -1000 {
			params.Add("limit", "Invalid")
		}
		if item.Request.Offset == -1000 {
			params.Add("offset", "Invalid")
		}

		addr.RawQuery = params.Encode()

		req := httptest.NewRequest(http.MethodGet, addr.String(), nil)

		_, err = ProccesQueryParams(mockWriter, req)

		assert.Error(t, err, "[%d] expected error, but don't", caseNum)
		assert.Equal(t, errInvalidResponseWriterObj.Error(), err.Error(), "[%d] invalid type of error")
	}
}

func TestIncorrectUserXMLData(t *testing.T) {
	w := httptest.NewRecorder()
	SendResultToClient(w, nil, false)

	body := strings.TrimSpace(w.Body.String())
	assert.Equal(t, http.StatusInternalServerError, w.Code, "expected StatusInternalServerError, but ok")
	assert.Equal(t, errInvalidUsersStruct.Error(), body, "expected %#s, but got %#s", errInvalidUsersStruct.Error(), body)

}

type InvalidUserXMLData struct {
}

func (s InvalidUserXMLData) MarshalJSON() ([]byte, error) {
	return nil, errInvalidResponseWriterObj
}

func TestInabilityToSendClient(t *testing.T) {
	w := httptest.NewRecorder()
	check := InvalidUserXMLData{}
	SendResultToClient(w, check, true)

	body := strings.TrimSpace(w.Body.String())
	assert.Equal(t, http.StatusInternalServerError, w.Code, "expected StatusInternalServerError, but ok")
	assert.Equal(t, errCantMarhalJSON.Error(), body, "expected %#s, but got %#s", errCantMarhalJSON.Error(), body)
}
