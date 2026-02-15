package main

import "github.com/spf13/cobra"

func newLoginCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "login",
		Short: "Launch OAuth flow for Cursor authentication",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runLogin()
		},
	}
}

func newLogoutCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "logout",
		Short: "Clear Cursor credentials",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runLogout()
		},
	}
}

func newStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Check auth and proxy status",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStatus()
		},
	}
}

func newStartCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start the proxy server",
		RunE: func(cmd *cobra.Command, args []string) error {
			daemon, _ := cmd.Flags().GetBool("daemon")
			return runStart(daemon)
		},
	}
	cmd.Flags().Bool("daemon", false, "Run in background")
	return cmd
}

func newStopCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "stop",
		Short: "Stop the proxy daemon",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStop()
		},
	}
}

func newModelsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "models",
		Short: "List available Cursor models",
		RunE: func(cmd *cobra.Command, args []string) error {
			jsonOut, _ := cmd.Flags().GetBool("json")
			return runModels(jsonOut)
		},
	}
	cmd.Flags().Bool("json", false, "Output as JSON")
	return cmd
}

func newTestCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "test",
		Short: "Send a test request to verify proxy works",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTest()
		},
	}
}

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runVersion()
		},
	}
}
