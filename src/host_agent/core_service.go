package host_agent

import (
	"log"
	"vm_manager/host_agent/src/modules"
	"vm_manager/vm_utils"
)

const (
	CurrentVersion = "1.3.1"
)

type CoreService struct {
	vm_utils.EndpointService //base class
	ConfigPath               string
	DataPath                 string
	resourceManager          *modules.ResourceManager
	transManager             *CoreTransactionManager
	apiModule                *modules.APIModule
}

func (core *CoreService) GetAPIServiceAddress() string {
	if nil != core.apiModule {
		return core.apiModule.GetServiceAddress()
	}
	return ""
}

func (core *CoreService) GetVersion() string {
	return CurrentVersion
}

func (core *CoreService) OnMessageReceived(msg vm_utils.Message) {

	if targetSession := msg.GetToSession(); targetSession != 0 {
		if err := core.transManager.PushMessage(msg); err != nil {
			log.Printf("<core> push message [%08X] from %s to session [%08X] fail: %s", msg.GetID(), msg.GetSender(), targetSession, err.Error())
		}
		return
	}
	var err error
	switch msg.GetID() {
	case vm_utils.QueryComputePoolRequest:
	case vm_utils.GetComputePoolRequest:
	case vm_utils.CreateComputePoolRequest:
	case vm_utils.DeleteComputePoolRequest:
	case vm_utils.ModifyComputePoolRequest:

	case vm_utils.QueryStoragePoolRequest:
	case vm_utils.GetStoragePoolRequest:
	case vm_utils.CreateStoragePoolRequest:
	case vm_utils.DeleteStoragePoolRequest:
	case vm_utils.ModifyStoragePoolRequest:
	case vm_utils.QueryAddressPoolRequest:
	case vm_utils.GetAddressPoolRequest:
	case vm_utils.CreateAddressPoolRequest:
	case vm_utils.ModifyAddressPoolRequest:
	case vm_utils.DeleteAddressPoolRequest:
	case vm_utils.QueryAddressRangeRequest:
	case vm_utils.GetAddressRangeRequest:
	case vm_utils.AddAddressRangeRequest:
	case vm_utils.RemoveAddressRangeRequest:

	case vm_utils.QueryComputePoolCellRequest:
	case vm_utils.GetComputePoolCellRequest:
	case vm_utils.AddComputePoolCellRequest:
	case vm_utils.RemoveComputePoolCellRequest:
	case vm_utils.QueryUnallocatedComputePoolCellRequest:
	case vm_utils.QueryZoneStatusRequest:
	case vm_utils.QueryComputePoolStatusRequest:
	case vm_utils.GetComputePoolStatusRequest:
	case vm_utils.QueryComputePoolCellStatusRequest:
	case vm_utils.GetComputePoolCellStatusRequest:
	case vm_utils.EnableComputePoolCellRequest:
	case vm_utils.DisableComputePoolCellRequest:
	case vm_utils.QueryCellStorageRequest:
	case vm_utils.ModifyCellStorageRequest:
	case vm_utils.MigrateInstanceRequest:
	case vm_utils.InstanceMigratedEvent:
	case vm_utils.InstancePurgedEvent:

	case vm_utils.ComputeCellAvailableEvent:
	case vm_utils.ImageServerAvailableEvent:

	case vm_utils.QueryGuestRequest:
	case vm_utils.GetGuestRequest:
	case vm_utils.CreateGuestRequest:
	case vm_utils.DeleteGuestRequest:
	case vm_utils.ResetSystemRequest:
	case vm_utils.SearchGuestRequest:
	case vm_utils.ModifyAutoStartRequest:
	case vm_utils.QueryInstanceStatusRequest:
	case vm_utils.GetInstanceStatusRequest:
	case vm_utils.StartInstanceRequest:
	case vm_utils.StopInstanceRequest:
	case vm_utils.ResetSecretRequest:
	case vm_utils.GuestCreatedEvent:
	case vm_utils.GuestDeletedEvent:
	case vm_utils.GuestStartedEvent:
	case vm_utils.GuestStoppedEvent:
	case vm_utils.GuestUpdatedEvent:
	case vm_utils.CellStatusReportEvent:
	case vm_utils.AddressChangedEvent:
	case vm_utils.SystemResetEvent:
	case vm_utils.StartBatchCreateGuestRequest:
	case vm_utils.GetBatchCreateGuestRequest:
	case vm_utils.StartBatchDeleteGuestRequest:
	case vm_utils.GetBatchDeleteGuestRequest:
	case vm_utils.StartBatchStopGuestRequest:
	case vm_utils.GetBatchStopGuestRequest:
	case vm_utils.ModifyPriorityRequest:
	case vm_utils.ModifyDiskThresholdRequest:
	case vm_utils.ModifyNetworkThresholdRequest:

	case vm_utils.InsertMediaRequest:
	case vm_utils.EjectMediaRequest:
	case vm_utils.MediaAttachedEvent:
	case vm_utils.MediaDetachedEvent:

	case vm_utils.ModifyGuestNameRequest:
	case vm_utils.ModifyCoreRequest:
	case vm_utils.ModifyMemoryRequest:
	case vm_utils.ModifyAuthRequest:
	case vm_utils.GetAuthRequest:
	case vm_utils.ResizeDiskRequest:
	case vm_utils.ShrinkDiskRequest:

	case vm_utils.QueryDiskImageRequest:
	case vm_utils.GetDiskImageRequest:
	case vm_utils.CreateDiskImageRequest:
	case vm_utils.DeleteDiskImageRequest:
	case vm_utils.ModifyDiskImageRequest:
	case vm_utils.SynchronizeDiskImageRequest:

	case vm_utils.QueryMediaImageRequest:
	case vm_utils.GetMediaImageRequest:
	case vm_utils.CreateMediaImageRequest:
	case vm_utils.DeleteMediaImageRequest:
	case vm_utils.ModifyMediaImageRequest:
	case vm_utils.SynchronizeMediaImageRequest:

	case vm_utils.QuerySnapshotRequest:
	case vm_utils.GetSnapshotRequest:
	case vm_utils.CreateSnapshotRequest:
	case vm_utils.DeleteSnapshotRequest:
	case vm_utils.RestoreSnapshotRequest:
	case vm_utils.SnapshotResumedEvent:

	case vm_utils.QueryMigrationRequest:
	case vm_utils.GetMigrationRequest:
	case vm_utils.CreateMigrationRequest:
	case vm_utils.QueryTemplateRequest:
	case vm_utils.GetTemplateRequest:
	case vm_utils.CreateTemplateRequest:
	case vm_utils.ModifyTemplateRequest:
	case vm_utils.DeleteTemplateRequest:
	case vm_utils.ComputeCellDisconnectedEvent:
	//security policy group
	case vm_utils.QueryPolicyGroupRequest:
	case vm_utils.GetPolicyGroupRequest:
	case vm_utils.CreatePolicyGroupRequest:
	case vm_utils.ModifyPolicyGroupRequest:
	case vm_utils.DeletePolicyGroupRequest:
	case vm_utils.QueryPolicyRuleRequest:
	case vm_utils.AddPolicyRuleRequest:
	case vm_utils.ModifyPolicyRuleRequest:
	case vm_utils.ChangePolicyRuleOrderRequest:
	case vm_utils.RemovePolicyRuleRequest:

	//guest security policy
	case vm_utils.GetGuestRuleRequest:
	case vm_utils.ChangeGuestRuleOrderRequest:
	case vm_utils.ChangeGuestRuleDefaultActionRequest:
	case vm_utils.AddGuestRuleRequest:
	case vm_utils.ModifyGuestRuleRequest:
	case vm_utils.RemoveGuestRuleRequest:
	default:
		core.handleIncomingMessage(msg)
		return
	}
	//Invoke transaction
	err = core.transManager.InvokeTask(msg)
	if err != nil {
		log.Printf("<core> invoke transaction with message [%08X] fail: %s", msg.GetID(), err.Error())
	}
}
func (core *CoreService) handleIncomingMessage(msg vm_utils.Message) {
	switch msg.GetID() {
	default:
		log.Printf("<core> message [%08X] from %s.[%08X] ignored", msg.GetID(), msg.GetSender(), msg.GetFromSession())
	}
}

func (core *CoreService) OnServiceConnected(name string, t vm_utils.ServiceType, remoteAddress string) {
	log.Printf("<core> service %s connected, type %d", name, t)
	switch t {
	case vm_utils.ServiceTypeCell:
		event, _ := vm_utils.CreateJsonMessage(vm_utils.ComputeCellAvailableEvent)
		event.SetString(vm_utils.ParamKeyCell, name)
		event.SetString(vm_utils.ParamKeyAddress, remoteAddress)
		core.SendToSelf(event)
	default:
		break
	}
}

func (core *CoreService) OnServiceDisconnected(nodeName string, t vm_utils.ServiceType, gracefullyClose bool) {
	if gracefullyClose {
		log.Printf("<core> service %s closed by remote, type %d", nodeName, t)
	} else {
		log.Printf("<core> service %s lost, type %d", nodeName, t)
	}

	switch t {
	case vm_utils.ServiceTypeCell:
		event, _ := vm_utils.CreateJsonMessage(vm_utils.ComputeCellDisconnectedEvent)
		event.SetString(vm_utils.ParamKeyCell, nodeName)
		event.SetBoolean(vm_utils.ParamKeyFlag, gracefullyClose)
		core.SendToSelf(event)
	case vm_utils.ServiceTypeImage:
		core.resourceManager.RemoveImageServer(nodeName)
	default:
		break
	}
}

func (core *CoreService) OnDependencyReady() {
	core.SetServiceReady()
}

func (core *CoreService) InitialEndpoint() (err error) {
	log.Printf("<core> initial core service, v %s", CurrentVersion)
	log.Printf("<core> domain %s, group address %s:%d", core.GetDomain(), core.GetGroupAddress(), core.GetGroupPort())

	core.resourceManager, err = modules.CreateResourceManager(core.DataPath)
	if err != nil {
		return err
	}
	core.transManager, err = CreateTransactionManager(core, core.resourceManager)
	if err != nil {
		return err
	}

	core.apiModule, err = modules.CreateAPIModule(core.ConfigPath, core, core.resourceManager)
	if err != nil {
		return err
	}
	//register submodules
	if err = core.RegisterSubmodule(core.apiModule.GetModuleName(), core.apiModule.GetResponseChannel()); err != nil {
		return err
	}
	return nil
}

func (core *CoreService) OnEndpointStarted() (err error) {
	if err = core.resourceManager.Start(); err != nil {
		return err
	}
	if err = core.transManager.Start(); err != nil {
		return err
	}
	if err = core.apiModule.Start(); err != nil {
		return err
	}
	log.Print("<core> started")
	return nil
}

func (core *CoreService) OnEndpointStopped() {
	if err := core.apiModule.Stop(); err != nil {
		log.Printf("<core> stop api module fail: %s", err.Error())
	}
	if err := core.transManager.Stop(); err != nil {
		log.Printf("<core> stop transaction manager fail: %s", err.Error())
	}
	if err := core.resourceManager.Stop(); err != nil {
		log.Printf("<core> stop compute pool module fail: %s", err.Error())
	}
	log.Print("<core> stopped")
}
