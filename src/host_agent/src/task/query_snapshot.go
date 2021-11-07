package task

import (
	"github.com/project-nano/framework"
	"log"
	"time"
	"github.com/project-nano/core/modules"
)

type QuerySnapshotExecutor struct {
	Sender         framework.MessageSender
	ResourceModule modules.ResourceModule
}

func (executor *QuerySnapshotExecutor)Execute(id framework.SessionID, request framework.Message,
	incoming chan framework.Message, terminate chan bool) error {
	instanceID, err := request.GetString(framework.ParamKeyInstance)
	if err != nil{
		return err
	}
	log.Printf("[%08X] request query snapshot of guest '%s' from %s.[%08X]", id, instanceID,
		request.GetSender(), request.GetFromSession())

	var ins modules.InstanceStatus
	resp, _ := framework.CreateJsonMessage(framework.QuerySnapshotResponse)
	resp.SetToSession(request.GetFromSession())
	resp.SetFromSession(id)
	resp.SetSuccess(false)
	{
		var respChan = make(chan modules.ResourceResult)
		executor.ResourceModule.GetInstanceStatus(instanceID, respChan)
		result := <- respChan
		if result.Error != nil{
			log.Printf("[%08X] fetch instance fail: %s", id, result.Error.Error())
			resp.SetError(result.Error.Error())
			return executor.Sender.SendMessage(resp, request.GetSender())
		}
		ins = result.Instance
	}
	{
		//forward request
		forward, _ := framework.CreateJsonMessage(framework.QuerySnapshotRequest)
		forward.SetFromSession(id)
		forward.SetString(framework.ParamKeyInstance, instanceID)
		if err = executor.Sender.SendMessage(forward, ins.Cell); err != nil{
			log.Printf("[%08X] forward query snapshot to cell '%s' fail: %s", id, ins.Cell, err.Error())
			resp.SetError(err.Error())
			return executor.Sender.SendMessage(resp, request.GetSender())
		}
		timer := time.NewTimer(modules.DefaultOperateTimeout)
		select{
		case cellResp := <- incoming:
			if !cellResp.IsSuccess(){
				//	log.Printf("[%08X] cell get snapshot success", id)
				//}else{
				log.Printf("[%08X] cell query snapshot fail: %s", id, cellResp.GetError())
			}
			cellResp.SetFromSession(id)
			cellResp.SetToSession(request.GetFromSession())
			//forward
			return executor.Sender.SendMessage(cellResp, request.GetSender())
		case <- timer.C:
			//timeout
			log.Printf("[%08X] wait query snapshot response timeout", id)
			resp.SetError("cell timeout")
			return executor.Sender.SendMessage(resp, request.GetSender())
		}
	}
}
