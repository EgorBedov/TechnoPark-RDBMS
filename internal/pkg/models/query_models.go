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

type PostQuery struct {
	SlugOrId	string
	Limit 	int
	Since 	int
	Sort	string
	Desc 	bool
}

func DecodePostQuery(r *http.Request) PostQuery {
	params := r.URL.Query()
	query := PostQuery{
		SlugOrId: 	mux.Vars(r)["slug_or_id"],
		Limit: 	100,
		Since: 	0,
		Sort:	"flat",
		Desc:  	false,
	}
	if limit := params.Get("limit"); limit != "" {
		query.Limit, _ = strconv.Atoi(limit)
	}
	if since := params.Get("since"); since != "" {
		query.Since, _ = strconv.Atoi(since)
	}
	if sort := params.Get("sort"); sort != "" {
		query.Sort = sort
	}
	if desc := params.Get("desc"); desc != "" {
		query.Desc, _ = strconv.ParseBool(desc)
	}

	return query
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
