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
        style="width: 100%">
        <el-table-column
          prop="id"
          label="ID">
        </el-table-column>
        <el-table-column
          prop="username"
          label="用户名">
        </el-table-column>
        <el-table-column
          prop="ip"
          label="登录IP">
        </el-table-column>
        <el-table-column
          label="登录时间"
          width="">
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

@media (max-width: 620px) {
  .system-log-main {
    overflow: auto;
  }
}
</style>
