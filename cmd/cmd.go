package cmd

import (
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	clientgoscheme.AddToScheme(scheme)
	cobra.OnInitialize(flagsFromEnv)
}

// flagsFromEnv allows flags to be set from environment variables.
func flagsFromEnv() {
	viper.SetEnvPrefix("ingress_yubikey")
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv()
}
