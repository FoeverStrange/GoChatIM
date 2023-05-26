package utils

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	DB  *gorm.DB
	Red *redis.Client
)

func InitConfig() {
	viper.SetConfigName("app")
	viper.AddConfigPath("config")
	err := viper.ReadInConfig()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("config app inited")
	// fmt.Println("config mysql init:", viper.Get("mysql"))
}

func InitMySQL() {
	//自定义日志模板，打印sql语句
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold: time.Second, //慢SQL查询阈值
			LogLevel:      logger.Info, //级别
			Colorful:      true,        //彩色
		},
	)
	DB, _ = gorm.Open(mysql.Open(viper.GetString("mysql.dns")), &gorm.Config{Logger: newLogger})
	fmt.Println("MySQL inited")
	// user := models.UserBasic{}
	// DB.Find(&user)
	// fmt.Println(user)
}
func InitRedis() {
	Red = redis.NewClient(&redis.Options{
		Addr:         viper.GetString("redis.addr"),
		DB:           viper.GetInt("redis.DB"),
		Password:     viper.GetString("redis.password"),
		PoolSize:     viper.GetInt("redis.poolSize"),
		MinIdleConns: viper.GetInt("redis.minIdleConn"),
	})
	// pong, err := Red.Ping().Result()
	// if err != nil {
	// 	fmt.Println("init redis err, ", err)
	// } else {
	// 	fmt.Println("init redis ", pong)
	// }
}

const (
	PublishKey = "websocket"
)

// Publish发布消息到redis
func Publish(ctx context.Context, channel string, msg string) error {
	// var err error
	err := Red.Publish(ctx, channel, msg).Err()
	fmt.Println("publish..", msg)
	if err != nil {
		fmt.Println(err)
	}
	return err
}

// Subscribe订阅消息到redis
func Subscribe(ctx context.Context, channel string) (string, error) {
	sub := Red.Subscribe(ctx, channel)
	// fmt.Println("subscribe..", ctx)
	msg, err := sub.ReceiveMessage(ctx)
	fmt.Println("subscribe..", msg.Payload)
	if err != nil {
		fmt.Println("sub.ReceiveMessage ERRO", err)
	}
	fmt.Println("subscribe..", msg.Payload)
	return msg.Payload, err
}
