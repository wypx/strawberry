package task

import (
	"fmt"
	"log"
	"time"
	"vm_manager/host_agent/src/modules"
	"vm_manager/vm_utils"
)

type InsertMediaExecutor struct {
	Sender         vm_utils.MessageSender
	ResourceModule modules.ResourceModule
}

func (executor *InsertMediaExecutor) Execute(id vm_utils.SessionID, request vm_utils.Message,
	incoming chan vm_utils.Message, terminate chan bool) error {
	instanceID, err := request.GetString(vm_utils.ParamKeyInstance)
	if err != nil {
		return err
	}
	mediaSource, err := request.GetString(vm_utils.ParamKeyMedia)
	if err != nil {
		return err
	}

	log.Printf("[%08X] request insert media '%s' into guest '%s' from %s.[%08X]", id, mediaSource, instanceID,
		request.GetSender(), request.GetFromSession())

	var ins modules.InstanceStatus
	resp, _ := vm_utils.CreateJsonMessage(vm_utils.InsertMediaResponse)
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
		if !ins.Running {
			err = fmt.Errorf("instance '%s' is stopped", instanceID)
			log.Printf("[%08X] instance '%s' is stopped", id, instanceID)
			resp.SetError(err.Error())
			return executor.Sender.SendMessage(resp, request.GetSender())
		}
	}
	{
		//forward request
		forward, _ := vm_utils.CreateJsonMessage(vm_utils.InsertMediaRequest)
		forward.SetFromSession(id)
		forward.SetString(vm_utils.ParamKeyInstance, instanceID)
		forward.SetString(vm_utils.ParamKeyMedia, mediaSource)

		//todo: get media name for display
		var respChan = make(chan modules.ResourceResult)
		executor.ResourceModule.GetImageServer(respChan)
		var result = <-respChan
		if result.Error != nil {
			errMsg := result.Error.Error()
			log.Printf("[%08X] get image server fail: %s", id, errMsg)
			resp.SetError(errMsg)
			return executor.Sender.SendMessage(resp, request.GetSender())
		}
		log.Printf("[%08X] select image server '%s:%d' for image '%s'", id, result.Host, result.Port, mediaSource)
		forward.SetString(vm_utils.ParamKeyHost, result.Host)
		forward.SetUInt(vm_utils.ParamKeyPort, uint(result.Port))

		if err = executor.Sender.SendMessage(forward, ins.Cell); err != nil {
			log.Printf("[%08X] forward insert media to cell '%s' fail: %s", id, ins.Cell, err.Error())
			resp.SetError(err.Error())
			return executor.Sender.SendMessage(resp, request.GetSender())
		}
		timer := time.NewTimer(modules.DefaultOperateTimeout)
		select {
		case cellResp := <-incoming:
			if cellResp.IsSuccess() {
				log.Printf("[%08X] cell insert media success", id)
			} else {
				log.Printf("[%08X] cell insert media fail: %s", id, cellResp.GetError())
			}
			cellResp.SetFromSession(id)
			cellResp.SetToSession(request.GetFromSession())
			//forward
			return executor.Sender.SendMessage(cellResp, request.GetSender())
		case <-timer.C:
			//timeout
			log.Printf("[%08X] wait insert response timeout", id)
			resp.SetError("cell timeout")
			return executor.Sender.SendMessage(resp, request.GetSender())
		}
	}
}
