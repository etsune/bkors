package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type EditStatus int

const (
	StatusNew EditStatus = iota
	StatusApproved
	StatusDeclined
)

type DictEntry struct {
	Hangul          string `bson:"hangul"`
	Hanja           string `bson:"hanja"`
	HomonymicNumber string `bson:"hn"`
	Transcription   string `bson:"ts"`
	Body            string `bson:"body"`
}

type Placement struct {
	Volume    int    `bson:"v"`
	Page      int    `bson:"pg"`
	Side      int    `bson:"s"`
	Paragraph int    `bson:"pr"`
	Coords    string `bson:"c"`
}

type CreateEntryRequest struct {
	Hangul          string
	Hanja           string
	HomonymicNumber string
	Body            string
	Image           string
	Placement       Placement
}

type DBEntry struct {
	Id            primitive.ObjectID `bson:"_id"`
	Entry         DictEntry          `bson:"entry"`
	IsReviewed    bool               `bson:"rev"`
	Placement     Placement          `bson:"placement"`
	HeaderSearch  []string           `bson:"header_search"`
	BodySearch    string             `bson:"body_search"`
	Image         string             `bson:"image"`
	Dict          string             `bson:"dict"`
	PlacementSort int                `bson:"psort"`
	CreatedAt     time.Time          `bson:"created_at"`
	UpdatedAt     time.Time          `bson:"updated_at"`
}

type DBPage struct {
	Id         primitive.ObjectID `bson:"_id"`
	Dictionary string             `bson:"dict"`
	Volume     int                `bson:"vol"`
	Page       int                `bson:"p"`
	File       string             `bson:"file"`
	Width      int                `bson:"w"`
	Height     int                `bson:"h"`
	Num        int                `bson:"num"`
}

type EditEntry struct {
	IsReviewed      bool   `bson:"is_reviwed" form:"is_reviewed"`
	Hangul          string `bson:"hangul" form:"hangul"`
	Hanja           string `bson:"hanja" form:"hanja"`
	HomonymicNumber string `bson:"hn" form:"hn"`
	Transcription   string `bson:"ts" form:"ts"`
	Body            string `bson:"body" form:"body"`
	Meta            string `bson:"meta" form:"meta"`
}

type DBEdit struct {
	Id        primitive.ObjectID `bson:"_id"`
	EntryId   primitive.ObjectID `bson:"entry_id"`
	Status    EditStatus         `bson:"status"`
	Source    EditEntry          `bson:"source"`
	Result    EditEntry          `bson:"result"`
	Author    string             `bson:"author"`
	Approver  string             `bson:"approver"`
	Image     string             `bson:"image"`
	CreatedAt time.Time          `bson:"created_at"`
}

// Body - loan source , senses-examples(kor, rus)

type Page struct {
	Volume     int
	PageNumber int
	Image      string
}

type Edit struct {
	OldEntry  string    `json:"old_entry,omitempty" bson:"old_entry,omitempty"`
	NewEntry  string    `json:"new_entry,omitempty" bson:"new_entry,omitempty"`
	Comment   string    `json:"comment,omitempty" bson:"comment,omitempty"`
	CreatedAt time.Time `json:"created_at,omitempty" bson:"created_at,omitempty"`
	UpdatedAt time.Time `json:"updated_at,omitempty" bson:"updated_at,omitempty"`
}

type DBUser struct {
	Id             primitive.ObjectID `bson:"_id"`
	Username       string             `bson:"username"`
	Password       string             `bson:"password"`
	IsAdmin        bool               `bson:"is_admin"`
	HasAutoApprove bool               `bson:"aappr"`
	CreatedAt      time.Time          `bson:"created_at"`
	UpdatedAt      time.Time          `bson:"updated_at"`
}

type DBDownload struct {
	Id       primitive.ObjectID `bson:"_id"`
	Filename string             `bson:"filename"`
	Path     string             `bson:"path"`
	Time     string             `bson:"time"`
	Size     string             `bson:"size"`
}
