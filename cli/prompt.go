package cli

import (
	"os"
	"strconv"
	"strings"

	"github.com/manifoldco/promptui"
)

// Prompt to select a value from a given set of values.
// Also displays the currently selected instance.
func selectOpt(acceptedOpts []string, msg string) (result string) {
	coloredExit := toSprintf("%sexit", string("\033[31m"))
	if acceptedOpts[len(acceptedOpts)-1] == "exit" {
		acceptedOpts[len(acceptedOpts)-1] = coloredExit
	}
	zlog.Debug().Msgf("Selection prompt with options %s:", acceptedOpts)
	if msg == "" {
		msg = toSprintf("%s%s%s%s Select one of the following", string("\033[31m"), string("\033[1m"), currentInstance, string("\033[0m"))
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
	err = textValidate(input)
	if err == nil {
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

// kind of opposite of instanceValidate
func newInstanceValidate(input string) (err error) {
	err = textValidate(input)
	if strings.ContainsRune(input, '.') {
		err = toError("cannot have `.` in an instance name")
	}
	if elementInSlice(input, &reseveredWords) > -1 {
		err = toError("this is a reserved word; pick another")
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
	existingInstances := append(allInstances(), "exit")
	if len(existingInstances) < 6 {
		instance = selectOpt(existingInstances, toSprintf("Please pick the instance to %s:", action))
	} else {
		zboth.Info().Msgf(strings.Join(append([]string{"The following instances exist: "}, allInstances()...), "\n"))
		zlog.Debug().Msgf("String prompt to select instance")
		instance = getString("Please name the instance to "+action, instanceValidate)
	}
	return
}
