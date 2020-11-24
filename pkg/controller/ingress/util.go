package ingress

import (
	"fmt"

	netv1 "k8s.io/api/networking/v1"
)

func validateClass(ing *netv1.Ingress) error {
	// Check Ingress class annotation, then Ingress Class spec
	class, exists := ing.Annotations["kubernetes.io/ingress.class"]
	if exists {
		if class == ControllerName {
			return nil
		} else {
			return fmt.Errorf("unexpected ingress class %s", class)
		}
	}
	if ing.Spec.IngressClassName != nil {
		class := *ing.Spec.IngressClassName
		if class == ControllerName {
			return nil
		} else {
			return fmt.Errorf("unexpected ingress class %s", class)
		}
	}
	return fmt.Errorf("no ingress class found")
}

func findRuleForHost(host string, ingress *netv1.Ingress) (upstream, error) {
	for _, rule := range ingress.Spec.Rules {
		if rule.Host == host {
			return upstream{
				serviceName:      rule.HTTP.Paths[0].Backend.Service.Name,
				serviceNamespace: ingress.ObjectMeta.Namespace,
				port:             rule.HTTP.Paths[0].Backend.Service.Port.Number,
			}, nil
		}
	}
	return upstream{}, fmt.Errorf("host %s not found in rules", host)
}
