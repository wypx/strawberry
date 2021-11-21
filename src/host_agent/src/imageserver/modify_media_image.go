package imageserver

import (
	"log"
	"vm_manager/vm_utils"
)

type ModifyMediaImageExecutor struct {
	Sender      vm_utils.MessageSender
	ImageServer *ImageManager
}

func (executor *ModifyMediaImageExecutor) Execute(id vm_utils.SessionID, request vm_utils.Message,
	incoming chan vm_utils.Message, terminate chan bool) (err error) {
	var imageID string
	if imageID, err = request.GetString(vm_utils.ParamKeyImage); err != nil {
		return
	}
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
	var respChan = make(chan error, 1)
	executor.ImageServer.ModifyMediaImage(imageID, config, respChan)
	err = <-respChan

	resp, _ := vm_utils.CreateJsonMessage(vm_utils.ModifyMediaImageResponse)
	resp.SetSuccess(false)
	resp.SetFromSession(id)
	resp.SetToSession(request.GetFromSession())

	if err != nil {
		resp.SetError(err.Error())
		log.Printf("[%08X] modify media image fail: %s", id, err.Error())
		return executor.Sender.SendMessage(resp, request.GetSender())
	}
	log.Printf("[%08X] media image '%s' modified", id, imageID)
	resp.SetSuccess(true)
	return executor.Sender.SendMessage(resp, request.GetSender())
}
