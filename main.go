package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
)

func main() {
	// Define a custom boolean flag to trigger installation and configuration
	installFlag := flag.Bool("i", false, "Install and configure the necessary components")
	ssmFlag := flag.String("ssm", "", "Connect to an SSM instance by name")
	flag.Parse()

	// If the -install flag is provided, perform installation and configuration
	if *installFlag {
		installAndConfigure()
		return
	}

	// Check if no flags are provided and print help
	if flag.NFlag() == 0 {
		printHelp()
		return
	}

	// If -ssm flag is provided with an environment, execute the SSM command
	if *ssmFlag != "" {
		environment := *ssmFlag
		connectToEnvironment(environment)
		return
	}

	if !checkAndInstall("aws", "AWS CLI", "awscli") {
		return
	}

	if !checkAndInstall("curl", "Curl", "curl") {
		return
	}

	if !isSSMPluginInstalled() {
		downloadAndInstallSSMPlugin()
	} else {
		fmt.Println("SSM Plugin already Installed")
	}

	// Everything is installed, so let's configure AWS.
	if !isAWSConfigured() {
		configureAWS()
	} else {
		fmt.Println("AWS CLI is already configured")
	}

}

func checkAndInstall(command, componentName, installationURL string) bool {
	cmd := exec.Command(command, "--version")

	// Capture and discard stdout and stderr
	cmd.Stdout = nil
	cmd.Stderr = nil

	err := cmd.Run()

	if err != nil {
		fmt.Printf("%s is not installed or not in the PATH. Attempting to install...\n", componentName)
		installCmd := exec.Command("sudo", "apt", "install", installationURL, "-y")

		// Capture and discard stdout and stderr
		installCmd.Stdout = nil
		installCmd.Stderr = nil

		if err := installCmd.Run(); err != nil {
			fmt.Printf("Failed to install %s: %v\n", componentName, err)
			return false
		} else {
			fmt.Printf("%s installed successfully.\n", componentName)
			return true
		}
	}

	fmt.Printf("%s is already installed.\n", componentName)
	return true
}

func downloadAndInstallSSMPlugin() bool {
	if _, err := os.Stat("session-manager-plugin.deb"); os.IsNotExist(err) {
		fmt.Println("Downloading AWS SSM Session Manager plugin...")
		downloadCmd := exec.Command("curl", "https://s3.amazonaws.com/session-manager-downloads/plugin/latest/ubuntu_64bit/session-manager-plugin.deb", "-o", "session-manager-plugin.deb")

		// Capture and discard stdout and stderr
		downloadCmd.Stdout = nil
		downloadCmd.Stderr = nil

		if err := downloadCmd.Run(); err != nil {
			fmt.Println("Failed to download AWS SSM Session Manager plugin:", err)
			return false
		}
		fmt.Println("AWS SSM Session Manager plugin downloaded successfully.")
	}

	fmt.Println("Installing AWS SSM Session Manager plugin...")
	installCmd := exec.Command("sudo", "dpkg", "-i", "session-manager-plugin.deb")

	// Capture and discard stdout and stderr
	installCmd.Stdout = nil
	installCmd.Stderr = nil

	if err := installCmd.Run(); err != nil {
		fmt.Printf("Failed to install AWS SSM Session Manager plugin: %v\n", err)
		return false
	}
	fmt.Println("AWS SSM Session Manager plugin installed successfully.")
	return true
}

func isSSMPluginInstalled() bool {
	cmd := exec.Command("dpkg", "-l", "session-manager-plugin")

	err := cmd.Run()

	if err == nil {
		return true // Exit code is zero, indicating the package is available
	}

	return false // Non-zero exit code means the package is not available
}

func configureAWS() {
	fmt.Println("Configuring AWS...")
	cmd := exec.Command("aws", "configure")

	cmd.Stdin = os.Stdin // Pass stdin to the command

	cmd.Stdout = os.Stdout // Pass stdout to the command

	cmd.Stderr = os.Stderr // Pass stderr to the command

	err := cmd.Run()

	if err != nil {
		fmt.Println("Failed to execute 'aws configure':", err)
	} else {
		fmt.Println("AWS configured successfully.")
	}
}

func isAWSConfigured() bool {
	usr, err := user.Current()
	if err != nil {
		fmt.Println("Failed to determine the user's home directory:", err)
		return false
	}

	configFilePath := filepath.Join(usr.HomeDir, ".aws", "config")

	_, err = os.Stat(configFilePath)
	return !os.IsNotExist(err)
}

func installAndConfigure() {
	// Include the installation and configuration logic here

	// After installation and configuration, continue with the rest of the program
	if !checkAndInstall("aws", "AWS CLI", "awscli") {
		return
	}

	if !checkAndInstall("curl", "Curl", "curl") {
		return
	}

	if !isSSMPluginInstalled() {
		downloadAndInstallSSMPlugin()
	} else {
		fmt.Println("SSM Plugin already Installed")
	}

	// Everything is installed, so let's configure AWS.
	if !isAWSConfigured() {
		configureAWS()
	} else {
		fmt.Println("AWS CLI is already configured")
	}
}

func printHelp() {
	fmt.Println("Usage:")
	fmt.Println("  To install and configure, run: your_program -i or --install")
	fmt.Println("  To connect to an SSM instance, run: your_program --ssm instance_name")
	fmt.Println("  To print this help message, run: your_program")
}

func connectToEnvironment(environment string) {
	// Define environment-to-instance-region mappings
	environmentToInstanceRegion := map[string]struct {
		InstanceID string
		Region     string
	}{
		"amp-af-stg": {
			InstanceID: "i-00c7da261367b0a31",
			Region:     "ap-southeast-1", // Example region for "dev"
		},
		"amp-af-prd": {
			InstanceID: "i-0b808b5c8b54ca924",
			Region:     "ap-southeast-2", // Example region for "prod"
		},
		"another": {
			InstanceID: "i-0123456789abcdef0", // Placeholder instance ID for "another"
			Region:     "us-east-1",           // Example region for "another"
		},
	}

	if envInfo, ok := environmentToInstanceRegion[environment]; ok {
		fmt.Printf("Executing SSM command for %s environment in region %s:\n", environment, envInfo.Region)
		executeSSMCommand(envInfo.InstanceID, envInfo.Region)
	} else {
		fmt.Println("Please specify a valid environment using the -ssm flag (dev, prod, or another).")
	}
}

func executeSSMCommand(instanceID, region string) {
	cmd := exec.Command("aws", "ssm", "start-session", "--target", instanceID,
		"--document-name", "AWS-StartPortForwardingSession",
		"--parameters", `{"portNumber":["8080"],"localPortNumber":["8080"]}`,
		"--region", region)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Printf("Error executing SSM command: %v\n", err)
	}
}
