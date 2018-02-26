package agent

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/zbiljic/optic/internal"
	"github.com/zbiljic/optic/internal/config"
	"github.com/zbiljic/optic/internal/models"
	"github.com/zbiljic/optic/internal/selfmetric"
	"github.com/zbiljic/optic/optic"
)

// Agent runs optic and collects data based on the given config
type Agent struct {
	Config *config.Config
}

// NewAgent returns an Agent struct based off the given Config.
func NewAgent(config *config.Config) (*Agent, error) {
	a := &Agent{
		Config: config,
	}

	if !a.Config.Agent.OmitHostname {
		if a.Config.Agent.Hostname == "" {
			hostname, err := os.Hostname()
			if err != nil {
				return nil, err
			}

			a.Config.Agent.Hostname = hostname
		}

		config.Tags["host"] = a.Config.Agent.Hostname
	}

	return a, nil
}

// Connect connects to all configured sinks
func (a *Agent) Connect() error {
	for _, sink := range a.Config.Sinks {
		switch st := sink.Sink.(type) {
		case optic.ServiceSink:
			if err := st.Start(); err != nil {
				log.Printf("ERROR Service for sink '%s' failed to start, exiting\n%s\n",
					sink.Name(), err.Error())
				return err
			}
		}

		log.Printf("DEBUG Attempting connection to sink: %s\n", sink.Config.Name)

		err := sink.Sink.Connect()
		if err != nil {
			log.Printf("ERROR Failed to connect to sink %s, retrying in 15s, error was '%s' \n",
				sink.Name(), err)
			time.Sleep(15 * time.Second)
			err = sink.Sink.Connect()
			if err != nil {
				return err
			}
		}
		log.Printf("DEBUG Successfully connected to sink: %s\n", sink.Config.Name)
	}
	return nil
}

// Close closes the connection to all configured sinks
func (a *Agent) Close() error {
	var err error
	for _, s := range a.Config.Sinks {
		err = s.Sink.Close()
		switch st := s.Sink.(type) {
		case optic.ServiceSink:
			st.Stop()
		}
	}
	return err
}

func panicRecover(input *models.RunningSource) {
	if err := recover(); err != nil {
		trace := make([]byte, 2048)
		runtime.Stack(trace, true)
		log.Printf("FATAL Input [%s] panicked: %s, Stack:\n%s\n",
			input.Name(), err, trace)
	}
}

// Test verifies that we can 'Gather' from all sources with their configured
// Config struct
func (a *Agent) Test() error {
	shutdown := make(chan struct{})
	defer close(shutdown)
	eventCh := make(chan optic.Event)

	// dummy receiver for the point channel
	go func() {
		for {
			select {
			case <-eventCh:
				// do nothing
			case <-shutdown:
				return
			}
		}
	}()

	for _, source := range a.Config.Sources {
		if _, ok := source.Source.(optic.ServiceSource); ok {
			fmt.Printf("\nWARNING: skipping plugin [[%s]]: service sources not supported in test mode\n",
				source.Name())
			continue
		}

		acc := NewAccumulator(source, eventCh)
		source.SetTrace(true)
		source.SetDefaultTags(a.Config.Tags)

		fmt.Printf("* Plugin: %s, Collection 1\n", source.Name())
		if source.Config.Interval != 0 {
			fmt.Printf("* Internal: %s\n", source.Config.Interval)
		}

		if err := source.Source.Gather(acc); err != nil {
			return err
		}

		// Special instructions for some sources. cpu metric, for example, needs
		// to be run twice in order to return cpu usage percentages.
		time.Sleep(500 * time.Millisecond)
		fmt.Printf("* Plugin: %s, Collection 2\n", source.Name())
		if err := source.Source.Gather(acc); err != nil {
			return err
		}
	}

	return nil
}

// Run runs the agent daemon, gathering every Interval
func (a *Agent) Run(shutdown chan struct{}) error {
	var wg sync.WaitGroup

	log.Printf("INFO Agent Config: Interval:%s, Hostname:%#v, Flush Interval:%s \n",
		a.Config.Agent.Interval, a.Config.Agent.Hostname, a.Config.Agent.FlushInterval)

	// configure all sources
	for _, source := range a.Config.Sources {
		source.SetDefaultTags(a.Config.Tags)
	}

	// Start all ServiceSources
	for _, source := range a.Config.Sources {
		switch p := source.Source.(type) {
		case optic.ServiceSource:
			acc := NewAccumulator(source, source.EventsCh())
			if err := p.Start(acc); err != nil {
				log.Printf("ERROR Service for source %s failed to start, exiting\n%s\n",
					source.Name(), err.Error())
				return err
			}
			defer p.Stop()
		}
	}

	wg.Add(len(a.Config.Sources))
	for _, source := range a.Config.Sources {
		interval := a.Config.Agent.Interval
		// overwrite global interval if this plugin has it's own
		if source.Config.Interval != 0 {
			interval = source.Config.Interval
		}
		go func(source *models.RunningSource, interval time.Duration) {
			defer wg.Done()
			a.gatherer(shutdown, source, interval)
		}(source, interval)
	}

	wg.Wait()
	a.Close()
	return nil
}

// flush writes a list of events to all configured sinks.
func (a *Agent) flush() {
	var wg sync.WaitGroup

	wg.Add(len(a.Config.Processors))
	for _, p := range a.Config.Processors {
		go func(processor *models.RunningProcessor) {
			defer wg.Done()
			processor.Flush()
		}(p)
	}

	wg.Add(len(a.Config.Sinks))
	for _, s := range a.Config.Sinks {
		go func(sink *models.RunningSink) {
			defer wg.Done()
			err := sink.Write()
			if err != nil {
				log.Printf("ERROR Error writing to sink [%s]: %s",
					sink.Name(), err.Error())
			}
		}(s)
	}

	wg.Wait()
}

func (a *Agent) flusher(
	shutdown chan struct{},
	eventCh chan optic.Event,
) error {
	// Sleep for one interval before starting flush
	time.Sleep(a.Config.Agent.FlushInterval)

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-shutdown:
				if len(eventCh) > 0 {
					// keep going until eventCh is flushed
					continue
				}
				return
			}
		}
	}()

	ticker := time.NewTicker(a.Config.Agent.FlushInterval)
	semaphore := make(chan struct{}, 1)
	for {
		select {
		case <-shutdown:
			log.Println("INFO Flushing any cached events before shutdown")
			// wait for eventCh to get flushed before flushing sinks
			wg.Wait()
			a.flush()
			return nil
		case <-ticker.C:
			go func() {
				select {
				case semaphore <- struct{}{}:
					internal.RandomSleep(a.Config.Agent.FlushJitter, shutdown)
					a.flush()
					<-semaphore
				default:
					// skipping this flush because one is already happening
					log.Println("WARNING Skipping a scheduled flush because there is already a flush ongoing.")
				}
			}()
		}
	}
}

// gatherer runs the sources that have been configured with their own reporting
// interval.
func (a *Agent) gatherer(
	shutdown chan struct{},
	source *models.RunningSource,
	interval time.Duration,
) {
	defer panicRecover(source)

	GatherTime := selfmetric.GetOrRegisterHistogram(
		"gather",
		"gather_time_nanoseconds",
		map[string]string{"source": source.Config.Name},
	)

	eventCh := source.EventsCh()

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case event := <-eventCh:
				events := []optic.Event{event}
				for _, processor := range source.Config.Processors {
					events = processor.Processor.Apply(event)
					if len(events) == 0 {
						continue
					}
				}
				for _, e := range events {
					go func(event optic.Event) {
						source.ForwardEvent(event)
					}(e)
				}
			case <-shutdown:
				if len(eventCh) > 0 {
					// keep going until eventCh is flushed
					continue
				}
				return
			case <-ticker.C:
				continue
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := a.flusher(shutdown, eventCh); err != nil {
			log.Printf("ERROR Flusher routine failed, exiting: %s\n", err.Error())
		}
	}()

	acc := NewAccumulator(source, eventCh)

	for {
		internal.RandomSleep(a.Config.Agent.CollectionJitter, shutdown)

		start := time.Now()
		gatherWithTimeout(shutdown, source, acc, interval)
		elapsed := time.Since(start)

		GatherTime.Update(elapsed.Nanoseconds())

		select {
		case <-shutdown:
			log.Println("INFO Flushing any cached events before shutdown")
			// wait for eventCh to get flushed before flushing sinks
			wg.Wait()
			a.flush()
			return
		case <-ticker.C:
			continue
		}
	}
}

// gatherWithTimeout gathers from the given source, with the given timeout.
// when the given timeout is reached, gatherWithTimeout logs an error message
// but continues waiting for it to return. This is to avoid leaving behind
// hung processes, and to prevent re-calling the same hung process over and
// over.
func gatherWithTimeout(
	shutdown chan struct{},
	source *models.RunningSource,
	acc optic.Accumulator,
	timeout time.Duration,
) {

	ticker := time.NewTicker(timeout)
	defer ticker.Stop()
	done := make(chan error)
	go func() {
		done <- source.Source.Gather(acc)
	}()

	for {
		select {
		case err := <-done:
			if err != nil {
				acc.AddError(err)
			}
			return
		case <-ticker.C:
			err := fmt.Errorf("took longer to collect than collection interval (%s)",
				timeout)
			acc.AddError(err)
			continue
		case <-shutdown:
			return
		}
	}
}
