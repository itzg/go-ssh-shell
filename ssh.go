package shell

import (
	"fmt"
	"github.com/gliderlabs/ssh"
	"io"
	"log"
)

// SshServer manages acceptance of and authenticating SSH connections and delegating input to
// a Handler for each session instantiated by the given HandlerFactory.
// The zero value of this struct will use reasonable defaults, but won't be very useful since
// the default handler just outputs remote inputs to logs.
type SshServer struct {
	*Config
	HandlerFactory
}

type HandlerFactory func(s *Shell) Handler

type Handler interface {
	// HandleLine is called with the next line that was consumed by the SSH shell. Typically this
	// is due the user typing a command string.
	// If an error is returned, then the error is reported back to the SSH client and the SSH
	// session is closed.
	HandleLine(line string) error

	// HandleEof is called when the user types Control-D
	// If an error is returned, then the error is reported back to the SSH client and the SSH
	// session is closed.
	HandleEof() error
}

type defaultHandler struct {
	s *Shell
}

func (h *defaultHandler) HandleLine(line string) error {
	log.Printf("LINE from %s: %s", h.s.InstanceName(), line)
	return nil
}

func (h *defaultHandler) HandleEof() error {
	log.Printf("EOF from %s", h.s.InstanceName())
	return nil
}

func defaultHandlerFactory(s *Shell) Handler {
	return &defaultHandler{s: s}
}

// ListenAndServe will block listening for new SSH connections and serving those with a new
// instance of Shell and a Handler each.
func (s *SshServer) ListenAndServe() error {
	config := s.Config
	if config == nil {
		config = &Config{}
	}

	handlerFactory := s.HandlerFactory
	if handlerFactory == nil {
		handlerFactory = defaultHandlerFactory
	}

	auth := NewAuth(config)

	hostKeyResolver := NewHostKeyResolver(config)

	bind := useOrDefaultString(config.Bind, DefaultBind)

	log.Printf("Accepting SSH connections at %s", bind)
	return ssh.ListenAndServe(bind,
		func(session ssh.Session) {
			instanceName := fmt.Sprintf("%s@%s", session.User(), session.RemoteAddr())
			log.Printf("I: New session for user=%s from=%s\n", session.User(), session.RemoteAddr())
			shell := NewShell(session, instanceName, config)
			shell.SetPrompt("> ")

			handler := handlerFactory(shell)

			for {
				line, err := shell.Read()

				if err != nil {
					if err == io.EOF {
						err = handler.HandleEof()
						if err == io.EOF {
							return
						} else if err != nil {
							endSessionWithError(session, shell, err)
							return
						}
					} else {
						endSessionWithError(session, shell, err)
						return
					}
				}

				err = handler.HandleLine(line)
				if err != nil {
					if err == io.EOF {
						return
					} else {
						endSessionWithError(session, shell, err)
					}
				}
			}
		},
		ssh.PasswordAuth(auth.PasswordHandler),
		hostKeyResolver.ResolveOption(),
	)
}

func endSessionWithError(s ssh.Session, shell *Shell, err error) {
	_ = shell.OutputLine("")
	_ = shell.OutputLine(err.Error())
	_ = s.Exit(1)
}
