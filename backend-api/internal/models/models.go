package models

import "time"

type User struct {
	ID        int       `json:"id"`
	NumericID string    `json:"numeric_id"`
	Name      string    `json:"name"`
	StreamKey string    `json:"stream_key,omitempty"`
	Bitrate   *int      `json:"bitrate,omitempty"` // User's configured upload bitrate in kbps
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Destination struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	RTMPURL   string    `json:"rtmp_url"`
	CreatedAt time.Time `json:"created_at"`
}

type UserDestination struct {
	ID             int          `json:"id"`
	UserID         int          `json:"user_id"`
	DestinationID  int          `json:"destination_id"`
	Destination    *Destination `json:"destination,omitempty"`
	StreamKey      string       `json:"stream_key,omitempty"`
	CreatedAt      time.Time    `json:"created_at"`
	UpdatedAt      time.Time    `json:"updated_at"`
}

type BroadcastSession struct {
	ID                int       `json:"id"`
	UserID            int       `json:"user_id"`
	UserDestinationID int       `json:"user_destination_id"`
	Date              time.Time `json:"date"`
	DurationMinutes  int       `json:"duration_minutes"`
	StartedAt         time.Time `json:"started_at"`
	EndedAt           *time.Time `json:"ended_at,omitempty"`
}

type UserMachine struct {
	ID         int       `json:"id"`
	MachineID  string    `json:"machine_id"`
	UserID     int       `json:"user_id"`
	LastUsedAt time.Time `json:"last_used_at"`
}

type UserSession struct {
	ID             int       `json:"id"`
	UserID         int       `json:"user_id"`
	MachineID      string    `json:"machine_id"`
	AccessToken    string    `json:"access_token,omitempty"`
	RefreshToken   string    `json:"refresh_token,omitempty"`
	TokenExpiresAt time.Time `json:"token_expires_at"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type UserWithDestinations struct {
	User         *User
	Destinations []UserDestination
}