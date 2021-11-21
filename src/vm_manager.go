package main

import (
	"log"
	HostAgent "vm_manager/host_agent"
	VmAdmin "vm_manager/vm_admin"
	VmAgent "vm_manager/vm_agent"
)

func main() {
	log.Printf("vm manager start\n")
	VmAdmin.Initialize()
	VmAgent.Initialize()
	HostAgent.Initialize()
	// host_agent.Initialize()
	// vm_admin.Initialize()
	log.Printf("vm manager exit\n")
}
