package http

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"ssh-cli/internal/domain"
	"time"
)

// HttpApiScanner implements ports.ApiScanner using net/http.
type HttpApiScanner struct {
	baseURL    string
	httpClient *http.Client
}

// NewHttpApiScanner creates a new HttpApiScanner instance.
func NewHttpApiScanner(baseURL string) *HttpApiScanner {
	return &HttpApiScanner{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Register registers a new user in the system (requires Admin JWT token).
func (h *HttpApiScanner) Register(token string, email string) error {
	url := fmt.Sprintf("%s/passh/auth/register", h.baseURL)
	reqBody, err := json.Marshal(map[string]string{
		"email": email,
	})
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := h.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return domain.ErrUnauthorized
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var apiResp domain.ApiResponse[interface{}]
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return fmt.Errorf("failed to parse register response: %w", err)
	}

	if !apiResp.Success {
		return fmt.Errorf("registration failed: %s (%s)", apiResp.Error.Message, apiResp.Error.Code)
	}

	return nil
}

// Login requests a JWT token from the API using email and password.
func (h *HttpApiScanner) Login(email, password string) (string, error) {
	url := fmt.Sprintf("%s/passh/auth/login", h.baseURL)
	reqBody, err := json.Marshal(map[string]string{
		"email":    email,
		"password": password,
	})
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := h.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return "", domain.ErrUnauthorized
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var apiResp domain.ApiResponse[map[string]string]
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return "", fmt.Errorf("failed to parse login response: %w", err)
	}

	if !apiResp.Success {
		if apiResp.Error.Code == "AUTH_FAILED" {
			return "", domain.ErrUnauthorized
		}
		return "", fmt.Errorf("login failed: %s (%s)", apiResp.Error.Message, apiResp.Error.Code)
	}

	token, ok := apiResp.Data["token"]
	if !ok {
		return "", errors.New("token not found in login response")
	}

	return token, nil
}

// ResetPasswordRequest triggers password reset token dispatch.
func (h *HttpApiScanner) ResetPasswordRequest(email string) error {
	url := fmt.Sprintf("%s/passh/auth/reset-password/request", h.baseURL)
	reqBody, err := json.Marshal(map[string]string{
		"email": email,
	})
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := h.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var apiResp domain.ApiResponse[interface{}]
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return fmt.Errorf("failed to parse reset request response: %w", err)
	}

	if !apiResp.Success {
		return fmt.Errorf("reset request failed: %s (%s)", apiResp.Error.Message, apiResp.Error.Code)
	}

	return nil
}

// ConfirmPasswordReset commits password updates using token.
func (h *HttpApiScanner) ConfirmPasswordReset(token, newPassword, confirmPassword string) error {
	url := fmt.Sprintf("%s/passh/auth/reset-password/confirm", h.baseURL)
	reqBody, err := json.Marshal(map[string]string{
		"token":            token,
		"new_password":     newPassword,
		"confirm_password": confirmPassword,
	})
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := h.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var apiResp domain.ApiResponse[interface{}]
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return fmt.Errorf("failed to parse reset confirmation response: %w", err)
	}

	if !apiResp.Success {
		return fmt.Errorf("reset confirmation failed: %s (%s)", apiResp.Error.Message, apiResp.Error.Code)
	}

	return nil
}

// GetConnections lists all connections for the authenticated user.
func (h *HttpApiScanner) GetConnections(token string) ([]domain.Connection, error) {
	url := fmt.Sprintf("%s/passh/connections", h.baseURL)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := h.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, domain.ErrUnauthorized
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var apiResp domain.ApiResponse[[]domain.Connection]
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse connections response: %w", err)
	}

	if !apiResp.Success {
		return nil, fmt.Errorf("failed to list connections: %s (%s)", apiResp.Error.Message, apiResp.Error.Code)
	}

	return apiResp.Data, nil
}

// GetConnectionDetail fetches specific connection details including the decrypted password.
func (h *HttpApiScanner) GetConnectionDetail(token string, id uint) (domain.ConnectionDetail, error) {
	url := fmt.Sprintf("%s/passh/connections/%d", h.baseURL, id)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return domain.ConnectionDetail{}, err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := h.httpClient.Do(req)
	if err != nil {
		return domain.ConnectionDetail{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return domain.ConnectionDetail{}, domain.ErrUnauthorized
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return domain.ConnectionDetail{}, err
	}

	var apiResp domain.ApiResponse[domain.ConnectionDetail]
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return domain.ConnectionDetail{}, fmt.Errorf("failed to parse connection detail response: %w", err)
	}

	if !apiResp.Success {
		return domain.ConnectionDetail{}, fmt.Errorf("failed to get connection details: %s (%s)", apiResp.Error.Message, apiResp.Error.Code)
	}

	return apiResp.Data, nil
}

// CreateConnection registers a new remote server connection in the backend API.
func (h *HttpApiScanner) CreateConnection(token string, name, ip string, port int, username, password, details string) error {
	url := fmt.Sprintf("%s/passh/connections", h.baseURL)
	reqBody, err := json.Marshal(map[string]interface{}{
		"name":     name,
		"ip":       ip,
		"port":     port,
		"username": username,
		"password": password,
		"details":  details,
	})
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := h.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return domain.ErrUnauthorized
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var apiResp domain.ApiResponse[interface{}]
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return fmt.Errorf("failed to parse create connection response: %w", err)
	}

	if !apiResp.Success {
		return fmt.Errorf("failed to create connection: %s (%s)", apiResp.Error.Message, apiResp.Error.Code)
	}

	return nil
}

// UpdateConnection updates an existing connection in the backend API.
func (h *HttpApiScanner) UpdateConnection(token string, id uint, name, ip string, port int, username, password, details string) error {
	url := fmt.Sprintf("%s/passh/connections/%d", h.baseURL, id)
	reqBody, err := json.Marshal(map[string]interface{}{
		"name":     name,
		"ip":       ip,
		"port":     port,
		"username": username,
		"password": password,
		"details":  details,
	})
	if err != nil {
		return err
	}

	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := h.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return domain.ErrUnauthorized
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var apiResp domain.ApiResponse[interface{}]
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return fmt.Errorf("failed to parse update connection response: %w", err)
	}

	if !apiResp.Success {
		return fmt.Errorf("failed to update connection: %s (%s)", apiResp.Error.Message, apiResp.Error.Code)
	}

	return nil
}

// DeleteConnection deletes a connection from the backend API.
func (h *HttpApiScanner) DeleteConnection(token string, id uint) error {
	url := fmt.Sprintf("%s/passh/connections/%d", h.baseURL, id)
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := h.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return domain.ErrUnauthorized
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var apiResp domain.ApiResponse[interface{}]
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return fmt.Errorf("failed to parse delete connection response: %w", err)
	}

	if !apiResp.Success {
		return fmt.Errorf("failed to delete connection: %s (%s)", apiResp.Error.Message, apiResp.Error.Code)
	}

	return nil
}

// GetPasswords retrieves all password entries from backend API.
func (h *HttpApiScanner) GetPasswords(token string) ([]domain.Password, error) {
	url := fmt.Sprintf("%s/passh/passwords", h.baseURL)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := h.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, domain.ErrUnauthorized
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var apiResp domain.ApiResponse[[]domain.Password]
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse passwords response: %w", err)
	}

	if !apiResp.Success {
		return nil, fmt.Errorf("failed to list passwords: %s (%s)", apiResp.Error.Message, apiResp.Error.Code)
	}

	return apiResp.Data, nil
}

// GetPasswordDetail retrieves the detailed password entry including decrypted value.
func (h *HttpApiScanner) GetPasswordDetail(token string, id uint) (domain.PasswordDetail, error) {
	url := fmt.Sprintf("%s/passh/passwords/%d", h.baseURL, id)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return domain.PasswordDetail{}, err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := h.httpClient.Do(req)
	if err != nil {
		return domain.PasswordDetail{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return domain.PasswordDetail{}, domain.ErrUnauthorized
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return domain.PasswordDetail{}, err
	}

	var apiResp domain.ApiResponse[domain.PasswordDetail]
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return domain.PasswordDetail{}, fmt.Errorf("failed to parse password detail response: %w", err)
	}

	if !apiResp.Success {
		return domain.PasswordDetail{}, fmt.Errorf("failed to get password details: %s (%s)", apiResp.Error.Message, apiResp.Error.Code)
	}

	return apiResp.Data, nil
}

// CreatePassword registers a new password configuration in backend.
func (h *HttpApiScanner) CreatePassword(token string, name, user, password string) error {
	url := fmt.Sprintf("%s/passh/passwords", h.baseURL)
	reqBody, err := json.Marshal(map[string]string{
		"name":     name,
		"user":     user,
		"password": password,
	})
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := h.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return domain.ErrUnauthorized
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var apiResp domain.ApiResponse[interface{}]
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return fmt.Errorf("failed to parse create password response: %w", err)
	}

	if !apiResp.Success {
		return fmt.Errorf("failed to create password: %s (%s)", apiResp.Error.Message, apiResp.Error.Code)
	}

	return nil
}

// UpdatePassword updates fields of a password entry in backend.
func (h *HttpApiScanner) UpdatePassword(token string, id uint, name, user, password string) error {
	url := fmt.Sprintf("%s/passh/passwords/%d", h.baseURL, id)
	reqBody, err := json.Marshal(map[string]string{
		"name":     name,
		"user":     user,
		"password": password,
	})
	if err != nil {
		return err
	}

	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := h.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return domain.ErrUnauthorized
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var apiResp domain.ApiResponse[interface{}]
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return fmt.Errorf("failed to parse update password response: %w", err)
	}

	if !apiResp.Success {
		return fmt.Errorf("failed to update password: %s (%s)", apiResp.Error.Message, apiResp.Error.Code)
	}

	return nil
}

// DeletePassword deletes a password entry from backend.
func (h *HttpApiScanner) DeletePassword(token string, id uint) error {
	url := fmt.Sprintf("%s/passh/passwords/%d", h.baseURL, id)
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := h.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return domain.ErrUnauthorized
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var apiResp domain.ApiResponse[interface{}]
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return fmt.Errorf("failed to parse delete password response: %w", err)
	}

	if !apiResp.Success {
		return fmt.Errorf("failed to delete password: %s (%s)", apiResp.Error.Message, apiResp.Error.Code)
	}

	return nil
}

// Groups Module

// CreateGroup sends a group creation request to the API.
func (h *HttpApiScanner) CreateGroup(token string, name string) (domain.Group, error) {
	url := fmt.Sprintf("%s/passh/groups", h.baseURL)
	reqBody, err := json.Marshal(map[string]string{
		"name": name,
	})
	if err != nil {
		return domain.Group{}, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return domain.Group{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := h.httpClient.Do(req)
	if err != nil {
		return domain.Group{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return domain.Group{}, domain.ErrUnauthorized
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return domain.Group{}, err
	}

	var apiResp domain.ApiResponse[domain.Group]
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return domain.Group{}, fmt.Errorf("failed to parse create group response: %w", err)
	}

	if !apiResp.Success {
		return domain.Group{}, fmt.Errorf("failed to create group: %s (%s)", apiResp.Error.Message, apiResp.Error.Code)
	}

	return apiResp.Data, nil
}

// GetGroups lists all groups the authenticated user belongs to.
func (h *HttpApiScanner) GetGroups(token string) ([]domain.Group, error) {
	url := fmt.Sprintf("%s/passh/groups", h.baseURL)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := h.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return nil, domain.ErrUnauthorized
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var apiResp domain.ApiResponse[[]domain.Group]
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse groups response: %w", err)
	}

	if !apiResp.Success {
		return nil, fmt.Errorf("failed to list groups: %s (%s)", apiResp.Error.Message, apiResp.Error.Code)
	}

	return apiResp.Data, nil
}

// GetGroupDetail retrieves group details, members list, and shared assets.
func (h *HttpApiScanner) GetGroupDetail(token string, id uint) (domain.GroupDetailResponse, error) {
	url := fmt.Sprintf("%s/passh/groups/%d", h.baseURL, id)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return domain.GroupDetailResponse{}, err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := h.httpClient.Do(req)
	if err != nil {
		return domain.GroupDetailResponse{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return domain.GroupDetailResponse{}, domain.ErrUnauthorized
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return domain.GroupDetailResponse{}, err
	}

	var apiResp domain.ApiResponse[domain.GroupDetailResponse]
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return domain.GroupDetailResponse{}, fmt.Errorf("failed to parse group details response: %w", err)
	}

	if !apiResp.Success {
		return domain.GroupDetailResponse{}, fmt.Errorf("failed to get group details: %s (%s)", apiResp.Error.Message, apiResp.Error.Code)
	}

	return apiResp.Data, nil
}

// AddGroupMember registers a new member using their email.
func (h *HttpApiScanner) AddGroupMember(token string, groupID uint, email string) error {
	url := fmt.Sprintf("%s/passh/groups/%d/members", h.baseURL, groupID)
	reqBody, err := json.Marshal(map[string]string{
		"email": email,
	})
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := h.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return domain.ErrUnauthorized
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var apiResp domain.ApiResponse[interface{}]
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return fmt.Errorf("failed to parse add member response: %w", err)
	}

	if !apiResp.Success {
		return fmt.Errorf("failed to add member: %s (%s)", apiResp.Error.Message, apiResp.Error.Code)
	}

	return nil
}

// RemoveGroupMember removes a user membership from the group.
func (h *HttpApiScanner) RemoveGroupMember(token string, groupID uint, userID uint) error {
	url := fmt.Sprintf("%s/passh/groups/%d/members/%d", h.baseURL, groupID, userID)
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := h.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return domain.ErrUnauthorized
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var apiResp domain.ApiResponse[interface{}]
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return fmt.Errorf("failed to parse remove member response: %w", err)
	}

	if !apiResp.Success {
		return fmt.Errorf("failed to remove member: %s (%s)", apiResp.Error.Message, apiResp.Error.Code)
	}

	return nil
}

// ShareConnection shares an SSH connection with a group.
func (h *HttpApiScanner) ShareConnection(token string, groupID uint, connectionID uint) error {
	url := fmt.Sprintf("%s/passh/groups/%d/connections", h.baseURL, groupID)
	reqBody, err := json.Marshal(map[string]uint{
		"ssh_connection_id": connectionID,
	})
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := h.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return domain.ErrUnauthorized
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var apiResp domain.ApiResponse[interface{}]
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return fmt.Errorf("failed to parse share connection response: %w", err)
	}

	if !apiResp.Success {
		return fmt.Errorf("failed to share connection: %s (%s)", apiResp.Error.Message, apiResp.Error.Code)
	}

	return nil
}

// UnshareConnection stops sharing an SSH connection with a group.
func (h *HttpApiScanner) UnshareConnection(token string, groupID uint, connectionID uint) error {
	url := fmt.Sprintf("%s/passh/groups/%d/connections/%d", h.baseURL, groupID, connectionID)
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := h.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return domain.ErrUnauthorized
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var apiResp domain.ApiResponse[interface{}]
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return fmt.Errorf("failed to parse unshare connection response: %w", err)
	}

	if !apiResp.Success {
		return fmt.Errorf("failed to unshare connection: %s (%s)", apiResp.Error.Message, apiResp.Error.Code)
	}

	return nil
}

// SharePassword shares a password credential with a group.
func (h *HttpApiScanner) SharePassword(token string, groupID uint, passwordID uint) error {
	url := fmt.Sprintf("%s/passh/groups/%d/passwords", h.baseURL, groupID)
	reqBody, err := json.Marshal(map[string]uint{
		"password_id": passwordID,
	})
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBody))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := h.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return domain.ErrUnauthorized
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var apiResp domain.ApiResponse[interface{}]
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return fmt.Errorf("failed to parse share password response: %w", err)
	}

	if !apiResp.Success {
		return fmt.Errorf("failed to share password: %s (%s)", apiResp.Error.Message, apiResp.Error.Code)
	}

	return nil
}

// UnsharePassword stops sharing a password credential with a group.
func (h *HttpApiScanner) UnsharePassword(token string, groupID uint, passwordID uint) error {
	url := fmt.Sprintf("%s/passh/groups/%d/passwords/%d", h.baseURL, groupID, passwordID)
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := h.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return domain.ErrUnauthorized
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var apiResp domain.ApiResponse[interface{}]
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return fmt.Errorf("failed to parse unshare password response: %w", err)
	}

	if !apiResp.Success {
		return fmt.Errorf("failed to unshare password: %s (%s)", apiResp.Error.Message, apiResp.Error.Code)
	}

	return nil
}

// DeleteGroup deletes a group configuration in backend.
func (h *HttpApiScanner) DeleteGroup(token string, groupID uint) error {
	url := fmt.Sprintf("%s/passh/groups/%d", h.baseURL, groupID)
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := h.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return domain.ErrUnauthorized
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var apiResp domain.ApiResponse[interface{}]
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return fmt.Errorf("failed to parse delete group response: %w", err)
	}

	if !apiResp.Success {
		return fmt.Errorf("failed to delete group: %s (%s)", apiResp.Error.Message, apiResp.Error.Code)
	}

	return nil
}
