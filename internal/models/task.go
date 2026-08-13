package models

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/go-xorm/xorm"
)

type TaskProtocol int8

const (
	TaskHTTP TaskProtocol = iota + 1 // HTTP协议
	TaskRPC                          // RPC方式执行命令
)

type TaskLevel int8

const (
	TaskLevelParent TaskLevel = 1 // 父任务
	TaskLevelChild  TaskLevel = 2 // 子任务(依赖任务)
)

type TaskDependencyStatus int8

const (
	TaskDependencyStatusStrong TaskDependencyStatus = 1 // 强依赖
	TaskDependencyStatusWeak   TaskDependencyStatus = 2 // 弱依赖
)

type TaskHTTPMethod int8

const (
	TaskHTTPMethodGet  TaskHTTPMethod = 1
	TaskHttpMethodPost TaskHTTPMethod = 2
)

type TaskReference struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

// 任务
type Task struct {
	Id               int                  `json:"id" xorm:"int pk autoincr"`
	Name             string               `json:"name" xorm:"varchar(32) notnull"`                            // 任务名称
	Level            TaskLevel            `json:"level" xorm:"tinyint notnull index default 1"`               // 任务等级 1: 主任务 2: 依赖任务
	DependencyTaskId string               `json:"dependency_task_id" xorm:"varchar(64) notnull default ''"`   // 依赖任务ID,多个ID逗号分隔
	DependencyStatus TaskDependencyStatus `json:"dependency_status" xorm:"tinyint notnull default 1"`         // 依赖关系 1:强依赖 主任务执行成功, 依赖任务才会被执行 2:弱依赖
	Spec             string               `json:"spec" xorm:"varchar(64) notnull"`                            // crontab
	Protocol         TaskProtocol         `json:"protocol" xorm:"tinyint notnull index"`                      // 协议 1:http 2:系统命令
	Command          string               `json:"command" xorm:"text notnull"`                                // URL地址或shell命令
	HttpMethod       TaskHTTPMethod       `json:"http_method" xorm:"tinyint notnull default 1"`               // http请求方法
	Timeout          int                  `json:"timeout" xorm:"mediumint notnull default 0"`                 // 任务执行超时时间(单位秒),0不限制
	Multi            int8                 `json:"multi" xorm:"tinyint notnull default 1"`                     // 是否允许多实例运行
	RetryTimes       int8                 `json:"retry_times" xorm:"tinyint notnull default 0"`               // 重试次数
	RetryInterval    int16                `json:"retry_interval" xorm:"smallint notnull default 0"`           // 重试间隔时间
	NotifyStatus     int8                 `json:"notify_status" xorm:"tinyint notnull default 1"`             // 任务执行结束是否通知 0: 不通知 1: 失败通知 2: 执行结束通知 3: 任务执行结果关键字匹配通知
	NotifyType       int8                 `json:"notify_type" xorm:"tinyint notnull default 0"`               // 通知类型 1: 邮件 2: slack 3: webhook
	NotifyReceiverId string               `json:"notify_receiver_id" xorm:"varchar(256) notnull default '' "` // 通知接受者ID, setting表主键ID，多个ID逗号分隔
	NotifyKeyword    string               `json:"notify_keyword" xorm:"text notnull"`
	Tag              string               `json:"tag" xorm:"varchar(32) notnull default ''"`
	Remark           string               `json:"remark" xorm:"varchar(100) notnull default ''"` // 备注
	Status           Status               `json:"status" xorm:"tinyint notnull index default 0"` // 状态 1:正常 0:停止
	Created          time.Time            `json:"created" xorm:"datetime notnull created"`       // 创建时间
	Deleted          time.Time            `json:"deleted" xorm:"datetime deleted"`               // 删除时间
	BaseModel        `json:"-" xorm:"-"`
	Hosts            []TaskHostDetail `json:"hosts" xorm:"-"`
	NextRunTime      time.Time        `json:"next_run_time" xorm:"-"`
	ParentTasks      []TaskReference  `json:"parent_tasks" xorm:"-"`
	LastRunStatus    *Status          `json:"last_run_status" xorm:"-"`
	LastRunTime      *time.Time       `json:"last_run_time" xorm:"-"`
	LastRunDuration  int              `json:"last_run_duration" xorm:"-"`
}

func taskHostTableName() []string {
	return []string{TablePrefix + "task_host", "th"}
}

// 新增
func (task *Task) Create() (insertId int, err error) {
	_, err = Db.Insert(task)
	if err == nil {
		insertId = task.Id
	}

	return
}

func (task *Task) UpdateBean(id int) (int64, error) {
	return Db.ID(id).
		Cols(`name,level,spec,protocol,command,timeout,multi,
			retry_times,retry_interval,remark,notify_status,
			notify_type,notify_receiver_id, dependency_task_id, dependency_status, tag,http_method, notify_keyword`).
		Update(task)
}

// 更新
func (task *Task) Update(id int, data CommonMap) (int64, error) {
	return Db.Table(task).ID(id).Update(data)
}

// 删除
func (task *Task) Delete(id int) (int64, error) {
	return Db.Id(id).Delete(task)
}

// 禁用
func (task *Task) Disable(id int) (int64, error) {
	return task.Update(id, CommonMap{"status": Disabled})
}

// 激活
func (task *Task) Enable(id int) (int64, error) {
	return task.Update(id, CommonMap{"status": Enabled})
}

// 获取所有激活任务
func (task *Task) ActiveList(page, pageSize int) ([]Task, error) {
	params := CommonMap{"Page": page, "PageSize": pageSize}
	task.parsePageAndPageSize(params)
	list := make([]Task, 0)
	err := Db.Where("status = ? AND level = ?", Enabled, TaskLevelParent).Limit(task.PageSize, task.pageLimitOffset()).
		Find(&list)

	if err != nil {
		return list, err
	}

	return task.setHostsForTasks(list)
}

// 获取某个主机下的所有激活任务
func (task *Task) ActiveListByHostId(hostId int16) ([]Task, error) {
	taskHostModel := new(TaskHost)
	taskIds, err := taskHostModel.GetTaskIdsByHostId(hostId)
	if err != nil {
		return nil, err
	}
	if len(taskIds) == 0 {
		return nil, nil
	}
	list := make([]Task, 0)
	err = Db.Where("status = ?  AND level = ?", Enabled, TaskLevelParent).
		In("id", taskIds...).
		Find(&list)
	if err != nil {
		return list, err
	}

	return task.setHostsForTasks(list)
}

func (task *Task) setHostsForTasks(tasks []Task) ([]Task, error) {
	taskHostModel := new(TaskHost)
	var err error
	for i, value := range tasks {
		taskHostDetails, err := taskHostModel.GetHostIdsByTaskId(value.Id)
		if err != nil {
			return nil, err
		}
		tasks[i].Hosts = taskHostDetails
	}

	return tasks, err
}

func (task *Task) setParentTasksForTasks(tasks []Task) ([]Task, error) {
	childIds := make(map[int]struct{})
	for _, item := range tasks {
		if item.Level == TaskLevelChild {
			childIds[item.Id] = struct{}{}
		}
	}
	if len(childIds) == 0 {
		return tasks, nil
	}

	parents := make([]Task, 0)
	err := Db.Where("level = ? AND dependency_task_id <> ?", TaskLevelParent, "").
		Asc("id").
		Cols("id", "name", "dependency_task_id").
		Find(&parents)
	if err != nil {
		return nil, err
	}

	parentTasks := make(map[int][]TaskReference)
	for _, parent := range parents {
		for _, value := range strings.Split(parent.DependencyTaskId, ",") {
			childId, err := strconv.Atoi(strings.TrimSpace(value))
			if err != nil {
				continue
			}
			if _, ok := childIds[childId]; ok {
				parentTasks[childId] = append(parentTasks[childId], TaskReference{
					Id:   parent.Id,
					Name: parent.Name,
				})
			}
		}
	}

	for i := range tasks {
		if tasks[i].Level == TaskLevelChild {
			tasks[i].ParentTasks = parentTasks[tasks[i].Id]
		}
	}

	return tasks, nil
}

func (task *Task) setLastRunsForTasks(tasks []Task) ([]Task, error) {
	if len(tasks) == 0 {
		return tasks, nil
	}

	taskIds := make([]interface{}, len(tasks))
	for i, item := range tasks {
		taskIds[i] = item.Id
	}

	type latestTaskLog struct {
		Id int64 `xorm:"id"`
	}
	latestLogs := make([]latestTaskLog, 0)
	err := Db.Table(new(TaskLog)).
		Select("MAX(id) AS id").
		In("task_id", taskIds...).
		GroupBy("task_id").
		Find(&latestLogs)
	if err != nil {
		return nil, err
	}
	if len(latestLogs) == 0 {
		return tasks, nil
	}

	logIds := make([]interface{}, len(latestLogs))
	for i, log := range latestLogs {
		logIds[i] = log.Id
	}
	logs := make([]TaskLog, 0, len(logIds))
	err = Db.Table(new(TaskLog)).
		In("id", logIds...).
		Cols("task_id", "status", "start_time", "end_time").
		Find(&logs)
	if err != nil {
		return nil, err
	}

	lastRuns := make(map[int]TaskLog, len(logs))
	for _, log := range logs {
		lastRuns[log.TaskId] = log
	}

	for i := range tasks {
		log, ok := lastRuns[tasks[i].Id]
		if !ok {
			continue
		}
		status := log.Status
		startTime := log.StartTime
		tasks[i].LastRunStatus = &status
		tasks[i].LastRunTime = &startTime
		tasks[i].LastRunDuration = log.executionSeconds(time.Now())
	}

	return tasks, nil
}

// 判断任务名称是否存在
func (task *Task) NameExist(name string, id int) (bool, error) {
	if id > 0 {
		count, err := Db.Where("name = ? AND status = ? AND id != ?", name, Enabled, id).Count(task)
		return count > 0, err
	}
	count, err := Db.Where("name = ? AND status = ?", name, Enabled).Count(task)

	return count > 0, err
}

func (task *Task) GetStatus(id int) (Status, error) {
	exist, err := Db.Id(id).Get(task)
	if err != nil {
		return 0, err
	}
	if !exist {
		return 0, errors.New("not exist")
	}

	return task.Status, nil
}

func (task *Task) Detail(id int) (Task, error) {
	t := Task{}
	_, err := Db.Where("id=?", id).Get(&t)

	if err != nil {
		return t, err
	}

	taskHostModel := new(TaskHost)
	t.Hosts, err = taskHostModel.GetHostIdsByTaskId(id)

	return t, err
}

func (task *Task) List(params CommonMap) ([]Task, error) {
	task.parsePageAndPageSize(params)
	list := make([]Task, 0)
	session := Db.Alias("t").Join("LEFT", taskHostTableName(), "t.id = th.task_id")
	task.parseWhere(session, params)
	if pinnedIds, ok := params["PinnedIds"].([]int); ok && len(pinnedIds) > 0 {
		session.OrderBy(pinnedTaskOrder(pinnedIds))
	}
	err := session.GroupBy("t.id").Desc("t.id").Cols("t.*").Limit(task.PageSize, task.pageLimitOffset()).Find(&list)

	if err != nil {
		return nil, err
	}

	list, err = task.setHostsForTasks(list)
	if err != nil {
		return nil, err
	}

	list, err = task.setParentTasksForTasks(list)
	if err != nil {
		return nil, err
	}

	return task.setLastRunsForTasks(list)
}

func pinnedTaskOrder(ids []int) string {
	parts := make([]string, 0, len(ids))
	for index, id := range ids {
		parts = append(parts, fmt.Sprintf("WHEN %d THEN %d", id, index))
	}
	return fmt.Sprintf("CASE t.id %s ELSE %d END ASC", strings.Join(parts, " "), len(ids))
}

// Children returns the fields needed by dependency task selectors.
func (task *Task) Children() ([]Task, error) {
	list := make([]Task, 0)
	err := Db.Where("level = ?", TaskLevelChild).
		Asc("id").
		Cols("id", "name").
		Find(&list)

	return list, err
}

// Tags returns all non-empty task tags for list filtering.
func (task *Task) Tags() ([]string, error) {
	rows := make([]struct {
		Tag string `xorm:"tag"`
	}, 0)
	err := Db.Table(task).
		Where("tag <> ?", "").
		Distinct("tag").
		Asc("tag").
		Find(&rows)
	if err != nil {
		return nil, err
	}

	tags := make([]string, len(rows))
	for i, row := range rows {
		tags[i] = row.Tag
	}

	return tags, nil
}

// 获取依赖任务列表
func (task *Task) GetDependencyTaskList(ids string) ([]Task, error) {
	list := make([]Task, 0)
	if ids == "" {
		return list, nil
	}
	idList := strings.Split(ids, ",")
	taskIds := make([]interface{}, len(idList))
	for i, v := range idList {
		taskIds[i] = v
	}
	fields := "t.*"
	err := Db.Alias("t").
		Where("t.level = ?", TaskLevelChild).
		In("t.id", taskIds).
		Cols(fields).
		Find(&list)

	if err != nil {
		return list, err
	}

	return task.setHostsForTasks(list)
}

func (task *Task) Total(params CommonMap) (int64, error) {
	session := Db.Alias("t").Join("LEFT", taskHostTableName(), "t.id = th.task_id")
	task.parseWhere(session, params)
	list := make([]Task, 0)

	err := session.GroupBy("t.id").Find(&list)

	return int64(len(list)), err
}

// 解析where
func (task *Task) parseWhere(session *xorm.Session, params CommonMap) {
	if len(params) == 0 {
		return
	}
	id, ok := params["Id"]
	if ok && id.(int) > 0 {
		session.And("t.id = ?", id)
	}
	hostId, ok := params["HostId"]
	if ok && hostId.(int) > 0 {
		session.And("th.host_id = ?", hostId)
	}
	name, ok := params["Name"]
	if ok && name.(string) != "" {
		session.And("t.name LIKE ?", "%"+name.(string)+"%")
	}
	keyword, ok := params["Keyword"]
	if ok && keyword.(string) != "" {
		likeKeyword := "%" + keyword.(string) + "%"
		session.And("(CAST(t.id AS CHAR) LIKE ? OR t.name LIKE ?)", likeKeyword, likeKeyword)
	}
	protocol, ok := params["Protocol"]
	if ok && protocol.(int) > 0 {
		session.And("protocol = ?", protocol)
	}
	status, ok := params["Status"]
	if ok && status.(int) > -1 {
		session.And("status = ?", status)
	}

	tag, ok := params["Tag"]
	if ok && tag.(string) != "" {
		session.And("tag = ? ", tag)
	}
}
