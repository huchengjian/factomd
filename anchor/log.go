// Copyright 2015 FactomProject Authors. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

package anchor

import (
	"os"

	"github.com/FactomProject/factomd/logger"
	"github.com/FactomProject/factomd/util"
)

var (
	logcfg     = util.ReadConfig("", "").Log
	logPath    = logcfg.LogPath
	logLevel   = logcfg.LogLevel
	logfile, _ = os.OpenFile(logPath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0660)
)

// setup subsystem loggers
var (
	anchorLog = logger.New(logfile, logLevel, "ANCH")
)
