package main

import (
	"context"

	"github.com/crossplane/crossplane-runtime/pkg/errors"
	"github.com/crossplane/crossplane-runtime/pkg/logging"

	fnv1beta1 "github.com/crossplane/function-sdk-go/proto/v1beta1"
	"github.com/crossplane/function-sdk-go/request"
	"github.com/crossplane/function-sdk-go/response"

	"github.com/crossplane/function-composition-mux/input/v1beta1"

	corev1 "k8s.io/api/core/v1"
)

// Function returns whatever response you ask it to.
type Function struct {
	fnv1beta1.UnimplementedFunctionRunnerServiceServer

	log logging.Logger
}

// RunFunction runs the Function.
func (f *Function) RunFunction(_ context.Context, req *fnv1beta1.RunFunctionRequest) (*fnv1beta1.RunFunctionResponse, error) {
	f.log.Info("Running Function", "tag", req.GetMeta().GetTag())

	// This creates a new response to the supplied request. Note that Functions
	// are run in a pipeline! Other Functions may have run before this one. If
	// they did, response.To will copy their desired state from req to rsp. Be
	// sure to pass through any desired state your Function is not concerned
	// with unmodified.
	rsp := response.To(req, response.DefaultTTL)

	// Input is supplied by the author of a Composition when they choose to run
	// your Function. Input is arbitrary, except that it must be a KRM-like
	// object. Supporting input is also optional - if you don't need to you can
	// delete this, and delete the input directory.
	in := &v1beta1.ProposedComposition{}
	if err := request.GetInput(req, in); err != nil {
		response.Fatal(rsp, errors.Wrapf(err, "cannot get Function input from %T", req))
		return rsp, nil
	}

	// Crossplane itself sets the Resource field of a request so we don't have to.
	// TODO: update this to accept the array of resources.
	observed, err := request.GetObservedCompositeResource(req)

	if err != nil {
		// This isn't a 500-style error; we tried to serve the request but something went wrong so we stop the pipeline.
		response.Fatal(rsp, errors.Wrapf(err, "cannot get observed composed resources from %T", req))
		return rsp, nil
	}

	desired, err := request.GetDesiredCompositeResource(req)

	if err != nil {
		response.Fatal(rsp, errors.Wrapf(err, "cannot get desired composed resources from %T", req))
		return rsp, nil
	}

	// If the new Composition is the same as the current compositionRef, we should short-circuit/no-op.
	if compositionRef := observed.Resource.GetCompositionReference(); compositionRef != nil && compositionRef.Name != in.Name {
		newCompositionRef := corev1.ObjectReference{
			Name: in.Name,
		}

		// TODO: we should verify that the new Composition actually exists in the current context, and that its compositeTypeRef matches.
		desired.Resource.SetCompositionReference(&newCompositionRef)

		response.SetDesiredCompositeResource(rsp, desired)
	}

	response.Normalf(rsp, "I switched my active composition reference from %q to %q", observed.Resource.GetCompositionReference().Name, in.Name)

	return rsp, nil
}
