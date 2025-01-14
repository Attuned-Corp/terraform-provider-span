package main

import (
	"context"
	"flag"
	"log"

	"github.com/attuned-corp/terraform-provider-span/span"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
)

var (
	// goreleaser inserts this at build time.
	Version = "dev"
)

func main() {
	var debug bool

	flag.BoolVar(&debug, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	opts := providerserver.ServeOpts{
		Address: "registry.terraform.io/attuned-corp/span",
		Debug:   debug,
	}

	err := providerserver.Serve(context.Background(), span.NewProviderFactory(Version), opts)

	if err != nil {
		log.Fatal(err.Error())
	}
}
