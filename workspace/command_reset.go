package cobra

import (
	"io"
	"reflect"

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
	v := reflect.ValueOf(f.Value)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() == reflect.Struct {
		changedField := v.FieldByName("changed")
		if changedField.IsValid() && changedField.CanSet() && changedField.Kind() == reflect.Bool {
			changedField.SetBool(false)
		}
		valueField := v.FieldByName("value")
		if valueField.IsValid() && valueField.CanSet() && valueField.Kind() == reflect.Ptr {
			sliceElem := valueField.Elem()
			if sliceElem.IsValid() && sliceElem.CanSet() && sliceElem.Kind() == reflect.Slice {
				sliceElem.Set(reflect.Zero(sliceElem.Type()))
			}
		}
	}
	f.Value.Set(f.DefValue)
	f.Changed = false
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
