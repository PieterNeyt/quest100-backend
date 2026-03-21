package api

type SubmitGuessRequest struct {
	Guess string `json:"guess" binding:"required"`
}
