package main

import (
	"log"
	"strconv"

	"github.com/xuri/excelize/v2"
)

func MakeSGValue(f *excelize.File, title string, rowindex int, accountName *string, result []*SecurityGroupRuleData, resource string) int {
	for _, sgdata := range result {

		style := ValueStyle(f)

		f.SetCellStyle(title, "B"+strconv.Itoa(rowindex), "J"+strconv.Itoa(rowindex), style)
		f.SetCellValue(title, "B"+strconv.Itoa(rowindex), *accountName)
		f.SetCellValue(title, "C"+strconv.Itoa(rowindex), title)
		f.SetCellValue(title, "D"+strconv.Itoa(rowindex), sgdata.GroupID)
		f.SetCellValue(title, "E"+strconv.Itoa(rowindex), sgdata.RuleID)
		f.SetCellValue(title, "F"+strconv.Itoa(rowindex), sgdata.InOutbound)
		f.SetCellValue(title, "G"+strconv.Itoa(rowindex), sgdata.SrcAddr)
		f.SetCellValue(title, "H"+strconv.Itoa(rowindex), sgdata.Protocol)
		f.SetCellValue(title, "I"+strconv.Itoa(rowindex), sgdata.Port)
		f.SetCellValue(title, "J"+strconv.Itoa(rowindex), sgdata.Description)
		rowindex++
	}

	return rowindex
}

func MakeSGColumn(f *excelize.File, title string, rowindex int, resource string) int {

	style := ColumnStyle(f)

	f.SetCellStyle(title, "B"+strconv.Itoa(rowindex), "J"+strconv.Itoa(rowindex), style)
	f.SetCellValue(title, "B"+strconv.Itoa(rowindex), "Account")
	f.SetCellValue(title, "C"+strconv.Itoa(rowindex), "Service")
	f.SetCellValue(title, "D"+strconv.Itoa(rowindex), "Group ID")
	f.SetCellValue(title, "E"+strconv.Itoa(rowindex), "Rule ID")
	f.SetCellValue(title, "F"+strconv.Itoa(rowindex), "In/Outbound")
	f.SetCellValue(title, "G"+strconv.Itoa(rowindex), "Source")
	f.SetCellValue(title, "H"+strconv.Itoa(rowindex), "Protocol")
	f.SetCellValue(title, "I"+strconv.Itoa(rowindex), "Port")
	f.SetCellValue(title, "J"+strconv.Itoa(rowindex), "Description")

	return rowindex + 1
}

func MakeSGSubtitle(f *excelize.File, title string, subtitle string, rowindex int) int {
	err := f.MergeCell(title, "B"+strconv.Itoa(rowindex), "D"+strconv.Itoa(rowindex))
	if err != nil {
		log.Printf("Make SubTitle Error: %s", err.Error())
	}
	style, index := SubtitleStyle(f, rowindex)
	f.SetCellStyle(title, "B"+strconv.Itoa(index), "D"+strconv.Itoa(index), style)
	f.SetCellValue(title, "B"+strconv.Itoa(index), subtitle)
	return index + 2
}
