package utils

import (
	"encoding/json"
	"net/http"
)

type openaiErrorResponse struct {
	Error struct {
		Message string      `json:"message"`
		Type    string      `json:"type"`
		Param   interface{} `json:"param"`
		Code    interface{} `json:"code"`
	} `json:"error"`
}

func JSONError(w http.ResponseWriter, message string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	
	resp := openaiErrorResponse{}
	resp.Error.Message = message
	resp.Error.Type = "invalid_request_error"
	
	if code >= 500 {
		resp.Error.Type = "server_error"
	}
	
	json.NewEncoder(w).Encode(resp)
}
