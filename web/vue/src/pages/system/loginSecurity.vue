<template>
  <el-container>
    <el-main class="security-main">
      <header class="security-page-header">
        <div>
          <h1>登录安全</h1>
          <p>当前访问 IP：<strong>{{peerIp || '-'}}</strong></p>
        </div>
        <el-button size="small" type="primary" icon="el-icon-check" @click="submit">保存设置</el-button>
      </header>

      <el-form ref="form" :model="form" label-position="top" class="security-form">
        <section class="security-section">
          <div class="security-section__title">
            <div>
              <h2><i class="el-icon-warning"></i>失败封禁</h2>
              <p>同时按来源 IP 和登录账号统计，任意一项达到阈值都会被封禁。</p>
            </div>
            <el-switch v-model="form.block_enabled"></el-switch>
          </div>
          <div class="security-policy-row" :class="{'is-disabled': !form.block_enabled}">
            <div class="security-policy-field">
              <span>统计周期</span>
              <el-input-number v-model="form.window_minutes" :min="1" :max="1440" :disabled="!form.block_enabled" size="small"></el-input-number>
              <em>分钟</em>
            </div>
            <div class="security-policy-field">
              <span>允许失败</span>
              <el-input-number v-model="form.max_failures" :min="1" :max="100" :disabled="!form.block_enabled" size="small"></el-input-number>
              <em>次</em>
            </div>
            <div class="security-policy-field">
              <span>封禁时长</span>
              <el-input-number v-model="form.block_minutes" :min="1" :max="10080" :disabled="!form.block_enabled" size="small"></el-input-number>
              <em>分钟</em>
            </div>
          </div>
        </section>

        <section class="security-section">
          <div class="security-section__title">
            <div>
              <h2><i class="el-icon-circle-check-outline"></i>后台 IP 白名单</h2>
              <p>启用后，只有名单内的真实 TCP 来源 IP 可以登录和访问后台。</p>
            </div>
            <el-switch v-model="form.whitelist_enabled"></el-switch>
          </div>
          <el-form-item class="security-whitelist-field">
            <el-input
              v-model="form.whitelist"
              type="textarea"
              :rows="5"
              :disabled="!form.whitelist_enabled"
              placeholder="每行一个 IP 或 CIDR，例如：&#10;192.168.1.10&#10;192.168.1.0/24">
            </el-input>
          </el-form-item>
          <p class="security-notice">
            经过 Nginx、Caddy 等反向代理时，应用看到的是代理服务器 IP；公网部署还应在反向代理或防火墙层配置相同白名单。
          </p>
        </section>
      </el-form>

      <section class="security-section security-blocks">
        <div class="security-section__title">
          <div>
            <h2><i class="el-icon-tickets"></i>当前封禁</h2>
            <p>这里只显示尚未到期的 IP 和账号封禁。</p>
          </div>
          <el-button size="small" icon="el-icon-refresh" @click="init">刷新</el-button>
        </div>
        <el-table :data="blocks" border empty-text="暂无封禁" class="security-block-table">
          <el-table-column label="类型" width="90">
            <template slot-scope="scope">{{scope.row.scope === 'ip' ? 'IP' : '账号'}}</template>
          </el-table-column>
          <el-table-column prop="value" label="对象"></el-table-column>
          <el-table-column label="封禁至" width="180">
            <template slot-scope="scope">{{scope.row.blocked_until | formatTime}}</template>
          </el-table-column>
          <el-table-column label="操作" width="90" align="center">
            <template slot-scope="scope">
              <el-button type="text" size="small" @click="removeBlock(scope.row)">解封</el-button>
            </template>
          </el-table-column>
        </el-table>
      </section>
    </el-main>
  </el-container>
</template>

<script>
import systemService from '../../api/system'

export default {
  name: 'login-security',
  data () {
    return {
      peerIp: '',
      blocks: [],
      form: {
        block_enabled: true,
        window_minutes: 10,
        max_failures: 5,
        block_minutes: 30,
        whitelist_enabled: false,
        whitelist: ''
      }
    }
  },
  created () {
    this.init()
  },
  methods: {
    init () {
      systemService.loginSecurity((data) => {
        const policy = data.policy || {}
        this.peerIp = data.peer_ip || ''
        this.blocks = data.blocks || []
        this.form.block_enabled = policy.block_enabled !== false
        this.form.window_minutes = policy.window_minutes || 10
        this.form.max_failures = policy.max_failures || 5
        this.form.block_minutes = policy.block_minutes || 30
        this.form.whitelist_enabled = policy.whitelist_enabled === true
        this.form.whitelist = (policy.whitelist || []).join('\n')
      })
    },
    submit () {
      if (this.form.whitelist_enabled && !this.form.whitelist.trim()) {
        this.$message.error('启用白名单前请至少填写一个 IP 或 CIDR')
        return
      }
      systemService.updateLoginSecurity(this.form, () => {
        this.$message.success('保存成功')
        this.init()
      })
    },
    removeBlock (item) {
      systemService.removeLoginBlock(item.id, () => {
        this.$message.success('已解除封禁')
        this.init()
      })
    }
  }
}
</script>

<style scoped>
.security-main {
  overflow: auto;
  color: #2c2c2b;
}

.security-page-header,
.security-section__title {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 20px;
}

.security-page-header {
  margin-bottom: 18px;
}

.security-page-header h1,
.security-section h2 {
  margin: 0;
  letter-spacing: 0;
}

.security-page-header h1 {
  font-size: 20px;
  line-height: 28px;
}

.security-page-header p,
.security-section__title p,
.security-notice {
  margin: 5px 0 0;
  color: #7a7d81;
  font-size: 12px;
  line-height: 18px;
}

.security-section {
  padding: 18px 0 20px;
  border-top: 1px solid #e4e6e8;
}

.security-section h2 {
  font-size: 15px;
  line-height: 22px;
}

.security-section h2 i {
  width: 18px;
  margin-right: 7px;
  color: #66727f;
  font-size: 15px;
  text-align: center;
}

.security-policy-row {
  display: flex;
  margin-top: 16px;
  gap: 28px;
  flex-wrap: wrap;
}

.security-policy-field {
  display: flex;
  align-items: center;
  gap: 9px;
  color: #55585c;
  font-size: 13px;
}

.security-policy-field em {
  color: #818489;
  font-style: normal;
}

.security-policy-field /deep/ .el-input-number {
  width: 116px;
}

.security-whitelist-field {
  max-width: 700px;
  margin: 16px 0 0;
}

.security-whitelist-field /deep/ textarea {
  padding: 10px 12px;
  border-radius: 4px;
  font-family: Consolas, "Courier New", monospace;
  font-size: 12px;
  line-height: 20px;
}

.security-notice {
  max-width: 760px;
  padding-left: 10px;
  border-left: 3px solid #e6a23c;
  color: #8a6525;
}

.security-block-table {
  width: 100%;
  margin-top: 14px;
}

@media (max-width: 760px) {
  .security-page-header,
  .security-section__title {
    gap: 12px;
  }

  .security-policy-row {
    display: grid;
    gap: 12px;
  }
}
</style>
