package main

import (
	"net/http"
)

func (app *application) internalServerError(rw http.ResponseWriter, r *http.Request, err error) {
	app.logger.Errorw("internal error", "method", r.Method, "path", r.URL.Path, "error", err.Error())
	writeJSONError(rw, http.StatusInternalServerError, "the server encountered a problem")
}

func (app *application) badRequestResponse(rw http.ResponseWriter, r *http.Request, err error) {
	app.logger.Warnf("bad request error", "method", r.Method, "path", r.URL.Path, "error", err.Error())
	writeJSONError(rw, http.StatusBadRequest, err.Error())
}

func (app *application) conflictResponse(rw http.ResponseWriter, r *http.Request, err error) {
	app.logger.Errorf("conflict response", "method", r.Method, "path", r.URL.Path, "error", err.Error())
	writeJSONError(rw, http.StatusConflict, err.Error())
}

func (app *application) notFoundResponse(rw http.ResponseWriter, r *http.Request, err error) {
	app.logger.Warnf("not found error", "method", r.Method, "path", r.URL.Path, "error", err.Error())
	writeJSONError(rw, http.StatusNotFound, "not found")
}

func (app *application) unauthorizedBasicErrorResponse(rw http.ResponseWriter, r *http.Request, err error) {
	app.logger.Warnf("unauthorized error", "method", r.Method, "path", r.URL.Path, "error", err.Error())
	rw.Header().Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)
	writeJSONError(rw, http.StatusUnauthorized, "unauthorized")
}

func (app *application) unauthorizedErrorResponse(rw http.ResponseWriter, r *http.Request, err error) {
	app.logger.Warnf("unauthorized error", "method", r.Method, "path", r.URL.Path, "error", err.Error())
	writeJSONError(rw, http.StatusUnauthorized, "unauthorized")
}

func (app *application) forbiddendResponse(rw http.ResponseWriter, r *http.Request) {
	app.logger.Warnw("forbidden", "method", r.Method, "path", r.URL, "error")
	writeJSONError(rw, http.StatusForbidden, "forbidden")
}

func (app *application) rateLimitExceededResponse(rw http.ResponseWriter, r *http.Request, retryAfter string) {
	app.logger.Warnw("rate limited exceeded", "method", r.Method, "path", r.URL.Path)
	rw.Header().Set("Retry-After", retryAfter)
	writeJSONError(rw, http.StatusTooManyRequests, "rate limit exceeded, retry after: "+retryAfter)
}
