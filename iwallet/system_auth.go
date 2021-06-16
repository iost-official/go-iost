package iwallet

import (
	"strconv"

	"github.com/spf13/cobra"
)

var addpermCmd = &cobra.Command{
	Use:     "add-permission permission threshold",
	Aliases: []string{"addperm"},
	Short:   "add permission to this account",
	Long:    "add permission to this account",
	Example: `  iwallet sys addperm myactive 100`,
	Args: func(cmd *cobra.Command, args []string) error {
		if err := checkArgsNumber(cmd, args, "permission", "threshold"); err != nil {
			return err
		}
		return checkAccount(cmd)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		threshold, err := strconv.Atoi(args[1])
		if err != nil {
			return err
		}
		return processMethod("auth.iost", "addPermission", accountName, args[0], threshold)
	},
}

var droppermCmd = &cobra.Command{
	Use:     "drop-permission permission",
	Aliases: []string{"dropperm"},
	Short:   "drop permission of this account",
	Long:    "drop permission of this account",
	Example: `  iwallet sys dropperm myactive`,
	Args: func(cmd *cobra.Command, args []string) error {
		if err := checkArgsNumber(cmd, args, "permission"); err != nil {
			return err
		}
		return checkAccount(cmd)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return processMethod("auth.iost", "dropPermission", accountName, args[0])
	},
}

var assignPermCmd = &cobra.Command{
	Use:     "assign-permission permission item weight",
	Aliases: []string{"assignperm"},
	Short:   "assign item to permission",
	Long:    "assign item to permission",
	Example: `  iwallet sys assignperm myactive someone@perm 50
  iwallet sys assignperm myactive EhNiaU4DzUmjCrvynV3gaUeuj2VjB1v2DCmbGD5U2nSE 50`,
	Args: func(cmd *cobra.Command, args []string) error {
		if err := checkArgsNumber(cmd, args, "permission", "item", "weight"); err != nil {
			return err
		}
		return checkAccount(cmd)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		threshold, err := strconv.Atoi(args[2])
		if err != nil {
			return err
		}
		return processMethod("auth.iost", "assignPermission", accountName, args[0], args[1], threshold)
	},
}

var revokePermCmd = &cobra.Command{
	Use:     "revoke-permission permission item",
	Aliases: []string{"revokeperm"},
	Short:   "revoke item to permission",
	Long:    "revoke item to permission",
	Example: `  iwallet sys revokeperm myactive someone@perm
  iwallet sys revokeperm myactive EhNiaU4DzUmjCrvynV3gaUeuj2VjB1v2DCmbGD5U2nSE`,
	Args: func(cmd *cobra.Command, args []string) error {
		if err := checkArgsNumber(cmd, args, "permission", "item"); err != nil {
			return err
		}
		return checkAccount(cmd)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return processMethod("auth.iost", "revokePermission", accountName, args[0], args[1])
	},
}

var addgroupCmd = &cobra.Command{
	Use:     "add-group group_name",
	Aliases: []string{"addgroup"},
	Short:   "add a group to account",
	Long:    "add a group to account",
	Example: `  iwallet sys addgroup mygroup`,
	Args: func(cmd *cobra.Command, args []string) error {
		if err := checkArgsNumber(cmd, args, "group"); err != nil {
			return err
		}
		return checkAccount(cmd)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return processMethod("auth.iost", "addGroup", accountName, args[0])
	},
}

var dropgroupCmd = &cobra.Command{
	Use:     "drop-group group_name",
	Aliases: []string{"dropgroup"},
	Short:   "drop group",
	Long:    "drop group",
	Example: `  iwallet sys dropgroup mygroup`,
	Args: func(cmd *cobra.Command, args []string) error {
		if err := checkArgsNumber(cmd, args, "group"); err != nil {
			return err
		}
		return checkAccount(cmd)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return processMethod("auth.iost", "dropGroup", accountName, args[0])
	},
}

var assignGroupCmd = &cobra.Command{
	Use:     "assign-group permission item weight",
	Aliases: []string{"assigngroup"},
	Short:   "assign item to group",
	Long:    "assign item to group",
	Example: `  iwallet sys assignperm myactive someone@perm 50
  iwallet sys assignperm myactive EhNiaU4DzUmjCrvynV3gaUeuj2VjB1v2DCmbGD5U2nSE 50`,
	Args: func(cmd *cobra.Command, args []string) error {
		if err := checkArgsNumber(cmd, args, "group", "item", "weight"); err != nil {
			return err
		}
		return checkAccount(cmd)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		threshold, err := strconv.Atoi(args[2])
		if err != nil {
			return err
		}
		return processMethod("auth.iost", "assignGroup", accountName, args[0], args[1], threshold)
	},
}

var revokeGroupCmd = &cobra.Command{
	Use:     "revoke-group group item",
	Aliases: []string{"revokegroup"},
	Short:   "revoke item to group",
	Long:    "revoke item to group",
	Example: `  iwallet sys revokeperm myactive someone@perm
  iwallet sys revokeperm myactive EhNiaU4DzUmjCrvynV3gaUeuj2VjB1v2DCmbGD5U2nSE`,
	Args: func(cmd *cobra.Command, args []string) error {
		if err := checkArgsNumber(cmd, args, "group", "item"); err != nil {
			return err
		}
		return checkAccount(cmd)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return processMethod("auth.iost", "revokeGroup", accountName, args[0], args[1])
	},
}

var bindPermCmd = &cobra.Command{
	Use:     "assign-permission-to-group permission group",
	Aliases: []string{"bindperm"},
	Short:   "bind permission into a group",
	Long:    "bind permission into a group",
	Example: `  iwallet sys bindperm myperm mygroup`,
	Args: func(cmd *cobra.Command, args []string) error {
		if err := checkArgsNumber(cmd, args, "permission", "group"); err != nil {
			return err
		}
		return checkAccount(cmd)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return processMethod("auth.iost", "assignPermissionToGroup", accountName, args[0], args[1])
	},
}

var unbindPermCmd = &cobra.Command{
	Use:     "revoke-pemission-in-group permission group",
	Aliases: []string{"unbindperm"},
	Short:   "unbind permission into a group",
	Long:    "unbind permission into a group",
	Example: `  iwallet sys unbindperm myperm mygroup`,
	Args: func(cmd *cobra.Command, args []string) error {
		if err := checkArgsNumber(cmd, args, "permission", "group"); err != nil {
			return err
		}
		return checkAccount(cmd)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return processMethod("auth.iost", "revokePermissionInGroup", accountName, args[0], args[1])
	},
}

func init() {
	systemCmd.AddCommand(addpermCmd)
	systemCmd.AddCommand(droppermCmd)
	systemCmd.AddCommand(assignPermCmd)
	systemCmd.AddCommand(revokePermCmd)
	systemCmd.AddCommand(addgroupCmd)
	systemCmd.AddCommand(dropgroupCmd)
	systemCmd.AddCommand(assignGroupCmd)
	systemCmd.AddCommand(revokeGroupCmd)
	systemCmd.AddCommand(bindPermCmd)
	systemCmd.AddCommand(unbindPermCmd)
}
