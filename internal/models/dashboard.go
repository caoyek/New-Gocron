package models

import (
	"strings"
	"time"
)

type DashboardStats struct {
	EnabledTasks    int64                    `json:"enabled_tasks"`
	DisabledTasks   int64                    `json:"disabled_tasks"`
	TodayExecutions int                      `json:"today_executions"`
	TodaySuccesses  int                      `json:"today_successes"`
	TodayFailures   int                      `json:"today_failures"`
	HourlyCounts    []int                    `json:"hourly_counts"`
	UpcomingTasks   []DashboardUpcomingTask  `json:"upcoming_tasks"`
	RecentFailures  []DashboardRecentFailure `json:"recent_failures"`
	Date            string                   `json:"date"`
}

type DashboardUpcomingTask struct {
	Id          int          `json:"id"`
	Name        string       `json:"name"`
	Tag         string       `json:"tag"`
	Protocol    TaskProtocol `json:"protocol"`
	Status      Status       `json:"status"`
	NextRunTime time.Time    `json:"next_run_time"`
}

type DashboardRecentFailure struct {
	Id            int64        `json:"id"`
	TaskId        int          `json:"task_id"`
	Name          string       `json:"name"`
	Tag           string       `json:"tag"`
	Protocol      TaskProtocol `json:"protocol"`
	FailureTime   time.Time    `json:"failure_time"`
	ResultSummary string       `json:"result_summary"`
}

func LoadDashboardStats(now time.Time) (DashboardStats, error) {
	stats := DashboardStats{
		HourlyCounts: make([]int, 24),
		Date:         now.Format("2006-01-02"),
	}

	var err error
	stats.EnabledTasks, err = Db.Where("status = ?", Enabled).Count(new(Task))
	if err != nil {
		return stats, err
	}
	stats.DisabledTasks, err = Db.Where("status = ?", Disabled).Count(new(Task))
	if err != nil {
		return stats, err
	}

	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	end := start.AddDate(0, 0, 1)
	logs := make([]TaskLog, 0)
	err = Db.Where("start_time >= ? AND start_time < ?", start.Format(DefaultTimeFormat), end.Format(DefaultTimeFormat)).
		Cols("start_time", "status").
		Find(&logs)
	if err != nil {
		return stats, err
	}

	stats.TodayExecutions = len(logs)
	for _, log := range logs {
		if log.Status == Finish {
			stats.TodaySuccesses++
		} else if log.Status == Failure {
			stats.TodayFailures++
		}
		hour := log.StartTime.In(now.Location()).Hour()
		if hour >= 0 && hour < len(stats.HourlyCounts) {
			stats.HourlyCounts[hour]++
		}
	}

	return stats, nil
}

func LoadDashboardScheduledTasks() ([]Task, error) {
	tasks := make([]Task, 0)
	err := Db.Where("level = ? AND status = ?", TaskLevelParent, Enabled).
		Cols("id", "name", "tag", "spec", "protocol", "status").
		Find(&tasks)

	return tasks, err
}

func LoadDashboardRecentFailures(limit int) ([]DashboardRecentFailure, error) {
	logs := make([]TaskLog, 0)
	err := Db.Where("status = ?", Failure).
		Desc("id").
		Limit(limit).
		Cols("id", "task_id", "name", "protocol", "start_time", "end_time", "result").
		Find(&logs)
	if err != nil {
		return nil, err
	}

	taskIds := make([]int, 0, len(logs))
	for _, log := range logs {
		taskIds = append(taskIds, log.TaskId)
	}
	tags := make(map[int]string)
	if len(taskIds) > 0 {
		tasks := make([]Task, 0)
		err = Db.In("id", taskIds).Cols("id", "tag").Find(&tasks)
		if err != nil {
			return nil, err
		}
		for _, task := range tasks {
			tags[task.Id] = task.Tag
		}
	}

	failures := make([]DashboardRecentFailure, 0, len(logs))
	for _, log := range logs {
		failureTime := log.EndTime
		if failureTime.IsZero() {
			failureTime = log.StartTime
		}
		summary := strings.Join(strings.Fields(log.Result), " ")
		runes := []rune(summary)
		if len(runes) > 120 {
			summary = string(runes[:120]) + "..."
		}
		failures = append(failures, DashboardRecentFailure{
			Id:            log.Id,
			TaskId:        log.TaskId,
			Name:          log.Name,
			Tag:           tags[log.TaskId],
			Protocol:      log.Protocol,
			FailureTime:   failureTime,
			ResultSummary: summary,
		})
	}

	return failures, nil
}
