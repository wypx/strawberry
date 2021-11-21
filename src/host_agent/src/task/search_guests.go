package task

import (
	"log"
	"vm_manager/host_agent/src/modules"
	"vm_manager/vm_utils"
)

type SearchGuestsExecutor struct {
	Sender         vm_utils.MessageSender
	ResourceModule modules.ResourceModule
}

func (executor *SearchGuestsExecutor) Execute(id vm_utils.SessionID, request vm_utils.Message,
	incoming chan vm_utils.Message, terminate chan bool) error {
	var err error
	var condition modules.SearchGuestsCondition
	if condition.Pool, err = request.GetString(vm_utils.ParamKeyPool); err != nil {
		return err
	}
	if condition.Cell, err = request.GetString(vm_utils.ParamKeyCell); err != nil {
		return err
	}
	if condition.Keyword, err = request.GetString(vm_utils.ParamKeyData); err != nil {
		return err
	}
	if condition.Limit, err = request.GetInt(vm_utils.ParamKeyLimit); err != nil {
		return err
	}
	if condition.Offset, err = request.GetInt(vm_utils.ParamKeyStart); err != nil {
		return err
	}

	resp, _ := vm_utils.CreateJsonMessage(vm_utils.SearchGuestResponse)
	resp.SetSuccess(false)
	resp.SetFromSession(id)
	resp.SetToSession(request.GetFromSession())

	{
		var respChan = make(chan modules.ResourceResult, 1)
		executor.ResourceModule.SearchGuests(condition, respChan)
		result := <-respChan
		if result.Error != nil {
			err = result.Error
			log.Printf("[%08X] search guests fail: %s", id, err.Error())
			resp.SetError(err.Error())
			return executor.Sender.SendMessage(resp, request.GetSender())
		}
		var guests = result.InstanceList
		if err = modules.MarshalInstanceStatusListToMessage(guests, resp); err != nil {
			log.Printf("[%08X] build response message for search guests result fail: %s", id, err.Error())
			resp.SetError(err.Error())
			return executor.Sender.SendMessage(resp, request.GetSender())
		}
		var flags = []uint64{uint64(result.Total), uint64(result.Limit), uint64(result.Offset)}
		resp.SetUIntArray(vm_utils.ParamKeyFlag, flags)
		//log.Printf("[%08X] %d guest(s) available", id, len(guests))
		resp.SetSuccess(true)
		return executor.Sender.SendMessage(resp, request.GetSender())
	}
}
