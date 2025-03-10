package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"
)

const filePath string = "dataset.xml"
const trueToken string = "oralCumshot"

type Users struct {
	List []XmlUser `xml:"row"`
}

type XmlUser struct {
	Id        int    `xml:"id"`
	FirstName string `xml:"first_name"`
	LastName  string `xml:"last_name"`
	Age       int    `xml:"age"`
	About     string `xml:"about"`
	Gender    string `xml:"gender"`
}

func Search(r *http.Request) []byte {

	if r.Header.Get("AccessToken") != trueToken {
		resp, err := json.Marshal(SearchErrorResponse{Error: "bad token :C"})
		if err != nil {
			panic(err)
		}
		return resp
	}

	limit, err := strconv.Atoi(r.URL.Query().Get("limit"))
	if err != nil {
		panic(err)
	}
	offset, err := strconv.Atoi(r.URL.Query().Get("offset"))
	if err != nil {
		panic(err)
	}
	orderBy, err := strconv.Atoi(r.URL.Query().Get("order_by"))
	if err != nil {
		panic(err)
	}

	req := SearchRequest{
		Limit:      limit,
		Offset:     offset,
		OrderBy:    orderBy,
		Query:      r.URL.Query().Get("query"),
		OrderField: r.URL.Query().Get("order_field"),
	}

	file, err := os.Open(filePath)
	if err != nil {
		panic(err)
	}

	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		panic(err)
	}

	userData := Users{}

	err = xml.Unmarshal(content, &userData)
	if err != nil {
		panic(err)
	}

	users := make([]User, len(userData.List))

	for i, user := range userData.List {
		users[i] = User{
			Id:     user.Id,
			Name:   user.FirstName + " " + user.LastName,
			Age:    user.Age,
			About:  user.About,
			Gender: user.Gender,
		}

	}

	entries := make([]User, 0, req.Limit)

	for _, user := range users {
		if strings.Contains(user.Name, req.Query) || strings.Contains(user.About, req.Query) {
			entries = append(entries, user)
		}
	}

	if req.OrderField == "" {
		req.OrderField = "Name"
	}

	type SortValue struct {
		IntVal int
		StrVal string
		IsStr  bool
	}

	var getField func(u User) SortValue

	switch req.OrderField {
	case "Name":
		getField = func(u User) SortValue { return SortValue{StrVal: u.Name, IsStr: true} }
	case "Id":
		getField = func(u User) SortValue { return SortValue{IntVal: u.Id} }
	case "Age":
		getField = func(u User) SortValue { return SortValue{IntVal: u.Age} }
	default:
		resp, err := json.Marshal(SearchErrorResponse{Error: "ErrorBadOrderField"})
		if err != nil {
			panic(err)
		}
		return resp
	}

	if req.OrderBy != 0 {

		if req.OrderBy > 1 || req.OrderBy < -1 {
			resp, err := json.Marshal(SearchErrorResponse{Error: "Wrong order_by number:" + strconv.Itoa(req.OrderBy)})
			if err != nil {
				panic(err)
			}
			return resp
		}

		sort.Slice(entries, func(i int, j int) bool {
			a, b := getField(entries[i]), getField(entries[j])

			if a.IsStr {
				if req.OrderBy == 1 {
					return a.StrVal > b.StrVal
				}
				return a.StrVal < b.StrVal
			}

			if req.OrderBy == 1 {
				return a.IntVal > b.IntVal
			}
			return a.IntVal < b.IntVal

		})
	}

	if req.Offset > len(entries) {
		resp, err := json.Marshal(SearchErrorResponse{Error: "Offset is more than number of entries, offset:" + strconv.Itoa(req.Offset) +
			"  , number of entries:" + strconv.Itoa(len(entries))})
		if err != nil {
			panic(err)
		}
		return resp
	}

	end := req.Limit + req.Offset

	if end > len(entries) {
		end = len(entries)
	}

	resp, err := json.Marshal(entries[req.Offset:end])
	if err != nil {
		panic(err)
	}

	return resp

}

func Dummy(w http.ResponseWriter, r *http.Request) {
	resp := Search(r)
	errData := SearchErrorResponse{}
	err := json.Unmarshal(resp, &errData)
	if err != nil {
		userData := []User{}
		err := json.Unmarshal(resp, &userData)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
		w.WriteHeader(http.StatusOK)
		w.Write(resp)
		return
	}
	switch errData.Error {
	case "bad token :C":
		w.WriteHeader(http.StatusUnauthorized)
	default:
		w.WriteHeader(http.StatusBadRequest)
		w.Write(resp)
	}

}

func FatalDummy(w http.ResponseWriter, r *http.Request) {
	key := r.FormValue("query")

	switch key {
	case "__internal_error":
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"Error": "Unexpected response format",`))
	case "__bad_request_json":
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"Error": "Unexpected response format",`))
	case "__good_request_bad_json":
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"sdsd": "Unexpted respd format"}`))
	}

}

func TestTimeout(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(3 * time.Second)
		resp := Search(r)
		w.WriteHeader(http.StatusOK)
		w.Write(resp)
	}))
	defer ts.Close()
	searchCases := []SearchClient{
		{
			URL:         ts.URL,
			AccessToken: trueToken,
		},
		{
			URL:         "http://localhost:9999",
			AccessToken: trueToken,
		},
	}

	req := SearchRequest{}

	for _, search := range searchCases {
		_, err := search.FindUsers(req)
		if err != nil {
			if strings.Contains(err.Error(), "timeout") {
				t.Logf("Expected timeout")
				continue

			}
			if strings.Contains(search.URL, "http://localhost:9999") {
				t.Logf("Expected unknown error, %v", err)
				return
			}
			t.Fatalf("Unexpected error, %v", err)
		}
	}

	t.Fatalf("Unexpected responce from server")

}

func TestSearch(t *testing.T) {
	cases := []struct {
		name    string
		req     SearchRequest
		isError bool
		isEmpty bool
	}{{"Normal query", SearchRequest{Limit: 10, Offset: 0, Query: "", OrderField: "Id", OrderBy: 0}, false, false},
		{"Check max limit = 25", SearchRequest{Limit: 30, Offset: 0, Query: "J", OrderField: "Name", OrderBy: 1}, false, false},
		{"Check negative limit", SearchRequest{Limit: -10, Offset: 0, Query: "", OrderField: "Age", OrderBy: 0}, true, false},
		{"Check wrong order_fieid", SearchRequest{Limit: 0, Offset: 0, Query: "", OrderField: "sdos", OrderBy: -1}, true, false},
		{"Check negative offset", SearchRequest{Limit: 0, Offset: -10, Query: "", OrderField: "Id", OrderBy: -1}, true, false},
		{"Check offset > len(entries)", SearchRequest{Limit: 0, Offset: 30, Query: "J", OrderField: "Id", OrderBy: -1}, true, false},
		{"Check invalid order_by", SearchRequest{Limit: 0, Offset: 30, Query: "J", OrderField: "Id", OrderBy: 999}, true, false},
		{"Check limit = 0", SearchRequest{Limit: 0, Offset: 0, Query: "J", OrderField: "Id", OrderBy: -1}, false, true},
	}

	ts := httptest.NewServer(http.HandlerFunc(Dummy))
	defer ts.Close()

	search := SearchClient{
		URL:         ts.URL,
		AccessToken: trueToken,
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			resp, err := search.FindUsers(test.req)
			if test.isEmpty {
				if err != nil {
					t.Fatalf("Unexpected error while parsing json: %v", err)
				}
				if len(resp.Users) != 0 {
					t.Fatalf("Expected empty response, but got: %v", resp.Users)
				}
			}

			if test.isError {
				if err == nil {
					t.Fatalf("Expected error, but got none")
				}
				t.Logf("Caught expected error: %v", err)
			}

		})

	}

	search = SearchClient{
		URL:         ts.URL,
		AccessToken: "aswsd",
	}

	_, err := search.FindUsers(SearchRequest{})
	if err != nil {
		if strings.Contains(err.Error(), "Bad AccessToken") {
			t.Logf("Expected Bad Token")
			return
		}
		t.Fatalf("Unexpected token check error")

	}

}

func TestFatalErrors(t *testing.T) {
	fatalErrors := map[string]string{
		"__internal_error":        "SearchServer fatal error",
		"__bad_request_json":      "cant unpack error json",
		"__good_request_bad_json": "cant unpack result json",
	}

	ts := httptest.NewServer(http.HandlerFunc(FatalDummy))
	defer ts.Close()

	search := &SearchClient{
		URL:         ts.URL,
		AccessToken: trueToken,
	}

	for query, expectedErr := range fatalErrors {
		t.Run(fmt.Sprintf("Query=%s", query), func(t *testing.T) {
			_, err := search.FindUsers(SearchRequest{Query: query})
			if err == nil {
				t.Fatalf("Expected error, but got nil")
			}
			if !strings.Contains(err.Error(), expectedErr) {
				t.Fatalf("Expected %q, but got %v", expectedErr, err)
			}
			t.Logf("Correctly got expected error: %v", err)
			t.Logf("Test passed for query=%s", query)
		})
	}
}

// код писать тут
