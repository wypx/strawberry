package vm_agent

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	VmAgentSvc "vm_manager/vm_agent/svc"
	VmUtils "vm_manager/vm_utils"

	"github.com/libvirt/libvirt-go"
)

const (
	CurrentVersion = "1.3.1"
)

type CellService struct {
	VmUtils.EndpointService
	DataPath       string
	collector      *CollectorModule
	insManager     *VmAgentSvc.InstanceManager
	storageManager *VmAgentSvc.StorageManager
	networkManager *VmAgentSvc.NetworkManager
	transManager   *TransactionManager
	virConnect     *libvirt.Connect
	initiator      *VmAgentSvc.GuestInitiator
	dhcpService    *VmAgentSvc.DHCPService
}

func CreateCellService(config DomainConfig, workingPath string) (service *CellService, err error) {
	var dataPath = filepath.Join(workingPath, DataPathName)
	if _, err = os.Stat(dataPath); os.IsNotExist(err) {
		if err = os.Mkdir(dataPath, DefaultPathPerm); err != nil {
			err = fmt.Errorf("create data path '%s' fail: %s", dataPath, err.Error())
			return
		} else {
			log.Printf("data path '%s' created", dataPath)
		}
	}
	service = &CellService{}
	service.DataPath = dataPath
	if service.EndpointService, err = VmUtils.CreatePeerEndpoint(config.GroupAddress, config.GroupPort, config.Domain); err != nil {
		err = fmt.Errorf("create new endpoint fail: %s", err.Error())
		return
	}
	return service, nil
}

func (cell *CellService) OnMessageReceived(msg VmUtils.Message) {
	if targetSession := msg.GetToSession(); targetSession != 0 {
		if err := cell.transManager.PushMessage(msg); err != nil {
			log.Printf("<cell> push message [%08X] from %s to session [%08X] fail: %s", msg.GetID(), msg.GetSender(), targetSession, err.Error())
		}
		return
	}
	switch msg.GetID() {
	case VmUtils.CreateGuestRequest:
	case VmUtils.DeleteGuestRequest:
	case VmUtils.GetGuestRequest:
	case VmUtils.QueryGuestRequest:
	case VmUtils.GetInstanceStatusRequest:
	case VmUtils.StartInstanceRequest:
	case VmUtils.StopInstanceRequest:
	case VmUtils.ComputePoolReadyEvent:
	case VmUtils.CreateDiskImageRequest:
	case VmUtils.ModifyCoreRequest:
	case VmUtils.ModifyMemoryRequest:
	case VmUtils.ModifyPriorityRequest:
	case VmUtils.ModifyDiskThresholdRequest:
	case VmUtils.ModifyNetworkThresholdRequest:
	case VmUtils.ModifyAuthRequest:
	case VmUtils.ModifyGuestNameRequest:
	case VmUtils.ModifyAutoStartRequest:
	case VmUtils.GetAuthRequest:
	case VmUtils.ResizeDiskRequest:
	case VmUtils.ShrinkDiskRequest:
	case VmUtils.ResetSystemRequest:
	case VmUtils.InsertMediaRequest:
	case VmUtils.EjectMediaRequest:
	case VmUtils.QuerySnapshotRequest:
	case VmUtils.GetSnapshotRequest:
	case VmUtils.CreateSnapshotRequest:
	case VmUtils.DeleteSnapshotRequest:
	case VmUtils.RestoreSnapshotRequest:
	case VmUtils.GetComputePoolCellRequest:
	case VmUtils.ComputeCellRemovedEvent:
	case VmUtils.AttachInstanceRequest:
	case VmUtils.DetachInstanceRequest:
	case VmUtils.ResetSecretRequest:
	case VmUtils.QueryCellStorageRequest:
	case VmUtils.ModifyCellStorageRequest:
	case VmUtils.AddressPoolChangedEvent:

	//security policy
	case VmUtils.GetGuestRuleRequest:
	case VmUtils.AddGuestRuleRequest:
	case VmUtils.ModifyGuestRuleRequest:
	case VmUtils.ChangeGuestRuleDefaultActionRequest:
	case VmUtils.ChangeGuestRuleOrderRequest:
	case VmUtils.RemoveGuestRuleRequest:

	default:
		cell.handleIncomingMessage(msg)
		return
	}
	var err = cell.transManager.InvokeTask(msg)
	if err != nil {
		log.Printf("<cell> invoke transaction with message [%08X] fail: %s", msg.GetID(), err.Error())
	}
}

func (cell *CellService) GetVersion() string {
	return CurrentVersion
}

func (cell *CellService) handleIncomingMessage(msg VmUtils.Message) {
	switch msg.GetID() {
	default:
		log.Printf("<cell> message [%08X] from %s.[%08X] ignored", msg.GetID(), msg.GetSender(), msg.GetFromSession())
	}
}

func (cell *CellService) OnServiceConnected(name string, t VmUtils.ServiceType, remoteAddress string) {
	log.Printf("<cell> service %s connected, type %d", name, t)
	if t == VmUtils.ServiceTypeCore {
		cell.collector.AddObserver(name)
	}
}

func (cell *CellService) OnServiceDisconnected(name string, t VmUtils.ServiceType, gracefullyClose bool) {
	if gracefullyClose {
		log.Printf("<cell> service %s closed by remote, type %d", name, t)
	} else {
		log.Printf("<cell> service %s lost, type %d", name, t)
	}
	if t == VmUtils.ServiceTypeCore {
		cell.collector.RemoveObserver(name)
	}
}

func (cell *CellService) OnDependencyReady() {
	cell.SetServiceReady()
}

func (cell *CellService) InitialEndpoint() error {
	log.Printf("<cell> initial cell service, v %s", CurrentVersion)
	log.Printf("<cell> domain %s, group address %s:%d", cell.GetDomain(), cell.GetGroupAddress(), cell.GetGroupPort())
	var err error

	const (
		DefaultLibvirtURL = "qemu:///system"
	)
	if cell.virConnect, err = libvirt.NewConnect(DefaultLibvirtURL); err != nil {
		return err
	}
	if cell.storageManager, err = VmAgentSvc.CreateStorageManager(cell.DataPath, cell.virConnect); err != nil {
		return err
	}

	if cell.insManager, err = VmAgentSvc.CreateInstanceManager(cell.DataPath, cell.virConnect); err != nil {
		return err
	}
	if cell.collector, err = CreateCollectorModule(cell,
		cell.insManager.GetEventChannel(), cell.storageManager.GetOutputEventChannel()); err != nil {
		return err
	}

	if cell.networkManager, err = VmAgentSvc.CreateNetworkManager(cell.DataPath, cell.virConnect); err != nil {
		return err
	}
	var networkResources = cell.insManager.GetInstanceNetworkResources()
	if err = cell.networkManager.SyncInstanceResources(networkResources); err != nil {
		return err
	}
	if cell.initiator, err = VmAgentSvc.CreateInitiator(cell.networkManager, cell.insManager); err != nil {
		return err
	}
	if cell.dhcpService, err = VmAgentSvc.CreateDHCPService(cell.networkManager); err != nil {
		return err
	}

	cell.transManager, err = CreateTransactionManager(cell, cell.insManager, cell.storageManager, cell.networkManager)
	if err != nil {
		return err
	}
	log.Println("<cell> all module ready")
	return nil
}
func (cell *CellService) OnEndpointStarted() (err error) {
	if err = cell.collector.Start(); err != nil {
		return err
	}
	if err = cell.insManager.Start(); err != nil {
		return err
	}
	if err = cell.storageManager.Start(); err != nil {
		return err
	}
	if err = cell.networkManager.Start(); err != nil {
		return err
	}
	if err = cell.initiator.Start(); err != nil {
		return
	}
	if err = cell.dhcpService.Start(); err != nil {
		return
	}
	if err = cell.transManager.Start(); err != nil {
		return err
	}
	log.Println("<cell> started")
	return nil
}
func (cell *CellService) OnEndpointStopped() {
	if err := cell.transManager.Stop(); err != nil {
		log.Printf("<cell> stop transaction manger fail: %s", err.Error())
	}
	if err := cell.dhcpService.Stop(); err != nil {
		log.Printf("<cell> stop dhcp service fail: %s", err.Error())
	}
	if err := cell.initiator.Stop(); err != nil {
		log.Printf("<cell> stop guest initiator fail: %s", err.Error())
	}
	if err := cell.networkManager.Stop(); err != nil {
		log.Printf("<cell> stop network manager fail: %s", err.Error())
	}
	if err := cell.storageManager.Stop(); err != nil {
		log.Printf("<cell> stop storage manager fail: %s", err.Error())
	}
	if err := cell.insManager.Stop(); err != nil {
		log.Printf("<cell> stop instance manager fail: %s", err.Error())
	}
	if err := cell.collector.Stop(); err != nil {
		log.Printf("<cell> stop collector fail: %s", err.Error())
	}
	log.Println("<cell> all module stopped")
}
