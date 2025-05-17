package routes

import "net/http"

type ErrorResponse struct {
	Status  int    `json:"status,omitempty"`
	Message string `json:"message,omitempty"`
	URL     string `json:"url,omitempty"`
}

func (u *ErrorResponse) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}
