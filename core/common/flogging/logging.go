package flogging

import (
	"fmt"
	"io"
	ll "log"
	"os"
	"regexp"
	"strings"
	"sync"

	"github.com/op/go-logging"
	//"runtime"
	//"fmt"
)

const (
	pkgLogID      = "flogging"
	defaultFormat = "%{color}%{time:2006-01-02 15:04:05.000 MST} [%{module}] %{shortfunc} -> %{level:.4s} %{id:03x}%{color:reset} %{message}"
	defaultLevel  = logging.INFO
)

var (
	logger *logging.Logger

	defaultOutput io.Writer

	modules          map[string]string
	peerStartModules map[string]string

	lock sync.RWMutex
	once sync.Once
)

//	=====================================================================
//	function name: init
//	function type: private
//	function receiver: na
//	initialize the grpclogger config
//	=====================================================================
func init() {
	logger = logging.MustGetLogger(pkgLogID)
	Reset()
	//initgrpclogger()
}

//	=====================================================================
//	function name: Reset
//	function type: public
//	function receiver: na
//	Reset sets to logging to the defaults defined in this package
//	=====================================================================
func Reset() {
	modules = make(map[string]string)
	lock = sync.RWMutex{}
	defaultOutput = os.Stderr
	InitBackend(SetFormat(defaultFormat), defaultOutput)
	InitFromSpec("")
}

func SetOuput(output io.Writer) {
	/*	fmt.Println("logging setouput ",runtime.GOOS)
		if runtime.GOOS != "windows" {
			fmt.Println("logging if")
			InitBackend(SetFormat(defaultFormat), output)
			InitFromSpec("")
		}*/
}

//	=====================================================================
//	function name: SetFormat
//	function type: public
//	function receiver: na
//	SetFormat sets the logging format
//	=====================================================================
func SetFormat(formatSpec string) logging.Formatter {
	if formatSpec == "" {
		formatSpec = defaultFormat
	}

	return logging.MustStringFormatter(formatSpec)
}

//	=====================================================================
//	function name: InitBackend
//	function type: public
//	function receiver: na
//	InitBackend sets up the logging backend based on the provided
//	logging formatter and I/O writer.
//	=====================================================================
func InitBackend(formatter logging.Formatter, output io.Writer) {
	backend := logging.NewLogBackend(output, "", 0)
	backendFormatter := logging.NewBackendFormatter(backend, formatter)
	logging.SetBackend(backendFormatter).SetLevel(defaultLevel, "")
}

//	=====================================================================
//	function name: DefaultLevel
//	function type: public
//	function receiver: na
//	DefaultLevel returns the fallback value for loggers to use if parsing
//	fails
//	=====================================================================
func DefaultLevel() string {
	return defaultLevel.String()
}

//	=====================================================================
//	function name: GetModuleLevel
//	function type: public
//	function receiver: na
//	GetModuleLevel gets the current logging level for the specified module
//	=====================================================================
func GetModuleLevel(module string) string {
	//	logging.GetLevel() returns the logging level for the module, if defined.
	//	Otherwise, it returns the default logging level, as set by
	//	`flogging/logging.go`.
	level := logging.GetLevel(module).String()

	return level
}

//	=====================================================================
//	function name: SetModuleLevel
//	function type: public
//	function receiver: na
//	SetModuleLevel sets the logging level for the modules that match the supplied
//	regular expression. Can be used to dynamically change the log level for the
//	module.
//	=====================================================================
func SetModuleLevel(moduleRegExp string, level string) (string, error) {
	return setModuleLevel(moduleRegExp, level, false, false)
}

//	=====================================================================
//	function name: setModuleLevel
//	function type: public
//	function receiver: na
//	implement set the logging level for the modules
//	=====================================================================
func setModuleLevel(moduleRegExp string, level string, isRegExp bool, revert bool) (string, error) {
	var re *regexp.Regexp
	logLevel, err := logging.LogLevel(level)
	if err != nil {
		logger.Warningf("Invalid logging level '%s' - ignored", level)
	} else {
		if !isRegExp || revert {
			logging.SetLevel(logLevel, moduleRegExp)
			logger.Debugf("Module '%s' logger enabled for log level '%s'", moduleRegExp, level)
		} else {
			re, err = regexp.Compile(moduleRegExp)
			if err != nil {
				logger.Warningf("Invalid regular expression: %s", moduleRegExp)
				return "", err
			}
			lock.Lock()
			defer lock.Unlock()
			for module := range modules {
				if re.MatchString(module) {
					logging.SetLevel(logging.Level(logLevel), module)
					modules[module] = logLevel.String()
					logger.Debugf("Module '%s' logger enabled for log level '%s'", module, logLevel)
				}
			}
		}
	}

	return logLevel.String(), err
}

//	=====================================================================
//	function name: MustGetLogger
//	function type: public
//	function receiver: na
//	MustGetLogger is used in place of `logging.MustGetLogger` to allow us to
//	store a map of all modules and submodules that have loggers in the system.
//	=====================================================================
func MustGetLogger(module string) *logging.Logger {
	l := logging.MustGetLogger(module)
	lock.Lock()
	defer lock.Unlock()
	modules[module] = GetModuleLevel(module)
	return l
}

//	=====================================================================
//	function name: InitFromSpec
//	function type: public
//	function receiver: na
//	InitFromSpec initializes the logging based on the supplied spec. It is
//	exposed externally so that consumers of the flogging package may parse their
//	own logging specification. The logging specification has the following form:
//		[<module>[,<module>...]=]<level>[:[<module>[,<module>...]=]<level>...]
//	=====================================================================
func InitFromSpec(spec string) string {
	levelAll := defaultLevel
	var err error

	if spec != "" {
		fields := strings.Split(spec, ":")
		for _, field := range fields {
			split := strings.Split(field, "=")
			switch len(split) {
			case 1:
				if levelAll, err = logging.LogLevel(field); err != nil {
					logger.Warningf("Logging level '%s' not recognized, defaulting to '%s': %s", field, defaultLevel, err)
					levelAll = defaultLevel // need to reset cause original value was overwritten
				}
			case 2:
				// <module>[,<module>...]=<level>
				levelSingle, err := logging.LogLevel(split[1])
				if err != nil {
					logger.Warningf("Invalid logging level in '%s' ignored", field)
					continue
				}

				if split[0] == "" {
					logger.Warningf("Invalid logging override specification '%s' ignored - no module specified", field)
				} else {
					modules := strings.Split(split[0], ",")
					for _, module := range modules {
						logger.Debugf("Setting logging level for module '%s' to '%s'", module, levelSingle)
						logging.SetLevel(levelSingle, module)
					}
				}
			default:
				logger.Warningf("Invalid logging override '%s' ignored - missing ':'?", field)
			}
		}
	}

	logging.SetLevel(levelAll, "") // set the logging level for all modules

	// iterate through modules to reload their level in the modules map based on
	// the new default level
	for k := range modules {
		MustGetLogger(k)
	}
	// register flogging logger in the modules map
	MustGetLogger(pkgLogID)

	return levelAll.String()
}

//	=====================================================================
//	function name: SetPeerStartupModulesMap
//	function type: public
//	function receiver: na
//	SetPeerStartupModulesMap saves the modules and their log levels.
//	this function should only be called at the end of peer startup.
//	=====================================================================
func SetPeerStartupModulesMap() {
	lock.Lock()
	defer lock.Unlock()

	once.Do(func() {
		peerStartModules = make(map[string]string)
		for k, v := range modules {
			peerStartModules[k] = v
		}
	})
}

//	=====================================================================
//	function name: GetPeerStartupLevel
//	function type: public
//	function receiver: na
//	GetPeerStartupLevel returns the peer startup level for the specified module.
//	It will return an empty string if the input parameter is empty or the module
//	is not found
//	=====================================================================
func GetPeerStartupLevel(module string) string {
	if module != "" {
		if level, ok := peerStartModules[module]; ok {
			return level
		}
	}

	return ""
}

//	=====================================================================
//	function name: GetPeerStartupLevel
//	function type: public
//	function receiver: na
//	RevertToPeerStartupLevels reverts the log levels for all modules to the level
//	defined at the end of peer startup.
//	=====================================================================
func RevertToPeerStartupLevels() error {
	lock.RLock()
	defer lock.RUnlock()
	for key := range peerStartModules {
		_, err := setModuleLevel(key, peerStartModules[key], false, true)
		if err != nil {
			return err
		}
	}
	logger.Info("Log levels reverted to the levels defined at the end of peer startup")

	return nil
}

// type Slog struct {
// 	hook *slog.SyslogHook
// }

// func NewSysLog(module string) (*Slog, error) {
// 	hook, err := slog.NewSyslogHook("udp", cmd.Config.Log.LogServer, syslog.LOG_INFO, module)
// 	if err != nil {
// 		logger.Fatal(err)
// 		return nil, err
// 	}

// 	return &Slog{
// 		hook: hook,
// 	}, nil
// }

// func (s *Slog) Info(m string) {
// 	err := s.hook.Writer.Info(m)
// 	if err != nil {
// 		logger.Error(err)
// 	}
// }

// func (s *Slog) Debug(m string) {
// 	err := s.hook.Writer.Debug(m)
// 	if err != nil {
// 		logger.Error(err)
// 	}
// }

// func (s *Slog) Notice(m string) {
// 	err := s.hook.Writer.Notice(m)
// 	if err != nil {
// 		logger.Error(err)
// 	}
// }

// func (s *Slog) Warning(m string) {
// 	err := s.hook.Writer.Notice(m)
// 	if err != nil {
// 		logger.Error(err)
// 	}
// }

// func (s *Slog) Error(m string) {
// 	err := s.hook.Writer.Err(m)
// 	if err != nil {
// 		logger.Error(err)
// 	}
// }

// func (s *Slog) Crit(m string) {
// 	err := s.hook.Writer.Crit(m)
// 	if err != nil {
// 		logger.Error(err)
// 	}
// }

func LogTxt(fileName string) {

	var file *os.File
	_, err := os.Stat(fileName)
	if os.IsExist(err) {
		file, err = os.OpenFile(fileName, os.O_WRONLY, 0666)
		if err != nil {
			fmt.Println(err)
		}
	} else {

		file, err = os.Create(fileName)
		if err != nil {
			fmt.Println(err)
		}
	}
	backend1 := logging.NewLogBackend(file, "", ll.LstdFlags|ll.Lshortfile)
	backend2 := logging.NewLogBackend(os.Stderr, "", 0)

	backend2Formatter := logging.NewBackendFormatter(backend2, logging.MustStringFormatter(defaultFormat))
	backend1Leveled := logging.AddModuleLevel(backend1)
	backend1Leveled.SetLevel(logging.INFO, "")
	logging.SetBackend(backend1Leveled, backend2Formatter)
}
