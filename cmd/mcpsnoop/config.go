package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

type config struct {
	Label         string
	TraceFile     string
	NoTrace       bool
	RedactSecrets bool
	RedactKeys    []string
}

func loadConfig() (config, bool, error) {
	const configFile = ".mcpsnoop.toml"

	f, err := os.Open(configFile)
	if os.IsNotExist(err) {
		return config{}, false, nil
	}
	if err != nil {
		return config{}, false, err
	}
	defer f.Close()

	cfg, err := parseConfig(f)
	if err != nil {
		return config{}, false, err
	}

	return cfg, true, nil
}

func parseConfig(r io.Reader) (config, error) {
	var cfg config

	scanner := bufio.NewScanner(r)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			return config{}, fmt.Errorf("invalid config line: %q", line)
		}

		key := strings.TrimSpace(parts[0])
		value := strings.Trim(strings.TrimSpace(parts[1]), `"`)

		switch key {
		case "label":
			cfg.Label = value

		case "trace-file":
			cfg.TraceFile = value

		case "no-trace":
			b, err := strconv.ParseBool(value)
			if err != nil {
				return config{}, fmt.Errorf("invalid value for %q: %w", key, err)
			}
			cfg.NoTrace = b

		case "redact-secrets":
			b, err := strconv.ParseBool(value)
			if err != nil {
				return config{}, fmt.Errorf("invalid value for %q: %w", key, err)
			}
			cfg.RedactSecrets = b

		case "redact-key":
			var keys redactKeysFlag
			if err := keys.Set(value); err != nil {
				return config{}, err
			}
			cfg.RedactKeys = []string(keys)

		default:
			return config{}, fmt.Errorf("unknown config key %q", key)
		}
	}

	if err := scanner.Err(); err != nil {
		return config{}, err
	}

	return cfg, nil
}

func applyConfig(
	fs *flag.FlagSet,
	cfg config,
	ok bool,
	label, traceFile *string,
	noTrace, redactSecrets *bool,
	redactKeys *redactKeysFlag,
) {
	if !ok {
		return
	}

	visited := map[string]bool{}

	fs.Visit(func(f *flag.Flag) {
		visited[f.Name] = true
	})

	if !visited["label"] {
		*label = cfg.Label
	}

	if !visited["trace-file"] {
		*traceFile = cfg.TraceFile
	}

	if !visited["no-trace"] {
		*noTrace = cfg.NoTrace
	}

	if !visited["redact-secrets"] {
		*redactSecrets = cfg.RedactSecrets
	}

	if !visited["redact-key"] {
		*redactKeys = redactKeysFlag(cfg.RedactKeys)
	}
}
