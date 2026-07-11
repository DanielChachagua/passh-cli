package ports

import "ssh-cli/internal/domain"

// ApiScanner represents the outbound port to interact with the external HTTP API.
type ApiScanner interface {
	Login(username, password string) (string, error)
	Register(token string, email string) error
	ResetPasswordRequest(email string) error
	ConfirmPasswordReset(token, newPassword, confirmPassword string) error

	// SSH Connections
	GetConnections(token string) ([]domain.Connection, error)
	GetConnectionDetail(token string, id uint) (domain.ConnectionDetail, error)
	CreateConnection(token string, name, ip string, port int, username, password, details string) error
	UpdateConnection(token string, id uint, name, ip string, port int, username, password, details string) error
	DeleteConnection(token string, id uint) error

	// Passwords Module
	GetPasswords(token string) ([]domain.Password, error)
	GetPasswordDetail(token string, id uint) (domain.PasswordDetail, error)
	CreatePassword(token string, name, user, password string) error
	UpdatePassword(token string, id uint, name, user, password string) error
	DeletePassword(token string, id uint) error

	// Groups Module
	CreateGroup(token string, name string) (domain.Group, error)
	GetGroups(token string) ([]domain.Group, error)
	GetGroupDetail(token string, id uint) (domain.GroupDetailResponse, error)
	AddGroupMember(token string, groupID uint, email string) error
	RemoveGroupMember(token string, groupID uint, userID uint) error
	ShareConnection(token string, groupID uint, connectionID uint) error
	UnshareConnection(token string, groupID uint, connectionID uint) error
	SharePassword(token string, groupID uint, passwordID uint) error
	UnsharePassword(token string, groupID uint, passwordID uint) error
	DeleteGroup(token string, groupID uint) error
}

// ConfigStorage represents the outbound port to read, write and clear local configuration files.
type ConfigStorage interface {
	SaveToken(token string) error
	LoadToken() (string, error)
	ClearToken() error
}

// SshDialer represents the outbound port to open an interactive SSH connection.
type SshDialer interface {
	Connect(config domain.SshSessionConfig) error
}

// SshManagerService represents the inbound port containing the application logic.
type SshManagerService interface {
	Login(username, password string) error
	Register(email string) error
	Logout() error
	ResetPasswordRequest(email string) error
	ConfirmPasswordReset(token, newPassword, confirmPassword string) error

	// SSH Connections
	ConnectInteractive(searchTerm string) error
	CreateConnection(name, ip string, port int, username, password, details string) error
	ListConnections() ([]domain.Connection, error)
	UpdateConnection(id uint, name, ip string, port int, username, password, details string) error
	DeleteConnection(id uint) error
	GetConnectionDetail(id uint) (domain.ConnectionDetail, error)

	// Passwords Module
	CreatePassword(name, user, password string) error
	ListPasswords() ([]domain.Password, error)
	GetPasswordDetail(id uint) (domain.PasswordDetail, error)
	UpdatePassword(id uint, name, user, password string) error
	DeletePassword(id uint) error

	// Groups Module
	CreateGroup(name string) (domain.Group, error)
	ListGroups() ([]domain.Group, error)
	GetGroupDetail(id uint) (domain.GroupDetailResponse, error)
	AddGroupMember(groupID uint, email string) error
	RemoveGroupMember(groupID uint, userID uint) error
	ShareConnection(groupID uint, connectionID uint) error
	UnshareConnection(groupID uint, connectionID uint) error
	SharePassword(groupID uint, passwordID uint) error
	UnsharePassword(groupID uint, passwordID uint) error
	DeleteGroup(groupID uint) error
}
