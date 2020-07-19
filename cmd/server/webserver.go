package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"fleek-technical-test/internal/fs"
	"fleek-technical-test/internal/store"
)

func webServer(ctx context.Context, port int, store store.Store, dst string) {
	mux := http.NewServeMux()

	h := &handler{
		store: store,
		dst:   dst,
	}

	mux.HandleFunc("/", index)
	mux.HandleFunc("/list", h.list)
	mux.HandleFunc("/file/", h.get)

	srv := &http.Server{
		Addr:    ":" + strconv.Itoa(port),
		Handler: mux,
	}

	log.Infof("http server start listening on port %d", port)

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("unable to start server: %v", err)
		}
	}()

	<-ctx.Done()

	ctxShutdown, _ := context.WithTimeout(context.Background(), 5*time.Second)

	if err := srv.Shutdown(ctxShutdown); err != nil {
		log.Fatal(err)
	}
}

func index(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./web/index.html")
}

type handler struct {
	store store.Store
	dst   string
}

type FileInfo struct {
	Hash string `json:"hash"`
	Name string `json:"name"`
	Size int64  `json:"size"`
	Key  string `json:"key"`
}

func (h *handler) list(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	files := make(map[string]*fs.File)
	h.store.All(func(key string, data interface{}) {
		files[key] = data.(*fs.File)
	})

	w.Header().Set("Content-Type", "application/json")

	var result struct {
		Files []FileInfo `json:"files"`
	}

	for _, file := range files {
		result.Files = append(result.Files, FileInfo{
			Hash: file.Hash(),
			Name: file.Name(),
			Size: file.Size(),
			Key:  base64.URLEncoding.EncodeToString(file.Key()),
		})
	}

	encoder := json.NewEncoder(w)
	_ = encoder.Encode(result)
}

func (h *handler) get(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")

	// hash is stored in the last part of the url
	hash := parts[len(parts)-1]

	// retrieve filename and key in query params
	filename := r.URL.Query().Get("filename")
	keyStr := r.URL.Query().Get("key")
	key, _ := base64.URLEncoding.DecodeString(keyStr)

	if keyStr == "" {
		//http.Error(w, "key is not provided", http.StatusBadRequest)
		http.ServeFile(w, r, "./web/key.html")
		return
	}

	// if filename is not provided, the filename will be the hash
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	w.Header().Set("Content-Type", "application/octet-stream")

	f, err := os.Open(h.dst + string(os.PathSeparator) + hash)

	if err != nil {
		// let's assume that the file is not found in case of error...
		http.Error(w, "file not found", http.StatusNotFound)
		return
	}

	defer f.Close()

	// try do decrypt the file with the key provided in query params
	if err = fs.EncryptDecrypt(key, f, w); err != nil {
		http.Error(w, "unable to decrypt file, key is maybe incorrect", http.StatusForbidden)
		return
	}

	// (optional) flush response writer just to be sure that the client receive correctly the content
	if wf, ok := w.(http.Flusher); ok {
		wf.Flush()
	}
}
