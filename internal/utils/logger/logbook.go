package logger

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"sync"
	"time"

	"github.com/wizarki972/myone/internal/config"
	"github.com/wizarki972/myone/internal/utils/fldir"
)

// LogBook - a book that stores all the logs and saves them when its told, usually at the end of any command execution.
type LogBook struct {
	mu              sync.Mutex
	userConfig      *config.Config
	invokedByFlags  string
	invokedBySubCmd string
	bookStartTime   time.Time

	logs     string
	logCount int

	save            bool
	saveOnError     bool
	savePath        string
	previouslySaved bool
	closeOnError    bool
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
		closeOnError:    true,
	}
}

// Enters given log into the log book, following log types are accepted: 1 - info, 2 - warning, 3 - error
func (book *LogBook) EnterLog(logMsg string, logType LogType, err error) {
	book.mu.Lock()
	defer book.mu.Unlock()

	if len(logMsg) == 0 {
		book.Print("Cannot enter an empty log.", LogTypes.Error, nil)
	}

	book.mu.Lock()
	book.logs += fmt.Sprintf("%s -- [%s] %s\n", time.Now().Format("02-01-2006 15:04:05"), logType.Type, logMsg)
	book.logCount += 1
	book.mu.Unlock()

	if logType == LogTypes.Error && (book.saveOnError || book.save) {
		book.SaveBook()
		book.Print(logMsg, logType, err)
	}
}

// It stores the log in the book and it also prints it.
func (book *LogBook) EnterLogAndPrint(logMsg string, logType LogType, err error) {
	book.mu.Lock()
	defer book.mu.Unlock()

	if len(logMsg) == 0 {
		book.Print("Cannot enter an empty log.", LogTypes.Error, nil)
	}

	if book.saveOnError || book.save {
		book.mu.Lock()
		book.logs += fmt.Sprintf("%s -- [%s] %s\n", time.Now().Format("02-01-2006 15:04:05"), logType.Type, logMsg)
		book.logCount += 1
		book.mu.Unlock()
	}

	if logType == LogTypes.Error && (book.saveOnError || book.save) {
		book.SaveBook()
	}

	book.Print(logMsg, logType, err)
}

// Add which sub command is running
func (book *LogBook) AddSubCommand(subCmd string) {
	if len(subCmd) == 0 {
		book.Print("Cannot add an empty sub command.", LogTypes.Error, nil)
	}
	book.invokedBySubCmd = subCmd
}

// Adds the next flag, which is running...
func (book *LogBook) AddFlag(flag string) {
	if len(flag) == 0 {
		book.Print("Cannot add an empty flag.", LogTypes.Error, nil)
	}
	book.invokedByFlags += flag + ","
	book.EnterLog("FROM HERE, LOGS FOR THE FOLLOWING FLAG - "+flag, LogTypes.Info, nil)
}

// Saves the log book in the specified location
func (book *LogBook) SaveBook() error {
	var err error
	book.mu.Lock()
	defer book.mu.Unlock()

	logHeader := fmt.Sprintf("title=MyOne Log\ninvokedBySubCommand=%s\ninvokedByFlags=%s\nlogStartedAt=%s\nlogCount=%d\n\n===LOGS===\n\n", book.invokedBySubCmd, book.invokedByFlags, book.bookStartTime.Format("02-01-2006 15:04:05"), book.logCount)
	if book.previouslySaved {
		err = fldir.WriteOrAppendToFile(logHeader+book.logs, book.savePath)
	} else {
		err = fldir.WriteStringToFile(logHeader+book.logs, book.savePath)
	}

	if err != nil {
		return err
	}

	if !book.previouslySaved {
		book.previouslySaved = true
	}
	return nil
}

// saves the logs after a certain amount of time.
// mainly used for background services running for a long time.
func (book *LogBook) StartAutoLogSaver(ctx context.Context) {
	var ticker *time.Ticker
	if book.userConfig.Logs.LogSaveInterval <= 0 || book.userConfig.Logs.LogSaveInterval > 59 {
		ticker = time.NewTicker(10 * time.Minute)
	} else {
		ticker = time.NewTicker(time.Duration(book.userConfig.Logs.LogSaveInterval) * time.Minute)
	}
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := book.SaveBook(); err != nil {
				book.EnterLogAndPrint("Failed to save logs, after 10 minute interval. Exact issue is printed below,", LogTypes.Warning, nil)
				fmt.Println("[ERROR] " + err.Error())
			}
		case <-ctx.Done():
			if err := book.SaveBook(); err != nil {
				book.EnterLogAndPrint("Failed to save logs, after 10 minute interval. Exact issue is printed below,", LogTypes.Warning, nil)
				fmt.Println("[ERROR] " + err.Error())
			}
			return
		}
	}
}

// tells the book to close/not close when an error is encoutered.
func (book *LogBook) SetCloseOnError(value bool) {
	book.closeOnError = value
}

// it prints the log
func (book *LogBook) Print(log string, logType LogType, err error) {
	if len(log) == 0 {
		fmt.Printf("-> [%s] IF YOU ARE SEEING THIS ERROR THAN THAT MEANS AN EMPTY LOG WAS PROVIDED.\n", logType.Type)
		return
	}

	if !slices.Contains(logLevels[book.userConfig.Logs.Level], logType) {
		return
	}
	fmt.Printf("-> %s\n", log)

	if logType == LogTypes.Error {
		if book.userConfig.Logs.Panic {
			if err == nil {
				panic(errors.New("no error was provided for this log"))
			}
			panic(err)
		}

		if book.closeOnError {
			os.Exit(1)
		}
	}
}
