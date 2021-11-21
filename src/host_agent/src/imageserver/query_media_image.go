package imageserver

import (
	"log"
	"vm_manager/vm_utils"
)

type QueryMediaImageExecutor struct {
	Sender      vm_utils.MessageSender
	ImageServer *ImageManager
}

func (executor *QueryMediaImageExecutor) Execute(id vm_utils.SessionID, request vm_utils.Message,
	incoming chan vm_utils.Message, terminate chan bool) (err error) {
	var filterOwner, filterGroup string
	filterOwner, _ = request.GetString(vm_utils.ParamKeyUser)
	filterGroup, _ = request.GetString(vm_utils.ParamKeyGroup)

	var respChan = make(chan ImageResult, 1)
	executor.ImageServer.QueryMediaImage(filterOwner, filterGroup, respChan)

	var result = <-respChan

	resp, _ := vm_utils.CreateJsonMessage(vm_utils.QueryMediaImageResponse)
	resp.SetSuccess(false)
	resp.SetFromSession(id)
	resp.SetToSession(request.GetFromSession())

	if result.Error != nil {
		err = result.Error
		resp.SetError(err.Error())
		log.Printf("[%08X] query media image fail: %s", id, err.Error())
		return executor.Sender.SendMessage(resp, request.GetSender())
	}

	var name, imageID, description, tags, createTime, modifyTime []string
	var size, tagCount []uint64
	for _, image := range result.MediaList {
		name = append(name, image.Name)
		imageID = append(imageID, image.ID)
		description = append(description, image.Description)
		size = append(size, uint64(image.Size))
		count := uint64(len(image.Tags))
		tagCount = append(tagCount, count)
		for _, tag := range image.Tags {
			tags = append(tags, tag)
		}
		createTime = append(createTime, image.CreateTime)
		modifyTime = append(modifyTime, image.ModifyTime)
	}

	resp.SetSuccess(true)
	resp.SetStringArray(vm_utils.ParamKeyName, name)
	resp.SetStringArray(vm_utils.ParamKeyImage, imageID)
	resp.SetStringArray(vm_utils.ParamKeyDescription, description)
	resp.SetStringArray(vm_utils.ParamKeyTag, tags)
	resp.SetStringArray(vm_utils.ParamKeyCreate, createTime)
	resp.SetStringArray(vm_utils.ParamKeyModify, modifyTime)

	resp.SetUIntArray(vm_utils.ParamKeySize, size)
	resp.SetUIntArray(vm_utils.ParamKeyCount, tagCount)
	//log.Printf("[%08X] query media image success, %d image(s) available", id, len(result.MediaList))
	return executor.Sender.SendMessage(resp, request.GetSender())

}
