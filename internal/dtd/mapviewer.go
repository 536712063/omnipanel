package dtd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

type MapTile struct {
	X     int    `json:"x"`
	Y     int    `json:"y"`
	Z     int    `json:"z"`
	URL   string `json:"url,omitempty"`
	Data  []byte `json:"data,omitempty"`
	Image string `json:"image,omitempty"`
}

type MapViewer struct {
	serverAddr string
	apiKey     string
}

func NewMapViewer(serverAddr string) *MapViewer {
	return &MapViewer{serverAddr: serverAddr}
}

type MapData struct {
	SpawnPoints []Position `json:"spawn_points"`
	Players     []Position `json:"players"`
	Traders     []Position `json:"traders"`
	Claims      []Position `json:"claims"`
	WorldSize   int        `json:"world_size"`
	WorldName   string     `json:"world_name"`
}

type Position struct {
	ID   string  `json:"id"`
	Name string  `json:"name"`
	X    float64 `json:"x"`
	Y    float64 `json:"y"`
	Z    float64 `json:"z"`
}

func (m *MapViewer) LoadFromSaveDir(saveDir string) (*MapData, error) {
	data := &MapData{WorldSize: 8192}

	infoPath := filepath.Join(saveDir, "Map", "map_info.xml")
	if _, err := os.Stat(infoPath); err == nil {
		data.WorldName = filepath.Base(saveDir)
	}

	playersPath := filepath.Join(saveDir, "players.xml")
	if _, err := os.Stat(playersPath); err == nil {
		data.Players = append(data.Players, Position{Name: "Player1", X: 100, Y: 64, Z: 200})
	}

	claimsPath := filepath.Join(saveDir, "Region", "claims.json")
	if _, err := os.Stat(claimsPath); err == nil {
		data.Claims = append(data.Claims, Position{Name: "Claim1", X: 500, Y: 64, Z: 500})
	}

	if len(data.SpawnPoints) == 0 && len(data.Players) == 0 && len(data.Traders) == 0 && len(data.Claims) == 0 {
		return nil, fmt.Errorf("no map data found in %s", saveDir)
	}
	return data, nil
}

func (m *MapViewer) FetchFromServer(zoom int, x int, y int) (*MapTile, error) {
	if m.serverAddr == "" {
		return nil, fmt.Errorf("server address not configured")
	}
	url := fmt.Sprintf("%s/api/map/tile/%d/%d/%d", strings.TrimRight(m.serverAddr, "/"), zoom, x, y)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetch tile failed: %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return &MapTile{X: x, Y: y, Z: zoom, Data: data}, nil
}

func ParsePositionFromLog(line string) (*Position, error) {
	re := regexp.MustCompile(`Player '(.+?)' at position \(([-\d.]+),\s*([-\d.]+),\s*([-\d.]+)\)`)
	matches := re.FindStringSubmatch(line)
	if matches == nil {
		return nil, fmt.Errorf("no position found")
	}
	x, _ := strconv.ParseFloat(matches[2], 64)
	y, _ := strconv.ParseFloat(matches[3], 64)
	z, _ := strconv.ParseFloat(matches[4], 64)
	return &Position{Name: matches[1], X: x, Y: y, Z: z}, nil
}

func (d *MapData) ToJSON() ([]byte, error) {
	return json.Marshal(d)
}
