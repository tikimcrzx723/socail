package main

import (
	"net/http"

	"github.com/tikimcrzx723/social/internal/store"
)

// getUserFeedHandler godoc
//
//	@Summary		Fetches the user feed
//	@Description	Fetches the user feed
//	@Tags			feed
//	@Accept			json
//	@Produce		json
//	@Param			since	query		string	false	"Since"
//	@Param			until	query		string	false	"Until"
//	@Param			limit	query		int		false	"Limit"
//	@Param			offset	query		int		false	"Offset"
//	@Param			sort	query		string	false	"Sort"
//	@Param			tags	query		string	false	"Tags"
//	@Param			search	query		string	false	"Search"
//	@Success		200		{object}	[]store.PostWithMetadata
//	@Failure		400		{object}	error
//	@Failure		500		{object}	error
//	@Security		ApiKeyAuth
//	@Router			/users/feed [get]
func (app *application) getUserFeedHandler(rw http.ResponseWriter, r *http.Request) {
	fq := store.PaginatedFeedQuery{
		Limit:  20,
		Offset: 0,
		Sort:   "desc",
	}

	fq, err := fq.Parse(r)
	if err != nil {
		app.badRequestResponse(rw, r, err)
		return
	}

	if err := Validate.Struct(fq); err != nil {
		app.badRequestResponse(rw, r, err)
		return
	}

	feed, err := app.store.Posts.GetUserFeed(r.Context(), int64(71), fq)
	if err != nil {
		app.internalServerError(rw, r, err)
		return
	}

	if err := app.jsonResponse(rw, http.StatusOK, feed); err != nil {
		app.internalServerError(rw, r, err)
	}
}
