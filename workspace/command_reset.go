package cobra

import (
	"encoding/csv"
	"io"
	"strings"

	"github.com/spf13/pflag"
)

type Command struct {
	Use             string
	Run             func(cmd *Command, args []string)
	flags           *pflag.FlagSet
	persistentFlags *pflag.FlagSet
	commands        []*Command
	args            []string
	out             io.Writer
	err             io.Writer
}

func (c *Command) Flags() *pflag.FlagSet {
	if c.flags == nil {
		c.flags = pflag.NewFlagSet(c.Use, pflag.ContinueOnError)
	}
	return c.flags
}

func (c *Command) PersistentFlags() *pflag.FlagSet {
	if c.persistentFlags == nil {
		c.persistentFlags = pflag.NewFlagSet(c.Use, pflag.ContinueOnError)
	}
	return c.persistentFlags
}

func (c *Command) AddCommand(cmds ...*Command) {
	c.commands = append(c.commands, cmds...)
}

func (c *Command) Commands() []*Command {
	return c.commands
}

func (c *Command) SetArgs(a []string) {
	c.args = a
}

func (c *Command) SetOut(w io.Writer) {
	c.out = w
}

func (c *Command) SetErr(w io.Writer) {
	c.err = w
}

func (c *Command) Execute() error {
	if len(c.args) > 0 {
		subName := c.args[0]
		for _, cmd := range c.commands {
			if cmd.Use == subName {
				// Parse flags from remaining args
				targetFlags := pflag.NewFlagSet(subName, pflag.ContinueOnError)
				if c.persistentFlags != nil {
					targetFlags.AddFlagSet(c.persistentFlags)
				}
				if cmd.flags != nil {
					targetFlags.AddFlagSet(cmd.flags)
				}
				if err := targetFlags.Parse(c.args[1:]); err != nil {
					return err
				}
				if cmd.Run != nil {
					cmd.Run(cmd, c.args[1:])
				}
				return nil
			}
		}
	}
	return nil
}

func resetFlagValue(f *pflag.Flag) {
	// Slice-backed values (stringSlice, intSlice, stringArray, durationSlice, ...)
	// implement pflag.SliceValue. Their Set method has append-on-repeated-call
	// semantics gated by an internal "changed" flag, so calling Set(DefValue) on an
	// already-parsed flag would append the default instead of replacing it. The
	// exported SliceValue.Replace method fully overwrites the underlying slice,
	// restoring the default value. Reflection cannot be used here because the
	// pflag internals are unexported fields (CanSet() is always false).
	if sv, ok := f.Value.(pflag.SliceValue); ok {
		if def := parseSliceDefault(f.DefValue); def != nil {
			_ = sv.Replace(def)
		}
	} else if err := f.Value.Set(f.DefValue); err != nil {
		// resetFlagValue runs as a VisitAll callback with no error channel, so a
		// failed reset leaves the value untouched rather than half-reset.
		return
	}
	f.Changed = false
}

// parseSliceDefault converts a slice flag's default string (e.g. "[a,b]") back
// into the individual comma-separated tokens that SliceValue.Replace expects.
// It returns nil when the default cannot be parsed, signalling the caller to
// skip the reset.
func parseSliceDefault(def string) []string {
	if len(def) >= 2 && strings.HasPrefix(def, "[") && strings.HasSuffix(def, "]") {
		def = def[1 : len(def)-1]
	}
	if def == "" {
		return []string{}
	}
	fields, err := csv.NewReader(strings.NewReader(def)).Read()
	if err != nil {
		return nil
	}
	return fields
}

// ResetFlags resets all flags (both local and persistent) of a command and its subcommands back to their default values.
func (c *Command) ResetFlags() {
	if c.flags != nil {
		c.flags.VisitAll(resetFlagValue)
	}
	if c.persistentFlags != nil {
		c.persistentFlags.VisitAll(resetFlagValue)
	}
	for _, cmd := range c.Commands() {
		cmd.ResetFlags()
	}
}
