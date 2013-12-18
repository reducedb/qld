// Copyright (c) 2013 Jian Zhen <zhenjl@gmail.com>
//
// All rights reserved.
//
// Use of this source code is governed by the Apache 2.0 license.

package qld

const (
	defaultHeaderSize      = 4
	defaultProtocolVersion = 0x0a
	defaultMaxPacketSize   = 1<<24 - 1
	defaultTimeFormat      = "2006-01-02 15:04:05"
)

// http://dev.mysql.com/doc/internals/en/generic-response-packets.html
const (
	okPacket  = 0x00
	eofPacket = 0xfe
	errPacket = 0xff
)

// mysql_com.h
// enum_server_command
// You should add new commands to the end of the list, otherwise old servers
// won't be able to handle them as 'unsupported'
// http://dev.mysql.com/doc/internals/en/command-phase.html
type serverCommand byte

const (
	comSleep serverCommand = iota
	comQuit
	comInitDB
	comComQuery
	comFieldList
	comCreateDB
	comDropDB
	comRefresh
	comShutdown
	comStatistics
	comProcessInfo
	comConnect
	comProcessKill
	comDebug
	comPing
	comTime
	comDelayedInsert
	comChangeUser
	comBinlogDump
	comTableDump
	comConnectOut
	comRegisterSlave
	comStmtPrepare
	comStmtExecute
	comStmtSendLongData
	comStmtClose
	comStmtReset
	comSetOption
	comStmtFetch
	comDaemon
	comBinlogDumpGTID
	comResetConnection

	// Must be last
	comEnd
)

var commandName []string = []string{
	"Sleep",
	"Quit",
	"Init DB",
	"Query",
	"Field List",
	"Create DB",
	"Drop DB",
	"Refresh",
	"Shutdown",
	"Statistics",
	"Processlist",
	"Connect",
	"Kill",
	"Debug",
	"Ping",
	"Time",
	"Delayed insert",
	"Change user",
	"Binlog Dump",
	"Table Dump",
	"Connect Out",
	"Register Slave",
	"Prepare",
	"Execute",
	"Long Data",
	"Close stmt",
	"Reset stmt",
	"Set option",
	"Fetch",
	"Daemon",
	"Binlog Dump GTID",
	"Reset Connection",
	"Error",
}

type fieldFlag uint32

const (
	flagNotNull         fieldFlag = 1 << iota // Field can't be NULL
	flagPriKey                                // Field is part of a primary key
	flagUniqueKey                             // Field is part of a unique key
	flagMultipleKey                           // Field is part of a key
	flagBlob                                  // Field is a blob
	flagUnsigned                              // Field is unsigned
	flagZeroFill                              // Field is zerofill
	flagBinary                                // Field is binary
	flagEnum                                  // Field is an enum
	flagAutoIncrement                         // Field is a autoincrement field
	flagTimestamp                             // Field is a timestamp
	flagSet                                   // Field is a set
	flagNoDefaultValue                        // Field doesn't have default value
	flagOnUpdateNow                           // Field is set to NOW on update
	flagPartKey                               // Field is part of some key
	flagNum                                   // Field is num (for clients)
	flagIntern1                               // Field is internal to mysql
	flagIntern2                               // Field is internal to mysql
	flagGetFixedFields                        // Used to get fields in item tree
	flagFieldInPartFunc                       // Field part of partition func
)

type refreshFlag uint32

const (
	refreshGrant      refreshFlag = 1 << iota // Refresh grant tables
	refreshLog                                // Start on new log file
	refreshTables                             // Close all tables
	refreshHosts                              // Flush host cache
	refreshStatus                             // Flush status variables
	refreshThreads                            // Flush thread cache
	refreshSlave                              // Reset master info and restart slave thread
	refreshMaster                             // Remove all bin logs in the index and truncate the index
	refreshErrorLog                           // Rotate only the error log
	refreshEngineLog                          // Flush all storage engine logs
	refreshBinaryLog                          // Flush the binary log
	refreshRelayLog                           // Flush the relay log
	refreshGeneralLog                         // Flush the general log
	refreshSlowLog                            // Flush the slow query log
	refreshReadLock                           // Lock tables for read
	refreshFast                               // Internal

	// Reset (remove all queries) from query cache
	refreshQueryCache
	refreshQueryCacheFree
	refreshDesKeyFile
	refreshUserResources
	refreshForExport
)

// http://dev.mysql.com/doc/internals/en/capability-flags.html
type clientFlag uint32

const (
	clientLongPassword               clientFlag = 1 << iota // new more secure passwords
	clientFoundRows                                         // Found instead of affected rows
	clientLongFlag                                          // Get all column flags
	clientConnectWithDB                                     // One can specify db on connect
	clientNoSchema                                          // Don't allow database.table.column
	clientCompress                                          // Can't use compression protocol
	clientODBC                                              // ODBC client
	clientLocalFiles                                        // Can use LOAD DATA LOCAL
	clientIgnoreSpace                                       // Ignore spaces before '('
	clientProtocol41                                        // New 4.1 protocol
	clientInteractive                                       // This is an interactive client
	clientSSL                                               // Switch to SSL after handshake
	clientIgnoreSIGPIPE                                     // Ignore sigpipes
	clientTransactions                                      // Client knows about transactions
	clientReserved                                          // Old flag for 4.1 protocol
	clientSecureConnection                                  // New 4.1 authentication
	clientMultiStatements                                   // Enable/disable multi-stmt support
	clientMultiResults                                      // Enable/disable multi-results
	clientPSMultiResults                                    // Multi-results in PS-protocol
	clientPluginAuth                                        // Client supports plugin authentication
	clientConnectAttrs                                      // Client supports connection attributes
	clientPluginAuthLenencClientData                        // Enable authentication response packet to be larger than 255 bytes
	clientCanHandleExpiredPasswords                         // Don't close the connection for a connection with expired password
	clientSSLVerifyServerCert        clientFlag = 1 << 30
	clientRememberOptions            clientFlag = 1 << 31
)

/* Gather all possible capabilites (flags) supported by the server */
const serverCapabilityFlags = clientLongPassword |
	clientFoundRows |
	clientConnectWithDB |
	//	clientCompress |
	clientLocalFiles |
	clientProtocol41 |
	clientInteractive |
	//	clientSSL |
	// clientTransactions |
	// clientReserved |
	clientSecureConnection

// http://dev.mysql.com/doc/internals/en/status-flags.html
type serverStatusFlag uint32

const (
	serverStatusInTrans    serverStatusFlag = 1 << iota
	serverStatusAutocommit                  // Server in auto_commit mode
	serverStatusUnused1
	serverMoreResultsExists // Multi-query - next query exists
	serverQueryNoGoodIndexUsed
	serverQueryNoIndexUsed

	// The server was able to fulfill the clients request and opened a read-only
	// non-scrollable cursor fora  query. This flag comes in reply to comStmtExecute
	// and comStmtFetch commands
	serverStatusCursorExists

	// This flag is sent when a read-only cursor is exhausted, in reply to the
	// comStmtFetch command
	serverStatusLastRowSent
	serverStatusDBDropped // A database was dropped
	serverStatusNoBackslashEscapes

	// Sent to the client if after a prepared statement reprepare we discovered
	// that the new statement returns a different numer of result set columns
	serverStatusMetadataChanged
	serverQueryWasSlow
	serverPSOutParams // To mark ResultSet containing output parameter values

	// Set at the same time as serverStatusInTrans if the started multi-statement
	// transaction is a read-only transaction. Cleared when the transaction commits
	// or aborts. Since this flag is sent to clients in OK and EOF packets, the flag
	// indicates the transaction status at the end of command execution.
	serverStatusInTransReadOnly
)

// Server status flags that must be cleared when starting execution of a new SQL
// statement. Flags from this set are only added to the current server status by the
// execution engine, but never removed -- the execution engine expects them to disappear
// automagically by the next command.
const serverStatusClearSet = serverQueryNoGoodIndexUsed |
	serverQueryNoIndexUsed |
	serverMoreResultsExists |
	serverStatusMetadataChanged |
	serverQueryWasSlow |
	serverStatusDBDropped |
	serverStatusCursorExists |
	serverStatusLastRowSent

type fieldType byte

const (
	fieldTypeDecimal fieldType = iota
	fieldTypeTiny
	fieldTypeShort
	fieldTypeLong
	fieldTypeFloat
	fieldTypeDouble
	fieldTypeNull
	fieldTypeTimestamp
	fieldTypeLongLong
	fieldTypeInt24
	fieldTypeDate
	fieldTypeTime
	fieldTypeDateTime
	fieldTypeYear
	fieldTypeNewDate
	fieldTypeVarchar
	fieldTypeBit
	fieldTypeTimestamp2
	fieldTypeDateTime2
	fieldTypeTime2
	fieldTypeNewDecimal fieldType = 246
	fieldTypeEnum       fieldType = 247
	fieldTypeSet        fieldType = 248
	fieldTypeTinyBlob   fieldType = 249
	fieldTypeMediumBlob fieldType = 250
	fieldTypeLongBlob   fieldType = 251
	fieldTypeBlob       fieldType = 252
	fieldTypeVarString  fieldType = 253
	fieldTypeString     fieldType = 254
	fieldTypeGeometry   fieldType = 255
)

// http://dev.mysql.com/doc/internals/en/character-set.html#packet-Protocol::CharacterSet
const (
	collationLatin1Swedish byte = 0x08 // 8
	collationUtf8General   byte = 0x21 // 33
	collationBinary        byte = 0x3f // 63
)
