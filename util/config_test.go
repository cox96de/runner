package util

import (
	"testing"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"gotest.tools/v3/assert"
)

func TestBindStringArg(t *testing.T) {
	t.Run("default", func(t *testing.T) {
		vv := viper.New()
		flagSet := pflag.NewFlagSet(t.Name(), pflag.ContinueOnError)
		err := BindStringArg(flagSet, vv, &StringArg{
			ArgKey:    "str_key",
			FlagName:  "str_key",
			FlagValue: "default",
			FlagUsage: "usage for str_key",
			Env:       "STR_KEY",
		})
		assert.NilError(t, err)
		value := vv.Get("str_key")
		assert.DeepEqual(t, value, "default")
	})
	t.Run("from_env", func(t *testing.T) {
		t.Setenv("STR_KEY", "from_env")
		vv := viper.New()
		flagSet := pflag.NewFlagSet(t.Name(), pflag.ContinueOnError)
		err := BindStringArg(flagSet, vv, &StringArg{
			ArgKey:    "str_key",
			FlagName:  "str_key",
			FlagValue: "default",
			FlagUsage: "usage for str_key",
			Env:       "STR_KEY",
		})
		assert.NilError(t, err)
		value := vv.Get("str_key")
		assert.DeepEqual(t, value, "from_env")
	})

	t.Run("from_flag", func(t *testing.T) {
		t.Setenv("STR_KEY", "from_env")
		vv := viper.New()
		flagSet := pflag.NewFlagSet(t.Name(), pflag.ContinueOnError)
		err := BindStringArg(flagSet, vv, &StringArg{
			ArgKey:    "str_key",
			FlagName:  "str_key",
			FlagValue: "default",
			FlagUsage: "usage for str_key",
			Env:       "STR_KEY",
		})
		assert.NilError(t, err)
		err = flagSet.Parse([]string{"--str_key=from_flag"})
		assert.NilError(t, err)
		value := vv.Get("str_key")
		assert.DeepEqual(t, value, "from_flag")
	})
}

func TestBindIntArg(t *testing.T) {
	t.Run("default", func(t *testing.T) {
		vv := viper.New()
		flagSet := pflag.NewFlagSet(t.Name(), pflag.ContinueOnError)
		err := BindIntArg(flagSet, vv, &IntArg{
			ArgKey:    "str_key",
			FlagName:  "str_key",
			FlagValue: 10,
			FlagUsage: "usage for str_key",
			Env:       "STR_KEY",
		})
		assert.NilError(t, err)
		value := vv.GetInt("str_key")
		assert.DeepEqual(t, value, 10)
	})
	t.Run("from_env", func(t *testing.T) {
		t.Setenv("STR_KEY", "1")
		vv := viper.New()
		flagSet := pflag.NewFlagSet(t.Name(), pflag.ContinueOnError)
		err := BindIntArg(flagSet, vv, &IntArg{
			ArgKey:    "str_key",
			FlagName:  "str_key",
			FlagValue: 10,
			FlagUsage: "usage for str_key",
			Env:       "STR_KEY",
		})
		assert.NilError(t, err)
		value := vv.GetInt("str_key")
		assert.DeepEqual(t, value, 1)
	})

	t.Run("from_flag", func(t *testing.T) {
		t.Setenv("STR_KEY", "from_env")
		vv := viper.New()
		flagSet := pflag.NewFlagSet(t.Name(), pflag.ContinueOnError)
		err := BindIntArg(flagSet, vv, &IntArg{
			ArgKey:    "str_key",
			FlagName:  "str_key",
			FlagValue: 12,
			FlagUsage: "usage for str_key",
			Env:       "STR_KEY",
		})
		assert.NilError(t, err)
		err = flagSet.Parse([]string{"--str_key=12"})
		assert.NilError(t, err)
		value := vv.GetInt("str_key")
		assert.DeepEqual(t, value, 12)
	})
}
