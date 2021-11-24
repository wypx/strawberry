package vm_admin2

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"
	"vm_manager/vm_utils"
)

type FrontEndConfig struct {
	ListenAddress string `json:"address"`
	ListenPort    int    `json:"port"`
	ServiceHost   string `json:"service_host"`
	ServicePort   int    `json:"service_port"`
	APIKey        string `json:"api_key"`
	APIID         string `json:"api_id"`
	WebRoot       string `json:"web_root"`
	CORSEnable    bool   `json:"cors_enable,omitempty"`
}

type MainService struct {
	frontend *FrontEndService
}

const (
	ExecuteName    = "frontend"
	ConfigFileName = "frontend.cfg"
	ConfigPathName = "config"
	WebRootName    = "web_root"
	DataPathName   = "data"
)

func (service *MainService) Start() (output string, err error) {
	if nil == service.frontend {
		err = errors.New("invalid service")
		return
	}
	if err = service.frontend.Start(); err != nil {
		return
	}
	output = fmt.Sprintf("Front-End Module %s\nCore API: %s\nNano Web Portal: http://%s\n",
		service.frontend.GetVersion(),
		service.frontend.GetBackendURL(),
		service.frontend.GetListenAddress())
	return
}

func (service *MainService) Stop() (output string, err error) {
	if nil == service.frontend {
		err = errors.New("invalid service")
		return
	}
	err = service.frontend.Stop()
	return
}

func (service *MainService) Snapshot() (output string, err error) {
	fmt.Printf("Snapshot interface not implement\n")
	return
}

func generateConfigure(workingPath string) (err error) {
	const (
		DefaultPathPerm = 0740
	)
	var configPath = filepath.Join(workingPath, ConfigPathName)
	if _, err = os.Stat(configPath); os.IsNotExist(err) {
		//create path
		err = os.Mkdir(configPath, DefaultPathPerm)
		if err != nil {
			return
		}
		fmt.Printf("config path %s created\n", configPath)
	}

	var configFile = filepath.Join(configPath, ConfigFileName)
	if _, err = os.Stat(configFile); os.IsNotExist(err) {
		fmt.Println("No configures available, following instructions to generate a new one.")
		const (
			DefaultConfigPerm   = 0640
			DefaultBackEndPort  = 5850
			DefaultFrontEndPort = 5870
		)
		// var defaultWebRoot = filepath.Join(workingPath, WebRootName)
		var defaultWebRoot = "/root/work/strawberry/vm_admin2/web_root"
		var config = FrontEndConfig{}
		if config.ListenAddress, err = vm_utils.ChooseIPV4Address("Portal listen address"); err != nil {
			return
		}
		if config.ListenPort, err = vm_utils.InputInteger("Portal listen port", DefaultFrontEndPort); err != nil {
			return
		}
		if config.ServiceHost, err = vm_utils.InputString("Backend service address", config.ListenAddress); err != nil {
			return
		}
		if config.ServicePort, err = vm_utils.InputInteger("Backend service port", DefaultBackEndPort); err != nil {
			return
		}
		fmt.Println("defaultWebRoot: " + defaultWebRoot)
		if config.WebRoot, err = vm_utils.InputString("Web Root Path", defaultWebRoot); err != nil {
			return
		}
		//write
		data, err := json.MarshalIndent(config, "", " ")
		if err != nil {
			return err
		}
		if err = ioutil.WriteFile(configFile, data, DefaultConfigPerm); err != nil {
			return err
		}
		fmt.Printf("default configure '%s' generated\n", configFile)
	}

	var dataPath = filepath.Join(workingPath, DataPathName)
	if _, err = os.Stat(dataPath); os.IsNotExist(err) {
		//create path
		err = os.Mkdir(dataPath, DefaultPathPerm)
		if err != nil {
			return
		}
		fmt.Printf("data path %s created\n", dataPath)
	}
	return
}

func createDaemon(workingPath string) (service vm_utils.DaemonizedService, err error) {
	var configPath = filepath.Join(workingPath, ConfigPathName)
	var dataPath = filepath.Join(workingPath, DataPathName)
	if _, err = os.Stat(configPath); os.IsNotExist(err) {
		err = fmt.Errorf("config path %s not available", configPath)
		return nil, err
	}
	if _, err = os.Stat(dataPath); os.IsNotExist(err) {
		err = fmt.Errorf("data path %s not available", dataPath)
		return nil, err
	}
	var s = MainService{}
	s.frontend, err = CreateFrontEnd(configPath, dataPath)
	return &s, err
}

func getWorkingPath() (path string, err error) {
	executable, err := os.Executable()
	if err != nil {
		return "", err
	}
	return filepath.Abs(filepath.Dir(executable))
}

func Initialize() {
	// vm_utils.ProcessDaemon(ExecuteName, generateConfigure, createDaemon)
	workingPath, err := getWorkingPath()
	if err != nil {
		fmt.Printf("get working path fail: %s\n", err.Error())
		return
	}
	if err := generateConfigure(workingPath); err != nil {
		fmt.Printf("generate config fail: %s\n", err.Error())
		return
	}
	daemonizedService, err := createDaemon(workingPath)
	if daemonizedService == nil || err != nil {
		log.Printf("generate service fail: %s", err.Error())
		return
	}
	log.Printf("vm admin started\n")

	// Wait for interrupt signal to gracefully shutdown the server with
	// a timeout of 5 seconds.
	quit := make(chan os.Signal, 1)
	// kill (no param) default send syscall.SIGTERM
	// kill -2 is syscall.SIGINT
	// kill -9 is syscall.SIGKILL but can't be catch, so don't need add it
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// The context is used to inform the server it has 5 seconds to finish
	// the request it is currently handling
	_, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	log.Println("Server exiting")
}
