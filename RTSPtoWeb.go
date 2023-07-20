package main

import (
	"fmt"
	"os"
   "os/signal"
	"path/filepath"
	"strings"
	"time"
   "syscall"

	"github.com/kardianos/service"
	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/debug"
	"golang.org/x/sys/windows/svc/eventlog"
)

const serviceName = "RTSPtoWebService"
const serviceDescription = "RTSPtoWebService"

var elog debug.Log

type program struct{}

func (p program) Start(s service.Service) error {
	fmt.Println(s.String() + " started")
	go p.run()
	return nil
}

func (p program) Stop(s service.Service) error {
	fmt.Println(s.String() + " stopped")
	return nil
}

func (p program) run() {

	// for {
	// 	elog.Info(1, "Service is running")
	// 	time.Sleep(1 * time.Second)
	// }

   log.Info("test")

	elog.Info(1, "Server CORE start")
	go HTTPAPIServer()
	go RTSPServer()
	go Storage.StreamChannelRunAll()
	signalChanel := make(chan os.Signal, 1)
	done := make(chan bool, 1)
	signal.Notify(signalChanel, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		// sig := <-signalChanel
		elog.Info(1, "Server receive signal")
		done <- true
	}()
	elog.Info(1, "Server start success a wait signals")
	<-done
	Storage.StopAll()
	time.Sleep(2 * time.Second)
	elog.Info(1, "Server stop working by signal")
}

func usage(errmsg string) {
	fmt.Fprintf(os.Stderr,
		"%s\n\n"+
			"usage: %s <command>\n"+
			"       where <command> is one of\n"+
			"       install, remove, debug, start, stop\n",
		errmsg, os.Args[0])
	os.Exit(2)
}

func main() {
	exepath, err1 := exePath()
	if err1 != nil {
		fmt.Println("Cannot start the service: " + err1.Error())
	}
	dir, _ := filepath.Split(exepath)
	configPath := dir + "config.json"
	argumentConfigPath := "-config=" + configPath

	serviceConfig := &service.Config{
		Name:        serviceName,
		DisplayName: serviceName,
		Description: serviceDescription,
		Arguments:   []string{argumentConfigPath},
	}
	prg := &program{}
	s, err := service.New(prg, serviceConfig)
	if err != nil {
		fmt.Println("Cannot create the service: " + err.Error())
	}

	inService, err := svc.IsWindowsService()
	if err != nil {
		log.Fatalf("failed to determine if we are running in service: %v", err)
	}
	if inService {
		elog, err = eventlog.Open(serviceName)
		if err != nil {
			return
		}

		elog.Info(1, fmt.Sprintf("configPath: %s", configPath))

		err = s.Run()
		if err != nil {
			elog.Error(1, fmt.Sprintf("%s service failed: %v", serviceName, err))
		}
		return
	} else {
      elog = debug.New(serviceName)
   }

	if len(os.Args) < 1 {
		usage("no command specified")
	}

	cmd := strings.ToLower(os.Args[1])
	// configPath := "config.json"
	// if len(os.Args) == 3 {
	// 	configPath = strings.ToLower(os.Args[2])
	// }

	switch cmd {
	case "debug":
		err = s.Run()
		if err != nil {
			fmt.Println("Cannot start the service: " + err.Error())
		}
		return
	case "install":
		err = s.Install()
		// err = installService(svcName, svcName, configPath)
	case "remove":
		err = s.Uninstall()
		// err = removeService(svcName)
	case "start":
		err = s.Start()
		// err = startService(svcName)
	case "stop":
		err = s.Stop()
		// err = controlService(svcName, svc.Stop, svc.Stopped)
	default:
		usage(fmt.Sprintf("invalid command %s", cmd))
	}
	if err != nil {
		log.Fatalf("failed to %s %s: %v", cmd, serviceName, err)
	}
	return
}
