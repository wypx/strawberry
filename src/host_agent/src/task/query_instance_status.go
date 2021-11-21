package task

import (
	"log"
	"vm_manager/host_agent/src/modules"
	"vm_manager/vm_utils"
)

type QueryInstanceStatusExecutor struct {
	Sender         vm_utils.MessageSender
	ResourceModule modules.ResourceModule
}

func (executor *QueryInstanceStatusExecutor) Execute(id vm_utils.SessionID, request vm_utils.Message,
	incoming chan vm_utils.Message, terminate chan bool) (err error) {
	poolName, err := request.GetString(vm_utils.ParamKeyPool)
	if err != nil {
		return err
	}
	var inCell = false
	cellName, err := request.GetString(vm_utils.ParamKeyCell)
	if err == nil {
		inCell = true
	}
	var respChan = make(chan modules.ResourceResult)
	if inCell {
		//log.Printf("[%08X] request query instance status in cell '%s' from %s.[%08X]", id, cellName,
		//	request.GetSender(), request.GetFromSession())
		executor.ResourceModule.QueryInstanceStatusInCell(poolName, cellName, respChan)
	} else {
		//log.Printf("[%08X] request query instance status in pool '%s' from %s.[%08X]", id, poolName,
		//	request.GetSender(), request.GetFromSession())
		executor.ResourceModule.QueryInstanceStatusInPool(poolName, respChan)
	}
	result := <-respChan

	resp, _ := vm_utils.CreateJsonMessage(vm_utils.QueryInstanceStatusResponse)
	resp.SetToSession(request.GetFromSession())
	resp.SetFromSession(id)
	resp.SetSuccess(false)
	if result.Error != nil {
		err = result.Error
		log.Printf("[%08X] query instance status fail: %s", id, err.Error())
		resp.SetError(err.Error())
		return executor.Sender.SendMessage(resp, request.GetSender())
	}

	var instances = result.InstanceList
	modules.MarshalInstanceStatusListToMessage(instances, resp)
	resp.SetSuccess(true)
	//log.Printf("[%08X] %d instance(s) available", id, len(instances))
	return executor.Sender.SendMessage(resp, request.GetSender())
}
