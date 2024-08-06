package main

import (
	"log"
	"strconv"

	"github.com/xuri/excelize/v2"
)

func MakeLBValue(f *excelize.File, title string, rowindex int, accountName *string, result []*LoadBalancerData, resource string) int {
	for _, lbdata := range result {

		style := ValueStyle(f)

		f.SetCellStyle(title, "B"+strconv.Itoa(rowindex), "G"+strconv.Itoa(rowindex), style)
		f.SetCellValue(title, "B"+strconv.Itoa(rowindex), *accountName)
		f.SetCellValue(title, "C"+strconv.Itoa(rowindex), title)
		f.SetCellValue(title, "D"+strconv.Itoa(rowindex), lbdata.Name)
		f.SetCellValue(title, "E"+strconv.Itoa(rowindex), lbdata.InExternal)
		f.SetCellValue(title, "F"+strconv.Itoa(rowindex), lbdata.PrdDev)
		f.SetCellValue(title, "G"+strconv.Itoa(rowindex), lbdata.Type)
		rowindex++
	}

	return rowindex
}

func MakeLBColumn(f *excelize.File, title string, rowindex int, resource string) int {

	style := ColumnStyle(f)

	f.SetCellStyle(title, "B"+strconv.Itoa(rowindex), "G"+strconv.Itoa(rowindex), style)
	f.SetCellValue(title, "B"+strconv.Itoa(rowindex), "Account")
	f.SetCellValue(title, "C"+strconv.Itoa(rowindex), "Service")
	f.SetCellValue(title, "D"+strconv.Itoa(rowindex), "LB Name")
	f.SetCellValue(title, "E"+strconv.Itoa(rowindex), "LB Scheme")
	f.SetCellValue(title, "F"+strconv.Itoa(rowindex), "LB Usage")
	f.SetCellValue(title, "G"+strconv.Itoa(rowindex), "LB Type")

	return rowindex + 1
}

func MakeLBSubtitle(f *excelize.File, title string, subtitle string, rowindex int) int {
	err := f.MergeCell(title, "B"+strconv.Itoa(rowindex), "D"+strconv.Itoa(rowindex))
	if err != nil {
		log.Printf("Make SubTitle Error: %s", err.Error())
	}
	style, index := SubtitleStyle(f, rowindex)
	f.SetCellStyle(title, "B"+strconv.Itoa(index), "D"+strconv.Itoa(index), style)
	f.SetCellValue(title, "B"+strconv.Itoa(index), subtitle)
	return index + 2
}
