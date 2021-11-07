package task

import (
	"log"
	"github.com/project-nano/framework"
	"github.com/project-nano/core/modules"
	"fmt"
	"time"
)

type ResizeGuestDiskExecutor struct {
	Sender         framework.MessageSender
	ResourceModule modules.ResourceModule
}

func (executor *ResizeGuestDiskExecutor)Execute(id framework.SessionID, request framework.Message,
	incoming chan framework.Message, terminate chan bool) error {
	guestID, err := request.GetString(framework.ParamKeyGuest)
	if err != nil{
		return err
	}
	index, err := request.GetUInt(framework.ParamKeyDisk)
	if err != nil{
		return err
	}
	var diskIndex = int(index)
	diskSize, err := request.GetUInt(framework.ParamKeySize)
	if err != nil{
		return err
	}
	log.Printf("[%08X] request resize disk of '%s' from %s.[%08X]", id, guestID,
		request.GetSender(), request.GetFromSession())

	var ins modules.InstanceStatus
	resp, _ := framework.CreateJsonMessage(framework.ResizeDiskResponse)
	resp.SetToSession(request.GetFromSession())
	resp.SetFromSession(id)
	resp.SetSuccess(false)
	{
		var respChan = make(chan modules.ResourceResult)
		executor.ResourceModule.GetInstanceStatus(guestID, respChan)
		result := <- respChan
		if result.Error != nil{
			log.Printf("[%08X] fetch instance fail: %s", id, result.Error.Error())
			resp.SetError(result.Error.Error())
			return executor.Sender.SendMessage(resp, request.GetSender())
		}
		ins = result.Instance
		if diskIndex >= len(ins.Disks){
			err = fmt.Errorf("invalid disk index %d", diskIndex)
			log.Printf("[%08X] %s", id, err.Error())
			resp.SetError(err.Error())
			return executor.Sender.SendMessage(resp, request.GetSender())
		}
		if ins.Disks[diskIndex] >= uint64(diskSize) {
			err = fmt.Errorf("target size must larger than %d GiB", diskSize >> 30)
			log.Printf("[%08X] %s", id, err.Error())
			resp.SetError(err.Error())
			return executor.Sender.SendMessage(resp, request.GetSender())
		}
	}
	{
		//request delete
		forward, _ := framework.CreateJsonMessage(framework.ResizeDiskRequest)
		forward.SetFromSession(id)
		forward.SetString(framework.ParamKeyGuest, guestID)
		forward.SetUInt(framework.ParamKeySize, diskSize)
		forward.SetUInt(framework.ParamKeyDisk, index)
		if err = executor.Sender.SendMessage(forward, ins.Cell); err != nil{
			log.Printf("[%08X] forward resize disk to cell '%s' fail: %s", id, ins.Cell, err.Error())
			resp.SetError(err.Error())
			return executor.Sender.SendMessage(resp, request.GetSender())
		}
		timer := time.NewTimer(modules.DefaultOperateTimeout)
		select{
		case cellResp := <- incoming:
			if cellResp.IsSuccess(){
				ins.Disks[diskIndex] = uint64(diskSize)
				//update
				var respChan = make(chan error)
				executor.ResourceModule.UpdateInstanceStatus(ins, respChan)
				err = <- respChan
				if err != nil{
					log.Printf("[%08X] update new disk size fail: %s", id, err.Error())
					resp.SetError(err.Error())
					return executor.Sender.SendMessage(resp, request.GetSender())
				}
				log.Printf("[%08X] resize disk success", id)
			}else{
				log.Printf("[%08X] resize disk fail: %s", id, cellResp.GetError())
			}
			cellResp.SetFromSession(id)
			cellResp.SetToSession(request.GetFromSession())
			//forward
			return executor.Sender.SendMessage(cellResp, request.GetSender())
		case <- timer.C:
			//timeout
			log.Printf("[%08X] wait resize disk response timeout", id)
			resp.SetError("request timeout")
			return executor.Sender.SendMessage(resp, request.GetSender())
		}
	}
}

