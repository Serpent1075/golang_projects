package main

import (
	"log"
	"strconv"

	"github.com/xuri/excelize/v2"
)

func MakeRDSValue(f *excelize.File, title string, rowindex int, accountName *string, result []*RDSData, resource string) int {
	for _, instance := range result {
		style := ValueStyle(f)

		f.SetCellStyle(title, "B"+strconv.Itoa(rowindex), "K"+strconv.Itoa(rowindex), style)
		f.SetCellValue(title, "B"+strconv.Itoa(rowindex), *accountName)
		f.SetCellValue(title, "C"+strconv.Itoa(rowindex), title)
		f.SetCellValue(title, "D"+strconv.Itoa(rowindex), instance.Identifier)
		f.SetCellValue(title, "E"+strconv.Itoa(rowindex), instance.InstanceClass)
		f.SetCellValue(title, "F"+strconv.Itoa(rowindex), instance.Engine+" "+instance.EngineVersion)
		f.SetCellValue(title, "G"+strconv.Itoa(rowindex), instance.AllocatedStorage)
		f.SetCellValue(title, "H"+strconv.Itoa(rowindex), instance.StorageType)
		f.SetCellValue(title, "I"+strconv.Itoa(rowindex), instance.MultiAZ)
		f.SetCellValue(title, "J"+strconv.Itoa(rowindex), instance.Usage)

		rowindex++
	}

	return rowindex
}

func MakeRDSColumn(f *excelize.File, title string, rowindex int, resource string) int {
	style := ColumnStyle(f)

	f.SetCellStyle(title, "B"+strconv.Itoa(rowindex), "K"+strconv.Itoa(rowindex), style)
	f.SetCellValue(title, "B"+strconv.Itoa(rowindex), "Account")
	f.SetCellValue(title, "C"+strconv.Itoa(rowindex), "Services")
	f.SetCellValue(title, "D"+strconv.Itoa(rowindex), "DB Identifier")
	f.SetCellValue(title, "E"+strconv.Itoa(rowindex), "Instance Class")
	f.SetCellValue(title, "F"+strconv.Itoa(rowindex), "Engine")
	f.SetCellValue(title, "G"+strconv.Itoa(rowindex), "Allocated Storage")
	f.SetCellValue(title, "H"+strconv.Itoa(rowindex), "Storage Type")
	f.SetCellValue(title, "I"+strconv.Itoa(rowindex), "MultiAZ")
	f.SetCellValue(title, "J"+strconv.Itoa(rowindex), "Usage")

	return rowindex + 1
}

func MakeRDSSubtitle(f *excelize.File, title string, subtitle string, rowindex int) int {
	rowindex = rowindex + 2
	err := f.MergeCell(title, "B"+strconv.Itoa(rowindex), "D"+strconv.Itoa(rowindex))
	if err != nil {
		log.Printf("Make SubTitle Error: %s", err.Error())
	}
	style, index := SubtitleStyle(f, rowindex)
	f.SetCellStyle(title, "B"+strconv.Itoa(index), "D"+strconv.Itoa(index), style)
	f.SetCellValue(title, "B"+strconv.Itoa(index), subtitle)
	return index + 2
}

// //////////////////////////// Cluster ///////////////////////////

func MakeClusterRDSColumn(f *excelize.File, title string, rowindex int, resource string) int {
	style := ColumnStyle(f)

	f.SetCellStyle(title, "B"+strconv.Itoa(rowindex), "K"+strconv.Itoa(rowindex), style)
	f.SetCellValue(title, "B"+strconv.Itoa(rowindex), "Account")
	f.SetCellValue(title, "C"+strconv.Itoa(rowindex), "Services")
	f.SetCellValue(title, "D"+strconv.Itoa(rowindex), "Cluster Identifier")
	f.SetCellValue(title, "E"+strconv.Itoa(rowindex), "DB Identifier")
	f.SetCellValue(title, "F"+strconv.Itoa(rowindex), "Instance Class")
	f.SetCellValue(title, "G"+strconv.Itoa(rowindex), "Engine")
	f.SetCellValue(title, "H"+strconv.Itoa(rowindex), "Allocated Storage")
	f.SetCellValue(title, "I"+strconv.Itoa(rowindex), "Storage Type")
	f.SetCellValue(title, "J"+strconv.Itoa(rowindex), "MultiAZ")
	f.SetCellValue(title, "K"+strconv.Itoa(rowindex), "Usage")

	return rowindex + 1
}

func MakeClusterValue(f *excelize.File, title string, rowindex int, accountName *string, result []*ClusterData, resource string) int {
	for _, cluster := range result {
		rowindex = MakeClusterColumn(f, title, rowindex, "rds")
		style := ValueStyle(f)
		f.SetCellStyle(title, "B"+strconv.Itoa(rowindex), "K"+strconv.Itoa(rowindex), style)
		f.SetCellValue(title, "B"+strconv.Itoa(rowindex), *accountName)
		f.SetCellValue(title, "C"+strconv.Itoa(rowindex), "Cluster")
		f.SetCellValue(title, "D"+strconv.Itoa(rowindex), cluster.Identifier)

		f.SetCellValue(title, "F"+strconv.Itoa(rowindex), cluster.Engine+" "+cluster.EngineVersion)
		f.SetCellValue(title, "G"+strconv.Itoa(rowindex), cluster.Port)
		f.SetCellValue(title, "H"+strconv.Itoa(rowindex), cluster.MultiAZ)
		f.SetCellValue(title, "J"+strconv.Itoa(rowindex), cluster.Status)
		f.SetCellValue(title, "K"+strconv.Itoa(rowindex), cluster.Usage)
		rowindex++
		if cluster.RDSData != nil {

			rowindex = MakeClusterRDSColumn(f, title, rowindex, "rds")
			for _, instance := range cluster.RDSData {
				style := ValueStyle(f)
				f.SetCellStyle(title, "B"+strconv.Itoa(rowindex), "K"+strconv.Itoa(rowindex), style)
				f.SetCellValue(title, "B"+strconv.Itoa(rowindex), *accountName)
				f.SetCellValue(title, "C"+strconv.Itoa(rowindex), title)
				f.SetCellValue(title, "E"+strconv.Itoa(rowindex), instance.Identifier)
				f.SetCellValue(title, "F"+strconv.Itoa(rowindex), instance.InstanceClass)
				f.SetCellValue(title, "G"+strconv.Itoa(rowindex), instance.Engine+" "+instance.EngineVersion)
				f.SetCellValue(title, "H"+strconv.Itoa(rowindex), instance.AllocatedStorage)
				f.SetCellValue(title, "I"+strconv.Itoa(rowindex), instance.StorageType)
				f.SetCellValue(title, "J"+strconv.Itoa(rowindex), instance.MultiAZ)
				f.SetCellValue(title, "K"+strconv.Itoa(rowindex), instance.Usage)

				rowindex++
			}
		}

	}

	return rowindex
}

func MakeClusterColumn(f *excelize.File, title string, rowindex int, resource string) int {

	style := ColumnStyle(f)

	f.SetCellStyle(title, "B"+strconv.Itoa(rowindex), "K"+strconv.Itoa(rowindex), style)
	f.SetCellValue(title, "B"+strconv.Itoa(rowindex), "Account")
	f.SetCellStyle(title, "C"+strconv.Itoa(rowindex), "K"+strconv.Itoa(rowindex), style)
	f.SetCellValue(title, "C"+strconv.Itoa(rowindex), "Service")
	f.SetCellValue(title, "D"+strconv.Itoa(rowindex), "Cluster ID")
	f.SetCellValue(title, "E"+strconv.Itoa(rowindex), "Cluster Class")
	f.SetCellValue(title, "F"+strconv.Itoa(rowindex), "Cluster Engine")
	f.SetCellValue(title, "G"+strconv.Itoa(rowindex), "Cluster Port")
	f.SetCellValue(title, "H"+strconv.Itoa(rowindex), "Cluster MultiAZ")
	f.SetCellValue(title, "I"+strconv.Itoa(rowindex), "Cluster Status")
	f.SetCellValue(title, "J"+strconv.Itoa(rowindex), "Cluster Storage Type")
	f.SetCellValue(title, "K"+strconv.Itoa(rowindex), "Cluster Usage")

	return rowindex + 1
}

func MakeClusterSubtitle(f *excelize.File, title string, subtitle string, rowindex int) int {
	err := f.MergeCell(title, "B"+strconv.Itoa(rowindex), "D"+strconv.Itoa(rowindex))
	if err != nil {
		log.Printf("Make SubTitle Error: %s", err.Error())
	}
	style, index := SubtitleStyle(f, rowindex)
	f.SetCellStyle(title, "B"+strconv.Itoa(index), "D"+strconv.Itoa(index), style)
	f.SetCellValue(title, "B"+strconv.Itoa(index), subtitle)
	return index + 2
}
