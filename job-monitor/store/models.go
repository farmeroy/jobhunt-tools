package store

import "time"

type Company struct {
	ID      int64
	Name    string
	Slug    string
	ATS     string
	Active  bool
	AddedAt time.Time
}

type Posting struct {
	ID               int64
	CompanyID        int64
	CompanyName      string
	ExternalID       string
	Title            string
	URL              string
	Location         string
	Description      string
	FirstSeen        time.Time
	LastSeen         time.Time
	TitleRelevant    *bool
	ContentScore     *float64
	ContentReasoning *string
	NotifiedAt       *time.Time
	AppliedAt        *time.Time
}
