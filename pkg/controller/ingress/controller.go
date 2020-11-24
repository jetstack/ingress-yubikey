package ingress

import (
	"context"
	"sync"

	"github.com/go-logr/logr"
	netv1 "k8s.io/api/networking/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const ControllerName = "yubikey-ingress"

// Controller is an Ingress controller
type Controller struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme

	// A reference to all hostnames specified in an ingress
	hostnames map[types.NamespacedName][]string

	// A reference of hostnames to upstreams (for now, just the first rule is used and paths are ignored)
	rules map[string]upstream

	// DefaultBackend address to proxy to
	DefaultBackend string

	sync.RWMutex
}

type upstream struct {
	serviceName      string
	serviceNamespace string
	port             int32
}

// SetupWithManager adds the Ingress controller to a controller-manager
func (c *Controller) SetupWithManager(mgr ctrl.Manager) error {
	if c.hostnames == nil {
		c.hostnames = make(map[types.NamespacedName][]string)
	}
	if c.rules == nil {
		c.rules = make(map[string]upstream)
	}
	c.Log.Info("Starting " + ControllerName + "controller")
	return ctrl.NewControllerManagedBy(mgr).
		For(&netv1.Ingress{}).Complete(c)
}

func (c *Controller) Reconcile(ctx context.Context, r ctrl.Request) (ctrl.Result, error) {
	c.Log.Info("saw ingress event", "name", r.NamespacedName.String())
	ing := &netv1.Ingress{}
	err := c.Client.Get(ctx, r.NamespacedName, ing)
	if apierrors.IsNotFound(err) {
		// handle deletion
		return c.delete(r)
	}
	if err := validateClass(ing); err != nil {
		// Not for us
		c.Log.Info("ignoring invalid ingress", "name", r.NamespacedName, "reason", err.Error())
		return ctrl.Result{}, nil
	}
	return c.createOrUpdate(r, ing)
}

func (c *Controller) createOrUpdate(r ctrl.Request, ing *netv1.Ingress) (reconcile.Result, error) {
	// Parse rules
	if len(ing.Spec.TLS) == 0 {
		c.Log.Info("no TLS block found", "name", r.NamespacedName)
		return ctrl.Result{}, nil
	}
	hostnames := []string{}
	c.Lock()
	defer c.Unlock()
	for _, tls := range ing.Spec.TLS {
		for _, host := range tls.Hosts {
			upstream, err := findRuleForHost(host, ing)
			if err != nil {
				c.Log.Error(err, "no rules for host found in Ingress spec", "name", r.NamespacedName)
				return ctrl.Result{}, nil
			}
			hostnames = append(hostnames, host)
			c.rules[host] = upstream
		}
	}
	c.hostnames[r.NamespacedName] = hostnames
	return ctrl.Result{}, nil
}

// delete handles cleanup when an Ingress is deleted
func (c *Controller) delete(r ctrl.Request) (reconcile.Result, error) {
	c.Lock()
	hostnames, exists := c.hostnames[r.NamespacedName]
	if exists {
		for _, host := range hostnames {
			delete(c.rules, host)
		}
	}
	delete(c.hostnames, r.NamespacedName)
	c.Unlock()
	c.Log.Info("ingress deleted", "name", r.NamespacedName.String())
	return ctrl.Result{}, nil
}
