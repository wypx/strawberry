package task

import (
	"fmt"
	"log"
	"vm_manager/host_agent/src/modules"
	"vm_manager/vm_utils"
)

type ModifyAddressPoolExecutor struct {
	Sender         vm_utils.MessageSender
	ResourceModule modules.ResourceModule
}

func (executor *ModifyAddressPoolExecutor) Execute(id vm_utils.SessionID, request vm_utils.Message,
	incoming chan vm_utils.Message, terminate chan bool) (err error) {
	var config modules.AddressPoolConfig
	var poolName string
	if poolName, err = request.GetString(vm_utils.ParamKeyAddress); err != nil {
		return
	}
	config.Name = poolName
	if config.Gateway, err = request.GetString(vm_utils.ParamKeyGateway); err != nil {
		return
	}
	if config.DNS, err = request.GetStringArray(vm_utils.ParamKeyServer); err != nil {
		return
	}
	if config.Provider, err = request.GetString(vm_utils.ParamKeyMode); err != nil {
		err = fmt.Errorf("get provider fail: %s", err.Error())
		return
	}
	var respChan = make(chan modules.ResourceResult, 1)
	executor.ResourceModule.ModifyAddressPool(config, respChan)
	resp, _ := vm_utils.CreateJsonMessage(vm_utils.ModifyAddressPoolResponse)
	resp.SetFromSession(id)
	resp.SetToSession(request.GetFromSession())
	resp.SetSuccess(false)

	var result = <-respChan
	if result.Error != nil {
		err = result.Error
		resp.SetError(err.Error())
		log.Printf("[%08X] request modify address pool from %s.[%08X] fail: %s",
			id, request.GetSender(), request.GetFromSession(), err.Error())
		return executor.Sender.SendMessage(resp, request.GetSender())
	}
	resp.SetSuccess(true)
	log.Printf("[%08X] address pool '%s' modified from %s.[%08X]",
		id, poolName, request.GetSender(), request.GetFromSession())
	if err = executor.Sender.SendMessage(resp, request.GetSender()); err != nil {
		log.Printf("[%08X] warning: send modify address pool response fail: %s", id, err.Error())
	}
	if 0 != len(result.ComputeCellInfoList) {
		notify, _ := vm_utils.CreateJsonMessage(vm_utils.AddressPoolChangedEvent)
		notify.SetString(vm_utils.ParamKeyAddress, poolName)
		notify.SetString(vm_utils.ParamKeyGateway, config.Gateway)
		notify.SetString(vm_utils.ParamKeyMode, config.Provider)
		notify.SetStringArray(vm_utils.ParamKeyServer, config.DNS)
		notify.SetFromSession(id)
		for _, cell := range result.ComputeCellInfoList {
			if err = executor.Sender.SendMessage(notify, cell.Name); err != nil {
				log.Printf("[%08X] warning: notify address pool change to '%s' fail: %s", id, cell.Name, err.Error())
			}
		}
		log.Printf("[%08X] notified address pool changed to %d affected cell", id, len(result.ComputeCellInfoList))
	}
	return nil
}
