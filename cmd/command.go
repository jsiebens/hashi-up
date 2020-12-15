package cmd

import (
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
)

var sshTargetAddr string
var sshTargetUser string
var sshTargetKey string
var runLocal bool

func Execute() error {
	var rootCmd = &cobra.Command{
		Use: "hashi-up",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	var certificate = &cobra.Command{
		Use: "cert",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	var consul = &cobra.Command{
		Use: "consul",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	var nomad = &cobra.Command{
		Use: "nomad",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	var vault = &cobra.Command{
		Use: "vault",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	certificate.AddCommand(CreateCertificateCommand())
	nomad.AddCommand(InstallNomadCommand())
	consul.AddCommand(InstallConsulCommand())
	vault.AddCommand(InstallVaultCommand())

	rootCmd.AddCommand(VersionCommand())
	rootCmd.AddCommand(certificate)
	rootCmd.AddCommand(nomad)
	rootCmd.AddCommand(consul)
	rootCmd.AddCommand(vault)

	rootCmd.PersistentFlags().StringVar(&sshTargetAddr, "ssh-target-addr", "", "Remote SSH target address (e.g. 127.0.0.1:22")
	rootCmd.PersistentFlags().StringVar(&sshTargetUser, "ssh-target-user", "root", "Username for SSH login")
	rootCmd.PersistentFlags().StringVar(&sshTargetKey, "ssh-target-key", "", "The ssh key to use for SSH login")
	rootCmd.PersistentFlags().BoolVar(&runLocal, "local", false, "Running the installation locally, without ssh")

	rootCmd.PersistentFlags().String("ip", "", "Public IP of node")
	rootCmd.PersistentFlags().String("user", "", "Username for SSH login")
	rootCmd.PersistentFlags().String("ssh-key", "", "The ssh key to use for remote login")
	rootCmd.PersistentFlags().String("ssh-port", "", "The port on which to connect for ssh")

	rootCmd.PersistentFlags().MarkDeprecated("ip", "use the new flag ssh-target-addr")
	rootCmd.PersistentFlags().MarkDeprecated("user", "use the new flag ssh-target-user")
	rootCmd.PersistentFlags().MarkDeprecated("ssh-key", "use the new flag ssh-target-key")
	rootCmd.PersistentFlags().MarkDeprecated("ssh-port", "use the new flag ssh-target-addr")

	return rootCmd.Execute()
}

func expandPath(path string) string {
	res, _ := homedir.Expand(path)
	return res
}