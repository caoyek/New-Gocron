<template>
  <el-container>
    <el-main class="system-settings-main">
      <notification-tab></notification-tab>
      <el-form ref="form" class="system-settings-form" :model="form" :rules="formRules" label-position="top">
        <section class="settings-section">
          <div class="settings-section__header">
            <h2>Slack 通知</h2>
          </div>
          <div class="settings-grid">
            <el-form-item class="settings-span-full" label="Webhook URL" prop="url">
              <el-input v-model="form.url"></el-input>
            </el-form-item>
            <el-form-item class="settings-span-full" label="通知模板" prop="template">
              <el-input v-model="form.template" type="textarea" :rows="7"></el-input>
            </el-form-item>
          </div>
          <div class="settings-actions">
            <el-button size="small" type="primary" icon="el-icon-check" @click="submit">保存</el-button>
          </div>
        </section>

        <section class="settings-section settings-collection-section">
          <div class="settings-section__header">
            <h2>Channel</h2>
            <el-button size="small" type="success" icon="el-icon-plus" @click="createChannel">新增</el-button>
          </div>
          <div class="settings-tags">
            <el-tag
              v-for="item in channels"
              :key="item.id"
              size="small"
              closable
              @close="deleteChannel(item)">
              {{item.name}}
            </el-tag>
            <span v-if="channels.length === 0" class="settings-empty">暂无 Channel</span>
          </div>
        </section>
      </el-form>
      <el-dialog
        class="compact-settings-dialog"
        title="新增 Channel"
        :visible.sync="dialogVisible"
        width="420px">
        <el-form class="settings-dialog-form" label-position="top">
          <el-form-item label="Channel名称">
            <el-input v-model.trim="channel" v-focus></el-input>
          </el-form-item>
        </el-form>
        <span slot="footer" class="settings-dialog-actions">
          <el-button size="small" @click="dialogVisible = false">取消</el-button>
          <el-button size="small" type="primary" @click="saveChannel">确定</el-button>
        </span>
      </el-dialog>
    </el-main>
  </el-container>
</template>

<script>
import notificationTab from './tab'
import notificationService from '../../../api/notification'
export default {
  name: 'notification-slack',
  data () {
    return {
      dialogVisible: false,
      form: {
        url: '',
        template: ''
      },
      formRules: {
        url: [
          {type: 'url', required: true, message: '请输入有效的通知URL', trigger: 'blur'}
        ],
        template: [
          {required: true, message: '请输入通知模板', trigger: 'blur'}
        ]
      },
      channels: [],
      channel: ''
    }
  },
  components: {notificationTab},
  created () {
    this.init()
  },
  methods: {
    createChannel () {
      this.dialogVisible = true
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
      notificationService.updateSlack(this.form, () => {
        this.$message.success('更新成功')
        this.init()
      })
    },
    saveChannel () {
      if (this.channel === '') {
        this.$message.error('请输入Channel名称')
        return
      }
      notificationService.createSlackChannel(this.channel, () => {
        this.dialogVisible = false
        this.init()
      })
    },
    deleteChannel (item) {
      notificationService.removeSlackChannel(item.id, () => {
        this.init()
      })
    },
    init () {
      this.channel = ''
      notificationService.slack((data) => {
        this.form.url = data.url
        this.form.template = data.template
        this.channels = data.channels || []
      })
    }
  }
}
</script>
