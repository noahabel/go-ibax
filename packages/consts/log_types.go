/*---------------------------------------------------------------------------------------------
 *  Copyright (c) IBAX. All rights reserved.
 *  See LICENSE in the project root for license information.
 *--------------------------------------------------------------------------------------------*/

package consts

// LogEventType is storing numeric event type
type LogEventType int

// Types of log errors
	MarshallingError         = "Marshall"
	UnmarshallingError       = "Unmarshall"
	ParseError               = "Parse"
	IOError                  = "IO"
	CryptoError              = "Crypto"
	ContractError            = "Contract"
	DBError                  = "DB"
	PanicRecoveredError      = "Panic"
	ConnectionError          = "Connection"
	ConfigError              = "Config"
	VMError                  = "VM"
	JustWaiting              = "JustWaiting"
	Ntpdate                  = "Ntpdate"
	BlockError               = "Block"
	ParserError              = "Parser"
	ContextError             = "Context"
	SessionError             = "Session"
	RouteError               = "Route"
	NotFound                 = "NotFound"
	Found                    = "Found"
	EmptyObject              = "EmptyObject"
	InvalidObject            = "InvalidObject"
	DuplicateObject          = "DuplicateObject"
	UnknownObject            = "UnknownObject"
	ParameterExceeded        = "ParameterExceeded"
	DivisionByZero           = "DivisionByZero"
	EvalError                = "Eval"
	JWTError                 = "JWT"
	AccessDenied             = "AccessDenied"
	SizeDoesNotMatch         = "SizeDoesNotMatch"
	NoIndex                  = "NoIndex"
	NoFunds                  = "NoFunds"
	BlockIsFirst             = "BlockIsFirst"
	IncorrectCallingContract = "IncorrectCallingContract"
	WritingFile              = "WritingFile"
	CentrifugoError          = "CentrifugoError"
	StatsdError              = "StatsdError"
	MigrationError           = "MigrationError"
	AutoupdateError          = "AutoupdateError"
	BCRelevanceError         = "BCRelevanceError"
	BCActualizationError     = "BCActualizationError"
	SchedulerError           = "SchedulerError"
	SyncProcess              = "SyncProcess"
	WrongModeError           = "WrongModeError"
	OBSManagerError          = "OBSManagerError"
	TCPClientError           = "TCPClientError"
	BadTxError               = "BadTxError"
	TimeCalcError            = "BlockTimeCounterError"
	RegisterError            = "RegisterError"
)
