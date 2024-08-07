package internal

import (
	"github.com/luskaner/aoe2DELanServer/battle-server-broadcast"
	"github.com/luskaner/aoe2DELanServer/common"
	commonProcess "github.com/luskaner/aoe2DELanServer/common/process"
	"golang.org/x/sys/windows"
	"time"
)

var processWaitInterval = 1 * time.Second

func waitUntilAnyProcessExist(names []string) (processesEntryNames map[string]windows.ProcessEntry32) {
	for i := 0; i < int((1*time.Minute)/processWaitInterval); i++ {
		processesEntryNames = commonProcess.ProcessesEntryNames(names)
		if len(processesEntryNames) > 0 {
			return
		}
		time.Sleep(processWaitInterval)
	}
	return
}

func waitForProcess(processesEntryName windows.ProcessEntry32) bool {
	handle, err := windows.OpenProcess(windows.SYNCHRONIZE, true, processesEntryName.ProcessID)

	if err != nil {
		return false
	}

	defer func(handle windows.Handle) {
		_ = windows.CloseHandle(handle)
	}(handle)

	_, err = windows.WaitForSingleObject(handle, windows.INFINITE)

	if err != nil {
		return false
	}

	return true
}

func Watch(steamProcess bool, microsoftStoreProcess bool, serverExe string, broadcastBattleServer bool, revertArgs []string, exitCode *int) {
	*exitCode = common.ErrSuccess
	if len(revertArgs) > 0 {
		defer func() {
			RunConfig(revertArgs)
		}()
	}
	processes := waitUntilAnyProcessExist(commonProcess.GameProcesses(steamProcess, microsoftStoreProcess))
	if len(processes) == 0 {
		*exitCode = ErrGameTimeoutStart
		return
	}
	if broadcastBattleServer {
		mostPriority, restInterfaces, err := battle_server_broadcast.RetrieveBsInterfaceAddresses()
		if err == nil && mostPriority != nil && len(restInterfaces) > 0 {
			if len(waitUntilAnyProcessExist([]string{"BattleServer.exe"})) > 0 {
				go func() {
					_ = battle_server_broadcast.CloneAnnouncements(mostPriority, restInterfaces)
				}()
			} else {
				*exitCode = ErrBattleServerTimeOutStart
			}
		}
	}
	var process windows.ProcessEntry32
	for _, p := range processes {
		process = p
		break
	}
	if waitForProcess(process) {
		if serverExe != "-" {
			if _, err := commonProcess.Kill(serverExe); err != nil {
				*exitCode = ErrFailedStopServer
			}
		}
	} else {
		*exitCode = ErrFailedWaitForProcess
	}
}
