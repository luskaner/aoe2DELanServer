package server

import (
	"crypto/tls"
	"errors"
	mapset "github.com/deckarep/golang-set/v2"
	"golang.org/x/sys/windows"
	"launcher/internal"
	"launcher/internal/executor"
	"net"
	"net/http"
	"net/netip"
	"os"
	"os/exec"
	"shared"
	"time"
)

const autoServerExecutable string = "server.exe"

var autoServerPaths = []string{`.\`, `..\`, `..\server\`}

func StartServer(config internal.ServerConfig) (bool, *exec.Cmd) {
	if config.Start == "false" {
		return false, nil
	}
	executablePath := GetExecutablePath(config)
	if executablePath == "" {
		return false, nil
	}
	var ok bool
	var cmd *exec.Cmd = nil
	if config.Stop == "true" {
		_, cmd = executor.StartCustomExecutable(executablePath, true)
		ok = cmd != nil
	} else {
		ok = executor.ShellExecute("open", executablePath, true, windows.SW_MINIMIZE) == nil
	}
	if ok {
		// Wait up to 30s for server to start
		for i := 0; i < 30; i++ {
			for ip := range shared.HostOrIpToIps(config.Host).Iter() {
				if LanServer(ip, true) {
					return true, cmd
				}
			}
			time.Sleep(time.Second)
		}
		if cmd != nil {
			return StopServer(cmd), cmd
		}
	}
	return false, nil
}

func StopServer(cmd *exec.Cmd) bool {
	err := cmd.Process.Kill()
	if err != nil {
		return false
	}
	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	select {
	case <-time.After(10 * time.Second):
		return false
	case err := <-done:
		if err != nil {
			var e *exec.ExitError
			if !errors.As(err, &e) {
				return false
			}
		}
		return true
	}
}

func GetExecutablePath(config internal.ServerConfig) string {
	if config.Executable == "auto" {
		for _, path := range autoServerPaths {
			fullPath := path + autoServerExecutable
			if f, err := os.Stat(fullPath); err == nil && !f.IsDir() {
				return fullPath
			}
		}
		return ""
	}
	return config.Executable
}

func LanServer(host string, insecureSkipVerify bool) bool {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: insecureSkipVerify},
	}
	client := &http.Client{Transport: tr}
	resp, err := client.Get("https://" + host + "/test")
	if err != nil {
		return false
	}
	return resp.StatusCode == http.StatusOK
}

func LanServersAnnounced() mapset.Set[string] {
	addr := net.UDPAddr{
		Port: 59999,
		IP:   netip.IPv4Unspecified().AsSlice(),
	}

	conn, err := net.ListenUDP("udp", &addr)
	if err != nil {
		return nil
	}
	defer func(conn *net.UDPConn) {
		_ = conn.Close()
	}(conn)

	err = conn.SetReadDeadline(time.Now().Add(15 * time.Second))
	if err != nil {
		return nil
	}

	addresses := mapset.NewSet[string]()
	buf := make([]byte, 1)
	for {
		n, addr, err := conn.ReadFromUDP(buf)
		if err != nil {
			break
		} else if n == 1 && buf[0] == 43 {
			addresses.Add(addr.IP.String())
		}
	}
	return addresses
}
