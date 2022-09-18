// Copyright 2021 Eryx <evorui at gmail dot com>, All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"os"
	"os/signal"
	"runtime"

	"github.com/hooto/hlog4g/hlog"

	"github.com/sysinner/indns/config"
	"github.com/sysinner/indns/server"
)

func init() {
	runtime.GOMAXPROCS(1)
}

var (
	version = "git"
	release = "1"
)

func main() {

	if err := config.Setup(version, release); err != nil {
		hlog.Printf("error", "config setup fail : %s", err.Error())
		hlog.Flush()
		os.Exit(1)
	}

	if err := server.Start(); err != nil {
		hlog.Printf("error", "server start fail : %s", err.Error())
		hlog.Flush()
		os.Exit(1)
	}

	sig := make(chan os.Signal)
	signal.Notify(sig, os.Interrupt)

	sg := <-sig
	hlog.Printf("warn", "server signal quit %s", sg.String())
	hlog.Flush()
}
