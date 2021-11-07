package main

import (
	"log"
	VmAdmin "vm_manager/vm_admin"
	VmAgent "vm_manager/vm_agent"
)

func main() {
	log.Printf("vm manager start\n")
	VmAdmin.Initialize()
	VmAgent.Initialize()
	// host_agent.Initialize()
	// vm_admin.Initialize()
	log.Printf("vm manager exit\n")
}
