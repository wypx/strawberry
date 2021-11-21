package task

import (
	"errors"
	"fmt"
	"log"
	"time"
	"vm_manager/host_agent/src/modules"
	"vm_manager/vm_utils"
)

type ResetMonitorSecretExecutor struct {
	Sender         vm_utils.MessageSender
	ResourceModule modules.ResourceModule
}

func (executor *ResetMonitorSecretExecutor) Execute(id vm_utils.SessionID, request vm_utils.Message,
	incoming chan vm_utils.Message, terminate chan bool) (err error) {
	var guestID string
	if guestID, err = request.GetString(vm_utils.ParamKeyGuest); err != nil {
		err = fmt.Errorf("get guest id fail: %s", err.Error())
		return err
	}

	var ins modules.InstanceStatus
	resp, _ := vm_utils.CreateJsonMessage(vm_utils.ResetSecretResponse)
	resp.SetToSession(request.GetFromSession())
	resp.SetFromSession(id)
	resp.SetSuccess(false)

	{
		var respChan = make(chan modules.ResourceResult)
		executor.ResourceModule.GetInstanceStatus(guestID, respChan)
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
			log.Printf("[%08X] redirect reset request to cell '%s' fail: %s", id, ins.Cell, err.Error())
			resp.SetError(err.Error())
			return executor.Sender.SendMessage(resp, request.GetSender())
		}
		timer := time.NewTimer(modules.DefaultOperateTimeout)
		select {
		case cellResp := <-incoming:
			if cellResp.IsSuccess() {
				var newSecret string
				if newSecret, err = cellResp.GetString(vm_utils.ParamKeySecret); err != nil {
					err = fmt.Errorf("get new secret from response fail: %s", err.Error())
				} else {
					var respChan = make(chan error, 1)
					executor.ResourceModule.UpdateInstanceMonitorSecret(guestID, newSecret, respChan)
					err = <-respChan
					if err == nil {
						log.Printf("[%08X] monitor secret of guest '%s' reset", id, ins.Name)
					}
				}
			} else {
				err = errors.New(cellResp.GetError())
			}
			if err != nil {
				log.Printf("[%08X] cell reset monitor secret fail: %s", id, cellResp.GetError())
			} else {
				cellResp.SetSuccess(true)
			}
			cellResp.SetFromSession(id)
			cellResp.SetToSession(fromSession)
			//forward
			return executor.Sender.SendMessage(cellResp, request.GetSender())
		case <-timer.C:
			//timeout
			log.Printf("[%08X] wait reset response timeout", id)
			resp.SetError("cell timeout")
			return executor.Sender.SendMessage(resp, request.GetSender())
		}
	}
}
