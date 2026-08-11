<template>
  <el-container>
    <el-main class="system-settings-main">
      <notification-tab></notification-tab>
      <el-form ref="form" class="system-settings-form" :model="form" :rules="formRules" label-position="top">
        <section class="settings-section">
          <div class="settings-section__header">
            <h2>邮件服务器</h2>
          </div>
          <div class="settings-grid">
            <el-form-item label="SMTP服务器" prop="host">
              <el-input v-model="form.host"></el-input>
            </el-form-item>
            <el-form-item label="端口" prop="port">
              <el-input v-model.number="form.port"></el-input>
            </el-form-item>
            <el-form-item label="用户名" prop="user">
              <el-input v-model="form.user"></el-input>
            </el-form-item>
            <el-form-item label="密码" prop="password">
              <el-input v-model="form.password" type="password"></el-input>
            </el-form-item>
            <el-form-item class="settings-span-full" prop="template">
              <template slot="label">
                通知模板
                <el-tooltip content="支持 HTML 内容" placement="top">
                  <i class="settings-label-help el-icon-warning"></i>
                </el-tooltip>
              </template>
              <el-input v-model="form.template" type="textarea" :rows="6"></el-input>
            </el-form-item>
          </div>
          <div class="settings-actions">
            <el-button size="small" type="primary" icon="el-icon-check" @click="submit()">保存</el-button>
          </div>
        </section>

        <section class="settings-section settings-collection-section">
          <div class="settings-section__header">
            <h2>通知用户</h2>
            <el-button size="small" type="success" icon="el-icon-plus" @click="createUser">新增</el-button>
          </div>
          <div class="settings-tags">
            <el-tag
              v-for="item in receivers"
              :key="item.email"
              size="small"
              closable
              @close="deleteUser(item)">
              {{item.username}} - {{item.email}}
            </el-tag>
            <span v-if="receivers.length === 0" class="settings-empty">暂无通知用户</span>
          </div>
        </section>
      </el-form>
      <el-dialog
        class="compact-settings-dialog"
        title="新增通知用户"
        :visible.sync="dialogVisible"
        width="420px">
        <el-form class="settings-dialog-form" label-position="top">
          <el-form-item label="用户名">
            <el-input v-model.trim="username" v-focus></el-input>
          </el-form-item>
          <el-form-item label="邮箱地址">
            <el-input v-model.trim="email"></el-input>
          </el-form-item>
        </el-form>
        <span slot="footer" class="settings-dialog-actions">
          <el-button size="small" @click="dialogVisible = false">取消</el-button>
          <el-button size="small" type="primary" @click="saveUser">确定</el-button>
        </span>
      </el-dialog>
    </el-main>
  </el-container>
</template>

<script>
import notificationTab from './tab'
import notificationService from '../../../api/notification'
export default {
  name: 'notification-email',
  data () {
    return {
      form: {
        host: '',
        port: 465,
        user: '',
        password: '',
        template: ''
      },
      formRules: {
        host: [
          {required: true, message: '请输入邮件服务器地址', trigger: 'blur'}
        ],
        port: [
          {type: 'number', required: true, message: '请输入有效的端口', trigger: 'blur'}
        ],
        user: [
          {required: true, message: '请输入用户email', trigger: 'blur'}
        ],
        password: [
          {required: true, message: '请输入密码', trigger: 'blur'}
        ],
        template: [
          {required: true, message: '请输入通知模板内容', trigger: 'blur'}
        ]
      },
      receivers: [],
      username: '',
      email: '',
      dialogVisible: false
    }
  },
  components: {notificationTab},
  created () {
    this.init()
  },
  methods: {
    createUser () {
      this.dialogVisible = true
    },
    saveUser () {
      if (this.username === '' || this.email === '') {
        this.$message.error('参数不完整')
        return
      }
      notificationService.createMailUser({
        username: this.username,
        email: this.email
      }, () => {
        this.dialogVisible = false
        this.init()
      })
    },
    deleteUser (item) {
      notificationService.removeMailUser(item.id, () => {
        this.init()
      })
    },
    submit () {
      this.$refs['form'].validate((valid) => {
        if (!valid) {
          return false
        }
        this.save()
      })
    },
    save () {
      notificationService.updateMail(this.form, () => {
        this.$message.success('更新成功')
        this.init()
      })
    },
    init () {
      this.username = ''
      this.email = ''
      notificationService.mail((data) => {
        this.form.host = data.host
        if (data.port) {
          this.form.port = data.port
        }
        this.form.user = data.user
        this.form.password = data.password
        this.form.template = data.template
        this.receivers = data.mail_users || []
      })
    }
  }
}
</script>
