package cmd

import (
	"context"
	"errors"
	"fmt"
	"github.com/gorilla/handlers"
	"github.com/luskaner/aoe2DELanServer/common"
	"github.com/luskaner/aoe2DELanServer/common/pidLock"
	"github.com/luskaner/aoe2DELanServer/server/internal"
	"github.com/luskaner/aoe2DELanServer/server/internal/files"
	"github.com/luskaner/aoe2DELanServer/server/internal/ip"
	"github.com/luskaner/aoe2DELanServer/server/internal/middleware"
	"github.com/luskaner/aoe2DELanServer/server/internal/routes"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"io"
	"net/http"
	"net/netip"
	"os"
	"os/signal"
	"path"
	"path/filepath"
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
				os.Exit(common.ErrPidLock)
			}
			host := viper.GetString("default.Host")
			addr := ip.ResolveHost(host)
			if addr == nil {
				fmt.Println("Failed to resolve host")
				_ = lock.Unlock()
				os.Exit(internal.ErrResolveHost)
			}
			mux := http.NewServeMux()
			files.Initialize()
			routes.Initialize(mux)
			sessionMux := middleware.SessionMiddleware(mux)

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

			server := &http.Server{
				Addr:    addr.String() + ":443",
				Handler: handlers.LoggingHandler(writer, sessionMux),
			}

			if viper.GetBool("Announcement.Enabled") {
				fmt.Println("Trying to announce server in", addr, "network to UDP port", viper.GetInt("Announcement.Port"))
				go func() {
					ip.Announce(addr, viper.GetInt("Announcement.Port"))
				}()
			}

			certificatePairFolder := common.CertificatePairFolder(os.Args[0])
			if certificatePairFolder == "" {
				fmt.Println("Failed to determine certificate pair folder")
				_ = lock.Unlock()
				os.Exit(internal.ErrCertDirectory)
			}
			stop := make(chan os.Signal, 1)
			signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
			fmt.Println("Listening on " + server.Addr)
			go func() {
				err := server.ListenAndServeTLS(filepath.Join(certificatePairFolder, common.Cert), filepath.Join(certificatePairFolder, common.Key))
				if err != nil && !errors.Is(err, http.ErrServerClosed) {
					fmt.Println("Failed to start server")
					fmt.Println(err)
					os.Exit(internal.ErrStartServer)
				}
			}()
			<-stop

			fmt.Printf("Server is shutting down...")

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			if err := server.Shutdown(ctx); err != nil {
				fmt.Printf("Server forced to shutdown: %v\n", err)
			}

			fmt.Println("Server stopped")

			_ = lock.Unlock()
		},
	}
)

func Execute() error {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", fmt.Sprintf(`config file (default config.ini in %s directories)`, strings.Join(configPaths, ", ")))
	rootCmd.PersistentFlags().BoolP("announce", "a", true, "Announce server in LAN. Disabling this will not allow launchers to discover it and will require specifying the host")
	rootCmd.PersistentFlags().IntP("announcePort", "p", common.AnnouncePort, "Port to announce to. If changed, the launchers will need to specify the port in Server.AnnouncePorts")
	rootCmd.PersistentFlags().StringP("host", "n", netip.IPv4Unspecified().String(), "The host the server will bind to.")
	rootCmd.PersistentFlags().BoolP("logToConsole", "l", false, "Log the requests to the console (stdout) or not.")
	rootCmd.PersistentFlags().BoolP("generatePlatformUserId", "g", false, "Generate the Platform User Id based on the user's IP.")
	rootCmd.PersistentFlags().StringVarP(&Version, "version", "v", Version, "Version")
	if err := viper.BindPFlag("Announcement.Enabled", rootCmd.PersistentFlags().Lookup("announce")); err != nil {
		return err
	}
	if err := viper.BindPFlag("Announcement.Port", rootCmd.PersistentFlags().Lookup("announcePort")); err != nil {
		return err
	}
	if err := viper.BindPFlag("default.Host", rootCmd.PersistentFlags().Lookup("host")); err != nil {
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
		viper.SetConfigType("ini")
		viper.SetConfigName("config")
	}
	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
