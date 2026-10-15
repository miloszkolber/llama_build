package main

import (
	"flag"
	"os"
	"os/exec"
	"strings"
	"testing"
)

// Run the real entry point in a subprocess so exit/flag behavior is observable.
func TestGateCLI(t *testing.T) {
	if os.Getenv("GATE_CLI_HELPER") == "1" {
		for i, arg := range os.Args {
			if arg == "--" {
				os.Args = append([]string{"gate"}, os.Args[i+1:]...)
				break
			}
		}
		flag.CommandLine = flag.NewFlagSet("gate", flag.ExitOnError)
		main()
		os.Exit(0)
	}
	for _, tc := range []struct {
		args []string
		code int
		want string
	}{
		{[]string{"--health"}, 0, ""},
		{[]string{"--healthcheck"}, 0, ""},
		{[]string{"healthcheck"}, 0, ""},
		{[]string{"--config", "missing.json"}, 2, "flag provided but not defined"},
		{[]string{"--health", "--config", "missing.json"}, 2, "flag provided but not defined"},
		{[]string{"missing.json"}, 2, "unexpected positional"},
		{nil, 2, "configuration error"},
	} {
		t.Run(strings.Join(tc.args, " "), func(t *testing.T) {
			cmd := exec.Command(os.Args[0], append([]string{"-test.run=^TestGateCLI$", "--"}, tc.args...)...)
			cmd.Env = []string{"GATE_CLI_HELPER=1", "GATE_PROFILE=invalid", "DOCKER_SOCKET=/missing/docker.sock"}
			out, err := cmd.CombinedOutput()
			code := 0
			if err != nil {
				if exit, ok := err.(*exec.ExitError); ok {
					code = exit.ExitCode()
				} else {
					t.Fatal(err)
				}
			}
			if code != tc.code || !strings.Contains(string(out), tc.want) {
				t.Fatalf("code=%d output=%s", code, out)
			}
		})
	}
}
