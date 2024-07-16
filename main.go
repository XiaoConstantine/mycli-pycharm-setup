package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
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
	ideaDir := filepath.Join(projectDir, ".idea")
	err := os.MkdirAll(ideaDir, 0755)
	if err != nil {
		return err
	}

	miscXML := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<project version="4">
  <component name="ProjectRootManager" version="2" project-jdk-name="Python 3 (%s)" project-jdk-type="Python SDK" />
</project>`, projectName)

	err = os.WriteFile(filepath.Join(ideaDir, "misc.xml"), []byte(miscXML), 0644)
	if err != nil {
		return err
	}

	imlXML := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<module type="PYTHON_MODULE" version="4">
  <component name="NewModuleRootManager">
    <content url="file://$MODULE_DIR$">
      <excludeFolder url="file://$MODULE_DIR$/%s" />
    </content>
    <orderEntry type="jdk" jdkName="Python 3 (%s)" jdkType="Python SDK" />
    <orderEntry type="sourceFolder" forTests="false" />
  </component>
</module>`, filepath.Base(venvPath), filepath.Base(venvPath))

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
