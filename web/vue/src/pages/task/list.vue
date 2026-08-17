<template>
<el-container>
  <el-main class="task-list-main">
    <el-form class="task-filter-form task-desktop-only">
      <div class="task-filter-grid">
        <el-form-item>
          <el-input
            v-model.trim="searchParams.keyword"
            clearable
            placeholder="搜索任务..."
            @keyup.enter.native="applySearch()">
          </el-input>
        </el-form-item>
        <el-form-item>
          <el-select
            v-model="searchParams.tag"
            filterable
            clearable
            placeholder="标签"
            @change="applySearch()">
            <el-option
              v-for="tag in tags"
              :key="tag"
              :label="tag"
              :value="tag">
            </el-option>
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-select v-model.trim="searchParams.host_id" clearable placeholder="任务节点" @change="applySearch()">
            <el-option
              v-for="item in hosts"
              :key="item.id"
              :label="item.alias + ' - ' + item.name + ':' + item.port "
              :value="item.id">
            </el-option>
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-select v-model.trim="searchParams.protocol" clearable placeholder="执行方式" @change="applySearch()">
            <el-option
              v-for="item in protocolList"
              :key="item.value"
              :label="item.label"
              :value="item.value">
            </el-option>
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-select v-model.trim="searchParams.status" clearable placeholder="状态" @change="applySearch()">
            <el-option
              v-for="item in statusList"
              :key="item.value"
              :label="item.label"
              :value="item.value">
            </el-option>
          </el-select>
        </el-form-item>
        <div class="task-filter-actions">
          <el-button size="small" type="primary" icon="el-icon-search" @click="applySearch()">搜索</el-button>
          <el-button
            v-if="this.$store.getters.user.isAdmin"
            size="small"
            type="success"
            icon="el-icon-plus"
            @click="toEdit(null)">新增</el-button>
        </div>
      </div>
    </el-form>
    <el-pagination
      class="task-pagination task-desktop-only"
      background
      layout="prev, pager, next, sizes, total"
      :total="taskTotal"
      :current-page="searchParams.page"
      :page-size="20"
      @size-change="changePageSize"
      @current-change="changePage"
      @prev-click="changePage"
      @next-click="changePage">
    </el-pagination>
    <el-table
      ref="taskTable"
      :data="tasks"
      :class="['task-desktop-only', {'has-expanded-task': expandedTaskIds.length > 0}]"
      tooltip-effect="dark"
      border
      @expand-change="handleExpandChange"
      @row-click="handleRowClick"
      row-class-name="task-table-row"
      style="width: 100%">
      <el-table-column
        type="expand"
        width="1"
        class-name="task-expand-column"
        label-class-name="task-expand-column">
        <template slot-scope="scope">
          <div class="task-detail-panel" :style="taskDetailPanelStyle">
            <div class="task-detail-settings">
              <section class="task-detail-section">
                <h3 class="task-detail-section-title"><span>1</span>基本信息</h3>
                <div class="task-detail-grid task-detail-grid--three">
                  <div class="task-detail-field">
                    <span class="task-detail-label">创建时间</span>
                    <span class="task-detail-value">{{scope.row.created | formatTime}}</span>
                  </div>
                  <div class="task-detail-field">
                    <span class="task-detail-label">任务类型</span>
                    <span class="task-detail-value">{{scope.row.level | formatLevel}}</span>
                  </div>
                  <div class="task-detail-field">
                    <span class="task-detail-label">单实例运行</span>
                    <span class="task-detail-value">{{scope.row.multi | formatMulti}}</span>
                  </div>
                </div>
              </section>

              <section class="task-detail-section">
                <h3 class="task-detail-section-title"><span>2</span>执行设置</h3>
                <div class="task-detail-grid task-detail-grid--three">
                  <div class="task-detail-field">
                    <span class="task-detail-label">超时时间</span>
                    <span class="task-detail-value">{{scope.row.timeout | formatTimeout}}</span>
                  </div>
                  <div class="task-detail-field">
                    <span class="task-detail-label">重试次数</span>
                    <span class="task-detail-value">{{scope.row.retry_times}} 次</span>
                  </div>
                  <div class="task-detail-field">
                    <span class="task-detail-label">重试间隔</span>
                    <span class="task-detail-value">{{scope.row.retry_interval | formatRetryTimesInterval}}</span>
                  </div>
                  <div class="task-detail-field task-detail-field--full">
                    <span class="task-detail-label">任务节点</span>
                    <div v-if="scope.row.hosts && scope.row.hosts.length" class="task-detail-hosts">
                      <el-tag v-for="item in scope.row.hosts" :key="item.host_id" size="mini" type="info">
                        {{item.alias}} · {{item.name}}:{{item.port}}
                      </el-tag>
                    </div>
                    <span v-else class="task-detail-empty">未配置任务节点</span>
                  </div>
                </div>
              </section>

              <section class="task-detail-section">
                <h3 class="task-detail-section-title"><span>3</span>备注</h3>
                <div class="task-detail-remark">{{scope.row.remark || '暂无备注'}}</div>
              </section>
            </div>

            <section class="task-detail-command-panel">
              <div class="task-detail-command-heading">
                <span>命令（执行内容）</span>
                <span class="task-detail-command-stats">
                  {{scope.row.protocol === 2 ? 'SHELL' : 'HTTP'}} · {{(scope.row.command || '').split('\n').length}} 行 · {{(scope.row.command || '').length}} 字符
                </span>
              </div>
              <pre class="task-detail-command">{{scope.row.command || '未配置命令'}}</pre>
            </section>
          </div>
        </template>
      </el-table-column>
      <el-table-column
        prop="id"
        label="ID"
        width="52">
      </el-table-column>
      <el-table-column
        label="任务名称"
        width="330">
        <template slot-scope="scope">
          <div
            class="task-name-cell"
            :title="scope.row.level === 2
              ? `${scope.row.name}  主任务：${formatParentTasks(scope.row.parent_tasks)}`
              : scope.row.name">
            <span class="task-name-line">
              <button
                type="button"
                class="task-pin-button"
                :class="{'is-pinned': isTaskPinned(scope.row.id)}"
                :title="isTaskPinned(scope.row.id) ? '取消置顶' : '置顶任务'"
                :aria-label="isTaskPinned(scope.row.id) ? '取消置顶' : '置顶任务'"
                @click.stop="toggleTaskPin(scope.row)">
                <i :class="isTaskPinned(scope.row.id) ? 'el-icon-star-on' : 'el-icon-star-off'"></i>
              </button>
              <span class="task-name-text">{{scope.row.name}}</span>
            </span>
            <span v-if="scope.row.level === 2" class="task-parent-name">
              主任务：{{formatParentTasks(scope.row.parent_tasks)}}
            </span>
          </div>
        </template>
      </el-table-column>
      <el-table-column label="标签" :width="tagColumnWidth">
        <template slot-scope="scope">
          <span class="task-tag-cell" :title="scope.row.tag || ''">{{scope.row.tag || '-'}}</span>
        </template>
      </el-table-column>
      <el-table-column
        label="cron表达式"
        width="140">
        <template slot-scope="scope">
          <el-tag v-if="scope.row.level === 2" size="mini" type="info">子任务</el-tag>
          <span v-else>{{scope.row.spec}}</span>
        </template>
      </el-table-column>
      <el-table-column label="下次执行时间" width="190">
        <template slot-scope="scope">
          <el-tag v-if="scope.row.level === 2" size="mini" type="info">子任务</el-tag>
          <span v-else>{{scope.row.next_run_time | formatTime}}</span>
        </template>
      </el-table-column>
      <el-table-column label="上次执行" width="250">
        <template slot-scope="scope">
          <div v-if="scope.row.last_run_time" class="task-last-run">
            <span
              class="task-last-run__summary"
              :class="'is-' + lastRunTagType(scope.row.last_run_status)">
              <span>{{lastRunStatusText(scope.row.last_run_status)}}</span>
              <span class="task-last-run__divider"></span>
              <span>{{formatRunDuration(scope.row.last_run_duration)}}</span>
            </span>
            <span class="task-last-run__time">{{scope.row.last_run_time | formatTime}}</span>
          </div>
          <span v-else class="task-last-run-empty">暂无执行记录</span>
        </template>
      </el-table-column>
      <el-table-column
        prop="protocol"
        :formatter="formatProtocol"
        label="执行方式"
        width="88">
      </el-table-column>
      <el-table-column label="执行命令" min-width="180">
        <template slot-scope="scope">
          <el-popover
            placement="top-start"
            width="420"
            trigger="hover"
            :open-delay="250">
            <div class="task-command-preview">
              <div class="task-command-preview__heading">
                <span>执行命令</span>
                <small>{{(scope.row.command || '').split('\n').length}} 行 · {{(scope.row.command || '').length}} 字符</small>
              </div>
              <pre>{{scope.row.command || '未配置命令'}}</pre>
            </div>
            <span
              slot="reference"
              class="task-command-summary"
              :class="{'is-empty': !scope.row.command}">
              {{scope.row.command || '未配置命令'}}
            </span>
          </el-popover>
        </template>
      </el-table-column>
      <el-table-column
        label="状态"
        width="57"
        class-name="task-status-column"
        label-class-name="task-status-column"
        fixed="right"
        v-if="this.isAdmin">
          <template slot-scope="scope">
            <el-switch
              v-if="scope.row.level === 1"
              v-model="scope.row.status"
              :active-value="1"
              :inactive-value="0"
              active-color="#13ce66"
              @change="changeStatus(scope.row)"
              inactive-color="#ff4949">
            </el-switch>
            <el-tag v-else size="mini" type="info">子任务</el-tag>
          </template>
      </el-table-column>
      <el-table-column
        label="状态"
        width="57"
        class-name="task-status-column"
        label-class-name="task-status-column"
        fixed="right"
        v-else>
        <template slot-scope="scope">
          <el-switch
            v-if="scope.row.level === 1"
            v-model="scope.row.status"
            :active-value="1"
            :inactive-value="0"
            active-color="#13ce66"
            :disabled="true"
            inactive-color="#ff4949">
          </el-switch>
          <el-tag v-else size="mini" type="info">子任务</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="156" fixed="right" v-if="this.isAdmin">
        <template slot-scope="scope">
          <div class="task-actions">
            <el-tooltip content="编辑" placement="top">
              <el-button
                class="task-action-button"
                size="mini"
                type="primary"
                icon="el-icon-edit"
                aria-label="编辑"
                @click="toEdit(scope.row)">
              </el-button>
            </el-tooltip>
            <el-tooltip content="执行" placement="top">
              <el-button
                class="task-action-button"
                size="mini"
                type="success"
                icon="el-icon-caret-right"
                aria-label="执行"
                @click="runTask(scope.row)">
              </el-button>
            </el-tooltip>
            <el-tooltip content="日志" placement="top">
              <el-button
                class="task-action-button"
                size="mini"
                type="info"
                icon="el-icon-document"
                aria-label="日志"
                @click="jumpToLog(scope.row)">
              </el-button>
            </el-tooltip>
            <el-tooltip content="删除" placement="top">
              <el-button
                class="task-action-button"
                size="mini"
                type="danger"
                icon="el-icon-delete"
                aria-label="删除"
                @click="remove(scope.row)">
              </el-button>
            </el-tooltip>
          </div>
        </template>
      </el-table-column>
    </el-table>

    <section class="task-mobile-view" aria-label="定时任务列表">
      <div class="task-mobile-search">
        <el-input
          v-model.trim="searchParams.keyword"
          clearable
          placeholder="搜索任务..."
          @keyup.enter.native="applyMobileSearch">
        </el-input>
        <el-button
          class="task-mobile-icon-button"
          type="primary"
          icon="el-icon-search"
          aria-label="搜索"
          @click="applyMobileSearch">
        </el-button>
        <button
          class="task-mobile-icon-button task-mobile-filter-toggle"
          type="button"
          aria-label="筛选"
          @click="mobileFiltersVisible = !mobileFiltersVisible">
          <i class="el-icon-setting"></i>
          <span v-if="activeMobileFilterCount">{{activeMobileFilterCount}}</span>
        </button>
      </div>

      <div v-if="mobileFiltersVisible" class="task-mobile-filters">
        <el-select v-model="searchParams.tag" filterable clearable placeholder="标签">
          <el-option v-for="tag in tags" :key="tag" :label="tag" :value="tag"></el-option>
        </el-select>
        <el-select v-model.trim="searchParams.host_id" clearable placeholder="任务节点">
          <el-option
            v-for="item in hosts"
            :key="item.id"
            :label="item.alias + ' - ' + item.name + ':' + item.port"
            :value="item.id">
          </el-option>
        </el-select>
        <div class="task-mobile-filter-row">
          <el-select v-model.trim="searchParams.protocol" clearable placeholder="执行方式">
            <el-option
              v-for="item in protocolList"
              :key="item.value"
              :label="item.label"
              :value="item.value">
            </el-option>
          </el-select>
          <el-select v-model.trim="searchParams.status" clearable placeholder="状态">
            <el-option
              v-for="item in statusList"
              :key="item.value"
              :label="item.label"
              :value="item.value">
            </el-option>
          </el-select>
        </div>
        <div class="task-mobile-filter-actions">
          <el-button size="small" @click="resetMobileFilters">重置</el-button>
          <el-button size="small" type="primary" @click="applyMobileSearch">应用</el-button>
        </div>
      </div>

      <div class="task-mobile-summary">
        <span>共 {{taskTotal}} 个任务</span>
        <el-button
          v-if="isAdmin"
          size="small"
          type="success"
          icon="el-icon-plus"
          @click="toEdit(null)">新增</el-button>
      </div>

      <div v-if="tasks.length" class="task-mobile-list">
        <article v-for="item in tasks" :key="item.id" class="task-mobile-card">
          <header class="task-mobile-card__header" @click="toggleMobileTask(item)">
            <button
              type="button"
              class="task-pin-button task-mobile-pin"
              :class="{'is-pinned': isTaskPinned(item.id)}"
              :aria-label="isTaskPinned(item.id) ? '取消置顶' : '置顶任务'"
              @click.stop="toggleTaskPin(item)">
              <i :class="isTaskPinned(item.id) ? 'el-icon-star-on' : 'el-icon-star-off'"></i>
            </button>
            <span class="task-mobile-id">#{{item.id}}</span>
            <div class="task-mobile-card__title">
              <strong>{{item.name}}</strong>
              <span v-if="item.level === 2">主任务：{{formatParentTasks(item.parent_tasks)}}</span>
            </div>
            <div class="task-mobile-card__meta">
              <el-tag v-if="item.tag" size="mini" type="info">{{item.tag}}</el-tag>
              <span
                class="task-mobile-protocol"
                :class="item.protocol === 2 ? 'is-shell' : 'is-http'">{{mobileProtocolText(item)}}</span>
            </div>
            <span class="task-mobile-status" @click.stop>
              <el-switch
                v-if="item.level === 1"
                v-model="item.status"
                :active-value="1"
                :inactive-value="0"
                active-color="#13ce66"
                inactive-color="#ff4949"
                :disabled="!isAdmin"
                @change="changeStatus(item)">
              </el-switch>
              <el-tag v-else size="mini" type="info">子任务</el-tag>
            </span>
          </header>

          <div class="task-mobile-card__schedule" @click="toggleMobileTask(item)">
            <div>
              <span>表达式</span>
              <strong>{{item.level === 2 ? '子任务' : item.spec}}</strong>
            </div>
            <div>
              <span>下次执行</span>
              <strong v-if="item.level === 2">随主任务</strong>
              <strong v-else>{{item.next_run_time | formatTime}}</strong>
            </div>
            <div class="task-mobile-card__last-run">
              <span>上次执行</span>
              <strong v-if="item.last_run_time" :class="'is-' + lastRunTagType(item.last_run_status)">
                {{lastRunStatusText(item.last_run_status)}} · {{formatRunDuration(item.last_run_duration)}}
                <small>{{item.last_run_time | formatTime}}</small>
              </strong>
              <strong v-else class="is-empty">暂无执行记录</strong>
            </div>
          </div>

          <div
            v-if="mobileExpandedTaskId === item.id"
            class="task-mobile-card__details">
            <dl>
              <div>
                <dt>任务节点</dt>
                <dd>{{mobileHostText(item.hosts)}}</dd>
              </div>
              <div>
                <dt>超时 / 重试</dt>
                <dd>{{item.timeout | formatTimeout}} · {{item.retry_times}} 次</dd>
              </div>
              <div>
                <dt>备注</dt>
                <dd>{{item.remark || '暂无备注'}}</dd>
              </div>
            </dl>
            <div class="task-mobile-command">
              <span>执行命令</span>
              <pre>{{item.command || '未配置命令'}}</pre>
            </div>
          </div>

          <footer v-if="isAdmin" class="task-mobile-card__actions">
            <el-button size="mini" type="primary" icon="el-icon-edit" @click.stop="toEdit(item)">编辑</el-button>
            <el-button size="mini" type="success" icon="el-icon-caret-right" @click.stop="runTask(item)">执行</el-button>
            <el-button size="mini" type="info" icon="el-icon-document" @click.stop="jumpToLog(item)">日志</el-button>
            <el-button size="mini" type="danger" icon="el-icon-delete" @click.stop="remove(item)">删除</el-button>
          </footer>
        </article>
      </div>
      <div v-else class="task-mobile-empty">暂无任务</div>

      <nav v-if="taskTotal > 0" class="task-mobile-pagination" aria-label="任务分页">
        <button
          type="button"
          aria-label="上一页"
          :disabled="searchParams.page <= 1"
          @click="changePage(searchParams.page - 1)">
          <i class="el-icon-arrow-left"></i>
        </button>
        <span>第 {{searchParams.page}} / {{mobileTotalPages}} 页</span>
        <button
          type="button"
          aria-label="下一页"
          :disabled="searchParams.page >= mobileTotalPages"
          @click="changePage(searchParams.page + 1)">
          <i class="el-icon-arrow-right"></i>
        </button>
      </nav>
    </section>

    <task-editor
      :visible.sync="editorVisible"
      :task-id="editingTaskId"
      :hosts="hosts"
      :tags="tags"
      @saved="handleTaskSaved">
    </task-editor>
  </el-main>
</el-container>
</template>

<script>
import taskEditor from './edit'
import taskService from '../../api/task'

export default {
  name: 'task-list',
  data () {
    return {
      tasks: [],
      hosts: [],
      tags: [],
      taskTotal: 0,
      editorVisible: false,
      editingTaskId: null,
      autoRefreshTimer: null,
      delayedRefreshTimer: null,
      searchInProgress: false,
      autoRefreshPaused: false,
      searchRequestId: 0,
      lastSearchParams: null,
      expandedTaskIds: [],
      pinnedTaskIds: [],
      taskDetailPanelWidth: 0,
      mobileFiltersVisible: false,
      mobileExpandedTaskId: null,
      searchParams: {
        page_size: 20,
        page: 1,
        keyword: '',
        protocol: '',
        tag: '',
        host_id: '',
        status: ''
      },
      isAdmin: this.$store.getters.user.isAdmin,
      protocolList: [
        {
          value: '1',
          label: 'http'
        },
        {
          value: '2',
          label: 'shell'
        }
      ],
      statusList: [
        {
          value: '2',
          label: '激活'
        },
        {
          value: '1',
          label: '停止'
        }
      ]
    }
  },
  computed: {
    taskDetailPanelStyle () {
      if (!this.taskDetailPanelWidth) {
        return null
      }
      return {width: this.taskDetailPanelWidth + 'px'}
    },
    tagColumnWidth () {
      if (!this.tags.length) {
        return 130
      }
      const canvas = document.createElement('canvas')
      const context = canvas.getContext('2d')
      context.font = '14px "Helvetica Neue", Helvetica, "PingFang SC", Arial, sans-serif'
      const textWidth = this.tags.reduce((width, tag) => {
        return Math.max(width, context.measureText(tag || '-').width)
      }, 0)
      return Math.max(130, Math.ceil(textWidth) + 32)
    },
    mobileTotalPages () {
      return Math.max(1, Math.ceil(this.taskTotal / this.searchParams.page_size))
    },
    activeMobileFilterCount () {
      return ['tag', 'host_id', 'protocol', 'status'].filter(key => this.searchParams[key] !== '').length
    }
  },
  components: {taskEditor},
  created () {
    this.loadPinnedTasks()
    const hostId = this.$route.query.host_id
    if (hostId) {
      this.searchParams.host_id = hostId
    }

    this.search()
    this.loadTags()
    this.startAutoRefresh()
    document.addEventListener('visibilitychange', this.handleVisibilityChange)
  },
  mounted () {
    window.addEventListener('resize', this.updateTaskDetailPanelWidth)
  },
  beforeDestroy () {
    if (this.autoRefreshTimer) {
      clearInterval(this.autoRefreshTimer)
    }
    if (this.delayedRefreshTimer) {
      clearTimeout(this.delayedRefreshTimer)
    }
    document.removeEventListener('visibilitychange', this.handleVisibilityChange)
    window.removeEventListener('resize', this.updateTaskDetailPanelWidth)
  },
  filters: {
    formatLevel (value) {
      if (value === 1) {
        return '主任务'
      }
      return '子任务'
    },
    formatTimeout (value) {
      if (value > 0) {
        return value + '秒'
      }
      return '不限制'
    },
    formatRetryTimesInterval (value) {
      if (value > 0) {
        return value + '秒'
      }
      return '系统默认'
    },
    formatMulti (value) {
      if (value > 0) {
        return '否'
      }
      return '是'
    }
  },
  methods: {
    pinnedTaskStorageKey () {
      const uid = this.$store.getters.user.uid || 'anonymous'
      return `new-gocron:pinned-tasks:${uid}`
    },
    loadPinnedTasks () {
      try {
        const ids = JSON.parse(localStorage.getItem(this.pinnedTaskStorageKey()) || '[]')
        this.pinnedTaskIds = Array.isArray(ids)
          ? ids.map(id => Number(id)).filter((id, index, list) => id > 0 && list.indexOf(id) === index).slice(0, 100)
          : []
      } catch (error) {
        this.pinnedTaskIds = []
      }
    },
    isTaskPinned (id) {
      return this.pinnedTaskIds.indexOf(Number(id)) !== -1
    },
    toggleTaskPin (task) {
      const id = Number(task.id)
      if (this.isTaskPinned(id)) {
        this.pinnedTaskIds = this.pinnedTaskIds.filter(item => item !== id)
        this.$message.success('已取消置顶')
      } else {
        this.pinnedTaskIds = [id].concat(this.pinnedTaskIds).slice(0, 100)
        this.$message.success('任务已置顶')
      }
      localStorage.setItem(this.pinnedTaskStorageKey(), JSON.stringify(this.pinnedTaskIds))
      this.searchParams.page = 1
      this.search()
    },
    formatRunDuration (seconds) {
      const totalSeconds = Math.max(0, Number(seconds) || 0)
      if (totalSeconds < 60) {
        return totalSeconds + 's'
      }
      const minutes = Math.floor(totalSeconds / 60)
      const remainingSeconds = totalSeconds % 60
      if (minutes < 60) {
        return minutes + 'm ' + remainingSeconds + 's'
      }
      const hours = Math.floor(minutes / 60)
      const remainingMinutes = minutes % 60
      return hours + 'h ' + remainingMinutes + 'm'
    },
    startAutoRefresh () {
      this.autoRefreshTimer = setInterval(() => {
        this.autoRefresh()
      }, 30000)
    },
    autoRefresh () {
      if (document.hidden || this.editorVisible || this.expandedTaskIds.length > 0 || this.searchInProgress || this.autoRefreshPaused) {
        return
      }
      this.search(null, this.lastSearchParams)
    },
    handleVisibilityChange () {
      if (!document.hidden) {
        this.autoRefresh()
      }
    },
    scheduleRefresh (delay = 1500) {
      if (this.delayedRefreshTimer) {
        clearTimeout(this.delayedRefreshTimer)
      }
      this.delayedRefreshTimer = setTimeout(() => {
        this.delayedRefreshTimer = null
        this.autoRefresh()
      }, delay)
    },
    handleExpandChange (row, expandedRows) {
      this.expandedTaskIds = expandedRows.map(item => item.id)
      if (expandedRows.length > 0) {
        this.$nextTick(this.updateTaskDetailPanelWidth)
      }
    },
    updateTaskDetailPanelWidth () {
      const table = this.$refs.taskTable && this.$refs.taskTable.$el
      if (!table) {
        return
      }
      const expandedCell = table.querySelector('.el-table__body-wrapper .el-table__expanded-cell')
      let horizontalPadding = 0
      if (expandedCell) {
        const style = window.getComputedStyle(expandedCell)
        horizontalPadding = parseFloat(style.paddingLeft) + parseFloat(style.paddingRight)
      }
      this.taskDetailPanelWidth = Math.max(0, table.clientWidth - horizontalPadding)
    },
    handleRowClick (row, event) {
      if (event.target.closest('button, a, input, .el-switch, .task-actions, .task-command-summary')) {
        return
      }
      const expanded = this.expandedTaskIds.indexOf(row.id) !== -1
      this.$refs.taskTable.toggleRowExpansion(row, !expanded)
    },
    lastRunStatusText (status) {
      const statusText = {
        0: '失败',
        1: '执行中',
        2: '成功',
        3: '取消'
      }
      return statusText[status] || '未知'
    },
    lastRunTagType (status) {
      const tagType = {
        0: 'danger',
        1: 'warning',
        2: 'success',
        3: 'info'
      }
      return tagType[status] || 'info'
    },
    formatParentTasks (parentTasks) {
      if (!Array.isArray(parentTasks) || parentTasks.length === 0) {
        return '未关联'
      }
      return parentTasks.map(item => `${item.id} · ${item.name}`).join('、')
    },
    mobileProtocolText (item) {
      if (item.protocol === 2) {
        return 'SHELL'
      }
      return 'HTTP'
    },
    mobileHostText (hosts) {
      if (!Array.isArray(hosts) || hosts.length === 0) {
        return '未配置任务节点'
      }
      return hosts.map(item => `${item.alias} · ${item.name}:${item.port}`).join('、')
    },
    toggleMobileTask (item) {
      this.mobileExpandedTaskId = this.mobileExpandedTaskId === item.id ? null : item.id
    },
    applyMobileSearch () {
      this.mobileFiltersVisible = false
      this.applySearch()
    },
    resetMobileFilters () {
      this.searchParams.tag = ''
      this.searchParams.host_id = ''
      this.searchParams.protocol = ''
      this.searchParams.status = ''
      this.applyMobileSearch()
    },
    applySearch () {
      this.searchParams.page = 1
      this.search()
    },
    changeStatus (item) {
      if (item.status) {
        taskService.enable(item.id, () => {
          this.search()
        })
      } else {
        taskService.disable(item.id, () => {
          this.search()
        })
      }
    },
    formatProtocol (row, col) {
      if (row[col.property] === 2) {
        return 'shell'
      }
      if (row.http_method === 1) {
        return 'http-get'
      }
      return 'http-post'
    },
    changePage (page) {
      this.searchParams.page = page
      this.search()
    },
    changePageSize (pageSize) {
      this.searchParams.page_size = pageSize
      this.search()
    },
    search (callback = null, params = null) {
      const requestId = ++this.searchRequestId
      const searchParams = Object.assign({}, params || this.searchParams)
      searchParams.pinned_ids = this.pinnedTaskIds.join(',')
      if (!params) {
        this.lastSearchParams = searchParams
      }
      this.searchInProgress = true
      taskService.list(searchParams, (tasks, hosts) => {
        if (requestId !== this.searchRequestId) {
          return
        }
        this.tasks = tasks.data
        this.taskTotal = tasks.total
        this.hosts = hosts
        if (callback) {
          callback()
        }
      }, (success) => {
        if (requestId === this.searchRequestId) {
          this.searchInProgress = false
          this.autoRefreshPaused = !success
        }
      })
    },
    loadTags () {
      taskService.tags((tags) => {
        this.tags = tags || []
      })
    },
    runTask (item) {
      this.$appConfirm(() => {
        taskService.run(item.id, () => {
          this.$message.success('任务已开始执行')
          this.search()
          this.scheduleRefresh()
        })
      }, true)
    },
    remove (item) {
      this.$appConfirm(() => {
        taskService.remove(item.id, () => {
          this.refresh()
        })
      })
    },
    jumpToLog (item) {
      this.$router.push(`/task/log?task_id=${item.id}`)
    },
    refresh () {
      this.search(() => {
        this.$message.success('刷新成功')
      })
      this.loadTags()
    },
    toEdit (item) {
      this.editingTaskId = item ? item.id : null
      this.editorVisible = true
    },
    handleTaskSaved () {
      this.search()
      this.loadTags()
    }
  }
}
</script>
<style scoped>
  .task-list-main {
    height: calc(100vh - 40px) !important;
    overflow-x: hidden;
    overflow-y: auto;
  }

  .task-pagination {
    margin-bottom: 10px;
  }

  .task-list-main /deep/ .task-expand-column {
    width: 1px;
    padding: 0 !important;
  }

  .task-list-main /deep/ .el-table__expand-column {
    padding: 0 !important;
    overflow: hidden;
  }

  .task-list-main /deep/ .el-table__expand-column .cell,
  .task-list-main /deep/ .el-table__expand-icon {
    display: none !important;
    padding: 0;
  }

  .task-list-main /deep/ .task-table-row {
    cursor: pointer;
  }

  .task-list-main /deep/ .task-status-column .cell {
    padding-right: 8px;
    padding-left: 8px;
    text-align: center;
  }

  .task-list-main /deep/ .task-status-column .el-switch__core {
    border: 0;
  }

  .task-list-main /deep/ .task-status-column .el-switch__button {
    top: 2px;
    left: 2px;
  }

  .task-list-main /deep/ .task-status-column .el-tag {
    padding-right: 2px;
    padding-left: 2px;
    font-size: 11px;
  }

  .task-list-main /deep/ .has-expanded-task .el-table__fixed-right {
    z-index: 2;
    box-shadow: none;
    pointer-events: none;
  }

  .task-list-main /deep/ .has-expanded-task .el-table__fixed-right .task-table-row {
    pointer-events: auto;
  }

  .task-list-main /deep/ .has-expanded-task .el-table__fixed-right .el-table__expanded-cell {
    border: 0;
    border-bottom: 1px solid #ebeef5;
    background: transparent !important;
    pointer-events: none;
  }

  .task-list-main /deep/ .has-expanded-task .el-table__fixed-right .el-table__expanded-cell > * {
    visibility: hidden;
  }

  .task-detail-panel {
    position: relative;
    z-index: 3;
    display: grid;
    grid-template-columns: minmax(0, 1.1fr) minmax(0, 0.9fr);
    color: #2c2c2b;
    background: #fff;
  }

  .task-detail-settings,
  .task-detail-command-panel {
    min-width: 0;
    padding: 16px 22px;
    box-sizing: border-box;
  }

  .task-detail-command-panel {
    border-left: 1px solid #dcd9d4;
  }

  .task-detail-section {
    margin-bottom: 14px;
  }

  .task-detail-section:last-child {
    margin-bottom: 0;
  }

  .task-detail-section-title {
    display: flex;
    align-items: center;
    gap: 7px;
    margin: 0 0 8px;
    color: #7d7a75;
    font-size: 12px;
    font-weight: 600;
  }

  .task-detail-section-title::after {
    height: 1px;
    flex: 1;
    background: #e6e5e3;
    content: '';
  }

  .task-detail-section-title span {
    display: inline-flex;
    width: 16px;
    height: 16px;
    align-items: center;
    justify-content: center;
    border-radius: 4px;
    background: #e5f2fc;
    color: #2783de;
    font-size: 11px;
  }

  .task-detail-grid {
    display: grid;
    gap: 10px 18px;
  }

  .task-detail-grid--three {
    grid-template-columns: repeat(3, minmax(0, 1fr));
  }

  .task-detail-field {
    min-width: 0;
  }

  .task-detail-field--full {
    grid-column: 1 / -1;
  }

  .task-detail-label {
    display: block;
    margin-bottom: 3px;
    color: #7d7a75;
    font-size: 12px;
    line-height: 18px;
  }

  .task-detail-value,
  .task-detail-empty {
    color: #2c2c2b;
    font-size: 13px;
    line-height: 20px;
  }

  .task-detail-empty {
    color: #9a9792;
  }

  .task-detail-hosts {
    display: flex;
    gap: 6px;
    align-items: center;
    flex-wrap: wrap;
  }

  .task-detail-hosts /deep/ .el-tag {
    height: 22px;
    font-size: 11px;
    line-height: 20px;
  }

  .task-detail-remark {
    min-height: 42px;
    padding: 8px 10px;
    border: 1px solid #e0dfdc;
    border-radius: 6px;
    color: #53514e;
    font-size: 13px;
    line-height: 20px;
    white-space: pre-wrap;
    overflow-wrap: anywhere;
  }

  .task-detail-command-heading {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 8px;
    color: #2c2c2b;
    font-size: 12px;
    font-weight: 600;
  }

  .task-detail-command-stats {
    color: #8f8c87;
    font-size: 11px;
    font-weight: 400;
  }

  .task-detail-command {
    min-height: 184px;
    margin: 0;
    padding: 12px 14px;
    border: 1px solid #343a46;
    border-radius: 6px;
    box-sizing: border-box;
    overflow: auto;
    background: #1f232b;
    color: #e6edf3;
    font-family: Consolas, 'Courier New', monospace;
    font-size: 12px;
    line-height: 20px;
    text-align: left;
    text-align-last: left;
    white-space: pre-wrap;
    word-break: break-word;
    overflow-wrap: break-word;
  }

  .task-filter-form {
    margin-bottom: 18px;
    padding: 0;
    border: 0;
    background: transparent;
  }

  .task-filter-grid {
    display: grid;
    grid-template-columns: 200px 200px 200px 120px 90px auto;
    gap: 8px;
  }

  .task-filter-grid .el-form-item {
    display: flex;
    align-items: center;
    margin-bottom: 0;
  }

  .task-filter-grid /deep/ .el-form-item__content {
    flex: 1;
    min-width: 0;
    margin-left: 0 !important;
    line-height: 40px;
  }

  .task-filter-grid /deep/ .el-input,
  .task-filter-grid /deep/ .el-select {
    width: 100%;
  }

  .task-filter-actions {
    display: flex;
    gap: 8px;
    align-items: center;
    align-self: end;
  }

  .task-filter-actions .el-button + .el-button {
    margin-left: 0;
  }

  .task-actions {
    display: flex;
    align-items: center;
    gap: 6px;
    white-space: nowrap;
  }

  .task-actions .el-button + .el-button {
    margin-left: 0;
  }

  .task-action-button {
    width: 28px;
    height: 28px;
    padding: 0;
    border-radius: 4px;
  }

  .task-action-button /deep/ i {
    font-size: 14px;
  }

  .task-last-run {
    display: flex;
    gap: 8px;
    align-items: center;
    white-space: nowrap;
  }

  .task-last-run__summary {
    display: inline-flex;
    flex: 0 0 auto;
    height: 22px;
    padding: 0 6px;
    border: 1px solid #d3d4d6;
    border-radius: 4px;
    align-items: center;
    gap: 6px;
    box-sizing: border-box;
    background: #f4f4f5;
    color: #73767a;
    font-size: 12px;
    font-weight: 600;
    line-height: 20px;
  }

  .task-last-run__summary.is-success {
    border-color: #c2e7b0;
    background: #f0f9eb;
    color: #4b9f2f;
  }

  .task-last-run__summary.is-danger {
    border-color: #fbc4c4;
    background: #fef0f0;
    color: #f25555;
  }

  .task-last-run__summary.is-warning {
    border-color: #f5dab1;
    background: #fdf6ec;
    color: #d48806;
  }

  .task-last-run__divider {
    width: 1px;
    height: 11px;
    background: currentColor;
    opacity: 0.35;
  }

  .task-last-run__time {
    color: #8a8d91;
    font-size: 12px;
    line-height: 17px;
  }

  .task-last-run-empty {
    color: #9a9792;
    font-size: 12px;
    white-space: nowrap;
  }

  .task-command-summary {
    display: block;
    overflow: hidden;
    color: #53514e;
    font-family: Consolas, 'Courier New', monospace;
    font-size: 12px;
    line-height: 20px;
    text-overflow: ellipsis;
    white-space: nowrap;
    cursor: help;
  }

  .task-command-summary.is-empty {
    color: #9a9792;
    font-family: inherit;
  }

  .task-command-preview {
    color: #2c2c2b;
  }

  .task-command-preview__heading {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
    padding-bottom: 8px;
    border-bottom: 1px solid #e6e5e3;
    font-size: 12px;
    font-weight: 600;
  }

  .task-command-preview__heading small {
    color: #8f8c87;
    font-size: 11px;
    font-weight: 400;
  }

  .task-command-preview pre {
    max-height: 220px;
    margin: 10px 0 0;
    padding: 12px 14px;
    border: 1px solid #343a46;
    border-radius: 6px;
    box-sizing: border-box;
    overflow: auto;
    background: #1f232b;
    color: #e6edf3;
    font-family: Consolas, 'Courier New', monospace;
    font-size: 12px;
    line-height: 20px;
    text-align: left;
    text-align-last: left;
    white-space: pre-wrap;
    word-break: break-word;
    overflow-wrap: break-word;
  }

  .task-name-cell {
    min-width: 0;
    line-height: 18px;
    white-space: nowrap;
  }

  .task-name-text,
  .task-parent-name,
  .task-tag-cell {
    white-space: nowrap;
  }

  .task-name-text {
    min-width: 0;
  }

  .task-name-line {
    display: flex;
    gap: 6px;
    align-items: center;
  }

  .task-pin-button {
    width: 18px;
    height: 18px;
    flex: 0 0 18px;
    padding: 0;
    border: 0;
    outline: none;
    background: transparent;
    color: #b8bbc0;
    cursor: pointer;
    font-size: 15px;
    line-height: 18px;
  }

  .task-pin-button:hover,
  .task-pin-button.is-pinned {
    color: #e6a23c;
  }

  .task-parent-name {
    display: block;
    margin-top: 2px;
    color: #8a8d91;
    font-size: 11px;
    line-height: 16px;
  }

  .task-tag-cell {
    display: block;
  }

  .task-mobile-view {
    display: none;
  }

  @media (max-width: 1200px) {
    .task-filter-grid {
      grid-template-columns: repeat(3, minmax(220px, 1fr));
    }

    .task-filter-actions {
      grid-column: 1 / -1;
      justify-content: space-between;
    }
  }

  @media (max-width: 960px) {
    .task-list-main {
      overflow-y: auto;
    }

    .task-detail-panel {
      grid-template-columns: 1fr;
    }

    .task-detail-command-panel {
      border-top: 1px solid #dcd9d4;
      border-left: 0;
    }

    .task-filter-grid {
      grid-template-columns: repeat(2, minmax(220px, 1fr));
    }
  }

  @media (max-width: 768px) {
    .task-list-main {
      height: calc(100vh - 80px) !important;
      margin: 12px !important;
      overflow-y: auto;
    }

    .task-desktop-only {
      display: none !important;
    }

    .task-mobile-view {
      display: block;
      color: #2c2c2b;
    }

    .task-mobile-search {
      display: grid;
      grid-template-columns: minmax(0, 1fr) 40px 40px;
      gap: 8px;
      align-items: center;
    }

    .task-mobile-search /deep/ .el-input__inner {
      height: 40px;
      line-height: 40px;
    }

    .task-mobile-icon-button {
      position: relative;
      display: inline-flex;
      width: 40px;
      height: 40px;
      padding: 0;
      align-items: center;
      justify-content: center;
      box-sizing: border-box;
      border-radius: 6px;
    }

    .task-mobile-filter-toggle {
      border: 1px solid #d8dce2;
      outline: 0;
      background: #ffffff;
      color: #4c5056;
      cursor: pointer;
      font-size: 14px;
    }

    .task-mobile-filter-toggle span {
      position: absolute;
      top: -5px;
      right: -5px;
      display: inline-flex;
      min-width: 16px;
      height: 16px;
      padding: 0 4px;
      align-items: center;
      justify-content: center;
      box-sizing: border-box;
      border-radius: 8px;
      background: #f25555;
      color: #ffffff;
      font-size: 10px;
      line-height: 16px;
    }

    .task-mobile-filters {
      display: grid;
      margin-top: 10px;
      padding: 12px;
      gap: 10px;
      border: 1px solid #e1e4e8;
      border-radius: 6px;
      background: #f7f8fa;
    }

    .task-mobile-filters /deep/ .el-select {
      width: 100%;
    }

    .task-mobile-filter-row {
      display: grid;
      grid-template-columns: repeat(2, minmax(0, 1fr));
      gap: 8px;
    }

    .task-mobile-filter-actions {
      display: flex;
      justify-content: flex-end;
      gap: 8px;
    }

    .task-mobile-filter-actions .el-button + .el-button {
      margin-left: 0;
    }

    .task-mobile-summary {
      display: flex;
      min-height: 44px;
      align-items: center;
      justify-content: space-between;
      color: #7d8085;
      font-size: 12px;
    }

    .task-mobile-list {
      display: grid;
      gap: 10px;
    }

    .task-mobile-card {
      overflow: hidden;
      border: 1px solid #e1e4e8;
      border-radius: 6px;
      background: #ffffff;
      box-shadow: 0 1px 3px rgba(31, 35, 40, 0.05);
    }

    .task-mobile-card__header {
      display: flex;
      min-height: 54px;
      padding: 11px 10px;
      align-items: start;
      gap: 4px;
      box-sizing: border-box;
      cursor: pointer;
    }

    .task-mobile-pin {
      margin-top: 2px;
    }

    .task-mobile-card__title {
      min-width: 0;
      max-width: 140px;
      flex: 0 1 auto;
    }

    .task-mobile-card__title strong {
      display: block;
      overflow: hidden;
      color: #25282d;
      font-size: 15px;
      line-height: 21px;
      text-overflow: ellipsis;
      white-space: nowrap;
    }

    .task-mobile-card__title span {
      display: block;
      margin-top: 2px;
      overflow: hidden;
      color: #8a8d91;
      font-size: 11px;
      line-height: 16px;
      text-overflow: ellipsis;
      white-space: nowrap;
    }

    .task-mobile-status {
      width: 36px;
      min-width: 36px;
      margin-left: auto;
      flex: 0 0 36px;
      text-align: right;
    }

    .task-mobile-status /deep/ .el-switch {
      transform: scale(0.9);
      transform-origin: right top;
    }

    .task-mobile-status /deep/ .el-switch__core {
      border: 0;
    }

    .task-mobile-status /deep/ .el-tag {
      width: 36px;
      padding: 0;
      box-sizing: border-box;
      font-size: 10px;
      text-align: center;
    }

    .task-mobile-status /deep/ .el-switch__button {
      top: 2px;
      left: 2px;
    }

    .task-mobile-card__meta {
      display: flex;
      min-width: 0;
      min-height: 22px;
      padding: 0;
      align-items: center;
      flex: 0 1 auto;
      gap: 4px;
      box-sizing: border-box;
      white-space: nowrap;
    }

    .task-mobile-card__meta /deep/ .el-tag {
      max-width: 66px;
      overflow: hidden;
      text-overflow: ellipsis;
    }

    .task-mobile-id {
      margin-top: 2px;
      flex: 0 0 auto;
      color: #92969b;
      font-size: 11px;
      line-height: 18px;
    }

    .task-mobile-protocol {
      display: inline-flex;
      height: 20px;
      padding: 0 6px;
      align-items: center;
      border: 1px solid;
      border-radius: 3px;
      font-size: 10px;
      font-weight: 600;
      line-height: 18px;
    }

    .task-mobile-protocol.is-http {
      border-color: #a9d2ef;
      background: #eef7fd;
      color: #2677a8;
    }

    .task-mobile-protocol.is-shell {
      border-color: #efd2a2;
      background: #fff8ea;
      color: #a56814;
    }

    .task-mobile-card__schedule {
      display: grid;
      padding: 10px 12px;
      grid-template-columns: minmax(0, 0.8fr) minmax(0, 1.2fr);
      gap: 10px 14px;
      border-top: 1px solid #eceef1;
      background: #fafbfc;
      cursor: pointer;
    }

    .task-mobile-card__schedule > div {
      min-width: 0;
    }

    .task-mobile-card__schedule span {
      display: block;
      margin-bottom: 2px;
      color: #8a8d91;
      font-size: 11px;
      line-height: 16px;
    }

    .task-mobile-card__schedule strong {
      display: block;
      overflow: hidden;
      color: #45484d;
      font-size: 12px;
      font-weight: 500;
      line-height: 18px;
      text-overflow: ellipsis;
      white-space: nowrap;
    }

    .task-mobile-card__last-run {
      grid-column: 1 / -1;
    }

    .task-mobile-card__last-run strong small {
      margin-left: 5px;
      color: #92969b;
      font-size: 11px;
      font-weight: 400;
    }

    .task-mobile-card__last-run strong.is-success {
      color: #4b9f2f;
    }

    .task-mobile-card__last-run strong.is-danger {
      color: #e34f4f;
    }

    .task-mobile-card__last-run strong.is-warning {
      color: #c98215;
    }

    .task-mobile-card__last-run strong.is-info,
    .task-mobile-card__last-run strong.is-empty {
      color: #7d8085;
    }

    .task-mobile-card__details {
      padding: 12px;
      border-top: 1px solid #e1e4e8;
    }

    .task-mobile-card__details dl {
      display: grid;
      margin: 0;
      gap: 10px;
    }

    .task-mobile-card__details dl > div {
      display: grid;
      grid-template-columns: 78px minmax(0, 1fr);
      gap: 8px;
    }

    .task-mobile-card__details dt {
      color: #8a8d91;
      font-size: 11px;
      line-height: 18px;
    }

    .task-mobile-card__details dd {
      margin: 0;
      color: #45484d;
      font-size: 12px;
      line-height: 18px;
      overflow-wrap: anywhere;
    }

    .task-mobile-command {
      margin-top: 12px;
    }

    .task-mobile-command > span {
      display: block;
      margin-bottom: 5px;
      color: #8a8d91;
      font-size: 11px;
    }

    .task-mobile-command pre {
      max-height: 160px;
      margin: 0;
      padding: 10px;
      overflow: auto;
      border-radius: 4px;
      background: #1f232b;
      color: #e6edf3;
      font-family: Consolas, 'Courier New', monospace;
      font-size: 11px;
      line-height: 18px;
      white-space: pre-wrap;
      overflow-wrap: anywhere;
    }

    .task-mobile-card__actions {
      display: grid;
      padding: 9px 10px;
      grid-template-columns: repeat(4, minmax(0, 1fr));
      gap: 6px;
      border-top: 1px solid #eceef1;
    }

    .task-mobile-card__actions .el-button {
      width: 100%;
      height: 34px;
      padding: 0 4px;
      border-radius: 4px;
      font-size: 12px;
    }

    .task-mobile-card__actions .el-button + .el-button {
      margin-left: 0;
    }

    .task-mobile-empty {
      padding: 48px 0;
      color: #9a9da1;
      font-size: 13px;
      text-align: center;
    }

    .task-mobile-pagination {
      display: grid;
      margin: 14px 0 4px;
      grid-template-columns: 40px minmax(0, 1fr) 40px;
      align-items: center;
      gap: 8px;
    }

    .task-mobile-pagination button {
      width: 40px;
      height: 36px;
      padding: 0;
      border: 1px solid #d8dce2;
      border-radius: 4px;
      outline: 0;
      background: #ffffff;
      color: #4c5056;
    }

    .task-mobile-pagination button:disabled {
      background: #f3f4f6;
      color: #b8bbc0;
    }

    .task-mobile-pagination span {
      color: #73767a;
      font-size: 12px;
      text-align: center;
    }
  }

  @media (max-width: 680px) {
    .task-filter-grid {
      grid-template-columns: 1fr;
    }

    .task-detail-grid--three {
      grid-template-columns: 1fr;
    }

    .task-detail-field--full {
      grid-column: auto;
    }
  }
</style>
