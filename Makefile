# Copyright 2021 Eryx <evorui аt gmail dοt com>, All rights reserved.
#

EXE_SERVER = bin/indnsd
APP_HOME = /opt/sysinner/indns
APP_USER = root

all:
	go build -o ${EXE_SERVER} cmd/server/main.go

install:
	mkdir -p ${APP_HOME}/bin
	mkdir -p ${APP_HOME}/etc/conf.d
	mkdir -p ${APP_HOME}/var/log
	mkdir -p ${APP_HOME}/var/data
	cp -rp misc ${APP_HOME}/ 
	install -m 755 ${EXE_SERVER} ${APP_HOME}/${EXE_SERVER}
	install -m 600 misc/systemd/systemd.service /lib/systemd/system/indnsd.service
	systemctl daemon-reload

clean:
	rm -f ${EXE_SERVER}

