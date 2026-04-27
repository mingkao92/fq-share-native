package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

type sessionContext struct {
	token           string
	rootDir         string
	uploadsDir      string
	sharedDir       string
	uploadTextsPath string
	sharedTextsPath string
}

func runHTTPMode(args []string) error {
	fs := flag.NewFlagSet("serve-http", flag.ContinueOnError)
	port := fs.Int("port", 0, "")
	token := fs.String("token", "", "")
	dataDir := fs.String("data-dir", "", "")
	publicBaseURL := fs.String("public-base-url", "", "")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *port <= 0 || *dataDir == "" {
		return errors.New("missing required args")
	}
	if strings.TrimSpace(*token) != "" && !validTokenPattern.MatchString(strings.TrimSpace(*token)) {
		return errors.New("invalid token format")
	}

	baseURL := strings.TrimRight(strings.TrimSpace(*publicBaseURL), "/")
	if baseURL != "" {
		parsed, err := url.Parse(baseURL)
		if err != nil || parsed.Scheme == "" || parsed.Host == "" {
			return errors.New("invalid --public-base-url")
		}
		if parsed.Scheme != "http" && parsed.Scheme != "https" {
			return errors.New("--public-base-url must be http/https")
		}
	}

	cfg := httpConfig{
		token:         strings.TrimSpace(*token),
		dataDir:       *dataDir,
		sessionsDir:   filepath.Join(*dataDir, "sessions"),
		publicBaseURL: baseURL,
	}
	if err := os.MkdirAll(cfg.sessionsDir, 0o755); err != nil {
		return err
	}
	if cfg.token != "" {
		if _, err := ensureSessionDirs(cfg, cfg.token); err != nil {
			return err
		}
	}

	h := &httpHandler{cfg: cfg, statsByTK: make(map[string]*accessStats)}
	srv := &http.Server{
		Addr:              fmt.Sprintf("0.0.0.0:%d", *port),
		Handler:           h,
		ReadHeaderTimeout: 8 * time.Second,
	}
	return srv.ListenAndServe()
}

func (h *httpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	if r.Method == http.MethodOptions {
		h.handleOptions(w)
		return
	}

	if r.Method == http.MethodGet && path == "/api/health" {
		h.sendJSON(w, http.StatusOK, map[string]any{"ok": true})
		return
	}

	if r.Method == http.MethodPost && path == "/api/session/start" {
		h.handleSessionStart(w, r)
		return
	}

	if r.Method == http.MethodPost && path == "/api/session/stop" {
		h.handleSessionStop(w, r)
		return
	}

	session, status, authErr := h.authorizeSession(r)
	if authErr != "" {
		h.sendJSON(w, status, map[string]any{"ok": false, "error": authErr})
		return
	}

	if r.Method == http.MethodGet && (path == "/" || path == "/index.html") {
		h.markPageVisit(session.token)
		body := []byte(renderMobilePage(session.token, requestLang(r)))
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Content-Length", strconv.Itoa(len(body)))
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(body)
		return
	}

	if r.Method == http.MethodGet && path == "/api/list" {
		h.handleList(w, session)
		return
	}

	if r.Method == http.MethodGet && strings.HasPrefix(path, "/api/download/") {
		h.handleDownload(w, r, session)
		return
	}

	if r.Method == http.MethodPost && path == "/api/clear" {
		h.handleClear(w, r, session)
		return
	}

	if r.Method == http.MethodPost && (path == "/api/upload" || path == "/api/share") {
		h.handleUpload(w, r, session)
		return
	}

	if r.Method == http.MethodPost && (path == "/api/text" || path == "/api/share-text") {
		h.handleText(w, r, session)
		return
	}

	h.sendJSON(w, http.StatusNotFound, map[string]any{"ok": false, "error": "not found"})
}

func requestLang(r *http.Request) string {
	if lang := strings.TrimSpace(r.URL.Query().Get("lang")); lang != "" {
		return normalizeLang(lang)
	}
	return normalizeLang(r.Header.Get("Accept-Language"))
}

func (h *httpHandler) handleOptions(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Click-Token")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.WriteHeader(http.StatusNoContent)
}

func (h *httpHandler) handleSessionStart(w http.ResponseWriter, r *http.Request) {
	token := h.cfg.token
	if token == "" {
		token = strings.TrimSpace(r.URL.Query().Get("token"))
		if token == "" {
			token = strings.TrimSpace(r.Header.Get("X-Click-Token"))
		}
		if token == "" {
			generated, err := generateToken(18)
			if err != nil {
				h.sendJSON(w, http.StatusInternalServerError, map[string]any{"ok": false, "error": "generate token failed"})
				return
			}
			token = generated
		}
	}

	if !validTokenPattern.MatchString(token) {
		h.sendJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "invalid token"})
		return
	}

	if _, err := ensureSessionDirs(h.cfg, token); err != nil {
		h.sendJSON(w, http.StatusInternalServerError, map[string]any{"ok": false, "error": "failed to initialize session"})
		return
	}

	h.sendJSON(w, http.StatusOK, h.buildSessionStatus(r, token))
}

func (h *httpHandler) handleSessionStop(w http.ResponseWriter, r *http.Request) {
	token := strings.TrimSpace(r.URL.Query().Get("token"))
	if token == "" {
		token = strings.TrimSpace(r.Header.Get("X-Click-Token"))
	}
	if token == "" {
		token = h.cfg.token
	}
	if !validTokenPattern.MatchString(token) {
		h.sendJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "invalid token"})
		return
	}
	if h.cfg.token != "" && token != h.cfg.token {
		h.sendJSON(w, http.StatusForbidden, map[string]any{"ok": false, "error": "invalid token"})
		return
	}

	session := buildSessionContext(h.cfg, token)
	deleted := clearFiles(session.uploadsDir) + clearFiles(session.sharedDir)
	deleted += clearTextEntries(session.uploadTextsPath) + clearTextEntries(session.sharedTextsPath)
	_ = os.RemoveAll(session.rootDir)
	h.removeTokenStats(token)

	h.sendJSON(w, http.StatusOK, map[string]any{"ok": true, "deleted": deleted, "token": token})
}

func (h *httpHandler) authorizeSession(r *http.Request) (sessionContext, int, string) {
	token := strings.TrimSpace(r.URL.Query().Get("token"))
	if token == "" {
		token = strings.TrimSpace(r.Header.Get("X-Click-Token"))
	}
	if token == "" {
		return sessionContext{}, http.StatusForbidden, "missing token"
	}
	if h.cfg.token != "" && token != h.cfg.token {
		return sessionContext{}, http.StatusForbidden, "invalid token"
	}
	if !validTokenPattern.MatchString(token) {
		return sessionContext{}, http.StatusForbidden, "invalid token"
	}
	session, err := ensureSessionDirs(h.cfg, token)
	if err != nil {
		return sessionContext{}, http.StatusInternalServerError, "failed to initialize session"
	}
	return session, 0, ""
}

func (h *httpHandler) handleList(w http.ResponseWriter, session sessionContext) {
	prunedUploads := pruneOldFiles(session.uploadsDir, maxFileAge)
	prunedShared := pruneOldFiles(session.sharedDir, maxFileAge)
	prunedUploadTexts := pruneOldTextEntries(session.uploadTextsPath, maxFileAge)
	prunedSharedTexts := pruneOldTextEntries(session.sharedTextsPath, maxFileAge)
	pageVisits, lastPageVisit := h.snapshotPageStats(session.token)

	uploads := listFiles(session.uploadsDir)
	shared := listFiles(session.sharedDir)
	uploadTexts := listTextEntries(session.uploadTextsPath)
	sharedTexts := listTextEntries(session.sharedTextsPath)

	h.sendJSON(w, http.StatusOK, map[string]any{
		"ok":              true,
		"uploads":         uploads,
		"shared":          shared,
		"upload_texts":    uploadTexts,
		"shared_texts":    sharedTexts,
		"timeline":        buildTimeline(shared, uploads, sharedTexts, uploadTexts),
		"page_visits":     pageVisits,
		"last_page_visit": lastPageVisit,
		"pruned": map[string]any{
			"removed_uploads":      prunedUploads,
			"removed_shared":       prunedShared,
			"removed_upload_texts": prunedUploadTexts,
			"removed_shared_texts": prunedSharedTexts,
			"removed_total":        prunedUploads + prunedShared + prunedUploadTexts + prunedSharedTexts,
		},
	})
}

func (h *httpHandler) handleDownload(w http.ResponseWriter, r *http.Request, session sessionContext) {
	parts := strings.SplitN(r.URL.Path, "/", 5)
	if len(parts) < 5 {
		h.sendJSON(w, http.StatusNotFound, map[string]any{"ok": false, "error": "not found"})
		return
	}

	category := parts[3]
	rawName, err := url.PathUnescape(parts[4])
	if err != nil {
		h.sendJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "invalid filename"})
		return
	}
	safeName := sanitizeFilename(rawName)
	if safeName != rawName {
		h.sendJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "invalid filename"})
		return
	}

	baseDir := session.uploadsDir
	if category == "shared" {
		baseDir = session.sharedDir
	} else if category != "uploads" {
		h.sendJSON(w, http.StatusNotFound, map[string]any{"ok": false, "error": "not found"})
		return
	}

	target := filepath.Join(baseDir, safeName)
	fi, err := os.Stat(target)
	if err != nil || fi.IsDir() {
		h.sendJSON(w, http.StatusNotFound, map[string]any{"ok": false, "error": "file not found"})
		return
	}

	f, err := os.Open(target)
	if err != nil {
		h.sendJSON(w, http.StatusInternalServerError, map[string]any{"ok": false, "error": "open file failed"})
		return
	}
	defer f.Close()

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Length", strconv.FormatInt(fi.Size(), 10))
	w.Header().Set("Content-Disposition", buildAttachmentDisposition(fi.Name()))
	w.WriteHeader(http.StatusOK)
	_, _ = io.Copy(w, f)
}

func (h *httpHandler) handleClear(w http.ResponseWriter, r *http.Request, session sessionContext) {
	scope := "all"
	if strings.Contains(strings.ToLower(r.Header.Get("Content-Type")), "application/json") {
		defer r.Body.Close()
		var payload map[string]any
		if err := json.NewDecoder(io.LimitReader(r.Body, 1<<20)).Decode(&payload); err != nil {
			h.sendJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "invalid json payload"})
			return
		}
		if v, ok := payload["scope"].(string); ok {
			scope = strings.ToLower(strings.TrimSpace(v))
		}
	}

	if scope != "all" && scope != "uploads" && scope != "shared" {
		h.sendJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "invalid scope"})
		return
	}

	deleted := 0
	if scope == "all" || scope == "uploads" {
		deleted += clearFiles(session.uploadsDir)
		deleted += clearTextEntries(session.uploadTextsPath)
	}
	if scope == "all" || scope == "shared" {
		deleted += clearFiles(session.sharedDir)
		deleted += clearTextEntries(session.sharedTextsPath)
	}

	h.sendJSON(w, http.StatusOK, map[string]any{"ok": true, "scope": scope, "deleted": deleted})
}

func (h *httpHandler) handleUpload(w http.ResponseWriter, r *http.Request, session sessionContext) {
	if !strings.Contains(strings.ToLower(r.Header.Get("Content-Type")), "multipart/form-data") {
		h.sendJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "Content-Type must be multipart/form-data"})
		return
	}

	pruneOldFiles(session.uploadsDir, maxFileAge)
	pruneOldFiles(session.sharedDir, maxFileAge)

	if err := r.ParseMultipartForm(32 << 20); err != nil {
		h.sendJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "invalid multipart form"})
		return
	}
	if r.MultipartForm != nil {
		defer r.MultipartForm.RemoveAll()
	}

	files := r.MultipartForm.File["file"]
	if len(files) == 0 {
		h.sendJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "missing file"})
		return
	}

	targetDir := session.uploadsDir
	if r.URL.Path == "/api/share" {
		targetDir = session.sharedDir
	}

	saved := make([]map[string]any, 0, len(files))
	for _, fileHeader := range files {
		if fileHeader == nil || strings.TrimSpace(fileHeader.Filename) == "" {
			continue
		}
		name, size, err := saveUploadedFile(fileHeader, targetDir)
		if err != nil {
			continue
		}
		saved = append(saved, map[string]any{"name": name, "size": size})
	}

	if len(saved) == 0 {
		h.sendJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": "no valid files"})
		return
	}

	if strings.Contains(strings.ToLower(r.Header.Get("Accept")), "text/html") {
		body := []byte("<html><meta charset='utf-8'><body><p>上传成功，返回上页查看列表。</p><p><a href='javascript:history.back()'>返回</a></p></body></html>")
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Content-Length", strconv.Itoa(len(body)))
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(body)
		return
	}

	h.sendJSON(w, http.StatusOK, map[string]any{"ok": true, "saved": saved})
}

func (h *httpHandler) handleText(w http.ResponseWriter, r *http.Request, session sessionContext) {
	text, err := readSubmittedText(r)
	if err != nil {
		h.sendJSON(w, http.StatusBadRequest, map[string]any{"ok": false, "error": err.Error()})
		return
	}

	targetPath := session.uploadTextsPath
	if r.URL.Path == "/api/share-text" {
		targetPath = session.sharedTextsPath
	}

	entry, err := appendTextEntry(targetPath, text)
	if err != nil {
		h.sendJSON(w, http.StatusInternalServerError, map[string]any{"ok": false, "error": "save text failed"})
		return
	}

	h.sendJSON(w, http.StatusOK, map[string]any{"ok": true, "saved": entry})
}

func readSubmittedText(r *http.Request) (string, error) {
	contentType := strings.ToLower(r.Header.Get("Content-Type"))
	var text string

	switch {
	case strings.Contains(contentType, "application/json"):
		defer r.Body.Close()
		var payload map[string]any
		if err := json.NewDecoder(io.LimitReader(r.Body, 1<<20)).Decode(&payload); err != nil {
			return "", errors.New("invalid json payload")
		}
		if v, ok := payload["text"].(string); ok {
			text = v
		}
	case strings.Contains(contentType, "application/x-www-form-urlencoded") || strings.Contains(contentType, "multipart/form-data"):
		if err := r.ParseForm(); err != nil {
			return "", errors.New("invalid form payload")
		}
		text = r.Form.Get("text")
	default:
		defer r.Body.Close()
		data, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
		if err != nil {
			return "", errors.New("read body failed")
		}
		text = string(data)
	}

	text = sanitizeText(text)
	if text == "" {
		return "", errors.New("text is empty")
	}
	return text, nil
}

func sanitizeText(raw string) string {
	text := strings.ReplaceAll(raw, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")
	text = strings.TrimSpace(text)
	if text == "" {
		return ""
	}
	runes := []rune(text)
	if len(runes) > 4000 {
		text = string(runes[:4000])
	}
	return text
}

func (h *httpHandler) sendJSON(w http.ResponseWriter, status int, payload map[string]any) {
	data, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Content-Length", strconv.Itoa(len(data)))
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Click-Token")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.WriteHeader(status)
	_, _ = w.Write(data)
}

func (h *httpHandler) buildSessionStatus(r *http.Request, token string) map[string]any {
	baseURL := h.resolveBaseURL(r)
	return map[string]any{
		"ok":        true,
		"running":   true,
		"token":     token,
		"local_url": baseURL,
		"lan_url":   baseURL,
		"phone_url": fmt.Sprintf("%s/?token=%s", baseURL, url.QueryEscape(token)),
	}
}

func (h *httpHandler) resolveBaseURL(r *http.Request) string {
	if h.cfg.publicBaseURL != "" {
		return h.cfg.publicBaseURL
	}

	host := strings.TrimSpace(r.Header.Get("X-Forwarded-Host"))
	if host == "" {
		host = strings.TrimSpace(r.Host)
	}
	if host == "" {
		host = "127.0.0.1"
	}

	proto := "http"
	if r.TLS != nil {
		proto = "https"
	}
	if forwardedProto := strings.TrimSpace(r.Header.Get("X-Forwarded-Proto")); forwardedProto != "" {
		parts := strings.Split(forwardedProto, ",")
		candidate := strings.TrimSpace(parts[0])
		if candidate == "http" || candidate == "https" {
			proto = candidate
		}
	}

	return fmt.Sprintf("%s://%s", proto, host)
}

func (h *httpHandler) markPageVisit(token string) {
	stats := h.getTokenStats(token)
	stats.markPageVisit()
}

func (h *httpHandler) snapshotPageStats(token string) (int, int64) {
	stats := h.getTokenStats(token)
	return stats.snapshot()
}

func (h *httpHandler) getTokenStats(token string) *accessStats {
	h.sessionMu.Lock()
	defer h.sessionMu.Unlock()
	stats, ok := h.statsByTK[token]
	if ok {
		return stats
	}
	stats = &accessStats{}
	h.statsByTK[token] = stats
	return stats
}

func (h *httpHandler) removeTokenStats(token string) {
	h.sessionMu.Lock()
	defer h.sessionMu.Unlock()
	delete(h.statsByTK, token)
}

func buildSessionContext(cfg httpConfig, token string) sessionContext {
	hashed := hashToken(token)
	rootDir := filepath.Join(cfg.sessionsDir, hashed)
	return sessionContext{
		token:           token,
		rootDir:         rootDir,
		uploadsDir:      filepath.Join(rootDir, "uploads"),
		sharedDir:       filepath.Join(rootDir, "shared"),
		uploadTextsPath: filepath.Join(rootDir, "uploads_texts.jsonl"),
		sharedTextsPath: filepath.Join(rootDir, "shared_texts.jsonl"),
	}
}

func ensureSessionDirs(cfg httpConfig, token string) (sessionContext, error) {
	session := buildSessionContext(cfg, token)
	if err := os.MkdirAll(session.uploadsDir, 0o755); err != nil {
		return sessionContext{}, err
	}
	if err := os.MkdirAll(session.sharedDir, 0o755); err != nil {
		return sessionContext{}, err
	}
	return session, nil
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:16])
}

func sanitizeFilename(rawName string) string {
	name := strings.TrimSpace(rawName)
	name = filepath.Base(name)
	name = invalidNameChars.ReplaceAllString(name, "_")
	name = strings.Trim(name, "._")
	if name == "" {
		name = fmt.Sprintf("file_%d", time.Now().Unix())
	}
	if utf8.RuneCountInString(name) > 180 {
		runes := []rune(name)
		return string(runes[:180])
	}
	return name
}

func buildAttachmentDisposition(name string) string {
	asciiFallback := strings.ToValidUTF8(name, "")
	asciiFallback = strings.Map(func(r rune) rune {
		if r < 0x20 || r == 0x7f || r == '"' || r == '\\' {
			return -1
		}
		if r > 0x7e {
			return '_'
		}
		return r
	}, asciiFallback)
	asciiFallback = strings.TrimSpace(asciiFallback)
	if asciiFallback == "" {
		asciiFallback = "download"
	}
	return fmt.Sprintf("attachment; filename=%q; filename*=UTF-8''%s", asciiFallback, url.PathEscape(name))
}

func saveUploadedFile(fileHeader *multipart.FileHeader, targetDir string) (string, int64, error) {
	filename := sanitizeFilename(fileHeader.Filename)

	src, err := fileHeader.Open()
	if err != nil {
		return "", 0, err
	}
	defer src.Close()

	dst, finalName, err := createUniqueFile(targetDir, filename)
	if err != nil {
		return "", 0, err
	}
	defer dst.Close()

	size, err := io.Copy(dst, src)
	if err != nil {
		return "", 0, err
	}
	return finalName, size, nil
}

func createUniqueFile(baseDir, filename string) (*os.File, string, error) {
	ext := filepath.Ext(filename)
	stem := strings.TrimSuffix(filename, ext)

	for index := 0; index < 5000; index++ {
		candidate := filename
		if index > 0 {
			candidate = fmt.Sprintf("%s_%d%s", stem, index, ext)
		}
		path := filepath.Join(baseDir, candidate)
		f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_EXCL, 0o644)
		if err == nil {
			return f, candidate, nil
		}
		if os.IsExist(err) {
			continue
		}
		return nil, "", err
	}

	return nil, "", errors.New("failed to allocate unique filename")
}

func listFiles(baseDir string) []fileEntry {
	entries, err := os.ReadDir(baseDir)
	if err != nil {
		return []fileEntry{}
	}
	result := make([]fileEntry, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		result = append(result, fileEntry{
			Name:  entry.Name(),
			Size:  info.Size(),
			MTime: info.ModTime().Unix(),
		})
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].MTime < result[j].MTime
	})
	return result
}

func pruneOldFiles(baseDir string, maxAge time.Duration) int {
	entries, err := os.ReadDir(baseDir)
	if err != nil {
		return 0
	}
	now := time.Now()
	removed := 0
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		if now.Sub(info.ModTime()) > maxAge {
			if err := os.Remove(filepath.Join(baseDir, entry.Name())); err == nil || os.IsNotExist(err) {
				removed++
			}
		}
	}
	return removed
}

func appendTextEntry(path, text string) (textEntry, error) {
	entry := textEntry{
		ID:    strconv.FormatInt(time.Now().UnixNano(), 36),
		Text:  text,
		MTime: time.Now().Unix(),
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return textEntry{}, err
	}
	defer f.Close()
	data, err := json.Marshal(entry)
	if err != nil {
		return textEntry{}, err
	}
	if _, err := f.Write(append(data, '\n')); err != nil {
		return textEntry{}, err
	}
	return entry, nil
}

func listTextEntries(path string) []textEntry {
	data, err := os.ReadFile(path)
	if err != nil {
		return []textEntry{}
	}
	lines := strings.Split(string(data), "\n")
	result := make([]textEntry, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var entry textEntry
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			continue
		}
		if entry.Text == "" || entry.MTime <= 0 {
			continue
		}
		result = append(result, entry)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].MTime < result[j].MTime
	})
	return result
}

func pruneOldTextEntries(path string, maxAge time.Duration) int {
	entries := listTextEntries(path)
	if len(entries) == 0 {
		return 0
	}
	keep := make([]textEntry, 0, len(entries))
	now := time.Now()
	removed := 0
	for _, entry := range entries {
		if now.Sub(time.Unix(entry.MTime, 0)) > maxAge {
			removed++
			continue
		}
		keep = append(keep, entry)
	}
	if removed == 0 {
		return 0
	}
	_ = writeTextEntries(path, keep)
	return removed
}

func clearTextEntries(path string) int {
	entries := listTextEntries(path)
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return 0
	}
	return len(entries)
}

func writeTextEntries(path string, entries []textEntry) error {
	if len(entries) == 0 {
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			return err
		}
		return nil
	}
	lines := make([]string, 0, len(entries))
	for _, entry := range entries {
		data, err := json.Marshal(entry)
		if err != nil {
			continue
		}
		lines = append(lines, string(data))
	}
	payload := strings.Join(lines, "\n") + "\n"
	return os.WriteFile(path, []byte(payload), 0o644)
}

func buildTimeline(sharedFiles []fileEntry, uploadFiles []fileEntry, sharedTexts []textEntry, uploadTexts []textEntry) []timelineEntry {
	result := make([]timelineEntry, 0, len(sharedFiles)+len(uploadFiles)+len(sharedTexts)+len(uploadTexts))
	for _, file := range sharedFiles {
		result = append(result, timelineEntry{
			Bucket: "shared",
			Kind:   "file",
			Name:   file.Name,
			Size:   file.Size,
			MTime:  file.MTime,
		})
	}
	for _, file := range uploadFiles {
		result = append(result, timelineEntry{
			Bucket: "uploads",
			Kind:   "file",
			Name:   file.Name,
			Size:   file.Size,
			MTime:  file.MTime,
		})
	}
	for _, entry := range sharedTexts {
		result = append(result, timelineEntry{
			Bucket: "shared",
			Kind:   "text",
			Text:   entry.Text,
			MTime:  entry.MTime,
		})
	}
	for _, entry := range uploadTexts {
		result = append(result, timelineEntry{
			Bucket: "uploads",
			Kind:   "text",
			Text:   entry.Text,
			MTime:  entry.MTime,
		})
	}
	sort.SliceStable(result, func(i, j int) bool {
		return result[i].MTime < result[j].MTime
	})
	return result
}

func clearFiles(baseDir string) int {
	entries, err := os.ReadDir(baseDir)
	if err != nil {
		return 0
	}
	deleted := 0
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if err := os.Remove(filepath.Join(baseDir, entry.Name())); err == nil || os.IsNotExist(err) {
			deleted++
		}
	}
	return deleted
}
