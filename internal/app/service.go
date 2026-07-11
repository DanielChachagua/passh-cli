package app

import (
	"errors"
	"fmt"
	"strings"

	"ssh-cli/internal/domain"
	"ssh-cli/internal/ports"

	"github.com/manifoldco/promptui"
)

// SshManagerServiceImpl implements ports.SshManagerService coordinating logic.
type SshManagerServiceImpl struct {
	apiScanner    ports.ApiScanner
	configStorage ports.ConfigStorage
	sshDialer     ports.SshDialer
}

// NewSshManagerService instantiates a new SshManagerServiceImpl.
func NewSshManagerService(apiScanner ports.ApiScanner, configStorage ports.ConfigStorage, sshDialer ports.SshDialer) *SshManagerServiceImpl {
	return &SshManagerServiceImpl{
		apiScanner:    apiScanner,
		configStorage: configStorage,
		sshDialer:     sshDialer,
	}
}

// Register registers a new user (requires active Admin user logged in).
func (s *SshManagerServiceImpl) Register(email string) error {
	token, err := s.configStorage.LoadToken()
	if err != nil {
		return fmt.Errorf("failed to read token: %w", err)
	}
	if token == "" {
		return errors.New("no active session found. You must login as an Admin first")
	}

	err = s.apiScanner.Register(token, email)
	if err != nil {
		if s.isAuthError(err) {
			s.handleSessionExpired()
			return errors.New("session expired")
		}
		return err
	}
	return nil
}

// Login verifies credentials against API scanner and saves resulting JWT token.
func (s *SshManagerServiceImpl) Login(email, password string) error {
	token, err := s.apiScanner.Login(email, password)
	if err != nil {
		return err
	}

	if err := s.configStorage.SaveToken(token); err != nil {
		return fmt.Errorf("failed to save token locally: %w", err)
	}

	return nil
}

// ConnectInteractive retrieves connections list, lets the user select one via a paginated CLI prompt, retrieves connection detail, and starts SSH session.
// If searchTerm is provided, it attempts to match directly by ID or filters the list by name/IP proximity (%like%).
func (s *SshManagerServiceImpl) ConnectInteractive(searchTerm string) error {
	token, err := s.configStorage.LoadToken()
	if err != nil {
		return fmt.Errorf("failed to read token: %w", err)
	}
	if token == "" {
		return errors.New("no active session found. Please run login first")
	}

	conns, err := s.apiScanner.GetConnections(token)
	if err != nil {
		if s.isAuthError(err) {
			s.handleSessionExpired()
			return errors.New("session expired")
		}
		return fmt.Errorf("failed to load connections: %w", err)
	}

	if len(conns) == 0 {
		return errors.New("no registered connections found in the database")
	}

	var targetConn domain.Connection
	foundDirectly := false

	if searchTerm != "" {
		var id uint
		_, err := fmt.Sscanf(searchTerm, "%d", &id)
		if err == nil {
			// Find connection by numeric ID
			for _, c := range conns {
				if c.ID == id {
					targetConn = c
					foundDirectly = true
					break
				}
			}
			if !foundDirectly {
				return fmt.Errorf("no connection found with ID %d", id)
			}
		} else {
			// Proximity search (%like% filter)
			var matches []domain.Connection
			termLower := strings.ToLower(searchTerm)
			for _, c := range conns {
				if strings.Contains(strings.ToLower(c.Name), termLower) ||
					strings.Contains(strings.ToLower(c.IP), termLower) ||
					strings.Contains(strings.ToLower(c.Details), termLower) {
					matches = append(matches, c)
				}
			}

			if len(matches) == 0 {
				return fmt.Errorf("no connections match search term '%s'", searchTerm)
			} else if len(matches) == 1 {
				// Single match: connect directly!
				targetConn = matches[0]
				foundDirectly = true
				fmt.Printf("Connecting directly to single match: \033[1;32m%s\033[0m (%s)\n", targetConn.Name, targetConn.IP)
			} else {
				// Multiple matches: narrow list to only matched connections
				conns = matches
				fmt.Printf("Multiple matches found for '%s'. Narrowing down selection list...\n", searchTerm)
			}
		}
	}

	if foundDirectly {
		detail, err := s.apiScanner.GetConnectionDetail(token, targetConn.ID)
		if err != nil {
			if s.isAuthError(err) {
				s.handleSessionExpired()
				return errors.New("session expired")
			}
			return fmt.Errorf("failed to retrieve connection details: %w", err)
		}

		sshConfig := domain.SshSessionConfig{
			IP:       detail.IP,
			Port:     detail.Port,
			User:     detail.Username,
			Password: detail.Password,
		}

		return s.sshDialer.Connect(sshConfig)
	}

	// Strategy A: Interactive prompt paging
	pageSize := 5 // Number of connections per page
	currentPage := 0
	totalItems := len(conns)

	for {
		start := currentPage * pageSize
		end := start + pageSize
		if end > totalItems {
			end = totalItems
		}

		pageItems := conns[start:end]

		// Wrap connections into custom item types so promptui templates can distinguish options
		type PromptItem struct {
			IsAction bool
			Action   string
			Conn     domain.Connection
		}

		var promptList []PromptItem
		for _, c := range pageItems {
			promptList = append(promptList, PromptItem{Conn: c})
		}

		// Inject virtual navigation options
		if end < totalItems {
			promptList = append(promptList, PromptItem{
				IsAction: true,
				Action:   "next",
				Conn:     domain.Connection{Name: "--> [Siguiente Página]", Details: "Ver los siguientes registros"},
			})
		}
		if currentPage > 0 {
			promptList = append([]PromptItem{{
				IsAction: true,
				Action:   "prev",
				Conn:     domain.Connection{Name: "<- [Página Anterior]", Details: "Ver los registros anteriores"},
			}}, promptList...)
		}

		templates := &promptui.SelectTemplates{
			Label: "{{ . }}",
			Active: `{{ if .IsAction }}➡️  ` + "\033[1;33m{{ .Conn.Name }}\033[0m" + `{{ else }}💻 ` + "\033[1;36m{{ .Conn.Name }}\033[0m (\033[33m{{ .Conn.Username }}@{{ .Conn.IP }}:{{ .Conn.Port }}\033[0m) [{{ .Conn.Details }}]" + `{{ end }}`,
			Inactive: `{{ if .IsAction }}    ` + "{{ .Conn.Name }}" + `{{ else }}    ` + "{{ .Conn.Name }} ({{ .Conn.Username }}@{{ .Conn.IP }}:{{ .Conn.Port }})" + `{{ end }}`,
			Selected: `🔌 Selected: \033[1;32m{{ .Conn.Name }}\033[0m`,
		}

		searcher := func(input string, index int) bool {
			item := promptList[index]
			if item.IsAction {
				return false
			}
			name := strings.ToLower(item.Conn.Name)
			ip := strings.ToLower(item.Conn.IP)
			details := strings.ToLower(item.Conn.Details)
			input = strings.ToLower(input)
			return strings.Contains(name, input) || strings.Contains(ip, input) || strings.Contains(details, input)
		}

		totalPages := (totalItems + pageSize - 1) / pageSize
		prompt := promptui.Select{
			Label:     fmt.Sprintf("Select server to connect (Page %d/%d, Total: %d):", currentPage+1, totalPages, totalItems),
			Items:     promptList,
			Templates: templates,
			Size:      10,
			Searcher:  searcher,
		}

		idx, _, err := prompt.Run()
		if err != nil {
			return err
		}

		selectedItem := promptList[idx]
		if selectedItem.IsAction {
			if selectedItem.Action == "next" {
				currentPage++
			} else if selectedItem.Action == "prev" {
				currentPage--
			}
			continue
		}

		// Selected a real connection configuration
		selected := selectedItem.Conn

		// Fetch details including decrypted password
		detail, err := s.apiScanner.GetConnectionDetail(token, selected.ID)
		if err != nil {
			if s.isAuthError(err) {
				s.handleSessionExpired()
				return errors.New("session expired")
			}
			return fmt.Errorf("failed to retrieve connection details: %w", err)
		}

		sshConfig := domain.SshSessionConfig{
			IP:       detail.IP,
			Port:     detail.Port,
			User:     detail.Username,
			Password: detail.Password,
		}

		return s.sshDialer.Connect(sshConfig)
	}
}

// CreateConnection adds a new server connection via the HTTP API using the stored JWT token.
func (s *SshManagerServiceImpl) CreateConnection(name, ip string, port int, username, password, details string) error {
	token, err := s.configStorage.LoadToken()
	if err != nil {
		return fmt.Errorf("failed to read token: %w", err)
	}
	if token == "" {
		return errors.New("no active session found. Please run login first")
	}

	err = s.apiScanner.CreateConnection(token, name, ip, port, username, password, details)
	if err != nil {
		if s.isAuthError(err) {
			s.handleSessionExpired()
			return errors.New("session expired")
		}
		return err
	}

	return nil
}

// ListConnections retrieves all user connections.
func (s *SshManagerServiceImpl) ListConnections() ([]domain.Connection, error) {
	token, err := s.configStorage.LoadToken()
	if err != nil {
		return nil, fmt.Errorf("failed to read token: %w", err)
	}
	if token == "" {
		return nil, errors.New("no active session found. Please run login first")
	}

	conns, err := s.apiScanner.GetConnections(token)
	if err != nil {
		if s.isAuthError(err) {
			s.handleSessionExpired()
			return nil, errors.New("session expired")
		}
		return nil, err
	}

	return conns, nil
}

// UpdateConnection updates an existing connection in the backend.
func (s *SshManagerServiceImpl) UpdateConnection(id uint, name, ip string, port int, username, password, details string) error {
	token, err := s.configStorage.LoadToken()
	if err != nil {
		return fmt.Errorf("failed to read token: %w", err)
	}
	if token == "" {
		return errors.New("no active session found. Please run login first")
	}

	err = s.apiScanner.UpdateConnection(token, id, name, ip, port, username, password, details)
	if err != nil {
		if s.isAuthError(err) {
			s.handleSessionExpired()
			return errors.New("session expired")
		}
		return err
	}

	return nil
}

// DeleteConnection deletes a connection in the backend.
func (s *SshManagerServiceImpl) DeleteConnection(id uint) error {
	token, err := s.configStorage.LoadToken()
	if err != nil {
		return fmt.Errorf("failed to read token: %w", err)
	}
	if token == "" {
		return errors.New("no active session found. Please run login first")
	}

	err = s.apiScanner.DeleteConnection(token, id)
	if err != nil {
		if s.isAuthError(err) {
			s.handleSessionExpired()
			return errors.New("session expired")
		}
		return err
	}

	return nil
}

// GetConnectionDetail retrieves full connection details including decrypted password.
func (s *SshManagerServiceImpl) GetConnectionDetail(id uint) (domain.ConnectionDetail, error) {
	token, err := s.configStorage.LoadToken()
	if err != nil {
		return domain.ConnectionDetail{}, fmt.Errorf("failed to read token: %w", err)
	}
	if token == "" {
		return domain.ConnectionDetail{}, errors.New("no active session found. Please run login first")
	}

	detail, err := s.apiScanner.GetConnectionDetail(token, id)
	if err != nil {
		if s.isAuthError(err) {
			s.handleSessionExpired()
			return domain.ConnectionDetail{}, errors.New("session expired")
		}
		return domain.ConnectionDetail{}, err
	}

	return detail, nil
}

// Logout clears the local session configurations.
func (s *SshManagerServiceImpl) Logout() error {
	return s.configStorage.ClearToken()
}

// Passwords Module

// CreatePassword registers a new password configuration in backend.
func (s *SshManagerServiceImpl) CreatePassword(name, user, password string) error {
	token, err := s.configStorage.LoadToken()
	if err != nil {
		return fmt.Errorf("failed to read token: %w", err)
	}
	if token == "" {
		return errors.New("no active session found. Please run login first")
	}

	err = s.apiScanner.CreatePassword(token, name, user, password)
	if err != nil {
		if s.isAuthError(err) {
			s.handleSessionExpired()
			return errors.New("session expired")
		}
		return err
	}

	return nil
}

// ListPasswords lists all stored passwords.
func (s *SshManagerServiceImpl) ListPasswords() ([]domain.Password, error) {
	token, err := s.configStorage.LoadToken()
	if err != nil {
		return nil, fmt.Errorf("failed to read token: %w", err)
	}
	if token == "" {
		return nil, errors.New("no active session found. Please run login first")
	}

	passwords, err := s.apiScanner.GetPasswords(token)
	if err != nil {
		if s.isAuthError(err) {
			s.handleSessionExpired()
			return nil, errors.New("session expired")
		}
		return nil, err
	}

	return passwords, nil
}

// GetPasswordDetail retrieves password detail including decrypted value.
func (s *SshManagerServiceImpl) GetPasswordDetail(id uint) (domain.PasswordDetail, error) {
	token, err := s.configStorage.LoadToken()
	if err != nil {
		return domain.PasswordDetail{}, fmt.Errorf("failed to read token: %w", err)
	}
	if token == "" {
		return domain.PasswordDetail{}, errors.New("no active session found. Please run login first")
	}

	detail, err := s.apiScanner.GetPasswordDetail(token, id)
	if err != nil {
		if s.isAuthError(err) {
			s.handleSessionExpired()
			return domain.PasswordDetail{}, errors.New("session expired")
		}
		return domain.PasswordDetail{}, err
	}

	return detail, nil
}

// UpdatePassword updates an existing password record.
func (s *SshManagerServiceImpl) UpdatePassword(id uint, name, user, password string) error {
	token, err := s.configStorage.LoadToken()
	if err != nil {
		return fmt.Errorf("failed to read token: %w", err)
	}
	if token == "" {
		return errors.New("no active session found. Please run login first")
	}

	err = s.apiScanner.UpdatePassword(token, id, name, user, password)
	if err != nil {
		if s.isAuthError(err) {
			s.handleSessionExpired()
			return errors.New("session expired")
		}
		return err
	}

	return nil
}

// DeletePassword deletes a password record.
func (s *SshManagerServiceImpl) DeletePassword(id uint) error {
	token, err := s.configStorage.LoadToken()
	if err != nil {
		return fmt.Errorf("failed to read token: %w", err)
	}
	if token == "" {
		return errors.New("no active session found. Please run login first")
	}

	err = s.apiScanner.DeletePassword(token, id)
	if err != nil {
		if s.isAuthError(err) {
			s.handleSessionExpired()
			return errors.New("session expired")
		}
		return err
	}

	return nil
}

// ResetPasswordRequest triggers password reset token dispatch.
func (s *SshManagerServiceImpl) ResetPasswordRequest(email string) error {
	return s.apiScanner.ResetPasswordRequest(email)
}

// ConfirmPasswordReset commits password updates using token.
func (s *SshManagerServiceImpl) ConfirmPasswordReset(token, newPassword, confirmPassword string) error {
	return s.apiScanner.ConfirmPasswordReset(token, newPassword, confirmPassword)
}

// Groups Module

// CreateGroup sends a group creation request to the API.
func (s *SshManagerServiceImpl) CreateGroup(name string) (domain.Group, error) {
	token, err := s.configStorage.LoadToken()
	if err != nil {
		return domain.Group{}, fmt.Errorf("failed to read token: %w", err)
	}
	if token == "" {
		return domain.Group{}, errors.New("no active session found. Please run login first")
	}

	group, err := s.apiScanner.CreateGroup(token, name)
	if err != nil {
		if s.isAuthError(err) {
			s.handleSessionExpired()
			return domain.Group{}, errors.New("session expired")
		}
		return domain.Group{}, err
	}
	return group, nil
}

// ListGroups lists all groups the authenticated user belongs to.
func (s *SshManagerServiceImpl) ListGroups() ([]domain.Group, error) {
	token, err := s.configStorage.LoadToken()
	if err != nil {
		return nil, fmt.Errorf("failed to read token: %w", err)
	}
	if token == "" {
		return nil, errors.New("no active session found. Please run login first")
	}

	groups, err := s.apiScanner.GetGroups(token)
	if err != nil {
		if s.isAuthError(err) {
			s.handleSessionExpired()
			return nil, errors.New("session expired")
		}
		return nil, err
	}
	return groups, nil
}

// GetGroupDetail retrieves group details, members, and shared assets.
func (s *SshManagerServiceImpl) GetGroupDetail(id uint) (domain.GroupDetailResponse, error) {
	token, err := s.configStorage.LoadToken()
	if err != nil {
		return domain.GroupDetailResponse{}, fmt.Errorf("failed to read token: %w", err)
	}
	if token == "" {
		return domain.GroupDetailResponse{}, errors.New("no active session found. Please run login first")
	}

	detail, err := s.apiScanner.GetGroupDetail(token, id)
	if err != nil {
		if s.isAuthError(err) {
			s.handleSessionExpired()
			return domain.GroupDetailResponse{}, errors.New("session expired")
		}
		return domain.GroupDetailResponse{}, err
	}
	return detail, nil
}

// AddGroupMember adds a member to a group using their email.
func (s *SshManagerServiceImpl) AddGroupMember(groupID uint, email string) error {
	token, err := s.configStorage.LoadToken()
	if err != nil {
		return fmt.Errorf("failed to read token: %w", err)
	}
	if token == "" {
		return errors.New("no active session found. Please run login first")
	}

	err = s.apiScanner.AddGroupMember(token, groupID, email)
	if err != nil {
		if s.isAuthError(err) {
			s.handleSessionExpired()
			return errors.New("session expired")
		}
		return err
	}
	return nil
}

// RemoveGroupMember removes a user membership from the group.
func (s *SshManagerServiceImpl) RemoveGroupMember(groupID uint, userID uint) error {
	token, err := s.configStorage.LoadToken()
	if err != nil {
		return fmt.Errorf("failed to read token: %w", err)
	}
	if token == "" {
		return errors.New("no active session found. Please run login first")
	}

	err = s.apiScanner.RemoveGroupMember(token, groupID, userID)
	if err != nil {
		if s.isAuthError(err) {
			s.handleSessionExpired()
			return errors.New("session expired")
		}
		return err
	}
	return nil
}

// ShareConnection shares an SSH connection with a group.
func (s *SshManagerServiceImpl) ShareConnection(groupID uint, connectionID uint) error {
	token, err := s.configStorage.LoadToken()
	if err != nil {
		return fmt.Errorf("failed to read token: %w", err)
	}
	if token == "" {
		return errors.New("no active session found. Please run login first")
	}

	err = s.apiScanner.ShareConnection(token, groupID, connectionID)
	if err != nil {
		if s.isAuthError(err) {
			s.handleSessionExpired()
			return errors.New("session expired")
		}
		return err
	}
	return nil
}

// UnshareConnection stops sharing an SSH connection with a group.
func (s *SshManagerServiceImpl) UnshareConnection(groupID uint, connectionID uint) error {
	token, err := s.configStorage.LoadToken()
	if err != nil {
		return fmt.Errorf("failed to read token: %w", err)
	}
	if token == "" {
		return errors.New("no active session found. Please run login first")
	}

	err = s.apiScanner.UnshareConnection(token, groupID, connectionID)
	if err != nil {
		if s.isAuthError(err) {
			s.handleSessionExpired()
			return errors.New("session expired")
		}
		return err
	}
	return nil
}

// SharePassword shares a password with a group.
func (s *SshManagerServiceImpl) SharePassword(groupID uint, passwordID uint) error {
	token, err := s.configStorage.LoadToken()
	if err != nil {
		return fmt.Errorf("failed to read token: %w", err)
	}
	if token == "" {
		return errors.New("no active session found. Please run login first")
	}

	err = s.apiScanner.SharePassword(token, groupID, passwordID)
	if err != nil {
		if s.isAuthError(err) {
			s.handleSessionExpired()
			return errors.New("session expired")
		}
		return err
	}
	return nil
}

// UnsharePassword stops sharing a password with a group.
func (s *SshManagerServiceImpl) UnsharePassword(groupID uint, passwordID uint) error {
	token, err := s.configStorage.LoadToken()
	if err != nil {
		return fmt.Errorf("failed to read token: %w", err)
	}
	if token == "" {
		return errors.New("no active session found. Please run login first")
	}

	err = s.apiScanner.UnsharePassword(token, groupID, passwordID)
	if err != nil {
		if s.isAuthError(err) {
			s.handleSessionExpired()
			return errors.New("session expired")
		}
		return err
	}
	return nil
}

// DeleteGroup deletes a group configuration in backend.
func (s *SshManagerServiceImpl) DeleteGroup(groupID uint) error {
	token, err := s.configStorage.LoadToken()
	if err != nil {
		return fmt.Errorf("failed to read token: %w", err)
	}
	if token == "" {
		return errors.New("no active session found. Please run login first")
	}

	err = s.apiScanner.DeleteGroup(token, groupID)
	if err != nil {
		if s.isAuthError(err) {
			s.handleSessionExpired()
			return errors.New("session expired")
		}
		return err
	}
	return nil
}

func (s *SshManagerServiceImpl) isAuthError(err error) bool {
	return errors.Is(err, domain.ErrUnauthorized)
}

func (s *SshManagerServiceImpl) handleSessionExpired() {
	_ = s.configStorage.ClearToken()
	fmt.Println("\n\033[1;31mError: Your session has expired or authentication failed!\033[0m")
	fmt.Println("\033[33mPlease log in again by running: passh user login\033[0m\n")
}
