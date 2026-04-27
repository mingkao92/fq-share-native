package main

import (
	"regexp"
	"sync"
	"time"
)

const (
	hostName       = "com.mingkao.fqshare"
	defaultPort    = 18500
	maxPort        = 18600
	maxFileAge     = 24 * time.Hour
	healthWaitTime = 5 * time.Second
)

var legacyHostNames = []string{}

var invalidNameChars = regexp.MustCompile(`[\x00-\x1f\x7f/\\:]+`)
var validTokenPattern = regexp.MustCompile(`^[A-Za-z0-9_-]{16,128}$`)

type appPaths struct {
	selfPath    string
	baseDir     string
	runtimeDir  string
	dataDir     string
	statePath   string
	httpLogPath string
	hostLogPath string
}

type serverState struct {
	PID       int    `json:"pid"`
	Port      int    `json:"port"`
	Token     string `json:"token"`
	LANIP     string `json:"lan_ip"`
	StartedAt int64  `json:"started_at"`
}

type accessStats struct {
	mu            sync.Mutex
	pageVisits    int
	lastPageVisit int64
}

func (a *accessStats) markPageVisit() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.pageVisits++
	a.lastPageVisit = time.Now().Unix()
}

func (a *accessStats) snapshot() (int, int64) {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.pageVisits, a.lastPageVisit
}

type httpConfig struct {
	token         string
	dataDir       string
	sessionsDir   string
	publicBaseURL string
}

type httpHandler struct {
	cfg       httpConfig
	sessionMu sync.Mutex
	statsByTK map[string]*accessStats
}

type fileEntry struct {
	Name  string `json:"name"`
	Size  int64  `json:"size"`
	MTime int64  `json:"mtime"`
}

type textEntry struct {
	ID    string `json:"id"`
	Text  string `json:"text"`
	MTime int64  `json:"mtime"`
}

type timelineEntry struct {
	Bucket string `json:"bucket"`
	Kind   string `json:"kind"`
	Name   string `json:"name,omitempty"`
	Text   string `json:"text,omitempty"`
	Size   int64  `json:"size,omitempty"`
	MTime  int64  `json:"mtime"`
}
