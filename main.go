package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func main() {
	projectDir, err := os.Getwd()
	if err != nil {
		fmt.Println("Error: Unable to get current directory.")
		os.Exit(1)
	}

	if !isPythonProject(projectDir) {
		fmt.Println("Error: Not a Python project directory. Make sure you're in a directory with setup.py or pyproject.toml.")
		os.Exit(1)
	}

	venvPath, err := findVenv(projectDir)
	if err != nil {
		fmt.Println("Error: Virtual environment not found. Please create a virtual environment named 'venv' or '.venv' in your project directory.")
		os.Exit(1)
	}

	interpreterPath := filepath.Join(venvPath, "bin", "python")
	if _, err := os.Stat(interpreterPath); os.IsNotExist(err) {
		fmt.Println("Error: Python interpreter not found in the virtual environment.")
		os.Exit(1)
	}

	projectName := filepath.Base(projectDir)

	err = createPyCharmConfig(projectDir, projectName, venvPath)
	if err != nil {
		fmt.Printf("Error creating PyCharm configuration: %v\n", err)
		os.Exit(1)
	}

	err = openPyCharm(projectDir)
	if err != nil {
		fmt.Printf("Error opening PyCharm: %v\n", err)
		fmt.Println("You can open the project manually in PyCharm.")
	} else {
		fmt.Println("PyCharm project opened successfully.")
	}

	fmt.Printf("Project setup complete. Virtual environment: %s\n", venvPath)
}

func isPythonProject(dir string) bool {
	_, err1 := os.Stat(filepath.Join(dir, "setup.py"))
	_, err2 := os.Stat(filepath.Join(dir, "pyproject.toml"))
	return err1 == nil || err2 == nil
}

func findVenv(dir string) (string, error) {
	for dir != "/" {
		venvPath := filepath.Join(dir, "venv")
		if _, err := os.Stat(venvPath); err == nil {
			return venvPath, nil
		}
		venvPath = filepath.Join(dir, ".venv")
		if _, err := os.Stat(venvPath); err == nil {
			return venvPath, nil
		}
		dir = filepath.Dir(dir)
	}
	return "", fmt.Errorf("virtual environment not found")
}

func createPyCharmConfig(projectDir, projectName, venvPath string) error {
	// Get Python version
	pythonVersion, err := getPythonVersion(venvPath)
	if err != nil {
		return fmt.Errorf("failed to get Python version: %w", err)
	}

	// Create .idea directory if it doesn't exist
	ideaDir := filepath.Join(projectDir, ".idea")
	if err := os.MkdirAll(ideaDir, 0755); err != nil {
		return fmt.Errorf("failed to create .idea directory: %w", err)
	}

	// Use the project name for the interpreter name
	interpreterName := fmt.Sprintf("Python %s (%s)", pythonVersion, projectName)

	// Create misc.xml
	miscXML := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<project version="4">
  <component name="ProjectRootManager" version="2" project-jdk-name="%s" project-jdk-type="Python SDK" />
</project>`, interpreterName)

	if err := os.WriteFile(filepath.Join(ideaDir, "misc.xml"), []byte(miscXML), 0644); err != nil {
		return fmt.Errorf("failed to write misc.xml: %w", err)
	}

	// Create <project_name>.iml
	imlXML := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<module type="PYTHON_MODULE" version="4">
  <component name="NewModuleRootManager">
    <content url="file://$MODULE_DIR$">
      <excludeFolder url="file://$MODULE_DIR$/%s" />
    </content>
    <orderEntry type="jdk" jdkName="%s" jdkType="Python SDK" />
    <orderEntry type="sourceFolder" forTests="false" />
  </component>
</module>`, filepath.Base(venvPath), interpreterName)

	if err := os.WriteFile(filepath.Join(ideaDir, projectName+".iml"), []byte(imlXML), 0644); err != nil {
		return fmt.Errorf("failed to write %s.iml: %w", projectName, err)
	}

	return nil
}

func openPyCharm(projectDir string) error {
	cmd := exec.Command("pycharm", projectDir)
	cmd.Stderr = os.Stderr
	return cmd.Start()
}

func getPythonVersion(venvPath string) (string, error) {
	pythonPath := filepath.Join(venvPath, "bin", "python")
	cmd := exec.Command(pythonPath, "--version")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	// Output is typically in the format "Python X.Y.Z"
	version := strings.TrimSpace(string(output))
	parts := strings.Split(version, " ")
	if len(parts) != 2 {
		return "", fmt.Errorf("unexpected Python version format: %s", version)
	}
	// Return just the "X.Y" part of the version
	versionParts := strings.Split(parts[1], ".")
	if len(versionParts) < 2 {
		return "", fmt.Errorf("unexpected Python version format: %s", parts[1])
	}
	return fmt.Sprintf("%s.%s", versionParts[0], versionParts[1]), nil
}
