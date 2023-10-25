package executor

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/colonyos/colonies/pkg/client"
	"github.com/colonyos/colonies/pkg/core"
	"github.com/colonyos/colonies/pkg/security/crypto"
	"github.com/colonyos/executors/common/pkg/debug"
	"github.com/colonyos/executors/common/pkg/failure"
	"github.com/colonyos/executors/common/pkg/parsers"
	"github.com/colonyos/executors/common/pkg/sync"
	log "github.com/sirupsen/logrus"
)

type Executor struct {
	verbose            bool
	coloniesServerHost string
	coloniesServerPort int
	coloniesInsecure   bool
	colonyID           string
	colonyPrvKey       string
	executorID         string
	executorPrvKey     string
	executorType       string
	logDir             string
	fsDir              string
	swName             string
	swType             string
	swVersion          string
	hwCPU              string
	hwModel            string
	hwNodes            int
	hwMem              string
	hwStorage          string
	hwGPUCount         int
	hwGPUNodesCount    int
	hwGPUName          string
	hwGPUMem           string
	long               float64
	lat                float64
	locDesc            string
	ctx                context.Context
	cancel             context.CancelFunc
	client             *client.ColoniesClient
	syncHandler        *sync.SyncHandler
	failureHandler     *failure.FailureHandler
	debugHandler       *debug.DebugHandler
}

type ExecutorOption func(*Executor)

func WithVerbose(verbose bool) ExecutorOption {
	return func(e *Executor) {
		e.verbose = verbose
	}
}

func WithColoniesServerHost(host string) ExecutorOption {
	return func(e *Executor) {
		e.coloniesServerHost = host
	}
}

func WithColoniesServerPort(port int) ExecutorOption {
	return func(e *Executor) {
		e.coloniesServerPort = port
	}
}

func WithExecutorType(executorType string) ExecutorOption {
	return func(e *Executor) {
		e.executorType = executorType
	}
}

func WithColoniesInsecure(insecure bool) ExecutorOption {
	return func(e *Executor) {
		e.coloniesInsecure = insecure
	}
}

func WithColonyID(id string) ExecutorOption {
	return func(e *Executor) {
		e.colonyID = id
	}
}

func WithColonyPrvKey(prvkey string) ExecutorOption {
	return func(e *Executor) {
		e.colonyPrvKey = prvkey
	}
}

func WithExecutorID(id string) ExecutorOption {
	return func(e *Executor) {
		e.executorID = id
	}
}

func WithExecutorPrvKey(key string) ExecutorOption {
	return func(e *Executor) {
		e.executorPrvKey = key
	}
}

func WithLogdir(logDir string) ExecutorOption {
	return func(e *Executor) {
		e.logDir = logDir
	}
}

func WithFsDir(fsDir string) ExecutorOption {
	return func(e *Executor) {
		e.fsDir = fsDir
	}
}

func WithSoftwareName(swName string) ExecutorOption {
	return func(e *Executor) {
		e.swName = swName
	}
}

func WithSoftwareType(swType string) ExecutorOption {
	return func(e *Executor) {
		e.swType = swType
	}
}

func WithSoftwareVersion(swVersion string) ExecutorOption {
	return func(e *Executor) {
		e.swVersion = swVersion
	}
}

func WithHardwareCPU(hwCPU string) ExecutorOption {
	return func(e *Executor) {
		e.hwCPU = hwCPU
	}
}

func WithHardwareModel(hwModel string) ExecutorOption {
	return func(e *Executor) {
		e.hwModel = hwModel
	}
}

func WithHardwareNodes(hwNodes int) ExecutorOption {
	return func(e *Executor) {
		e.hwNodes = hwNodes
	}
}

func WithHardwareMemory(hwMem string) ExecutorOption {
	return func(e *Executor) {
		e.hwMem = hwMem
	}
}

func WithHardwareStorage(hwStorage string) ExecutorOption {
	return func(e *Executor) {
		e.hwStorage = hwStorage
	}
}

func WithHardwareGPUCount(hwGPUCount int) ExecutorOption {
	return func(e *Executor) {
		e.hwGPUCount = hwGPUCount
	}
}

func WithHardwareGPUNodesCount(hwGPUNodesCount int) ExecutorOption {
	return func(e *Executor) {
		e.hwGPUNodesCount = hwGPUNodesCount
	}
}

func WithHardwareGPUName(hwGPUName string) ExecutorOption {
	return func(e *Executor) {
		e.hwGPUName = hwGPUName
	}
}

func WithHardwareGPUMemory(hwGPUMem string) ExecutorOption {
	return func(e *Executor) {
		e.hwGPUMem = hwGPUMem
	}
}

func WithLong(long float64) ExecutorOption {
	return func(e *Executor) {
		e.long = long
	}
}

func WithLat(lat float64) ExecutorOption {
	return func(e *Executor) {
		e.lat = lat
	}
}

func WithLocDesc(locDesc string) ExecutorOption {
	return func(e *Executor) {
		e.locDesc = locDesc
	}
}

func (e *Executor) createColoniesExecutorWithKey(colonyID string) (*core.Executor, string, string, error) {
	crypto := crypto.CreateCrypto()
	executorPrvKey, err := crypto.GeneratePrivateKey()
	if err != nil {
		return nil, "", "", err
	}

	executorID, err := crypto.GenerateID(executorPrvKey)
	if err != nil {
		return nil, "", "", err
	}

	executor := core.CreateExecutor(executorID, e.executorType, e.executorType+"-"+core.GenerateRandomID(), colonyID, time.Now(), time.Now())
	executor.Capabilities.Software.Name = e.swName
	executor.Capabilities.Software.Type = e.swType
	executor.Capabilities.Software.Version = e.swVersion
	executor.Capabilities.Hardware.CPU = e.hwCPU
	executor.Capabilities.Hardware.Model = e.hwModel
	executor.Capabilities.Hardware.Nodes = e.hwNodes
	executor.Capabilities.Hardware.Storage = e.hwStorage
	executor.Capabilities.Hardware.Memory = e.hwMem
	executor.Capabilities.Hardware.GPU.Count = e.hwGPUCount
	executor.Capabilities.Hardware.GPU.NodeCount = e.hwGPUNodesCount
	executor.Capabilities.Hardware.GPU.Name = e.hwGPUName
	executor.Capabilities.Hardware.GPU.Memory = e.hwGPUMem
	executor.Location.Description = e.locDesc
	executor.Location.Long = e.long
	executor.Location.Lat = e.lat

	return executor, executorID, executorPrvKey, nil
}

func CreateExecutor(opts ...ExecutorOption) (*Executor, error) {
	e := &Executor{}
	for _, opt := range opts {
		opt(e)
	}

	ctx, cancel := context.WithCancel(context.Background())
	e.ctx = ctx
	e.cancel = cancel

	sigc := make(chan os.Signal)
	signal.Notify(sigc, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT, syscall.SIGSEGV)
	go func() {
		<-sigc
		e.Shutdown()
		os.Exit(1)
	}()

	e.client = client.CreateColoniesClient(e.coloniesServerHost, e.coloniesServerPort, e.coloniesInsecure, false)

	if e.colonyPrvKey != "" {
		spec, executorID, executorPrvKey, err := e.createColoniesExecutorWithKey(e.colonyID)
		if err != nil {
			return nil, err
		}
		e.executorID = executorID
		e.executorPrvKey = executorPrvKey

		_, err = e.client.AddExecutor(spec, e.colonyPrvKey)
		if err != nil {
			return nil, err
		}
		err = e.client.ApproveExecutor(e.executorID, e.colonyPrvKey)
		if err != nil {
			return nil, err
		}

		log.WithFields(log.Fields{"ExecutorID": e.executorID}).Info("Self-registered")
	}
	function := &core.Function{ExecutorID: e.executorID, ColonyID: e.colonyID, FuncName: "execute"}
	e.client.AddFunction(function, e.executorPrvKey)

	var err error
	e.failureHandler, err = failure.CreateFailureHandler(e.executorPrvKey, e.client)
	if err != nil {
		return nil, err
	}

	e.debugHandler, err = debug.CreateDebugHandler(e.executorPrvKey, e.client)
	if err != nil {
		return nil, err
	}

	e.syncHandler, err = sync.CreateSyncHandler(e.colonyID, e.executorPrvKey, e.client, e.fsDir, e.failureHandler, e.debugHandler)
	if err != nil {
		return nil, err
	}

	log.WithFields(log.Fields{
		"Verbose":               e.verbose,
		"ColoniesServerHost":    e.coloniesServerHost,
		"ColoniesServerPort":    e.coloniesServerPort,
		"ColoniesInsecure":      e.coloniesInsecure,
		"LogDir":                e.logDir,
		"FsDir":                 e.fsDir,
		"ColonyId":              e.colonyID,
		"ColonyPrvKey":          "***********************",
		"ExecutorId":            e.executorID,
		"ExecutorPrvKey":        "***********************",
		"Longitude":             e.long,
		"Latitude":              e.lat,
		"LocationDesc":          e.locDesc,
		"HardwareModel":         e.hwModel,
		"HardwareNodes":         e.hwNodes,
		"HardwareCPU":           e.hwCPU,
		"HardwareMemory":        e.hwMem,
		"HardwareStorage":       e.hwStorage,
		"HardwareGPUName":       e.hwGPUName,
		"HardwareGPUCount":      e.hwGPUCount,
		"HardwareGPUNodesCount": e.hwGPUNodesCount,
		"HardwareGPUMemory":     e.hwGPUMem,
		"SoftwareName":          e.swName,
		"SoftwareVersion":       e.swVersion,
		"SoftwareType":          e.swType,
		"ExecutorType":          e.executorType}).
		Info("Kube Executor started")

	return e, nil
}

func (e *Executor) Shutdown() error {
	log.Info("Shutting down")
	if e.colonyPrvKey != "" {
		err := e.client.DeleteExecutor(e.executorID, e.colonyPrvKey)
		if err != nil {
			log.WithFields(log.Fields{"ExecutorID": e.executorID}).Warning("Failed to deregistered")
		}

		log.WithFields(log.Fields{"ExecutorID": e.executorID}).Info("Deregistered")
	}
	e.cancel()
	return nil
}

func (e *Executor) executeK8s(process *core.Process) error {
	kwArgs, err := parsers.ParseKwArgs(process, e.failureHandler, e.debugHandler)
	if err != nil {
		return err
	}

	fmt.Println(kwArgs)

	return nil
}

func (e *Executor) ServeForEver() error {
	for {
		process, err := e.client.AssignWithContext(e.colonyID, 100, e.ctx, e.executorPrvKey)
		if err != nil {
			var coloniesError *core.ColoniesError
			if errors.As(err, &coloniesError) {
				if coloniesError.Status == 404 { // No processes can be selected for executor
					log.Info(err)
					continue
				}
			}

			log.Error(err)
			log.Error("Retrying in 5 seconds ...")
			time.Sleep(5 * time.Second)
			continue
		}

		log.WithFields(log.Fields{"ProcessID": process.ID, "ExecutorID": e.executorID}).Info("Assigned process to executor")

		if process.FunctionSpec.FuncName == "execute" {
			fmt.Println("Exec process")
			err = e.executeK8s(process)
			if err != nil {
				e.failureHandler.HandleError(process, err, "Failed to execute K8s")
			}

			e.client.Close(process.ID, e.executorPrvKey)
		} else {
			log.WithFields(log.Fields{"FuncName": process.FunctionSpec.FuncName}).Error("Unsupported funcname")
			err := e.client.Fail(process.ID, []string{"Unsupported funcname"}, e.executorPrvKey)
			if err != nil {
				log.WithFields(log.Fields{"ProcessId": process.ID, "Error": err}).Error("Failed to close process as failed")
			}
			continue
		}
	}
}