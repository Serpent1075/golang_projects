package config

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	logger "github.com/whitewhale1075/urmylogger"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gopkg.in/redis.v5"
)

type Handler struct {
	ReadDB  *sqlx.DB
	WriteDB *sqlx.DB
	RedisDB *redis.Client
	MongoDB *mongo.Client
}

func ConnectToReadDB(readhost string, readport int, readuser string, readpasswd string, readdbname string) *sqlx.DB {
	readpsqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+"password=%s dbname=%s sslmode=disable", readhost, readport, readuser, readpasswd, readdbname)
	readdb, readdberr := sqlx.Open("postgres", readpsqlInfo)
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
	return readdb
}

func ConnectToWriteDB(writehost string, writeport int, writeuser string, writepasswd string, writedbname string) *sqlx.DB {

	writepsqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+"password=%s dbname=%s sslmode=disable", writehost, writeport, writeuser, writepasswd, writedbname)
	writedb, writedberr := sqlx.Open("postgres", writepsqlInfo)
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
	return writedb
}

func ConnectToRedis(readaddr string, readpsswd string) *redis.Client {
	//Initializing redis
	var redisoption = redis.Options{
		Addr:     readaddr,
		Password: readpsswd, // no password set
		DB:       0,         // use default DB
	}

	rdb := redis.NewClient(&redisoption)
	conrediserr := ping(rdb)
	if conrediserr != nil {
		logger.Error("Error", logger.String("conrediserr", conrediserr.Error()))
		panic(conrediserr)
	}
	return rdb
}

func ping(client *redis.Client) error {
	pong, redispingerr := client.Ping().Result()
	if redispingerr != nil {
		fmt.Println(redispingerr)
		panic(redispingerr)
	}
	fmt.Println("Redis: " + pong)
	// Output: PONG <nil>

	return nil
}

func ConnectToMongo(mongohost string) *mongo.Client {

	clientOptions := options.Client().ApplyURI(mongohost)
	mongoconn, mongoconnerr := mongo.Connect(context.TODO(), clientOptions)
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
	return mongoconn
}

func Init() *Handler {
	host := "192.168.92.130"
	psqlport := 5432
	psqluser := "koreaogh"
	psqlpw := "test1234"
	psqldb := "urmydb"
	redisport := "6379"
	redispasswd := "qwer1234"
	mongoport := "27017"

	readdb := ConnectToReadDB(host, psqlport, psqluser, psqlpw, psqldb)
	writedb := ConnectToWriteDB(host, psqlport, psqluser, psqlpw, psqldb)
	redisdb := ConnectToRedis(host+":"+redisport, redispasswd)
	mongodb := ConnectToMongo("mongodb://" + host + ":" + mongoport)

	return &Handler{
		ReadDB:  readdb,
		WriteDB: writedb,
		RedisDB: redisdb,
		MongoDB: mongodb,
	}
}
