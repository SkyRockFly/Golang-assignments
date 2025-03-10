package main

/*

import (
	"fmt"
	"net/http"
	"strconv"
)

const filePath string = "dataset.xml"

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

func main() {

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/favicon.ico" {
			http.NotFound(w, r)
			return
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

		resp, err := Search(req)

		if err != nil {
			panic(err)
		}

		fmt.Println(resp)
	})

	fmt.Println("starting server at :8080")
	http.ListenAndServe(":8080", nil)

}

func Search(req SearchRequest) (*SearchResponse, error) {
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

	foundByQuery := make([]User, 0, req.Limit)

	for _, user := range users {
		if strings.Contains(user.Name, req.Query) || strings.Contains(user.About, req.Query) {
			foundByQuery = append(foundByQuery, user)
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
		return &SearchResponse{}, fmt.Errorf("wrong order field: %v  ", req.OrderField)
	}

	if req.OrderBy != 0 {

		sort.Slice(foundByQuery, func(i int, j int) bool {
			a, b := getField(foundByQuery[i]), getField(foundByQuery[j])

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

	if req.Offset > len(foundByQuery) {
		return &SearchResponse{}, nil
	}

	end := req.Limit + req.Offset
	nextPage := true

	if end > len(foundByQuery) {
		end = len(foundByQuery)
		nextPage = false
	}

	return &SearchResponse{Users: foundByQuery[req.Offset:end],
		NextPage: nextPage}, nil

} */
