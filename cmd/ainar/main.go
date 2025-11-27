// AINAR - AI Model Router
// Provides a unified OpenAI-compatible interface for multiple AI providers
package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/HeavySnowJakarta/ainar/internal/api"
	"github.com/HeavySnowJakarta/ainar/internal/config"
	"github.com/HeavySnowJakarta/ainar/internal/webui"
	"github.com/spf13/cobra"
)

var (
	// Version information (set during build)
	Version   = "dev"
	GitCommit = "unknown"
	BuildDate = "unknown"

	// Global flags
	configPath string
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "ainar",
		Short: "AINAR - AI Model Router",
		Long: `AINAR (Ainar Is Not an AI Router) is a unified OpenAI-compatible 
API gateway that routes requests to multiple AI providers.`,
		Run: func(cmd *cobra.Command, args []string) {
			runServer()
		},
	}

	// Global flags
	rootCmd.PersistentFlags().StringVarP(&configPath, "config", "c", "", "Path to configuration file")

	// Add subcommands
	rootCmd.AddCommand(versionCmd())
	rootCmd.AddCommand(serveCmd())
	rootCmd.AddCommand(providerCmd())
	rootCmd.AddCommand(keyCmd())
	rootCmd.AddCommand(aliasCmd())
	rootCmd.AddCommand(initCmd())

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func versionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("AINAR version %s\n", Version)
			fmt.Printf("Git commit: %s\n", GitCommit)
			fmt.Printf("Build date: %s\n", BuildDate)
		},
	}
}

func serveCmd() *cobra.Command {
	var apiPort, webuiPort int
	var host string

	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Start the AINAR server",
		Run: func(cmd *cobra.Command, args []string) {
			runServerWithFlags(host, apiPort, webuiPort)
		},
	}

	cmd.Flags().StringVar(&host, "host", "", "Host to bind to (overrides config)")
	cmd.Flags().IntVar(&apiPort, "port", 0, "API port (overrides config)")
	cmd.Flags().IntVar(&webuiPort, "webui-port", 0, "WebUI port (overrides config)")

	return cmd
}

func providerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "provider",
		Short: "Manage upstream providers",
	}

	// List providers
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List all providers",
		Run: func(cmd *cobra.Command, args []string) {
			cfg := loadConfig()
			if len(cfg.Providers) == 0 {
				fmt.Println("No providers configured")
				return
			}
			for _, p := range cfg.Providers {
				status := "enabled"
				if !p.Enabled {
					status = "disabled"
				}
				fmt.Printf("%-15s %-20s %-10s %s [%s]\n", p.Code, p.Name, p.APIType, p.BaseURL, status)
			}
		},
	}

	// Add provider
	addCmd := &cobra.Command{
		Use:   "add",
		Short: "Add a new provider",
		Run: func(cmd *cobra.Command, args []string) {
			name, _ := cmd.Flags().GetString("name")
			code, _ := cmd.Flags().GetString("code")
			apiType, _ := cmd.Flags().GetString("type")
			baseURL, _ := cmd.Flags().GetString("url")
			apiKey, _ := cmd.Flags().GetString("key")

			cfg := loadConfig()
			p := config.Provider{
				Name:    name,
				Code:    code,
				APIType: apiType,
				BaseURL: baseURL,
				APIKey:  apiKey,
				Enabled: true,
			}

			if err := cfg.AddProvider(p); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}

			if err := cfg.Save(getConfigPath()); err != nil {
				fmt.Fprintf(os.Stderr, "Error saving config: %v\n", err)
				os.Exit(1)
			}

			fmt.Printf("Provider %s added successfully\n", code)
		},
	}
	addCmd.Flags().StringP("name", "n", "", "Provider name")
	addCmd.Flags().StringP("code", "c", "", "Provider code")
	addCmd.Flags().StringP("type", "t", "openai", "API type (openai, anthropic, google, ollama)")
	addCmd.Flags().StringP("url", "u", "", "Base URL")
	addCmd.Flags().StringP("key", "k", "", "API key")
	addCmd.MarkFlagRequired("name")
	addCmd.MarkFlagRequired("code")
	addCmd.MarkFlagRequired("url")

	// Remove provider
	removeCmd := &cobra.Command{
		Use:   "remove [code]",
		Short: "Remove a provider",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			cfg := loadConfig()
			if err := cfg.RemoveProvider(args[0]); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}

			if err := cfg.Save(getConfigPath()); err != nil {
				fmt.Fprintf(os.Stderr, "Error saving config: %v\n", err)
				os.Exit(1)
			}

			fmt.Printf("Provider %s removed successfully\n", args[0])
		},
	}

	cmd.AddCommand(listCmd, addCmd, removeCmd)
	return cmd
}

func keyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "key",
		Short: "Manage downstream API keys",
	}

	// List keys
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List all API keys",
		Run: func(cmd *cobra.Command, args []string) {
			cfg := loadConfig()
			if len(cfg.DownstreamKeys) == 0 {
				fmt.Println("No API keys configured")
				return
			}
			for _, k := range cfg.DownstreamKeys {
				status := "enabled"
				if !k.Enabled {
					status = "disabled"
				}
				limitInfo := "no limit"
				if k.Limit != nil {
					limitInfo = fmt.Sprintf("%s: %d tokens", k.Limit.Type, k.Limit.MaxTokens)
				}
				usageInfo := ""
				if k.Usage != nil && k.Limit != nil {
					percentage := float64(k.Usage.TokensUsed) / float64(k.Limit.MaxTokens) * 100
					usageInfo = fmt.Sprintf(" (%.1f%% used)", percentage)
				}
				fmt.Printf("%-20s [%s] %s%s\n", k.Name, status, limitInfo, usageInfo)
			}
		},
	}

	// Generate key
	generateCmd := &cobra.Command{
		Use:   "generate [name]",
		Short: "Generate a new API key",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			limitType, _ := cmd.Flags().GetString("limit-type")
			limitTokens, _ := cmd.Flags().GetInt64("limit-tokens")

			cfg := loadConfig()

			var limit *config.TokenLimit
			if limitType != "" && limitTokens > 0 {
				limit = &config.TokenLimit{
					Type:      limitType,
					MaxTokens: limitTokens,
				}
			}

			key, err := cfg.GenerateDownstreamKey(args[0], limit)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}

			if err := cfg.Save(getConfigPath()); err != nil {
				fmt.Fprintf(os.Stderr, "Error saving config: %v\n", err)
				os.Exit(1)
			}

			fmt.Printf("API key generated for %s:\n", args[0])
			fmt.Printf("\n  %s\n\n", key)
			fmt.Println("⚠️  Save this key now. It will only be shown once!")
		},
	}
	generateCmd.Flags().String("limit-type", "", "Limit type (daily, weekly, monthly, total)")
	generateCmd.Flags().Int64("limit-tokens", 0, "Maximum tokens")

	// Delete key
	deleteCmd := &cobra.Command{
		Use:   "delete [name]",
		Short: "Delete an API key",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			cfg := loadConfig()
			found := false
			for i := range cfg.DownstreamKeys {
				if cfg.DownstreamKeys[i].Name == args[0] {
					cfg.DownstreamKeys = append(cfg.DownstreamKeys[:i], cfg.DownstreamKeys[i+1:]...)
					found = true
					break
				}
			}

			if !found {
				fmt.Fprintf(os.Stderr, "Key %s not found\n", args[0])
				os.Exit(1)
			}

			if err := cfg.Save(getConfigPath()); err != nil {
				fmt.Fprintf(os.Stderr, "Error saving config: %v\n", err)
				os.Exit(1)
			}

			fmt.Printf("API key %s deleted\n", args[0])
		},
	}

	cmd.AddCommand(listCmd, generateCmd, deleteCmd)
	return cmd
}

func aliasCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "alias",
		Short: "Manage model aliases",
	}

	// List aliases
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List all aliases",
		Run: func(cmd *cobra.Command, args []string) {
			cfg := loadConfig()
			if len(cfg.Aliases) == 0 {
				fmt.Println("No aliases configured")
				return
			}
			for _, a := range cfg.Aliases {
				fmt.Printf("%s -> %s\n", a.From, a.To)
			}
		},
	}

	// Add alias
	addCmd := &cobra.Command{
		Use:   "add [from] [to]",
		Short: "Add a new alias",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			cfg := loadConfig()
			if err := cfg.AddAlias(args[0], args[1]); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}

			if err := cfg.Save(getConfigPath()); err != nil {
				fmt.Fprintf(os.Stderr, "Error saving config: %v\n", err)
				os.Exit(1)
			}

			fmt.Printf("Alias added: %s -> %s\n", args[0], args[1])
		},
	}

	// Remove alias
	removeCmd := &cobra.Command{
		Use:   "remove [from]",
		Short: "Remove an alias",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			cfg := loadConfig()
			found := false
			for i := range cfg.Aliases {
				if cfg.Aliases[i].From == args[0] {
					cfg.Aliases = append(cfg.Aliases[:i], cfg.Aliases[i+1:]...)
					found = true
					break
				}
			}

			if !found {
				fmt.Fprintf(os.Stderr, "Alias %s not found\n", args[0])
				os.Exit(1)
			}

			if err := cfg.Save(getConfigPath()); err != nil {
				fmt.Fprintf(os.Stderr, "Error saving config: %v\n", err)
				os.Exit(1)
			}

			fmt.Printf("Alias %s removed\n", args[0])
		},
	}

	cmd.AddCommand(listCmd, addCmd, removeCmd)
	return cmd
}

func initCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Initialize a new configuration file",
		Run: func(cmd *cobra.Command, args []string) {
			path := getConfigPath()
			if _, err := os.Stat(path); err == nil {
				fmt.Printf("Configuration file already exists at %s\n", path)
				os.Exit(1)
			}

			cfg := config.DefaultConfig()
			if err := cfg.Save(path); err != nil {
				fmt.Fprintf(os.Stderr, "Error creating config: %v\n", err)
				os.Exit(1)
			}

			fmt.Printf("Configuration file created at %s\n", path)
		},
	}
}

func runServer() {
	runServerWithFlags("", 0, 0)
}

func runServerWithFlags(host string, apiPort, webuiPort int) {
	cfg := loadConfig()
	cfgPath := getConfigPath()

	// Override with command line flags
	if host != "" {
		cfg.Server.Host = host
	}
	if apiPort != 0 {
		cfg.Server.Port = apiPort
	}
	if webuiPort != 0 {
		cfg.WebUI.Port = webuiPort
	}

	// Create API router
	router := api.NewRouter(cfg, cfgPath)

	// Start API server
	apiAddr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	apiServer := &http.Server{
		Addr:    apiAddr,
		Handler: router,
	}

	go func() {
		fmt.Printf("API server listening on %s\n", apiAddr)
		if err := apiServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Fprintf(os.Stderr, "API server error: %v\n", err)
		}
	}()

	// Start WebUI server if not disabled
	var webuiServer *http.Server
	if cfg.WebUI.Status != "disabled" {
		localOnly := cfg.WebUI.Status == "local"
		if cfg.WebUI.Status == "external" {
			fmt.Println("⚠️  WARNING: WebUI is set to external mode. Anyone can access sensitive data!")
		}

		webuiHandler := webui.NewHandler(cfg, cfgPath, localOnly)
		webuiAddr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.WebUI.Port)
		webuiServer = &http.Server{
			Addr:    webuiAddr,
			Handler: webuiHandler,
		}

		go func() {
			fmt.Printf("WebUI server listening on %s\n", webuiAddr)
			if err := webuiServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				fmt.Fprintf(os.Stderr, "WebUI server error: %v\n", err)
			}
		}()
	}

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	fmt.Println("\nShutting down...")
	apiServer.Close()
	if webuiServer != nil {
		webuiServer.Close()
	}
}

func loadConfig() *config.Config {
	path := getConfigPath()
	cfg, err := config.Load(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}
	return cfg
}

func getConfigPath() string {
	if configPath != "" {
		return configPath
	}
	return config.GetConfigPath()
}
