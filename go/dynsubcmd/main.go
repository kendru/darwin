package main

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type Operation interface {
	Invoke() error
}

type PersonOp struct {
	Name string
	Age  int
}

func (o *PersonOp) Invoke() error {
	fmt.Printf("This is a Person operation:\nName\t= %s\nAge\t= %d\n", o.Name, o.Age)
	return nil
}

type CompanyOp struct {
	Name  string
	State string
}

func (o *CompanyOp) Invoke() error {
	fmt.Printf("This is a Company operation:\nName\t= %s\nState\t= %s\n", o.Name, o.State)
	return nil
}

var opRegistry = make(map[string]Operation)

func RegisterOperation(name string, op Operation) {
	opRegistry[name] = op
}

var (
	cfgFile string
	rootCmd = &cobra.Command{
		Use:   "dynsubcmd",
		Short: "A test of using Viper and Cobra to expose dynamic subcommands",
	}
)

func init() {
	RegisterOperation("person", &PersonOp{})
	RegisterOperation("company", &CompanyOp{})

	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file")
	for cmdName, op := range opRegistry {
		cmdName := cmdName
		op := op
		cmd := &cobra.Command{
			Use:   cmdName,
			Short: fmt.Sprintf("A test command for %s", cmdName),
			Run: func(cmd *cobra.Command, args []string) {
				v := reflect.ValueOf(op).Elem().Type()
				for i := 0; i < v.NumField(); i++ {
					fieldName := strings.ToLower(v.Field(i).Name)
					viper.BindPFlag(fieldName, cmd.Flags().Lookup(fieldName))
				}

				newOp := reflect.New(v).Interface().(Operation)
				viper.Unmarshal(newOp)

				if err := newOp.Invoke(); err != nil {
					panic(err)
				}
			},
		}

		v := reflect.ValueOf(op).Elem().Type()
		for i := 0; i < v.NumField(); i++ {
			field := v.Field(i)
			fieldName := strings.ToLower(v.Field(i).Name)
			switch field.Type.Kind() {
			case reflect.String:
				cmd.Flags().String(fieldName, "", fmt.Sprintf("String Arg: %s", fieldName))
			case reflect.Int:
				cmd.Flags().Int(fieldName, 0, fmt.Sprintf("Int Arg: %s", fieldName))
			}
		}

		rootCmd.AddCommand(cmd)
	}
}

func main() {
	rootCmd.Execute()
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := homedir.Dir()
		if err != nil {
			panic(err)
		}

		viper.AddConfigPath(home)
		viper.SetConfigName(".dynsubcmd")
	}

	viper.SetEnvPrefix("DEREF")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
