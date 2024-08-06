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
	EC2         []*EC2Data
	VPN         []*VPNData
	ClusterData []*ClusterData
	RDSData     []*RDSData
	LB          []*LoadBalancerData
	EBSData     []*EBSData
	EIPData     []*EIPData
	//LBCount        LBCountData
	NatGatewayData        []*NatGatewayData
	SecurityGroupRuleData []*SecurityGroupRuleData
}

type SecurityGroupRuleData struct {
	GroupName   string
	GroupID     string
	RuleID      string
	SrcAddr     string
	Protocol    string
	Port        string
	InOutbound  string
	Description string
}

type EIPData struct {
	IPAddress     string
	AssociationId string
}

type EC2Data struct {
	InstanceName    string
	InstanceId      string
	PrivateIP       string
	InstanceType    string
	PlatformDetails string
	Usage           string
	Hostname        string
	OSVersion       string
	Status          string
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

type EBSData struct {
	ID    string
	Name  string
	Type  string
	Size  int32
	IOPS  int32
	State string
}

type LBCountData struct {
	PrdExtAlb int
	PrdExtNlb int
	PrdIntAlb int
	PrdIntNlb int
	DevExtAlb int
	DevExtNlb int
	DevIntAlb int
	DevIntNlb int
}

type ClusterData struct {
	Identifier     string
	Engine         string
	EngineVersion  string
	Status         string
	EndPoint       string
	ReaderEndpoint string
	MultiAZ        bool
	Port           string
	Usage          string
	RDSData        []RDSData
}

type RDSData struct {
	Identifier       string
	Engine           string
	EngineVersion    string
	EndpointAddress  string
	MultiAZ          bool
	StorageType      string
	InstanceClass    string
	InstancePort     int32
	AllocatedStorage int32
	Usage            string
}
