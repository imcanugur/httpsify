package proxy

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"html"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/imcanugur/httpsify/internal/netutil"
	"github.com/imcanugur/httpsify/internal/version"
)

//go:embed landing.html
var landingPageHTML string

func (s *Server) serveLandingPage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.WriteHeader(http.StatusOK)

	services := s.getListeningPorts()
	sort.Slice(services, func(i, j int) bool {
		return services[i].Port < services[j].Port
	})

	var webServices []ServiceInfo
	var systemServices []ServiceInfo
	for _, svc := range services {
		if svc.IsWeb {
			webServices = append(webServices, svc)
		} else {
			systemServices = append(systemServices, svc)
		}
	}

	var httpHTML strings.Builder
	if len(webServices) == 0 {
		httpHTML.WriteString("<div style=\"font-size: 13px; color: var(--muted); font-style: italic;\">No proxy-ready services detected.</div>")
	} else {
		for _, svc := range webServices {
			url := fmt.Sprintf("https://%d.localhost", svc.Port)
			svcJSON, _ := json.Marshal(svc)
			httpHTML.WriteString(fmt.Sprintf(`
            <a href="%s" class="port-item" data-service='%s'>
                <span class="port-name">%d.localhost</span>
                <span class="port-action">Proxy Ready</span>
            </a>`, url, html.EscapeString(string(svcJSON)), svc.Port))
		}
	}

	var otherHTML strings.Builder
	for _, svc := range systemServices {
		svcJSON, _ := json.Marshal(svc)
		otherHTML.WriteString(fmt.Sprintf(`
        <div class="port-item other-service" data-service='%s'>
            <span class="port-name">Port %d</span>
            <span class="port-action">System</span>
        </div>`, html.EscapeString(string(svcJSON)), svc.Port))
	}

	otherSectionClass := ""
	if len(systemServices) == 0 {
		otherSectionClass = "hidden"
	}

	uptime := time.Since(s.startTime).Round(time.Second).String()
	reqCount := s.requestCount.Load()
	localIPs := netutil.GetLocalIPs()
	var ipsHTML strings.Builder
	if len(localIPs) == 0 {
		ipsHTML.WriteString("<span class=\"ip-badge\">Localhost Only</span>")
	} else {
		for _, ip := range localIPs {
			ipsHTML.WriteString(fmt.Sprintf("<span class=\"ip-badge\">%s</span>", ip))
		}
	}

	ver := version.Get()
	output := strings.NewReplacer(
		"{{.HTTP_LIST}}", httpHTML.String(),
		"{{.OTHER_SECTION_CLASS}}", otherSectionClass,
		"{{.OTHER_LIST}}", otherHTML.String(),
		"{{.VERSION}}", ver.Version,
		"{{.UPTIME}}", uptime,
		"{{.REQUEST_COUNT}}", fmt.Sprintf("%d", reqCount),
		"{{.LOCAL_IPS}}", ipsHTML.String(),
	).Replace(landingPageHTML)

	w.Write([]byte(output))
}
