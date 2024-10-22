package server

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/google/uuid"
	"github.com/luskaner/aoe2DELanServer/common"
	commonProcess "github.com/luskaner/aoe2DELanServer/common/process"
	launcherCommon "github.com/luskaner/aoe2DELanServer/launcher-common"
	commonExecutor "github.com/luskaner/aoe2DELanServer/launcher-common/executor/exec"
	"golang.org/x/net/ipv4"
	"net"
	"net/netip"
	"os"
	"path"
	"path/filepath"
	"time"
)

var autoServerDir = []string{fmt.Sprintf("%c", filepath.Separator), fmt.Sprintf(`%c%s%c`, filepath.Separator, common.Server, filepath.Separator), fmt.Sprintf(`%c..%c`, filepath.Separator, filepath.Separator), fmt.Sprintf(`%c..%c%s%c`, filepath.Separator, filepath.Separator, common.Server, filepath.Separator)}
var autoServerName = []string{common.GetExeFileName(true, common.Server)}

func StartServer(stop string, executable string, args []string) (result *commonExecutor.Result, executablePath string, ip string) {
	executablePath = GetExecutablePath(executable)
	if executablePath == "" {
		return
	}
	var showWindow bool
	if stop == "true" {
		showWindow = false
	} else {
		showWindow = true
	}
	result = commonExecutor.Options{File: executablePath, Args: args, ShowWindow: showWindow, Pid: true}.Exec()
	if result.Success() {
		// Wait up to 30s for server to start
		for i := 0; i < 30; i++ {
			for curIp := range launcherCommon.HostOrIpToIps(netip.IPv4Unspecified().String()).Iter() {
				if LanServer(curIp, true) {
					ip = curIp
					return
				}
			}
			time.Sleep(time.Second)
		}
		if proc, err := commonProcess.Kill(executablePath); err != nil {
			fmt.Println("Failed to stop server")
			fmt.Println("Error message: " + result.Err.Error())
			if proc != nil {
				fmt.Println("You may try killing it manually. Search for the process PID inside server.pid if it exists")
			}
		}
		result = nil
	}
	return
}

func GetExecutablePath(executable string) string {
	if executable == "auto" {
		ex, err := os.Executable()
		if err != nil {
			return ""
		}
		exePath := filepath.Dir(ex)
		var f os.FileInfo
		for _, dir := range autoServerDir {
			dirPath := exePath + dir
			for _, name := range autoServerName {
				p := dirPath + name
				if f, err = os.Stat(p); err == nil && !f.IsDir() {
					return path.Clean(p)
				}
			}
		}
		return ""
	}
	return executable
}

func LanServer(host string, insecureSkipVerify bool) bool {
	return HttpGet(fmt.Sprintf("https://%s/test", host), insecureSkipVerify) == common.ErrSuccess
}

func announcementConnections(multicastIPs []net.IP, ports []int) []*net.UDPConn {
	var connections []*net.UDPConn
	var multicastIfs []*net.Interface
	if len(multicastIPs) > 0 {
		interfaces, err := net.Interfaces()
		if err == nil {
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
					var IP net.IP
					if IP = v.IP.To4(); IP == nil {
						continue
					}
					if i.Flags&net.FlagRunning != 0 && i.Flags&net.FlagMulticast != 0 {
						multicastIfs = append(multicastIfs, &i)
					}
				}
			}
		}
	}
	for _, port := range ports {
		addr := &net.UDPAddr{
			IP:   netip.IPv4Unspecified().AsSlice(),
			Port: port,
		}
		conn, err := net.ListenUDP("udp4", addr)
		if err != nil {
			continue
		}
		if len(multicastIPs) > 0 {
			p := ipv4.NewPacketConn(conn)
			for _, multicastIP := range multicastIPs {
				multicastAddr := &net.UDPAddr{
					IP:   multicastIP,
					Port: port,
				}
				for _, multicastIf := range multicastIfs {
					_ = p.JoinGroup(multicastIf, multicastAddr)
				}
			}
		}
		connections = append(connections, conn)
	}
	return connections
}

func LanServersAnnounced(multicastIPs []net.IP, ports []int) map[uuid.UUID]*common.AnnounceMessage {
	results := make(chan map[uuid.UUID]*common.AnnounceMessage)
	connections := announcementConnections(multicastIPs, ports)
	for _, conn := range connections {
		go func() {
			defer func(conn *net.UDPConn) {
				_ = conn.Close()
			}(conn)

			err := conn.SetReadDeadline(time.Now().Add(15 * time.Second))
			if err != nil {
				return
			}

			packetBuffer := make([]byte, 65_536)
			headerBuffer := make([]byte, len(common.AnnounceHeader))
			var messageLenBuffer uint16
			var messageBuffer *bytes.Buffer
			servers := make(map[uuid.UUID]*common.AnnounceMessage)
			var n int
			var serverAddr *net.UDPAddr

			for {
				_, serverAddr, err = conn.ReadFromUDP(packetBuffer)
				if err != nil {
					break
				}
				n = copy(headerBuffer, packetBuffer)
				if n < len(common.AnnounceHeader) || string(headerBuffer) != common.AnnounceHeader {
					continue
				}
				remainingPacketBuffer := packetBuffer[n:]
				version := remainingPacketBuffer[:common.AnnounceVersionLength][0]
				remainingPacketBuffer = remainingPacketBuffer[common.AnnounceVersionLength:]
				var id uuid.UUID
				id, err = uuid.FromBytes(remainingPacketBuffer[:common.AnnounceIdLength])
				if err != nil {
					continue
				}
				remainingPacketBuffer = remainingPacketBuffer[common.AnnounceIdLength:]
				err = binary.Read(bytes.NewReader(remainingPacketBuffer[2:]), binary.LittleEndian, &messageLenBuffer)
				if err != nil {
					continue
				}
				remainingPacketBuffer = remainingPacketBuffer[2:]
				messageBuffer = bytes.NewBuffer(remainingPacketBuffer[:messageLenBuffer])
				var data interface{}
				switch version {
				case common.AnnounceVersion0:
					var msg common.AnnounceMessageData000
					dec := gob.NewDecoder(messageBuffer)
					if err = dec.Decode(&msg); err == nil {
						data = msg
					}
				case common.AnnounceVersion1:
					var msg common.AnnounceMessageData001
					dec := gob.NewDecoder(messageBuffer)
					if err = dec.Decode(&msg); err == nil {
						data = msg
					}
				}
				ip := serverAddr.IP.String()
				var m *common.AnnounceMessage
				var ok bool
				if m, ok = servers[id]; !ok {
					m = &common.AnnounceMessage{
						Version: version,
						Data:    data,
						Ips:     mapset.NewSet[string](),
					}
					servers[id] = m
				}
				m.Ips.Add(ip)
			}

			results <- servers
		}()
	}

	servers := make(map[uuid.UUID]*common.AnnounceMessage)
	for range ports {
		for id, server := range <-results {
			if _, ok := servers[id]; !ok {
				servers[id] = server
			} else {
				for ip := range server.Ips.Iter() {
					servers[id].Ips.Add(ip)
				}
			}
		}
	}

	return servers
}
