package cmd

import (
	"fmt"
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/luskaner/aoe2DELanServer/common"
	"github.com/luskaner/aoe2DELanServer/common/executor"
	commonProcess "github.com/luskaner/aoe2DELanServer/common/process"
	launcherCommon "github.com/luskaner/aoe2DELanServer/launcher-common"
	"github.com/luskaner/aoe2DELanServer/launcher-common/cmd"
	"github.com/luskaner/aoe2DELanServer/launcher-config/internal"
	"github.com/luskaner/aoe2DELanServer/launcher-config/internal/cmd/wrapper"
	"github.com/luskaner/aoe2DELanServer/launcher-config/internal/userData"
	"github.com/spf13/cobra"
	"os"
	"os/signal"
	"runtime"
	"syscall"
)

func removeUserCert() bool {
	fmt.Println("Removing previously added user certificate, authorize it if needed ...")
	if _, err := wrapper.RemoveUserCert(); err == nil {
		fmt.Println("Successfully removed user certificate")
		return true
	} else {
		fmt.Println("Failed to remove user certificate")
		return false
	}
}

func restoreMetadata() bool {
	fmt.Println("Restoring previously backed up metadata")
	if userData.Metadata(gameId).Restore(gameId) {
		fmt.Println("Successfully restored metadata")
		return true
	} else {
		fmt.Println("Failed to restore metadata")
		return false
	}
}

func restoreProfiles() bool {
	fmt.Println("Restoring previously backed up profiles")
	if userData.RestoreProfiles(gameId, true) {
		fmt.Println("Successfully restored profiles")
		return true
	} else {
		fmt.Println("Failed to restore profiles")
		return false
	}
}

func undoSetUp(addedUserCert bool, backedUpMetadata bool, backedUpProfiles bool) {
	if addedUserCert {
		removeUserCert()
	}
	if backedUpMetadata {
		restoreMetadata()
	}
	if backedUpProfiles {
		restoreProfiles()
	}
}

var AddUserCertData []byte
var BackupMetadata bool
var BackupProfiles bool
var agentStart bool
var agentEndOnError bool
var storeString = "local"

var setUpCmd = &cobra.Command{
	Use:   "setup",
	Short: "Setups configuration",
	Long:  "Adds any of the following:\n* One or more host mappings to the local DNS server\n* Certificate to the " + storeString + " machine's trusted root store\n* Backup user metadata\n* Backup user profiles",
	Run: func(_ *cobra.Command, _ []string) {
		if len(cmd.MapIPs) > 9 {
			fmt.Println("Too many IPs. Up to 9 can be mapped")
			os.Exit(launcherCommon.ErrIpMapAddTooMany)
		}
		var addedUserCert bool
		var backedUpMetadata bool
		var backedUpProfiles bool
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
		go func() {
			_, ok := <-sigs
			if ok {
				undoSetUp(addedUserCert, backedUpMetadata, backedUpProfiles)
				os.Exit(common.ErrSignal)
			}
		}()
		if (BackupMetadata || BackupProfiles) && !common.ValidGame(gameId) {
			fmt.Println("Invalid game type")
			os.Exit(launcherCommon.ErrInvalidGame)
		}
		isAdmin := executor.IsAdmin()
		if AddUserCertData != nil {
			fmt.Println("Adding user certificate, authorize it if needed...")
			crt := wrapper.BytesToCertificate(AddUserCertData)
			if crt == nil {
				fmt.Println("Failed to parse certificate")
				os.Exit(internal.ErrUserCertAddParse)
			}
			if err := wrapper.AddUserCert(crt); err == nil {
				fmt.Println("Successfully added user certificate")
				addedUserCert = true
			} else {
				fmt.Println("Failed to add user certificate")
				fmt.Println("Error message: " + err.Error())
				os.Exit(internal.ErrUserCertAdd)
			}
		}
		if BackupMetadata {
			fmt.Println("Backing up metadata")
			if userData.Metadata(gameId).Backup(gameId) {
				fmt.Println("Successfully backed up metadata")
				backedUpMetadata = true
			} else {
				errorCode := internal.ErrMetadataBackup
				if addedUserCert {
					if !removeUserCert() {
						errorCode = internal.ErrMetadataBackupRevert
					}
				}
				fmt.Println("Failed to back up metadata")
				os.Exit(errorCode)
			}
		}
		if BackupProfiles {
			fmt.Println("Backing up profiles")
			if userData.BackupProfiles(gameId) {
				fmt.Println("Successfully backed up profiles")
				backedUpProfiles = true
			} else {
				errorCode := internal.ErrProfilesBackup
				if addedUserCert {
					if !removeUserCert() {
						errorCode = internal.ErrProfilesBackupRevert
					}
				}
				if backedUpMetadata {
					if !restoreMetadata() {
						errorCode = internal.ErrProfilesBackupRevert
					}
				}
				fmt.Println("Failed to back up profiles")
				os.Exit(errorCode)
			}
		}
		hostMappings := mapset.NewSet[string]()
		if len(cmd.MapIPs) > 0 {
			for _, ip := range cmd.MapIPs {
				hostMappings.Add(ip.String())
			}
		}
		if cmd.AddLocalCertData != nil || !hostMappings.IsEmpty() || cmd.MapCDN {
			agentStarted := internal.ConnectAgentIfNeeded() == nil
			if !agentStarted && agentStart && !isAdmin {
				result := internal.StartAgentIfNeeded()
				if !result.Success() {
					fmt.Println("Failed to start config-admin-agent")
					if result.Err != nil {
						fmt.Println(result.Err)
					}
					if result.ExitCode != common.ErrSuccess {
						fmt.Println(result.ExitCode)
					}
					os.Exit(internal.ErrStartAgent)
				} else {
					agentStarted = internal.ConnectAgentIfNeededWithRetries(true)
					if !agentStarted {
						fmt.Println("Failed to connect to config-admin-agent after starting it. Kill the 'config-admin-agent' process manually")
						os.Exit(internal.ErrStartAgentVerify)
					}
				}
			}
			if agentStarted {
				fmt.Println("Communicating with config-admin-agent to add local cert and/or host mappings...")
			} else {
				if isAdmin {
					fmt.Println("Running config-admin to add local cert and/or host mappings...")
				} else {
					fmt.Println("Running config-admin to add local cert and/or host mappings, authorize it if needed...")
				}
			}
			err, exitCode := internal.RunSetUp(hostMappings, cmd.AddLocalCertData, cmd.MapCDN)
			if err == nil && exitCode == common.ErrSuccess {
				if agentStarted {
					fmt.Println("Successfully communicated with config-admin-agent")
				} else {
					fmt.Println("Successfully ran config-admin")
				}
			} else {
				if err != nil {
					fmt.Println("Received error:")
					fmt.Println(err)
				}
				if exitCode != common.ErrSuccess {
					fmt.Println("Received exit code:")
					fmt.Println(exitCode)
				}
				errorCode := internal.ErrAdminSetup
				if addedUserCert {
					if !removeUserCert() {
						errorCode = internal.ErrAdminSetupRevert
					}
				}
				if backedUpMetadata {
					if !restoreMetadata() {
						errorCode = internal.ErrAdminSetupRevert
					}
				}
				if backedUpProfiles {
					if !restoreProfiles() {
						errorCode = internal.ErrAdminSetupRevert
					}
				}
				if agentStarted {
					fmt.Println("Failed to communicate with config-admin-agent. Communicating with it to shutdown...")
					if agentEndOnError {
						if err := internal.StopAgentIfNeeded(); err != nil {
							failedStopAgent := true
							if isAdmin {
								_, err := commonProcess.Kill(common.GetExeFileName(true, common.LauncherConfigAdminAgent))
								if err == nil {
									fmt.Println("Successfully killed config-admin-agent.")
									failedStopAgent = false
								}
							}
							if failedStopAgent {
								fmt.Println("Failed to stop config-admin-agent. Kill the 'config-admin-agent' process manually")
								fmt.Println("Error message: " + err.Error())
								os.Exit(internal.ErrStartAgentRevert)
							}
						} else {
							fmt.Println("Successfully stopped config-admin-agent.")
						}
					}
				} else {
					fmt.Println("Failed to run config-admin")
				}
				os.Exit(errorCode)
			}
		}
	},
}

func InitSetUp() {
	if runtime.GOOS != "linux" {
		storeString = "user/" + storeString
	}
	cmd.InitSetUp(setUpCmd)
	setUpCmd.Flags().StringVarP(
		&gameId,
		"game",
		"e",
		common.GameAoE2,
		fmt.Sprintf(`Game type. Only "%s" is currently supported.`, common.GameAoE2),
	)
	if runtime.GOOS != "linux" {
		setUpCmd.Flags().BytesBase64VarP(
			&AddUserCertData,
			"userCert",
			"u",
			nil,
			"Add the certificate to the user's trusted root store",
		)
	}
	setUpCmd.Flags().BoolVarP(
		&BackupMetadata,
		"metadata",
		"m",
		false,
		"Backup metadata",
	)
	setUpCmd.Flags().BoolVarP(
		&BackupProfiles,
		"profiles",
		"p",
		false,
		"Backup profiles",
	)
	setUpCmd.Flags().BoolVarP(
		&agentStart,
		"agentStart",
		"g",
		false,
		"Start the config-admin-agent if it is not running, we are not admin and is needed for admin action.",
	)
	setUpCmd.Flags().BoolVarP(
		&agentEndOnError,
		"agentEndOnError",
		"r",
		false,
		"Stop the config-admin-agent if it is running and any admin action failed.",
	)
	err := setUpCmd.Flags().MarkHidden("agentStart")
	if err != nil {
		panic(err)
	}
	err = setUpCmd.Flags().MarkHidden("agentEndOnError")
	if err != nil {
		panic(err)
	}
	RootCmd.AddCommand(setUpCmd)
}
