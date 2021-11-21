package imageserver

import (
	"vm_manager/vm_utils"

	"log"
)

type CreateDiskImageExecutor struct {
	Sender      vm_utils.MessageSender
	ImageServer *ImageManager
}

func (executor *CreateDiskImageExecutor) Execute(id vm_utils.SessionID, request vm_utils.Message,
	incoming chan vm_utils.Message, terminate chan bool) (err error) {
	var config ImageConfig
	if config.Name, err = request.GetString(vm_utils.ParamKeyName); err != nil {
		return err
	}
	if config.Owner, err = request.GetString(vm_utils.ParamKeyUser); err != nil {
		return err
	}
	if config.Group, err = request.GetString(vm_utils.ParamKeyGroup); err != nil {
		return err
	}
	if config.Description, err = request.GetString(vm_utils.ParamKeyDescription); err != nil {
		return err
	}
	if config.Tags, err = request.GetStringArray(vm_utils.ParamKeyTag); err != nil {
		return err
	}
	var respChan = make(chan ImageResult, 1)
	executor.ImageServer.CreateDiskImage(config, respChan)
	result := <-respChan

	resp, _ := vm_utils.CreateJsonMessage(vm_utils.CreateDiskImageResponse)
	resp.SetSuccess(false)
	resp.SetFromSession(id)
	resp.SetToSession(request.GetFromSession())

	if result.Error != nil {
		err = result.Error
		resp.SetError(err.Error())
		log.Printf("[%08X] create disk image fail: %s", id, err.Error())
		return executor.Sender.SendMessage(resp, request.GetSender())
	}
	log.Printf("[%08X] new disk image '%s' created(id '%s')", id, config.Name, result.ID)
	resp.SetString(vm_utils.ParamKeyImage, result.ID)
	resp.SetSuccess(true)
	return executor.Sender.SendMessage(resp, request.GetSender())
}
