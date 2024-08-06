package main

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/costexplorer"
	costextype "github.com/aws/aws-sdk-go-v2/service/costexplorer/types"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

var accountnum string
var bucket_name string

func init() {
	accountnum = os.Getenv("ACCOUNTNUM")
	bucket_name = os.Getenv("BUCKETNAME")
}

func main() {
	lambda.Start(InvoiceMonthlyReport)
}

func InvoiceMonthlyReport() {
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion("ap-northeast-2"))
	if err != nil {
		log.Fatalf("failed to load configuration, %v", err)
	}
	client := s3.NewFromConfig(cfg)

	listLinkedAccount := ReadAccount(client)
	exchangeratedata := ReadExchangeRateData(client)
	customerdata := ReadCustomerData(client)
	countrydata := ReadCountryData(client)
	countrycfdata := ReadCountryCFData(client)

	dateofdata := time.Now().AddDate(0, -1, 0)
	monthlyresultdata := make([]MonthlyResultData, 0)
	for i := 0; i < 5; i++ {
		var tempMonthlyResultData MonthlyResultData
		linkedAccountID, detailServices, region, refactoredInvoiceData, err := invoiceHandler(countrydata, countrycfdata, dateofdata, client)
		if err != nil {
			log.Println("no key")
			break
		}
		output := makeInvoiceReport(linkedAccountID, detailServices, region, refactoredInvoiceData, customerdata, exchangeratedata, dateofdata)
		if i == 0 {
			templateInvoiceReport(output, client)
		}
		tempMonthlyResultData.Month = int(dateofdata.Month())
		tempMonthlyResultData.ResultData = output
		monthlyresultdata = append(monthlyresultdata, tempMonthlyResultData)
		dateofdata = dateofdata.AddDate(0, -1, 0)
	}
	monthlyreport := monthlyReportHandler(monthlyresultdata, customerdata, listLinkedAccount, countrydata)

	templateMonthlyReport(monthlyreport, client)
}

func invoiceHandler(countrydata []CountryData, countrydatacf []CountryDataCF, date time.Time, client *s3.Client) ([]string, []string, []string, []RefactoredInvoiceData, error) {
	var month int = int(date.Month())
	input := s3.GetObjectInput{
		Bucket:              aws.String(bucket_name),
		Key:                 aws.String("data/billing/ecsv_" + strconv.Itoa(month) + "_" + strconv.Itoa(date.Year()) + ".csv"),
		ExpectedBucketOwner: aws.String(accountnum),
	}
	output, outputerr := client.GetObject(context.TODO(), &input)
	if outputerr != nil {
		log.Printf("failed to get object, %v", outputerr)
		return nil, nil, nil, nil, outputerr
	}

	invoicedata, err := ReadCsv(output.Body)
	if err != nil {
		check(err)
	}

	var indata []InvoiceData
	for _, e := range invoicedata {
		tmpcost, err := strconv.ParseFloat(e[28], 32)
		if err != nil {
			//fmt.Println(err)
		}
		ignore := ignoreinvoice(tmpcost)
		if !ignore {
			cdata := InvoiceData{
				InvoiceID:              e[0],
				PayerAccountId:         e[1],
				LinkedAccountId:        e[2],
				RecordType:             e[3],
				RecordID:               e[4],
				BillingPeriodStartDate: e[5],
				BillingPeriodEndDate:   e[6],
				InvoiceDate:            e[7],
				PayerAccountName:       e[8],
				LinkedAccountName:      e[9],
				TaxationAddress:        e[10],
				PayerPONumber:          e[11],
				ProductCode:            e[12],
				ProductName:            e[13],
				SellerOfRecord:         e[14],
				UsageType:              e[15],
				Operation:              e[16],
				RateId:                 e[17],
				ItemDescription:        e[18],
				UsageStartDate:         e[19],
				UsageEndDate:           e[20],
				UsageQuantity:          e[21],
				BlendedRate:            e[22],
				CurrencyCode:           e[23],
				CostBeforeTax:          e[24],
				Credits:                e[25],
				TaxAmount:              e[26],
				TaxType:                e[27],
				TotalCost:              e[28],
			}
			indata = append(indata, cdata)
		}
	}

	var refactoredInvoiceData []RefactoredInvoiceData
	var linkedAccountID []string
	var detailServices []string
	var region []string
	for _, e := range indata {
		countrysearchresult := GetCountryName(e.ProductCode, e.UsageType, e.ItemDescription, countrydata, countrydatacf)

		detailservice := GetDetailServiceName(e.ProductName)

		tmpUsageQuantity, err := strconv.ParseFloat(e.UsageQuantity, 32)
		if err != nil {
			//fmt.Println(err)
		}
		tmpCostBeforeTax, err := strconv.ParseFloat(e.CostBeforeTax, 32)
		if err != nil {
			//fmt.Println(err)
		}

		tmpCredits, err := strconv.ParseFloat(e.Credits, 32)
		if err != nil {
			//fmt.Println(err)
		}
		tmpTaxAmount, err := strconv.ParseFloat(e.TaxAmount, 32)
		if err != nil {
			//fmt.Println(err)
		}
		tmpTotalCost, err := strconv.ParseFloat(e.TotalCost, 32)
		if err != nil {
			//fmt.Println(err)
		}
		linkedAccountID = append(linkedAccountID, e.LinkedAccountId)
		detailServices = append(detailServices, detailservice)
		region = append(region, countrysearchresult)
		cdata := RefactoredInvoiceData{
			LinkedAccountId:        e.LinkedAccountId,
			RecordID:               e.RecordID,
			BillingPeriodStartDate: e.BillingPeriodStartDate,
			BillingPeriodEndDate:   e.BillingPeriodEndDate,
			InvoiceDate:            e.InvoiceDate,
			LinkedAccountName:      e.LinkedAccountName,
			ProductCode:            e.ProductCode,
			ProductName:            detailservice,
			SellerOfRecord:         e.SellerOfRecord,
			UsageType:              e.UsageType,
			Operation:              e.Operation,
			RateId:                 e.RateId,
			ItemDescription:        e.ItemDescription,
			Region:                 countrysearchresult,
			DetailService:          detailservice,
			UsageStartDate:         e.UsageStartDate,
			UsageEndDate:           e.UsageEndDate,
			UsageQuantity:          tmpUsageQuantity,
			CurrencyCode:           e.CurrencyCode,
			CostBeforeTax:          tmpCostBeforeTax,
			Credits:                tmpCredits,
			TaxAmount:              tmpTaxAmount,
			TaxType:                e.TaxType,
			TotalCost:              tmpTotalCost,
		}
		refactoredInvoiceData = append(refactoredInvoiceData, cdata)
	}

	return linkedAccountID, detailServices, region, refactoredInvoiceData, nil
}

func monthlyReportHandler(monthlyresultdata []MonthlyResultData, customerdata []CustomerData, listLinkedAccount []LinkedCredential, countrydata []CountryData) []MonthlyReport {
	listrecori := GetRecommendationsRI()
	listmonthlyreport := make([]MonthlyReport, 0)

	for _, account := range listLinkedAccount {
		var cusdata CustomerData
		for _, data := range customerdata {
			if data.CustomerLinkedAccountId == account.LinkedAccount {
				cusdata = data
				break
			}
		}
		recori := make([]TypeRecommendation, 0)
		listcountrycode := make([]string, 0)
		serviceusage, totalserviceusage, listregion := GetServiceUsage(account, monthlyresultdata)
		for _, data := range listregion {
			for _, data1 := range countrydata {
				if data == data1.CountryName {
					listcountrycode = append(listcountrycode, data1.CountryCode)
				}
			}
		}
		listcountrycode = makeSliceUnique(listcountrycode)
		unusedeiplist := make([]UnusedEip, 0)
		unattachedvolume := make([]UnattachedVolume, 0)
		for _, data := range listcountrycode {
			config := GetConfig(account, data)
			unusedeiplist = append(unusedeiplist, GetUnusedEIP(account, config, data)...)
			unattachedvolume = append(unattachedvolume, GetUnattachedVolume(account, config, data)...)
		}
		for _, data := range listrecori {
			if data.LinkedAccountID == account.LinkedAccount {
				recori = append(recori, data)
			}
		}
		monthlyReport := MonthlyReport{
			LinkedAccount:         account.LinkedAccount,
			CustomerName:          cusdata.CustomerName,
			CustomerEmail:         cusdata.CustomerEmail,
			CustomerKoreaName:     cusdata.CustomerKoreaName,
			CustomerAddress:       cusdata.CustomerAddress,
			CustomerInChargeName:  cusdata.CustomerInChargeName,
			CustomerInChargePhone: cusdata.CustomerInChargePhone,
			YearofReport:          strconv.Itoa(time.Now().Year()),
			MonthofReport:         strconv.Itoa(int(time.Now().Month())),
			TotalServiceUsage:     totalserviceusage,
			ServiceUsage:          serviceusage,
			UnusedEip:             unusedeiplist,
			UnattachedVolume:      unattachedvolume,
			RIRecomendation:       recori,
		}
		listmonthlyreport = append(listmonthlyreport, monthlyReport)
	}

	return listmonthlyreport
}

func makeInvoiceReport(linkedAccountID []string, detailServices []string, region []string, refactoredInvoiceData []RefactoredInvoiceData, customerdata []CustomerData, exchangeratedata []CurrencyRate, reportdate time.Time) []ResultData {

	linkedAccountID = makeSliceUnique(linkedAccountID)
	detailServices = makeSliceUnique(detailServices)
	region = makeSliceUnique(region)

	////////////////////////보고서 생성에 필요한 데이터 가져와서 각 계정별 서비스, 국가, 총 사용비용 등 구하기////////////////////////////
	if len(linkedAccountID) != 0 {
		var resultdata = make([]ResultData, len(linkedAccountID))

		for index, account := range linkedAccountID {

			var cusInvoiceData []RefactoredInvoiceData
			var serviceInvoiceData []RefactoredInvoiceData
			var serviceAndRegionInvoiceData []RefactoredInvoiceData
			var customerInfo []CustomerData
			var totalcostbeforetax float64 = 0
			var totalcostaftertax float64 = 0
			var servicecount int = 0
			var servicecountindex int = 0

			cusInvoiceData = GetCustomerInvoiceData(account, refactoredInvoiceData) //Invoice별
			customerInfo = GetCustomerName(account, customerdata)                   // LinkedAccount계정으로 회사정보 알아오기

			if len(customerInfo) != 0 {
				//// 알아온 회사 정보를 구조체에 넣기
				resultdata[index].CustomerEmailAddress = make([]CustomerEmailAddress, len(customerInfo))
				for emailindex, i := range customerInfo {
					resultdata[index].CustomerEmailAddress[emailindex].Email = i.CustomerEmail
				}
				var duedate time.Time
				switch customerInfo[0].DueDate {
				case "current":
					duedate = time.Date(time.Now().Year(), time.Now().Month()+1, 1, 0, 0, 0, 0, time.Local).AddDate(0, 0, -1)
				case "nextmiddle":
					duedate = time.Date(time.Now().Year(), time.Now().Month()+1, 15, 0, 0, 0, 0, time.Local)
				case "nextend":
					duedate = time.Date(time.Now().Year(), time.Now().Month()+2, 1, 0, 0, 0, 0, time.Local).AddDate(0, 0, -1)
				}
				resultdata[index].LinkedAccountID = account
				resultdata[index].LinkedAccountName = customerInfo[0].CustomerName
				resultdata[index].CustomerIncharge = customerInfo[0].CustomerInChargeName
				resultdata[index].CustomerKoreaName = customerInfo[0].CustomerKoreaName
				resultdata[index].CustomoreAddress = customerInfo[0].CustomerAddress
				resultdata[index].Customerid = customerInfo[0].CustomerName
				resultdata[index].StartDate = cusInvoiceData[0].BillingPeriodStartDate
				resultdata[index].EndDate = cusInvoiceData[0].BillingPeriodEndDate
				resultdata[index].DueDate = duedate.String()
				resultdata[index].TaxAmount = "10"
				resultdata[index].ExchangeRate = GetCurrencyRate(exchangeratedata, reportdate)
				///////////////////////// 해당 계정의 사용 서비스 개수 세기///////////////
				for _, data := range detailServices {
					if data != "" {
						serviceInvoiceData = GetCustomerServiceInvoiceData(account, data, refactoredInvoiceData)
						if len(serviceInvoiceData) != 0 {
							servicecount += 1
						}
					}
				}
				/////////////////// 해당 계정이 사용하는 서비스를 조회해서 서비스당 부과되는 요금 및 크레딧 사용 구하기

				if servicecount != 0 { ///// 조회되는 서비스가 있다면
					resultdata[index].Services = make([]Service, 0)
					for _, data := range detailServices {
						serviceInvoiceData = GetCustomerServiceInvoiceData(account, data, refactoredInvoiceData)
						if len(serviceInvoiceData) != 0 { /////////해당 링크드 계정의 서비스로 조회된게 없으면 넘어가기/////////
							var runTotalCode bool = true
							var tmpService Service
							var tmpserviceTotalCostCreditUsageBefore float64 = 0
							var creditusage float64 = 0
							var tmpserviceTotalCostCreditUsageAfter float64 = 0
							tmpService.DetailService = serviceInvoiceData[0].DetailService

							for _, servicedata := range serviceInvoiceData {
								if servicedata.RecordID != "AccountTotal:"+servicedata.LinkedAccountId {
									var regioncount int = 0

									creditusage += servicedata.Credits
									tmpserviceTotalCostCreditUsageBefore += servicedata.CostBeforeTax
									for _, regiondata := range region { //서비스 내에 리전이 있는지 조회

										serviceAndRegionInvoiceData = GetCustomerServiceAndRegionInvoiceData(account, data, regiondata, refactoredInvoiceData)
										if len(serviceAndRegionInvoiceData) != 0 {
											regioncount += 1
										}
									}

									if regioncount != 0 { //서비스내에 리전별 사용이 존재하다면
										tmpService.Region = make([]Region, 0)
										for _, regiondata := range region {
											var tmpRegion Region
											serviceAndRegionInvoiceData = GetCustomerServiceAndRegionInvoiceData(account, data, regiondata, refactoredInvoiceData)
											if len(serviceAndRegionInvoiceData) != 0 {
												var regionservicetotalcost float64
												tmpRegion.Region = serviceAndRegionInvoiceData[0].Region
												tmpRegion.Description = make([]Description, 0)

												for _, serviceregiondata := range serviceAndRegionInvoiceData {
													var tmpRegionDescription Description
													tmpRegionDescription.Description = serviceregiondata.ItemDescription
													tmpRegionDescription.DescriptionTotalCost = fmt.Sprintf("%10.2f", serviceregiondata.CostBeforeTax)
													regionservicetotalcost += serviceregiondata.CostBeforeTax
													tmpRegion.Description = append(tmpRegion.Description, tmpRegionDescription)
												}
												tmpRegion.RegionTotalCost = fmt.Sprintf("%10.2f", regionservicetotalcost)
												tmpService.Region = append(tmpService.Region, tmpRegion)

											}
										}
									} else {
										tmpService.Region = make([]Region, 1)
										tmpService.Region[0].Region = "Global"
									}

								} else {
									runTotalCode = false
								}

							}

							////////////////////////각 계정의 서비스별 총 사용량 구하기///////////////////
							if runTotalCode {
								applyspp := true
								if totalcostbeforetax < 0 {
									totalcostbeforetax = 0
								}
								creditusage = checkZero(math.Round(creditusage*100) / 100)
								tmpserviceTotalCostCreditUsageBefore = checkZero(math.Round(tmpserviceTotalCostCreditUsageBefore*100) / 100)
								//SPP 값을 구하기 위해 먼저 크레딧 적용된 서비스의 사용량을 구해야함
								tmpserviceTotalCostCreditUsageAfter += tmpserviceTotalCostCreditUsageBefore + creditusage
								if tmpserviceTotalCostCreditUsageAfter < 0 {
									tmpserviceTotalCostCreditUsageAfter = 0
								}

								if strings.Contains(tmpService.DetailService, "Support") || strings.Contains(tmpService.DetailService, "Registrar") {
									applyspp = false
									///Support, Registrar 서비스에 대해서는 spp 적용 제외
								}

								if applyspp { ////SPP 계산////
									spp, err := strconv.ParseFloat(customerInfo[0].SPPDiscount, 32)
									if err != nil {
										log.Printf("err: %s", err.Error())
									}

									sppdiscount := tmpserviceTotalCostCreditUsageAfter * 100 / (100 - spp) * (spp / 100)
									//소수 2번째 자리에서 반올림 후 더하기
									sppdiscount = checkZero(math.Round(sppdiscount*100) / 100)
									tmpserviceTotalCostCreditUsageBefore += sppdiscount
									tmpserviceTotalCostCreditUsageAfter = tmpserviceTotalCostCreditUsageBefore + creditusage
									if tmpserviceTotalCostCreditUsageAfter < 0 {
										tmpserviceTotalCostCreditUsageAfter = 0
									}
								}

								tmpService.ServiceTotalCostBeforeCredit = fmt.Sprintf("$ %10.2f", tmpserviceTotalCostCreditUsageBefore)
								tmpService.ServiceTotalCostBeforeCreditrow = tmpserviceTotalCostCreditUsageBefore
								tmpService.CreaditUsage = fmt.Sprintf("$ %10.2f", creditusage)
								totalcostbeforetax += tmpserviceTotalCostCreditUsageBefore + creditusage
								tmpService.ServiceTotalCostAfterCredit = fmt.Sprintf("$ %10.2f", tmpserviceTotalCostCreditUsageAfter)
								servicecountindex += 1
								resultdata[index].Services = append(resultdata[index].Services, tmpService)
							}

						}
					}

					if strings.Contains(customerInfo[0].ManagedService, "managed") {
						resultdata[index].ManagedService = fmt.Sprintf("%10.2f", checkZero(math.Round(totalcostbeforetax*10)/100))
						totalcostbeforetax += totalcostbeforetax * 0.1
					} else {
						resultdata[index].ManagedService = "0"
					}
					resultdata[index].TotalCostBeforeTax = fmt.Sprintf("%10.2f", checkZero(math.Round(totalcostbeforetax*100)/100))
					totalcostaftertax = totalcostbeforetax * 1.1
					resultdata[index].TaxAmount = fmt.Sprintf("%10.2f", checkZero(math.Round(math.Abs(totalcostbeforetax*0.1)*100)/100))
					resultdata[index].TotalCostAfterTax = fmt.Sprintf("%10.2f", checkZero(totalcostaftertax))
					totalkrw, totalkrwerr := strconv.ParseFloat(strings.TrimSpace(resultdata[index].ExchangeRate.ExchangeRate), 32)
					check(totalkrwerr)
					resultdata[index].ReportMonth = GetReportMonth(strings.TrimSpace(resultdata[index].ExchangeRate.Month))
					resultdata[index].TotalCostAfterTaxKRW = Commaf(checkZero(math.Round(totalcostaftertax * totalkrw)))
				}
			}

			///////////////////// 템플릿 만들기 //////////////////////////

		}
		return resultdata

	} else {
		return nil
	}

}

func templateMonthlyReport(monthlyreport []MonthlyReport, client *s3.Client) {
	//date := time.Now().AddDate(0, -1, 0)
	mkdirerr := os.MkdirAll("/tmp/result/monthly/html/", 0770)
	check(mkdirerr)
	for _, data := range monthlyreport {
		t, err := template.New(data.CustomerKoreaName).Parse(monthlyreporttemplatesample)
		check(err)
		f, err := os.Create("/tmp/result/monthly/html/" + "월간보고서" + "_" + data.CustomerKoreaName + "_" + strconv.Itoa(int(time.Now().Month())) + "월_" + strconv.Itoa(time.Now().Year()) + "년_" + data.LinkedAccount + ".html")

		check(err)
		err = t.Execute(f, data)
		check(err)
	}
	newfile, err := os.Create("/tmp/result/monthly/monthly_report_result.zip")
	if err != nil {
		panic(err)
	}

	ZipDir(newfile, "/tmp/result/monthly/html/")

	file, err := os.Open("/tmp/result/monthly/monthly_report_result.zip")
	defer file.Close()

	putinput := s3.PutObjectInput{
		Bucket: aws.String(bucket_name),
		Key:    aws.String("result/monthly/" + time.Now().Local().Format("2006-01-02T15:05:37") + "/monthly_report_result.zip"),
		Body:   file,
	}
	putoutput, err := client.PutObject(context.Background(), &putinput)
	check(err)
	if putoutput != nil {

	}
}

func templateInvoiceReport(resultdatas []ResultData, client *s3.Client) {
	mkdirerr := os.MkdirAll("/tmp/result/invoice/html", 0770)
	check(mkdirerr)
	for _, data := range resultdatas {
		t, err := template.New(data.CustomerKoreaName).Parse(invoicetemplatesample)
		check(err)

		f, err := os.Create("/tmp/result/invoice/html/" + data.CustomerKoreaName + "_" + data.LinkedAccountName + "_" + strconv.Itoa(int(time.Now().Month())) + "월_" + strconv.Itoa(time.Now().Year()) + "년_" + data.LinkedAccountID + ".html")
		check(err)

		err = t.Execute(f, data)
		check(err)
	}
	newfile, err := os.Create("/tmp/result/invoice/invoice_result.zip")
	if err != nil {
		panic(err)
	}

	ZipDir(newfile, "/tmp/result/invoice/html/")

	file, err := os.Open("/tmp/result/invoice/invoice_result.zip")
	defer file.Close()
	putinput := s3.PutObjectInput{
		Bucket: aws.String(bucket_name),
		Key:    aws.String("result/invoice/" + time.Now().Local().Format("2006-12-02_15:05:37") + "/invoice_result.zip"),
		Body:   file,
	}
	putoutput, err := client.PutObject(context.Background(), &putinput)
	check(err)
	log.Println(*putoutput)
}

func GetRecommendationsRI() []TypeRecommendation {
	cfg, err := config.LoadDefaultConfig(context.Background(), config.WithRegion("ap-northeast-2"))
	if err != nil {
		log.Printf("failed to load configuration, %v", err)
	}
	client := costexplorer.NewFromConfig(cfg)
	input := costexplorer.GetReservationPurchaseRecommendationInput{
		Service:      aws.String("Amazon Elastic Compute Cloud - Compute"),
		AccountId:    aws.String("708595888134"),
		AccountScope: costextype.AccountScopeLinked.Values()[1],
	}
	output, err := client.GetReservationPurchaseRecommendation(context.Background(), &input)
	if err != nil {
		log.Printf("Get RI Recommend Error: %s", err.Error())
		return nil
	}
	result := make([]TypeRecommendation, 0)
	for _, data := range output.Recommendations {
		for _, data1 := range data.RecommendationDetails {
			estimatedondemandcost, estimatedondemandcosterr := strconv.ParseFloat(*data1.EstimatedMonthlyOnDemandCost, 32)
			check(estimatedondemandcosterr)
			estimatedmonthlysavingamount, estimatedmonthlysavingamounterr := strconv.ParseFloat(*data1.EstimatedMonthlySavingsAmount, 32)
			check(estimatedmonthlysavingamounterr)
			estimatedsavingpercentage, estimatedsavingpercentageerr := strconv.ParseFloat(*data1.EstimatedMonthlySavingsPercentage, 32)
			check(estimatedsavingpercentageerr)
			temp := TypeRecommendation{
				LinkedAccountID:                        *data1.AccountId,
				MonthlyOndemandCost:                    fmt.Sprintf("%.2f", math.Round(estimatedondemandcost*10)/10),
				SavingAmount:                           fmt.Sprintf("%.2f", math.Round(estimatedmonthlysavingamount*10)/10),
				SavingPercentage:                       fmt.Sprintf("%.2f", math.Round(estimatedsavingpercentage*10)/10),
				CurrentInstanceType:                    *data1.InstanceDetails.EC2InstanceDetails.InstanceType,
				RecommendedNumberOfInstancesToPurchase: *data1.RecommendedNumberOfInstancesToPurchase,
			}
			result = append(result, temp)
		}
	}
	return result
}

func GetUnattachedVolume(linkedcredential LinkedCredential, config *aws.Config, countrycode string) []UnattachedVolume {
	client := ec2.NewFromConfig(*config)
	input := ec2.DescribeVolumesInput{}
	output, err := client.DescribeVolumes(context.TODO(), &input)
	if err != nil {
		log.Printf("Get Unattached Volume Error(%s): %s", linkedcredential.LinkedAccount, err.Error())
		return nil
	}
	listunattachedvolume := make([]UnattachedVolume, 0)
	for _, data := range output.Volumes {
		var unattachedvolume UnattachedVolume
		if data.State == data.State.Values()[1] {
			unattachedvolume.UnattachedVolumeID = *data.VolumeId
			unattachedvolume.UnattachedVolumeStatus = string(data.State)
			unattachedvolume.Countrycode = countrycode
			listunattachedvolume = append(listunattachedvolume, unattachedvolume)
		}
	}
	return listunattachedvolume
}

func GetUnusedEIP(linkedcredential LinkedCredential, config *aws.Config, countrycode string) []UnusedEip {
	client := ec2.NewFromConfig(*config)
	result, err := client.DescribeAddresses(context.Background(), &ec2.DescribeAddressesInput{})
	if err != nil {
		log.Printf("Get Unused EIP Error(%s): %s", linkedcredential.LinkedAccount, err.Error())
		return nil
	}
	listunusedeip := make([]UnusedEip, 0)
	if len(result.Addresses) != 0 {
		for _, data := range result.Addresses {
			if data.AssociationId == nil {
				if data.InstanceId == nil || *data.InstanceId == "" {
					var unusedeip UnusedEip
					unusedeip.Eip = *data.PublicIp
					if data.AllocationId == nil {
						unusedeip.AllocationID = ""
					} else {
						unusedeip.AllocationID = *data.AllocationId
					}

					unusedeip.Countrycode = countrycode
					listunusedeip = append(listunusedeip, unusedeip)
				}
			}
		}
	}

	return listunusedeip
}

func GetServiceUsage(account LinkedCredential, monthlyresultdata []MonthlyResultData) ([]ServiceUsage, ServiceUsage, []string) {
	listserviceusage := make([]ServiceUsage, 0)
	services := make([]string, 0)
	var totalserviceusage ServiceUsage
	listregion := make([]string, 0)
	listtotalmonthlyservice := make([]MonthlyService, 0)
	for i := range monthlyresultdata {
		var monthlyservices MonthlyService
		for _, data := range monthlyresultdata[len(monthlyresultdata)-1-i].ResultData {
			if account.LinkedAccount == data.LinkedAccountID {
				for _, service := range data.Services {
					if service.ServiceTotalCostBeforeCreditrow > 0.1 {
						services = append(services, service.DetailService)
						monthlyservices.ServiceUsage += service.ServiceTotalCostBeforeCreditrow
					}
					for _, region := range service.Region {
						listregion = append(listregion, region.Region)
					}
				}
			}
		}
		monthlyservices.Month = monthlyresultdata[len(monthlyresultdata)-1-i].Month
		listtotalmonthlyservice = append(listtotalmonthlyservice, monthlyservices)
	}
	listregion = makeSliceUnique(listregion)
	services = makeSliceUnique(services)
	totalserviceusage.ServiceName = "Total Cost"
	totalserviceusage.ServiceDataName = "data" + strconv.Itoa(len(services)+1)
	totalserviceusage.ServiceConfigName = "config" + strconv.Itoa(len(services)+1)
	totalserviceusage.ServiceChartName = "chart" + strconv.Itoa(len(services)+1)
	totalserviceusage.Service = listtotalmonthlyservice
	totalserviceusage.ServiceYear = strconv.Itoa(time.Now().Year())

	for index, service := range services {
		var serviceusage ServiceUsage
		listmonthlyservice := make([]MonthlyService, 0)
		serviceusage.ServiceName = service
		serviceusage.ServiceDataName = "data" + strconv.Itoa(index+1)
		serviceusage.ServiceConfigName = "config" + strconv.Itoa(index+1)
		serviceusage.ServiceChartName = "chart" + strconv.Itoa(index+1)
		serviceusage.ServiceYear = strconv.Itoa(time.Now().Year())
		for i := range monthlyresultdata {
			for _, data := range monthlyresultdata[len(monthlyresultdata)-1-i].ResultData {
				if account.LinkedAccount == data.LinkedAccountID {
					for _, data2 := range data.Services {
						if data2.DetailService == service {
							var monthlyservice MonthlyService
							monthlyservice.Month = monthlyresultdata[len(monthlyresultdata)-1-i].Month
							monthlyservice.ServiceUsage = math.Round(data2.ServiceTotalCostBeforeCreditrow*10) / 10
							if len(monthlyresultdata)-1-i == 0 {
								serviceusage.ServiceCost = math.Round(data2.ServiceTotalCostBeforeCreditrow*10) / 10
							}
							listmonthlyservice = append(listmonthlyservice, monthlyservice)
						}
					}
				}
			}
		}
		serviceusage.Service = listmonthlyservice
		listserviceusage = append(listserviceusage, serviceusage)
	}
	return listserviceusage, totalserviceusage, listregion
}

func GetConfig(linkedcredential LinkedCredential, countrycode string) *aws.Config {
	cfg, err := config.LoadDefaultConfig(
		context.Background(),
		config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(linkedcredential.AccessKey, linkedcredential.SecretKey, ""),
		),
		config.WithRegion(countrycode),
	)
	check(err)
	return &cfg
}

func GetReportMonth(reportmonth string) string {
	switch reportmonth {
	case "1":
		return "January"
	case "2":
		return "February"
	case "3":
		return "March"
	case "4":
		return "April"
	case "5":
		return "May"
	case "6":
		return "June"
	case "7":
		return "July"
	case "8":
		return "August"
	case "9":
		return "September"
	case "10":
		return "October"
	case "11":
		return "November"
	case "12":
		return "December"
	default:
		return ""
	}
}

func ReadCountryData(client *s3.Client) []CountryData {
	input := s3.GetObjectInput{
		Bucket:              aws.String(bucket_name),
		Key:                 aws.String("data/countrydata/countrydata.csv"),
		ExpectedBucketOwner: aws.String(accountnum),
	}
	output, outputerr := client.GetObject(context.TODO(), &input)
	if outputerr != nil {
		log.Fatalf("failed to get object, %v", outputerr)
	}
	countrydata, err := ReadCsv(output.Body)
	if err != nil {
		panic(err)
	}

	var codata []CountryData
	for _, e := range countrydata {
		cdata := CountryData{
			CountrySite:  e[0],
			CountryName:  e[1],
			CountryCode:  e[2],
			CountryCode2: e[3],
		}
		codata = append(codata, cdata)
	}
	return codata
}

func ReadCountryCFData(client *s3.Client) []CountryDataCF {
	input := s3.GetObjectInput{
		Bucket:              aws.String(bucket_name),
		Key:                 aws.String("data/countrydata/countrydata2.csv"),
		ExpectedBucketOwner: aws.String(accountnum),
	}
	output, outputerr := client.GetObject(context.TODO(), &input)
	if outputerr != nil {
		log.Fatalf("failed to get object, %v", outputerr)
	}
	countrydatacf, err := ReadCsv(output.Body)
	if err != nil {
		panic(err)
	}

	var codatacf []CountryDataCF
	for _, e := range countrydatacf {
		cdata := CountryDataCF{
			CountryName: e[0],
			CountryCode: e[1],
		}
		codatacf = append(codatacf, cdata)
	}
	return codatacf
}

func ReadAccount(client *s3.Client) []LinkedCredential {
	input := s3.GetObjectInput{
		Bucket:              aws.String(bucket_name),
		Key:                 aws.String("data/credential/account.csv"),
		ExpectedBucketOwner: aws.String(accountnum),
	}
	output, outputerr := client.GetObject(context.TODO(), &input)
	if outputerr != nil {
		log.Fatalf("failed to get object, %v", outputerr)
	}
	linkedcredentialdatas, err := ReadCsv(output.Body)
	if err != nil {
		panic(err)
	}
	var result []LinkedCredential
	for _, e := range linkedcredentialdatas {
		linkedcredentialdata := LinkedCredential{
			LinkedAccount: e[0],
			AccessKey:     e[1],
			SecretKey:     e[2],
		}
		result = append(result, linkedcredentialdata)
	}

	return result
}

func ReadExchangeRateData(client *s3.Client) []CurrencyRate {
	input := s3.GetObjectInput{
		Bucket:              aws.String(bucket_name),
		Key:                 aws.String("data/exchangerate/exchange_rate.csv"),
		ExpectedBucketOwner: aws.String(accountnum),
	}
	output, outputerr := client.GetObject(context.TODO(), &input)
	if outputerr != nil {
		log.Fatalf("failed to get object, %v", outputerr)
	}
	exchangedata, err := ReadCsv(output.Body)
	if err != nil {
		panic(err)
	}

	var ratedata []CurrencyRate
	for _, e := range exchangedata {
		rdata := CurrencyRate{
			Year:         e[0],
			Month:        e[1],
			ExchangeRate: e[2],
		}
		ratedata = append(ratedata, rdata)
	}
	return ratedata
}

func ReadCustomerData(client *s3.Client) []CustomerData {
	input := s3.GetObjectInput{
		Bucket:              aws.String(bucket_name),
		Key:                 aws.String("data/customerdata/Contacts_CustomerInChargeName.csv"),
		ExpectedBucketOwner: aws.String(accountnum),
	}
	output, outputerr := client.GetObject(context.TODO(), &input)
	if outputerr != nil {
		log.Fatalf("failed to get object, %v", outputerr)
	}
	contactdata, err := ReadCsv(output.Body)
	if err != nil {
		panic(err)
	}

	var rcontactdata []ReadContact
	startReadData := false
	contactindex := make([]int, 0)
	for _, e := range contactdata {
		if startReadData == true {
			cdata := ReadContact{
				CustomerInChargeName:  e[contactindex[0]],
				CustomerInChargePhone: e[contactindex[1]],
				CustomerEmail:         e[contactindex[2]],
				LinkedAccountID:       strings.Replace(e[contactindex[3]], "i", "", 1),
			}
			rcontactdata = append(rcontactdata, cdata)
		}

		if e[0] == "Name" {
			for index, data := range e {
				switch data {
				case "Name":
					contactindex = append(contactindex, index)
				case "CustomerInChargePhone":
					contactindex = append(contactindex, index)
				case "CustomerEmail":
					contactindex = append(contactindex, index)
				case "AWS 계정":
					contactindex = append(contactindex, index)
				}
			}
			startReadData = true
		}

	}

	input = s3.GetObjectInput{
		Bucket:              aws.String(bucket_name),
		Key:                 aws.String("data/customerdata/CustomerKoreaName.csv"),
		ExpectedBucketOwner: aws.String(accountnum),
	}
	output, outputerr = client.GetObject(context.TODO(), &input)
	if outputerr != nil {
		log.Fatalf("failed to get object, %v", outputerr)
	}

	customerdata, err := ReadCsv(output.Body)
	if err != nil {
		panic(err)
	}

	var rcustomerdata []ReadCustomer
	startReadData = false
	customerindex := make([]int, 0)
	for _, e := range customerdata {
		if startReadData == true {
			if e[customerindex[0]] == "" {
				break
			}

			cdata := ReadCustomer{
				LinkedAccountID:   strings.Replace(e[contactindex[0]], "i", "", 1),
				CustomerKoreaName: e[customerindex[1]],
				RegisteredName:    e[customerindex[2]],
				RegisteredDate:    e[customerindex[3]],
				ManagedService:    e[customerindex[4]],
				SPPDiscount:       e[customerindex[5]],
				CustomerAddress:   e[customerindex[6]],
				DueDate:           e[customerindex[7]],
			}
			rcustomerdata = append(rcustomerdata, cdata)
		}

		if e[0] == "Name" {
			for index, data := range e {
				switch data {
				case "Name":
					customerindex = append(customerindex, index)
				case "KoreaName":
					customerindex = append(customerindex, index)
				case "RegisteredName":
					customerindex = append(customerindex, index)
				case "RegisteredDate":
					customerindex = append(customerindex, index)
				case "Managed Service":
					customerindex = append(customerindex, index)
				case "CustomerAddress":
					customerindex = append(customerindex, index)
				case "SPP Discount":
					customerindex = append(customerindex, index)
				case "Due Date":
					customerindex = append(customerindex, index)
				}
			}
			startReadData = true
		}
	}

	//fmt.Println(rcontactdata)
	//fmt.Println(rcustomerdata)

	var integrateddata []CustomerData
	for _, condata := range rcontactdata {
		for _, cusdata := range rcustomerdata {
			if strings.Contains(condata.LinkedAccountID, cusdata.LinkedAccountID) {

				singleCutomerData := CustomerData{
					CustomerLinkedAccountId: cusdata.LinkedAccountID,
					CustomerName:            cusdata.RegisteredName,
					CustomerEmail:           condata.CustomerEmail,
					RegisteredDate:          cusdata.RegisteredDate,
					CustomerKoreaName:       cusdata.CustomerKoreaName,
					CustomerAddress:         cusdata.CustomerAddress,
					CustomerInChargeName:    condata.CustomerInChargeName,
					CustomerInChargePhone:   condata.CustomerInChargePhone,
					ManagedService:          cusdata.ManagedService,
					SPPDiscount:             cusdata.SPPDiscount,
					DueDate:                 cusdata.DueDate,
				}
				integrateddata = append(integrateddata, singleCutomerData)
			}
		}
	}

	return integrateddata
}

func ReadCsv(f io.ReadCloser) ([][]string, error) {

	// Read File into a Variable
	lines, err := csv.NewReader(f).ReadAll()
	if err != nil {
		return [][]string{}, err
	}

	return lines, nil
}

func makeSliceUnique(s []string) []string {
	keys := make(map[string]struct{})
	res := make([]string, 0)
	for _, val := range s {
		if _, ok := keys[val]; ok {
			continue
		} else {
			keys[val] = struct{}{}
			res = append(res, val)
		}
	}
	return res
}

func ignoreinvoice(cost float64) bool {
	var epsilon = 0.001

	if cost > epsilon {
		if cost-epsilon <= epsilon {
			return true
		} else {
			return false
		}
	} else {
		if epsilon-cost <= epsilon {
			return true
		} else {
			return false
		}
	}
}

func GetDetailServiceName(productname string) string {
	var result string
	result = strings.Replace(productname, "AWS ", "", 1)
	result = strings.Replace(result, "Amazon ", "", 1)
	result = strings.Replace(result, "Amazon", "", 1)
	return result
}

func GetCountryName(productcode string, usagetype string, description string, countrydata []CountryData, countrydatacf []CountryDataCF) string {

	if productcode == "AmazonCloudFront" {
		for _, data := range countrydatacf {
			if data.CountryCode == description[:1] {
				return data.CountryName
			}
		}
	}

	if productcode == "AmazonRoute53" {
		return "Global"
	}

	if strings.Contains(productcode, "Support") {
		return ""
	}

	if len(usagetype) > 4 {
		for _, data := range countrydata {
			if usagetype[:4] == data.CountryCode2 {
				return data.CountryName
			}
		}

		if usagetype[:1] == "CA" {
			for _, data := range countrydata {
				if data.CountryCode2 == "CAN1" {
					return data.CountryName
				}
			}
		} else if strings.Contains(usagetype, "us-") ||
			strings.Contains(usagetype, "ap-") ||
			strings.Contains(usagetype, "af-") ||
			strings.Contains(usagetype, "ca-") ||
			strings.Contains(usagetype, "eu-") ||
			strings.Contains(usagetype, "sa-") ||
			strings.Contains(usagetype, "me-") {
			var tmpcode string
			switch {
			case strings.Contains(usagetype, "us-east-1"):
				tmpcode = "us-east-1"
			case strings.Contains(usagetype, "us-east-2"):
				tmpcode = "us-east-2"
			case strings.Contains(usagetype, "us-west-1"):
				tmpcode = "us-west-1"
			case strings.Contains(usagetype, "us-west-2"):
				tmpcode = "us-west-2"
			case strings.Contains(usagetype, "af-south-1"):
				tmpcode = "af-south-1"
			case strings.Contains(usagetype, "ap-east-1"):
				tmpcode = "ap-east-1"
			case strings.Contains(usagetype, "ap-south-1"):
				tmpcode = "ap-south-1"
			case strings.Contains(usagetype, "ap-northeast-3"):
				tmpcode = "ap-northeast-3"
			case strings.Contains(usagetype, "ap-northeast-2"):
				tmpcode = "ap-northeast-2"
			case strings.Contains(usagetype, "ap-southeast-1"):
				tmpcode = "ap-southeast-1"
			case strings.Contains(usagetype, "ap-southeast-2"):
				tmpcode = "ap-southeast-2"
			case strings.Contains(usagetype, "ap-northeast-1"):
				tmpcode = "ap-northeast-1"
			case strings.Contains(usagetype, "ca-central-1"):
				tmpcode = "ca-central-1"
			case strings.Contains(usagetype, "eu-central-1"):
				tmpcode = "eu-central-1"
			case strings.Contains(usagetype, "eu-west-1"):
				tmpcode = "eu-west-1"
			case strings.Contains(usagetype, "eu-west-2"):
				tmpcode = "eu-west-2"
			case strings.Contains(usagetype, "eu-south-1"):
				tmpcode = "eu-south-1"
			case strings.Contains(usagetype, "eu-west-3"):
				tmpcode = "eu-west-3"
			case strings.Contains(usagetype, "eu-north-1"):
				tmpcode = "eu-north-1"
			case strings.Contains(usagetype, "me-south-1"):
				tmpcode = "me-south-1"
			case strings.Contains(usagetype, "sa-east-1"):
				tmpcode = "sa-east-1"
			}

			for _, data := range countrydata {
				if data.CountryCode == tmpcode {
					return data.CountryName
				}
			}
		} else {
			for _, data := range countrydata {
				if data.CountryCode2 == "USE1" {
					return data.CountryName
				}
			}
		}
	} else {
		return "Not Registered"
	}
	return "Country Exception"
}

func GetCurrencyRate(exchangeratedata []CurrencyRate, reportdate time.Time) CurrencyRate {

	var currencyrate CurrencyRate
	for _, exchangeindex := range exchangeratedata {
		year, yearerr := strconv.Atoi(exchangeindex.Year)
		check(yearerr)
		month, montherr := strconv.Atoi(strings.TrimSpace(exchangeindex.Month))
		check(montherr)
		exchangeindex.ExchangeRate = strings.TrimSpace(exchangeindex.ExchangeRate)
		if year == reportdate.Year() && month == int(reportdate.Month()) {
			currencyrate = exchangeindex
		}
	}
	return currencyrate
}

func GetCustomerName(customerlinkedaccountid string, customerdata []CustomerData) []CustomerData {
	var result []CustomerData
	for _, data := range customerdata {
		if data.CustomerLinkedAccountId == customerlinkedaccountid {

			result = append(result, data)
		}
	}

	return result
}

func GetCustomerServiceAndRegionInvoiceData(linkedaccount string, productname string, region string, refactoreddata []RefactoredInvoiceData) []RefactoredInvoiceData {
	var result []RefactoredInvoiceData
	for _, data := range refactoreddata {
		if data.LinkedAccountId == linkedaccount && data.ProductName == productname && data.Region == region {
			result = append(result, data)
		}
	}
	return result
}

func GetCustomerServiceInvoiceData(linkedaccount string, productname string, refactoreddata []RefactoredInvoiceData) []RefactoredInvoiceData {
	var result []RefactoredInvoiceData
	for _, data := range refactoreddata {
		if data.LinkedAccountId == linkedaccount && data.ProductName == productname {
			result = append(result, data)
		}
	}

	return result
}

func GetCustomerInvoiceData(linkedaccount string, refactoreddata []RefactoredInvoiceData) []RefactoredInvoiceData {
	var result []RefactoredInvoiceData
	for _, data := range refactoreddata {
		if data.LinkedAccountId == linkedaccount {
			result = append(result, data)
		}
	}
	return result
}

func Commaf(v float64) string {
	buf := &bytes.Buffer{}
	if v < 0 {
		buf.Write([]byte{'-'})
		v = 0 - v
	}

	comma := []byte{','}

	parts := strings.Split(strconv.FormatFloat(v, 'f', -1, 64), ".")
	pos := 0
	if len(parts[0])%3 != 0 {
		pos += len(parts[0]) % 3
		buf.WriteString(parts[0][:pos])
		buf.Write(comma)
	}
	for ; pos < len(parts[0]); pos += 3 {
		buf.WriteString(parts[0][pos : pos+3])
		buf.Write(comma)
	}
	buf.Truncate(buf.Len() - 1)

	if len(parts) > 1 {
		buf.Write([]byte{'.'})
		buf.WriteString(parts[1])
	}
	return buf.String()
}

func checkZero(cost float64) float64 {
	if 0 >= cost {
		return 0
	}
	return cost
}

func ZipDir(saveFile *os.File, savePath string) error {
	zipWriter := zip.NewWriter(saveFile)
	defer zipWriter.Close()
	z := zipType{zipWriter}
	return z.dir(savePath, "")
}

func (z zipType) dir(dirPath string, zipPath string) error {
	files, err := ioutil.ReadDir(dirPath)
	if err != nil {
		return err
	}
	for _, file := range files {
		fullPath := dirPath + string(filepath.Separator) + file.Name()
		if file.IsDir() {
			if err != nil {
				return err
			}
			z.dir(fullPath, zipPath+file.Name()+string(filepath.Separator))
		} else {
			if err := z.file(file, fullPath, zipPath); err != nil {
				return err
			}
		}
	}
	return nil
}

func (z zipType) file(file os.FileInfo, filePath string, zipPath string) error {
	header, _ := zip.FileInfoHeader(file)
	header.Name = zipPath + header.Name

	w, err := z.writer.CreateHeader(header)
	if err != nil {
		return err
	}
	f, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer f.Close()
	io.Copy(w, f)

	return nil
}

func check(err error) {
	if err != nil {
		log.Println(err)
	}
}

const monthlyreporttemplatesample = `
<html style="box-sizing: border-box; -moz-box-sizing: border-box;">

<head style="box-sizing: border-box; -moz-box-sizing: border-box;">
<meta charset="UTF-8">  
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>MonthlyReport</title>
<style>
body {
	width: 100%;
	height: 100%;
	margin: 0;
	padding: 0;
	background-color: #ddd;
}
* {
	box-sizing: border-box;
	-moz-box-sizing: border-box;
}

	.paper_cover {
		width: 210mm;
		min-height: 297mm;
		margin: 10mm auto;
		border-radius: 5px;
		background: #fff;
		box-shadow: 0 0 5px rgba(0, 0, 0, 0.1);
	}

	.paper {
		width: 210mm;
		min-height: 297mm;
		padding: 20mm; /* set contents area */
		margin: 10mm auto;
		background: #fff;
		box-shadow: 0 0 5px rgba(0, 0, 0, 0.1);
	}
	.content_cover {
		padding: 0;
		height: 257mm;
		position: relative;
	}

	.content {
		padding: 0;
		height: 257mm;
	}

	/* cover 페이지 */
	.cover{width: 210mm; min-height: 297mm;}
	.cover_title{position: absolute; top: 80mm; left: 23mm; width: 500px; height: auto;}
	.report{color: rgb(24, 97, 255); font-size: 30px; margin-bottom:20px; float: left;}
	.bar{font-size: 74px; color: rgb(153, 206, 255); margin: 0 13px 0 0; float: left;}
	.year{color: rgb(24, 97, 255); font-family: 'Y_Spotlight'; font-size: 90px; margin:0px;float: none;}
	.m_report{color: rgb(24, 97, 255); font-size: 22px; margin: 0 0 0 20px; font-family: 'Pretendard-Light';}
	.chartMenu p {padding: 10px; font-size: 20px;}
	.company{font-family: 'Eoe_Zno_L'; font-size: 20px; margin: 0 0 0 20px; transform: skew(-0.1deg); line-height:28px;}
	.title{color: #000; font-size: 50px; float: left;}
	.sub{color:#000; font-size: 20px; margin-top: 220px;}

	/* 차트 페이지 머리글*/
	.header_Rectangle{background-color: rgb(211, 211, 211);width: auto; height: 4px;}
	#header_cmt{float: left;}
	#header_year{font-weight: bold;margin-left: 234px;}
	.page_title{font-size: 25px; font-weight: bold; margin: 25px 0px;}
	.page_description{background-color: rgb(231, 231, 231); padding-top: 5px; font-size: 15px; font-family: 'Pretendard-Light'; margin-bottom: 50px;}
	.pre_description{ font-family: 'Pretendard-Regular';}

	/* 표 페이지 머리글 */
	.page_description_Graph{background-color: rgb(231, 231, 231); padding-top: 5px; font-size: 15px; font-family: 'Pretendard-Light'; margin-bottom: 30px;}

	/* 도넛차트 사이즈 */
	.doughnut_size{width: 500px; margin-left: 12%;}

	/* 차트 사이즈 및 컬러 */
	.tftable {font-size:12px;color:#000;width:100%;border-width: 1px;border-color: #a9a9a9;border-collapse: collapse;}
	.tftable th {font-size:12px;background-color:rgb(211, 211, 211);border-width: 1px;padding: 8px;border-style: solid;border-color: #a9a9a9;text-align:center;}
	.tftable tr {background-color:#fff; border-color: #a9a9a9;}
	.tftable td {font-size:12px;border-width: 1px;padding: 8px;border-style: solid;border-color: #a9a9a9;}

	@font-face {
	font-family: 'Y_Spotlight';
	src: url('https://cdn.jsdelivr.net/gh/projectnoonnu/noonfonts-20-12@1.0/Y_Spotlight.woff') format('woff');
	font-weight: normal;
	font-style: normal;
	}
	@font-face {
    font-family: 'Pretendard-Light';
    src: url('https://cdn.jsdelivr.net/gh/Project-Noonnu/noonfonts_2107@1.1/Pretendard-Light.woff') format('woff');
    font-weight: 300;
    font-style: normal;
	}

	
	@page {
		size: A4;
		margin: 0;
	}
	@media print {
		html, body {
			width: 210mm;
			height: 297mm;
			background: rgb(255, 255, 255);
		}
		.paper {
			margin: 0;
			border: initial;
			border-radius: initial;
			width: initial;
			min-height: initial;
			box-shadow: initial;
			background: initial;
			page-break-after: always;
		}
	}
	</style>
	<meta charset="utf-8" />
</head>
<body class="body">
<div>
<!-- 표지 -->
<div class="paper_cover">
	<div class="content_cover"> 
		<img class="cover" src="https://cmtinfo-report.s3.ap-northeast-2.amazonaws.com/data/assets/images/coverpage.png">
		<div class="cover_title">
		<div class="report"><b>{{.YearofReport}} MONTHLY REPORT</b></div><br>
		<div class="title">클라우드<br>
		<b>{{.MonthofReport}}월 보고서</b><br></div>
		<br>
		<div class="sub"><b>고객명:</b> {{.CustomerKoreaName}}<br>
		<b>계정명:</b> {{.CustomerName}}<br>
		<b>계정 ID:</b> {{.LinkedAccount}}<br></div>   
		</div>
	</div>    
</div>
<main>

{{ template "announce_doughnut" .}}
{{ template "announce_linegraph" .TotalServiceUsage}}
{{range .ServiceUsage}}{{ template "announce_linegraph" .}}{{end}}

		
<script type="text/javascript" src="https://cdn.jsdelivr.net/npm/chart.js"></script>
<script src="https://cdn.jsdelivr.net/npm/chartjs-adapter-date-fns/dist/chartjs-adapter-date-fns.bundle.min.js"></script>
<script>
{{ template "write_doughnut_graph" . }}
{{ template "write_line_service" .TotalServiceUsage }}
{{range .ServiceUsage}}
{{ template "write_line_service" . }}
{{end}}
</script>


<div class="paper">
	<div class="content">
		<div class="page_header">
			<div class="header_Rectangle"></div>
			<p id="header_year">{{.YearofReport}} MONTHLY REPORT</p>
			<div class="header_Rectangle"></div>
		</div>
		<div class="page_title">Unused EIP</div>
		<div class="page_description">
			<pre class="pre_description">
	현재 사용하지 않는 Elastic IP 리스트입니다.<br>
	사용하지 않으신다면 해당 EIP 해제를 권장드립니다.
			</pre>
		</div>
		<table class="tftable" border="1">
			<thead>
				<th>EIP</th>
				<th>Allocation ID</th>
				<th>Country Code</th>
			</thead>
			<tbody>
				{{range .UnusedEip}}
				<tr>{{ template "write_unusedeip" . }}</tr>
				{{end}}	
			</tbody>
		</table>
	</div>
</div>
<div class="paper">
	<div class="content">
		<div class="page_header">
			<div class="header_Rectangle"></div>
				<p id="header_year">{{.YearofReport}} MONTHLY REPORT</p>
				<div class="header_Rectangle"></div>
			</div>
		<div class="page_title">Unattached Volume</div>
		<div class="page_description_Graph">
			<pre class="pre_description">
	현재 사용하지 않는 볼륨 리스트입니다.<br>
	사용하지 않으신다면 해당 볼륨 종료를 권장드립니다.
			</pre>
		</div>
		<table class="tftable" border="1">
			<thead>
				<th>Volume ID</th>
				<th>Volume Status</th>
				<th>Country Code</th>
			</thead>
			<tbody>
				{{range .UnattachedVolume}}
				<tr>{{ template "write_unattachedvolume" . }}</tr>
				{{end}}
			</tbody>
		</table>
	</div>
</div>
</main>
</body>
</html>

{{define "write_rirecommendation"}}
<tr><td>{{.CurrentInstanceType}}</td>
<td>{{.LinkedAccountID}}</td>
<td>${{.MonthlyOndemandCost}}</td>
<td>${{.SavingAmount}}</td>
<td>{{.SavingPercentage}}%</td>
<td>{{.RecommendedNumberOfInstancesToPurchase}}</td></tr>

{{end}}

{{define "write_unattachedvolume"}}
<td>{{.UnattachedVolumeID}}</td><td>{{.UnattachedVolumeStatus}}</td><td>{{.Countrycode}}</td>
{{end}}


{{define "write_unusedeip"}}
<td>{{.Eip}}</td><td>{{.AllocationID}}</td><td>{{.Countrycode}}</td>
{{end}}

{{define "announce_doughnut"}}
<div class="paper">
<div class="content">
	<div class="page_header">
		<div class="header_Rectangle"></div>
		<p id="header_year">{{.YearofReport}} MONTHLY REPORT</p>
		<div class="header_Rectangle"></div>
	</div>
	<div class="page_title">AWS Service Cost</div>
	<div class="page_description">
		<pre class="pre_description">
	지난 달의 AWS 서비스별 비용 분석입니다.<br>
	(해당 비용은 Credit이 포함되지 않았습니다.)
		</pre>
	</div>
	<div>
		<div class="chartCard">
			<div class="chartBox">
				<div class="doughnut_size"><canvas id="doughnutchart"></canvas></div>
			<br>
			</div>
		</div>
	</div>
</div>
</div>
{{end}}

{{define "announce_linegraph"}}
<div class="paper">
<div class="content">
	<div class="page_header">
		<div class="header_Rectangle"></div>
		<p id="header_year">{{.ServiceYear}} MONTHLY REPORT</p>
		<div class="header_Rectangle"></div>
	</div>
	<div class="page_title">{{.ServiceName}}</div>
	<div class="page_description">
		<pre class="pre_description">
	지난 5개월 간의 {{.ServiceName}} 비용 분석입니다.<br>
	(해당 비용은 Credit이 포함되지 않았습니다.)
		</pre>
	</div>
	<div>
		<div class="chartCard">
			<div class="chartBox">
			<canvas id="{{.ServiceChartName}}"></canvas>
			<br>
			</div>
		</div>
	</div>
</div>
</div>
{{end}}

{{define "write_doughnut_graph"}}
    const doughnutchartconfig = {           
      	type: 'doughnut',
      	data: {
			
			labels: [{{range .ServiceUsage}}{{template "write_service_name" .}}{{end}}],
			datasets: [{
				label: "AWS 사용량",
				data: [{{range .ServiceUsage}}{{template "write_service_previous_cost" .}}{{end}}],
				backgroundColor: [
					'rgba(244, 67, 54, 100)',
					'rgba(3, 169, 244, 100)',
					'rgba(255, 255, 59, 100)',
					'rgba(233, 30, 99, 100)',
					'rgba(0, 188, 212, 100)',
					'rgba(255, 193, 7, 100)',
					'rgba(156, 39, 176, 100)',
					'rgba(0, 150, 136, 100)',
					'rgba(255, 152, 0, 100)',
					'rgba(103, 58, 183, 100)',
					'rgba(76, 175, 80, 100)',
					'rgba(77, 44, 64, 100)',
					'rgba(63, 81, 181, 100)',
					'rgba(139, 195, 74, 100)',
					'rgba(77, 44, 64, 100)',
					'rgba(33, 150, 243, 100)',
					'rgba(205, 220, 57, 100)',
					'rgba(158, 158, 158, 100)'
					],
					borderColor: [
					'rgba(244, 67, 54, 100)',
					'rgba(3, 169, 244, 100)',
					'rgba(255, 255, 59, 100)',
					'rgba(233, 30, 99, 100)',
					'rgba(0, 188, 212, 100)',
					'rgba(255, 193, 7, 100)',
					'rgba(156, 39, 176, 100)',
					'rgba(0, 150, 136, 100)',
					'rgba(255, 152, 0, 100)',
					'rgba(103, 58, 183, 100)',
					'rgba(76, 175, 80, 100)',
					'rgba(77, 44, 64, 100)',
					'rgba(63, 81, 181, 100)',
					'rgba(139, 195, 74, 100)',
					'rgba(77, 44, 64, 100)',
					'rgba(33, 150, 243, 100)',
					'rgba(205, 220, 57, 100)',
					'rgba(158, 158, 158, 100)'
					],
					borderWidth: 3
				}],
		},
    };
    
    const doughnutchart = new Chart(
      document.getElementById('doughnutchart'),
      doughnutchartconfig
    );

{{end}}


{{define "write_line_service"}}
  
	// config 
    const {{.ServiceConfigName}} = {               // 추가시 'config' -> 'config1',2,3..등으로 수정
      type: 'line',
      data: {                                
		labels: [{{range .Service}}{{ template "write_service_date" . }}{{end}}],
		datasets: [{
			label: '{{.ServiceName}}',
			data: [{{range .Service}}{{ template "write_service_cost" . }}{{end}}],
			backgroundColor: [
			'rgba(255, 26, 104)'
			],
			borderColor: [
			'rgba(255, 26, 104)',
			],
			borderWidth: 3
			}]
		},
      	options: {
       		scales: {
          		y: {
            	beginAtZero: true
          	}
        }
      }
    };
    // render init block
    const {{.ServiceChartName}} = new Chart(
      document.getElementById('{{.ServiceChartName}}'),
      {{.ServiceConfigName}}
    );

{{end}}

{{define "write_service_name"}}'{{.ServiceName}}',{{end}}
{{define "write_service_previous_cost"}}{{.ServiceCost}},{{end}}
{{define "write_service_date"}}'{{.Month}}월',{{end}}

{{define "write_service_cost"}}{{.ServiceUsage}},{{end}}
`

const invoicetemplatesample = `
<!DOCTYPE html>
<html lang="en">
<style>
.clearfix:after {
	content: "";
	display: table;
	clear: both;
  }
  
  a {
	color: #20409A;
	text-decoration: underline;
  }
  
  body {
	position: relative;
	width: 21cm;  
	height: 29.7cm; 
	margin: 0 auto; 
	color: #001028;
	background: #FFFFFF; 
	font-family: Arial, sans-serif; 
	font-size: 12px; 
	font-family: Arial;
  }
  
  header {
	padding: 10px 0;
	margin-bottom: 0px;
  }
  
  #logo {
	text-align: center;
	padding: 10px;
  }
  
  #logo img {
	width: 100px;
  }
  
  h1 {
	border-top: 2px solid  #20409A;
	border-bottom: 2px solid  #20409A;
	color: #20409A;
	font-size: 2.4em;
	line-height: 1.4em;
	font-weight: normal;
	text-align: center;
	margin: 0 0 10px 0;
	background: url(dimension.png);
  }
  
  #project {
	float: left;
  }
  
  #project span {
	color: #001028;
	text-align: left;
	width: 52px;
	margin-right: 18px;
	display: inline-block;
	font-size: 11px;
	font-weight: bold;
  }
  
  #company {
	float: right;
	text-align: right;
  }
  
  #project div,
  #company div {
	white-space: nowrap;        
  }
  
  #center {
	text-align: center;
	background-color: #F5F5F5;
  }
  
  table {
	width: 100%;
	border-collapse: collapse;
	border-spacing: 0;
	border: 1px solid #001028;
  }
  
  
  table th,
  table td {
	text-align: center;
	border: 1px solid #001028;
  }
  
  table th {
	padding: 10px 20px;
	color: #001028;
	border-bottom: 1px solid #001028;
	white-space: nowrap;        
	font-weight: bold;
	text-align: center;
	background-color: #F5F5F5;
  }
  
  table .service,
  table .desc {
	text-align: left;
  }
  
  table td {
	padding: 10px;
	text-align: right;
  }
  
  table td.service,
  table td.desc {
	vertical-align: top;
  }
  
  table td.unit,
  table td.qty,
  table td.total {
	font-size: 12px;
  }
  
  table td.grand {
	border-top: 1px solid #5D6975;;
  }
  
  footer {
	color: #5D6975;
	width: 100%;
	height: 30px;
	position: absolute;
	bottom: 0;
	border-top: 1px solid #C1CED9;
	padding: 10px 0;
	text-align: center;
  }
  .content {
		padding: 0;
		height: 257mm;
	}
  @media print {
		html, body {
			width: 210mm;
			height: 297mm;
			background: rgb(255, 255, 255);
		}
		.paper {
			margin: 0;
			border: initial;
			border-radius: initial;
			width: initial;
			min-height: initial;
			box-shadow: initial;
			background: initial;
			page-break-after: always;
		}
	}
</style>
<head>
<meta charset="UTF-8">  
<title>Invoice</title>
</head>
<body>
	<div class="paper">
		<div class="content">
<header class="clearfix">
      <h1>INVOICE</h1>        
      <div id="company" class="clearfix">
        <div>씨엠티정보통신(주)</div>           
        <div>서울시 성동구 성수이로22길 37<br/>아크밸리, 803호</div>         
        <div><a href="mailto:cmtcloud-support@cmtinfo.co.kr">cmtcloud-support@cmtinfo.co.kr</a></div>
      </div>       

      <div id="project"> 
        <div><span>COMPANY</span>{{.CustomerKoreaName}}({{.LinkedAccountName}})</div>        
        <div><span>ADDRESS</span>{{.CustomoreAddress}}</div>   
        <div><span>DATE</span>{{.StartDate}} - {{.EndDate}}</div>   
        <div><span>DUE DATE</span>{{.DueDate}}</div>    
		<div style="color:red"><span>1 USD</span>{{.ExchangeRate.ExchangeRate}} KRW</div>	
      </div>
    </header>


<main>
<div></div>
<table>
<thead>
  <tr>
    <th class="service" style="text-align: center;">SERVICE</th>
    <th>CHARGE</th>
    <th>CREDIT</th>
    <th>TOTAL</th>
  </tr>
</thead>

<tbody>
{{range .Services}}
<div>{{ template "write_service" . }}</div>
{{end}}

<tr>
<td class="service">Managed Service</td> 
<td class="unit">${{.ManagedService}}</td>
<td class="credit">$0.0</td>
<td class="total">${{.ManagedService}}</td>
</tr>
<tr>
<td colspan="2" rowspan='4'></td>
<td id="center"><b>SUB TOTAL</b></td>
<td class="total">${{.TotalCostBeforeTax}}</td>
</tr>

<tr>
<td id="center"><b>TAX (10%)</b></td> 
<td class="tax">${{.TaxAmount}}</td>
</tr>

<tr>

<td id="center"><b>T O T A L</b></td>
<td class="TOTAL"><b>${{.TotalCostAfterTax}}</b></td>
</tr>

<tr>
<td id="center"><b>KRW</b></td>
<td class="krw" style="color: red;"><b>&#x20a9; {{.TotalCostAfterTaxKRW}}</b></td>
</tr>
</tbody>
</table>
<div id="logo">
  <img src="http://cmtinfo.co.kr/kor/image/logo.svg" style="width: 250px; height: 35px;">      
</div>
</main>
</div>
	</div>
</body>
</html>

{{define "write_service"}}
<tr>
<td class="service">{{.DetailService}}</td>    
<td class="unit">{{.ServiceTotalCostBeforeCredit}}</td>   
<td class="credit">{{.CreaditUsage}}</td>   
<td class="total">{{.ServiceTotalCostAfterCredit}}</td>  
</tr>
{{end}}

{{define "write_region"}}
- {{.Region}} ${{.RegionTotalCost}}</br>
{{end}}

{{define "write_description"}}
- {{.Description}} $ {{.DescriptionTotalCost}}</br>
{{end}}

{{define "email"}}
<a href="mailto:{{.Email}}">{{.Email}}</a>
{{end}}
`

type ReadContact struct {
	LinkedAccountID       string
	CustomerInChargeName  string
	CustomerEmail         string
	CustomerInChargePhone string
}

type ReadCustomer struct {
	CustomerKoreaName string
	RegisteredName    string
	LinkedAccountID   string
	RegisteredDate    string
	ManagedService    string
	CustomerAddress   string
	SPPDiscount       string
	DueDate           string
}

type CustomerData struct {
	CustomerLinkedAccountId string
	CustomerName            string
	CustomerEmail           string
	RegisteredDate          string
	CustomerKoreaName       string
	CustomerAddress         string
	CustomerInChargeName    string
	CustomerInChargePhone   string
	ManagedService          string
	SPPDiscount             string
	DueDate                 string
}

type CountryData struct {
	CountrySite  string
	CountryName  string
	CountryCode  string
	CountryCode2 string
}

type CountryDataCF struct {
	CountryName string
	CountryCode string
}

type CurrencyRate struct {
	Year         string
	Month        string
	ExchangeRate string
}

type InvoiceData struct {
	InvoiceID              string
	PayerAccountId         string
	LinkedAccountId        string
	RecordType             string
	RecordID               string
	BillingPeriodStartDate string
	BillingPeriodEndDate   string
	InvoiceDate            string
	PayerAccountName       string
	LinkedAccountName      string
	TaxationAddress        string
	PayerPONumber          string
	ProductCode            string
	ProductName            string
	SellerOfRecord         string
	UsageType              string
	Operation              string
	RateId                 string
	ItemDescription        string
	UsageStartDate         string
	UsageEndDate           string
	UsageQuantity          string
	BlendedRate            string
	CurrencyCode           string
	CostBeforeTax          string
	Credits                string
	TaxAmount              string
	TaxType                string
	TotalCost              string
}

type RefactoredInvoiceData struct {
	LinkedAccountId        string
	RecordID               string
	BillingPeriodStartDate string
	BillingPeriodEndDate   string
	InvoiceDate            string
	LinkedAccountName      string
	ProductCode            string
	ProductName            string
	SellerOfRecord         string
	UsageType              string
	Operation              string
	RateId                 string
	ItemDescription        string
	Region                 string
	DetailService          string
	UsageStartDate         string
	UsageEndDate           string
	UsageQuantity          float64
	CurrencyCode           string
	CostBeforeTax          float64
	Credits                float64
	TaxAmount              float64
	TaxType                string
	TotalCost              float64
}

type MonthlyReport struct {
	LinkedAccount         string
	CustomerName          string
	CustomerEmail         string
	RegisteredDate        string
	CustomerKoreaName     string
	CustomerAddress       string
	CustomerInChargeName  string
	CustomerInChargePhone string
	YearofReport          string
	MonthofReport         string
	TotalServiceUsage     ServiceUsage
	ServiceUsage          []ServiceUsage
	UnusedEip             []UnusedEip
	UnattachedVolume      []UnattachedVolume
	//OldGenerationInstanceType []OldGenerationInstanceType
	RIRecomendation []TypeRecommendation
}

type OldGenerationInstanceType struct {
	InstanceID    string
	InstanceType  string
	RecommendType string
}

type TypeRecommendation struct {
	LinkedAccountID                        string
	MonthlyOndemandCost                    string
	SavingAmount                           string
	SavingPercentage                       string
	CurrentInstanceType                    string
	RecommendedNumberOfInstancesToPurchase string
}

type UnattachedVolume struct {
	UnattachedVolumeID     string
	UnattachedVolumeStatus string
	Countrycode            string
}

type UnusedEip struct {
	Eip          string
	AllocationID string
	Countrycode  string
}

type ServiceUsage struct {
	ServiceName       string
	ServiceDataName   string
	ServiceConfigName string
	ServiceChartName  string
	ServiceCost       float64
	ServiceYear       string
	Service           []MonthlyService
}

type MonthlyService struct {
	Month        int
	ServiceUsage float64
}

type LinkedCredential struct {
	LinkedAccount string
	AccessKey     string
	SecretKey     string
}

type MonthlyResultData struct {
	Month      int
	ResultData []ResultData
}

type ResultData struct {
	LinkedAccountID      string
	CustomerKoreaName    string
	LinkedAccountName    string
	Customerid           string
	CustomoreAddress     string
	CustomerEmailAddress []CustomerEmailAddress
	CustomerIncharge     string
	StartDate            string
	EndDate              string
	DueDate              string
	ReportMonth          string
	ManagedService       string
	TotalCostBeforeTax   string
	TaxAmount            string
	TotalCostAfterTax    string
	TotalCostAfterTaxKRW string
	Services             []Service
	ExchangeRate         CurrencyRate
}

type Service struct {
	DetailService                   string
	ServiceTotalCostBeforeCredit    string
	ServiceTotalCostBeforeCreditrow float64
	CreaditUsage                    string
	ServiceTotalCostAfterCredit     string
	Region                          []Region
}

type Region struct {
	Region          string
	RegionTotalCost string
	Description     []Description
}

type Description struct {
	Description          string
	DescriptionTotalCost string
}

type CustomerEmailAddress struct {
	Email string
}

type zipType struct {
	writer *zip.Writer
}
