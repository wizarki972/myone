package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/wizarki972/myone/internal/config"
	"github.com/wizarki972/myone/internal/utils/fldir"
)

// LogBook - a book that stores all the logs and saves them when its told, usually at the end of any command execution.
type LogBook struct {
	userConfig      *config.Config
	invokedByFlags  string
	invokedBySubCmd string
	bookStartTime   time.Time
	logs            string
	logCount        int
	save            bool
	saveOnError     bool
	savePath        string
}

// Constructor for LogBook. LogBook - a book that stores all the logs and saves them when its told, usually at the end of any command execution.
func NewLogBook(savePath string, save, saveOnError bool, userConfig *config.Config) *LogBook {
	bookTime := time.Now()
	if savePath == "" && (save || saveOnError) {
		if len(userConfig.Logs.DirectoryPath) != 0 && fldir.IsPathExist(userConfig.Logs.DirectoryPath) {
			savePath = filepath.Join(userConfig.Logs.DirectoryPath, bookTime.Format("02-01-2006 15:04:05")+"_myone.log")
		} else {
			savePath = filepath.Join(config.DefaultConfig.Logs.DirectoryPath, bookTime.Format("02-01-2006 15:04:05")+"_myone.log")
		}
	}

	return &LogBook{
		userConfig:      userConfig,
		bookStartTime:   bookTime,
		invokedByFlags:  "",
		invokedBySubCmd: "",
		logs:            "",
		logCount:        0,
		save:            save,
		saveOnError:     saveOnError,
		savePath:        savePath,
	}
}

// Enters given log into the log book, following log types are accepted: 1 - info, 2 - warning, 3 - error
func (book *LogBook) EnterLog(logMsg string, logType LogType, err error) {
	if len(logMsg) == 0 {
		book.Log("Cannot enter an empty log.", LogTypes.Error, nil)
	}

	book.logs += fmt.Sprintf("%s -- [%s] %s\n", time.Now().Format("02-01-2006 15:04:05"), logType.Type, logMsg)
	book.logCount += 1

	if logType == LogTypes.Error && (book.saveOnError || book.save) {
		book.SaveBook()
		book.Log(logMsg, logType, err)
	}
}

// It stores the log in the book and it also prints it.
func (book *LogBook) EnterLogAndPrint(logMsg string, logType LogType, err error) {
	if len(logMsg) == 0 {
		book.Log("Cannot enter an empty log.", LogTypes.Error, nil)
	}

	if book.saveOnError || book.save {
		book.logs += fmt.Sprintf("%s -- [%s] %s\n", time.Now().Format("02-01-2006 15:04:05"), logType.Type, logMsg)
		book.logCount += 1
	}

	if logType == LogTypes.Error && (book.saveOnError || book.save) {
		book.SaveBook()
	}

	book.Log(logMsg, logType, err)
}

// Add which sub command is running
func (book *LogBook) AddSubCommand(subCmd string) {
	if len(subCmd) == 0 {
		book.Log("Cannot add an empty sub command.", LogTypes.Error, nil)
	}
	book.invokedBySubCmd = subCmd
}

// Adds the next flag, which is running...
func (book *LogBook) AddFlag(flag string) {
	if len(flag) == 0 {
		book.Log("Cannot add an empty flag.", LogTypes.Error, nil)
	}
	book.invokedByFlags += flag + ","
	book.EnterLog("FROM HERE, LOGS FOR THE FOLLOWING FLAG - "+flag, LogTypes.Info, nil)
}

// Saves the log book in the specified location
func (book *LogBook) SaveBook() {
	logHeader := fmt.Sprintf("title=MyOne Log\ninvokedBySubCommand=%s\ninvokedByFlags=%s\nlogStartedAt=%s\nlogCount=%d\n\n===LOGS===\n\n", book.invokedBySubCmd, strings.TrimSpace(book.invokedByFlags), book.bookStartTime.Format("02-01-2006 15:04:05"), book.logCount)
	fldir.WriteStringToFile(logHeader+book.logs, book.savePath)
}

// it prints the log
func (book *LogBook) Log(log string, logType LogType, err error) {
	if len(log) == 0 {
		fmt.Printf("-> [%s] IF YOU ARE SEEING THIS ERROR THAN THAT MEANS AN EMPTY LOG WAS PROVIDED.\n", logType.Type)
		return
	}

	if !slices.Contains(logLevels[book.userConfig.Logs.Level], logType) {
		return
	}
	fmt.Printf("-> %s\n", log)

	if logType == LogTypes.Error {
		if book.userConfig.Logs.Panic && err != nil {
			panic(err)
		}
		os.Exit(1)
	}
}

// common log printing function - only place where this used is in cmd/common.go
func Log(log string, logType LogType, err error) {
	userConfig := config.GetConfig()
	if len(log) == 0 {
		fmt.Printf("-> [%s] IF YOU ARE SEEING THIS ERROR THAN THAT MEANS AN EMPTY LOG WAS PROVIDED.\n", logType.Type)
		return
	}

	if !slices.Contains(logLevels[userConfig.Logs.Level], logType) {
		return
	}
	fmt.Printf("-> %s\n", log)

	if logType == LogTypes.Error {
		if userConfig.Logs.Panic && err != nil {
			panic(err)
		}
		os.Exit(1)
	}
}
