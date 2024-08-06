package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	"github.com/aws/aws-sdk-go-v2/service/rds"
)

type ClientInstanceType struct {
	instancetype       string
	instancegeneration string
}

type KeyData struct {
	AccessKey string
	SecretKey string
}

func DescribeRecord(ctx context.Context, accountids []Account) {

	for i, account := range accountids {

		file, _ := os.Open("./access_key/" + strings.TrimSpace(account.Filename))
		keydata, err := ReadCsv(file)
		check(err)

		cfg, err := config.LoadDefaultConfig(
			context.TODO(),
			config.WithCredentialsProvider(
				credentials.NewStaticCredentialsProvider(keydata[1][0], keydata[1][1], ""),
			),
		)
		check(err)
		accountids[i].AwsCfg = cfg
	}

	//var resultresources []ResultResource = make([]ResultResource, 0)
	for _, account := range accountids {
		fmt.Println(account.AccountName)
		//ec2datas := GetEC2Data(account.AwsCfg)
		//vpndatas := GetVPNData(account.AwsCfg)
		//clusterdatas, rdsdatas := GetRDSData(account.AwsCfg)
		//natgatewaydatas := GetNatGateway(account.AwsCfg)
		lbdatas := GetLoadBalancer(account.AwsCfg)
		/*
			if ec2datas != nil {
				fmt.Println(ec2datas)
			}
			if vpndatas != nil {
				fmt.Println(vpndatas)
			}
			if clusterdatas != nil {
				fmt.Println(clusterdatas)
			}
			if rdsdatas != nil {
				fmt.Println(rdsdatas)
			}
			if natgatewaydatas != nil {
				fmt.Println(natgatewaydatas)
			}
		*/
		if lbdatas != nil {
			fmt.Println(lbdatas)
		}
		fmt.Println("------------------------------------------------------")
	}

	/*
		resultresource := ResultResource{
			AccountID: accountids[0],
			AccountName: ,
			EC2:
		}
	*/

}

func GetLoadBalancer(cfg aws.Config) []LoadBalancerData {
	lbclient := elasticloadbalancingv2.NewFromConfig(cfg)
	lbinput := elasticloadbalancingv2.DescribeLoadBalancersInput{}
	lboutput, lberr := lbclient.DescribeLoadBalancers(context.Background(), &lbinput)
	check(lberr)

	var lbdatas []LoadBalancerData = make([]LoadBalancerData, 0)

	for _, data := range lboutput.LoadBalancers {
		lbname := *data.LoadBalancerName
		var lbprddev string
		if strings.Contains(strings.ToLower(lbname), "dev") {
			lbprddev = "Dev"
		} else {
			lbprddev = "Prod"
		}

		lbdata := LoadBalancerData{
			Name:       *data.LoadBalancerName,
			Type:       string(data.Type),
			InExternal: string(data.Scheme),
			PrdDev:     lbprddev,
		}
		lbdatas = append(lbdatas, lbdata)
	}
	return lbdatas
}

func GetNatGateway(cfg aws.Config) []NatGatewayData {
	ec2client := ec2.NewFromConfig(cfg)
	natgatewayinput := ec2.DescribeNatGatewaysInput{}
	natgatewayoutput, natgatewayerr := ec2client.DescribeNatGateways(context.Background(), &natgatewayinput)
	check(natgatewayerr)

	var natgatewaydatas []NatGatewayData = make([]NatGatewayData, 0)
	for _, data := range natgatewayoutput.NatGateways {
		natgatewaydata := NatGatewayData{
			Name: *data.Tags[0].Value,
		}
		natgatewaydatas = append(natgatewaydatas, natgatewaydata)
		fmt.Println(data.Tags[0].Value)
	}
	return natgatewaydatas
}

func GetVPNData(cfg aws.Config) []VPNData {
	ec2client := ec2.NewFromConfig(cfg)
	vpninput := ec2.DescribeVpnConnectionsInput{}
	vpnoutput, vpnerr := ec2client.DescribeVpnConnections(context.Background(), &vpninput)
	check(vpnerr)

	var vpndatas []VPNData = make([]VPNData, 0)
	for _, data := range vpnoutput.VpnConnections {
		vpndata := VPNData{
			Name: *data.Tags[0].Value,
		}
		vpndatas = append(vpndatas, vpndata)
	}
	return vpndatas
}

func GetRDSData(cfg aws.Config) ([]ClusterData, []RDSData) {
	rdsclient := rds.NewFromConfig(cfg)
	rdsclusterinput := rds.DescribeDBClustersInput{}
	rdsclusteroutput, rdsclusteroutputerr := rdsclient.DescribeDBClusters(context.Background(), &rdsclusterinput)
	check(rdsclusteroutputerr)
	rdsinstanceinput := rds.DescribeDBInstancesInput{}
	rdsinstanceoutput, rdsinstanceoutputerr := rdsclient.DescribeDBInstances(context.Background(), &rdsinstanceinput)
	check(rdsinstanceoutputerr)
	var clusterdatas []ClusterData = make([]ClusterData, 0)

	for _, data := range rdsclusteroutput.DBClusters {
		var instancedata []RDSData = make([]RDSData, 0)
		clusterdata := ClusterData{
			Identifier:           *data.DBClusterIdentifier,
			Engine:               *data.Engine,
			EngineVersion:        *data.EngineVersion,
			Status:               *data.Status,
			EndPoint:             *data.Endpoint,
			ReaderEndpoint:       *data.ReaderEndpoint,
			MultiAZ:              *data.MultiAZ,
			ClusterInstanceClass: *data.DBClusterInstanceClass,
			StorageType:          *data.StorageType,
			RDSData:              instancedata,
		}

		for _, tag := range data.TagList {
			if *tag.Key == "Usage" {
				clusterdata.Usage = *tag.Value
				log.Println(*tag.Value)
			}
		}
		clusterdatas = append(clusterdatas, clusterdata)
	}

	var rdsdatas []RDSData = make([]RDSData, 0)
	for _, data := range rdsinstanceoutput.DBInstances {
		var added bool = false
		for i, cluster := range clusterdatas {
			if *data.DBClusterIdentifier == cluster.Identifier {
				rdsdata := RDSData{
					ClusterIdentifier: *data.DBClusterIdentifier,
					Identifier:        *data.DBInstanceIdentifier,
					Engine:            *data.Engine,
					EngineVersion:     *data.EngineVersion,
					EndpointAddress:   *data.Endpoint.Address,
					MultiAZ:           *data.MultiAZ,
					StorageType:       *data.StorageType,
					InstanceClass:     *data.DBInstanceClass,
					InstancePort:      *data.DbInstancePort,
					AllocatedStorage:  *data.AllocatedStorage,
				}
				for _, tag := range data.TagList {
					if *tag.Key == "Usage" {

						rdsdata.Usage = *tag.Value
					}
				}
				clusterdatas[i].RDSData = append(clusterdatas[i].RDSData, rdsdata)
				added = true
			}
		}
		if !added {
			rdsdata := RDSData{
				Identifier:       *data.DBInstanceIdentifier,
				Engine:           *data.Engine,
				EngineVersion:    *data.EngineVersion,
				EndpointAddress:  *data.Endpoint.Address,
				MultiAZ:          *data.MultiAZ,
				StorageType:      *data.StorageType,
				InstanceClass:    *data.DBInstanceClass,
				InstancePort:     *data.DbInstancePort,
				AllocatedStorage: *data.AllocatedStorage,
			}
			for _, tag := range data.TagList {
				if *tag.Key == "Usage" {
					rdsdata.Usage = *tag.Value
				}
			}
			rdsdatas = append(rdsdatas, rdsdata)
		}

	}
	return clusterdatas, rdsdatas
}

func GetEC2Data(cfg aws.Config) []EC2Data {
	ec2client := ec2.NewFromConfig(cfg)
	ec2input := ec2.DescribeInstancesInput{}
	ec2output, ec2err := ec2client.DescribeInstances(context.TODO(), &ec2input)
	check(ec2err)
	var ec2datas []EC2Data = make([]EC2Data, 0)

	for _, data := range ec2output.Reservations {
		ec2data := EC2Data{
			InstanceId:   *data.Instances[0].InstanceId,
			PrivateIP:    *data.Instances[0].PrivateIpAddress,
			InstanceType: string(data.Instances[0].InstanceType.Values()[0]),
		}

		for _, tag := range data.Instances[0].Tags {
			if *tag.Key == "Usage" {
				ec2data.Usage = *tag.Value
			}

			if *tag.Key == "Hostname" {
				ec2data.Hostname = *tag.Value
			}

			if *tag.Key == "OS Version" {
				ec2data.OSVersion = *tag.Value
			}
		}
		ec2datas = append(ec2datas, ec2data)
	}

	return ec2datas
}

func ReadAccount() []Account {
	file, _ := os.Open("./client_list/client.csv")
	clientdatas, err := ReadCsv(file)
	check(err)
	var accountlist []Account = make([]Account, 0)
	for i, data := range clientdatas {
		if i != 0 {
			accountdata := Account{
				AccountID:   data[0],
				AccountName: data[1],
				Filename:    data[2],
			}
			accountlist = append(accountlist, accountdata)
		}
	}
	return accountlist
}

func ReadCsv(f io.ReadCloser) ([][]string, error) {

	// Read File into a Variable
	lines, readcsverr := csv.NewReader(f).ReadAll()
	if readcsverr != nil {
		return [][]string{}, readcsverr
	}

	return lines, nil
}

func main() {
	accountlist := ReadAccount()
	DescribeRecord(context.TODO(), accountlist)
}

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
