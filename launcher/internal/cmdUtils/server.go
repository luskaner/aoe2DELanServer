package cmdUtils

import (
	"fmt"
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/luskaner/aoe2DELanServer/common"
	launcherCommon "github.com/luskaner/aoe2DELanServer/launcher-common"
	commonExecutor "github.com/luskaner/aoe2DELanServer/launcher-common/executor/exec"
	"github.com/luskaner/aoe2DELanServer/launcher/internal"
	"github.com/luskaner/aoe2DELanServer/launcher/internal/server"
	"net"
	"sort"
	"strings"
)

func SelectBestServerIp(ips []string) (ok bool, ip string) {
	var successIps []net.IP

	for _, curIp := range ips {
		if server.LanServer(curIp, true) {
			parsedIp := net.ParseIP(curIp)
			if parsedIp.IsLoopback() {
				return true, curIp
			}
			successIps = append(successIps, net.ParseIP(curIp).To4())
		}
	}

	countSuccessIps := len(successIps)
	if countSuccessIps == 0 {
		return
	}

	ok = true
	ip = successIps[0].String()
	interfaces, err := net.Interfaces()

	if err != nil {
		return
	}

	var addrs []net.Addr
	for _, i := range interfaces {
		addrs, err = i.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			v, addrOk := addr.(*net.IPNet)
			if !addrOk {
				continue
			}

			for _, curIp := range successIps {
				if v.Contains(curIp) {
					ip = curIp.String()
					return
				}
			}
		}
	}

	return
}

func ListenToServerAnnouncementsAndSelectBestIp(gameId string, multicastIPs []net.IP, ports []int) (errorCode int, ip string) {
	errorCode = common.ErrSuccess
	servers := server.LanServersAnnounced(multicastIPs, ports)
	if servers == nil {
		fmt.Println("Could not listen to server announcements. Maybe the UDP port", common.AnnouncePort, "is blocked or already in use.")
		errorCode = internal.ErrListenServerAnnouncements
	}
	if servers != nil && len(servers) > 0 {
		var ok bool
		var serverTags []string
		var serversStr [][]string
		announcedNewerVersion := false
		announcedOlderVersion := false
		for _, data := range servers {
			if data.Version >= common.AnnounceVersion1 {
				announceData := data.Data.(common.AnnounceMessageData001)
				gameIdSet := mapset.NewSet[string](announceData.GameIds...)
				if !gameIdSet.ContainsOne(gameId) {
					continue
				}
			}
			ips := data.Ips.ToSlice()
			sort.Strings(ips)
			hosts := mapset.NewSet[string]()
			for _, ip := range ips {
				hosts.Append(launcherCommon.IpToHosts(ip).ToSlice()...)
			}
			ipsStr := strings.Join(ips, ", ")
			hostsStr := ""
			suffix := ""
			if !hosts.IsEmpty() {
				hostsSlice := hosts.ToSlice()
				sort.Strings(hostsSlice)
				hostsStr = strings.Join(hostsSlice, ", ")
			}
			suffix = fmt.Sprintf("- v. %d", data.Version)
			if data.Version > common.AnnounceVersionLatest {
				announcedNewerVersion = true
			} else if data.Version < common.AnnounceVersionLatest {
				announcedOlderVersion = true
			}
			var strVars []interface{}
			strVars = append(strVars, ipsStr)
			format := "%s"
			if len(hostsStr) > 0 {
				format += " (%s)"
				strVars = append(strVars, hostsStr)
			}
			format += " %s"
			strVars = append(strVars, suffix)
			serverTags = append(serverTags, fmt.Sprintf(format, strVars...))
			serversStr = append(serversStr, ips)
		}
		if announcedNewerVersion {
			fmt.Println("Found at least a server with a newer version than the client. The launcher should be upgraded.")
		}
		if announcedOlderVersion {
			fmt.Println("Found at least a server with an older version than the client. The server(s) should be upgraded.")
		}
		if len(servers) == 1 {
			fmt.Printf("Found a single server \"%s\", will connect to it...\n", serverTags[0])
			ok, ip = SelectBestServerIp(serversStr[0])
			if !ok {
				fmt.Println("Server is not reachable. Check the client can connect to", ip, "on TCP port 443 (HTTPS)")
				errorCode = internal.ErrServerUnreachable
				return
			}
		} else {
			var option int
			for {
				fmt.Println("Found the following servers:")
				for i := range serversStr {
					fmt.Printf("%d. %s\n", i+1, serverTags[i])
				}
				fmt.Printf("Enter the number of the server (1-%d): ", len(serversStr))
				_, err := fmt.Scan(&option)
				if err != nil || option < 1 || option > len(serversStr) {
					fmt.Println("Invalid (or error reading) option. Please enter a number from the list.")
					continue
				}
				if option == len(serversStr) {
					break
				}
				ips := serversStr[option-1]
				ok, ip = SelectBestServerIp(ips)
				if ok {
					break
				} else {
					fmt.Println(fmt.Sprintf("Server #%d is not reachable. Check the client can connect to it on TCP port 443 (HTTPS).", option))
					fmt.Println("Please enter the same (to retry) or another number from the list")
				}
			}
		}
	}
	return
}

func (c *Config) StartServer(executable string, args []string, stop bool, canTrustCertificate bool) (errorCode int, ip string) {
	serverExecutablePath := server.GetExecutablePath(executable)
	if serverExecutablePath == "" {
		fmt.Println("Cannot find server executable path. Set it manually in Server.Executable.")
		errorCode = internal.ErrServerExecutable
		return
	}
	if executable != serverExecutablePath {
		fmt.Println("Found server executable path:", serverExecutablePath)
	}
	if !common.HasCertificatePair(serverExecutablePath) {
		if !canTrustCertificate {
			fmt.Println("serverStart is true and canTrustCertificate is false. Certificate pair is missing. Generate your own certificates manually.")
			errorCode = internal.ErrServerCertMissing
			return
		}
		certificateFolder := common.CertificatePairFolder(serverExecutablePath)
		if certificateFolder == "" {
			fmt.Println("Cannot find certificate folder of the server. Make sure the folder structure of the server is correct.")
			errorCode = internal.ErrServerCertDirectory
			return
		}
		if result := server.GenerateCertificatePair(certificateFolder); !result.Success() {
			fmt.Println("Failed to generate certificate pair. Check the folder and its permissions")
			errorCode = internal.ErrServerCertCreate
			if result.Err != nil {
				fmt.Println("Error message: " + result.Err.Error())
			}
			if result.ExitCode != common.ErrSuccess {
				fmt.Printf(`Exit code: %d. See documentation for "genCert" to check what it means.`+"\n", result.ExitCode)
			}
			return
		}
	}
	fmt.Println("Starting server, authorize 'server' in firewall if needed...")
	var stopStr string
	if stop {
		stopStr = "true"
	} else {
		stopStr = "false"
	}
	var result *commonExecutor.Result
	var serverExe string
	result, serverExe, ip = server.StartServer(stopStr, executable, args)
	if result.Success() {
		fmt.Println("Server started.")
		if stop {
			c.SetServerExe(serverExe)
		}
	} else {
		fmt.Println("Could not start server.")
		errorCode = internal.ErrServerStart
		if result != nil {
			if result.Err != nil {
				fmt.Println("Error message: " + result.Err.Error())
			}
			if result.ExitCode != common.ErrSuccess {
				fmt.Printf(`Exit code: %d. See documentation for "server" to check what it means`+"\n", result.ExitCode)
			}
		} else {
			fmt.Println("Try running the server manually.")
		}
	}
	return
}
