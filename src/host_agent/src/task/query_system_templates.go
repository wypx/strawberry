package task

import (
	"log"
	"vm_manager/host_agent/src/modules"
	"vm_manager/vm_utils"
)

type QuerySystemTemplatesExecutor struct {
	Sender         vm_utils.MessageSender
	ResourceModule modules.ResourceModule
}

func (executor *QuerySystemTemplatesExecutor) Execute(id vm_utils.SessionID, request vm_utils.Message,
	incoming chan vm_utils.Message, terminate chan bool) (err error) {
	var respChan = make(chan modules.ResourceResult, 1)
	executor.ResourceModule.QuerySystemTemplates(respChan)
	var result = <-respChan

	resp, _ := vm_utils.CreateJsonMessage(vm_utils.QueryTemplateResponse)
	resp.SetSuccess(false)
	resp.SetFromSession(id)
	resp.SetToSession(request.GetFromSession())

	if result.Error != nil {
		err = result.Error
		resp.SetError(err.Error())
		log.Printf("[%08X] handle query system templates from %s.[%08X] fail: %s",
			id, request.GetSender(), request.GetFromSession(), err.Error())
	} else {
		var idList, nameList, osList, createList, modifiedList []string
		for _, t := range result.TemplateList {
			idList = append(idList, t.ID)
			nameList = append(nameList, t.Name)
			osList = append(osList, t.OperatingSystem)
			createList = append(createList, t.CreatedTime)
			modifiedList = append(modifiedList, t.ModifiedTime)
		}
		resp.SetSuccess(true)
		resp.SetStringArray(vm_utils.ParamKeyID, idList)
		resp.SetStringArray(vm_utils.ParamKeyName, nameList)
		resp.SetStringArray(vm_utils.ParamKeySystem, osList)
		resp.SetStringArray(vm_utils.ParamKeyCreate, createList)
		resp.SetStringArray(vm_utils.ParamKeyModify, modifiedList)
	}

	return executor.Sender.SendMessage(resp, request.GetSender())
}
