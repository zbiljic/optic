package plugins

import (
	_ "github.com/zbiljic/optic/plugins/buffers/all"    // load buffers
	_ "github.com/zbiljic/optic/plugins/codecs/all"     // load codecs
	_ "github.com/zbiljic/optic/plugins/processors/all" // load processors
	_ "github.com/zbiljic/optic/plugins/sinks/all"      // load sinks
	_ "github.com/zbiljic/optic/plugins/sources/all"    // load sources
)
