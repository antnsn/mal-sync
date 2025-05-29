package lokirules

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/antnsn/mal-sync/internal/common"
)

const (
	lokitoolCmd = "lokitool"
)

// Sync performs the Loki rules synchronization.
func Sync(rulesPath, lokiAddress, orgID, tempBaseDir string) error {
	log.Printf("Starting Loki rules sync for Loki instance: %s (OrgID: %s)", lokiAddress, orgID)
	log.Printf("Rules path: %s", rulesPath)

	// 1. Prepare temporary directory for this sync operation
	syncTempDir := filepath.Join(tempBaseDir, fmt.Sprintf("mal-sync-lokirules-%d", os.Getpid()))
	if err := common.EnsureDir(syncTempDir); err != nil {
		return fmt.Errorf("failed to create temporary sync directory %s: %w", syncTempDir, err)
	}
	defer func() {
		log.Printf("Cleaning up temporary directory: %s", syncTempDir)
		if err := os.RemoveAll(syncTempDir); err != nil {
			log.Printf("Warning: failed to clean up temporary directory %s: %v", syncTempDir, err)
		}
	}()
	log.Printf("Using temporary directory: %s", syncTempDir)

	// 2. Collect and copy rule files to temporary location
	var tempRuleFiles []string
	fileInfo, err := os.Stat(rulesPath)
	if err != nil {
		return fmt.Errorf("failed to stat rules path %s: %w", rulesPath, err)
	}

	if fileInfo.IsDir() {
		log.Printf("Processing rules from directory: %s", rulesPath)
		entries, err := os.ReadDir(rulesPath)
		if err != nil {
			return fmt.Errorf("failed to read rules directory %s: %w", rulesPath, err)
		}
		for _, entry := range entries {
			if !entry.IsDir() && (strings.HasSuffix(entry.Name(), ".yaml") || strings.HasSuffix(entry.Name(), ".yml")) {
				srcPath := filepath.Join(rulesPath, entry.Name())
				dstPath := filepath.Join(syncTempDir, entry.Name())
				log.Printf("Copying rule file %s to %s", srcPath, dstPath)
				if err := common.CopyFile(srcPath, dstPath); err != nil {
					return fmt.Errorf("failed to copy rule file %s to %s: %w", srcPath, dstPath, err)
				}
				tempRuleFiles = append(tempRuleFiles, dstPath)
			}
		}
		if len(tempRuleFiles) == 0 {
			log.Printf("No .yaml or .yml files found in directory %s. Nothing to sync.", rulesPath)
			return nil // Not an error, just nothing to do
		}
	} else {
		// Single file case
		if !(strings.HasSuffix(rulesPath, ".yaml") || strings.HasSuffix(rulesPath, ".yml")) {
			return fmt.Errorf("rules.path points to a file but it is not a .yaml or .yml file: %s", rulesPath)
		}
		dstPath := filepath.Join(syncTempDir, filepath.Base(rulesPath))
		log.Printf("Copying single rule file %s to %s", rulesPath, dstPath)
		if err := common.CopyFile(rulesPath, dstPath); err != nil {
			return fmt.Errorf("failed to copy rule file %s to %s: %w", rulesPath, dstPath, err)
		}
		tempRuleFiles = append(tempRuleFiles, dstPath)
	}

	log.Printf("Copied %d rule file(s) to %s", len(tempRuleFiles), syncTempDir)

	// 3. Lint each rule file before attempting to sync
	log.Println("Linting Loki rule files...")
	for _, ruleFile := range tempRuleFiles {
		log.Printf("Linting rule file: %s", ruleFile)
		lintArgs := []string{
			"rules",
			"lint",
			ruleFile,
			// lokitool lint does not require --address or --org-id for file linting
		}
		if output, err := common.ExecuteCommand(lokitoolCmd, lintArgs...); err != nil {
			// lokitool lint exits with non-zero on lint errors
			log.Printf("Linting failed for %s:\n%s", ruleFile, output) // Log output which contains lint errors
			return fmt.Errorf("linting failed for rule file %s: %w", ruleFile, err)
		}
		log.Printf("Linting successful for %s", ruleFile)
	}

	// 4. Sync the Loki rules with Loki using --rule-dirs
	log.Println("Syncing Loki rules with Loki...")
	syncArgs := []string{
		"rules",
		"sync",
		"--address=" + lokiAddress,
		"--org-id=" + orgID,
		"--rule-dirs=" + syncTempDir, // Point to the temporary directory containing all rule files
	}

	if len(tempRuleFiles) == 0 {
		log.Println("No rule files to sync. If --rule-dirs is empty, lokitool might remove all rules for the given org-id. Proceeding with empty rule set.")
		// Potentially, one might want to confirm this action or have a flag to prevent accidental deletion of all rules.
		// For now, we proceed as `lokitool rules sync` would with an empty directory.
	}

	if output, err := common.ExecuteCommand(lokitoolCmd, syncArgs...); err != nil {
		return fmt.Errorf("failed to sync Loki rules with Loki: %w\nOutput:\n%s", err, output)
	}

	log.Println("Loki rules synced successfully.")
	return nil
}
