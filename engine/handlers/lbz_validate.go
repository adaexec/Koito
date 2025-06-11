package handlers

import (
	"net/http"

	"github.com/gabehf/koito/engine/middleware"
	"github.com/gabehf/koito/internal/db"
	"github.com/gabehf/koito/internal/logger"
	"github.com/gabehf/koito/internal/utils"
)

type LbzValidateResponse struct {
	Code     int    `json:"code"`
	Error    string `json:"error,omitempty"`
	Message  string `json:"message,omitempty"`
	Valid    bool   `json:"valid,omitempty"`
	UserName string `json:"user_name,omitempty"`
}

func LbzValidateTokenHandler(store db.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		l := logger.FromContext(ctx)
		l.Debug().Msg("Validating user token...")

		u := middleware.GetUserFromContext(ctx)
		var response LbzValidateResponse
		if u == nil {
			response.Code = http.StatusUnauthorized
			response.Error = "Incorrect Authorization"
			w.WriteHeader(http.StatusUnauthorized)
			utils.WriteJSON(w, http.StatusOK, response)
		} else {
			response.Code = 200
			response.Message = "Token valid."
			response.Valid = true
			response.UserName = u.Username
			utils.WriteJSON(w, http.StatusOK, response)
		}
	}
}
