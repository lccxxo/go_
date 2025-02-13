package task

import (
	"github.com/go_/lccxxo/goResourceWatcher/internal/database"
	"time"

	"github.com/go_/lccxxo/goResourceWatcher/internal/logger"
)

type TaskEntity struct {
	ID        string    `gorm:"primary_key;unique;type:varchar(255)"` // 任务ID，主键且唯一
	TaskName  string    `gorm:"column:task_name;type:varchar(255)"`   // 任务名称
	CreatedAt time.Time `gorm:"column:created_at"`                    // 创建时间
	UpdatedAt time.Time `gorm:"column:updated_at"`                    // 更新时间
	Status    string    `gorm:"column:status;type:varchar(50)"`       // 运行为1 暂停为2
}

func AutoMigrateTaskEntity() {
	db := database.GetDB()
	if err := db.AutoMigrate(&TaskEntity{}); err != nil {
		logger.Logger.Errorf("Failed to auto migrate task entity: %v", err)
		return
	}
}

func Create(entity *TaskEntity) error {
	db := database.GetDB()
	entity.CreatedAt = time.Now()
	entity.Status = "1"
	return db.Create(&entity).Error
}

func Pause(entity *TaskEntity) error {
	db := database.GetDB()
	entity.UpdatedAt = time.Now()
	entity.Status = "2"
	return db.Updates(&entity).Where("id = ?", entity.ID).Error
}

func Resume(entity *TaskEntity) error {
	db := database.GetDB()
	entity.UpdatedAt = time.Now()
	entity.Status = "1"
	return db.Updates(&entity).Where("id = ?", entity.ID).Error
}

func Delete(entity *TaskEntity) error {
	db := database.GetDB()
	return db.Delete(&entity).Where("id = ?", entity.ID).Error
}

func ClearTaskTable() error {
	db := database.GetDB()
	return db.Where("1 = 1").Delete(&TaskEntity{}).Error
}

func ToTaskEntity(task Task) *TaskEntity {
	return &TaskEntity{
		ID:       task.ID,
		TaskName: task.TaskName,
	}
}
