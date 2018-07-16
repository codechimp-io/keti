package version

import (
	"bytes"
	"fmt"
	"runtime"
	"strings"
	"text/template"

	"github.com/prometheus/client_golang/prometheus"
)

// Build information. Populated at build-time.
var (
	Version      string
	Revision     string
	Branch       string
	BuildDate    string
	GoVersion    = runtime.Version()
	OS           = runtime.GOOS
	Architecture = runtime.GOARCH
	tpl          *template.Template
)

// versionInfoTmpl contains the template used by Info.
var versionInfoTmpl = `
{{.program}}, version {{.version}} (branch: {{.branch}}, revision: {{.revision}})
  build date:       {{.buildDate}}
  go version:       {{.go_version}}
  OS/Arch:          {{.os}}/{{.architecture}}
`

func init() {
	tpl = template.Must(template.New("version").Parse(versionInfoTmpl))
}

// Print returns version information.
func Print(name string) string {
	m := map[string]string{
		"program":      name,
		"version":      Version,
		"revision":     Revision,
		"branch":       Branch,
		"buildDate":    BuildDate,
		"go_version":   GoVersion,
		"os":           OS,
		"architecture": Architecture,
	}

	var buf bytes.Buffer
	if err := tpl.ExecuteTemplate(&buf, "version", m); err != nil {
		panic(err)
	}
	return strings.TrimSpace(buf.String())
}

// Info returns version, branch and revision information.
func Info(name string) string {
	return fmt.Sprintf("%s (version: %s, branch: %s, revision: %s)", name, Version, Branch, Revision)
}

// BuildContext returns goVersion and buildDate information.
func BuildContext() string {
	return fmt.Sprintf("(Go: %s, Date: %s)", GoVersion, BuildDate)
}

// NewMetricsCollector returns a prometheus.Collector which represents current build information.
func NewMetricsCollector(name string) *prometheus.GaugeVec {
	labels := map[string]string{
		"program":      name,
		"version":      Version,
		"revision":     Revision,
		"branch":       Branch,
		"buildDate":    BuildDate,
		"go_version":   GoVersion,
		"os":           OS,
		"architecture": Architecture,
	}

	labelNames := make([]string, 0, len(labels))
	for n := range labels {
		labelNames = append(labelNames, n)
	}

	buildInfo := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "version_info",
			Help: "A metric with a constant '1' value labeled by different build stats fields.",
		},
		labelNames,
	)
	buildInfo.With(labels).Set(1)

	return buildInfo
}
