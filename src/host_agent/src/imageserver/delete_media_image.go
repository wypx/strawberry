package imageserver

import (
	"vm_manager/vm_utils"

	"log"
)

type DeleteMediaImageExecutor struct {
	Sender      vm_utils.MessageSender
	ImageServer *ImageManager
}

func (executor *DeleteMediaImageExecutor) Execute(id vm_utils.SessionID, request vm_utils.Message,
	incoming chan vm_utils.Message, terminate chan bool) (err error) {
	imageID, err := request.GetString(vm_utils.ParamKeyImage)
	if err != nil {
		return err
	}
	resp, _ := vm_utils.CreateJsonMessage(vm_utils.DeleteMediaImageResponse)
	resp.SetSuccess(false)
	resp.SetFromSession(id)
	resp.SetToSession(request.GetFromSession())

	var respChan = make(chan error, 1)
	executor.ImageServer.DeleteMediaImage(imageID, respChan)
	err = <-respChan
	if err != nil {
		resp.SetError(err.Error())
		log.Printf("[%08X] delete media image fail: %s", id, err.Error())
		return executor.Sender.SendMessage(resp, request.GetSender())
	}
	//deallocated
	resp.SetSuccess(true)
	log.Printf("[%08X] media image '%s' deleted", id, imageID)
	return executor.Sender.SendMessage(resp, request.GetSender())
}
