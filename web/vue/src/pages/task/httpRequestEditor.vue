<template>
  <div class="http-editor">
    <el-form-item prop="command" class="request-address">
      <div class="request-address-row">
        <el-select v-model="form.http_method" class="request-method" :class="{'is-post': form.http_method === 2}" size="small">
          <el-option label="GET" :value="1"></el-option>
          <el-option label="POST" :value="2"></el-option>
        </el-select>
        <el-input v-model="form.command" size="small" placeholder="https://example.com/api" @input="readParams"></el-input>
      </div>
    </el-form-item>
    <el-tabs v-model="activeTab" class="request-tabs">
      <el-tab-pane name="params">
        <span slot="label">Params <small v-if="paramCount">{{paramCount}}</small></span>
        <div class="request-section-heading">{{isFormPost ? '表单参数' : 'Query 参数'}}</div>
        <div class="request-grid request-grid-heading"><span>参数名</span><span>参数值</span><span></span></div>
        <div v-for="(row, index) in params" :key="row.id" class="request-grid">
          <el-input v-model="row.name" size="small" placeholder="参数名" @input="writeParams"></el-input>
          <el-input v-model="row.value" size="small" placeholder="参数值" @input="writeParams"></el-input>
          <el-button type="text" icon="el-icon-delete" aria-label="删除参数" @click="removeParam(index)"></el-button>
        </div>
        <el-button class="request-add" type="text" icon="el-icon-plus" @click="addRow(params)">添加参数</el-button>
        <p v-if="isFormPost" class="request-hint">兼容原有 POST：URL 问号后的参数作为表单 Body 提交。</p>
      </el-tab-pane>
      <el-tab-pane label="Body" name="body">
        <div v-if="form.http_method === 1" class="request-empty"><i class="el-icon-document"></i><span>GET 请求不发送 Body</span><small>请在参数或 Headers 中配置请求内容</small></div>
        <template v-else>
          <div class="body-toolbar">
            <span class="body-type">JSON <small>选填 · 留空不发送 Body</small></span>
            <el-button type="text" size="small" @click="formatJSON">格式化</el-button>
          </div>
          <div class="json-caption">application/json</div>
          <textarea v-model="form.http_body" class="json-editor" placeholder='{
  "key": "value"
}' spellcheck="false" aria-label="JSON 请求体" @input="form.http_body_mode = 'json'"></textarea>
          <p v-if="legacyForm && form.http_body_mode !== 'json'" class="request-hint">此旧任务保留原有表单参数提交方式；填写 JSON 后改用 JSON Body。</p>
        </template>
      </el-tab-pane>
      <el-tab-pane name="headers">
        <span slot="label">Headers <small v-if="headerCount">{{headerCount}}</small></span>
        <div class="request-section-heading">请求头</div>
        <el-alert v-if="headerError" :title="headerError" type="error" :closable="false"></el-alert>
        <div class="request-grid request-grid-heading"><span>Header 名称</span><span>Header 值</span><span></span></div>
        <div v-for="(row, index) in headers" :key="row.id" class="request-grid">
          <el-input v-model="row.name" size="small" placeholder="如 Authorization"></el-input>
          <el-input v-model="row.value" size="small" placeholder="如 Bearer token"></el-input>
          <el-button type="text" icon="el-icon-delete" aria-label="删除请求头" @click="headers.splice(index, 1)"></el-button>
        </div>
        <el-button class="request-add" type="text" icon="el-icon-plus" @click="addRow(headers)">添加请求头</el-button>
        <p class="request-hint">JSON Body 默认使用 application/json，无需重复填写 Content-Type。</p>
      </el-tab-pane>
    </el-tabs>
  </div>
</template>

<script>
let rowId = 0
function row (name, value) {
  return {id: ++rowId, name: name || '', value: value || ''}
}

export default {
  name: 'http-request-editor',
  props: {
    form: {type: Object, required: true},
    legacyForm: {type: Boolean, default: false}
  },
  data () {
    return {activeTab: 'params', params: [], headers: [], headerError: ''}
  },
  computed: {
    isFormPost () { return this.form.http_method === 2 && this.form.http_body_mode !== 'json' },
    paramCount () { return this.params.filter(item => item.name !== '').length },
    headerCount () { return this.headers.filter(item => item.name !== '').length }
  },
  watch: {
    'form.http_method' (value) {
      if (value === 2 && !this.legacyForm) this.form.http_body_mode = 'json'
    },
    'form.command': 'readParams',
    'form.http_headers': 'readHeaders'
  },
  mounted () {
    if (this.form.http_method === 2 && !this.legacyForm) this.form.http_body_mode = 'json'
    this.readParams()
    this.readHeaders()
  },
  methods: {
    addRow (rows) { rows.push(row()) },
    readParams () {
      const command = this.form.command || ''
      const address = command.split('#')[0]
      const question = address.indexOf('?')
      const parsed = []
      if (question >= 0) {
        new URLSearchParams(address.slice(question + 1)).forEach((value, name) => parsed.push(row(name, value)))
      }
      // Do not rebuild rows on our own changes: preserve focus and incomplete input.
      if (this.lastWrittenCommand === command) return
      this.params = parsed.length ? parsed : [row()]
    },
    writeParams () {
      const command = this.form.command || ''
      const hashAt = command.indexOf('#')
      const hash = hashAt >= 0 ? command.slice(hashAt) : ''
      const base = (hashAt >= 0 ? command.slice(0, hashAt) : command).split('?')[0]
      const query = new URLSearchParams()
      this.params.forEach(item => { if (item.name !== '') query.append(item.name, item.value) })
      const encoded = query.toString()
      this.lastWrittenCommand = base + (encoded ? '?' + encoded : '') + hash
      this.form.command = this.lastWrittenCommand
    },
    removeParam (index) {
      this.params.splice(index, 1)
      this.writeParams()
    },
    readHeaders () {
      this.headerError = ''
      try {
        const data = this.form.http_headers ? JSON.parse(this.form.http_headers) : {}
        if (!data || Array.isArray(data) || typeof data !== 'object' || Object.keys(data).some(name => typeof data[name] !== 'string')) throw new Error('invalid')
        this.headers = Object.keys(data).map(name => row(name, data[name]))
        if (!this.headers.length) this.headers.push(row())
      } catch (e) {
        this.headerError = '原请求头配置格式无效，请先修正原配置。'
      }
    },
    formatJSON () {
      if (!this.form.http_body.trim()) return
      try {
        this.form.http_body = JSON.stringify(JSON.parse(this.form.http_body), null, 2)
        this.form.http_body_mode = 'json'
      } catch (e) {
        this.$message.error('JSON 格式有误，请检查后再格式化')
      }
    },
    validate () {
      if (this.headerError) {
        this.activeTab = 'headers'
        this.$message.error(this.headerError)
        return false
      }
      const headers = Object.create(null)
      const seen = new Set()
      for (const item of this.headers) {
        if (!item.name && !item.value) continue
        const name = item.name.trim()
        if (!name || seen.has(name.toLowerCase())) {
          this.activeTab = 'headers'
          this.$message.error('请填写请求头名称，同名请求头不能重复')
          return false
        }
        seen.add(name.toLowerCase())
        headers[name] = item.value
      }
      if (this.params.some(item => !item.name && item.value)) {
        this.activeTab = 'params'
        this.$message.error('请填写参数名称')
        return false
      }
      if (this.form.http_method === 2 && this.form.http_body.trim()) {
        try { JSON.parse(this.form.http_body) } catch (e) {
          this.activeTab = 'body'
          this.$message.error('请填写有效的 JSON Body')
          return false
        }
        this.form.http_body_mode = 'json'
      }
      if (this.form.http_method === 2 && !this.legacyForm) this.form.http_body_mode = 'json'
      this.form.http_headers = Object.keys(headers).length ? JSON.stringify(headers) : ''
      return true
    }
  }
}
</script>

<style scoped>
.http-editor { min-width: 0; padding-top: 4px; }
.http-editor .request-address { margin-bottom: 24px; }
.request-address /deep/ .el-form-item__error { position: absolute; padding-top: 6px; line-height: 18px; }
.request-address-row { display: flex; border: 1px solid #dcdfe6; border-radius: 6px; overflow: hidden; background: #fff; }
.request-address-row /deep/ .el-input__inner { border: 0; height: 36px; line-height: 36px; }
.request-method { width: 100px; flex: 0 0 100px; border-right: 1px solid #e6e9ef; }
.request-method /deep/ .el-input__inner { color: #26996b; font-weight: 700; background: #f3faf6; }
.request-method.is-post /deep/ .el-input__inner { color: #2783de; background: #f0f7ff; }
.request-tabs /deep/ .el-tabs__active-bar { background: #2783de; }
.request-tabs /deep/ .el-tabs__item { font-size: 13px; }
.request-tabs /deep/ .el-tabs__nav-wrap::after { height: 1px; background: #e6e9ef; }
.request-tabs small { display: inline-block; min-width: 16px; padding: 0 3px; border-radius: 4px; background: #edf4fb; color: #2783de; text-align: center; line-height: 18px; }
.request-section-heading { display: flex; flex-wrap: wrap; gap: 8px; align-items: baseline; margin: 8px 0 12px; color: #555e6b; font-size: 13px; font-weight: 500; }
.request-section-heading span, .request-hint { color: #9299a3; font-size: 12px; font-weight: 400; }
.request-grid { display: grid; grid-template-columns: minmax(0, 1fr) minmax(0, 1.4fr) 28px; gap: 8px; align-items: center; margin-bottom: 8px; }
.request-grid-heading { padding: 8px 0; border-bottom: 1px solid #ebeef5; color: #909399; font-size: 12px; }
.request-grid /deep/ .el-input__inner { height: 32px; line-height: 32px; }
.request-grid .el-button { padding: 7px 0; color: #a4abb5; }
.request-grid .el-button:hover { color: #f56c6c; }
.request-add { padding: 7px 0; font-size: 12px; }
.request-hint { margin: 12px 0; line-height: 20px; }
.request-empty { display: flex; flex-direction: column; gap: 12px; align-items: center; padding: 65px 12px; color: #909399; font-size: 13px; }
.request-empty i { font-size: 28px; color: #bcc6d3; }
.request-empty small { color: #a9afb7; }
.body-toolbar { display: flex; align-items: center; justify-content: space-between; margin: 10px 0 14px; }
.body-type { color: #606266; font-size: 13px; }
.body-type small { margin-left: 8px; color: #9299a3; font-size: 12px; }
.json-caption { padding: 8px 12px; border: 1px solid #343a46; border-bottom: 0; border-radius: 6px 6px 0 0; background: #282e38; color: #8e9aaa; font: 12px Consolas, monospace; }
.json-editor { display: block; width: 100%; min-height: 320px; box-sizing: border-box; padding: 14px; border: 1px solid #343a46; border-radius: 0 0 6px 6px; outline: none; resize: vertical; background: #1f232b; color: #e6edf3; caret-color: #7dcfff; font: 13px/22px Consolas, monospace; }
.json-editor:focus { border-color: #2783de; }
.json-editor::placeholder { color: #7f8997; }
@media (max-width: 600px) {
  .request-method { width: 85px; flex-basis: 85px; }
  .request-grid { gap: 5px; grid-template-columns: minmax(0, 1fr) minmax(0, 1fr) 24px; }
  .json-editor { min-height: 200px; }
}
</style>
