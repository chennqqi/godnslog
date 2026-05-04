package openapi

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/getkin/kin-openapi/openapi3"
)

// Importer handles OpenAPI spec imports for OAST payload injection
type Importer struct {
	apiURL    string
	apiKey    string
	caseID    string
	template  string
	expiresIn string
}

// NewImporter creates a new OpenAPI importer
func NewImporter(apiURL, apiKey, caseID, template, expiresIn string) *Importer {
	return &Importer{
		apiURL:    apiURL,
		apiKey:    apiKey,
		caseID:    caseID,
		template:  template,
		expiresIn: expiresIn,
	}
}

// ImportSpec imports an OpenAPI spec and generates OAST payloads
func (i *Importer) ImportSpec(specURL string) (*ImportResult, error) {
	// Load OpenAPI spec
	loader := openapi3.NewLoader()
	u, err := url.Parse(specURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse spec URL: %w", err)
	}
	doc, err := loader.LoadFromURI(u)
	if err != nil {
		return nil, fmt.Errorf("failed to load OpenAPI spec: %w", err)
	}

	result := &ImportResult{
		SpecURL:   specURL,
		CaseID:    i.caseID,
		Payloads:  []GeneratedPayload{},
		Endpoints: []Endpoint{},
	}

	// Iterate through all paths
	for path, pathItem := range doc.Paths.Map() {
		if pathItem == nil {
			continue
		}

		// Check each operation (GET, POST, PUT, DELETE, etc.)
		operations := map[string]*openapi3.Operation{
			"get":    pathItem.Get,
			"post":   pathItem.Post,
			"put":    pathItem.Put,
			"delete": pathItem.Delete,
			"patch":  pathItem.Patch,
		}

		for method, operation := range operations {
			if operation == nil {
				continue
			}

			endpoint := Endpoint{
				Path:   path,
				Method: strings.ToUpper(method),
			}

			// Extract parameters
			if operation.Parameters != nil {
				for _, paramRef := range operation.Parameters {
					if paramRef == nil {
						continue
					}
					param := paramRef.Value
					if param == nil {
						continue
					}

					// Generate payload for each parameter
					payload := i.generatePayloadForParameter(param, path, method)
					if payload != nil {
						result.Payloads = append(result.Payloads, *payload)
						endpoint.InjectionPoints = append(endpoint.InjectionPoints, param.Name)
					}
				}
			}

			// Check request body
			if operation.RequestBody != nil {
				requestBody := operation.RequestBody.Value
				if requestBody != nil && requestBody.Content != nil {
					for contentType, mediaType := range requestBody.Content {
						if mediaType.Schema != nil {
							schema := mediaType.Schema.Value
							if schema != nil {
								// Generate payload for schema properties
								payloads := i.generatePayloadsForSchema(schema, path, method, "body", contentType)
								result.Payloads = append(result.Payloads, payloads...)
								endpoint.InjectionPoints = append(endpoint.InjectionPoints, "body")
							}
						}
					}
				}
			}

			if len(endpoint.InjectionPoints) > 0 {
				result.Endpoints = append(result.Endpoints, endpoint)
			}
		}
	}

	return result, nil
}

// generatePayloadForParameter generates an OAST payload for a parameter
func (i *Importer) generatePayloadForParameter(param *openapi3.Parameter, path, method string) *GeneratedPayload {
	if param == nil {
		return nil
	}

	// Skip non-injectable parameters
	if param.In != "query" && param.In != "header" && param.In != "path" {
		return nil
	}

	// Generate payload
	token := generateRandomToken()
	rendered := strings.ReplaceAll(i.template, "{{.Token}}", token)

	return &GeneratedPayload{
		Token:           token,
		Location:        param.In,
		ParameterName:   param.Name,
		Endpoint:        fmt.Sprintf("%s %s", method, path),
		RenderedPayload: rendered,
		Description:     fmt.Sprintf("Inject into %s parameter '%s'", param.In, param.Name),
	}
}

// generatePayloadsForSchema generates OAST payloads for schema properties
func (i *Importer) generatePayloadsForSchema(schema *openapi3.Schema, path, method, location, contentType string) []GeneratedPayload {
	var payloads []GeneratedPayload

	if schema == nil || schema.Properties == nil {
		return payloads
	}

	for propName, propRef := range schema.Properties {
		if propRef == nil {
			continue
		}
		prop := propRef.Value
		if prop == nil {
			continue
		}

		// Check if property is a string type suitable for injection
		if prop.Type == nil || !prop.Type.Is("string") {
			continue
		}

		// Generate payload
		token := generateRandomToken()
		rendered := strings.ReplaceAll(i.template, "{{.Token}}", token)

		payload := GeneratedPayload{
			Token:           token,
			Location:        location,
			ParameterName:   propName,
			Endpoint:        fmt.Sprintf("%s %s", method, path),
			RenderedPayload: rendered,
			ContentType:     contentType,
			Description:     fmt.Sprintf("Inject into %s property '%s'", location, propName),
		}

		payloads = append(payloads, payload)
	}

	return payloads
}

// ImportFromJSON imports OpenAPI spec from JSON string
func (i *Importer) ImportFromJSON(jsonStr string) (*ImportResult, error) {
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData([]byte(jsonStr))
	if err != nil {
		return nil, fmt.Errorf("failed to load OpenAPI JSON: %w", err)
	}

	// Convert to temporary spec URL for processing
	// This is a simplified approach - in production, save to temp file
	return i.processDoc(doc, "inline.json")
}

// processDoc processes an OpenAPI document
func (i *Importer) processDoc(doc *openapi3.T, source string) (*ImportResult, error) {
	result := &ImportResult{
		SpecURL:   source,
		CaseID:    i.caseID,
		Payloads:  []GeneratedPayload{},
		Endpoints: []Endpoint{},
	}

	for path, pathItem := range doc.Paths.Map() {
		if pathItem == nil {
			continue
		}

		operations := map[string]*openapi3.Operation{
			"get":    pathItem.Get,
			"post":   pathItem.Post,
			"put":    pathItem.Put,
			"delete": pathItem.Delete,
		}

		for method, operation := range operations {
			if operation == nil {
				continue
			}

			endpoint := Endpoint{
				Path:   path,
				Method: strings.ToUpper(method),
			}

			if operation.Parameters != nil {
				for _, paramRef := range operation.Parameters {
					if paramRef == nil {
						continue
					}
					param := paramRef.Value
					if param == nil {
						continue
					}

					payload := i.generatePayloadForParameter(param, path, method)
					if payload != nil {
						result.Payloads = append(result.Payloads, *payload)
						endpoint.InjectionPoints = append(endpoint.InjectionPoints, param.Name)
					}
				}
			}

			if len(endpoint.InjectionPoints) > 0 {
				result.Endpoints = append(result.Endpoints, endpoint)
			}
		}
	}

	return result, nil
}

// ImportResult represents the result of an OpenAPI import
type ImportResult struct {
	SpecURL   string             `json:"spec_url"`
	CaseID    string             `json:"case_id"`
	Payloads  []GeneratedPayload `json:"payloads"`
	Endpoints []Endpoint         `json:"endpoints"`
}

// GeneratedPayload represents a generated OAST payload
type GeneratedPayload struct {
	Token           string `json:"token"`
	Location        string `json:"location"` // query, header, path, body
	ParameterName   string `json:"parameter_name"`
	Endpoint        string `json:"endpoint"`
	RenderedPayload string `json:"rendered_payload"`
	ContentType     string `json:"content_type,omitempty"`
	Description     string `json:"description"`
}

// Endpoint represents an API endpoint with injection points
type Endpoint struct {
	Path            string   `json:"path"`
	Method          string   `json:"method"`
	InjectionPoints []string `json:"injection_points"`
}

// generateRandomToken generates a random token (simplified)
func generateRandomToken() string {
	// In production, use crypto/rand
	return fmt.Sprintf("%x", time.Now().UnixNano())
}
