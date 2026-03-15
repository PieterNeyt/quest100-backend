package domain

import (
	"fmt"
	"time"
)

type Date struct {
	StartDate int64 `json:"startDate"`
	EndDate   int64 `json:"endDate"`
}

type SearchObject struct {
	TypeID   TypeID `json:"typeId"`
	ObjectID string `json:"objectId"`
}

type RequestBody struct {
	Date          Date           `json:"date"`
	IDFormat      string         `json:"idFormat"`
	SearchObjects []SearchObject `json:"searchObjects"`
}

type TypeID string

const (
	Student  TypeID = "person.student"
	Lector          = "person.staff"
	Teaching        = "activitytype.teaching"
)

type Object struct {
	ObjectID string `json:"objectId"`
	TypeID   string `json:"typeId"`
}

type Result struct {
	ID      int   `json:"id"`
	Begin   int64 `json:"begin"`
	End     int64 `json:"end"`
	Objects []Object
}

type ResponseBody struct {
	Limit        int      `json:"limit"`
	Page         int      `json:"page"`
	TotalPages   int      `json:"totalPages"`
	TotalResults int      `json:"totalResults"`
	Results      []Result `json:"results"`
}

func CreateRequestBody(typeID TypeID, employeeId int) *RequestBody {
	timeNow := time.Now().Unix()
	timeNow = 1773739800
	return &RequestBody{
		Date: Date{
			StartDate: timeNow,
			EndDate:   timeNow,
		},
		IDFormat: "EXTERNAL",
		//TODO nog verder aanvullen voor alle teaching type objects
		SearchObjects: []SearchObject{
			{
				TypeID:   typeID,
				ObjectID: fmt.Sprintf("person_%d", employeeId),
			},
			{
				TypeID:   Teaching,
				ObjectID: "_te_47675",
			},
			{
				TypeID:   Teaching,
				ObjectID: "_te_47677",
			},
		},
	}
}
