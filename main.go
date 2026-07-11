package main

import (
	"context"
	"flag"
	"log"

	"github.com/amrutp24/terraform-provider-fireworks/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
)

// Regenerate docs/ from the provider schema and examples/ (CI fails on drift).
//go:generate go tool tfplugindocs generate --provider-name fireworks --rendered-provider-name Fireworks

// Set at build time by GoReleaser.
var version = "dev"

func main() {
	var debug bool
	flag.BoolVar(&debug, "debug", false, "run the provider with support for debuggers like delve")
	flag.Parse()

	opts := providerserver.ServeOpts{
		Address: "registry.terraform.io/amrutp24/fireworks",
		Debug:   debug,
	}

	if err := providerserver.Serve(context.Background(), provider.New(version), opts); err != nil {
		log.Fatal(err.Error())
	}
}
