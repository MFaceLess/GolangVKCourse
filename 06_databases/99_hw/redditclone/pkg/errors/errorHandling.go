package errors

import (
	"encoding/json"
	"io"
)

type ErrorResponse struct {
	Location string `json:"location"`
	Param    string `json:"param"`
	Value    string `json:"value"`
	Msg      string `json:"msg"`
}

func ErrorJSON(w io.Writer, location, param, value, msg string) {
	errors := []ErrorResponse{{
		Location: location,
		Param:    param,
		Value:    value,
		Msg:      msg,
	}}

	resp, err := json.Marshal(map[string][]ErrorResponse{"errors": errors})
	if err != nil {
		return
	}

	if _, err = w.Write(resp); err != nil {
		return
	}
}
