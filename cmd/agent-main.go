package cmd

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"
	"syscall"

	"github.com/kardianos/service"

	"github.com/zbiljic/optic/agent"
	"github.com/zbiljic/optic/internal/config"
	"github.com/zbiljic/optic/logger"
	_ "github.com/zbiljic/optic/plugins" // load all plugins
)

func agentMain() error {

	startProfilerServerIfConfigured()

	if globalOsIsWindows {
		svcConfig := &service.Config{
			Name:        "optic",
			DisplayName: "Optic Data Collector Service",
			Description: "Collects data using a series of plugins and forwards it to another series of plugins.",
			Arguments:   []string{"-config", "C:\\Program Files\\Optic\\optic.conf"},
		}

		prg := &program{}
		s, err := service.New(prg, svcConfig)
		if err != nil {
			fatalIf(err, "Failed to create service:")
		}
		// Handle the --service flag here to prevent any issues with tooling that
		// may not have an interactive session, e.g. installing from Ansible.
		if globalServiceCommand != "" {
			if globalConfigFile {
				(*svcConfig).Arguments = []string{"-config", globalConfig}
			}
			err := service.Control(s, globalServiceCommand)
			if err != nil {
				fatalIf(err, "Service control failed:")
			}
			os.Exit(0)
		} else {
			err = s.Run()
			if err != nil {
				log.Println("ERROR", err.Error())
			}
		}
	} else {
		stop = make(chan struct{})
		return reloadLoop(stop)
	}

	return nil
}

var stop chan struct{}

var activeThreadCount = 0

type program struct{}

func (p *program) Start(_ service.Service) error {
	go p.run()
	return nil
}

func (p *program) run() {
	stop = make(chan struct{})
	reloadLoop(stop)
}

func (p *program) Stop(_ service.Service) error {
	close(stop)
	return nil
}

func reloadLoop(stop chan struct{}) error {
	reload := make(chan bool, 1)
	reload <- true
	for <-reload {
		reload <- false

		// If no other options are specified, load the config file and run.
		c := config.NewConfig()

		err := c.LoadConfig(globalConfig)
		if err != nil {
			errorIf(err, "Failed to load configuration:")
			return exitStatus(globalErrorExitStatus)
		}

		if len(c.Sources) == 0 {
			errorIf(errDummy(), "No sources found, did you provide a valid config file?")
			return exitStatus(globalErrorExitStatus)
		}
		if len(c.Sinks) == 0 {
			errorIf(errDummy(), "No sinks found, did you provide a valid config file?")
			return exitStatus(globalErrorExitStatus)
		}

		if int64(c.Agent.Interval) <= 0 {
			errorIf(errDummy(), fmt.Sprintf("Agent interval must be positive, found %s",
				c.Agent.Interval))
			return exitStatus(globalErrorExitStatus)
		}

		if int64(c.Agent.FlushInterval) <= 0 {
			errorIf(errDummy(), fmt.Sprintf("Agent flush_interval must be positive, found %s",
				c.Agent.FlushInterval))
			return exitStatus(globalErrorExitStatus)
		}

		// limit number of operating system threads
		if c.Agent.ThreadCount > 0 {
			if activeThreadCount != c.Agent.ThreadCount {
				activeThreadCount = c.Agent.ThreadCount
				log.Printf("DEBUG Update number of operating system threads used to: %d",
					activeThreadCount)
				runtime.GOMAXPROCS(activeThreadCount)
			}
		}

		// Setup logging again, due to possible logfile update
		updateGlobals()
		logger.SetupLogging(globalDebug, globalQuiet, globalLogfile)

		ag, err := agent.NewAgent(c)
		if err != nil {
			log.Fatal("ERROR ", err.Error())
		}

		if globalTestCommand {
			err = ag.Test()
			if err != nil {
				log.Fatal("ERROR ", err.Error())
			}
			return nil
		}

		err = ag.Connect()
		if err != nil {
			log.Fatal("ERROR ", err.Error())
		}

		shutdown := make(chan struct{})

		trapCh := signalTrap(os.Interrupt, syscall.SIGHUP)
		go func() {
			select {
			case sig := <-trapCh:
				if sig == os.Interrupt {
					close(shutdown)
				}
				if sig == syscall.SIGHUP {
					log.Println("INFO Reloading Optic config")
					<-reload
					reload <- true
					close(shutdown)
				}
			case <-stop:
				close(shutdown)
			}
		}()

		log.Printf("INFO Starting Optic %s", Version)
		log.Printf("INFO Loaded sources: %s", strings.Join(c.SourceNames(), " "))
		log.Printf("INFO Loaded processors: %s", strings.Join(c.ProcessorNames(), " "))
		log.Printf("INFO Loaded sinks: %s", strings.Join(c.SinkNames(), " "))
		log.Printf("INFO Global tags: %s", c.GlobalTags())

		ag.Run(shutdown)
	}

	return nil
}

// Check the interfaces are satisfied
var (
	_ service.Interface = &program{}
)
