package cmd

import (
	"context"
	"errors"
	"fmt"
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/gorilla/handlers"
	"github.com/luskaner/aoe2DELanServer/common"
	"github.com/luskaner/aoe2DELanServer/common/executor"
	"github.com/luskaner/aoe2DELanServer/common/pidLock"
	"github.com/luskaner/aoe2DELanServer/server/internal"
	"github.com/luskaner/aoe2DELanServer/server/internal/ip"
	"github.com/luskaner/aoe2DELanServer/server/internal/middleware"
	"github.com/luskaner/aoe2DELanServer/server/internal/models/initializer"
	"github.com/luskaner/aoe2DELanServer/server/internal/routes"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"io"
	"log"
	"net"
	"net/http"
	"net/netip"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"
)

var configPaths = []string{path.Join("resources", "config"), "."}

var (
	Version string
	cfgFile string
	rootCmd = &cobra.Command{
		Use:   filepath.Base(os.Args[0]),
		Short: "server is a service acting as " + common.Domain + " for LAN features in AoE 2:DE.",
		Run: func(_ *cobra.Command, _ []string) {
			lock := &pidLock.Lock{}
			if err := lock.Lock(); err != nil {
				fmt.Println("Failed to lock pid file. You may try checking if the process in PID file exists (and killing it).")
				fmt.Println(err.Error())
				os.Exit(common.ErrPidLock)
			}
			gameSet := mapset.NewSet[string](viper.GetStringSlice("default.Games")...)
			if gameSet.IsEmpty() {
				fmt.Println("No games specified")
				_ = lock.Unlock()
				os.Exit(internal.ErrGames)
			}
			for game := range gameSet.Iter() {
				if !common.ValidGame(game) {
					fmt.Println("Invalid game specified:", game)
					_ = lock.Unlock()
					os.Exit(internal.ErrGames)
				}
			}
			if executor.IsAdmin() {
				fmt.Println("Running as administrator, this is not recommended for security reasons.")
				if runtime.GOOS == "linux" {
					fmt.Println(fmt.Sprintf("If the issue is that you cannot listen on the port, then run `sudo setcap CAP_NET_BIND_SERVICE=+eip '%s'`, before re-running the server", os.Args[0]))
				}
			}
			hosts := viper.GetStringSlice("default.Hosts")
			addrs := ip.ResolveHosts(hosts)
			if addrs == nil || len(addrs) == 0 {
				fmt.Println("Failed to resolve host (or it was an Ipv6 address)")
				_ = lock.Unlock()
				os.Exit(internal.ErrResolveHost)
			}
			mux := http.NewServeMux()
			initializer.InitializeGames(gameSet)
			routes.Initialize(mux)
			gameMux := middleware.GameMiddleware(mux)
			sessionMux := middleware.SessionMiddleware(gameMux)
			logToConsole := viper.GetBool("default.LogToConsole")
			var writer io.Writer
			if logToConsole {
				writer = os.Stdout
			} else {
				err := os.MkdirAll("logs", 0755)
				if err != nil {
					fmt.Println("Failed to create logs directory")
					_ = lock.Unlock()
					os.Exit(internal.ErrCreateLogsDir)
				}
				t := time.Now()
				fileName := fmt.Sprintf("logs/access_log_%d-%02d-%02dT%02d-%02d-%02d.txt", t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second())
				file, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY, 0644)
				if err != nil {
					fmt.Println("Failed to create log file")
					_ = lock.Unlock()
					os.Exit(internal.ErrCreateLogFile)
				}
				writer = file
			}
			certificatePairFolder := common.CertificatePairFolder(os.Args[0])
			if certificatePairFolder == "" {
				fmt.Println("Failed to determine certificate pair folder")
				_ = lock.Unlock()
				os.Exit(internal.ErrCertDirectory)
			}
			stop := make(chan os.Signal, 1)
			signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

			handler := handlers.LoggingHandler(writer, sessionMux)
			certFile := filepath.Join(certificatePairFolder, common.Cert)
			keyFile := filepath.Join(certificatePairFolder, common.Key)
			var servers []*http.Server
			customLogger := log.New(&internal.CustomWriter{OriginalWriter: os.Stderr}, "", log.LstdFlags)
			var multicastIP net.IP
			multicast := viper.GetBool("Announcement.Multicast")
			if multicast {
				multicastIP = net.ParseIP(viper.GetString("Announcement.MulticastGroup"))
				if multicastIP == nil || multicastIP.To4() == nil || !multicastIP.IsMulticast() {
					fmt.Println("Invalid multicast IP")
					_ = lock.Unlock()
					os.Exit(internal.ErrMulticastGroup)
				}
			}
			broadcast := viper.GetBool("Announcement.Broadcast")
			announcePort := viper.GetInt("Announcement.Port")
			if broadcast || multicast {
				fmt.Println("Announcing on port", announcePort)
			}
			for _, addr := range addrs {
				server := &http.Server{
					Addr:     addr.String() + ":443",
					Handler:  handler,
					ErrorLog: customLogger,
				}

				fmt.Println("Listening on " + server.Addr)
				go func() {
					if broadcast || multicast {
						go func() {
							ip.Announce(addr, multicastIP, announcePort, broadcast, multicast)
						}()
					}
					err := server.ListenAndServeTLS(certFile, keyFile)
					if err != nil && !errors.Is(err, http.ErrServerClosed) {
						fmt.Println("Failed to start server")
						fmt.Println(err)
						os.Exit(internal.ErrStartServer)
					}
				}()
				servers = append(servers, server)
			}

			<-stop

			fmt.Println("Servers are shutting down...")

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			for _, server := range servers {
				if err := server.Shutdown(ctx); err != nil {
					fmt.Printf("Server %s forced to shutdown: %v\n", server.Addr, err)
				}

				fmt.Println("Server", server.Addr, "stopped")
			}

			_ = lock.Unlock()
		},
	}
)

func Execute() error {
	cobra.OnInitialize(initConfig)
	rootCmd.Version = Version
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", fmt.Sprintf(`config file (default config.toml in %s directories)`, strings.Join(configPaths, ", ")))
	rootCmd.PersistentFlags().BoolP("announce", "a", true, "Announce server in LAN. Disabling this will not allow launchers to discover it and will require specifying the host")
	rootCmd.PersistentFlags().IntP("announcePort", "p", common.AnnouncePort, "Port to announce to. If changed, the launchers will need to specify the port in Server.AnnouncePorts")
	rootCmd.PersistentFlags().BoolP("announceMulticast", "m", true, "Whether to announce the server using Multicast.")
	rootCmd.PersistentFlags().BoolP("announceBroadcast", "b", false, "Whether to announce the server using Broadcast.")
	rootCmd.PersistentFlags().StringP("announceMulticastGroup", "i", "239.31.97.8", "Whether to announce the server using Multicast or Broadcast.")
	rootCmd.PersistentFlags().StringArrayP("games", "e", []string{common.GameAoE2}, fmt.Sprintf(`Games that the server will accept. Currently, only "%s" is supported.`, common.GameAoE2))
	rootCmd.PersistentFlags().StringArrayP("host", "n", []string{netip.IPv4Unspecified().String()}, "The host the server will bind to. Can be set multiple times.")
	rootCmd.PersistentFlags().BoolP("logToConsole", "l", false, "Log the requests to the console (stdout) or not.")
	rootCmd.PersistentFlags().BoolP("generatePlatformUserId", "g", false, "Generate the Platform User Id based on the user's IP.")
	if err := viper.BindPFlag("Announcement.Enabled", rootCmd.PersistentFlags().Lookup("announce")); err != nil {
		return err
	}
	if err := viper.BindPFlag("Announcement.Port", rootCmd.PersistentFlags().Lookup("announcePort")); err != nil {
		return err
	}
	if err := viper.BindPFlag("Announcement.Broadcast", rootCmd.PersistentFlags().Lookup("announceBroadcast")); err != nil {
		return err
	}
	if err := viper.BindPFlag("Announcement.Multicast", rootCmd.PersistentFlags().Lookup("announceMulticast")); err != nil {
		return err
	}
	if err := viper.BindPFlag("Announcement.MulticastGroup", rootCmd.PersistentFlags().Lookup("announceMulticastGroup")); err != nil {
		return err
	}
	if err := viper.BindPFlag("default.Hosts", rootCmd.PersistentFlags().Lookup("host")); err != nil {
		return err
	}
	if err := viper.BindPFlag("default.Games", rootCmd.PersistentFlags().Lookup("games")); err != nil {
		return err
	}
	if err := viper.BindPFlag("default.LogToConsole", rootCmd.PersistentFlags().Lookup("logToConsole")); err != nil {
		return err
	}
	if err := viper.BindPFlag("default.GeneratePlatformUserId", rootCmd.PersistentFlags().Lookup("generatePlatformUserId")); err != nil {
		return err
	}
	return rootCmd.Execute()
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		for _, configPath := range configPaths {
			viper.AddConfigPath(configPath)
		}
		viper.SetConfigType("toml")
		viper.SetConfigName("config")
	}
	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
