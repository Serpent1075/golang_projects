// Connect will receive the $connect request
// It will handle the authorization also
func Connect(request APIGatewayWebsocketProxyRequest) (interface{}, error) {
	if request.RequestContext.Authorizer == nil {
		return Authorizer(request)
	}

	id := request.RequestContext.Authorizer.(map[string]interface{})["cognito:username"].(string)
	connectionID := request.RequestContext.ConnectionID
	StoreSocket(id, connectionID)

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
	}, nil
}

// StoreSocket will store the id,connectionid map in dynamodb
func StoreSocket(id, connectionID string) error {
	m := models.UserSocket{
		ID:           id,
		ConnectionID: connectionID,
	}

	av, err := dynamodbattribute.MarshalMap(m)
	if err != nil {
		log.Fatalln("Unable to marshal user socket map", err.Error())
	}

	input := &dynamodb.PutItemInput{
		TableName: aws.String("GoChatSockets"),
		Item:      av,
	}

	sess := GetSession()

	db := dynamodb.New(sess)

	_, err = db.PutItem(input)
	if err != nil {
		log.Fatal("INSERT ERROR", err.Error())
	}

	return nil
}