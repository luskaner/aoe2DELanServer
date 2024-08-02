package internal

import (
	"github.com/luskaner/aoe2DELanServer/common"
	launcherCommon "github.com/luskaner/aoe2DELanServer/launcherCommon"
	"github.com/luskaner/aoe2DELanServer/launcherCommon/executor"
)

func RunConfig(revertFlags []string) {
	args := []string{launcherCommon.ConfigRevertCmd}
	args = append(args, revertFlags...)
	executor.ExecOptions{File: common.GetExeFileName(true, common.LauncherConfig), ExitCode: true, Wait: true, Args: args}.Exec()
}
