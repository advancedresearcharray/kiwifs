package docexport

import (
	"fmt"
	"log"
	"os/exec"
	"strings"
	"sync"
)

// DepStatus describes whether an external tool is available.
type DepStatus struct {
	Name      string // human-readable name
	Binary    string // executable name
	Available bool
	Version   string // output of --version if available
	Required  bool   // true if the tool is required for a registered format
}

// CheckDeps probes for all external tools used by the document export
// pipeline and returns their availability. This is called once at startup
// so the server can log clear messages about missing tools.
func CheckDeps() []DepStatus {
	deps := []struct {
		name      string
		binary    string
		versionFl string // flag to get version
		required  bool
	}{
		{"Pandoc", "pandoc", "--version", true},
		{"Marp CLI", "marp", "--version", false},
		{"MkDocs", "mkdocs", "--version", false},
		{"Typst", "typst", "--version", false},
		{"XeLaTeX", "xelatex", "--version", false},
		{"pandoc-crossref", "pandoc-crossref", "--version", false},
	}

	results := make([]DepStatus, len(deps))
	var wg sync.WaitGroup
	for i, d := range deps {
		wg.Add(1)
		go func(idx int, name, binary, versionFlag string, required bool) {
			defer wg.Done()
			ds := DepStatus{
				Name:     name,
				Binary:   binary,
				Required: required,
			}
			out, err := exec.Command(binary, versionFlag).CombinedOutput()
			if err == nil {
				ds.Available = true
				// Extract first line of version output.
				lines := strings.SplitN(string(out), "\n", 2)
				ds.Version = strings.TrimSpace(lines[0])
			}
			results[idx] = ds
		}(i, d.name, d.binary, d.versionFl, d.required)
	}
	wg.Wait()
	return results
}

// LogDeps logs the availability of external tools at startup.
func LogDeps(prefix string) []DepStatus {
	deps := CheckDeps()
	for _, d := range deps {
		if d.Available {
			log.Printf("%sdocexport: %s available (%s)", prefix, d.Name, d.Version)
		} else {
			severity := "optional"
			if d.Required {
				severity = "REQUIRED"
			}
			log.Printf("%sdocexport: %s not found (%s) — %s export features disabled",
				prefix, d.Name, d.Binary, severity)
		}
	}
	return deps
}

// depCache caches tool availability after first probe to avoid repeated
// exec calls during the server's lifetime.
var depCache struct {
	once   sync.Once
	status map[string]bool
}

// IsAvailable returns true if a tool binary is on the PATH.
// Results are cached after the first call per binary.
func IsAvailable(binary string) bool {
	depCache.once.Do(func() {
		depCache.status = make(map[string]bool)
		for _, ds := range CheckDeps() {
			depCache.status[ds.Binary] = ds.Available
		}
	})
	if avail, ok := depCache.status[binary]; ok {
		return avail
	}
	// Probe dynamically for tools not in the standard set.
	_, err := exec.LookPath(binary)
	avail := err == nil
	depCache.status[binary] = avail
	return avail
}

// RequireTool returns an error if the named binary is not available.
func RequireTool(name, binary string) error {
	if !IsAvailable(binary) {
		return fmt.Errorf("%s (%s) is not installed; please install it and ensure it is on your PATH", name, binary)
	}
	return nil
}
