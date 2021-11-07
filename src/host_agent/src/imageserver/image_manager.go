package imageserver

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/project-nano/framework"
	"github.com/satori/go.uuid"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type ImageConfig struct {
	Name        string   `json:"name"`
	Owner       string   `json:"owner"`
	Group       string   `json:"group"`
	Description string   `json:"description,omitempty"`
	Tags        []string `json:"tags,omitempty"`
}

type ImageStatus struct {
	ImageConfig
	ID         string `json:"id"`
	Format     string `json:"format"`
	Path       string `json:"path"`
	Size       uint   `json:"size"`
	Version    uint   `json:"version"`
	Locked     bool   `json:"-"`
	CreateTime string `json:"create_time,omitempty"`
	ModifyTime string `json:"modify_time,omitempty"`
}

type DiskStatus struct {
	ImageStatus
	CheckSum   string `json:"check_sum,omitempty"`
	Created    bool   `json:"-"`
	Progress   uint   `json:"-"`
}

type imageCommand struct {
	Type             ImageCommandType
	ID               string
	CheckSum         string
	Progress         uint
	User             string
	Group            string
	Tags             []string
	MediaImageConfig ImageConfig
	DiskImageConfig  ImageConfig
	ResultChan       chan ImageResult
	ErrorChan        chan error
}

type ImageCommandType int

const (
	cmdQueryMediaImage  = iota
	cmdCreateMediaImage
	cmdModifyMediaImage
	cmdDeleteMediaImage
	cmdLockMediaImage
	cmdFinishMediaImage
	cmdUnlockMediaImage
	cmdGetMediaImageFile
	cmdGetMediaImage
	cmdSyncMediaImages
	cmdQueryDiskImage
	cmdCreateDiskImage
	cmdModifyDiskImage
	cmdDeleteDiskImage
	cmdLockDiskImage
	cmdFinishDiskImage
	cmdUnlockDiskImage
	cmdGetDiskImage
	cmdGetDiskImageFile
	cmdUpdateDiskImageProgress
	cmdSyncDiskImages
)

type ImageResult struct {
	Error      error
	ID         string
	Path       string
	Size       uint
	CheckSum   string
	MediaList  []ImageStatus
	DiskList   []DiskStatus
	MediaImage ImageStatus
	DiskImage  DiskStatus
}

type ImageManager struct {
	mediaImages     map[string]ImageStatus //key = image id
	mediaImageNames map[string]bool        //key = group.name
	mediaPath       string
	diskImages      map[string]DiskStatus
	diskImageNames  map[string]bool //key = group.name
	diskPath        string
	dataFile        string
	commands        chan imageCommand
	runner          *framework.SimpleRunner
}

const (
	TimeFormatLayout   = "2006-01-02 15:04:05"
	FormatExtQCOW2     = "qcow2"
	FormatExtISO       = "iso"
	DefaultDiskFormat  = FormatExtQCOW2
	DefaultMediaFormat = FormatExtISO
)

func CreateImageManager(dataPath string) (manager *ImageManager, err error){
	const (
		DefaultQueueSize = 1 << 10
		PathPerm = 0700
		MediaPathName = "media_images"
		DiskPathName = "disk_images"
		DataFileName = "image.data"
	)
	manager = &ImageManager{}
	manager.runner = framework.CreateSimpleRunner(manager.Routine)
	manager.mediaImages = map[string]ImageStatus{}
	manager.mediaImageNames = map[string]bool{}
	manager.diskImages = map[string]DiskStatus{}
	manager.diskImageNames = map[string]bool{}

	manager.commands = make(chan imageCommand, DefaultQueueSize)
	manager.dataFile = filepath.Join(dataPath, DataFileName)
	manager.mediaPath = filepath.Join(dataPath, MediaPathName)
	manager.diskPath = filepath.Join(dataPath, DiskPathName)
	if _, err := os.Stat(manager.mediaPath);os.IsNotExist(err){
		if err = os.Mkdir(manager.mediaPath, PathPerm);err != nil{
			return nil, err
		}else{
			log.Printf("<image> new media path '%s' created", manager.mediaPath)
		}
	}
	if _, err := os.Stat(manager.diskPath);os.IsNotExist(err){
		if err = os.Mkdir(manager.diskPath, PathPerm);err != nil{
			return nil, err
		}else{
			log.Printf("<image> new disk path '%s' created", manager.diskPath)
		}
	}
	if err = manager.LoadData();err != nil{
		return nil, err
	}
	return manager, nil
}

func (manager *ImageManager) Start() error{
	return manager.runner.Start()
}

func (manager *ImageManager) Stop() error{
	return manager.runner.Stop()
}

func (manager *ImageManager) Routine(c framework.RoutineController)  {
	log.Printf("<image> started")
	for !c.IsStopping(){
		select {
		case <- c.GetNotifyChannel():
			c.SetStopping()
		case cmd := <- manager.commands:
			manager.handleCommand(cmd)
		}
	}
	c.NotifyExit()
	log.Printf("<image> stopped")
}

type imageSavedData struct {
	MediaImages []ImageStatus `json:"media_images"`
	DiskImages  []DiskStatus  `json:"disk_images"`
}

func (manager *ImageManager) SaveData() error{
	const (
		FilePerm = 0640
	)
	var saved imageSavedData
	for _, media := range manager.mediaImages{
		saved.MediaImages = append(saved.MediaImages, media)
	}
	for _, image := range manager.diskImages{
		saved.DiskImages = append(saved.DiskImages, image)
	}
	data, err := json.MarshalIndent(saved, "", " ")
	if err != nil{
		return err
	}
	if err = ioutil.WriteFile(manager.dataFile, data, FilePerm);err != nil{
		return err
	}
	log.Printf("<image> %d media image(s), %d disk image(s) saved into '%s'", 
		len(saved.MediaImages), len(saved.DiskImages), manager.dataFile)
	return nil
}

func (manager *ImageManager) LoadData() error{
	if _, err := os.Stat(manager.dataFile);os.IsNotExist(err){
		log.Println("<image> no images configured")
		return nil
	}
	data, err := ioutil.ReadFile(manager.dataFile)
	if err != nil{
		return err
	}
	var saved imageSavedData
	if err = json.Unmarshal(data, &saved);err != nil{
		return err
	}
	for _, image := range saved.MediaImages{
		image.Locked = false
		manager.mediaImages[image.ID] = image
		var nameWithGroup = fmt.Sprintf("%s.%s", image.Group, image.Name)
		manager.mediaImageNames[nameWithGroup] = true
	}
	for _, image := range saved.DiskImages{
		image.Locked = false
		image.Created = true
		manager.diskImages[image.ID] = image
		var nameWithGroup = fmt.Sprintf("%s.%s", image.Group, image.Name)
		manager.diskImageNames[nameWithGroup] = true
	}
	log.Printf("<image> %d media image(s), %d disk image(s) loaded from '%s'", 
		len(saved.MediaImages), len(saved.DiskImages), manager.dataFile)
	return nil
}

func (manager *ImageManager) handleCommand(cmd imageCommand){
	var err error
	switch cmd.Type {
	case cmdQueryMediaImage:
		err = manager.handleQueryMediaImage(cmd.User, cmd.Group, cmd.ResultChan)
	case cmdCreateMediaImage:
		err = manager.handleCreateMediaImage(cmd.MediaImageConfig, cmd.ResultChan)
	case cmdDeleteMediaImage:
		err = manager.handleDeleteMediaImage(cmd.ID, cmd.ErrorChan)
	case cmdLockMediaImage:
		err = manager.handleLockMediaImageForUpdate(cmd.ID, cmd.ResultChan)
	case cmdFinishMediaImage:
		err = manager.handleFinishMediaImage(cmd.ID, cmd.ErrorChan)
	case cmdUnlockMediaImage:
		err = manager.handleUnlockMediaImage(cmd.ID, cmd.ErrorChan)
	case cmdGetMediaImageFile:
		err = manager.handleGetMediaImageFile(cmd.ID, cmd.ResultChan)
	case cmdGetMediaImage:
		err = manager.handleGetMediaImage(cmd.ID, cmd.ResultChan)
	case cmdModifyMediaImage:
		err = manager.handleModifyMediaImage(cmd.ID, cmd.MediaImageConfig, cmd.ErrorChan)
	case cmdQueryDiskImage:
		err = manager.handleQueryDiskImage(cmd.User, cmd.Group, cmd.Tags, cmd.ResultChan)
	case cmdCreateDiskImage:
		err = manager.handleCreateDiskImage(cmd.DiskImageConfig, cmd.ResultChan)
	case cmdModifyDiskImage:
		err = manager.handleModifyDiskImage(cmd.ID, cmd.DiskImageConfig, cmd.ErrorChan)
	case cmdDeleteDiskImage:
		err = manager.handleDeleteDiskImage(cmd.ID, cmd.ErrorChan)
	case cmdLockDiskImage:
		err = manager.handleLockDiskImageForUpdate(cmd.ID, cmd.ResultChan)
	case cmdFinishDiskImage:
		err = manager.handleFinishDiskImage(cmd.ID, cmd.CheckSum, cmd.ErrorChan)
	case cmdUnlockDiskImage:
		err = manager.handleUnlockDiskImage(cmd.ID, cmd.ErrorChan)
	case cmdGetDiskImage:
		err = manager.handleGetDiskImage(cmd.ID, cmd.ResultChan)
	case cmdGetDiskImageFile:
		err = manager.handleGetDiskImageFile(cmd.ID, cmd.ResultChan)
	case cmdUpdateDiskImageProgress:
		err = manager.handleUpdateDiskImageProgress(cmd.ID, cmd.Progress, cmd.ErrorChan)
	case cmdSyncMediaImages:
		err = manager.handleSyncMediaImages(cmd.User, cmd.Group, cmd.ErrorChan)
	case cmdSyncDiskImages:
		err = manager.handleSyncDiskImages(cmd.User, cmd.Group, cmd.ErrorChan)
	default:
		log.Printf("<image> unsupported command type %d", cmd.Type)
		break
	}
	if err != nil {
		log.Printf("<image> handle command %d fail: %s", cmd.Type, err.Error())
	}
}

func (manager *ImageManager) QueryMediaImage(owner, group string, respChan chan ImageResult){
	cmd := imageCommand{Type: cmdQueryMediaImage, User:owner, Group:group,  ResultChan:respChan}
	manager.commands <- cmd
}

func (manager *ImageManager) CreateMediaImage(config ImageConfig, respChan chan ImageResult){
	cmd := imageCommand{Type: cmdCreateMediaImage, MediaImageConfig:config, ResultChan:respChan}
	manager.commands <- cmd
}

func (manager *ImageManager) DeleteMediaImage(id string, respChan chan error){
	cmd := imageCommand{Type: cmdDeleteMediaImage, ID:id, ErrorChan:respChan}
	manager.commands <- cmd
}

func (manager *ImageManager) LockMediaImageForUpdate(id string, respChan chan ImageResult){
	cmd := imageCommand{Type: cmdLockMediaImage, ID:id, ResultChan:respChan}
	manager.commands <- cmd
}

func (manager *ImageManager) FinishMediaImage(id string, respChan chan error){
	cmd := imageCommand{Type: cmdFinishMediaImage, ID:id, ErrorChan:respChan}
	manager.commands <- cmd
}

func (manager *ImageManager) UnlockMediaImage(id string, respChan chan error){
	cmd := imageCommand{Type: cmdUnlockMediaImage, ID:id, ErrorChan:respChan}
	manager.commands <- cmd
}


func (manager * ImageManager) GetMediaImage(id string, respChan chan ImageResult){
	cmd := imageCommand{Type: cmdGetMediaImage, ID:id, ResultChan:respChan}
	manager.commands <- cmd
}

func (manager * ImageManager) GetMediaImageFile(id string, respChan chan ImageResult){
	cmd := imageCommand{Type: cmdGetMediaImageFile, ID:id, ResultChan:respChan}
	manager.commands <- cmd
}

func (manager * ImageManager) ModifyMediaImage(id string, config ImageConfig, respChan chan error){
	manager.commands <- imageCommand{Type: cmdModifyMediaImage, ID: id, MediaImageConfig:config, ErrorChan: respChan}
}


func (manager *ImageManager) QueryDiskImage(owner, group string, tags []string, respChan chan ImageResult){
	cmd := imageCommand{Type: cmdQueryDiskImage, User:owner, Group: group, Tags: tags, ResultChan:respChan}
	manager.commands <- cmd
}

func (manager *ImageManager) CreateDiskImage(config ImageConfig, respChan chan ImageResult){
	cmd := imageCommand{Type: cmdCreateDiskImage, DiskImageConfig:config, ResultChan:respChan}
	manager.commands <- cmd
}

func (manager * ImageManager) ModifyDiskImage(id string, config ImageConfig, respChan chan error){
	manager.commands <- imageCommand{Type: cmdModifyDiskImage, ID: id, DiskImageConfig:config, ErrorChan: respChan}
}

func (manager *ImageManager) DeleteDiskImage(id string, respChan chan error){
	cmd := imageCommand{Type: cmdDeleteDiskImage, ID:id, ErrorChan:respChan}
	manager.commands <- cmd
}

func (manager *ImageManager) LockDiskImageForUpdate(id string, respChan chan ImageResult){
	cmd := imageCommand{Type: cmdLockDiskImage, ID:id, ResultChan:respChan}
	manager.commands <- cmd
}

func (manager *ImageManager) FinishDiskImage(id, checksum string, respChan chan error){
	cmd := imageCommand{Type: cmdFinishDiskImage, ID:id, CheckSum:checksum, ErrorChan:respChan}
	manager.commands <- cmd
}

func (manager *ImageManager) UnlockDiskImage(id string, respChan chan error){
	cmd := imageCommand{Type: cmdUnlockDiskImage, ID:id, ErrorChan:respChan}
	manager.commands <- cmd
}

func (manager * ImageManager) GetDiskImage(id string, respChan chan ImageResult){
	manager.commands <- imageCommand{Type: cmdGetDiskImage, ID:id, ResultChan:respChan}
}

func (manager * ImageManager) GetDiskImageFile(id string, respChan chan ImageResult){
	cmd := imageCommand{Type: cmdGetDiskImageFile, ID:id, ResultChan:respChan}
	manager.commands <- cmd
}

func (manager * ImageManager) UpdateDiskImageProgress(id string, progress uint, respChan chan error){
	manager.commands <- imageCommand{Type: cmdUpdateDiskImageProgress, ID:id, Progress: progress, ErrorChan: respChan}
}

func (manager * ImageManager) SyncMediaImages(owner, group string, respChan chan error){
	manager.commands <- imageCommand{Type: cmdSyncMediaImages, User: owner, Group: group, ErrorChan: respChan}
}

func (manager * ImageManager) SyncDiskImages(owner, group string, respChan chan error){
	manager.commands <- imageCommand{Type: cmdSyncDiskImages, User: owner, Group: group, ErrorChan: respChan}
}

func (manager *ImageManager) handleQueryMediaImage(owner, group string, respChan chan ImageResult) (err error){
	var result []ImageStatus
	var names []string
	var nameToID = map[string]string{}
	var filterByOwner = 0 != len(owner)
	var filterByGroup = 0 != len(group)
	for id, image := range manager.mediaImages{
		if !(filterByOwner && owner == image.Owner) && !(filterByGroup && group == image.Group ) {
			//both owner and group unmatched
			continue
		}
		var nameWithGroup = fmt.Sprintf("%s.%s", image.Group, image.Name)
		nameToID[nameWithGroup] = id
		names = append(names, nameWithGroup)
	}

	//sort
	sort.Stable(sort.StringSlice(names))
	for _, name := range names{
		imageID, exists := nameToID[name]
		if !exists{
			err = fmt.Errorf("invalid image name '%s'", name)
			respChan <- ImageResult{Error:err}
			return
		}
		image, exists := manager.mediaImages[imageID]
		if !exists{
			err = fmt.Errorf("invalid image id '%s'", imageID)
			respChan <- ImageResult{Error:err}
			return
		}
		result = append(result, image)
	}

	respChan <- ImageResult{MediaList:result}
	return nil
}

func (manager *ImageManager) handleCreateMediaImage(config ImageConfig, respChan chan ImageResult) (err error){
	var nameWithGroup = fmt.Sprintf("%s.%s", config.Group, config.Name)
	if _, exists := manager.mediaImageNames[nameWithGroup]; exists{
		err = fmt.Errorf("media image '%s' already exists in group '%s'", config.Name, config.Group)
		respChan <- ImageResult{Error:err}
		return
	}
	var newID = uuid.NewV4()
	var image = ImageStatus{}
	image.ImageConfig = config
	image.ID = newID.String()
	image.Size = 0
	image.Version = 0
	image.Locked = false
	image.Format = DefaultMediaFormat
	image.CreateTime = time.Now().Format(TimeFormatLayout)
	manager.mediaImages[image.ID] = image
	manager.mediaImageNames[nameWithGroup] = true
	respChan <- ImageResult{ID:image.ID}
	log.Printf("<image> new media image '%s'(id '%s') created", config.Name, image.ID)
	return manager.SaveData()
}

func (manager *ImageManager) handleDeleteMediaImage(id string, respChan chan error) error{
	image, exists := manager.mediaImages[id]
	if !exists{
		err := fmt.Errorf("invalid media image '%s'", id)
		respChan <- err
		return err
	}
	if image.Locked{
		err := fmt.Errorf("media image '%s' locked", id)
		respChan <- err
		return err
	}
	if _, err := os.Stat(image.Path); !os.IsNotExist(err){
		//delete image file
		if err = os.Remove(image.Path); err != nil{
			log.Printf("<image> delete media image fail: %s", err.Error())
		}
	}
	var nameWithGroup = fmt.Sprintf("%s.%s", image.Group, image.Name)
	delete(manager.mediaImageNames, nameWithGroup)
	delete(manager.mediaImages, id)

	log.Printf("<image> media image '%s' deleted", id)
	respChan <- nil
	return manager.SaveData()
}

func (manager *ImageManager) handleLockMediaImageForUpdate(id string, respChan chan ImageResult) error{
	image, exists := manager.mediaImages[id]
	if !exists{
		err := fmt.Errorf("invalid media image '%s'", id)
		respChan <- ImageResult{Error:err}
		return err
	}
	if image.Locked{
		err := fmt.Errorf("media image '%s' locked", id)
		respChan <- ImageResult{Error:err}
		return err
	}
	//target path
	var newVersion = image.Version + 1
	var targetFile = fmt.Sprintf("%s_v%d.%s", image.ID, newVersion, image.Format)
	var targetPath = filepath.Join(manager.mediaPath, targetFile)
	//lock for update
	image.Locked = true
	manager.mediaImages[image.ID] = image
	log.Printf("<image> media image '%s' locked", id)
	respChan <- ImageResult{Path:targetPath}
	return nil
}

func (manager *ImageManager) handleFinishMediaImage(id string, respChan chan error) error{
	image, exists := manager.mediaImages[id]
	if !exists{
		err := fmt.Errorf("invalid media image '%s'", id)
		respChan <- err
		return err
	}
	if !image.Locked{
		err := fmt.Errorf("media image '%s' is not locked", id)
		respChan <- err
		return err
	}
	var newVersion = image.Version + 1
	var targetFile = fmt.Sprintf("%s_v%d.%s", image.ID, newVersion, image.Format)
	var targetPath = filepath.Join(manager.mediaPath, targetFile)
	if stat, err := os.Stat(targetPath);os.IsNotExist(err){
		err := fmt.Errorf("new file '%s' not available for media image '%s'", targetPath, id)
		respChan <- err
		return err
	}else{
		image.Size = uint(stat.Size())
	}
	var previousFile = fmt.Sprintf("%s_v%d.%s", image.ID, image.Version, image.Format)
	var previousPath = filepath.Join(manager.mediaPath, previousFile)
	if _, err := os.Stat(previousPath);!os.IsNotExist(err){
		if err = os.Remove(previousPath);err != nil{
			log.Printf("<image> warning: delete previous version '%s' fail: %s", previousPath, err.Error())
		}else{
			log.Printf("<image> previous version '%s' deleted", previousPath)
		}
	}
	image.Version = newVersion
	image.Path = targetPath
	image.Locked = false
	image.ModifyTime = time.Now().Format(TimeFormatLayout)
	manager.mediaImages[id] = image
	log.Printf("<image> media image '%s' updated to version %d, file '%s'", id, newVersion, targetPath)
	respChan <- nil
	return manager.SaveData()
}

func (manager *ImageManager) handleUnlockMediaImage(id string, respChan chan error) error{
	image, exists := manager.mediaImages[id]
	if !exists{
		err := fmt.Errorf("invalid media image '%s'", id)
		respChan <- err
		return err
	}
	if !image.Locked{
		err := fmt.Errorf("media image '%s' not locked", id)
		respChan <- err
		return err
	}
	image.Locked = false
	manager.mediaImages[id] = image
	log.Printf("<image> media image '%s' unlocked", id)
	respChan <- nil
	return nil
}

func (manager * ImageManager) handleGetMediaImage(id string, respChan chan ImageResult) (err error){
	image, exists := manager.mediaImages[id]
	if !exists{
		err = fmt.Errorf("invalid media image '%s'", id)
		respChan <- ImageResult{Error:err}
		return err
	}
	respChan <- ImageResult{MediaImage:image}
	return nil
}

func (manager * ImageManager) handleGetMediaImageFile(id string, respChan chan ImageResult) error{
	image, exists := manager.mediaImages[id]
	if !exists{
		err := fmt.Errorf("invalid media image '%s'", id)
		respChan <- ImageResult{Error:err}
		return err
	}
	if 0 == image.Version{
		err := fmt.Errorf("no content available for media image '%s'", id)
		respChan <- ImageResult{Error:err}
		return err
	}
	if image.Locked{
		err := fmt.Errorf("media image '%s' is locked for update", id)
		respChan <- ImageResult{Error:err}
		return err
	}
	respChan <- ImageResult{Path:image.Path, Size:image.Size}
	return nil
}

func (manager * ImageManager) handleModifyMediaImage(id string, config ImageConfig, respChan chan error) (err error){
	image, exists := manager.mediaImages[id]
	if !exists{
		err := fmt.Errorf("invalid media image '%s'", id)
		respChan <- err
		return err
	}
	if image.Locked{
		err := fmt.Errorf("media image '%s' is locked for update", id)
		respChan <- err
		return err
	}
	if config.Name != ""{
		image.Name = config.Name
	}
	if config.Owner != ""{
		image.Owner = config.Owner
	}
	if config.Group != ""{
		image.Group = config.Group
	}
	if config.Description != ""{
		image.Description = config.Description
	}
	if 0 != len(config.Tags){
		image.Tags = config.Tags
	}
	manager.mediaImages[id] = image
	log.Printf("<image> media image '%s' modified", id)
	respChan <- nil
	return manager.SaveData()
}

func (manager *ImageManager) handleQueryDiskImage(owner, group string, tags []string, respChan chan ImageResult) (err error){
	var result []DiskStatus
	var names []string
	var nameToID = map[string]string{}
	var filterByOwner = owner != ""
	var filterByGroup = group != ""
	var filterByTags = 0 != len(tags)
	for id, image := range manager.diskImages {
		if !(filterByOwner && owner == image.Owner) && !(filterByGroup && group == image.Group ) {
			//both owner and group unmatched
			continue
		}
		if filterByTags {
			var tagMatched = false
			for _, target := range tags {
				for _, tag := range image.Tags {
					if tag == target {
						tagMatched = true
						break
					}
				}
				if tagMatched {
					break
				}
			}
			if !tagMatched {
				continue
			}
		}
		var nameWithGroup = fmt.Sprintf("%s.%s", image.Group, image.Name)
		nameToID[nameWithGroup] = id
		names = append(names, nameWithGroup)
	}

	//sort
	sort.Stable(sort.StringSlice(names))
	for _, name := range names{
		imageID, exists := nameToID[name]
		if !exists{
			err = fmt.Errorf("invalid disk image name '%s'", name)
			respChan <- ImageResult{Error:err}
			return
		}
		image, exists := manager.diskImages[imageID]
		if !exists{
			err = fmt.Errorf("invalid disk image id '%s'", imageID)
			respChan <- ImageResult{Error:err}
			return
		}
		result = append(result, image)
	}

	respChan <- ImageResult{DiskList:result}
	return nil
}

func (manager *ImageManager) handleCreateDiskImage(config ImageConfig, respChan chan ImageResult) (err error){
	var nameWithGroup = fmt.Sprintf("%s.%s", config.Group, config.Name)
	if _, exists := manager.diskImageNames[nameWithGroup]; exists{
		err = fmt.Errorf("disk image '%s' already exists in group '%s'", config.Name, config.Group)
		respChan <- ImageResult{Error:err}
		return
	}
	var newID = uuid.NewV4()

	var image = DiskStatus{}
	image.ImageConfig = config
	image.ID = newID.String()
	image.Size = 0
	image.Version = 0
	image.Locked = false
	image.Created = false
	image.Progress = 0
	//todo: more format support
	image.Format = DefaultDiskFormat
	image.CreateTime = time.Now().Format(TimeFormatLayout)
	manager.diskImages[image.ID] = image
	manager.diskImageNames[nameWithGroup] = true
	log.Printf("<image> new disk image '%s'(id '%s') created", config.Name, image.ID)
	respChan <- ImageResult{ID: image.ID}
	return manager.SaveData()
}


func (manager * ImageManager) handleModifyDiskImage(id string, config ImageConfig, respChan chan error) (err error){
	image, exists := manager.diskImages[id]
	if !exists{
		err := fmt.Errorf("invalid disk image '%s'", id)
		respChan <- err
		return err
	}
	if image.Locked{
		err := fmt.Errorf("disk image '%s' is locked for update", id)
		respChan <- err
		return err
	}
	if config.Name != ""{
		image.Name = config.Name
	}
	if config.Owner != ""{
		image.Owner = config.Owner
	}
	if config.Group != ""{
		image.Group = config.Group
	}
	if config.Description != ""{
		image.Description = config.Description
	}
	if 0 != len(config.Tags){
		image.Tags = config.Tags
	}
	manager.diskImages[id] = image
	log.Printf("<image> disk image '%s' modified", id)
	respChan <- nil
	return manager.SaveData()
}


func (manager *ImageManager) handleDeleteDiskImage(id string, respChan chan error) error{
	image, exists := manager.diskImages[id]
	if !exists{
		err := fmt.Errorf("invalid disk image '%s'", id)
		respChan <- err
		return err
	}
	if image.Locked{
		err := fmt.Errorf("disk image '%s' locked", id)
		respChan <- err
		return err
	}
	if _, err := os.Stat(image.Path); !os.IsNotExist(err){
		//delete image file
		if err = os.Remove(image.Path); err != nil{
			log.Printf("<image> delete disk image fail: %s", err.Error())
		}
	}
	var nameWithGroup = fmt.Sprintf("%s.%s", image.Group, image.Name)
	delete(manager.diskImageNames, nameWithGroup)
	delete(manager.diskImages, id)
	log.Printf("<image> disk image '%s' deleted", id)
	respChan <- nil
	return manager.SaveData()
}

func (manager *ImageManager) handleLockDiskImageForUpdate(id string, respChan chan ImageResult) error{
	image, exists := manager.diskImages[id]
	if !exists{
		err := fmt.Errorf("invalid disk image '%s'", id)
		respChan <- ImageResult{Error:err}
		return err
	}
	if image.Locked{
		err := fmt.Errorf("disk image '%s' locked", id)
		respChan <- ImageResult{Error:err}
		return err
	}
	//target path
	var newVersion = image.Version + 1
	var targetFile = fmt.Sprintf("%s_v%d.%s", image.ID, newVersion, image.Format)
	var targetPath = filepath.Join(manager.diskPath, targetFile)
	//lock for update
	image.Locked = true
	manager.diskImages[image.ID] = image
	log.Printf("<image> disk image '%s' locked", id)
	respChan <- ImageResult{Path:targetPath}
	return nil
}

func (manager *ImageManager) handleFinishDiskImage(id, checksum string, respChan chan error) error{
	image, exists := manager.diskImages[id]
	if !exists{
		err := fmt.Errorf("invalid disk image '%s'", id)
		respChan <- err
		return err
	}
	if !image.Locked{
		err := fmt.Errorf("disk image '%s' is not locked", id)
		respChan <- err
		return err
	}
	var newVersion = image.Version + 1
	var targetFile = fmt.Sprintf("%s_v%d.%s", image.ID, newVersion, image.Format)
	var targetPath = filepath.Join(manager.diskPath, targetFile)
	if stat, err := os.Stat(targetPath);os.IsNotExist(err){
		err := fmt.Errorf("new file '%s' not available for disk image '%s'", targetPath, id)
		respChan <- err
		return err
	}else{
		image.Size = uint(stat.Size())
	}
	var previousFile = fmt.Sprintf("%s_v%d.%s", image.ID, image.Version, image.Format)
	var previousPath = filepath.Join(manager.diskPath, previousFile)
	if _, err := os.Stat(previousPath);!os.IsNotExist(err){
		if err = os.Remove(previousPath);err != nil{
			log.Printf("<image> warning: delete previous version '%s' fail: %s", previousPath, err.Error())
		}else{
			log.Printf("<image> previous version '%s' deleted", previousPath)
		}
	}
	image.Version = newVersion
	image.Path = targetPath
	image.CheckSum = checksum
	image.Locked = false
	image.Created = true
	image.ModifyTime = time.Now().Format(TimeFormatLayout)
	manager.diskImages[id] = image
	log.Printf("<image> disk image '%s' updated to version %d, file '%s'", id, newVersion, targetPath)
	respChan <- nil
	return manager.SaveData()
}

func (manager * ImageManager) handleUpdateDiskImageProgress(id string, progress uint, respChan chan error) (err error){
	image, exists := manager.diskImages[id]
	if !exists{
		err = fmt.Errorf("invalid disk image '%s'", id)
		respChan <- err
		return err
	}
	if image.Created{
		err = fmt.Errorf("disk image '%s' already created", id)
		respChan <- err
		return err
	}
	if !image.Locked{
		err = fmt.Errorf("lock image '%s' before update", id)
		respChan <- err
		return err
	}
	if image.Progress != progress{
		image.Progress = progress
		manager.diskImages[id] = image
		log.Printf("<image> disk image '%s' updated to %d%%", image.Name, progress)
	}
	respChan <- nil
	return nil
}

func (manager *ImageManager) handleUnlockDiskImage(id string, respChan chan error) error{
	image, exists := manager.diskImages[id]
	if !exists{
		err := fmt.Errorf("invalid disk image '%s'", id)
		respChan <- err
		return err
	}
	if !image.Locked{
		err := fmt.Errorf("disk image '%s' not locked", id)
		respChan <- err
		return err
	}
	image.Locked = false
	manager.diskImages[id] = image
	log.Printf("<image> disk image '%s' unlocked", id)
	respChan <- nil
	return nil
}

func (manager * ImageManager) handleGetDiskImage(id string, respChan chan ImageResult) (err error){
	image, exists := manager.diskImages[id]
	if !exists{
		err = fmt.Errorf("invalid disk image '%s'", id)
		respChan <- ImageResult{Error:err}
		return err
	}
	respChan <- ImageResult{DiskImage:image}
	return nil
}

func (manager * ImageManager) handleGetDiskImageFile(id string, respChan chan ImageResult) error{
	image, exists := manager.diskImages[id]
	if !exists{
		err := fmt.Errorf("invalid disk image '%s'", id)
		respChan <- ImageResult{Error:err}
		return err
	}
	if 0 == image.Version{
		err := fmt.Errorf("no content available for disk image '%s'", id)
		respChan <- ImageResult{Error:err}
		return err
	}
	if image.Locked{
		err := fmt.Errorf("disk image '%s' is locked for update", id)
		respChan <- ImageResult{Error:err}
		return err
	}
	respChan <- ImageResult{Path:image.Path, Size:image.Size, CheckSum:image.CheckSum}
	return nil
}

func (manager * ImageManager) handleSyncMediaImages(owner, group string, respChan chan error) (err error){
	if "" == owner{
		err = errors.New("image owner required")
		respChan <- err
		return
	}
	if "" == group{
		err = errors.New("image group required")
		respChan <- err
		return
	}
	var existed = map[string]string{}
	for _, image := range manager.mediaImages{
		var name = fmt.Sprintf("%s_v%d", image.ID, image.Version)
		existed[name] = image.ID
	}
	var newFiles, lostID []string
	if newFiles, lostID, err = compareCurrentFiles(manager.mediaPath, DefaultMediaFormat, existed); err != nil{
		err = fmt.Errorf("find absent media images fail: %s", err.Error())
		respChan <- err
		return
	}
	if 0 == len(newFiles) && 0 == len(lostID){
		respChan <- nil
		log.Println("<image> all media images synchronized, no absent file discovered")
		return
	}
	for _, filename := range newFiles {
		var now = time.Now()
		var timestamp = now.Format(TimeFormatLayout)
		var image ImageStatus
		image.Owner = owner
		image.Group = group
		image.Name = fmt.Sprintf("%s_%d", filename, now.Unix())
		image.Description = fmt.Sprintf("generated by synchronize media images on %s", timestamp)
		image.ID = uuid.NewV4().String()
		image.Version = 1
		image.Tags = []string{}
		image.Format = DefaultMediaFormat
		image.Locked = false
		image.CreateTime = timestamp
		image.ModifyTime = timestamp
		image.Path = filepath.Join(manager.mediaPath, fmt.Sprintf("%s_v%d.%s", image.ID, image.Version, image.Format))
		var info os.FileInfo
		var sourceFile = filepath.Join(manager.mediaPath, fmt.Sprintf("%s.%s", filename, DefaultMediaFormat))
		if info, err = os.Stat(sourceFile); err != nil{
			err = fmt.Errorf("check source media file '%s' fail: %s", sourceFile, err.Error())
			respChan <- err
			return
		}
		image.Size = uint(info.Size())
		if err = os.Rename(sourceFile, image.Path); err != nil{
			err = fmt.Errorf("rename '%s' to '%s' fail: %s", sourceFile, image.Path, err.Error())
			respChan <- err
			return
		}
		var nameWithGroup = fmt.Sprintf("%s.%s", group, image.Name)
		manager.mediaImages[image.ID] = image
		manager.mediaImageNames[nameWithGroup] = true
		log.Printf("<image> synchronize %s to media image '%s'(%s)", filename, image.Name, image.ID)
	}
	if 0 != len(lostID){
		var image ImageStatus
		var exists bool
		for _, imageID := range lostID{
			if image, exists = manager.mediaImages[imageID]; !exists{
				log.Printf("<image> warning: found an invalid media image '%s'", imageID)
				continue
			}
			log.Printf("<image> remove invalid media image '%s'(id '%s')", image.Name, imageID)
			var nameWithGroup = fmt.Sprintf("%s.%s", image.Group, image.Name)
			delete(manager.mediaImageNames, nameWithGroup)
			delete(manager.mediaImages, imageID)
		}
	}
	respChan <- nil
	log.Printf("<image> %d new/ %d lost media image(s) synchronized", len(newFiles), len(lostID))
	return manager.SaveData()
}

func (manager * ImageManager) handleSyncDiskImages(owner, group string, respChan chan error) (err error){
	if "" == owner{
		err = errors.New("image owner required")
		respChan <- err
		return
	}
	if "" == group{
		err = errors.New("image group required")
		respChan <- err
		return
	}
	var existed = map[string]string{}
	for _, image := range manager.diskImages{
		var name = fmt.Sprintf("%s_v%d", image.ID, image.Version)
		existed[name] = image.ID
	}
	var newFiles, lostID []string
	if newFiles, lostID, err = compareCurrentFiles(manager.diskPath, DefaultDiskFormat, existed); err != nil{
		err = fmt.Errorf("find absent disk images fail: %s", err.Error())
		respChan <- err
		return
	}
	if 0 == len(newFiles) && 0 == len(lostID){
		respChan <- nil
		log.Println("<image> all disk images synchronized, no absent file discovered")
		return
	}
	for _, filename := range newFiles {
		var now = time.Now()
		var timestamp = now.Format(TimeFormatLayout)
		var image DiskStatus
		image.Owner = owner
		image.Group = group
		image.Name = fmt.Sprintf("%s_%d", filename, now.Unix())
		image.Description = fmt.Sprintf("generated by synchronize disk images on %s", timestamp)
		image.ID = uuid.NewV4().String()
		image.Version = 1
		image.Tags = []string{}
		image.Format = DefaultDiskFormat
		image.CreateTime = timestamp
		image.ModifyTime = timestamp
		image.Created = true
		image.Locked = false
		image.Path = filepath.Join(manager.diskPath, fmt.Sprintf("%s_v%d.%s", image.ID, image.Version, image.Format))
		var info os.FileInfo
		var sourceFile = filepath.Join(manager.diskPath, fmt.Sprintf("%s.%s", filename, DefaultDiskFormat))
		if info, err = os.Stat(sourceFile); err != nil{
			err = fmt.Errorf("check source disk file '%s' fail: %s", sourceFile, err.Error())
			respChan <- err
			return
		}
		image.Size = uint(info.Size())
		log.Printf("<image> compute checksum for '%s'...", sourceFile)
		if image.CheckSum, err = computeCheckSum(sourceFile); err != nil{
			err = fmt.Errorf("compute checksum for '%s' fail: %s", sourceFile, err.Error())
			respChan <- err
			return
		}
		if err = os.Rename(sourceFile, image.Path); err != nil{
			err = fmt.Errorf("rename '%s' to '%s' fail: %s", sourceFile, image.Path, err.Error())
			respChan <- err
			return
		}

		var nameWithGroup = fmt.Sprintf("%s.%s", group, image.Name)
		manager.diskImages[image.ID] = image
		manager.diskImageNames[nameWithGroup] = true
		log.Printf("<image> synchronize %s to disk image '%s'(%s)", filename, image.Name, image.ID)
	}
	if 0 != len(lostID){
		var image DiskStatus
		var exists bool
		for _, imageID := range lostID{
			if image, exists = manager.diskImages[imageID]; !exists{
				log.Printf("<image> warning: found an invalid disk image '%s'", imageID)
				continue
			}
			log.Printf("<image> remove invalid disk image '%s'(id '%s')", image.Name, imageID)
			var nameWithGroup = fmt.Sprintf("%s.%s", image.Group, image.Name)
			delete(manager.diskImageNames, nameWithGroup)
			delete(manager.diskImages, imageID)
		}
	}
	respChan <- nil
	log.Printf("<image> %d new/ %d lost disk image(s) synchronized", len(newFiles), len(lostID))
	return manager.SaveData()
}

func compareCurrentFiles(targetPath, ext string, existed map[string]string) (newFiles, lostID []string, err error){
	var suffix = fmt.Sprintf(".%s", ext)
	var targets = existed
	var exists bool
	err = filepath.Walk(targetPath, func(currentFile string, info os.FileInfo, accessErr error) error {
		if accessErr != nil{
			return fmt.Errorf("access '%s' fail: %s", currentFile, accessErr.Error())
		}
		if targetPath == currentFile{
			return nil
		}
		if info.IsDir(){
			return filepath.SkipDir
		}
		var base = filepath.Base(currentFile)
		if !strings.HasSuffix(base, suffix){
			return nil
		}
		var filename = strings.TrimSuffix(base, suffix)
		if _, exists = targets[filename]; !exists{
			newFiles = append(newFiles, filename)
		}else{
			delete(targets, filename)
		}
		return nil
	})
	for _, id := range targets{
		lostID = append(lostID, id)
	}
	return
}
