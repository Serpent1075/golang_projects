// APIGatewayWebsocketProxyRequest contains data coming from the API Gateway proxy
type APIGatewayWebsocketProxyRequest struct {
	MethodArn                       string                                 `json:"methodArn"`
	Resource                        string                                 `json:"resource"` // The resource path defined in API Gateway
	Path                            string                                 `json:"path"`     // The url path for the caller
	HTTPMethod                      string                                 `json:"httpMethod"`
	Headers                         map[string]string                      `json:"headers"`
	MultiValueHeaders               map[string][]string                    `json:"multiValueHeaders"`
	QueryStringParameters           map[string]string                      `json:"queryStringParameters"`
	MultiValueQueryStringParameters map[string][]string                    `json:"multiValueQueryStringParameters"`
	PathParameters                  map[string]string                      `json:"pathParameters"`
	StageVariables                  map[string]string                      `json:"stageVariables"`
	RequestContext                  APIGatewayWebsocketProxyRequestContext `json:"requestContext"`
	Body                            string                                 `json:"body"`
	IsBase64Encoded                 bool                                   `json:"isBase64Encoded,omitempty"`
}

// APIGatewayWebsocketProxyRequestContext contains the information to identify
// the AWS account and resources invoking the Lambda function. It also includes
// Cognito identity information for the caller.
type APIGatewayWebsocketProxyRequestContext struct {
	AccountID          string                           `json:"accountId"`
	ResourceID         string                           `json:"resourceId"`
	Stage              string                           `json:"stage"`
	RequestID          string                           `json:"requestId"`
	Identity           events.APIGatewayRequestIdentity `json:"identity"`
	ResourcePath       string                           `json:"resourcePath"`
	Authorizer         interface{}                      `json:"authorizer"`
	HTTPMethod         string                           `json:"httpMethod"`
	APIID              string                           `json:"apiId"` // The API Gateway rest API Id
	ConnectedAt        int64                            `json:"connectedAt"`
	ConnectionID       string                           `json:"connectionId"`
	DomainName         string                           `json:"domainName"`
	Error              string                           `json:"error"`
	EventType          string                           `json:"eventType"`
	ExtendedRequestId  string                           `json:"extendedRequestId"`
	IntegrationLatency string                           `json:"integrationLatency"`
	MessageDirection   string                           `json:"messageDirection"`
	MessageID          interface{}                      `json:"messageId"`
	RequestTime        string                           `json:"requestTime"`
	RequestTimeEpoch   int64                            `json:"requestTimeEpoch"`
	RouteKey           string                           `json:"routeKey"`
	Status             string                           `json:"status"`
}
