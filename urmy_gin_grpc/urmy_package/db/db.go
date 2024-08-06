package db

import (
	"log"

	//"github.com/jackc/pgx/v5"

	"github.com/whitewhale1075/urmy_package/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Handler struct {
	DB *gorm.DB
	//Psql *pgx.Conn
	//Redis *redis.Client
	//Mongo *mongo.Client
}

func Init(dburl string, redisurl string, mongourl string) Handler {

	dbconn := PsqlConn(dburl)
	//redisconn := RedisConn(redisurl)
	//mongoconn := MongoConn(mongourl)

	return Handler{
		DB: dbconn,
		//Psql: dbconn,
		//Redis: redisconn,
		//Mongo: mongoconn,
	}
}

func PsqlConn(psqlurl string) *gorm.DB {
	db, err := gorm.Open(postgres.Open(psqlurl), &gorm.Config{})

	if err != nil {
		log.Fatalln(err)
	}

	db.AutoMigrate(&models.User{})
	return db
}

/*
func PsqlConn(psqlurl string) *pgx.Conn {
	dbconn, err := pgx.Connect(context.Background(), psqlurl)
	if err != nil {
		panic(err)
	}
	defer dbconn.Close(context.Background())
	return dbconn
}
*/
/*
func RedisConn(redisurl string) *redis.Client {
	//Initializing redis
	var redisoption = redis.Options{
		Addr:     redisurl,
		Password: "", // no password set
		DB:       0,  // use default DB
	}
	client := redis.NewClient(&redisoption)
	if pong := client.Ping(); pong.String() != "ping: Poing" {
		os.Exit(1)
	}

	return client
}

func MongoConn(mongourl string) *mongo.Client {
	clientOptions := options.Client().ApplyURI(mongourl)
	mongoconn, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		panic(err)
	}

	return mongoconn
}
*/
