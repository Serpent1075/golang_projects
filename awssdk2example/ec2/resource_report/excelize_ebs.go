package main

import (
	"log"
	"strconv"

	"github.com/xuri/excelize/v2"
)

func MakeEBSValue(f *excelize.File, title string, rowindex int, accountName *string, result []*EBSData, resource string) int {
	for _, data := range result {

		style := ValueStyle(f)

		f.SetCellStyle(title, "B"+strconv.Itoa(rowindex), "I"+strconv.Itoa(rowindex), style)
		f.SetCellValue(title, "B"+strconv.Itoa(rowindex), *accountName)
		f.SetCellValue(title, "C"+strconv.Itoa(rowindex), title)
		f.SetCellValue(title, "D"+strconv.Itoa(rowindex), data.ID)
		f.SetCellValue(title, "E"+strconv.Itoa(rowindex), data.Name)
		f.SetCellValue(title, "F"+strconv.Itoa(rowindex), data.Type)
		f.SetCellValue(title, "G"+strconv.Itoa(rowindex), data.IOPS)
		f.SetCellValue(title, "H"+strconv.Itoa(rowindex), data.Size)
		f.SetCellValue(title, "I"+strconv.Itoa(rowindex), data.State)

		rowindex++
	}

	return rowindex
}

func MakeEBSColumn(f *excelize.File, title string, rowindex int, resource string) int {

	style := ColumnStyle(f)

	f.SetCellStyle(title, "B"+strconv.Itoa(rowindex), "I"+strconv.Itoa(rowindex), style)
	f.SetCellValue(title, "B"+strconv.Itoa(rowindex), "Account")
	f.SetCellValue(title, "C"+strconv.Itoa(rowindex), "Service")
	f.SetCellValue(title, "D"+strconv.Itoa(rowindex), "ID")
	f.SetCellValue(title, "E"+strconv.Itoa(rowindex), "Name")
	f.SetCellValue(title, "F"+strconv.Itoa(rowindex), "Type")
	f.SetCellValue(title, "G"+strconv.Itoa(rowindex), "IOPS")
	f.SetCellValue(title, "H"+strconv.Itoa(rowindex), "Size")
	f.SetCellValue(title, "I"+strconv.Itoa(rowindex), "State")

	return rowindex + 1
}

func MakeEBSSubtitle(f *excelize.File, title string, subtitle string, rowindex int) int {
	err := f.MergeCell(title, "B"+strconv.Itoa(rowindex), "D"+strconv.Itoa(rowindex))
	if err != nil {
		log.Printf("Make SubTitle Error: %s", err.Error())
	}
	style, index := SubtitleStyle(f, rowindex)
	f.SetCellStyle(title, "B"+strconv.Itoa(index), "D"+strconv.Itoa(index), style)
	f.SetCellValue(title, "B"+strconv.Itoa(index), subtitle)
	return index + 2
}

func MakeEBSTitle(f *excelize.File, title string, rowindex int) int {
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
