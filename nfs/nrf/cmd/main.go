package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/vlong2000/trea-one/nfs/nrf/pkg/factory"
	"github.com/vlong2000/trea-one/nfs/nrf/pkg/service"
)

func main() {
	cfgPath := flag.String("c", factory.DefaultConfigPath, "NRF config file")
	flag.StringVar(cfgPath, "config", factory.DefaultConfigPath, "NRF config file")
	flag.Parse()

	if err := service.Run(*cfgPath); err != nil {
		fmt.Fprintf(os.Stderr, "nrf: %v\n", err)
		os.Exit(1)
	}
}
