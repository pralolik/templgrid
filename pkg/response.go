package pkg

type ErrorResponse struct {
	Ok    bool   `json:"ok"`
	Error string `json:"error"`
}

type SuccessfulResponse struct {
	Ok      bool   `json:"ok"`
	Message string `json:"message"`
}
