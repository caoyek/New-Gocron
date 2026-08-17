<template>
  <el-container>
    <el-main class="system-settings-main">
      <notification-tab></notification-tab>

      <section class="settings-section settings-collection-section webhook-template-section">
        <div class="settings-section__header">
          <div class="settings-heading-with-help">
            <h2>通知模板</h2>
            <el-tooltip content="使用 POST 请求，Content-Type 为 application/json" placement="top">
              <button class="settings-help-button" type="button" aria-label="企微群推送请求说明">
                <i class="el-icon-warning"></i>
              </button>
            </el-tooltip>
          </div>
          <el-button size="small" type="success" icon="el-icon-plus" @click="openCreateTemplate">新增</el-button>
        </div>
        <el-table :data="templates" size="small" empty-text="暂无通知模板，请先新增">
          <el-table-column label="名称" width="200">
            <template slot-scope="scope">
              <span>{{scope.row.name}}</span>
              <el-tag v-if="scope.row.is_default" class="default-template-tag" size="mini" type="info">默认</el-tag>
            </template>
          </el-table-column>
          <el-table-column label="模板内容" min-width="360" show-overflow-tooltip>
            <template slot-scope="scope">
              <span class="template-content">{{scope.row.content}}</span>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="110" align="center">
            <template slot-scope="scope">
              <div class="webhook-actions">
                <el-tooltip content="编辑" placement="top">
                  <el-button size="mini" type="primary" icon="el-icon-edit" aria-label="编辑" @click="openEditTemplate(scope.row)"></el-button>
                </el-tooltip>
                <el-tooltip v-if="!scope.row.is_default" content="删除" placement="top">
                  <el-button size="mini" type="danger" icon="el-icon-delete" aria-label="删除" @click="deleteTemplate(scope.row)"></el-button>
                </el-tooltip>
              </div>
            </template>
          </el-table-column>
        </el-table>
      </section>

      <section class="settings-section settings-collection-section webhook-group-section">
        <div class="settings-section__header">
          <h2>企微群机器人</h2>
          <el-button size="small" type="success" icon="el-icon-plus" @click="openCreateGroup">新增</el-button>
        </div>
        <el-table :data="groups" size="small" empty-text="暂无企微群，请先新增">
          <el-table-column prop="name" label="名称" width="200"></el-table-column>
          <el-table-column label="Webhook URL" min-width="360" show-overflow-tooltip>
            <template slot-scope="scope">
              <span class="webhook-url">{{scope.row.url}}</span>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="110" align="center">
            <template slot-scope="scope">
              <div class="webhook-actions">
                <el-tooltip content="编辑" placement="top">
                  <el-button size="mini" type="primary" icon="el-icon-edit" aria-label="编辑" @click="openEditGroup(scope.row)"></el-button>
                </el-tooltip>
                <el-tooltip content="删除" placement="top">
                  <el-button size="mini" type="danger" icon="el-icon-delete" aria-label="删除" @click="deleteGroup(scope.row)"></el-button>
                </el-tooltip>
              </div>
            </template>
          </el-table-column>
        </el-table>
      </section>

      <el-dialog
        class="compact-settings-dialog"
        :title="templateForm.id ? '编辑通知模板' : '新增通知模板'"
        :visible.sync="templateDialogVisible"
        width="620px"
        append-to-body>
        <el-form ref="templateForm" class="settings-dialog-form" :model="templateForm" :rules="templateRules" label-position="top">
          <el-form-item label="名称" prop="name">
            <el-input v-model.trim="templateForm.name" :disabled="templateForm.is_default" placeholder="例如：任务失败告警"></el-input>
          </el-form-item>
          <el-form-item label="模板内容" prop="content">
            <el-input v-model.trim="templateForm.content" type="textarea" :rows="9" placeholder="请输入企微机器人接收的 JSON 模板"></el-input>
          </el-form-item>
        </el-form>
        <span slot="footer" class="settings-dialog-actions">
          <el-button size="small" @click="templateDialogVisible = false">取消</el-button>
          <el-button size="small" type="primary" @click="submitTemplate">保存</el-button>
        </span>
      </el-dialog>

      <el-dialog
        class="compact-settings-dialog"
        :title="groupForm.id ? '编辑企微群' : '新增企微群'"
        :visible.sync="groupDialogVisible"
        width="560px"
        append-to-body>
        <el-form ref="groupForm" class="settings-dialog-form" :model="groupForm" :rules="groupRules" label-position="top">
          <el-form-item label="名称" prop="name">
            <el-input v-model.trim="groupForm.name" placeholder="例如：运维告警群"></el-input>
          </el-form-item>
          <el-form-item label="Webhook URL" prop="url">
            <el-input v-model.trim="groupForm.url" placeholder="请输入企微群机器人 Webhook URL"></el-input>
          </el-form-item>
        </el-form>
        <span slot="footer" class="settings-dialog-actions">
          <el-button size="small" @click="groupDialogVisible = false">取消</el-button>
          <el-button size="small" type="primary" @click="submitGroup">保存</el-button>
        </span>
      </el-dialog>
    </el-main>
  </el-container>
</template>

<script>
import notificationTab from './tab'
import notificationService from '../../../api/notification'

function emptyTemplate () {
  return {
    id: 0,
    name: '',
    content: '',
    is_default: false
  }
}

function emptyGroup () {
  return {
    id: 0,
    name: '',
    url: ''
  }
}

export default {
  name: 'notification-webhook',
  components: {notificationTab},
  data () {
    return {
      templateDialogVisible: false,
      groupDialogVisible: false,
      templates: [],
      groups: [],
      templateForm: emptyTemplate(),
      groupForm: emptyGroup(),
      templateRules: {
        name: [
          {required: true, message: '请输入模板名称', trigger: 'blur'},
          {max: 48, message: '模板名称不能超过48个字符', trigger: 'blur'}
        ],
        content: [
          {required: true, message: '请输入通知模板', trigger: 'blur'},
          {max: 4096, message: '通知模板不能超过4096个字符', trigger: 'blur'}
        ]
      },
      groupRules: {
        name: [
          {required: true, message: '请输入企微群名称', trigger: 'blur'},
          {max: 64, message: '企微群名称不能超过64个字符', trigger: 'blur'}
        ],
        url: [
          {type: 'url', required: true, message: '请输入有效的 Webhook URL', trigger: 'blur'}
        ]
      }
    }
  },
  created () {
    this.init()
  },
  methods: {
    openCreateTemplate () {
      this.templateForm = emptyTemplate()
      this.templateDialogVisible = true
      this.clearValidation('templateForm')
    },
    openEditTemplate (template) {
      this.templateForm = Object.assign({}, template)
      this.templateDialogVisible = true
      this.clearValidation('templateForm')
    },
    submitTemplate () {
      this.$refs.templateForm.validate((valid) => {
        if (!valid) {
          return false
        }
        const callback = () => {
          this.templateDialogVisible = false
          this.$message.success('通知模板保存成功')
          this.init()
        }
        if (this.templateForm.id) {
          notificationService.updateWebhookTemplate(this.templateForm.id, this.templateForm, callback)
          return
        }
        notificationService.createWebhookTemplate(this.templateForm, callback)
      })
    },
    deleteTemplate (template) {
      this.$confirm(`确定删除通知模板“${template.name}”吗？正在被任务使用的通知模板无法删除。`, '删除确认', {
        confirmButtonText: '删除',
        cancelButtonText: '取消',
        type: 'warning'
      }).then(() => {
        notificationService.removeWebhookTemplate(template.id, () => {
          this.$message.success('删除成功')
          this.init()
        })
      }).catch(() => {})
    },
    openCreateGroup () {
      this.groupForm = emptyGroup()
      this.groupDialogVisible = true
      this.clearValidation('groupForm')
    },
    openEditGroup (group) {
      this.groupForm = Object.assign({}, group)
      this.groupDialogVisible = true
      this.clearValidation('groupForm')
    },
    clearValidation (formName) {
      this.$nextTick(() => {
        if (this.$refs[formName]) {
          this.$refs[formName].clearValidate()
        }
      })
    },
    submitGroup () {
      this.$refs.groupForm.validate((valid) => {
        if (!valid) {
          return false
        }
        const callback = () => {
          this.groupDialogVisible = false
          this.$message.success('企微群保存成功')
          this.init()
        }
        if (this.groupForm.id) {
          notificationService.updateWebhookGroup(this.groupForm.id, this.groupForm, callback)
          return
        }
        notificationService.createWebhookGroup(this.groupForm, callback)
      })
    },
    deleteGroup (group) {
      this.$confirm(`确定删除企微群“${group.name}”吗？正在被任务使用的企微群无法删除。`, '删除确认', {
        confirmButtonText: '删除',
        cancelButtonText: '取消',
        type: 'warning'
      }).then(() => {
        notificationService.removeWebhookGroup(group.id, () => {
          this.$message.success('删除成功')
          this.init()
        })
      }).catch(() => {})
    },
    init () {
      notificationService.webhook((data) => {
        data = data || {}
        this.templates = data.templates || []
        this.groups = data.groups || []
      })
    }
  }
}
</script>

<style scoped>
.webhook-template-section {
  margin-top: 16px;
}

.settings-heading-with-help {
  display: flex;
  align-items: center;
  gap: 8px;
}

.webhook-group-section {
  margin-top: 16px;
}

.template-content,
.webhook-url {
  display: block;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.default-template-tag {
  margin-left: 8px;
}

.webhook-actions {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
}

.webhook-actions /deep/ .el-button {
  width: 28px;
  height: 28px;
  padding: 0;
}
</style>
