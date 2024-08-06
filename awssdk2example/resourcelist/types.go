package main

import "github.com/aws/aws-sdk-go-v2/aws"

type Account struct {
	AccountID   string
	AccountName string
	Filename    string
	AwsCfg      aws.Config
}

type ResultResource struct {
	AccountID   string
	AccountName string
	EC2         []EC2Data
	VPN         []VPNData
	ClusterData []ClusterData
	RDSData     []RDSData
}

type EC2Data struct {
	InstanceId      string
	PrivateIP       string
	InstanceType    string
	PlatformDetails string
	Usage           string
	Hostname        string
	OSVersion       string
}

type VPNData struct {
	Name string
}

type NatGatewayData struct {
	Name string
}

type LoadBalancerData struct {
	Name       string
	Type       string
	InExternal string
	PrdDev     string
}

type ClusterData struct {
	Identifier           string
	Engine               string
	EngineVersion        string
	Status               string
	EndPoint             string
	ReaderEndpoint       string
	MultiAZ              bool
	StorageType          string
	Port                 string
	ClusterInstanceClass string
	Usage                string
	RDSData              []RDSData
}

type RDSData struct {
	ClusterIdentifier string
	Identifier        string
	Engine            string
	EngineVersion     string
	EndpointAddress   string
	MultiAZ           bool
	StorageType       string
	InstanceClass     string
	InstancePort      int32
	AllocatedStorage  int32
	Usage             string
}
