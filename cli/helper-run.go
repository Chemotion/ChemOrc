package cli

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// check if the CLI is running interactively; if interactive == true && fail == true, then exit. Wrapper around conf.GetBool(joinKey(stateWord,"quiet")).
func isInteractive(fail bool) (interactive bool) {
	interactive = true
	if conf.GetBool(joinKey(stateWord, "quiet")) { // if the key does not exist, this returns false which implies that the value of `interactive` remains unchanged
		interactive = false
		if fail {
			zboth.Fatal().Err(toError("incomplete in quiet mode")).Msgf("%s is in quiet mode. Give all arguments to specify the desired action; use '--help' flag for more. ABORT!", nameCLI)
		}
	}
	return
}

// check if an element is in an array of type(element), if yes, return the 1st index, else -1.
func elementInSlice[T uint64 | int | float64 | string](elem T, slice *[]T) int {
	for index, element := range *slice {
		if element == elem {
			return index
		}
	}
	return -1
}

// generate a new UID (of the form xxxxxxxx) as a string
func getNewUniqueID() string {
	id, _ := uuid.NewRandom()
	return strings.Split(id.String(), "-")[0]
}

// to manage config as loaded into Viper
func getSubHeadings(yamlConf *viper.Viper, key string) (subheadings []string) {
	for k := range yamlConf.GetStringMapString(key) {
		subheadings = append(subheadings, k)
	}
	return
}

// join keys so as to access them in a viper configuration
func joinKey(s ...string) (result string) {
	result = strings.Join(s, ".")
	return
}

// to lower case, same as strings.ToLower
var toLower = strings.ToLower

// to shorten fmt statements
var toError = fmt.Errorf
var toSprintf = fmt.Sprintf

// toBool
func toBool(s string) (value bool) {
	if toLower(s) == "true" {
		value = true
	} else if toLower(s) == "false" {
		value = false
	} else {
		zboth.Fatal().Msgf("cannot convert %s to bool", s)
	}
	return
}

// determine if the command was called on its own (true) or accessed via a menu (false);
// only works if the command has no children of its own
func ownCall(cmd *cobra.Command) bool {
	return len(cmd.Commands()) == 0 // a command is accessed on its own if there are no child commands
}

// to get all existing instances as determined by the configuration file
func allInstances() (instances []string) {
	instances = getSubHeadings(&conf, instancesWord)
	return
}

// to get all existing used ports
func allPorts() (ports []uint64) {
	existingInstances := allInstances()
	for _, instance := range existingInstances {
		ports = append(ports, conf.GetUint64(joinKey(instancesWord, instance, "port")))
	}
	return
}

// get internal name for an instance
func getInternalName(givenName string) (name string) {
	if err := instanceValidate(givenName); err == nil {
		name = conf.GetString(joinKey(instancesWord, givenName, "name"))
	} else {
		zboth.Fatal().Err(err).Msgf("No such instance: %s", givenName)
	}
	return
}

// get column associated with `ps` output for a given instance of chemotion
func getColumn(givenName, column, service string) (values []string) {
	name := getInternalName(givenName)
	filterStr := toSprintf("--filter \"label=net.chemotion.cli.project=%s\"", name)
	if service != "" { // for a specific service
		filterStr += toSprintf(" --filter name=%s", service)
	}
	if column == "Status" {
		if res, err := execShell(toSprintf("%s ps -a %s --format \"{{.Status}} {{.Names}}\"", virtualizer, filterStr)); err == nil {
			lines := strings.Split(string(res), "\n")
			for _, line := range lines {
				if strings.Contains(line, "dbupgrade") { // ignore dbupgrade containers when checking for instance status
					continue
				} else {
					values = append(values, line)
				}
			}
		}
	} else { // for all other kinds of column, currently the only one used is "Names", for getInternalName
		if res, err := execShell(toSprintf("%s ps -a %s --format \"{{.%s}}\"", virtualizer, filterStr, column)); err == nil {
			values = strings.Split(string(res), "\n")
		}
	}
	return
}

// get services associated with a given `instance` of Chemotion
func getServices(givenName string) (services []string) {
	name, out := getInternalName(givenName), getColumn(givenName, "Names", "")
	for _, line := range out { // determine what are the status messages for all associated containers
		l := strings.TrimSpace(line) // use only the first word
		if len(l) > 0 {
			l = strings.TrimPrefix(l, toSprintf("%s-", name))
			l = strings.TrimSuffix(l, toSprintf("-%d", rollNum))
			services = append(services, l)
		}
	}
	return
}

// get container ID associated with a given `instance` and `service` of Chemotion
// func getContainerID(givenName, service string) (id string) {
// 	out := getColumn(givenName, "ID", service)
// 	if len(out) == 2 {
// 		id = out[0]
// 	} else {
// 		id = "not found"
// 	}
// 	return
// }

// split address into subcomponents
func splitAddress(address string) (protocol string, domain string, port uint64) {
	if err := addressValidate(address); err == nil {
		var portStr string
		protocol, address, _ = strings.Cut(address, ":")
		address = strings.TrimPrefix(address, "//")
		domain, portStr, _ = strings.Cut(address, ":")
		if port = 0; portStr != "" {
			p, _ := strconv.Atoi(portStr)
			port = uint64(p)
		}
	} else {
		zboth.Fatal().Err(err).Msgf("Given address %s is invalid.", address)
	}
	return
}
