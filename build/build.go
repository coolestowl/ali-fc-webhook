package build

var (
	Version   = "Unknown"
	GoVersion = "Unknown"
	GitHash   = "Unknown"
	BuildTime = "Unknown"
	OSArch    = "Unknown"
	InfoMap   map[string]string
)

func init() {
	InfoMap = map[string]string{
		"version": Version,
		"go":      GoVersion,
		"os/arch": OSArch,
		"commit":  GitHash,
		"built":   BuildTime,
	}
}
