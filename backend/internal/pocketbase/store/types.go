package store

import (
	"time"
)

type GroupStatistics struct {
	TotalUsers        int     `json:"totalUsers"`
	TotalItems        int     `json:"totalItems"`
	TotalLocations    int     `json:"totalLocations"`
	TotalLabels       int     `json:"totalLabels"`
	TotalItemPrice    float64 `json:"totalItemPrice"`
	TotalWithWarranty int     `json:"totalWithWarranty"`
}

type ValueOverTimeEntry struct {
	Date  time.Time `json:"date"`
	Value float64   `json:"value"`
	Name  string    `json:"name"`
}

type ValueOverTime struct {
	PriceAtStart float64              `json:"valueAtStart"`
	PriceAtEnd   float64              `json:"valueAtEnd"`
	Start        time.Time            `json:"start"`
	End          time.Time            `json:"end"`
	Entries      []ValueOverTimeEntry `json:"entries"`
}

type TotalsByOrganizer struct {
	ID    string  `json:"id"`
	Name  string  `json:"name"`
	Total float64 `json:"total"`
}

type Group struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Currency  string    `json:"currency"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type GroupUpdate struct {
	Name     string `json:"name"`
	Currency string `json:"currency"`
}

type GroupInvitationCreate struct {
	ExpiresAt time.Time `json:"expiresAt"`
	Uses      int       `json:"uses"`
}

type GroupInvitation struct {
	ID        string    `json:"id"`
	ExpiresAt time.Time `json:"expiresAt"`
	Uses      int       `json:"uses"`
	Token     string    `json:"token,omitempty"`
	Group     Group     `json:"group"`
}

type UserRegistration struct {
	GroupToken string `json:"token"`
	Name       string `json:"name"`
	Email      string `json:"email"`
	Password   string `json:"password"`
}

type TokenResponse struct {
	Token           string    `json:"token"`
	AttachmentToken string    `json:"attachmentToken"`
	ExpiresAt       time.Time `json:"expiresAt"`
}

type ItemPath struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type LocationTree struct {
	ID       string         `json:"id"`
	Name     string         `json:"name"`
	Type     string         `json:"type"`
	Children []LocationTree `json:"children"`
}

type ActionAmountResult struct {
	Completed int `json:"completed"`
}

type Build struct {
	Version   string `json:"version"`
	Commit    string `json:"commit"`
	BuildTime string `json:"buildTime"`
}

type APISummary struct {
	Health            bool     `json:"health"`
	Build             Build    `json:"build"`
	Versions          []string `json:"versions"`
	Title             string   `json:"title"`
	Message           string   `json:"message"`
	AllowRegistration bool     `json:"allowRegistration"`
	Demo              bool     `json:"demo"`
}
