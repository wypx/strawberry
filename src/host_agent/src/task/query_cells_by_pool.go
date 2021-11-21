package task

import (
	"log"
	"vm_manager/host_agent/src/modules"
	"vm_manager/vm_utils"
)

type QueryCellsByPoolExecutor struct {
	Sender         vm_utils.MessageSender
	ResourceModule modules.ResourceModule
}

func (executor *QueryCellsByPoolExecutor) Execute(id vm_utils.SessionID, request vm_utils.Message,
	incoming chan vm_utils.Message, terminate chan bool) error {
	poolName, err := request.GetString(vm_utils.ParamKeyPool)
	if err != nil {
		return err
	}
	//log.Printf("[%08X] query cells by pool from %s.[%08X]", id, request.GetSender(), request.GetFromSession())
	var respChan = make(chan modules.ResourceResult)
	executor.ResourceModule.QueryCellsInPool(poolName, respChan)
	result := <-respChan
	resp, _ := vm_utils.CreateJsonMessage(vm_utils.QueryComputePoolResponse)
	resp.SetSuccess(false)
	resp.SetFromSession(id)
	resp.SetToSession(request.GetFromSession())

	if result.Error != nil {
		resp.SetError(result.Error.Error())
		log.Printf("[%08X] query cells fail: %s", id, result.Error.Error())
		return executor.Sender.SendMessage(resp, request.GetSender())
	}

	//log.Printf("[%08X] %d cells available in pool '%s'", id, len(result.ComputeCellInfoList), poolName)
	resp.SetSuccess(true)
	modules.CellsToMessage(resp, result.ComputeCellInfoList)

	return executor.Sender.SendMessage(resp, request.GetSender())
}
