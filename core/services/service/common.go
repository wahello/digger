///////////////////////////////////////////
// Copyright(C) 2020
// Author : Jason He
// Version: 0.0.1
///////////////////////////////////////////
package service

import (
	"context"
	"database/sql"
	"digger/services"
	"digger/utils"
	"fmt"
	"github.com/go-redis/redis/v7"
	"github.com/hetianyi/gox"
	"github.com/hetianyi/gox/logger"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"sync"
)

var (
	dbConn               *gorm.DB
	RedisClient          *redis.Client
	initLock             = new(sync.Mutex)
	cacheService         services.CacheService
	projectService       services.ProjectService
	projectConfigService services.ProjectConfigService
	resultService        services.ResultService
	taskService          services.TaskService
	queueService         services.QueueService
	pluginService        services.PluginService
	dbService            services.DBService
	configService        services.ConfigService
	statisticService     services.StatisticService
	proxyService         services.ProxyService
	pushService          services.PushSourceService
)

// 初始化数据库连接
func InitDb(connString string) {
	db, err := gorm.Open("postgres", connString)
	if err != nil {
		logger.Fatal(err)
	}
	//	db.LogMode(true)
	dbConn = db
	logger.Info("数据库连接成功")
}

func InitRedis(connString string) {

	redisConfig := utils.ParseRedisConnStr(connString)
	if redisConfig == nil {
		logger.Fatal("redis配置错误: ", connString)
	}

	RedisClient = redis.NewClient(&redis.Options{
		// Addr: "39.101.143.224:20021",
		Addr:     redisConfig.Address,
		Password: redisConfig.Password, // no password set
		DB:       redisConfig.DB,       // use default DB
	})
	_, err := RedisClient.Ping().Result()
	if err != nil {
		logger.Fatal(err)
	}
	logger.Info("Redis连接成功")
}

func DBService() services.DBService {
	if dbService == nil {
		initLock.Lock()
		defer initLock.Unlock()
		if dbService == nil {
			dbService = &dbServiceImp{}
		}
	}
	return dbService
}

func CacheService() services.CacheService {
	if cacheService == nil {
		initLock.Lock()
		defer initLock.Unlock()
		if cacheService == nil {
			cacheService = &cacheServiceImp{
				cache:     make(map[string]interface{}),
				cacheLock: new(sync.Mutex),
			}
		}
	}
	return cacheService
}

func ProjectService() services.ProjectService {
	if projectService == nil {
		initLock.Lock()
		defer initLock.Unlock()
		if projectService == nil {
			projectService = &projectServiceImp{}
		}
	}
	return projectService
}

func ProjectConfigService() services.ProjectConfigService {
	if projectConfigService == nil {
		initLock.Lock()
		defer initLock.Unlock()
		if projectConfigService == nil {
			projectConfigService = &projectConfigServiceImp{}
		}
	}
	return projectConfigService
}

func ResultService() services.ResultService {
	if resultService == nil {
		initLock.Lock()
		defer initLock.Unlock()
		if resultService == nil {
			resultService = &resultServiceImp{}
		}
	}
	return resultService
}

func TaskService() services.TaskService {
	if taskService == nil {
		initLock.Lock()
		defer initLock.Unlock()
		if taskService == nil {
			taskService = &taskServiceImp{}
		}
	}
	return taskService
}

func QueueService() services.QueueService {
	if queueService == nil {
		initLock.Lock()
		defer initLock.Unlock()
		if queueService == nil {
			queueService = &queueServiceImpl{}
		}
	}
	return queueService
}

func PluginService() services.PluginService {
	if pluginService == nil {
		initLock.Lock()
		defer initLock.Unlock()
		if pluginService == nil {
			pluginService = &pluginServiceImp{}
		}
	}
	return pluginService
}

func ConfigService() services.ConfigService {
	if configService == nil {
		initLock.Lock()
		defer initLock.Unlock()
		if configService == nil {
			configService = &configServiceImpl{}
		}
	}
	return configService
}

func StatisticService() services.StatisticService {
	if statisticService == nil {
		initLock.Lock()
		defer initLock.Unlock()
		if statisticService == nil {
			statisticService = &statisticServiceImp{}
		}
	}
	return statisticService
}

func ProxyService() services.ProxyService {
	if proxyService == nil {
		initLock.Lock()
		defer initLock.Unlock()
		if proxyService == nil {
			proxyService = &proxyServiceImp{}
		}
	}
	return proxyService
}

func PushService() services.PushSourceService {
	if pushService == nil {
		initLock.Lock()
		defer initLock.Unlock()
		if pushService == nil {
			pushService = &pushServiceImp{}
		}
	}
	return pushService
}

func transformNotFoundErr(err error) error {
	if err == nil {
		return nil
	}
	if gorm.IsRecordNotFoundError(err) {
		return nil
	}
	return err
}

// 执行事务
func DoTransaction(work func(tx *gorm.DB) error) error {
	// 开始事务
	var err error
	tx := dbConn.BeginTx(context.Background(), &sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
		ReadOnly:  false,
	})
	gox.Try(func() {
		if err := work(tx); err != nil {
			panic(err)
		}
		// 提交事务
		if err := tx.Commit().Error; err != nil {
			panic(err)
		}
	}, func(e interface{}) {
		// 回滚
		tx.Rollback()
		err = e.(error)
		logger.Debug(fmt.Sprintf("rollback tx due to: %s", err.Error()))
	})
	return err
}
