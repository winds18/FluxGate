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
	s.mux.HandleFunc("POST /api/nodes/import", s.handleImportNodes)
	s.mux.HandleFunc("PATCH /api/nodes/{id}", s.handleUpdateNode)
	s.mux.HandleFunc("POST /api/nodes/{id}/reset-display-name", s.handleResetNodeDisplayName)

	s.mux.HandleFunc("GET /api/teams", s.handleListTeams)
	s.mux.HandleFunc("POST /api/teams", s.handleCreateTeam)
	s.mux.HandleFunc("GET /api/users", s.handleListUsers)
	s.mux.HandleFunc("POST /api/users", s.handleCreateUser)
	s.mux.HandleFunc("GET /api/tokens", s.handleListTokens)
	s.mux.HandleFunc("POST /api/tokens", s.handleCreateToken)
	s.mux.HandleFunc("POST /api/tokens/{id}/revoke", s.handleRevokeToken)
	s.mux.HandleFunc("POST /api/tokens/{id}/restore", s.handleRestoreToken)
	s.mux.HandleFunc("POST /api/tokens/{id}/extend", s.handleExtendToken)
	s.mux.HandleFunc("POST /api/tokens/{id}/quota", s.handleAddTokenQuota)

	s.mux.HandleFunc("GET /api/virtual-nodes", s.handleListVirtualNodes)
	s.mux.HandleFunc("POST /api/virtual-nodes", s.handleCreateVirtualNode)
	s.mux.HandleFunc("GET /api/policies", s.handleListPolicies)
	s.mux.HandleFunc("POST /api/policies", s.handleCreatePolicy)
	s.mux.HandleFunc("GET /api/traffic/tokens", s.handleListTokenTraffic)

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
	var input struct {
		DisplayPrefix string `json:"display_prefix"`
	}
	if !decodeJSON(w, r, &input) {
		return
	}
	source, err := s.store.UpdateSourcePrefix(r.Context(), id, input.DisplayPrefix)
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
	refresher := upstreamsync.Refresher{Store: s.store}
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
	var input struct {
		DisplayName string `json:"display_name"`
	}
	if !decodeJSON(w, r, &input) {
		return
	}
	node, err := s.store.UpdateNodeDisplayName(r.Context(), id, input.DisplayName)
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

func (s *Server) handleListTokens(w http.ResponseWriter, r *http.Request) {
	tokens, err := s.store.ListTokens(r.Context())
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
	result, err := s.store.CreateToken(r.Context(), s.cfg.TokenSecret, s.cfg.PublicBaseURL, input)
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

func (s *Server) handleListTokenTraffic(w http.ResponseWriter, r *http.Request) {
	summaries, err := s.store.ListTokenTraffic(r.Context())
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
	tokens, err := s.store.ListTokens(ctx)
	if err != nil {
		return singbox.Config{}, err
	}
	virtualNodes, err := s.store.ListVirtualNodes(ctx)
	if err != nil {
		return singbox.Config{}, err
	}
	upstreamNodes, err := s.store.ListNodes(ctx)
	if err != nil {
		return singbox.Config{}, err
	}
	policies, err := s.store.ListPolicies(ctx)
	if err != nil {
		return singbox.Config{}, err
	}
	return singbox.BuildConfigWithPolicies(tokens, virtualNodes, upstreamNodes, policies), nil
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
		GatewayHost:   s.cfg.GatewayHost,
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
