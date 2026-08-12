<template>
  <el-container>
    <el-main class="system-log-main">
      <div class="system-log-toolbar">
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
      </div>
      <el-table
        :data="logs"
        border
        ref="table"
        class="login-log-table"
        style="width: 100%">
        <el-table-column
          prop="id"
          label="ID"
          width="72">
        </el-table-column>
        <el-table-column
          prop="username"
          label="账号"
          min-width="150">
        </el-table-column>
        <el-table-column
          prop="ip"
          label="来源 IP"
          min-width="150">
        </el-table-column>
        <el-table-column
          label="结果"
          width="120">
          <template slot-scope="scope">
            <span class="login-result" :class="`is-${scope.row.result}`">{{resultText(scope.row.result)}}</span>
          </template>
        </el-table-column>
        <el-table-column
          prop="message"
          label="说明"
          min-width="180">
        </el-table-column>
        <el-table-column
          label="时间"
          width="180">
          <template slot-scope="scope">
            {{scope.row.created | formatTime}}
          </template>
        </el-table-column>
      </el-table>
    </el-main>
  </el-container>
</template>

<script>
import systemService from '../../api/system'
export default {
  name: 'login-log',
  data () {
    return {
      logs: [],
      logTotal: 0,
      searchParams: {
        page_size: 20,
        page: 1
      }
    }
  },
  created () {
    this.search()
  },
  methods: {
    changePage (page) {
      this.searchParams.page = page
      this.search()
    },
    changePageSize (pageSize) {
      this.searchParams.page_size = pageSize
      this.search()
    },
    search () {
      systemService.loginLogList(this.searchParams, (data) => {
        this.logs = data.data
        this.logTotal = data.total
      })
    },
    resultText (result) {
      const labels = {
        success: '成功',
        password_failure: '密码错误',
        blocked: '封禁拒绝',
        whitelist_rejected: '白名单拒绝'
      }
      return labels[result] || result || '-'
    }
  }
}
</script>

<style scoped>
.system-log-main {
  overflow: hidden;
}

.system-log-toolbar {
  margin-bottom: 10px;
}

.login-result {
  display: inline-flex;
  min-width: 64px;
  height: 24px;
  padding: 0 8px;
  box-sizing: border-box;
  align-items: center;
  justify-content: center;
  border: 1px solid #dcdfe3;
  border-radius: 4px;
  background: #f7f8f9;
  color: #606266;
  font-size: 12px;
  line-height: 22px;
}

.login-result.is-success {
  border-color: #b8dfc8;
  background: #f0f9f4;
  color: #21834c;
}

.login-result.is-password_failure,
.login-result.is-blocked,
.login-result.is-whitelist_rejected {
  border-color: #f0b9b9;
  background: #fff2f2;
  color: #c43d3d;
}

@media (max-width: 620px) {
  .system-log-main {
    overflow: auto;
  }
}
</style>
