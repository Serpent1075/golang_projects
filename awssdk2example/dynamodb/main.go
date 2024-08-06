// Handler is the base handler that will receive all web socket request
func Handler(request APIGatewayWebsocketProxyRequest) (interface{}, error) {
	switch request.RequestContext.RouteKey {
	case "$connect":
		return Connect(request)
	case "$disconnect":
		return Disconnect(request)
	default:
		return Default(request)
	}
}

func main() {
	lambda.Start(Handler)
}

// GetSession creates a new aws session and returns it
func GetSession() *session.Session {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("ap-south-1"),
	})
	if err != nil {
		log.Fatalln("Unable to create AWS session", err.Error())
	}

	return sess
}