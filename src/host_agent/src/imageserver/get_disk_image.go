package imageserver

import (
	"vm_manager/vm_utils"

	"log"
)

type GetDiskImageExecutor struct {
	Sender      vm_utils.MessageSender
	ImageServer *ImageManager
}

func (executor *GetDiskImageExecutor) Execute(id vm_utils.SessionID, request vm_utils.Message,
	incoming chan vm_utils.Message, terminate chan bool) (err error) {
	imageID, err := request.GetString(vm_utils.ParamKeyImage)
	if err != nil {
		return err
	}
	var respChan = make(chan ImageResult, 1)
	executor.ImageServer.GetDiskImage(imageID, respChan)
	var result = <-respChan
	resp, _ := vm_utils.CreateJsonMessage(vm_utils.GetDiskImageResponse)
	resp.SetSuccess(false)
	resp.SetFromSession(id)
	resp.SetToSession(request.GetFromSession())

	if result.Error != nil {
		err := result.Error
		log.Printf("[%08X] get disk image fail: %s", id, err.Error())
		resp.SetError(err.Error())
		return executor.Sender.SendMessage(resp, request.GetSender())
	}
	var image = result.DiskImage
	resp.SetSuccess(true)
	resp.SetString(vm_utils.ParamKeyName, image.Name)
	resp.SetString(vm_utils.ParamKeyDescription, image.Description)
	resp.SetStringArray(vm_utils.ParamKeyTag, image.Tags)
	resp.SetString(vm_utils.ParamKeyUser, image.Owner)
	resp.SetString(vm_utils.ParamKeyGroup, image.Group)

	resp.SetUInt(vm_utils.ParamKeySize, uint(image.Size))
	resp.SetUInt(vm_utils.ParamKeyProgress, image.Progress)

	resp.SetBoolean(vm_utils.ParamKeyEnable, image.Created)
	return executor.Sender.SendMessage(resp, request.GetSender())
}
