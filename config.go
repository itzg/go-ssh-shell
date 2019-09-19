package shell

type User struct {
	Password string
}

type Config struct {
	HistorySize int
	HostKeyFile string `mapstructure:"host-key-file"`
	Users       map[string]User
	Bind        string
}

const (
	DefaultConfigHistorySize = 100
	DefaultBind = ":22"
)

func useOrDefaultInt(value int, defaultValue int) int {
	if value == 0 {
		return defaultValue
	} else {
		return value
	}
}

func useOrDefaultString(value string, defaultValue string) string {
	if value == "" {
		return defaultValue
	} else {
		return value
	}
}