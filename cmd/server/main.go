package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"

	"fleek-technical-test/internal/signal"
	"fleek-technical-test/internal/store"
)

func main() {
	ctx := signal.WatchInterrupt(context.Background(), 10*time.Second)

	var src, dst string
	var port int
	var jsonLog, help bool

	flag.StringVar(&src, "src", "", "source directory")
	flag.StringVar(&dst, "dst", "", "destination directory")
	flag.IntVar(&port, "p", 80, "http server port")
	flag.BoolVar(&jsonLog, "json", false, "format log in json")
	flag.BoolVar(&help, "h", false, "show this help")
	flag.Parse()

	log.SetOutput(os.Stdout)
	log.SetLevel(log.DebugLevel)

	if jsonLog {
		log.SetFormatter(&log.JSONFormatter{})
	}

	if help || src == "" || dst == "" {
		fmt.Println(os.Args[0], "-src {SRC_PATH} -dst {DST_PATH}")
		fmt.Println("Mirroring encryption server")
		flag.PrintDefaults()
		os.Exit(1)
	}

	if _, err := os.Stat(src); os.IsNotExist(err) {
		log.Fatalf("source directory does not exist: %v", err)
	}

	if _, err := os.Stat(dst); os.IsNotExist(err) {
		log.Fatalf("destination directory does not exist: %v", err)
	}

	memoryStore, err := store.NewMemoryStore()

	if err != nil {
		log.Fatal(err)
	}

	wg := sync.WaitGroup{}

	wg.Add(1)
	go func() {
		watchFileSystem(ctx, memoryStore, src, dst)
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		webServer(ctx, port, memoryStore, dst)
		wg.Done()
	}()

	wg.Wait()

	// TODO golang doc
}
