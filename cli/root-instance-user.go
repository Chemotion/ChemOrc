package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

func runRailsCommand(givenName, service, command string) (output string) {
	gotoFolder(givenName)
	escapedCommand := command
	if strings.HasSuffix(shell, "powershell.exe") || strings.HasSuffix(shell, "pwsh.exe") {
		escapedCommand = toSprintf("`\"%s`\"", command) // escape " with `
	} else {
		escapedCommand = toSprintf("\\\"%s\\\"", command) // escape " with \
	}
	bOutput, err := execShell(toSprintf("%s compose exec --workdir /chemotion/app %s bash -c \"echo %s | bundle exec rails c\"", virtualizer, service, escapedCommand))
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

func userExists(givenName, email string) (exists bool) {
	output := runRailsCommand(givenName, primaryService, "User.find_by(email:'"+email+"')")
	if output == "nil" {
		exists = false
	} else if strings.Contains(output, email) {
		exists = true
	} else {
		zboth.Fatal().Err(toError("output not understood")).Msgf("Failed to understand output from Rails. It is %s", output)
	}
	return
}

func createUser(givenName, email, firstname, lastname, typeOfUser, abbreviation, password string) {
	output := strings.Split(runRailsCommand(givenName, primaryService, toSprintf("User.create(email:'%s', password:'%s', first_name:'%s', last_name:'%s', type:'%s', name_abbreviation:'%s').save", email, password, firstname, lastname, typeOfUser, abbreviation)), " ")
	if toBool(output[len(output)-1]) {
		zboth.Info().Msgf("User created successfully.")
	} else {
		zboth.Warn().Err(toError("user creation failed")).Msgf("Failed to create user. Please ensure that all conditions for abbreviation and password are met.")
	}
}

func updateUserInteraction(givenName, email string, firstname *string, lastname *string, abbreviation *string, password *string) {
	details := getUserDetails(givenName, email)
	if *firstname == "" {
		if selectYesNo("Do you wish to change the firstname for the user (current value is: "+details["firstname"]+")", false) {
			*firstname = getString("Please enter new value of firstname for the user", textValidate)
		}
	}
	if *lastname == "" {
		if selectYesNo("Do you wish to change the lastname for the user (current value is: "+details["lastname"]+")", false) {
			*lastname = getString("Please enter new value of lastname for the user", textValidate)
		}
	}
	if *abbreviation == "" {
		if selectYesNo("Do you wish to change the abbreviation for the user (current value is: "+details["abbreviation"]+")", false) {
			*abbreviation = getString("Please enter new value of abbreviation for the user", textValidate)
		}
	}
	if *password == "" {
		if selectYesNo("Do you wish to change the password for the user", false) {
			*password = getPassword()
		}
	}
}

func updateUser(givenName, email, subStr string) (err error) {
	output := strings.Split(runRailsCommand(givenName, primaryService, toSprintf("User.find_by(email:'%s').update(%s)", email, subStr)), " ")
	if toBool(output[len(output)-1]) {
		return nil
	} else {
		return toError("user detail modification failed")
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

func unlockUser(givenName, email string) {
	output := runRailsCommand(givenName, primaryService, "User.find_by(email: '"+email+"').update(locked_at: nil)")
	if strings.Contains(output, "true") {
		zboth.Info().Msgf("User associated with address %s unlocked successfully.", email)
	} else {
		zboth.Warn().Err(toError("unlocked failed")).Msgf("Failed to unlock the user with this address: %s.", email)
	}
}

func deleteUser(givenName, email string) {
	output := runRailsCommand(givenName, primaryService, "User.find_by(email:'"+email+"').destroy")
	if strings.Contains(output, "@deleted") {
		zboth.Info().Msgf("User associated with address %s deleted successfully.", email)
	} else {
		zboth.Warn().Err(toError("delete failed")).Msgf("Failed to delete a user with this address: %s.", email)
	}
}

var userInstanceRootCmd = &cobra.Command{
	Use:       "user",
	Aliases:   []string{"users"},
	Short:     "Manage user such as create, add, update and remove user and reset password for " + nameProject,
	ValidArgs: []string{"list", "unlock", "create", "update", "describe", "delete"},
	Run: func(cmd *cobra.Command, args []string) {
		status := instanceStatus(currentInstance)
		if elementInSlice(status, &[]string{"Created", "Exited"}) > 0 {
			zboth.Fatal().Err(toError("instance is %s", status)).Msgf("Cannot perform operations on %s. Instance is %s.", currentInstance, status)
		}
		acceptedOpts := []string{"list", "unlock", "create", "update", "describe", "delete"}
		selected, firstname, lastname, email, password, abbreviation, typeOfUser := "", "", "", "", "", "", "Person"
		if ownCall(cmd) {
			if len(args) > 0 {
				selected = args[0]
				if elementInSlice(selected, &cmd.ValidArgs) == -1 {
					zboth.Fatal().Err(toError("invalid argument")).Msgf("`user` expects one of the following: %s.", strings.Join(cmd.ValidArgs, ", "))
				}
			}
			if cmd.Flag("firstname").Changed {
				switch selected {
				case "list", "unlock", "describe", "delete":
					zboth.Warn().Err(toError("useless flag")).Msgf("The flag --firstname is useless when used with `%s`", selected)
				case "create", "update":
					firstname = cmd.Flag("firstname").Value.String()
					if err := textValidate(firstname); err != nil {
						zboth.Fatal().Err(err).Msgf("firstname is invalid")
					}
				}
			}
			if cmd.Flag("lastname").Changed {
				switch selected {
				case "list", "unlock", "describe", "delete":
					zboth.Warn().Err(toError("useless flag")).Msgf("The flag --lastname is useless when used with `%s`", selected)
				case "create", "update":
					lastname = cmd.Flag("lastname").Value.String()
					if err := textValidate(lastname); err != nil {
						zboth.Fatal().Err(err).Msgf("lastname is invalid")
					}
				}
			}
			if cmd.Flag("email").Changed {
				switch selected {
				case "list":
					zboth.Warn().Err(toError("useless flag")).Msgf("The flag --email is useless when used with `%s`", selected)
				case "create", "update", "describe", "delete", "unlock":
					email = cmd.Flag("email").Value.String()
					if err := emailValidate(email); err != nil {
						zboth.Fatal().Err(err).Msgf("Invalid email address.")
					}
				}
			}
			if cmd.Flag("password").Changed {
				switch selected {
				case "list", "unlock", "describe", "delete":
					zboth.Warn().Err(toError("useless flag")).Msgf("The flag --password is useless when used with `%s`", selected)
				case "create", "update":
					password = cmd.Flag("password").Value.String()
					if err := textValidate(firstname); err != nil {
						zboth.Fatal().Err(err).Msgf("password is invalid")
					}
				}
			}
			if cmd.Flag("abbreviation").Changed {
				switch selected {
				case "list", "unlock", "describe", "delete":
					zboth.Warn().Err(toError("useless flag")).Msgf("The flag --abbreviatione is useless when used with `%s`", selected)
				case "create", "update":
					abbreviation = cmd.Flag("abbreviation").Value.String()
					if err := textValidate(abbreviation); err != nil {
						zboth.Fatal().Err(err).Msgf("abbreviation is invalid")
					}
				}
			}
		}
		if selected == "" && isInteractive(false) {
			if ownCall(cmd) {
				selected = selectOpt(append(acceptedOpts, coloredExit), "")
			} else {
				selected = selectOpt(append(acceptedOpts, []string{"back", coloredExit}...), "")
			}
		}
		switch selected {
		case "list":
			names := listUsers(currentInstance)
			if len(names) > 0 {
				zboth.Info().Msgf("The following user(s) exist for instance %s:\n%s", currentInstance, strings.Join(names, "\n"))
			} else {
				zboth.Warn().Err(toError("no users found")).Msgf("No users were gathered from the instance %s.", currentInstance)
			}
		case "unlock":
			if email == "" && isInteractive(true) {
				email = getString("Please enter email address of the user you wish to unlock", emailValidate)
			}
			if userExists(currentInstance, email) {
				unlockUser(currentInstance, email)
			} else {
				zboth.Fatal().Err(toError("no associated user")).Msgf("No user with email %s was found in instance %s.", email, currentInstance)
			}
		case "create":
			if email == "" && isInteractive(true) {
				email = getString("Please enter email address of the user you wish to create", emailValidate)
				if selectYesNo("Is this user an Admin?", false) {
					typeOfUser = "Admin"
				}
			}
			if userExists(currentInstance, email) {
				zboth.Fatal().Err(toError("user exists")).Msgf("A user with email %s already exists in instance %s.", email, currentInstance)
			}
			if firstname == "" && isInteractive(true) {
				firstname = getString("Please enter first name for the user", textValidate)
			}
			if lastname == "" && isInteractive(true) {
				lastname = getString("Please enter last name for the user", textValidate)
			}
			if abbreviation == "" && isInteractive(true) {
				abbreviation = getString("Please enter abbreviation name for the user", textValidate)
			}
			if password == "" {
				if isInteractive(false) {
					password = getPassword()
				} else {
					password = getNewUniqueID() + getNewUniqueID()
					fmt.Printf("Setting password as for %s. Please take note - this will not be stored in logs. Password is:\n%s\n", email, password)
				}
			}
			createUser(currentInstance, email, firstname, lastname, typeOfUser, abbreviation, password)
		case "describe":
			if email == "" && isInteractive(true) {
				email = getString("Please enter email address of the user you wish to describe", emailValidate)
			}
			if userExists(currentInstance, email) {
				// Print on screen - avoids putting this personal information in log unless debug mode is on
				fmt.Println("The user has following details: ")
				for k, v := range getUserDetails(currentInstance, email) {
					fmt.Printf("%s: %s\n", k, v)
				}
			} else {
				zboth.Fatal().Err(toError("no associated user")).Msgf("No user with address %s was found in instance %s.", email, currentInstance)
			}
		case "update":
			if email == "" && isInteractive(true) {
				email = getString("Please enter email address of the user you wish to update", emailValidate)
			}
			if isInteractive(false) {
				if firstname == "" && lastname == "" && abbreviation == "" && password == "" {
					zboth.Fatal().Err(toError("no flag used")).Msgf("You must specify --email and at least one of these to use `update` in quiet mode: --firstname, --lastname, --abbreviation, --password.")
				}
			}
			if userExists(currentInstance, email) {
				if isInteractive(false) {
					if firstname == "" && lastname == "" && abbreviation == "" && password == "" {
						updateUserInteraction(currentInstance, email, &firstname, &lastname, &abbreviation, &password)
					}
				}
				if firstname != "" {
					subStr := toSprintf("first_name:'%s'", firstname)
					if err := updateUser(currentInstance, email, subStr); err != nil {
						zboth.Fatal().Err(err).Msgf("Failed to change user's firstname.")
					} else {
						zboth.Info().Msgf("Firstname changed successfully.")
					}
				}
				if lastname != "" {
					subStr := toSprintf("last_name:'%s'", lastname)
					if err := updateUser(currentInstance, email, subStr); err != nil {
						zboth.Fatal().Err(err).Msgf("Failed to change user's lastname.")
					} else {
						zboth.Info().Msgf("Lastname changed successfully.")
					}
				}
				if abbreviation != "" {
					subStr := toSprintf("name_abbreviation:'%s'", abbreviation)
					if err := updateUser(currentInstance, email, subStr); err != nil {
						zboth.Fatal().Err(err).Msgf("Failed to change user's abbreviation. Please ensure that all conditions for abbreviation are met.")
					} else {
						zboth.Info().Msgf("Abbreviation changed successfully.")
					}
				}
				if password != "" {
					subStr := toSprintf("password:'%s'", password)
					if err := updateUser(currentInstance, email, subStr); err != nil {
						zboth.Fatal().Err(err).Msgf("Failed to change user's password. Please ensure that all conditions for password are met.")
					} else {
						zboth.Info().Msgf("Password changed successfully.")
					}
				}
			} else {
				zboth.Fatal().Err(toError("no associated user")).Msgf("No user with address %s was found in instance %s.", email, currentInstance)
			}
		case "delete":
			if email == "" && isInteractive(true) {
				email = getString("Please enter email address of the user you wish to delete", emailValidate)
			}
			if userExists(currentInstance, email) {
				deleteUser(currentInstance, email)
			} else {
				zboth.Fatal().Err(toError("no associated user")).Msgf("No user with address %s was found in instance %s.", email, currentInstance)
			}
		case "back":
			cmd.Run(cmd, args)
		case coloredExit:
			os.Exit(0)
		case "multiple arguments":
			zboth.Warn().Msgf("user subcommand expects only ONE argument of the following: %s.", strings.Join(acceptedOpts, ", "))
		default:
			zboth.Info().Msgf("user subcommand expects one of the following: %s.", strings.Join(acceptedOpts, ", "))
		}
	},
}

func init() {
	instanceRootCmd.AddCommand(userInstanceRootCmd)
	userInstanceRootCmd.Flags().String("firstname", "", "(updated) firstname of the new (existing) user")
	userInstanceRootCmd.Flags().String("lastname", "", "(updated) lastname of the new (existing) user")
	userInstanceRootCmd.Flags().String("email", "", "email address of the user to unlock/create/update/delete")
	userInstanceRootCmd.Flags().String("password", "", "new password for the user")
	userInstanceRootCmd.Flags().String("abbreviation", "", "new/updated abbreviation for the user")
}
