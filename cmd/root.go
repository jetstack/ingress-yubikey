package cmd

import (
	"github.com/jakexks/ingress-yubikey/pkg/controller/ingress"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

var (
	rootCmd = &cobra.Command{
		Use:   "ingress-yubikey",
		Short: "A PoC ingress controller that terminates TLS without direct access to the private key",
		Long:  "A PoC ingress controller that terminates TLS without direct access to the private key.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return root()
		},
	}
)

func init() {
	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))
	rootCmd.PersistentFlags().String("metrics-addr", ":8080", "The address the metric endpoint binds to.")
	rootCmd.PersistentFlags().Bool("enable-leader-election", false, "Enable leader election for controller manager.")
	rootCmd.PersistentFlags().String("smartcard-pin", "123456", "The PIN to unlock the smartcard")
	if err := viper.BindPFlags(rootCmd.PersistentFlags()); err != nil {
		// Flags collide or other viper weirdness
		setupLog.Error(err, "couldn't bind viper config to flags")
		os.Exit(1)
	}
}

// Execute is the entrypoint from main
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

// root sets up and starts the controller manager
func root() error {
	ctx := ctrl.SetupSignalHandler()
	opts := ctrl.Options{
		Scheme:             scheme,
		MetricsBindAddress: viper.GetString("metrics-addr"),
	}
	if viper.GetBool("enable-leader-election") {
		hostname, err := os.Hostname()
		if err != nil {
			setupLog.Error(err, "couldn't get hostname")
			return err
		}
		opts.LeaderElection = true
		opts.LeaderElectionID = hostname
	}
	mgr, err := ctrl.NewManager(config.GetConfigOrDie(), opts)
	if err != nil {
		setupLog.Error(err, "couldn't create controller manager")
		return err
	}

	// Start the ingress controller
	ing := &ingress.Controller{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("yubikey-ingress"),
		Scheme: mgr.GetScheme(),
	}
	if err := ing.SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "yubikey-ingress")
	}
	go ing.Listen(ctx)
	if err := mgr.Start(ctx); err != nil {
		setupLog.Error(err, "unable to start controller-manager", "controller", "yubikey-ingress")
	}
	<-ctx.Done()
	return nil
}
