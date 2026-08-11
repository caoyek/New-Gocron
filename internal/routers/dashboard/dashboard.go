package dashboard

import (
	"sort"
	"time"

	"github.com/caoyek/New-Gocron/internal/models"
	"github.com/caoyek/New-Gocron/internal/modules/logger"
	"github.com/caoyek/New-Gocron/internal/modules/utils"
	"github.com/jakecoffman/cron"
	"github.com/ouqiang/goutil"
	"gopkg.in/macaron.v1"
)

func Index(ctx *macaron.Context) string {
	now := time.Now()
	stats, err := models.LoadDashboardStats(now)
	jsonResp := utils.JsonResponse{}
	if err != nil {
		logger.Error(err)
		return jsonResp.CommonFailure(utils.FailureContent, err)
	}
	tasks, err := models.LoadDashboardScheduledTasks()
	if err != nil {
		logger.Error(err)
		return jsonResp.CommonFailure(utils.FailureContent, err)
	}

	upcomingTasks := make([]models.DashboardUpcomingTask, 0, len(tasks))
	for _, task := range tasks {
		var nextRunTime time.Time
		err := goutil.PanicToError(func() {
			nextRunTime = cron.Parse(task.Spec).Next(now)
		})
		if err != nil || nextRunTime.IsZero() || !nextRunTime.After(now) {
			continue
		}
		upcomingTasks = append(upcomingTasks, models.DashboardUpcomingTask{
			Id:          task.Id,
			Name:        task.Name,
			Tag:         task.Tag,
			Protocol:    task.Protocol,
			Status:      task.Status,
			NextRunTime: nextRunTime,
		})
	}
	sort.Slice(upcomingTasks, func(i, j int) bool {
		if upcomingTasks[i].NextRunTime.Equal(upcomingTasks[j].NextRunTime) {
			return upcomingTasks[i].Id < upcomingTasks[j].Id
		}
		return upcomingTasks[i].NextRunTime.Before(upcomingTasks[j].NextRunTime)
	})
	if len(upcomingTasks) > 10 {
		upcomingTasks = upcomingTasks[:10]
	}
	stats.UpcomingTasks = upcomingTasks
	stats.RecentFailures, err = models.LoadDashboardRecentFailures(10)
	if err != nil {
		logger.Error(err)
		return jsonResp.CommonFailure(utils.FailureContent, err)
	}

	return jsonResp.Success(utils.SuccessContent, stats)
}
