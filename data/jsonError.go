package data

import "net/http"

/**
structured this way for legacy reasons. Millicore expects a response in this format
*/
type ErrorJSON struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
	Status  string `json:"status"`
}

func (e *ErrorJSON) Error() string {
	return e.Message
}

func NewErrorJSON(message string, code int) *ErrorJSON {
	return &ErrorJSON{Message: message, Code: code, Status: "error"}
}

func NewErrorJSONBadRequest(msg string) *ErrorJSON {
	err := &ErrorJSON{Message: "could not parse request data " + msg, Code: http.StatusBadRequest, Status: "error"}
	return err
}

func NewErrorJSONUnexpectedError(msg string) *ErrorJSON {
	return NewErrorJSON("unexpected error occured  "+msg, http.StatusInternalServerError)
}

func NewErrorJSONNotFound(msg string) *ErrorJSON {
	return NewErrorJSON("resource not found "+msg, http.StatusNotFound)
}

func NewErrorJsonNoContent(msg string) *ErrorJSON {
	return NewErrorJSON("no content "+msg, http.StatusNoContent)
}
