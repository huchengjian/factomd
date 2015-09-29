// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package daemon

import (
	"fmt"
	"github.com/FactomProject/factomd/btcd"
	"github.com/FactomProject/factomd/btcd/limits"
	"github.com/FactomProject/factomd/btcd/wire"
	. "github.com/FactomProject/factomd/common/constants"
	cp "github.com/FactomProject/factomd/controlpanel"
	"github.com/FactomProject/factomd/database"
	"github.com/FactomProject/factomd/process"
	"github.com/FactomProject/factomd/state"
	"github.com/FactomProject/factomd/state/stateinit"
	"github.com/FactomProject/factomd/util"
	"github.com/FactomProject/factomd/wsapi"
	"os"
	"runtime"
	"strings"
	"time"
)

// winServiceMain is only invoked on Windows.  It detects when btcd is running
// as a service and reacts accordingly.
//var winServiceMain func() (bool, error)

func RunDaemon() {
	ftmdLog.Info("//////////////////////// Copyright 2015 Factom Foundation")
	ftmdLog.Info("//////////////////////// Use of this source code is governed by the MIT")
	ftmdLog.Info("//////////////////////// license that can be found in the LICENSE file.")

	ftmdLog.Warning("Go compiler version: %s", runtime.Version())
	fmt.Println("Go compiler version: ", runtime.Version())
	cp.CP.AddUpdate("gocompiler",
		"system",
		fmt.Sprintln("Go compiler version: ", runtime.Version()),
		"",
		0)
	cp.CP.AddUpdate("copyright",
		"system",
		"Legal",
		"Copyright 2015 Factom Foundation\n"+
			"Use of this source code is governed by the MIT\n"+
			"license that can be found in the LICENSE file.",
		0)

	if !isCompilerVersionOK() {
		for i := 0; i < 30; i++ {
			fmt.Println("!!! !!! !!! ERROR: unsupported compiler version !!! !!! !!!")
		}
		time.Sleep(time.Second)
		os.Exit(1)
	}

	s := state.InitState()

	// Use all processor cores.
	runtime.GOMAXPROCS(runtime.NumCPU())

	//Up some limits.
	if err := limits.SetLimits(); err != nil {
		os.Exit(1)
	}

	// Work around defer not working after os.Exit()
	if err := factomdMain(s); err != nil {
		os.Exit(1)
	}

}

func factomdMain(s *state.State) error {
	var inMsgQueue = make(chan wire.FtmInternalMsg, 100)     //incoming message queue for factom application messages
	var outMsgQueue = make(chan wire.FtmInternalMsg, 100)    //outgoing message queue for factom application messages
	var inCtlMsgQueue = make(chan wire.FtmInternalMsg, 100)  //incoming message queue for factom application messages
	var outCtlMsgQueue = make(chan wire.FtmInternalMsg, 100) //outgoing message queue for factom application messages

	// Start the processor module
	go process.Start_Processor(s.DB, inMsgQueue, outMsgQueue, inCtlMsgQueue, outCtlMsgQueue)

	// Start the wsapi server module in a separate go-routine
	wsapi.Start(s.DB, inMsgQueue)

	// wait till the initialization is complete in processor
	hash, _ := s.DB.FetchDBHashByHeight(0)
	if hash != nil {
		for true {
			latestDirBlockHash, _, _ := s.DB.FetchBlockHeightCache()
			if latestDirBlockHash == nil {
				ftmdLog.Info("Waiting for the processor to be initialized...")
				time.Sleep(2 * time.Second)
			} else {
				break
			}
		}
	}

	if len(os.Args) >= 2 {
		if os.Args[1] == "initializeonly" {
			time.Sleep(time.Second)
			fmt.Println("Initializing only.")
			os.Exit(0)
		}
	} else {
		fmt.Println("\n'factomd initializeonly' will do just that.  Initialize and stop.")
	}

	// Start the factoid (btcd) component and P2P component
	btcd.Start_btcd(s.DB, inMsgQueue, outMsgQueue, inCtlMsgQueue, outCtlMsgQueue, process.FactomdUser, process.FactomdPass, SERVER_NODE != cfg.App.NodeMode)

	return nil
}

func isCompilerVersionOK() bool {
	goodenough := false

	if strings.Contains(runtime.Version(), "1.4") {
		goodenough = true
	}

	if strings.Contains(runtime.Version(), "1.5") {
		goodenough = true
	}

	if strings.Contains(runtime.Version(), "1.6") {
		goodenough = true
	}

	return goodenough
}
