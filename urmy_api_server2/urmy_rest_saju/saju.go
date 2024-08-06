package saju

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"math"
	"strconv"
	"sync"
	"time"
)

type SaJuAnalyzerInterface interface {
	ExtractSaju(saju *Saju) *SaJuPalJa
	ExtractDaeUnSaeUn(saju *Saju, palja *SaJuPalJa) *SaJuPalJa
	ParseSaju(year int, month int, day int, hour int, min int, gender bool) *Saju
	CheckUpdateUrMyDaeSaeUn(lastlogin time.Time) bool
	Evaluate_GoonbHab(host Person, opponent Person) (float64, float64, string, string)
	Find_GoongHab(mysaju *SaJuPalJa, mydaesaeun *DaeSaeUn, friendsaju *SaJuPalJa, frienddaesaeun *DaeSaeUn, saju_table *SajuTable, sajuanalyzer *SajuAnalyzer) (*Person, *Person)
}

func (sa *SajuAnalyzer) CheckUpdateUrMyDaeSaeUn(lastlogin time.Time) bool {
	index := time.Now().Year() - sa.Manse[0].WhichYear
	Month := int(time.Now().Month())
	julgy := time.Date(sa.Julgys[index].Year, time.Month(sa.Julgys[index].MonthJulgys[Month-1].Month), sa.Julgys[index].MonthJulgys[Month-1].DayJulgys[0].Day, 0, 0, 0, 0, time.UTC)
	if lastlogin.Before(julgy) && time.Now().After(julgy) {
		return true
	}
	return false
}

func (sa *SajuAnalyzer) ParseSaju(year int, month int, day int, hour int, min int, gender bool) *Saju {
	var saju Saju
	var yajasi bool
	saju.Gender = gender
	temp := time.Date(year, time.Month(month), day, hour, min, 0, 0, time.UTC)
	_, saju.Time, yajasi = inTimeSpan(formatTime(hour, min))
	if yajasi {
		temp.AddDate(0, 0, 1)
	}
	saju.Year, saju.Month, saju.Day = year, month, day
	return &saju
}

func (sa *SajuAnalyzer) ExtractSaju(saju *Saju) *SaJuPalJa {
	var t SaJuPalJa
	index := saju.Year - sa.Manse[0].WhichYear

	tbirth := time.Date(saju.Year, time.Month(saju.Month), saju.Day, 0, 0, 0, 0, time.UTC)
	yjulgy := time.Date(sa.Julgys[index].Year, time.Month(sa.Julgys[index].MonthJulgys[1].Month), sa.Julgys[index].MonthJulgys[1].DayJulgys[0].Day, 0, 0, 0, 0, time.UTC)
	mjulgy := time.Date(sa.Julgys[index].Year, time.Month(sa.Julgys[index].MonthJulgys[saju.Month-1].Month), sa.Julgys[index].MonthJulgys[saju.Month-1].DayJulgys[0].Day, 0, 0, 0, 0, time.UTC)
	indexTime, _, _ := inTimeSpan(saju.Time)

	if tbirth.Before(yjulgy) {
		t.YearChun = sa.Manse[index-1].Chungan_Title
		t.YearJi = sa.Manse[index-1].Jiji_Title
	} else {
		t.YearChun = sa.Manse[index].Chungan_Title
		t.YearJi = sa.Manse[index].Jiji_Title
	}

	if tbirth.Before(mjulgy) {
		if saju.Month == 1 {
			t.MonthChun = sa.Manse[index-1].MonthGanji[11].Chungan_Title
			t.MonthJi = sa.Manse[index-1].MonthGanji[11].Jiji_Title
		} else {
			// 월 배열 -1 , 절기 전 -1
			t.MonthChun = sa.Manse[index].MonthGanji[saju.Month-2].Chungan_Title
			t.MonthJi = sa.Manse[index].MonthGanji[saju.Month-2].Jiji_Title
		}
	} else {
		t.MonthChun = sa.Manse[index].MonthGanji[saju.Month-1].Chungan_Title
		t.MonthJi = sa.Manse[index].MonthGanji[saju.Month-1].Jiji_Title
	}

	t.DayChun = sa.Manse[index].MonthGanji[saju.Month-1].Day_Ganji[saju.Day-1].Chungan_Title
	t.DayJi = sa.Manse[index].MonthGanji[saju.Month-1].Day_Ganji[saju.Day-1].Jiji_Title

	if indexTime == -1 {
		t.TimeChun = "none"
		t.TimeJi = "none"
	} else {
		t.TimeChun = sa.Manse[index].MonthGanji[saju.Month-1].Day_Ganji[saju.Day-1].Time_Ganji[indexTime].Chungan_Title
		t.TimeJi = sa.Manse[index].MonthGanji[saju.Month-1].Day_Ganji[saju.Day-1].Time_Ganji[indexTime].Jiji_Title
	}
	return &t
}

func (sa *SajuAnalyzer) ExtractDaeUnSaeUn(saju *Saju, palja *SaJuPalJa) *DaeSaeUn {

	var daesaeun DaeSaeUn
	var index int
	var daeunindex int
	currentdaeunindex := 0
	age := time.Now().Year() - saju.Year
	if time.Now().Before(time.Date(saju.Year, time.Month(saju.Month), saju.Day, 0, 0, 0, 0, time.Local)) {
		age--
	}
	daesaeun.SaeUnChun = sa.SaeUn[time.Now().Year()-1950].SaeUnGanjis.Chun
	daesaeun.SaeUnJi = sa.SaeUn[time.Now().Year()-1950].SaeUnGanjis.Ji

	for i := 0; i < 10; i++ {
		if sa.Chungan[i].Title == palja.YearChun {
			index = i
			break
		}
	}
	switch saju.Gender {
	case true:
		switch sa.Chungan[index].Properties.Umyang { //1991807
		case 1:
			daeunindex = sa.forwardDaeunIndex(saju, palja)
		case 0:
			daeunindex = sa.reverseDaeunIndex(saju, palja)
		}

	case false:
		switch sa.Chungan[index].Properties.Umyang { //1991807
		case 1:
			daeunindex = sa.reverseDaeunIndex(saju, palja)
		case 0:
			daeunindex = sa.forwardDaeunIndex(saju, palja)
		}
	}

	for i := 0; (age - (i*10 + daeunindex)) >= 0; i++ {
		currentdaeunindex += 1
	}

	//if gender is male go true
	switch saju.Gender {
	case true:
		switch sa.Chungan[index].Properties.Umyang {
		case 1:
			daesaeun = sa.forwardDaeunGanji(currentdaeunindex, palja, daesaeun)
		case 0:
			daesaeun = sa.reverseDaeunGanji(currentdaeunindex, palja, daesaeun)
		}
	case false:
		switch sa.Chungan[index].Properties.Umyang {
		case 1:
			daesaeun = sa.reverseDaeunGanji(currentdaeunindex, palja, daesaeun)
		case 0:
			daesaeun = sa.forwardDaeunGanji(currentdaeunindex, palja, daesaeun)
		}
	}

	return &daesaeun
}

func (sa *SajuAnalyzer) forwardDaeunIndex(saju *Saju, palja *SaJuPalJa) int {
	var daeunindex int
	tjulgy := time.Date(saju.Year, time.Month(saju.Month), sa.Julgys[saju.Year-1950].MonthJulgys[saju.Month-1].DayJulgys[0].Day, 0, 0, 0, 0, time.UTC)
	tbirth := time.Date(saju.Year, time.Month(saju.Month), saju.Day, 0, 0, 0, 0, time.UTC)

	if tbirth.After(tjulgy) {
		if saju.Month+1 == 13 {
			tjulgy = time.Date(saju.Year+1, time.Month(saju.Month-11), sa.Julgys[saju.Year-1949].MonthJulgys[0].DayJulgys[0].Day, 0, 0, 0, 0, time.UTC)
		} else {
			tjulgy = time.Date(saju.Year, time.Month(saju.Month+1), sa.Julgys[saju.Year-1950].MonthJulgys[saju.Month].DayJulgys[0].Day, 0, 0, 0, 0, time.UTC)
		}
	} else if tjulgy == tbirth {
		daeunindex = 1
		return daeunindex
	}
	e := tjulgy.Sub(tbirth).Hours() / 72
	daeunindex = int(math.Round(e))
	return daeunindex
}

func (sa *SajuAnalyzer) reverseDaeunIndex(saju *Saju, palja *SaJuPalJa) int {
	var daeunindex int
	tjulgy := time.Date(saju.Year, time.Month(saju.Month), sa.Julgys[saju.Year-1950].MonthJulgys[saju.Month-1].DayJulgys[0].Day, 0, 0, 0, 0, time.UTC)
	tbirth := time.Date(saju.Year, time.Month(saju.Month), saju.Day, 0, 0, 0, 0, time.UTC)
	if tbirth.Before(tjulgy) {
		if saju.Month-1 == 0 {
			tjulgy = time.Date(saju.Year-1, time.Month(12), sa.Julgys[saju.Year-1950].MonthJulgys[11].DayJulgys[0].Day, 0, 0, 0, 0, time.UTC)
		} else {
			tjulgy = time.Date(saju.Year, time.Month(saju.Month-1), sa.Julgys[saju.Year-1950].MonthJulgys[saju.Month-2].DayJulgys[0].Day, 0, 0, 0, 0, time.UTC)
		}
	} else if tjulgy == tbirth {
		daeunindex = 1
		return daeunindex
	}
	e := tbirth.Sub(tjulgy).Hours() / 72
	daeunindex = int(math.Round(e))
	return daeunindex
}

func (sa *SajuAnalyzer) forwardDaeunGanji(currentdaeunindex int, palja *SaJuPalJa, daesaeun DaeSaeUn) DaeSaeUn {

	for i := 0; i < 10; i++ {
		if palja.MonthChun == sa.Chungan[i].Title {
			if i+currentdaeunindex < 10 {
				daesaeun.DaeUnChun = sa.Chungan[i+currentdaeunindex].Title
			} else {
				daesaeun.DaeUnChun = sa.Chungan[i+currentdaeunindex-10].Title
			}
		}
	}
	for i := 0; i < 12; i++ {
		if palja.MonthJi == sa.Jiji[i].Title {
			if i+currentdaeunindex < 12 {
				daesaeun.DaeUnJi = sa.Jiji[i+currentdaeunindex].Title
			} else {
				daesaeun.DaeUnJi = sa.Jiji[i+currentdaeunindex-12].Title
			}
		}
	}
	return daesaeun
}

func (sa *SajuAnalyzer) reverseDaeunGanji(currentdaeunindex int, palja *SaJuPalJa, daesaeun DaeSaeUn) DaeSaeUn {
	for i := 9; i >= 0; i-- {
		if palja.MonthChun == sa.Chungan[i].Title {
			if i-currentdaeunindex < 0 {
				daesaeun.DaeUnChun = sa.Chungan[i-currentdaeunindex+10].Title
			} else {
				daesaeun.DaeUnChun = sa.Chungan[i-currentdaeunindex].Title
			}
		}
	}
	for i := 11; i >= 0; i-- {
		if palja.MonthJi == sa.Jiji[i].Title {
			if i-currentdaeunindex < 0 {
				daesaeun.DaeUnJi = sa.Jiji[i-currentdaeunindex+12].Title
			} else {
				daesaeun.DaeUnJi = sa.Jiji[i-currentdaeunindex].Title
			}
		}
	}
	return daesaeun
}

func inTimeSpan(check string) (int, string, bool) {
	//valueofChar := [12]rune{'자', '축', '인', '묘', '진', '사', '오', '미', '신', '유', '술', '해'}
	valueofChar := [13]int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 0}
	data := []struct {
		start string
		end   string
	}{
		{"00:00", "01:30"},
		{"01:29", "03:30"},
		{"03:29", "05:30"},
		{"05:29", "07:30"},
		{"07:29", "09:30"},
		{"09:29", "11:30"},
		{"11:29", "13:30"},
		{"13:29", "15:30"},
		{"15:29", "17:30"},
		{"17:29", "19:30"},
		{"19:29", "21:30"},
		{"21:29", "23:30"},
		{"23:29", "24:00"},
	}

	resultdata := []struct {
		time string
	}{
		{"00:00"},
		{"01:30"},
		{"03:30"},
		{"05:30"},
		{"07:30"},
		{"09:30"},
		{"11:30"},
		{"13:30"},
		{"15:30"},
		{"17:30"},
		{"19:30"},
		{"21:30"},
	}

	if check == "25:00" {
		return -1, "25:00", false
	}

	if check == "00:00" {
		return 0, resultdata[0].time, false
	}

	newLayout := "15:04"
	checked, _ := time.Parse(newLayout, check)

	a := 0
	for _, t := range data {
		start, _ := time.Parse(newLayout, t.start)
		end, _ := time.Parse(newLayout, t.end)

		if checked.After(start) && checked.Before(end) {
			break
		}
		a++
	}
	if a == 12 {
		return valueofChar[0], resultdata[0].time, true
	}
	return valueofChar[a], resultdata[a].time, false
}

func formatTime(hour int, min int) string {
	formattedhour := strconv.Itoa(hour)
	formattedmin := strconv.Itoa(min)
	if formattedhour == "0" {
		formattedhour = "00"
	}
	if formattedmin == "0" {
		formattedmin = "00"
	}
	if len(formattedhour) == 1 {
		formattedhour = "0" + formattedhour
	}

	if len(formattedmin) == 1 {
		formattedmin = "0" + formattedmin
	}

	return formattedhour + ":" + formattedmin

}

func (saju *SajuAnalyzer) person_chungan_input(a []*string) *Person {
	var c Person
	sajutableChungan := saju.Chungan
	sajutableJiji := saju.Jiji
	sajutableSibsung := saju.Sibsung
	sajutableSib2Unsung := saju.Sib2Unsung
	c.Chun = make([]Chungan, 6)
	c.Ji = make([]Jiji, 6)
	c.Result = make([]Result_record, 6)
	for i := 0; i < len(a); i++ {
		if i < 6 {
			for j := 0; j < 10; j++ {
				if *a[i] == sajutableChungan[j].Title {
					c.Chun[i] = sajutableChungan[j]

					break
				}
			}
		} else {
			for j := 0; j < 12; j++ {
				if *a[i] == sajutableJiji[j].Title {
					c.Ji[i-len(a)/2] = sajutableJiji[j]

					break
				}
			}
		}
	}

	for i := 0; i < 4; i++ {
		switch c.Ji[i].Title {
		case "진":
			for j := 0; j < 4; j++ {
				if c.Chun[j].Title == "신" || c.Chun[j].Title == "임" {
					c.Chun[j].Properties.IpMyo = WhichMyJi
				}
			}
		case "술":
			for j := 0; j < 4; j++ {
				if c.Chun[j].Title == "병" || c.Chun[j].Title == "무" || c.Chun[j].Title == "을" {
					c.Chun[j].Properties.IpMyo = WhichMyJi
				}
			}
		case "축":
			for j := 0; j < 4; j++ {
				if c.Chun[j].Title == "경" || c.Chun[j].Title == "기" || c.Chun[j].Title == "정" {
					c.Chun[j].Properties.IpMyo = WhichMyJi
				}
			}
		case "미":
			for j := 0; j < 4; j++ {
				if c.Chun[j].Title == "갑" || c.Chun[j].Title == "해" {
					c.Chun[j].Properties.IpMyo = WhichMyJi
				}
			}
		}
	}

	for i := 0; i < 5; i++ { //십성 (본관을 넣어야함)
		if c.Chun[2].Properties.Prop == sajutableSibsung[i].Prop {
			for j := 0; j < 6; j++ {
				for k := 0; k < 5; k++ {
					if c.Chun[j].Properties.Prop == sajutableSibsung[i].Comp_Prop[k].Comp_Prop {
						//c.Chun[j].Properties.Sibsung = "본관"
						c.Chun[j].Properties.Sibsung = sajutableSibsung[i].Comp_Prop[k].Title
					}

					if c.Ji[j].Properties.Prop == sajutableSibsung[i].Comp_Prop[k].Comp_Prop {
						c.Ji[j].Properties.Sibsung = sajutableSibsung[i].Comp_Prop[k].Title
					}

				}
			}
			break
		}
	}

	//십이운성
	for i := 0; i < len(a)/2; i++ {
		for j := 0; j < 10; j++ {
			if c.Chun[i].Title == sajutableSib2Unsung[j].Title {
				for k := 0; k < 12; k++ {
					if c.Ji[i].Title == sajutableSib2Unsung[j].Properties[k].Jiji_char {
						c.Chun[i].Properties.Unsung_by_Jiji.level = sajutableSib2Unsung[j].Properties[k].Level
						c.Chun[i].Properties.Unsung_by_Jiji.Unsung_title = sajutableSib2Unsung[j].Properties[k].Prop
					}
				}
			}

			if c.Chun[2].Title == sajutableSib2Unsung[j].Title {
				for k := 0; k < 12; k++ {
					if c.Ji[i].Title == sajutableSib2Unsung[j].Properties[k].Jiji_char {
						c.Chun[i].Properties.Unsung_Me.level = sajutableSib2Unsung[j].Properties[k].Level
						c.Chun[i].Properties.Unsung_Me.Unsung_title = sajutableSib2Unsung[j].Properties[k].Prop
					}
				}
			}
		}
	}

	//창고

	for i := 0; i < len(a)/2; i++ {
		c.Ji[i].Properties.ChangGo = 0
		if c.Ji[i].Properties.Prop == "earth" {
			switch c.Chun[2].Properties.Prop {
			case "tree":
				switch c.Ji[i].Properties.Jijanggans[1].Chungan_char.Properties.Prop {
				case "tree":
					c.Ji[i].Properties.ChangGo = YangInGo
				case "fire":
					c.Ji[i].Properties.ChangGo = SikSangGo
				case "iron":
					c.Ji[i].Properties.ChangGo = GwanGo
				case "water":
					c.Ji[i].Properties.ChangGo = InSungGo
				}
			case "fire":
				switch c.Ji[i].Properties.Jijanggans[1].Chungan_char.Properties.Prop {
				case "tree":
					c.Ji[i].Properties.ChangGo = InSungGo
				case "fire":
					c.Ji[i].Properties.ChangGo = YangInGo
				case "iron":
					c.Ji[i].Properties.ChangGo = JaeGo
				case "water":
					c.Ji[i].Properties.ChangGo = GwanGo
				}
			case "earth":
				switch c.Ji[i].Properties.Jijanggans[1].Chungan_char.Properties.Prop {
				case "tree":
					c.Ji[i].Properties.ChangGo = GwanGo
				case "fire":
					c.Ji[i].Properties.ChangGo = InSungGo
				case "iron":
					c.Ji[i].Properties.ChangGo = SikSangGo
				case "water":
					c.Ji[i].Properties.ChangGo = JaeGo
				}
			case "iron":
				switch c.Ji[i].Properties.Jijanggans[1].Chungan_char.Properties.Prop {
				case "tree":
					c.Ji[i].Properties.ChangGo = JaeGo
				case "fire":
					c.Ji[i].Properties.ChangGo = GwanGo
				case "iron":
					c.Ji[i].Properties.ChangGo = YangInGo
				case "water":
					c.Ji[i].Properties.ChangGo = SikSangGo
				}
			case "water":
				switch c.Ji[i].Properties.Jijanggans[1].Chungan_char.Properties.Prop {
				case "tree":
					c.Ji[i].Properties.ChangGo = SikSangGo
				case "fire":
					c.Ji[i].Properties.ChangGo = JaeGo
				case "iron":
					c.Ji[i].Properties.ChangGo = InSungGo
				case "water":
					c.Ji[i].Properties.ChangGo = YangInGo
				}
			}
		}
	}
	return &c
}

func Find_Chungan_hab(wg *sync.WaitGroup, mutex *sync.Mutex, a *Person, b *Person) {
	mutex.Lock()
	for i := 1; i < 3; i++ {
		a.Result[i].ChunGanHab = 0
		b.Result[i].ChunGanHab = 0
		if math.Abs(float64(a.Chun[i].Id-b.Chun[i].Id)) == 5 {
			switch {
			case a.Chun[i].Title == "갑" || a.Chun[i].Title == "기":
				a.Result[i].ChunGanHab = GabGi
				b.Result[i].ChunGanHab = GabGi
			case a.Chun[i].Title == "을" || a.Chun[i].Title == "경":
				a.Result[i].ChunGanHab = ElGyeong
				b.Result[i].ChunGanHab = ElGyeong
			case a.Chun[i].Title == "병" || a.Chun[i].Title == "신":
				a.Result[i].ChunGanHab = ByeongSin
				b.Result[i].ChunGanHab = ByeongSin
			case a.Chun[i].Title == "정" || a.Chun[i].Title == "임":
				a.Result[i].ChunGanHab = JeongIm
				b.Result[i].ChunGanHab = JeongIm
			case a.Chun[i].Title == "무" || a.Chun[i].Title == "계":
				a.Result[i].ChunGanHab = MuGye
				b.Result[i].ChunGanHab = MuGye
			}
		}
	}
	mutex.Unlock()
	wg.Done()
	//return a, b
}

func Find_Chungan_Geok(wg *sync.WaitGroup, mutex *sync.Mutex, a *Person, b *Person) {
	mutex.Lock()
	for i := 1; i < 3; i++ {
		a.Result[i].ChunGanGeok = 0
		b.Result[i].ChunGanGeok = 0
		if math.Abs(float64(a.Chun[i].Id-b.Chun[i].Id)) == 6 { //내가 극 당하는 것
			switch {
			case a.Chun[i].Title == "갑":
				a.Result[i].ChunGanGeok = (-1 * GabGyeong)
			case a.Chun[i].Title == "을":
				a.Result[i].ChunGanGeok = (-1 * ElSin)

			case a.Chun[i].Title == "병":
				a.Result[i].ChunGanGeok = (-1 * ByeongIm)
			case a.Chun[i].Title == "정":
				a.Result[i].ChunGanGeok = (-1 * JeongHye)
			case a.Chun[i].Title == "무":
				a.Result[i].ChunGanGeok = (-1 * MuGab)
			case a.Chun[i].Title == "기":
				a.Result[i].ChunGanGeok = (-1 * GiEl)
			case a.Chun[i].Title == "경":
				a.Result[i].ChunGanGeok = (-1 * GyeongByeong)
			case a.Chun[i].Title == "신":
				a.Result[i].ChunGanGeok = (-1 * SinJeong)
			case a.Chun[i].Title == "임":
				a.Result[i].ChunGanGeok = (-1 * ImMu)
			case a.Chun[i].Title == "계":
				a.Result[i].ChunGanGeok = (-1 * GyeGi)
			}
		}

		if math.Abs(float64(a.Chun[i].Id-b.Chun[i].Id)) == 4 { // 상대방이 극 당하는 것
			switch {
			case a.Chun[i].Title == "갑":
				b.Result[i].ChunGanGeok = (-1 * GabGyeong)
			case a.Chun[i].Title == "을":
				b.Result[i].ChunGanGeok = (-1 * ElSin)
			case a.Chun[i].Title == "병":
				b.Result[i].ChunGanGeok = (-1 * ByeongIm)
			case a.Chun[i].Title == "정":
				b.Result[i].ChunGanGeok = (-1 * JeongHye)
			case a.Chun[i].Title == "무":
				b.Result[i].ChunGanGeok = (-1 * MuGab)
			case a.Chun[i].Title == "기":
				b.Result[i].ChunGanGeok = (-1 * GiEl)
			case a.Chun[i].Title == "경":
				b.Result[i].ChunGanGeok = (-1 * GyeongByeong)
			case a.Chun[i].Title == "신":
				b.Result[i].ChunGanGeok = (-1 * SinJeong)
			case a.Chun[i].Title == "임":
				b.Result[i].ChunGanGeok = (-1 * ImMu)
			case a.Chun[i].Title == "계":
				b.Result[i].ChunGanGeok = (-1 * GyeGi)
			}
		}
	}
	mutex.Unlock()
	wg.Done()
}

func (saju *SajuAnalyzer) Find_Sibsung_Goonghab(wg *sync.WaitGroup, mutex *sync.Mutex, a *Person, b *Person) {
	e := saju.Sibsung
	mutex.Lock()
	for i := 0; i < 4; i++ {
		if a.Chun[1].Properties.Prop == e[i].Prop {
			for k := 0; k < 4; k++ {
				if b.Chun[1].Properties.Prop == e[i].Comp_Prop[k].Comp_Prop {
					a.Chun[1].Properties.Sibsung = e[i].Comp_Prop[k].Title
				}
			}
		}

		if a.Chun[2].Properties.Prop == e[i].Prop {
			for k := 0; k < 4; k++ {
				if b.Chun[2].Properties.Prop == e[i].Comp_Prop[k].Comp_Prop {
					a.Chun[2].Properties.Sibsung = e[i].Comp_Prop[k].Title
				}
			}
		}

		if b.Chun[1].Properties.Prop == e[i].Prop {
			for k := 0; k < 4; k++ {
				if a.Chun[1].Properties.Prop == e[i].Comp_Prop[k].Comp_Prop {
					b.Chun[1].Properties.Sibsung = e[i].Comp_Prop[k].Title
				}
			}
		}

		if b.Chun[2].Properties.Prop == e[i].Prop {
			for k := 0; k < 4; k++ {
				if a.Chun[2].Properties.Prop == e[i].Comp_Prop[k].Comp_Prop {
					b.Chun[2].Properties.Sibsung = e[i].Comp_Prop[k].Title
				}
			}
		}
	}
	a.Result[1].Sibsung = 0
	b.Result[2].Sibsung = 0
	a.Result[1].Sibsung = 0
	b.Result[2].Sibsung = 0
	switch a.Chun[2].Properties.Sibsung {
	case "비겁":
		switch b.Chun[2].Properties.Sibsung {
		case "식상":
			if a.Chun[1].Properties.Umyang == a.Chun[2].Properties.Umyang {
				if a.Chun[2].Properties.Umyang == b.Chun[2].Properties.Umyang {
					b.Result[2].Sibsung = SibsungPlus2
				} else {
					b.Result[2].Sibsung = SibsungPlus4
				}
			} else {
				if a.Chun[2].Properties.Umyang == b.Chun[2].Properties.Umyang {
					b.Result[2].Sibsung = SibsungPlus2
				} else {
					b.Result[2].Sibsung = SibsungPlus4
				}
			}
		case "재성":
			if a.Chun[1].Properties.Umyang == a.Chun[2].Properties.Umyang {
				if a.Chun[2].Properties.Umyang == b.Chun[2].Properties.Umyang {
					b.Result[2].Sibsung = SibsungMinus3
				} else {
					b.Result[2].Sibsung = SibsungMinus1
				}
			} else {
				b.Result[2].Sibsung = SibsungMinus4
			}
		case "관성":
			if a.Chun[1].Properties.Umyang == a.Chun[2].Properties.Umyang {
				if a.Chun[2].Properties.Umyang == b.Chun[2].Properties.Umyang {
					a.Result[2].Sibsung = SibsungMinus4
				} else {
					a.Result[2].Sibsung = SibsungMinus1
				}
			} else {
				if a.Chun[2].Properties.Umyang == b.Chun[2].Properties.Umyang {
					a.Result[2].Sibsung = SibsungMinus2
				} else {
					a.Result[2].Sibsung = SibsungMinus4
				}
			}
		case "인성":
			if a.Chun[1].Properties.Umyang == a.Chun[2].Properties.Umyang { //비견
				if a.Chun[2].Properties.Umyang == b.Chun[2].Properties.Umyang { //편
					a.Result[2].Sibsung = SibsungPlus2
				} else { //정
					a.Result[2].Sibsung = SibsungPlus4
				}
			} else { //겁재
				if a.Chun[2].Properties.Umyang == b.Chun[2].Properties.Umyang { //정
					a.Result[2].Sibsung = SibsungPlus2
				} else { //편
					a.Result[2].Sibsung = SibsungPlus4
				}
			}
		}
	case "식상":
		switch b.Chun[2].Properties.Sibsung {
		case "비겁":
			if a.Chun[1].Properties.Umyang == a.Chun[2].Properties.Umyang {
				if a.Chun[2].Properties.Umyang == b.Chun[2].Properties.Umyang {
					a.Result[2].Sibsung = SibsungPlus2
				} else {
					a.Result[2].Sibsung = SibsungPlus4
				}
			} else {
				if a.Chun[2].Properties.Umyang == b.Chun[2].Properties.Umyang {
					a.Result[2].Sibsung = SibsungPlus2
				} else {
					a.Result[2].Sibsung = SibsungPlus4
				}
			}
		case "재성":
			if a.Chun[1].Properties.Umyang == a.Chun[2].Properties.Umyang {
				if a.Chun[2].Properties.Umyang == b.Chun[2].Properties.Umyang {
					b.Result[2].Sibsung = SibsungPlus3
				} else {
					b.Result[2].Sibsung = SibsungPlus3
				}
			} else {
				if a.Chun[2].Properties.Umyang == b.Chun[2].Properties.Umyang {
					b.Result[2].Sibsung = SibsungPlus3
				} else {
					b.Result[2].Sibsung = SibsungPlus3
				}
			}

		case "관성":
			if a.Chun[1].Properties.Umyang == a.Chun[2].Properties.Umyang { //편
				if a.Chun[2].Properties.Umyang == b.Chun[2].Properties.Umyang { //편
					b.Result[2].Sibsung = SibsungPlus4
				} else { //정
					b.Result[2].Sibsung = SibsungMinus4
				}
			} else { //정
				if a.Chun[2].Properties.Umyang == b.Chun[2].Properties.Umyang { // 정
					b.Result[2].Sibsung = SibsungMinus4
				} else { //편
					b.Result[2].Sibsung = SibsungPlus1
				}
			}

		case "인성":
			if a.Chun[1].Properties.Umyang == a.Chun[2].Properties.Umyang { //편
				if a.Chun[2].Properties.Umyang == b.Chun[2].Properties.Umyang { //편
					a.Result[2].Sibsung = SibsungMinus3
				} else { //정
					if a.Chun[2].Properties.Umyang == 1 {
						a.Result[2].Sibsung = SibsungMinus3
					} else {
						a.Result[2].Sibsung = SibsungPlus2
						b.Result[2].Sibsung = SibsungPlus2
					}
				}
			} else { //정
				if a.Chun[2].Properties.Umyang == b.Chun[2].Properties.Umyang { // 정
					a.Result[2].Sibsung = SibsungMinus4
				} else { //편
					if a.Chun[2].Properties.Umyang == 1 {
						a.Result[2].Sibsung = SibsungPlus2
						b.Result[2].Sibsung = SibsungPlus2
					} else {
						a.Result[2].Sibsung = SibsungMinus4
					}
				}
			}
		}
	case "재성":
		switch b.Chun[2].Properties.Sibsung {
		case "비겁":
			if a.Chun[1].Properties.Umyang == a.Chun[2].Properties.Umyang { //편
				if a.Chun[2].Properties.Umyang == b.Chun[2].Properties.Umyang { //편
					a.Result[2].Sibsung = SibsungMinus3
				} else { //정
					a.Result[2].Sibsung = SibsungMinus4
				}
			} else { //정
				if a.Chun[2].Properties.Umyang == b.Chun[2].Properties.Umyang { //정
					a.Result[2].Sibsung = SibsungMinus4
				} else { //편
					a.Result[2].Sibsung = SibsungMinus1
				}
			}
		case "식상":
			if a.Chun[1].Properties.Umyang == a.Chun[2].Properties.Umyang { //편
				if a.Chun[2].Properties.Umyang == b.Chun[2].Properties.Umyang { //편
					a.Result[2].Sibsung = SibsungPlus3
				} else { //정
					a.Result[2].Sibsung = SibsungPlus3
				}
			} else { //정
				if a.Chun[2].Properties.Umyang == b.Chun[2].Properties.Umyang { //정
					a.Result[2].Sibsung = SibsungPlus3
				} else { //편
					a.Result[2].Sibsung = SibsungPlus3
				}
			}
		case "관성":
			if a.Chun[1].Properties.Umyang == a.Chun[2].Properties.Umyang { //편
				if a.Chun[2].Properties.Umyang == b.Chun[2].Properties.Umyang { //편
					b.Result[2].Sibsung = SibsungMinus3
				} else { //정
					b.Result[2].Sibsung = SibsungPlus3
				}
			} else { //정
				if a.Chun[2].Properties.Umyang == b.Chun[2].Properties.Umyang { //정
					a.Result[2].Sibsung = SibsungPlus3
				} else { //편
					b.Result[2].Sibsung = SibsungMinus3
				}
			}
		case "인성":
			if a.Chun[1].Properties.Umyang == a.Chun[2].Properties.Umyang { //편
				if a.Chun[2].Properties.Umyang == b.Chun[2].Properties.Umyang { //편
					b.Result[2].Sibsung = SibsungMinus2
				} else { //정
					b.Result[2].Sibsung = SibsungMinus2
				}
			} else { //정
				if a.Chun[2].Properties.Umyang == b.Chun[2].Properties.Umyang { //정
					a.Result[2].Sibsung = SibsungMinus2
				} else { //편
					b.Result[2].Sibsung = SibsungMinus2
				}
			}
		}
	case "관성":
		switch b.Chun[2].Properties.Sibsung {
		case "비겁":
			if a.Chun[1].Properties.Umyang == a.Chun[2].Properties.Umyang { //편
				if a.Chun[2].Properties.Umyang == b.Chun[2].Properties.Umyang { //편
					b.Result[2].Sibsung = SibsungMinus4
				} else { //정
					b.Result[2].Sibsung = SibsungMinus4
				}
			} else { //정
				if a.Chun[2].Properties.Umyang == b.Chun[2].Properties.Umyang { //정
					b.Result[2].Sibsung = SibsungMinus2
				} else { //편
					b.Result[2].Sibsung = SibsungMinus1
				}
			}
		case "식상":
			if a.Chun[1].Properties.Umyang == a.Chun[2].Properties.Umyang { //편
				if a.Chun[2].Properties.Umyang == b.Chun[2].Properties.Umyang { //편
					a.Result[2].Sibsung = SibsungMinus4
				} else { //정
					a.Result[2].Sibsung = SibsungPlus1
				}
			} else { //정
				if a.Chun[2].Properties.Umyang == b.Chun[2].Properties.Umyang { //정
					a.Result[2].Sibsung = SibsungMinus4
				} else { //편
					a.Result[2].Sibsung = SibsungPlus1
				}
			}
		case "재성":
			if a.Chun[1].Properties.Umyang == a.Chun[2].Properties.Umyang { //편
				if a.Chun[2].Properties.Umyang == b.Chun[2].Properties.Umyang { //편
					a.Result[2].Sibsung = SibsungMinus3
				} else { //정
					a.Result[2].Sibsung = SibsungMinus3
				}
			} else { //정
				if a.Chun[2].Properties.Umyang == b.Chun[2].Properties.Umyang { //정
					a.Result[2].Sibsung = SibsungPlus3
				} else { //편
					a.Result[2].Sibsung = SibsungPlus3
				}
			}
		case "인성":
			if a.Chun[1].Properties.Umyang == a.Chun[2].Properties.Umyang { //편
				if a.Chun[2].Properties.Umyang == b.Chun[2].Properties.Umyang { //편
					a.Result[2].Sibsung = SibsungPlus2
					b.Result[2].Sibsung = SibsungPlus2
				} else { //정
					a.Result[2].Sibsung = SibsungPlus2
					b.Result[2].Sibsung = SibsungPlus2
				}
			} else { //정
				if a.Chun[2].Properties.Umyang == b.Chun[2].Properties.Umyang { //정
					a.Result[2].Sibsung = SibsungPlus2
					b.Result[2].Sibsung = SibsungPlus2
				} else { //편
					a.Result[2].Sibsung = SibsungPlus2
					b.Result[2].Sibsung = SibsungPlus2
				}
			}
		}
	case "인성":
		switch b.Chun[2].Properties.Sibsung {
		case "비겁":
			if a.Chun[1].Properties.Umyang == a.Chun[2].Properties.Umyang { //편
				if a.Chun[2].Properties.Umyang == b.Chun[2].Properties.Umyang { //편
					b.Result[2].Sibsung = SibsungPlus1
				} else { //정
					b.Result[2].Sibsung = SibsungPlus3
				}
			} else { //정
				if a.Chun[2].Properties.Umyang == b.Chun[2].Properties.Umyang { //정
					b.Result[2].Sibsung = SibsungPlus1
				} else { //편
					b.Result[2].Sibsung = SibsungPlus3
				}
			}
		case "식상":
			if a.Chun[1].Properties.Umyang == a.Chun[2].Properties.Umyang { //편
				if a.Chun[2].Properties.Umyang == b.Chun[2].Properties.Umyang { //편
					b.Result[2].Sibsung = SibsungMinus4
				} else { //정
					if a.Chun[2].Properties.Umyang == 1 {
						a.Result[2].Sibsung = SibsungPlus2
						b.Result[2].Sibsung = SibsungPlus2
					} else {
						b.Result[2].Sibsung = SibsungMinus3
					}
				}
			} else { //정
				if a.Chun[2].Properties.Umyang == b.Chun[2].Properties.Umyang { // 정
					b.Result[2].Sibsung = SibsungMinus4
				} else { //편
					if a.Chun[2].Properties.Umyang != 1 {
						a.Result[2].Sibsung = SibsungPlus2
						b.Result[2].Sibsung = SibsungPlus2
					} else {
						b.Result[2].Sibsung = SibsungMinus3
					}
				}
			}
		case "재성":
			if a.Chun[1].Properties.Umyang == a.Chun[2].Properties.Umyang { //편
				if a.Chun[2].Properties.Umyang == b.Chun[2].Properties.Umyang { //편
					a.Result[2].Sibsung = SibsungMinus2
				} else { //정
					a.Result[2].Sibsung = SibsungMinus2
				}
			} else { //정
				if a.Chun[2].Properties.Umyang == b.Chun[2].Properties.Umyang { //정
					a.Result[2].Sibsung = SibsungMinus2
				} else { //편
					a.Result[2].Sibsung = SibsungMinus2
				}
			}
		case "관성":
			if a.Chun[1].Properties.Umyang == a.Chun[2].Properties.Umyang { //편
				if a.Chun[2].Properties.Umyang == b.Chun[2].Properties.Umyang { //편
					a.Result[2].Sibsung = SibsungPlus2
					b.Result[2].Sibsung = SibsungPlus2
				} else { //정
					a.Result[2].Sibsung = SibsungPlus2
					b.Result[2].Sibsung = SibsungPlus2
				}
			} else { //정
				if a.Chun[2].Properties.Umyang == b.Chun[2].Properties.Umyang { //정
					a.Result[2].Sibsung = SibsungPlus2
					b.Result[2].Sibsung = SibsungPlus2
				} else { //편
					a.Result[2].Sibsung = SibsungPlus2
					b.Result[2].Sibsung = SibsungPlus2
				}
			}
		}
	}
	mutex.Unlock()
	wg.Done()
	//return a, b
}

func (saju *SajuAnalyzer) Find_Unsung_Goonghab(wg *sync.WaitGroup, mutex *sync.Mutex, a *Person, b *Person) {
	f := saju.Sib2Unsung
	mutex.Lock()
	for j := 0; j < 10; j++ {
		if a.Chun[1].Title == f[j].Title {
			for k := 0; k < 12; k++ {
				if b.Ji[1].Title == f[j].Properties[k].Jiji_char {
					a.Chun[1].Properties.Unsung_Me.level = f[j].Properties[k].Level
					a.Chun[1].Properties.Unsung_Me.Unsung_title = f[j].Properties[k].Prop
				}
			}
		}
		if b.Chun[1].Title == f[j].Title {
			for k := 0; k < 12; k++ {
				if a.Ji[1].Title == f[j].Properties[k].Jiji_char {
					b.Chun[1].Properties.Unsung_Me.level = f[j].Properties[k].Level
					b.Chun[1].Properties.Unsung_Me.Unsung_title = f[j].Properties[k].Prop
				}
			}
		}
	}

	for j := 0; j < 10; j++ {
		if a.Chun[2].Title == f[j].Title {
			for k := 0; k < 12; k++ {
				if b.Ji[2].Title == f[j].Properties[k].Jiji_char {
					a.Chun[2].Properties.Unsung_Me.level = f[j].Properties[k].Level
					a.Chun[2].Properties.Unsung_Me.Unsung_title = f[j].Properties[k].Prop
				}
			}
		}
		if b.Chun[2].Title == f[j].Title {
			for k := 0; k < 12; k++ {
				if a.Ji[2].Title == f[j].Properties[k].Jiji_char {
					b.Chun[2].Properties.Unsung_Me.level = f[j].Properties[k].Level
					b.Chun[2].Properties.Unsung_Me.Unsung_title = f[j].Properties[k].Prop
				}
			}
		}
	}
	mutex.Unlock()
	wg.Done()
	//return a, b
}
func Find_Banghab_Goonghab(wg *sync.WaitGroup, mutex *sync.Mutex, a *Person, b *Person) {
	mutex.Lock()
	for i := 0; i < 4; i++ {
		a.Result[i].BangHab = 0
		b.Result[i].BangHab = 0
		//인묘진
		if a.Ji[i].Title == "인" || a.Ji[i].Title == "묘" || a.Ji[i].Title == "진" {
			if b.Ji[i].Title == "인" || b.Ji[i].Title == "묘" || b.Ji[i].Title == "진" {
				if a.Ji[i].Title != b.Ji[i].Title {
					a.Result[i].BangHab = InMyoJin
					b.Result[i].BangHab = InMyoJin
				}
			}
		}

		//사오미
		if a.Ji[i].Title == "사" || a.Ji[i].Title == "오" || a.Ji[i].Title == "미" {
			if b.Ji[i].Title == "사" || b.Ji[i].Title == "오" || b.Ji[i].Title == "미" {
				if a.Ji[i].Title != b.Ji[i].Title {
					a.Result[i].BangHab = SaOMi
					b.Result[i].BangHab = SaOMi
				}
			}
		}

		//신유술
		if a.Ji[i].Title == "신" || a.Ji[i].Title == "유" || a.Ji[i].Title == "술" {
			if b.Ji[i].Title == "신" || b.Ji[i].Title == "유" || b.Ji[i].Title == "술" {
				if a.Ji[i].Title != b.Ji[i].Title {
					a.Result[i].BangHab = SinYuSul
					b.Result[i].BangHab = SinYuSul
				}
			}
		}

		//해자축
		if a.Ji[i].Title == "해" || a.Ji[i].Title == "자" || a.Ji[i].Title == "축" {
			if b.Ji[i].Title == "해" || b.Ji[i].Title == "자" || b.Ji[i].Title == "축" {
				if a.Ji[i].Title != b.Ji[i].Title {
					a.Result[i].BangHab = HaeJaChuk
					b.Result[i].BangHab = HaeJaChuk
				}
			}
		}
	}
	mutex.Unlock()
	wg.Done()
	//return a, b
}

func Find_Samhab_Goonghab(wg *sync.WaitGroup, mutex *sync.Mutex, a *Person, b *Person) {
	mutex.Lock()
	for i := 0; i < 4; i++ {
		a.Result[i].SamHab = 0
		b.Result[i].SamHab = 0
		//신자진
		if a.Ji[i].Title == "신" || a.Ji[i].Title == "자" || a.Ji[i].Title == "진" {
			if b.Ji[i].Title == "신" || b.Ji[i].Title == "자" || b.Ji[i].Title == "진" {
				if a.Ji[i].Title != b.Ji[i].Title {
					a.Result[i].SamHab = SinJaJin
					b.Result[i].SamHab = SinJaJin
				}
			}
		}

		//사유축
		if a.Ji[i].Title == "사" || a.Ji[i].Title == "유" || a.Ji[i].Title == "축" {
			if b.Ji[i].Title == "사" || b.Ji[i].Title == "유" || b.Ji[i].Title == "축" {
				if a.Ji[i].Title != b.Ji[i].Title {
					a.Result[i].SamHab = SaYuChuk
					b.Result[i].SamHab = SaYuChuk
				}
			}
		}

		//해묘미
		if a.Ji[i].Title == "해" || a.Ji[i].Title == "묘" || a.Ji[i].Title == "미" {
			if b.Ji[i].Title == "해" || b.Ji[i].Title == "묘" || b.Ji[i].Title == "미" {
				if a.Ji[i].Title != b.Ji[i].Title {
					a.Result[i].SamHab = HaeMyoMi
					b.Result[i].SamHab = HaeMyoMi
				}
			}
		}

		//인오술
		if a.Ji[i].Title == "인" || a.Ji[i].Title == "오" || a.Ji[i].Title == "술" {
			if b.Ji[i].Title == "인" || b.Ji[i].Title == "오" || b.Ji[i].Title == "술" {
				if a.Ji[i].Title != b.Ji[i].Title {
					a.Result[i].SamHab = InOSul
					b.Result[i].SamHab = InOSul
				}
			}
		}
	}
	mutex.Unlock()
	wg.Done()
	//return a, b
}

func Find_Yukhab_Goonghab(wg *sync.WaitGroup, mutex *sync.Mutex, a *Person, b *Person) {
	mutex.Lock()
	for i := 0; i < 4; i++ {
		a.Result[i].YukHab = 0
		b.Result[i].YukHab = 0
		//축자
		if a.Ji[i].Title == "자" || a.Ji[i].Title == "축" {
			if b.Ji[i].Title == "자" || b.Ji[i].Title == "축" {
				if a.Ji[i].Title != b.Ji[i].Title {

					a.Ji[i].Properties.Prop = "earth"
					a.Result[i].YukHab = JaChuk
					b.Ji[i].Properties.Prop = "earth"
					b.Result[i].YukHab = JaChuk
				}
			}
		}

		//인해
		if a.Ji[i].Title == "인" || a.Ji[i].Title == "해" {
			if b.Ji[i].Title == "인" || b.Ji[i].Title == "해" {
				if a.Ji[i].Title != b.Ji[i].Title {

					a.Ji[i].Properties.Prop = "tree"
					b.Ji[i].Properties.Prop = "tree"
					a.Result[i].YukHab = InHye
					b.Result[i].YukHab = InHye
				}
			}
		}

		//진유
		if a.Ji[i].Title == "진" || a.Ji[i].Title == "유" {
			if b.Ji[i].Title == "진" || b.Ji[i].Title == "유" {
				if a.Ji[i].Title != b.Ji[i].Title {

					a.Ji[i].Properties.Prop = "iron"
					b.Ji[i].Properties.Prop = "iron"
					a.Result[i].YukHab = JinYu
					b.Result[i].YukHab = JinYu
				}
			}
		}

		//사신
		if a.Ji[i].Title == "사" || a.Ji[i].Title == "신" {
			if b.Ji[i].Title == "사" || b.Ji[i].Title == "신" {
				if a.Ji[i].Title != b.Ji[i].Title {

					a.Ji[i].Properties.Prop = "water"
					b.Ji[i].Properties.Prop = "water"
					a.Result[i].YukHab = SaSin
					b.Result[i].YukHab = SaSin
				}
			}
		}

		//오미
		if a.Ji[i].Title == "오" || a.Ji[i].Title == "미" {
			if b.Ji[i].Title == "오" || b.Ji[i].Title == "미" {
				if a.Ji[i].Title != b.Ji[i].Title {

					a.Ji[i].Properties.Prop = "fire"
					b.Ji[i].Properties.Prop = "fire"
					a.Result[i].YukHab = OMi
					b.Result[i].YukHab = OMi
				}
			}
		}
	}
	mutex.Unlock()
	wg.Done()
	//return a, b
}

func Find_Hyungsal_Goonghab(wg *sync.WaitGroup, mutex *sync.Mutex, a *Person, b *Person) {
	mutex.Lock()
	for i := 0; i < 4; i++ {
		a.Result[i].Hyung = 0
		b.Result[i].Hyung = 0
		//인사
		if a.Ji[i].Title == "사" || a.Ji[i].Title == "인" {
			if b.Ji[i].Title == "사" || b.Ji[i].Title == "인" {
				if a.Ji[i].Title != b.Ji[i].Title {
					a.Result[i].Hyung = InSa
					b.Result[i].Hyung = InSa
				}
			}
		}

		//사신
		if a.Ji[i].Title == "사" || a.Ji[i].Title == "신" {
			if b.Ji[i].Title == "사" || b.Ji[i].Title == "신" {
				if a.Ji[i].Title != b.Ji[i].Title {
					a.Result[i].Hyung = SaSin
					b.Result[i].Hyung = SaSin
				}
			}
		}

		//축술
		if a.Ji[i].Title == "술" || a.Ji[i].Title == "축" {
			if b.Ji[i].Title == "술" || b.Ji[i].Title == "축" {
				if a.Ji[i].Title != b.Ji[i].Title {
					a.Result[i].Hyung = ChukSul
					b.Result[i].Hyung = ChukSul
				}
			}
		}

		//술미
		if a.Ji[i].Title == "술" || a.Ji[i].Title == "미" {
			if b.Ji[i].Title == "술" || b.Ji[i].Title == "미" {
				if a.Ji[i].Title != b.Ji[i].Title {
					a.Result[i].Hyung = SulMi
					b.Result[i].Hyung = SulMi
				}
			}
		}

		//자묘
		if a.Ji[i].Title == "자" || a.Ji[i].Title == "묘" {
			if b.Ji[i].Title == "자" || b.Ji[i].Title == "묘" {
				if a.Ji[i].Title != b.Ji[i].Title {
					a.Result[i].Hyung = JaMyo
					b.Result[i].Hyung = JaMyo
				}
			}
		}

		//진진
		if a.Ji[i].Title == "진" {
			if b.Ji[i].Title == "진" {
				a.Result[i].Hyung = JinJin
				b.Result[i].Hyung = JinJin
			}
		}

		//오오
		if a.Ji[i].Title == "오" {
			if b.Ji[i].Title == "오" {
				a.Result[i].Hyung = OO
				b.Result[i].Hyung = OO
			}
		}

		//유유
		if a.Ji[i].Title == "유" {
			if b.Ji[i].Title == "유" {
				a.Result[i].Hyung = YuYu
				b.Result[i].Hyung = YuYu
			}
		}

		//해해
		if a.Ji[i].Title == "해" {
			if b.Ji[i].Title == "해" {
				a.Result[i].Hyung = HaeHae
				b.Result[i].Hyung = HaeHae
			}
		}
	}
	mutex.Unlock()
	wg.Done()
	//return a, b
}

func Find_Choongsal_Goonghab(wg *sync.WaitGroup, mutex *sync.Mutex, a *Person, b *Person) {
	mutex.Lock()
	for i := 0; i < 4; i++ {
		a.Result[i].Choong = 0
		b.Result[i].Choong = 0
		//인신
		if a.Ji[i].Title == "인" || a.Ji[i].Title == "신" {
			if b.Ji[i].Title == "인" || b.Ji[i].Title == "신" {
				if a.Ji[i].Title != b.Ji[i].Title {
					a.Result[i].Choong = InSin
					b.Result[i].Choong = InSin
				}
			}
		}

		//묘유
		if a.Ji[i].Title == "묘" || a.Ji[i].Title == "유" {
			if b.Ji[i].Title == "묘" || b.Ji[i].Title == "유" {
				if a.Ji[i].Title != b.Ji[i].Title {
					a.Result[i].Choong = MyoYu
					b.Result[i].Choong = MyoYu
				}
			}
		}

		//진술
		if a.Ji[i].Title == "진" || a.Ji[i].Title == "술" {
			if b.Ji[i].Title == "진" || b.Ji[i].Title == "술" {
				if a.Ji[i].Title != b.Ji[i].Title {
					a.Result[i].Choong = JinSul
					b.Result[i].Choong = JinSul
				}
			}
		}

		//사해
		if a.Ji[i].Title == "사" || a.Ji[i].Title == "해" {
			if b.Ji[i].Title == "사" || b.Ji[i].Title == "해" {
				if a.Ji[i].Title != b.Ji[i].Title {
					a.Result[i].Choong = SaHae
					b.Result[i].Choong = SaHae
				}
			}
		}

		//자오
		if a.Ji[i].Title == "자" || a.Ji[i].Title == "오" {
			if b.Ji[i].Title == "자" || b.Ji[i].Title == "오" {
				if a.Ji[i].Title != b.Ji[i].Title {
					a.Result[i].Choong = JaO
					b.Result[i].Choong = JaO
				}
			}
		}

		//축미
		if a.Ji[i].Title == "축" || a.Ji[i].Title == "미" {
			if b.Ji[i].Title == "축" || b.Ji[i].Title == "미" {
				if a.Ji[i].Title != b.Ji[i].Title {
					a.Result[i].Choong = ChukMi
					b.Result[i].Choong = ChukMi
				}
			}
		}

	}
	mutex.Unlock()
	wg.Done()
	//return a, b
}

func Find_Pasal_Goonghab(wg *sync.WaitGroup, mutex *sync.Mutex, a *Person, b *Person) {
	mutex.Lock()
	for i := 0; i < 4; i++ {
		a.Result[i].Pa = 0
		b.Result[i].Pa = 0
		//자유 귀문이 우선
		if a.Ji[i].Title == "자" || a.Ji[i].Title == "유" {
			if b.Ji[i].Title == "자" || b.Ji[i].Title == "유" {
				if a.Ji[i].Title != b.Ji[i].Title {
					a.Result[i].Pa = JaYu
					b.Result[i].Pa = JaYu
				}
			}
		}

		//묘오
		if a.Ji[i].Title == "묘" || a.Ji[i].Title == "오" {
			if b.Ji[i].Title == "묘" || b.Ji[i].Title == "오" {
				if a.Ji[i].Title != b.Ji[i].Title {
					a.Result[i].Pa = MyoO
					b.Result[i].Pa = MyoO
				}
			}
		}

		//사신  형합이 우선
		//사신
		if a.Ji[i].Title == "사" || a.Ji[i].Title == "신" {
			if b.Ji[i].Title == "사" || b.Ji[i].Title == "신" {
				if a.Ji[i].Title != b.Ji[i].Title {
					a.Result[i].Pa = SaSin
					b.Result[i].Pa = SaSin
				}
			}
		}

		//축진
		if a.Ji[i].Title == "사" || a.Ji[i].Title == "신" {
			if b.Ji[i].Title == "사" || b.Ji[i].Title == "신" {
				if a.Ji[i].Title != b.Ji[i].Title {
					a.Result[i].Pa = JinChuk
					b.Result[i].Pa = JinChuk
				}
			}
		}

		//술미 형이 우선
		if a.Ji[i].Title == "술" || a.Ji[i].Title == "미" {
			if b.Ji[i].Title == "술" || b.Ji[i].Title == "미" {
				if a.Ji[i].Title != b.Ji[i].Title {
					a.Result[i].Pa = PaSulMi
					b.Result[i].Pa = PaSulMi
				}
			}
		}

		//인해 인중 병 손상
		if a.Ji[i].Title == "인" || a.Ji[i].Title == "해" {
			if b.Ji[i].Title == "인" || b.Ji[i].Title == "해" {
				if a.Ji[i].Title != b.Ji[i].Title {
					a.Result[i].Pa = InHae
					b.Result[i].Pa = InHae
				}
			}
		}
	}
	mutex.Unlock()
	wg.Done()
	//return a, b
}

func Find_Haesal_Goonghab(wg *sync.WaitGroup, mutex *sync.Mutex, a *Person, b *Person) {
	//자미
	mutex.Lock()
	for i := 0; i < 4; i++ {
		a.Result[i].Hae = 0
		b.Result[i].Hae = 0
		if a.Ji[i].Title == "자" || a.Ji[i].Title == "미" {
			if b.Ji[i].Title == "자" || b.Ji[i].Title == "미" {
				if a.Ji[i].Title != b.Ji[i].Title {
					a.Result[i].Hae = JaMi
					b.Result[i].Hae = JaMi
				}
			}
		}

		//축오 탕화중독
		if a.Ji[i].Title == "축" || a.Ji[i].Title == "오" {
			if b.Ji[i].Title == "축" || b.Ji[i].Title == "오" {
				if a.Ji[i].Title != b.Ji[i].Title {
					a.Result[i].Hae = OChuk
					b.Result[i].Hae = OChuk
				}
			}
		}

		//인사 형이우선
		if a.Ji[i].Title == "인" || a.Ji[i].Title == "사" {
			if b.Ji[i].Title == "인" || b.Ji[i].Title == "사" {
				if a.Ji[i].Title != b.Ji[i].Title {
					a.Result[i].Hae = HaeInSa
					b.Result[i].Hae = HaeInSa
				}
			}
		}

		//묘진
		if a.Ji[i].Title == "묘" || a.Ji[i].Title == "진" {
			if b.Ji[i].Title == "묘" || b.Ji[i].Title == "진" {
				if a.Ji[i].Title != b.Ji[i].Title {
					a.Result[i].Hae = MyoJin
					b.Result[i].Hae = MyoJin
				}
			}
		}

		//신해 신이 손상
		if a.Ji[i].Title == "신" || a.Ji[i].Title == "해" {
			if b.Ji[i].Title == "신" || b.Ji[i].Title == "해" {
				if a.Ji[i].Title != b.Ji[i].Title {
					a.Result[i].Hae = HaeSin
					b.Result[i].Hae = HaeSin
				}
			}
		}

		//유술
		if a.Ji[i].Title == "유" || a.Ji[i].Title == "술" {
			if b.Ji[i].Title == "유" || b.Ji[i].Title == "술" {
				if a.Ji[i].Title != b.Ji[i].Title {
					a.Result[i].Hae = YuSul
					b.Result[i].Hae = YuSul
				}
			}
		}
	}
	mutex.Unlock()
	wg.Done()
	//return a, b
}

func Find_Wonzin_Goonghab(wg *sync.WaitGroup, mutex *sync.Mutex, a *Person, b *Person) {
	mutex.Lock()
	for i := 0; i < 4; i++ {
		a.Result[i].WonJin = 0
		b.Result[i].WonJin = 0
		//인유
		if a.Ji[i].Title == "인" || a.Ji[i].Title == "유" {
			if b.Ji[i].Title == "인" || b.Ji[i].Title == "유" {
				if a.Ji[i].Title != b.Ji[i].Title {
					a.Result[i].WonJin = InYu
					b.Result[i].WonJin = InYu
				}
			}
		}

		//묘신 금목상쟁
		if a.Ji[i].Title == "묘" || a.Ji[i].Title == "신" {
			if b.Ji[i].Title == "묘" || b.Ji[i].Title == "신" {
				if a.Ji[i].Title != b.Ji[i].Title {
					a.Result[i].WonJin = MyoSin
					b.Result[i].WonJin = MyoSin
				}
			}
		}

		//진해 입묘형
		if a.Ji[i].Title == "진" || a.Ji[i].Title == "해" {
			if b.Ji[i].Title == "진" || b.Ji[i].Title == "해" {
				if a.Ji[i].Title != b.Ji[i].Title {
					a.Result[i].WonJin = JinHae
					b.Result[i].WonJin = JinHae
				}
			}
		}

		//사술 입묘형
		if a.Ji[i].Title == "사" || a.Ji[i].Title == "술" {
			if b.Ji[i].Title == "사" || b.Ji[i].Title == "술" {
				if a.Ji[i].Title != b.Ji[i].Title {
					a.Result[i].WonJin = SaSul
					b.Result[i].WonJin = SaSul
				}
			}
		}

		//오축 탕화형
		if a.Ji[i].Title == "오" || a.Ji[i].Title == "축" {
			if b.Ji[i].Title == "오" || b.Ji[i].Title == "축" {
				if a.Ji[i].Title != b.Ji[i].Title {
					a.Result[i].WonJin = WonOChuk
					b.Result[i].WonJin = WonOChuk
				}
			}
		}

		//자미탕화원진귀문
		if a.Ji[2].Title == "자" || a.Ji[2].Title == "미" {
			if b.Ji[2].Title == "자" || b.Ji[2].Title == "미" {
				if a.Ji[2].Title != b.Ji[2].Title {
					a.Result[2].WonJin = WonJaMi
					b.Result[2].WonJin = WonJaMi
				}
			}
		}
	}
	mutex.Unlock()
	wg.Done()
	//return a, b
}

func Find_Guimun_Goonghab(wg *sync.WaitGroup, mutex *sync.Mutex, a *Person, b *Person) {
	//인미
	mutex.Lock()
	for i := 0; i < 4; i++ {
		a.Result[i].GuiMoon = 0
		b.Result[i].GuiMoon = 0
		if a.Ji[i].Title == "인" || a.Ji[i].Title == "미" {
			if b.Ji[i].Title == "인" || b.Ji[i].Title == "미" {
				if a.Ji[i].Title != b.Ji[i].Title {
					a.Result[i].GuiMoon = InMi
					b.Result[i].GuiMoon = InMi
				}
			}
		}

		//자유 음주형
		if a.Ji[i].Title == "자" || a.Ji[i].Title == "유" {
			if b.Ji[i].Title == "자" || b.Ji[i].Title == "유" {
				if a.Ji[i].Title != b.Ji[i].Title {
					a.Result[i].GuiMoon = WonJaYu
					b.Result[i].GuiMoon = WonJaYu
				}
			}
		}
	}
	mutex.Unlock()
	wg.Done()
	//return a, b
}

func Find_Gyeokgak_Goonghab(wg *sync.WaitGroup, mutex *sync.Mutex, a *Person, b *Person) {
	//사묘축
	mutex.Lock()
	for i := 0; i < 4; i++ {
		a.Result[i].GyeokGak = 0
		b.Result[i].GyeokGak = 0
		if a.Ji[i].Title == "묘" || a.Ji[i].Title == "사" {
			if b.Ji[i].Title == "묘" || b.Ji[i].Title == "사" {
				if a.Ji[i].Title != b.Ji[i].Title {

					a.Result[i].GyeokGak = MyoSa
					b.Result[i].GyeokGak = MyoSa
				}
			}
		}

		if a.Ji[i].Title == "묘" || a.Ji[i].Title == "축" {
			if b.Ji[i].Title == "묘" || b.Ji[i].Title == "축" {
				if a.Ji[i].Title != b.Ji[i].Title {

					a.Result[i].GyeokGak = MyoChuk
					b.Result[i].GyeokGak = MyoChuk
				}
			}
		}

		//신오진
		if a.Ji[i].Title == "오" || a.Ji[i].Title == "신" {
			if b.Ji[i].Title == "오" || b.Ji[i].Title == "신" {
				if a.Ji[i].Title != b.Ji[i].Title {

					a.Result[i].GyeokGak = MyoChuk
					b.Result[i].GyeokGak = OSin
				}
			}
		}

		if a.Ji[i].Title == "오" || a.Ji[i].Title == "진" {
			if b.Ji[i].Title == "오" || b.Ji[i].Title == "진" {
				if a.Ji[i].Title != b.Ji[i].Title {

					a.Result[i].GyeokGak = MyoChuk
					b.Result[i].GyeokGak = JinO
				}
			}
		}

		//해유미
		if a.Ji[i].Title == "유" || a.Ji[i].Title == "해" {
			if b.Ji[i].Title == "유" || b.Ji[i].Title == "해" {
				if a.Ji[i].Title != b.Ji[i].Title {

					a.Result[i].GyeokGak = MyoChuk
					b.Result[i].GyeokGak = YuHae
				}
			}
		}

		if a.Ji[i].Title == "유" || a.Ji[i].Title == "미" {
			if b.Ji[i].Title == "유" || b.Ji[i].Title == "미" {
				if a.Ji[i].Title != b.Ji[i].Title {

					a.Result[i].GyeokGak = MyoChuk
					b.Result[i].GyeokGak = MiYu
				}
			}
		}

		//인자술
		if a.Ji[i].Title == "자" || a.Ji[i].Title == "인" {
			if b.Ji[i].Title == "자" || b.Ji[i].Title == "인" {
				if a.Ji[i].Title != b.Ji[i].Title {

					a.Result[i].GyeokGak = MyoChuk
					b.Result[i].GyeokGak = InJa
				}
			}
		}

		if a.Ji[i].Title == "자" || a.Ji[i].Title == "술" {
			if b.Ji[i].Title == "자" || b.Ji[i].Title == "술" {
				if a.Ji[i].Title != b.Ji[i].Title {

					a.Result[i].GyeokGak = MyoChuk
					b.Result[i].GyeokGak = SulJa
				}
			}
		}

	}
	mutex.Unlock()
	wg.Done()
	//return a, b
}

func Find_AmHab_Goonghab(wg *sync.WaitGroup, mutex *sync.Mutex, a *Person, b *Person) {
	mutex.Lock()
	for i := 0; i < 4; i++ {
		switch a.Ji[i].Title {
		case "인":
			switch b.Ji[i].Title {
			case "유":

				a.Result[i].AmHab = AmHabValue
				b.Result[i].AmHab = AmHabValue
			case "술":

				a.Result[i].AmHab = AmHabValue
				b.Result[i].AmHab = AmHabValue
			case "축":

				a.Result[i].AmHab = AmHabValue
				b.Result[i].AmHab = AmHabValue
			default:
				a.Result[i].AmHab = 0
				b.Result[i].AmHab = 0
			}
		case "묘":
			switch b.Ji[i].Title {
			case "사":

				a.Result[i].AmHab = AmHabValue
				b.Result[i].AmHab = AmHabValue
			case "신":

				a.Result[i].AmHab = AmHabValue
				b.Result[i].AmHab = AmHabValue
			default:
				a.Result[i].AmHab = 0
				b.Result[i].AmHab = 0
			}
		case "진":
			switch b.Ji[i].Title {
			case "사":

				a.Result[i].AmHab = AmHabValue
				b.Result[i].AmHab = AmHabValue
			default:
				a.Result[i].AmHab = 0
				b.Result[i].AmHab = 0
			}
		case "사":
			switch b.Ji[i].Title {
			case "묘":

				a.Result[i].AmHab = AmHabValue
				b.Result[i].AmHab = AmHabValue
			case "진":

				a.Result[i].AmHab = AmHabValue
				b.Result[i].AmHab = AmHabValue
			case "자":

				a.Result[i].AmHab = AmHabValue
				b.Result[i].AmHab = AmHabValue
			case "축":

				a.Result[i].AmHab = AmHabValue
				b.Result[i].AmHab = AmHabValue
			default:
				a.Result[i].AmHab = 0
				b.Result[i].AmHab = 0
			}
		case "오":
			switch b.Ji[i].Title {
			case "신":

				a.Result[i].AmHab = AmHabValue
				b.Result[i].AmHab = AmHabValue
			case "해":

				a.Result[i].AmHab = AmHabValue
				b.Result[i].AmHab = AmHabValue
			default:
				a.Result[i].AmHab = 0
				b.Result[i].AmHab = 0
			}
		case "미":
			switch b.Ji[i].Title {
			case "신":

				a.Result[i].AmHab = AmHabValue
				b.Result[i].AmHab = AmHabValue
			case "유":

				a.Result[i].AmHab = AmHabValue
				b.Result[i].AmHab = AmHabValue
			default:
				a.Result[i].AmHab = 0
				b.Result[i].AmHab = 0
			}
		case "신":
			switch b.Ji[i].Title {
			case "묘":

				a.Result[i].AmHab = AmHabValue
				b.Result[i].AmHab = AmHabValue
			case "오":

				a.Result[i].AmHab = AmHabValue
				b.Result[i].AmHab = AmHabValue
			case "미":

				a.Result[i].AmHab = AmHabValue
				b.Result[i].AmHab = AmHabValue
			default:
				a.Result[i].AmHab = 0
				b.Result[i].AmHab = 0
			}
		case "유":
			switch b.Ji[i].Title {
			case "인":

				a.Result[i].AmHab = AmHabValue
				b.Result[i].AmHab = AmHabValue
			default:
				a.Result[i].AmHab = 0
				b.Result[i].AmHab = 0
			}
		case "술":
			switch b.Ji[i].Title {
			case "해":

				a.Result[i].AmHab = AmHabValue
				b.Result[i].AmHab = AmHabValue
			default:
				a.Result[i].AmHab = 0
				b.Result[i].AmHab = 0
			}
		case "해":
			switch b.Ji[i].Title {
			case "오":

				a.Result[i].AmHab = AmHabValue
				b.Result[i].AmHab = AmHabValue
			case "미":

				a.Result[i].AmHab = AmHabValue
				b.Result[i].AmHab = AmHabValue
			case "술":

				a.Result[i].AmHab = AmHabValue
				b.Result[i].AmHab = AmHabValue
			default:
				a.Result[i].AmHab = 0
				b.Result[i].AmHab = 0
			}
		case "자":
			switch b.Ji[i].Title {
			case "인":

				a.Result[i].AmHab = AmHabValue
				b.Result[i].AmHab = AmHabValue
			case "사":

				a.Result[i].AmHab = AmHabValue
				b.Result[i].AmHab = AmHabValue
			default:
				a.Result[i].AmHab = 0
				b.Result[i].AmHab = 0
			}
		case "축":
			switch b.Ji[i].Title {
			case "인":

				a.Result[i].AmHab = AmHabValue
				b.Result[i].AmHab = AmHabValue
			case "사":

				a.Result[i].AmHab = AmHabValue
				b.Result[i].AmHab = AmHabValue
			default:
				a.Result[i].AmHab = 0
				b.Result[i].AmHab = 0
			}
		}
	}
	mutex.Unlock()
	wg.Done()
	//return a, b
}

func (sa *SajuAnalyzer) Find_GoongHab(mysaju *SaJuPalJa, mydaesaeun *DaeSaeUn, friendsaju *SaJuPalJa, frienddaesaeun *DaeSaeUn) (*Person, *Person) {
	host_chungan_received := []*string{&mysaju.YearChun, &mysaju.MonthChun, &mysaju.DayChun, &mysaju.TimeChun, &mydaesaeun.DaeUnChun, &mydaesaeun.SaeUnChun, &mysaju.YearJi, &mysaju.MonthJi, &mysaju.DayJi, &mysaju.TimeJi, &mydaesaeun.DaeUnJi, &mydaesaeun.SaeUnJi}
	opponent_chungan_received := []*string{&friendsaju.YearChun, &friendsaju.MonthChun, &friendsaju.DayChun, &friendsaju.TimeChun, &frienddaesaeun.DaeUnChun, &frienddaesaeun.SaeUnChun, &friendsaju.YearJi, &friendsaju.MonthJi, &friendsaju.DayJi, &friendsaju.TimeJi, &frienddaesaeun.DaeUnJi, &frienddaesaeun.SaeUnJi}

	host := sa.person_chungan_input(host_chungan_received)
	opponent := sa.person_chungan_input(opponent_chungan_received)

	var wg sync.WaitGroup
	wg.Add(15)
	var mutex sync.Mutex
	go func() { sa.Find_Sibsung_Goonghab(&wg, &mutex, host, opponent) }()
	go func() { sa.Find_Unsung_Goonghab(&wg, &mutex, host, opponent) }()
	go func() { Find_Chungan_hab(&wg, &mutex, host, opponent) }()
	go func() { Find_Chungan_Geok(&wg, &mutex, host, opponent) }()
	go func() { Find_Banghab_Goonghab(&wg, &mutex, host, opponent) }()
	go func() { Find_Samhab_Goonghab(&wg, &mutex, host, opponent) }()
	go func() { Find_Yukhab_Goonghab(&wg, &mutex, host, opponent) }()
	go func() { Find_Hyungsal_Goonghab(&wg, &mutex, host, opponent) }()
	go func() { Find_Choongsal_Goonghab(&wg, &mutex, host, opponent) }()
	go func() { Find_Pasal_Goonghab(&wg, &mutex, host, opponent) }()
	go func() { Find_Haesal_Goonghab(&wg, &mutex, host, opponent) }()
	go func() { Find_Wonzin_Goonghab(&wg, &mutex, host, opponent) }()
	go func() { Find_Guimun_Goonghab(&wg, &mutex, host, opponent) }()
	go func() { Find_Gyeokgak_Goonghab(&wg, &mutex, host, opponent) }()
	go func() { Find_AmHab_Goonghab(&wg, &mutex, host, opponent) }()
	wg.Wait()
	return host, opponent
}

func (sa *SajuAnalyzer) Evaluate_GoonbHab(host *Person, opponent *Person) (float64, float64, string, string) {
	var host_grade = make([]float64, 4)
	var opponent_grade = make([]float64, 4)
	//var host_description [4]string
	//var opponent_description [4]string
	for i := 0; i < 4; i++ {
		var hostchungan, opponentchungan, hostjiji, opponentjiji float64 = 0.0, 0.0, 0.0, 0.0
		if host.Result[i].ChunGanHab != 0 {
			if host.Result[i].ChunGanHab == GabGi {
				//host_description[i], opponent_description[i] = "생각 또는 가치관이 잘 맞다. ", "생각 또는 가치관이 잘 맞다. "
				hostchungan, opponentchungan = 10, 10
			} else if host.Result[i].ChunGanHab == ElGyeong {
				//host_description[i], opponent_description[i] = "생각 또는 가치관이 잘 맞다. ", "생각 또는 가치관이 잘 맞다. "
				hostchungan, opponentchungan = 10, 10
			} else if host.Result[i].ChunGanHab == ByeongSin {
				//host_description[i], opponent_description[i] = "생각 또는 가치관이 잘 맞다. ", "생각 또는 가치관이 잘 맞다. "
				hostchungan, opponentchungan = 10, 10
			} else if host.Result[i].ChunGanHab == JeongIm {
				//host_description[i], opponent_description[i] = "생각 또는 가치관이 잘 맞다. ", "생각 또는 가치관이 잘 맞다. "
				hostchungan, opponentchungan = 10, 10
			} else if host.Result[i].ChunGanHab == MuGye {
				//host_description[i], opponent_description[i] = "생각 또는 가치관이 잘 맞다. ", "생각 또는 가치관이 잘 맞다. "
				hostchungan, opponentchungan = 10, 10
			}

			if host.Result[i].IpMyo != 0 || opponent.Result[i].IpMyo != 0 {
				/*
					if host.Result[i].IpMyo == WhichJi && opponent.Result[i].IpMyo == WhichJi {
						host_description[i] = "생각 또는 가치관이 잘 맞지만, 스스로가 상대방에게 묻힌다는 기분이다. "
						opponent_description[i] = "생각 또는 가치관이 잘 맞지만, 상대방이 나에게 묻힌다는 기분이다. "
					} else if host.Result[i].IpMyo.WhichJi == 2 && opponent.Result[i].IpMyo.WhichJi == 2 {
						host_description[i] = "생각 또는 가치관이 잘 맞지만, 스스로가 상대방에게 묻힌다는 기분이다. "
						opponent_description[i] = "생각 또는 가치관이 잘 맞지만, 상대방이 나에게 묻힌다는 기분이다. "
					}*/
				hostchungan -= 3
				opponentchungan -= 3
			}
		} else if host.Result[i].ChunGanGeok != 0 || opponent.Result[i].ChunGanGeok != 0 {
			if host.Result[i].ChunGanGeok != 0 {
				//host_description[i] = "생각 또는 가치관이 안 맞으며, 상대방이 나의 생각 또는 가치관을 제어하고자 한다. "
				//opponent_description[i] = "생각 또는 가치관이 안 맞으며, 내가 상대방의 생각 또는 가치관을 제어하고자 한다. "
				hostchungan = 3

				if host.Result[i].IpMyo != 0 || opponent.Result[i].IpMyo != 0 {
					if host.Result[i].IpMyo == WhichMyJi && opponent.Result[i].IpMyo == WhichOppJi {
						//host_description[i] = "생각 또는 가치관이 잘 안 맞지만, 스스로가 상대방을 잘 받아들인다. "
						//opponent_description[i] = "생각 또는 가치관이 잘 맞지만, 상대방이 나를 잘 받아들인다. "
						hostchungan += 3
					} else if host.Result[i].IpMyo == WhichOppJi && opponent.Result[i].IpMyo == WhichMyJi {
						hostchungan += 3
						//host_description[i] = "생각 또는 가치관이 잘 안 맞으며, 스스로가 상대방에게 묻힌다는 기분이다. "
						//opponent_description[i] = "생각 또는 가치관이 잘 안 맞으며, 상대방이 나에게 묻힌다는 기분이다. "
					}
				}

			} else if opponent.Result[i].ChunGanGeok != 0 {
				//host_description[i] = "생각 또는 가치관이 잘 안 맞으며, 내가 상대방의 생각 또는 가치관을 제어하고자 한다. "
				//opponent_description[i] = "생각 또는 가치관이 안 맞으며, 상대방이 나의 생각 또는 가치관을 제어하고자 한다. "
				opponentchungan = 3
				if host.Result[i].IpMyo != 0 || opponent.Result[i].IpMyo != 0 {
					if host.Result[i].IpMyo == WhichOppJi && opponent.Result[i].IpMyo == WhichMyJi {
						//host_description[i] = "생각 또는 가치관이 잘 안 맞지만, 상대방이 나를 잘 받아들인다. "
						//opponent_description[i] = "생각 또는 가치관이 잘 맞지만, 스스로가 상대방을 잘 받아들인다. "
						opponentchungan += 3
					} else if host.Result[i].IpMyo == WhichMyJi && opponent.Result[i].IpMyo == WhichOppJi {
						//host_description[i] = "생각 또는 가치관이 잘 안 맞으며, 상대방이 나에게 묻힌다는 기분이다. "
						//opponent_description[i] = "생각 또는 가치관이 잘 안 맞으며, 스스로가 상대방에게 묻힌다는 기분이다. "
						opponentchungan += 3
					}
				}
			}
		} else {
			hostchungan = 5
			opponentchungan = 5
		}

		if host.Result[i].YukHab != 0 {
			//host_description[i] += "서로가 부부처럼 잘 맞다. "
			//opponent_description[i] += "서로가 부부처럼 잘 맞다. "
			hostjiji += 10
			opponentjiji += 10
		} else if host.Result[i].SamHab != 0 {
			//host_description[i] += "실질적으로 서로가 합심해서 어떤 일을 했을 때 잘 맞다. "
			//opponent_description[i] += "실실적으로 서로가 합심해서 어떤 일을 했을 때 잘 맞다. "

			hostjiji += 9
			opponentjiji += 9
		} else if host.Result[i].BangHab != 0 {
			//host_description[i] += "가족처럼 잘 맞다. "
			//opponent_description[i] += "가족처럼 잘 맞다. "
			hostjiji += 8
			opponentjiji += 8
		} else if host.Result[i].WonJin != 0 {
			if host.Result[i].AmHab != 0 {
				//host_description[i] += "서로가 만나면 애증의 관계가 수 있다. "
				//opponent_description[i] += "서로가 만나면 애증의 관계가 수 있다. "
				hostjiji += 7
				opponentjiji += 7
			} else {
				//host_description[i] += "서로가 만나면 서로를 원망할 수 있다. "
				//opponent_description[i] += "서로가 만나면 서로를 원망할 수 있다. "
				hostjiji += 7
				opponentjiji += 7
			}
		} else if host.Result[i].AmHab != 0 {
			//host_description[i] += "서로가 겉으로 드러나지 않는 속 정이 있을 수 있다. "
			//opponent_description[i] += "서로가 겉으로 드러나지 않는 속 정이 있을 수 있다. "
			hostjiji += 7
			opponentjiji += 7
		} else if host.Result[i].GuiMoon != 0 {
			if host.Result[i].GuiMoon == InMi {
				//host_description[i] += "한쪽이 상대방에게 집착할 수 있다. "
				//opponent_description[i] += "한쪽이 상대방에게 집착할 수 있다. "
				hostjiji += 6
				opponentjiji += 6
			} else if host.Result[i].GuiMoon == JaYu {
				//host_description[i] += "술을 마시고 해프닝이 일어나듯, 인지하지 못할 때 문제가 생길 수 있다. "
				//opponent_description[i] += "술을 마시고 해프닝이 일어나듯, 인지하지 못할 때 문제가 생길 수 있다. "
				hostjiji += 5
				opponentjiji += 5
			}

		} else if host.Result[i].GyeokGak != 0 {
			hostjiji += 4
			opponentjiji += 4
		} else if host.Result[i].Hae != 0 {
			hostjiji += 3
			opponentjiji += 3
		} else if host.Result[i].Hyung != 0 {
			//host_description[i] += "서로가 만나면 어떤 분쟁이자 소소한 다툼거리가 만들어 질 수 있다. "
			//opponent_description[i] += "서로가 만나면 어떤 분쟁이자 소소한 다툼거리가 만들어 질 수 있다. "
			hostjiji += 2
			opponentjiji += 2
		} else if host.Result[i].Pa != 0 {
			if host.Result[i].YukHab != 0 {
				hostjiji += 1
				opponentjiji += 1
			} /*else if host.Result[i].Hyung != 0 {

			} else {
				//host_description[i] += "서로가 만나면 자잘한 관계가 깨어질 수 있다. "
				//opponent_description[i] += "서로가 만나면 자잘한 관계가 깨어질 수 있다. "
				hostjiji += 0.5
				opponentjiji += 0.5
			}*/
		} else {
			hostjiji += 0
			opponentjiji += 0
		}

		//log.Println(strconv.Itoa(hostchungan) + "  " + strconv.Itoa(opponentchungan) + "  " + strconv.Itoa(hostjiji) + "  " + strconv.Itoa(opponentjiji))
		/*
			if strings.TrimSpace(opponent.LoginID) == "maletesta3daf207-cdb8-4" || strings.TrimSpace(opponent.LoginID) == "femaletest0f9533f4-5dd2-4" {
				log.Printf("%d   %d", hostchungan, hostjiji)
				log.Printf("%d   %d", opponentchungan, opponentjiji)
			}*/
		host_grade[i] = float64(((hostchungan * 2) + (hostjiji * 3)) / 5)
		opponent_grade[i] = float64(((opponentchungan * 2) + (opponentjiji * 3)) / 5)

		//log.Printf("%.2f    %.2f", host_grade[i], opponent_grade[i])
		if host.Result[i].Choong != 0 {
			//host_description[i] += "충돌이 일어날 수 있다. "
			//opponent_description[i] += "충돌이 일어날 수 있다. "
			host_grade[i] = 10 - host_grade[i]
			opponent_grade[i] = 10 - opponent_grade[i]
		}
	}

	var final_host_grade float64
	var final_opponent_grade float64
	var final_host_description string
	var final_opponent_description string

	final_host_grade = (host_grade[0] + 3*host_grade[1] + 5*host_grade[2] + host_grade[3]) / 10
	final_opponent_grade = (opponent_grade[0] + 3*opponent_grade[1] + 5*opponent_grade[2] + opponent_grade[3]) / 10
	//log.Printf("%.2f   %.2f", final_host_grade, final_opponent_grade)
	//log.Println("//////////////////////////////////////////////")
	/*
		if host_description[i] != "" {
			switch i {
			case 1:
				final_host_description += "각자의 " + host_description[i]
				final_opponent_description += "각자의 " + opponent_description[i]
			case 2:
				final_host_description += "서로가 살아가는 세상에서 " + host_description[i]
				final_opponent_description += "서로가 살아가는 세상에서 " + opponent_description[i]
			}
		}
	*/
	if final_host_grade > 10 || final_opponent_grade > 10 {
		log.Println(host.LoginID)
		log.Println(opponent.LoginID)
		for i := 0; i < 4; i++ {
			log.Println(i)
			log.Printf("host: %.2f oppo: %.2f", host_grade[i], opponent_grade[i])
			log.Printf("Yukhab: %v", host.Result[i].YukHab)
			log.Printf("BangHab: %v", host.Result[i].BangHab)
			log.Printf("SamHab: %v", host.Result[i].SamHab)
			log.Printf("Choong: %v", host.Result[i].Choong)
			log.Printf("Hyung: %v", host.Result[i].Hyung)
			log.Printf("Pa: %v", host.Result[i].Pa)
			log.Printf("Hae: %v", host.Result[i].Hae)
			log.Printf("GyeokGak: %v", host.Result[i].GyeokGak)
			log.Printf("WonJin: %v", host.Result[i].WonJin)
			log.Printf("GuiMoon: %v", host.Result[i].GuiMoon)
			log.Printf("IpMyo: %v", host.Result[i].IpMyo)
			log.Printf("AmHab: %v", host.Result[i].AmHab)
			log.Printf("ChunGanHab: %v", host.Result[i].ChunGanHab)
		}
		log.Printf("host: %.2f oppo: %.2f", final_host_grade, final_opponent_grade)
		log.Println("---------------------------------------------------------------")
	}

	return math.Round(final_host_grade*100.0) / 100, math.Round(final_opponent_grade*100) / 100, final_host_description, final_opponent_description
}

func NewSaJuAnalyzer(path string) *SajuAnalyzer {
	return newSaJuAnalyzer(path)
}

func newSaJuAnalyzer(path string) *SajuAnalyzer {
	chungan_table := make([]Chungan, 10)
	jiji_table := make([]Jiji, 12)
	manse_table := make([]YearGanji, 100)
	julgys_table := make([]YearJulgy, 74)
	saeun_table := make([]SaeUn, 100)
	sibsung_table := make([]*Sibsung, 5)
	sib2unsung_table := make([]*Chungan_Unsung, 10)

	b, err := ioutil.ReadFile(path + "/config/chungan.json")
	if err != nil {
		log.Println(err)
		return nil
	}
	c, err := ioutil.ReadFile(path + "/config/jiji.json")
	if err != nil {
		log.Println(err)
		return nil
	}

	json.Unmarshal(b, &chungan_table)
	json.Unmarshal(c, &jiji_table)

	for i := 0; i < 100; i++ {
		d, err := ioutil.ReadFile(path + "/config/manse/manse" + strconv.Itoa(1950+i) + ".json")
		if err != nil {
			panic(err)
		}
		json.Unmarshal(d, &manse_table[i])
	}

	log.Println("Manse Done")

	for i := 0; i < 74; i++ {
		f, err := ioutil.ReadFile(path + "/config/julgy/julgy" + strconv.Itoa(1950+i) + ".json")
		if err != nil {
			panic(err)
		}
		json.Unmarshal(f, &julgys_table[i])

	}

	log.Println("julgy Done")

	g, err := ioutil.ReadFile(path + "/config/saeun/saeun.json")
	if err != nil {
		panic(err)
	}
	json.Unmarshal(g, &saeun_table)

	d, err := ioutil.ReadFile(path + "/config/sibsung.json")
	if err != nil {
		log.Println(err)
		return nil
	}

	e, err := ioutil.ReadFile(path + "/config/sib2Unsung.json")
	if err != nil {
		log.Println(err)
		return nil
	}

	json.Unmarshal(d, &sibsung_table)
	json.Unmarshal(e, &sib2unsung_table)

	return &SajuAnalyzer{
		Chungan:    chungan_table,
		Jiji:       jiji_table,
		Manse:      manse_table,
		Julgys:     julgys_table,
		SaeUn:      saeun_table,
		Sibsung:    sibsung_table,
		Sib2Unsung: sib2unsung_table,
	}
}

func Umyang_balance(a *Person, b *Person) {
	count_plus := 0
	count_minus := 0

	for i := 0; i < 5; i++ {
		if a.Chun[i].Properties.Umyang == 1 {
			count_plus++
		} else if a.Chun[i].Properties.Umyang == 0 {
			count_minus++
		}

		if b.Chun[i].Properties.Umyang == 1 {
			count_plus++
		} else if b.Chun[i].Properties.Umyang == 0 {
			count_minus++
		}

		if a.Ji[i].Properties.Umyang == 1 {
			count_plus++
		} else if a.Ji[i].Properties.Umyang == 0 {
			count_minus++
		}

		if b.Ji[i].Properties.Umyang == 1 {
			count_plus++
		} else if b.Ji[i].Properties.Umyang == 0 {
			count_minus++
		}

	}

}

func Ohang_balance(a *Person, b *Person) {
	count_tree := 0
	count_fire := 0
	count_earth := 0
	count_iron := 0
	count_water := 0

	for i := 0; i < 4; i++ {
		if a.Chun[i].Properties.Prop == "tree" {
			count_tree++
		} else if a.Chun[i].Properties.Prop == "fire" {
			count_fire++
		} else if a.Chun[i].Properties.Prop == "earth" {
			count_earth++
		} else if a.Chun[i].Properties.Prop == "iron" {
			count_iron++
		} else if a.Chun[i].Properties.Prop == "water" {
			count_water++
		}

		if b.Chun[i].Properties.Prop == "tree" {
			count_tree++
		} else if b.Chun[i].Properties.Prop == "fire" {
			count_fire++
		} else if b.Chun[i].Properties.Prop == "earth" {
			count_earth++
		} else if b.Chun[i].Properties.Prop == "iron" {
			count_iron++
		} else if b.Chun[i].Properties.Prop == "water" {
			count_water++
		}

		if a.Ji[i].Properties.Prop == "tree" {
			count_tree++
		} else if a.Ji[i].Properties.Prop == "fire" {
			count_fire++
		} else if a.Ji[i].Properties.Prop == "earth" {
			count_earth++
		} else if a.Ji[i].Properties.Prop == "iron" {
			count_iron++
		} else if a.Ji[i].Properties.Prop == "water" {
			count_water++
		}

		if b.Ji[i].Properties.Prop == "tree" {
			count_tree++
		} else if b.Ji[i].Properties.Prop == "fire" {
			count_fire++
		} else if b.Ji[i].Properties.Prop == "earth" {
			count_earth++
		} else if b.Ji[i].Properties.Prop == "iron" {
			count_iron++
		} else if b.Ji[i].Properties.Prop == "water" {
			count_water++
		}
	}
}

func chungan_hab(a *Person) {
	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			a.Result[i].ChunGanHab = 0
			a.Result[j].ChunGanHab = 0
			if math.Abs(float64(a.Chun[i].Id-a.Chun[j].Id)) == 5 {
				switch {
				case a.Chun[i].Title == "갑" || a.Chun[i].Title == "기":
					a.Result[i].ChunGanHab = GabGi
					a.Result[j].ChunGanHab = GabGi
				case a.Chun[i].Title == "을" || a.Chun[i].Title == "경":
					a.Result[i].ChunGanHab = ElGyeong
					a.Result[j].ChunGanHab = ElGyeong
				case a.Chun[i].Title == "병" || a.Chun[i].Title == "신":
					a.Result[i].ChunGanHab = ByeongSin
					a.Result[j].ChunGanHab = ByeongSin
				case a.Chun[i].Title == "정" || a.Chun[i].Title == "임":
					a.Result[i].ChunGanHab = JeongIm
					a.Result[j].ChunGanHab = JeongIm
				case a.Chun[i].Title == "무" || a.Chun[i].Title == "계":
					a.Result[i].ChunGanHab = MuGye
					a.Result[j].ChunGanHab = MuGye

				}
			}
		}
	}
}

func chungan_geok(a *Person) *Person {
	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			a.Result[i].ChunGanGeok = 0
			a.Result[j].ChunGanGeok = 0
			if i != j {
				if math.Abs(float64(a.Chun[i].Id-a.Chun[j].Id)) == 6 {
					switch {
					case a.Chun[i].Title == "갑":
						a.Result[i].ChunGanGeok = -1 * GabGyeong
						a.Result[j].ChunGanGeok = GabGyeong
					case a.Chun[i].Title == "을":
						a.Result[i].ChunGanGeok = -1 * ElSin
						a.Result[j].ChunGanGeok = ElSin
					case a.Chun[i].Title == "병":
						a.Result[i].ChunGanGeok = -1 * ByeongIm
						a.Result[j].ChunGanGeok = ByeongIm
					case a.Chun[i].Title == "정":
						a.Result[i].ChunGanGeok = -1 * JeongHye
						a.Result[j].ChunGanGeok = JeongHye
					case a.Chun[i].Title == "무":
						a.Result[i].ChunGanGeok = -1 * MuGab
						a.Result[j].ChunGanGeok = MuGab
					case a.Chun[i].Title == "기":
						a.Result[i].ChunGanGeok = -1 * GiEl
						a.Result[j].ChunGanGeok = GiEl
					case a.Chun[i].Title == "경":
						a.Result[i].ChunGanGeok = -1 * GyeongByeong
						a.Result[j].ChunGanGeok = GyeongByeong
					case a.Chun[i].Title == "신":
						a.Result[i].ChunGanGeok = -1 * SinJeong
						a.Result[j].ChunGanGeok = SinJeong
					case a.Chun[i].Title == "임":
						a.Result[i].ChunGanGeok = -1 * ImMu
						a.Result[j].ChunGanGeok = ImMu
					case a.Chun[i].Title == "계":
						a.Result[i].ChunGanGeok = -1 * GyeGi
						a.Result[j].ChunGanGeok = GyeGi
					}
				}
			}
		}
	}
	return a
}

func Find_Ipmyo(a *Person, b *Person) (*Person, *Person) {
	for i := 1; i < 3; i++ {
		a.Result[i].IpMyo = 0
		b.Result[i].IpMyo = 0

		switch a.Ji[i].Title {
		case "진":
			if a.Chun[i].Title == "신" || a.Chun[i].Title == "임" {
				a.Result[i].IpMyo = WhichMyJi
			} else if b.Chun[i].Title == "신" || b.Chun[i].Title == "임" {
				b.Result[i].IpMyo = WhichOppJi
			}
		case "술":
			if a.Chun[i].Title == "병" || a.Chun[i].Title == "무" || a.Chun[i].Title == "을" {
				a.Result[i].IpMyo = WhichMyJi
			} else if b.Chun[i].Title == "병" || b.Chun[i].Title == "무" || b.Chun[i].Title == "을" {
				b.Result[i].IpMyo = WhichOppJi
			}
		case "축":
			if a.Chun[i].Title == "경" || a.Chun[i].Title == "기" || a.Chun[i].Title == "정" {
				a.Result[i].IpMyo = WhichMyJi
			} else if b.Chun[i].Title == "경" || b.Chun[i].Title == "기" || b.Chun[i].Title == "정" {
				b.Result[i].IpMyo = WhichOppJi
			}
		case "미":
			if a.Chun[i].Title == "갑" || a.Chun[i].Title == "해" {
				a.Result[i].IpMyo = WhichMyJi
			} else if b.Chun[i].Title == "갑" || b.Chun[i].Title == "해" {
				b.Result[i].IpMyo = WhichOppJi
			}
		}

		switch b.Ji[i].Title {
		case "진":
			if a.Chun[i].Title == "신" || a.Chun[i].Title == "임" {
				a.Result[i].IpMyo = 1310
			} else if b.Chun[i].Title == "신" || b.Chun[i].Title == "임" {
				b.Result[i].IpMyo = WhichMyJi
			}
		case "술":

			if a.Chun[i].Title == "병" || a.Chun[i].Title == "무" || a.Chun[i].Title == "을" {
				a.Result[i].IpMyo = WhichOppJi
			} else if b.Chun[i].Title == "병" || b.Chun[i].Title == "무" || b.Chun[i].Title == "을" {
				b.Result[i].IpMyo = WhichMyJi
			}

		case "축":
			if a.Chun[i].Title == "경" || a.Chun[i].Title == "기" || a.Chun[i].Title == "정" {
				a.Result[i].IpMyo = WhichOppJi
			} else if b.Chun[i].Title == "경" || b.Chun[i].Title == "기" || b.Chun[i].Title == "정" {
				b.Result[i].IpMyo = WhichMyJi
			}
		case "미":
			if a.Chun[i].Title == "갑" || a.Chun[i].Title == "해" {
				a.Result[i].IpMyo = WhichOppJi
			} else if b.Chun[i].Title == "갑" || b.Chun[i].Title == "해" {
				b.Result[i].IpMyo = WhichMyJi
			}
		}
	}
	return a, b
}

func Find_Banghab(a *Person) {
	num1 := 5
	num2 := 5
	num3 := 5
	//인묘진
	for i := 0; i < 4; i++ {
		if a.Ji[i].Title == "인" {
			num1 = i
		}
	}
	for j := 0; j < 4; j++ {
		if a.Ji[j].Title == "묘" {
			num2 = j
		}
	}
	for k := 0; k < 4; k++ {
		if a.Ji[k].Title == "진" {
			num3 = k
		}
	}
	a.Result[num3].BangHab = 0
	a.Result[num2].BangHab = 0
	a.Result[num1].BangHab = 0
	a.Result[4].BangHab = 0
	a.Result[5].BangHab = 0
	if num1 != 5 || num2 != 5 || num3 != 5 {
		switch {
		case num1 == 5 && num2 == 5 && num3 == 5:
			{
				a.Result[num1].BangHab = InMyoJin
				a.Result[num2].BangHab = InMyoJin
				a.Result[num3].BangHab = InMyoJin
			}
		case num1 == 5:
			if num2 == 5 {
				if a.Ji[4].Title == "인" && a.Ji[5].Title == "묘" {
					a.Result[4].BangHab = InMyoJin
					a.Result[5].BangHab = InMyoJin
					a.Result[num3].BangHab = InMyoJin
				} else if a.Ji[4].Title == "묘" && a.Ji[5].Title == "인" {
					a.Result[4].BangHab = InMyoJin
					a.Result[5].BangHab = InMyoJin
					a.Result[num3].BangHab = InMyoJin
				}

			} else if num3 == 5 {
				if a.Ji[4].Title == "인" && a.Ji[5].Title == "진" {
					a.Result[4].BangHab = InMyoJin
					a.Result[5].BangHab = InMyoJin
					a.Result[num2].BangHab = InMyoJin
				} else if a.Ji[4].Title == "진" && a.Ji[5].Title == "인" {
					a.Result[4].BangHab = InMyoJin
					a.Result[5].BangHab = InMyoJin
					a.Result[num2].BangHab = InMyoJin
				}

			} else if num2 != 5 && num3 != 5 {
				if a.Ji[4].Title == "인" {
					a.Result[4].BangHab = InMyoJin
					a.Result[num2].BangHab = InMyoJin
					a.Result[num3].BangHab = InMyoJin
				}
				if a.Ji[5].Title == "인" {
					a.Result[5].BangHab = InMyoJin
					a.Result[num2].BangHab = InMyoJin
					a.Result[num3].BangHab = InMyoJin
				}

			}
			fallthrough

		case num2 == 5:
			if num3 == 5 {
				if a.Ji[4].Title == "묘" && a.Ji[5].Title == "진" {
					a.Result[4].BangHab = InMyoJin
					a.Result[5].BangHab = InMyoJin
					a.Result[num1].BangHab = InMyoJin
				} else if a.Ji[4].Title == "진" && a.Ji[5].Title == "묘" {
					a.Result[4].BangHab = InMyoJin
					a.Result[5].BangHab = InMyoJin
					a.Result[num1].BangHab = InMyoJin
				}

			} else if num1 != 5 && num3 != 5 {
				if a.Ji[4].Title == "묘" {
					a.Result[4].BangHab = InMyoJin
					a.Result[num1].BangHab = InMyoJin
					a.Result[num3].BangHab = InMyoJin
				} else if a.Ji[5].Title == "묘" {
					a.Result[5].BangHab = InMyoJin
					a.Result[num1].BangHab = InMyoJin
					a.Result[num3].BangHab = InMyoJin
				}

			}
			fallthrough

		case num3 == 5:
			if num1 != 5 && num2 != 5 {
				if a.Ji[4].Title == "진" {
					a.Result[4].BangHab = InMyoJin
					a.Result[num1].BangHab = InMyoJin
					a.Result[num2].BangHab = InMyoJin
				} else if a.Ji[5].Title == "진" {
					a.Result[5].BangHab = InMyoJin
					a.Result[num1].BangHab = InMyoJin
					a.Result[num2].BangHab = InMyoJin
				}

			}

		}
	}

	//사오미
	num1 = 5
	num2 = 5
	num3 = 5
	for i := 0; i < 4; i++ {
		if a.Ji[i].Title == "사" {
			num1 = i
		}
	}
	for j := 0; j < 4; j++ {
		if a.Ji[j].Title == "오" {
			num2 = j
		}
	}
	for k := 0; k < 4; k++ {
		if a.Ji[k].Title == "미" {
			num3 = k
		}
	}

	if num1 != 5 || num2 != 5 || num3 != 5 {
		switch {
		case num1 == 5 && num2 == 5 && num3 == 5:
			{
				a.Result[num1].BangHab = SaOMi
				a.Result[num2].BangHab = SaOMi
				a.Result[num3].BangHab = SaOMi
			}
		case num1 == 5:
			if num2 == 5 {
				if a.Ji[4].Title == "사" && a.Ji[5].Title == "오" {
					a.Result[4].BangHab = SaOMi
					a.Result[5].BangHab = SaOMi
					a.Result[num3].BangHab = SaOMi
				} else if a.Ji[4].Title == "오" && a.Ji[5].Title == "사" {
					a.Result[4].BangHab = SaOMi
					a.Result[5].BangHab = SaOMi
					a.Result[num3].BangHab = SaOMi
				}
			} else if num3 == 5 {
				if a.Ji[4].Title == "사" && a.Ji[5].Title == "미" {
					a.Result[4].BangHab = SaOMi
					a.Result[5].BangHab = SaOMi
					a.Result[num2].BangHab = SaOMi
				} else if a.Ji[4].Title == "미" && a.Ji[5].Title == "사" {
					a.Result[4].BangHab = SaOMi
					a.Result[5].BangHab = SaOMi
					a.Result[num2].BangHab = SaOMi
				}
			} else if num2 != 5 && num3 != 5 {
				if a.Ji[4].Title == "사" {
					a.Result[4].BangHab = SaOMi
					a.Result[num2].BangHab = SaOMi
					a.Result[num3].BangHab = SaOMi
				}
				if a.Ji[5].Title == "사" {
					a.Result[5].BangHab = SaOMi
					a.Result[num2].BangHab = SaOMi
					a.Result[num3].BangHab = SaOMi
				}
			}
			fallthrough

		case num2 == 5:
			if num3 == 5 {
				if a.Ji[4].Title == "오" && a.Ji[5].Title == "미" {
					a.Result[4].BangHab = SaOMi
					a.Result[5].BangHab = SaOMi
					a.Result[num1].BangHab = SaOMi
				} else if a.Ji[4].Title == "미" && a.Ji[5].Title == "오" {
					a.Result[4].BangHab = SaOMi
					a.Result[5].BangHab = SaOMi
					a.Result[num1].BangHab = SaOMi
				}
			} else if num1 != 5 && num3 != 5 {
				if a.Ji[4].Title == "오" {
					a.Result[4].BangHab = SaOMi
					a.Result[num1].BangHab = SaOMi
					a.Result[num3].BangHab = SaOMi
				} else if a.Ji[5].Title == "오" {
					a.Result[5].BangHab = SaOMi
					a.Result[num1].BangHab = SaOMi
					a.Result[num3].BangHab = SaOMi
				}
			}
			fallthrough

		case num3 == 5:
			if num1 != 5 && num2 != 5 {
				if a.Ji[4].Title == "미" {
					a.Result[4].BangHab = SaOMi
					a.Result[num1].BangHab = SaOMi
					a.Result[num2].BangHab = SaOMi
				} else if a.Ji[5].Title == "미" {
					a.Result[5].BangHab = SaOMi
					a.Result[num1].BangHab = SaOMi
					a.Result[num2].BangHab = SaOMi
				}
			}

		}
	}

	//신유술
	num1 = 5
	num2 = 5
	num3 = 5
	for i := 0; i < 4; i++ {
		if a.Ji[i].Title == "신" {
			num1 = i
		}
	}
	for j := 0; j < 4; j++ {
		if a.Ji[j].Title == "유" {
			num2 = j
		}
	}
	for k := 0; k < 4; k++ {
		if a.Ji[k].Title == "술" {
			num3 = k
		}
	}

	if num1 != 5 || num2 != 5 || num3 != 5 {
		switch {
		case num1 == 5 && num2 == 5 && num3 == 5:
			{
				a.Result[num1].BangHab = SaOMi
				a.Result[num2].BangHab = SaOMi
				a.Result[num3].BangHab = SaOMi
			}
		case num1 == 5:
			if num2 == 5 {
				if a.Ji[4].Title == "신" && a.Ji[5].Title == "유" {
					a.Result[4].BangHab = SinYuSul
					a.Result[5].BangHab = SinYuSul
					a.Result[num3].BangHab = SinYuSul
				} else if a.Ji[4].Title == "유" && a.Ji[5].Title == "신" {
					a.Result[4].BangHab = SinYuSul
					a.Result[5].BangHab = SinYuSul
					a.Result[num3].BangHab = SinYuSul
				}
			} else if num3 == 5 {
				if a.Ji[4].Title == "신" && a.Ji[5].Title == "술" {
					a.Result[4].BangHab = SinYuSul
					a.Result[5].BangHab = SinYuSul
					a.Result[num2].BangHab = SinYuSul
				} else if a.Ji[4].Title == "술" && a.Ji[5].Title == "신" {
					a.Result[4].BangHab = SinYuSul
					a.Result[5].BangHab = SinYuSul
					a.Result[num2].BangHab = SinYuSul
				}
			} else if num2 != 5 && num3 != 5 {
				if a.Ji[4].Title == "신" {
					a.Result[4].BangHab = SinYuSul
					a.Result[num2].BangHab = SinYuSul
					a.Result[num3].BangHab = SinYuSul
				}
				if a.Ji[5].Title == "신" {
					a.Result[5].BangHab = SinYuSul
					a.Result[num2].BangHab = SinYuSul
					a.Result[num3].BangHab = SinYuSul
				}
			}
			fallthrough

		case num2 == 5:
			if num3 == 5 {
				if a.Ji[4].Title == "유" && a.Ji[5].Title == "술" {
					a.Result[4].BangHab = SinYuSul
					a.Result[5].BangHab = SinYuSul
					a.Result[num1].BangHab = SinYuSul
				} else if a.Ji[4].Title == "술" && a.Ji[5].Title == "유" {
					a.Result[4].BangHab = SinYuSul
					a.Result[5].BangHab = SinYuSul
					a.Result[num1].BangHab = SinYuSul
				}
			} else if num1 != 5 && num3 != 5 {
				if a.Ji[4].Title == "유" {
					a.Result[4].BangHab = SinYuSul
					a.Result[num1].BangHab = SinYuSul
					a.Result[num3].BangHab = SinYuSul
				} else if a.Ji[5].Title == "유" {
					a.Result[5].BangHab = SinYuSul
					a.Result[num1].BangHab = SinYuSul
					a.Result[num3].BangHab = SinYuSul
				}
			}
			fallthrough

		case num3 == 5:
			if num1 != 5 && num2 != 5 {
				if a.Ji[4].Title == "술" {
					a.Result[4].BangHab = SinYuSul
					a.Result[num1].BangHab = SinYuSul
					a.Result[num2].BangHab = SinYuSul
				} else if a.Ji[5].Title == "술" {
					a.Result[5].BangHab = SinYuSul
					a.Result[num1].BangHab = SinYuSul
					a.Result[num2].BangHab = SinYuSul
				}
			}
		}

		//해자축
		num1 = 5
		num2 = 5
		num3 = 5
		for i := 0; i < 4; i++ {
			if a.Ji[i].Title == "해" {
				num1 = i
			}
		}
		for j := 0; j < 4; j++ {
			if a.Ji[j].Title == "자" {
				num2 = j
			}
		}
		for k := 0; k < 4; k++ {
			if a.Ji[k].Title == "축" {
				num3 = k
			}
		}

		if num1 != 5 || num2 != 5 || num3 != 5 {
			switch {
			case num1 == 5 && num2 == 5 && num3 == 5:
				{
					a.Result[num1].BangHab = SaOMi
					a.Result[num2].BangHab = SaOMi
					a.Result[num3].BangHab = SaOMi
				}
			case num1 == 5:
				if num2 == 5 {
					if a.Ji[4].Title == "해" && a.Ji[5].Title == "자" {
						a.Result[4].BangHab = HaeJaChuk
						a.Result[5].BangHab = HaeJaChuk
						a.Result[num3].BangHab = HaeJaChuk
					} else if a.Ji[4].Title == "자" && a.Ji[5].Title == "해" {
						a.Result[4].BangHab = HaeJaChuk
						a.Result[5].BangHab = HaeJaChuk
						a.Result[num3].BangHab = HaeJaChuk
					}
				} else if num3 == 5 {
					if a.Ji[4].Title == "해" && a.Ji[5].Title == "축" {
						a.Result[4].BangHab = HaeJaChuk
						a.Result[5].BangHab = HaeJaChuk
						a.Result[num2].BangHab = HaeJaChuk
					} else if a.Ji[4].Title == "축" && a.Ji[5].Title == "해" {
						a.Result[4].BangHab = HaeJaChuk
						a.Result[5].BangHab = HaeJaChuk
						a.Result[num2].BangHab = HaeJaChuk
					}
				} else if num2 != 5 && num3 != 5 {
					if a.Ji[4].Title == "해" {
						a.Result[4].BangHab = HaeJaChuk
						a.Result[num2].BangHab = HaeJaChuk
						a.Result[num3].BangHab = HaeJaChuk
					}
					if a.Ji[5].Title == "해" {
						a.Result[5].BangHab = HaeJaChuk
						a.Result[num2].BangHab = HaeJaChuk
						a.Result[num3].BangHab = HaeJaChuk
					}
				}
				fallthrough

			case num2 == 5:
				if num3 == 5 {
					if a.Ji[4].Title == "자" && a.Ji[5].Title == "축" {
						a.Result[4].BangHab = HaeJaChuk
						a.Result[5].BangHab = HaeJaChuk
						a.Result[num1].BangHab = HaeJaChuk
					} else if a.Ji[4].Title == "축" && a.Ji[5].Title == "자" {

						a.Result[4].BangHab = HaeJaChuk
						a.Result[5].BangHab = HaeJaChuk
						a.Result[num1].BangHab = HaeJaChuk
					}
				} else if num1 != 5 && num3 != 5 {
					if a.Ji[4].Title == "자" {
						a.Result[4].BangHab = HaeJaChuk
						a.Result[num1].BangHab = HaeJaChuk
						a.Result[num3].BangHab = HaeJaChuk
					} else if a.Ji[5].Title == "자" {
						a.Result[5].BangHab = HaeJaChuk
						a.Result[num1].BangHab = HaeJaChuk
						a.Result[num3].BangHab = HaeJaChuk
					}
				}
				fallthrough

			case num3 == 5:
				if num1 != 5 && num2 != 5 {
					if a.Ji[4].Title == "축" {
						a.Result[4].BangHab = HaeJaChuk
						a.Result[num1].BangHab = HaeJaChuk
						a.Result[num2].BangHab = HaeJaChuk
					} else if a.Ji[5].Title == "축" {
						a.Result[5].BangHab = HaeJaChuk
						a.Result[num1].BangHab = HaeJaChuk
						a.Result[num2].BangHab = HaeJaChuk
					}
				}
			}
		}
	}
}

func Find_Samhab(a *Person) {
	num1 := 5
	num2 := 5
	num3 := 5
	//신자진
	for i := 0; i < 4; i++ {
		if a.Ji[i].Title == "신" {
			num1 = i
		}
	}
	for j := 0; j < 4; j++ {
		if a.Ji[j].Title == "자" {
			num2 = j
		}
	}
	for k := 0; k < 4; k++ {
		if a.Ji[k].Title == "진" {
			num3 = k
		}
	}

	a.Result[num1].SamHab = 0
	a.Result[num2].SamHab = 0
	a.Result[num3].SamHab = 0
	a.Result[4].SamHab = 0
	a.Result[5].SamHab = 0

	if num1 != 5 || num2 != 5 || num3 != 5 {
		switch {
		case num1 == 5 && num2 == 5 && num3 == 5:
			{
				a.Ji[num1].Properties.Hae += 1
				a.Ji[num2].Properties.Go += 1
				a.Ji[num3].Properties.Ji += 1
				a.Result[num1].SamHab = SinJaJin
				a.Result[num2].SamHab = SinJaJin
				a.Result[num3].SamHab = SinJaJin
			}
		case num1 == 5:
			if num2 == 5 {
				if a.Ji[4].Title == "신" && a.Ji[5].Title == "자" {

					a.Result[4].SamHab = SinJaJin
					a.Result[5].SamHab = SinJaJin
					a.Ji[num3].Properties.Ji += 1
					a.Result[num3].SamHab = SinJaJin
				} else if a.Ji[4].Title == "자" && a.Ji[5].Title == "신" {
					a.Result[4].SamHab = SinJaJin
					a.Result[5].SamHab = SinJaJin
					a.Ji[num3].Properties.Ji += 1
					a.Result[num3].SamHab = SinJaJin
				}

			} else if num3 == 5 {
				if a.Ji[4].Title == "신" && a.Ji[5].Title == "진" {
					a.Result[4].SamHab = SinJaJin
					a.Result[5].SamHab = SinJaJin
					a.Ji[num2].Properties.Go += 1
					a.Result[num2].SamHab = SinJaJin
				} else if a.Ji[4].Title == "진" && a.Ji[5].Title == "신" {
					a.Result[4].SamHab = SinJaJin
					a.Result[5].SamHab = SinJaJin
					a.Ji[num2].Properties.Go += 1
					a.Result[num2].SamHab = SinJaJin
				}

			} else if num2 != 5 && num3 != 5 {
				if a.Ji[4].Title == "신" {
					a.Result[4].SamHab = SinJaJin
					a.Ji[num2].Properties.Go += 1
					a.Ji[num3].Properties.Ji += 1
					a.Result[num2].SamHab = SinJaJin
					a.Result[num3].SamHab = SinJaJin
				} else {
					a.Result[5].SamHab = SinJaJin
					a.Ji[num2].Properties.Go += 1
					a.Ji[num3].Properties.Ji += 1
					a.Result[num2].SamHab = SinJaJin
					a.Result[num3].SamHab = SinJaJin
				}
			}
			fallthrough

		case num2 == 5:
			if num3 == 5 {
				if a.Ji[4].Title == "자" && a.Ji[5].Title == "진" {
					a.Result[4].SamHab = SinJaJin
					a.Result[5].SamHab = SinJaJin
					a.Ji[num1].Properties.Hae += 1
					a.Result[num1].SamHab = SinJaJin
				} else if a.Ji[4].Title == "진" && a.Ji[5].Title == "자" {
					a.Result[4].SamHab = SinJaJin
					a.Result[5].SamHab = SinJaJin
					a.Ji[num1].Properties.Hae += 1
					a.Result[num1].SamHab = SinJaJin
				}
			} else if num1 != 5 && num3 != 5 {
				if a.Ji[4].Title == "자" {
					a.Result[4].SamHab = SinJaJin
					a.Ji[num1].Properties.Hae += 1
					a.Ji[num3].Properties.Ji += 1
					a.Result[num1].SamHab = SinJaJin
					a.Result[num3].SamHab = SinJaJin
				} else if a.Ji[5].Title == "자" {
					a.Result[5].SamHab = SinJaJin
					a.Ji[num1].Properties.Hae += 1
					a.Ji[num3].Properties.Ji += 1
					a.Result[num1].SamHab = SinJaJin
					a.Result[num3].SamHab = SinJaJin
				}
			}
			fallthrough

		case num3 == 5:
			if num1 != 5 && num2 != 5 {
				if a.Ji[4].Title == "진" {
					a.Result[4].SamHab = SinJaJin
					a.Ji[num1].Properties.Hae += 1
					a.Ji[num2].Properties.Go += 1
					a.Result[num1].SamHab = SinJaJin
					a.Result[num2].SamHab = SinJaJin
				} else if a.Ji[5].Title == "진" {
					a.Result[5].SamHab = SinJaJin
					a.Ji[num1].Properties.Hae += 1
					a.Ji[num2].Properties.Go += 1
					a.Result[num1].SamHab = SinJaJin
					a.Result[num2].SamHab = SinJaJin
				}
			}

		}
	}

	//사유축
	num1 = 5
	num2 = 5
	num3 = 5

	for i := 0; i < 4; i++ {
		if a.Ji[i].Title == "사" {
			num1 = i
		}
	}
	for j := 0; j < 4; j++ {
		if a.Ji[j].Title == "유" {
			num2 = j
		}
	}
	for k := 0; k < 4; k++ {
		if a.Ji[k].Title == "축" {
			num3 = k
		}
	}

	if num1 != 5 || num2 != 5 || num3 != 5 {
		switch {
		case num1 == 5 && num2 == 5 && num3 == 5:
			{
				a.Ji[num1].Properties.Hae += 1
				a.Ji[num2].Properties.Go += 1
				a.Ji[num3].Properties.Ji += 1
				a.Result[num1].SamHab = SaYuChuk
				a.Result[num2].SamHab = SaYuChuk
				a.Result[num3].SamHab = SaYuChuk
			}
		case num1 == 5:
			if num2 == 5 {
				if a.Ji[4].Title == "사" && a.Ji[5].Title == "유" {
					a.Result[4].SamHab = SaYuChuk
					a.Result[5].SamHab = SaYuChuk
					a.Ji[num3].Properties.Ji += 1
					a.Result[num3].SamHab = SaYuChuk
				} else if a.Ji[4].Title == "유" && a.Ji[5].Title == "사" {
					a.Result[4].SamHab = SaYuChuk
					a.Result[5].SamHab = SaYuChuk
					a.Ji[num3].Properties.Ji += 1
					a.Result[num3].SamHab = SaYuChuk
				}
			} else if num3 == 5 {
				if a.Ji[4].Title == "사" && a.Ji[5].Title == "축" {
					a.Result[4].SamHab = SaYuChuk
					a.Result[5].SamHab = SaYuChuk
					a.Ji[num2].Properties.Go += 1
					a.Result[num2].SamHab = SaYuChuk
				} else if a.Ji[4].Title == "축" && a.Ji[5].Title == "사" {
					a.Result[4].SamHab = SaYuChuk
					a.Result[5].SamHab = SaYuChuk
					a.Ji[num2].Properties.Go += 1
					a.Result[num2].SamHab = SaYuChuk
				}
			} else if num2 != 5 && num3 != 5 {
				if a.Ji[4].Title == "사" || a.Ji[5].Title == "사" {
					a.Result[4].SamHab = SaYuChuk
					a.Result[5].SamHab = SaYuChuk
					a.Ji[num2].Properties.Go += 1
					a.Ji[num3].Properties.Ji += 1
					a.Result[num2].SamHab = SaYuChuk
					a.Result[num3].SamHab = SaYuChuk
				}
			}
			fallthrough

		case num2 == 5:
			if num3 == 5 {
				if a.Ji[4].Title == "유" && a.Ji[5].Title == "축" {
					a.Result[4].SamHab = SaYuChuk
					a.Result[5].SamHab = SaYuChuk
					a.Ji[num1].Properties.Hae += 1
					a.Result[num1].SamHab = SaYuChuk
				} else if a.Ji[4].Title == "축" && a.Ji[5].Title == "유" {
					a.Result[4].SamHab = SaYuChuk
					a.Result[5].SamHab = SaYuChuk
					a.Ji[num1].Properties.Hae += 1
					a.Result[num1].SamHab = SaYuChuk
				}
			} else if num1 != 5 && num3 != 5 {
				if a.Ji[4].Title == "유" {
					a.Result[4].SamHab = SaYuChuk
					a.Ji[num1].Properties.Hae += 1
					a.Ji[num3].Properties.Ji += 1
					a.Result[num1].SamHab = SaYuChuk
					a.Result[num3].SamHab = SaYuChuk
				} else if a.Ji[5].Title == "유" {
					a.Result[5].SamHab = SaYuChuk
					a.Ji[num1].Properties.Hae += 1
					a.Ji[num3].Properties.Ji += 1
					a.Result[num1].SamHab = SaYuChuk
					a.Result[num3].SamHab = SaYuChuk
				}
			}
			fallthrough

		case num3 == 5:
			if num1 != 5 && num2 != 5 {
				if a.Ji[4].Title == "축" {
					a.Result[4].SamHab = SaYuChuk
					a.Ji[num1].Properties.Hae += 1
					a.Ji[num2].Properties.Go += 1
					a.Result[num1].SamHab = SaYuChuk
					a.Result[num2].SamHab = SaYuChuk
				} else if a.Ji[5].Title == "축" {
					a.Result[5].SamHab = SaYuChuk
					a.Ji[num1].Properties.Hae += 1
					a.Ji[num2].Properties.Go += 1
					a.Result[num1].SamHab = SaYuChuk
					a.Result[num2].SamHab = SaYuChuk
				}
			}

		}
	}

	//해묘미
	num1 = 5
	num2 = 5
	num3 = 5
	for i := 0; i < 4; i++ {
		if a.Ji[i].Title == "해" {
			num1 = i
		}
	}
	for j := 0; j < 4; j++ {
		if a.Ji[j].Title == "묘" {
			num2 = j
		}
	}
	for k := 0; k < 4; k++ {
		if a.Ji[k].Title == "미" {
			num3 = k
		}
	}

	if num1 != 5 || num2 != 5 || num3 != 5 {
		switch {
		case num1 == 5 && num2 == 5 && num3 == 5:
			{
				a.Ji[num1].Properties.Hae += 1
				a.Ji[num2].Properties.Go += 1
				a.Ji[num3].Properties.Ji += 1
				a.Result[num1].SamHab = HaeMyoMi
				a.Result[num2].SamHab = HaeMyoMi
				a.Result[num3].SamHab = HaeMyoMi
			}
		case num1 == 5:
			if num2 == 5 {
				if a.Ji[4].Title == "해" && a.Ji[5].Title == "묘" {
					a.Result[4].SamHab = HaeMyoMi
					a.Result[5].SamHab = HaeMyoMi
					a.Ji[num3].Properties.Ji += 1
					a.Result[num3].SamHab = HaeMyoMi
				} else if a.Ji[4].Title == "묘" && a.Ji[5].Title == "해" {
					a.Result[4].SamHab = HaeMyoMi
					a.Result[5].SamHab = HaeMyoMi
					a.Ji[num3].Properties.Ji += 1
					a.Result[num3].SamHab = HaeMyoMi
				}
			} else if num3 == 5 {
				if a.Ji[4].Title == "해" && a.Ji[5].Title == "미" {
					a.Result[4].SamHab = HaeMyoMi
					a.Result[5].SamHab = HaeMyoMi
					a.Ji[num2].Properties.Go += 1
					a.Result[num2].SamHab = HaeMyoMi
				} else if a.Ji[4].Title == "미" && a.Ji[5].Title == "해" {
					a.Result[4].SamHab = HaeMyoMi
					a.Result[5].SamHab = HaeMyoMi
					a.Ji[num2].Properties.Go += 1
					a.Result[num2].SamHab = HaeMyoMi
				}
			} else if num2 != 5 && num3 != 5 {
				if a.Ji[4].Title == "해" {
					a.Result[4].SamHab = HaeMyoMi
					a.Ji[num2].Properties.Go += 1
					a.Ji[num3].Properties.Ji += 1
					a.Result[num2].SamHab = HaeMyoMi
					a.Result[num3].SamHab = HaeMyoMi
				}

				if a.Ji[5].Title == "해" {
					a.Result[5].SamHab = HaeMyoMi
					a.Ji[num2].Properties.Go += 1
					a.Ji[num3].Properties.Ji += 1
					a.Result[num2].SamHab = HaeMyoMi
					a.Result[num3].SamHab = HaeMyoMi
				}
			}
			fallthrough

		case num2 == 5:
			if num3 == 5 {
				if a.Ji[4].Title == "묘" && a.Ji[5].Title == "미" {
					a.Result[4].SamHab = HaeMyoMi
					a.Result[5].SamHab = HaeMyoMi
					a.Ji[num1].Properties.Hae += 1
					a.Result[num1].SamHab = HaeMyoMi
				} else if a.Ji[4].Title == "미" && a.Ji[5].Title == "묘" {
					a.Result[4].SamHab = HaeMyoMi
					a.Result[5].SamHab = HaeMyoMi
					a.Ji[num1].Properties.Hae += 1
					a.Result[num1].SamHab = HaeMyoMi
				}
			} else if num1 != 5 && num3 != 5 {
				if a.Ji[4].Title == "묘" {
					a.Result[4].SamHab = HaeMyoMi
					a.Ji[num1].Properties.Hae += 1
					a.Ji[num3].Properties.Ji += 1
					a.Result[num1].SamHab = HaeMyoMi
					a.Result[num3].SamHab = HaeMyoMi
				} else if a.Ji[5].Title == "묘" {
					a.Result[5].SamHab = HaeMyoMi
					a.Ji[num1].Properties.Hae += 1
					a.Ji[num3].Properties.Ji += 1
					a.Result[num1].SamHab = HaeMyoMi
					a.Result[num3].SamHab = HaeMyoMi
				}
			}
			fallthrough

		case num3 == 5:
			if num1 != 5 && num2 != 5 {
				if a.Ji[4].Title == "미" {
					a.Result[4].SamHab = HaeMyoMi
					a.Ji[num1].Properties.Hae += 1
					a.Ji[num2].Properties.Go += 1
					a.Result[num1].SamHab = HaeMyoMi
					a.Result[num2].SamHab = HaeMyoMi
				} else if a.Ji[5].Title == "미" {
					a.Result[5].SamHab = HaeMyoMi
					a.Ji[num1].Properties.Hae += 1
					a.Ji[num2].Properties.Go += 1
					a.Result[num1].SamHab = HaeMyoMi
					a.Result[num2].SamHab = HaeMyoMi
				}
			}

		}

	}

	//인오술
	num1 = 5
	num2 = 5
	num3 = 5
	for i := 0; i < 4; i++ {
		if a.Ji[i].Title == "인" {
			num1 = i
		}
	}
	for j := 0; j < 4; j++ {
		if a.Ji[j].Title == "오" {
			num2 = j
		}
	}
	for k := 0; k < 4; k++ {
		if a.Ji[k].Title == "술" {
			num3 = k
		}
	}

	if num1 != 5 || num2 != 5 || num3 != 5 {
		switch {
		case num1 == 5 && num2 == 5 && num3 == 5:
			{
				a.Ji[num1].Properties.Hae += 1
				a.Ji[num2].Properties.Go += 1
				a.Ji[num3].Properties.Ji += 1
				a.Result[num1].SamHab = InOSul
				a.Result[num2].SamHab = InOSul
				a.Result[num3].SamHab = InOSul
			}
		case num1 == 5:
			if num2 == 5 {
				if a.Ji[4].Title == "인" && a.Ji[5].Title == "오" {
					a.Result[4].SamHab = InOSul
					a.Result[5].SamHab = InOSul
					a.Ji[num3].Properties.Ji += 1
					a.Result[num3].SamHab = InOSul
				} else if a.Ji[4].Title == "오" && a.Ji[5].Title == "인" {
					a.Result[4].SamHab = InOSul
					a.Result[5].SamHab = InOSul
					a.Ji[num3].Properties.Ji += 1
					a.Result[num3].SamHab = InOSul
				}
			} else if num3 == 5 {
				if a.Ji[4].Title == "인" && a.Ji[5].Title == "술" {
					a.Result[4].SamHab = InOSul
					a.Result[5].SamHab = InOSul
					a.Ji[num2].Properties.Go += 1
					a.Result[num2].SamHab = InOSul
				} else if a.Ji[4].Title == "술" && a.Ji[5].Title == "인" {
					a.Result[4].SamHab = InOSul
					a.Result[5].SamHab = InOSul
					a.Ji[num2].Properties.Go += 1
					a.Result[num2].SamHab = InOSul
				}
			} else if num2 != 5 && num3 != 5 {
				if a.Ji[4].Title == "인" {
					a.Result[4].SamHab = InOSul
					a.Ji[num2].Properties.Go += 1
					a.Ji[num3].Properties.Ji += 1
					a.Result[num2].SamHab = InOSul
					a.Result[num3].SamHab = InOSul
				}
				if a.Ji[5].Title == "인" {
					a.Result[5].SamHab = InOSul
					a.Ji[num2].Properties.Go += 1
					a.Ji[num3].Properties.Ji += 1
					a.Result[num2].SamHab = InOSul
					a.Result[num3].SamHab = InOSul
				}
			}
			fallthrough

		case num2 == 5:
			if num3 == 5 {
				if a.Ji[4].Title == "오" && a.Ji[5].Title == "술" {
					a.Result[4].SamHab = InOSul
					a.Result[5].SamHab = InOSul
					a.Ji[num1].Properties.Hae += 1
					a.Result[num1].SamHab = InOSul
				} else if a.Ji[4].Title == "술" && a.Ji[5].Title == "오" {
					a.Result[4].SamHab = InOSul
					a.Result[5].SamHab = InOSul
					a.Ji[num1].Properties.Hae += 1
					a.Result[num1].SamHab = InOSul
				}
			} else if num1 != 5 && num3 != 5 {
				if a.Ji[4].Title == "오" {
					a.Result[4].SamHab = InOSul
					a.Ji[num1].Properties.Hae += 1
					a.Ji[num3].Properties.Ji += 1
					a.Result[num1].SamHab = InOSul
					a.Result[num3].SamHab = InOSul
				} else if a.Ji[5].Title == "오" {
					a.Result[5].SamHab = InOSul
					a.Ji[num1].Properties.Hae += 1
					a.Ji[num3].Properties.Ji += 1
					a.Result[num1].SamHab = InOSul
					a.Result[num3].SamHab = InOSul
				}
			}
			fallthrough

		case num3 == 5:
			if num1 != 5 && num2 != 5 {
				if a.Ji[4].Title == "술" {
					a.Result[4].SamHab = InOSul
					a.Ji[num1].Properties.Hae += 1
					a.Ji[num2].Properties.Go += 1
					a.Result[num1].SamHab = InOSul
					a.Result[num2].SamHab = InOSul
				} else if a.Ji[5].Title == "술" {
					a.Result[5].SamHab = InOSul
					a.Ji[num1].Properties.Hae += 1
					a.Ji[num2].Properties.Go += 1
					a.Result[num1].SamHab = InOSul
					a.Result[num2].SamHab = InOSul
				}
			}

		}

	}
}

func Find_Yukhab(a *Person) {
	//축자
	num1 := 5
	num2 := 5
	for i := 0; i < 4; i++ {
		if a.Ji[i].Title == "축" {
			num1 = i
		}
	}
	for j := 0; j < 4; j++ {
		if a.Ji[j].Title == "자" {
			num2 = j
		}
	}
	a.Result[num1].YukHab = 0
	a.Result[num2].YukHab = 0
	a.Result[4].YukHab = 0
	a.Result[5].YukHab = 0
	if num1 != 5 || num2 != 5 {
		switch {
		case num1 != 5 && num2 != 5:
			a.Ji[num1].Properties.Prop = "earth"
			a.Ji[num2].Properties.Prop = "earth"
			a.Result[num1].YukHab = JaChuk
			a.Result[num2].YukHab = JaChuk
		case num1 == 5:
			if a.Ji[4].Title == "축" {
				a.Result[4].YukHab = JaChuk
				a.Ji[num2].Properties.Prop = "earth"
				a.Result[num2].YukHab = JaChuk
			} else if a.Ji[5].Title == "축" {
				a.Result[5].YukHab = JaChuk
				a.Ji[num2].Properties.Prop = "earth"
				a.Result[num2].YukHab = JaChuk
			}

			fallthrough
		case num2 == 5:
			if a.Ji[4].Title == "자" {
				a.Result[num1].YukHab = JaChuk
				a.Ji[num1].Properties.Prop = "earth"
				a.Result[4].YukHab = JaChuk
			} else if a.Ji[5].Title == "자" {
				a.Result[num1].YukHab = JaChuk
				a.Ji[num1].Properties.Prop = "earth"
				a.Result[5].YukHab = JaChuk
			}
		}
	}

	//인해
	num1 = 5
	num2 = 5

	for i := 0; i < 4; i++ {
		if a.Ji[i].Title == "인" {
			num1 = i
		}
	}
	for j := 0; j < 4; j++ {
		if a.Ji[j].Title == "해" {
			num2 = j
		}
	}

	if num1 != 5 || num2 != 5 {
		switch {
		case num1 != 5 && num2 != 5:
			a.Ji[num1].Properties.Prop = "tree"
			a.Ji[num2].Properties.Prop = "tree"
			a.Result[num1].YukHab = InHye
			a.Result[num2].YukHab = InHye
		case num1 == 5:
			if a.Ji[4].Title == "인" {
				a.Result[4].YukHab = InHye
				a.Ji[num2].Properties.Prop = "tree"
				a.Result[num2].YukHab = InHye
			} else if a.Ji[5].Title == "인" {
				a.Result[5].YukHab = InHye
				a.Ji[num2].Properties.Prop = "tree"
				a.Result[num2].YukHab = InHye
			}

			fallthrough
		case num2 == 5:
			if a.Ji[4].Title == "해" {
				a.Result[num1].YukHab = InHye
				a.Ji[num1].Properties.Prop = "tree"
				a.Result[4].YukHab = InHye
			} else if a.Ji[5].Title == "해" {
				a.Result[num1].YukHab = InHye
				a.Ji[num1].Properties.Prop = "tree"
				a.Result[5].YukHab = InHye
			}
		}
	}

	//묘술
	num1 = 5
	num2 = 5

	for i := 0; i < 4; i++ {
		if a.Ji[i].Title == "묘" {
			num1 = i
		}
	}
	for j := 0; j < 4; j++ {
		if a.Ji[j].Title == "술" {
			num2 = j
		}
	}

	if num1 != 5 || num2 != 5 {
		switch {
		case num1 != 5 && num2 != 5:
			a.Ji[num1].Properties.Prop = "fire"
			a.Ji[num2].Properties.Prop = "fire"
			a.Result[num1].YukHab = MyoSul
			a.Result[num2].YukHab = MyoSul
		case num1 == 5:
			if a.Ji[4].Title == "묘" {
				a.Result[4].YukHab = MyoSul
				a.Ji[num2].Properties.Prop = "fire"
				a.Result[num2].YukHab = MyoSul
			} else if a.Ji[5].Title == "묘" {
				a.Result[5].YukHab = MyoSul
				a.Ji[num2].Properties.Prop = "fire"
				a.Result[num2].YukHab = MyoSul
			}
			fallthrough
		case num2 == 5:
			if a.Ji[4].Title == "술" {
				a.Result[num1].YukHab = MyoSul
				a.Ji[num1].Properties.Prop = "fire"
				a.Result[4].YukHab = MyoSul
			} else if a.Ji[5].Title == "술" {
				a.Result[num1].YukHab = MyoSul
				a.Ji[num1].Properties.Prop = "fire"
				a.Result[5].YukHab = MyoSul
			}

		}
	}

	//진유
	num1 = 5
	num2 = 5

	for i := 0; i < 4; i++ {
		if a.Ji[i].Title == "진" {
			num1 = i
		}
	}
	for j := 0; j < 4; j++ {
		if a.Ji[j].Title == "유" {
			num2 = j
		}
	}

	if num1 != 5 || num2 != 5 {
		switch {
		case num1 != 5 && num2 != 5:
			a.Ji[num1].Properties.Prop = "iron"
			a.Ji[num2].Properties.Prop = "iron"
			a.Result[num1].YukHab = JinYu
			a.Result[num2].YukHab = JinYu
		case num1 == 5:
			if a.Ji[4].Title == "진" {
				a.Result[4].YukHab = JinYu
				a.Ji[num2].Properties.Prop = "iron"
				a.Result[num2].YukHab = JinYu
			} else if a.Ji[5].Title == "진" {
				a.Result[5].YukHab = JinYu
				a.Ji[num2].Properties.Prop = "iron"
				a.Result[num2].YukHab = JinYu
			}

			fallthrough
		case num2 == 5:
			if a.Ji[4].Title == "유" {
				a.Result[num1].YukHab = JinYu
				a.Ji[num1].Properties.Prop = "iron"
				a.Result[4].YukHab = JinYu
			} else if a.Ji[5].Title == "유" {
				a.Result[num1].YukHab = JinYu
				a.Ji[num1].Properties.Prop = "iron"
				a.Result[5].YukHab = JinYu
			}
		}
	}

	//사신
	num1 = 5
	num2 = 5

	for i := 0; i < 4; i++ {
		if a.Ji[i].Title == "사" {
			num1 = i
		}
	}
	for j := 0; j < 4; j++ {
		if a.Ji[j].Title == "신" {
			num2 = j
		}
	}

	if num1 != 5 || num2 != 5 {
		switch {
		case num1 != 5 && num2 != 5:
			a.Ji[num1].Properties.Prop = "water"
			a.Ji[num2].Properties.Prop = "water"
			a.Result[num1].YukHab = SaSin
			a.Result[num2].YukHab = SaSin
		case num1 == 5:
			if a.Ji[4].Title == "사" {
				a.Result[4].YukHab = SaSin
				a.Ji[num2].Properties.Prop = "water"
				a.Result[num2].YukHab = SaSin
			} else if a.Ji[5].Title == "사" {
				a.Result[5].YukHab = SaSin
				a.Ji[num2].Properties.Prop = "water"
				a.Result[num2].YukHab = SaSin
			}

			fallthrough
		case num2 == 5:
			if a.Ji[4].Title == "신" {
				a.Result[num1].YukHab = SaSin
				a.Ji[num1].Properties.Prop = "water"
				a.Result[4].YukHab = SaSin
			} else if a.Ji[5].Title == "신" {
				a.Result[num1].YukHab = SaSin
				a.Ji[num1].Properties.Prop = "water"
				a.Result[5].YukHab = SaSin
			}
		}
	}

	//오미
	num1 = 5
	num2 = 5

	for i := 0; i < 4; i++ {
		if a.Ji[i].Title == "오" {
			num1 = i
		}
	}
	for j := 0; j < 4; j++ {
		if a.Ji[j].Title == "미" {
			num2 = j
		}
	}

	if num1 != 5 || num2 != 5 {
		switch {
		case num1 != 5 && num2 != 5:
			a.Ji[num1].Properties.Prop = "fire"
			a.Ji[num2].Properties.Prop = "fire"
			a.Result[num1].YukHab = OMi
			a.Result[num2].YukHab = OMi
		case num1 == 5:
			if a.Ji[4].Title == "오" {
				a.Result[4].YukHab = OMi
				a.Ji[num2].Properties.Prop = "fire"
				a.Result[num2].YukHab = OMi
			} else if a.Ji[5].Title == "오" {
				a.Result[5].YukHab = OMi
				a.Ji[num2].Properties.Prop = "fire"
				a.Result[num2].YukHab = OMi
			}

			fallthrough
		case num2 == 5:
			if a.Ji[4].Title == "미" {
				a.Result[num1].YukHab = OMi
				a.Ji[num1].Properties.Prop = "fire"
				a.Result[4].YukHab = OMi
			} else if a.Ji[5].Title == "미" {
				a.Result[num1].YukHab = OMi
				a.Ji[num1].Properties.Prop = "fire"
				a.Result[5].YukHab = OMi
			}
		}
	}

}

func Find_Hyungsal(a *Person) {

	//인사
	num1 := 5
	num2 := 5

	for i := 0; i < 4; i++ {
		if a.Ji[i].Title == "인" {
			num1 = i
		}
	}
	for j := 0; j < 4; j++ {
		if a.Ji[j].Title == "사" {
			num2 = j
		}
	}

	a.Result[num1].Hyung = 0
	a.Result[num2].Hyung = 0

	if num1 != 5 || num2 != 5 {
		switch {
		case num1 != 5 && num2 != 5:
			a.Result[num1].Hyung = InSa
			a.Result[num2].Hyung = InSa
		case num1 == 5:
			if a.Ji[4].Title == "인" {
				a.Result[4].Hyung = InSa
				a.Result[num2].Hyung = InSa
			} else if a.Ji[5].Title == "인" {
				a.Result[5].Hyung = InSa
				a.Result[num2].Hyung = InSa
			}

			fallthrough
		case num2 == 5:
			if a.Ji[4].Title == "사" {
				a.Result[num1].Hyung = InSa
				a.Result[4].Hyung = InSa
			} else if a.Ji[5].Title == "사" {
				a.Result[num1].Hyung = InSa
				a.Result[5].Hyung = InSa
			}
		}
	}
	//사신
	num1 = 5
	num2 = 5

	for i := 0; i < 4; i++ {
		if a.Ji[i].Title == "사" {
			num1 = i
		}
	}
	for j := 0; j < 4; j++ {
		if a.Ji[j].Title == "신" {
			num2 = j
		}
	}

	if num1 != 5 || num2 != 5 {
		switch {
		case num1 != 5 && num2 != 5:
			a.Result[num1].Hyung = SaSin
			a.Result[num2].Hyung = SaSin
		case num1 == 5:
			if a.Ji[4].Title == "사" {
				a.Result[4].Hyung = SaSin
				a.Result[num2].Hyung = SaSin
			} else if a.Ji[5].Title == "사" {
				a.Result[5].Hyung = SaSin
				a.Result[num2].Hyung = SaSin
			}

			fallthrough
		case num2 == 5:
			if a.Ji[4].Title == "신" {
				a.Result[num1].Hyung = SaSin
				a.Result[4].Hyung = SaSin
			} else if a.Ji[5].Title == "신" {
				a.Result[num1].Hyung = SaSin
				a.Result[5].Hyung = SaSin
			}
		}
	}

	//축술
	num1 = 5
	num2 = 5

	for i := 0; i < 4; i++ {
		if a.Ji[i].Title == "축" {
			num1 = i
		}
	}
	for j := 0; j < 4; j++ {
		if a.Ji[j].Title == "술" {
			num2 = j
		}
	}

	if num1 != 5 || num2 != 5 {
		switch {
		case num1 != 5 && num2 != 5:
			a.Result[num1].Hyung = ChukSul
			a.Result[num2].Hyung = ChukSul
		case num1 == 5:
			if a.Ji[4].Title == "축" {
				a.Result[4].Hyung = ChukSul
				a.Result[num2].Hyung = ChukSul
			} else if a.Ji[5].Title == "축" {
				a.Result[5].Hyung = ChukSul
				a.Result[num2].Hyung = ChukSul
			}

			fallthrough
		case num2 == 5:
			if a.Ji[4].Title == "술" {
				a.Result[num1].Hyung = ChukSul
				a.Result[4].Hyung = ChukSul
			} else if a.Ji[5].Title == "술" {
				a.Result[num1].Hyung = ChukSul
				a.Result[5].Hyung = ChukSul
			}
		}
	}

	//술미
	num1 = 5
	num2 = 5

	for i := 0; i < 4; i++ {
		if a.Ji[i].Title == "술" {
			num1 = i
		}
	}
	for j := 0; j < 4; j++ {
		if a.Ji[j].Title == "미" {
			num2 = j
		}
	}

	if num1 != 5 || num2 != 5 {
		switch {
		case num1 != 5 && num2 != 5:
			a.Result[num1].Hyung = SulMi
			a.Result[num2].Hyung = SulMi
		case num1 == 5:
			if a.Ji[4].Title == "술" {
				a.Result[4].Hyung = SulMi
				a.Result[num2].Hyung = SulMi
			} else if a.Ji[5].Title == "술" {
				a.Result[5].Hyung = SulMi
				a.Result[num2].Hyung = SulMi
			}

			fallthrough
		case num2 == 5:
			if a.Ji[4].Title == "미" {
				a.Result[num1].Hyung = SulMi
				a.Result[4].Hyung = SulMi
			} else if a.Ji[5].Title == "미" {
				a.Result[num1].Hyung = SulMi
				a.Result[5].Hyung = SulMi
			}
		}
	}

	//자묘
	num1 = 5
	num2 = 5

	for i := 0; i < 4; i++ {
		if a.Ji[i].Title == "자" {
			num1 = i
		}
	}
	for j := 0; j < 4; j++ {
		if a.Ji[j].Title == "묘" {
			num2 = j
		}
	}

	if num1 != 5 || num2 != 5 {
		switch {
		case num1 != 5 && num2 != 5:
			a.Result[num1].Hyung = JaMyo
			a.Result[num2].Hyung = JaMyo
		case num1 == 5:
			if a.Ji[4].Title == "자" {
				a.Result[4].Hyung = JaMyo
				a.Result[num2].Hyung = JaMyo
			} else if a.Ji[5].Title == "자" {
				a.Result[5].Hyung = JaMyo
				a.Result[num2].Hyung = JaMyo
			}
			fallthrough
		case num2 == 5:
			if a.Ji[4].Title == "묘" {
				a.Result[num1].Hyung = JaMyo
				a.Result[4].Hyung = JaMyo
			} else if a.Ji[5].Title == "묘" {
				a.Result[num1].Hyung = JaMyo
				a.Result[5].Hyung = JaMyo
			}
		}
	}

	//진진
	num1 = 5
	num2 = 5

	for i := 0; i < 4; i++ {
		if a.Ji[i].Title == "진" {
			num1 = i
		}
	}
	for j := 0; j < 4; j++ {
		if a.Ji[j].Title == "진" {
			num2 = j
		}
	}

	if num1 != 5 || num2 != 5 {
		switch {
		case num1 != 5 && num2 != 5:
			a.Result[num1].Hyung = JinJin
			a.Result[num2].Hyung = JinJin
		case num1 == 5:
			if a.Ji[4].Title == "진" {
				a.Result[4].Hyung = JinJin
				a.Result[num2].Hyung = JinJin
			} else if a.Ji[5].Title == "진" {
				a.Result[5].Hyung = JinJin
				a.Result[num2].Hyung = JinJin
			}

			fallthrough
		case num2 == 5:
			if a.Ji[4].Title == "진" {
				a.Result[num1].Hyung = JinJin
				a.Result[4].Hyung = JinJin
			} else if a.Ji[5].Title == "진" {
				a.Result[num1].Hyung = JinJin
				a.Result[5].Hyung = JinJin
			}
		}
	}

	//오오
	num1 = 5
	num2 = 5

	for i := 0; i < 4; i++ {
		if a.Ji[i].Title == "오" {
			num1 = i
		}
	}
	for j := 0; j < 4; j++ {
		if a.Ji[j].Title == "오" {
			num2 = j
		}
	}

	if num1 != 5 || num2 != 5 {
		switch {
		case num1 != 5 && num2 != 5:
			a.Result[num1].Hyung = OO
			a.Result[num2].Hyung = OO
		case num1 == 5:
			if a.Ji[4].Title == "오" {
				a.Result[4].Hyung = OO
				a.Result[num2].Hyung = OO
			} else if a.Ji[5].Title == "오" {
				a.Result[5].Hyung = OO
				a.Result[num2].Hyung = OO
			}

			fallthrough
		case num2 == 5:
			if a.Ji[4].Title == "오" {
				a.Result[num1].Hyung = OO
				a.Result[4].Hyung = OO
			} else if a.Ji[5].Title == "오" {
				a.Result[num1].Hyung = OO
				a.Result[5].Hyung = OO
			}
		}
	}

	//유유
	for i := 0; i < 4; i++ {
		if a.Ji[i].Title == "유" {
			num1 = i
		}
	}
	for j := 0; j < 4; j++ {
		if a.Ji[j].Title == "유" {
			num2 = j
		}
	}

	if num1 != 5 || num2 != 5 {
		switch {
		case num1 != 5 && num2 != 5:
			a.Result[num1].Hyung = YuYu
			a.Result[num2].Hyung = YuYu
		case num1 == 5:
			if a.Ji[4].Title == "유" {
				a.Result[4].Hyung = YuYu
				a.Result[num2].Hyung = YuYu
			} else if a.Ji[5].Title == "유" {
				a.Result[5].Hyung = YuYu
				a.Result[num2].Hyung = YuYu
			}

			fallthrough
		case num2 == 5:
			if a.Ji[4].Title == "유" {
				a.Result[num1].Hyung = YuYu
				a.Result[4].Hyung = YuYu
			} else if a.Ji[5].Title == "유" {
				a.Result[num1].Hyung = YuYu
				a.Result[5].Hyung = YuYu
			}
		}
	}

	//해해
	num1 = 5
	num2 = 5

	for i := 0; i < 4; i++ {
		if a.Ji[i].Title == "해" {
			num1 = i
		}
	}
	for j := 0; j < 4; j++ {
		if a.Ji[j].Title == "해" {
			num2 = j
		}
	}

	if num1 != 5 || num2 != 5 {
		switch {
		case num1 != 5 && num2 != 5:
			a.Result[num1].Hyung = HaeHae
			a.Result[num2].Hyung = HaeHae
		case num1 == 5:
			if a.Ji[4].Title == "해" {
				a.Result[4].Hyung = HaeHae
				a.Result[num2].Hyung = HaeHae
			} else if a.Ji[5].Title == "해" {
				a.Result[5].Hyung = HaeHae
				a.Result[num2].Hyung = HaeHae
			}

			fallthrough
		case num2 == 5:
			if a.Ji[4].Title == "해" {
				a.Result[num1].Hyung = HaeHae
				a.Result[4].Hyung = HaeHae
			} else if a.Ji[5].Title == "해" {
				a.Result[num1].Hyung = HaeHae
				a.Result[5].Hyung = HaeHae
			}

		}
	}
}

func Find_Choongsal(a *Person) {
	//인신
	num1 := 5
	num2 := 5

	for i := 0; i < 4; i++ {
		if a.Ji[i].Title == "인" {
			num1 = i
		}
	}
	for j := 0; j < 4; j++ {
		if a.Ji[j].Title == "신" {
			num2 = j
		}
	}
	a.Result[num1].Choong = 0
	a.Result[num2].Choong = 0

	if num1 != 5 || num2 != 5 {
		switch {
		case num1 != 5 && num2 != 5:
			a.Result[num1].Choong = InSin
			a.Result[num2].Choong = InSin
		case num1 == 5:
			if a.Ji[4].Title == "인" {
				a.Result[4].Choong = InSin
				a.Result[num2].Choong = InSin
			} else if a.Ji[5].Title == "인" {
				a.Result[5].Choong = InSin
				a.Result[num2].Choong = InSin
			}

			fallthrough
		case num2 == 5:
			if a.Ji[4].Title == "신" {
				a.Result[num1].Choong = InSin
				a.Result[4].Choong = InSin
			} else if a.Ji[5].Title == "신" {
				a.Result[num1].Choong = InSin
				a.Result[5].Choong = InSin
			}
		}
	}
	//묘유
	num1 = 5
	num2 = 5

	for i := 0; i < 4; i++ {
		if a.Ji[i].Title == "묘" {
			num1 = i
		}
	}
	for j := 0; j < 4; j++ {
		if a.Ji[j].Title == "유" {
			num2 = j
		}
	}

	if num1 != 5 || num2 != 5 {
		switch {
		case num1 != 5 && num2 != 5:
			a.Result[num1].Choong = MyoYu
			a.Result[num2].Choong = MyoYu
		case num1 == 5:
			if a.Ji[4].Title == "묘" {
				a.Result[4].Choong = MyoYu
				a.Result[num2].Choong = MyoYu
			} else if a.Ji[5].Title == "묘" {
				a.Result[5].Choong = MyoYu
				a.Result[num2].Choong = MyoYu
			}

			fallthrough
		case num2 == 5:
			if a.Ji[4].Title == "유" {
				a.Result[num1].Choong = MyoYu
				a.Result[4].Choong = MyoYu
			} else if a.Ji[5].Title == "유" {
				a.Result[num1].Choong = MyoYu
				a.Result[5].Choong = MyoYu
			}

		}
	}

	//진술
	num1 = 5
	num2 = 5

	for i := 0; i < 4; i++ {
		if a.Ji[i].Title == "진" {
			num1 = i
		}
	}
	for j := 0; j < 4; j++ {
		if a.Ji[j].Title == "술" {
			num2 = j
		}
	}

	if num1 != 5 || num2 != 5 {
		switch {
		case num1 != 5 && num2 != 5:
			a.Result[num1].Choong = JinSul
			a.Result[num2].Choong = JinSul
		case num1 == 5:
			if a.Ji[4].Title == "진" {
				a.Result[4].Choong = JinSul
				a.Result[num2].Choong = JinSul
			} else if a.Ji[5].Title == "진" {
				a.Result[5].Choong = JinSul
				a.Result[num2].Choong = JinSul
			}

			fallthrough
		case num2 == 5:
			if a.Ji[4].Title == "술" {
				a.Result[num1].Choong = JinSul
				a.Result[4].Choong = JinSul
			} else if a.Ji[5].Title == "술" {
				a.Result[num1].Choong = JinSul
				a.Result[5].Choong = JinSul
			}
		}
	}

	//사해
	num1 = 5
	num2 = 5

	for i := 0; i < 4; i++ {
		if a.Ji[i].Title == "사" {
			num1 = i
		}
	}
	for j := 0; j < 4; j++ {
		if a.Ji[j].Title == "해" {
			num2 = j
		}
	}

	if num1 != 5 || num2 != 5 {
		switch {
		case num1 != 5 && num2 != 5:
			a.Result[num1].Choong = SaHae
			a.Result[num2].Choong = SaHae
		case num1 == 5:
			if a.Ji[4].Title == "사" {
				a.Result[4].Choong = SaHae
				a.Result[num2].Choong = SaHae
			} else if a.Ji[5].Title == "사" {
				a.Result[5].Choong = SaHae
				a.Result[num2].Choong = SaHae
			}

			fallthrough
		case num2 == 5:
			if a.Ji[4].Title == "해" {
				a.Result[num1].Choong = SaHae
				a.Result[4].Choong = SaHae
			} else if a.Ji[5].Title == "해" {
				a.Result[num1].Choong = SaHae
				a.Result[5].Choong = SaHae
			}
		}
	}
	//자오
	num1 = 5
	num2 = 5

	for i := 0; i < 4; i++ {
		if a.Ji[i].Title == "자" {
			num1 = i
		}
	}
	for j := 0; j < 4; j++ {
		if a.Ji[j].Title == "오" {
			num2 = j
		}
	}

	if num1 != 5 || num2 != 5 {
		switch {
		case num1 != 5 && num2 != 5:
			a.Result[num1].Choong = JaO
			a.Result[num2].Choong = JaO
		case num1 == 5:
			if a.Ji[4].Title == "자" {
				a.Result[4].Choong = JaO
				a.Result[num2].Choong = JaO
			} else if a.Ji[5].Title == "자" {
				a.Result[5].Choong = JaO
				a.Result[num2].Choong = JaO
			}

			fallthrough
		case num2 == 5:
			if a.Ji[4].Title == "오" {
				a.Result[num1].Choong = JaO
				a.Result[4].Choong = JaO
			} else if a.Ji[5].Title == "오" {
				a.Result[num1].Choong = JaO
				a.Result[5].Choong = JaO
			}

		}
	}

	//축미
	num1 = 5
	num2 = 5

	for i := 0; i < 4; i++ {
		if a.Ji[i].Title == "축" {
			num1 = i
		}
	}
	for j := 0; j < 4; j++ {
		if a.Ji[j].Title == "미" {
			num2 = j
		}
	}

	if num1 != 5 || num2 != 5 {
		switch {
		case num1 != 5 && num2 != 5:
			a.Result[num1].Choong = ChukMi
			a.Result[num2].Choong = ChukMi
		case num1 == 5:
			if a.Ji[4].Title == "축" {
				a.Result[4].Choong = ChukMi
				a.Result[num2].Choong = ChukMi
			} else if a.Ji[5].Title == "축" {
				a.Result[5].Choong = ChukMi
				a.Result[num2].Choong = ChukMi
			}

			fallthrough
		case num2 == 5:
			if a.Ji[4].Title == "미" {
				a.Result[num1].Choong = ChukMi
				a.Result[4].Choong = ChukMi
			} else if a.Ji[5].Title == "미" {
				a.Result[num1].Choong = ChukMi
				a.Result[5].Choong = ChukMi
			}
		}
	}
}

func Find_Pasal(a *Person) {
	//자유 귀문이 우선
	num1 := 5
	num2 := 5

	for i := 0; i < 4; i++ {
		if a.Ji[i].Title == "자" {
			num1 = i
		}
	}
	for j := 0; j < 4; j++ {
		if a.Ji[j].Title == "유" {
			num2 = j
		}
	}
	a.Result[num1].Pa = 0
	a.Result[num2].Pa = 0

	if num1 != 5 || num2 != 5 {
		switch {
		case num1 != 5 && num2 != 5:
			a.Result[num1].Pa = JaYu
			a.Result[num2].Pa = JaYu
		case num1 == 5:
			if a.Ji[4].Title == "자" {
				a.Result[4].Pa = JaYu
				a.Result[num2].Pa = JaYu
			} else if a.Ji[5].Title == "자" {
				a.Result[5].Pa = JaYu
				a.Result[num2].Pa = JaYu
			}

			fallthrough
		case num2 == 5:
			if a.Ji[4].Title == "유" {
				a.Result[num1].Pa = JaYu
				a.Result[4].Pa = JaYu
			} else if a.Ji[5].Title == "유" {
				a.Result[num1].Pa = JaYu
				a.Result[5].Pa = JaYu
			}
		}
	}

	//묘오
	num1 = 5
	num2 = 5

	for i := 0; i < 4; i++ {
		if a.Ji[i].Title == "묘" {
			num1 = i
		}
	}
	for j := 0; j < 4; j++ {
		if a.Ji[j].Title == "오" {
			num2 = j
		}
	}

	if num1 != 5 || num2 != 5 {
		switch {
		case num1 != 5 && num2 != 5:
			a.Result[num1].Pa = MyoO
			a.Result[num2].Pa = MyoO
		case num1 == 5:
			if a.Ji[4].Title == "묘" {
				a.Result[4].Pa = MyoO
				a.Result[num2].Pa = MyoO
			} else if a.Ji[5].Title == "묘" {
				a.Result[5].Pa = MyoO
				a.Result[num2].Pa = MyoO
			}

			fallthrough
		case num2 == 5:
			if a.Ji[4].Title == "오" {
				a.Result[num1].Pa = MyoO
				a.Result[4].Pa = MyoO
			} else if a.Ji[5].Title == "오" {
				a.Result[num1].Pa = MyoO
				a.Result[5].Pa = MyoO
			}

		}
	}

	//사신  형합이 우선
	//사신
	num1 = 5
	num2 = 5

	for i := 0; i < 4; i++ {
		if a.Ji[i].Title == "사" {
			num1 = i
		}
	}
	for j := 0; j < 4; j++ {
		if a.Ji[j].Title == "신" {
			num2 = j
		}
	}

	if num1 != 5 || num2 != 5 {
		switch {
		case num1 != 5 && num2 != 5:
			a.Result[num1].Pa = PaSaSin
			a.Result[num2].Pa = PaSaSin
		case num1 == 5:
			if a.Ji[4].Title == "사" {
				a.Result[4].Pa = PaSaSin
				a.Result[num2].Pa = PaSaSin
			} else if a.Ji[5].Title == "사" {
				a.Result[5].Pa = PaSaSin
				a.Result[num2].Pa = PaSaSin
			}

			fallthrough
		case num2 == 5:
			if a.Ji[4].Title == "신" {
				a.Result[num1].Pa = PaSaSin
				a.Result[4].Pa = PaSaSin
			} else if a.Ji[5].Title == "신" {
				a.Result[num1].Pa = PaSaSin
				a.Result[5].Pa = PaSaSin
			}

		}
	}

	//축진
	num1 = 5
	num2 = 5

	for i := 0; i < 4; i++ {
		if a.Ji[i].Title == "축" {
			num1 = i
		}
	}
	for j := 0; j < 4; j++ {
		if a.Ji[j].Title == "진" {
			num2 = j
		}
	}

	if num1 != 5 || num2 != 5 {
		switch {
		case num1 != 5 && num2 != 5:
			a.Result[num1].Pa = JinChuk
			a.Result[num2].Pa = JinChuk
		case num1 == 5:
			if a.Ji[4].Title == "축" {
				a.Result[4].Pa = JinChuk
				a.Result[num2].Pa = JinChuk
			} else if a.Ji[5].Title == "축" {
				a.Result[5].Pa = JinChuk
				a.Result[num2].Pa = JinChuk
			}

			fallthrough
		case num2 == 5:
			if a.Ji[4].Title == "진" {
				a.Result[num1].Pa = JinChuk
				a.Result[4].Pa = JinChuk
			} else if a.Ji[5].Title == "진" {
				a.Result[num1].Pa = JinChuk
				a.Result[5].Pa = JinChuk
			}
		}
	}

	//술미 형이 우선
	num1 = 5
	num2 = 5

	for i := 0; i < 4; i++ {
		if a.Ji[i].Title == "술" {
			num1 = i
		}
	}
	for j := 0; j < 4; j++ {
		if a.Ji[j].Title == "미" {
			num2 = j
		}
	}

	if num1 != 5 || num2 != 5 {
		switch {
		case num1 != 5 && num2 != 5:
			a.Result[num1].Pa = PaSulMi
			a.Result[num2].Pa = PaSulMi
		case num1 == 5:
			if a.Ji[4].Title == "술" {
				a.Result[4].Pa = PaSulMi
				a.Result[num2].Pa = PaSulMi
			} else if a.Ji[5].Title == "술" {
				a.Result[5].Pa = PaSulMi
				a.Result[num2].Pa = PaSulMi
			}

			fallthrough
		case num2 == 5:
			if a.Ji[4].Title == "미" {
				a.Result[num1].Pa = PaSulMi
				a.Result[4].Pa = PaSulMi
			} else if a.Ji[5].Title == "미" {
				a.Result[num1].Pa = PaSulMi
				a.Result[5].Pa = PaSulMi
			}
		}
	}

	//인해 인중 병 손상
	num1 = 5
	num2 = 5

	for i := 0; i < 4; i++ {
		if a.Ji[i].Title == "인" {
			num1 = i
		}
	}
	for j := 0; j < 4; j++ {
		if a.Ji[j].Title == "해" {
			num2 = j
		}
	}

	if num1 != 5 || num2 != 5 {
		switch {
		case num1 != 5 && num2 != 5:
			a.Result[num1].Pa = InHae
			a.Result[num2].Pa = InHae
		case num1 == 5:
			if a.Ji[4].Title == "인" {
				a.Result[4].Pa = InHae
				a.Result[num2].Pa = InHae
			} else if a.Ji[5].Title == "인" {
				a.Result[5].Pa = InHae
				a.Result[num2].Pa = InHae
			}

			fallthrough
		case num2 == 5:
			if a.Ji[4].Title == "해" {
				a.Result[num1].Pa = InHae
				a.Result[4].Pa = InHae
			} else if a.Ji[5].Title == "해" {
				a.Result[num1].Pa = InHae
				a.Result[5].Pa = InHae
			}
		}
	}
}

func Find_Haesal(a *Person) {
	//자미
	num1 := 5
	num2 := 5

	for i := 0; i < 4; i++ {
		if a.Ji[i].Title == "자" {
			num1 = i
		}
	}
	for j := 0; j < 4; j++ {
		if a.Ji[j].Title == "미" {
			num2 = j
		}
	}

	a.Result[num1].Hae = 0
	a.Result[num2].Hae = 0
	if num1 != 5 || num2 != 5 {
		switch {
		case num1 != 5 && num2 != 5:
			a.Result[num1].Hae = JaMi
			a.Result[num2].Hae = JaMi
		case num1 == 5:
			if a.Ji[4].Title == "자" {
				a.Result[4].Hae = JaMi
				a.Result[num2].Hae = JaMi
			} else if a.Ji[5].Title == "자" {
				a.Result[5].Hae = JaMi
				a.Result[num2].Hae = JaMi
			}

			fallthrough
		case num2 == 5:
			if a.Ji[4].Title == "미" {
				a.Result[num1].Hae = JaMi
				a.Result[4].Hae = JaMi
			} else if a.Ji[5].Title == "미" {
				a.Result[num1].Hae = JaMi
				a.Result[5].Hae = JaMi
			}
		}
	}

	//축오 탕화중독
	num1 = 5
	num2 = 5

	for i := 0; i < 4; i++ {
		if a.Ji[i].Title == "축" {
			num1 = i
		}
	}
	for j := 0; j < 4; j++ {
		if a.Ji[j].Title == "오" {
			num2 = j
		}
	}

	if num1 != 5 || num2 != 5 {
		switch {
		case num1 != 5 && num2 != 5:
			a.Result[num1].Hae = OChuk
			a.Result[num2].Hae = OChuk
		case num1 == 5:
			if a.Ji[4].Title == "축" {
				a.Result[4].Hae = OChuk
				a.Result[num2].Hae = OChuk
			} else if a.Ji[5].Title == "축" {
				a.Result[5].Hae = OChuk
				a.Result[num2].Hae = OChuk
			}

			fallthrough
		case num2 == 5:
			if a.Ji[4].Title == "오" {
				a.Result[num1].Hae = OChuk
				a.Result[4].Hae = OChuk
			} else if a.Ji[5].Title == "오" {
				a.Result[num1].Hae = OChuk
				a.Result[5].Hae = OChuk
			}
		}
	}

	//인사 형이우선
	num1 = 5
	num2 = 5

	for i := 0; i < 4; i++ {
		if a.Ji[i].Title == "인" {
			num1 = i
		}
	}
	for j := 0; j < 4; j++ {
		if a.Ji[j].Title == "사" {
			num2 = j
		}
	}

	if num1 != 5 || num2 != 5 {
		switch {
		case num1 != 5 && num2 != 5:
			a.Result[num1].Hae = HaeInSa
			a.Result[num2].Hae = HaeInSa
		case num1 == 5:
			if a.Ji[4].Title == "인" {
				a.Result[4].Hae = HaeInSa
				a.Result[num2].Hae = HaeInSa
			} else if a.Ji[5].Title == "인" {
				a.Result[5].Hae = HaeInSa
				a.Result[num2].Hae = HaeInSa
			}

			fallthrough
		case num2 == 5:
			if a.Ji[4].Title == "사" {
				a.Result[num1].Hae = HaeInSa
				a.Result[4].Hae = HaeInSa
			} else if a.Ji[5].Title == "사" {
				a.Result[num1].Hae = HaeInSa
				a.Result[5].Hae = HaeInSa
			}
		}
	}

	//묘진
	num1 = 5
	num2 = 5

	for i := 0; i < 4; i++ {
		if a.Ji[i].Title == "묘" {
			num1 = i
		}
	}
	for j := 0; j < 4; j++ {
		if a.Ji[j].Title == "진" {
			num2 = j
		}
	}

	if num1 != 5 || num2 != 5 {
		switch {
		case num1 != 5 && num2 != 5:
			a.Result[num1].Hae = MyoJin
			a.Result[num2].Hae = MyoJin
		case num1 == 5:
			if a.Ji[4].Title == "묘" {
				a.Result[4].Hae = MyoJin
				a.Result[num2].Hae = MyoJin
			} else if a.Ji[5].Title == "묘" {
				a.Result[5].Hae = MyoJin
				a.Result[num2].Hae = MyoJin
			}

			fallthrough
		case num2 == 5:
			if a.Ji[4].Title == "진" {
				a.Result[num1].Hae = MyoJin
				a.Result[4].Hae = MyoJin
			} else if a.Ji[5].Title == "진" {
				a.Result[num1].Hae = MyoJin
				a.Result[5].Hae = MyoJin
			}
		}
	}

	//신해 신이 손상
	num1 = 5
	num2 = 5

	for i := 0; i < 4; i++ {
		if a.Ji[i].Title == "신" {
			num1 = i
		}
	}
	for j := 0; j < 4; j++ {
		if a.Ji[j].Title == "해" {
			num2 = j
		}
	}

	if num1 != 5 || num2 != 5 {
		switch {
		case num1 != 5 && num2 != 5:
			a.Result[num1].Hae = HaeSin
			a.Result[num2].Hae = HaeSin
		case num1 == 5:
			if a.Ji[4].Title == "신" {
				a.Result[4].Hae = HaeSin
				a.Result[num2].Hae = HaeSin
			} else if a.Ji[5].Title == "신" {
				a.Result[5].Hae = HaeSin
				a.Result[num2].Hae = HaeSin
			}

			fallthrough
		case num2 == 5:
			if a.Ji[4].Title == "해" {
				a.Result[num1].Hae = HaeSin
				a.Result[4].Hae = HaeSin
			} else if a.Ji[5].Title == "해" {
				a.Result[num1].Hae = HaeSin
				a.Result[5].Hae = HaeSin
			}
		}
	}

	//유술
	num1 = 5
	num2 = 5

	for i := 0; i < 4; i++ {
		if a.Ji[i].Title == "유" {
			num1 = i
		}
	}
	for j := 0; j < 4; j++ {
		if a.Ji[j].Title == "술" {
			num2 = j
		}
	}

	if num1 != 5 || num2 != 5 {
		switch {
		case num1 != 5 && num2 != 5:
			a.Result[num1].Hae = YuSul
			a.Result[num2].Hae = YuSul
		case num1 == 5:
			if a.Ji[4].Title == "유" {
				a.Result[4].Hae = YuSul
				a.Result[num2].Hae = YuSul
			} else if a.Ji[5].Title == "유" {
				a.Result[5].Hae = YuSul
				a.Result[num2].Hae = YuSul
			}

			fallthrough
		case num2 == 5:
			if a.Ji[4].Title == "술" {
				a.Result[num1].Hae = YuSul
				a.Result[4].Hae = YuSul
			} else if a.Ji[5].Title == "술" {
				a.Result[num1].Hae = YuSul
				a.Result[5].Hae = YuSul
			}
		}
	}
}

func Find_Wonzin(a *Person) {
	//인유
	num1 := 5
	num2 := 5

	for i := 0; i < 4; i++ {
		if a.Ji[i].Title == "인" {
			num1 = i
		}
	}
	for j := 0; j < 4; j++ {
		if a.Ji[j].Title == "유" {
			num2 = j
		}
	}
	a.Result[num1].WonJin = 0
	a.Result[num2].WonJin = 0

	if num1 != 5 || num2 != 5 {
		switch {
		case num1 != 5 && num2 != 5:
			a.Result[num1].WonJin = InYu
			a.Result[num2].WonJin = InYu
		case num1 == 5:
			if a.Ji[4].Title == "인" {
				a.Result[4].WonJin = InYu
				a.Result[num2].WonJin = InYu
			} else if a.Ji[5].Title == "인" {
				a.Result[5].WonJin = InYu
				a.Result[num2].WonJin = InYu
			}

			fallthrough
		case num2 == 5:
			if a.Ji[4].Title == "유" {
				a.Result[num1].WonJin = InYu
				a.Result[4].WonJin = InYu
			} else if a.Ji[5].Title == "유" {
				a.Result[num1].WonJin = InYu
				a.Result[5].WonJin = InYu
			}
		}
	}

	//묘신 금목상쟁
	num1 = 5
	num2 = 5

	for i := 0; i < 4; i++ {
		if a.Ji[i].Title == "묘" {
			num1 = i
		}
	}
	for j := 0; j < 4; j++ {
		if a.Ji[j].Title == "신" {
			num2 = j
		}
	}

	if num1 != 5 || num2 != 5 {
		switch {
		case num1 != 5 && num2 != 5:
			a.Result[num1].WonJin = MyoSin
			a.Result[num2].WonJin = MyoSin
		case num1 == 5:
			if a.Ji[4].Title == "묘" {
				a.Result[4].WonJin = MyoSin
				a.Result[num2].WonJin = MyoSin
			} else if a.Ji[5].Title == "묘" {
				a.Result[5].WonJin = MyoSin
				a.Result[num2].WonJin = MyoSin
			}

			fallthrough
		case num2 == 5:
			if a.Ji[4].Title == "신" {
				a.Result[num1].WonJin = MyoSin
				a.Result[4].WonJin = MyoSin
			} else if a.Ji[5].Title == "신" {
				a.Result[num1].WonJin = MyoSin
				a.Result[5].WonJin = MyoSin
			}
		}
	}

	//진해 입묘형
	num1 = 5
	num2 = 5

	for i := 0; i < 4; i++ {
		if a.Ji[i].Title == "진" {
			num1 = i
		}
	}
	for j := 0; j < 4; j++ {
		if a.Ji[j].Title == "해" {
			num2 = j
		}
	}

	if num1 != 5 || num2 != 5 {
		switch {
		case num1 != 5 && num2 != 5:
			a.Result[num1].WonJin = JinHae
			a.Result[num2].WonJin = JinHae
		case num1 == 5:
			if a.Ji[4].Title == "진" {
				a.Result[4].WonJin = JinHae
				a.Result[num2].WonJin = JinHae
			} else if a.Ji[5].Title == "진" {
				a.Result[5].WonJin = JinHae
				a.Result[num2].WonJin = JinHae
			}

			fallthrough
		case num2 == 5:
			if a.Ji[4].Title == "해" {
				a.Result[num1].WonJin = JinHae
				a.Result[4].WonJin = JinHae
			} else if a.Ji[5].Title == "해" {
				a.Result[num1].WonJin = JinHae
				a.Result[5].WonJin = JinHae
			}
		}
	}

	//사술 입묘형
	num1 = 5
	num2 = 5

	for i := 0; i < 4; i++ {
		if a.Ji[i].Title == "사" {
			num1 = i
		}
	}
	for j := 0; j < 4; j++ {
		if a.Ji[j].Title == "술" {
			num2 = j
		}
	}

	if num1 != 5 || num2 != 5 {
		switch {
		case num1 != 5 && num2 != 5:
			a.Result[num1].WonJin = SaSul
			a.Result[num2].WonJin = SaSul
		case num1 == 5:
			if a.Ji[4].Title == "사" {
				a.Result[4].WonJin = SaSul
				a.Result[num2].WonJin = SaSul
			} else if a.Ji[5].Title == "사" {
				a.Result[5].WonJin = SaSul
				a.Result[num2].WonJin = SaSul
			}

			fallthrough
		case num2 == 5:
			if a.Ji[4].Title == "술" {
				a.Result[num1].WonJin = SaSul
				a.Result[4].WonJin = SaSul
			} else if a.Ji[5].Title == "술" {
				a.Result[num1].WonJin = SaSul
				a.Result[5].WonJin = SaSul
			}
		}
	}

	//오축 탕화형
	num1 = 5
	num2 = 5

	for i := 0; i < 4; i++ {
		if a.Ji[i].Title == "오" {
			num1 = i
		}
	}
	for j := 0; j < 4; j++ {
		if a.Ji[j].Title == "축" {
			num2 = j
		}
	}

	if num1 != 5 || num2 != 5 {
		switch {
		case num1 != 5 && num2 != 5:
			a.Result[num1].WonJin = WonOChuk
			a.Result[num2].WonJin = WonOChuk
		case num1 == 5:
			if a.Ji[4].Title == "오" {
				a.Result[4].WonJin = WonOChuk
				a.Result[num2].WonJin = WonOChuk
			} else if a.Ji[5].Title == "오" {
				a.Result[5].WonJin = WonOChuk
				a.Result[num2].WonJin = WonOChuk
			}

			fallthrough
		case num2 == 5:
			if a.Ji[4].Title == "축" {
				a.Result[num1].WonJin = WonOChuk
				a.Result[4].WonJin = WonOChuk
			} else if a.Ji[5].Title == "축" {
				a.Result[num1].WonJin = WonOChuk
				a.Result[5].WonJin = WonOChuk
			}
		}
	}

	//자미탕화원진귀문
	num1 = 5
	num2 = 5

	for i := 0; i < 4; i++ {
		if a.Ji[i].Title == "자" {
			num1 = i
		}
	}
	for j := 0; j < 4; j++ {
		if a.Ji[j].Title == "미" {
			num2 = j
		}
	}

	if num1 != 5 || num2 != 5 {
		switch {
		case num1 != 5 && num2 != 5:
			a.Result[num1].WonJin = WonJaMi
			a.Result[num2].WonJin = WonJaMi
		case num1 == 5:
			if a.Ji[4].Title == "자" {
				a.Result[4].WonJin = WonJaMi
				a.Result[num2].WonJin = WonJaMi
			} else if a.Ji[5].Title == "자" {
				a.Result[5].WonJin = WonJaMi
				a.Result[num2].WonJin = WonJaMi
			}

			fallthrough
		case num2 == 5:
			if a.Ji[4].Title == "미" {
				a.Result[num1].WonJin = WonJaMi
				a.Result[4].WonJin = WonJaMi
			} else if a.Ji[5].Title == "미" {
				a.Result[num1].WonJin = WonJaMi
				a.Result[5].WonJin = WonJaMi
			}
		}
	}
}

func Find_Guimun(a *Person) {
	//인미
	num1 := 5
	num2 := 5

	for i := 0; i < 4; i++ {
		if a.Ji[i].Title == "인" {
			num1 = i
		}
	}
	for j := 0; j < 4; j++ {
		if a.Ji[j].Title == "미" {
			num2 = j
		}
	}
	a.Result[num1].GuiMoon = 0
	a.Result[num2].GuiMoon = 0

	if num1 != 5 || num2 != 5 {
		switch {
		case num1 != 5 && num2 != 5:
			a.Result[num1].GuiMoon = InMi
			a.Result[num2].GuiMoon = InMi
		case num1 == 5:
			if a.Ji[4].Title == "인" {
				a.Result[4].GuiMoon = InMi
				a.Result[num2].GuiMoon = InMi
			} else if a.Ji[5].Title == "인" {
				a.Result[5].GuiMoon = InMi
				a.Result[num2].GuiMoon = InMi
			}

			fallthrough
		case num2 == 5:
			if a.Ji[4].Title == "미" {
				a.Result[num1].GuiMoon = InMi
				a.Result[4].GuiMoon = InMi
			} else if a.Ji[5].Title == "미" {
				a.Result[num1].GuiMoon = InMi
				a.Result[5].GuiMoon = InMi
			}
		}
	}

	//자유 음주형
	num1 = 5
	num2 = 5

	for i := 0; i < 4; i++ {
		if a.Ji[i].Title == "자" {
			num1 = i
		}
	}
	for j := 0; j < 4; j++ {
		if a.Ji[j].Title == "유" {
			num2 = j
		}
	}

	if num1 != 5 || num2 != 5 {
		switch {
		case num1 != 5 && num2 != 5:
			a.Result[num1].GuiMoon = WonJaYu
			a.Result[num2].GuiMoon = WonJaYu
		case num1 == 5:
			if a.Ji[4].Title == "자" {
				a.Result[4].GuiMoon = WonJaYu
				a.Result[num2].GuiMoon = WonJaYu
			} else if a.Ji[5].Title == "자" {
				a.Result[5].GuiMoon = WonJaYu
				a.Result[num2].GuiMoon = WonJaYu
			}

			fallthrough
		case num2 == 5:
			if a.Ji[4].Title == "유" {
				a.Result[num1].GuiMoon = WonJaYu
				a.Result[4].GuiMoon = WonJaYu
			} else if a.Ji[5].Title == "유" {
				a.Result[num1].GuiMoon = WonJaYu
				a.Result[5].GuiMoon = WonJaYu
			}
		}
	}
}

func Find_Gyeokgak(a *Person) {
	//사묘축
	num1 := 5
	num2 := 5

	for i := 0; i < 4; i++ {
		if a.Ji[i].Title == "사" {
			num1 = i
		}
	}
	for j := 0; j < 4; j++ {
		if a.Ji[j].Title == "묘" {
			num2 = j
		}
	}
	a.Result[num1].GyeokGak = 0
	a.Result[num2].GyeokGak = 0

	if num1 != 5 || num2 != 5 {
		switch {
		case num1 != 5 && num2 != 5:
			a.Result[num1].GyeokGak = MyoSa
			a.Result[num2].GyeokGak = MyoSa
		case num1 == 5:
			if a.Ji[4].Title == "사" {
				a.Result[4].GyeokGak = MyoSa
				a.Result[num2].GyeokGak = MyoSa
			} else if a.Ji[5].Title == "사" {
				a.Result[5].GyeokGak = MyoSa
				a.Result[num2].GyeokGak = MyoSa
			}

			fallthrough
		case num2 == 5:
			if a.Ji[4].Title == "묘" {
				a.Result[num1].GyeokGak = MyoSa
				a.Result[4].GyeokGak = MyoSa
			} else if a.Ji[5].Title == "묘" {
				a.Result[num1].GyeokGak = MyoSa
				a.Result[5].GyeokGak = MyoSa
			}
		}
	}

	num1 = 5
	num2 = 5

	for i := 0; i < 4; i++ {
		if a.Ji[i].Title == "묘" {
			num1 = i
		}
	}
	for j := 0; j < 4; j++ {
		if a.Ji[j].Title == "축" {
			num2 = j
		}
	}

	if num1 != 5 || num2 != 5 {
		switch {
		case num1 != 5 && num2 != 5:
			a.Result[num1].GyeokGak = MyoChuk
			a.Result[num2].GyeokGak = MyoChuk
		case num1 == 5:
			if a.Ji[4].Title == "묘" {
				a.Result[4].GyeokGak = MyoChuk
				a.Result[num2].GyeokGak = MyoChuk
			} else if a.Ji[5].Title == "묘" {
				a.Result[5].GyeokGak = MyoChuk
				a.Result[num2].GyeokGak = MyoChuk
			}

			fallthrough
		case num2 == 5:
			if a.Ji[4].Title == "축" {
				a.Result[num1].GyeokGak = MyoChuk
				a.Result[4].GyeokGak = MyoChuk
			} else if a.Ji[5].Title == "축" {
				a.Result[num1].GyeokGak = MyoChuk
				a.Result[5].GyeokGak = MyoChuk
			}

		}
	}

	//신오진
	num1 = 5
	num2 = 5

	for i := 0; i < 4; i++ {
		if a.Ji[i].Title == "신" {
			num1 = i
		}
	}
	for j := 0; j < 4; j++ {
		if a.Ji[j].Title == "오" {
			num2 = j
		}
	}

	if num1 != 5 || num2 != 5 {
		switch {
		case num1 != 5 && num2 != 5:
			a.Result[num1].GyeokGak = OSin
			a.Result[num2].GyeokGak = OSin
		case num1 == 5:
			if a.Ji[4].Title == "신" {
				a.Result[4].GyeokGak = OSin
				a.Result[num2].GyeokGak = OSin
			} else if a.Ji[5].Title == "신" {
				a.Result[5].GyeokGak = OSin
				a.Result[num2].GyeokGak = OSin
			}

			fallthrough
		case num2 == 5:
			if a.Ji[4].Title == "오" {
				a.Result[num1].GyeokGak = OSin
				a.Result[4].GyeokGak = OSin
			} else if a.Ji[5].Title == "오" {
				a.Result[num1].GyeokGak = OSin
				a.Result[5].GyeokGak = OSin
			}
		}
	}

	num1 = 5
	num2 = 5

	for i := 0; i < 4; i++ {
		if a.Ji[i].Title == "오" {
			num1 = i
		}
	}
	for j := 0; j < 4; j++ {
		if a.Ji[j].Title == "진" {
			num2 = j
		}
	}

	if num1 != 5 || num2 != 5 {
		switch {
		case num1 != 5 && num2 != 5:
			a.Result[num1].GyeokGak = JinO
			a.Result[num2].GyeokGak = JinO
		case num1 == 5:
			if a.Ji[4].Title == "오" {
				a.Result[4].GyeokGak = JinO
				a.Result[num2].GyeokGak = JinO
			} else if a.Ji[5].Title == "오" {
				a.Result[5].GyeokGak = JinO
				a.Result[num2].GyeokGak = JinO
			}

			fallthrough
		case num2 == 5:
			if a.Ji[4].Title == "진" {
				a.Result[num1].GyeokGak = JinO
				a.Result[4].GyeokGak = JinO
			} else if a.Ji[5].Title == "진" {
				a.Result[num1].GyeokGak = JinO
				a.Result[5].GyeokGak = JinO
			}
		}
	}

	//해유미
	num1 = 5
	num2 = 5

	for i := 0; i < 4; i++ {
		if a.Ji[i].Title == "유" {
			num1 = i
		}
	}
	for j := 0; j < 4; j++ {
		if a.Ji[j].Title == "미" {
			num2 = j
		}
	}

	if num1 != 5 || num2 != 5 {
		switch {
		case num1 != 5 && num2 != 5:
			a.Result[num1].GyeokGak = MiYu
			a.Result[num2].GyeokGak = MiYu
		case num1 == 5:
			if a.Ji[4].Title == "유" {
				a.Result[4].GyeokGak = MiYu
				a.Result[num2].GyeokGak = MiYu
			} else if a.Ji[5].Title == "유" {
				a.Result[5].GyeokGak = MiYu
				a.Result[num2].GyeokGak = MiYu
			}

			fallthrough
		case num2 == 5:
			if a.Ji[4].Title == "미" {
				a.Result[num1].GyeokGak = MiYu
				a.Result[4].GyeokGak = MiYu
			} else if a.Ji[5].Title == "미" {
				a.Result[num1].GyeokGak = MiYu
				a.Result[5].GyeokGak = MiYu
			}
		}
	}

	num1 = 5
	num2 = 5

	for i := 0; i < 4; i++ {
		if a.Ji[i].Title == "유" {
			num1 = i
		}
	}
	for j := 0; j < 4; j++ {
		if a.Ji[j].Title == "해" {
			num2 = j
		}
	}

	if num1 != 5 || num2 != 5 {
		switch {
		case num1 != 5 && num2 != 5:
			a.Result[num1].GyeokGak = YuHae
			a.Result[num2].GyeokGak = YuHae
		case num1 == 5:
			if a.Ji[4].Title == "유" {
				a.Result[4].GyeokGak = YuHae
				a.Result[num2].GyeokGak = YuHae
			} else if a.Ji[5].Title == "유" {
				a.Result[5].GyeokGak = YuHae
				a.Result[num2].GyeokGak = YuHae
			}

			fallthrough
		case num2 == 5:
			if a.Ji[4].Title == "해" {
				a.Result[num1].GyeokGak = YuHae
				a.Result[4].GyeokGak = YuHae
			} else if a.Ji[5].Title == "해" {
				a.Result[num1].GyeokGak = YuHae
				a.Result[5].GyeokGak = YuHae
			}
		}
	}

	//인자술
	num1 = 5
	num2 = 5

	for i := 0; i < 4; i++ {
		if a.Ji[i].Title == "자" {
			num1 = i
		}
	}
	for j := 0; j < 4; j++ {
		if a.Ji[j].Title == "인" {
			num2 = j
		}
	}

	if num1 != 5 || num2 != 5 {
		switch {
		case num1 != 5 && num2 != 5:
			a.Result[num1].GyeokGak = InJa
			a.Result[num2].GyeokGak = InJa
		case num1 == 5:
			if a.Ji[4].Title == "자" {
				a.Result[4].GyeokGak = InJa
				a.Result[num2].GyeokGak = InJa
			} else if a.Ji[5].Title == "자" {
				a.Result[5].GyeokGak = InJa
				a.Result[num2].GyeokGak = InJa
			}

			fallthrough
		case num2 == 5:
			if a.Ji[4].Title == "인" {
				a.Result[num1].GyeokGak = InJa
				a.Result[4].GyeokGak = InJa
			} else if a.Ji[5].Title == "인" {
				a.Result[num1].GyeokGak = InJa
				a.Result[5].GyeokGak = InJa
			}
		}
	}

	num1 = 5
	num2 = 5

	for i := 0; i < 4; i++ {
		if a.Ji[i].Title == "자" {
			num1 = i
		}
	}
	for j := 0; j < 4; j++ {
		if a.Ji[j].Title == "술" {
			num2 = j
		}
	}

	if num1 != 5 || num2 != 5 {
		switch {
		case num1 != 5 && num2 != 5:
			a.Result[num1].GyeokGak = SulJa
			a.Result[num2].GyeokGak = SulJa
		case num1 == 5:
			if a.Ji[4].Title == "자" {
				a.Result[4].GyeokGak = SulJa
				a.Result[num2].GyeokGak = SulJa
			} else if a.Ji[5].Title == "자" {
				a.Result[5].GyeokGak = SulJa
				a.Result[num2].GyeokGak = SulJa
			}

			fallthrough
		case num2 == 5:
			if a.Ji[4].Title == "술" {
				a.Result[num1].GyeokGak = SulJa
				a.Result[4].GyeokGak = SulJa
			} else if a.Ji[5].Title == "술" {
				a.Result[num1].GyeokGak = SulJa
				a.Result[5].GyeokGak = SulJa
			}
		}
	}
}

func Fing_AmHab(a *Person) *Person {

	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			a.Result[i].AmHab = 0
			a.Result[j].AmHab = 0
			if i != j {
				switch a.Ji[i].Title {
				case "인":
					switch a.Ji[j].Title {
					case "유":
						a.Result[i].AmHab = AmHabValue
						a.Result[j].AmHab = AmHabValue
					case "술":
						a.Result[i].AmHab = AmHabValue
						a.Result[j].AmHab = AmHabValue
					case "축":
						a.Result[i].AmHab = AmHabValue
						a.Result[j].AmHab = AmHabValue
					default:
						a.Result[i].AmHab = 0
						a.Result[j].AmHab = 0
					}
				case "묘":
					switch a.Ji[j].Title {
					case "사":
						a.Result[i].AmHab = AmHabValue
						a.Result[j].AmHab = AmHabValue
					case "신":
						a.Result[i].AmHab = AmHabValue
						a.Result[j].AmHab = AmHabValue
					default:
						a.Result[i].AmHab = 0
						a.Result[j].AmHab = 0
					}
				case "진":
					switch a.Ji[j].Title {
					case "사":
						a.Result[i].AmHab = AmHabValue
						a.Result[j].AmHab = AmHabValue
					default:
						a.Result[i].AmHab = 0
						a.Result[j].AmHab = 0
					}
				case "사":
					switch a.Ji[j].Title {
					case "묘":
						a.Result[i].AmHab = AmHabValue
						a.Result[j].AmHab = AmHabValue
					case "진":
						a.Result[i].AmHab = AmHabValue
						a.Result[j].AmHab = AmHabValue
					case "자":
						a.Result[i].AmHab = AmHabValue
						a.Result[j].AmHab = AmHabValue
					case "축":
						a.Result[i].AmHab = AmHabValue
						a.Result[j].AmHab = AmHabValue
					default:
						a.Result[i].AmHab = 0
						a.Result[j].AmHab = 0
					}
				case "오":
					switch a.Ji[j].Title {
					case "신":
						a.Result[i].AmHab = AmHabValue
						a.Result[j].AmHab = AmHabValue
					case "해":
						a.Result[i].AmHab = AmHabValue
						a.Result[j].AmHab = AmHabValue
					default:
						a.Result[i].AmHab = 0
						a.Result[j].AmHab = 0
					}
				case "미":
					switch a.Ji[j].Title {
					case "신":
						a.Result[i].AmHab = AmHabValue
						a.Result[j].AmHab = AmHabValue
					case "유":
						a.Result[i].AmHab = AmHabValue
						a.Result[j].AmHab = AmHabValue
					default:
						a.Result[i].AmHab = 0
						a.Result[j].AmHab = 0
					}
				case "신":
					switch a.Ji[j].Title {
					case "묘":
						a.Result[i].AmHab = AmHabValue
						a.Result[j].AmHab = AmHabValue
					case "오":
						a.Result[i].AmHab = AmHabValue
						a.Result[j].AmHab = AmHabValue
					case "미":
						a.Result[i].AmHab = AmHabValue
						a.Result[j].AmHab = AmHabValue
					default:
						a.Result[i].AmHab = 0
						a.Result[j].AmHab = 0
					}
				case "유":
					switch a.Ji[j].Title {
					case "인":
						a.Result[i].AmHab = AmHabValue
						a.Result[j].AmHab = AmHabValue
					default:
						a.Result[i].AmHab = 0
						a.Result[j].AmHab = 0
					}
				case "술":
					switch a.Ji[j].Title {
					case "해":
						a.Result[i].AmHab = AmHabValue
						a.Result[j].AmHab = AmHabValue
					default:
						a.Result[i].AmHab = 0
						a.Result[j].AmHab = 0
					}
				case "해":
					switch a.Ji[j].Title {
					case "오":

						a.Result[i].AmHab = AmHabValue
						a.Result[j].AmHab = AmHabValue
					case "미":

						a.Result[i].AmHab = AmHabValue
						a.Result[j].AmHab = AmHabValue
					case "술":

						a.Result[i].AmHab = AmHabValue
						a.Result[j].AmHab = AmHabValue
					default:
						a.Result[i].AmHab = 0
						a.Result[j].AmHab = 0
					}
				case "자":
					switch a.Ji[j].Title {
					case "인":

						a.Result[i].AmHab = AmHabValue
						a.Result[j].AmHab = AmHabValue
					case "사":

						a.Result[i].AmHab = AmHabValue
						a.Result[j].AmHab = AmHabValue
					default:
						a.Result[i].AmHab = 0
						a.Result[j].AmHab = 0
					}
				case "축":
					switch a.Ji[j].Title {
					case "인":

						a.Result[i].AmHab = AmHabValue
						a.Result[j].AmHab = AmHabValue
					case "사":

						a.Result[i].AmHab = AmHabValue
						a.Result[j].AmHab = AmHabValue
					default:
						a.Result[i].AmHab = 0
						a.Result[j].AmHab = 0
					}
				}
			}
		}
	}

	return a
}

func Find_Characteristics(host *Person) *Person {
	Find_Banghab(host)
	Find_Samhab(host)
	Find_Yukhab(host)
	Find_Hyungsal(host)
	Find_Choongsal(host)
	Find_Pasal(host)
	Find_Haesal(host)
	Find_Wonzin(host)
	Find_Guimun(host)
	Find_Gyeokgak(host)
	return host
}
