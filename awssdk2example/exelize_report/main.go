// s3://wdms-backup-logging/Backup/459666865111/ap-northeast-2/2023/03/31/backup_jobs_report_01_12_2022/
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/efs"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/xuri/excelize/v2"
)

var bucketname string
var accountnum string
var region string

func init() {
	accountnum = os.Getenv("accountnum")
	bucketname = os.Getenv("bucketname")
	region = os.Getenv("region")
}

// 459666865111
func main() {
	lambda.Start(handler)
}

func handler(ctx context.Context) {
	cfg := GetConfig()
	if cfg == nil {
		os.Exit(1)
	}
	previousdate := time.Now().AddDate(0, -1, 0)

	uploadkey := "Backup/" + accountnum + "/" + region + "/" + getDateString(previousdate.Year()) + "/" + getDateString(int(previousdate.Month())) + "/"

	ec2result := make([]*Result, 0)
	rdsresult := make([]*Result, 0)
	efsresult := make([]*Result, 0)

	listitem := ParseJsonToStruct(ctx, cfg, previousdate)

	ec2ids := makeSliceUnique(GetInstanceIds(listitem))
	rdsids := makeSliceUnique(GetRDSId(listitem))
	efsids := makeSliceUnique(GetEFSId(listitem))

	ec2tags := GetEC2Information(ctx, cfg, ec2ids)
	efstags := GetEfSInformation(ctx, cfg, efsids)

	ec2result = ParseStructToResultWithTags(listitem, ec2ids, ec2result, ec2tags)
	rdsresult = ParseStructToResult(listitem, rdsids, rdsresult)
	efsresult = ParseStructToResultWithTags(listitem, efsids, efsresult, efstags)

	filename := MakeExcelReport("BMW_WDMS", ec2result, rdsresult, efsresult)
	PutReport(ctx, cfg, uploadkey, filename)
	/*
		for _, data := range listitem {
			if data.ResourceArn == "arn:aws:ec2:ap-northeast-2:459666865111:instance/i-08ef3fcbbae0cc154" {
				log.Println(*data)
			}
		}
	*/
}

func PutReport(ctx context.Context, cfg *aws.Config, uploadkey string, filename string) {
	client := s3.NewFromConfig(*cfg)

	file, err := os.Open("/tmp/" + filename)
	if err != nil {
		log.Printf("PutReport Error: %s", err.Error())
		return
	}
	defer file.Close()
	putinput := s3.PutObjectInput{
		Bucket: aws.String(bucketname),
		Key:    aws.String(uploadkey + filename),
		Body:   file,
	}

	putoutput, err := client.PutObject(context.Background(), &putinput)
	if err != nil {
		log.Printf("PutReport Error: %s", err.Error())
		return
	}
	log.Printf("Version ID: %s", *putoutput.VersionId)
}

func MakeExcelReport(title string, ec2result []*Result, rdsresult []*Result, efsresult []*Result) string {
	f := excelize.NewFile()
	defer func() {
		if err := f.Close(); err != nil {
			log.Printf("MakeExcelReport Error: %s", err.Error())
		}
	}()
	// 워크시트 만들기
	index, err := f.NewSheet(title)
	if err != nil {
		log.Printf("MakeExcelReport Error: %s", err.Error())
		return ""
	}
	// 셀 값 설정
	rowindex := 1
	rowindex = MakeTitle(f, title, rowindex)
	rowindex = MakeSubtitle(f, title, "1) EC2 AMI Backup", rowindex)
	rowindex = MakeColumn(f, title, rowindex, "ec2")
	rowindex = MakeValue(f, title, rowindex, ec2result, "ec2")
	rowindex = MakeSubtitle(f, title, "2) RDS Backup", rowindex)
	rowindex = MakeColumn(f, title, rowindex, "rds")
	rowindex = MakeValue(f, title, rowindex, rdsresult, "rds")
	rowindex = MakeSubtitle(f, title, "3) EFS Backup", rowindex)
	rowindex = MakeColumn(f, title, rowindex, "efs")
	MakeValue(f, title, rowindex, efsresult, "efs")

	// 통합 문서에 대 한 기본 워크시트를 설정 합니다
	f.SetActiveSheet(index)
	// 지정 된 경로를 기반으로 파일 저장

	filename := title + "_monthly_backup_" + getDateString(time.Now().Year()) + "_" + getDateString(int(time.Now().Month())-1) + ".xlsx"
	if err := f.SaveAs("/tmp/" + filename); err != nil {
		log.Printf("MakeExcelReport Error: %s", err.Error())
	}
	return filename
}

func MakeValue(f *excelize.File, title string, rowindex int, result []*Result, resource string) int {
	for _, instance := range result {
		for i, data := range instance.Attributes {
			bordersetting := make([]excelize.Border, 0)
			bordersetting = append(bordersetting, excelize.Border{
				Type:  "left",
				Style: 2,
				Color: "#000000",
			})
			bordersetting = append(bordersetting, excelize.Border{
				Type:  "right",
				Style: 2,
				Color: "#000000",
			})
			bordersetting = append(bordersetting, excelize.Border{
				Type:  "top",
				Style: 2,
				Color: "#000000",
			})

			if i == len(instance.Attributes)-1 {
				bordersetting = append(bordersetting, excelize.Border{
					Type:  "bottom",
					Style: 5,
					Color: "#000000",
				})
			} else {
				bordersetting = append(bordersetting, excelize.Border{
					Type:  "bottom",
					Style: 2,
					Color: "#000000",
				})
			}

			style, err := f.NewStyle(&excelize.Style{
				Border: bordersetting,
				Alignment: &excelize.Alignment{
					Horizontal: "center",
					Vertical:   "center",
				},
				Font: &excelize.Font{
					Bold: false,
					Size: 10,
				},
			})
			if err != nil {
				log.Printf("MakeValue Error: %s", err.Error())
				return 0
			}

			f.SetCellStyle(title, "B"+strconv.Itoa(rowindex), "I"+strconv.Itoa(rowindex), style)
			f.SetCellValue(title, "B"+strconv.Itoa(rowindex), title)
			f.SetCellValue(title, "C"+strconv.Itoa(rowindex), instance.ServerName)
			f.SetCellValue(title, "D"+strconv.Itoa(rowindex), instance.InstanceID)
			f.SetCellValue(title, "E"+strconv.Itoa(rowindex), data.StartTime)
			f.SetCellValue(title, "F"+strconv.Itoa(rowindex), data.EndTime)
			f.SetCellValue(title, "G"+strconv.Itoa(rowindex), data.Duration)
			if resource == "rds" {
				f.SetCellValue(title, "H"+strconv.Itoa(rowindex), data.Status)
			} else {
				f.SetCellValue(title, "H"+strconv.Itoa(rowindex), data.Volume)
				f.SetCellValue(title, "I"+strconv.Itoa(rowindex), data.Status)
			}

			rowindex++
		}
	}
	return rowindex + 2
}

func MakeColumn(f *excelize.File, title string, rowindex int, resource string) int {
	bordersetting := make([]excelize.Border, 0)
	bordersetting = append(bordersetting, excelize.Border{
		Type:  "left",
		Style: 2,
		Color: "#000000",
	})
	bordersetting = append(bordersetting, excelize.Border{
		Type:  "right",
		Style: 2,
		Color: "#000000",
	})
	bordersetting = append(bordersetting, excelize.Border{
		Type:  "top",
		Style: 2,
		Color: "#000000",
	})
	bordersetting = append(bordersetting, excelize.Border{
		Type:  "bottom",
		Style: 2,
		Color: "#000000",
	})
	style, err := f.NewStyle(&excelize.Style{
		Border: bordersetting,
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
		Font: &excelize.Font{
			Bold: true,
			Size: 10,
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Pattern: 1,
			Color:   []string{"#b8b8b8"},
		},
	})
	if err != nil {
		log.Printf("MakeColumn Error: %s", err.Error())
		return 0
	}
	f.SetCellStyle(title, "B"+strconv.Itoa(rowindex), "I"+strconv.Itoa(rowindex), style)
	f.SetCellValue(title, "B"+strconv.Itoa(rowindex), "서비스")
	f.SetCellValue(title, "C"+strconv.Itoa(rowindex), "서버")
	f.SetCellValue(title, "D"+strconv.Itoa(rowindex), "Instance ID")
	f.SetCellValue(title, "E"+strconv.Itoa(rowindex), "시작 (UTC)")
	f.SetCellValue(title, "F"+strconv.Itoa(rowindex), "종료 (UTC)")
	f.SetCellValue(title, "G"+strconv.Itoa(rowindex), "소요")
	if resource == "rds" {
		f.SetCellValue(title, "H"+strconv.Itoa(rowindex), "상태")
	} else {
		f.SetCellValue(title, "H"+strconv.Itoa(rowindex), "용량")
		f.SetCellValue(title, "I"+strconv.Itoa(rowindex), "상태")
	}

	return rowindex + 1
}

func MakeSubtitle(f *excelize.File, title string, subtitle string, rowindex int) int {
	err := f.MergeCell(title, "B"+strconv.Itoa(rowindex), "D"+strconv.Itoa(rowindex))
	if err != nil {
		log.Printf("Make SubTitle Error: %s", err.Error())
	}
	bordersetting := make([]excelize.Border, 0)
	bordersetting = append(bordersetting, excelize.Border{
		Type:  "left",
		Style: 2,
		Color: "#fcff96",
	})
	bordersetting = append(bordersetting, excelize.Border{
		Type:  "right",
		Style: 2,
		Color: "#fcff96",
	})
	bordersetting = append(bordersetting, excelize.Border{
		Type:  "top",
		Style: 2,
		Color: "#fcff96",
	})
	bordersetting = append(bordersetting, excelize.Border{
		Type:  "bottom",
		Style: 2,
		Color: "#fcff96",
	})
	style, err := f.NewStyle(&excelize.Style{
		Border: bordersetting,
		Alignment: &excelize.Alignment{
			Horizontal: "left",
			Vertical:   "center",
			Indent:     1,
		},
		Font: &excelize.Font{
			Bold: true,
			Size: 10,
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Pattern: 1,
			Color:   []string{"#fcff96"},
		},
	})
	if err != nil {
		log.Printf("MakeSubtitle Error: %s", err.Error())
		return rowindex + 1
	}
	f.SetCellStyle(title, "B"+strconv.Itoa(rowindex), "D"+strconv.Itoa(rowindex), style)
	f.SetCellValue(title, "B"+strconv.Itoa(rowindex), subtitle)
	return rowindex + 2
}

func MakeTitle(f *excelize.File, title string, rowindex int) int {
	err := f.MergeCell(title, "B1", "I1")
	if err != nil {
		log.Printf("Make Title Error: %s", err.Error())
	}
	bordersetting := make([]excelize.Border, 0)
	bordersetting = append(bordersetting, excelize.Border{
		Type:  "left",
		Style: 2,
		Color: "#000000",
	})
	bordersetting = append(bordersetting, excelize.Border{
		Type:  "right",
		Style: 2,
		Color: "#000000",
	})
	bordersetting = append(bordersetting, excelize.Border{
		Type:  "top",
		Style: 2,
		Color: "#000000",
	})
	bordersetting = append(bordersetting, excelize.Border{
		Type:  "bottom",
		Style: 2,
		Color: "#000000",
	})
	style, err := f.NewStyle(&excelize.Style{
		Border: bordersetting,
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
		Font: &excelize.Font{
			Bold: true,
			Size: 20,
		},
	})
	if err != nil {
		log.Printf("Make Title Error: %s", err.Error())
		return 0
	}
	f.SetCellStyle(title, "B"+strconv.Itoa(rowindex), "I"+strconv.Itoa(rowindex), style)
	f.SetCellValue(title, "B"+strconv.Itoa(rowindex), title+" 백업 일지")
	return rowindex + 2
}

func ParseStructToResultWithTags(listItem []*BackupReportItem, mapids []string, result []*Result, tags map[string]string) []*Result {

	for _, id := range mapids {
		var resultunit Result
		resultunit.InstanceID = id
		if tags[id] != "" {
			resultunit.ServerName = tags[id]
		} else {
			resultunit.ServerName = id
		}

		resattlist := make([]ResultAttributes, 0)
		for _, item := range listItem {
			if item.ResourceId == id {

				createtime := strings.ReplaceAll(strings.ReplaceAll(item.CreationDate, "T", " "), "Z", "")
				endtime := strings.ReplaceAll(strings.ReplaceAll(item.CompletionDate, "T", " "), "Z", "")
				jobruntime := strings.ReplaceAll(strings.ReplaceAll(item.JobRunTime, "T", " "), "Z", "")
				volume := GetHumanReadableVolume(item.BackupSizeInBytes)
				status := item.JobStatus
				resatt := ResultAttributes{
					StartTime: createtime,
					EndTime:   endtime,
					Duration:  jobruntime,
					Volume:    volume,
					Status:    status,
				}

				resattlist = append(resattlist, resatt)
			}
		}
		resultunit.Attributes = resattlist
		result = append(result, &resultunit)
	}

	return result
}

func ParseStructToResult(listItem []*BackupReportItem, mapids []string, result []*Result) []*Result {

	for _, id := range mapids {
		var resultunit Result
		resultunit.InstanceID = id
		resultunit.ServerName = id
		resattlist := make([]ResultAttributes, 0)
		for _, item := range listItem {
			if item.ResourceId == id {

				createtime := strings.ReplaceAll(strings.ReplaceAll(item.CreationDate, "T", " "), "Z", "")
				endtime := strings.ReplaceAll(strings.ReplaceAll(item.CompletionDate, "T", " "), "Z", "")
				jobruntime := strings.ReplaceAll(strings.ReplaceAll(item.JobRunTime, "T", " "), "Z", "")
				volume := GetHumanReadableVolume(item.BackupSizeInBytes)
				status := item.JobStatus
				resatt := ResultAttributes{
					StartTime: createtime,
					EndTime:   endtime,
					Duration:  jobruntime,
					Volume:    volume,
					Status:    status,
				}

				resattlist = append(resattlist, resatt)
			}
		}
		resultunit.Attributes = resattlist
		result = append(result, &resultunit)
	}

	return result
}

func GetHumanReadableVolume(bytesize int) string {
	unit := []string{"Byte", "KB", "MB", "GB", "TB", "PB"}
	count := 0
	for {
		if count > 10 {
			return "uncountable"
		}
		if bytesize >= 1024 {
			bytesize = bytesize / 1024
			count++
		} else {
			return fmt.Sprintf("%.2f %s", math.Round(float64(bytesize)*10)/10, unit[count])
		}
	}
}

func GetEFSId(listItem []*BackupReportItem) []string {
	ids := make([]string, 0)
	for i, item := range listItem {
		if item.ResourceType == "EFS" {
			result := goReSplit(item.ResourceArn, `:`)
			id := strings.Replace(result[len(result)-1], "file-system/", "", -1)
			ids = append(ids, id)
			listItem[i].ResourceId = id
		}
	}

	return ids
}

func GetRDSId(listItem []*BackupReportItem) []string {
	ids := make([]string, 0)
	for i, item := range listItem {
		if item.ResourceType == "RDS" {
			result := goReSplit(item.ResourceArn, `:`)
			id := result[len(result)-1]
			ids = append(ids, id)
			listItem[i].ResourceId = id
		}
	}

	return ids
}

func GetInstanceIds(listItem []*BackupReportItem) []string {
	ids := make([]string, 0)
	for i, item := range listItem {
		if item.ResourceType == "EC2" {
			result := goReSplit(item.ResourceArn, `:`)
			id := strings.Replace(result[len(result)-1], "instance/", "", -1)
			ids = append(ids, id)
			listItem[i].ResourceId = id
		}
	}

	return ids
}

func ParseJsonToStruct(ctx context.Context, cfg *aws.Config, previousdate time.Time) []*BackupReportItem {

	//var listitem2 ListItem

	backupReportItems := make([]*BackupReportItem, 0)
	client := s3.NewFromConfig(*cfg)

	reportdate := time.Date(previousdate.Year(), previousdate.Month(), 1, 0, 0, 0, 0, time.Local)
	count := 0
	for {

		var listitem ListItem
		listitem.ReportItems = make([]*BackupReportItem, 0)
		if reportdate.Month() == time.Now().Month() || count > 33 {
			break
		}
		var fullkey *string
		keyyear := getDateString(reportdate.Year())
		keymonth := getDateString(int(reportdate.Month()))
		keyday := getDateString(reportdate.Day())
		key := "Backup/" + accountnum + "/" + region + "/" + keyyear + "/" + keymonth + "/" + keyday + "/backup_jobs_report_01_12_2022/"
		listobjinput := s3.ListObjectsInput{
			Bucket: aws.String(bucketname),
			Prefix: aws.String(key),
		}
		listobjresp, listobjerr := client.ListObjects(ctx, &listobjinput)
		if listobjerr != nil {
			log.Printf("ParseJsonToStruct listobjerr: %s", listobjerr.Error())
		}
		for _, data := range listobjresp.Contents {
			if strings.Contains(*data.Key, ".json") {
				fullkey = data.Key
			}
		}
		getobjinput := s3.GetObjectInput{
			Bucket:              aws.String(bucketname),
			Key:                 fullkey,
			ExpectedBucketOwner: aws.String(accountnum),
		}

		getobjresp, err := client.GetObject(ctx, &getobjinput)
		if err != nil {
			log.Printf("ParseJsonToStruct getobjresp: %s", err.Error())
		}
		defer getobjresp.Body.Close()
		d, err := ioutil.ReadAll(getobjresp.Body)
		if err != nil {
			log.Printf("ParseJsonToStruct fileread: %s", err.Error())

		}
		json.Unmarshal(d, &listitem)
		backupReportItems = append(backupReportItems, listitem.ReportItems...)
		reportdate = reportdate.AddDate(0, 0, 1)
		count++
	}
	/*
		for _, data := range backupReportItems {
			log.Printf("%s   %s   %s", data.ResourceId, data.CreationDate, data.CompletionDate)
		}
	*/
	return backupReportItems
}

func getDateString(date int) string {
	if date < 10 {
		return "0" + strconv.Itoa(date)
	} else {
		return strconv.Itoa(date)
	}
}

func GetEfSInformation(ctx context.Context, cfg *aws.Config, filesystemIds []string) map[string]string {
	client := efs.NewFromConfig(*cfg)
	mapids := make(map[string]string, 0)
	for _, efsid := range filesystemIds {
		input := efs.DescribeFileSystemsInput{
			FileSystemId: aws.String(efsid),
		}
		resp, err := client.DescribeFileSystems(ctx, &input)
		if err != nil {
			log.Printf("GetEfSInformation: %s", err.Error())
		} else {
			mapids[efsid] = *resp.FileSystems[0].Name
		}
	}
	return mapids
}

func GetEC2Information(ctx context.Context, cfg *aws.Config, listInstanceIds []string) map[string]string {
	mapids := make(map[string]string, 0)
	client := ec2.NewFromConfig(*cfg)
	for _, id := range listInstanceIds {
		input := ec2.DescribeInstancesInput{
			InstanceIds: []string{id},
		}
		resp, err := client.DescribeInstances(ctx, &input)
		if err != nil {
			log.Printf("GetEC2Information: %s", err.Error())
		} else {
			for _, data := range resp.Reservations {
				if len(data.Instances) != 0 {
					for _, instance := range data.Instances {
						for _, tag := range instance.Tags {
							if *tag.Key == "Name" {
								mapids[*instance.InstanceId] = *tag.Value
							}
						}
					}
				}
			}
		}

	}

	return mapids
}

func GetConfig() *aws.Config {
	cfg, err := config.LoadDefaultConfig(
		context.TODO(),
	)
	if err != nil {
		log.Printf("PutReport Error: %s", err.Error())
		return nil
	}
	return &cfg
}

func goReSplit(text string, pattern string) []string {
	regex := regexp.MustCompile(pattern)
	result := regex.Split(text, -1)
	return result
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

type ListItem struct {
	ReportItems []*BackupReportItem `json:"reportItems"`
}

type BackupReportItem struct {
	ReportTimePeriodStart string `json:"reportTimePeriodStart"`
	ReportTimePeriodEnd   string `json:"reportTimePeriodEnd"`
	AccountId             string `json:"accountId"`
	Region                string `json:"region"`
	BackupJobId           string `json:"backupJobId"`
	JobStatus             string `json:"jobStatus"`
	ResourceType          string `json:"resourceType"`
	ResourceArn           string `json:"resourceArn"`
	BackupPlanArn         string `json:"backupPlanArn"`
	BackupRuleId          string `json:"backupRuleId"`
	CreationDate          string `json:"creationDate"`
	CompletionDate        string `json:"completionDate"`
	RecoveryPointArn      string `json:"recoveryPointArn"`
	JobRunTime            string `json:"jobRunTime"`
	BackupSizeInBytes     int    `json:"backupSizeInBytes"`
	BackupVaultName       string `json:"backupVaultName"`
	BackupVaultArn        string `json:"backupVaultArn"`
	IamRoleArn            string `json:"iamRoleArn"`
	ResourceId            string
}

type Result struct {
	ServerName string
	InstanceID string
	Attributes []ResultAttributes
}

type ResultAttributes struct {
	StartTime string
	EndTime   string
	Duration  string
	Volume    string
	Status    string
}
