package task

import (
	"errors"
	"fmt"
	"log"
	"time"
	"vm_manager/host_agent/src/modules"
	"vm_manager/vm_utils"
)

type ModifyGuestNameExecutor struct {
	Sender         vm_utils.MessageSender
	ResourceModule modules.ResourceModule
}

func (executor *ModifyGuestNameExecutor) Execute(id vm_utils.SessionID, request vm_utils.Message,
	incoming chan vm_utils.Message, terminate chan bool) error {
	guestID, err := request.GetString(vm_utils.ParamKeyGuest)
	if err != nil {
		return err
	}
	name, err := request.GetString(vm_utils.ParamKeyName)
	if err != nil {
		return err
	}
	log.Printf("[%08X] request rename guest '%s' from %s.[%08X]", id, guestID,
		request.GetSender(), request.GetFromSession())

	var ins modules.InstanceStatus
	resp, _ := vm_utils.CreateJsonMessage(vm_utils.ModifyCoreResponse)
	resp.SetToSession(request.GetFromSession())
	resp.SetFromSession(id)
	resp.SetSuccess(false)
	var poolName string
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
		if ins.Running {
			err = fmt.Errorf("guest '%s' is still running", ins.Name)
			log.Printf("[%08X] %s", id, err.Error())
			resp.SetError(err.Error())
			return executor.Sender.SendMessage(resp, request.GetSender())
		}
		if ins.Name == name {
			err = errors.New("no need to change")
			log.Printf("[%08X] %s", id, err.Error())
			resp.SetError(err.Error())
			return executor.Sender.SendMessage(resp, request.GetSender())
		}
		poolName = ins.Pool
	}
	{
		//check new name
		var respChan = make(chan modules.ResourceResult)
		executor.ResourceModule.GetInstanceByName(poolName, name, respChan)
		var result = <-respChan
		if result.Error == nil {
			err = fmt.Errorf("instance '%s' already exists in pool '%s'", name, poolName)
			log.Printf("[%08X] %s", id, err.Error())
			resp.SetError(err.Error())
			return executor.Sender.SendMessage(resp, request.GetSender())
		}
	}
	{
		//forward rename request
		forward, _ := vm_utils.CreateJsonMessage(vm_utils.ModifyGuestNameRequest)
		forward.SetFromSession(id)
		forward.SetString(vm_utils.ParamKeyGuest, guestID)
		forward.SetString(vm_utils.ParamKeyName, name)
		if err = executor.Sender.SendMessage(forward, ins.Cell); err != nil {
			log.Printf("[%08X] forward modify name to cell '%s' fail: %s", id, ins.Cell, err.Error())
			resp.SetError(err.Error())
			return executor.Sender.SendMessage(resp, request.GetSender())
		}
		timer := time.NewTimer(modules.DefaultOperateTimeout)
		select {
		case cellResp := <-incoming:
			if cellResp.IsSuccess() {
				//update
				var respChan = make(chan error)
				executor.ResourceModule.RenameInstance(guestID, name, respChan)
				err = <-respChan
				if err != nil {
					log.Printf("[%08X] update new new fail: %s", id, err.Error())
					resp.SetError(err.Error())
					return executor.Sender.SendMessage(resp, request.GetSender())
				}
				log.Printf("[%08X] modify guest name success", id)
			} else {
				log.Printf("[%08X] modify guest name fail: %s", id, cellResp.GetError())
			}
			cellResp.SetFromSession(id)
			cellResp.SetToSession(request.GetFromSession())
			//forward
			return executor.Sender.SendMessage(cellResp, request.GetSender())
		case <-timer.C:
			//timeout
			log.Printf("[%08X] wait modify guest name response timeout", id)
			resp.SetError("request timeout")
			return executor.Sender.SendMessage(resp, request.GetSender())
		}
	}
}
