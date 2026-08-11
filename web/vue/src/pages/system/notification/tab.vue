<template>
  <div class="notification-nav">
    <el-tabs v-model="activeName" class="notification-tabs" @tab-click="changeTab">
      <el-tab-pane label="邮件" name="email"></el-tab-pane>
      <el-tab-pane label="Slack" name="slack"></el-tab-pane>
      <el-tab-pane label="企微群推送" name="webhook"></el-tab-pane>
    </el-tabs>
    <el-tooltip placement="bottom-end">
      <div slot="content">
        通知模板变量<br>
        TaskId：任务ID<br>
        TaskName：任务名称<br>
        Status：执行状态<br>
        Result：执行输出
      </div>
      <button class="notification-help-button" type="button" aria-label="通知模板变量说明">
        <i class="el-icon-warning"></i>
      </button>
    </el-tooltip>
  </div>
</template>

<script>
import './settings.css'

export default {
  name: 'notification-tab',
  data () {
    return {
      activeName: ''
    }
  },
  created () {
    const segments = this.$route.path.split('/')
    if (segments.length !== 4) {
      return 'email'
    }
    this.activeName = segments[3]
  },
  methods: {
    changeTab (item) {
      this.$router.push(`/system/notification/${item.name}`)
    }
  }
}
</script>
