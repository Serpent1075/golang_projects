package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"hash/fnv"
	"io"

	"log"
	"net/http"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"

	logger "github.com/whitewhale1075/urmylogger"

	firebase "firebase.google.com/go"
	//"firebase.google.com/go/messaging"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	uuid "github.com/uuid6/uuid6go-proto"
	country "github.com/whitewhale1075/urmycountry"
	saju "github.com/whitewhale1075/urmysaju"
	"google.golang.org/api/option"

	_ "github.com/lib/pq"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gopkg.in/redis.v5"
)

//const version = "1.0.0"

var writedb *sqlx.DB
var readdb *sqlx.DB
var rdb *redis.Client
var mongoconn *mongo.Client

var firebaseapp *firebase.App
var compserviceid = "s10001"

//var pathofserviceky string = "./urmytalk-prod-nul5o-aa697a30c0.json"

var pathofserviceky string
var mySigningKey = []byte("ALSKJEN12389NAjnashj101j9snvaahLWIDhuwlKSNDKSITROLAKShQPASKDB12687")
var s3client *s3.Client
var applesecret string
var cfurl string = "https://contents.urmycorp.com/"

type Environment struct {
	Readhost     string
	Readport     int
	Readuser     string
	ReadPassword string
	ReadDBName   string
	RedisAddr    string
	RedisPass    string
	MongoUrl     string
}

var env Environment

type AppHandler struct {
	sa *saju.SajuAnalyzer
	co *country.CountryList
}

var a *AppHandler

type SecretValue struct {
	CFurl          string `json:"cfurl"`
	RedisAddr      string `json:"redis-addr"`
	RedisPass      string `json:"redis-pass"`
	RdbmsHost      string `json:"rdbms-host"`
	RdbmsPort      string `json:"rdbms-port"`
	RdbmsUser      string `json:"rdbms-user"`
	RdbmsPassword  string `json:"rdbms-password"`
	RdbmsDBName    string `json:"rdbms-dbname"`
	PremiumService string `json:"premium-service"`
	ConfigPath     string `json:"config-path"`
}

const urmyconfigfilepath = "/data" // /data .

func getenv() {

	env.Readhost = "127.0.0.1"
	env.Readport = 5432
	env.Readuser = "koreaogh"
	env.ReadPassword = "test1234"
	env.RedisAddr = "127.0.0.1:6379"
	env.ReadDBName = "urmydb"
	env.RedisPass = "qwer1234"
	env.MongoUrl = "mongodb://127.0.0.1:27017"

	/*
		env.Readhost = "192.168.92.130"
		env.Readport = 5432
		env.Readuser = "koreaogh"
		env.ReadPassword = "test1234"
		env.RedisAddr = "192.168.92.130:6379"
		env.ReadDBName = "urmydb"
		env.RedisPass = "qwer1234"
		env.MongoUrl = "mongodb://192.168.92.130:27017"
	*/
}

func init() {
	pathofserviceky = urmyconfigfilepath + "/config/urmytalk-prod-nul5o-aa697a30c0.json"
	getenv()

	var err error
	s3client = s3.NewFromConfig(*GetConfig())
	a = &AppHandler{
		sa: saju.NewSaJuAnalyzer(urmyconfigfilepath),
		co: country.NewCountrySelector(urmyconfigfilepath),
	}
	firebaseapp, err = initFirebase()
	if err != nil {
		logger.Error("Error", logger.String("initfirebase", err.Error()))
		return
	}

	var tops = []logger.TeeOption{
		{
			Filename: urmyconfigfilepath + "/logs/info.log",
			Ropt: logger.RotateOptions{
				MaxSize:    1,
				MaxAge:     1,
				MaxBackups: 3,
				Compress:   true,
			},
			Lef: func(lvl logger.Level) bool {
				return lvl <= logger.InfoLevel
			},
		},
		{
			Filename: urmyconfigfilepath + "/logs/error.log",
			Ropt: logger.RotateOptions{
				MaxSize:    1,
				MaxAge:     1,
				MaxBackups: 3,
				Compress:   true,
			},
			Lef: func(lvl logger.Level) bool {
				return lvl > logger.InfoLevel
			},
		},
	}
	logging := logger.NewTeeWithRotate(tops)
	logger.ResetDefault(logging)
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	fmt.Println(runtime.GOMAXPROCS(0))
	ConnectToDB()
	ConnectToRedis()
	MongoConn()
	defer readdb.Close()
	defer writedb.Close()
	defer rdb.Close()
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.Use(JSONMiddleware())

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	v1 := r.Group("/v1")
	{

		//v1.POST("/signin", signin)
		v1.POST("/checkneedinsert", checkNeedInsert)
		v1.POST("/getcfurl", getCFUrl)

	}

	v2 := r.Group("/v2")
	v2.Use(FirebaseAuthTokenMiddleware)
	v2.Use(TokenAuthMiddleware)
	{
		v2.POST("/getudata", getudata)
		v2.POST("/insertpropic", insertpropic)
		v2.POST("/service", Service)
		v2.POST("/signout", signout)
		v2.POST("/deleteuser", deleteUser)

		//////UpdateProfile////////
		v2.POST("/updateprofile", updateInfo)
		v2.POST("/sendmessage", sendMessage)

	}
	v3 := r.Group("/v3")
	v3.Use(TokenAuthMiddleware)
	{
		v3.POST("/listcandidatelist", ListCandidate)
		v3.POST("/listfriendlist", ListFriendList)
		v3.POST("/addfriendlist", AddFriendList)
		v3.POST("/addanonymousfriendlist", AddAnonymousFriendList)
		v3.POST("/deletefriendlist", DeleteFriendList)
		v3.POST("/listblacklist", ListBlackList)
		v3.POST("/addblacklist", AddBlackList)
		v3.POST("/deleteblacklist", DeleteBlackList)
		v3.POST("/listservice", ListService)
		v3.POST("/purchaseservice", PurchaseService)
		v3.POST("/refundrequest", RefundRequest)
		v3.POST("/numberofoppositesex", NumberOfOppositeSex)
		v3.POST("/numberofurmy", NumberOfUrMy)

		v3.POST("/insertcandidatereport", InsertCandidateReport)
		v3.POST("/insertchatreport", InsertChatReport)
		v3.POST("/insertinquiry", InsertInquiry)
		v3.POST("/insertinquirycontent", InsertInquiryCotent)
		v3.POST("/getinquiry", GetInquiry)
		v3.POST("/getinquirycontent", GetInquiryContent)

		v3.POST("/updateprofilepic", updateProfilePicture)
		v3.POST("/deleteprofilepic", deleteProfilePicture)
	}
	v4 := r.Group("/v4")
	v4.Use(FirebaseAuthTokenMiddleware)
	{

		v4.POST("/signup", signup)
	}
	r.Run(":3010")
}

func Service(c *gin.Context) {
	fmt.Println("ping")
}

func JSONMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, post-check=0")
		c.Writer.Header().Set("Last-Modified", time.Now().String())
		c.Writer.Header().Set("Pragma", "no-cache")
		c.Writer.Header().Set("Expires", "-1")
		c.Writer.Header().Set("Content-Type", "application/json")
		c.Next()
	}
}

func FirebaseAuthTokenMiddleware(c *gin.Context) {

	ctx := context.TODO()
	firebasetoken := c.Request.Header.Get("firebasetoken")
	client, firebaseappAutherr := firebaseapp.Auth(ctx)
	if firebaseappAutherr != nil {
		logger.Error("Error", logger.String("firebasetokenerr", firebaseappAutherr.Error()), logger.String("path", c.FullPath()))
		c.AbortWithStatus(http.StatusUnauthorized)
	}
	_, verifyidtokenerr := client.VerifyIDToken(ctx, firebasetoken)
	if verifyidtokenerr != nil {
		logger.Error("Error", logger.String("firebasetokenerr", verifyidtokenerr.Error()), logger.String("path", c.FullPath()))
		c.AbortWithStatus(http.StatusUnauthorized)
	}

	//log.Printf("Verified ID token: %v\n", token)
	c.Set("firebasetoken", firebasetoken)
	c.Next()
}

func TokenAuthMiddleware(c *gin.Context) {
	tokenAuth, tokenAutherr := ExtractAccessTokenMetadata(c.Request)
	if tokenAutherr != nil {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	userId, fetchaccessautherr := FetchAccessAuth(tokenAuth)
	//log.Println("TokenAuthMiddle: " + userId)
	if fetchaccessautherr != nil {
		log.Println(fetchaccessautherr.Error())
		if fetchaccessautherr == jwt.ErrSignatureInvalid {
			logger.Error("Error", logger.String("fetchaccessautherr", tokenAutherr.Error()), logger.String("path", c.Request.URL.Path))
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"message": "fetchaccessautherr",
			})

		} else {
			logger.Error("Error", logger.String("fetchaccessautherr", tokenAutherr.Error()))
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

	}
	//log.Printf("userid: %s", userId)
	c.Set("userId", userId)
	c.Next()

}

//////////////////// Firebase Send Message //////////////////

func sendMessage(c *gin.Context) {
	ctx := context.Background()
	urmyuuid := c.MustGet("userId").(string)
	if threatable := checkTheValueThreatable(urmyuuid); threatable {
		logger.Error("Error", logger.String("uuid", urmyuuid), logger.String("threatable", c.ClientIP()))
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	var sendmsgerr error
	var message SendMessage
	if sendmsgerr := c.ShouldBindJSON(&message); sendmsgerr != nil {
		logger.Error("Error", logger.String("uuid", urmyuuid), logger.String("bind", sendmsgerr.Error()))
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	ok, sendmsgerr := CheckBlackList(urmyuuid, message.MsgUser)
	if sendmsgerr != nil {
		logger.Error("Error", logger.String("sendmsgerr", sendmsgerr.Error()), logger.String("uuid", urmyuuid))
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	client, sendmsgerr := firebaseapp.Firestore(ctx)
	if sendmsgerr != nil {
		logger.Error("Error", logger.String("sendmsgerr", sendmsgerr.Error()), logger.String("uuid", urmyuuid))
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	if !ok {
		_, sendmsgerr = client.Collection("chatroomId").Doc(message.ChatroomId).Collection(message.ChatroomId).Doc(strconv.Itoa(message.MsgSentdate)).Set(ctx, map[string]interface{}{
			"sender":       message.MsgSender,
			"receiver":     message.MsgUser,
			"msgtimestamp": message.MsgSentdate,
			"msgcontent":   message.MsgContent,
			"msgtype":      message.MsgType,
			"isread":       false,
		})
		if sendmsgerr != nil {
			// Handle any errors in an appropriate way, such as returning them.
			logger.Error("Error", logger.String("sendmsgerr", sendmsgerr.Error()), logger.String("uuid", urmyuuid))
			c.AbortWithStatus(http.StatusNotFound)
			return
		}

		_, sendmsgerr = client.Collection("user").Doc(message.MsgUser).Collection("chatroomId").Doc(message.ChatroomId).Set(ctx, map[string]interface{}{
			"chatroomId":   message.ChatroomId,
			"user":         message.MsgSender,
			"lastChat":     message.MsgContent,
			"msgtimestamp": message.MsgSentdate,
			"msgtype":      message.MsgType,
		})
		if sendmsgerr != nil {
			// Handle any errors in an appropriate way, such as returning them.
			logger.Error("Error", logger.String("sendmsgerr", sendmsgerr.Error()), logger.String("uuid", urmyuuid))
			c.AbortWithStatus(http.StatusNotFound)
			return
		}
	}

	_, sendmsgerr = client.Collection("user").Doc(message.MsgSender).Collection("chatroomId").Doc(message.ChatroomId).Set(ctx, map[string]interface{}{
		"chatroomId":   message.ChatroomId,
		"user":         message.MsgUser,
		"lastChat":     message.MsgContent,
		"msgtimestamp": message.MsgSentdate,
		"msgtype":      message.MsgType,
	})
	if sendmsgerr != nil {
		// Handle any errors in an appropriate way, such as returning them.
		logger.Error("Error", logger.String("sendmsgerr", sendmsgerr.Error()), logger.String("uuid", urmyuuid))
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	/*
		var sendmessage SendMessage
		if err := c.ShouldBindJSON(&sendmessage); err != nil {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

		registrationToken, getnotificationtokenerr := getUserNotificationToken(c, sendmessage.MsgUser)
		if getnotificationtokenerr != nil {
			c.JSON(http.StatusUnprocessableEntity, "Invalid json provided")
			log.Println("Invalid json provided")
			return

		}

		message := &messaging.Message{
			Data: map[string]string{
				"content": sendmessage.MsgContent,
				"time":    sendmessage.MsgSentdate,
				"sender":  sendmessage.MsgSender,
			},

			Token: registrationToken,
		}

		client, err := firebaseapp.Messaging(ctx)
		if err != nil {
			log.Fatalf("error getting Messaging client: %v\n", err)
		}

		response, err := client.Send(ctx, message)
		if err != nil {
			log.Fatalln(err)
		}
	*/
	c.JSON(http.StatusOK, "Message sent")
}

func CheckPremiumUser(premiumserviceid string, userid string) bool {
	var premium []string
	err := readdb.Select(&premium, "SELECT regexp_replace(purchaseuser, '\\s+$','') FROM urmysales WHERE serviceid=$1 AND purchasedate between current_timestamp(0)+'-1 months' and current_timestamp(0)", premiumserviceid)
	if err != nil {
		log.Printf("select my data error %s", err.Error())
	}

	for _, data := range premium {
		if strings.Contains(data, userid) {
			return true
		}
	}
	return false
}

/*
	func getUserNotificationToken(ctx context.Context, userId string) (string, error) {
		notificationconn := mongoconn.Database("urmyusers").Collection("usersignin")

		var signinUser SignInUser
		notifinderr := notificationconn.FindOne(ctx, bson.M{"useruid": strings.TrimSpace(userId)}).Decode(&signinUser)
		if notifinderr != nil {
			return "", WrapError("getunotitokfinderr", notifinderr.Error())
		}
		return strings.TrimSpace(signinUser.FirebaseNotificationToken), nil

}
*/
func deleteFirebaseUser(ctx context.Context, useruid string) error {
	client, err := firebaseapp.Auth(ctx)
	if err != nil {
		return WrapError("deleteFirebaseUser", err.Error())
	}
	err = client.DeleteUser(ctx, useruid)
	if err != nil {
		return WrapError("deleteFirebaseUser", err.Error())
	}
	return nil
}

func checkFirebaseUserEmail(ctx context.Context, email string, useruid string) bool {
	client, err := firebaseapp.Auth(ctx)
	if err != nil {
		log.Fatalf("error getting Auth client: %v\n", err)
	}

	u, err := client.GetUserByEmail(ctx, email)
	if err != nil {
		log.Fatalf("error getting user by email %s: %v\n", email, err)
	}

	if u.UID == useruid {
		return true
	} else {
		return false
	}
}

func initFirebase() (*firebase.App, error) {
	opt := option.WithCredentialsFile(pathofserviceky)
	app, initfberr := firebase.NewApp(context.Background(), nil, opt)
	if initfberr != nil {
		return nil, WrapError("initfberr ", initfberr.Error())
	}
	return app, nil
}

////////////////////   AUTH      /////////////////////////////

func deleteUser(c *gin.Context) {
	var u SignUpUser
	if delbuserbinderr := c.ShouldBindJSON(&u); delbuserbinderr != nil {
		logger.Error("Error", logger.String("delbuserbinderr", delbuserbinderr.Error()), logger.String("uuid", u.Useruid))
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	urmyuuid := c.MustGet("userId").(string)                       //uuid generated by apiserver
	delfberr := deleteFirebaseUser(context.Background(), urmyuuid) //uuid from firebase
	if delfberr != nil {
		logger.Error("Error", logger.String("delfberr", delfberr.Error()), logger.String("uuid", urmyuuid))
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	uinfo, getusererr := GetUserDataToDelete(urmyuuid)
	if getusererr != nil {
		logger.Error("Error", logger.String("getusererr", getusererr.Error()), logger.String("uuid", urmyuuid))
	}

	delerr := DeleteUrMyUser(urmyuuid, uinfo)
	if delerr != nil {
		logger.Error("Error", logger.String("delerr", delerr.Error()), logger.String("uuid", urmyuuid))
	}
	c.JSON(http.StatusOK, gin.H{})
}

func updateInfo(c *gin.Context) {
	var updateprofile UpdateProfile
	var upinfoerr error
	urmyuuid := c.MustGet("userId").(string)
	if threatable := checkTheValueThreatable(urmyuuid); threatable {
		logger.Error("Error", logger.String("uuid", urmyuuid), logger.String("threatable", c.ClientIP()))
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	if upinfoerr = c.ShouldBindJSON(&updateprofile); upinfoerr != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	result, upinfoerr := UpdateUrMyProfile(urmyuuid, &updateprofile)
	if upinfoerr != nil {
		if upinfoerr.Error() == "upuinfoupdatablerr: notabletochangebirthdate" {
			logger.Error("Error", logger.String("upinfoerr", upinfoerr.Error()), logger.String("uuid", urmyuuid))
			c.AbortWithStatus(http.StatusNotAcceptable)
			return
		}
		logger.Error("Error", logger.String("upinfoerr", upinfoerr.Error()), logger.String("uuid", urmyuuid))
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	} else {
		ExecRefreshMaterializedView()
	}

	if result {
		userdata, upinfoerr := GetUrMyUserData(urmyuuid)
		if upinfoerr != nil {
			logger.Error("Error", logger.String("upinfoerr", upinfoerr.Error()), logger.String("uuid", urmyuuid))
			c.AbortWithStatus(http.StatusNotFound)
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"urdata": userdata,
		})

	}
}

func updateProfilePicture(c *gin.Context) {
	var upropicerr error
	urmyuuid, _ := c.MustGet("userId").(string)

	file, header, upropicerr := c.Request.FormFile("profile")
	if upropicerr != nil {
		logger.Error("Error", logger.String("upropicerr", upropicerr.Error()), logger.String("uuid", urmyuuid))
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	defer file.Close()
	// 파일 저장

	// 방법 1.
	// 기본 제공 함수로 파일 저장
	//file, _ := c.FormFile("file")
	//c.SaveUploadedFile(file, filepath.Join("./uploaded", file.Filename))
	//c.String(http.StatusOK, fmt.Sprintf("'%s' uploaded!", file.Filename))
	buffer := make([]byte, header.Size)
	file.Read(buffer)
	var tmpfilename string
	//detectedcontenttype := http.DetectContentType(buffer)
	//fmt.Println(detectedcontenttype)

	if strings.Contains(header.Filename, ".jpeg") || strings.Contains(header.Filename, ".jpg") {
		tmpfilename = time.Now().Format("2006-01-02T15:04:05") + ".jpg"
	} else if strings.Contains(header.Filename, ".png") {
		tmpfilename = time.Now().Format("2006-01-02T15:04:05") + ".png"
	} else {
		logger.Error("Error", logger.String("upropicerr", upropicerr.Error()), logger.String("uuid", urmyuuid))
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	c.SaveUploadedFile(header, filepath.Join("./uploaded.jpg"))
	upropicerr = PutProfilePicS3(buffer, urmyuuid, tmpfilename)
	if upropicerr != nil {
		logger.Error("Error", logger.String("upropicerr", upropicerr.Error()), logger.String("uuid", urmyuuid))
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	_, upropicerr = UpdateProfilePicData(urmyuuid, tmpfilename)
	if upropicerr != nil {
		logger.Error("Error", logger.String("upropicerr", upropicerr.Error()), logger.String("uuid", urmyuuid))
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	result, upropicerr := GetUrMyUserProfilePicture(urmyuuid)
	if upropicerr != nil {
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"profilepiclist": result,
	})
}

func deleteProfilePicture(c *gin.Context) {
	var delpropicerr error
	urmyuuid := c.MustGet("userId").(string)
	if threatable := checkTheValueThreatable(urmyuuid); threatable {
		logger.Error("Error", logger.String("uuid", urmyuuid), logger.String("threatable", c.ClientIP()))
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	var filerequest UpdateProfile
	if delpropicerr = c.ShouldBindJSON(&filerequest); delpropicerr != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	_, delpropicerr = DeleteProfilePicData(urmyuuid, filerequest.Value)
	if delpropicerr != nil {
		logger.Error("Error", logger.String("delpropicerr", delpropicerr.Error()), logger.String("uuid", urmyuuid))
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	delpropicerr = DeleteProfilePicS3(urmyuuid, filerequest.Value)
	if delpropicerr != nil {
		logger.Error("Error", logger.String("delpropicerr", delpropicerr.Error()), logger.String("uuid", urmyuuid))
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	result, delpropicerr := GetUrMyUserProfilePicture(urmyuuid)
	if delpropicerr != nil {

		logger.Error("Error", logger.String("delpropicerr", delpropicerr.Error()), logger.String("uuid", urmyuuid))
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"profilepiclist": result,
	})
}

func checkNeedInsert(c *gin.Context) {
	var u SignUpUser
	var checkneedinserterr error
	if checkneedinserterr = c.ShouldBindJSON(&u); checkneedinserterr != nil {
		logger.Error("Error", logger.String("checkneedinserterr", checkneedinserterr.Error()), logger.String("uuid", u.Useruid))
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	if threatable := checkTheValueThreatable(u.LoginID); threatable {
		logger.Error("Error", logger.String("uuid", u.LoginID), logger.String("threatable", c.ClientIP()))
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	checked := checkFirebaseUserEmail(c, u.LoginID, u.Useruid)
	if !checked {
		logger.Error("Error", logger.String("checkneedinserterr", "checkFirebaseUserEmailError"), logger.String("uuid", u.Useruid))
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	var updatedprofile bool

	updatedprofile, checkneedinserterr = getNeedInert(u.LoginID)
	if checkneedinserterr != nil && checkneedinserterr != sql.ErrNoRows {
		logger.Error("Error", logger.String("checkneedinserterr", checkneedinserterr.Error()), logger.String("uuid", u.Useruid))
		c.AbortWithStatus(http.StatusNotAcceptable)
	}
	token, checkneedinsertokenerr := CreateToken(u.Useruid)
	if checkneedinsertokenerr != nil {
		logger.Error("Error", logger.String("checkneedinsertokenerr", checkneedinsertokenerr.Error()), logger.String("uuid", u.Useruid))
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	checkneedinsertokenerr = CreateAccessAuth(u.Useruid, token)
	if checkneedinsertokenerr != nil {
		logger.Error("Error", logger.String("checkneedinsertokenerr", checkneedinsertokenerr.Error()), logger.String("uuid", u.Useruid))
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	if updatedprofile {
		c.JSON(http.StatusCreated, gin.H{
			"authtoken": token.AccessToken,
		})
		return
	} else {
		c.JSON(http.StatusAccepted, gin.H{
			"authtoken": token.AccessToken,
		})
		return
	}
}

func getCFUrl(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"ContentUrl": cfurl})
}

func getNeedInert(email string) (bool, error) {
	var updatedprofile bool
	updateprofileerr := readdb.Get(&updatedprofile, "SELECT updatedprofile FROM urmyuserinfo WHERE email=$1", email)
	if updateprofileerr == sql.ErrNoRows {
		return false, sql.ErrNoRows
	} else if updateprofileerr != nil {
		return false, WrapError("updateprofileerr", updateprofileerr.Error())
	} else {

		return updatedprofile, nil
	}
}

func signup(c *gin.Context) {
	var u *InsertUserProfile

	var signuperr error
	if signuperr := c.ShouldBindJSON(&u); signuperr != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	saju, uuid, signuperr := AddUrMyUser(u)
	if signuperr != nil {
		logger.Error("Error", logger.String("signuperr", signuperr.Error()), logger.String("uuid", u.Useruid))
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	} else {

		palja := a.sa.ExtractSaju(saju)

		dasaeun := a.sa.ExtractDaeUnSaeUn(saju, palja)
		signuperr = InputUrMySaJuInfo(uuid, palja, dasaeun)
		if signuperr != nil {
			logger.Error("Error", logger.String("signuperr", signuperr.Error()), logger.String("uuid", u.Useruid))
			c.AbortWithStatus(http.StatusNotImplemented)
			return
		} else {
			ExecRefreshMaterializedView()
			go sendCandidateNotification(uuid)
			c.JSON(http.StatusCreated, gin.H{
				"message": "created",
			})

		}
	}
}

func insertpropic(c *gin.Context) {
	var upropicerr error

	urmyuuid, _ := c.MustGet("userId").(string)
	file, header, upropicerr := c.Request.FormFile("profile")
	if upropicerr != nil {
		logger.Error("Error", logger.String("upropicerr", upropicerr.Error()), logger.String("uuid", urmyuuid))
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	defer file.Close()
	//log.Println(urmyuuid)
	buffer := make([]byte, header.Size)
	file.Read(buffer)
	var tmpfilename string
	//detectedcontenttype := http.DetectContentType(buffer)
	//fmt.Println(detectedcontenttype)

	if strings.Contains(header.Filename, ".jpeg") || strings.Contains(header.Filename, ".jpg") {
		tmpfilename = time.Now().Format("2006-01-02T15:04:05") + ".jpg"
	} else if strings.Contains(header.Filename, ".png") {
		tmpfilename = time.Now().Format("2006-01-02T15:04:05") + ".png"
	} else {
		logger.Error("Error", logger.String("upropicerr", upropicerr.Error()), logger.String("uuid", urmyuuid))
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	c.SaveUploadedFile(header, filepath.Join("./uploaded.jpg"))
	upropicerr = PutProfilePicS3(buffer, urmyuuid, tmpfilename)
	if upropicerr != nil {
		logger.Error("Error", logger.String("upropicerr", upropicerr.Error()), logger.String("uuid", urmyuuid))
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	_, upropicerr = UpdateProfilePicData(urmyuuid, tmpfilename)
	if upropicerr != nil {
		logger.Error("Error", logger.String("upropicerr", upropicerr.Error()), logger.String("uuid", urmyuuid))
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, gin.H{})

}

func getudata(c *gin.Context) {
	urmyuuid := c.MustGet("userId").(string)
	var u SignInUser

	var getudataerr error
	if getudataerr = c.ShouldBindJSON(&u); getudataerr != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	u.Useruuid = urmyuuid
	if threatable := checkTheValueThreatable(u.LoginID); threatable {
		logger.Error("Error", logger.String("uuid", u.LoginID), logger.String("threatable", c.ClientIP()))
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	urmyuuid, getudataerr = GetUrMyUserEmail(&u)
	if getudataerr != nil {
		logger.Error("Error", logger.String("getudataerr", getudataerr.Error()), logger.String("uuid", u.Useruuid))
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	go sendCandidateNotification(urmyuuid)
	userdata, getudataerr := GetUrMyUserData(urmyuuid)
	if getudataerr != nil {
		logger.Error("Error", logger.String("getudataerr", getudataerr.Error()), logger.String("uuid", urmyuuid))
		c.AbortWithStatus(http.StatusNotFound)
		return
	} else {
		c.JSON(http.StatusOK, gin.H{
			"urdata": userdata,
		})
	}
}

func sendCandidateNotification(urmyuuid string) {

	var sendcandinotierr error
	mydata, sendcandinotierr := GetSajuData(urmyuuid)
	if sendcandinotierr != nil {
		logger.Error("Error", logger.String("sendcandinotierr", sendcandinotierr.Error()), logger.String("uuid", urmyuuid))
		return
	}
	currenturmycount, _ := geturmycount(urmyuuid)

	candidatelist, sendcandinotierr := getNotificationCandidatelist(urmyuuid, mydata)
	if sendcandinotierr != nil {
		logger.Error("Error", logger.String("sendcandinotierr", sendcandinotierr.Error()), logger.String("uuid", urmyuuid))
		return
	}

	if len(candidatelist) == 0 {
		return
	}
	if currenturmycount == len(candidatelist) {
		return
	}

	sendcandinotierr = updateurmycount(urmyuuid, len(candidatelist))
	if sendcandinotierr != nil {
		logger.Error("Error", logger.String("sendcandinotierr", sendcandinotierr.Error()), logger.String("uuid", urmyuuid))
	}
	/*
		for _, candi := range candidatelist {
			count, sendcandinotierr := geturmycount(candi)
			if sendcandinotierr != nil {
				logger.Error("Error", logger.String("sendcandinotierr", sendcandinotierr.Error()), logger.String("uuid", urmyuuid))
			}
			sendcandinotierr = updateurmycount(candi, count+1)
			if sendcandinotierr != nil {
				logger.Error("Error", logger.String("sendcandinotierr", sendcandinotierr.Error()), logger.String("uuid", urmyuuid))
			}

				ctx := context.Background()
				registrationToken, sendcandinotierr := getUserNotificationToken(ctx, candi)
				if sendcandinotierr != nil {
					logger.Error("Error", logger.String("sendcandinotierr", sendcandinotierr.Error()), logger.String("uuid", urmyuuid))
				}

				message := &messaging.Message{
					Data: map[string]string{
						"content":      "Urmy has been signed up",
						"time":         time.Now().UTC().String(),
						"click_action": "FLUTTER_NOTIFICATION_CLICK",
					},

					Token: registrationToken,
				}

				client, sendcandinotierr := firebaseapp.Messaging(ctx)
				if sendcandinotierr != nil {
					logger.Error("Error", logger.String("sendcandinotierr", sendcandinotierr.Error()), logger.String("uuid", urmyuuid))
				}
				_, sendcandinotierr = client.Send(ctx, message)
				if sendcandinotierr != nil {
					logger.Error("Error", logger.String("sendcandinotierr", sendcandinotierr.Error()), logger.String("uuid", urmyuuid))

				}

		}
	*/
	ExecRefreshMaterializedView()
}

type Count struct {
	Count int `db:"urmycount"`
}

func geturmycount(uuid string) (int, error) {
	var count int
	geturmycounterr := readdb.Get(&count, "SELECT urmycount FROM urmyuserinfo_kr WHERE loginuuid=$1", uuid)
	if geturmycounterr != nil {
		return 0, WrapError("geturmycounterr", geturmycounterr.Error())
	}
	return count, nil
}

func updateurmycount(uuid string, count int) error {
	rows, upurmycounterr := writedb.Query("UPDATE urmyuserinfo SET urmycount=$1 WHERE loginuuid=$2", count, uuid)
	if upurmycounterr != nil {
		return WrapError("upurmycounterr", upurmycounterr.Error())
	}
	defer rows.Close()
	return nil
}

/*
	func signin(c *gin.Context) {
		var u SignInUser
		var urmyuuid string
		var signinerr error
		if signinerr = c.ShouldBindJSON(&u); signinerr != nil {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

		if strings.Contains(u.LoginID, "@") {
			urmyuuid, signinerr = GetUrMyUserEmail(&u)
			if signinerr != nil {
				checked := checkFirebaseUserEmail(c, u.LoginID, u.FirebaseUserUid)
				if checked {
					signinerr = deleteFirebaseUser(c, u.FirebaseUserUid)
					if signinerr != nil {
						logger.Error("Error", logger.String("signinerr", signinerr.Error()), logger.String("uuid", u.FirebaseUserUid))
					}
					signinerr = inputCompensationList(u.LoginID)
					if signinerr != nil {
						logger.Error("Error", logger.String("signinerr", signinerr.Error()), logger.String("uuid", u.FirebaseUserUid))
					}
					c.AbortWithStatus(http.StatusInternalServerError)
				}
				logger.Error("Error", logger.String("signinerr", signinerr.Error()), logger.String("uuid", u.FirebaseUserUid))
				c.AbortWithStatus(http.StatusNotFound)
				return
			}
		} else {
			urmyuuid, signinerr = GetUrMyUserPhone(&u)
			if signinerr != nil {
				logger.Error("Error", logger.String("signinerr", signinerr.Error()), logger.String("uuid", u.FirebaseUserUid))
				c.AbortWithStatus(http.StatusNotFound)
				return
			}
		}

		userdata, signinerr := GetUrMyUserData(urmyuuid)
		if signinerr != nil {
			logger.Error("Error", logger.String("signinerr", signinerr.Error()), logger.String("uuid", u.FirebaseUserUid))
			c.AbortWithStatus(http.StatusNotFound)
			return
		} else {

			c.JSON(http.StatusOK, gin.H{
				"urdata": userdata,
			})
		}
	}

	func applyCompensation(useruid string) error {
		purserverr := InsertPurchaseService(useruid, compserviceid)
		if purserverr != nil {
			// DB에 구매 완료가 정상적으로 등록되지 않음
			return WrapError("applycompensation", purserverr.Error())
		}
		return nil
	}

	func inputCompensationList(useremail string) error {
		_, err := writedb.Exec("INSERT INTO urmycompensationlist (email) VALUES ($1)", useremail)
		if err != nil {
			return WrapError("inputcompensationlist", err.Error())
		}
		return nil
	}

	func checkCompensationList(useremail string) (bool, error) {
		var email string
		err := readdb.Get(&email, "SELECT email FROM urmycompensationlist WHERE email = $1", useremail)
		if err != nil {
			if err == sql.ErrNoRows {
				return false, nil
			} else {
				return false, WrapError("checkgetcompensationlist", err.Error())
			}
		}

		_, delusererr := writedb.Exec("DELETE FROM urmycompensationlist Where email=$1", email)
		if delusererr != nil {
			return false, WrapError("checkdeletecompensationlist", delusererr.Error())
		}

		return true, nil
	}
*/
func signout(c *gin.Context) {
	var signouterr error
	urmyuuid := c.MustGet("userId").(string)
	if threatable := checkTheValueThreatable(urmyuuid); threatable {
		logger.Error("Error", logger.String("uuid", urmyuuid), logger.String("threatable", c.ClientIP()))
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	tokenAuth, signouterr := ExtractAccessTokenMetadata(c.Request)
	if signouterr != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		c.Abort()
		return
	}

	deleted, signouterr := DeleteAccessAuth(tokenAuth.AccessUuid)
	if signouterr != nil || deleted == 0 {
		logger.Error("Error", logger.String("signouterr", signouterr.Error()), logger.String("uuid", urmyuuid))
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	} else {
		c.JSON(http.StatusOK, gin.H{})
	}
}

////////////////// Candidate //////////////////////////////

func ListPremiumcandidate(c *gin.Context) {
	var listprecandierr error
	urmyuuid := c.MustGet("userId").(string)
	if threatable := checkTheValueThreatable(urmyuuid); threatable {
		logger.Error("Error", logger.String("uuid", urmyuuid), logger.String("threatable", c.ClientIP()))
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	purchasable, listprecandierr := CheckPurchasedService(urmyuuid, "s10001") /// 서비스 구매했는지 확인
	if listprecandierr != nil {
		logger.Error("Error", logger.String("listprecandierr", listprecandierr.Error()), logger.String("uuid", urmyuuid))
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	if purchasable {
		c.AbortWithStatus(http.StatusBadRequest)
	}

	premiumcandidatelist, listprecandierr := GetPremiumCandidateListFromCache(urmyuuid) //캐시에 저장되어있는 결과인지 확인
	if listprecandierr != nil {
		logger.Error("Error", logger.String("listprecandierr", listprecandierr.Error()), logger.String("uuid", urmyuuid))
		c.AbortWithStatus(http.StatusNotFound)
		return
	} else {

		c.JSON(http.StatusOK, gin.H{
			"premiumcandidatelist": premiumcandidatelist,
		})
	}

}

func ListBlackList(c *gin.Context) {
	var listblacklisterr error
	urmyuuid := c.MustGet("userId").(string)
	if threatable := checkTheValueThreatable(urmyuuid); threatable {
		logger.Error("Error", logger.String("uuid", urmyuuid), logger.String("threatable", c.ClientIP()))
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	blacklistdatabase := mongoconn.Database("urmyfriend").Collection("blacklist")
	blacklist, listblacklisterr := BlacklistRead(blacklistdatabase, urmyuuid)
	if listblacklisterr != nil {
		logger.Error("Error", logger.String("listblacklisterr", listblacklisterr.Error()), logger.String("uuid", urmyuuid))
		c.AbortWithStatus(http.StatusNotImplemented)
		return
	}

	blacklistinfo, listblacklisterr := GetBlackList(blacklist)
	if listblacklisterr != nil {
		logger.Error("Error", logger.String("listblacklisterr", listblacklisterr.Error()), logger.String("uuid", urmyuuid))
		c.AbortWithStatus(http.StatusNotImplemented)
	}
	if len(blacklist) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"candidatelist": []CandidateList{},
		})
		return
	}
	for i, data := range blacklistinfo {
		blacklistinfo[i].ProfilePicPath, listblacklisterr = GetUrMyUserProfilePicture(data.CandidateUUID)
		if listblacklisterr != nil {
			logger.Error("Error", logger.String("listblacklisterr", listblacklisterr.Error()), logger.String("uuid", urmyuuid))
			blacklistinfo[i].ProfilePicPath = []ProfilePic{}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"candidatelist": blacklistinfo,
	})
}

func AddBlackList(c *gin.Context) {
	var addblacklisterr error
	urmyuuid := c.MustGet("userId").(string)
	if threatable := checkTheValueThreatable(urmyuuid); threatable {
		logger.Error("Error", logger.String("uuid", urmyuuid), logger.String("threatable", c.ClientIP()))
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	var blacklist BlackList
	var friend FriendList
	if addblacklisterr = c.ShouldBindJSON(&blacklist); addblacklisterr != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	blacklistdatabase := mongoconn.Database("urmyfriend").Collection("blacklist")
	addblacklisterr = BlacklistWrite(blacklistdatabase, urmyuuid, blacklist)
	if addblacklisterr != nil {
		logger.Error("Error", logger.String("addblacklisterr", addblacklisterr.Error()), logger.String("uuid", urmyuuid))
		c.AbortWithStatus(http.StatusInternalServerError)
	}
	friend.HostUUID = blacklist.HostUUID
	friend.FriendUUID = blacklist.BlacklistUUID
	frienddatabase := mongoconn.Database("urmyfriend").Collection("friendlist")
	addblacklisterr = FriendDelete(frienddatabase, urmyuuid, friend)
	if addblacklisterr != nil {
		logger.Error("Error", logger.String("addblacklisterr", addblacklisterr.Error()), logger.String("uuid", urmyuuid))
		c.AbortWithStatus(http.StatusInternalServerError)
	}

	blacklists, addblacklisterr := BlacklistRead(blacklistdatabase, urmyuuid)
	if addblacklisterr != nil {
		logger.Error("Error", logger.String("addblacklisterr", addblacklisterr.Error()), logger.String("uuid", urmyuuid))
		c.AbortWithStatus(http.StatusNotImplemented)
	}

	blacklistsinfo, addblacklisterr := GetBlackList(blacklists)
	if addblacklisterr != nil {
		logger.Error("Error", logger.String("addblacklisterr", addblacklisterr.Error()), logger.String("uuid", urmyuuid))
		c.AbortWithStatus(http.StatusNotImplemented)
	}
	if len(blacklistsinfo) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"candidatelist": []CandidateList{},
		})
		return
	}
	for i, data := range blacklists {
		blacklistsinfo[i].ProfilePicPath, addblacklisterr = GetUrMyUserProfilePicture(data.BlacklistUUID)
		if addblacklisterr != nil {
			logger.Error("Error", logger.String("addblacklisterr", addblacklisterr.Error()), logger.String("uuid", urmyuuid))
			blacklistsinfo[i].ProfilePicPath = []ProfilePic{}
		}
	}
	c.JSON(http.StatusOK, gin.H{
		"candidatelist": blacklistsinfo,
	})
}

func DeleteBlackList(c *gin.Context) {
	var delblacklisterr error
	urmyuuid := c.MustGet("userId").(string)
	if threatable := checkTheValueThreatable(urmyuuid); threatable {
		logger.Error("Error", logger.String("uuid", urmyuuid), logger.String("threatable", c.ClientIP()))
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	var blacklist BlackList
	if delblacklisterr = c.ShouldBindJSON(&blacklist); delblacklisterr != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	blacklistdatabase := mongoconn.Database("urmyfriend").Collection("blacklist")
	delblacklisterr = BlacklistDelete(blacklistdatabase, urmyuuid, blacklist)
	if delblacklisterr != nil {
		logger.Error("Error", logger.String("delblacklisterr", delblacklisterr.Error()), logger.String("uuid", urmyuuid))
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	blacklists, delblacklisterr := BlacklistRead(blacklistdatabase, urmyuuid)
	if delblacklisterr != nil {
		logger.Error("Error", logger.String("delblacklisterr", delblacklisterr.Error()), logger.String("uuid", urmyuuid))
		c.AbortWithStatus(http.StatusNotImplemented)
		return
	}

	blacklistsinfo, delblacklisterr := GetBlackList(blacklists)
	if delblacklisterr != nil {
		logger.Error("Error", logger.String("delblacklisterr", delblacklisterr.Error()), logger.String("uuid", urmyuuid))
		c.AbortWithStatus(http.StatusNotImplemented)
		return
	}
	if len(blacklistsinfo) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"candidatelist": []CandidateList{},
		})
		return
	}
	for i, data := range blacklists {
		blacklistsinfo[i].ProfilePicPath, delblacklisterr = GetUrMyUserProfilePicture(data.BlacklistUUID)
		if delblacklisterr != nil {
			logger.Error("Error", logger.String("delblacklisterr", delblacklisterr.Error()), logger.String("uuid", urmyuuid))
			blacklistsinfo[i].ProfilePicPath = []ProfilePic{}
		}
	}
	c.JSON(http.StatusOK, gin.H{
		"candidatelist": blacklistsinfo,
	})
}

func ListFriendList(c *gin.Context) {
	var listfrilisterr error
	urmyuuid := c.MustGet("userId").(string)
	if threatable := checkTheValueThreatable(urmyuuid); threatable {
		logger.Error("Error", logger.String("uuid", urmyuuid), logger.String("threatable", c.ClientIP()))
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	frienddatabase := mongoconn.Database("urmyfriend").Collection("friendlist")
	friendlist, listfrilisterr := FriendRead(frienddatabase, urmyuuid)
	if listfrilisterr != nil {
		logger.Error("Error", logger.String("listfrilisterr", listfrilisterr.Error()), logger.String("uuid", urmyuuid))
		c.AbortWithStatus(http.StatusNotImplemented)
		return
	}

	var friendsinfo []CandidateList

	friendsinfo, getfrilisterr := GetFriendList(friendlist)
	if getfrilisterr != nil {
		logger.Error("Error", logger.String("listfrilisterr", getfrilisterr.Error()), logger.String("uuid", urmyuuid))
		c.AbortWithStatus(http.StatusNotImplemented)
		return
	}
	if len(friendsinfo) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"candidatelist": []CandidateList{},
		})
		return
	}
	for i, data := range friendsinfo {
		friendsinfo[i].ProfilePicPath, listfrilisterr = GetUrMyUserProfilePicture(data.CandidateUUID)
		if listfrilisterr != nil {
			logger.Error("Error", logger.String("listfrilisterr", listfrilisterr.Error()), logger.String("uuid", urmyuuid))
			friendsinfo[i].ProfilePicPath = []ProfilePic{}
		}
	}

	//log.Printf("list friendlist: %v", friendsinfo)
	c.JSON(http.StatusOK, gin.H{
		"candidatelist": friendsinfo,
	})
}

func AddFriendList(c *gin.Context) {
	var addfrilisterr error
	urmyuuid := c.MustGet("userId").(string)
	if threatable := checkTheValueThreatable(urmyuuid); threatable {
		logger.Error("Error", logger.String("uuid", urmyuuid), logger.String("threatable", c.ClientIP()))
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	var friend FriendList

	if addfrilisterr = c.ShouldBindJSON(&friend); addfrilisterr != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	frienddatabase := mongoconn.Database("urmyfriend").Collection("friendlist")
	prefriendlist, preaddfrilisterr := FriendRead(frienddatabase, urmyuuid)
	if preaddfrilisterr != nil {
		logger.Error("Error", logger.String("preaddfrilisterr", preaddfrilisterr.Error()), logger.String("uuid", urmyuuid))
		c.AbortWithStatus(http.StatusNotImplemented)
		return
	}

	if len(prefriendlist) > 5 {
		c.AbortWithStatus(http.StatusForbidden)
		return
	}

	addfrilisterr = FriendWrite(frienddatabase, urmyuuid, friend)
	if addfrilisterr != nil {
		logger.Error("Error", logger.String("addfrilisterr", addfrilisterr.Error()), logger.String("uuid", urmyuuid))
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	friendlist, addfrilisterr := FriendRead(frienddatabase, urmyuuid)
	if addfrilisterr != nil {
		logger.Error("Error", logger.String("addfrilisterr", addfrilisterr.Error()), logger.String("uuid", urmyuuid))
		c.AbortWithStatus(http.StatusNotImplemented)
		return
	}

	friendsinfo, addfrilisterr := GetFriendList(friendlist)
	if addfrilisterr != nil {
		logger.Error("Error", logger.String("addfrilisterr", addfrilisterr.Error()), logger.String("uuid", urmyuuid))
		c.AbortWithStatus(http.StatusNotImplemented)
		return
	}

	for i, data := range friendlist {
		friendsinfo[i].ProfilePicPath, addfrilisterr = GetUrMyUserProfilePicture(data.FriendUUID)
		if addfrilisterr != nil {
			logger.Error("Error", logger.String("addfrilisterr", addfrilisterr.Error()), logger.String("uuid", urmyuuid))
			friendsinfo[i].ProfilePicPath = []ProfilePic{}
		}
	}
	//log.Printf("add friendlist: %v", friendsinfo)
	c.JSON(http.StatusOK, gin.H{
		"candidatelist": friendsinfo,
	})
}

func AddAnonymousFriendList(c *gin.Context) {
	var addfrilisterr error
	urmyuuid := c.MustGet("userId").(string)
	if threatable := checkTheValueThreatable(urmyuuid); threatable {
		logger.Error("Error", logger.String("uuid", urmyuuid), logger.String("threatable", c.ClientIP()))
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	var friend FriendList

	if addfrilisterr = c.ShouldBindJSON(&friend); addfrilisterr != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	hostsajupalja, addanonyfrienderr := GetSajuPalja(urmyuuid)
	if addanonyfrienderr != nil {
		logger.Error("Error", logger.String("addanonyfrienderr", addanonyfrienderr.Error()), logger.String("uuid", urmyuuid))
		c.AbortWithStatus(http.StatusNotImplemented)
		return
	}
	hostdaesaeun, addanonyfrienderr := GetDaeSaeUn(urmyuuid)
	if addanonyfrienderr != nil {
		logger.Error("Error", logger.String("addanonyfrienderr", addanonyfrienderr.Error()), logger.String("uuid", urmyuuid))
		c.AbortWithStatus(http.StatusNotImplemented)
		return
	}
	opposajupalja, addanonyfrienderr := GetSajuPalja(friend.FriendUUID)
	if addanonyfrienderr != nil {
		logger.Error("Error", logger.String("addanonyfrienderr", addanonyfrienderr.Error()), logger.String("uuid", urmyuuid))
		c.AbortWithStatus(http.StatusNotImplemented)
		return
	}
	oppodaesaeun, addanonyfrienderr := GetDaeSaeUn(friend.FriendUUID)
	if addanonyfrienderr != nil {
		logger.Error("Error", logger.String("addanonyfrienderr", addanonyfrienderr.Error()), logger.String("uuid", urmyuuid))
		c.AbortWithStatus(http.StatusNotImplemented)
		return
	}

	host, opponent := a.sa.Find_GoongHab(hostsajupalja, hostdaesaeun, opposajupalja, oppodaesaeun)

	final_host_grade, final_opponent_grade, _, _ := a.sa.Evaluate_GoonbHab(host, opponent)
	friend.HostUUID = urmyuuid
	friend.HostGrade = fmt.Sprintf("%.2f", final_host_grade)
	friend.OpponentGrade = fmt.Sprintf("%.2f", final_opponent_grade)
	friend.ResultRecord = opponent.Result

	frienddatabase := mongoconn.Database("urmyfriend").Collection("friendlist")
	addfrilisterr = FriendWrite(frienddatabase, urmyuuid, friend)
	if addfrilisterr != nil {
		logger.Error("Error", logger.String("addfrilisterr", addfrilisterr.Error()), logger.String("uuid", urmyuuid))
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	friendlist, addfrilisterr := FriendRead(frienddatabase, urmyuuid)
	if addfrilisterr != nil {
		logger.Error("Error", logger.String("addfrilisterr", addfrilisterr.Error()), logger.String("uuid", urmyuuid))
		c.AbortWithStatus(http.StatusNotImplemented)
		return
	}

	friendsinfo, addfrilisterr := GetFriendList(friendlist)
	if addfrilisterr != nil {
		logger.Error("Error", logger.String("addfrilisterr", addfrilisterr.Error()), logger.String("uuid", urmyuuid))
		c.AbortWithStatus(http.StatusNotImplemented)
		return
	}

	for i, data := range friendlist {
		friendsinfo[i].ProfilePicPath, addfrilisterr = GetUrMyUserProfilePicture(data.FriendUUID)
		if addfrilisterr != nil {
			logger.Error("Error", logger.String("addfrilisterr", addfrilisterr.Error()), logger.String("uuid", urmyuuid))
			friendsinfo[i].ProfilePicPath = []ProfilePic{}
		}
	}
	//log.Printf("add anony friendlist: %v", friendsinfo)
	c.JSON(http.StatusOK, gin.H{
		"candidatelist": friendsinfo,
	})
}

func DeleteFriendList(c *gin.Context) {
	var delfrilisterr error
	urmyuuid := c.MustGet("userId").(string)
	if threatable := checkTheValueThreatable(urmyuuid); threatable {
		logger.Error("Error", logger.String("uuid", urmyuuid), logger.String("threatable", c.ClientIP()))
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	var friend FriendList
	if delfrilisterr = c.ShouldBindJSON(&friend); delfrilisterr != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	frienddatabase := mongoconn.Database("urmyfriend").Collection("friendlist")
	delfrilisterr = FriendDelete(frienddatabase, urmyuuid, friend)
	if delfrilisterr != nil {
		logger.Error("Error", logger.String("delfrilisterr", delfrilisterr.Error()), logger.String("uuid", urmyuuid))
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	friendlist, delfrilisterr := FriendRead(frienddatabase, urmyuuid)
	if delfrilisterr != nil {
		logger.Error("Error", logger.String("delfrilisterr", delfrilisterr.Error()), logger.String("uuid", urmyuuid))
		c.AbortWithStatus(http.StatusNotImplemented)
		return
	}

	friendsinfo, delfrilisterr := GetFriendList(friendlist)
	if delfrilisterr != nil {
		logger.Error("Error", logger.String("delfrilisterr", delfrilisterr.Error()), logger.String("uuid", urmyuuid))
		c.AbortWithStatus(http.StatusNotImplemented)
	}
	if len(friendsinfo) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"candidatelist": []CandidateList{},
		})
		return
	}
	for i, data := range friendlist {
		friendsinfo[i].ProfilePicPath, delfrilisterr = GetUrMyUserProfilePicture(data.FriendUUID)
		if delfrilisterr != nil {
			logger.Error("Error", logger.String("delfrilisterr", delfrilisterr.Error()), logger.String("uuid", urmyuuid))
			friendsinfo[i].ProfilePicPath = []ProfilePic{}
		}
	}
	c.JSON(http.StatusOK, gin.H{
		"candidatelist": friendsinfo,
	})
}

func ListCandidate(c *gin.Context) {
	urmyuuid := c.MustGet("userId").(string)

	if threatable := checkTheValueThreatable(urmyuuid); threatable {
		logger.Error("Error", logger.String("uuid", urmyuuid), logger.String("threatable", c.ClientIP()))
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	candidatelist, getcandidatefromcacheerr := GetCandidateListFromCache(urmyuuid)
	if getcandidatefromcacheerr != nil || len(candidatelist) == 0 {
		mydata, listcandierr := GetSajuData(urmyuuid)
		if listcandierr != nil {
			logger.Error("Error", logger.String("listcandierr", listcandierr.Error()), logger.String("uuid", urmyuuid))
			c.AbortWithStatus(http.StatusNotImplemented)
			return
		}

		candidatelist, listcandierr = GetCandidateList(urmyuuid, mydata)
		if listcandierr != nil {
			logger.Error("Error", logger.String("listcandierr", listcandierr.Error()), logger.String("uuid", urmyuuid))
			c.AbortWithStatus(http.StatusNotImplemented)
			return
		}
		if len(candidatelist) == 0 {
			c.JSON(http.StatusOK, gin.H{
				"candidatelist": []CandidateList{},
			})
			return
		}
		for i, data := range candidatelist {
			candidatelist[i].ProfilePicPath, listcandierr = GetUrMyUserProfilePicture(data.CandidateUUID)
			if listcandierr != nil {
				logger.Error("Error", logger.String("listcandierr", listcandierr.Error()), logger.String("uuid", urmyuuid))
				candidatelist[i].ProfilePicPath = []ProfilePic{}
			}
		}

	}

	/*
		blacklistdatabase := mongoconn.Database("urmyfriend").Collection("blacklist")
		blacklist, listblacklisterr := BlacklistRead(blacklistdatabase, urmyuuid)
		if listblacklisterr != nil {
			logger.Error("Error", logger.String("listcandierr", listblacklisterr.Error()), logger.String("uuid", urmyuuid))
		}

		frienddatabase := mongoconn.Database("urmyfriend").Collection("friendlist")
		friendlist, chkfrilistreaderr := FriendRead(frienddatabase, urmyuuid)
		if chkfrilistreaderr != nil {
			logger.Error("Error", logger.String("listcandierr", chkfrilistreaderr.Error()), logger.String("uuid", urmyuuid))
		}


		for i, data := range candidatelist {
			for _, friend := range friendlist {
				if data.CandidateUUID == friend.FriendUUID {
					candidatelist = RemoveIndex(candidatelist, i)
					break
				}
			}
			for _, black := range blacklist {
				if data.CandidateUUID == black.BlacklistUUID {
					candidatelist = RemoveIndex(candidatelist, i)
					break
				}
			}
		}
	*/
	if len(candidatelist) > 100 {
		getcandiputcandicacheerr := PutCandidateListToCache(urmyuuid, candidatelist)
		if getcandiputcandicacheerr != nil {
			logger.Error("Error", logger.String("listcandierr", getcandiputcandicacheerr.Error()), logger.String("uuid", urmyuuid))
		}
	}
	//log.Println(candidatelist)
	c.JSON(http.StatusOK, gin.H{
		"candidatelist": candidatelist,
	})

}

// ////////////////////Report //////////////////////////////////////
type CandidateReport struct {
	ReportedUser string `json:"reporteduser" db:"reporteduser"`
	Reporttype   string `json:"reporttype" db:"reporttype"`
	Contents     string `json:"contents" db:"contents"`
}

type ChatReport struct {
	ChatroomId   string `json:"chatroomid" db:"chatroomid"`
	Reporttype   string `json:"reporttype" db:"reporttype"`
	Messages     string `json:"messages" db:"messages"`
	MessageTime  string `json:"messagetime" db:"messagetime"`
	Contents     string `json:"contents" db:"contents"`
	ReportedUser string `json:"reporteduser" db:"reporteduser"`
}

type Inquiry struct {
	CaseID         int    `json:"caseid" db:"caseid"`
	UserID         string `json:"userid" db:"userid"`
	Inquirytype    string `json:"inquirytype" db:"inquirytype"`
	Title          string `json:"title" db:"title"`
	ManagerID      string `json:"managerid" db:"managerid"`
	InquiryState   string `json:"inquirystate" db:"inquirystate"`
	Contents       string `json:"contents" db:"contents"`
	RegisteredDate string `json:"registereddate" db:"registereddate"`
}

func InsertCandidateReport(c *gin.Context) {
	var report CandidateReport
	urmyuuid := c.MustGet("userId").(string)
	if bindreporterr := c.ShouldBindJSON(&report); bindreporterr != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	insertreporterr := InsertCandidateReportIntoDB(urmyuuid, &report)
	if insertreporterr != nil {
		logger.Error("Error", logger.String("insertreporterr", insertreporterr.Error()), logger.String("uuid", urmyuuid))
		c.AbortWithStatus(http.StatusBadRequest)
		return
	} else {
		c.JSON(http.StatusOK, gin.H{})
	}
}

func InsertChatReport(c *gin.Context) {
	var report ChatReport
	urmyuuid := c.MustGet("userId").(string)
	if bindreporterr := c.ShouldBindJSON(&report); bindreporterr != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	insertreporterr := InsertChatReportIntoDB(urmyuuid, &report)
	if insertreporterr != nil {
		logger.Error("Error", logger.String("insertreporterr", insertreporterr.Error()), logger.String("uuid", urmyuuid))
		c.AbortWithStatus(http.StatusBadRequest)
		return
	} else {
		c.JSON(http.StatusOK, gin.H{})
	}
}

func InsertInquiry(c *gin.Context) {
	urmyuuid := c.MustGet("userId").(string)
	var inquiry Inquiry
	if bindinquiryerr := c.ShouldBindJSON(&inquiry); bindinquiryerr != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	//fmt.Println(inquiry.Inquirytype)
	insertinquiryerr := InsertInquiryIntoDB(urmyuuid, &inquiry)
	if insertinquiryerr != nil {
		logger.Error("Error", logger.String("insertinquiryerr", insertinquiryerr.Error()), logger.String("uuid", urmyuuid))
		c.AbortWithStatus(http.StatusBadRequest)
		return
	} else {
		c.JSON(http.StatusOK, gin.H{})
	}
}

func InsertInquiryCotent(c *gin.Context) {
	urmyuuid := c.MustGet("userId").(string)
	var inquiry Inquiry
	if bindinquiryerr := c.ShouldBindJSON(&inquiry); bindinquiryerr != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	//fmt.Println(inquiry.Inquirytype)
	insertinquirycontenterr := InsertInquiryContentIntoDB(urmyuuid, &inquiry)
	if insertinquirycontenterr != nil {
		logger.Error("Error", logger.String("insertinquirycontenterr", insertinquirycontenterr.Error()), logger.String("uuid", urmyuuid))
		c.AbortWithStatus(http.StatusBadRequest)
		return
	} else {
		c.JSON(http.StatusOK, gin.H{})
	}
}

func GetInquiry(c *gin.Context) {
	urmyuuid := c.MustGet("userId").(string)
	inquirylist, inquirylisterr := GetInquiryFromDB(urmyuuid)
	if inquirylisterr != nil {
		logger.Error("Error", logger.String("inquirylisterr", inquirylisterr.Error()), logger.String("uuid", urmyuuid))
		c.AbortWithStatus(http.StatusBadRequest)
		return
	} else {
		c.JSON(http.StatusOK, gin.H{"inquirylist": inquirylist})
	}
}

func GetInquiryContent(c *gin.Context) {
	urmyuuid := c.MustGet("userId").(string)
	var inquiry Inquiry
	if bindinquiryerr := c.ShouldBindJSON(&inquiry); bindinquiryerr != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	inquirycontentlist, inquirycontentlisterr := GetInquiryContentFromDB(inquiry.CaseID)
	if inquirycontentlisterr != nil {
		logger.Error("Error", logger.String("inquirycontentlisterr", inquirycontentlisterr.Error()), logger.String("uuid", urmyuuid))
		c.AbortWithStatus(http.StatusBadRequest)
		return
	} else {
		c.JSON(http.StatusOK, gin.H{"inquirycontentlist": inquirycontentlist})
	}
}

func InsertCandidateReportIntoDB(urmyuuid string, report *CandidateReport) error {
	insertreport, insertreporterr := writedb.Query("INSERT INTO urmycandidatereport (reportinguser, reporteduser, reporttype, contents, registereddate, states) VALUES ($1, $2, $3, $4,  current_timestamp, $5)", urmyuuid, report.ReportedUser, report.Reporttype, report.Contents, "registered")
	if insertreporterr != nil {
		return WrapError("insertreporterr", insertreporterr.Error())
	}
	defer insertreport.Close()
	return nil
}

func InsertChatReportIntoDB(urmyuuid string, report *ChatReport) error {
	insertreport, insertreporterr := writedb.Query("INSERT INTO urmychatreport (reportinguser, reporteduser, chatroomid, reporttype, messages, messagetime, contents, registereddate, states) VALUES ($1, $2, $3, $4, $5, $6, $7, current_timestamp, $8)", urmyuuid, report.ReportedUser, report.ChatroomId, report.Reporttype, report.Messages, report.MessageTime, report.Contents, "registered")
	if insertreporterr != nil {
		return WrapError("insertreporterr", insertreporterr.Error())
	}
	defer insertreport.Close()
	return nil
}

func InsertInquiryIntoDB(urmyuuid string, inquiry *Inquiry) error {
	var caseid int
	ctx := context.Background()
	tx, txbeginerr := writedb.BeginTx(ctx, nil)
	if txbeginerr != nil {
		return WrapError("txbeginerr", txbeginerr.Error())
	}
	{
		stmt, insertinquirytodberr := tx.Prepare("INSERT INTO urmyinquiry (userid, inquirytype, title, inquirystate) VALUES ($1, $2, $3, $4) RETURNING caseid")
		if insertinquirytodberr != nil {
			tx.Rollback()
			return WrapError("insertinquirytodberr", insertinquirytodberr.Error())
		}

		defer stmt.Close()

		queryerr := stmt.QueryRow(urmyuuid, inquiry.Inquirytype, inquiry.Title, "registered").Scan(&caseid)
		if queryerr != nil {
			tx.Rollback()
			return WrapError("queryerr", queryerr.Error())
		}
	}
	{
		txcommiterr := tx.Commit()
		if txcommiterr != nil {
			return WrapError("txcommiterr", txcommiterr.Error())
		}

		_, insertinquirycontentserr := writedb.Query("INSERT INTO urmyinquirycontent (caseid, managerid, contents, registereddate) VALUES ($1, $2,$3, current_timestamp)", caseid, urmyuuid, inquiry.Contents)
		if insertinquirycontentserr != nil {

			return WrapError("insertinquirycontentserr", insertinquirycontentserr.Error())
		}

	}
	return nil
}

func InsertInquiryContentIntoDB(urmyuuid string, inquiry *Inquiry) error {
	ctx := context.Background()
	tx, txbeginerr := writedb.BeginTx(ctx, nil)
	if txbeginerr != nil {
		return WrapError("txbeginerr", txbeginerr.Error())
	}

	_, updateinquiryerr := tx.ExecContext(ctx, "UPDATE urmyinquiry SET inquirystate=$1 WHERE caseid=$2", "replied", inquiry.CaseID)
	if updateinquiryerr != nil {
		tx.Rollback()
		return WrapError("updateinquiryerr", updateinquiryerr.Error())
	}

	_, insertinquirycontentserr := tx.ExecContext(ctx, "INSERT INTO urmyinquirycontent (caseid, managerid, contents, registereddate) VALUES ($1, $2, $3, current_timestamp)", inquiry.CaseID, urmyuuid, inquiry.Contents)
	if insertinquirycontentserr != nil {

		return WrapError("insertinquirycontentserr", insertinquirycontentserr.Error())
	}

	txcommiterr := tx.Commit()
	if txcommiterr != nil {
		return WrapError("txcommiterr", txcommiterr.Error())
	}

	return nil
}

func GetInquiryFromDB(urmyuuid string) ([]Inquiry, error) {
	var inquirylist []Inquiry
	getinquirylisterr := readdb.Select(&inquirylist, "SELECT caseid, inquirytype, title, inquirystate FROM urmyinquiry WHERE userid=$1", urmyuuid)
	if getinquirylisterr != nil {
		return nil, WrapError("geturmycounterr", getinquirylisterr.Error())
	}
	return inquirylist, nil
}

func GetInquiryContentFromDB(caseid int) ([]Inquiry, error) {
	var inquirylist []Inquiry
	getinquirycontentlisterr := readdb.Select(&inquirylist, "SELECT caseid, contents, managerid, registereddate FROM urmyinquirycontent WHERE caseid=$1 ORDER BY registereddate", caseid)
	if getinquirycontentlisterr == sql.ErrNoRows {
		return nil, nil
	} else if getinquirycontentlisterr != nil {
		return nil, WrapError("getinquirycontentlisterr", getinquirycontentlisterr.Error())
	}

	return inquirylist, nil
}

//////////////////////Message/////////////////////////////////////
/*
func updateChatPicture(c *gin.Context) {
	urmyuuid := c.MustGet("userId").(string)
	if threatable := checkTheValueThreatable(urmyuuid); threatable {
		logger.Error("Error", logger.String("uuid", urmyuuid), logger.String("threatable", c.ClientIP()))
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	var upchatpicerr error
	file, header, upchatpicerr := c.Request.FormFile("chatpicture")
	if upchatpicerr != nil {
		logger.Error("Error", logger.String("upchatpicerr", upchatpicerr.Error()), logger.String("uuid", urmyuuid))
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	defer file.Close()
	chatroomId := c.Request.Form.Get("chatId")
	buffer := make([]byte, header.Size)
	file.Read(buffer)

	var tmpfilename string

	if strings.Contains(header.Filename, ".jpeg") || strings.Contains(header.Filename, ".jpg") {
		tmpfilename = time.Now().Format("2006-01-02T15:04:05") + ".jpg"

	} else if strings.Contains(header.Filename, ".png") {
		tmpfilename = time.Now().Format("2006-01-02T15:04:05") + ".png"

	} else {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	putinput := s3.PutObjectInput{
		Bucket:             aws.String("urmy-tf-cloudfront-s3"),
		Key:                aws.String("chatlist/" + chatroomId + "/" + tmpfilename),
		Body:               bytes.NewReader(buffer),
		ContentType:        aws.String(http.DetectContentType(buffer)),
		ContentDisposition: aws.String("attachment"),
	}

	_, upchatpicerr = s3client.PutObject(context.Background(), &putinput)
	if upchatpicerr != nil {
		logger.Error("Error", logger.String("upchatpicerr", upchatpicerr.Error()), logger.String("uuid", urmyuuid))
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"imagepath": cfurl + "dev/chatlist/" + chatroomId + "/" + tmpfilename,
	})
}
*/
/////////////////////// Service //////////////////////////////////
//ListService PurchaseService RefundRequest NumberOfOppositeSex

func ListService(c *gin.Context) {
	urmyuuid := c.MustGet("userId").(string)
	if threatable := checkTheValueThreatable(urmyuuid); threatable {
		logger.Error("Error", logger.String("uuid", urmyuuid), logger.String("threatable", c.ClientIP()))
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	service, lserverr := GetService()
	if lserverr != nil {
		logger.Error("Error", logger.String("lserverr", lserverr.Error()), logger.String("uuid", urmyuuid))
		c.AbortWithStatus(http.StatusNotImplemented)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"servicelist": service,
	})
}

func PurchaseService(c *gin.Context) {
	var purserverr error
	urmyuuid := c.MustGet("userId").(string)
	if threatable := checkTheValueThreatable(urmyuuid); threatable {
		logger.Error("Error", logger.String("uuid", urmyuuid), logger.String("threatable", c.ClientIP()))
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	var service ServiceList
	if purserverr = c.ShouldBindJSON(&service); purserverr != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	switch service.ServiceUUID {
	case "s10000":
		ok, purserverr := CheckPurchasedService(urmyuuid, service.ServiceUUID)
		if purserverr != nil {
			logger.Error("Error", logger.String("purserverr", purserverr.Error()), logger.String("uuid", urmyuuid))
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		if ok {
			purserverr := InsertPurchaseService(urmyuuid, service.ServiceUUID)
			if purserverr != nil {
				// DB에 구매 완료가 정상적으로 등록되지 않음
				logger.Error("Error", logger.String("purserverr", purserverr.Error()), logger.String("uuid", urmyuuid))
				c.AbortWithStatus(http.StatusInternalServerError)
				return
			}
			c.JSON(http.StatusOK, gin.H{})
		}
	}
}

func RefundRequest(c *gin.Context) {
	var serviceid string
	var refreqerr error
	urmyuuid := c.MustGet("userId").(string)
	if threatable := checkTheValueThreatable(urmyuuid); threatable {
		logger.Error("Error", logger.String("uuid", urmyuuid), logger.String("threatable", c.ClientIP()))
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	if refreqerr = c.ShouldBindJSON(&serviceid); refreqerr != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	invoiceid, refreqerr := CheckRefundableItemExist(urmyuuid, serviceid)
	if refreqerr != nil {
		logger.Error("Error", logger.String("refreqerr", refreqerr.Error()), logger.String("uuid", urmyuuid))
		c.AbortWithStatus(http.StatusNotImplemented)
		return
	}
	if invoiceid == "urmyovertime" {
		logger.Error("Error", logger.String("refreqerr", "urmyovertime"), logger.String("uuid", urmyuuid))
		c.AbortWithStatus(http.StatusNotAcceptable)
		return
	}
	refreqerr = InsertRefundRequest(urmyuuid, invoiceid)
	if refreqerr != nil {
		logger.Error("Error", logger.String("refreqerr", refreqerr.Error()), logger.String("uuid", urmyuuid))
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	c.JSON(http.StatusOK, gin.H{})

}

func ListUrMyServicePurchased(c *gin.Context) {
	urmyuuid := c.MustGet("userId").(string)
	if threatable := checkTheValueThreatable(urmyuuid); threatable {
		logger.Error("Error", logger.String("uuid", urmyuuid), logger.String("threatable", c.ClientIP()))
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	servicepurchased, lservpurerr := GetUrMyServicePurchased(urmyuuid)
	if lservpurerr != nil {
		logger.Error("Error", logger.String("lservpurerr", lservpurerr.Error()), logger.String("uuid", urmyuuid))
		c.AbortWithStatus(http.StatusNotImplemented)
	}
	c.JSON(http.StatusOK, gin.H{
		"servicepurchased": servicepurchased,
	})

}

func NumberOfOppositeSex(c *gin.Context) {
	urmyuuid := c.MustGet("userId").(string)
	if threatable := checkTheValueThreatable(urmyuuid); threatable {
		logger.Error("Error", logger.String("uuid", urmyuuid), logger.String("threatable", c.ClientIP()))
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	count, numoppsexerr := GetNumberOfOppositeSex(urmyuuid)
	if numoppsexerr != nil {
		logger.Error("Error", logger.String("numoppsexerr", numoppsexerr.Error()), logger.String("uuid", urmyuuid))
		c.AbortWithStatus(http.StatusNotImplemented)
	}
	c.JSON(http.StatusOK, gin.H{
		"count": count,
	})
}

func NumberOfUrMy(c *gin.Context) {
	urmyuuid := c.MustGet("userId").(string)
	if threatable := checkTheValueThreatable(urmyuuid); threatable {
		logger.Error("Error", logger.String("uuid", urmyuuid), logger.String("threatable", c.ClientIP()))
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	count, numberofurmyerr := geturmycount(urmyuuid)
	if numberofurmyerr != nil {
		logger.Error("Error", logger.String("numberofurmyerr", numberofurmyerr.Error()), logger.String("uuid", urmyuuid))
		c.AbortWithStatus(http.StatusNotImplemented)
	}
	c.JSON(http.StatusOK, gin.H{
		"count": count,
	})
}

//////////////////   Relational Database   /////////////////////////////

func ExecRefreshMaterializedView() {
	_, execrefreshmaterializedviewerr := writedb.Exec("REFRESH MATERIALIZED VIEW urmyuserinfo_kr")
	if execrefreshmaterializedviewerr != nil {
		//log.Println("Refresh view error: $v", execrefreshmaterializedviewerr)
		logger.Error("Error", logger.String("execrefreshmaterializedviewerr", execrefreshmaterializedviewerr.Error()))
	}
}

func UpdateProfilePicData(userid string, filename string) (bool, error) {

	var sequencemax int
	uppropicselmaxprosqerr := readdb.Get(&sequencemax, "SELECT MAX(profilesequence) FROM urmyuserprofile WHERE loginuuid=$1", userid)
	if uppropicselmaxprosqerr != nil {
		if uppropicselmaxprosqerr == sql.ErrNoRows {
			sequencemax = 0
			inserturmyuserprofile, uppropicnorawsinserr := writedb.Query("INSERT INTO urmyuserprofile (loginuuid, profilename, updateddate, profilesequence) VALUES ($1, $2, current_timestamp, $3)", userid, filename, sequencemax)
			if uppropicnorawsinserr != nil {
				return false, WrapError("uppropicnorawsinserr", uppropicnorawsinserr.Error())
			}
			defer inserturmyuserprofile.Close()
			return true, nil
		}
	}

	inserturmyuserprofile, uppropicinserr := writedb.Query("INSERT INTO urmyuserprofile (loginuuid, profilename, updateddate, profilesequence) VALUES ($1, $2, current_timestamp, $3)", userid, filename, sequencemax+1)
	if uppropicinserr != nil {
		return false, WrapError("uppropicinserr", uppropicinserr.Error())
	}
	defer inserturmyuserprofile.Close()

	return true, nil
}

func UpdateProfilePicSequenceData(userid string, profilePicPath []*ProfilePic) (bool, error) {
	for i := 0; i < len(profilePicPath); i++ {
		rows, upropicseqerr := writedb.Query("UPDATE urmyuserprofile SET profilesequence=$1 WHERE loginuuid=$2 AND profilename=$3", profilePicPath[i].ProfileSequence, userid, profilePicPath[i].Filename)
		if upropicseqerr != nil {
			return false, WrapError("upropicseqerr", upropicseqerr.Error())
		}
		defer rows.Close()
	}

	return true, nil
}

func DeleteProfilePicData(userid string, filename string) (bool, error) {
	rows, delprofpicerr := writedb.Query("DELETE FROM urmyuserprofile Where loginuuid=$1 and profilename=$2", userid, filename)
	if delprofpicerr != nil {
		return false, WrapError("delprofpicerr", delprofpicerr.Error())
	}
	defer rows.Close()
	return true, nil
}

func GetService() ([]ServiceList, error) {
	var servicelist []ServiceList
	getservicerr := readdb.Select(&servicelist, "SELECT serviceid, servicename, serviceprice FROM urmyservice")
	if getservicerr == sql.ErrNoRows {
		return []ServiceList{}, nil
	}
	if getservicerr != nil {
		return nil, WrapError("getservicerr", getservicerr.Error())
	}
	return servicelist, nil
}

func CheckPurchasedService(urmyuuid string, serviceid string) (bool, error) {
	var servicesales ServiceSales
	chkpurserviceselerr := readdb.Get(&servicesales, "SELECT serviceid, servicename, serviceprice FROM urmysales WHERE purchaseuser=$1 AND serviceid=$2", urmyuuid, serviceid)
	if chkpurserviceselerr == sql.ErrNoRows {
		return true, nil
	}
	if chkpurserviceselerr != nil {
		return false, WrapError("chkpurserviceselerr", chkpurserviceselerr.Error())
	}
	temp, chkpurserviceparserr := time.Parse(time.RFC3339, servicesales.PurchasedDate)
	if chkpurserviceparserr != nil {
		return false, WrapError("chkpurserviceparserr", chkpurserviceparserr.Error())
	}
	///결제일 제한시간///
	if temp.AddDate(0, 1, 0).Before(time.Now().Local()) {
		return true, nil
	}
	return false, nil
}

func InsertPurchaseService(urmyuuid string, serviceid string) error {
	inserturmysales, inspurservicerr := writedb.Query("INSERT INTO urmysales (serviceid, purchaseuser, purchasedate) VALUES ($1, $2, current_timestamp)", urmyuuid, serviceid)
	if inspurservicerr != nil {
		return WrapError("inspurservicerr", inspurservicerr.Error())
	}
	defer inserturmysales.Close()
	return nil
}

func CheckRefundableItemExist(urmyuuid string, serviceid string) (string, error) {
	var servicesales ServiceSales
	chkrefitemexistselerr := readdb.Get(&servicesales, "SELECT serviceid, servicename, serviceprice FROM urmysales WHERE purchaseuser=$1 AND serviceid=$2", urmyuuid, serviceid)
	if chkrefitemexistselerr == sql.ErrNoRows {
		return "", nil
	}
	if chkrefitemexistselerr != nil {
		return "", WrapError("chkrefitemexistselerr", chkrefitemexistselerr.Error())
	}
	temp, chkrefitemexistparserr := time.Parse(time.RFC3339, servicesales.PurchasedDate)
	if chkrefitemexistparserr != nil {
		return "", WrapError("chkrefitemexistparserr", chkrefitemexistparserr.Error())
	}
	///결제일 제한시간///
	if time.Since(temp) > 72 {
		return "urmyovertime", nil
	} else {
		return servicesales.InvoiceId, nil
	}
}

func InsertRefundRequest(urmyuuid string, invoiceid string) error {
	inserturmysales, insrefreqerr := writedb.Query("INSERT INTO urmyrefundrequest (invoiceid, refunduser, refundrequestdate) VALUES ($1, $2, current_timestamp)", invoiceid, urmyuuid)
	if insrefreqerr != nil {
		return WrapError("insrefreqerr", insrefreqerr.Error())
	}
	defer inserturmysales.Close()
	return nil
}

func GetUrMyServicePurchased(urmyuuid string) ([]ServiceSales, error) {
	var purchased []ServiceSales
	getpurservicerr := readdb.Get(&purchased, "SELECT serviceid, purchasedate FROM urmyusurmysaleserinfo WHERE purchaseuser=$1", urmyuuid)
	if getpurservicerr == sql.ErrNoRows {
		return []ServiceSales{}, nil
	}
	if getpurservicerr != nil {
		return nil, WrapError("getpurservicerr", getpurservicerr.Error())
	}
	return purchased, nil
}

func GetNumberOfOppositeSex(urmyuuid string) (int, error) {
	var gender bool
	var count int
	getnumoppsexgenerr := readdb.Get(&gender, "SELECT gender FROM urmyuserinfo_kr WHERE loginuuid=$1", urmyuuid)
	if getnumoppsexgenerr != nil {
		return -1, WrapError("getnumoppsexgenerr", getnumoppsexgenerr.Error())
	}

	getnumoppsexcounterr := readdb.Get(&count, "SELECT COUNT(*) FROM urmyuserinfo_kr WHERE gender=$1", !gender)
	if getnumoppsexcounterr != nil {
		return -1, WrapError("getnumoppsexcounterr", getnumoppsexcounterr.Error())
	}
	return count, nil
}

func UpdateUrMyDaeSaeUn(useruuid string, daesaeun saju.DaeSaeUn) error {
	_, updaesaeunerr := writedb.Exec("UPDATE urmysajudaesaeun SET daeunchun=$1, daeunji=$2, saeunchun=$3, saeunji=$4 WHERE loginuuid=$5", daesaeun.DaeUnChun, daesaeun.DaeUnJi, daesaeun.SaeUnChun, daesaeun.SaeUnJi, useruuid)
	if updaesaeunerr != nil {
		return WrapError("updaesaeunerr", updaesaeunerr.Error())
	}
	return nil
}

func UpdateUrMySaJuInfo(loginuuid string, sajudata *saju.SaJuPalJa, daesaeun *saju.DaeSaeUn) error {
	ctx := context.Background()
	tx, upsajubegintxerr := writedb.BeginTx(ctx, nil)
	if upsajubegintxerr != nil {
		return WrapError("upsajubegintxerr", upsajubegintxerr.Error())
	}
	_, upsajupaljaerr := tx.ExecContext(ctx, "UPDATE urmysajupalja SET yearChun=$1, yearJi=$2, monthChun=$3, monthJi=$4, dayChun=$5, dayJi=$6, timeChun=$7, timeJi=$8 WHERE loginuuid=$9", sajudata.YearChun, sajudata.YearJi, sajudata.MonthChun, sajudata.MonthJi, sajudata.DayChun, sajudata.DayJi, sajudata.TimeChun, sajudata.TimeJi, loginuuid)
	if upsajupaljaerr != nil {
		tx.Rollback()
		return WrapError("upsajupaljaerr", upsajupaljaerr.Error())
	}

	_, upsajudaesaeunerr := tx.ExecContext(ctx, "UPDATE urmysajudaesaeun SET daeunchun=$1, daeunji=$2, saeunchun=$3, saeunji=$4 WHERE loginuuid=$5", daesaeun.DaeUnChun, daesaeun.DaeUnJi, daesaeun.SaeUnChun, daesaeun.SaeUnJi, loginuuid)
	if upsajudaesaeunerr != nil {
		tx.Rollback()
		return WrapError("upsajudaesaeunerr", upsajudaesaeunerr.Error())
	}

	upsajutxerr := tx.Commit()
	if upsajutxerr != nil {
		return WrapError("upsajutxerr", upsajutxerr.Error())
	} else {
		fmt.Println("....saju Transaction committed")
	}
	return nil
}

func InputUrMySaJuInfo(loginuuid string, sajudata *saju.SaJuPalJa, daesaeun *saju.DaeSaeUn) error {
	ctx := context.Background()
	tx, insajubegintxerr := writedb.BeginTx(ctx, nil)
	if insajubegintxerr != nil {
		return WrapError("insajubegintxerr", insajubegintxerr.Error())
	}
	_, insajupaljaerr := tx.ExecContext(ctx, "INSERT INTO urmysajupalja (loginuuid, yearChun, yearJi, monthChun, monthJi, dayChun, dayJi, timeChun, timeJi) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)", loginuuid, sajudata.YearChun, sajudata.YearJi, sajudata.MonthChun, sajudata.MonthJi, sajudata.DayChun, sajudata.DayJi, sajudata.TimeChun, sajudata.TimeJi)
	if insajupaljaerr != nil {
		tx.Rollback()
		return WrapError("insajupaljaerr", insajupaljaerr.Error())
	}

	_, insajudaesaeunerr := tx.ExecContext(ctx, "INSERT INTO urmysajudaesaeun (loginuuid, daeunchun, daeunji, saeunchun, saeunji) VALUES ($1,$2,$3,$4,$5)", loginuuid, daesaeun.DaeUnChun, daesaeun.DaeUnJi, daesaeun.SaeUnChun, daesaeun.SaeUnJi)
	if insajudaesaeunerr != nil {
		tx.Rollback()
		return WrapError("insajudaesaeunerr", insajudaesaeunerr.Error())
	}

	insajutxerr := tx.Commit()
	if insajutxerr != nil {
		return WrapError("insajutxerr", insajutxerr.Error())
	} else {
		fmt.Println("....saju Transaction committed")
	}
	return nil
}

func GetSajuData(uuid string) (*saju.Saju, error) {
	var mydata CandidateList
	getsajuselmydataerr := readdb.Get(&mydata, "SELECT birthday, gender FROM urmyuserinfo_kr WHERE loginuuid=$1", uuid)
	if getsajuselmydataerr != nil {
		return nil, WrapError("getsajuselmydataerr", getsajuselmydataerr.Error())
	}

	var saju saju.Saju
	saju.Gender = mydata.CandidateGender
	getsajuselbirtherr := readdb.Get(&saju, "SELECT urmyyear, urmymonth, urmyday, urmytime FROM urmybirthday WHERE birthday=$1", mydata.CandidateBirthDate)
	if getsajuselbirtherr != nil {
		return nil, WrapError("getsajuselmydataerr", getsajuselmydataerr.Error())
	}
	return &saju, nil
}

func CheckBlackList(urmyuuid string, userid string) (bool, error) {
	blacklistdatabase := mongoconn.Database("urmyfriend").Collection("blacklist")
	if blacklistdatabase == nil {
		return false, errors.New("chkblacklistconnectionerr")
	}

	blacklist, chkblacklistreaderr := BlacklistRead(blacklistdatabase, urmyuuid)
	if chkblacklistreaderr != nil {
		return false, WrapError("chkblacklistreaderr", chkblacklistreaderr.Error())
	}

	for _, blacked := range blacklist {
		if blacked.BlacklistUUID == userid {
			return true, nil
		}
	}
	return false, nil
}

func CheckFriendList(urmyuuid string, userid string) (bool, error) {
	frienddatabase := mongoconn.Database("urmyfriend").Collection("friendlist")
	if frienddatabase == nil {
		return false, errors.New("chkfrilistconnectionerr")
	}

	friendlist, chkfrilistreaderr := FriendRead(frienddatabase, urmyuuid)
	if chkfrilistreaderr != nil {
		return false, WrapError("chkfrilistreaderr", chkfrilistreaderr.Error())
	}

	for _, friended := range friendlist {
		if friended.FriendUUID == userid {
			return true, nil
		}
	}
	return false, nil
}

func GetBlackList(blacklist []BlackList) ([]CandidateList, error) {
	var candidatelist []CandidateList = make([]CandidateList, 0)
	var candidates CandidateList
	for _, data := range blacklist {
		getblacklisterr := readdb.Get(&candidates, "SELECT loginuuid, name, birthday, gender FROM urmyuserinfo_kr WHERE loginuuid=$1", data.BlacklistUUID)
		if getblacklisterr != nil && getblacklisterr != sql.ErrNoRows {
			return nil, WrapError("getblacklisterr", getblacklisterr.Error())
		} else if getblacklisterr == sql.ErrNoRows {
			continue
		}
		candidates.ContentUrl = cfurl
		candidates.HostGrade = data.HostGrade
		candidates.OpponentGrade = data.OpponentGrade
		candidatelist = append(candidatelist, candidates)
	}

	for i := 0; i < len(candidatelist); i++ {
		candidatelist[i].ContentUrl = cfurl
	}

	return candidatelist, nil
}

func GetFriendList(friendlist []FriendList) ([]CandidateList, error) {
	var candidatelist []CandidateList = make([]CandidateList, 0)
	var candidates CandidateList
	for _, data := range friendlist {
		getfrilisterr := readdb.Get(&candidates, "SELECT loginuuid, name, birthday, gender, residentcountry,residentstate,residentcity, description, birthday, mbti FROM urmyuserinfo_kr WHERE loginuuid=$1", data.FriendUUID)
		if getfrilisterr != nil && getfrilisterr != sql.ErrNoRows {
			return nil, WrapError("getfrilisterr", getfrilisterr.Error())
		} else if getfrilisterr == sql.ErrNoRows {
			continue
		}
		candidates.ContentUrl = cfurl
		candidates.HostGrade = data.HostGrade
		candidates.OpponentGrade = data.OpponentGrade
		candidates.ResultRecord = data.ResultRecord
		candidatelist = append(candidatelist, candidates)
	}

	return candidatelist, nil
}

func GetCandidateList(urmyuid string, hostsaju *saju.Saju) ([]CandidateList, error) {

	var candidatelist []CandidateList

	var urmyuser GetCandidate
	getcandigenerr := readdb.Get(&urmyuser, "SELECT gender, birthday, tester FROM urmyuserinfo_kr WHERE loginuuid=$1", urmyuid)
	if getcandigenerr != nil {
		return nil, WrapError("getcandigenerr", getcandigenerr.Error())
	}
	year, getcandigetyearerr := strconv.Atoi(urmyuser.Birthdate[:4])
	if getcandigetyearerr != nil {
		return nil, WrapError("getcandigetyearerr", getcandigetyearerr.Error())
	}

	if urmyuser.Tester {
		getcandicandilerr := readdb.Select(&candidatelist, "SELECT loginuuid, name, gender, residentcountry,residentstate,residentcity, description, birthday, mbti FROM urmyuserinfo_kr where gender=$1 and tester=$2 and urmyyear between $3 and $4 order by RANDOM() limit 20", !urmyuser.Gender, true, year-12, year+12)
		if getcandicandilerr != nil {
			return nil, WrapError("getcandicandilerr", getcandicandilerr.Error())
		}
	} else {
		getcandicandilerr := readdb.Select(&candidatelist, "SELECT loginuuid, name, gender, residentcountry,residentstate,residentcity, description, birthday, mbti FROM urmyuserinfo_kr where gender=$1 and tester=$2 and urmyyear between $3 and $4 order by RANDOM() limit 20", !urmyuser.Gender, false, year-12, year+12)
		if getcandicandilerr != nil {
			return nil, WrapError("getcandicandilerr", getcandicandilerr.Error())
		}
	}

	hostsajupalja, getcandihostsajuerr := GetSajuPalja(urmyuid)
	if getcandihostsajuerr != nil {
		return nil, WrapError("getcandihostsajuerr", getcandihostsajuerr.Error())
	}

	hostdaesaeun, getcandihostdaesaeunerr := GetDaeSaeUn(urmyuid)
	if getcandihostdaesaeunerr != nil {
		return nil, WrapError("getcandihostdaesaeunerr", getcandihostdaesaeunerr.Error())
	}

	for i := 0; i < len(candidatelist); i++ {
		candidatelist[i].ContentUrl = cfurl
		opponentsajupalja, getcandisajuerr := GetSajuPalja(candidatelist[i].CandidateUUID)
		if getcandisajuerr != nil {
			return nil, WrapError("getcandisajuerr", getcandisajuerr.Error())
		}

		opponentdaesaeun, getcandidaesaeunerr := GetDaeSaeUn(candidatelist[i].CandidateUUID)
		if getcandidaesaeunerr != nil {
			return nil, WrapError("getcandidaesaeunerr", getcandidaesaeunerr.Error())
		}

		host, opponent := a.sa.Find_GoongHab(hostsajupalja, hostdaesaeun, opponentsajupalja, opponentdaesaeun)
		host.LoginID = urmyuid
		opponent.LoginID = candidatelist[i].CandidateUUID

		final_host_grade, final_opponent_grade, _, _ := a.sa.Evaluate_GoonbHab(host, opponent)
		candidatelist[i].HostGrade = fmt.Sprintf("%.2f", final_host_grade)
		candidatelist[i].OpponentGrade = fmt.Sprintf("%.2f", final_opponent_grade)
		candidatelist[i].ResultRecord = opponent.Result
	}

	ok := true

	if ok {
		sort.SliceStable(candidatelist, func(i, j int) bool {
			hosti, hostierr := strconv.ParseFloat(candidatelist[i].HostGrade, 64)
			if hostierr != nil {
				//log.Println(hostierr) // 3.1415927410125732
				return false
			}
			oppoi, oppoierr := strconv.ParseFloat(candidatelist[i].OpponentGrade, 64)
			if oppoierr != nil {
				//log.Println(oppoierr) // 3.1415927410125732
				return false
			}
			hostj, hostjerr := strconv.ParseFloat(candidatelist[j].HostGrade, 64)
			if hostjerr != nil {
				//log.Println(hostjerr) // 3.1415927410125732
				return false
			}
			oppoj, oppojerr := strconv.ParseFloat(candidatelist[j].OpponentGrade, 64)
			if oppojerr != nil {
				//log.Println(oppojerr) // 3.1415927410125732
				return false
			}
			return (hosti+oppoi)/2 > (hostj+oppoj)/2
		})
	}

	return candidatelist, nil
}

func RemoveIndex(s []CandidateList, index int) []CandidateList {
	ret := make([]CandidateList, 0)
	ret = append(ret, s[:index]...)
	return append(ret, s[index+1:]...)
}

func GetSajuPalja(useruuid string) (*saju.SaJuPalJa, error) {
	var sajupalja saju.SaJuPalJa
	getsajupaljaerr := readdb.Get(&sajupalja, "SELECT yearchun, yearji, monthchun, monthji, daychun, dayji, timechun, timeji FROM urmysajupalja WHERE loginuuid=$1", useruuid)
	if getsajupaljaerr != nil {

		return nil, WrapError("getsajupaljaerr", getsajupaljaerr.Error())
	}
	return &sajupalja, nil
}

func GetDaeSaeUn(useruuid string) (*saju.DaeSaeUn, error) {
	var sajupalja saju.DaeSaeUn
	getdaesaeunerr := readdb.Get(&sajupalja, "SELECT daeunchun, daeunji, saeunchun, saeunji FROM urmysajudaesaeun WHERE loginuuid=$1", useruuid)
	if getdaesaeunerr != nil {
		return nil, WrapError("getdaesaeunerr", getdaesaeunerr.Error())
	}
	return &sajupalja, nil
}

type GetCountry struct {
	Country string `db:"country"`
	City    string `db:"city"`
	Gender  bool   `db:"gender"`
}

func CheckBirthdateUpdatable(useruuid string, birthdate string) (bool, error) {
	var birthdateupdate sql.NullString

	parseinputbirthdate, chkupdatableparseinputbirtherr := time.Parse("2006-01-02 15:04:05", birthdate)
	if chkupdatableparseinputbirtherr != nil {
		return false, WrapError("chkupdatableselbirtherr", chkupdatableparseinputbirtherr.Error())
	}
	if parseinputbirthdate.Year() > time.Now().Year()-19 {
		return false, errors.New("young")
	}

	if parseinputbirthdate.Year() < 1951 {
		return false, errors.New("over")
	}

	chkupdatableselbirtherr := readdb.Get(&birthdateupdate, "SELECT birthdayupdate FROM urmyuserinfo_kr WHERE loginuuid=$1", useruuid)
	if chkupdatableselbirtherr != nil {
		return false, WrapError("chkupdatableselbirtherr", chkupdatableselbirtherr.Error())
	}

	if !birthdateupdate.Valid {
		return true, nil
	} else {
		parsebirthdateupdate, chkupdatableparsebirthupdateerr := time.Parse(time.RFC3339, birthdateupdate.String)
		if chkupdatableparsebirthupdateerr != nil {
			return false, WrapError("chkupdatableparsebirthupdateerr", chkupdatableparsebirthupdateerr.Error())
		}

		if parsebirthdateupdate.AddDate(0, 1, 0).Before(time.Now().Local()) {
			return true, nil
		} else {
			return false, errors.New("notabletochangebirthdate")
		}

	}
}

func UpdateUrMyProfile(useruuid string, updateprofile *UpdateProfile) (bool, error) {
	var result sql.Result

	switch updateprofile.Key {
	case "email":
		var upuinfoemailerr error
		result, upuinfoemailerr = writedb.Exec("UPDATE urmyuserinfo SET email=$1 WHERE loginuuid=$2", updateprofile.Value, useruuid)
		if upuinfoemailerr != nil {
			return false, WrapError("upuinfoemailerr", upuinfoemailerr.Error())
		}
	case "birthdate":
		updatable, upuinfoupdatablerr := CheckBirthdateUpdatable(useruuid, updateprofile.Value)
		if upuinfoupdatablerr != nil {
			return false, WrapError("upuinfoupdatablerr", upuinfoupdatablerr.Error())
		}
		if updatable {
			temp, upuinfotimeparserr := time.Parse("2006-01-02 15:04:05", updateprofile.Value)
			if upuinfotimeparserr != nil {
				return false, WrapError("upuinfotimeparserr", upuinfotimeparserr.Error())
			}

			var getcountry GetCountry
			upuinfogetcongenerr := readdb.Get(&getcountry, "SELECT country, city, gender FROM urmyuserinfo WHERE loginuuid=$1", useruuid)
			if upuinfogetcongenerr != nil {
				return false, WrapError("upuinfogetcongenerr", upuinfogetcongenerr.Error())
			}

			operation, indexhour, indexmin := setTimeByLocation(getcountry.Country, getcountry.City)
			if !operation {
				temp = temp.Add(time.Hour * time.Duration(-1*indexhour))
				temp = temp.Add(time.Minute * time.Duration(-1*indexmin))

			} else {
				temp = temp.Add(time.Hour * time.Duration(indexhour))
				temp = temp.Add(time.Minute * time.Duration(indexmin))
			}

			hour, min, _ := temp.Clock()
			year, month, day := temp.Date()

			saju := a.sa.ParseSaju(year, int(month), day, hour, min, getcountry.Gender)

			var tempMonth, tempDay string
			if saju.Month < 10 {
				tempMonth = "0" + strconv.Itoa(saju.Month)
			} else {
				tempMonth = strconv.Itoa(saju.Month)
			}
			if saju.Day < 10 {
				tempDay = "0" + strconv.Itoa(saju.Day)
			} else {
				tempDay = strconv.Itoa(saju.Day)
			}

			tempBirth := strconv.Itoa(saju.Year) + tempMonth + tempDay + saju.Time

			var selecturmybirthday string
			ctx := context.Background()
			upuinfoselbirtherr := readdb.QueryRowContext(ctx, "SELECT birthday FROM urmybirthday WHERE birthday=$1", tempBirth).Scan(&selecturmybirthday)
			if upuinfoselbirtherr != nil {
				if upuinfoselbirtherr == sql.ErrNoRows {
					_, upuinfoinsbirtherr := writedb.Exec("INSERT INTO urmybirthday (birthday, urmyyear, urmymonth, urmyday, urmytime) VALUES ($1, $2, $3, $4, $5)",
						tempBirth, saju.Year, saju.Month, saju.Day, saju.Time)
					if upuinfoinsbirtherr != nil {
						return false, WrapError("upuinfoinsbirtherr", upuinfoinsbirtherr.Error())
					}
				} else {
					return false, WrapError("upuinfoselbirtherr", upuinfoselbirtherr.Error())
				}
			}

			var upuinfobirtherr error
			result, upuinfobirtherr = writedb.Exec("UPDATE urmyuserinfo SET birthday=$1, birthdayupdate=current_timestamp WHERE loginuuid=$2", tempBirth, useruuid)
			if upuinfobirtherr != nil {
				return false, WrapError("upuinfobirtherr", upuinfobirtherr.Error())
			}

			palja := a.sa.ExtractSaju(saju)
			dasaeun := a.sa.ExtractDaeUnSaeUn(saju, palja)
			upuinfosajuerr := UpdateUrMySaJuInfo(useruuid, palja, dasaeun)
			if upuinfosajuerr != nil {
				return false, WrapError("upuinfosajuerr", upuinfosajuerr.Error())
			}
		} else {
			return false, errors.New("notabletochangebirthdate")
		}

	case "name":
		var upuinfonameerr error
		result, upuinfonameerr = writedb.Exec("UPDATE urmyuserinfo SET name=$1 WHERE loginuuid=$2", updateprofile.Value, useruuid)
		if upuinfonameerr != nil {
			return false, WrapError("upuinfombtierr", upuinfonameerr.Error())
		}
	case "mbti":
		var upuinfombtierr error
		result, upuinfombtierr = writedb.Exec("UPDATE urmyuserinfo SET mbti=$1 WHERE loginuuid=$2", updateprofile.Value, useruuid)
		if upuinfombtierr != nil {
			return false, WrapError("upuinfombtierr", upuinfombtierr.Error())
		}
	case "phoneno":
		var upuinfophonenoerr error
		result, upuinfophonenoerr = writedb.Exec("UPDATE urmyuserinfo SET phoneno=$1 WHERE loginuuid=$2", updateprofile.Value, useruuid)
		if upuinfophonenoerr != nil {
			return false, WrapError("upuinfophonenoerr", upuinfophonenoerr.Error())
		}
	case "description":
		var upuinfodescriptionerr error
		result, upuinfodescriptionerr = writedb.Exec("UPDATE urmyuserinfo SET description=$1 WHERE loginuuid=$2", updateprofile.Value, useruuid)
		if upuinfodescriptionerr != nil {
			return false, WrapError("upuinfodescriptionerr", upuinfodescriptionerr.Error())
		}
	case "country":
		var upuinfocountryerr error
		slice := strings.Split(updateprofile.Value, ",")
		result, upuinfocountryerr = writedb.Exec("UPDATE urmyuserinfo SET residentcountry=$1, residentstate=$2, residentcity=$3 WHERE loginuuid=$4", slice[0], slice[1], slice[2], useruuid)
		if upuinfocountryerr != nil {
			return false, WrapError("upuinfocountryerr", upuinfocountryerr.Error())
		}
	}

	value, updaterawaffectederr := result.RowsAffected()
	if value != 1 {
		return false, WrapError("updaterawaffectederr", updaterawaffectederr.Error())
	} else {
		return true, nil
	}

}

func DeleteUrMyUser(urmyuuid string, uinfo UserinfoToDelete) error {
	var delusererr error

	_, adduserinsbirtherr := writedb.Exec("INSERT INTO urmyuserdeleted (loginuuid, country, city, birthday, mbti, gender, deldate) VALUES ($1, $2, $3, $4, $5, $6, current_timestamp)",
		uinfo.LoginUuid, uinfo.Country, uinfo.City, uinfo.Birthdate, uinfo.Mbti, uinfo.Gender)
	if adduserinsbirtherr != nil {
		return WrapError("adduserindeltablerr", adduserinsbirtherr.Error())
	}

	deletemongoconn := mongoconn.Database("urmyusers").Collection("usersignin")
	go func() {
		var filesname []string
		getpropicnamerr := readdb.Select(&filesname, "SELECT profilename FROM urmyuserprofile WHERE loginuuid=$1", urmyuuid)
		if getpropicnamerr != nil {
			if getpropicnamerr != sql.ErrNoRows {
				logger.Warn("selectdelpropicerr", logger.String(urmyuuid, getpropicnamerr.Error()))
			}
		}
		for _, file := range filesname {
			delpropicerr := DeleteProfilePicS3(urmyuuid, strings.TrimSpace(file))
			if delpropicerr != nil {
				logger.Warn("delpropicerr", logger.String(urmyuuid, file))
			}
		}
		_, delusererr = writedb.Exec("DELETE FROM urmyuserprofile Where loginuuid=$1", urmyuuid)
	}()

	go func() {
		_, delusererr = DeleteIdMongo(deletemongoconn, urmyuuid)
	}()

	go func() {
		ctx := context.Background()
		client, _ := firebaseapp.Firestore(ctx)
		deletefriendmongoconn := mongoconn.Database("urmyfriend").Collection("friendlist")

		hosts, hostreaderr := HostRead(deletefriendmongoconn, urmyuuid)
		if hostreaderr != nil {
			logger.Warn("hostreaderr", logger.String(urmyuuid, hostreaderr.Error()))
		}

		for _, data := range hosts {
			chatid := makeChatId(urmyuuid, data)
			_, deletechatlisterr := client.Collection("user").Doc(data).Collection("chatroomId").Doc(chatid).Delete(ctx)
			if deletechatlisterr != nil {
				logger.Error("Error", logger.String("deletechatlisterr", deletechatlisterr.Error()), logger.String("chatid", chatid), logger.String("host", data), logger.String("user", urmyuuid))
			}

		}

		_, delusererr = DeleteIdFromFriendMongo(deletefriendmongoconn, urmyuuid)
	}()
	go func() {
		deleteblackmongoconn := mongoconn.Database("urmyfriend").Collection("blacklist")
		_, delusererr = DeleteIdFromBlackMongo(deleteblackmongoconn, urmyuuid)
	}()

	_, delurmyusersagreement := writedb.Exec("DELETE FROM urmyusersagreement WHERE loginuuid=$1", urmyuuid)
	if delurmyusersagreement != nil && delurmyusersagreement != sql.ErrNoRows {
		delusererr = WrapError("delurmyusersagreement", delurmyusersagreement.Error())
	}
	_, delurmyuserprofile := writedb.Exec("DELETE FROM urmyuserprofile WHERE loginuuid=$1", urmyuuid)
	if delurmyuserprofile != nil && delurmyuserprofile != sql.ErrNoRows {
		delusererr = WrapError("delurmyuserprofile", delurmyuserprofile.Error())
	}
	_, delurmychatreport := writedb.Exec("DELETE FROM urmychatreport WHERE reportinguser=$1", urmyuuid)
	if delurmychatreport != nil && delurmychatreport != sql.ErrNoRows {
		delusererr = WrapError("delurmychatreport", delurmychatreport.Error())
	}
	_, delurmyinquiry := writedb.Exec("DELETE FROM urmyinquiry WHERE userid=$1", urmyuuid)
	if delurmyinquiry != nil && delurmyinquiry != sql.ErrNoRows {
		delusererr = WrapError("delurmyinquiry", delurmyinquiry.Error())
	}
	_, delusererruserinfo := writedb.Exec("DELETE FROM urmyuserinfo WHERE loginuuid=$1", urmyuuid)
	if delusererruserinfo != nil {
		delusererr = WrapError("delusererruserinfo", delusererruserinfo.Error())
	}
	ExecRefreshMaterializedView()
	if delusererr != nil {
		return WrapError("delusererr", delusererr.Error())
	}
	return nil
}

func makeChatId(myuuid string, opponenduuid string) string {
	myuuid = strings.ReplaceAll(myuuid, "-", "")
	opponenduuid = strings.ReplaceAll(opponenduuid, "-", "")
	if hash(myuuid) > hash(opponenduuid) {
		return "urmy" + opponenduuid + myuuid
	} else {
		return "urmy" + myuuid + opponenduuid
	}
}

func GetUrMyUserEmail(signinuser *SignInUser) (string, error) {
	var urmy SignInResult

	singinemailerr := readdb.Get(&urmy, "SELECT deviceos, loginuuid, birthday, gender FROM urmyuserinfo WHERE email=$1 AND loginuuid=$2", signinuser.LoginID, signinuser.Useruuid)
	if singinemailerr != nil {
		return "", WrapError("singinemailerr", singinemailerr.Error())
	}

	if signinuser.DeviceOS != strings.TrimSpace(urmy.DeviceOS) {
		_, updevinfoerr := writedb.Exec("UPDATE urmyuserinfo SET deviceos=$1 WHERE loginuuid=$2", signinuser.DeviceOS, signinuser.Useruuid)
		if updevinfoerr != nil {
			return "", WrapError("updevinfoerr", updevinfoerr.Error())
		}
	}

	signinmongoconn := mongoconn.Database("urmyusers").Collection("usersignin")
	_, singinmongoerr := SignInMongo(signinmongoconn, &urmy, signinuser.FirebaseNotificationToken, signinuser.UrMyPlatform)
	if singinmongoerr != nil {
		return "", WrapError("singinemailmongoerr", singinmongoerr.Error())
	}

	return urmy.LoginUuid, nil
}

func GetUrMyUserPhone(signinuser *SignInUser) (string, error) {
	var urmy SignInResult

	signinphoneerr := readdb.Get(&urmy, "SELECT deviceos, loginuuid, birthday, gender FROM urmyuserinfo WHERE email=$1 AND loginuuid=$2", signinuser.LoginID, signinuser.Useruuid)
	if signinphoneerr != nil {
		return "signinphoneerr", WrapError("signinphoneerr", signinphoneerr.Error())
	}

	if signinuser.DeviceOS != strings.TrimSpace(urmy.DeviceOS) {
		_, updevinfoerr := writedb.Exec("UPDATE urmyuserinfo SET deviceos=$1 WHERE loginuuid=$2", signinuser.DeviceOS, signinuser.Useruuid)
		if updevinfoerr != nil {
			return "", WrapError("updevinfoerr", updevinfoerr.Error())
		}
	}

	signinmongoconn := mongoconn.Database("urmyusers").Collection("usersignin")
	_, singinmongoerr := SignInMongo(signinmongoconn, &urmy, signinuser.FirebaseNotificationToken, signinuser.UrMyPlatform)
	if singinmongoerr != nil {
		return "", WrapError("singinphonemongoerr", singinmongoerr.Error())
	}
	return urmy.LoginUuid, nil
}

func GetUserDataToDelete(uuid string) (UserinfoToDelete, error) {
	var userinfo UserinfoToDelete
	getuserdataerr := readdb.Get(&userinfo, "SELECT loginuuid, gender, mbti, country, city, birthday FROM urmyuserinfo WHERE loginuuid=$1", uuid)
	if getuserdataerr != nil {
		return UserinfoToDelete{}, WrapError("getuserdataerr", getuserdataerr.Error())
	} else {
		return userinfo, nil
	}
}

func GetUrMyUserData(uuid string) (Userinfo, error) {
	var userinfo Userinfo

	getuserdataerr := readdb.Get(&userinfo, "SELECT loginuuid, email, name, birthday, gender, mbti, residentcountry, residentstate, residentcity, description FROM urmyuserinfo WHERE loginuuid=$1", uuid)
	if getuserdataerr != nil {
		return Userinfo{}, WrapError("getuserdataerr", getuserdataerr.Error())
	}

	saju, getsajupaljaerr := GetSajuPalja(uuid)
	if getsajupaljaerr != nil {
		return Userinfo{}, WrapError("getsajupaljaerr", getuserdataerr.Error())
	}
	userinfo.SajuPalja.YearChun = base64.StdEncoding.EncodeToString([]byte(saju.YearChun))
	userinfo.SajuPalja.YearJi = base64.StdEncoding.EncodeToString([]byte(saju.YearJi))
	userinfo.SajuPalja.MonthChun = base64.StdEncoding.EncodeToString([]byte(saju.MonthChun))
	userinfo.SajuPalja.MonthJi = base64.StdEncoding.EncodeToString([]byte(saju.MonthJi))
	userinfo.SajuPalja.DayChun = base64.StdEncoding.EncodeToString([]byte(saju.DayChun))
	userinfo.SajuPalja.DayJi = base64.StdEncoding.EncodeToString([]byte(saju.DayJi))
	userinfo.SajuPalja.TimeChun = base64.StdEncoding.EncodeToString([]byte(saju.TimeChun))
	userinfo.SajuPalja.TimeJi = base64.StdEncoding.EncodeToString([]byte(saju.TimeJi))

	var userprofile []ProfilePic
	userprofile, userprofileerr := GetUrMyUserProfilePicture(uuid)
	if userprofileerr != nil {
		userprofile = []ProfilePic{}
	}

	userinfo.ProfilePicPath = userprofile

	return userinfo, nil
}

func GetUrMyUserProfilePicture(useruuid string) ([]ProfilePic, error) {
	var userprofile []ProfilePic
	getuserpropicerr := readdb.Select(&userprofile, "SELECT profilename, updateddate, profilesequence FROM urmyuserprofile WHERE loginuuid=$1", useruuid)
	if getuserpropicerr != nil && getuserpropicerr != sql.ErrNoRows {
		return nil, WrapError("getuserpropicerr", getuserpropicerr.Error())
	}
	return userprofile, nil
}

func setTimeByLocation(countryname string, cityname string) (bool, int, int) {
	for _, country := range a.co.Country {
		if country.CountryName == countryname {
			for _, city := range country.City {
				if city.CityName == cityname {
					return city.PlusMinus, city.Hour, city.Min
				}
			}
		}
	}
	return false, 0, 0
}

func AddUrMyUser(signupuser *InsertUserProfile) (*saju.Saju, string, error) {
	//birthdateinsert := pq.QuoteLiteral(signupuser.Birthdate)

	temp, adduserbirthparserr := time.Parse("2006-01-02 15:04:05", signupuser.Birthdate)
	if adduserbirthparserr != nil {
		return nil, "", WrapError("adduserbirthparserr", adduserbirthparserr.Error())
	}

	if temp.Year() > time.Now().Year()-19 {
		return nil, "", errors.New("young")
	}

	if temp.Year() < 1951 {
		return nil, "", errors.New("over")
	}

	operation, indexhour, indexmin := setTimeByLocation(signupuser.Country, signupuser.City)
	if !operation {
		temp = temp.Add(time.Hour * time.Duration(-1*indexhour))
		temp = temp.Add(time.Minute * time.Duration(-1*indexmin))

	} else {
		temp = temp.Add(time.Hour * time.Duration(indexhour))
		temp = temp.Add(time.Minute * time.Duration(indexmin))
	}

	hour, min, _ := temp.Clock()
	year, month, day := temp.Date()

	saju := a.sa.ParseSaju(year, int(month), day, hour, min, signupuser.Gender)

	var tempMonth, tempDay string
	if saju.Month < 10 {
		tempMonth = "0" + strconv.Itoa(saju.Month)
	} else {
		tempMonth = strconv.Itoa(saju.Month)
	}
	if saju.Day < 10 {
		tempDay = "0" + strconv.Itoa(saju.Day)
	} else {
		tempDay = strconv.Itoa(saju.Day)
	}

	tempBirth := strconv.Itoa(saju.Year) + tempMonth + tempDay + saju.Time

	var selecturmybirthday string
	ctx := context.Background()
	adduserselbirtherr := readdb.QueryRowContext(ctx, "SELECT birthday FROM urmybirthday WHERE birthday=$1", tempBirth).Scan(&selecturmybirthday)
	switch {
	case adduserselbirtherr == sql.ErrNoRows:
		_, adduserinsbirtherr := writedb.Exec("INSERT INTO urmybirthday (birthday, urmyyear, urmymonth, urmyday, urmytime) VALUES ($1, $2, $3, $4, $5)",
			tempBirth, saju.Year, saju.Month, saju.Day, saju.Time)
		if adduserinsbirtherr != nil {
			return nil, "", WrapError("adduserinsbirtherr", adduserinsbirtherr.Error())
		}
	case adduserselbirtherr != nil:
		return nil, "", WrapError("adduserselbirtherr", adduserselbirtherr.Error())
	}

	tx, adduserbegintxerr := writedb.BeginTx(ctx, nil)
	if adduserbegintxerr != nil {
		return nil, "", WrapError("adduserbegintxerr", adduserbegintxerr.Error())
	}
	_, adduserinfoerr := tx.ExecContext(ctx, "INSERT INTO urmyuserinfo (loginuuid, deviceos, email, name, birthday, gender, mbti, country, city, description, updatedprofile) VALUES ($1, $2, $3, $4, (SELECT birthday from urmybirthday WHERE birthday=$5), $6, $7, $8, $9, $10,$11)",
		signupuser.Useruid, signupuser.DeviceOS, signupuser.LoginID, signupuser.Name, tempBirth, signupuser.Gender, signupuser.Mbti, signupuser.Country, signupuser.City, signupuser.Description, true)
	if adduserinfoerr != nil {
		tx.Rollback()
		var value string = signupuser.Useruid + "/" + signupuser.DeviceOS + "/" + signupuser.LoginID + "/" + signupuser.Name + "/" + tempBirth + "/" + strconv.FormatBool(signupuser.Gender) + "/" + signupuser.Mbti + "/" + signupuser.Country + "/" + signupuser.City + "/" + signupuser.Description
		return nil, "", WrapError("adduserinfoerr", WrapError(value, adduserinfoerr.Error()).Error())
	}

	_, adduseragreerr := tx.ExecContext(ctx, "INSERT INTO urmyusersagreement (loginuuid, isoverage, urmyprivacy, urmyservice, urmyillegal ,createdat) VALUES ($1, $2, $3, $4, $5, current_timestamp)",
		signupuser.Useruid, signupuser.Isoverage, signupuser.Privacy, signupuser.Service, signupuser.Illegal)
	if adduseragreerr != nil {
		tx.Rollback()
		return nil, "", WrapError("adduseragreerr", adduseragreerr.Error())
	}

	_, adduserefresherr := tx.ExecContext(ctx, "REFRESH MATERIALIZED VIEW urmyuserinfo_kr")
	if adduserefresherr != nil {
		return nil, "", WrapError("adduserefresherr", adduserefresherr.Error())
	}

	addusertxcommiterr := tx.Commit()
	if addusertxcommiterr != nil {
		return nil, "", WrapError("addusertxcommiterr", addusertxcommiterr.Error())
	} else {
		fmt.Println("....Signup Transaction committed")
	}

	return saju, signupuser.Useruid, nil
}

func getNotificationCandidatelist(urmyuid string, hostsaju *saju.Saju) ([]string, error) {
	var urmyuser GetCandidate
	getnoticandigenerr := readdb.Get(&urmyuser, "SELECT gender, birthday FROM urmyuserinfo_kr WHERE loginuuid=$1", urmyuid)
	if getnoticandigenerr != nil {
		return nil, WrapError("getnoticandiyearerr", getnoticandigenerr.Error())
	}

	year, getnoticandiyearerr := strconv.Atoi(urmyuser.Birthdate[:4])
	if getnoticandiyearerr != nil {
		return nil, WrapError("getnoticandiyearerr", getnoticandiyearerr.Error())
	}

	var birthdaylist []Birthdate
	getnoticandilistbirtherr := readdb.Select(&birthdaylist, "SELECT birthday,urmyyear, urmymonth, urmyday, urmytime FROM urmyuserinfo_kr where gender=$1 and urmyyear between $2 and $3", !urmyuser.Gender, year-12, year+12)
	if getnoticandilistbirtherr != nil && getnoticandilistbirtherr != sql.ErrNoRows {
		return nil, WrapError("getnoticandilistbirtherr", getnoticandilistbirtherr.Error())
	}

	hostsajupalja, getnoticandisajuerr := GetSajuPalja(urmyuid)
	if getnoticandisajuerr != nil {
		return nil, WrapError("getnoticandisajuerr", getnoticandisajuerr.Error())
	}

	hostdaesaeun, getnoticandidaesaeerr := GetDaeSaeUn(urmyuid)
	if getnoticandidaesaeerr != nil {
		return nil, WrapError("getnoticandidaesaeerr", getnoticandidaesaeerr.Error())
	}

	birthdaylist = makeBirthdateSliceUnique(birthdaylist)

	if len(birthdaylist) == 0 {
		return []string{}, nil
	}

	for i, birthdata := range birthdaylist {
		temphour, _ := strconv.Atoi(birthdata.Urmytime[:2])
		tempmin, _ := strconv.Atoi(birthdata.Urmytime[3:5])
		temp := time.Date(birthdata.Urmyyear, time.Month(birthdata.Urmymonth), birthdata.Urmyday, temphour, tempmin, 0, 0, time.UTC)
		hour, min, _ := temp.Clock()
		year, month, day := temp.Date()

		saju := a.sa.ParseSaju(year, int(month), day, hour, min, !urmyuser.Gender)

		palja := a.sa.ExtractSaju(saju)

		dasaeun := a.sa.ExtractDaeUnSaeUn(saju, palja)

		host, opponent := a.sa.Find_GoongHab(hostsajupalja, hostdaesaeun, palja, dasaeun)

		final_host_grade, final_opponent_grade, _, _ := a.sa.Evaluate_GoonbHab(host, opponent)
		birthdaylist[i].Grade = (final_host_grade + final_opponent_grade) / 2
		birthdaylist[i].HostGrade = final_host_grade
		birthdaylist[i].OpponentGrade = final_opponent_grade
	}

	sort.Slice(birthdaylist, func(i, j int) bool {
		return birthdaylist[i].Grade > birthdaylist[j].Grade
	})

	var candidatelist []string
	if birthdaylist[0].Grade > 9 {
		getnoticandiseleerr := readdb.Select(&candidatelist, "SELECT loginuuid FROM urmyuserinfo_kr where gender=$1 and birthday=$2", !urmyuser.Gender, birthdaylist[0].Birthday)
		if getnoticandiseleerr != nil {
			return nil, WrapError("getnoticandiseleerr", getnoticandiseleerr.Error())
		}
	}

	return candidatelist, nil
}

func Close() {
	readdb.Close()
	writedb.Close()
}

func ConnectToDB() {
	var readdberr, writedberr error

	readpsqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+"password=%s dbname=%s sslmode=disable", env.Readhost, env.Readport, env.Readuser, env.ReadPassword, env.ReadDBName)
	readdb, readdberr = sqlx.Open("postgres", readpsqlInfo)
	if readdb != nil {
		readdb.SetMaxOpenConns(100) //최대 커넥션
		readdb.SetMaxIdleConns(10)  //대기 커넥션
	}
	if readdberr != nil {
		logger.Error("Error", logger.String("readdberr", readdberr.Error()))
		panic(readdberr)
	}
	readpingerr := readdb.Ping()
	if readpingerr != nil {
		logger.Error("Error", logger.String("readpingerr", readpingerr.Error()))
		panic(readpingerr)
	} else {
		fmt.Println("Read Database Connected")
	}

	writepsqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+"password=%s dbname=%s sslmode=disable", env.Readhost, env.Readport, env.Readuser, env.ReadPassword, env.ReadDBName)
	writedb, writedberr = sqlx.Open("postgres", writepsqlInfo)
	if writedb != nil {
		writedb.SetMaxOpenConns(100) //최대 커넥션
		writedb.SetMaxIdleConns(10)  //대기 커넥션
	}
	if writedberr != nil {
		logger.Error("Error", logger.String("writedberr", writedberr.Error()))
		panic(writedberr)
	}

	writepingerr := writedb.Ping()
	if writepingerr != nil {
		logger.Error("Error", logger.String("writepingerr", writepingerr.Error()))
		panic(writepingerr)
	} else {
		fmt.Println("Write Database Connected")
	}
}

////////////////   Redis      ////////////////////////////

func GetPremiumCandidateListFromCache(useruid string) ([]CandidateList, error) {

	candidatelistrow, getprecandifromcachegeterr := rdb.Get(useruid + "urmypremiumcandidate" + time.Now().Month().String()).Result()
	if getprecandifromcachegeterr != nil {
		return nil, WrapError("getprecandifromcachegeterr", getprecandifromcachegeterr.Error())
	}
	//userID, _ := strconv.ParseUint(userid, 10, 64)

	var candidatelist []CandidateList
	if getprecandifromcacheunmarshalerr := json.Unmarshal([]byte(candidatelistrow), &candidatelist); getprecandifromcacheunmarshalerr != nil {
		return nil, WrapError("getprecandifromcacheunmarshalerr", getprecandifromcacheunmarshalerr.Error())
	}
	return candidatelist, nil
}

func GetCandidateListFromCache(useruid string) ([]CandidateList, error) {

	candidatelistrow, getcandifromcachegeterr := rdb.Get(useruid + "urmycandidate").Result()
	if getcandifromcachegeterr != nil {
		return nil, WrapError("putcanditocachseterr", getcandifromcachegeterr.Error())
	}
	//userID, _ := strconv.ParseUint(userid, 10, 64)

	var candidatelist []CandidateList
	if getcandifromcacheunmarshalerr := json.Unmarshal([]byte(candidatelistrow), &candidatelist); getcandifromcacheunmarshalerr != nil {
		return nil, WrapError("getcandifromcacheunmarshalerr", getcandifromcacheunmarshalerr.Error())
	}

	return candidatelist, nil
}

func PutCandidateListToCache(useruid string, candidatelist []CandidateList) error {
	//at := time.Now().Add(time.Second * 240).Unix() //converting Unix to UTC(to Time object)
	//now := time.Now()
	result, putcanditocachemarshalerr := json.Marshal(candidatelist)
	if putcanditocachemarshalerr != nil {
		return WrapError("putcanditocachemarshalerr", putcanditocachemarshalerr.Error())
	}
	putcanditocachseterr := rdb.Set(useruid+"urmycandidate", result, time.Hour*60).Err()
	if putcanditocachseterr != nil {
		return WrapError("putcanditocachseterr", putcanditocachseterr.Error())
	}
	return nil
}

func CreateToken(userid string) (*AccessTokenDetails, error) {
	td := &AccessTokenDetails{}
	td.AtExpires = time.Now().Add(time.Hour * 10).Unix()
	//16070400
	u := generateuuid()
	td.AccessUuid = u

	atclaims := jwt.MapClaims{}
	atclaims["authorized"] = true
	atclaims["access_uuid"] = td.AccessUuid
	atclaims["user_id"] = userid
	atclaims["exp"] = td.AtExpires
	atoken := jwt.NewWithClaims(jwt.SigningMethodHS256, atclaims)

	var cretoksignedsterr error
	td.AccessToken, cretoksignedsterr = atoken.SignedString(mySigningKey)
	if cretoksignedsterr != nil {
		return nil, WrapError("cretoksignedsterr", cretoksignedsterr.Error())
	}
	return td, nil
}

func ExtractAccessToken(r *http.Request) string {
	bearToken := r.Header.Get("authtoken")
	//normally Authorization the_token_xxx

	return bearToken
}

func VerifyAccessToken(r *http.Request) (*jwt.Token, error) {
	tokenString := ExtractAccessToken(r)

	token, veritokparserr := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		//Make sure that the token method conform to "SigningMethodHMAC"

		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(mySigningKey), nil
	})
	if veritokparserr != nil {
		return nil, WrapError("veritokparserr", veritokparserr.Error())
	}
	return token, nil
}

func ExtractAccessTokenMetadata(r *http.Request) (*AccessDetails, error) {
	token, extacctokenmetaerr := VerifyAccessToken(r)
	if extacctokenmetaerr != nil {
		return nil, WrapError("extacctokenmetaerr", extacctokenmetaerr.Error())
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if ok && token.Valid {
		accessUuid, ok := claims["access_uuid"].(string)
		if !ok {
			return nil, WrapError("extacctokenmetavaliderr", extacctokenmetaerr.Error())
		}
		//userId, err := strconv.ParseUint(fmt.Sprintf("%.f", claims["user_id"]), 10, 64)
		userId := claims["user_id"].(string)
		return &AccessDetails{
			AccessUuid: accessUuid,
			UserId:     userId,
		}, nil
	}
	return nil, extacctokenmetaerr
}

func FetchAccessAuth(authD *AccessDetails) (string, error) {
	userid, fetaccautherr := rdb.Get(authD.AccessUuid).Result()
	if fetaccautherr != nil {
		return "", WrapError("fetaccautherr", fetaccautherr.Error())
	}
	//userID, _ := strconv.ParseUint(userid, 10, 64)

	return userid, nil
}

func CreateAccessAuth(userid string, td *AccessTokenDetails) error {
	at := time.Unix(td.AtExpires, 0) //converting Unix to UTC(to Time object)
	now := time.Now()
	creaccautherr := rdb.Set(td.AccessUuid, userid, at.Sub(now)).Err()
	if creaccautherr != nil {
		return WrapError("creaccautherr", creaccautherr.Error())
	}
	return nil
}

/*
	func ResetAccessAuth(authD *AccessDetails) error {
		//converting Unix to UTC(to Time object)
		rt := time.Duration(time.Second * 240)
		//.AtExpires = time.Now().Add(time.Minute * 15).Unix()

		resetaccautherr := rdb.Expire(authD.AccessUuid, rt).Err()
		if resetaccautherr != nil {
			return WrapError("delaccautherr", resetaccautherr.Error())
		}
		return nil
	}
*/
func DeleteAccessAuth(givenAccessUuid string) (int64, error) {
	accessdeleted, delaccautherr := rdb.Del(givenAccessUuid).Result()
	if delaccautherr != nil {
		return 0, WrapError("delaccautherr", delaccautherr.Error())
	}
	return accessdeleted, nil
}

func ConnectToRedis() {
	//Initializing redis
	var redisoption = redis.Options{
		Addr:     env.RedisAddr,
		Password: env.RedisPass, // no password set
		DB:       0,             // use default DB
	}

	rdb = redis.NewClient(&redisoption)
	conrediserr := ping(rdb)
	if conrediserr != nil {
		logger.Error("Error", logger.String("conrediserr", conrediserr.Error()))
		panic(conrediserr)
	}
}

///////////////    Mongo Database   //////////////

func DeleteIdMongo(deleteDb *mongo.Collection, urmyuuid string) (*mongo.DeleteResult, error) {
	ctx, cf := context.WithTimeout(context.Background(), 10*time.Second)
	defer cf()
	deleteResult, deleteidonemongoerr := deleteDb.DeleteOne(ctx, bson.M{
		"useruid": urmyuuid,
	})
	if deleteidonemongoerr != nil {
		return nil, WrapError("signinuserreaderr", deleteidonemongoerr.Error())
	}

	return deleteResult, nil
}

func DeleteIdFromFriendMongo(deleteDb *mongo.Collection, urmyuuid string) (*mongo.DeleteResult, error) {
	ctx, cf := context.WithTimeout(context.Background(), 10*time.Second)
	defer cf()
	deletefriendResult, deletefriendmongoerr := deleteDb.DeleteMany(ctx, bson.M{
		"friend": urmyuuid,
	})
	if deletefriendmongoerr != nil {
		return nil, WrapError("deletefriendmongoerr", deletefriendmongoerr.Error())
	}

	_, deletehostmongoerr := deleteDb.DeleteMany(ctx, bson.M{
		"host": urmyuuid,
	})
	if deletehostmongoerr != nil {
		return nil, WrapError("deletehostmongoerr", deletehostmongoerr.Error())
	}
	return deletefriendResult, nil
}

func DeleteIdFromBlackMongo(deleteDb *mongo.Collection, urmyuuid string) (*mongo.DeleteResult, error) {
	ctx, cf := context.WithTimeout(context.Background(), 10*time.Second)
	defer cf()
	deleteblackResult, deleteblackmongoerr := deleteDb.DeleteMany(ctx, bson.M{
		"blacklist": urmyuuid,
	})
	if deleteblackmongoerr != nil {
		return nil, WrapError("deleteblackmongoerr", deleteblackmongoerr.Error())
	}

	_, deleteblackhostmongoerr := deleteDb.DeleteMany(ctx, bson.M{
		"blacklist": urmyuuid,
	})
	if deleteblackhostmongoerr != nil {
		return nil, WrapError("deleteblackhostmongoerr", deleteblackhostmongoerr.Error())
	}
	return deleteblackResult, nil
}
func SignUpMongo(ctx context.Context, writeDb *mongo.Collection, userUid string, notification string, platform string) (*mongo.InsertOneResult, error) {

	insertResult, signupmongoinserterr := writeDb.InsertOne(ctx, bson.M{
		"useruid":           userUid,
		"notificationtoken": notification,
		"platform":          platform,
		"lastlogin":         time.Now().Format("2006-01-02 15:04:05"),
	})
	if signupmongoinserterr != nil {
		return nil, WrapError("signupmongoinserterr", signupmongoinserterr.Error())
	}
	return insertResult, nil
}

func SignInMongo(writeDb *mongo.Collection, user *SignInResult, notification string, platform string) (*mongo.UpdateResult, error) {
	ctx, cf := context.WithTimeout(context.Background(), 10*time.Second)
	defer cf()

	userinfo, signinuserreaderr := UserRead(writeDb, strings.TrimSpace(user.LoginUuid))
	if signinuserreaderr != nil {
		_, reinserterr := SignUpMongo(ctx, writeDb, strings.TrimSpace(user.LoginUuid), notification, platform)
		if reinserterr != nil {
			return nil, WrapError("reinserterr", reinserterr.Error())
		}
		userinfo, reinserterr = UserRead(writeDb, strings.TrimSpace(user.LoginUuid))
		if reinserterr != nil {
			return nil, WrapError("reinserterr", reinserterr.Error())
		}
	}

	lastlogin, lastloginparserr := time.Parse("2006-01-02 15:04:05", userinfo.LastLogin)
	if lastloginparserr != nil {
		return nil, WrapError("lastlogintimeparseerr", lastloginparserr.Error())
	}

	//log.Println(lastlogin)
	//log.Println(julgy)
	if a.sa.CheckUpdateUrMyDaeSaeUn(lastlogin) {
		year, signinyearparserr := strconv.Atoi(user.Birthday[0:4])
		if signinyearparserr != nil {
			return nil, WrapError("signinyearparserr", signinyearparserr.Error())
		}
		//log.Println(user.Birthday[0:4])
		month, signinmonthparserr := strconv.Atoi(user.Birthday[4:6])
		if signinmonthparserr != nil {
			return nil, WrapError("signinmonthparserr", signinmonthparserr.Error())
		}
		//log.Println(user.Birthday[4:6])
		day, signindayparserr := strconv.Atoi(user.Birthday[6:8])
		if signindayparserr != nil {
			return nil, WrapError("signindayparserr", signindayparserr.Error())
		}
		//log.Println(user.Birthday[6:8])
		hour, signinhourparserr := strconv.Atoi(user.Birthday[8:10])
		if signinhourparserr != nil {
			return nil, WrapError("signindayparserr", signinhourparserr.Error())
		}
		//log.Println(user.Birthday[8:10])
		min, signinminparserr := strconv.Atoi(user.Birthday[11:13])
		if signinminparserr != nil {
			return nil, WrapError("signinminparserr", signinminparserr.Error())
		}
		//log.Println(user.Birthday[11:13])
		//log.Printf("%d %d %d %d %d", year, month, day, hour, min)

		//log.Println(user.Birthday)

		saju := a.sa.ParseSaju(year, month, day, hour, min, user.Gender)

		saju.Year, saju.Month, saju.Day = year, month, day

		palja := a.sa.ExtractSaju(saju)
		dasaeun := a.sa.ExtractDaeUnSaeUn(saju, palja)

		signinupdatedaesaeunerr := UpdateUrMyDaeSaeUn(user.LoginUuid, *dasaeun)
		return nil, signinupdatedaesaeunerr
	}

	filter := bson.M{"useruid": strings.TrimSpace(user.LoginUuid)}
	update := bson.M{
		"$set": bson.M{
			"notificationtoken": notification,
			"platform":          platform,
			"lastlogin":         time.Now().Format("2006-01-02 15:04:05"),
		},
	}
	updateResult, signinupdateoneerr := writeDb.UpdateOne(ctx,
		filter,
		update,
	)
	if signinupdateoneerr != nil {
		return nil, WrapError("signinupdateoneerr", signinupdateoneerr.Error())
	}

	return updateResult, nil
}

func UserRead(readDb *mongo.Collection, urmyuuid string) (*SignInUser, error) {
	//ctx, cf := context.WithTimeout(context.Background(), 10*time.Second)
	//defer cf()
	ctx := context.Background()
	var signinuser SignInUser

	userfinderr := readDb.FindOne(ctx, bson.M{"useruid": urmyuuid}).Decode(&signinuser)
	if userfinderr != nil {
		return nil, WrapError("userfinderr", userfinderr.Error())
	}

	return &signinuser, nil
}

func HostRead(readDb *mongo.Collection, urmyuuid string) ([]string, error) {
	//ctx, cf := context.WithTimeout(context.Background(), 10*time.Second)
	//defer cf()
	ctx := context.Background()
	var hostlist []string
	var datas []bson.M
	result, hostfinderr := readDb.Find(ctx, bson.M{"friend": urmyuuid})
	if hostfinderr != nil {
		return nil, WrapError("hostfinderr", hostfinderr.Error())
	}
	hostallerr := result.All(ctx, &datas)
	if hostallerr != nil {
		return nil, WrapError("hostallerr", hostallerr.Error())
	}
	for _, data := range datas {

		hostlist = append(hostlist, data["host"].(string))
	}
	return hostlist, nil
}

func FriendRead(readDb *mongo.Collection, urmyuuid string) ([]FriendList, error) {
	//ctx, cf := context.WithTimeout(context.Background(), 10*time.Second)
	//defer cf()
	ctx := context.Background()
	var friendlist []FriendList
	var datas []bson.M
	result, friendfinderr := readDb.Find(ctx, bson.M{"host": urmyuuid})
	if friendfinderr != nil {
		return nil, WrapError("friendreadfinderr", friendfinderr.Error())
	}
	friendallerr := result.All(ctx, &datas)
	if friendallerr != nil {
		return nil, WrapError("friendallerr", friendallerr.Error())
	}
	for _, data := range datas {

		friendlist = append(friendlist, FriendList{
			HostUUID:      data["host"].(string),
			FriendUUID:    data["friend"].(string),
			HostGrade:     data["hostgrade"].(string),
			OpponentGrade: data["opponentgrade"].(string),
			ResultRecord:  data["resultrecord"].(primitive.A),
		})
	}
	return friendlist, nil
}

func FriendWrite(writeDb *mongo.Collection, urmyuuid string, concord FriendList) error {
	ctx, cf := context.WithTimeout(context.Background(), 10*time.Second)
	defer cf()

	insertResult, friendadderr := writeDb.InsertOne(ctx, bson.M{
		"host":          urmyuuid,
		"friend":        concord.FriendUUID,
		"hostgrade":     concord.HostGrade,
		"opponentgrade": concord.OpponentGrade,
		"resultrecord":  concord.ResultRecord,
	})
	if friendadderr != nil {
		return WrapError("friendadderr", friendadderr.Error())
	}
	fmt.Printf("Inserted a single document: %s, %s\n", insertResult.InsertedID, concord.FriendUUID)
	return nil
}

func FriendDelete(deleteDb *mongo.Collection, urmyuuid string, friend FriendList) error {
	ctx, cf := context.WithTimeout(context.Background(), 10*time.Second)
	defer cf()
	deleteResult, frienddelonerr := deleteDb.DeleteOne(ctx, bson.M{
		"host":   urmyuuid,
		"friend": friend.FriendUUID,
	})
	if frienddelonerr != nil {
		return WrapError("frienddelonerr", frienddelonerr.Error())
	}
	fmt.Printf("Deleted %v documents in the trainers collection\n", deleteResult.DeletedCount)
	return nil
}

func BlacklistRead(readDb *mongo.Collection, urmyuuid string) ([]BlackList, error) {
	//ctx, cf := context.WithTimeout(context.Background(), 10*time.Second)
	//defer cf()
	ctx := context.Background()
	var blacklist []BlackList
	var datas []bson.M
	result, blackfinderr := readDb.Find(ctx, bson.M{"host": urmyuuid})
	if blackfinderr != nil {
		return nil, WrapError("blackfinderr", blackfinderr.Error())

	}

	blackallerr := result.All(ctx, &datas)
	if blackallerr != nil {
		return nil, WrapError("blackallerr", blackallerr.Error())
	}
	for _, data := range datas {
		blacklist = append(blacklist, BlackList{
			HostUUID:      data["host"].(string),
			BlacklistUUID: data["blacklist"].(string),
			HostGrade:     data["hostgrade"].(string),
			OpponentGrade: data["opponentgrade"].(string),
		})
	}
	return blacklist, nil
}

func BlacklistWrite(writeDb *mongo.Collection, urmyuuid string, blacklist BlackList) error {
	ctx, cf := context.WithTimeout(context.Background(), 10*time.Second)
	defer cf()

	insertResult, blackaddoneerr := writeDb.InsertOne(ctx, bson.M{
		"host":          urmyuuid,
		"blacklist":     blacklist.BlacklistUUID,
		"hostgrade":     blacklist.HostGrade,
		"opponentgrade": blacklist.OpponentGrade,
	})
	if blackaddoneerr != nil {
		return WrapError("blackaddoneerr", blackaddoneerr.Error())
	}
	fmt.Printf("Inserted a single document: %s, %s\n", insertResult.InsertedID, blacklist.BlacklistUUID)
	return nil
}

func BlacklistDelete(deleteDb *mongo.Collection, urmyuuid string, blacklist BlackList) error {
	ctx, cf := context.WithTimeout(context.Background(), 10*time.Second)
	defer cf()
	deleteResult, blackdeloneerr := deleteDb.DeleteOne(ctx, bson.M{
		"host":      urmyuuid,
		"blacklist": blacklist.BlacklistUUID,
	})
	if blackdeloneerr != nil {
		return WrapError("blackdeloneerr", blackdeloneerr.Error())
	}
	fmt.Printf("Deleted %v documents in the trainers collection\n", deleteResult.DeletedCount)
	return nil
}

func MongoConn() {
	var mongoconnerr error
	clientOptions := options.Client().ApplyURI(env.MongoUrl)
	mongoconn, mongoconnerr = mongo.Connect(context.TODO(), clientOptions)
	if mongoconnerr != nil {
		logger.Error("Error", logger.String("mongoconnerr: %v", mongoconnerr.Error()))
		panic(mongoconnerr)
	}

	// Check the connection
	mongoconnpingerr := mongoconn.Ping(context.TODO(), nil)
	if mongoconnpingerr != nil {
		logger.Error("Error", logger.String("mongoconnpingerr: %v", mongoconnpingerr.Error()))
		panic(mongoconnpingerr)
	} else {
		fmt.Println("MongoDB Connection Made")
	}

}

/*
// ///////////////// In App Purchase ///////////////////////
var appstoreclient *appstore.Client
var googleplayclient *playstore.Client

func InAppAppStoreConnect() {
	appstoreclient = appstore.New()
	//appstoreclient.Verify()
}

func InAppGooglePlayConnect() {
	jsonKey, keyjsonreaderr := ioutil.ReadFile("jsonKey.json")
	if keyjsonreaderr != nil {
		log.Fatal(keyjsonreaderr)
	}
	var playstorerr error
	googleplayclient, playstorerr = playstore.New(jsonKey)
	if playstorerr != nil {
		log.Fatal(playstorerr)
	}
	googleplayclient.AcknowledgeProduct()
	googleplayclient.AcknowledgeSubscription()
	googleplayclient.CancelSubscription()
	googleplayclient.RefundSubscription()
	googleplayclient.RevokeSubscription()
	googleplayclient.VerifyProduct()
	googleplayclient.VerifySubscription()
	googleplayclient.VerifySubscriptionV2()
}
*/
///////////////////   utilities   //////////////////////
func hash(s string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()
}

func checkTheValueThreatable(value string) bool {
	if strings.Contains(strings.ToLower(value), "select") && strings.Contains(strings.ToLower(value), "from") {
		return true
	}
	if strings.Contains(strings.ToLower(value), "delete from") {
		return true
	}
	if strings.Contains(strings.ToLower(value), "update") && strings.Contains(strings.ToLower(value), "set") {
		return true
	}
	if strings.Contains(strings.ToLower(value), "insert into") {
		return true
	}
	if strings.Contains(strings.ToLower(value), "'") {
		return true
	}
	if strings.Contains(strings.ToLower(value), ";") {
		return true
	}
	if strings.Contains(strings.ToLower(value), "\"") {
		return true
	}
	return false
}

func makeBirthdateSliceUnique(s []Birthdate) []Birthdate {
	keys := make(map[string]struct{})
	res := make([]Birthdate, 0)
	for _, val := range s {
		if _, ok := keys[val.Birthday]; ok {
			continue
		} else {
			keys[val.Birthday] = struct{}{}
			res = append(res, val)
		}
	}
	return res
}
func generateuuid() string {
	var gen uuid.UUIDv7Generator

	gen.CounterPrecisionLength = 12

	id := gen.Next()

	return strings.ReplaceAll(id.ToString(), "-", "")
}

/*
func generateuuid() (string, error) {
	u, err := uuid.NewV4()
	if err != nil {
		return "", err
	} else {
		return u.String(), nil
	}
}
*/
/*
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
*/
func ping(client *redis.Client) error {
	pong, redispingerr := client.Ping().Result()
	if redispingerr != nil {
		fmt.Println(redispingerr)
		return WrapError("redispingerr", redispingerr.Error())

	}
	fmt.Println("Redis: " + pong)
	// Output: PONG <nil>

	return nil
}

func ReadCsv(f io.ReadCloser) ([][]string, error) {

	// Read File into a Variable
	lines, readcsverr := csv.NewReader(f).ReadAll()
	if readcsverr != nil {
		return [][]string{}, WrapError("readcsverr", readcsverr.Error())
	}

	return lines, nil
}

/*
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
*/
/////////////////////AWS//////////////////////////////

func PutProfilePicS3(buffer []byte, userId string, tmpfilename string) error {
	putinput := s3.PutObjectInput{
		Bucket:             aws.String("urmy-tf-cloudfront-s3"),
		Key:                aws.String("users/pic/" + strings.ReplaceAll(userId, "-", "") + "/" + tmpfilename),
		Body:               bytes.NewReader(buffer),
		ContentType:        aws.String(http.DetectContentType(buffer)),
		ContentDisposition: aws.String("attachment"),
	}

	_, putpropics3putobjerr := s3client.PutObject(context.Background(), &putinput)
	if putpropics3putobjerr != nil {

		return WrapError("putpropics3putobjerr", putpropics3putobjerr.Error())
	}
	return nil
}

func DeleteProfilePicS3(userId string, tmpfilename string) error {
	deleteinput := s3.DeleteObjectInput{
		Bucket: aws.String("urmy-tf-cloudfront-s3"),
		Key:    aws.String("users/pic/" + strings.ReplaceAll(userId, "-", "") + "/" + tmpfilename),
	}

	_, delpropics3delobjerr := s3client.DeleteObject(context.Background(), &deleteinput)
	if delpropics3delobjerr != nil {
		return WrapError("delpropics3delobjerr", delpropics3delobjerr.Error())
	}
	return nil
}

func GetConfig() *aws.Config {
	cfg, getcfgloadcfgerr := config.LoadDefaultConfig(context.Background(), config.WithRegion("ap-northeast-2"))
	if getcfgloadcfgerr != nil {
		logger.Error("Error", logger.String("getcfgloadcfgerr: %v", getcfgloadcfgerr.Error()))
	}
	return &cfg
}

/////////////////////logger////////////////////

func WrapError(errortype string, message string) error {
	return errors.New(errortype + ": " + message)
}

//////////////////// struct ///////////////

type SendMessage struct {
	ChatroomId  string `json:"chatroomid"`
	MsgUser     string `json:"msguser"`
	MsgContent  string `json:"msgcontent"`
	MsgSentdate int    `json:"sentdate"`
	MsgSender   string `json:"msgsender"`
	MsgType     string `json:"msgtype"`
}

type ServiceList struct {
	ServiceUUID  string `json:"ServiceUUID" db:"serviceid"`
	ServiceName  string `json:"ServiceName" db:"servicename"`
	ServicePrice string `json:"ServicePrice" db:"serviceprice"`
}

type ServiceSales struct {
	InvoiceId     string `db:"invoiceid"`
	ServiceId     string `json:"ServiceId" db:"serviceid"`
	PurchasedUser string `db:"purchaseuser"`
	PurchasedDate string `json:"PurchasedDate" db:"purchasedate"`
}

type RefundRequestList struct {
	RefundUUID  string `json:"refundid"`
	InvoiceUUID string `json:"invoiceid"`
	RefundUser  string `json:"refunduser"`
}

type SignUpUser struct {
	LoginID string `json:"loginId"`
	Useruid string `json:"userUid"`
}

type UpdateProfile struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type SignInUser struct {
	DeviceOS                  string `json:"deviceos"`
	LoginID                   string `json:"loginId" bson:"loginId"`
	Useruuid                  string `json:"userUid" bson:"userUid"`
	FirebaseNotificationToken string `json:"notificationtoken" bson:"notificationtoken"`
	UrMyPlatform              string `json:"urmyplatform" bson:"platform"`
	LastLogin                 string `json:"lastlogin" bson:"lastlogin"`
}

type InsertUserProfile struct {
	DeviceOS    string `json:"deviceos"`
	LoginID     string `json:"loginId"`
	Useruid     string `json:"userUid"`
	Name        string `json:"name"`
	PhoneNo     string `json:"phoneNo"`
	Phonecode   string `json:"phonecode"`
	Gender      bool   `json:"genderState"`
	Birthdate   string `json:"birthdate"`
	Description string `json:"description"`
	Country     string `json:"country"`
	City        string `json:"city"`
	Mbti        string `json:"mbti"`
	Isoverage   bool   `json:"isoverage"`
	Privacy     bool   `json:"privacy"`
	Service     bool   `json:"service"`
	Illegal     bool   `json:"illegal"`
}

type AccessDetails struct {
	AccessUuid string
	UserId     string
}

type AccessTokenDetails struct {
	AccessToken string
	AccessUuid  string
	AtExpires   int64
}

type Userinfo struct {
	LoginUuid      string `db:"loginuuid"`
	Email          string `db:"email"`
	PhoneNo        string `db:"phoneno"`
	PhoneCode      string `db:"phonecode"`
	Name           string `db:"name"`
	Birthdate      string `db:"birthday"`
	Gender         bool   `db:"gender"`
	Mbti           string `db:"mbti"`
	Description    string `db:"description"`
	SajuPalja      saju.SaJuPalJa
	CurrentCountry sql.NullString `db:"residentcountry"`
	CurrentState   sql.NullString `db:"residentstate"`
	CurrentCity    sql.NullString `db:"residentcity"`
	ProfilePicPath []ProfilePic
}

type UserinfoToDelete struct {
	LoginUuid string `db:"loginuuid"`
	Gender    bool   `db:"gender"`
	Birthdate string `db:"birthday"`
	Country   string `db:"country"`
	City      string `db:"city"`
	Mbti      string `db:"mbti"`
}

type SignInResult struct {
	LoginUuid string `db:"loginuuid"`
	DeviceOS  string `db:"deviceos"`
	Birthday  string `db:"birthday"`
	Gender    bool   `db:"gender"`
}

type ProfilePic struct {
	Filename        string    `json:"Filename" db:"profilename"`
	UploadedDate    time.Time `json:"UploadedDate" db:"updateddate"`
	ProfileSequence int       `json:"ProfileSequence" db:"profilesequence"`
}

type PersonSaju struct {
	LoginUuid string `json:"loginuuid"`
	YearChun  string `json:"YearChun"`
	YearJi    string `json:"YearJi"`
	MonthChun string `json:"MonthChun"`
	MonthJi   string `json:"MonthJi"`
	DayChun   string `json:"DayChun"`
	DayJi     string `json:"DayJi"`
	TimeChun  string `json:"TimeChun"`
	TimeJi    string `json:"TimeJi"`
	DaeunChun string `json:"DaeunChun"`
	DaeUnJi   string `json:"DaeUnJi"`
	SaeunChun string `json:"SaeunChun"`
	SaeunJi   string `json:"SaeunJi"`
}

type GetCandidate struct {
	Gender    bool   `db:"gender"`
	Birthdate string `db:"birthday"`
	Tester    bool   `db:"tester"`
}

type CandidateList struct {
	CandidateUUID      string `bson:"loginuuid" db:"loginuuid"`
	CandidateName      string `db:"name"`
	CandidateBirthDate string `db:"birthday"`
	CandidateGender    bool   `db:"gender"`
	CandidateCountry   string `db:"residentcountry"`
	CandidateState     string `db:"residentstate"`
	CandidateCity      string `db:"residentcity"`
	CandidateMBTI      string `db:"mbti"`
	Description        string `db:"description"`
	HostGrade          string
	OpponentGrade      string
	ResultRecord       interface{}
	ContentUrl         string
	ProfilePicPath     []ProfilePic
}

type FriendList struct {
	HostUUID      string      `bson:"host"`
	FriendUUID    string      `json:"CandidateUUID" bson:"friend"`
	HostGrade     string      `json:"HostGrade" bson:"hostgrade"`
	OpponentGrade string      `json:"OpponentGrade" bson:"opponentgrade"`
	ResultRecord  interface{} `json:"ResultRecord" bson:"resultrecord"`
}

type BlackList struct {
	HostUUID      string `bson:"host"`
	BlacklistUUID string `json:"CandidateUUID" bson:"friend"`
	HostGrade     string `json:"HostGrade" bson:"hostgrade"`
	OpponentGrade string `json:"OpponentGrade" bson:"opponentgrade"`
}

type Birthdate struct {
	Birthday      string `db:"birthday"`
	Urmyyear      int    `db:"urmyyear"`
	Urmymonth     int    `db:"urmymonth"`
	Urmyday       int    `db:"urmyday"`
	Urmytime      string `db:"urmytime"`
	Grade         float64
	HostGrade     float64
	OpponentGrade float64
}
