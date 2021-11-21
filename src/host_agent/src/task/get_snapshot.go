package task

import (
	"fmt"
	"log"
	"time"
	"vm_manager/host_agent/src/modules"
	"vm_manager/vm_utils"
)

type GetSnapshotExecutor struct {
	Sender         vm_utils.MessageSender
	ResourceModule modules.ResourceModule
}

func (executor *GetSnapshotExecutor) Execute(id vm_utils.SessionID, request vm_utils.Message,
	incoming chan vm_utils.Message, terminate chan bool) error {
	instanceID, err := request.GetString(vm_utils.ParamKeyInstance)
	if err != nil {
		return err
	}
	snapshot, err := request.GetString(vm_utils.ParamKeyName)
	if err != nil {
		return err
	}
	log.Printf("[%08X] request get snapshot '%s' of guest '%s' from %s.[%08X]", id, snapshot, instanceID,
		request.GetSender(), request.GetFromSession())

	var ins modules.InstanceStatus
	resp, _ := vm_utils.CreateJsonMessage(vm_utils.GetSnapshotResponse)
	resp.SetToSession(request.GetFromSession())
	resp.SetFromSession(id)
	resp.SetSuccess(false)
	if err = QualifySnapshotName(snapshot); err != nil {
		log.Printf("[%08X] invalid snapshot name '%s' : %s", id, snapshot, err.Error())
		err = fmt.Errorf("invalid snapshot name '%s': %s", snapshot, err.Error())
		resp.SetError(err.Error())
		return executor.Sender.SendMessage(resp, request.GetSender())
	}
	{
		var respChan = make(chan modules.ResourceResult)
		executor.ResourceModule.GetInstanceStatus(instanceID, respChan)
		result := <-respChan
		if result.Error != nil {
			log.Printf("[%08X] fetch instance fail: %s", id, result.Error.Error())
			resp.SetError(result.Error.Error())
			return executor.Sender.SendMessage(resp, request.GetSender())
		}
		ins = result.Instance
	}
	{
		//forward request
		forward, _ := vm_utils.CreateJsonMessage(vm_utils.GetSnapshotRequest)
		forward.SetFromSession(id)
		forward.SetString(vm_utils.ParamKeyInstance, instanceID)
		forward.SetString(vm_utils.ParamKeyName, snapshot)
		if err = executor.Sender.SendMessage(forward, ins.Cell); err != nil {
			log.Printf("[%08X] forward get snapshot to cell '%s' fail: %s", id, ins.Cell, err.Error())
			resp.SetError(err.Error())
			return executor.Sender.SendMessage(resp, request.GetSender())
		}
		timer := time.NewTimer(modules.DefaultOperateTimeout)
		select {
		case cellResp := <-incoming:
			if !cellResp.IsSuccess() {
				//	log.Printf("[%08X] cell get snapshot success", id)
				//}else{
				log.Printf("[%08X] cell get snapshot fail: %s", id, cellResp.GetError())
			}
			cellResp.SetFromSession(id)
			cellResp.SetToSession(request.GetFromSession())
			//forward
			return executor.Sender.SendMessage(cellResp, request.GetSender())
		case <-timer.C:
			//timeout
			log.Printf("[%08X] wait get snapshot response timeout", id)
			resp.SetError("cell timeout")
			return executor.Sender.SendMessage(resp, request.GetSender())
		}
	}
}
