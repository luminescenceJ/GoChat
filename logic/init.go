package logic

import (
	"Go-Chat/common/e"
	"Go-Chat/config"
	"Go-Chat/tools"
	"github.com/go-redis/redis"
	"github.com/sirupsen/logrus"
	"golang.org/x/time/rate"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
	"strings"
	"time"
)

var (
	RedisClient *redis.Client
	GormDB      *gorm.DB
)

func (logic *Logic) InitRedisClient() (err error) {
	redisOpt := tools.RedisOption{
		Address:  config.Conf.Redis.RedisAddress,
		Password: config.Conf.Redis.RedisPassword,
		Db:       config.Conf.Redis.Db,
	}
	RedisClient = tools.GetRedisInstance(redisOpt)
	if pong, err := RedisClient.Ping().Result(); err != nil {
		logrus.Infof("RedisCli Ping Result pong: %s,  err: %s", pong, err)
	}
	return err
}

func (logic *Logic) InitRpcServer() (err error) {
	var network, addr string
	// a host multi port case
	rpcAddressList := strings.Split(config.Conf.Logic.RpcAddress, ",")
	for _, bind := range rpcAddressList {
		if network, addr, err = tools.ParseNetwork(bind); err != nil {
			logrus.Panicf("InitLogicRpc ParseNetwork error : %s", err.Error())
		}
		logrus.Infof("logic start run at-->%s:%s", network, addr)
		go logic.createRpcServer(network, addr)
	}
	return
}

func (logic *Logic) InitMysql() {
	ormLogger := logger.Default // 数据库日志打印
	conf := config.Conf.Mysql
	dsn := conf.Username + ":" + conf.Password + "@tcp(" + conf.Host + ":" + conf.Port + ")/" + conf.DBName + "?" + conf.Config
	var err error
	GormDB, err = gorm.Open(mysql.New(mysql.Config{
		DSN:                       dsn,   // DSN data source name
		DefaultStringSize:         256,   // string 类型字段的默认长度
		DisableDatetimePrecision:  true,  // 禁用 datetime 精度，MySQL 5.6 之前的数据库不支持
		DontSupportRenameIndex:    true,  // 重命名索引时采用删除并新建的方式，MySQL 5.7 之前的数据库和 MariaDB 不支持重命名索引
		DontSupportRenameColumn:   true,  // 用 `change` 重命名列，MySQL 8 之前的数据库和 MariaDB 不支持重命名列
		SkipInitializeWithVersion: false, // 根据版本自动配置
	}), &gorm.Config{
		Logger: ormLogger,
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
	})
	if err != nil {
		logrus.Fatal(err)
		return
	}

	sqlDB, _ := GormDB.DB()
	sqlDB.SetMaxIdleConns(20)  //设置连接池，空闲
	sqlDB.SetMaxOpenConns(100) //打开
	sqlDB.SetConnMaxLifetime(time.Second * 30)

	// 慢日志中间件
	SlowQueryLog(GormDB)
	// 限流器中间件
	GormRateLimiter(GormDB, rate.NewLimiter(500, 1000))
}

// SlowQueryLog 慢查询日志
func SlowQueryLog(db *gorm.DB) {
	err := db.Callback().Query().Before("*").Register("slow_query_start", func(d *gorm.DB) {
		now := time.Now()
		d.Set("start_time", now)
	})
	if err != nil {
		panic(err)
	}

	err = db.Callback().Query().After("*").Register("slow_query_end", func(d *gorm.DB) {
		now := time.Now()
		start, ok := d.Get("start_time")
		if ok {
			duration := now.Sub(start.(time.Time))
			// 一般认为 200 Ms 为Sql慢查询
			if duration > time.Millisecond*200 {
				logrus.Error("慢查询", "SQL:", d.Statement.SQL.String())
			}
		}
	})
	if err != nil {
		panic(err)
	}
}

// GormRateLimiter Gorm限流器 此限流器不能终止GORM查询链。
func GormRateLimiter(db *gorm.DB, r *rate.Limiter) {
	err := db.Callback().Query().Before("*").Register("RateLimitGormMiddleware", func(d *gorm.DB) {
		if !r.Allow() {
			err := d.AddError(e.GormToManyRequestError)
			if err != nil {
				return
			}
			logrus.Error(e.GormToManyRequestError.Error())
			return
		}
	})
	if err != nil {
		panic(err)
	}
}
