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
	query := Query{
		Slug: mux.Vars(r)["slug"],
		Limit: 100,
		Since: "",
		Desc:  false,
	}
	if limit := r.URL.Query().Get("limit"); limit != "" {
		query.Limit, _ = strconv.Atoi(limit)
	}
	if since := r.URL.Query().Get("since"); since != "" {
		query.Since = since
	}
	if desc := r.URL.Query().Get("desc"); desc != "" {
		query.Desc, _ = strconv.ParseBool(desc)
	}

	return query
}
