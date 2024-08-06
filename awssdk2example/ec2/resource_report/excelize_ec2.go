package main

import (
	"log"
	"strconv"

	"github.com/xuri/excelize/v2"
)

func MakeEC2Value(f *excelize.File, title string, rowindex int, accountName *string, result []*EC2Data, resource string) int {
	for _, instance := range result {

		style := ValueStyle(f)

		f.SetCellStyle(title, "B"+strconv.Itoa(rowindex), "L"+strconv.Itoa(rowindex), style)
		f.SetCellValue(title, "B"+strconv.Itoa(rowindex), *accountName)
		f.SetCellValue(title, "C"+strconv.Itoa(rowindex), title)
		f.SetCellValue(title, "D"+strconv.Itoa(rowindex), instance.InstanceName)
		f.SetCellValue(title, "E"+strconv.Itoa(rowindex), instance.Hostname)
		f.SetCellValue(title, "F"+strconv.Itoa(rowindex), instance.InstanceId)
		f.SetCellValue(title, "G"+strconv.Itoa(rowindex), instance.InstanceType)
		f.SetCellValue(title, "H"+strconv.Itoa(rowindex), instance.OSVersion)
		f.SetCellValue(title, "I"+strconv.Itoa(rowindex), instance.PlatformDetails)
		f.SetCellValue(title, "J"+strconv.Itoa(rowindex), instance.PrivateIP)
		f.SetCellValue(title, "K"+strconv.Itoa(rowindex), instance.Status)
		f.SetCellValue(title, "L"+strconv.Itoa(rowindex), instance.Usage)

		rowindex++
	}

	return rowindex
}

func MakeEC2Column(f *excelize.File, title string, rowindex int, resource string) int {

	style := ColumnStyle(f)

	f.SetCellStyle(title, "B"+strconv.Itoa(rowindex), "L"+strconv.Itoa(rowindex), style)
	f.SetCellValue(title, "B"+strconv.Itoa(rowindex), "Account")
	f.SetCellValue(title, "C"+strconv.Itoa(rowindex), "Service")
	f.SetCellValue(title, "D"+strconv.Itoa(rowindex), "ServerName")
	f.SetCellValue(title, "E"+strconv.Itoa(rowindex), "Hostname")
	f.SetCellValue(title, "F"+strconv.Itoa(rowindex), "Instance ID")
	f.SetCellValue(title, "G"+strconv.Itoa(rowindex), "Instance Type")
	f.SetCellValue(title, "H"+strconv.Itoa(rowindex), "OS Version")
	f.SetCellValue(title, "I"+strconv.Itoa(rowindex), "Platform detail")
	f.SetCellValue(title, "J"+strconv.Itoa(rowindex), "Private IP")
	f.SetCellValue(title, "K"+strconv.Itoa(rowindex), "State")
	f.SetCellValue(title, "L"+strconv.Itoa(rowindex), "Usage")

	return rowindex + 1
}

func MakeEC2Subtitle(f *excelize.File, title string, subtitle string, rowindex int) int {
	err := f.MergeCell(title, "B"+strconv.Itoa(rowindex), "D"+strconv.Itoa(rowindex))
	if err != nil {
		log.Printf("Make SubTitle Error: %s", err.Error())
	}
	style, index := SubtitleStyle(f, rowindex)
	f.SetCellStyle(title, "B"+strconv.Itoa(index), "D"+strconv.Itoa(index), style)
	f.SetCellValue(title, "B"+strconv.Itoa(index), subtitle)
	return index + 2
}

func MakeEC2Title(f *excelize.File, title string, rowindex int) int {
	err := f.MergeCell(title, "B1", "D1")
	if err != nil {
		log.Printf("Make Title Error: %s", err.Error())
	}
	style := TitleStyle(f)
	f.SetCellStyle(title, "B"+strconv.Itoa(rowindex), "D"+strconv.Itoa(rowindex), style)
	f.SetCellValue(title, "B"+strconv.Itoa(rowindex), title)
	return rowindex + 2
}

//////////////////////////EBS/////////////////////////////////

/////////////////////////EIP///////////////////////////////////

///////////////////Security Group///////////////////////////////
