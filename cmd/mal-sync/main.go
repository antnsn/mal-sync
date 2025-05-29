package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/antnsn/mal-sync/internal/alertmanager"
	"github.com/antnsn/mal-sync/internal/mimirrules"
	"github.com/antnsn/mal-sync/internal/lokirules"
)

func main() {
	// Define command-line flags
	// For Alertmanager
	alertmanagerCmd := flag.NewFlagSet("alertmanager", flag.ExitOnError)
	// Note: flag.String returns a pointer. We'll dereference after parsing.
	_ = alertmanagerCmd.String("config.file", "", "Path to the Alertmanager configuration file (e.g., /config/alertmanager.yaml). Env: MALSYNC_ALERTMANAGER_CONFIG_FILE")
	_ = alertmanagerCmd.String("templates.dir", "", "Path to the directory containing Alertmanager template files (e.g., /etc/alertmanager/templates). Env: MALSYNC_ALERTMANAGER_TEMPLATES_DIR")
	_ = alertmanagerCmd.String("mimir.address", "", "Address of the Mimir instance (e.g., http://mimir-nginx.mimir.svc.cluster.local:80). Env: MALSYNC_ALERTMANAGER_MIMIR_ADDRESS")
	_ = alertmanagerCmd.String("mimir.id", "anonymous", "Mimir tenant ID. Env: MALSYNC_ALERTMANAGER_MIMIR_ID")
	_ = alertmanagerCmd.String("temp.dir", "/tmp", "Temporary directory for staging files. Env: MALSYNC_ALERTMANAGER_TEMP_DIR")

	// For Mimir Rules
	mimirRulesCmd := flag.NewFlagSet("mimir-rules", flag.ExitOnError)
	_ = mimirRulesCmd.String("rules.path", "", "Path to a directory containing Mimir rule files (*.yaml, *.yml) or a single rule file. Env: MALSYNC_MIMIRRULES_RULES_PATH")
	_ = mimirRulesCmd.String("mimir.address", "", "Address of the Mimir instance. Env: MALSYNC_MIMIRRULES_MIMIR_ADDRESS")
	_ = mimirRulesCmd.String("mimir.id", "anonymous", "Mimir tenant ID. Env: MALSYNC_MIMIRRULES_MIMIR_ID")
	_ = mimirRulesCmd.String("temp.dir", "/tmp", "Temporary directory for staging files. Env: MALSYNC_MIMIRRULES_TEMP_DIR")
	_ = mimirRulesCmd.String("rules.namespace", "", "Mimir namespace to load the rules into. Env: MALSYNC_MIMIRRULES_RULES_NAMESPACE")

	// For Loki Rules
	lokiRulesCmd := flag.NewFlagSet("loki-rules", flag.ExitOnError)
	_ = lokiRulesCmd.String("rules.path", "", "Path to a directory containing Loki rule files (*.yaml, *.yml) or a single rule file. Env: MALSYNC_LOKIRULES_RULES_PATH")
	_ = lokiRulesCmd.String("loki.address", "", "Address of the Loki instance (e.g., http://loki.loki.svc.cluster.local:3100). Env: MALSYNC_LOKIRULES_LOKI_ADDRESS")
	_ = lokiRulesCmd.String("loki.org-id", "fake", "Loki Organization ID. Env: MALSYNC_LOKIRULES_LOKI_ORG_ID") // Loki often uses 'fake' as a default/common org-id for single-tenant setups
	_ = lokiRulesCmd.String("temp.dir", "/tmp", "Temporary directory for staging files. Env: MALSYNC_LOKIRULES_TEMP_DIR")
	// Add Loki specific flags here ...

	if len(os.Args) < 2 {
		log.Println("Expected 'alertmanager' or 'loki' subcommands")
		fmt.Println("Usage: mal-sync <subcommand> [options]")
		fmt.Println("\nSubcommands:")
		fmt.Println("  alertmanager  Sync Alertmanager configurations")
		fmt.Println("  mimir-rules   Sync Mimir rule files")
		fmt.Println("  loki-rules    Sync Loki rule files") // For future
		fmt.Println("\nAlertmanager options:")
		alertmanagerCmd.PrintDefaults()
		fmt.Println("\nMimir Rules options:")
		mimirRulesCmd.PrintDefaults()
		fmt.Println("\nLoki Rules options:")
		lokiRulesCmd.PrintDefaults()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "alertmanager":
		alertmanagerCmd.Parse(os.Args[2:])
		// Helper to determine if a flag was set on the command line
		alertmanagerFlagsSet := make(map[string]bool)
		alertmanagerCmd.Visit(func(f *flag.Flag) { alertmanagerFlagsSet[f.Name] = true })

		getAMValue := func(flagName, envVarName string) string {
			val := alertmanagerCmd.Lookup(flagName).Value.String()
			defVal := alertmanagerCmd.Lookup(flagName).DefValue
			if alertmanagerFlagsSet[flagName] { // Flag was explicitly set
				return val
			}
			env := os.Getenv(envVarName)
			if env != "" {
				log.Printf("Using %s from environment variable %s: %s", flagName, envVarName, env)
				return env
			}
			return defVal // or val, which is defVal if not set and no env
		}

		configFileVal := getAMValue("config.file", "MALSYNC_ALERTMANAGER_CONFIG_FILE")
		templatesDirVal := getAMValue("templates.dir", "MALSYNC_ALERTMANAGER_TEMPLATES_DIR")
		mimirAddressValAM := getAMValue("mimir.address", "MALSYNC_ALERTMANAGER_MIMIR_ADDRESS")
		mimirIDValAM := getAMValue("mimir.id", "MALSYNC_ALERTMANAGER_MIMIR_ID")
		tempDirValAM := getAMValue("temp.dir", "MALSYNC_ALERTMANAGER_TEMP_DIR")

		if configFileVal == "" {
			log.Fatal("Error: -config.file flag or MALSYNC_ALERTMANAGER_CONFIG_FILE env var is required for alertmanager sync")
		}
		if mimirAddressValAM == "" {
			log.Fatal("Error: -mimir.address flag or MALSYNC_ALERTMANAGER_MIMIR_ADDRESS env var is required for alertmanager sync")
		}

		err := alertmanager.Sync(configFileVal, templatesDirVal, mimirAddressValAM, mimirIDValAM, tempDirValAM)
		if err != nil {
			log.Fatalf("Alertmanager sync failed: %v", err)
		}
		log.Println("Alertmanager sync completed successfully.")
	case "mimir-rules":
		mimirRulesCmd.Parse(os.Args[2:])
		// Helper to determine if a flag was set on the command line
		mimirRulesFlagsSet := make(map[string]bool)
		mimirRulesCmd.Visit(func(f *flag.Flag) { mimirRulesFlagsSet[f.Name] = true })

		getMRValue := func(flagName, envVarName string) string {
			val := mimirRulesCmd.Lookup(flagName).Value.String()
			defVal := mimirRulesCmd.Lookup(flagName).DefValue
			if mimirRulesFlagsSet[flagName] { // Flag was explicitly set
				return val
			}
			env := os.Getenv(envVarName)
			if env != "" {
				log.Printf("Using %s from environment variable %s: %s", flagName, envVarName, env)
				return env
			}
			return defVal // or val, which is defVal if not set and no env
		}

		rulesPathValMR := getMRValue("rules.path", "MALSYNC_MIMIRRULES_RULES_PATH")
		mimirAddressValMR := getMRValue("mimir.address", "MALSYNC_MIMIRRULES_MIMIR_ADDRESS")
		mimirIDValMR := getMRValue("mimir.id", "MALSYNC_MIMIRRULES_MIMIR_ID")
		tempDirValMR := getMRValue("temp.dir", "MALSYNC_MIMIRRULES_TEMP_DIR")
		namespaceValMR := getMRValue("rules.namespace", "MALSYNC_MIMIRRULES_RULES_NAMESPACE")

		if rulesPathValMR == "" {
			log.Fatal("Error: -rules.path flag or MALSYNC_MIMIRRULES_RULES_PATH env var is required for mimir-rules sync")
		}
		if mimirAddressValMR == "" {
			log.Fatal("Error: -mimir.address flag or MALSYNC_MIMIRRULES_MIMIR_ADDRESS env var is required for mimir-rules sync")
		}
		if namespaceValMR == "" {
			log.Fatal("Error: -rules.namespace flag or MALSYNC_MIMIRRULES_RULES_NAMESPACE env var is required for mimir-rules sync")
		}

		err := mimirrules.Sync(rulesPathValMR, mimirAddressValMR, mimirIDValMR, namespaceValMR, tempDirValMR)
		if err != nil {
			log.Fatalf("Mimir rules sync failed: %v", err)
		}
		log.Println("Mimir rules sync completed successfully.")
	case "loki-rules":
		lokiRulesCmd.Parse(os.Args[2:])
		// Helper to determine if a flag was set on the command line
		lokiRulesFlagsSet := make(map[string]bool)
		lokiRulesCmd.Visit(func(f *flag.Flag) { lokiRulesFlagsSet[f.Name] = true })

		getLRValue := func(flagName, envVarName string) string {
			val := lokiRulesCmd.Lookup(flagName).Value.String()
			defVal := lokiRulesCmd.Lookup(flagName).DefValue
			if lokiRulesFlagsSet[flagName] { // Flag was explicitly set
				return val
			}
			env := os.Getenv(envVarName)
			if env != "" {
				log.Printf("Using %s from environment variable %s: %s", flagName, envVarName, env)
				return env
			}
			return defVal
		}

		rulesPathValLR := getLRValue("rules.path", "MALSYNC_LOKIRULES_RULES_PATH")
		lokiAddressValLR := getLRValue("loki.address", "MALSYNC_LOKIRULES_LOKI_ADDRESS")
		lokiOrgIDValLR := getLRValue("loki.org-id", "MALSYNC_LOKIRULES_LOKI_ORG_ID")
		tempDirValLR := getLRValue("temp.dir", "MALSYNC_LOKIRULES_TEMP_DIR")

		if rulesPathValLR == "" {
			log.Fatal("Error: -rules.path flag or MALSYNC_LOKIRULES_RULES_PATH env var is required for loki-rules sync")
		}
		if lokiAddressValLR == "" {
			log.Fatal("Error: -loki.address flag or MALSYNC_LOKIRULES_LOKI_ADDRESS env var is required for loki-rules sync")
		}
		if lokiOrgIDValLR == "" { // Though it has a default, it's good practice to ensure it's explicitly handled if cleared
			log.Fatal("Error: -loki.org-id flag or MALSYNC_LOKIRULES_LOKI_ORG_ID env var is required for loki-rules sync")
		}

		err := lokirules.Sync(rulesPathValLR, lokiAddressValLR, lokiOrgIDValLR, tempDirValLR)
		if err != nil {
			log.Fatalf("Loki rules sync failed: %v", err)
		}
		log.Println("Loki rules sync completed successfully.")
	default:
		log.Fatalf("Unknown subcommand: %s. Expected 'alertmanager', 'mimir-rules', or 'loki-rules'.", os.Args[1])
	}
}

