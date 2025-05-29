package alertmanager

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/antnsn/mal-sync/internal/common" // Adjusted import path
)

const (
	mimirtoolCmd = "mimirtool" // Assuming mimirtool is in PATH
)

// Sync performs the Alertmanager synchronization.
func Sync(configFile, templatesDir, mimirAddress, mimirID, tempBaseDir string) error {
	log.Printf("Starting Alertmanager sync for Mimir instance: %s (ID: %s)", mimirAddress, mimirID)
	log.Printf("Config file: %s", configFile)
	if templatesDir != "" {
		log.Printf("Templates directory: %s", templatesDir)
	}

	// 1. Prepare temporary directory for this sync operation
	syncTempDir := filepath.Join(tempBaseDir, fmt.Sprintf("mal-sync-alertmanager-%d", os.Getpid()))
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

	// 2. Copy main config file to temporary location (snapshot)
	tempConfigFile := filepath.Join(syncTempDir, "alertmanager-config.yml")
	log.Printf("Copying main config file %s to %s", configFile, tempConfigFile)
	if err := common.CopyFile(configFile, tempConfigFile); err != nil {
		return fmt.Errorf("failed to copy config file %s to %s: %w", configFile, tempConfigFile, err)
	}

	// 3. Verify the temporary config file
	log.Printf("Verifying Alertmanager config: %s", tempConfigFile)
	verifyArgs := []string{"alertmanager", "verify", tempConfigFile}
	if output, err := common.ExecuteCommand(mimirtoolCmd, verifyArgs...); err != nil {
		return fmt.Errorf("Alertmanager config verification failed for %s: %w\nOutput:\n%s", tempConfigFile, err, output)
	}
	log.Println("Alertmanager config verified successfully.")

	// 4. Handle templates
	var templateFileArgs []string
	if templatesDir != "" {
		tempTemplatesDir := filepath.Join(syncTempDir, "templates")
		if err := common.EnsureDir(tempTemplatesDir); err != nil {
			return fmt.Errorf("failed to create temporary templates directory %s: %w", tempTemplatesDir, err)
		}
		log.Printf("Processing templates from %s, copying to %s", templatesDir, tempTemplatesDir)

		entries, err := os.ReadDir(templatesDir)
		if err != nil {
			log.Printf("Warning: could not read templates directory %s: %v. Proceeding without templates.", templatesDir, err)
		} else {
			foundTemplates := false
			for _, entry := range entries {
				if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".tmpl") {
					srcPath := filepath.Join(templatesDir, entry.Name())
					dstPath := filepath.Join(tempTemplatesDir, entry.Name())
					log.Printf("Copying template %s to %s", srcPath, dstPath)
					if err := common.CopyFile(srcPath, dstPath); err != nil {
						return fmt.Errorf("failed to copy template file %s to %s: %w", srcPath, dstPath, err)
					}
					templateFileArgs = append(templateFileArgs, dstPath)
					foundTemplates = true
				}
			}
			if !foundTemplates {
				log.Printf("No .tmpl files found in %s", templatesDir)
			} else {
				log.Printf("Copied %d template(s) to %s", len(templateFileArgs), tempTemplatesDir)
			}
		}
	}

	// 5. Load the Alertmanager configuration and templates into Mimir
	log.Println("Loading Alertmanager config and templates into Mimir...")
	loadArgs := []string{
		"alertmanager",
		"load",
		tempConfigFile, // The copied main config file
	}
	loadArgs = append(loadArgs, templateFileArgs...) // Add copied template files
	loadArgs = append(loadArgs, "--address="+mimirAddress, "--id="+mimirID)

	if output, err := common.ExecuteCommand(mimirtoolCmd, loadArgs...); err != nil {
		return fmt.Errorf("failed to load Alertmanager config to Mimir: %w\nOutput:\n%s", err, output)
	}

	log.Println("Alertmanager configuration and templates loaded successfully.")
	return nil
}
