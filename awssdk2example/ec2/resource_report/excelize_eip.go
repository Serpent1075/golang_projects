package main

import (
	"log"
	"strconv"

	"github.com/xuri/excelize/v2"
)

func MakeEIPValue(f *excelize.File, title string, rowindex int, accountName *string, result []*EIPData, resource string) int {
	for _, data := range result {

		style := ValueStyle(f)

		f.SetCellStyle(title, "B"+strconv.Itoa(rowindex), "E"+strconv.Itoa(rowindex), style)
		f.SetCellValue(title, "B"+strconv.Itoa(rowindex), *accountName)
		f.SetCellValue(title, "C"+strconv.Itoa(rowindex), title)
		f.SetCellValue(title, "D"+strconv.Itoa(rowindex), data.IPAddress)
		f.SetCellValue(title, "E"+strconv.Itoa(rowindex), data.AssociationId)
		rowindex++
	}

	return rowindex
}

func MakeEIPColumn(f *excelize.File, title string, rowindex int, resource string) int {

	style := ColumnStyle(f)

	f.SetCellStyle(title, "B"+strconv.Itoa(rowindex), "E"+strconv.Itoa(rowindex), style)
	f.SetCellValue(title, "B"+strconv.Itoa(rowindex), "Account")
	f.SetCellValue(title, "C"+strconv.Itoa(rowindex), "Service")
	f.SetCellValue(title, "D"+strconv.Itoa(rowindex), "IP Address")
	f.SetCellValue(title, "E"+strconv.Itoa(rowindex), "Association ID")

	return rowindex + 1
}

func MakeEIPSubtitle(f *excelize.File, title string, subtitle string, rowindex int) int {
	err := f.MergeCell(title, "B"+strconv.Itoa(rowindex), "D"+strconv.Itoa(rowindex))
	if err != nil {
		log.Printf("Make SubTitle Error: %s", err.Error())
	}
	style, index := SubtitleStyle(f, rowindex)
	f.SetCellStyle(title, "B"+strconv.Itoa(index), "D"+strconv.Itoa(index), style)
	f.SetCellValue(title, "B"+strconv.Itoa(index), subtitle)
	return index + 2
}

func MakeEIPTitle(f *excelize.File, title string, rowindex int) int {
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
