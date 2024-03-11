package flow

import (
	"bytes"
	"encoding/json"
	"gograph/internal/log"
	"gograph/internal/schema"
	"gograph/internal/template"
	"gograph/internal/util"
	"io"
	"net/http"
)

var DEBUG bool = true

type GraphqlRequestBody struct {
	OperationName string                 `json:"operationName"`
	Query         string                 `json:"query"`
	Variables     map[string]interface{} `json:"variables"`
}

type GraphqlRequest struct {
	Endpoint  *FlowEndpoint          `json:"endpoint"`
	QueryName string                 `json:"queryName"`
	Depth     int                    `json:"depth"`
	Variables map[string]interface{} `json:"variables"`
	Headers   map[string]interface{} `json:"headers"`
	context   *StepTemplateContext
}

func (g *GraphqlRequest) GenerateQuery(options *schema.QuerySelectorOptions) *schema.QueryString {
	queryName := template.RunTemplateOrUnparsed(g.QueryName, g.context)
	operation := g.Endpoint.schema.FindOperationByName(queryName)
	if operation == nil {
		log.Fatalln("GraphqlRequest: operation not found", queryName)
	}

	return operation.QueryString(options)
}

type GraphqlRunResult struct {
	Request *GraphqlRunResult_Request  `json:"request"`
	Reponse *GraphqlRunResult_Response `json:"response"`
}

type GraphqlRunResult_Request struct {
	Body *GraphqlRequestBody `json:"body"`
}

type GraphqlRunResult_Response struct {
	StatusCode    int            `json:"statusCode"`
	Status        string         `json:"status"`
	ContentLength int64          `json:"contentLength"`
	Header        http.Header    `json:"header"`
	Cookies       []*http.Cookie `json:"cookies"`
	Body          []byte         `json:"body"`
}

func (resp *GraphqlRunResult_Response) String() string {
	return string(resp.Body)
}

func (resp *GraphqlRunResult_Response) Json() (map[string]interface{}, error) {
	variables := make(map[string]interface{})
	err := json.Unmarshal(resp.Body, &variables)
	if err != nil {
		log.Println("Unable to decode response body to json", err)
		return variables, err
	}
	return variables, nil
}

func (g *GraphqlRequest) Run() (*GraphqlRunResult, error) {

	depth := 3
	if g.Depth > 0 {
		depth = g.Depth
	}
	query := g.GenerateQuery(&schema.QuerySelectorOptions{
		IgnoreUnderscored: true,
		MaxDepth:          uint8(depth),
	})

	result := &GraphqlRunResult{
		Request: &GraphqlRunResult_Request{
			Body: &GraphqlRequestBody{
				OperationName: query.Name,
				Query:         query.Text,
				Variables:     g.Variables,
			},
		},
		Reponse: nil,
	}

	// Marshal the query into a JSON request body
	requestBody, err := json.Marshal(result.Request.Body)
	if err != nil {
		return result, err
	}

	url := g.Endpoint.UrlParsed(g.context)

	// Create a new HTTP request
	log.Verboseln("GraphqlRequest: calling", url)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		return result, err
	}

	// Prepare the header
	req.Header.Set("Content-Type", "application/json")
	if g.Headers != nil {

		for name, value := range g.Headers {

			// convert the value to string
			var v string
			switch t := value.(type) {
			case string:
				v = t
			default:
				v = util.JsonPrint(t)
			}
			v = template.RunTemplateOrUnparsed(v, g.context)
			log.Debugf("Setting header: %v=%v", name, v)
			req.Header.Set(name, v)
		}
	}
	// Set request headers

	// Add your authorization token here if needed
	// req.Header.Set("Authorization", "Bearer YOUR_ACCESS_TOKEN")

	// Initialize HTTP client
	client := &http.Client{}

	// Send the request
	resp, err := client.Do(req)
	if err != nil {
		return result, err
	}
	defer resp.Body.Close()

	result.Reponse = &GraphqlRunResult_Response{
		Status:     resp.Status,
		StatusCode: resp.StatusCode,
		Header:     resp.Header,
		Cookies:    resp.Cookies(),
	}

	// resp.Header

	// Get the response text
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return result, err
	}

	result.Reponse.ContentLength = resp.ContentLength

	result.Reponse.Body = responseBody

	if DEBUG {
		log.Verbose("GraphqlRequest: query")
		log.Verbose(" Header          :")
		// TODO
		log.Verbose(" Body           :")
		log.Verbose(string(requestBody))
		log.Debug(" Query          :")
		log.Debug(query.Text)

		log.Verbose("GraphqlRequest: result")
		log.Verbose(" Status          :", result.Reponse.Status)
		log.Verbose(" Code            :", result.Reponse.StatusCode)
		log.Verbose(" ContentLength   :", result.Reponse.ContentLength)
		log.Verbosef(" Header          : %v", len(result.Reponse.Header))
		for k, header := range result.Reponse.Header {
			log.Verbosef("  %v: %v ", k, header)
		}
		log.Verbosef(" Cookies         : %v", len(result.Reponse.Cookies))
		for i, cookie := range result.Reponse.Cookies {
			log.Verbosef("  %v: %v ", i, cookie.Raw)
		}
		log.Verbosef(" Body            : %v", len(result.Reponse.Body))
		log.Verbosef(string(result.Reponse.Body))
	}

	return result, nil
}
