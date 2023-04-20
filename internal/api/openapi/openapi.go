// Copyright 2022 Harness Inc. All rights reserved.
// Use of this source code is governed by the Polyform Free Trial License
// that can be found in the LICENSE.md file for this repository.

package openapi

import (
	"github.com/harness/gitness/version"

	"github.com/swaggest/openapi-go/openapi3"
)

type (
	paginationRequest struct {
		Page int `query:"page"     default:"1"`
		Size int `query:"limit"    default:"30"`
	}
)

// Generate is a helper function that constructs the
// openapi specification object, which can be marshaled
// to json or yaml, as needed.
func Generate() *openapi3.Spec {
	reflector := openapi3.Reflector{}
	reflector.Spec = &openapi3.Spec{Openapi: "3.0.0"}
	reflector.Spec.Info.
		WithTitle("API Specification").
		WithVersion(version.Version.String())
	reflector.Spec.Servers = []openapi3.Server{{
		URL: "/api/v1/",
	}}

	//
	// register endpoints
	//

	buildAccount(&reflector)
	buildUser(&reflector)
	buildAdmin(&reflector)
	buildPrincipals(&reflector)
	spaceOperations(&reflector)
	repoOperations(&reflector)
	resourceOperations(&reflector)
	pullReqOperations(&reflector)
	webhookOperations(&reflector)

	//
	// define security scheme
	//

	scheme := openapi3.SecuritySchemeOrRef{
		SecurityScheme: &openapi3.SecurityScheme{
			HTTPSecurityScheme: &openapi3.HTTPSecurityScheme{
				Scheme: "bearerAuth",
				Bearer: &openapi3.Bearer{},
			},
		},
	}
	security := openapi3.ComponentsSecuritySchemes{}
	security.WithMapOfSecuritySchemeOrRefValuesItem("bearerAuth", scheme)
	reflector.Spec.Components.WithSecuritySchemes(security)

	//
	// enforce security scheme globally
	//

	reflector.Spec.WithSecurity(map[string][]string{
		"bearerAuth": {},
	})

	return reflector.Spec
}