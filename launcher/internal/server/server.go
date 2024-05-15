package server

import (
	"crypto/tls"
	"golang.org/x/sys/windows"
	"launcher/internal"
	"launcher/internal/executor"
	"net"
	"net/http"
	"os"
	"os/exec"
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
	var ok = false
	var cmd *exec.Cmd = nil
	if config.Stop == "true" {
		cmd := executor.StartCustomExecutable(executablePath, true)
		ok = cmd != nil
	} else {
		ok = executor.ShellExecute("open", executablePath, true, windows.SW_MINIMIZE)
	}
	if ok {
		for {
			if LanServer(config.Host, true) {
				break
			}
			time.Sleep(1 * time.Second)
		}
	}
	return ok, cmd
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

func WaitForLanServerAnnounce() *net.UDPAddr {
	addr := net.UDPAddr{
		Port: 59999,
		IP:   net.ParseIP("0.0.0.0"),
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

	buf := make([]byte, 1)
	for {
		n, addr, err := conn.ReadFromUDP(buf)
		if err != nil {
			return nil
		} else if n == 1 && buf[0] == 43 {
			return addr
		} else {
			return nil
		}
	}
}
