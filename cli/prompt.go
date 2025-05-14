package cli

import (
	"os"
	"strconv"
	"strings"

	"github.com/manifoldco/promptui"
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
	selection := promptui.Select{
		Label: msg,
		Items: acceptedOpts,
	}
	_, result, err := selection.Run()
	if err == nil {
		zlog.Debug().Msgf("Selected option: %s", result)
	} else if err == promptui.ErrInterrupt || err == promptui.ErrEOF {
		zboth.Fatal().Err(err).Msgf("Selection cancelled!")
	} else {
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
	var defValueStr string
	if defValue {
		defValueStr = "y"
	} else {
		defValueStr = "n"
	}
	answer := promptui.Prompt{
		Label:     question,
		IsConfirm: true,
		Default:   defValueStr,
	}
	if _, err := answer.Run(); err == nil {
		result = true
	} else if err == promptui.ErrAbort {
		result = false
	} else if err == promptui.ErrInterrupt || err == promptui.ErrEOF {
		zboth.Fatal().Err(toError("yesno prompt cancelled")).Msgf("Selection cancelled.")
	} else {
		zboth.Fatal().Err(err).Msgf("Selection failed! Check log. ABORT!")
	}
	zlog.Debug().Msgf("Selected answer: %t", result)
	return
}

func emailValidate(input string) (err error) {
	if err = textValidate(input); err == nil {
		if strings.Count(input, "@") != 1 || strings.Count(input, ".") < 1 {
			err = toError("please input a valid email address")
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
func getString(message string, validator promptui.ValidateFunc) (result string) {
	zlog.Debug().Msgf("String prompt with message: %s", message)
	prompt := promptui.Prompt{
		Label:    message,
		Validate: validator,
	}
	if res, err := prompt.Run(); err == nil {
		zlog.Debug().Msgf("Given answer: %s", res)
		result = res
	} else if err == promptui.ErrInterrupt || err == promptui.ErrEOF {
		zboth.Fatal().Err(toError("prompt cancelled")).Msgf("Prompt cancelled. Can't proceed without. ABORT!")
	} else {
		zboth.Fatal().Err(err).Msgf("Prompt failed because: %s.", err.Error())
	}
	return
}

// to select an instance, gives a list to select from when less than 5, else a text input
func selectInstance(action string) (instance string) {
	existingInstances := append(allInstances(), coloredExit)
	if len(existingInstances) < 6 {
		instance = selectOpt(existingInstances, toSprintf("Please pick the instance to %s:", action))
	} else {
		zboth.Info().Msg(strings.Join(append([]string{"The following instances exist: "}, allInstances()...), "\n"))
		zlog.Debug().Msg("String prompt to select instance")
		instance = getString("Please name the instance to "+action, instanceValidate)
	}
	return
}

// to get a new password
func getPassword() (password string) {
	if zerolog.GlobalLevel() == zerolog.DebugLevel {
		zboth.Warn().Err(toError("password in debug mode")).Msg(color.Color("You are gathering password while in debug mode. [red]!!! The password will be stored in the log file as plain-text !!![reset] Exit now to avoid this."))
	}
	prompt := promptui.Prompt{
		Label:       "Please enter new password",
		Mask:        '*',
		HideEntered: true,
	}
	var confirm string
	var err error
	if confirm, err = prompt.Run(); err == nil {
		zlog.Debug().Msgf("Password, first attempt gathered")
		prompt.Label = "Please re-enter the same password to confirm it"
		if password, err = prompt.Run(); err == nil {
			zlog.Debug().Msgf("Password, second attempt gathered")
			if password != confirm {
				zboth.Warn().Err(toError("password mismatch")).Msgf("The re-entered password does not match, please try again.")
				password = getPassword()
			}
		}
	}
	if err == promptui.ErrInterrupt || err == promptui.ErrEOF {
		zboth.Fatal().Err(toError("prompt cancelled")).Msgf("Prompt cancelled. Can't proceed without. ABORT!")
	} else if err != nil {
		zboth.Fatal().Err(err).Msgf("Prompt failed because: %s.", err.Error())
	}
	return
}
