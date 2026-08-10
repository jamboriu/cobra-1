package cobra

import (
	"reflect"

	"github.com/spf13/pflag"
)

// ResetFlags resets all flags (both local and persistent) of a command and its subcommands back to their default values.
func (c *Command) ResetFlags() {
	if c.flags != nil {
		c.flags.VisitAll(func(f *pflag.Flag) {
			// For slice/array flags, pflag values often keep an internal 'changed' representation
			// to know when to clear the default value before appending.
			// We use reflection to reset this 'changed' field if present.
			v := reflect.ValueOf(f.Value)
			if v.Kind() == reflect.Ptr {
				v = v.Elem()
			}
			if v.Kind() == reflect.Struct {
				changedField := v.FieldByName("changed")
				if changedField.IsValid() && changedField.CanSet() && changedField.Kind() == reflect.Bool {
					changedField.SetBool(false)
				}
			}
			f.Value.Set(f.DefValue)
			f.Changed = false
		})
	}
	if c.persistentFlags != nil {
		c.persistentFlags.VisitAll(func(f *pflag.Flag) {
			v := reflect.ValueOf(f.Value)
			if v.Kind() == reflect.Ptr {
				v = v.Elem()
			}
			if v.Kind() == reflect.Struct {
				changedField := v.FieldByName("changed")
				if changedField.IsValid() && changedField.CanSet() && changedField.Kind() == reflect.Bool {
					changedField.SetBool(false)
				}
			}
			f.Value.Set(f.DefValue)
			f.Changed = false
		})
	}
	for _, cmd := range c.Commands() {
		cmd.ResetFlags()
	}
}
