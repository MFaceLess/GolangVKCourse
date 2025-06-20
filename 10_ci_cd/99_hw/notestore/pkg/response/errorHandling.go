package response

import (
	"encoding/json"
	"io"
)

type Response struct {
	StatusCode int    `json:"status_code"`
	Msg        string `json:"msg"`
}

func RespJSON(w io.Writer, statusCode int, message string) {
	respJSON := Response{
		StatusCode: statusCode,
		Msg:        message,
	}

	resp, err := json.Marshal(respJSON)
	if err != nil {
		return
	}

	if _, err = w.Write(resp); err != nil {
		return
	}
}
