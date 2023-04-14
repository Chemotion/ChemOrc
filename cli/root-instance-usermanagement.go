package cli

import (
	"github.com/spf13/cobra"
)

var usermanagementCmdTable = make(cmdTable)

var usermanagementCmd = &cobra.Command{
	Use:     "manage",
	Aliases: []string{"um", "manage"},
	Args:    cobra.NoArgs,
	Short:   "Manage user such as create, add, update and remove user and reset password for " + nameCLI,
	Run: func(cmd *cobra.Command, args []string) {
		isInteractive(true)
		var acceptedOpts []string
		if elementInSlice(instanceStatus(currentInstance), &[]string{"Exited", "Created"}) == -1 { // checks if the instance is running
			acceptedOpts = []string{"create", "list", "update", "delete"}
			usermanagementCmdTable["create"] = createUserManagementInstanceRootCmd.Run
			usermanagementCmdTable["list"] = listUserManagementInstanceRootCmd.Run
			usermanagementCmdTable["update"] = updateUserManagementInstanceRootCmd.Run
			usermanagementCmdTable["delete"] = deleteUserManagementInstanceRootCmd.Run
		} else {
			acceptedOpts = []string{"logs"}
			usermanagementCmdTable["logs"] = logInstanceRootCmd.Run
		}

		usermanagementCmdTable[selectOpt(acceptedOpts, "")](cmd, args)
	},
}

func init() {
	instanceRootCmd.AddCommand(usermanagementCmd)
}
