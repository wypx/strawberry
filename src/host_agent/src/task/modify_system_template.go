package task

import (
	"fmt"
	"log"
	"vm_manager/host_agent/src/modules"
	"vm_manager/vm_utils"
)

type ModifySystemTemplateExecutor struct {
	Sender         vm_utils.MessageSender
	ResourceModule modules.ResourceModule
}

func (executor *ModifySystemTemplateExecutor) Execute(id vm_utils.SessionID, request vm_utils.Message,
	incoming chan vm_utils.Message, terminate chan bool) (err error) {
	var templateID string
	if templateID, err = request.GetString(vm_utils.ParamKeyTemplate); err != nil {
		err = fmt.Errorf("get template id fail: %s", err.Error())
		return
	}
	var config modules.SystemTemplateConfig
	if config.Name, err = request.GetString(vm_utils.ParamKeyName); err != nil {
		err = fmt.Errorf("get template name fail: %s", err.Error())
		return
	}
	if config.Admin, err = request.GetString(vm_utils.ParamKeyAdmin); err != nil {
		err = fmt.Errorf("get admin name fail: %s", err.Error())
		return
	}
	if config.OperatingSystem, err = request.GetString(vm_utils.ParamKeySystem); err != nil {
		err = fmt.Errorf("get oprating system fail: %s", err.Error())
		return
	}
	if config.Disk, err = request.GetString(vm_utils.ParamKeyDisk); err != nil {
		err = fmt.Errorf("get disk option fail: %s", err.Error())
		return
	}
	if config.Network, err = request.GetString(vm_utils.ParamKeyNetwork); err != nil {
		err = fmt.Errorf("get network option fail: %s", err.Error())
		return
	}
	if config.Display, err = request.GetString(vm_utils.ParamKeyDisplay); err != nil {
		err = fmt.Errorf("get display option fail: %s", err.Error())
		return
	}
	if config.Control, err = request.GetString(vm_utils.ParamKeyMonitor); err != nil {
		err = fmt.Errorf("get control option fail: %s", err.Error())
		return
	}
	if config.USB, err = request.GetString(vm_utils.ParamKeyDevice); err != nil {
		err = fmt.Errorf("get usb option fail: %s", err.Error())
		return
	}
	if config.Tablet, err = request.GetString(vm_utils.ParamKeyInterface); err != nil {
		err = fmt.Errorf("get tablet option fail: %s", err.Error())
		return
	}

	var respChan = make(chan error, 1)
	executor.ResourceModule.ModifySystemTemplate(templateID, config, respChan)
	err = <-respChan

	resp, _ := vm_utils.CreateJsonMessage(vm_utils.ModifyTemplateResponse)
	resp.SetSuccess(false)
	resp.SetFromSession(id)
	resp.SetToSession(request.GetFromSession())

	if err != nil {
		resp.SetError(err.Error())
		log.Printf("[%08X] handle modify system template from %s.[%08X] fail: %s",
			id, request.GetSender(), request.GetFromSession(), err.Error())
	} else {
		resp.SetSuccess(true)
	}
	return executor.Sender.SendMessage(resp, request.GetSender())
}
