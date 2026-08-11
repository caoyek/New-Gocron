<template>
  <el-container class="dashboard-page">
    <el-main v-loading="loading">
      <header class="dashboard-header">
        <div>
          <h1>数据看板</h1>
        </div>
        <time class="dashboard-clock" :datetime="currentTime.toISOString()">
          <small>当前时间</small>
          <strong>{{currentTime | formatTime}}</strong>
        </time>
      </header>

      <section class="metric-grid" aria-label="任务统计">
        <article class="metric-card metric-card--enabled">
          <span class="metric-card__icon"><i class="el-icon-circle-check"></i></span>
          <div>
            <span class="metric-card__label">启用任务数</span>
            <strong>{{enabledTasks}}</strong>
          </div>
        </article>
        <article class="metric-card metric-card--disabled">
          <span class="metric-card__icon"><i class="el-icon-circle-close"></i></span>
          <div>
            <span class="metric-card__label">停用任务数</span>
            <strong>{{disabledTasks}}</strong>
          </div>
        </article>
        <article class="metric-card metric-card--success">
          <span class="metric-card__icon"><i class="el-icon-success"></i></span>
          <div>
            <span class="metric-card__label">今日执行成功数</span>
            <strong>{{todaySuccesses}}</strong>
          </div>
        </article>
        <article class="metric-card metric-card--failure">
          <span class="metric-card__icon"><i class="el-icon-error"></i></span>
          <div>
            <span class="metric-card__label">今日执行失败数</span>
            <strong>{{todayFailures}}</strong>
          </div>
        </article>
      </section>

      <div class="dashboard-activity-grid">
      <section class="activity-panel upcoming-panel" aria-label="即将执行的任务">
        <header class="upcoming-panel__header">
          <div>
            <h2>即将执行</h2>
            <span>未来最近执行的 10 个已启用主任务</span>
          </div>
        </header>

        <div v-if="upcomingTasks.length" class="upcoming-list">
          <div
            v-for="(task, index) in upcomingTasks"
            :key="task.id"
            class="upcoming-list__row">
            <span class="upcoming-list__order">{{index + 1}}</span>
            <time class="upcoming-list__time" :datetime="task.next_run_time">
              {{task.next_run_time | formatTime}}
            </time>
            <div class="upcoming-list__details">
              <span v-if="task.tag" class="upcoming-list__tag" :title="task.tag">{{task.tag}}</span>
              <span v-else class="upcoming-list__tag is-empty">未设置标签</span>
              <strong class="upcoming-list__name" :title="task.name">{{task.name}}</strong>
              <span
                class="upcoming-list__protocol"
                :class="task.protocol === 1 ? 'is-http' : 'is-shell'">
                {{task.protocol === 1 ? 'HTTP' : 'SHELL'}}
              </span>
              <span
                class="upcoming-list__status"
                :class="task.status === 1 ? 'is-enabled' : 'is-disabled'">
                {{task.status === 1 ? '启用' : '停用'}}
              </span>
            </div>
          </div>
        </div>
        <div v-else class="upcoming-list__empty">暂无即将执行的任务</div>
      </section>

      <section class="activity-panel recent-failure-panel" aria-label="最近执行异常">
        <header class="upcoming-panel__header">
          <div>
            <h2>最近执行异常</h2>
            <span>最近 10 条失败记录</span>
          </div>
          <strong class="recent-failure-panel__count">{{recentFailures.length}} 条</strong>
        </header>

        <div v-if="recentFailures.length" class="recent-failure-list">
          <button
            v-for="(failure, index) in recentFailures"
            :key="failure.id"
            type="button"
            class="upcoming-list__row recent-failure-list__row"
            :title="failureTitle(failure)"
            @click="openFailureLog(failure)">
            <span class="upcoming-list__order">{{index + 1}}</span>
            <time class="upcoming-list__time" :datetime="failure.failure_time">
              {{failure.failure_time | formatTime}}
            </time>
            <div class="upcoming-list__details">
              <span v-if="failure.tag" class="upcoming-list__tag" :title="failure.tag">{{failure.tag}}</span>
              <span v-else class="upcoming-list__tag is-empty">未设置标签</span>
              <strong class="upcoming-list__name">{{failure.name}}</strong>
              <span
                class="upcoming-list__protocol"
                :class="failure.protocol === 1 ? 'is-http' : 'is-shell'">
                {{failure.protocol === 1 ? 'HTTP' : 'SHELL'}}
              </span>
              <span class="upcoming-list__status is-failure">失败</span>
            </div>
          </button>
        </div>
        <div v-else class="upcoming-list__empty">暂无执行异常</div>
      </section>
      </div>

      <section class="chart-panel">
        <header class="chart-panel__header">
          <div>
            <h2>今日执行统计</h2>
            <span>按任务开始执行时间统计</span>
          </div>
          <strong>共 {{todayExecutions}} 次</strong>
        </header>

        <div class="chart-wrap">
          <svg
            class="execution-chart"
            :viewBox="`0 0 ${chart.width} ${chart.height}`"
            preserveAspectRatio="xMidYMid meet"
            role="img"
            :aria-label="`${date} 每小时任务执行次数折线图`"
            @mouseleave="hoveredHour = null">
            <g v-for="tick in yTicks" :key="`y-${tick.value}`">
              <line
                class="chart-grid-line"
                :x1="chart.left"
                :x2="chart.width - chart.right"
                :y1="tick.y"
                :y2="tick.y">
              </line>
              <text class="chart-axis-label" :x="chart.left - 10" :y="tick.y + 4" text-anchor="end">
                {{tick.value}}
              </text>
            </g>

            <line
              class="chart-axis-line"
              :x1="chart.left"
              :x2="chart.width - chart.right"
              :y1="chart.height - chart.bottom"
              :y2="chart.height - chart.bottom">
            </line>

            <text
              v-for="label in xLabels"
              :key="`x-${label.hour}`"
              class="chart-axis-label"
              :x="label.x"
              :y="chart.height - 14"
              text-anchor="middle">
              {{label.hour}}时
            </text>

            <polyline class="chart-line" :points="linePoints"></polyline>

            <g v-for="point in chartPoints" :key="`point-${point.hour}`">
              <circle
                class="chart-point-hit"
                :cx="point.x"
                :cy="point.y"
                r="10"
                @mouseenter="hoveredHour = point.hour">
              </circle>
              <circle
                class="chart-point"
                :class="{'is-active': hoveredHour === point.hour}"
                :cx="point.x"
                :cy="point.y"
                r="3.5">
              </circle>
            </g>

            <g v-if="hoveredPoint" class="chart-tooltip">
              <rect
                :x="tooltipPosition.x"
                :y="tooltipPosition.y"
                width="88"
                height="30"
                rx="4">
              </rect>
              <text
                :x="tooltipPosition.x + 44"
                :y="tooltipPosition.y + 19"
                text-anchor="middle">
                {{hoveredPoint.hour}}时 · {{hoveredPoint.count}}次
              </text>
            </g>
          </svg>
        </div>
      </section>
    </el-main>
  </el-container>
</template>

<script>
import dashboardService from '../../api/dashboard'

export default {
  name: 'dashboard-index',
  data () {
    return {
      loading: false,
      date: '',
      enabledTasks: 0,
      disabledTasks: 0,
      todayExecutions: 0,
      todaySuccesses: 0,
      todayFailures: 0,
      hourlyCounts: new Array(24).fill(0),
      upcomingTasks: [],
      recentFailures: [],
      currentTime: new Date(),
      clockTimer: null,
      refreshTimer: null,
      hoveredHour: null,
      chart: {
        width: 960,
        height: 280,
        left: 54,
        right: 18,
        top: 20,
        bottom: 42
      }
    }
  },
  computed: {
    chartMaximum () {
      const maximum = Math.max.apply(null, this.hourlyCounts)
      if (maximum <= 4) {
        return 4
      }
      return Math.ceil(maximum / 4) * 4
    },
    chartPoints () {
      const plotWidth = this.chart.width - this.chart.left - this.chart.right
      const plotHeight = this.chart.height - this.chart.top - this.chart.bottom
      return this.hourlyCounts.map((count, index) => ({
        hour: index + 1,
        count,
        x: this.chart.left + ((index + 1) / 24) * plotWidth,
        y: this.chart.top + plotHeight - (count / this.chartMaximum) * plotHeight
      }))
    },
    linePoints () {
      return this.chartPoints.map((point) => `${point.x},${point.y}`).join(' ')
    },
    yTicks () {
      const plotHeight = this.chart.height - this.chart.top - this.chart.bottom
      const ticks = []
      for (let index = 0; index <= 4; index++) {
        ticks.push({
          value: this.chartMaximum - (this.chartMaximum / 4) * index,
          y: this.chart.top + (plotHeight / 4) * index
        })
      }
      return ticks
    },
    xLabels () {
      const plotWidth = this.chart.width - this.chart.left - this.chart.right
      const labels = []
      for (let hour = 1; hour <= 24; hour++) {
        labels.push({
          hour,
          x: this.chart.left + (hour / 24) * plotWidth
        })
      }
      return labels
    },
    hoveredPoint () {
      if (this.hoveredHour === null) {
        return null
      }
      return this.chartPoints[this.hoveredHour - 1]
    },
    tooltipPosition () {
      if (!this.hoveredPoint) {
        return {x: 0, y: 0}
      }
      const x = Math.max(
        this.chart.left,
        Math.min(this.hoveredPoint.x - 44, this.chart.width - this.chart.right - 88)
      )
      const y = Math.max(this.chart.top, this.hoveredPoint.y - 40)
      return {x, y}
    }
  },
  created () {
    this.load()
    this.clockTimer = window.setInterval(() => {
      this.currentTime = new Date()
    }, 1000)
    this.refreshTimer = window.setInterval(() => {
      this.load(true)
    }, 30000)
  },
  beforeDestroy () {
    window.clearInterval(this.clockTimer)
    window.clearInterval(this.refreshTimer)
  },
  methods: {
    load (silent) {
      if (!silent) {
        this.loading = true
      }
      dashboardService.index((data) => {
        this.loading = false
        this.date = data.date || ''
        this.enabledTasks = data.enabled_tasks || 0
        this.disabledTasks = data.disabled_tasks || 0
        this.todayExecutions = data.today_executions || 0
        this.todaySuccesses = data.today_successes || 0
        this.todayFailures = data.today_failures || 0
        this.hourlyCounts = Array.isArray(data.hourly_counts) && data.hourly_counts.length === 24
          ? data.hourly_counts
          : new Array(24).fill(0)
        this.upcomingTasks = Array.isArray(data.upcoming_tasks) ? data.upcoming_tasks : []
        this.recentFailures = Array.isArray(data.recent_failures) ? data.recent_failures : []
      })
    },
    failureTitle (failure) {
      return failure.result_summary
        ? `${failure.name}\n${failure.result_summary}`
        : failure.name
    },
    openFailureLog (failure) {
      this.$router.push(`/task/log?task_id=${failure.task_id}`)
    }
  }
}
</script>

<style scoped>
.dashboard-page {
  width: 100%;
  height: 100%;
  color: #2c2c2b;
}

.dashboard-page > .el-main {
  overflow-y: auto;
}

.dashboard-header {
  display: flex;
  min-height: 38px;
  align-items: flex-start;
  justify-content: space-between;
  margin-bottom: 14px;
}

.dashboard-header h1 {
  margin: 0;
  font-size: 18px;
  font-weight: 600;
  letter-spacing: 0;
  line-height: 24px;
}

.dashboard-clock {
  display: inline-flex;
  min-height: 32px;
  padding: 5px 10px;
  align-items: baseline;
  gap: 8px;
  box-sizing: border-box;
  border-left: 3px solid #2783de;
  border-radius: 4px;
  background: #f4f8fc;
  line-height: 20px;
  white-space: nowrap;
}

.dashboard-clock small {
  color: #5f6266;
  font-size: 11px;
}

.dashboard-clock strong {
  color: #2377c9;
  font-size: 14px;
  font-weight: 600;
  font-variant-numeric: tabular-nums;
}

.metric-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 12px;
  margin-bottom: 14px;
}

.metric-card {
  display: flex;
  min-height: 94px;
  padding: 18px 20px;
  align-items: center;
  gap: 14px;
  box-sizing: border-box;
  border: 1px solid #e5e6e8;
  border-radius: 7px;
  background: #ffffff;
}

.metric-card__icon {
  display: inline-flex;
  width: 38px;
  height: 38px;
  align-items: center;
  justify-content: center;
  flex: 0 0 38px;
  border-radius: 6px;
  font-size: 20px;
}

.metric-card--enabled .metric-card__icon {
  background: #eaf7ef;
  color: #2b9b5f;
}

.metric-card--disabled .metric-card__icon {
  background: #f5eeee;
  color: #b85d5d;
}

.metric-card--success .metric-card__icon {
  background: #eaf7ef;
  color: #2b9b5f;
}

.metric-card--failure .metric-card__icon {
  background: #faeeee;
  color: #c84e4e;
}

.metric-card__label {
  display: block;
  margin-bottom: 3px;
  color: #74777b;
  font-size: 12px;
  line-height: 18px;
}

.metric-card strong {
  display: block;
  font-size: 26px;
  font-weight: 600;
  letter-spacing: 0;
  line-height: 30px;
}

.dashboard-activity-grid {
  display: grid;
  margin-bottom: 14px;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 12px;
}

.activity-panel {
  min-width: 0;
  border: 1px solid #e5e6e8;
  border-radius: 7px;
  background: #ffffff;
  overflow: hidden;
}

.upcoming-panel__header {
  display: flex;
  min-height: 56px;
  padding: 10px 16px;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  box-sizing: border-box;
  border-bottom: 1px solid #eceef0;
}

.upcoming-panel__header h2 {
  margin: 0;
  font-size: 14px;
  font-weight: 600;
  letter-spacing: 0;
  line-height: 21px;
}

.upcoming-panel__header span {
  color: #8a8d91;
  font-size: 11px;
  line-height: 17px;
}

.recent-failure-panel__count {
  color: #b85d5d;
  font-size: 12px;
  font-weight: 500;
  line-height: 20px;
  white-space: nowrap;
}

.upcoming-list__row {
  display: grid;
  min-height: 39px;
  padding: 0 16px;
  align-items: center;
  grid-template-columns: 24px 148px minmax(0, 1fr);
  column-gap: 12px;
  border-bottom: 1px solid #f0f1f2;
  box-sizing: border-box;
  font-size: 12px;
}

.upcoming-list__row:last-child {
  border-bottom: 0;
}

.upcoming-list__order {
  color: #a0a3a7;
  font-variant-numeric: tabular-nums;
  text-align: center;
}

.upcoming-list__time {
  color: #4e5358;
  font-variant-numeric: tabular-nums;
  white-space: nowrap;
}

.upcoming-list__details {
  display: grid;
  min-width: 0;
  align-items: center;
  grid-template-columns: 145px minmax(0, 1fr) 44px 46px;
  gap: 14px;
}

.upcoming-list__name,
.upcoming-list__tag {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.upcoming-list__name {
  flex: 0 1 auto;
  color: #303236;
  font-weight: 500;
}

.upcoming-list__tag {
  justify-self: start;
  max-width: 100%;
  padding: 2px 7px;
  border: 1px solid #d9e5f1;
  border-radius: 3px;
  background: #f4f8fc;
  color: #477498;
  line-height: 18px;
}

.upcoming-list__tag.is-empty {
  border-color: transparent;
  background: transparent;
  color: #a0a3a7;
}

.upcoming-list__protocol {
  display: inline-flex;
  width: 44px;
  height: 22px;
  align-items: center;
  justify-content: center;
  box-sizing: border-box;
  border: 1px solid transparent;
  border-radius: 3px;
  font-size: 10px;
  line-height: 20px;
  white-space: nowrap;
}

.upcoming-list__protocol.is-http {
  border-color: #91c4f2;
  background: #e7f2ff;
  color: #1769aa;
}

.upcoming-list__protocol.is-shell {
  border-color: #efc37c;
  background: #fff3df;
  color: #9a4f00;
}

.upcoming-list__status {
  display: inline-flex;
  width: 38px;
  height: 22px;
  align-items: center;
  justify-content: center;
  box-sizing: border-box;
  border: 1px solid transparent;
  border-radius: 3px;
  font-size: 11px;
  line-height: 20px;
  white-space: nowrap;
  justify-self: end;
}

.upcoming-list__status.is-enabled {
  border-color: #c9e8d5;
  background: #eef8f2;
  color: #278655;
}

.upcoming-list__status.is-disabled {
  border-color: #ead2d2;
  background: #faf1f1;
  color: #b85d5d;
}

.upcoming-list__status.is-failure {
  border-color: #e5bcbc;
  background: #faeded;
  color: #b54040;
}

.recent-failure-list__row {
  width: 100%;
  border-top: 0;
  border-right: 0;
  border-left: 0;
  background: transparent;
  color: inherit;
  cursor: pointer;
  font-family: inherit;
  text-align: left;
}

.recent-failure-list__row:hover,
.recent-failure-list__row:focus-visible {
  background: #faf6f6;
  outline: 0;
}

.upcoming-list__empty {
  padding: 28px 16px;
  color: #a0a3a7;
  font-size: 12px;
  text-align: center;
}

.chart-panel {
  min-height: 0;
  padding: 18px 20px 14px;
  box-sizing: border-box;
  border: 1px solid #e5e6e8;
  border-radius: 7px;
  background: #ffffff;
}

.chart-panel__header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
  margin-bottom: 8px;
}

.chart-panel__header h2 {
  margin: 0;
  font-size: 14px;
  font-weight: 600;
  letter-spacing: 0;
  line-height: 21px;
}

.chart-panel__header span {
  color: #8a8d91;
  font-size: 11px;
  line-height: 17px;
}

.chart-panel__header > strong {
  color: #5f6266;
  font-size: 12px;
  font-weight: 500;
  line-height: 21px;
  white-space: nowrap;
}

.chart-wrap {
  width: 100%;
  min-height: 280px;
  overflow-x: auto;
}

.execution-chart {
  display: block;
  width: 100%;
  min-width: 700px;
  height: auto;
  aspect-ratio: 24 / 7;
}

.chart-grid-line {
  stroke: #eceef0;
  stroke-width: 1;
}

.chart-axis-line {
  stroke: #cfd2d5;
  stroke-width: 1;
}

.chart-axis-label {
  fill: #8a8d91;
  font-size: 11px;
}

.chart-line {
  fill: none;
  stroke: #2783de;
  stroke-linecap: round;
  stroke-linejoin: round;
  stroke-width: 2;
}

.chart-point-hit {
  fill: transparent;
  cursor: crosshair;
}

.chart-point {
  fill: #ffffff;
  stroke: #2783de;
  stroke-width: 2;
  pointer-events: none;
}

.chart-point.is-active {
  fill: #2783de;
  stroke-width: 3;
}

.chart-tooltip rect {
  fill: #2c2c2b;
}

.chart-tooltip text {
  fill: #ffffff;
  font-size: 11px;
}

@media (max-width: 760px) {
  .metric-grid {
    grid-template-columns: 1fr;
  }

  .metric-card {
    min-height: 78px;
    padding: 13px 16px;
  }

  .chart-panel {
    padding-right: 14px;
    padding-left: 14px;
  }

  .upcoming-panel__header {
    align-items: flex-start;
    flex-direction: column;
    gap: 4px;
  }

  .upcoming-list__row {
    padding: 8px 12px;
    grid-template-columns: 20px minmax(0, 1fr);
    gap: 4px 8px;
  }

  .upcoming-list__time {
    grid-column: 2;
    grid-row: 2;
    font-size: 11px;
  }

  .upcoming-list__details {
    grid-column: 2;
  }

  .upcoming-list__tag {
    max-width: 110px;
  }
}

@media (max-width: 1100px) {
  .dashboard-activity-grid {
    grid-template-columns: 1fr;
  }
}

@media (min-width: 761px) and (max-width: 1100px) {
  .metric-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}
</style>
