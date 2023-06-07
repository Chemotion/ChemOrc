package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

func runRailsCommand(givenName, service, command string) (output string) {
	gotoFolder(givenName)
	bOutput, err := execShell(toSprintf("%s compose exec --workdir /chemotion/app %s bash -c \"echo \\\"%s\\\" | bundle exec rails c\"", virtualizer, service, command))
	gotoFolder("work.dir")
	if err == nil {
		outputLines := strings.Split(string(bOutput), "\n")
		for outputStartsAt := range outputLines {
			if strings.HasPrefix(outputLines[outputStartsAt], command) {
				output = strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(strings.Join(outputLines[outputStartsAt:], " ")), command))
				break
			}
		}
	} else {
		zboth.Fatal().Err(err).Msgf("Failed to execute Rails command: %s in service :%s.", command, service)
	}
	return
}

func userExists(givenName, email string) (err error) {
	output := runRailsCommand(givenName, primaryService, "User.find_by(email:'"+email+"')")
	if output == "nil" {
		err = toError("email not found")
	} else if strings.Contains(output, email) {
		err = nil
	} else {
		zboth.Fatal().Err(toError("output not understood")).Msgf("Failed to understand output from Rails. It is %s", output)
	}
	return
}

func createUser(givenName, email string) {
	firstname := getString("Please enter first name for the user", textValidate)
	lastname := getString("Please enter last name for the user", textValidate)
	abbreviation := getString("Please enter abbreviation name for the user", textValidate)
	typeOfUser := "Person"
	if selectYesNo("Is this user an Admin?", false) {
		typeOfUser = "Admin"
	}
	password := getPassword()
	output := strings.Split(runRailsCommand(givenName, primaryService, toSprintf("User.create(email:'%s', password:'%s', first_name:'%s', last_name:'%s', type:'%s', name_abbreviation:'%s').save", email, password, firstname, lastname, typeOfUser, abbreviation)), " ")
	if toBool(output[len(output)-1]) {
		zboth.Info().Msgf("User created successfully.")
	} else {
		zboth.Warn().Err(toError("user creation failed")).Msgf("Failed to create user. Please ensure that all conditions for abbreviation and password are met.")
	}
}

func modifyUser(givenName, email string) {
	details := getUserDetails(givenName, email)
	options := []string{"First name: " + details["firstname"], "Last name: " + details["lastname"], "Abbreviation: " + details["abbreviation"], "Password"}
	var subStr string
	switch selectOpt(options, "Which value do you want to change") {
	case "First name: " + details["firstname"]:
		subStr = toSprintf("first_name:'%s'", getString("Please enter new first name for the user", textValidate))
	case "Last name: " + details["lastname"]:
		subStr = toSprintf("last_name:'%s'", getString("Please enter new last name for the user", textValidate))
	case "Abbreviation: " + details["abbreviation"]:
		subStr = toSprintf("name_abbreviation:'%s'", getString("Please enter new abbreviation for the user", textValidate))
	case "Password":
		subStr = toSprintf("password:'%s'", getPassword())
	}
	output := strings.Split(runRailsCommand(givenName, primaryService, toSprintf("User.find_by(email:'%s').update(%s)", email, subStr)), " ")
	if toBool(output[len(output)-1]) {
		zboth.Info().Msgf("User details modified successfully.")
	} else {
		zboth.Warn().Err(toError("user detail modification failed")).Msgf("Failed to change user details. Please ensure that all conditions for abbreviation and password are met.")
	}
}

func getUserDetails(givenName, email string) (details map[string]string) {
	details = make(map[string]string)
	detail := strings.Split(strings.TrimFunc(runRailsCommand(givenName, primaryService, "User.where(email:'"+email+"').map {|u| u.first_name + '%' + u.last_name + '%'+ u.name_abbreviation + '%' + u.type }"), func(r rune) bool { return r == '[' || r == ']' || r == '"' }), "%")
	details["firstname"] = detail[0]
	details["lastname"] = detail[1]
	details["fullname"] = details["firstname"] + " " + details["lastname"]
	details["abbreviation"] = detail[2]
	details["type"] = detail[3]
	return
}

func listUsers(givenName string) (names []string) {
	output := runRailsCommand(givenName, primaryService, "User.where(deleted_at:nil).map {|u| u.type + ': ' + u.first_name + ' ' + u.last_name}")
	names = strings.Split(strings.TrimFunc(output, func(r rune) bool { return r == '[' || r == ']' }), ", ")
	for i := range names {
		names[i] = strings.TrimFunc(names[i], func(r rune) bool { return r == '"' })
	}
	return
}

func deleteUser(givenName, email string) {
	if err := userExists(givenName, email); err == nil {
		output := runRailsCommand(givenName, primaryService, "User.find_by(email:'"+email+"').destroy")
		if strings.Contains(output, "@deleted") {
			zboth.Info().Msgf("User associated with address %s deleted successfully.", email)
		} else {
			zboth.Warn().Err(toError("delete failed")).Msgf("Failed to delete a user with this address: %s.", email)
		}
	} else {
		zboth.Warn().Err(err).Msgf("A User associated with email %s was not found.", email)
	}
}

var userInstanceRootCmd = &cobra.Command{
	Use:       "user",
	Aliases:   []string{"users"},
	Short:     "Manage user such as create, add, update and remove user and reset password for " + nameCLI,
	ValidArgs: []string{"create", "list", "update", "describe", "delete"},
	Run: func(cmd *cobra.Command, args []string) {
		var selected string
		if ownCall(cmd) {
			if selected = "multiple arguments"; len(args) == 1 {
				selected = args[0]
			}
		} else {
			if isInteractive(true) {
				acceptedOpts := []string{"create", "list", "update", "describe", "delete"}
				if ownCall(cmd) {
					acceptedOpts = append(acceptedOpts, coloredExit)
				} else {
					acceptedOpts = append(acceptedOpts, []string{"back", coloredExit}...)
				}
				selected = selectOpt(acceptedOpts, "")
			}
		}
		switch selected {
		case "create":
			email := getString("Please enter email address of the user you wish to create", emailValidate)
			if err := userExists(currentInstance, email); err == nil {
				if selectYesNo(toSprintf("A user [%s] with the same address exists, modify this user?", getUserDetails(currentInstance, email)["fullname"]), false) {
					modifyUser(currentInstance, email)
				} else {
					zboth.Info().Msgf("Nothing was done.")
				}
			} else {
				createUser(currentInstance, email)
			}
		case "list":
			names := listUsers(currentInstance)
			if len(names) > 0 {
				zboth.Info().Msgf("The following users exist for instance %s:\n%s", currentInstance, strings.Join(names, "\n"))
			} else {
				zboth.Warn().Err(toError("no users found")).Msgf("No users were gathered from the instance %s.", currentInstance)
			}
		case "describe":
			email := getString("Please enter email address of the user you wish to describe", emailValidate)
			if err := userExists(currentInstance, email); err == nil {
				zboth.Info().Msgf("The user has following details: ")
				for k, v := range getUserDetails(currentInstance, email) {
					// Print on screen - avoids putting this in log unless debug mode is on
					fmt.Printf("%s: %s\n", k, v)
				}
			} else {
				zboth.Fatal().Err(err).Msgf("No user with address %s was found in instance %s.", email, currentInstance)
			}
		case "update":
			email := getString("Please enter email address of the user you wish to modify", emailValidate)
			if err := userExists(currentInstance, email); err == nil {
				modifyUser(currentInstance, email)
			} else {
				if selectYesNo(toSprintf("A user with the address `%s` does not exist, create new user?", email), false) {
					createUser(currentInstance, email)
				} else {
					zboth.Info().Msgf("Nothing was done.")
				}
			}
		case "delete":
			email := getString("Please enter email address of the user you wish to delete", emailValidate)
			deleteUser(currentInstance, email)
		case coloredExit:
			os.Exit(0)
		case "multiple arguments":
			zboth.Warn().Msgf("user subcommand expects only ONE argument of the following: %s.", strings.Join(cmd.ValidArgs, ", "))
		default:
			zboth.Info().Msgf("user subcommand expects one of the following: %s.", strings.Join(cmd.ValidArgs, ", "))
		}
	},
}

func init() {
	instanceRootCmd.AddCommand(userInstanceRootCmd)
}
