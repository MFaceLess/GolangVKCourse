package main

import (
	"cmp"
	"encoding/json"
	"encoding/xml"
	"errors"
	"net/http"
	"os"
	"slices"
	"strconv"
	"strings"
)

const (
	LimitNotSet = -1

	ID     = "Id"
	Name   = "Name"
	Age    = "Age"
	About  = "About"
	Gender = "Gender"
)

var (
	DataFileName = "dataset.xml"

	errAccessDenied = errors.New("access Denied")

	errInvalidOrder       = errors.New("некорректно задано поле order")
	errInvalidLimit       = errors.New("некорректно задано поле limit")
	errInvalidOffset      = errors.New("некорректно задано поле offset")
	errQueryError         = errors.New("ошибка при парсинге параметров query")
	errInvalidUsersStruct = errors.New("ошибка при преобразовании к типу []UserXMLData")
)

type UserXMLData struct {
	ID        int    `xml:"id"`
	FirstName string `xml:"first_name"`
	LastName  string `xml:"last_name"`
	Age       int    `xml:"age"`
	About     string `xml:"about"`
	Gender    string `xml:"gender"`
}

func (user *UserXMLData) MarshalJSON() ([]byte, error) {
	serializeMap := map[string]interface{}{
		ID:     user.ID,
		Name:   user.FirstName + " " + user.LastName,
		Age:    user.Age,
		About:  user.About,
		Gender: user.Gender,
	}

	return json.Marshal(serializeMap)
}

type Users struct {
	UserList []UserXMLData `xml:"row"`
}

func GetFileData(fileName string) ([]byte, error) {
	fContent, err := os.ReadFile(fileName)

	if err != nil {
		return nil, err
	}

	return fContent, nil
}

func OrderUsers(users *[]UserXMLData, orderField string, orderBy int) {
	if orderBy == OrderByAsIs {
		return
	}

	slices.SortFunc(*users, func(a, b UserXMLData) int {
		var result int
		switch orderField {
		case ID:
			result = cmp.Compare(a.ID, b.ID)
		case Age:
			result = cmp.Compare(a.Age, b.Age)
		default:
			result = strings.Compare(a.FirstName+" "+a.LastName, b.FirstName+" "+b.LastName)
		}

		if orderBy == OrderByDesc {
			return -result
		}
		return result
	})
}

func ApplyQueryToUsers(request *SearchRequest, users *Users) []UserXMLData {
	queryUsers := []UserXMLData{}

	if request.Query != "" {
		for _, user := range users.UserList {
			if strings.Contains(user.FirstName+" "+user.LastName, request.Query) || strings.Contains(user.About, request.Query) {
				queryUsers = append(queryUsers, user)
			}
		}
	} else {
		queryUsers = users.UserList
	}

	if request.Limit == LimitNotSet {
		request.Limit = len(queryUsers)
	}

	OrderUsers(&queryUsers, request.OrderField, request.OrderBy)

	if request.Offset < len(queryUsers) && request.Offset+request.Limit <= len(queryUsers) {
		return queryUsers[request.Offset:(request.Offset + request.Limit)]
	}

	if request.Offset < len(queryUsers) {
		return queryUsers[request.Offset:]
	}

	return []UserXMLData{}
}

func SendResultToClient(w http.ResponseWriter, usersData any, test bool) {
	_, ok := usersData.([]UserXMLData)
	if !ok && !test {
		http.Error(w, errInvalidUsersStruct.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	err := json.NewEncoder(w).Encode(usersData)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func ProccesQueryParams(w http.ResponseWriter, r *http.Request) (*SearchRequest, error) {
	orderField := r.URL.Query().Get("order_field")
	if orderField == "" {
		orderField = Name
	}
	if orderField != Name && orderField != ID && orderField != Age {
		checkErr := SearchErrorResponse{Error: ErrorBadOrderField}
		w.WriteHeader(http.StatusBadRequest)
		err := json.NewEncoder(w).Encode(checkErr)
		if err != nil {
			return nil, err
		}
		return nil, errQueryError
	}

	var err error

	orderString := r.URL.Query().Get("order_by")
	var order int
	if orderString != "" {
		order, err = strconv.Atoi(orderString)
		if err != nil || (order != OrderByAsc && order != OrderByAsIs && order != OrderByDesc) {
			checkErr := SearchErrorResponse{Error: errInvalidOrder.Error()}
			w.WriteHeader(http.StatusBadRequest)
			err = json.NewEncoder(w).Encode(checkErr)
			if err != nil {
				return nil, err
			}
			return nil, errQueryError
		}
	}

	limitString := r.URL.Query().Get("limit")

	limit := LimitNotSet
	if limitString != "" {
		limit, err = strconv.Atoi(limitString)
		if err != nil {
			checkErr := SearchErrorResponse{Error: errInvalidLimit.Error()}
			w.WriteHeader(http.StatusBadRequest)
			err = json.NewEncoder(w).Encode(checkErr)
			if err != nil {
				return nil, err
			}
			return nil, errQueryError
		}
	}

	offsetString := r.URL.Query().Get("offset")
	var offset int
	if offsetString != "" {
		offset, err = strconv.Atoi(offsetString)
		if err != nil {
			checkErr := SearchErrorResponse{Error: errInvalidOffset.Error()}
			w.WriteHeader(http.StatusBadRequest)
			err = json.NewEncoder(w).Encode(checkErr)
			if err != nil {
				return nil, err
			}
			return nil, errQueryError
		}
	}

	query := r.URL.Query().Get("query")

	result := SearchRequest{
		Limit:      limit,
		Offset:     offset,
		Query:      query,
		OrderField: orderField,
		OrderBy:    order,
	}

	return &result, nil
}

func SearchServer(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("AccessToken")
	if token == "" {
		http.Error(w, errAccessDenied.Error(), http.StatusUnauthorized)
		return
	}

	searchRequest, err := ProccesQueryParams(w, r)
	if err != nil {
		return
	}

	xmlData, err := GetFileData(DataFileName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	users := new(Users)
	err = xml.Unmarshal(xmlData, &users)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	usersResult := ApplyQueryToUsers(searchRequest, users)

	SendResultToClient(w, usersResult, false)
}
