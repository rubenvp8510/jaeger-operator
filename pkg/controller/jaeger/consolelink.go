package jaeger

import (
	"context"

	osconsolev1 "github.com/openshift/api/console/v1"
	osroutev1 "github.com/openshift/api/route/v1"
	log "github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/global"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	v1 "github.com/jaegertracing/jaeger-operator/pkg/apis/jaegertracing/v1"
	"github.com/jaegertracing/jaeger-operator/pkg/consolelink"
	"github.com/jaegertracing/jaeger-operator/pkg/inventory"
	"github.com/jaegertracing/jaeger-operator/pkg/tracing"
)

func (r *ReconcileJaeger) updateHref(ctx context.Context,
	jaeger v1.Jaeger, links []osconsolev1.ConsoleLink) []osconsolev1.ConsoleLink {

	for i, cl := range links {
		routeName := cl.Annotations[consolelink.RouteAnnotation]
		route := osroutev1.Route{}
		if err := r.rClient.Get(ctx, types.NamespacedName{Name: routeName, Namespace: cl.Namespace}, &route); err != nil {
			jaeger.Logger().WithError(err).WithFields(log.Fields{
				"consoleLink": cl.Name,
				"namespace":   cl.Namespace,
			}).Warn("updating console link href")
		}
		consolelink.UpdateHref(&links[i], route)
	}
	return links
}

func (r *ReconcileJaeger) applyConsoleLink(ctx context.Context, jaeger v1.Jaeger, desired []osconsolev1.ConsoleLink) error {
	tracer := global.TraceProvider().GetTracer(v1.ReconciliationTracer)
	ctx, span := tracer.Start(ctx, "applyConsoleLink")
	defer span.End()

	opts := []client.ListOption{
		client.InNamespace(jaeger.Namespace),
		client.MatchingLabels(map[string]string{
			"app.kubernetes.io/instance":   jaeger.Name,
			"app.kubernetes.io/managed-by": "jaeger-operator",
		}),
	}
	list := &osconsolev1.ConsoleLinkList{}
	if err := r.rClient.List(ctx, list, opts...); err != nil {
		return tracing.HandleError(err, span)
	}

	desiredWithHref := r.updateHref(ctx, jaeger, desired)
	inv := inventory.ForConsoleLinks(list.Items, desiredWithHref)
	for _, d := range inv.Create {
		jaeger.Logger().WithFields(log.Fields{
			"consoleLink": d.Name,
			"namespace":   d.Namespace,
		}).Debug("creating console link")
		if err := r.client.Create(ctx, &d); err != nil {
			return tracing.HandleError(err, span)
		}
	}

	for _, d := range inv.Update {
		jaeger.Logger().WithFields(log.Fields{
			"consoleLink": d.Name,
			"namespace":   d.Namespace,
		}).Debug("updating console link")
		if err := r.client.Update(ctx, &d); err != nil {
			return tracing.HandleError(err, span)
		}
	}

	for _, d := range inv.Delete {
		jaeger.Logger().WithFields(log.Fields{
			"consoleLink": d.Name,
			"namespace":   d.Namespace,
		}).Debug("deleting console link")
		if err := r.client.Delete(ctx, &d); err != nil {
			return tracing.HandleError(err, span)
		}
	}

	return nil
}
