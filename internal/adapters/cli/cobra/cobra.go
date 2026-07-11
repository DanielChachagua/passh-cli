package cobra

import (
	"errors"
	"fmt"

	"ssh-cli/internal/ports"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

var (
	emailFlag    string
	passwordFlag string

	// Flags for connection / password creation / editing
	nameFlag    string
	ipFlag      string
	portFlag    int
	userFlag    string
	passFlag    string
	detailsFlag string

	// Flags for password resets
	tokenFlag       string
	newPassFlag     string
	confirmPassFlag string
)

// CLIAdapter orchestrates Cobra command registration.
type CLIAdapter struct {
	service ports.SshManagerService
	rootCmd *cobra.Command
}

// NewCLIAdapter registers the commands and returns a CLIAdapter instance.
func NewCLIAdapter(service ports.SshManagerService) *CLIAdapter {
	adapter := &CLIAdapter{
		service: service,
	}

	rootCmd := &cobra.Command{
		Use:   "passh",
		Short: "passh is a CLI tool to manage SSH connections and credentials securely.",
		Long:  `A modern command-line interface to authenticate, manage SSH endpoints, and securely save/retrieve passwords.`,
	}

	// ==========================================================
	// USER SUBCOMMANDS (savessh user ...)
	// ==========================================================
	userCmd := &cobra.Command{
		Use:   "user",
		Short: "Manage accounts, session state, and password resets",
	}

	userRegisterCmd := &cobra.Command{
		Use:   "register",
		Short: "Register a new user account with the central server (Admin token required)",
		RunE:  adapter.runUserRegister,
	}
	userRegisterCmd.Flags().StringVarP(&emailFlag, "email", "e", "", "Email of the new user")

	userLoginCmd := &cobra.Command{
		Use:   "login",
		Short: "Authenticate with the server and save session token",
		RunE:  adapter.runUserLogin,
	}
	userLoginCmd.Flags().StringVarP(&emailFlag, "email", "e", "", "Email of the user")
	userLoginCmd.Flags().StringVarP(&passwordFlag, "password", "p", "", "Password of the user")

	userLogoutCmd := &cobra.Command{
		Use:   "logout",
		Short: "Clear local session token and log out",
		RunE:  adapter.runUserLogout,
	}

	userResetRequestCmd := &cobra.Command{
		Use:   "reset-request",
		Short: "Request a password reset token to your email address",
		RunE:  adapter.runUserResetRequest,
	}
	userResetRequestCmd.Flags().StringVarP(&emailFlag, "email", "e", "", "Registered email address")

	userResetConfirmCmd := &cobra.Command{
		Use:   "reset-confirm",
		Short: "Confirm password reset using the token received in your email",
		RunE:  adapter.runUserResetConfirm,
	}
	userResetConfirmCmd.Flags().StringVar(&tokenFlag, "token", "", "Password reset token")
	userResetConfirmCmd.Flags().StringVar(&newPassFlag, "new-password", "", "New password")
	userResetConfirmCmd.Flags().StringVar(&confirmPassFlag, "confirm-password", "", "Confirm new password")

	userCmd.AddCommand(userRegisterCmd)
	userCmd.AddCommand(userLoginCmd)
	userCmd.AddCommand(userLogoutCmd)
	userCmd.AddCommand(userResetRequestCmd)
	userCmd.AddCommand(userResetConfirmCmd)
	rootCmd.AddCommand(userCmd)

	// ==========================================================
	// SSH SUBCOMMANDS (savessh ssh ...)
	// ==========================================================
	sshCmd := &cobra.Command{
		Use:   "ssh",
		Short: "Manage remote server SSH connections",
	}

	sshAddCmd := &cobra.Command{
		Use:   "add",
		Short: "Add a new server connection to the database",
		RunE:  adapter.runSshAdd,
	}
	sshAddCmd.Flags().StringVar(&nameFlag, "name", "", "Name of the connection")
	sshAddCmd.Flags().StringVar(&ipFlag, "ip", "", "IP address or hostname of the server")
	sshAddCmd.Flags().IntVar(&portFlag, "port", 0, "SSH port of the server")
	sshAddCmd.Flags().StringVar(&userFlag, "user", "", "SSH username")
	sshAddCmd.Flags().StringVar(&passFlag, "password", "", "SSH password")
	sshAddCmd.Flags().StringVar(&detailsFlag, "details", "", "Optional description details")

	sshListCmd := &cobra.Command{
		Use:   "list",
		Short: "List all server connections in the database",
		RunE:  adapter.runSshList,
	}

	sshEditCmd := &cobra.Command{
		Use:   "edit [connection_id]",
		Short: "Edit an existing server connection in the database",
		Args:  cobra.ExactArgs(1),
		RunE:  adapter.runSshEdit,
	}
	sshEditCmd.Flags().StringVar(&nameFlag, "name", "", "Name of the connection")
	sshEditCmd.Flags().StringVar(&ipFlag, "ip", "", "IP address or hostname of the server")
	sshEditCmd.Flags().IntVar(&portFlag, "port", 0, "SSH port of the server")
	sshEditCmd.Flags().StringVar(&userFlag, "user", "", "SSH username")
	sshEditCmd.Flags().StringVar(&passFlag, "password", "", "SSH password")
	sshEditCmd.Flags().StringVar(&detailsFlag, "details", "", "Optional description details")

	sshDeleteCmd := &cobra.Command{
		Use:   "delete [connection_id]",
		Short: "Delete a connection from the database",
		Args:  cobra.ExactArgs(1),
		RunE:  adapter.runSshDelete,
	}

	sshConnectCmd := &cobra.Command{
		Use:   "connect [id | search_term]",
		Short: "Select server and start interactive SSH session",
		Long:  `Establish an SSH session. You can pass a connection ID or name search term to connect directly or filter the list.`,
		Args:  cobra.MaximumNArgs(1),
		RunE:  adapter.runSshConnect,
	}

	sshCmd.AddCommand(sshAddCmd)
	sshCmd.AddCommand(sshListCmd)
	sshCmd.AddCommand(sshEditCmd)
	sshCmd.AddCommand(sshDeleteCmd)
	sshCmd.AddCommand(sshConnectCmd)
	rootCmd.AddCommand(sshCmd)

	// ==========================================================
	// PASSWORD SUBCOMMANDS (savessh pass ...)
	// ==========================================================
	passCmd := &cobra.Command{
		Use:   "pass",
		Short: "Manage stored passwords and credentials securely",
	}

	passAddCmd := &cobra.Command{
		Use:   "add",
		Short: "Add a new password entry to the secure storage",
		RunE:  adapter.runPassAdd,
	}
	passAddCmd.Flags().StringVar(&nameFlag, "name", "", "Name of the credential / service")
	passAddCmd.Flags().StringVar(&userFlag, "user", "", "Username for the service")
	passAddCmd.Flags().StringVar(&passFlag, "password", "", "Password value")

	passListCmd := &cobra.Command{
		Use:   "list",
		Short: "List all password configurations (passwords will be masked)",
		RunE:  adapter.runPassList,
	}

	passViewCmd := &cobra.Command{
		Use:   "view [password_id]",
		Short: "Retrieve and view decrypted password details",
		Args:  cobra.ExactArgs(1),
		RunE:  adapter.runPassView,
	}

	passEditCmd := &cobra.Command{
		Use:   "edit [password_id]",
		Short: "Edit an existing password entry",
		Args:  cobra.ExactArgs(1),
		RunE:  adapter.runPassEdit,
	}
	passEditCmd.Flags().StringVar(&nameFlag, "name", "", "Name of the service")
	passEditCmd.Flags().StringVar(&userFlag, "user", "", "Username for the service")
	passEditCmd.Flags().StringVar(&passFlag, "password", "", "New password value")

	passDeleteCmd := &cobra.Command{
		Use:   "delete [password_id]",
		Short: "Delete a password entry from secure database",
		Args:  cobra.ExactArgs(1),
		RunE:  adapter.runPassDelete,
	}

	passCmd.AddCommand(passAddCmd)
	passCmd.AddCommand(passListCmd)
	passCmd.AddCommand(passViewCmd)
	passCmd.AddCommand(passEditCmd)
	passCmd.AddCommand(passDeleteCmd)
	rootCmd.AddCommand(passCmd)

	// ==========================================================
	// GROUP SUBCOMMANDS (savessh group ...)
	// ==========================================================
	groupCmd := &cobra.Command{
		Use:   "group",
		Short: "Manage groups and credential sharing",
	}

	groupCreateCmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new group",
		RunE:  adapter.runGroupCreate,
	}
	groupCreateCmd.Flags().StringVar(&nameFlag, "name", "", "Name of the group")

	groupListCmd := &cobra.Command{
		Use:   "list",
		Short: "List all groups you belong to",
		RunE:  adapter.runGroupList,
	}

	groupViewCmd := &cobra.Command{
		Use:   "view [group_id]",
		Short: "View details of a group, including members and shared assets",
		Args:  cobra.ExactArgs(1),
		RunE:  adapter.runGroupView,
	}

	groupAddMemberCmd := &cobra.Command{
		Use:   "add-member [group_id] [email]",
		Short: "Add a member to a group",
		Args:  cobra.ExactArgs(2),
		RunE:  adapter.runGroupAddMember,
	}

	groupRemoveMemberCmd := &cobra.Command{
		Use:   "remove-member [group_id] [user_id]",
		Short: "Remove a member from a group",
		Args:  cobra.ExactArgs(2),
		RunE:  adapter.runGroupRemoveMember,
	}

	groupShareSshCmd := &cobra.Command{
		Use:   "share-ssh [group_id] [connection_id]",
		Short: "Share an SSH connection with a group",
		Args:  cobra.ExactArgs(2),
		RunE:  adapter.runGroupShareSsh,
	}

	groupUnshareSshCmd := &cobra.Command{
		Use:   "unshare-ssh [group_id] [connection_id]",
		Short: "Stop sharing an SSH connection with a group",
		Args:  cobra.ExactArgs(2),
		RunE:  adapter.runGroupUnshareSsh,
	}

	groupSharePassCmd := &cobra.Command{
		Use:   "share-pass [group_id] [password_id]",
		Short: "Share a secure password entry with a group",
		Args:  cobra.ExactArgs(2),
		RunE:  adapter.runGroupSharePass,
	}

	groupUnsharePassCmd := &cobra.Command{
		Use:   "unshare-pass [group_id] [password_id]",
		Short: "Stop sharing a secure password entry with a group",
		Args:  cobra.ExactArgs(2),
		RunE:  adapter.runGroupUnsharePass,
	}

	groupDeleteCmd := &cobra.Command{
		Use:   "delete [group_id]",
		Short: "Delete a group (Cascades deletion to memberships and shares)",
		Args:  cobra.ExactArgs(1),
		RunE:  adapter.runGroupDelete,
	}

	groupCmd.AddCommand(groupCreateCmd)
	groupCmd.AddCommand(groupListCmd)
	groupCmd.AddCommand(groupViewCmd)
	groupCmd.AddCommand(groupAddMemberCmd)
	groupCmd.AddCommand(groupRemoveMemberCmd)
	groupCmd.AddCommand(groupShareSshCmd)
	groupCmd.AddCommand(groupUnshareSshCmd)
	groupCmd.AddCommand(groupSharePassCmd)
	groupCmd.AddCommand(groupUnsharePassCmd)
	groupCmd.AddCommand(groupDeleteCmd)
	rootCmd.AddCommand(groupCmd)

	adapter.rootCmd = rootCmd
	return adapter
}

// Execute runs the Cobra command parser.
func (a *CLIAdapter) Execute() error {
	return a.rootCmd.Execute()
}

func (a *CLIAdapter) runUserRegister(cmd *cobra.Command, args []string) error {
	email := emailFlag

	if email == "" {
		prompt := promptui.Prompt{
			Label: "Email of the new user",
			Validate: func(input string) error {
				if len(input) == 0 {
					return errors.New("email cannot be empty")
				}
				return nil
			},
		}
		var err error
		email, err = prompt.Run()
		if err != nil {
			return err
		}
	}

	fmt.Println("Registering account...")
	err := a.service.Register(email)
	if err != nil {
		return fmt.Errorf("registration failed: %w", err)
	}

	fmt.Println("\033[1;32mUser registered successfully! A welcome email was sent with their temporary credentials.\033[0m")
	return nil
}

func (a *CLIAdapter) runUserLogin(cmd *cobra.Command, args []string) error {
	email := emailFlag
	password := passwordFlag

	if email == "" {
		prompt := promptui.Prompt{
			Label: "Email",
			Validate: func(input string) error {
				if len(input) == 0 {
					return errors.New("email cannot be empty")
				}
				return nil
			},
		}
		var err error
		email, err = prompt.Run()
		if err != nil {
			return err
		}
	}

	if password == "" {
		prompt := promptui.Prompt{
			Label: "Password",
			Mask:  '*',
			Validate: func(input string) error {
				if len(input) == 0 {
					return errors.New("password cannot be empty")
				}
				return nil
			},
		}
		var err error
		password, err = prompt.Run()
		if err != nil {
			return err
		}
	}

	fmt.Println("Authenticating...")
	err := a.service.Login(email, password)
	if err != nil {
		return fmt.Errorf("login failed: %w", err)
	}

	fmt.Println("\033[1;32mSuccessfully authenticated and saved token!\033[0m")
	return nil
}

func (a *CLIAdapter) runUserLogout(cmd *cobra.Command, args []string) error {
	err := a.service.Logout()
	if err != nil {
		return fmt.Errorf("logout failed: %w", err)
	}
	fmt.Println("\033[1;32mLogged out successfully (session cleared).\033[0m")
	return nil
}

func (a *CLIAdapter) runUserResetRequest(cmd *cobra.Command, args []string) error {
	email := emailFlag

	if email == "" {
		prompt := promptui.Prompt{
			Label: "Enter your registered Email",
			Validate: func(input string) error {
				if len(input) == 0 {
					return errors.New("email cannot be empty")
				}
				return nil
			},
		}
		var err error
		email, err = prompt.Run()
		if err != nil {
			return err
		}
	}

	fmt.Println("Requesting password reset...")
	err := a.service.ResetPasswordRequest(email)
	if err != nil {
		return err
	}

	fmt.Println("\033[1;32mReset token sent! Check your email inbox.\033[0m")
	return nil
}

func (a *CLIAdapter) runUserResetConfirm(cmd *cobra.Command, args []string) error {
	token := tokenFlag
	newPass := newPassFlag
	confirmPass := confirmPassFlag

	if token == "" {
		prompt := promptui.Prompt{
			Label: "Enter the reset token",
			Validate: func(input string) error {
				if len(input) == 0 {
					return errors.New("token cannot be empty")
				}
				return nil
			},
		}
		var err error
		token, err = prompt.Run()
		if err != nil {
			return err
		}
	}

	if newPass == "" {
		prompt := promptui.Prompt{
			Label: "Enter new password",
			Mask:  '*',
			Validate: func(input string) error {
				if len(input) == 0 {
					return errors.New("password cannot be empty")
				}
				return nil
			},
		}
		var err error
		newPass, err = prompt.Run()
		if err != nil {
			return err
		}
	}

	if confirmPass == "" {
		prompt := promptui.Prompt{
			Label: "Confirm new password",
			Mask:  '*',
			Validate: func(input string) error {
				if len(input) == 0 {
					return errors.New("confirm password cannot be empty")
				}
				if input != newPass {
					return errors.New("passwords must match")
				}
				return nil
			},
		}
		var err error
		confirmPass, err = prompt.Run()
		if err != nil {
			return err
		}
	}

	if newPass != confirmPass {
		return errors.New("new password and confirmation password must match")
	}

	fmt.Println("Updating password...")
	err := a.service.ConfirmPasswordReset(token, newPass, confirmPass)
	if err != nil {
		return err
	}

	fmt.Println("\033[1;32mPassword successfully updated! You can now login with your new credentials.\033[0m")
	return nil
}

// SSH command runners

func (a *CLIAdapter) runSshAdd(cmd *cobra.Command, args []string) error {
	name := nameFlag
	ip := ipFlag
	port := portFlag
	user := userFlag
	password := passFlag
	details := detailsFlag

	if name == "" {
		prompt := promptui.Prompt{
			Label: "Connection Name (e.g. MyServer)",
			Validate: func(input string) error {
				if len(input) == 0 {
					return errors.New("name cannot be empty")
				}
				return nil
			},
		}
		var err error
		name, err = prompt.Run()
		if err != nil {
			return err
		}
	}

	if ip == "" {
		prompt := promptui.Prompt{
			Label: "IP Address / Hostname",
			Validate: func(input string) error {
				if len(input) == 0 {
					return errors.New("IP or hostname cannot be empty")
				}
				return nil
			},
		}
		var err error
		ip, err = prompt.Run()
		if err != nil {
			return err
		}
	}

	if port == 0 {
		prompt := promptui.Prompt{
			Label:   "Port",
			Default: "22",
			Validate: func(input string) error {
				if len(input) == 0 {
					return errors.New("port cannot be empty")
				}
				return nil
			},
		}
		res, err := prompt.Run()
		if err != nil {
			return err
		}
		var p int
		_, err = fmt.Sscanf(res, "%d", &p)
		if err != nil {
			return errors.New("invalid port number")
		}
		port = p
	}

	if user == "" {
		prompt := promptui.Prompt{
			Label:   "Username",
			Default: "root",
			Validate: func(input string) error {
				if len(input) == 0 {
					return errors.New("username cannot be empty")
				}
				return nil
			},
		}
		var err error
		user, err = prompt.Run()
		if err != nil {
			return err
		}
	}

	if password == "" {
		prompt := promptui.Prompt{
			Label: "SSH Password",
			Mask:  '*',
			Validate: func(input string) error {
				if len(input) == 0 {
					return errors.New("password cannot be empty")
				}
				return nil
			},
		}
		var err error
		password, err = prompt.Run()
		if err != nil {
			return err
		}
	}

	if details == "" {
		prompt := promptui.Prompt{
			Label: "Description details (optional)",
		}
		var err error
		details, err = prompt.Run()
		if err != nil {
			return err
		}
	}

	fmt.Println("Adding connection configuration to database...")
	err := a.service.CreateConnection(name, ip, port, user, password, details)
	if err != nil {
		return fmt.Errorf("failed to save connection: %w", err)
	}

	fmt.Println("\033[1;32mConnection created successfully!\033[0m")
	return nil
}

func (a *CLIAdapter) runSshList(cmd *cobra.Command, args []string) error {
	conns, err := a.service.ListConnections()
	if err != nil {
		return err
	}

	if len(conns) == 0 {
		fmt.Println("No connections found.")
		return nil
	}

	fmt.Println("\n\033[1;36m=== SSH Server Connections ===\033[0m")
	for _, c := range conns {
		fmt.Printf("ID: %-4d | Name: %-25s | Host: %s@%s:%d\n  Details: %s\n", c.ID, c.Name, c.Username, c.IP, c.Port, c.Details)
	}
	fmt.Println("")
	return nil
}

func (a *CLIAdapter) runSshEdit(cmd *cobra.Command, args []string) error {
	var id uint
	_, err := fmt.Sscanf(args[0], "%d", &id)
	if err != nil {
		return errors.New("invalid connection ID, must be a number")
	}

	fmt.Println("Retrieving current connection details...")
	detail, err := a.service.GetConnectionDetail(id)
	if err != nil {
		return err
	}

	name := nameFlag
	ip := ipFlag
	port := portFlag
	user := userFlag
	password := passFlag
	details := detailsFlag

	if name == "" {
		prompt := promptui.Prompt{
			Label:   "Connection Name",
			Default: detail.Name,
		}
		name, err = prompt.Run()
		if err != nil {
			return err
		}
	}

	if ip == "" {
		prompt := promptui.Prompt{
			Label:   "IP Address / Hostname",
			Default: detail.IP,
		}
		ip, err = prompt.Run()
		if err != nil {
			return err
		}
	}

	if port == 0 {
		prompt := promptui.Prompt{
			Label:   "Port",
			Default: fmt.Sprintf("%d", detail.Port),
		}
		res, err := prompt.Run()
		if err != nil {
			return err
		}
		var p int
		_, err = fmt.Sscanf(res, "%d", &p)
		if err != nil {
			return errors.New("invalid port number")
		}
		port = p
	}

	if user == "" {
		prompt := promptui.Prompt{
			Label:   "Username",
			Default: detail.Username,
		}
		user, err = prompt.Run()
		if err != nil {
			return err
		}
	}

	if password == "" {
		prompt := promptui.Prompt{
			Label: "SSH Password (leave empty to keep current)",
			Mask:  '*',
		}
		password, err = prompt.Run()
		if err != nil {
			return err
		}
		if password == "" {
			password = detail.Password
		}
	}

	if details == "" {
		prompt := promptui.Prompt{
			Label:   "Description details",
			Default: detail.Details,
		}
		details, err = prompt.Run()
		if err != nil {
			return err
		}
	}

	fmt.Println("Updating connection configuration...")
	err = a.service.UpdateConnection(id, name, ip, port, user, password, details)
	if err != nil {
		return err
	}

	fmt.Println("\033[1;32mConnection updated successfully!\033[0m")
	return nil
}

func (a *CLIAdapter) runSshDelete(cmd *cobra.Command, args []string) error {
	var id uint
	_, err := fmt.Sscanf(args[0], "%d", &id)
	if err != nil {
		return errors.New("invalid connection ID, must be a number")
	}

	prompt := promptui.Prompt{
		Label:     fmt.Sprintf("Are you sure you want to delete connection ID %d", id),
		IsConfirm: true,
	}
	_, err = prompt.Run()
	if err != nil {
		fmt.Println("Deletion aborted.")
		return nil
	}

	fmt.Printf("Deleting connection %d...\n", id)
	err = a.service.DeleteConnection(id)
	if err != nil {
		return err
	}

	fmt.Println("\033[1;32mConnection deleted successfully!\033[0m")
	return nil
}

func (a *CLIAdapter) runSshConnect(cmd *cobra.Command, args []string) error {
	searchTerm := ""
	if len(args) > 0 {
		searchTerm = args[0]
	}
	err := a.service.ConnectInteractive(searchTerm)
	if err != nil {
		return err
	}
	return nil
}

// PASS (Passwords) command runners

func (a *CLIAdapter) runPassAdd(cmd *cobra.Command, args []string) error {
	name := nameFlag
	user := userFlag
	password := passFlag

	if name == "" {
		prompt := promptui.Prompt{
			Label: "Service/Site Name (e.g. GitHub)",
			Validate: func(input string) error {
				if len(input) == 0 {
					return errors.New("name cannot be empty")
				}
				return nil
			},
		}
		var err error
		name, err = prompt.Run()
		if err != nil {
			return err
		}
	}

	if user == "" {
		prompt := promptui.Prompt{
			Label: "Username / Email",
			Validate: func(input string) error {
				if len(input) == 0 {
					return errors.New("username cannot be empty")
				}
				return nil
			},
		}
		var err error
		user, err = prompt.Run()
		if err != nil {
			return err
		}
	}

	if password == "" {
		prompt := promptui.Prompt{
			Label: "Password value",
			Mask:  '*',
			Validate: func(input string) error {
				if len(input) == 0 {
					return errors.New("password cannot be empty")
				}
				return nil
			},
		}
		var err error
		password, err = prompt.Run()
		if err != nil {
			return err
		}
	}

	fmt.Println("Storing password securely in database...")
	err := a.service.CreatePassword(name, user, password)
	if err != nil {
		return fmt.Errorf("failed to save password: %w", err)
	}

	fmt.Println("\033[1;32mPassword saved securely!\033[0m")
	return nil
}

func (a *CLIAdapter) runPassList(cmd *cobra.Command, args []string) error {
	passwords, err := a.service.ListPasswords()
	if err != nil {
		return err
	}

	if len(passwords) == 0 {
		fmt.Println("No stored passwords found.")
		return nil
	}

	fmt.Println("\n\033[1;35m=== Secure Stored Credentials ===\033[0m")
	for _, p := range passwords {
		fmt.Printf("ID: %-4d | Service/Site: %-25s | User: %s\n", p.ID, p.Name, p.User)
	}
	fmt.Println("")
	return nil
}

func (a *CLIAdapter) runPassView(cmd *cobra.Command, args []string) error {
	var id uint
	_, err := fmt.Sscanf(args[0], "%d", &id)
	if err != nil {
		return errors.New("invalid credential ID, must be a number")
	}

	fmt.Println("Retrieving credential details...")
	detail, err := a.service.GetPasswordDetail(id)
	if err != nil {
		return err
	}

	fmt.Println("\n\033[1;35m=== Decrypted Credential Details ===\033[0m")
	fmt.Printf("ID:       %d\n", detail.ID)
	fmt.Printf("Service:  %s\n", detail.Name)
	fmt.Printf("Username: %s\n", detail.User)
	fmt.Printf("Password: \033[1;32m%s\033[0m\n\n", detail.Password)
	return nil
}

func (a *CLIAdapter) runPassEdit(cmd *cobra.Command, args []string) error {
	var id uint
	_, err := fmt.Sscanf(args[0], "%d", &id)
	if err != nil {
		return errors.New("invalid credential ID, must be a number")
	}

	fmt.Println("Retrieving current credential details...")
	detail, err := a.service.GetPasswordDetail(id)
	if err != nil {
		return err
	}

	name := nameFlag
	user := userFlag
	password := passFlag

	if name == "" {
		prompt := promptui.Prompt{
			Label:   "Service/Site Name",
			Default: detail.Name,
		}
		name, err = prompt.Run()
		if err != nil {
			return err
		}
	}

	if user == "" {
		prompt := promptui.Prompt{
			Label:   "Username / Email",
			Default: detail.User,
		}
		user, err = prompt.Run()
		if err != nil {
			return err
		}
	}

	if password == "" {
		prompt := promptui.Prompt{
			Label: "Password (leave empty to keep current)",
			Mask:  '*',
		}
		password, err = prompt.Run()
		if err != nil {
			return err
		}
		if password == "" {
			password = detail.Password
		}
	}

	fmt.Println("Updating password record...")
	err = a.service.UpdatePassword(id, name, user, password)
	if err != nil {
		return err
	}

	fmt.Println("\033[1;32mPassword updated successfully!\033[0m")
	return nil
}

func (a *CLIAdapter) runPassDelete(cmd *cobra.Command, args []string) error {
	var id uint
	_, err := fmt.Sscanf(args[0], "%d", &id)
	if err != nil {
		return errors.New("invalid credential ID, must be a number")
	}

	prompt := promptui.Prompt{
		Label:     fmt.Sprintf("Are you sure you want to delete credential ID %d", id),
		IsConfirm: true,
	}
	_, err = prompt.Run()
	if err != nil {
		fmt.Println("Deletion aborted.")
		return nil
	}

	fmt.Printf("Deleting credential %d...\n", id)
	err = a.service.DeletePassword(id)
	if err != nil {
		return err
	}

	fmt.Println("\033[1;32mCredential deleted successfully!\033[0m")
	return nil
}

// GROUP command runners

func (a *CLIAdapter) runGroupCreate(cmd *cobra.Command, args []string) error {
	name := nameFlag

	if name == "" {
		prompt := promptui.Prompt{
			Label: "Group Name (e.g. work)",
			Validate: func(input string) error {
				if len(input) == 0 {
					return errors.New("group name cannot be empty")
				}
				return nil
			},
		}
		var err error
		name, err = prompt.Run()
		if err != nil {
			return err
		}
	}

	fmt.Printf("Creating group '%s'...\n", name)
	group, err := a.service.CreateGroup(name)
	if err != nil {
		return err
	}

	fmt.Printf("\033[1;32mGroup created successfully! ID: %d, Name: %s\033[0m\n", group.ID, group.Name)
	return nil
}

func (a *CLIAdapter) runGroupList(cmd *cobra.Command, args []string) error {
	groups, err := a.service.ListGroups()
	if err != nil {
		return err
	}

	if len(groups) == 0 {
		fmt.Println("You do not belong to any groups.")
		return nil
	}

	fmt.Println("\n\033[1;34m=== My Groups ===\033[0m")
	for _, g := range groups {
		fmt.Printf("ID: %-4d | Name: %-25s | Owner ID: %d\n", g.ID, g.Name, g.OwnerID)
	}
	fmt.Println("")
	return nil
}

func (a *CLIAdapter) runGroupView(cmd *cobra.Command, args []string) error {
	var id uint
	_, err := fmt.Sscanf(args[0], "%d", &id)
	if err != nil {
		return errors.New("invalid group ID, must be a number")
	}

	fmt.Println("Retrieving group details...")
	detail, err := a.service.GetGroupDetail(id)
	if err != nil {
		return err
	}

	fmt.Printf("\n\033[1;34m=== Group: %s (ID: %d, Owner: %d) ===\033[0m\n", detail.Group.Name, detail.Group.ID, detail.Group.OwnerID)

	// Members
	fmt.Println("\n\033[1;33m--- Members ---\033[0m")
	if len(detail.Members) == 0 {
		fmt.Println("No members.")
	} else {
		for _, m := range detail.Members {
			fmt.Printf("  User ID: %-4d | Email: %-30s | Role: %s\n", m.ID, m.Email, m.Role)
		}
	}

	// Connections
	fmt.Println("\n\033[1;36m--- Shared SSH Connections ---\033[0m")
	if len(detail.Connections) == 0 {
		fmt.Println("No shared SSH connections.")
	} else {
		for _, c := range detail.Connections {
			fmt.Printf("  ID: %-4d | Name: %-20s | Host: %s@%s:%d\n", c.ID, c.Name, c.Username, c.IP, c.Port)
		}
	}

	// Passwords
	fmt.Println("\n\033[1;35m--- Shared Passwords ---\033[0m")
	if len(detail.Passwords) == 0 {
		fmt.Println("No shared passwords.")
	} else {
		for _, p := range detail.Passwords {
			fmt.Printf("  ID: %-4d | Service: %-20s | User: %s\n", p.ID, p.Name, p.User)
		}
	}
	fmt.Println("")
	return nil
}

func (a *CLIAdapter) runGroupAddMember(cmd *cobra.Command, args []string) error {
	var groupID uint
	_, err := fmt.Sscanf(args[0], "%d", &groupID)
	if err != nil {
		return errors.New("invalid group ID, must be a number")
	}

	email := args[1]

	fmt.Printf("Adding member %s to group %d...\n", email, groupID)
	err = a.service.AddGroupMember(groupID, email)
	if err != nil {
		return err
	}

	fmt.Println("\033[1;32mMember added successfully to group!\033[0m")
	return nil
}

func (a *CLIAdapter) runGroupRemoveMember(cmd *cobra.Command, args []string) error {
	var groupID, userID uint
	_, err := fmt.Sscanf(args[0], "%d", &groupID)
	if err != nil {
		return errors.New("invalid group ID, must be a number")
	}
	_, err = fmt.Sscanf(args[1], "%d", &userID)
	if err != nil {
		return errors.New("invalid user ID, must be a number")
	}

	fmt.Printf("Removing user %d from group %d...\n", userID, groupID)
	err = a.service.RemoveGroupMember(groupID, userID)
	if err != nil {
		return err
	}

	fmt.Println("\033[1;32mMember removed successfully from group!\033[0m")
	return nil
}

func (a *CLIAdapter) runGroupShareSsh(cmd *cobra.Command, args []string) error {
	var groupID, connectionID uint
	_, err := fmt.Sscanf(args[0], "%d", &groupID)
	if err != nil {
		return errors.New("invalid group ID, must be a number")
	}
	_, err = fmt.Sscanf(args[1], "%d", &connectionID)
	if err != nil {
		return errors.New("invalid SSH connection ID, must be a number")
	}

	fmt.Printf("Sharing SSH connection %d with group %d...\n", connectionID, groupID)
	err = a.service.ShareConnection(groupID, connectionID)
	if err != nil {
		return err
	}

	fmt.Println("\033[1;32mSSH connection shared successfully!\033[0m")
	return nil
}

func (a *CLIAdapter) runGroupUnshareSsh(cmd *cobra.Command, args []string) error {
	var groupID, connectionID uint
	_, err := fmt.Sscanf(args[0], "%d", &groupID)
	if err != nil {
		return errors.New("invalid group ID, must be a number")
	}
	_, err = fmt.Sscanf(args[1], "%d", &connectionID)
	if err != nil {
		return errors.New("invalid SSH connection ID, must be a number")
	}

	fmt.Printf("Unsharing SSH connection %d from group %d...\n", connectionID, groupID)
	err = a.service.UnshareConnection(groupID, connectionID)
	if err != nil {
		return err
	}

	fmt.Println("\033[1;32mSSH connection unshared successfully!\033[0m")
	return nil
}

func (a *CLIAdapter) runGroupSharePass(cmd *cobra.Command, args []string) error {
	var groupID, passwordID uint
	_, err := fmt.Sscanf(args[0], "%d", &groupID)
	if err != nil {
		return errors.New("invalid group ID, must be a number")
	}
	_, err = fmt.Sscanf(args[1], "%d", &passwordID)
	if err != nil {
		return errors.New("invalid password ID, must be a number")
	}

	fmt.Printf("Sharing password %d with group %d...\n", passwordID, groupID)
	err = a.service.SharePassword(groupID, passwordID)
	if err != nil {
		return err
	}

	fmt.Println("\033[1;32mPassword shared successfully!\033[0m")
	return nil
}

func (a *CLIAdapter) runGroupUnsharePass(cmd *cobra.Command, args []string) error {
	var groupID, passwordID uint
	_, err := fmt.Sscanf(args[0], "%d", &groupID)
	if err != nil {
		return errors.New("invalid group ID, must be a number")
	}
	_, err = fmt.Sscanf(args[1], "%d", &passwordID)
	if err != nil {
		return errors.New("invalid password ID, must be a number")
	}

	fmt.Printf("Unsharing password %d from group %d...\n", passwordID, groupID)
	err = a.service.UnsharePassword(groupID, passwordID)
	if err != nil {
		return err
	}

	fmt.Println("\033[1;32mPassword unshared successfully!\033[0m")
	return nil
}

func (a *CLIAdapter) runGroupDelete(cmd *cobra.Command, args []string) error {
	var groupID uint
	_, err := fmt.Sscanf(args[0], "%d", &groupID)
	if err != nil {
		return errors.New("invalid group ID, must be a number")
	}

	prompt := promptui.Prompt{
		Label:     fmt.Sprintf("Are you sure you want to delete group ID %d (will cascade delete memberships and shares)", groupID),
		IsConfirm: true,
	}
	_, err = prompt.Run()
	if err != nil {
		fmt.Println("Deletion aborted.")
		return nil
	}

	fmt.Printf("Deleting group %d...\n", groupID)
	err = a.service.DeleteGroup(groupID)
	if err != nil {
		return err
	}

	fmt.Println("\033[1;32mGroup deleted successfully!\033[0m")
	return nil
}
