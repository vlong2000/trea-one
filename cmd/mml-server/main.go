package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"mml-server/internal/mapping"
	"mml-server/internal/server"
)

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	fs := flag.NewFlagSet("mml-server", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	host := fs.String("host", "0.0.0.0", "listen host")
	port := fs.Int("port", 2323, "listen port")
	mappingPath := fs.String("mapping", "configs/mapping.txt", "mapping file path")
	if err := fs.Parse(args); err != nil {
		return 1
	}
	if *port < 1 || *port > 65535 {
		fmt.Fprintf(os.Stderr, "invalid --port %d: must be 1-65535\n", *port)
		return 1
	}

	table, err := mapping.Load(*mappingPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "load mapping: %v\n", err)
		return 1
	}

	addr := net.JoinHostPort(*host, strconv.Itoa(*port))
	srv := server.New(addr, table)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	log.Printf("listening on %s mapping=%s", addr, *mappingPath)
	if err := srv.ListenAndServe(ctx); err != nil && !errors.Is(err, context.Canceled) {
		fmt.Fprintf(os.Stderr, "listen: %v\n", err)
		return 1
	}
	return 0
}
