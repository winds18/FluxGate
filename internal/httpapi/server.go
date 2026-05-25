package httpapi

import (
	"context"
	"database/sql"
	"embed"
	"encoding/json"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/winds18/FluxGate/internal/config"
	"github.com/winds18/FluxGate/internal/observability"
	"github.com/winds18/FluxGate/internal/policy"
	"github.com/winds18/FluxGate/internal/security"
	"github.com/winds18/FluxGate/internal/singbox"
	"github.com/winds18/FluxGate/internal/store"
	"github.com/winds18/FluxGate/internal/subscription"
	"github.com/winds18/FluxGate/internal/substore"
	"github.com/winds18/FluxGate/internal/upstreamsync"
)

//go:embed static/*
var staticFiles embed.FS

type Server struct {
	cfg    config.Config
	store  *store.Store
	logger *slog.Logger
	mux    *http.ServeMux
}

const (
	adminSessionCookie = "fg_admin_session"
	adminSessionTTL    = 12 * time.Hour
)

type contextAdminKey struct{}

func NewServer(cfg config.Config, store *store.Store, logger *slog.Logger) http.Handler {
	server := &Server{
		cfg:    cfg,
		store:  store,
		logger: logger,
		mux:    http.NewServeMux(),
	}
	server.routes()
	return server.middleware(server.mux)
}

func (s *Server) routes() {
	s.mux.HandleFunc("GET /", s.handleIndex)
	s.mux.HandleFunc("GET /assets/app.js", s.handleStatic("static/app.js", "application/javascript; charset=utf-8"))
	s.mux.HandleFunc("GET /assets/styles.css", s.handleStatic("static/styles.css", "text/css; charset=utf-8"))

	s.mux.HandleFunc("GET /healthz", s.handleHealth)
	s.mux.HandleFunc("GET /readyz", s.handleReady)
	s.mux.HandleFunc("GET /api/auth/session", s.handleAuthSession)
	s.mux.HandleFunc("POST /api/auth/login", s.handleAuthLogin)
	s.mux.HandleFunc("POST /api/auth/logout", s.handleAuthLogout)
	s.mux.HandleFunc("GET /api/overview", s.handleOverview)

	s.mux.HandleFunc("GET /api/sources", s.handleListSources)
	s.mux.HandleFunc("POST /api/sources", s.handleCreateSource)
	s.mux.HandleFunc("PATCH /api/sources/{id}", s.handleUpdateSource)
	s.mux.HandleFunc("POST /api/sources/{id}/refresh", s.handleRefreshSource)
	s.mux.HandleFunc("POST /api/sources/{id}/regenerate-node-names", s.handleRegenerateSourceNodeNames)

	s.mux.HandleFunc("GET /api/nodes", s.handleListNodes)
	s.mux.HandleFunc("GET /api/nodes/{id}", s.handleGetNode)
	s.mux.HandleFunc("POST /api/nodes/import", s.handleImportNodes)
	s.mux.HandleFunc("PATCH /api/nodes/{id}", s.handleUpdateNode)
	s.mux.HandleFunc("POST /api/nodes/{id}/reset-display-name", s.handleResetNodeDisplayName)

	s.mux.HandleFunc("GET /api/teams", s.handleListTeams)
	s.mux.HandleFunc("POST /api/teams", s.handleCreateTeam)
	s.mux.HandleFunc("PATCH /api/teams/{id}", s.handleUpdateTeam)
	s.mux.HandleFunc("GET /api/users", s.handleListUsers)
	s.mux.HandleFunc("POST /api/users", s.handleCreateUser)
	s.mux.HandleFunc("PATCH /api/users/{id}", s.handleUpdateUser)
	s.mux.HandleFunc("GET /api/tokens", s.handleListTokens)
	s.mux.HandleFunc("POST /api/tokens", s.handleCreateToken)
	s.mux.HandleFunc("POST /api/tokens/{id}/revoke", s.handleRevokeToken)
	s.mux.HandleFunc("POST /api/tokens/{id}/restore", s.handleRestoreToken)
	s.mux.HandleFunc("POST /api/tokens/{id}/rotate-subscription", s.handleRotateTokenSubscription)
	s.mux.HandleFunc("POST /api/tokens/{id}/extend", s.handleExtendToken)
	s.mux.HandleFunc("POST /api/tokens/{id}/quota", s.handleAddTokenQuota)

	s.mux.HandleFunc("GET /api/virtual-nodes", s.handleListVirtualNodes)
	s.mux.HandleFunc("POST /api/virtual-nodes", s.handleCreateVirtualNode)
	s.mux.HandleFunc("PATCH /api/virtual-nodes/{id}", s.handleUpdateVirtualNode)
	s.mux.HandleFunc("GET /api/policies", s.handleListPolicies)
	s.mux.HandleFunc("POST /api/policies", s.handleCreatePolicy)
	s.mux.HandleFunc("PATCH /api/policies/{id}", s.handleUpdatePolicy)
	s.mux.HandleFunc("GET /api/traffic/tokens", s.handleListTokenTraffic)
	s.mux.HandleFunc("GET /api/traffic/daily", s.handleListTrafficDaily)
	s.mux.HandleFunc("GET /api/traffic/hourly", s.handleListTrafficHourly)
	s.mux.HandleFunc("GET /api/traffic/outbounds", s.handleListOutboundTraffic)
	s.mux.HandleFunc("GET /api/delivery/readiness", s.handleDeliveryReadiness)

	s.mux.HandleFunc("POST /api/sing-box/config/generate", s.handleGenerateSingBoxConfig)
	s.mux.HandleFunc("POST /api/sing-box/config/check", s.handleCheckSingBoxConfig)
	s.mux.HandleFunc("POST /api/sing-box/config/publish", s.handlePublishSingBoxConfig)
	s.mux.HandleFunc("POST /api/sing-box/config/rollback", s.handleRollbackSingBoxConfig)
	s.mux.HandleFunc("POST /api/sing-box/restart", s.handleRestartSingBox)
	s.mux.HandleFunc("GET /sub/{token}", s.handleSubscription)
}

func (s *Server) middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := observability.NewRequestID()
		started := time.Now()
		w.Header().Set("x-request-id", requestID)
		ctx := context.WithValue(r.Context(), requestIDKey{}, requestID)
		if s.requiresAdmin(r.URL.Path) {
			admin, reason, ok := s.currentAdmin(r)
			if !ok {
				s.logger.Info("admin session rejected",
					"request_id", requestID,
					"path", redactedPath(r.URL.Path),
					"remote_addr", r.RemoteAddr,
					"reason", reason,
				)
				writeError(w, http.StatusUnauthorized, "login required")
				s.logger.Info("request completed",
					"request_id", requestID,
					"method", r.Method,
					"path", redactedPath(r.URL.Path),
					"remote_addr", r.RemoteAddr,
					"duration_ms", time.Since(started).Milliseconds(),
					"status", http.StatusUnauthorized,
				)
				return
			}
			ctx = context.WithValue(ctx, contextAdminKey{}, admin)
		}
		next.ServeHTTP(w, r.WithContext(ctx))
		s.logger.Info("request completed",
			"request_id", requestID,
			"method", r.Method,
			"path", redactedPath(r.URL.Path),
			"remote_addr", r.RemoteAddr,
			"duration_ms", time.Since(started).Milliseconds(),
		)
	})
}

func (s *Server) requiresAdmin(path string) bool {
	return strings.HasPrefix(path, "/api/") && !strings.HasPrefix(path, "/api/auth/")
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	s.serveEmbedded(w, "static/index.html", "text/html; charset=utf-8")
}

func (s *Server) handleStatic(name, contentType string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s.serveEmbedded(w, name, contentType)
	}
}

func (s *Server) serveEmbedded(w http.ResponseWriter, name, contentType string) {
	bytes, err := staticFiles.ReadFile(name)
	if err != nil {
		http.NotFound(w, nil)
		return
	}
	w.Header().Set("content-type", contentType)
	_, _ = w.Write(bytes)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"status": "ok", "version": s.cfg.Version})
}

func (s *Server) handleReady(w http.ResponseWriter, r *http.Request) {
	if err := s.store.Ping(r.Context()); err != nil {
		writeError(w, http.StatusServiceUnavailable, "database is not ready")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"status": "ready"})
}

func (s *Server) handleAuthSession(w http.ResponseWriter, r *http.Request) {
	admin, reason, ok := s.currentAdmin(r)
	if !ok {
		s.logger.Info("admin session rejected",
			"request_id", requestIDFromContext(r.Context()),
			"path", redactedPath(r.URL.Path),
			"remote_addr", r.RemoteAddr,
			"reason", reason,
		)
		writeError(w, http.StatusUnauthorized, "login required")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"admin": admin})
}

func (s *Server) handleAuthLogin(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if !decodeJSON(w, r, &input) {
		return
	}
	admin, err := s.store.AuthenticateAdmin(r.Context(), input.Username, input.Password)
	if err != nil {
		s.logger.Info("admin login failed",
			"request_id", requestIDFromContext(r.Context()),
			"username_present", strings.TrimSpace(input.Username) != "",
			"remote_addr", r.RemoteAddr,
			"reason", "invalid_credentials",
		)
		writeError(w, http.StatusUnauthorized, "invalid username or password")
		return
	}
	session, err := security.NewSessionToken(s.cfg.SessionSecret, admin.ID, time.Now().UTC(), adminSessionTTL)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create session")
		return
	}
	http.SetCookie(w, s.sessionCookie(r, session, int(adminSessionTTL.Seconds())))
	s.logger.Info("admin login succeeded",
		"request_id", requestIDFromContext(r.Context()),
		"admin_id", admin.ID,
		"remote_addr", r.RemoteAddr,
		"secure_cookie", isHTTPSRequest(r),
	)
	writeJSON(w, http.StatusOK, map[string]any{"admin": admin})
}

func (s *Server) handleAuthLogout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, s.sessionCookie(r, "", -1))
	s.logger.Info("admin logout",
		"request_id", requestIDFromContext(r.Context()),
		"remote_addr", r.RemoteAddr,
	)
	writeJSON(w, http.StatusOK, map[string]any{"status": "logged_out"})
}

func (s *Server) currentAdmin(r *http.Request) (store.Admin, string, bool) {
	cookie, err := r.Cookie(adminSessionCookie)
	if err != nil || cookie.Value == "" {
		return store.Admin{}, "missing_cookie", false
	}
	adminID, ok := security.VerifySessionToken(s.cfg.SessionSecret, cookie.Value, time.Now().UTC())
	if !ok {
		return store.Admin{}, "invalid_or_expired_session", false
	}
	admin, err := s.store.GetAdmin(r.Context(), adminID)
	if err != nil {
		return store.Admin{}, "admin_not_found", false
	}
	if admin.Status != "active" {
		return store.Admin{}, "admin_inactive", false
	}
	return admin, "", true
}

func (s *Server) sessionCookie(r *http.Request, value string, maxAge int) *http.Cookie {
	return &http.Cookie{
		Name:     adminSessionCookie,
		Value:    value,
		Path:     "/",
		MaxAge:   maxAge,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   isHTTPSRequest(r),
	}
}

func isHTTPSRequest(r *http.Request) bool {
	if r.TLS != nil {
		return true
	}
	if strings.EqualFold(r.Header.Get("x-forwarded-proto"), "https") {
		return true
	}
	return strings.Contains(strings.ToLower(r.Header.Get("forwarded")), "proto=https")
}

func (s *Server) handleOverview(w http.ResponseWriter, r *http.Request) {
	overview, err := s.store.Overview(r.Context(), s.cfg.Version)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, overview)
}

type deliveryReadinessResponse struct {
	Ready       bool                     `json:"ready"`
	ReadyCount  int                      `json:"ready_count"`
	TotalChecks int                      `json:"total_checks"`
	Checks      []deliveryReadinessCheck `json:"checks"`
	Config      deliveryConfigReadiness  `json:"config"`
	NextActions []string                 `json:"next_actions"`
	UpdatedAt   string                   `json:"updated_at"`
}

type deliveryReadinessCheck struct {
	Key     string `json:"key"`
	Symbol  string `json:"symbol"`
	Title   string `json:"title"`
	Ready   bool   `json:"ready"`
	Status  string `json:"status"`
	Summary string `json:"summary"`
}

type deliveryConfigReadiness struct {
	Valid                 bool     `json:"valid"`
	ConfigHash            string   `json:"config_hash"`
	InboundCount          int      `json:"inbound_count"`
	OutboundCount         int      `json:"outbound_count"`
	UpstreamOutboundCount int      `json:"upstream_outbound_count"`
	UserCount             int      `json:"user_count"`
	Messages              []string `json:"messages"`
}

func (s *Server) handleDeliveryReadiness(w http.ResponseWriter, r *http.Request) {
	readiness, err := s.deliveryReadiness(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, readiness)
}

func (s *Server) deliveryReadiness(ctx context.Context) (deliveryReadinessResponse, error) {
	overview, err := s.store.Overview(ctx, s.cfg.Version)
	if err != nil {
		return deliveryReadinessResponse{}, err
	}
	nodes, err := s.store.ListNodes(ctx)
	if err != nil {
		return deliveryReadinessResponse{}, err
	}
	virtualNodes, err := s.store.ListVirtualNodes(ctx)
	if err != nil {
		return deliveryReadinessResponse{}, err
	}
	tokens, err := s.store.ListTokens(ctx)
	if err != nil {
		return deliveryReadinessResponse{}, err
	}
	policies, err := s.store.ListPolicies(ctx)
	if err != nil {
		return deliveryReadinessResponse{}, err
	}
	config, err := s.buildSingBoxConfig(ctx)
	if err != nil {
		return deliveryReadinessResponse{}, err
	}
	configCheck, err := singbox.CheckConfig(config)
	if err != nil {
		return deliveryReadinessResponse{}, err
	}

	activeNodes := 0
	for _, node := range nodes {
		if node.Status == "active" {
			activeNodes++
		}
	}
	activeVirtualNodes := 0
	for _, node := range virtualNodes {
		if node.Status == "active" && node.ListenPort > 0 {
			activeVirtualNodes++
		}
	}
	usableTokens := 0
	now := time.Now().UTC()
	for _, token := range tokens {
		if ok, _ := subscription.TokenUsable(token, now); ok {
			usableTokens++
		}
	}
	activePolicies := 0
	for _, policy := range policies {
		if policy.Status == "active" {
			activePolicies++
		}
	}
	configReady := configCheck.Valid &&
		configCheck.InboundCount > 0 &&
		configCheck.UpstreamOutboundCount > 0 &&
		configCheck.UserCount > 0
	checks := []deliveryReadinessCheck{
		deliveryCheck("sources", "源", "接入来源", overview.Sources > 0, countSummary(overview.Sources, "个来源")),
		deliveryCheck("nodes", "点", "可用节点", activeNodes > 0, countSummary(int64(activeNodes), "个可用节点")),
		deliveryCheck("virtual_nodes", "网", "虚拟网关", activeVirtualNodes > 0, countSummary(int64(activeVirtualNodes), "个活跃入口")),
		deliveryCheck("tokens", "钥", "可用 Token", usableTokens > 0, countSummary(int64(usableTokens), "个可用 Token")),
		deliveryCheck("policies", "策", "访问策略", activePolicies > 0, countSummary(int64(activePolicies), "条活跃策略")),
		deliveryCheck("config", "运", "网关配置", configReady, configSummary(configCheck)),
	}

	readyCount := 0
	var nextActions []string
	for _, check := range checks {
		if check.Ready {
			readyCount++
			continue
		}
		nextActions = append(nextActions, "补齐"+check.Title)
	}
	if len(nextActions) == 0 {
		nextActions = []string{"复制 Token 订阅地址导入客户端", "从客户端连接网关并观察流量摘要"}
	}

	return deliveryReadinessResponse{
		Ready:       readyCount == len(checks),
		ReadyCount:  readyCount,
		TotalChecks: len(checks),
		Checks:      checks,
		Config: deliveryConfigReadiness{
			Valid:                 configCheck.Valid,
			ConfigHash:            configCheck.ConfigHash,
			InboundCount:          configCheck.InboundCount,
			OutboundCount:         configCheck.OutboundCount,
			UpstreamOutboundCount: configCheck.UpstreamOutboundCount,
			UserCount:             configCheck.UserCount,
			Messages:              configCheck.Messages,
		},
		NextActions: nextActions,
		UpdatedAt:   time.Now().UTC().Format(time.RFC3339),
	}, nil
}

func deliveryCheck(key, symbol, title string, ready bool, summary string) deliveryReadinessCheck {
	status := "待补"
	if ready {
		status = "就绪"
	}
	return deliveryReadinessCheck{
		Key:     key,
		Symbol:  symbol,
		Title:   title,
		Ready:   ready,
		Status:  status,
		Summary: summary,
	}
}

func countSummary(value int64, unit string) string {
	return strconv.FormatInt(value, 10) + " " + unit
}

func configSummary(result singbox.CheckResult) string {
	if !result.Valid {
		return "配置待修复"
	}
	return strconv.Itoa(result.InboundCount) + " 入站 / " +
		strconv.Itoa(result.UserCount) + " 用户 / " +
		strconv.Itoa(result.UpstreamOutboundCount) + " 上游"
}

func (s *Server) handleListSources(w http.ResponseWriter, r *http.Request) {
	items, err := s.store.ListSources(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, items)
}

func (s *Server) handleCreateSource(w http.ResponseWriter, r *http.Request) {
	var input store.CreateSourceInput
	if !decodeJSON(w, r, &input) {
		return
	}
	source, err := s.store.CreateSource(r.Context(), input)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, source)
}

func (s *Server) handleUpdateSource(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	var input store.UpdateSourceInput
	if !decodeJSON(w, r, &input) {
		return
	}
	source, err := s.store.UpdateSource(r.Context(), id, input)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, source)
}

func (s *Server) handleRefreshSource(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	source, err := s.store.GetSource(r.Context(), id)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	refresher := s.sourceRefresher()
	result, err := refresher.RefreshSource(r.Context(), source)
	if err != nil {
		_ = s.store.SetSourceSyncError(r.Context(), source.ID, err.Error())
		s.logger.Info("source refresh failed",
			"request_id", requestIDFromContext(r.Context()),
			"source_id", source.ID,
			"source_type", source.Type,
			"reason", err.Error(),
		)
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	updatedSource, err := s.store.GetSource(r.Context(), source.ID)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	s.logger.Info("source refresh succeeded",
		"request_id", requestIDFromContext(r.Context()),
		"source_id", source.ID,
		"source_type", source.Type,
		"imported", result.Imported,
		"updated", result.Updated,
		"skipped", result.Skipped,
		"inactivated", result.Inactivated,
	)
	writeJSON(w, http.StatusOK, map[string]any{
		"source": updatedSource,
		"result": result,
	})
}

func (s *Server) sourceRefresher() upstreamsync.Refresher {
	return upstreamsync.Refresher{
		Store: s.store,
		SubStore: substore.Extractor{
			URLTemplate: s.cfg.SubStoreExtractURLTemplate,
			Timeout:     s.cfg.SubStoreTimeout,
		},
		Logger: s.logger,
	}
}

func (s *Server) handleRegenerateSourceNodeNames(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	if err := s.store.RegenerateSourceNodeNames(r.Context(), id); err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"status": "regenerated"})
}

func (s *Server) handleListNodes(w http.ResponseWriter, r *http.Request) {
	nodes, err := s.store.ListNodes(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, nodes)
}

func (s *Server) handleGetNode(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	node, err := s.store.GetNode(r.Context(), id)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, node)
}

func (s *Server) handleImportNodes(w http.ResponseWriter, r *http.Request) {
	var input store.ImportNodesInput
	if !decodeJSON(w, r, &input) {
		return
	}
	result, err := s.store.ImportNodes(r.Context(), input)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (s *Server) handleUpdateNode(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	var input store.UpdateNodeInput
	if !decodeJSON(w, r, &input) {
		return
	}
	node, err := s.store.UpdateNode(r.Context(), id, input)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, node)
}

func (s *Server) handleResetNodeDisplayName(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	node, err := s.store.ResetNodeDisplayName(r.Context(), id)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, node)
}

func (s *Server) handleListTeams(w http.ResponseWriter, r *http.Request) {
	teams, err := s.store.ListTeams(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, teams)
}

func (s *Server) handleCreateTeam(w http.ResponseWriter, r *http.Request) {
	var input store.CreateTeamInput
	if !decodeJSON(w, r, &input) {
		return
	}
	team, err := s.store.CreateTeam(r.Context(), input)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, team)
}

func (s *Server) handleUpdateTeam(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	var input store.UpdateTeamInput
	if !decodeJSON(w, r, &input) {
		return
	}
	team, err := s.store.UpdateTeam(r.Context(), id, input)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, team)
}

func (s *Server) handleListUsers(w http.ResponseWriter, r *http.Request) {
	users, err := s.store.ListUsers(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, users)
}

func (s *Server) handleCreateUser(w http.ResponseWriter, r *http.Request) {
	var input store.CreateUserInput
	if !decodeJSON(w, r, &input) {
		return
	}
	user, err := s.store.CreateUser(r.Context(), input)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, user)
}

func (s *Server) handleUpdateUser(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	var input store.UpdateUserInput
	if !decodeJSON(w, r, &input) {
		return
	}
	user, err := s.store.UpdateUser(r.Context(), id, input)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, user)
}

func (s *Server) handleListTokens(w http.ResponseWriter, r *http.Request) {
	tokens, err := s.store.ListTokensForAdmin(r.Context(), s.cfg.TokenSecret, s.publicBaseURL(r))
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, tokens)
}

func (s *Server) handleCreateToken(w http.ResponseWriter, r *http.Request) {
	var input store.CreateTokenInput
	if !decodeJSON(w, r, &input) {
		return
	}
	result, err := s.store.CreateToken(r.Context(), s.cfg.TokenSecret, s.publicBaseURL(r), input)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, result)
}

func (s *Server) handleRevokeToken(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	token, err := s.store.RevokeToken(r.Context(), id)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, token)
}

func (s *Server) handleRotateTokenSubscription(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	result, err := s.store.RotateTokenSubscription(r.Context(), s.cfg.TokenSecret, s.publicBaseURL(r), id)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (s *Server) handleRestoreToken(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	token, err := s.store.RestoreToken(r.Context(), id)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, token)
}

func (s *Server) handleExtendToken(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	var input struct {
		ExtendDays int `json:"extend_days"`
	}
	if !decodeJSON(w, r, &input) {
		return
	}
	token, err := s.store.ExtendToken(r.Context(), id, input.ExtendDays)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, token)
}

func (s *Server) handleAddTokenQuota(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	var input struct {
		QuotaBytes int64 `json:"quota_bytes"`
	}
	if !decodeJSON(w, r, &input) {
		return
	}
	token, err := s.store.AddTokenQuota(r.Context(), id, input.QuotaBytes)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, token)
}

func (s *Server) handleListVirtualNodes(w http.ResponseWriter, r *http.Request) {
	nodes, err := s.store.ListVirtualNodes(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, nodes)
}

func (s *Server) handleCreateVirtualNode(w http.ResponseWriter, r *http.Request) {
	var input store.CreateVirtualNodeInput
	if !decodeJSON(w, r, &input) {
		return
	}
	if input.ListenPort == 0 {
		input.ListenPort = s.cfg.DefaultVLESSPort
	}
	node, err := s.store.CreateVirtualNode(r.Context(), input)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, node)
}

func (s *Server) handleUpdateVirtualNode(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	var input store.UpdateVirtualNodeInput
	if !decodeJSON(w, r, &input) {
		return
	}
	node, err := s.store.UpdateVirtualNode(r.Context(), id, input)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, node)
}

func (s *Server) handleListPolicies(w http.ResponseWriter, r *http.Request) {
	policies, err := s.store.ListPolicies(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, policies)
}

func (s *Server) handleCreatePolicy(w http.ResponseWriter, r *http.Request) {
	var input store.CreatePolicyInput
	if !decodeJSON(w, r, &input) {
		return
	}
	policy, err := s.store.CreatePolicy(r.Context(), input)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, policy)
}

func (s *Server) handleUpdatePolicy(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(w, r)
	if !ok {
		return
	}
	var input store.UpdatePolicyInput
	if !decodeJSON(w, r, &input) {
		return
	}
	policy, err := s.store.UpdatePolicy(r.Context(), id, input)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, policy)
}

func (s *Server) handleListTokenTraffic(w http.ResponseWriter, r *http.Request) {
	summaries, err := s.store.ListTokenTraffic(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, summaries)
}

func (s *Server) handleListTrafficDaily(w http.ResponseWriter, r *http.Request) {
	days := 14
	if raw := strings.TrimSpace(r.URL.Query().Get("days")); raw != "" {
		value, err := strconv.Atoi(raw)
		if err != nil || value <= 0 {
			writeError(w, http.StatusBadRequest, "days must be a positive integer")
			return
		}
		days = value
	}
	summaries, err := s.store.ListTrafficDaily(r.Context(), days)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, summaries)
}

func (s *Server) handleListTrafficHourly(w http.ResponseWriter, r *http.Request) {
	hours := 24
	if raw := strings.TrimSpace(r.URL.Query().Get("hours")); raw != "" {
		value, err := strconv.Atoi(raw)
		if err != nil || value <= 0 {
			writeError(w, http.StatusBadRequest, "hours must be a positive integer")
			return
		}
		hours = value
	}
	summaries, err := s.store.ListTrafficHourly(r.Context(), hours)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, summaries)
}

func (s *Server) handleListOutboundTraffic(w http.ResponseWriter, r *http.Request) {
	days := 14
	if raw := strings.TrimSpace(r.URL.Query().Get("days")); raw != "" {
		value, err := strconv.Atoi(raw)
		if err != nil || value <= 0 {
			writeError(w, http.StatusBadRequest, "days must be a positive integer")
			return
		}
		days = value
	}
	summaries, err := s.store.ListOutboundTraffic(r.Context(), days)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, summaries)
}

func (s *Server) handleGenerateSingBoxConfig(w http.ResponseWriter, r *http.Request) {
	config, err := s.buildSingBoxConfig(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	body, err := singbox.Marshal(config)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.Header().Set("content-type", "application/json; charset=utf-8")
	_, _ = w.Write(body)
}

func (s *Server) handleCheckSingBoxConfig(w http.ResponseWriter, r *http.Request) {
	config, err := s.buildSingBoxConfig(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	result, err := singbox.CheckConfig(config)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	status := http.StatusOK
	if !result.Valid {
		status = http.StatusUnprocessableEntity
	}
	writeJSON(w, status, result)
}

func (s *Server) handlePublishSingBoxConfig(w http.ResponseWriter, r *http.Request) {
	config, err := s.buildSingBoxConfig(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	result, err := singbox.PublishConfig(config, s.cfg.SingBoxConfigPath, s.cfg.SingBoxPreviousConfigPath)
	if err != nil {
		if !result.Valid {
			writeJSON(w, http.StatusUnprocessableEntity, result)
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if s.cfg.SingBoxAutoRestart {
		restart := s.restartSingBox(r.Context())
		result.Restart = &restart
		result.RestartRequired = !restart.Success
		if !restart.Success {
			writeJSON(w, http.StatusInternalServerError, result)
			return
		}
	}
	writeJSON(w, http.StatusOK, result)
}

func (s *Server) handleRollbackSingBoxConfig(w http.ResponseWriter, r *http.Request) {
	result, err := singbox.RollbackConfig(s.cfg.SingBoxConfigPath, s.cfg.SingBoxPreviousConfigPath)
	if err != nil {
		if !result.Valid {
			writeJSON(w, http.StatusUnprocessableEntity, result)
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if s.cfg.SingBoxAutoRestart {
		restart := s.restartSingBox(r.Context())
		result.Restart = &restart
		result.RestartRequired = !restart.Success
		if !restart.Success {
			writeJSON(w, http.StatusInternalServerError, result)
			return
		}
	}
	writeJSON(w, http.StatusOK, result)
}

func (s *Server) handleRestartSingBox(w http.ResponseWriter, r *http.Request) {
	result := s.restartSingBox(r.Context())
	status := http.StatusOK
	if result.Enabled && !result.Success {
		status = http.StatusInternalServerError
	}
	writeJSON(w, status, result)
}

func (s *Server) restartSingBox(ctx context.Context) singbox.RestartResult {
	result := singbox.Restart(ctx, singbox.RestartOptions{
		Enabled:       s.cfg.SingBoxAutoRestart,
		Driver:        s.cfg.SingBoxRestartDriver,
		DockerSocket:  s.cfg.SingBoxDockerSocket,
		ContainerName: s.cfg.SingBoxContainerName,
		Command:       s.cfg.SingBoxRestartCommand,
		Args:          s.cfg.SingBoxRestartArgs,
		Timeout:       s.cfg.SingBoxRestartTimeout,
	})
	s.logger.Info("sing-box restart evaluated",
		"enabled", result.Enabled,
		"executed", result.Executed,
		"success", result.Success,
		"skipped", result.Skipped,
		"duration_ms", result.DurationMS,
		"message", result.Message,
	)
	return result
}

func (s *Server) buildSingBoxConfig(ctx context.Context) (singbox.Config, error) {
	return singbox.BuildConfigFromStore(ctx, s.store)
}

func (s *Server) handleSubscription(w http.ResponseWriter, r *http.Request) {
	plainToken := r.PathValue("token")
	tokenHash := security.TokenHash(s.cfg.TokenSecret, plainToken)
	token, err := s.store.TokenByHash(r.Context(), tokenHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusUnauthorized, "invalid token")
			return
		}
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if ok, reason := subscription.TokenUsable(token, time.Now().UTC()); !ok {
		writeError(w, http.StatusForbidden, reason)
		return
	}
	virtualNodes, err := s.store.ListVirtualNodes(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	policies, err := s.store.ListPolicies(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	virtualNodes = policy.FilterVirtualNodes(virtualNodes, token, policies)
	target := r.URL.Query().Get("target")
	if target == "" {
		target = inferTarget(r.UserAgent())
	}
	response, err := subscription.Build(subscription.Request{
		Target:        target,
		GatewayHost:   s.publicGatewayHost(r),
		PublicBaseURL: s.cfg.PublicBaseURL,
	}, token, virtualNodes)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	_ = s.store.TouchTokenUsed(r.Context(), token.ID)
	subscription.SetUserInfoHeader(w.Header(), token)
	w.Header().Set("content-type", response.ContentType)
	_, _ = w.Write(response.Body)
}

func (s *Server) publicBaseURL(r *http.Request) string {
	host := publicRequestHost(r)
	if !safePublicHost(host) {
		return strings.TrimRight(s.cfg.PublicBaseURL, "/")
	}
	proto := strings.ToLower(firstForwardedValue(r.Header.Get("x-forwarded-proto")))
	if proto != "http" && proto != "https" {
		if r.TLS != nil {
			proto = "https"
		} else {
			proto = "http"
		}
	}
	return proto + "://" + host
}

func (s *Server) publicGatewayHost(r *http.Request) string {
	if host := hostWithoutPort(publicRequestHost(r)); host != "" {
		return host
	}
	if host := hostWithoutPort(s.cfg.GatewayHost); host != "" {
		return host
	}
	if host := publicURLHost(s.cfg.PublicBaseURL); host != "" {
		return host
	}
	return strings.TrimSpace(s.cfg.GatewayHost)
}

func publicRequestHost(r *http.Request) string {
	host := firstForwardedValue(r.Header.Get("x-forwarded-host"))
	if host == "" {
		host = r.Host
	}
	return strings.TrimSpace(host)
}

func hostWithoutPort(host string) string {
	host = strings.TrimSpace(host)
	if !safePublicHost(host) {
		return ""
	}
	if strings.Contains(host, ":") {
		name, _, err := net.SplitHostPort(host)
		if err != nil {
			return ""
		}
		return strings.Trim(name, "[]")
	}
	return strings.Trim(host, "[]")
}

func publicURLHost(rawURL string) string {
	parsed, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil {
		return ""
	}
	return hostWithoutPort(parsed.Host)
}

func firstForwardedValue(value string) string {
	if index := strings.Index(value, ","); index >= 0 {
		value = value[:index]
	}
	return strings.TrimSpace(value)
}

func safePublicHost(host string) bool {
	if host == "" || strings.ContainsAny(host, "/\\@ \t\r\n") {
		return false
	}
	if strings.Contains(host, ":") {
		name, port, err := net.SplitHostPort(host)
		if err != nil || name == "" || port == "" {
			return false
		}
		if _, err := strconv.Atoi(port); err != nil {
			return false
		}
		return true
	}
	return true
}

func inferTarget(userAgent string) string {
	ua := strings.ToLower(userAgent)
	if strings.Contains(ua, "sing-box") || strings.Contains(ua, "singbox") {
		return "sing-box"
	}
	return "clash"
}

func redactedPath(path string) string {
	if strings.HasPrefix(path, "/sub/") {
		return "/sub/<redacted>"
	}
	return path
}

func parseID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	raw := r.PathValue("id")
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, "invalid id")
		return 0, false
	}
	return id, true
}

func decodeJSON(w http.ResponseWriter, r *http.Request, target any) bool {
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(target); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json")
		return false
	}
	return true
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("content-type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeStoreError(w http.ResponseWriter, err error) {
	if errors.Is(err, sql.ErrNoRows) {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	writeError(w, http.StatusBadRequest, err.Error())
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]any{
		"error": message,
	})
}

func remoteIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

type requestIDKey struct{}

func requestIDFromContext(ctx context.Context) string {
	requestID, _ := ctx.Value(requestIDKey{}).(string)
	return requestID
}
