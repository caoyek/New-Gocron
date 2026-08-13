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
          prop="id"
          label="ID"
          width="88">
        </el-table-column>
        <el-table-column
          prop="task_id"
          label="任务ID"
          width="72">
        </el-table-column>
        <el-table-column
          prop="name"
          label="任务名称"
          min-width="220">
        </el-table-column>
        <el-table-column
          prop="protocol"
          label="执行方式"
          width="88"
          :formatter="formatProtocol">
        </el-table-column>
        <el-table-column
          label="任务节点"
          min-width="240">
          <template slot-scope="scope">
            <div class="task-log-host" v-html="scope.row.hostname">{{scope.row.hostname}}</div>
          </template>
        </el-table-column>
        <el-table-column
          label="开始时间"
          width="165">
          <template slot-scope="scope">
            {{scope.row.start_time | formatTime}}
          </template>
        </el-table-column>
        <el-table-column
          label="结束时间"
          width="165">
          <template slot-scope="scope">
            <span v-if="scope.row.status !== 1">{{scope.row.end_time | formatTime}}</span>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column
          label="时长"
          width="72">
          <template slot-scope="scope">
            {{scope.row.total_time > 0 ? scope.row.total_time : 1}}秒
          </template>
        </el-table-column>
        <el-table-column
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
    if (this.$route.query.task_id) {
      this.searchParams.keyword = this.$route.query.task_id
    }
    this.search()
  },
  methods: {
    formatProtocol (row, col) {
      if (row[col.property] === 1) {
        return 'http'
      }
      return 'shell'
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

  .task-log-host /deep/ br {
    display: none;
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
