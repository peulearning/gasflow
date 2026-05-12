package httputil

import (
	"encoding/json"
	"net/http"
)

func JSON(w http.ResponseWriter, code int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(data)
}

func OK(w http.ResponseWriter, data any)      { JSON(w, http.StatusOK, data) }
func Created(w http.ResponseWriter, data any) { JSON(w, http.StatusCreated, data) }

func Error(w http.ResponseWriter, code int, msg string) {
	JSON(w, code, map[string]string{"error": msg})
}

func BadRequest(w http.ResponseWriter, msg string)    { Error(w, http.StatusBadRequest, msg) }
func Unauthorized(w http.ResponseWriter, msg string)  { Error(w, http.StatusUnauthorized, msg) }
func Forbidden(w http.ResponseWriter, msg string)     { Error(w, http.StatusForbidden, msg) }
func NotFound(w http.ResponseWriter, msg string)      { Error(w, http.StatusNotFound, msg) }
func Unprocessable(w http.ResponseWriter, msg string) { Error(w, http.StatusUnprocessableEntity, msg) }