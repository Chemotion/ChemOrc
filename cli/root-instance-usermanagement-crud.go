package cli

import (
	"fmt"
	"log"
	"os/exec"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var Reset = "\033[0m"
var Green = "\033[32m"
var Yellow = "\033[33m"
var Red = "\033[31m"
var Cyan = "\033[36m"

func init() {
	if runtime.GOOS == "windows" {
		Reset = ""
		Green = ""
		Yellow = ""
		Red = ""
	}
}

func setDefaultValue(someValue string, defaultValue string) string {

	if string(someValue) == "" {
		someValue = string(defaultValue)
	}
	fmt.Println(Cyan + "value entered:: " + Yellow + someValue + Reset + "\n")
	return someValue
}

// This function prompt user to type in different input parameters such as password, first name, last name, abbeviation.
func userDataInput() ([]byte, string, string, string) {

	var fname, lname, abbreviation string

	fmt.Println(Green + "Please type in the Password or press Enter for defaults: " + Reset)
	passwd, err := term.ReadPassword(0)

	if err != nil {
		fmt.Println("An error occured while reading input. Please try again: ", err)
	}

	if string(passwd) == "" {
		fmt.Println(Cyan + "setting default password " + Yellow + "[chemotion]" + Reset + "\n")
	}

	fmt.Print(Green + "Please enter first name: " + Reset)
	fmt.Scanln(&fname)
	fname = setDefaultValue(fname, "ELN")
	fmt.Print(Green + "Please enter last name: " + Reset)
	fmt.Scanln(&lname)
	lname = setDefaultValue(lname, "User")
	fmt.Print(Green + "Please enter an " + Yellow + "(unique) " + Green + "abbreviation: " + Reset)
	fmt.Scanln(&abbreviation)
	abbreviation = setDefaultValue(abbreviation, "CU1")
	return passwd, fname, lname, abbreviation
}

// get container ID associated with a given `instance` and `service` of Chemotion
func getContainerID(givenName, service string) (id string) {
	out := getColumn(givenName, "ID", service)
	if len(out) == 2 {
		id = out[0]
	} else {
		id = "not found"
	}
	return
}

func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}
	return false
}

func handleCreateUserLogic() {
	containerID := getContainerID(currentInstance, "eln")
	sourcePath := "./payload/createUser.sh"
	destinationPath := "/embed/scripts"
	scriptPath := destinationPath + "/createUser.sh"

	cmd := exec.Command("docker", "cp", sourcePath, fmt.Sprintf("%s:%s", containerID, destinationPath))
	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}

	// prompt user to input required params
	var email string
	fmt.Print(Green + "\nPlease enter a " + Yellow + "(UNIQUE)" + Reset + " Email Address: " + Reset)
	fmt.Scanln(&email)
	email = setDefaultValue(email, "eln-user@kit.edu")
	passwd, fname, lname, abbreviation := userDataInput()

	args := []string{email, string(passwd), fname, lname, abbreviation}
	argString := strings.Join(args, " ")

	// Set the file permissions to executable
	cmd = exec.Command("docker", "exec", containerID, "chmod", "+x", scriptPath)
	err = cmd.Run()
	if err != nil {
		zboth.Fatal().Err(err).Msgf("Failed to assign proper rights to the script file %s ", scriptPath)
	}

	// Execute script inside docker container
	cmd = exec.Command("docker", "exec", containerID, "bash", "-c", scriptPath+" "+argString)
	stdoutStderr, err := cmd.CombinedOutput()

	if err != nil {
		zboth.Fatal().Err(err).Msgf("Failed to execute script file %s ", scriptPath)
	}

	stdErrConsole := strings.Split(string(stdoutStderr), "\n")
	if contains(stdErrConsole, "false") {
		fmt.Printf("%s\n", stdoutStderr)
		fmt.Printf("It seems, email: %s or abbreviation: %s is already exited\n", Yellow+email+Reset, Yellow+abbreviation+Reset)
		fmt.Println("Please choose unique attributes for new user")
	} else {
		fmt.Printf("Script %s executed successfully inside container %s.\n", Yellow+scriptPath+Reset, Yellow+containerID+Reset)
		fmt.Printf("New user created with email: %s and abbreviation: %s\n", Yellow+email+Reset, Yellow+abbreviation+Reset)
	}
}

func handleDeleteUserLogic() {
	containerID := getContainerID(currentInstance, "eln")
	sourcePath := "./payload/deleteUser.sh"
	destinationPath := "/embed/scripts"
	scriptPath := destinationPath + "/deleteUser.sh"

	cmd := exec.Command("docker", "cp", sourcePath, fmt.Sprintf("%s:%s", containerID, destinationPath))
	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}

	// Set the file permissions to executable
	cmd = exec.Command("docker", "exec", containerID, "chmod", "+x", scriptPath)
	err = cmd.Run()
	if err != nil {
		zboth.Fatal().Err(err).Msgf("Failed to assign proper rights to the script file %s ", scriptPath)
	}

	// prompt user to input required params
	var email string

	print(Red + "WARNING! User will be deleted permanently" + Reset)
	fmt.Print(Yellow + "\nPlease enter an Email Address of a USER you wish to delete: " + Reset)

	fmt.Scanln(&email)

	// Execute script inside docker container
	cmd = exec.Command("docker", "exec", containerID, "/bin/bash", "-c", scriptPath+" "+email)
	stdOutStderr, err := cmd.Output()

	if err != nil {
		zboth.Fatal().Err(err).Msgf("Failed to execute script file %s ", scriptPath)
	}

	fmt.Println(strings.Split(string(stdOutStderr), "\n"))
}

// create a new user of type Admin, Person and Device (for now only type:Person is supported)
var createUserManagementInstanceRootCmd = &cobra.Command{
	Use:     "create",
	Aliases: []string{"c", "create"},
	Args:    cobra.NoArgs,
	Short:   "Manage user actions such as create, add, update and remove user and reset password for " + nameCLI,
	Run: func(cmd *cobra.Command, args []string) {
		if ownCall(cmd) {
			fmt.Println("asd")
			handleCreateUserLogic()

		} else {
			handleCreateUserLogic()
			fmt.Println("asd")
		}
	},
}

// Update an existing user (i,e. first name, last name, password, abbrevitation)
var updateUserManagementInstanceRootCmd = &cobra.Command{
	Use:     "update",
	Aliases: []string{"u", "update"},
	Args:    cobra.NoArgs,
	Short:   "Manage user such as create, add, update and remove user and reset password for " + nameCLI,
	Run: func(cmd *cobra.Command, args []string) {
		// Handle Add update logic here
		if ownCall(cmd) {
			fmt.Println("handleUpdateUserLogic()")
		} else {
			fmt.Println("handleUpdateUserLogic()")
		}
	},
}

// Destroy a particular user from the user management list
var deleteUserManagementInstanceRootCmd = &cobra.Command{
	Use:     "delete",
	Aliases: []string{"d", "delete"},
	Args:    cobra.NoArgs,
	Short:   "Manage user such as create, add, update and remove user and reset password for " + nameCLI,
	Run: func(cmd *cobra.Command, args []string) {
		// Handle delete user logic here
		if ownCall(cmd) {
			handleDeleteUserLogic()
		} else {
			handleDeleteUserLogic()
		}
	},
}

var listUserManagementInstanceRootCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"l", "list"},
	Args:    cobra.NoArgs,
	Short:   "Manage user such as create, add, update and remove user and reset password for " + nameCLI,
	Run: func(cmd *cobra.Command, args []string) {
		// Handle list all users logic here
		fmt.Println("List of all users are as follows")
		if ownCall(cmd) {
			fmt.Println("triggered by own call")
		} else {
			fmt.Println("triggered via cli menu")
		}
	},
}

func init() {
	usermanagementCmd.AddCommand(createUserManagementInstanceRootCmd)
	usermanagementCmd.AddCommand(deleteUserManagementInstanceRootCmd)
	usermanagementCmd.AddCommand(listUserManagementInstanceRootCmd)
	usermanagementCmd.AddCommand(updateUserManagementInstanceRootCmd)
}
