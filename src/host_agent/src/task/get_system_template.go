package task

import (
	"fmt"
	"log"
	"vm_manager/host_agent/src/modules"
	"vm_manager/vm_utils"
)

type GetSystemTemplateExecutor struct {
	Sender         vm_utils.MessageSender
	ResourceModule modules.ResourceModule
}

func (executor *GetSystemTemplateExecutor) Execute(id vm_utils.SessionID, request vm_utils.Message,
	incoming chan vm_utils.Message, terminate chan bool) (err error) {
	var templateID string
	if templateID, err = request.GetString(vm_utils.ParamKeyTemplate); err != nil {
		err = fmt.Errorf("get template id fail: %s", err.Error())
		return
	}
	var respChan = make(chan modules.ResourceResult, 1)
	executor.ResourceModule.GetSystemTemplate(templateID, respChan)
	var result = <-respChan

	resp, _ := vm_utils.CreateJsonMessage(vm_utils.GetTemplateResponse)
	resp.SetSuccess(false)
	resp.SetFromSession(id)
	resp.SetToSession(request.GetFromSession())

	if result.Error != nil {
		err = result.Error
		resp.SetError(err.Error())
		log.Printf("[%08X] handle get system template from %s.[%08X] fail: %s",
			id, request.GetSender(), request.GetFromSession(), err.Error())
	} else {
		var t = result.Template
		resp.SetSuccess(true)
		resp.SetString(vm_utils.ParamKeyID, t.ID)
		resp.SetString(vm_utils.ParamKeyName, t.Name)
		resp.SetString(vm_utils.ParamKeyAdmin, t.Admin)
		resp.SetString(vm_utils.ParamKeySystem, t.OperatingSystem)
		resp.SetString(vm_utils.ParamKeyDisk, t.Disk)
		resp.SetString(vm_utils.ParamKeyNetwork, t.Network)
		resp.SetString(vm_utils.ParamKeyDisplay, t.Display)
		resp.SetString(vm_utils.ParamKeyMonitor, t.Control)
		resp.SetString(vm_utils.ParamKeyDevice, t.USB)
		resp.SetString(vm_utils.ParamKeyInterface, t.Tablet)
		resp.SetString(vm_utils.ParamKeyCreate, t.CreatedTime)
		resp.SetString(vm_utils.ParamKeyModify, t.ModifiedTime)
	}
	return executor.Sender.SendMessage(resp, request.GetSender())
}
