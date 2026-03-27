package domain

import (
	"fmt"
	"strings"
	"time"
)

type TypeID string

const (
	Student  TypeID = "person.student"
	Lector          = "person.staff"
	Teaching        = "activitytype.teaching"
)

type Field struct {
	FieldID string   `json:"fieldId"`
	Values  []string `json:"values"`
}

type Object struct {
	ObjectID string `json:"objectId"`
	TypeID   string `json:"typeId"`
}

type Result struct {
	ID      int      `json:"id"`
	Begin   int64    `json:"begin"`
	End     int64    `json:"end"`
	Status  []string `json:"status"`
	Objects []Object `json:"objects"`
	Fields  []Field  `json:"fields"`
}

type ResponseBody struct {
	Limit        int      `json:"limit"`
	Page         int      `json:"page"`
	TotalPages   int      `json:"totalPages"`
	TotalResults int      `json:"totalResults"`
	Results      []Result `json:"results"`
}

type CourseObject struct {
	ExtID  string  `json:"extId"`
	Fields []Field `json:"fields"`
}

type CourseObjectsResponse struct {
	Results []CourseObject `json:"results"`
}

type Date struct {
	StartDate int64 `json:"startDate"`
	EndDate   int64 `json:"endDate"`
}

type SearchObject struct {
	TypeID   TypeID `json:"typeId"`
	ObjectID string `json:"objectId"`
}

type TypeField struct {
	Type   string   `json:"type"`
	Fields []string `json:"fields"`
}

type RequestBody struct {
	Date          Date           `json:"date"`
	IDFormat      string         `json:"idFormat"`
	SearchObjects []SearchObject `json:"searchObjects"`
	TypeFields    []TypeField    `json:"typeFields,omitempty"`
	Statuses      []string       `json:"statuses,omitempty"`
}

type AgendaItem struct {
	ID         int      `json:"id"`
	Begin      int64    `json:"begin"`
	End        int64    `json:"end"`
	Status     string   `json:"status"`
	CourseName string   `json:"courseName"`
	Rooms      []string `json:"rooms"`
	Activity   string   `json:"activity"`
	Attended   bool     `json:"attended"`
}

func CreateRequestBody(typeID TypeID, employeeId int) *RequestBody {
	timeNow := time.Now().Unix()
	// TODO dit gebruiken voor een valide les voor een andere les te gebruiken moet je een datum naar unix (int64) formaat zetten
	//timeNow = 1773925800
	return &RequestBody{
		Date: Date{
			StartDate: timeNow,
			EndDate:   timeNow,
		},
		IDFormat: "EXTERNAL",
		SearchObjects: []SearchObject{
			{
				TypeID:   typeID,
				ObjectID: fmt.Sprintf("person_%d", employeeId),
			},
			{TypeID: Teaching, ObjectID: "_te_47675"},
			{TypeID: Teaching, ObjectID: "_te_47676"},
			{TypeID: Teaching, ObjectID: "_te_47677"},
			{TypeID: Teaching, ObjectID: "_te_47678"},
			{TypeID: Teaching, ObjectID: "_te_47679"},
			{TypeID: Teaching, ObjectID: "_te_47680"},
		},
	}
}

func CreateAgendaRequestBody(employeeId int, isStudent bool) *RequestBody {
	typeID := TypeID(Lector)
	if isStudent {
		typeID = Student
	}

	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).Unix()
	endOfDay := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, now.Location()).Unix()

	return &RequestBody{
		Date:     Date{StartDate: startOfDay, EndDate: endOfDay},
		IDFormat: "EXTERNAL",
		SearchObjects: []SearchObject{
			{TypeID: typeID, ObjectID: fmt.Sprintf("person_%d", employeeId)},
			{TypeID: Teaching, ObjectID: "_te_47675"},
			{TypeID: Teaching, ObjectID: "_te_47676"},
			{TypeID: Teaching, ObjectID: "_te_47677"},
			{TypeID: Teaching, ObjectID: "_te_47678"},
			{TypeID: Teaching, ObjectID: "_te_47679"},
			{TypeID: Teaching, ObjectID: "_te_47680"},
		},
	}
}

func AgendaItemFromResult(r Result) AgendaItem {
	item := AgendaItem{
		ID:    r.ID,
		Begin: r.Begin,
		End:   r.End,
	}

	if len(r.Status) > 0 {
		item.Status = r.Status[0]
	}

	for _, f := range r.Fields {
		if f.FieldID == "r.title" && len(f.Values) > 0 && f.Values[0] != "" {
			item.CourseName = f.Values[0]
			break
		}
	}

	for _, obj := range r.Objects {
		switch obj.TypeID {
		case "rooms":
			item.Rooms = append(item.Rooms, strings.TrimPrefix(obj.ObjectID, "room_"))
		case "activitytype.teaching":
			item.Activity = obj.ObjectID
		}
	}

	return item
}

func (item *AgendaItem) ResolveCourseFromObjects(objects []Object, courseNames map[string]string) {
	if item.CourseName != "" {
		return
	}
	for _, obj := range objects {
		if obj.TypeID == "course" {
			if name, ok := courseNames[obj.ObjectID]; ok {
				item.CourseName = name
			} else {
				item.CourseName = obj.ObjectID
			}
			return
		}
	}
}
func (rb *ResponseBody) UniqueCourseIDs() []string {
	seen := map[string]bool{}
	var ids []string
	for _, r := range rb.Results {
		for _, obj := range r.Objects {
			if obj.TypeID == "course" && !seen[obj.ObjectID] {
				ids = append(ids, obj.ObjectID)
				seen[obj.ObjectID] = true
			}
		}
	}
	return ids
}

func (r *CourseObjectsResponse) ToNameMap() map[string]string {
	names := make(map[string]string, len(r.Results))
	for _, obj := range r.Results {
		for _, f := range obj.Fields {
			if f.FieldID == "course.short.name" && len(f.Values) > 0 && f.Values[0] != "" {
				names[obj.ExtID] = f.Values[0]
				break
			}
		}
	}
	return names
}
