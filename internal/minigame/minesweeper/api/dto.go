package api

type SubmitMoveRequest struct {
	Action string `json:"action" binding:"required"`
	Row    int    `json:"row"`
	Col    int    `json:"col"`
}
