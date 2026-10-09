<template>
  <el-container>
    <el-main>
      <el-form class="task-log-filter-form">
        <div class="task-log-filter-grid">
        <el-form-item>
          <el-date-picker
            v-model="searchParams.start_time"
            type="datetime"
            clearable
            value-format="yyyy-MM-dd HH:mm:ss"
            placeholder="开始时间">
          </el-date-picker>
        </el-form-item>
        <el-form-item>
          <el-date-picker
            v-model="searchParams.end_time"
            type="datetime"
            clearable
            value-format="yyyy-MM-dd HH:mm:ss"
            placeholder="结束时间">
          </el-date-picker>
        </el-form-item>
        <el-form-item>
          <el-input
            v-model.trim="searchParams.keyword"
            clearable
            placeholder="搜索任务..."
            @keyup.enter.native="applyFilters">
          </el-input>
        </el-form-item>
        <el-form-item>
          <el-select v-model.trim="searchParams.protocol" clearable placeholder="执行方式">
            <el-option
            v-for="item in protocolList"
            :key="item.value"
            :label="item.label"
            :value="item.value">
            </el-option>
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-select v-model.trim="searchParams.status" clearable placeholder="状态">
            <el-option
              v-for="item in statusList"
              :key="item.value"
              :label="item.label"
              :value="item.value">
            </el-option>
          </el-select>
        </el-form-item>
        <div class="task-log-filter-actions">
          <el-button size="small" type="primary" icon="el-icon-search" @click="applyFilters">搜索</el-button>
          <el-button
            v-if="this.$store.getters.user.isAdmin"
            size="small"
            type="danger"
            icon="el-icon-delete"
            @click="clearLog">清空日志</el-button>
        </div>
        </div>
      </el-form>
      <el-pagination
        background
        layout="prev, pager, next, sizes, total"
        :total="logTotal"
        :page-size="20"
        @size-change="changePageSize"
        @current-change="changePage"
        @prev-click="changePage"
        @next-click="changePage">
      </el-pagination>
      <div class="log-table-toolbar">
        <el-popover placement="bottom-end" trigger="click">
          <div class="log-column-picker">
            <div class="log-column-picker__title">选择显示字段</div>
            <el-checkbox-group v-model="visibleLogColumns" @change="saveVisibleLogColumns">
              <el-checkbox v-for="column in logColumnOptions" :key="column.key" :label="column.key">{{column.label}}</el-checkbox>
            </el-checkbox-group>
          </div>
          <el-button slot="reference" class="log-column-picker-button" type="text" aria-label="选择显示字段" title="选择显示字段">
            <svg viewBox="0 0 48 48" aria-hidden="true">
              <path d="M18.2838 43.1713C14.9327 42.1736 11.9498 40.3213 9.58787 37.867C10.469 36.8227 11 35.4734 11 34.0001C11 30.6864 8.31371 28.0001 5 28.0001C4.79955 28.0001 4.60139 28.01 4.40599 28.0292C4.13979 26.7277 4 25.3803 4 24.0001C4 21.9095 4.32077 19.8938 4.91579 17.9995C4.94381 17.9999 4.97188 18.0001 5 18.0001C8.31371 18.0001 11 15.3138 11 12.0001C11 11.0488 10.7786 10.1493 10.3846 9.35011C12.6975 7.1995 15.5205 5.59002 18.6521 4.72314C19.6444 6.66819 21.6667 8.00013 24 8.00013C26.3333 8.00013 28.3556 6.66819 29.3479 4.72314C32.4795 5.59002 35.3025 7.1995 37.6154 9.35011C37.2214 10.1493 37 11.0488 37 12.0001C37 15.3138 39.6863 18.0001 43 18.0001C43.0281 18.0001 43.0562 17.9999 43.0842 17.9995C43.6792 19.8938 44 21.9095 44 24.0001C44 25.3803 43.8602 26.7277 43.594 28.0292C43.3986 28.01 43.2005 28.0001 43 28.0001C39.6863 28.0001 37 30.6864 37 34.0001C37 35.4734 37.531 36.8227 38.4121 37.867C36.0502 40.3213 33.0673 42.1736 29.7162 43.1713C28.9428 40.752 26.676 39.0001 24 39.0001C21.324 39.0001 19.0572 40.752 18.2838 43.1713Z" fill="none" stroke="currentColor" stroke-width="3.5" stroke-linejoin="round"></path>
              <path d="M24 31C27.866 31 31 27.866 31 24C31 20.134 27.866 17 24 17C20.134 17 17 20.134 17 24C17 27.866 20.134 31 24 31Z" fill="none" stroke="currentColor" stroke-width="3.5" stroke-linejoin="round"></path>
            </svg>
          </el-button>
        </el-popover>
      </div>
      <el-table
        :data="logs"
        border
        ref="table"
        style="width: 100%">
        <el-table-column type="expand">
          <template slot-scope="scope">
            <el-form label-position="left">
              <el-form-item>
                  重试次数: {{scope.row.retry_times}} <br>
                  cron表达式: {{scope.row.spec}} <br>
                  命令: {{scope.row.command}}
              </el-form-item>
            </el-form>
          </template>
        </el-table-column>
        <el-table-column
          v-if="isLogColumnVisible('id')" key="id"
          prop="id"
          label="ID"
          width="88">
        </el-table-column>
        <el-table-column
          v-if="isLogColumnVisible('task_id')" key="task_id"
          prop="task_id"
          label="任务ID"
          width="72">
        </el-table-column>
        <el-table-column
          v-if="isLogColumnVisible('name')" key="name"
          prop="name"
          label="任务名称"
          min-width="220">
        </el-table-column>
        <el-table-column
          v-if="isLogColumnVisible('protocol')" key="protocol"
          prop="protocol"
          label="执行方式"
          width="88"
          :formatter="formatProtocol">
        </el-table-column>
        <el-table-column
          v-if="isLogColumnVisible('hostname')" key="hostname"
          label="任务节点"
          min-width="240">
          <template slot-scope="scope">
            <div class="task-log-host">{{formatTaskHosts(scope.row.hostname)}}</div>
          </template>
        </el-table-column>
        <el-table-column
          v-if="isLogColumnVisible('start_time')" key="start_time"
          label="开始时间"
          width="165">
          <template slot-scope="scope">
            {{scope.row.start_time | formatTime}}
          </template>
        </el-table-column>
        <el-table-column
          v-if="isLogColumnVisible('end_time')" key="end_time"
          label="结束时间"
          width="165">
          <template slot-scope="scope">
            <span v-if="scope.row.status !== 1">{{scope.row.end_time | formatTime}}</span>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column
          v-if="isLogColumnVisible('total_time')" key="total_time"
          label="时长"
          width="72">
          <template slot-scope="scope">
            {{scope.row.total_time > 0 ? scope.row.total_time : 1}}秒
          </template>
        </el-table-column>
        <el-table-column
          v-if="isLogColumnVisible('status')" key="status"
          label="状态"
          width="72">
          <template slot-scope="scope">
            <span style="color:red" v-if="scope.row.status === 0">失败</span>
            <span style="color:green" v-else-if="scope.row.status === 1">执行中</span>
            <span v-else-if="scope.row.status === 2">成功</span>
            <span style="color:#4499EE" v-else-if="scope.row.status === 3">取消</span>
          </template>
        </el-table-column>
        <el-table-column
          label="执行结果"
          width="90" v-if="this.isAdmin">
          <template slot-scope="scope">
            <el-button type="success"
                       size="mini"
                       v-if="scope.row.status === 2"
                       @click="showTaskResult(scope.row)">结果</el-button>
            <el-button type="warning"
                       size="mini"
                       v-if="scope.row.status === 0"
                       @click="showTaskResult(scope.row)" >结果</el-button>
            <el-button type="danger"
                       size="mini"
                       v-if="scope.row.status === 1 && scope.row.protocol === 2"
                       @click="stopTask(scope.row)">停止
            </el-button>
          </template>
        </el-table-column>
        <el-table-column
          label="执行结果"
          width="90" v-else>
          <template slot-scope="scope">
            <el-button type="success"
                       size="mini"
                       v-if="scope.row.status === 2"
                       @click="showTaskResult(scope.row)">结果</el-button>
            <el-button type="warning"
                       size="mini"
                       v-if="scope.row.status === 0"
                       @click="showTaskResult(scope.row)" >结果</el-button>
          </template>
        </el-table-column>
      </el-table>
      <el-dialog
        class="task-result-dialog"
        :visible.sync="dialogVisible"
        width="68%"
        top="8vh">
        <span slot="title" class="task-result-dialog-title">任务执行结果</span>
        <div class="task-result-shell">
          <section class="task-result-section">
            <div class="task-result-section-heading">
              <h3><span>1</span>执行命令</h3>
              <small>{{(currentTaskResult.command || '').split('\n').length}} 行 · {{(currentTaskResult.command || '').length}} 字符</small>
            </div>
            <pre class="task-result-code task-result-command">{{currentTaskResult.command || '未记录执行命令'}}</pre>
          </section>

          <section class="task-result-section">
            <div class="task-result-section-heading">
              <h3><span>2</span>执行输出</h3>
              <small>{{(currentTaskResult.result || '').split('\n').length}} 行 · {{(currentTaskResult.result || '').length}} 字符</small>
            </div>
            <pre class="task-result-code task-result-output">{{currentTaskResult.result || '未返回执行结果'}}</pre>
          </section>
        </div>
        <span slot="footer" class="task-result-dialog-footer">
          <el-button size="small" type="primary" @click="dialogVisible = false">关闭</el-button>
        </span>
      </el-dialog>
    </el-main>
  </el-container>
</template>

<script>
import taskLogService from '../../api/taskLog'

export default {
  name: 'task-log',
  data () {
    return {
      logs: [],
      logColumnOptions: [
        {key: 'id', label: 'ID'},
        {key: 'task_id', label: '任务ID'},
        {key: 'name', label: '任务名称'},
        {key: 'protocol', label: '执行方式'},
        {key: 'hostname', label: '任务节点'},
        {key: 'start_time', label: '开始时间'},
        {key: 'end_time', label: '结束时间'},
        {key: 'total_time', label: '时长'},
        {key: 'status', label: '状态'}
      ],
      visibleLogColumns: [],
      logTotal: 0,
      searchParams: {
        page_size: 20,
        page: 1,
        keyword: '',
        protocol: '',
        status: '',
        start_time: '',
        end_time: ''
      },
      isAdmin: this.$store.getters.user.isAdmin,
      dialogVisible: false,
      currentTaskResult: {
        command: '',
        result: ''
      },
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
          value: '1',
          label: '失败'
        },
        {
          value: '2',
          label: '执行中'
        },
        {
          value: '3',
          label: '成功'
        },
        {
          value: '4',
          label: '取消'
        }
      ]
    }
  },
  created () {
    this.loadVisibleLogColumns()
    if (this.$route.query.task_id) {
      this.searchParams.keyword = this.$route.query.task_id
    }
    this.search()
  },
  methods: {
    logColumnStorageKey () {
      const uid = this.$store.getters.user.uid || 'anonymous'
      return `new-gocron:task-log-columns:${uid}`
    },
    loadVisibleLogColumns () {
      const defaults = this.logColumnOptions.map(column => column.key)
      try {
        const saved = JSON.parse(localStorage.getItem(this.logColumnStorageKey()) || 'null')
        this.visibleLogColumns = Array.isArray(saved)
          ? defaults.filter(key => saved.indexOf(key) !== -1)
          : defaults
      } catch (error) {
        this.visibleLogColumns = defaults
      }
    },
    saveVisibleLogColumns () {
      try {
        localStorage.setItem(this.logColumnStorageKey(), JSON.stringify(this.visibleLogColumns))
      } catch (error) {
        this.$message.warning('无法保存显示字段设置，请检查浏览器存储权限')
      }
      this.$nextTick(() => this.$refs.table.doLayout())
    },
    isLogColumnVisible (key) {
      return this.visibleLogColumns.indexOf(key) !== -1
    },
    formatProtocol (row, col) {
      if (row[col.property] === 1) {
        return 'http'
      }
      return 'shell'
    },
    formatTaskHosts (hostname) {
      return String(hostname || '').replace(/<br\s*\/?>/gi, '\n')
    },
    changePage (page) {
      this.searchParams.page = page
      this.search()
    },
    changePageSize (pageSize) {
      this.searchParams.page_size = pageSize
      this.search()
    },
    applyFilters () {
      this.searchParams.page = 1
      this.search()
    },
    search (callback = null) {
      if (this.searchParams.start_time && this.searchParams.end_time &&
        this.searchParams.start_time > this.searchParams.end_time) {
        this.$message.warning('开始时间不能晚于结束时间')
        return
      }
      taskLogService.list(this.searchParams, (data) => {
        this.logs = data.data
        this.logTotal = data.total

        if (callback) {
          callback()
        }
      })
    },
    clearLog () {
      this.$appConfirm(() => {
        taskLogService.clear(() => {
          this.searchParams.page = 1
          this.search()
        })
      })
    },
    stopTask (item) {
      taskLogService.stop(item.id, item.task_id, () => {
        this.search()
      })
    },
    showTaskResult (item) {
      this.dialogVisible = true
      this.currentTaskResult.command = item.command
      this.currentTaskResult.result = item.result
    }
  }
}
</script>
<style scoped>
  .log-table-toolbar {
    display: flex;
    height: 26px;
    margin-top: -12px;
    margin-bottom: 4px;
    justify-content: flex-end;
    align-items: center;
  }

  .log-column-picker-button {
    display: inline-flex;
    width: 30px;
    height: 28px;
    padding: 0;
    justify-content: center;
    align-items: center;
    border-radius: 5px;
    color: #858b94;
    transition: color 0.15s ease, background-color 0.15s ease;
  }

  .log-column-picker-button svg {
    width: 17px;
    height: 17px;
  }

  .log-column-picker-button:hover,
  .log-column-picker-button:focus {
    background: #f1f3f5;
    color: #50565f;
  }

  .log-column-picker {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  .log-column-picker__title {
    color: #606266;
    font-size: 13px;
    font-weight: 600;
  }

  .log-column-picker /deep/ .el-checkbox-group {
    display: flex;
    flex-direction: column;
    align-items: flex-start;
    gap: 8px;
  }

  .log-column-picker /deep/ .el-checkbox {
    margin: 0 !important;
  }

  .task-log-filter-form {
    margin-bottom: 18px;
    padding: 0;
    border: 0;
    background: transparent;
  }

  .task-log-filter-grid {
    display: grid;
    grid-template-columns: 210px 210px 180px 120px 90px auto;
    gap: 8px;
    align-items: center;
  }

  .task-log-filter-grid .el-form-item {
    margin-bottom: 0;
  }

  .task-log-filter-grid /deep/ .el-input,
  .task-log-filter-grid /deep/ .el-select,
  .task-log-filter-grid /deep/ .el-date-editor {
    width: 100%;
  }

  .task-log-filter-actions {
    display: flex;
    gap: 8px;
    align-items: center;
  }

  .task-log-filter-actions .el-button + .el-button {
    margin-left: 0;
  }

  .task-log-host {
    display: block;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .task-result-dialog-title {
    color: #2c2c2b;
    font-size: 16px;
    font-weight: 600;
  }

  .task-result-shell {
    color: #2c2c2b;
  }

  .task-result-section + .task-result-section {
    margin-top: 18px;
  }

  .task-result-section-heading {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
    margin-bottom: 8px;
  }

  .task-result-section-heading h3 {
    display: flex;
    min-width: 0;
    flex: 1;
    align-items: center;
    gap: 7px;
    margin: 0;
    color: #7d7a75;
    font-size: 12px;
    font-weight: 600;
  }

  .task-result-section-heading h3::after {
    height: 1px;
    flex: 1;
    background: #e6e5e3;
    content: '';
  }

  .task-result-section-heading h3 span {
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

  .task-result-section-heading small {
    flex: 0 0 auto;
    color: #8f8c87;
    font-size: 11px;
    font-weight: 400;
  }

  .task-result-code {
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
    white-space: pre-wrap;
    overflow-wrap: anywhere;
  }

  .task-result-command {
    min-height: 54px;
    max-height: 140px;
  }

  .task-result-output {
    min-height: 180px;
    max-height: 42vh;
  }

  .task-result-dialog-footer {
    display: flex;
    justify-content: flex-end;
  }

  .task-result-dialog /deep/ .el-dialog {
    max-width: 1080px;
    border-radius: 6px;
  }

  .task-result-dialog /deep/ .el-dialog__header {
    padding: 16px 20px 12px;
    border-bottom: 1px solid #e6e5e3;
  }

  .task-result-dialog /deep/ .el-dialog__body {
    padding: 18px 20px;
  }

  .task-result-dialog /deep/ .el-dialog__footer {
    padding: 10px 20px 12px;
    border-top: 1px solid #e6e5e3;
  }

  @media (max-width: 900px) {
    .task-log-filter-grid {
      grid-template-columns: repeat(2, minmax(180px, 1fr));
    }

    .task-log-filter-actions {
      grid-column: 1 / -1;
    }
  }

  @media (max-width: 620px) {
    .task-log-filter-grid {
      grid-template-columns: 1fr;
    }

    .task-result-dialog /deep/ .el-dialog {
      width: calc(100% - 24px) !important;
      margin-top: 12px !important;
    }

    .task-result-section-heading {
      align-items: flex-start;
      flex-direction: column;
      gap: 4px;
    }

    .task-result-section-heading h3 {
      width: 100%;
    }
  }
</style>
