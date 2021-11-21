package imageserver

import (
	"fmt"
	"log"
	"vm_manager/vm_utils"
)

type SyncDiskImagesExecutor struct {
	Sender      vm_utils.MessageSender
	ImageServer *ImageManager
}

func (executor *SyncDiskImagesExecutor) Execute(id vm_utils.SessionID, request vm_utils.Message,
	incoming chan vm_utils.Message, terminate chan bool) (err error) {
	var owner, group string
	if owner, err = request.GetString(vm_utils.ParamKeyUser); err != nil {
		err = fmt.Errorf("get owner fail: %s", err.Error())
		return err
	}
	if group, err = request.GetString(vm_utils.ParamKeyGroup); err != nil {
		err = fmt.Errorf("get group fail: %s", err.Error())
		return err
	}
	log.Printf("[%08X] %s.[%08X] request synchronize disk images...",
		id, request.GetSender(), request.GetFromSession())
	var respChan = make(chan error, 1)
	executor.ImageServer.SyncDiskImages(owner, group, respChan)
	err = <-respChan

	resp, _ := vm_utils.CreateJsonMessage(vm_utils.SynchronizeDiskImageResponse)
	resp.SetSuccess(false)
	resp.SetFromSession(id)
	resp.SetToSession(request.GetFromSession())

	if err != nil {
		resp.SetError(err.Error())
		log.Printf("[%08X] sync disk images fail: %s", id, err.Error())
	} else {
		log.Printf("[%08X] disk images synchronized", id)
		resp.SetSuccess(true)
	}
	return executor.Sender.SendMessage(resp, request.GetSender())
}
