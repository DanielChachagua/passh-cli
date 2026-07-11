package crypto

import (
	"fmt"
	"os"
	"ssh-cli/internal/domain"
	"time"

	"golang.org/x/crypto/ssh"
	"golang.org/x/term"
)

// SshCryptoDialer implements ports.SshDialer using native Go crypto/ssh.
type SshCryptoDialer struct{}

// NewSshCryptoDialer instantiates a new SshCryptoDialer.
func NewSshCryptoDialer() *SshCryptoDialer {
	return &SshCryptoDialer{}
}

// Connect establishes the SSH tunnel, sets the local terminal to raw mode, and binds streams.
func (d *SshCryptoDialer) Connect(config domain.SshSessionConfig) error {
	sshConfig := &ssh.ClientConfig{
		User: config.User,
		Auth: []ssh.AuthMethod{
			ssh.Password(config.Password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         15 * time.Second,
	}

	addr := fmt.Sprintf("%s:%d", config.IP, config.Port)
	fmt.Printf("Connecting to %s@%s...\n", config.User, addr)

	client, err := ssh.Dial("tcp", addr, sshConfig)
	if err != nil {
		return fmt.Errorf("failed to dial: %w", err)
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	// Link input/output/error streams directly to OS standard files
	session.Stdout = os.Stdout
	session.Stdin = os.Stdin
	session.Stderr = os.Stderr

	fd := int(os.Stdin.Fd())
	if term.IsTerminal(fd) {
		// Put local terminal in raw mode so keypresses are sent instantly (arrows, tab, backspace, ctrl+c)
		oldState, err := term.MakeRaw(fd)
		if err != nil {
			return fmt.Errorf("failed to put terminal in raw mode: %w", err)
		}
		defer term.Restore(fd, oldState)

		// Fetch terminal size dynamically
		w, h, err := term.GetSize(fd)
		if err != nil {
			w, h = 80, 40
		}

		termEnv := os.Getenv("TERM")
		if termEnv == "" {
			termEnv = "xterm-256color"
		}

		modes := ssh.TerminalModes{
			ssh.ECHO:          1,     // Enable echo
			ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
			ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
		}

		if err := session.RequestPty(termEnv, h, w, modes); err != nil {
			return fmt.Errorf("failed to request pseudo terminal: %w", err)
		}
	}

	// Launch remote shell session
	if err := session.Shell(); err != nil {
		return fmt.Errorf("failed to start remote shell: %w", err)
	}

	// Wait for remote command/shell to finish executing
	if err := session.Wait(); err != nil {
		// Avoid returning error if it's a standard exit status
		if _, ok := err.(*ssh.ExitError); ok {
			return nil
		}
		return err
	}

	return nil
}
