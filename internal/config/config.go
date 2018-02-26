package config

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/kr/pretty"
	"github.com/spf13/cast"
	"github.com/spf13/viper"

	"github.com/zbiljic/optic/internal/models"
	"github.com/zbiljic/optic/plugins/buffers"
	"github.com/zbiljic/optic/plugins/codecs"
	"github.com/zbiljic/optic/plugins/processors"
	"github.com/zbiljic/optic/plugins/sinks"
	"github.com/zbiljic/optic/plugins/sources"
)

const opticEnvironmentPrefix = "optic"

var (
	// envVarRegex is a regex to find environment variables in the config file
	envVarRegex = regexp.MustCompile(`^\$\w+$|^\$\{\w+\}$`)

	envVarEscaper = strings.NewReplacer(
		`"`, `\"`,
		`\`, `\\`,
	)
)

// Config provides a container with configuration parameters for Optic
type Config struct {
	Tags map[string]string `mapstructure:"global_tags"`

	Agent *AgentConfig `mapstructure:"agent"`

	Sources    map[string]*models.RunningSource    `mapstructure:"-"`
	Processors map[string]*models.RunningProcessor `mapstructure:"-"`
	Sinks      map[string]*models.RunningSink      `mapstructure:"-"`
}

func NewConfig() *Config {
	c := &Config{
		// Agent defaults:
		Agent: &AgentConfig{
			ThreadCount:   0,
			Interval:      10 * time.Second,
			FlushInterval: 10 * time.Second,
		},
		Sources:    make(map[string]*models.RunningSource),
		Processors: make(map[string]*models.RunningProcessor),
		Sinks:      make(map[string]*models.RunningSink),
	}
	return c
}

type AgentConfig struct {
	// Number of threads to be used by the agent process.
	// Should be in the range of [1, runtime.NumCPU()).
	ThreadCount int `mapstructure:"thread_count"`

	// Interval at which to gather information.
	Interval time.Duration `mapstructure:"interval"`

	// CollectionJitter is used to jitter the collection by a random amount.
	// Each plugin will sleep for a random time within jitter before collecting.
	// This can be used to avoid many plugins querying things like sysfs at the
	// same time, which can have a measurable effect on the system.
	CollectionJitter time.Duration `mapstructure:"collection_jitter"`

	// Default flushing interval for all sinks. You shouldn't set this below
	// interval. Maximum flush_interval will be flush_interval + flush_jitter.
	FlushInterval time.Duration `mapstructure:"flush_interval"`

	// FlushJitter Jitters the flush interval by a random amount.
	// This is primarily to avoid large write spikes for users running a large
	// number of Optic instances.
	// ie, a jitter of 5s and interval 10s means flushes will happen every 10-15s
	FlushJitter time.Duration `mapstructure:"flush_jitter"`

	// Override default hostname, if empty use os.Hostname().
	Hostname string `mapstructure:"hostname"`
	// If set to true, do no set the "host" tag in the Optic agent.
	OmitHostname bool `mapstructure:"omit_hostname"`
}

// SourceNames returns a list of strings of the configured sources.
func (c *Config) SourceNames() []string {
	var name []string
	for _, source := range c.Sources {
		name = append(name, source.Name())
	}
	return name
}

// ProcessorNames returns a list of strings of the configured processors.
func (c *Config) ProcessorNames() []string {
	var name []string
	for _, processor := range c.Processors {
		name = append(name, processor.Name())
	}
	return name
}

// SinkNames returns a list of strings of the configured sinks.
func (c *Config) SinkNames() []string {
	var name []string
	for _, sink := range c.Sinks {
		name = append(name, sink.Name())
	}
	return name
}

// GlobalTags returns a string of tags specified in the config.
func (c *Config) GlobalTags() string {
	var tags []string

	for k, v := range c.Tags {
		tags = append(tags, fmt.Sprintf("%s=%s", k, v))
	}
	sort.Strings(tags)

	return strings.Join(tags, " ")
}

// Try to find a default config file at these locations (in order):
//   1. $OPTIC_CONFIG_PATH
//   2. $HOME/.optic/optic.conf
//   3. /etc/optic/optic.conf
//
func getDefaultConfigPath() (string, error) {
	envfile := os.Getenv("OPTIC_CONFIG_PATH")
	homefile := os.ExpandEnv("${HOME}/.optic/optic.conf")
	etcfile := "/etc/optic/optic.conf"
	if runtime.GOOS == "windows" {
		etcfile = `C:\Program Files\Optic\optic.conf`
	}
	for _, path := range []string{envfile, homefile, etcfile} {
		if _, err := os.Stat(path); err == nil {
			log.Printf("INFO Using config file: %s", path)
			return path, nil
		}
	}

	// if we got here, we didn't find a file in a default location
	return "", fmt.Errorf("No config file specified, and could not find one"+
		" in $OPTIC_CONFIG_PATH, %s, or %s", homefile, etcfile)
}

// LoadConfig loads the given config file and applies it to c.
func (c *Config) LoadConfig(path string) error {
	var err error
	if path == "" {
		if path, err = getDefaultConfigPath(); err != nil {
			return err
		}
	}
	err = readViperConfig(path)
	if err != nil {
		return fmt.Errorf("Error reading %s, %s", path, err)
	}

	// NOTE: error is ignored from unmarshal, some fields will be extracted
	// manually
	viper.Unmarshal(c)

	builder := newPluginBuilder()
	defer builder.stop()

	// Parse all the plugins
	settings := viper.AllSettings()
	for name, value := range settings {
		subMap, err := cast.ToStringMapE(value)
		if err != nil {
			return fmt.Errorf("%s: invalid configuration", path)
		}

		var pluginType string

		switch name {
		case "global", "agent":
			// should already be extracted during unmarshal
			continue
		case "sources":
			pluginType = "source"
		case "processors":
			pluginType = "processor"
		case "sinks":
			pluginType = "sink"
		default:
			continue
		}

		if pluginType != "" {
			for pluginName, pluginValue := range subMap {
				pluginConfig, err := cast.ToStringMapE(pluginValue)
				if err != nil {
					return fmt.Errorf("Unsupported config format: %s, file %s",
						pluginName, path)
				}
				// queue build
				builder.addPlugin(&plugin{
					pluginType: pluginType,
					name:       pluginName,
					config:     pluginConfig,
				})
			}
		}
	}

	// goroutine for building configuration
	go func() {
		for {
			var (
				plugin *plugin
				err    error
			)
			select {
			case plugin = <-builder.ch:
				switch plugin.pluginType {
				case "source":
					err = c.addSource(builder, plugin.name, plugin.config)
					break
				case "processor":
					err = c.addProcessor(builder, plugin.name, plugin.config)
					break
				case "sink":
					err = c.addSink(builder, plugin.name, plugin.config)
					break
				}
			default:
				// no more
				return
			}
			if err != nil {
				if err == errPluginReferenceNotFound {
					// SPECIAL CASE: requires another plugin to be build before this one
					if !plugin.buildAttemptsLimitReached() {
						plugin.buildAttempts++
						builder.ch <- plugin
						continue
					}
					err = fmt.Errorf("Possible circular reference, %s", err)
				}
				builder.err = fmt.Errorf("Error parsing %s, %s", path, err)
				builder.endCh <- struct{}{}
				builder.wg.Done()
				return
			}
			builder.wg.Done()
		}
	}()

	// goroutine for stoping build
	go func() {
		var err error
		select {
		case <-builder.endCh:
			if builder.err != nil {
				err = builder.err
				break
			}
			// protection against infinite configuration
		case <-time.After(time.Second * 2):
			err = fmt.Errorf("Timeout reached while building configuration, remaining: %d",
				len(builder.ch))
			builder.err = err
			log.Println("FATAL", err)
			break
		}
		if err != nil {
			doneDelta := len(builder.ch)
			if doneDelta == 0 {
				builder.wg.Done()
			} else {
				builder.wg.Add(-doneDelta)
			}
		}
	}()

	builder.wg.Wait()

	if builder.err != nil {
		return builder.err
	}

	return err
}

func readViperConfig(path string) error {

	viper.SetEnvPrefix(opticEnvironmentPrefix) // will be uppercased automatically
	viper.AutomaticEnv()                       // read in environment variables that match

	viper.SetConfigFile(path)

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err != nil {
		return err
	}
	log.Println("TRACE", "Read config file:", viper.ConfigFileUsed())

	replaceEnvVars()

	settings := viper.AllSettings()
	log.Printf("TRACE %# v", pretty.Formatter(settings))

	return nil
}

// escapeEnv escapes a value for inserting into config.
func escapeEnv(value string) string {
	return envVarEscaper.Replace(value)
}

func replaceEnvVars() {
	for _, key := range viper.AllKeys() {
		value := viper.GetString(key)
		if value != "" {
			isEnvVar := envVarRegex.MatchString(value)
			if isEnvVar {
				envVar := value
				envVar = strings.TrimPrefix(envVar, "$")
				envVar = strings.TrimPrefix(envVar, "{")
				envVar = strings.TrimSuffix(envVar, "}")
				envValue, ok := os.LookupEnv(envVar)
				if ok {
					envValue = escapeEnv(envValue)
					log.Printf("TRACE Replaced environment variable '%s' to '%s'",
						value, envValue)
					viper.Set(key, envValue)
				}
			}
		}
	}
}

func (c *Config) addSource(
	builder *pluginBuilder,
	name string,
	config map[string]interface{},
) error {
	if _, ok := c.Sources[name]; ok {
		return fmt.Errorf("Cannot have multiple sources with the same name: %s", name)
	}

	kind, err := cast.ToStringE(config["kind"])
	if err != nil || kind == "" {
		return fmt.Errorf("Undefined source kind for: %s", name)
	}

	creator, ok := sources.Sources[kind]
	if !ok {
		return fmt.Errorf("Undefined but requested source kind: %s", kind)
	}
	source := creator()

	pluginConfig, err := c.buildSourceConfig(builder, kind, name, config)
	if err != nil {
		return err
	}

	// unmarshal configuration for concrete plugin
	var lv = viper.New()
	lv.Set("config", config)
	lv.UnmarshalKey("config", source)

	rs := models.NewRunningSource(source, pluginConfig)
	c.Sources[rs.Name()] = rs
	return nil
}

func (c *Config) addProcessor(
	builder *pluginBuilder,
	name string,
	config map[string]interface{},
) error {
	if _, ok := c.Processors[name]; ok {
		return fmt.Errorf("Cannot have multiple processors with the same name: %s", name)
	}

	kind, err := cast.ToStringE(config["kind"])
	if err != nil || kind == "" {
		return fmt.Errorf("Undefined processor kind for: %s", name)
	}

	creator, ok := processors.Processors[kind]
	if !ok {
		return fmt.Errorf("Undefined but requested processor kind: %s", kind)
	}
	processor := creator()

	pluginConfig, err := c.buildProcessorConfig(builder, kind, name, config)
	if err != nil {
		return err
	}

	// unmarshal configuration for concrete plugin
	var lv = viper.New()
	lv.Set("config", config)
	lv.UnmarshalKey("config", processor)

	// initialize processor
	if err := processor.Init(); err != nil {
		return err
	}

	rf := models.NewRunningProcessor(processor, pluginConfig)
	c.Processors[rf.Name()] = rf
	return nil
}

func (c *Config) addSink(
	builder *pluginBuilder,
	name string,
	config map[string]interface{},
) error {
	if _, ok := c.Sinks[name]; ok {
		return fmt.Errorf("Cannot have multiple sinks with the same name: %s", name)
	}

	kind, err := cast.ToStringE(config["kind"])
	if err != nil || kind == "" {
		return fmt.Errorf("Undefined sink kind for: %s", name)
	}

	creator, ok := sinks.Sinks[kind]
	if !ok {
		return fmt.Errorf("Undefined but requested sink kind: %s", kind)
	}
	sink := creator()

	pluginConfig, err := c.buildSinkConfig(builder, kind, name, config)
	if err != nil {
		return err
	}

	// unmarshal configuration for concrete plugin
	var lv = viper.New()
	lv.Set("config", config)
	lv.UnmarshalKey("config", sink)

	rs := models.NewRunningSink(sink, pluginConfig)
	c.Sinks[rs.Name()] = rs
	return nil
}

func (c *Config) buildSourceConfig(
	builder *pluginBuilder,
	kind string,
	name string,
	config map[string]interface{},
) (*models.SourceConfig, error) {
	log.Printf("TRACE Building source config: '%s' (%s)", name, kind)

	conf := &models.SourceConfig{Kind: kind, Name: name}

	// interval - OPTIONAL
	if node, ok := config["interval"]; ok {
		dur, err := cast.ToDurationE(node)
		if err != nil {
			return nil, err
		}

		conf.Interval = dur
	}

	// tags - OPTIONAL
	conf.Tags = make(map[string]string)
	if node, ok := config["tags"]; ok {
		tagMap, err := cast.ToStringMapStringE(node)
		if err != nil {
			return nil, fmt.Errorf("Unable to parse tags for source '%s': '%s'", name, err)
		}
		for k, v := range tagMap {
			conf.Tags[k] = v
		}
	}

	// processors - OPTIONAL
	conf.Processors = make([]*models.RunningProcessor, 0)
	if node, ok := config["processors"]; ok {
		processors := cast.ToSlice(node)
		for _, processorConfig := range processors {
			switch v := processorConfig.(type) {
			case string:
				// processor reference, connect it
				if processor, ok := c.Processors[v]; ok {
					conf.Processors = append(conf.Processors, processor)
					break
				}
				log.Printf("TRACE Required processor '%s' not found", v)
				return nil, errPluginReferenceNotFound
			default:
				return nil, fmt.Errorf("Unable to parse processors for source '%s', type: %s",
					name, v)
			}
		}
	}

	// forwards - REQUIRED
	conf.ForwardProcessors = make([]*models.RunningProcessor, 0)
	conf.ForwardSinks = make([]*models.RunningSink, 0)
	if node, ok := config["forwards"]; ok {
		forwards := cast.ToSlice(node)
		for _, forwardConfig := range forwards {
			switch v := forwardConfig.(type) {
			case string:
				// processor reference, connect it
				if processor, ok := c.Processors[v]; ok {
					conf.ForwardProcessors = append(conf.ForwardProcessors, processor)
					break
				}
				// sink reference, connect it
				if sink, ok := c.Sinks[v]; ok {
					conf.ForwardSinks = append(conf.ForwardSinks, sink)
					break
				}
				log.Printf("TRACE Required forward '%s' not found", v)
				return nil, errPluginReferenceNotFound
			default:
				return nil, fmt.Errorf("Unable to parse forwards for source '%s', type: %s",
					name, v)
			}
		}
	}

	// codec - OPTIONAL
	if codecConfig, ok := config["codec"]; ok {
		codecConfigMap, err := cast.ToStringMapE(codecConfig)
		if err != nil {
			return nil, fmt.Errorf("Unable to parse codec for source '%s': %s", name, err)
		}
		codec, err := codecs.NewCodec(codecConfigMap)
		if err != nil {
			return nil, fmt.Errorf("Unable to create codec for source '%s': %s", name, err)
		}

		conf.Decoder = codec
	}

	delete(config, "kind")
	delete(config, "interval")
	delete(config, "tags")
	delete(config, "processors")
	delete(config, "forwards")
	delete(config, "codec")

	return conf, nil
}

func (c *Config) buildProcessorConfig(
	builder *pluginBuilder,
	kind string,
	name string,
	config map[string]interface{},
) (*models.ProcessorConfig, error) {
	log.Printf("TRACE Building processor config: '%s' (%s)", name, kind)

	conf := &models.ProcessorConfig{Kind: kind, Name: name}

	// forwards - OPTIONAL
	conf.ForwardProcessors = make([]*models.RunningProcessor, 0)
	conf.ForwardSinks = make([]*models.RunningSink, 0)
	if node, ok := config["forwards"]; ok {
		forwards := cast.ToSlice(node)
		for _, forwardConfig := range forwards {
			switch v := forwardConfig.(type) {
			case string:
				// processor reference, connect it
				if processor, ok := c.Processors[v]; ok {
					conf.ForwardProcessors = append(conf.ForwardProcessors, processor)
					break
				}
				// sink reference, connect it
				if sink, ok := c.Sinks[v]; ok {
					conf.ForwardSinks = append(conf.ForwardSinks, sink)
					break
				}
				log.Printf("TRACE Required forward '%s' not found", v)
				return nil, errPluginReferenceNotFound
			default:
				return nil, fmt.Errorf("Unable to parse forwards for processor '%s', type: %s",
					name, v)
			}
		}
	}

	delete(config, "kind")
	delete(config, "forwards")

	return conf, nil
}

func (c *Config) buildSinkConfig(
	builder *pluginBuilder,
	kind string,
	name string,
	config map[string]interface{},
) (*models.SinkConfig, error) {
	log.Printf("TRACE Building sink config: '%s' (%s)", name, kind)

	conf := &models.SinkConfig{Kind: kind, Name: name}

	// batch_size - OPTIONAL
	if node, ok := config["batch_size"]; ok {
		batchSize, err := cast.ToIntE(node)
		if err != nil {
			return nil, err
		}

		conf.EventBatchSize = batchSize
	}

	// buffer - OPTIONAL
	if bufferConfig, ok := config["buffer"]; ok {
		bufferConfigMap, err := cast.ToStringMapE(bufferConfig)
		if err != nil {
			return nil, fmt.Errorf("Unable to parse buffer for source '%s': %s", name, err)
		}
		buffer, err := buffers.NewBuffer(bufferConfigMap)
		if err != nil {
			return nil, fmt.Errorf("Unable to create buffer for source '%s': %s", name, err)
		}

		conf.Buffer = buffer
	} else {
		log.Printf("DEBUG Using default buffer for sink: '%s' (%s)", name, kind)
		buffer, err := buffers.NewDefaultBuffer()
		if err != nil {
			return nil, err
		}
		conf.Buffer = buffer
	}

	// codec - OPTIONAL
	if codecConfig, ok := config["codec"]; ok {
		codecConfigMap, err := cast.ToStringMapE(codecConfig)
		if err != nil {
			return nil, fmt.Errorf("Unable to parse codec for source '%s': %s", name, err)
		}
		codec, err := codecs.NewCodec(codecConfigMap)
		if err != nil {
			return nil, fmt.Errorf("Unable to create codec for source '%s': %s", name, err)
		}

		conf.Encoder = codec
	}

	delete(config, "kind")
	delete(config, "batch_size")
	delete(config, "buffer")
	delete(config, "codec")

	return conf, nil
}
