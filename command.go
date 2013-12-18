// Copyright (c) 2013 Jian Zhen <zhenjl@gmail.com>
//
// All rights reserved.
//
// Use of this source code is governed by the Apache 2.0 license.

package qld

import (
	"bytes"
	"github.com/golang/glog"
	"io"
)

type command struct {
	cmd     serverCommand
	cmdName string
	stmt    string
}

func newCommand(data bytes.Buffer) (*command, error) {
	cmd := &command{}

	if err := cmd.parseCommand(data); err != nil {
		return nil, err
	}

	return cmd, nil
}

func (this *command) parseCommand(data bytes.Buffer) error {
	// 1 byte - character set
	this.cmd = serverCommand(data.Next(1)[0])
	this.cmdName = commandName[this.cmd]
	glog.V(3).Infof("Cmd = %s(%d)", this.cmdName, this.cmd)

	if tmp, err := data.ReadBytes(0x00); err != nil && err != io.EOF {
		return err
	} else {
		this.stmt = string(tmp)
	}
	glog.V(3).Infof("Cmd stmt = %s", this.stmt)

	return nil
}

func (this *command) execute() error {
	switch this.cmd {
	case comQuit:
		// should really never get here because handleCommandPhase() should have taken care of it
		return nil
		/*
			case comInitDB:
			case comComQuery:
			case comFieldList:
			case comCreateDB:
			case comDropDB:
			case comRefresh:
			case comShutdown:
			case comStatistics:
			case comProcessInfo:
			case comConnect:
			case comProcessKill:
			case comDebug:
			case comPing:
			case comTime:
			case comDelayedInsert:
			case comChangeUser:
			case comBinlogDump:
			case comTableDump:
			case comConnectOut:
			case comRegisterSlave:
			case comStmtPrepare:
			case comStmtExecute:
			case comStmtSendLongData:
			case comStmtClose:
			case comStmtReset:
			case comSetOption:
			case comStmtFetch:
			case comDaemon:
			case comBinlogDumpGTID:
			case comResetConnection:
		*/
	default:
		//case comSleep:
		return SQLErrors[1047]
	}
	return nil
}
