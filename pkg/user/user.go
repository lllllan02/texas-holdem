package user

import (
	"log"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// User 代表系统中的一个真实用户（不论是否在参与游戏）
type User struct {
	ID       string `json:"id" gorm:"primaryKey"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
}

var db *gorm.DB

func init() {
	var err error
	// 初始化 SQLite 数据库，文件名为 texas.db
	db, err = gorm.Open(sqlite.Open("texas.db"), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	// 自动迁移模式
	if err := db.AutoMigrate(&User{}); err != nil {
		log.Fatalf("failed to migrate database: %v", err)
	}
}

// GetUserByID 获取用户信息。如果用户不存在，则自动创建一个默认用户。
func GetUserByID(id string) *User {
	var u User
	result := db.First(&u, "id = ?", id)
	
	if result.Error == nil {
		return &u
	}

	// 如果不存在，创建一个默认的
	nickname := "Player_" + id
	if len(id) > 6 {
		nickname = "Player_" + id[:6]
	}

	u = User{
		ID:       id,
		Nickname: nickname,
		Avatar:   "", // 默认头像
	}

	if err := db.Create(&u).Error; err != nil {
		log.Printf("failed to create default user: %v", err)
	}

	return &u
}

// UpdateUser 更新用户信息
func UpdateUser(id, nickname, avatar string) *User {
	var u User
	result := db.First(&u, "id = ?", id)
	
	if result.Error != nil {
		// 如果用户不存在，则创建
		u = User{ID: id}
		if nickname != "" {
			u.Nickname = nickname
		}
		if avatar != "" {
			u.Avatar = avatar
		}
		db.Create(&u)
		return &u
	}

	// 更新字段
	updates := make(map[string]interface{})
	if nickname != "" {
		updates["nickname"] = nickname
	}
	if avatar != "" {
		updates["avatar"] = avatar
	}

	if len(updates) > 0 {
		db.Model(&u).Updates(updates)
	}

	// 重新查询最新数据
	db.First(&u, "id = ?", id)
	return &u
}
