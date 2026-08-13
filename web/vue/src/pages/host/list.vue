<template>
  <el-container>
    <el-main>
      <el-form class="host-filter-form">
        <div class="host-filter-grid">
          <el-form-item>
            <el-input
              v-model.trim="searchParams.keyword"
              clearable
              placeholder="搜索节点..."
              @keyup.enter.native="search()">
            </el-input>
          </el-form-item>
          <div class="host-filter-actions">
            <el-button size="small" type="primary" icon="el-icon-search" @click="search()">搜索</el-button>
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
        background
        layout="prev, pager, next, sizes, total"
        :total="hostTotal"
        :page-size="20"
        @size-change="changePageSize"
        @current-change="changePage"
        @prev-click="changePage"
        @next-click="changePage">
      </el-pagination>
      <el-table
        :data="hosts"
        tooltip-effect="dark"
        border
        style="width: 100%">
        <el-table-column
          prop="id"
          label="节点ID">
        </el-table-column>
        <el-table-column
          prop="alias"
          label="节点名称">
        </el-table-column>
        <el-table-column
          prop="name"
          label="主机名">
        </el-table-column>
        <el-table-column
          prop="port"
          label="端口">
        </el-table-column>
        <el-table-column label="查看任务">
          <template slot-scope="scope">
            <el-button size="mini" type="success" @click="toTasks(scope.row)">任务</el-button>
          </template>
        </el-table-column>
        <el-table-column
          prop="remark"
          label="备注">
        </el-table-column>
        <el-table-column label="操作" width="220" v-if="this.isAdmin">
          <template slot-scope="scope">
            <div class="host-actions">
              <el-button size="mini" type="primary" @click="toEdit(scope.row)">编辑</el-button>
              <el-button size="mini" type="info" @click="ping(scope.row)">测试</el-button>
              <el-button size="mini" type="danger" @click="remove(scope.row)">删除</el-button>
            </div>
          </template>
        </el-table-column>
      </el-table>
      <host-editor
        :visible.sync="editorVisible"
        :host-id="editingHostId"
        @saved="handleHostSaved">
      </host-editor>
    </el-main>
  </el-container>
</template>

<script>
import hostService from '../../api/host'
import hostEditor from './edit'
export default {
  name: 'host-list',
  data () {
    return {
      hosts: [],
      hostTotal: 0,
      editorVisible: false,
      editingHostId: null,
      searchParams: {
        page_size: 20,
        page: 1,
        keyword: ''
      },
      isAdmin: this.$store.getters.user.isAdmin
    }
  },
  components: {hostEditor},
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
    search (callback = null) {
      hostService.list(this.searchParams, (data) => {
        this.hosts = data.data
        this.hostTotal = data.total
        if (callback) {
          callback()
        }
      })
    },
    remove (item) {
      this.$appConfirm(() => {
        hostService.remove(item.id, () => this.search())
      })
    },
    ping (item) {
      hostService.ping(item.id, () => {
        this.$message.success('连接成功')
      })
    },
    toEdit (item) {
      this.editingHostId = item ? item.id : null
      this.editorVisible = true
    },
    handleHostSaved () {
      this.search()
    },
    toTasks (item) {
      this.$router.push(
        {
          path: '/task',
          query: {
            host_id: item.id
          }
        })
    }
  }
}
</script>
<style scoped>
.host-filter-form {
  margin-bottom: 18px;
  padding: 0;
  border: 0;
  background: transparent;
}

.host-filter-grid {
  display: grid;
  grid-template-columns: 180px auto;
  gap: 8px;
  align-items: center;
}

.host-filter-grid .el-form-item {
  margin-bottom: 0;
}

.host-filter-actions,
.host-actions {
  display: flex;
  gap: 8px;
  align-items: center;
  white-space: nowrap;
}

.host-filter-actions .el-button + .el-button,
.host-actions .el-button + .el-button {
  margin-left: 0;
}

@media (max-width: 620px) {
  .host-filter-grid {
    grid-template-columns: 1fr;
  }
}
</style>
