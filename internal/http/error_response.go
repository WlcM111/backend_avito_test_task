package httpapi

import (
	"encoding/json"
	"net/http"

	"pr-reviewer-service/internal/domain"
)

// ErrorBody — обёртка для объекта ошибки в HTTP-ответе.
type ErrorBody struct {
	Error ErrorItem `json:"error"`
}

// ErrorItem описывает код и сообщение об ошибке.
type ErrorItem struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// WriteError мапит доменные ошибки в HTTP-статусы и JSON-ответ.
func WriteError(w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	code := domain.ErrorCodeInternal
	msg := "internal error"

	if derr, ok := err.(*domain.DomainError); ok {
		code = derr.Code

		if derr.Err != nil {
			msg = derr.Err.Error()
		}

		switch derr.Code {
		case domain.ErrorCodeTeamExists:
			status = http.StatusBadRequest

		case domain.ErrorCodePRExists,
			domain.ErrorCodePRMerged,
			domain.ErrorCodeNotAssigned,
			domain.ErrorCodeNoCandidate:
			status = http.StatusConflict

		case domain.ErrorCodeNotFound:
			status = http.StatusNotFound

		default:
			status = http.StatusInternalServerError
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	_ = json.NewEncoder(w).Encode(ErrorBody{
		Error: ErrorItem{
			Code:    code,
			Message: msg,
		},
	})
}
