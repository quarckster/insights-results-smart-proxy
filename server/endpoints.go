// Copyright 2020 Red Hat, Inc
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package server

import (
	"fmt"
	"net/http"
	"path/filepath"
	"regexp"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	// MainEndpoint returns status ok
	MainEndpoint = ""
	// DeleteOrganizationsEndpoint deletes all {organizations}(comma separated array). DEBUG only
	DeleteOrganizationsEndpoint = "organizations/{organizations}"
	// DeleteClustersEndpoint deletes all {clusters}(comma separated array). DEBUG only
	DeleteClustersEndpoint = "clusters/{clusters}"
	// OrganizationsEndpoint returns all organizations
	OrganizationsEndpoint = "organizations"
	// ReportEndpoint returns report for provided {organization} and {cluster}
	ReportEndpoint = "report/{organization}/{cluster}"
	// LikeRuleEndpoint likes rule with {rule_id} for {cluster} using current user(from auth header)
	LikeRuleEndpoint = "clusters/{cluster}/rules/{rule_id}/like"
	// DislikeRuleEndpoint dislikes rule with {rule_id} for {cluster} using current user(from auth header)
	DislikeRuleEndpoint = "clusters/{cluster}/rules/{rule_id}/dislike"
	// ResetVoteOnRuleEndpoint resets vote on rule with {rule_id} for {cluster} using current user(from auth header)
	ResetVoteOnRuleEndpoint = "clusters/{cluster}/rules/{rule_id}/reset_vote"
	// GetVoteOnRuleEndpoint is an endpoint to get vote on rule. DEBUG only
	GetVoteOnRuleEndpoint = "clusters/{cluster}/rules/{rule_id}/get_vote"
	// RuleEndpoint is an endpoint to create&delete a rule. DEBUG only
	RuleEndpoint = "rules/{rule_id}"
	// RuleErrorKeyEndpoint is for endpoints to create&delete a rule_error_key (DEBUG only)
	// and for endpoint to get a rule
	RuleErrorKeyEndpoint = "rules/{rule_id}/error_keys/{error_key}"
	// RuleGroupsEndpoint is a simple redirect endpoint to the insights-content-service API specified in configruation
	RuleGroupsEndpoint = "groups"
	// ClustersForOrganizationEndpoint returns all clusters for {organization}
	ClustersForOrganizationEndpoint = "organizations/{organization}/clusters"
	// DisableRuleForClusterEndpoint disables a rule for specified cluster
	DisableRuleForClusterEndpoint = "clusters/{cluster}/rules/{rule_id}/disable"
	// EnableRuleForClusterEndpoint re-enables a rule for specified cluster
	EnableRuleForClusterEndpoint = "clusters/{cluster}/rules/{rule_id}/enable"
	// MetricsEndpoint returns prometheus metrics
	MetricsEndpoint = "metrics"
)

func (server *HTTPServer) addDebugEndpointsToRouter(router *mux.Router) {
	apiPrefix := server.Config.APIPrefix
	aggregatorEndpoint := server.ServicesConfig.AggregatorBaseEndpoint

	router.HandleFunc(apiPrefix+OrganizationsEndpoint, server.proxyTo(aggregatorEndpoint)).Methods(http.MethodGet)
	router.HandleFunc(apiPrefix+DeleteOrganizationsEndpoint, server.proxyTo(aggregatorEndpoint)).Methods(http.MethodDelete)
	router.HandleFunc(apiPrefix+DeleteClustersEndpoint, server.proxyTo(aggregatorEndpoint)).Methods(http.MethodDelete)
	router.HandleFunc(apiPrefix+GetVoteOnRuleEndpoint, server.proxyTo(aggregatorEndpoint)).Methods(http.MethodGet)
	router.HandleFunc(apiPrefix+RuleEndpoint, server.proxyTo(aggregatorEndpoint)).Methods(http.MethodPost)
	router.HandleFunc(apiPrefix+RuleErrorKeyEndpoint, server.proxyTo(aggregatorEndpoint)).Methods(http.MethodPost)
	router.HandleFunc(apiPrefix+RuleEndpoint, server.proxyTo(aggregatorEndpoint)).Methods(http.MethodDelete)
	router.HandleFunc(apiPrefix+RuleErrorKeyEndpoint, server.proxyTo(aggregatorEndpoint)).Methods(http.MethodDelete)

	// endpoints for pprof - needed for profiling, ie. usually in debug mode
	router.PathPrefix("/debug/pprof/").Handler(http.DefaultServeMux)
}

func (server *HTTPServer) addEndpointsToRouter(router *mux.Router) {
	apiPrefix := server.Config.APIPrefix
	openAPIURL := apiPrefix + filepath.Base(server.Config.APISpecFile)
	aggregatorEndpoint := server.ServicesConfig.AggregatorBaseEndpoint
	contentServiceEndpoint := server.ServicesConfig.ContentBaseEndpoint

	// it is possible to use special REST API endpoints in debug mode
	if server.Config.Debug {
		server.addDebugEndpointsToRouter(router)
	}

	// common REST API endpoints
	router.HandleFunc(apiPrefix+MainEndpoint, server.mainEndpoint).Methods(http.MethodGet)
	router.HandleFunc(apiPrefix+ReportEndpoint, server.proxyTo(aggregatorEndpoint)).Methods(http.MethodGet, http.MethodOptions)
	router.HandleFunc(apiPrefix+LikeRuleEndpoint, server.proxyTo(aggregatorEndpoint)).Methods(http.MethodPut, http.MethodOptions)
	router.HandleFunc(apiPrefix+DislikeRuleEndpoint, server.proxyTo(aggregatorEndpoint)).Methods(http.MethodPut, http.MethodOptions)
	router.HandleFunc(apiPrefix+ResetVoteOnRuleEndpoint, server.proxyTo(aggregatorEndpoint)).Methods(http.MethodPut, http.MethodOptions)
	router.HandleFunc(apiPrefix+ClustersForOrganizationEndpoint, server.proxyTo(aggregatorEndpoint)).Methods(http.MethodGet)
	router.HandleFunc(apiPrefix+DisableRuleForClusterEndpoint, server.proxyTo(aggregatorEndpoint)).Methods(http.MethodPut, http.MethodOptions)
	router.HandleFunc(apiPrefix+EnableRuleForClusterEndpoint, server.proxyTo(aggregatorEndpoint)).Methods(http.MethodPut, http.MethodOptions)
	router.HandleFunc(apiPrefix+RuleGroupsEndpoint, server.proxyTo(contentServiceEndpoint)).Methods(http.MethodGet, http.MethodOptions)
	router.HandleFunc(apiPrefix+RuleErrorKeyEndpoint, server.proxyTo(aggregatorEndpoint)).Methods(http.MethodGet)

	// Prometheus metrics
	router.Handle(apiPrefix+MetricsEndpoint, promhttp.Handler()).Methods(http.MethodGet)

	// OpenAPI specs
	router.HandleFunc(openAPIURL, server.serveAPISpecFile).Methods(http.MethodGet)
}

// MakeURLToEndpoint creates URL to endpoint, use constants from file endpoints.go
func MakeURLToEndpoint(apiPrefix, endpoint string, args ...interface{}) string {
	re := regexp.MustCompile(`\{[a-zA-Z_0-9]+\}`)
	endpoint = re.ReplaceAllString(endpoint, "%v")
	return apiPrefix + fmt.Sprintf(endpoint, args...)
}
