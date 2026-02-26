package cli

import (
	"os"
	"strconv"
	"strings"

	prompt "github.com/charmbracelet/huh"
	color "github.com/mitchellh/colorstring"
	"github.com/rs/zerolog"
)

// Prompt to select a value from a given set of values.
// Also displays the currently selected instance.
func selectOpt(acceptedOpts []string, msg string) (result string) {
	zlog.Debug().Msgf("Selection prompt with options %s:", acceptedOpts)
	if msg == "" {
		if currentInstance == "" {
			msg = "Select one of the following"
		} else {
			msg = color.Color(toSprintf("[green][dim]{%s} ", currentInstance)) + "Select one of the following"
		}
	}
	options := make([]prompt.Option[string], len(acceptedOpts))
	for i, opt := range acceptedOpts {
		options[i] = prompt.NewOption(opt, opt)
	}
	err := prompt.NewSelect[string]().Title(msg).Options(options...).Value(&result).Run()
	switch err {
	case nil:
		zlog.Debug().Msgf("Selected option: %s", result)
	case prompt.ErrUserAborted:
		zboth.Fatal().Err(toError("selection prompt cancelled")).Msgf("Selection cancelled.")
	default:
		zboth.Fatal().Err(err).Msgf("Selection failed! Check log. ABORT!")
	}
	if result == coloredExit {
		zboth.Debug().Msgf("Chose to exit")
		os.Exit(0)
	}
	return
}

// A simple yes or no question prompt. Yes = True, No = False.
func selectYesNo(question string, defValue bool) (result bool) {
	zlog.Debug().Msgf("Binary question: %s; default is: %t", question, defValue)
	result = defValue
	err := prompt.NewConfirm().Title(question).Inline(true).Value(&result).Run()
	switch err {
	case nil:
		zlog.Debug().Msgf("Selected option: %t", result)
	case prompt.ErrUserAborted:
		zboth.Fatal().Err(toError("yesno prompt cancelled")).Msgf("Selection cancelled.")
	default:
		zboth.Fatal().Err(err).Msgf("Selection failed! Check log. ABORT!")
	}
	return
}

func emailValidate(input string) (err error) {
	if err = textValidate(input); err == nil {
		if strings.Count(input, "@") != 1 || strings.Count(input, ".") < 1 {
			err = toError("please input a valid email address")
		} else if strings.Count(input, " ") > 0 {
			err = toError("email address can not contain any spaces")
		} else {
			err = nil
		}
	}
	return
}

func textValidate(input string) (err error) {
	if len(strings.ReplaceAll(input, " ", "")) == 0 {
		err = toError("can not accept empty value")
	} else if len(strings.Fields(input)) > 1 || strings.ContainsRune(input, ' ') {
		err = toError("can not have spaces in this input")
	} else {
		err = nil
	}
	return
}

func instanceValidate(input string) (err error) {
	if err = textValidate(input); err == nil {
		if len(getSubHeadings(&conf, joinKey(instancesWord, input))) == 0 {
			err = toError("there is no instance called %s", input)
		}
	}
	return
}

func addressValidate(input string) (err error) {
	if err = textValidate(input); err == nil {
		protocol, address, found := strings.Cut(input, "://")
		if found && ((protocol == "http") || (protocol == "https")) {
			address, port, portGiven := strings.Cut(address, ":")
			if err = textValidate(address); err == nil {
				if portGiven {
					if p, errConv := strconv.Atoi(port); errConv != nil || p < 1 {
						err = toError("port must an integer above 0")
					}
				}
			} else {
				err = toError("address cannot be empty")
			}
		} else {
			err = toError("address must start with protocol i.e. as `http://` or as `https://`")
		}
	}
	return
}

func fileValidate(input string) (err error) {
	err = textValidate(input)
	if !existingFile(input) {
		err = toError("this file does not exist")
	}
	return
}

// kind of opposite of instanceValidate
func newInstanceValidate(input string) (err error) {
	err = textValidate(input)
	for _, char := range []rune{'.', '/', '\\', ':'} {
		if strings.ContainsRune(input, char) {
			err = toError("cannot have `%c` in an instance name", char)
		}
	}
	if toLower(input) != input {
		err = toError("cannot have an uppercase letter") // because case is not preserved by Viper v1.x
	}
	if err == nil {
		if exists := instanceValidate(input); exists == nil {
			err = toError("this value is already taken")
		} else {
			err = nil
		}
	}
	return
}

// Get user input in form of a string by giving them the message.
func getString(message string, suggestions []string, validator func(string) error) (result string) {
	zlog.Debug().Msgf("String prompt with message: %s", message)
	var err error
	if len(suggestions) == 0 {
		err = prompt.NewInput().Title(message).Inline(true).Validate(validator).Value(&result).Run()
	} else {
		err = prompt.NewForm(prompt.NewGroup(prompt.
			NewInput().Title(message).
			Suggestions(suggestions).
			Description("Suggested options:\n" + strings.Join(suggestions[:len(suggestions)-1], ", ") + " and " + suggestions[len(suggestions)-1] + ".").
			Inline(false).Validate(validator).Value(&result))).Run()
	}
	switch err {
	case nil:
		zlog.Debug().Msgf("Given answer: %s", result)
	case prompt.ErrUserAborted:
		zboth.Fatal().Err(toError("get string prompt cancelled")).Msgf("Selection cancelled.")
	default:
		zboth.Fatal().Err(err).Msgf("Selection failed! Check log. ABORT!")
	}
	return
}

// to select an instance, gives a list to select from if 4 or less, else a text input
func selectInstance(action string) (instance string) {
	listInstances := allInstances()
	if len(listInstances) < 5 {
		instance = selectOpt(append(listInstances, coloredExit), toSprintf("Please pick the instance to %s:", action))
	} else {
		zlog.Debug().Msg("String prompt to select instance")
		instance = getString("Please name the instance to "+action, listInstances, instanceValidate)
	}
	return
}

// to get a new password
func getPassword() (password string) {
	if zerolog.GlobalLevel() == zerolog.DebugLevel {
		zboth.Warn().Err(toError("password in debug mode")).Msg(color.Color("You are gathering password while in debug mode. [red]!!! The password will be stored in the log file as plain-text !!![reset] Exit now to avoid this."))
	}
	var confirm string
	err := prompt.NewInput().Title("Please enter new password").
		Validate(textValidate).
		EchoMode(prompt.EchoModePassword).
		Inline(true).Value(&password).Run()
	switch err {
	case nil:
		zlog.Debug().Msgf("Password, first attempt gathered")
		confirm = password
		password = "" // reset value because pointer is used to display default value
		err = prompt.NewInput().Title("Please re-enter the same password to confirm it").
			Validate(textValidate).
			EchoMode(prompt.EchoModePassword).
			Inline(true).Value(&password).Run()
		switch err {
		case nil:
			zlog.Debug().Msgf("Password, second attempt gathered")
			if password != confirm {
				zboth.Warn().Err(toError("password mismatch")).Msgf("The re-entered password does not match, please try again.")
				password = getPassword()
			}
		case prompt.ErrUserAborted:
			zboth.Fatal().Err(toError("password prompt cancelled")).Msgf("Password prompt cancelled. Can't proceed without.")
		default:
			zboth.Fatal().Err(err).Msgf("Password prompt failed! Check log. ABORT!")
		}
	case prompt.ErrUserAborted:
		zboth.Fatal().Err(toError("password prompt cancelled")).Msgf("Prompt cancelled. Can't proceed without.")
	default:
		zboth.Fatal().Err(err).Msgf("Password prompt failed! Check log. ABORT!")
	}
	return
}
