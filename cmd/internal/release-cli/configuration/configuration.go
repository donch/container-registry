package configuration

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type Envs struct {
	K8sEnv     K8s     `mapstructure:"k8s"`
	CngEnv     Cng     `mapstructure:"cng"`
	ChartsEnv  Charts  `mapstructure:"charts"`
	OmnibusEnv Omnibus `mapstructure:"omnibus"`
	GdkEnv     Gdk     `mapstructure:"gdk"`
}

type Cng struct {
	ProjectID string `mapstructure:"id"`
	Ref       string `mapstructure:"ref"`
}

type Charts struct {
	ProjectID string `mapstructure:"id"`
	Ref       string `mapstructure:"ref"`
}

type Omnibus struct {
	ProjectID string `mapstructure:"id"`
	Ref       string `mapstructure:"ref"`
}

type K8s struct {
	Stages []Stage `mapstructure:"stages"`
}

type Stage struct {
	StageName     string `mapstructure:"name"`
	ProjectID     string `mapstructure:"id"`
	Ref           string `mapstructure:"ref"`
	BranchName    string `mapstructure:"branch_name"`
	CommitMessage string `mapstructure:"commit_message"`
	MRTitle       string `mapstructure:"mr_title"`
	Paths         []Path `mapstructure:"paths"`
}

type Gdk struct {
	ProjectID     string `mapstructure:"id"`
	Ref           string `mapstructure:"ref"`
	BranchName    string `mapstructure:"branch_name"`
	CommitMessage string `mapstructure:"commit_message"`
	MRTitle       string `mapstructure:"mr_title"`
	Paths         []Path `mapstructure:"paths"`
}

type Path struct {
	Filename string `mapstructure:"filename"`
}

type Issue struct {
	Title     string `mapstructure:"title"`
	Template  string `mapstructure:"template"`
	ProjectID string `mapstructure:"id"`
	Ref       string `mapstructure:"ref"`
}

type Global struct {
	Version   string
	AuthToken string
}

var (
	envs   Envs
	issue  Issue
	global Global
)

func SetConfig(cfg string) {
	viper.SetConfigFile(cfg)
}

func InitConfig(tag string, token string) {
	global = Global{
		Version:   tag,
		AuthToken: token,
	}

	home, err := os.UserHomeDir()
	cobra.CheckErr(err)

	viper.AddConfigPath(".")
	viper.AddConfigPath(home)
	viper.SetConfigType("yaml")
	viper.SetConfigName(".config")

	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}

	err = viper.UnmarshalKey("environments", &envs)
	if err != nil {
		panic(fmt.Errorf("Unable to decode config for environments: %s \n", err))
	}

	err = viper.UnmarshalKey("issue", &issue)
	if err != nil {
		panic(fmt.Errorf("Unable to decode config for release issue: %s \n", err))
	}
}

func GetK8sEnvConfig() K8s {
	return envs.K8sEnv
}

func GetGDKEnvConfig() Gdk {
	return envs.GdkEnv
}

func GetChartsEnvConfig() Charts {
	return envs.ChartsEnv
}

func GetOmnibusEnvConfig() Omnibus {
	return envs.OmnibusEnv
}

func GetCNGEnvConfig() Cng {
	return envs.CngEnv
}

func GetIssueConfig() Issue {
	return issue
}

func GetVersion() string {
	return global.Version
}

func GetAuthToken() string {
	return global.AuthToken
}
