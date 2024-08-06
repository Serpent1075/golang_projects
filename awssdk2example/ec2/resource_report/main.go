package main

import (
	"context"
	"encoding/csv"
	"io"
	"log"
	"os"
	"sort"
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

func DescribeRecord(ctx context.Context, accountids []Account) []*ResultResource {

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

	var resultresources []*ResultResource = make([]*ResultResource, 0)
	for _, account := range accountids {

		ec2datas := GetEC2Data(account.AwsCfg)
		vpndatas := GetVPNData(account.AwsCfg)
		eipdatas := GetEIPData(account.AwsCfg)
		ebsdatas := GetEBSData(account.AwsCfg)
		sgdatas := GetSecurityGroupData(account.AwsCfg)
		natgatewaydatas := GetNatGateway(account.AwsCfg)
		lbdatas := GetLoadBalancer(account.AwsCfg)
		//var lbcount *LBCountData
		clusterdatas, rdsdatas := GetRDSData(account.AwsCfg)
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

			if lbdatas != nil {
				lbcount = LBCount(lbdatas)
			}
		*/
		resultresource := ResultResource{
			AccountID:             account.AccountID,
			AccountName:           account.AccountName,
			EC2:                   ec2datas,
			VPN:                   vpndatas,
			ClusterData:           clusterdatas,
			RDSData:               rdsdatas,
			LB:                    lbdatas,
			EIPData:               eipdatas,
			EBSData:               ebsdatas,
			SecurityGroupRuleData: sgdatas,
			NatGatewayData:        natgatewaydatas,
		}

		if strings.TrimSpace(resultresource.AccountID) != "" {
			resultresources = append(resultresources, &resultresource)
		}

	}
	return resultresources
}

func LBCount(lbdatas []*LoadBalancerData) *LBCountData {
	var lbcount LBCountData

	for _, lbdata := range lbdatas {
		if lbdata.PrdDev == "Dev" {
			if lbdata.InExternal == "internal" {
				if lbdata.Type == "application" {
					lbcount.DevIntAlb++
				} else {
					lbcount.DevIntNlb++
				}
			} else {
				if lbdata.Type == "application" {
					lbcount.DevExtAlb++
				} else {
					lbcount.DevExtNlb++
				}
			}
		} else {
			if lbdata.InExternal == "internal" {
				if lbdata.Type == "application" {
					lbcount.PrdIntAlb++
				} else {
					lbcount.PrdIntNlb++
				}
			} else {
				if lbdata.Type == "application" {
					lbcount.PrdExtAlb++
				} else {
					lbcount.PrdExtNlb++
				}
			}
		}
	}

	return &lbcount
}

func GetEBSData(cfg aws.Config) []*EBSData {
	ebsclient := ec2.NewFromConfig(cfg)
	ebsinput := ec2.DescribeVolumesInput{}
	ebsoutput, ebserr := ebsclient.DescribeVolumes(context.Background(), &ebsinput)
	check(ebserr)

	var ebsdatas []*EBSData = make([]*EBSData, 0)

	for _, data := range ebsoutput.Volumes {
		ebsdata := EBSData{
			ID:    *data.VolumeId,
			Type:  string(data.VolumeType),
			Size:  *data.Size,
			State: string(data.State),
		}

		if data.Iops != nil {
			ebsdata.IOPS = *data.Iops
		}

		for _, tag := range data.Tags {
			if *tag.Key == "Name" {
				ebsdata.Name = *tag.Value
				break
			}
		}
		ebsdatas = append(ebsdatas, &ebsdata)
	}
	return ebsdatas
}

func GetLoadBalancer(cfg aws.Config) []*LoadBalancerData {
	lbclient := elasticloadbalancingv2.NewFromConfig(cfg)
	lbinput := elasticloadbalancingv2.DescribeLoadBalancersInput{}
	lboutput, lberr := lbclient.DescribeLoadBalancers(context.Background(), &lbinput)
	check(lberr)

	var lbdatas []*LoadBalancerData = make([]*LoadBalancerData, 0)

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
		lbdatas = append(lbdatas, &lbdata)
	}
	return lbdatas
}

func GetNatGateway(cfg aws.Config) []*NatGatewayData {
	ec2client := ec2.NewFromConfig(cfg)
	natgatewayinput := ec2.DescribeNatGatewaysInput{}
	natgatewayoutput, natgatewayerr := ec2client.DescribeNatGateways(context.Background(), &natgatewayinput)
	check(natgatewayerr)

	var natgatewaydatas []*NatGatewayData = make([]*NatGatewayData, 0)
	for _, data := range natgatewayoutput.NatGateways {
		natgatewaydata := NatGatewayData{
			Name: *data.Tags[0].Value,
		}
		natgatewaydatas = append(natgatewaydatas, &natgatewaydata)
	}
	return natgatewaydatas
}

func GetVPNData(cfg aws.Config) []*VPNData {
	ec2client := ec2.NewFromConfig(cfg)
	vpninput := ec2.DescribeVpnConnectionsInput{}
	vpnoutput, vpnerr := ec2client.DescribeVpnConnections(context.Background(), &vpninput)
	check(vpnerr)

	var vpndatas []*VPNData = make([]*VPNData, 0)
	for _, data := range vpnoutput.VpnConnections {
		vpndata := VPNData{
			Name: *data.Tags[0].Value,
		}
		vpndatas = append(vpndatas, &vpndata)
	}
	return vpndatas
}

func GetRDSData(cfg aws.Config) ([]*ClusterData, []*RDSData) {
	rdsclient := rds.NewFromConfig(cfg)
	rdsclusterinput := rds.DescribeDBClustersInput{}
	rdsclusteroutput, rdsclusteroutputerr := rdsclient.DescribeDBClusters(context.Background(), &rdsclusterinput)
	check(rdsclusteroutputerr)
	rdsinstanceinput := rds.DescribeDBInstancesInput{}
	rdsinstanceoutput, rdsinstanceoutputerr := rdsclient.DescribeDBInstances(context.Background(), &rdsinstanceinput)
	check(rdsinstanceoutputerr)
	var clusterdatas []*ClusterData = make([]*ClusterData, 0)

	for _, data := range rdsclusteroutput.DBClusters {
		var instancedata []RDSData = make([]RDSData, 0)
		for _, member := range data.DBClusterMembers {
			rdsdata := RDSData{
				Identifier: *member.DBInstanceIdentifier,
			}
			instancedata = append(instancedata, rdsdata)
		}
		clusterdata := ClusterData{
			Identifier:     *data.DBClusterIdentifier,
			Engine:         *data.Engine,
			EngineVersion:  *data.EngineVersion,
			Status:         *data.Status,
			EndPoint:       *data.Endpoint,
			ReaderEndpoint: *data.ReaderEndpoint,
			MultiAZ:        *data.MultiAZ,
			RDSData:        instancedata,
		}

		for _, tag := range data.TagList {
			if *tag.Key == "Usage" {
				clusterdata.Usage = *tag.Value
			}
		}
		clusterdatas = append(clusterdatas, &clusterdata)
	}

	var rdsdatas []*RDSData = make([]*RDSData, 0)
	for _, data := range rdsinstanceoutput.DBInstances {
		var added bool = false
		for i, cluster := range clusterdatas {
			for j, member := range cluster.RDSData {

				if member.Identifier == *data.DBInstanceIdentifier {
					//fmt.Printf("member %s\n", member.Identifier)
					//fmt.Printf("cluster data!@# %s\n", *data.DBInstanceIdentifier)

					rdsdata := RDSData{
						Identifier:       *data.DBInstanceIdentifier,
						Engine:           *data.Engine,
						EngineVersion:    *data.EngineVersion,
						MultiAZ:          *data.MultiAZ,
						StorageType:      *data.StorageType,
						InstanceClass:    *data.DBInstanceClass,
						InstancePort:     *data.DbInstancePort,
						AllocatedStorage: *data.AllocatedStorage,
					}

					if data.Endpoint != nil {
						rdsdata.EndpointAddress = *data.Endpoint.Address
					}

					for _, tag := range data.TagList {
						if *tag.Key == "Usage" {

							rdsdata.Usage = *tag.Value
						}
					}

					clusterdatas[i].RDSData[j] = rdsdata
					added = true
					break
				}
			}
		}

		if !added {
			rdsdata := RDSData{
				Identifier:       *data.DBInstanceIdentifier,
				Engine:           *data.Engine,
				EngineVersion:    *data.EngineVersion,
				MultiAZ:          *data.MultiAZ,
				StorageType:      *data.StorageType,
				InstanceClass:    *data.DBInstanceClass,
				InstancePort:     *data.DbInstancePort,
				AllocatedStorage: *data.AllocatedStorage,
			}
			if data.Endpoint != nil {
				rdsdata.EndpointAddress = *data.Endpoint.Address
			}
			for _, tag := range data.TagList {
				if *tag.Key == "Usage" {
					rdsdata.Usage = *tag.Value
				}
			}
			rdsdatas = append(rdsdatas, &rdsdata)
		}
	}

	return clusterdatas, rdsdatas
}

func GetEIPData(cfg aws.Config) []*EIPData {
	ec2client := ec2.NewFromConfig(cfg)
	eipinput := ec2.DescribeAddressesInput{}
	eipoutput, eiperr := ec2client.DescribeAddresses(context.Background(), &eipinput)
	check(eiperr)

	var eipdatas []*EIPData = make([]*EIPData, 0)

	for _, data := range eipoutput.Addresses {
		var eipdata EIPData
		eipdata.IPAddress = *data.PublicIp
		if data.AssociationId != nil {
			eipdata.AssociationId = *data.AssociationId
		} else {
			eipdata.AssociationId = "not using"
		}
		eipdatas = append(eipdatas, &eipdata)
	}
	return eipdatas
}

func GetSecurityGroupData(cfg aws.Config) []*SecurityGroupRuleData {
	ec2client := ec2.NewFromConfig(cfg)
	/*
		sginput := ec2.DescribeSecurityGroupsInput{}
		sgoutput, sgerr := ec2client.DescribeSecurityGroups(context.Background(), &sginput)
		check(sgerr)
		for _, data := range sgoutput.SecurityGroups {
			log.Println(data.GroupName)
		}*/
	sgruleinput := ec2.DescribeSecurityGroupRulesInput{}
	sgruleoutput, sgruleerr := ec2client.DescribeSecurityGroupRules(context.Background(), &sgruleinput)
	check(sgruleerr)
	var sgdatas []*SecurityGroupRuleData = make([]*SecurityGroupRuleData, 0)
	for _, sgrule := range sgruleoutput.SecurityGroupRules {
		sgdata := SecurityGroupRuleData{
			GroupID: *sgrule.GroupId,
			RuleID:  *sgrule.SecurityGroupRuleId,
		}

		if *sgrule.IpProtocol == "-1" {
			sgdata.Protocol = "all"
		} else {
			sgdata.Protocol = *sgrule.IpProtocol
		}

		if sgrule.CidrIpv4 != nil {
			sgdata.SrcAddr = *sgrule.CidrIpv4
		} else if sgrule.ReferencedGroupInfo != nil {
			sgdata.SrcAddr = *sgrule.ReferencedGroupInfo.GroupId
		} else {
			sgdata.SrcAddr = *sgrule.CidrIpv6
		}

		if *sgrule.FromPort == *sgrule.ToPort {
			if *sgrule.FromPort == -1 {
				sgdata.Port = "all"
			} else {
				sgdata.Port = Int32ToString(sgrule.FromPort)
			}
		} else {
			sgdata.Port = Int32ToString(sgrule.FromPort) + " - " + Int32ToString(sgrule.ToPort)
		}

		if sgrule.Description != nil {
			sgdata.Description = *sgrule.Description
		}

		for _, tag := range sgrule.Tags {
			if *tag.Key == "Name" {
				sgdata.GroupName = *tag.Value
			}
		}
		if *sgrule.IsEgress {
			sgdata.InOutbound = "Outbound"
		} else {
			sgdata.InOutbound = "Inbound"
		}

		sgdatas = append(sgdatas, &sgdata)
	}
	SortByGIDandRuleId(sgdatas)
	return sgdatas
}

func SortByGIDandRuleId(sgdatas []*SecurityGroupRuleData) {
	sort.Slice(sgdatas, func(i, j int) bool {
		if sgdatas[i].GroupID != sgdatas[j].GroupID {
			return sgdatas[i].GroupID < sgdatas[j].GroupID
		}
		if sgdatas[i].InOutbound != sgdatas[j].InOutbound {
			return sgdatas[i].InOutbound < sgdatas[j].InOutbound
		}
		return sgdatas[i].RuleID < sgdatas[j].RuleID

	})
}

func Int32ToString(n *int32) string {
	buf := [11]byte{}
	pos := len(buf)
	i := int64(*n)
	signed := i < 0
	if signed {
		i = -i
	}
	for {
		pos--
		buf[pos], i = '0'+byte(i%10), i/10
		if i == 0 {
			if signed {
				pos--
				buf[pos] = '-'
			}
			return string(buf[pos:])
		}
	}
}

func GetEC2Data(cfg aws.Config) []*EC2Data {
	ec2client := ec2.NewFromConfig(cfg)
	ec2input := ec2.DescribeInstancesInput{}
	ec2output, ec2err := ec2client.DescribeInstances(context.TODO(), &ec2input)
	check(ec2err)
	var ec2datas []*EC2Data = make([]*EC2Data, 0)

	for _, data := range ec2output.Reservations {

		if len(data.Instances) == 1 {
			if *data.Instances[0].State.Code != 48 {
				//fmt.Println(*data.Instances[0].InstanceId)
				//fmt.Println(*data.Instances[0].State.Code)

				//fmt.Println(*data.Instances[0].PrivateIpAddress)
				ec2data := EC2Data{
					InstanceId:      *data.Instances[0].InstanceId,
					PrivateIP:       *data.Instances[0].PrivateIpAddress,
					InstanceType:    string(data.Instances[0].InstanceType),
					PlatformDetails: string(data.Instances[0].Platform),
					Status:          string(data.Instances[0].State.Name),
				}

				for _, tag := range data.Instances[0].Tags {
					if *tag.Key == "Name" {
						ec2data.InstanceName = *tag.Value
					}

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
				ec2datas = append(ec2datas, &ec2data)
			}
		} else {
			/*
				EKS 노드그룹
				for _, data2 := range data.Instances {
					log.Println(data2.InstanceId)
				}
			*/
		}
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
	log.Println("start program")
	accountlist := ReadAccount()
	//DescribeRecord(context.TODO(), accountlist)
	resultresources := DescribeRecord(context.TODO(), accountlist)
	title := "resource_state_"
	filename := MakeExcelReport(title, resultresources)
	log.Println(filename)
}

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
