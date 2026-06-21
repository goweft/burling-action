// Command entrypoint is the container entrypoint for the burling GitHub
// Action. It runs the burling CLI in SARIF mode against the configured
// target, writes the SARIF document into the workspace, exposes its path
// as a step output, and optionally fails the step on ERROR findings.
//
// Configuration arrives as environment variables set by action.yml:
//
//	BURLING_COMMAND        burling subcommand (default "lint")
//	BURLING_TOKEN          path to the artifact to validate (required)
//	BURLING_STRICT         "true" to pass --strict
//	BURLING_FAIL_ON_ERROR  "true" to exit non-zero when burling reports ERROR
//	BURLING_OUTPUT         SARIF output path (default "burling.sarif")
//
// The burling binary is expected on PATH; the image installs it.
//
// Exit-code contract mirrors burling: 0 clean, 1 ERROR findings, 2 usage
// or I/O error. A code of 2 means burling produced no SARIF, so the
// action fails outright; codes 0 and 1 always yield a valid SARIF, and
// only fail-on-error turns a 1 into a step failure.
package main

import (
	"fmt"
	"os"
	"os/exec"
)

// exitUsageError is burling's exit code for a usage or I/O error, where
// no SARIF is produced. Distinct from finding-driven exit code 1.
const exitUsageError = 2

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "burling-action:", err)
		os.Exit(1)
	}
}

// run executes the action. It returns an error for operational failures
// (missing input, burling not runnable, no SARIF produced, I/O); a
// finding-driven step failure is signalled by exiting via os.Exit inside.
func run() error {
	command := envOr("BURLING_COMMAND", "lint")
	token := os.Getenv("BURLING_TOKEN")
	output := envOr("BURLING_OUTPUT", "burling.sarif")
	strict := isTrue(os.Getenv("BURLING_STRICT"))
	failOnError := isTrue(os.Getenv("BURLING_FAIL_ON_ERROR"))

	if token == "" {
		return fmt.Errorf("no target provided (set the 'token' input)")
	}

	sarif, code, err := runBurling(buildArgs(command, token, strict))
	if err != nil {
		return err
	}
	if code == exitUsageError {
		return fmt.Errorf("burling exited %d (usage or I/O error); no SARIF produced", code)
	}

	if err := os.WriteFile(output, sarif, 0o644); err != nil {
		return fmt.Errorf("write SARIF to %s: %w", output, err)
	}
	if err := setOutput("sarif-file", output); err != nil {
		return fmt.Errorf("set step output: %w", err)
	}
	fmt.Printf("burling %s wrote %s (exit %d)\n", command, output, code)

	if shouldFail(failOnError, code) {
		os.Exit(code)
	}
	return nil
}

// buildArgs assembles the burling argument vector. SARIF output is always
// requested; --strict is appended when enabled; the target is last.
func buildArgs(command, token string, strict bool) []string {
	args := []string{command, "--format", "sarif"}
	if strict {
		args = append(args, "--strict")
	}
	return append(args, token)
}

// runBurling invokes the burling binary, returning its stdout (the SARIF
// document), its exit code, and an error only if the process could not be
// run at all. A non-zero exit from burling (ERROR findings) is returned
// in code, not as an error, because the captured SARIF is still valid.
func runBurling(args []string) (stdout []byte, code int, err error) {
	cmd := exec.Command("burling", args...)
	cmd.Stderr = os.Stderr
	out, runErr := cmd.Output()
	if runErr != nil {
		if exit, ok := runErr.(*exec.ExitError); ok {
			return out, exit.ExitCode(), nil
		}
		return nil, 0, fmt.Errorf("run burling: %w", runErr)
	}
	return out, 0, nil
}

// shouldFail reports whether the action step should exit non-zero: only
// when fail-on-error is enabled and burling exited non-zero.
func shouldFail(failOnError bool, code int) bool {
	return failOnError && code != 0
}

// setOutput appends a name=value pair to the file named by GITHUB_OUTPUT,
// the mechanism GitHub Actions uses for step outputs. When GITHUB_OUTPUT
// is unset (e.g. a local run) it is a no-op.
func setOutput(name, value string) error {
	path := os.Getenv("GITHUB_OUTPUT")
	if path == "" {
		return nil
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = fmt.Fprintf(f, "%s=%s\n", name, value)
	return err
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func isTrue(s string) bool {
	return s == "true" || s == "1" || s == "yes"
}
