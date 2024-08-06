// Authorizer custom api authorizer
func Authorizer(request APIGatewayWebsocketProxyRequest) (events.APIGatewayCustomAuthorizerResponse, error) {
	token := request.QueryStringParameters["token"]

	// Fetch all keys
	jwkSet, err := jwk.Fetch("https://cognito-idp.ap-south-1.amazonaws.com/ap-south-1_vvx4f42sK/.well-known/jwks.json")
	if err != nil {
		log.Fatalln("Unable to fetch keys")
	}

	// Verify
	t, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
		keys := jwkSet.LookupKeyID(t.Header["kid"].(string))
		return keys[0].Materialize()
	})
	if err != nil || !t.Valid {
		log.Fatalln("Unauthorized")
	}

	claims := t.Claims.(jwt.MapClaims)

	return events.APIGatewayCustomAuthorizerResponse{
		PrincipalID: "me",
		PolicyDocument: events.APIGatewayCustomAuthorizerPolicy{
			Version: "2012-10-17",
			Statement: []events.IAMPolicyStatement{
				events.IAMPolicyStatement{
					Action:   []string{"execute-api:*"},
					Effect:   "Allow",
					Resource: []string{request.MethodArn},
				},
			},
		},
		Context: claims,
	}, nil
}