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

package config

import (
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/hooto/hlog4g/hlog"
	"github.com/hooto/htoml4g/htoml"
)

type ConfigCommon struct {
	Server struct {
		Bind            string   `toml:"bind"`
		NameServers     []string `toml:"name_servers"`
		ConfigDirectory string   `toml:"config_directory"`
	} `toml:"server"`
}

var (
	AppName = "indns"
	Prefix  = "/opt/sysinner/indns"
	Config  = ConfigCommon{}
	Version = ""
	Release = ""
	Records = ConfigRecordManager{
		items: map[string]*ConfigRecordEntry{},
	}
	confFile = ""
	err      error
)

func Setup(ver, rel string) error {

	Version, Release = ver, rel

	//
	if Prefix, err = filepath.Abs(filepath.Dir(os.Args[0]) + "/.."); err != nil {
		Prefix = "/opt/sysinner/" + AppName
	}
	confFile = Prefix + "/etc/indnsd.toml"

	if err := htoml.DecodeFromFile(confFile, &Config); err != nil && !os.IsNotExist(err) {
		return err
	}

	if Config.Server.Bind == "" {
		Config.Server.Bind = "127.0.0.1:53"
	}

	if Config.Server.ConfigDirectory == "" {
		Config.Server.ConfigDirectory = Prefix + "/etc/conf.d"
	}

	htoml.EncodeToFile(Config, confFile)

	{
		exec.Command("systemd", "stop", "systemd-resolved").Output()
		exec.Command("systemd", "disable", "systemd-resolved").Output()
	}

	return watcher()
}

type ConfigZone struct {
	Records []*ConfigZoneRecord `toml:"records"`
}

type ConfigZoneRecord struct {
	Name string   `toml:"name"`
	IPs  []net.IP `toml:"ips"`
}

type ConfigRecordManager struct {
	mu      sync.RWMutex
	version int64
	items   map[string]*ConfigRecordEntry
}

type ConfigRecordEntry struct {
	Version int64
	Name    string
	IPs     []net.IP
}

func (it *ConfigRecordManager) Get(name string) []net.IP {
	it.mu.RLock()
	defer it.mu.RUnlock()
	rc, ok := it.items[name]
	if ok {
		return rc.IPs
	}
	return nil
}

func (it *ConfigRecordManager) set(name string, ips []net.IP) error {

	var (
		ar     = map[string]bool{}
		ipSets = []net.IP{}
	)
	for _, v := range ips {
		if v != nil {
			if _, ok := ar[v.String()]; !ok {
				ipSets = append(ipSets, v)
			}
		}
	}

	it.mu.Lock()
	defer it.mu.Unlock()

	rc, ok := it.items[name]
	if !ok {
		rc = &ConfigRecordEntry{}
		it.items[name] = rc
	}

	if len(ipSets) == 0 && rc.Version == 0 {
		return nil
	}

	equal := func(ips1, ips2 []net.IP) bool {
		if len(ips1) != len(ips2) {
			return false
		}
		for _, v1 := range ips1 {
			hit := false
			for _, v2 := range ips2 {
				if v1.Equal(v2) {
					hit = true
					break
				}
			}
			if !hit {
				return false
			}
		}
		return true
	}

	if !equal(rc.IPs, ipSets) {
		rc.IPs = ipSets
		it.version++
		rc.Version = it.version
		hlog.Printf("info", "record set %s to ip %v", name, ipSets)
	}

	return nil
}

func (it *ConfigRecordManager) setZone(v ConfigZone) error {
	for _, v := range v.Records {
		it.set(v.Name, v.IPs)
	}
	return nil
}

func fileReload(path string) (*ConfigZone, error) {
	var cfg ConfigZone
	if err := htoml.DecodeFromFile(path, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func parseFile(path string) error {

	ext := filepath.Ext(path)
	if ext != ".toml" {
		return nil
	}

	cfg, err := fileReload(path)
	if err != nil {
		return err
	}
	return Records.setZone(*cfg)
}

func watcher() error {

	filepath.Walk(Config.Server.ConfigDirectory, func(path string, info os.FileInfo, err error) error {
		if info == nil || info.IsDir() {
			return nil
		}
		cfg, err := fileReload(path)
		if err != nil {
			hlog.Printf("warn", "load config file (%s) error %s", path, err.Error())
			return err
		}
		return Records.setZone(*cfg)
	})

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	// defer watcher.Close()

	go func() {
		for {
			select {
			case ev, ok := <-watcher.Events:
				if !ok {
					return
				}

				hlog.Printf("info", "config event %v", ev)

				if ev.Op&fsnotify.Create == fsnotify.Create {
					if err := parseFile(ev.Name); err != nil {
						hlog.Printf("warn", "conf event (%s) reload error %s", ev.Name, err.Error())
					}
				}

				if ev.Op&fsnotify.Write == fsnotify.Write {
					if err := parseFile(ev.Name); err != nil {
						hlog.Printf("warn", "conf event (%s) reload error %s", ev.Name, err.Error())
					}
				}

				if ev.Op&fsnotify.Remove == fsnotify.Remove {
				}

				if ev.Op&fsnotify.Rename == fsnotify.Rename {
				}

				if ev.Op&fsnotify.Chmod == fsnotify.Chmod {
				}

			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				hlog.Printf("error", err.Error())
			}
		}
	}()

	return watcher.Add(Config.Server.ConfigDirectory)
}
