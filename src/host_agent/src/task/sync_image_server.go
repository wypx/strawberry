package task

import (
	"log"
	"net/http"
	"vm_manager/host_agent/src/modules"
	"vm_manager/vm_utils"
)

type SyncImageServerExecutor struct {
	Sender         vm_utils.MessageSender
	ResourceModule modules.ResourceModule
	Client         *http.Client
}

func (executor *SyncImageServerExecutor) Execute(id vm_utils.SessionID, request vm_utils.Message,
	incoming chan vm_utils.Message, terminate chan bool) (err error) {
	const (
		Protocol = "https"
	)
	var serverName, mediaHost string
	var mediaPort uint
	if serverName, err = request.GetString(vm_utils.ParamKeyName); err != nil {
		log.Printf("[%08X] sync image server fail: %s", id, err.Error())
		return nil
	}
	if mediaHost, err = request.GetString(vm_utils.ParamKeyHost); err != nil {
		log.Printf("[%08X] sync image server fail: %s", id, err.Error())
		return nil
	}
	if mediaPort, err = request.GetUInt(vm_utils.ParamKeyPort); err != nil {
		log.Printf("[%08X] sync image server fail: %s", id, err.Error())
		return nil
	}
	executor.ResourceModule.AddImageServer(serverName, mediaHost, int(mediaPort))
	log.Printf("[%08X] new imager server '%s' available (%s:%d)", id, serverName, mediaHost, mediaPort)

	return nil
}
