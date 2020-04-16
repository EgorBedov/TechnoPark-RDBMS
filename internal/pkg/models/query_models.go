package models

import (
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
)

type Query struct {
	Slug	string
	Limit 	int
	Since 	string
	Desc	bool
}

func DecodeQuery(r *http.Request) Query {
	params := r.URL.Query()
	query := Query{
		Slug: mux.Vars(r)["slug"],
		Limit: 100,
		Since: "",
		Desc:  false,
	}
	if limit := params.Get("limit"); limit != "" {
		query.Limit, _ = strconv.Atoi(limit)
	}
	if since := params.Get("since"); since != "" {
		query.Since = since
	}
	if desc := params.Get("desc"); desc != "" {
		query.Desc, _ = strconv.ParseBool(desc)
	}

	return query
}
