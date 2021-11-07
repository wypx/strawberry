package virt

import (
	"fmt"
	"log"
	VmAgentSvc "vm_manager/vm_agent/svc"
	VmUtils "vm_manager/vm_utils"
)

type DeleteInstanceExecutor struct {
	Sender         VmUtils.MessageSender
	InstanceModule VmAgentSvc.InstanceModule
	StorageModule  VmAgentSvc.StorageModule
	NetworkModule  VmAgentSvc.NetworkModule
}

func (executor *DeleteInstanceExecutor) Execute(id VmUtils.SessionID, request VmUtils.Message,
	incoming chan VmUtils.Message, terminate chan bool) (err error) {
	var instanceID string
	instanceID, err = request.GetString(VmUtils.ParamKeyInstance)
	if err != nil {
		return err
	}
	log.Printf("[%08X] request delete instance '%s' from %s.[%08X]", id,
		instanceID, request.GetSender(), request.GetFromSession())

	resp, _ := VmUtils.CreateJsonMessage(VmUtils.DeleteGuestResponse)
	resp.SetSuccess(false)
	resp.SetFromSession(id)
	resp.SetToSession(request.GetFromSession())

	var config VmAgentSvc.GuestConfig
	{
		var respChan = make(chan VmAgentSvc.InstanceResult)
		executor.InstanceModule.GetInstanceStatus(instanceID, respChan)
		result := <-respChan
		if result.Error != nil {
			log.Printf("[%08X] get config fail: %s", id, result.Error.Error())
			return executor.ResponseToFail(request.GetSender(), resp, result.Error)
		}
		if result.Instance.Running {
			err := fmt.Errorf("instance '%s' still running", instanceID)
			log.Printf("[%08X] delete instance fail: %s", id, err.Error())
			return executor.ResponseToFail(request.GetSender(), resp, err)
		}
		config = result.Instance.GuestConfig
	}
	{
		//todo: detach network
		switch config.NetworkMode {
		case VmAgentSvc.NetworkModePlain:
			var respChan = make(chan error)
			executor.NetworkModule.DeallocateAllResource(instanceID, respChan)
			err := <-respChan
			if err != nil {
				log.Printf("[%08X] release network resource fail: %s", id, err.Error())
				return executor.ResponseToFail(request.GetSender(), resp, err)
			}
			log.Printf("[%08X] network resource released", id)
			break
		default:
			return fmt.Errorf("unsupported network mode %d", config.NetworkMode)
		}
	}
	{
		//delete guest config
		var respChan = make(chan error)
		executor.InstanceModule.DeleteInstance(instanceID, respChan)
		err := <-respChan
		if err != nil {
			log.Printf("[%08X] delete instance fail: %s", id, err)
			return executor.ResponseToFail(request.GetSender(), resp, err)
		}
		log.Printf("[%08X] instance deleted", id)
	}
	{
		//delete volumes
		switch config.StorageMode {
		case VmAgentSvc.StorageModeLocal:
			{
				var respChan = make(chan error)
				executor.StorageModule.DeleteVolumes(instanceID, respChan)
				err := <-respChan
				if err != nil {
					log.Printf("[%08X] delete volumes fail: %s", id, err)
					return executor.ResponseToFail(request.GetSender(), resp, err)
				}
				log.Printf("[%08X] disk volumes deleted", id)
			}
		default:
			return fmt.Errorf("unsupported storage mode %d", config.StorageMode)
		}
	}
	resp.SetSuccess(true)
	log.Printf("[%08X] delete finish, all resource released", id)
	if err = executor.Sender.SendMessage(resp, request.GetSender()); err != nil {
		log.Printf("[%08X] warning: send response fail: %s", id, err.Error())
		return err
	}
	event, _ := VmUtils.CreateJsonMessage(VmUtils.GuestDeletedEvent)
	event.SetFromSession(id)
	event.SetString(VmUtils.ParamKeyInstance, instanceID)
	if err = executor.Sender.SendMessage(event, request.GetSender()); err != nil {
		log.Printf("[%08X] warning: notify instance deleted fail: %s", id, err.Error())
		return err
	}
	return nil
}

func (executor *DeleteInstanceExecutor) ResponseToFail(target string, resp VmUtils.Message, err error) error {
	resp.SetSuccess(false)
	resp.SetError(err.Error())
	return executor.Sender.SendMessage(resp, target)
}
