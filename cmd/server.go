package main

import (
	"log"

	"github.com/JoeQiao666/inventory"
	"github.com/cloudstateio/go-support/cloudstate"
	"github.com/cloudstateio/go-support/cloudstate/eventsourced"
	"github.com/cloudstateio/go-support/cloudstate/protocol"
)

func main() {
	server, err := cloudstate.New(protocol.Config{
		ServiceName:    "inventory.inventory",
		ServiceVersion: "0.2.0",
	})
	if err != nil {
		log.Fatalf("cloudstate.New failed: %s", err)
	}
	err = server.RegisterEventSourced(&eventsourced.Entity{
		ServiceName:   "inventory.Inventory",
		PersistenceID: "Inventory",
		EntityFunc:    inventory.NewInventory,
	}, protocol.DescriptorConfig{
		Service: "service.proto",
	}.AddDomainDescriptor("domain.proto"))
	if err != nil {
		log.Fatalf("CloudState failed to register entity: %s", err)
	}
	err = server.Run()
	if err != nil {
		log.Fatalf("Cloudstate failed to run: %v", err)
	}
}
