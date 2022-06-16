// This file is safe to edit. Once it exists it will not be overwritten

package restapi

import (
	"crypto/tls"
	"net/http"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime"
	"github.com/go-openapi/runtime/middleware"

	"github.com/openclarity/apiclarity/api/server/restapi/operations"
)

//go:generate swagger generate server --target ../../server --name APIClarityAPIs --spec ../../swagger.yaml --principal interface{}

func configureFlags(api *operations.APIClarityAPIsAPI) {
	// api.CommandLineOptionsGroups = []swag.CommandLineOptionsGroup{ ... }
}

func configureAPI(api *operations.APIClarityAPIsAPI) http.Handler {
	// configure the api here
	api.ServeError = errors.ServeError

	// Set your custom logger if needed. Default one is log.Printf
	// Expected interface func(string, ...interface{})
	//
	// Example:
	// api.Logger = log.Printf

	api.UseSwaggerUI()
	// To continue using redoc as your UI, uncomment the following line
	// api.UseRedoc()

	api.JSONConsumer = runtime.JSONConsumer()

	api.JSONProducer = runtime.JSONProducer()

	if api.DeleteAPIInventoryAPIIDSpecsProvidedSpecHandler == nil {
		api.DeleteAPIInventoryAPIIDSpecsProvidedSpecHandler = operations.DeleteAPIInventoryAPIIDSpecsProvidedSpecHandlerFunc(func(params operations.DeleteAPIInventoryAPIIDSpecsProvidedSpecParams) middleware.Responder {
			return middleware.NotImplemented("operation operations.DeleteAPIInventoryAPIIDSpecsProvidedSpec has not yet been implemented")
		})
	}
	if api.DeleteAPIInventoryAPIIDSpecsReconstructedSpecHandler == nil {
		api.DeleteAPIInventoryAPIIDSpecsReconstructedSpecHandler = operations.DeleteAPIInventoryAPIIDSpecsReconstructedSpecHandlerFunc(func(params operations.DeleteAPIInventoryAPIIDSpecsReconstructedSpecParams) middleware.Responder {
			return middleware.NotImplemented("operation operations.DeleteAPIInventoryAPIIDSpecsReconstructedSpec has not yet been implemented")
		})
	}
	if api.GetAPIEventsHandler == nil {
		api.GetAPIEventsHandler = operations.GetAPIEventsHandlerFunc(func(params operations.GetAPIEventsParams) middleware.Responder {
			return middleware.NotImplemented("operation operations.GetAPIEvents has not yet been implemented")
		})
	}
	if api.GetAPIEventsEventIDHandler == nil {
		api.GetAPIEventsEventIDHandler = operations.GetAPIEventsEventIDHandlerFunc(func(params operations.GetAPIEventsEventIDParams) middleware.Responder {
			return middleware.NotImplemented("operation operations.GetAPIEventsEventID has not yet been implemented")
		})
	}
	if api.GetAPIEventsEventIDProvidedSpecDiffHandler == nil {
		api.GetAPIEventsEventIDProvidedSpecDiffHandler = operations.GetAPIEventsEventIDProvidedSpecDiffHandlerFunc(func(params operations.GetAPIEventsEventIDProvidedSpecDiffParams) middleware.Responder {
			return middleware.NotImplemented("operation operations.GetAPIEventsEventIDProvidedSpecDiff has not yet been implemented")
		})
	}
	if api.GetAPIEventsEventIDReconstructedSpecDiffHandler == nil {
		api.GetAPIEventsEventIDReconstructedSpecDiffHandler = operations.GetAPIEventsEventIDReconstructedSpecDiffHandlerFunc(func(params operations.GetAPIEventsEventIDReconstructedSpecDiffParams) middleware.Responder {
			return middleware.NotImplemented("operation operations.GetAPIEventsEventIDReconstructedSpecDiff has not yet been implemented")
		})
	}
	if api.GetAPIInventoryHandler == nil {
		api.GetAPIInventoryHandler = operations.GetAPIInventoryHandlerFunc(func(params operations.GetAPIInventoryParams) middleware.Responder {
			return middleware.NotImplemented("operation operations.GetAPIInventory has not yet been implemented")
		})
	}
	if api.GetAPIInventoryAPIIDAPIInfoHandler == nil {
		api.GetAPIInventoryAPIIDAPIInfoHandler = operations.GetAPIInventoryAPIIDAPIInfoHandlerFunc(func(params operations.GetAPIInventoryAPIIDAPIInfoParams) middleware.Responder {
			return middleware.NotImplemented("operation operations.GetAPIInventoryAPIIDAPIInfo has not yet been implemented")
		})
	}
	if api.GetAPIInventoryAPIIDHostPortHandler == nil {
		api.GetAPIInventoryAPIIDHostPortHandler = operations.GetAPIInventoryAPIIDHostPortHandlerFunc(func(params operations.GetAPIInventoryAPIIDHostPortParams) middleware.Responder {
			return middleware.NotImplemented("operation operations.GetAPIInventoryAPIIDHostPort has not yet been implemented")
		})
	}
	if api.GetAPIInventoryAPIIDProvidedSwaggerJSONHandler == nil {
		api.GetAPIInventoryAPIIDProvidedSwaggerJSONHandler = operations.GetAPIInventoryAPIIDProvidedSwaggerJSONHandlerFunc(func(params operations.GetAPIInventoryAPIIDProvidedSwaggerJSONParams) middleware.Responder {
			return middleware.NotImplemented("operation operations.GetAPIInventoryAPIIDProvidedSwaggerJSON has not yet been implemented")
		})
	}
	if api.GetAPIInventoryAPIIDReconstructedSwaggerJSONHandler == nil {
		api.GetAPIInventoryAPIIDReconstructedSwaggerJSONHandler = operations.GetAPIInventoryAPIIDReconstructedSwaggerJSONHandlerFunc(func(params operations.GetAPIInventoryAPIIDReconstructedSwaggerJSONParams) middleware.Responder {
			return middleware.NotImplemented("operation operations.GetAPIInventoryAPIIDReconstructedSwaggerJSON has not yet been implemented")
		})
	}
	if api.GetAPIInventoryAPIIDSpecsHandler == nil {
		api.GetAPIInventoryAPIIDSpecsHandler = operations.GetAPIInventoryAPIIDSpecsHandlerFunc(func(params operations.GetAPIInventoryAPIIDSpecsParams) middleware.Responder {
			return middleware.NotImplemented("operation operations.GetAPIInventoryAPIIDSpecs has not yet been implemented")
		})
	}
	if api.GetAPIInventoryAPIIDSuggestedReviewHandler == nil {
		api.GetAPIInventoryAPIIDSuggestedReviewHandler = operations.GetAPIInventoryAPIIDSuggestedReviewHandlerFunc(func(params operations.GetAPIInventoryAPIIDSuggestedReviewParams) middleware.Responder {
			return middleware.NotImplemented("operation operations.GetAPIInventoryAPIIDSuggestedReview has not yet been implemented")
		})
	}
	if api.GetAPIUsageHitCountHandler == nil {
		api.GetAPIUsageHitCountHandler = operations.GetAPIUsageHitCountHandlerFunc(func(params operations.GetAPIUsageHitCountParams) middleware.Responder {
			return middleware.NotImplemented("operation operations.GetAPIUsageHitCount has not yet been implemented")
		})
	}
	if api.GetDashboardAPIUsageHandler == nil {
		api.GetDashboardAPIUsageHandler = operations.GetDashboardAPIUsageHandlerFunc(func(params operations.GetDashboardAPIUsageParams) middleware.Responder {
			return middleware.NotImplemented("operation operations.GetDashboardAPIUsage has not yet been implemented")
		})
	}
	if api.GetDashboardAPIUsageLatestDiffsHandler == nil {
		api.GetDashboardAPIUsageLatestDiffsHandler = operations.GetDashboardAPIUsageLatestDiffsHandlerFunc(func(params operations.GetDashboardAPIUsageLatestDiffsParams) middleware.Responder {
			return middleware.NotImplemented("operation operations.GetDashboardAPIUsageLatestDiffs has not yet been implemented")
		})
	}
	if api.GetDashboardAPIUsageMostUsedHandler == nil {
		api.GetDashboardAPIUsageMostUsedHandler = operations.GetDashboardAPIUsageMostUsedHandlerFunc(func(params operations.GetDashboardAPIUsageMostUsedParams) middleware.Responder {
			return middleware.NotImplemented("operation operations.GetDashboardAPIUsageMostUsed has not yet been implemented")
		})
	}
	if api.PostAPIInventoryHandler == nil {
		api.PostAPIInventoryHandler = operations.PostAPIInventoryHandlerFunc(func(params operations.PostAPIInventoryParams) middleware.Responder {
			return middleware.NotImplemented("operation operations.PostAPIInventory has not yet been implemented")
		})
	}
	if api.PostAPIInventoryReviewIDApprovedReviewHandler == nil {
		api.PostAPIInventoryReviewIDApprovedReviewHandler = operations.PostAPIInventoryReviewIDApprovedReviewHandlerFunc(func(params operations.PostAPIInventoryReviewIDApprovedReviewParams) middleware.Responder {
			return middleware.NotImplemented("operation operations.PostAPIInventoryReviewIDApprovedReview has not yet been implemented")
		})
	}
	if api.PutAPIInventoryAPIIDSpecsProvidedSpecHandler == nil {
		api.PutAPIInventoryAPIIDSpecsProvidedSpecHandler = operations.PutAPIInventoryAPIIDSpecsProvidedSpecHandlerFunc(func(params operations.PutAPIInventoryAPIIDSpecsProvidedSpecParams) middleware.Responder {
			return middleware.NotImplemented("operation operations.PutAPIInventoryAPIIDSpecsProvidedSpec has not yet been implemented")
		})
	}

	api.PreServerShutdown = func() {}

	api.ServerShutdown = func() {}

	return setupGlobalMiddleware(api.Serve(setupMiddlewares))
}

// The TLS configuration before HTTPS server starts.
func configureTLS(tlsConfig *tls.Config) {
	// Make all necessary changes to the TLS configuration here.
}

// As soon as server is initialized but not run yet, this function will be called.
// If you need to modify a config, store server instance to stop it individually later, this is the place.
// This function can be called multiple times, depending on the number of serving schemes.
// scheme value will be set accordingly: "http", "https" or "unix".
func configureServer(s *http.Server, scheme, addr string) {
}

// The middleware configuration is for the handler executors. These do not apply to the swagger.json document.
// The middleware executes after routing but before authentication, binding and validation.
func setupMiddlewares(handler http.Handler) http.Handler {
	return handler
}

// The middleware configuration happens before anything, this middleware also applies to serving the swagger.json document.
// So this is a good place to plug in a panic handling middleware, logging and metrics.
func setupGlobalMiddleware(handler http.Handler) http.Handler {
	return handler
}
