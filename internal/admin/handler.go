package admin

import (
	"embed"
	"gate-limiter/config/settings"
	"html/template"
	"log/slog"
	"net/http"
)

//go:embed template/status.html
var templateFS embed.FS

var statusTemplate = template.Must(template.ParseFS(templateFS, "template/status.html"))

type StatusHandler struct {
	Config *settings.RootRateLimiterConfig
}

func NewStatusHandler(config *settings.RootRateLimiterConfig) *StatusHandler {
	return &StatusHandler{Config: config}
}

type statusPageData struct {
	TargetConfigured bool
	Target           string
	Port             int
	AdminPort        int
	Strategy         string
	IdentityKey      string
	IdentityHeader   string
	HasClientLimit   bool
	ClientLimit      int
	ClientWindow     int
	Apis             []apiData
	RedisHost        string
	RedisPort        int
	RedisDB          int
}

type apiData struct {
	Identifier     string
	Method         string
	PathExpression string
	PathValue      string
	Limit          int
	WindowSeconds  int
	RefillSeconds  int
	ExpireSeconds  int
	Target         string
}

func (s *StatusHandler) ServeHTTP(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	data := s.buildPageData()

	if err := statusTemplate.Execute(w, data); err != nil {
		slog.Error("admin template execute error", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func (s *StatusHandler) buildPageData() statusPageData {
	rl := s.Config.RateLimiter
	rc := s.Config.RedisConfig

	apis := make([]apiData, 0, len(rl.Apis))
	for _, api := range rl.Apis {
		target := rl.Target
		if api.Target != "" {
			target = api.Target
		}
		if target == "" {
			target = "-"
		}
		apis = append(apis, apiData{
			Identifier:     api.Identifier,
			Method:         api.Method,
			PathExpression: api.Path.Expression,
			PathValue:      api.Path.Value,
			Limit:          api.Limit,
			WindowSeconds:  api.WindowSeconds,
			RefillSeconds:  api.RefillSeconds,
			ExpireSeconds:  api.ExpireSeconds,
			Target:         target,
		})
	}

	identityHeader := ""
	if rl.Identity.Key == "ipv4" {
		identityHeader = rl.Identity.Header
	}

	return statusPageData{
		TargetConfigured: rl.Target != "",
		Target:           rl.Target,
		Port:             rl.Port,
		AdminPort:        rl.AdminPort,
		Strategy:         rl.Strategy,
		IdentityKey:      rl.Identity.Key,
		IdentityHeader:   identityHeader,
		HasClientLimit:   rl.Client.Limit > 0 && rl.Client.WindowSeconds > 0,
		ClientLimit:      rl.Client.Limit,
		ClientWindow:     rl.Client.WindowSeconds,
		Apis:             apis,
		RedisHost:        rc.Host,
		RedisPort:        rc.Port,
		RedisDB:          rc.DB,
	}
}
