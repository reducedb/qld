// Copyright (c) 2013 Jian Zhen <zhenjl@gmail.com>
//
// All rights reserved.
//
// Use of this source code is governed by the Apache 2.0 license.

package qld

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/golang/glog"
	"io"
	"net"
)

type connection struct {
	// Embedded struct so effectively Connection inherits all net.Conn functions
	net.Conn

	id int

	// Random bytes generator
	rand io.Reader

	// Pointer to the database configurations
	cfg *config

	// Buffer holding the incoming or outgoing packet. This can technical get very big
	// since it's never released. If a result set is, say 1 GB, then this buffer will
	// be 1 GB forever. Basically it's the size of the biggest MySQL packet.
	buf bytes.Buffer

	// Contains the initial challenge sent to the user
	cipher [20]byte

	// Packet sequence ID
	// http://dev.mysql.com/doc/internals/en/sequence-id.html
	sequence byte

	clientCapabilities clientFlag
	username           string
	maxPktSize         uint32
	charset            byte
	schema             string
	authResp           string

	quitChan chan bool
}

func (this *connection) handleConnectionPhase() error {
	if err := this.handlePlainHandshake(); err != nil {
		return err
	}

	glog.V(3).Info("Handshake successful")
	return nil
}

func (this *connection) handleCommandPhase() error {
	for {
		cmd, err := this.nextCommand()
		if err != nil {
			glog.Error(err.Error())
			return err
		}

		if cmd.cmd == comQuit {
			glog.V(3).Info("Quiting")
			return nil
		}

		if err := cmd.execute(); err != nil {
			glog.Error(err.Error())
			if err2 := this.writeErrPacket(err); err2 != nil {
				glog.Error(err2.Error())
			}
		} else {
			if err2 := this.writeOkPacket(); err2 != nil {
				glog.Error(err2.Error())
			}
		}
	}

	return nil
}

func (this *connection) nextCommand() (*command, error) {
	this.sequence = 0

	if err := this.readPacket(); err != nil {
		return nil, err
	}

	return newCommand(this.buf)
}

func (this *connection) handlePlainHandshake() error {
	if err := this.initHandshakeV10(); err != nil {
		return err
	}

	if err := this.writePacket(); err != nil {
		return err
	}

	if err := this.readPacket(); err != nil {
		return err
	}

	if err := this.parseHandshakeResponse41(); err != nil {
		return err
	}

	if err := this.writeOkPacket(); err != nil {
		return err
	}

	return nil
}

func (this *connection) writePacket() error {
	var header [defaultHeaderSize]byte

	// Split packets as needed
	for {
		data := this.buf.Next(defaultMaxPacketSize)
		pktLen := len(data)

		header[0] = byte(pktLen)
		header[1] = byte(pktLen >> 8)
		header[2] = byte(pktLen >> 16)
		header[3] = this.sequence

		// Write header
		if n, err := this.Write(header[:]); err != nil {
			return err
		} else if n != len(header) {
			return fmt.Errorf("Connection/writePacket: Error writing header. Expect %d bytes. Got %d.", len(header), n)
		}

		// Write data body
		if n, err := this.Write(data); err != nil {
			return err
		} else if n != len(data) {
			return fmt.Errorf("Connection/writePacket: Error writing body. Expect %d bytes. Got %d.", len(data), n)
		}

		glog.V(3).Infof("Wrote %d bytes", len(data)+len(header))

		this.sequence++

		if pktLen < defaultMaxPacketSize {
			return nil
		}
	}

	// Should never reach here
	return nil
}

func (this *connection) readPacket() (err error) {
	defer func() {
		// Looks like Grow() will panic, so let's recover from it
		if r := recover(); r != nil {
			var ok bool
			err, ok = r.(error)
			if !ok {
				err = fmt.Errorf("Connection/readPacket: %v", r)
			}
		}
	}()

	// Reset the buffer so we are clear for read/write
	this.buf.Reset()
	var header [defaultHeaderSize]byte

	// http://dev.mysql.com/doc/internals/en/sending-more-than-16mbyte.html
	// We are in a for loop to continue reading until payload is < defaultMaxPacketSize
	for {
		// 4 bytes - packet header
		if n, err := this.Read(header[:]); err != nil {
			return err
		} else if n != defaultHeaderSize {
			return fmt.Errorf("Connection/readPacket: Error reading header. Expect %d bytes. Got %d.", defaultHeaderSize, n)
		} else {
			glog.V(3).Infof("read data, n = %d", n)
		}

		if header[3] != this.sequence {
			return fmt.Errorf("Connecton/readPacket: sequence number mismatch")
		}

		pktLen := int(uint32(header[0]) | uint32(header[1])<<8 | uint32(header[2])<<16)
		glog.V(3).Infof("Packet length without header = %d bytes", pktLen)

		/*
			this.buf.Grow(pktLen)

			data := this.buf.Bytes()[:pktLen]

			if n, err := this.Read(data); err != nil {
				return err
			} else if n != pktLen {
				return fmt.Errorf("Connection/readPacket: Error reading body. Expect %d bytes. Got %d.", pktLen, n)
			}

			this.buf.SetBuffer(data)
		*/

		if n, err := io.CopyN(&this.buf, this, int64(pktLen)); err != nil {
			return err
		} else if n != int64(pktLen) {
			return fmt.Errorf("Connection/readPacket: Error reading body. Expect %d bytes. Got %d.", pktLen, n)
		}

		glog.V(3).Infof("Read %d bytes", this.buf.Len()+4)

		this.sequence++

		if pktLen < defaultMaxPacketSize {
			return nil
		}
	}

	// Should really never get here
	return nil
}

//http://dev.mysql.com/doc/internals/en/generic-response-packets.html#packet-ERR_Packet
func (this *connection) writeErrPacket(e error) error {
	sqlerr, ok := e.(*SQLError)
	if !ok {
		return fmt.Errorf("Connection/writeErrPacket: Cannot type assert sql error")
	}

	// Reset the buffer so we are clear for read/write
	this.buf.Reset()

	/*
			1              [ff] the ERR header
			2              error code
		  	if capabilities & CLIENT_PROTOCOL_41 {
				string[1]      '#' the sql-state marker
				string[5]      sql-state
			  }
			string[EOF]    error-message
	*/

	if err := this.buf.WriteByte(errPacket); err != nil {
		return err
	}

	if n, err := this.buf.Write([]byte{byte(sqlerr.Code), byte(sqlerr.Code >> 8)}); err != nil {
		return err
	} else if n != 2 {
		return fmt.Errorf("Connection/writeErrPacket: Error writing error code. Expecting %d, got %d", 2, n)
	}

	if err := this.buf.WriteByte('#'); err != nil {
		return err
	}

	if n, err := this.buf.Write([]byte(sqlerr.State)[:5]); err != nil {
		return err
	} else if n != 5 {
		return fmt.Errorf("Connection/writeErrPacket: Error writing sql state. Expecting %d, got %d", 5, n)
	}

	if n, err := this.buf.Write([]byte(sqlerr.Message)); err != nil {
		return err
	} else if n != len(sqlerr.Message) {
		return fmt.Errorf("Connection/writeErrPacket: Error writing sql state. Expecting %d, got %d", len(sqlerr.Message), n)
	}

	return this.writePacket()
}

// http://dev.mysql.com/doc/internals/en/generic-response-packets.html#packet-OK_Packet
func (this *connection) writeOkPacket() error {
	// Reset the buffer so we are clear for read/write
	this.buf.Reset()

	/*
			1              [00] the OK header
			lenenc-int     affected rows
			lenenc-int     last-insert-id
		  	if capabilities & CLIENT_PROTOCOL_41 {
			2              status_flags
			2              warnings
		  	} elseif capabilities & CLIENT_TRANSACTIONS {
			2              status_flags
		  	}
			string[EOF]    info
	*/

	if err := this.buf.WriteByte(okPacket); err != nil {
		return err
	}

	if n, err := this.buf.Write([]byte{0, 0, 0, 0, 0, 0}); err != nil {
		return err
	} else if n != 6 {
		return fmt.Errorf("Connection/writeOkPacket: Error writing okPacket. Expecting %d, got %d", 6, n)
	}

	//glog.V(3).Infof("ok packet = %#v", this.buf.Bytes())
	return this.writePacket()
}

// http://dev.mysql.com/doc/internals/en/connection-phase-packets.html
func (this *connection) initHandshakeV10() error {
	// Reset the buffer so we are clear for read/write
	this.buf.Reset()

	// 1 byte - [0a] protocol version
	if err := this.buf.WriteByte(defaultProtocolVersion); err != nil {
		return err
	}

	// 6 bytes - string[NUL]    server version
	if n, err := this.buf.Write([]byte{'0', '.', '1', '.', '0', 0}); err != nil {
		return err
	} else if n != 6 {
		return fmt.Errorf("Connection/writeInitialHandshakePacket: Error writing server version. Expecting %d, got %d", 6, n)
	}

	// 4 bytes - connection id
	if n, err := this.buf.Write([]byte{byte(this.id), byte(this.id >> 8), byte(this.id >> 16), byte(this.id >> 24)}); err != nil {
		return err
	} else if n != 4 {
		return fmt.Errorf("Connection/writeInitialHandshakePacket: Error writing connection id. Expecting %d, got %d", 4, n)
	}

	// 8 bytes - string[8] auth-plugin-data-part-1
	// Write 8 bytes now, 12 bytes later
	if n, err := this.buf.Write(this.cipher[:8]); err != nil {
		return err
	} else if n != 8 {
		return fmt.Errorf("Connection/writeInitialHandshakePacket: Error writing cipher. Expecting %d, got %d", 8, n)
	}

	// 1 byte - [00] filler
	if err := this.buf.WriteByte(0x00); err != nil {
		return err
	}

	// 2 bytes - capability flags (lower 2 bytes)
	if err := binary.Write(&this.buf, binary.LittleEndian, uint16(serverCapabilityFlags)); err != nil {
		return err
	}

	// if more data in the packet:
	// 1 byte - character set (optional)
	// 2 bytes - status flags (optional)
	// 2 bytes - capability flags (upper 2 bytes)
	// if capabilities & CLIENT_PLUGIN_AUTH {
	// 	1 byte - length of auth-plugin-data
	// } else {
	// 	1 byte - [00]
	// }
	// 10 bytes - string[10] reserved (all [00])
	// So 16 bytes of 0x00
	if n, err := this.buf.Write([]byte{collationUtf8General, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}); err != nil {
		return err
	} else if n != 16 {
		return fmt.Errorf("Connection/writeInitialHandshakePacket: Error writing charset, status and capability flags. Expecting %d, got %d", 16, n)
	}

	// Write the last 12 bytes of cipher
	// string[$len]   auth-plugin-data-part-2 ($len=MAX(13, length of auth-plugin-data - 8))
	if n, err := this.buf.Write(this.cipher[8:20]); err != nil {
		return err
	} else if n != 12 {
		return fmt.Errorf("Connection/writeInitialHandshakePacket: Error writing cipher. Expecting %d, got %d", 12, n)
	}

	// 1 byte - [00] byte 13 of auth-plugin-data-part-2
	if err := this.buf.WriteByte(0x00); err != nil {
		return err
	}

	/*
		  if capabilities & CLIENT_SECURE_CONNECTION {
		string[$len]   auth-plugin-data-part-2 ($len=MAX(13, length of auth-plugin-data - 8))
		  if capabilities & CLIENT_PLUGIN_AUTH {
		    if version >= (5.5.7 and < 5.5.10) or (>= 5.6.0 and < 5.6.2) {
		string[EOF]    auth-plugin name
		    } elseif version >= 5.5.10 or >= 5.6.2 {
		string[NUL]    auth-plugin name
		    }
		  }
	*/

	return nil
}

func (this *connection) parseHandshakeResponse41() error {
	// 4 bytes - capability flags, CLIENT_PROTOCOL_41 always set
	this.clientCapabilities = clientFlag(binary.LittleEndian.Uint32(this.buf.Next(4)))
	if this.clientCapabilities&clientProtocol41 == 0 {
		glog.V(3).Infof("%#v", errNotProtocol41)
		return errNotProtocol41
	}
	glog.V(3).Infof("Client capabilities = 0x%x, %032b", this.clientCapabilities, this.clientCapabilities)

	// 4 bytes - max-packet size
	this.maxPktSize = binary.LittleEndian.Uint32(this.buf.Next(4))
	if this.maxPktSize == 0 {
		this.maxPktSize = defaultMaxPacketSize
		glog.V(3).Infof("Setting packet size to default %d", this.maxPktSize)
	}
	glog.V(3).Infof("Received maxPktSize = %d", this.maxPktSize)

	// 1 byte - character set
	this.charset = this.buf.Next(1)[0]
	if this.charset != collationUtf8General {
		glog.Errorf("Charset = %d != %d", this.charset, collationUtf8General)
		// TODO: problem, error
	}
	glog.V(3).Infof("Character set = %b", this.charset)

	// 23 bytes - string[23]     reserved (all [0])
	// skipping
	if len(this.buf.Next(23)) != 23 {
		// TODO: problem
	}

	// string[NUL]    username
	tmp, err := this.buf.ReadBytes(0x00)
	if err != nil {
		return err
	}
	this.username = string(tmp)
	glog.V(3).Infof("User name = %s", this.username)

	// if capabilities & CLIENT_PLUGIN_AUTH_LENENC_CLIENT_DATA {
	//	lenenc-int     length of auth-response
	//	string[n]      auth-response
	// } else if capabilities & CLIENT_SECURE_CONNECTION {
	//  1              length of auth-response
	//  string[n]      auth-response
	// } else {
	//	string[NUL]    auth-response
	// }

	var authResp []byte
	if serverCapabilityFlags&clientPluginAuthLenencClientData != 0 {
		glog.Error("Should not reach here since we cannot handle length encoded integers")
		authResp = []byte{}
	} else if serverCapabilityFlags&clientSecureConnection != 0 {
		// We should definitely be here
		// 1 byte - length of auth-response
		n, err := this.buf.ReadByte()
		if err != nil {
			return err
		}
		glog.V(3).Infof("Auth Response length = %d", n)

		// n bytes      auth-response
		authResp = this.buf.Next(int(n))
		if len(authResp) != int(n) {
			return fmt.Errorf("Connection/readHandshakeResponse41: Insufficient data length. Expect %d, received %d", int(n), len(authResp))
		}
	} else {
		var err error
		authResp, err = this.buf.ReadBytes(0x00)
		if err != io.EOF && err != nil {
			return err
		}
	}
	this.authResp = string(authResp)
	glog.V(3).Infof("Auth-response = %s", this.authResp)

	// if capabilities & CLIENT_CONNECT_WITH_DB {
	// 	string[NUL]    database
	// }

	if serverCapabilityFlags&clientConnectWithDB != 0 {
		if tmp, err := this.buf.ReadBytes(0x00); err != nil && err != io.EOF {
			return err
		} else {
			this.schema = string(tmp)
		}
	}
	glog.V(3).Infof("Schema = %s", this.schema)

	// This should be the end of the packet

	return nil
}
