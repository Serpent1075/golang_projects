package saju

type SajuAnalyzer struct {
	Chungan    []Chungan
	Jiji       []Jiji
	Manse      []YearGanji
	Julgys     []YearJulgy
	SaeUn      []SaeUn
	Sibsung    []*Sibsung
	Sib2Unsung []*Chungan_Unsung
}

type Sibsung struct {
	Prop      string
	Comp_Prop []Compare_Prop
}

type Compare_Prop struct {
	Comp_Prop string
	Title     string
}

type Person struct {
	LoginID string
	Chun    []Chungan
	Ji      []Jiji
	Result  []Result_record
}

type Result_record struct {
	ChunGanHab  int32 `json:"ChunGanHab" bson:"chunganhab"`
	ChunGanGeok int32 `json:"ChunGanGeok" bson:"chungangeok"`
	Sibsung     int32 `json:"Sibsung" bson:"sibsung"`
	YukHab      int32 `json:"YukHab" bson:"yukhab"`
	SamHab      int32 `json:"SamHab" bson:"samhab"`
	//BanHab      BanHab      `json:"BanHab" bson:"banhab"`
	BangHab int32 `json:"BangHab" bson:"banghab"`
	AmHab   int32 `json:"AmHab" bson:"amhab"`
	//MyeongAmHab MyeongAmHab `json:"MyeongAmHab"`
	WonJin   int32 `json:"WonJin" bson:"wonjin"`
	GuiMoon  int32 `json:"GuiMoon" bson:"guimoon"`
	Hyung    int32 `json:"Hyung" bson:"hyung"`
	Choong   int32 `json:"Choong" bson:"choong"`
	Pa       int32 `json:"Pa" bson:"pa"`
	Hae      int32 `json:"Hae" bson:"hae"`
	GyeokGak int32 `json:"GyeokGak" bson:"geokgak"`
	IpMyo    int32 `json:"IpMyo" bson:"ipmyo"`
}

type IpMyo struct {
	Exist   bool
	WhichJi int
}

type ChunGanHab struct {
	Exist     bool
	GabGi     int
	ElGyeong  int
	ByeongSin int
	JeongIm   int
	MuGye     int
}

type ChunGanGeok struct {
	Exist bool

	GabGyeong    int
	ElSin        int
	ByeongIm     int
	JeongHye     int
	MuGab        int
	GiEl         int
	GyeongByeong int
	SinJeong     int
	ImMu         int
	GyeGi        int
}

type SibsungResult struct {
	Exist   bool
	Sibsung int
}

type YukHab struct {
	Exist  bool
	InHye  int
	MyoSul int
	JinYu  int
	SaSin  int
	OMi    int
	JaChuk int
}

type SamHab struct {
	Exist    bool
	InOSul   int
	HaeMyoMi int
	SinJaJin int
	SaYuChuk int
}

type BanHab struct {
	Exist  bool
	InO    int
	InSul  int
	OSul   int
	HaeMyo int
	MyoMi  int
	HaeMi  int
	SinJa  int
	JaJin  int
	SinJin int
	SaYu   int
	YuChuk int
	SaChuk int
}

type BangHab struct {
	Exist     bool
	InMyoJin  int
	SaOMi     int
	SinYuSul  int
	HaeJaChuk int
}

type AmHab struct {
	Exist bool
	AmHab int
}

type MyeongAmHab struct {
	Exist     bool
	GabGi     int
	ElGyeong  int
	ByeongSin int
	JeongIm   int
	MuGye     int
}

type WonJin struct {
	Exist  bool
	InYu   int
	MyoSin int
	JinHae int
	SaSul  int
	OChuk  int
	JaMi   int
}

type GuiMoon struct {
	Exist bool
	InMi  int
	JaYu  int
}

type Hyung struct {
	Exist   bool
	InSa    int
	JaMyo   int
	JinJin  int
	SaSin   int
	OO      int
	YuYu    int
	SulMi   int
	ChukSul int
	HaeHae  int
}

type Choong struct {
	Exist  bool
	InSin  int
	MyoYu  int
	JinSul int
	SaHae  int
	JaO    int
	ChukMi int
}

type Pa struct {
	Exist   bool
	InHae   int
	MyoO    int
	JinChuk int
	SaSin   int
	SulMi   int
	JaYu    int
}

type Haesal struct {
	Exist  bool
	InSa   int
	MyoJin int
	OChuk  int
	JaMi   int
	HaeSin int
	YuSul  int
}

type GyeokGak struct {
	Exist   bool
	InJa    int
	MyoSa   int
	MyoChuk int
	JinO    int
	OSin    int
	MiYu    int
	YuHae   int
	SulJa   int
}

type SajuTable struct {
	Chungan []Chungan
	Jiji    []Jiji
	Manse   []YearGanji
	Julgys  []YearJulgy
	SaeUn   []SaeUn
}

type Chungan struct {
	Id         int
	Title      string
	Properties Property_Chun
}

type Property_Chun struct {
	Umyang         int
	Prop           string
	Sibsung        string
	IpMyo          int32
	Unsung_Me      Unsung_Me
	Unsung_by_Jiji Unsung_by_Jiji
}

type Unsung_Me struct {
	level        int
	Unsung_title string
}

type Unsung_by_Jiji struct {
	level        int
	Unsung_title string
}

type Chungan_Unsung struct {
	Title      string
	Properties []Property_Unsung
}

type Property_Unsung struct {
	Level     int
	Jiji_char string
	Prop      string
}

type Jiji struct {
	Id         int
	Title      string
	Properties Property_Ji
}

type Property_Ji struct {
	Umyang     int
	Prop       string
	Sibsung    string
	Hae        int
	Go         int
	Ji         int
	ChangGo    int
	Jijanggans []Jijanggan
}

type ChangGo struct {
	Exist     bool
	InSungGo  int
	YangInGo  int
	JaeGo     int
	GwanGo    int
	SikSangGo int
}

type Jijanggan struct {
	Chungan_char Chungan_Jijanggan
	Value_len    int
}

type Chungan_Jijanggan struct {
	Id         int
	Title      string
	Properties Property_Chun_Jijanggan
}

type Property_Chun_Jijanggan struct {
	Umyang int
	Prop   string
}

type YearGanji struct {
	WhichYear     int
	Chungan_Title string
	Jiji_Title    string
	MonthGanji    []MonthGanji
}

type MonthGanji struct {
	WhichMonth    int
	Chungan_Title string
	Jiji_Title    string
	Day_Ganji     []Day_Ganji
}

type Day_Ganji struct {
	WhichDay      int
	Chungan_Title string
	Jiji_Title    string
	Time_Ganji    []Time_Ganji
}

type Time_Ganji struct {
	Chungan_Title string
	Jiji_Title    string
}

type YearJulgy struct {
	Year        int
	MonthJulgys []MonthJulgy
}

type MonthJulgy struct {
	Month     int
	DayJulgys []DayJulgy
}

type DayJulgy struct {
	Day   int
	Title string
}

type SaeUn struct {
	Year        int
	SaeUnGanjis SaeUnGanji
}

type SaeUnGanji struct {
	Chun string
	Ji   string
}

type Saju struct {
	Gender bool
	Year   int    `db:"urmyyear"`
	Month  int    `db:"urmymonth"`
	Day    int    `db:"urmyday"`
	Time   string `db:"urmytime"`
}

type SaJuPalJa struct {
	YearChun  string `db:"yearchun"`
	YearJi    string `db:"yearji"`
	MonthChun string `db:"monthchun"`
	MonthJi   string `db:"monthji"`
	DayChun   string `db:"daychun"`
	DayJi     string `db:"dayji"`
	TimeChun  string `db:"timechun"`
	TimeJi    string `db:"timeji"`
}

type DaeSaeUn struct {
	DaeUnChun string `db:"daeunchun"`
	DaeUnJi   string `db:"daeunji"`
	SaeUnChun string `db:"saeunchun"`
	SaeUnJi   string `db:"saeunji"`
}
