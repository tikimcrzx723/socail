package main

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/tikimcrzx723/social/internal/store"
)

type userKey string

const userCtx userKey = "user"

// GetUser godoc
//
//	@Summary		Fetches a user profile
//	@Description	Fetches a user profile by ID
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			id	path		int	true	"User ID"
//	@Success		200	{object}	store.User
//	@Failure		400	{object}	error
//	@Failure		404	{object}	error
//	@Failure		500	{object}	error
//	@Security		ApiKeyAuth
//	@Router			/users/{id} [get]
func (app *application) getUserHandler(rw http.ResponseWriter, r *http.Request) {
	userID, err := strconv.ParseInt(chi.URLParam(r, "userID"), 10, 64)
	if err != nil {
		app.badRequestResponse(rw, r, err)
		return
	}

	user, err := app.getUser(r.Context(), userID)
	if err != nil {
		switch err {
		case store.ErrNotFound:
			app.notFoundResponse(rw, r, err)
		default:
			app.internalServerError(rw, r, err)
		}
		return
	}
	if err := app.jsonResponse(rw, http.StatusOK, user); err != nil {
		app.internalServerError(rw, r, err)
	}
}

// FollowUser godoc
//
//	@Summary		Follows a user
//	@Description	Follows a user by ID
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			userID	path		int		true	"User ID"
//	@Success		204		{string}	string	"User followed"
//	@Failure		400		{object}	error	"User payload missing"
//	@Failure		404		{object}	error	"User not found"
//	@Security		ApiKeyAuth
//	@Router			/users/{userID}/follow [put]
func (app *application) followUserHandler(rw http.ResponseWriter, r *http.Request) {
	followerUser := getUserFromContext(r)
	followedID, err := strconv.ParseInt(chi.URLParam(r, "userID"), 10, 64)
	if err != nil {
		app.badRequestResponse(rw, r, err)
		return
	}

	if err := app.store.Followers.Follow(r.Context(), followerUser.ID, followedID); err != nil {
		switch err {
		case store.ErrConflict:
			app.conflictResponse(rw, r, err)
		default:
			app.internalServerError(rw, r, err)
		}
		return
	}

	if err := app.jsonResponse(rw, http.StatusNoContent, nil); err != nil {
		app.internalServerError(rw, r, err)
	}
}

// UnfollowUser godoc
//
//	@Summary		Unfollows a user
//	@Description	Unfollows a user by ID
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Param			userID	path		int			true	"User ID"
//	@Success		204		{string}	{string}	"User followed"
//	@Failure		400		{object}	error		"User payload missing"
//	@Failure		404		{object}	error		"User not found"
//	@Router			/users/{userID}/unfollow [put]
func (app *application) unfollowUserHandler(rw http.ResponseWriter, r *http.Request) {
	followerUser := getUserFromContext(r)
	unfollowedID, err := strconv.ParseInt(chi.URLParam(r, "userID"), 10, 64)
	if err != nil {
		app.badRequestResponse(rw, r, err)
		return
	}

	if err := app.store.Followers.Unfollow(r.Context(), followerUser.ID, unfollowedID); err != nil {
		app.internalServerError(rw, r, err)
		return
	}
}

// ActivateUser godoc
//
//	@Summary		Activates/Register a user
//	@Description	Activates/Register a user by invitation token
//
//	@Tags			users
//	@Produce		json
//	@Param			token	path		string	true	"Invitation token"
//	@Success		204		{string}	string	"User activated"
//	@Failure		404		{object}	error
//	@Failure		500		{object}	error
//	@Security		ApiKeyAuth
//	@Router			/users/activate/{token} [put]
func (app *application) activateUserHandler(rw http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")

	err := app.store.Users.Activate(r.Context(), token)
	if err != nil {
		switch err {
		case store.ErrNotFound:
			app.notFoundResponse(rw, r, err)
		default:
			app.internalServerError(rw, r, err)
		}
		return
	}

	if err := app.jsonResponse(rw, http.StatusNoContent, ""); err != nil {
		app.internalServerError(rw, r, err)
	}
}

// func (app *application) userContextMiddleware(next http.Handler) http.Handler {
// 	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
// 		userID, err := strconv.ParseInt(chi.URLParam(r, "userID"), 10, 64)
// 		if err != nil {
// 			app.badRequestResponse(rw, r, err)
// 			return
// 		}

// 		ctx := r.Context()

// 		user, err := app.store.Users.GetByID(ctx, userID)
// 		if err != nil {
// 			switch {
// 			case errors.Is(err, store.ErrNotFound):
// 				app.badRequestResponse(rw, r, err)
// 			default:
// 				app.internalServerError(rw, r, err)
// 			}
// 			return
// 		}

// 		ctx = context.WithValue(ctx, userCtx, user)
// 		next.ServeHTTP(rw, r.WithContext(ctx))
// 	})
// }

func getUserFromContext(r *http.Request) *store.User {
	user, _ := r.Context().Value(userCtx).(*store.User)
	return user
}
