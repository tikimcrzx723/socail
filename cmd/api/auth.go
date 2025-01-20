package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/tikimcrzx723/social/internal/mailer"
	"github.com/tikimcrzx723/social/internal/store"
)

type RegisterUserPayload struct {
	Username string `json:"username" validate:"required,max=100"`
	Email    string `json:"email" validate:"required,email,max=255"`
	Password string `json:"password" validate:"required,min=3,max=72"`
}

type UserWithToken struct {
	*store.User
	Token string `json:"token"`
}

// registerUserHandler godoc
//
//	@Summary		Registers a user
//	@Description	Registers a user
//	@Tags			authentication
//	@Accept			json
//	@Produce		json
//	@Param			payload	body		RegisterUserPayload	true	"User credentials"
//	@Success		201		{object}	UserWithToken		"User registered"
//	@Failure		400		{object}	error
//	@Failure		500		{object}	error
//	@Router			/authentication/user [post]
func (app *application) registerUserHandler(rw http.ResponseWriter, r *http.Request) {
	var payload RegisterUserPayload
	if err := readJSON(rw, r, &payload); err != nil {
		app.badRequestResponse(rw, r, err)
		return
	}

	if err := Validate.Struct(payload); err != nil {
		app.badRequestResponse(rw, r, err)
		return
	}

	user := &store.User{
		Username: payload.Username,
		Email:    payload.Email,
		Role: store.Role{
			Name: "user",
		},
	}

	// hash, the user password
	if err := user.Password.Set(payload.Password); err != nil {
		app.internalServerError(rw, r, err)
		return
	}

	plainToken := uuid.New().String()

	hash := sha256.Sum256([]byte(plainToken))
	hashToken := hex.EncodeToString(hash[:])

	// store user
	err := app.store.Users.CreateAndInvate(r.Context(), user, hashToken, time.Hour*72)
	if err != nil {
		switch err {
		case store.ErrDuplicateEmail:
			app.badRequestResponse(rw, r, err)
		case store.ErrDuplicateUsername:
			app.badRequestResponse(rw, r, err)
		default:
			app.internalServerError(rw, r, err)
		}
		return
	}

	userWithToken := UserWithToken{
		User:  user,
		Token: plainToken,
	}
	activationURL := fmt.Sprintf("%s/confirm/%s", app.config.frontedURL, plainToken)

	// isProdEnv := app.config.env == "production"
	// vars := struct {
	// 	Username      string
	// 	ActivationURL string
	// }{
	// 	Username:      user.Username,
	// 	ActivationURL: activationURL,
	// }

	// send email
	// err = app.mailer.Send(mailer.UserWelcomeTemplate, user.Username, user.Email, vars, !isProdEnv)
	data := map[string]any{
		"username":      user.Email,
		"activationURL": activationURL,
	}
	err = app.mailer.Send(user.Email, mailer.UserWelcomeTemplate, data)
	if err != nil {
		app.logger.Errorf("error sending welcome email", "error", err.Error())
		// rollback user creation if email fails (SAGA pattern)
		if err := app.store.Users.Delete(r.Context(), user.ID); err != nil {
			app.logger.Errorw("error deleting user", "error", err)
		}
		app.internalServerError(rw, r, err)
		return
	}

	if err := app.jsonResponse(rw, http.StatusCreated, userWithToken); err != nil {
		app.internalServerError(rw, r, err)
	}
}

type CreateUserTokenPayload struct {
	Email    string `json:"email" validate:"required,email,max=255"`
	Password string `json:"password" validate:"required,min=3,max=72"`
}

// createTokenHandler godoc
//
//	@Summary		Creates a token
//	@Description	Creates a token for a user
//	@Tags			authentication
//	@Accept			json
//	@Produce		json
//	@Param			payload	body		CreateUserTokenPayload	true	"User credentials"
//	@Success		200		{string}	string					"Token"
//	@Failure		400		{object}	error
//	@Failure		401		{object}	error
//	@Failure		500		{object}	error
//	@Router			/authentication/token [post]
func (app *application) createTokenHandler(rw http.ResponseWriter, r *http.Request) {
	// parse payload credentials
	var payload CreateUserTokenPayload
	if err := readJSON(rw, r, &payload); err != nil {
		app.badRequestResponse(rw, r, err)
		return
	}

	if err := Validate.Struct(payload); err != nil {
		app.badRequestResponse(rw, r, err)
		return
	}
	// fetch the user (check if the user exists) from the payload
	user, err := app.store.Users.GetByEmail(r.Context(), payload.Email)
	if err != nil {
		switch err {
		case store.ErrNotFound:
			app.unauthorizedErrorResponse(rw, r, err)
		default:
			app.internalServerError(rw, r, err)
		}
		return
	}
	// generate the token -> add claims

	if err := user.Password.Compare(payload.Password); err != nil {
		fmt.Print(err)
		app.unauthorizedErrorResponse(rw, r, err)
		return
	}

	claims := jwt.MapClaims{
		"sub": user.ID,
		"exp": time.Now().Add(app.config.auth.token.exp).Unix(),
		"iat": time.Now().Unix(),
		"nbf": time.Now().Unix(),
		"iss": app.config.auth.token.iss,
		"aud": app.config.auth.token.iss,
	}
	token, err := app.authenticator.GenerateToken(claims)
	if err != nil {
		app.internalServerError(rw, r, err)
		return
	}
	// send it to the client
	if err := app.jsonResponse(rw, http.StatusCreated, token); err != nil {
		app.internalServerError(rw, r, err)
	}
}
