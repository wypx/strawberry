package imageserver

import (
	"log"
	"vm_manager/vm_utils"
)

type DeleteDiskImageExecutor struct {
	Sender      vm_utils.MessageSender
	ImageServer *ImageManager
}

func (executor *DeleteDiskImageExecutor) Execute(id vm_utils.SessionID, request vm_utils.Message,
	incoming chan vm_utils.Message, terminate chan bool) (err error) {
	imageID, err := request.GetString(vm_utils.ParamKeyImage)
	if err != nil {
		return err
	}
	resp, _ := vm_utils.CreateJsonMessage(vm_utils.DeleteDiskImageResponse)
	resp.SetSuccess(false)
	resp.SetFromSession(id)
	resp.SetToSession(request.GetFromSession())

	var respChan = make(chan error, 1)
	executor.ImageServer.DeleteDiskImage(imageID, respChan)
	err = <-respChan
	if err != nil {
		resp.SetError(err.Error())
		log.Printf("[%08X] delete disk image fail: %s", id, err.Error())
		return executor.Sender.SendMessage(resp, request.GetSender())
	}
	resp.SetSuccess(true)
	log.Printf("[%08X] disk image '%s' deleted", id, imageID)
	return executor.Sender.SendMessage(resp, request.GetSender())
}
