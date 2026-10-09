package service

import (
	"time"

	"github.com/caoyek/New-Gocron/internal/models"
	"github.com/caoyek/New-Gocron/internal/modules/logger"
)

// PauseExpired prevents new executions; already-running executions may finish.
func (task Task) PauseExpired(item models.Task) {
	affected, err := models.Db.Table(new(models.Task)).ID(item.Id).
		Where("expires_at = ? AND expires_at <> '' AND expires_at <= ?", item.ExpiresAt, time.Now().Format(models.DefaultTimeFormat)).
		Update(models.CommonMap{"status": models.Disabled})
	if err != nil {
		logger.Error("暂停到期任务失败: ", err)
	}
	if affected > 0 && serviceCron != nil {
		task.Remove(item.Id)
	}
}

func (task Task) pauseExpiredTasks() {
	var tasks []models.Task
	if err := models.Db.Where("status = ? AND expires_at <> '' AND expires_at <= ?", models.Enabled, time.Now().Format(models.DefaultTimeFormat)).Find(&tasks); err != nil {
		logger.Error("读取到期任务失败: ", err)
		return
	}
	for _, item := range tasks {
		task.PauseExpired(item)
	}
}

func (task Task) maintain(stop <-chan struct{}) {
	expiration := time.NewTicker(time.Minute)
	defer expiration.Stop()
	// A large cleanup must not delay expiration checks.
	go task.maintainLogs(stop)
	for {
		select {
		case <-stop:
			return
		case <-expiration.C:
			task.pauseExpiredTasks()
		}
	}
}

func (task Task) maintainLogs(stop <-chan struct{}) {
	cleanup := time.NewTicker(time.Hour)
	defer cleanup.Stop()
	task.cleanupLogs(stop)
	for {
		select {
		case <-stop:
			return
		case <-cleanup.C:
			task.cleanupLogs(stop)
		}
	}
}

func (task Task) cleanupLogs(stop <-chan struct{}) {
	var tasks []models.Task
	if err := models.Db.Where("log_retention_days > 0").Find(&tasks); err != nil {
		logger.Error("读取日志保留配置失败: ", err)
		return
	}
	for _, item := range tasks {
		cutoff := time.Now().AddDate(0, 0, -item.LogRetentionDays)
		for {
			select {
			case <-stop:
				return
			default:
			}
			var logs []models.TaskLog
			if err := models.Db.Cols("id").Where("task_id = ? AND status <> ? AND end_time < ?", item.Id, models.Running, cutoff).Limit(500).Find(&logs); err != nil {
				logger.Error("读取过期日志失败: ", err)
				break
			}
			if len(logs) == 0 {
				break
			}
			ids := make([]int64, len(logs))
			for i, entry := range logs {
				ids[i] = entry.Id
			}
			affected, err := deleteExpiredLogBatch(item, cutoff, ids)
			if err != nil {
				logger.Error("清理过期日志失败: ", err)
				break
			}
			if affected == 0 {
				break
			}
		}
	}
}

func deleteExpiredLogBatch(item models.Task, cutoff time.Time, ids []int64) (int64, error) {
	session := models.Db.NewSession()
	defer session.Close()
	if err := session.Begin(); err != nil {
		return 0, err
	}
	defer session.Rollback()
	// Lock the setting until deletion commits, so a saved retention change
	// cannot race between checking the current setting and deleting the batch.
	var current models.Task
	found, err := session.ID(item.Id).ForUpdate().Get(&current)
	if err != nil {
		return 0, err
	}
	if !found || current.LogRetentionDays <= 0 || current.LogRetentionDays != item.LogRetentionDays {
		return 0, nil
	}
	affected, err := session.In("id", ids).
		Where("task_id = ? AND status <> ? AND end_time < ?", item.Id, models.Running, cutoff).
		Delete(new(models.TaskLog))
	if err != nil {
		return 0, err
	}
	return affected, session.Commit()
}
