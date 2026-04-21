package logger

var logLevels = map[int][]LogType{
	1: {LogTypes.Error},
	2: {LogTypes.Error, LogTypes.Warning},
	3: {LogTypes.Error, LogTypes.Info, LogTypes.Warning},
}

type LogType struct {
	TypeCode int
	Type     string
}

var LogTypes = struct {
	Info    LogType
	Warning LogType
	Error   LogType
}{
	Info: LogType{
		TypeCode: 1,
		Type:     "INFO",
	},
	Warning: LogType{
		TypeCode: 2,
		Type:     "WARNING",
	},
	Error: LogType{
		TypeCode: 3,
		Type:     "ERROR",
	},
}

// logbook - to store logs and save them at the end,
//  config options - log save location
