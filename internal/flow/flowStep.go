package flow

import (
	"encoding/json"
	"gograph/internal/template"
	"gograph/internal/util"
	"regexp"
	"slices"
	"strings"

	"github.com/PaesslerAG/jsonpath"
)

type StepTemplateContext struct {
	State map[string]interface{}
	Step  *FlowStep
}

// FlowStep
// ----------------------------------------
type FlowStep struct {
	Name     string   `yaml:",omitempty"`
	Endpoint string   `yaml:",omitempty"`
	Query    string   `yaml:",omitempty"`
	Queries  []string `yaml:",flow,omitempty"`
	Depth    int      `yaml:",omitempty"`

	Input string `yaml:",omitempty"`

	Headers map[string]interface{}

	Result struct {
		Status            []int `yaml:",flow,omitempty"`
		ExpectError       bool  `yaml:"error,omitempty"`
		ContinueOnFailure bool  `yaml:"continueOnFailure,omitempty"`
		Values            []struct {
			Name  string `yaml:",omitempty"`
			Path  string `yaml:",omitempty"`
			Match string `yaml:",omitempty"`
			Exact string `yaml:",omitempty"`
		}
	}
}

func (e *FlowStep) EndpointParsed(context *StepTemplateContext) string {
	return template.RunTemplateOrUnparsed(e.Endpoint, context)
}

func (e *FlowStep) NameParsed(context *StepTemplateContext) string {
	return template.RunTemplateOrUnparsed(e.Name, context)
}

func (e FlowStep) InputParsed(context *StepTemplateContext) (map[string]interface{}, error) {
	// Extract the input to a map
	parsedInput, err := template.RunTemplate(e.Input, context)

	if err != nil {
		return nil, err
	}

	variables := make(map[string]interface{})

	if len(parsedInput) > 0 {
		err = json.Unmarshal([]byte(parsedInput), &variables)
		if err != nil {
			return variables, err
		}
	}

	return variables, nil
}

func (e *FlowStep) SelectEndpoint(options []FlowEndpoint, context *StepTemplateContext) *FlowEndpoint {
	if len(options) == 0 {
		return nil
	}

	// Return the first if no endpoint is specified
	if len(e.Endpoint) == 0 {
		return &options[0]
	}

	endpointName := e.EndpointParsed(context)

	idx := slices.IndexFunc(options, func(f FlowEndpoint) bool { return f.Name == endpointName })
	if idx >= 0 {
		return &options[idx]
	}
	return nil
}

func (step *FlowStep) Run(flow *FlowDefinition) *StepResult {

	templateContext := &StepTemplateContext{
		State: flow.State,
		Step:  step,
	}

	result := &StepResult{
		Name:  step.NameParsed(templateContext),
		State: make(map[string]interface{}),
	}

	result.Verbosef("running")

	// Get the endpoint
	endpoint := step.SelectEndpoint(flow.Endpoints, templateContext)
	if endpoint == nil {
		result.Errorf("unable to find endpoint: %v", step.EndpointParsed(templateContext))
		return result
	}
	endpoint.LoadSchema(flow.BasePath)

	// Get the list of queries
	//   either defined as a single `query` or as a list of `queries` or both
	var queries []string
	if len(step.Query) > 0 {
		queries = []string{step.Query}
	}
	if len(step.Queries) > 0 {
		if queries == nil {
			queries = step.Queries
		} else {
			queries = append(queries, step.Queries...)
		}
	}

	result.Debugf("found %v queries", len(queries))

	// Iterate over the queries
	for _, queryName := range queries {

		// Get the query input
		var input map[string]interface{}
		if len(step.Input) > 0 {
			// Input provided by user
			v, err := step.InputParsed(templateContext)
			if err != nil {
				result.Errorf("[%v] unable to decode query input: %v in \n'%v'", queryName, err, step.Input)
				v = make(map[string]interface{})
			}
			input = v
		} else {
			// Input not provided by user => generate
			operation := endpoint.schema.FindOperationByName(queryName)
			if operation != nil {

				input = operation.Variables()

				result.Debugf("[%v] Generating variables: %v", queryName, util.PrettyPrint(input))
			} else {
				result.Errorf("[%v] unable to find the operation", queryName)
			}

			if input == nil {
				input = make(map[string]interface{})
			}
		}

		// Prepare the query
		query := &GraphqlRequest{
			Endpoint:  endpoint,
			QueryName: queryName,
			Depth:     step.Depth,
			Variables: input,
			Headers:   step.Headers,
			context:   templateContext,
		}

		result.Verbosef("[%v] executing query on %v", query.QueryName, query.Endpoint.Url)
		queryResult, err := query.Run()
		if err != nil {
			result.Errorf("[%v] query failed: %v", query.QueryName, err)
		}

		result.Result = queryResult

		if queryResult != nil && queryResult.Reponse != nil {

			// Check Status Code
			// ----------------------------------------
			expectedStatus := []int{200}
			if len(step.Result.Status) > 0 {
				expectedStatus = step.Result.Status
			}

			if slices.Index(expectedStatus, queryResult.Reponse.StatusCode) < 0 {
				result.Errorf("[%v] invalid response status: %v", query.QueryName, queryResult.Reponse.Status)
			}

			// Check for graphql Error
			// ----------------------------------------
			responseJson, err := queryResult.Reponse.Json()
			if err != nil {
				result.Errorf("[%v] unable to parse response json", query.QueryName)
			}

			graphQlErrors := responseJson["errors"]

			// The user expect a graphql error
			if step.Result.ExpectError {
				if graphQlErrors == nil {
					result.Errorf("[%v] expected graphql error not found in response", query.QueryName)
				}
			} else {
				// The user doesn't expect a graphql error
				if graphQlErrors != nil {
					switch graphQlErrors := graphQlErrors.(type) {
					case []interface{}:
						for _, gqlError := range graphQlErrors {
							switch gqlError := gqlError.(type) {
							case map[string]interface{}:
								result.Errorf("[%v] GraphQL error: %v", query.QueryName, gqlError["message"])
							default:
								result.Errorf("[%v] unknown graphql error[] response format", query.QueryName)
							}
						}
					default:
						result.Errorf("[%v] unknown graphql error response format", query.QueryName)
					}
				}
			}

			// Process the responses
			// ----------------------------------------
			for _, value := range step.Result.Values {
				// Extract the data
				data, err := jsonpath.Get(value.Path, responseJson)
				if err != nil {
					result.Errorf("[%v] unable to load jsonpath: %v, %v", query.QueryName, value.Path, err)
				} else {
					result.Debugf("[%v] jsonpath %v=%v", query.QueryName, value.Path, util.PrettyPrint(data))
					// Store the result if it is named
					if len(value.Name) > 0 {
						result.State[value.Name] = data
						flow.State[value.Name] = data
						result.Debugf("[%v] stored [%v]=%v", query.QueryName, value.Name, util.PrettyPrint(data))
					}

					// Check if the result match a provided regexp
					match := strings.TrimSpace(value.Match)
					if len(match) > 0 {
						result.Debugf("[%v] checking if data is a regex match %v=%v", query.QueryName, data, match)
						// The type of the data MUST be string
						var str string
						if data == nil {
							str = ""
						}
						switch t := data.(type) {
						case string:
							str = t
						default:
							// convert it back to JSON to get a value
							str = util.JsonPrint(data)
						}
						re, err := regexp.Compile(match)
						if err != nil {
							result.Errorf("[%v] invalid regexp: %v, %v, %v", query.QueryName, value.Path, match, err)
						}
						if !re.MatchString(str) {
							result.Errorf("[%v] Value doesn't match expected format: %v=%v, %v", query.QueryName, value.Path, str, match)
						}
					}

					// Check if the value is an exact match
					exact := strings.TrimSpace(value.Exact)
					if len(exact) > 0 {
						exact = template.RunTemplateOrUnparsed(exact, templateContext)
						result.Debugf("[%v] checking if data is an exact match %v=%v", query.QueryName, data, exact)
						if data != exact {
							result.Errorf("[%v] value error: [%v](%v) != (%v)", query.QueryName, value.Path, data, exact)
						}
					}
				}
			}
		}
		// range queries
	}
	return result
}
