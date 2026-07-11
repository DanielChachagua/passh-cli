package domain

import "errors"

// ErrUnauthorized represents a session expiration or invalid authentication token.
var ErrUnauthorized = errors.New("AUTH_FAILED")

// SshSessionConfig represents the SSH connection detail needed to establish a session.
type SshSessionConfig struct {
	IP       string `json:"ip"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
}

// Connection represents a brief representation of a remote server.
type Connection struct {
	ID       uint   `json:"id"`
	Name     string `json:"name"`
	IP       string `json:"ip"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Details  string `json:"details"`
}

// ConnectionDetail represents the full detail of a remote server including credentials.
type ConnectionDetail struct {
	ID       uint   `json:"id"`
	Name     string `json:"name"`
	IP       string `json:"ip"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
	Details  string `json:"details"`
}

// Password represents a brief representation of a credential entry.
type Password struct {
	ID     uint   `json:"id"`
	UserID uint   `json:"user_id"`
	Name   string `json:"name"`
	User   string `json:"user"`
}

// PasswordDetail represents the full detail of a credential including the decrypted password.
type PasswordDetail struct {
	ID       uint   `json:"id"`
	UserID   uint   `json:"user_id"`
	Name     string `json:"name"`
	User     string `json:"user"`
	Password string `json:"password"`
}

// Group represents a group entity.
type Group struct {
	ID        uint   `json:"id"`
	Name      string `json:"name"`
	OwnerID   uint   `json:"owner_id"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// GroupMember represents a member of a group.
type GroupMember struct {
	ID    uint   `json:"id"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

// GroupDetailResponse holds the payload returned by GET /api/groups/:id.
type GroupDetailResponse struct {
	Group       Group        `json:"group"`
	Members     []GroupMember `json:"members"`
	Connections []Connection `json:"connections"`
	Passwords   []Password   `json:"passwords"`
}

// APIError represents structural API errors returned in case Success is false.
type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// ApiResponse defines the wrapper for any response from the external API.
type ApiResponse[T any] struct {
	Success bool     `json:"success"`
	Data    T        `json:"data"`
	Error   APIError `json:"error"`
}
