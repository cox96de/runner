package util

import (
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// StringArg is a struct that represents a string argument
type StringArg struct {
	// ArgKey is the key used to bind the flag to viper
	ArgKey string
	// FlagName is the name of the flag. If empty, disable loading from flags
	FlagName string
	// FlagValue is the default value of the flag
	FlagValue string
	// FlagUsage is the usage of the flag
	FlagUsage string
	// Env is the environment variable name, if empty, disable loading from env
	Env string
}

// BindStringArg binds a string argument to the flagset and viper
func BindStringArg(flags *pflag.FlagSet, viper *viper.Viper, a *StringArg) error {
	if len(a.FlagName) > 0 {
		_ = flags.String(a.FlagName, a.FlagValue, a.FlagUsage)
		err := viper.BindPFlag(a.ArgKey, flags.Lookup(a.FlagName))
		if err != nil {
			return err
		}
	}
	if len(a.Env) > 0 {
		err := viper.BindEnv(a.ArgKey, a.Env)
		if err != nil {
			return err
		}
	}

	return nil
}

// IntArg is a struct that represents a int argument
type IntArg struct {
	// ArgKey is the key used to bind the flag to viper
	ArgKey string
	// FlagName is the name of the flag. If empty, disable loading from flags
	FlagName string
	// FlagValue is the default value of the flag
	FlagValue int
	FlagUsage string
	// Env is the environment variable name, if empty, disable loading from env
	Env string
}

// BindIntArg binds a int argument to the flagset and viper
func BindIntArg(flags *pflag.FlagSet, viper *viper.Viper, a *IntArg) error {
	_ = flags.Int(a.FlagName, a.FlagValue, a.FlagUsage)
	err := viper.BindPFlag(a.ArgKey, flags.Lookup(a.FlagName))
	if err != nil {
		return err
	}
	err = viper.BindEnv(a.ArgKey, a.Env)
	if err != nil {
		return err
	}
	return nil
}
