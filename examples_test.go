package shell_test

import (
	shell "github.com/itzg/go-ssh-shell"
	"log"
)

type exampleHandler struct {
	s shell.Shell
}

func exampleHandlerFactory(s *shell.Shell) shell.Handler {
	return &exampleHandler{}
}

func (h *exampleHandler) HandleLine(line string) error {
	log.Printf("LINE from %s: %s", h.s.InstanceName(), line)
	return nil
}

func (h *exampleHandler) HandleEof() error {
	log.Printf("EOF from %s", h.s.InstanceName())
	return nil
}

func ExampleSshServer_ListenAndServe() {
	sshServer := &shell.SshServer{
		Config: &shell.Config{
			Bind: ":2222",
			Users: map[string]shell.User{
				"user": {Password: "notsecure"},
			},
		},
		HandlerFactory: exampleHandlerFactory,
	}

	log.Fatal(sshServer.ListenAndServe())
}
