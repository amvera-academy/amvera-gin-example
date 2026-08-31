package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
)

type Item struct {
	ID int `json:"id"`
	Name string `json:"name"`
}

var dataDir string
var dataFile string
var gate sync.Mutex

func readItems() ([]Item, error) {
	if err := os.MkdirAll(dataDir, 0755); err != nil { return nil, err }
	data, err := os.ReadFile(dataFile)
	if errors.Is(err, os.ErrNotExist) {
		if err = os.WriteFile(dataFile, []byte("[]"), 0644); err != nil { return nil, err }
		return []Item{}, nil
	}
	if err != nil { return nil, err }
	var items []Item
	err = json.Unmarshal(data, &items)
	return items, err
}

func writeItems(items []Item) error {
	data, err := json.MarshalIndent(items, "", "  ")
	if err != nil { return err }
	return os.WriteFile(dataFile, data, 0644)
}

func main() {
	dataDir = os.Getenv("DATA_DIR")
	if dataDir == "" {
		if os.Getenv("AMVERA") != "" { dataDir = "/data" } else { dataDir = "data" }
	}
	dataFile = filepath.Join(dataDir, "items.json")
	router := gin.Default()
	router.StaticFile("/", "static/index.html")
	router.StaticFile("/app.js", "static/app.js")
	router.StaticFile("/styles.css", "static/styles.css")
	router.GET("/api/health", func(context *gin.Context) { context.JSON(http.StatusOK, gin.H{"ok": true, "framework": "Gin", "storage": dataFile}) })
	router.GET("/api/items", func(context *gin.Context) {
		items, err := readItems()
		if err != nil { context.JSON(500, gin.H{"error": "Internal server error"}); return }
		for left, right := 0, len(items)-1; left < right; left, right = left+1, right-1 { items[left], items[right] = items[right], items[left] }
		context.JSON(200, gin.H{"items": items, "count": len(items)})
	})
	router.POST("/api/items", func(context *gin.Context) {
		var data struct { Name string `json:"name"` }
		if context.ShouldBindJSON(&data) != nil { context.JSON(400, gin.H{"error": "Invalid JSON"}); return }
		name := strings.TrimSpace(data.Name)
		if len(name) < 1 || len(name) > 120 { context.JSON(400, gin.H{"error": "Name must contain from 1 to 120 characters"}); return }
		gate.Lock(); defer gate.Unlock()
		items, err := readItems()
		if err != nil { context.JSON(500, gin.H{"error": "Internal server error"}); return }
		id := 1
		for _, item := range items { if item.ID >= id { id = item.ID + 1 } }
		item := Item{ID: id, Name: name}
		items = append(items, item)
		if writeItems(items) != nil { context.JSON(500, gin.H{"error": "Internal server error"}); return }
		context.JSON(201, gin.H{"item": item})
	})
	router.DELETE("/api/items/:id", func(context *gin.Context) {
		id, err := strconv.Atoi(context.Param("id"))
		if err != nil { context.JSON(400, gin.H{"error": "Invalid id"}); return }
		gate.Lock(); defer gate.Unlock()
		items, err := readItems()
		if err != nil { context.JSON(500, gin.H{"error": "Internal server error"}); return }
		next := make([]Item, 0, len(items))
		for _, item := range items { if item.ID != id { next = append(next, item) } }
		if len(next) == len(items) { context.JSON(404, gin.H{"error": "Item not found"}); return }
		if writeItems(next) != nil { context.JSON(500, gin.H{"error": "Internal server error"}); return }
		context.JSON(200, gin.H{"deleted": true, "id": id})
	})
	router.Run("0.0.0.0:5000")
}
