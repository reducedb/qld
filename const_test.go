// Copyright (c) 2013 Jian Zhen <zhenjl@gmail.com>
//
// All rights reserved.
//
// Use of this source code is governed by the Apache 2.0 license.

package qld

import (
	"testing"
)

func TestServerCommandConstants(t *testing.T) {
	if comInitDB != 2 {
		t.Error("comInitDB != 2")
	}

	if commandName[comInitDB] != "Init DB" {
		t.Error("commandName[comInitDB] != 'Init DB'")
	}

	if comProcessInfo != 10 {
		t.Error("comProcessInfo != 10")
	}

	if commandName[comProcessInfo] != "Processlist" {
		t.Error("commandName[comProcessInfo] != 'Processlist'")
	}

	if comTime != 15 {
		t.Error("comTime != 15")
	}

	if commandName[comTime] != "Time" {
		t.Error("commandName[comTime] != 'Time'")
	}

	if comStmtExecute != 23 {
		t.Error("comStmtExecute != 23")
	}

	if commandName[comStmtExecute] != "Execute" {
		t.Error("commandName[comStmtExecute] != 'Execute'")
	}

	if comResetConnection != 31 {
		t.Error("comResetConnection != 31")
	}

	if commandName[comResetConnection] != "Reset Connection" {
		t.Error("commandName[comResetConnection] != 'Reset Connection'")
	}
}

func TestFieldFlagConstants(t *testing.T) {
	if flagNotNull != 1 {
		t.Error("flagNotNull != 1")
	}

	if flagBlob != 16 {
		t.Error("flagBlob != 16")
	}

	if flagTimestamp != 1024 {
		t.Error("flagTimestamp != 1024")
	}

	if flagOnUpdateNow != 8192 {
		t.Error("flagOnUpdateNow != 8192")
	}

	if flagFieldInPartFunc != 524288 {
		t.Error("flagFieldInPartFunc != 524288")
	}

}

func TestRefreshFlagConstants(t *testing.T) {
	if refreshGrant != 1 {
		t.Error("refreshGrant != 1")
	}

	if refreshSlave != 64 {
		t.Error("refreshSlave != 64")
	}

	if refreshEngineLog != 512 {
		t.Error("refreshEngineLog != 512")
	}

	if refreshSlowLog != 8192 {
		t.Error("refreshSlowLog != 8192")
	}

	if refreshForExport != 0x100000 {
		t.Error("refreshForExport != 0x100000")
	}
}

func TestClientFlagConstants(t *testing.T) {
	if clientLongPassword != 1 {
		t.Error("clientLongPassword != 1")
	}

	if clientProtocol41 != 512 {
		t.Error("clientProtocol41 != 512")
	}

	if clientMultiStatements != 65536 {
		t.Error("clientMultiStatements != 65536")
	}

	if clientCanHandleExpiredPasswords != 4194304 {
		t.Error("clientCanHandleExpiredPasswords != 4194304")
	}

	if clientRememberOptions != 1<<31 {
		t.Error("clientRememberOptions != 1<<31")
	}
}

func TestServerStatusFlagConstants(t *testing.T) {
	if serverStatusInTrans != 1 {
		t.Error("serverStatusInTrans != 1")
	}

	if serverStatusLastRowSent != 128 {
		t.Error("serverStatusLastRowSent != 128")
	}

	if serverStatusInTransReadOnly != 8192 {
		t.Error("serverStatusInTrans != 8192")
	}
}
