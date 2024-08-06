package main

import (
	"log"
	"strconv"
	"time"

	"github.com/xuri/excelize/v2"
)

func MakeExcelReport(title string, resultresources []*ResultResource) string {
	f := excelize.NewFile(excelize.Options{})
	f.SetSheetName("Sheet1", "Default")
	defer func() {

		if err := f.Close(); err != nil {
			log.Printf("MakeExcelReport Error: %s", err.Error())
		}
	}()
	sheetTitles := []string{"EC2", "EIP", "EBS", "SG", "VPN", "LB", "NatGateway", "RDS"}
	for _, sheetTitle := range sheetTitles {
		// 워크시트 만들기

		_, err := f.NewSheet(sheetTitle)
		if err != nil {
			log.Printf("MakeExcelReport Error: %s", err.Error())
			return ""
		}
		// 셀 값 설정
		rowindex := 1
		switch sheetTitle {
		case "EC2":
			rowindex = MakeEC2Subtitle(f, sheetTitle, "1) EC2", rowindex)
			rowindex = MakeEC2Column(f, sheetTitle, rowindex, "ec2")
			for _, resultdata := range resultresources {

				f.SetCellValue(sheetTitle, "B"+strconv.Itoa(rowindex), resultdata.AccountName)
				rowindex = MakeEC2Value(f, sheetTitle, rowindex, &resultdata.AccountName, resultdata.EC2, "ec2")
			}
		case "EIP":
			rowindex = MakeEIPSubtitle(f, sheetTitle, "2) EIP", rowindex)
			rowindex = MakeEIPColumn(f, sheetTitle, rowindex, "eip")
			for _, resultdata := range resultresources {

				f.SetCellValue(sheetTitle, "B"+strconv.Itoa(rowindex), resultdata.AccountName)
				rowindex = MakeEIPValue(f, sheetTitle, rowindex, &resultdata.AccountName, resultdata.EIPData, "eip")
			}
		case "EBS":
			rowindex = MakeEBSSubtitle(f, sheetTitle, "3) EBS", rowindex)
			rowindex = MakeEBSColumn(f, sheetTitle, rowindex, "ebs")
			for _, resultdata := range resultresources {

				f.SetCellValue(sheetTitle, "B"+strconv.Itoa(rowindex), resultdata.AccountName)
				rowindex = MakeEBSValue(f, sheetTitle, rowindex, &resultdata.AccountName, resultdata.EBSData, "ebs")
			}
		case "SG":
			rowindex = MakeSGSubtitle(f, sheetTitle, "4) SG", rowindex)
			rowindex = MakeSGColumn(f, sheetTitle, rowindex, "sg")
			for _, resultdata := range resultresources {

				f.SetCellValue(sheetTitle, "B"+strconv.Itoa(rowindex), resultdata.AccountName)
				rowindex = MakeSGValue(f, sheetTitle, rowindex, &resultdata.AccountName, resultdata.SecurityGroupRuleData, "sg")
			}
		case "VPN":
			rowindex = MakeVPNSubtitle(f, sheetTitle, "5) VPN", rowindex)
			rowindex = MakeVPNColumn(f, sheetTitle, rowindex, "vpn")
			for _, resultdata := range resultresources {

				f.SetCellValue(sheetTitle, "B"+strconv.Itoa(rowindex), resultdata.AccountName)
				rowindex = MakeVPNValue(f, sheetTitle, rowindex, &resultdata.AccountName, resultdata.VPN, "vpn")
			}
		case "LB":
			rowindex = MakeLBSubtitle(f, sheetTitle, "6) LB", rowindex)
			rowindex = MakeLBColumn(f, sheetTitle, rowindex, "lb")
			for _, resultdata := range resultresources {

				f.SetCellValue(sheetTitle, "B"+strconv.Itoa(rowindex), resultdata.AccountName)
				rowindex = MakeLBValue(f, sheetTitle, rowindex, &resultdata.AccountName, resultdata.LB, "lb")
			}
		case "NatGateway":
			rowindex = MakeNatGatewaySubtitle(f, sheetTitle, "7) NAT Gateawy", rowindex)
			rowindex = MakeNatGatewayColumn(f, sheetTitle, rowindex, "ngw")
			for _, resultdata := range resultresources {

				f.SetCellValue(sheetTitle, "B"+strconv.Itoa(rowindex), resultdata.AccountName)
				rowindex = MakeNatGatewayValue(f, sheetTitle, rowindex, &resultdata.AccountName, resultdata.NatGatewayData, "ngw")
			}
		case "RDS":
			rowindex = MakeClusterSubtitle(f, sheetTitle, "8) Cluster", rowindex)

			for _, resultdata := range resultresources {
				f.SetCellValue(sheetTitle, "B"+strconv.Itoa(rowindex), resultdata.AccountName)
				rowindex = MakeClusterValue(f, sheetTitle, rowindex, &resultdata.AccountName, resultdata.ClusterData, "rds")

			}

			rowindex = MakeRDSSubtitle(f, sheetTitle, "9) RDS", rowindex)
			rowindex = MakeRDSColumn(f, sheetTitle, rowindex, "rds")
			for _, resultdata := range resultresources {
				f.SetCellValue(sheetTitle, "B"+strconv.Itoa(rowindex), resultdata.AccountName)
				rowindex = MakeRDSValue(f, sheetTitle, rowindex, &resultdata.AccountName, resultdata.RDSData, "rds")
			}
		default:
			break
		}

	}

	// 통합 문서에 대 한 기본 워크시트를 설정 합니다
	f.SetActiveSheet(0)
	// 지정 된 경로를 기반으로 파일 저장
	f.DeleteSheet("Default")
	filename := title + getDateString(time.Now().Year()) + "_" + getDateString(int(time.Now().Month())) + getDateString(time.Now().Day()) + ".xlsx"
	if err := f.SaveAs("./" + filename); err != nil {
		log.Printf("MakeExcelReport Error: %s", err.Error())
	}
	return filename
}

func getDateString(date int) string {
	if date < 10 {
		return "0" + strconv.Itoa(date)
	} else {
		return strconv.Itoa(date)
	}
}

func ValueStyle(f *excelize.File) int {
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
			Bold: false,
			Size: 10,
		},
	})
	if err != nil {
		log.Printf("MakeValue Error: %s", err.Error())
		return 0
	}
	return style
}

func ColumnStyle(f *excelize.File) int {
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
	return style
}

func MergedTitleStyle(f *excelize.File) int {
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
		log.Printf("MergedTitle Error: %s", err.Error())
		return 0
	}
	return style
}

func SubtitleStyle(f *excelize.File, rowindex int) (int, int) {
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
		return 0, rowindex + 1
	}
	return style, rowindex
}

func TitleStyle(f *excelize.File) int {
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
	return style
}
