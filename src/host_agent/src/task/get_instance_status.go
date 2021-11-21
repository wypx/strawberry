package task

import (
	"fmt"
	"log"
	"time"
	"vm_manager/host_agent/src/modules"
	"vm_manager/vm_utils"
)

type GetInstanceStatusExecutor struct {
	Sender         vm_utils.MessageSender
	ResourceModule modules.ResourceModule
}

func (executor *GetInstanceStatusExecutor) Execute(id vm_utils.SessionID, request vm_utils.Message,
	incoming chan vm_utils.Message, terminate chan bool) error {
	instanceID, err := request.GetString(vm_utils.ParamKeyInstance)
	if err != nil {
		return err
	}

	//log.Printf("[%08X] request get instance '%s' status from %s.[%08X]", id, instanceID,
	//	request.GetSender(), request.GetFromSession())

	var ins modules.InstanceStatus

	resp, _ := vm_utils.CreateJsonMessage(vm_utils.GetInstanceStatusResponse)
	resp.SetToSession(request.GetFromSession())
	resp.SetFromSession(id)
	resp.SetSuccess(false)

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
	var fromSession = request.GetFromSession()
	{
		//redirect request
		request.SetFromSession(id)
		if err = executor.Sender.SendMessage(request, ins.Cell); err != nil {
			log.Printf("[%08X] redirect get instance to cell '%s' fail: %s", id, ins.Cell, err.Error())
			resp.SetError(err.Error())
			return executor.Sender.SendMessage(resp, request.GetSender())
		}
		timer := time.NewTimer(modules.DefaultOperateTimeout)
		select {
		case cellResp := <-incoming:
			if cellResp.IsSuccess() {
				//log.Printf("[%08X] cell get instance status success", id)
				//modify network info
				var internalMonitor = fmt.Sprintf("%s:%d", ins.InternalNetwork.MonitorAddress, ins.InternalNetwork.MonitorPort)
				var externalMonitor = fmt.Sprintf("%s:%d", ins.ExternalNetwork.MonitorAddress, ins.ExternalNetwork.MonitorPort)
				cellResp.SetStringArray(vm_utils.ParamKeyMonitor, []string{internalMonitor, externalMonitor})
				cellResp.SetStringArray(vm_utils.ParamKeyAddress, []string{ins.InternalNetwork.InstanceAddress, ins.ExternalNetwork.InstanceAddress})

			} else {
				log.Printf("[%08X] cell get instance status  fail: %s", id, cellResp.GetError())
			}
			cellResp.SetFromSession(id)
			cellResp.SetToSession(fromSession)
			//forward
			return executor.Sender.SendMessage(cellResp, request.GetSender())
		case <-timer.C:
			//timeout
			log.Printf("[%08X] wait query response timeout", id)
			resp.SetError("cell timeout")
			return executor.Sender.SendMessage(resp, request.GetSender())
		}
	}
}
