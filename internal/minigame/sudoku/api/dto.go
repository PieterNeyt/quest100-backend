package api

type SubmitMoveRequest struct {
	Row   int `json:"row" binding:"min=0,max=8"`
	Col   int `json:"col" binding:"min=0,max=8"`
	Value int `json:"value" binding:"min=0,max=9"`
}
