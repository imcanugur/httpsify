package proxy

import (
	_ "embed"
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/imcanugur/httpsify/internal/version"
)

//go:embed landing.html
var landingPageHTML string

func (s *Server) serveLandingPage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.WriteHeader(http.StatusOK)

	httpPorts, otherPorts := s.getListeningPorts()
	sort.Ints(httpPorts)
	sort.Ints(otherPorts)

	var httpHTML strings.Builder
	if len(httpPorts) == 0 {
		httpHTML.WriteString("<div style=\"font-size: 13px; color: var(--muted); font-style: italic;\">No proxy-ready services detected.</div>")
	} else {
		for _, port := range httpPorts {
			url := fmt.Sprintf("https://%d.localhost", port)
			httpHTML.WriteString(fmt.Sprintf(`
            <a href="%s" class="port-item">
                <span class="port-name">%d.localhost</span>
                <span class="port-action">Proxy Ready</span>
            </a>`, url, port))
		}
	}

	var otherHTML strings.Builder
	for _, port := range otherPorts {
		otherHTML.WriteString(fmt.Sprintf(`
        <div class="port-item other-service">
            <span class="port-name">Port %d</span>
            <span class="port-action">System</span>
        </div>`, port))
	}

	otherSectionClass := ""
	if len(otherPorts) == 0 {
		otherSectionClass = "hidden"
	}

	ver := version.Get()
	output := strings.NewReplacer(
		"{{.HTTP_LIST}}", httpHTML.String(),
		"{{.OTHER_SECTION_CLASS}}", otherSectionClass,
		"{{.OTHER_LIST}}", otherHTML.String(),
		"{{.VERSION}}", ver.Version,
	).Replace(landingPageHTML)

	w.Write([]byte(output))
}
