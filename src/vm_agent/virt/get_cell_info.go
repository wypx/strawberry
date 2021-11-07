package virt

import (
	"log"
	VmAgentSvc "vm_manager/vm_agent/svc"
	VmUtils "vm_manager/vm_utils"
)

type GetCellInfoExecutor struct {
	Sender         VmUtils.MessageSender
	InstanceModule VmAgentSvc.InstanceModule
	StorageModule  VmAgentSvc.StorageModule
	NetworkModule  VmAgentSvc.NetworkModule
}

func (executor *GetCellInfoExecutor) Execute(id VmUtils.SessionID, request VmUtils.Message,
	incoming chan VmUtils.Message, terminate chan bool) (err error) {

	//todo: add instance/network info
	resp, _ := VmUtils.CreateJsonMessage(VmUtils.GetComputePoolCellResponse)
	resp.SetToSession(request.GetFromSession())
	resp.SetFromSession(id)
	resp.SetSuccess(false)

	{
		//storage
		var respChan = make(chan VmAgentSvc.StorageResult, 1)
		executor.StorageModule.GetAttachDevices(respChan)
		var result = <-respChan
		if result.Error != nil {
			err = result.Error
			log.Printf("[%08X] fetch attach device fail: %s", id, err.Error())
			resp.SetError(err.Error())
			return executor.Sender.SendMessage(resp, request.GetSender())
		}
		var names, errMessages []string
		var attached []uint64
		for _, device := range result.Devices {
			names = append(names, device.Name)
			errMessages = append(errMessages, device.Error)
			if device.Attached {
				attached = append(attached, 1)
			} else {
				attached = append(attached, 0)
			}
		}
		resp.SetStringArray(VmUtils.ParamKeyStorage, names)
		resp.SetStringArray(VmUtils.ParamKeyError, errMessages)
		resp.SetUIntArray(VmUtils.ParamKeyAttach, attached)
		log.Printf("[%08X] %d device(s) available", id, len(names))
	}
	resp.SetSuccess(true)
	return executor.Sender.SendMessage(resp, request.GetSender())
}
