<template>
  <el-container>
    <el-main class="system-settings-main">
      <notification-tab></notification-tab>
      <el-form ref="form" class="system-settings-form" :model="form" :rules="formRules" label-position="top">
        <section class="settings-section">
          <div class="settings-section__header">
            <h2>企微群推送</h2>
            <el-tooltip content="使用 POST 请求，Content-Type 为 application/json" placement="top">
              <button class="settings-help-button" type="button" aria-label="企微群推送请求说明">
                <i class="el-icon-warning"></i>
              </button>
            </el-tooltip>
          </div>
          <div class="settings-grid">
            <el-form-item class="settings-span-full" label="URL" prop="url">
              <el-input v-model.trim="form.url"></el-input>
            </el-form-item>
            <el-form-item class="settings-span-full" label="通知模板" prop="template">
              <el-input v-model.trim="form.template" type="textarea" :rows="7"></el-input>
            </el-form-item>
          </div>
          <div class="settings-actions">
            <el-button size="small" type="primary" icon="el-icon-check" @click="submit()">保存</el-button>
          </div>
        </section>
      </el-form>
    </el-main>
  </el-container>
</template>

<script>
import notificationTab from './tab'
import notificationService from '../../../api/notification'
export default {
  name: 'notification-webhook',
  data () {
    return {
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
      }
    }
  },
  components: {notificationTab},
  created () {
    this.init()
  },
  methods: {
    submit () {
      this.$refs['form'].validate((valid) => {
        if (!valid) {
          return false
        }
        this.save()
      })
    },
    save () {
      notificationService.updateWebHook(this.form, () => {
        this.$message.success('更新成功')
        this.init()
      })
    },
    init () {
      notificationService.webhook((data) => {
        this.form.url = data.url
        this.form.template = data.template
      })
    }
  }
}
</script>
