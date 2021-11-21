package task

import (
	"fmt"
	"log"
	"vm_manager/host_agent/src/modules"
	"vm_manager/vm_utils"
)

type DeleteSystemTemplateExecutor struct {
	Sender         vm_utils.MessageSender
	ResourceModule modules.ResourceModule
}

func (executor *DeleteSystemTemplateExecutor) Execute(id vm_utils.SessionID, request vm_utils.Message,
	incoming chan vm_utils.Message, terminate chan bool) (err error) {

	var templateID string
	if templateID, err = request.GetString(vm_utils.ParamKeyTemplate); err != nil {
		err = fmt.Errorf("get template id fail: %s", err.Error())
		return
	}
	var respChan = make(chan error, 1)
	executor.ResourceModule.DeleteSystemTemplate(templateID, respChan)
	err = <-respChan

	resp, _ := vm_utils.CreateJsonMessage(vm_utils.DeleteTemplateResponse)
	resp.SetSuccess(false)
	resp.SetFromSession(id)
	resp.SetToSession(request.GetFromSession())

	if err != nil {
		resp.SetError(err.Error())
		log.Printf("[%08X] handle delete system template from %s.[%08X] fail: %s",
			id, request.GetSender(), request.GetFromSession(), err.Error())
	} else {
		resp.SetSuccess(true)
	}
	return executor.Sender.SendMessage(resp, request.GetSender())
}
