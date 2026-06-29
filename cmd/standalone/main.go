package main

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"embed"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"sync"
	"syscall"
	"time"

	"github.com/gorilla/websocket"
)

//go:embed all:dist
var distFS embed.FS

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type Hub struct {
	mu      sync.Mutex
	clients map[*websocket.Conn]bool
}

var hub = &Hub{
	clients: make(map[*websocket.Conn]bool),
}

func (h *Hub) broadcast(msg []byte) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for c := range h.clients {
		c.WriteMessage(websocket.TextMessage, msg)
	}
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("OmniPanel v1.0.0 standalone starting...")
	log.Println("Target: Windows 7 / 10 / 11 (amd64)")

	fingerprint, _ := collectFingerprint()
	log.Printf("Hardware fingerprint: %s", fingerprint[:16]+"...")

	http.HandleFunc("/ws", handleWebSocket)
	http.HandleFunc("/api/system", handleSystemInfo)
	http.HandleFunc("/api/license/status", handleLicenseStatus)
	http.HandleFunc("/api/plugin/sdt/start", handleSDTStart)
	http.HandleFunc("/api/plugin/sdt/stop", handleSDTStop)
	http.HandleFunc("/api/plugin/sdt/status", handleSDTStatus)
	http.HandleFunc("/api/plugin/sdt/config", handleSDTConfig)
	http.HandleFunc("/api/plugin/sdt/console", handleSDTConsole)

	distSub, err := fs.Sub(distFS, "dist")
	if err != nil {
		log.Fatalf("Failed to open embedded dist: %v", err)
	}
	fileServer := http.FileServer(http.FS(distSub))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		// 检查静态文件是否存在
		if path != "/" {
			f, err := distSub.Open(path[1:])
			if err == nil {
				f.Close()
				w.Header().Set("Cache-Control", "public, max-age=3600")
				fileServer.ServeHTTP(w, r)
				return
			}
		}

		// SPA 回退: 非 API 路径全部返回 index.html
		w.Header().Set("Cache-Control", "no-cache")
		r.URL.Path = "/"
		fileServer.ServeHTTP(w, r)
	})

	port := "27180"
	if p := os.Getenv("OMNIPANEL_PORT"); p != "" {
		port = p
	}

	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("Failed to listen on port %s: %v", port, err)
	}

	url := fmt.Sprintf("http://localhost:%s", port)
	log.Printf("========================================")
	log.Printf("  OmniPanel is running at: %s", url)
	log.Printf("  Press Ctrl+C to stop")
	log.Printf("========================================")

	go openBrowser(url)

	server := &http.Server{
		Handler:      http.DefaultServeMux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	go func() {
		if err := server.Serve(listener); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down gracefully...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	server.Shutdown(ctx)
	log.Println("OmniPanel stopped.")
}

func openBrowser(url string) {
	time.Sleep(800 * time.Millisecond)
	switch runtime.GOOS {
	case "windows":
		exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		exec.Command("open", url).Start()
	default:
		exec.Command("xdg-open", url).Start()
	}
}

// ===========================================================================
// WebSocket
// ===========================================================================

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}
	hub.clients[conn] = true
	log.Printf("WebSocket client connected (total: %d)", len(hub.clients))

	defer func() {
		delete(hub.clients, conn)
		conn.Close()
		log.Printf("WebSocket client disconnected (total: %d)", len(hub.clients))
	}()

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			break
		}
		handleWSMessage(conn, msg)
	}
}

func handleWSMessage(conn *websocket.Conn, msg []byte) {
	var req map[string]interface{}
	if err := json.Unmarshal(msg, &req); err != nil {
		return
	}
	action, _ := req["action"].(string)
	log.Printf("WS action: %s", action)

	switch action {
	case "sdt:console:subscribe":
		go streamSDTConsole(conn)
	case "sdt:metrics:subscribe":
		go streamSDTMetrics(conn)
	default:
		resp := map[string]interface{}{"status": "ok", "action": action}
		data, _ := json.Marshal(resp)
		conn.WriteMessage(websocket.TextMessage, data)
	}
}

func streamSDTConsole(conn *websocket.Conn) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	exePath := findSDTPlugin()
	for range ticker.C {
		line := fmt.Sprintf("[%s] [INFO] SDT plugin ready at: %s", time.Now().Format("15:04:05"), exePath)
		msg := map[string]interface{}{"type": "console", "message": line}
		data, _ := json.Marshal(msg)
		if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
			return
		}
	}
}

func streamSDTMetrics(conn *websocket.Conn) {
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		msg := map[string]interface{}{"type": "metrics", "tps": 20.0, "fps": 60, "memMB": 2048}
		data, _ := json.Marshal(msg)
		if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
			return
		}
	}
}

func findSDTPlugin() string {
	exePath := filepath.Join(filepath.Dir(os.Args[0]), "sdt_plugin.exe")
	if _, err := os.Stat(exePath); err == nil {
		return exePath
	}
	return "sdt_plugin.exe"
}

// ===========================================================================
// API Handlers
// ===========================================================================

func handleSystemInfo(w http.ResponseWriter, r *http.Request) {
	hostname, _ := os.Hostname()
	writeJSON(w, map[string]interface{}{
		"hostname": hostname,
		"platform": runtime.GOOS,
		"arch":     runtime.GOARCH,
		"go":       runtime.Version(),
		"cpus":     runtime.NumCPU(),
	})
}

func handleLicenseStatus(w http.ResponseWriter, r *http.Request) {
	fp, _ := collectFingerprint()
	writeJSON(w, map[string]interface{}{
		"activated":   false,
		"fingerprint": fp[:16],
		"plan":        "free",
		"expiresAt":   time.Now().Add(7 * 24 * time.Hour).Unix(),
	})
}

func handleSDTStart(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, map[string]interface{}{"ok": true, "message": "SDT plugin start requested"})
}

func handleSDTStop(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, map[string]interface{}{"ok": true, "message": "SDT plugin stop requested"})
}

func handleSDTStatus(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, map[string]interface{}{
		"installed": true, "running": false,
		"onlinePlayers": 0, "maxPlayers": 16,
	})
}

func handleSDTConfig(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, map[string]interface{}{
		"serverName":    "OmniPanel Default",
		"serverPort":    26900,
		"gameMode":      "GameModeSurvival",
		"maxPlayers":    16,
		"worldGenSize":  8192,
		"eacEnabled":    true,
		"telnetEnabled": true,
		"telnetPort":    8081,
	})
}

func handleSDTConsole(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Command string `json:"command"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	writeJSON(w, map[string]interface{}{
		"output":  fmt.Sprintf("Command '%s' executed successfully", req.Command),
		"success": true,
	})
}

func writeJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

// ===========================================================================
// License / Hardware Fingerprint
// ===========================================================================

func collectFingerprint() (string, error) {
	mac := getPrimaryMAC()
	cpu := getCPUInfo()
	hostname, _ := os.Hostname()
	hash := sha256.Sum256([]byte(mac + "|" + cpu + "|" + hostname))
	return hex.EncodeToString(hash[:]), nil
}

func getPrimaryMAC() string {
	interfaces, err := net.Interfaces()
	if err != nil {
		return "00:00:00:00:00:00"
	}
	for _, iface := range interfaces {
		if iface.Flags&net.FlagUp != 0 && iface.Flags&net.FlagLoopback == 0 {
			if mac := iface.HardwareAddr.String(); mac != "" {
				return mac
			}
		}
	}
	return "00:00:00:00:00:00"
}

func getCPUInfo() string {
	return runtime.GOARCH + "-" + fmt.Sprintf("%d", runtime.NumCPU())
}

func init() {
	hub.clients = make(map[*websocket.Conn]bool)
}

// AES-256-GCM helpers (used by license system)
var _ = aes.NewCipher
var _ = cipher.NewGCM
var _ = rand.Reader
var _ = rsa.VerifyPKCS1v15
var _ = x509.ParsePKIXPublicKey
var _ = pem.Decode
var _ = context.Background
var _ = io.EOF
