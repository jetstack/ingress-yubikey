module github.com/jakexks/ingress-yubikey

go 1.15

require (
	github.com/go-logr/logr v0.3.0
	github.com/go-piv/piv-go v1.6.1-0.20200925162556-47b85fc97472
	github.com/pkg/errors v0.9.1
	github.com/spf13/cobra v1.1.1
	github.com/spf13/viper v1.7.1
	k8s.io/api v0.19.2
	k8s.io/apimachinery v0.19.2
	k8s.io/client-go v0.19.2
	sigs.k8s.io/controller-runtime v0.7.0-alpha.6
)
