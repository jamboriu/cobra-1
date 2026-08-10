package cobra

import (
	"bytes"
	"testing"
)

func TestResetFlags(t *testing.T) {
	var host string

	rootCmd := &Command{
		Use: "root",
		Run: func(cmd *Command, args []string) {},
	}
	rootCmd.PersistentFlags().StringVar(&host, "host", "localhost", "host address")

	subCmd := &Command{
		Use: "sub",
		Run: func(cmd *Command, args []string) {},
	}
	rootCmd.AddCommand(subCmd)

	// 1. Execute with --host remotehost
	rootCmd.SetArgs([]string{"sub", "--host", "remotehost"})
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if host != "remotehost" {
		t.Errorf("expected host to be 'remotehost', got '%s'", host)
	}

	// 2. Reset flags
	rootCmd.ResetFlags()

	// 3. Execute again with no arguments
	rootCmd.SetArgs([]string{"sub"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if host != "localhost" {
		t.Errorf("expected host to be reset to 'localhost', got '%s'", host)
	}
}

func TestResetFlagsSlice(t *testing.T) {
	var hosts []string

	rootCmd := &Command{
		Use: "root",
		Run: func(cmd *Command, args []string) {},
	}
	rootCmd.PersistentFlags().StringSliceVar(&hosts, "hosts", []string{"localhost"}, "host addresses")

	subCmd := &Command{
		Use: "sub",
		Run: func(cmd *Command, args []string) {},
	}
	rootCmd.AddCommand(subCmd)

	// 1. Execute with --hosts remote1,remote2
	rootCmd.SetArgs([]string{"sub", "--hosts", "remote1,remote2"})
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(hosts) != 2 || hosts[0] != "remote1" || hosts[1] != "remote2" {
		t.Errorf("expected hosts to be [remote1, remote2], got %v", hosts)
	}

	// 2. Reset flags
	rootCmd.ResetFlags()

	// 3. Execute again with no arguments
	rootCmd.SetArgs([]string{"sub"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(hosts) != 1 || hosts[0] != "localhost" {
		t.Errorf("expected hosts to be reset to [localhost], got %v", hosts)
	}
}
