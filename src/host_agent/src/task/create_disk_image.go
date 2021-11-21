package task

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"
	"vm_manager/host_agent/src/modules"
	"vm_manager/vm_utils"
)

type CreateDiskImageExecutor struct {
	Sender         vm_utils.MessageSender
	ResourceModule modules.ResourceModule
	Client         *http.Client
}

func (executor *CreateDiskImageExecutor) Execute(id vm_utils.SessionID, request vm_utils.Message,
	incoming chan vm_utils.Message, terminate chan bool) error {

	var config modules.DiskImageConfig
	var guestID string
	var err error
	if config.Name, err = request.GetString(vm_utils.ParamKeyName); err != nil {
		return err
	}
	if guestID, err = request.GetString(vm_utils.ParamKeyGuest); err != nil {
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
	if guestID != "" {
		log.Printf("[%08X] request create disk image '%s' from guest '%s'", id, config.Name, guestID)
	} else {
		log.Printf("[%08X] request create disk image '%s' for uploading", id, config.Name)
	}

	resp, _ := vm_utils.CreateJsonMessage(vm_utils.CreateDiskImageResponse)
	resp.SetToSession(request.GetFromSession())
	resp.SetFromSession(id)
	resp.SetSuccess(false)
	if err = QualifyImageName(config.Name); err != nil {
		log.Printf("[%08X] invalid image name '%s' : %s", id, config.Name, err.Error())
		err = fmt.Errorf("invalid image name '%s': %s", config.Name, err.Error())
		resp.SetError(err.Error())
		return executor.Sender.SendMessage(resp, request.GetSender())
	}
	if err = QualifyNormalName(config.Owner); err != nil {
		log.Printf("[%08X] invalid owner name '%s' : %s", id, config.Owner, err.Error())
		err = fmt.Errorf("invalid owner name '%s': %s", config.Owner, err.Error())
		resp.SetError(err.Error())
		return executor.Sender.SendMessage(resp, request.GetSender())
	}
	if err = QualifyNormalName(config.Group); err != nil {
		log.Printf("[%08X] invalid group name '%s' : %s", id, config.Group, err.Error())
		err = fmt.Errorf("invalid group name '%s': %s", config.Group, err.Error())
		resp.SetError(err.Error())
		return executor.Sender.SendMessage(resp, request.GetSender())
	}

	var targetCell string
	if guestID != "" {
		//check guest
		var respChan = make(chan modules.ResourceResult)
		executor.ResourceModule.GetInstanceStatus(guestID, respChan)
		var result = <-respChan
		if result.Error != nil {
			err = result.Error
			log.Printf("[%08X] get instance status fail: %s", id, err.Error())
			resp.SetError(err.Error())
			return executor.Sender.SendMessage(resp, request.GetSender())
		}
		var status = result.Instance
		if !status.Created {
			err = fmt.Errorf("guest '%s' not created", guestID)
			log.Printf("[%08X] check guest status fail: %s", id, err.Error())
			resp.SetError(err.Error())
			return executor.Sender.SendMessage(resp, request.GetSender())
		}
		if status.Running {
			err = fmt.Errorf("guest '%s' still running", guestID)
			log.Printf("[%08X] check guest status fail: %s", id, err.Error())
			resp.SetError(err.Error())
			return executor.Sender.SendMessage(resp, request.GetSender())
		}
		targetCell = status.Cell
	}
	var imageServer, mediaHost string
	var mediaPort int
	{
		//get image server
		var respChan = make(chan modules.ResourceResult)
		executor.ResourceModule.GetImageServer(respChan)
		var result = <-respChan
		if result.Error != nil {
			err = result.Error
			log.Printf("[%08X] get image server fail: %s", id, err.Error())
			resp.SetError(err.Error())
			return executor.Sender.SendMessage(resp, request.GetSender())
		}
		imageServer = result.Name
		mediaHost = result.Host
		mediaPort = result.Port

	}
	var imageID string
	{
		var forwardCreate = vm_utils.CloneJsonMessage(request)
		forwardCreate.SetFromSession(id)
		forwardCreate.SetToSession(0)
		if err = executor.Sender.SendMessage(forwardCreate, imageServer); err != nil {
			log.Printf("[%08X] request create disk image to imageserver fail: %s", id, err.Error())
			resp.SetError(err.Error())
			return executor.Sender.SendMessage(resp, request.GetSender())
		}
		//wait response
		timer := time.NewTimer(modules.DefaultOperateTimeout)
		select {
		case forwardResp := <-incoming:
			if !forwardResp.IsSuccess() {
				log.Printf("[%08X] create disk image fail: %s", id, forwardResp.GetError())
				resp.SetError(forwardResp.GetError())
				return executor.Sender.SendMessage(resp, request.GetSender())
			}
			if imageID, err = forwardResp.GetString(vm_utils.ParamKeyImage); err != nil {
				log.Printf("[%08X] parse disk image ID fail: %s", id, forwardResp.GetError())
				resp.SetError(forwardResp.GetError())
				return executor.Sender.SendMessage(resp, request.GetSender())
			}
			log.Printf("[%08X] new disk image '%s'('%s') created at image server '%s'",
				id, config.Name, imageID, imageServer)

		case <-timer.C:
			//timeout
			log.Printf("[%08X] create disk image timeout", id)
			resp.SetError("time out")
			return executor.Sender.SendMessage(resp, request.GetSender())
		}

		if "" == guestID {
			//directly uploading
			resp.SetSuccess(true)
			resp.SetString(vm_utils.ParamKeyImage, imageID)
			return executor.Sender.SendMessage(resp, request.GetSender())
		}
	}
	{
		//request cell to transport data
		redirectedRequest, _ := vm_utils.CreateJsonMessage(vm_utils.CreateDiskImageRequest)
		redirectedRequest.SetFromSession(id)
		redirectedRequest.SetString(vm_utils.ParamKeyImage, imageID)
		redirectedRequest.SetString(vm_utils.ParamKeyGuest, guestID)
		redirectedRequest.SetString(vm_utils.ParamKeyHost, mediaHost)
		redirectedRequest.SetUInt(vm_utils.ParamKeyPort, uint(mediaPort))
		if err = executor.Sender.SendMessage(redirectedRequest, targetCell); err != nil {
			log.Printf("[%08X] redirect create request to cell '%s' fail: %s", id, targetCell, err.Error())
			resp.SetError(err.Error())
			executor.releaseResource(id, imageID, imageServer)
			return executor.Sender.SendMessage(resp, request.GetSender())
		}

		//wait for start
		timer := time.NewTimer(modules.DefaultOperateTimeout)
		select {
		case cellResp := <-incoming:
			if !cellResp.IsSuccess() {
				var errMsg = cellResp.GetError()
				log.Printf("[%08X] create disk image remotely fail: %s", id, errMsg)
				resp.SetError(errMsg)
				executor.releaseResource(id, imageID, imageServer)
				return executor.Sender.SendMessage(resp, request.GetSender())
			}
			log.Printf("[%08X] remote disk image creation started", id)
			resp.SetSuccess(true)
			resp.SetString(vm_utils.ParamKeyImage, imageID)
			if err = executor.Sender.SendMessage(resp, request.GetSender()); err != nil {
				log.Printf("[%08X] notify creation started fail: %s", id, err.Error())
				executor.releaseResource(id, imageID, imageServer)
				return err
			}

		case <-timer.C:
			err = fmt.Errorf("wait create response from cell '%s' timeout", targetCell)
			log.Printf("[%08X] wait create request response fail: %s", id, err.Error())
			resp.SetError(err.Error())
			executor.releaseResource(id, imageID, imageServer)
			return executor.Sender.SendMessage(resp, request.GetSender())
		}
	}
	{
		//keep waiting for progress

		var created bool
		var latestUpdate = time.Now()
		var checkTicker = time.NewTicker(1 * time.Second)
		for {
			select {
			case event := <-incoming:
				if !event.IsSuccess() {
					err = errors.New(event.GetError())
					log.Printf("[%08X] update progress fail: %s", id, err.Error())
					executor.releaseResource(id, imageID, imageServer)
					return err
				}
				if created, err = event.GetBoolean(vm_utils.ParamKeyEnable); err != nil {
					log.Printf("[%08X] parse event status fail: %s", id, err.Error())
					executor.releaseResource(id, imageID, imageServer)
					return err
				}
				if _, err = event.GetUInt(vm_utils.ParamKeyProgress); err != nil {
					log.Printf("[%08X] parse event progress fail: %s", id, err.Error())
					executor.releaseResource(id, imageID, imageServer)
					return err
				}
				event.SetString(vm_utils.ParamKeyImage, imageID)
				event.SetFromSession(id)
				event.SetToSession(0)
				if err = executor.Sender.SendMessage(event, imageServer); err != nil {
					log.Printf("[%08X] warning: forward disk image updated fail: %s", id, err.Error())
				}
				if created {
					//finished
					return nil
				} else {
					//unfinished
					latestUpdate = time.Now()
				}

			case <-checkTicker.C:
				//check
				if time.Now().Sub(latestUpdate) > modules.DefaultOperateTimeout {
					err = errors.New("wait update timeout")
					log.Printf("[%08X] wait create finish fail: %s", id, err.Error())
					executor.releaseResource(id, imageID, imageServer)
					return err
				}
			}
		}
	}
}

func (executor *CreateDiskImageExecutor) releaseResource(id vm_utils.SessionID, imageID, imageServer string) {
	if imageID != "" {
		delete, _ := vm_utils.CreateJsonMessage(vm_utils.DeleteDiskImageRequest)
		delete.SetString(vm_utils.ParamKeyImage, imageID)
		delete.SetFromSession(id)
		delete.SetToSession(0)
		if err := executor.Sender.SendMessage(delete, imageServer); err != nil {
			log.Printf("[%08X] warning: request delete disk image to imageserver fail: %s", id, err.Error())
			return
		}
		log.Printf("[%08X] try release disk image '%s' to imageserver '%s'", id, imageID, imageServer)
	}
}
