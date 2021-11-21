package task

import (
	"log"
	"vm_manager/host_agent/src/modules"
	"vm_manager/vm_utils"
)

type HandleCellStatusUpdatedExecutor struct {
	ResourceModule modules.ResourceModule
}

func (executor *HandleCellStatusUpdatedExecutor) Execute(id vm_utils.SessionID, event vm_utils.Message,
	incoming chan vm_utils.Message, terminate chan bool) error {
	var usage = modules.CellStatusReport{}
	if err := usage.FromMessage(event); err != nil {
		log.Printf("handle cell usage fail: %s", err.Error())
		return err
	}
	executor.ResourceModule.UpdateCellStatus(usage)
	return nil
}
