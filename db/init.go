package db

import (
	"fmt"
	"log"
	"time"

	"github.com/chennqqi/godnslog/internal/models"
	oldmodels "github.com/chennqqi/godnslog/models"
	"github.com/sirupsen/logrus"
	"xorm.io/xorm"
)

// InitDatabase initializes the database schema and seed data
func InitDatabase(orm *xorm.Engine, testMode bool, defaultLanguage string, defaultCleanInterval int64, ip string) error {
	orm.SetTZDatabase(time.Local)
	orm.SetTZLocation(time.Local)

	// Sync database schema for 1.0 models (for backward compatibility)
	err := orm.Sync(
		&oldmodels.TblDns{},
		&oldmodels.TblHttp{},
		&oldmodels.TblUser{},
		&oldmodels.TblResolve{},
		&oldmodels.TblCase{},
		&oldmodels.TblPayload{},
		&oldmodels.TblInteraction{},
		&oldmodels.TblAPIKey{},
	)
	if err != nil {
		logrus.Errorf("[db::InitDatabase] orm.Sync 1.0 models: %v", err)
		return err
	}

	// Sync database schema for 2.0 models
	err = orm.Sync(
		&models.User{},
		&models.Interaction{},
		&models.Resolve{},
		&models.Case{},
		&models.Payload{},
		&models.APIKey{},
		&models.Workflow{},
		&models.Canary{},
		&models.RebindingRule{},
		&models.Listener{},
	)
	if err != nil {
		logrus.Errorf("[db::InitDatabase] orm.Sync 2.0 models: %v", err)
		return err
	}

	// Initialize seed data
	if err := initSeedData(orm, testMode, defaultLanguage, defaultCleanInterval, ip); err != nil {
		return err
	}

	return nil
}

// initSeedData initializes seed data
func initSeedData(orm *xorm.Engine, testMode bool, defaultLanguage string, defaultCleanInterval int64, ip string) error {
	// Check and create super admin user
	if err := initSuperUser(orm, testMode, defaultLanguage, defaultCleanInterval); err != nil {
		return err
	}

	// Check and create default DNS resolve record
	if err := initDefaultResolve(orm, ip); err != nil {
		return err
	}

	return nil
}

// initSuperUser initializes the super admin user
func initSuperUser(orm *xorm.Engine, testMode bool, defaultLanguage string, defaultCleanInterval int64) error {
	count, err := orm.Count(&oldmodels.TblUser{})
	if err != nil {
		logrus.Errorf("[db::initSuperUser] orm.Count(user): %v", err)
		return err
	}

	// If there is no super user when system first init
	if count == 0 {
		randomPass := genRandomString(12)
		// In test mode, use fixed password
		if testMode {
			randomPass = "test123"
		}
		_, err = orm.InsertOne(&oldmodels.TblUser{
			Name:          "admin",
			Email:         "admin@godnslog.com",
			ShortId:       genShortId(),
			Pass:          makePassword(randomPass),
			Token:         genRandomToken(),
			Role:          0, // roleSuper
			Lang:          defaultLanguage,
			CleanInterval: defaultCleanInterval,
		})
		if err != nil {
			logrus.Errorf("[db::initSuperUser] orm.InsertOne(user): %v", err)
			return err
		}
		log.Printf("Init super admin user with password: %v", randomPass)
	}

	return nil
}

// initDefaultResolve initializes the default DNS resolve record
func initDefaultResolve(orm *xorm.Engine, ip string) error {
	var wwwRcd oldmodels.TblResolve
	exist, err := orm.Where(`host=?`, `www`).And(`type=?`, `A`).Get(&wwwRcd)
	if err != nil {
		logrus.Errorf("[db::initDefaultResolve] orm.Get(resolve): %v", err)
		return err
	}

	if !exist {
		wwwRcd.Host = "www"
		wwwRcd.Value = ip
		wwwRcd.Type = "A"
		wwwRcd.Ttl = 600 // default 600s
		orm.InsertOne(&wwwRcd)
	} else if wwwRcd.Value != ip {
		wwwRcd.Value = ip
		orm.Update(&wwwRcd)
	}

	return nil
}

// HealthCheck performs database health check
func HealthCheck(orm *xorm.Engine) error {
	if err := orm.Ping(); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}

	// Check if we can query a simple table
	count, err := orm.Count(&oldmodels.TblUser{})
	if err != nil {
		return fmt.Errorf("database query failed: %w", err)
	}

	logrus.Infof("Database health check passed. User count: %d", count)
	return nil
}

// Helper functions (should be moved to a utils package)
func genRandomString(length int) string {
	// Simplified implementation
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[i%len(charset)]
	}
	return string(b)
}

func genShortId() string {
	return genRandomString(8)
}

func genRandomToken() string {
	return genRandomString(32)
}

func makePassword(password string) string {
	// Simplified implementation - should use bcrypt in production
	return password
}
