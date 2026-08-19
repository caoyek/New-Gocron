<template>
  <el-dialog
    class="task-edit-dialog"
    :visible="visible"
    width="86%"
    top="5vh"
    :close-on-click-modal="false"
    @close="cancel">
    <span slot="title" class="task-dialog-title">
      {{taskId ? '编辑任务' : '新增任务'}}
      <small v-if="taskId">#{{taskId}}</small>
    </span>
    <div class="task-editor-shell" v-loading="loading">
      <el-form ref="form" class="task-form compact-edit-form" :model="form" :rules="formRules" label-position="top">
        <el-input v-model="form.id" type="hidden"></el-input>
        <div class="task-editor-layout">
          <div class="task-settings-panel">
            <section class="task-form-section">
              <h3 class="section-title"><span>1</span>基本信息</h3>
              <div class="form-grid">
                <el-form-item class="span-7" label="任务名称" prop="name">
                  <el-input v-model.trim="form.name" placeholder="请输入任务名称"></el-input>
                </el-form-item>
                <el-form-item class="span-5" label="标签">
                  <el-select
                    v-model.trim="form.tag"
                    class="task-tag-select"
                    filterable
                    allow-create
                    default-first-option
                    clearable
                    placeholder="选择或输入标签">
                    <el-option v-for="tag in tags" :key="tag" :label="tag" :value="tag"></el-option>
                  </el-select>
                </el-form-item>
                <el-form-item class="span-6" label="任务类型">
                  <div class="field-with-help">
                    <el-radio-group v-model="form.level" size="small">
                      <el-radio-button v-for="item in levelList" :key="item.value" :label="item.value">{{item.label}}</el-radio-button>
                    </el-radio-group>
                    <el-tooltip content="主任务可以配置多个子任务，主任务执行完成后自动执行子任务" placement="top">
                      <button class="field-help-button" type="button" aria-label="任务类型说明"><i class="el-icon-warning"></i></button>
                    </el-tooltip>
                  </div>
                </el-form-item>
                <el-form-item v-if="form.protocol === 2" class="span-6" label="任务节点">
                  <el-select
                    key="shell"
                    v-model="selectedHosts"
                    class="task-host-select"
                    filterable
                    multiple
                    placeholder="请选择任务节点">
                    <el-option v-for="item in availableHosts" :key="item.id" :label="item.alias + ' - ' + item.name" :value="item.id"></el-option>
                  </el-select>
                </el-form-item>
                <el-form-item v-else class="span-6" label="请求方法">
                  <el-radio-group v-model="form.http_method" size="small">
                    <el-radio-button v-for="item in httpMethods" :key="item.value" :label="item.value">{{item.label.toUpperCase()}}</el-radio-button>
                  </el-radio-group>
                </el-form-item>
                <el-form-item v-if="form.level === 1" class="span-6" label="依赖关系">
                  <div class="field-with-help">
                    <el-radio-group v-model="form.dependency_status" size="small">
                      <el-radio-button v-for="item in dependencyStatusList" :key="item.value" :label="item.value">{{item.label}}</el-radio-button>
                    </el-radio-group>
                    <el-tooltip content="强依赖：主任务执行成功才运行子任务；弱依赖：无论主任务是否成功都会运行子任务" placement="top">
                      <button class="field-help-button" type="button" aria-label="依赖关系说明"><i class="el-icon-warning"></i></button>
                    </el-tooltip>
                  </div>
                </el-form-item>
                <el-form-item v-if="form.level === 1" class="span-6" label="子任务">
                  <el-select
                    v-model="selectedDependencyTaskIds"
                    class="dependency-task-select"
                    filterable
                    multiple
                    clearable
                    no-data-text="暂无子任务"
                    placeholder="选择子任务">
                    <el-option
                      v-for="item in availableChildTasks"
                      :key="item.id"
                      :label="'#' + item.id + ' ' + item.name"
                      :value="item.id">
                    </el-option>
                  </el-select>
                </el-form-item>
              </div>
            </section>

            <section class="task-form-section">
              <h3 class="section-title"><span>2</span>调度设置</h3>
              <div class="form-grid">
                <el-form-item v-if="form.level === 1" class="span-12" label="cron表达式" prop="spec">
                  <div class="cron-field">
                    <el-input v-model.trim="form.spec" class="cron-input" placeholder="秒 分 时 天 月 周"></el-input>
                    <div class="cron-shortcuts">
                      <el-tooltip content="每天 00:00 执行" placement="top">
                        <el-button size="small" plain @click="applyCronShortcut('0 0 0 * * *')">每天</el-button>
                      </el-tooltip>
                      <el-tooltip content="每周一 00:00 执行" placement="top">
                        <el-button size="small" plain @click="applyCronShortcut('0 0 0 * * 1')">每周</el-button>
                      </el-tooltip>
                      <el-tooltip content="每月 1 日 00:00 执行" placement="top">
                        <el-button size="small" plain @click="applyCronShortcut('0 0 0 1 * *')">每月</el-button>
                      </el-tooltip>
                    </div>
                  </div>
                </el-form-item>
                <el-form-item class="span-12 switch-form-item" label="单实例运行">
                  <div class="switch-field">
                    <el-switch v-model="form.multi" :active-value="2" :inactive-value="1" active-color="#2783de"></el-switch>
                    <span>单实例运行 · {{form.multi === 2 ? '已开启' : '未开启'}}</span>
                    <el-tooltip content="前次任务未执行完成时，控制下一调度时间是否允许再次执行同一任务" placement="top">
                      <button class="field-help-button" type="button" aria-label="单实例运行说明"><i class="el-icon-warning"></i></button>
                    </el-tooltip>
                  </div>
                </el-form-item>
              </div>
            </section>

            <section class="task-form-section">
              <h3 class="section-title"><span>3</span>异常与重试</h3>
              <div class="form-grid">
                <el-form-item class="span-4" label="超时时间" prop="timeout">
                  <div class="field-with-help">
                    <el-input v-model.number.trim="form.timeout"><template slot="append">秒</template></el-input>
                    <el-tooltip content="任务执行超时后强制结束，取值 0–86400 秒；0 表示不限制，新增 HTTP 任务默认 3600 秒" placement="top">
                      <button class="field-help-button" type="button" aria-label="任务超时时间说明"><i class="el-icon-warning"></i></button>
                    </el-tooltip>
                  </div>
                </el-form-item>
                <el-form-item class="span-4" label="重试次数" prop="retry_times">
                  <el-input v-model.number.trim="form.retry_times" placeholder="0 - 10"><template slot="append">次</template></el-input>
                </el-form-item>
                <el-form-item class="span-4" label="重试间隔" prop="retry_interval">
                  <el-input v-model.number.trim="form.retry_interval" placeholder="0 - 3600"><template slot="append">秒</template></el-input>
                </el-form-item>
              </div>
            </section>

            <section class="task-form-section">
              <h3 class="section-title"><span>4</span>通知与备注</h3>
              <div class="form-grid">
                <el-form-item class="span-12" label="任务通知">
                  <el-radio-group v-model="form.notify_status" size="small" class="wide-segmented-control">
                    <el-radio-button v-for="item in notifyStatusList" :key="item.value" :label="item.value">{{item.label}}</el-radio-button>
                  </el-radio-group>
                </el-form-item>
                <div v-if="form.notify_status !== 1" class="notification-settings-row" :class="{'is-webhook': form.notify_type === 4}">
                  <el-form-item label="通知类型">
                    <el-radio-group v-model="form.notify_type" size="small">
                      <el-radio-button v-for="item in notifyTypes" :key="item.value" :label="item.value">{{item.label}}</el-radio-button>
                    </el-radio-group>
                  </el-form-item>
                  <el-form-item v-if="form.notify_type === 2" class="notification-recipient" label="接收用户">
                    <el-select key="notify-mail" v-model="selectedMailNotifyIds" filterable multiple collapse-tags placeholder="请选择">
                      <el-option v-for="item in mailUsers" :key="item.id" :label="item.username" :value="item.id"></el-option>
                    </el-select>
                  </el-form-item>
                  <el-form-item v-if="form.notify_type === 3" class="notification-recipient" label="发送Channel">
                    <el-select key="notify-slack" v-model="selectedSlackNotifyIds" filterable multiple collapse-tags placeholder="请选择">
                      <el-option v-for="item in slackChannels" :key="item.id" :label="item.name" :value="item.id"></el-option>
                    </el-select>
                  </el-form-item>
                  <el-form-item v-if="form.notify_type === 4" class="notification-recipient notification-recipient--scroll" label="企微群">
                    <el-select key="notify-webhook" v-model="selectedWebhookNotifyIds" filterable multiple placeholder="请选择">
                      <el-option v-for="item in webhookGroups" :key="item.id" :label="item.name" :value="item.id"></el-option>
                    </el-select>
                  </el-form-item>
                  <el-form-item v-if="form.notify_type === 4" class="notification-recipient" label="通知模板">
                    <el-select key="notify-webhook-template" v-model="selectedWebhookTemplateId" filterable placeholder="请选择">
                      <el-option v-for="item in webhookTemplates" :key="item.id" :label="item.name" :value="item.id"></el-option>
                    </el-select>
                  </el-form-item>
                </div>
                <el-form-item v-if="form.notify_status === 4" class="span-12 notification-rule-form-item" label="输出匹配规则">
                  <div class="notification-rule-toolbar">
                    <span>满足</span>
                    <el-radio-group v-model="notificationRuleMode" size="mini">
                      <el-radio-button label="any">任意规则</el-radio-button>
                      <el-radio-button label="all">全部规则</el-radio-button>
                    </el-radio-group>
                    <el-button size="mini" type="primary" plain icon="el-icon-plus" @click="addNotificationRule">添加规则</el-button>
                  </div>
                  <div class="notification-rule-list">
                    <div v-for="(rule, index) in notificationRules" :key="rule.id" class="notification-rule-row">
                      <el-select v-model="rule.type" class="notification-rule-type" size="small" @change="resetNotificationRule(rule)">
                        <el-option v-for="item in notificationRuleTypes" :key="item.value" :label="item.label" :value="item.value"></el-option>
                      </el-select>
                      <template v-if="rule.type === 'number'">
                        <el-input v-model.trim="rule.field" class="notification-rule-field" size="small" placeholder="字段名，如库存数量"></el-input>
                        <el-select v-model="rule.operator" class="notification-rule-operator" size="small">
                          <el-option v-for="operator in notificationNumberOperators" :key="operator" :label="operator" :value="operator"></el-option>
                        </el-select>
                        <el-input v-model.number="rule.number" class="notification-rule-number" size="small" placeholder="比较值"></el-input>
                      </template>
                      <template v-else>
                        <el-input v-model="rule.value" class="notification-rule-value" size="small" :placeholder="notificationRulePlaceholder(rule.type)"></el-input>
                      </template>
                      <el-checkbox v-model="rule.case_sensitive" class="notification-rule-case" size="small">区分大小写</el-checkbox>
                      <el-tooltip content="删除规则" placement="top">
                        <el-button class="notification-rule-delete" size="small" type="text" icon="el-icon-delete" :disabled="notificationRules.length === 1" @click="removeNotificationRule(index)"></el-button>
                      </el-tooltip>
                    </div>
                  </div>
                </el-form-item>
                <el-form-item class="span-12 remark-form-item" label="备注">
                  <el-input v-model="form.remark" type="textarea" :rows="4" placeholder="选填，说明用途与注意事项"></el-input>
                </el-form-item>
              </div>
            </section>
          </div>

          <div class="task-command-panel">
            <div class="command-panel-heading">
              <span class="command-panel-label"><b>*</b> 命令（执行内容）</span>
              <el-radio-group v-model="form.protocol" size="small">
                <el-radio-button v-for="item in protocolList" :key="item.value" :label="item.value">{{item.label.toUpperCase()}}</el-radio-button>
              </el-radio-group>
            </div>
            <el-form-item class="command-form-item" prop="command">
              <div class="command-editor">
                <div class="command-editor__body">
                  <textarea ref="commandEditor" v-model="form.command" class="command-editor__input" :placeholder="commandPlaceholder" spellcheck="false" wrap="soft" @keydown.tab.prevent="insertTab"></textarea>
                </div>
              </div>
            </el-form-item>
          </div>
        </div>
      </el-form>
    </div>
    <span slot="footer" class="compact-dialog-footer">
      <span class="required-note"><b>*</b> 为必填项</span>
      <span class="footer-actions">
        <el-button size="small" @click="cancel">取消</el-button>
        <el-button size="small" type="primary" icon="el-icon-check" @click="submit">保存</el-button>
      </span>
    </span>
  </el-dialog>
</template>

<script>
import taskService from '../../api/task'
import notificationService from '../../api/notification'

function emptyForm () {
  return {
    id: '',
    name: '',
    tag: '',
    level: 1,
    dependency_status: 1,
    dependency_task_id: '',
    spec: '',
    protocol: 2,
    http_method: 1,
    command: '',
    host_id: '',
    timeout: 0,
    multi: 2,
    notify_status: 1,
    notify_type: 2,
    notify_receiver_id: '',
    notify_keyword: '',
    retry_times: 0,
    retry_interval: 0,
    remark: ''
  }
}

let notificationRuleId = 0

function emptyNotificationRule (type) {
  return {
    id: ++notificationRuleId,
    type: type || 'contains',
    value: '',
    case_sensitive: false,
    field: '',
    operator: '>',
    number: null
  }
}

function parseWebhookNotifyTarget (value) {
  try {
    const target = JSON.parse(value)
    if (target && target.format === 'webhook_target' && target.version === 1) {
      return target
    }
  } catch (e) {}

  return null
}

export default {
  name: 'task-edit',
  props: {
    visible: {
      type: Boolean,
      default: false
    },
    taskId: {
      type: [Number, String],
      default: null
    },
    hosts: {
      type: Array,
      default: () => []
    },
    tags: {
      type: Array,
      default: () => []
    }
  },
  data () {
    return {
      form: emptyForm(),
      loading: false,
      formRules: {
        name: [
          {required: true, message: '请输入任务名称', trigger: 'blur'}
        ],
        spec: [
          {required: true, message: '请输入crontab表达式', trigger: 'blur'}
        ],
        command: [
          {required: true, message: '请输入命令', trigger: 'blur'}
        ],
        timeout: [
          {type: 'number', required: true, message: '请输入有效的任务超时时间', trigger: 'blur'}
        ],
        retry_times: [
          {type: 'number', required: true, message: '请输入有效的任务执行失败重试次数', trigger: 'blur'}
        ],
        retry_interval: [
          {type: 'number', required: true, message: '请输入有效的任务执行失败，重试间隔时间', trigger: 'blur'}
        ]
      },
      httpMethods: [
        {
          value: 1,
          label: 'get'
        },
        {
          value: 2,
          label: 'post'
        }
      ],
      protocolList: [
        {
          value: 1,
          label: 'http'
        },
        {
          value: 2,
          label: 'shell'
        }
      ],
      levelList: [
        {
          value: 1,
          label: '主任务'
        },
        {
          value: 2,
          label: '子任务'
        }
      ],
      dependencyStatusList: [
        {
          value: 1,
          label: '强依赖'
        },
        {
          value: 2,
          label: '弱依赖'
        }
      ],
      notifyStatusList: [
        {
          value: 1,
          label: '不通知'
        },
        {
          value: 2,
          label: '失败通知'
        },
        {
          value: 3,
          label: '总是通知'
        },
        {
          value: 4,
          label: '关键字匹配通知'
        }
      ],
      notifyTypes: [
        {
          value: 2,
          label: '邮件'
        },
        {
          value: 3,
          label: 'Slack'
        },
        {
          value: 4,
          label: '企微群推送'
        }
      ],
      notificationRuleTypes: [
        {value: 'contains', label: '包含文本'},
        {value: 'not_contains', label: '不包含文本'},
        {value: 'wildcard', label: '通配符'},
        {value: 'regex', label: '正则表达式'},
        {value: 'number', label: '数值比较'}
      ],
      notificationNumberOperators: ['>', '>=', '<', '<=', '=', '!='],
      notificationRuleMode: 'any',
      notificationRules: [emptyNotificationRule()],
      availableHosts: [],
      availableChildTasks: [],
      mailUsers: [],
      slackChannels: [],
      webhookGroups: [],
      webhookTemplates: [],
      selectedHosts: [],
      selectedDependencyTaskIds: [],
      selectedMailNotifyIds: [],
      selectedSlackNotifyIds: [],
      selectedWebhookNotifyIds: [],
      selectedWebhookTemplateId: null
    }
  },
  computed: {
    commandPlaceholder () {
      if (this.form.protocol === 1) {
        return '请输入URL地址'
      }

      return '请输入shell命令'
    }
  },
  watch: {
    visible (value) {
      if (value) {
        this.open()
      }
    },
    'form.protocol' (value) {
      if (!this.taskId && value === 1 && this.form.timeout === 0) {
        this.form.timeout = 3600
      }
    }
  },
  methods: {
    open () {
      this.reset()
      this.availableHosts = this.hosts || []
      this.loadNotifications()
      this.loadChildTasks()
      if (!this.taskId) {
        return
      }

      this.loading = true
      taskService.detail(this.taskId, (taskData, hosts) => {
        this.loading = false
        if (!taskData) {
          this.$message.error('数据不存在')
          this.cancel()
          return
        }
        this.availableHosts = hosts || this.hosts || []
        this.populate(taskData)
      })
    },
    reset () {
      this.form = emptyForm()
      this.selectedHosts = []
      this.selectedDependencyTaskIds = []
      this.selectedMailNotifyIds = []
      this.selectedSlackNotifyIds = []
      this.selectedWebhookNotifyIds = []
      this.selectedWebhookTemplateId = null
      this.notificationRuleMode = 'any'
      this.notificationRules = [emptyNotificationRule()]
      this.$nextTick(() => {
        if (this.$refs.form) {
          this.$refs.form.clearValidate()
        }
      })
    },
    populate (taskData) {
      this.form.id = taskData.id
      this.form.name = taskData.name
      this.form.tag = taskData.tag
      this.form.level = taskData.level
      if (taskData.dependency_status) {
        this.form.dependency_status = taskData.dependency_status
      }
      this.form.dependency_task_id = taskData.dependency_task_id == null ? '' : String(taskData.dependency_task_id)
      this.selectedDependencyTaskIds = this.form.dependency_task_id
        .split(',')
        .filter((id) => id !== '')
        .map((id) => parseInt(id))
        .filter((id) => id > 0)
      this.form.spec = taskData.spec
      this.form.protocol = taskData.protocol
      if (taskData.http_method) {
        this.form.http_method = taskData.http_method
      }
      this.form.command = taskData.command
      this.form.timeout = taskData.timeout
      this.form.multi = taskData.multi ? 1 : 2
      this.populateNotificationRules(taskData.notify_keyword)
      this.form.notify_status = taskData.notify_status + 1
      this.form.notify_receiver_id = taskData.notify_receiver_id == null ? '' : String(taskData.notify_receiver_id)
      if (taskData.notify_type) {
        this.form.notify_type = taskData.notify_type + 1
      }
      this.form.retry_times = taskData.retry_times
      this.form.retry_interval = taskData.retry_interval
      this.form.remark = taskData.remark
      taskData.hosts = taskData.hosts || []
      if (this.form.protocol === 2) {
        taskData.hosts.forEach((v) => {
          this.selectedHosts.push(v.host_id)
        })
      }

      if (this.form.notify_status > 1) {
        const notifyReceiverIds = this.form.notify_receiver_id.split(',')
        if (this.form.notify_type === 2) {
          notifyReceiverIds.forEach((v) => {
            if (v !== '') {
              this.selectedMailNotifyIds.push(parseInt(v))
            }
          })
        } else if (this.form.notify_type === 3) {
          notifyReceiverIds.forEach((v) => {
            if (v !== '') {
              this.selectedSlackNotifyIds.push(parseInt(v))
            }
          })
        } else if (this.form.notify_type === 4) {
          const target = parseWebhookNotifyTarget(this.form.notify_receiver_id)
          if (target) {
            this.selectedWebhookNotifyIds = (target.group_ids || []).map((id) => parseInt(id)).filter((id) => id > 0)
            this.selectedWebhookTemplateId = parseInt(target.template_id) || this.defaultWebhookTemplateId()
          } else {
            notifyReceiverIds.forEach((v) => {
              if (v !== '') {
                this.selectedWebhookNotifyIds.push(parseInt(v))
              }
            })
            this.selectedWebhookTemplateId = this.defaultWebhookTemplateId()
          }
        }
      }
    },
    loadNotifications () {
      notificationService.mail((data) => {
        this.mailUsers = data.mail_users || []
      })
      notificationService.slack((data) => {
        this.slackChannels = data.channels || []
      })
      notificationService.webhook((data) => {
        this.webhookGroups = data && data.groups ? data.groups : []
        this.webhookTemplates = data && data.templates ? data.templates : []
        this.normalizeWebhookTemplateSelection()
      })
    },
    defaultWebhookTemplateId () {
      return this.webhookTemplates.length > 0 ? this.webhookTemplates[0].id : null
    },
    normalizeWebhookTemplateSelection () {
      if (!this.webhookTemplates.some((template) => template.id === this.selectedWebhookTemplateId)) {
        this.selectedWebhookTemplateId = this.defaultWebhookTemplateId()
      }
    },
    loadChildTasks () {
      taskService.children((tasks) => {
        this.availableChildTasks = tasks || []
      })
    },
    insertTab (event) {
      const editor = event.target
      const start = editor.selectionStart
      const end = editor.selectionEnd
      this.form.command = this.form.command.slice(0, start) + '  ' + this.form.command.slice(end)
      this.$nextTick(() => {
        editor.selectionStart = start + 2
        editor.selectionEnd = start + 2
      })
    },
    applyCronShortcut (spec) {
      this.form.spec = spec
      this.$nextTick(() => {
        if (this.$refs.form) {
          this.$refs.form.clearValidate('spec')
        }
      })
    },
    populateNotificationRules (value) {
      const keyword = value == null ? '' : String(value).trim()
      if (!keyword) {
        this.notificationRuleMode = 'any'
        this.notificationRules = [emptyNotificationRule()]
        return
      }
      try {
        const parsed = JSON.parse(keyword)
        if (parsed.format === 'notification_rules' && parsed.version === 1 && (parsed.mode === 'any' || parsed.mode === 'all') && Array.isArray(parsed.rules) && parsed.rules.length > 0) {
          this.notificationRuleMode = parsed.mode
          this.notificationRules = parsed.rules.map((rule) => Object.assign(emptyNotificationRule(rule.type), rule))
          return
        }
      } catch (e) {
        // Existing plain text values remain compatible as a contains rule.
      }
      const legacyRule = emptyNotificationRule('contains')
      legacyRule.value = keyword
      legacyRule.case_sensitive = true
      this.notificationRuleMode = 'any'
      this.notificationRules = [legacyRule]
    },
    addNotificationRule () {
      if (this.notificationRules.length >= 20) {
        this.$message.error('通知匹配规则不能超过20条')
        return
      }
      this.notificationRules.push(emptyNotificationRule())
    },
    removeNotificationRule (index) {
      if (this.notificationRules.length > 1) {
        this.notificationRules.splice(index, 1)
      }
    },
    resetNotificationRule (rule) {
      rule.value = ''
      rule.field = ''
      rule.operator = '>'
      rule.number = null
      rule.case_sensitive = false
    },
    notificationRulePlaceholder (type) {
      const placeholders = {
        contains: '输出中需要包含的文本',
        not_contains: '输出中不能包含的文本',
        wildcard: '例如：订单*失败，? 匹配单个字符',
        regex: '请输入 Go 正则表达式'
      }
      return placeholders[type] || '请输入匹配内容'
    },
    serializeNotificationRules () {
      const rules = []
      for (let index = 0; index < this.notificationRules.length; index++) {
        const source = this.notificationRules[index]
        if (source.type === 'number') {
          const field = source.field == null ? '' : String(source.field).trim()
          const number = Number(source.number)
          if (!field) {
            this.$message.error(`第${index + 1}条规则请输入数值字段`)
            return false
          }
          if (source.number === '' || source.number == null || !Number.isFinite(number)) {
            this.$message.error(`第${index + 1}条规则请输入有效比较值`)
            return false
          }
          rules.push({
            type: 'number',
            field: field,
            operator: source.operator,
            number: number,
            case_sensitive: source.case_sensitive === true
          })
          continue
        }
        const value = source.value == null ? '' : String(source.value).trim()
        if (!value) {
          this.$message.error(`第${index + 1}条规则请输入匹配内容`)
          return false
        }
        rules.push({
          type: source.type,
          value: value,
          case_sensitive: source.case_sensitive === true
        })
      }
      return JSON.stringify({format: 'notification_rules', version: 1, mode: this.notificationRuleMode, rules: rules})
    },
    submit () {
      this.$refs['form'].validate((valid) => {
        if (!valid) {
          return false
        }
        if (this.form.protocol === 2 && this.selectedHosts.length === 0) {
          this.$message.error('请选择任务节点')
          return false
        }
        if (this.form.notify_status > 1) {
          if (this.form.notify_type === 2 && this.selectedMailNotifyIds.length === 0) {
            this.$message.error('请选择邮件接收用户')
            return false
          }
          if (this.form.notify_type === 3 && this.selectedSlackNotifyIds.length === 0) {
            this.$message.error('请选择Slack Channel')
            return false
          }
          if (this.form.notify_type === 4 && this.selectedWebhookNotifyIds.length === 0) {
            this.$message.error('请选择企微群')
            return false
          }
          if (this.form.notify_type === 4 && !this.selectedWebhookTemplateId) {
            this.$message.error('请选择通知模板')
            return false
          }
        }
        if (this.form.notify_status === 4) {
          const serializedRules = this.serializeNotificationRules()
          if (serializedRules === false) {
            return false
          }
          this.form.notify_keyword = serializedRules
        }

        this.save()
      })
    },
    save () {
      this.form.dependency_task_id = this.form.level === 1 ? this.selectedDependencyTaskIds.join(',') : ''
      if (this.form.protocol === 2 && this.selectedHosts.length > 0) {
        this.form.host_id = this.selectedHosts.join(',')
      }
      if (this.form.notify_status > 1 && this.form.notify_type === 2) {
        this.form.notify_receiver_id = this.selectedMailNotifyIds.join(',')
      }
      if (this.form.notify_status > 1 && this.form.notify_type === 3) {
        this.form.notify_receiver_id = this.selectedSlackNotifyIds.join(',')
      }
      if (this.form.notify_status > 1 && this.form.notify_type === 4) {
        this.form.notify_receiver_id = JSON.stringify({
          format: 'webhook_target',
          version: 1,
          group_ids: this.selectedWebhookNotifyIds,
          template_id: this.selectedWebhookTemplateId
        })
      }
      taskService.update(this.form, () => {
        this.$message.success('保存成功')
        this.$emit('saved')
        this.$emit('update:visible', false)
      })
    },
    cancel () {
      this.$emit('update:visible', false)
    }
  }
}
</script>

<style scoped>
.task-editor-shell,
.task-editor-layout {
  height: auto;
  min-height: 0;
  max-height: calc(94vh - 104px);
  box-sizing: border-box;
}

.compact-edit-form {
  height: auto;
}

.task-dialog-title {
  font-size: 16px;
  font-weight: 600;
}

.task-dialog-title small {
  margin-left: 6px;
  color: #7d7a75;
  font-size: 13px;
  font-weight: 400;
}

.task-editor-layout {
  display: grid;
  grid-template-columns: minmax(0, 0.9fr) minmax(0, 1.1fr);
  position: relative;
  overflow: hidden;
  color: #2c2c2b;
}

.task-editor-layout::after {
  position: absolute;
  top: 0;
  bottom: 0;
  left: 45%;
  width: 1px;
  background: #dcd9d4;
  content: '';
  pointer-events: none;
}

.task-settings-panel {
  min-height: 0;
  box-sizing: border-box;
  overflow-y: auto;
  padding: 8px 20px 10px;
}

.task-command-panel {
  display: flex;
  min-width: 0;
  min-height: 0;
  box-sizing: border-box;
  flex-direction: column;
  overflow: hidden;
  padding: 0 20px 10px;
  background: #ffffff;
}

.task-form-section {
  margin-bottom: 5px;
}

.task-form-section:last-child {
  margin-bottom: 0;
}

.section-title {
  display: flex;
  align-items: center;
  gap: 7px;
  margin: 0 0 3px;
  color: #7d7a75;
  font-size: 12px;
  font-weight: 600;
  letter-spacing: 0;
}

.section-title::after {
  height: 1px;
  flex: 1;
  background: #e6e5e3;
  content: '';
}

.section-title span {
  display: inline-flex;
  width: 16px;
  height: 16px;
  align-items: center;
  justify-content: center;
  border-radius: 4px;
  background: #e5f2fc;
  color: #2783de;
  font-size: 11px;
}

.form-grid {
  display: grid;
  grid-template-columns: repeat(12, minmax(0, 1fr));
  gap: 3px 10px;
}

.span-4 {
  grid-column: span 4;
}

.span-5 {
  grid-column: span 5;
}

.span-6 {
  grid-column: span 6;
}

.span-7 {
  grid-column: span 7;
}

.span-12 {
  grid-column: span 12;
}

.compact-edit-form /deep/ .el-form-item {
  margin-bottom: 0;
}

.compact-edit-form /deep/ .el-form-item__label {
  height: 20px;
  padding: 0 0 2px;
  color: #7d7a75;
  font-size: 12px;
  line-height: 18px;
}

.compact-edit-form /deep/ .el-form-item__content {
  line-height: 30px;
}

.task-settings-panel /deep/ .el-form-item__error {
  position: static;
  padding-top: 3px;
  line-height: 16px;
}

.compact-edit-form /deep/ .el-input__inner {
  height: 30px;
  border-color: #e0dfdc;
  border-radius: 6px;
  color: #2c2c2b;
  line-height: 30px;
}

.compact-edit-form /deep/ .el-input-group__append {
  padding: 0 9px;
  border-color: #e0dfdc;
  background: #f9f8f7;
  color: #7d7a75;
  font-size: 12px;
}

.compact-edit-form /deep/ .el-textarea__inner {
  min-height: 42px !important;
  padding: 5px 10px;
  border-color: #e0dfdc;
  border-radius: 6px;
  font-family: inherit;
  line-height: 18px;
}

.compact-edit-form /deep/ .el-select {
  width: 100%;
}

.task-tag-select /deep/ .el-input,
.task-host-select /deep/ .el-input,
.dependency-task-select /deep/ .el-input {
  width: 100%;
}

.task-host-select /deep/ .el-input__inner,
.dependency-task-select /deep/ .el-input__inner {
  height: 30px !important;
  min-height: 30px !important;
}

.task-host-select /deep/ .el-select__tags,
.dependency-task-select /deep/ .el-select__tags {
  display: flex;
  height: 30px;
  max-width: calc(100% - 32px) !important;
  align-items: center;
  flex-wrap: nowrap;
  overflow: hidden;
  white-space: nowrap;
}

.task-host-select /deep/ .el-tag,
.dependency-task-select /deep/ .el-tag {
  height: 20px;
  margin-top: 0;
  margin-bottom: 0;
  flex: 0 0 auto;
  font-size: 11px;
  line-height: 18px;
}

.task-host-select /deep/ .el-select__input,
.dependency-task-select /deep/ .el-select__input {
  height: 20px;
  margin-left: 5px;
  font-size: 11px;
}

.notification-recipient /deep/ .el-select,
.notification-recipient /deep/ .el-input {
  display: block;
  width: 100%;
}

.notification-settings-row {
  display: grid;
  min-width: 0;
  grid-column: span 12;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 10px;
}

.notification-settings-row.is-webhook {
  grid-template-columns: calc((100% - 10px) / 2) repeat(2, minmax(0, 1fr));
}

.notification-recipient /deep/ .el-select__tags {
  display: flex;
  height: 30px;
  max-width: calc(100% - 32px) !important;
  align-items: center;
  flex-wrap: nowrap;
  overflow: hidden;
  white-space: nowrap;
}

.notification-recipient /deep/ .el-input__inner {
  height: 30px !important;
  min-height: 30px !important;
}

.notification-recipient /deep/ .el-select__tags > span {
  display: flex;
  min-width: 0;
  align-items: center;
  overflow: hidden;
}

.notification-recipient /deep/ .el-tag {
  height: 20px;
  margin-top: 0;
  margin-bottom: 0;
  min-width: 0;
  flex: 0 1 auto;
  line-height: 18px;
}

.notification-recipient /deep/ .el-select__tags-text {
  display: inline-block;
  max-width: 100%;
  overflow: hidden;
  text-overflow: ellipsis;
  vertical-align: top;
  white-space: nowrap;
}

.notification-recipient /deep/ .el-select__input {
  height: 20px;
  margin-left: 5px;
  min-width: 0;
  font-size: 11px;
}

.notification-recipient--scroll /deep/ .el-select__tags {
  overflow-x: auto;
  overflow-y: hidden;
  scrollbar-width: none;
}

.notification-recipient--scroll /deep/ .el-select__tags::-webkit-scrollbar {
  display: none;
}

.notification-recipient--scroll /deep/ .el-select__tags > span {
  flex: 0 0 auto;
  overflow: visible;
}

.notification-recipient--scroll /deep/ .el-tag {
  max-width: none;
  flex: 0 0 auto;
}

.notification-recipient--scroll /deep/ .el-select__tags-text {
  max-width: none;
  overflow: visible;
  text-overflow: clip;
}

.notification-rule-form-item /deep/ .el-form-item__content {
  line-height: normal;
}

.notification-rule-toolbar {
  display: flex;
  min-height: 28px;
  align-items: center;
  gap: 8px;
  margin-bottom: 7px;
  color: #77736e;
  font-size: 12px;
}

.notification-rule-toolbar > span {
  white-space: nowrap;
}

.notification-rule-toolbar > .el-button {
  margin-left: auto;
}

.notification-rule-toolbar > .el-radio-group {
  width: auto;
}

.notification-rule-list {
  display: grid;
  gap: 6px;
}

.notification-rule-row {
  display: grid;
  min-width: 0;
  align-items: center;
  grid-template-columns: 96px minmax(80px, 1fr) 56px minmax(70px, 0.55fr) 88px 24px;
  gap: 5px;
}

.notification-rule-type,
.notification-rule-field,
.notification-rule-value,
.notification-rule-operator,
.notification-rule-number {
  min-width: 0;
}

.notification-rule-value {
  grid-column: 2 / 5;
}

.notification-rule-case {
  margin: 0;
  white-space: nowrap;
}

.notification-rule-case /deep/ .el-checkbox__label {
  padding-left: 4px;
  color: #77736e;
  font-size: 12px;
}

.notification-rule-delete {
  width: 24px;
  box-sizing: border-box;
  padding: 5px 4px;
  color: #d04b44;
}

.remark-form-item /deep/ .el-textarea__inner {
  min-height: 76px !important;
}

.compact-edit-form /deep/ .el-radio-group {
  display: flex;
  width: 100%;
  box-sizing: border-box;
  padding: 2px;
  border-radius: 6px;
  background: #f0efed;
}

.compact-edit-form /deep/ .el-radio-button {
  flex: 1;
}

.compact-edit-form /deep/ .el-radio-button__inner {
  width: 100%;
  height: 26px;
  padding: 5px 7px;
  border: 0;
  border-radius: 5px;
  background: transparent;
  box-shadow: none;
  color: #7d7a75;
  font-size: 12px;
}

.compact-edit-form /deep/ .el-radio-button:first-child .el-radio-button__inner,
.compact-edit-form /deep/ .el-radio-button:last-child .el-radio-button__inner {
  border-radius: 5px;
}

.compact-edit-form /deep/ .el-radio-button__orig-radio:checked + .el-radio-button__inner {
  background: #ffffff;
  color: #2c2c2b;
  font-weight: 500;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.08);
}

.wide-segmented-control /deep/ .el-radio-button__inner {
  padding-right: 4px;
  padding-left: 4px;
}

.cron-field {
  display: flex;
  width: 100%;
  align-items: center;
  gap: 6px;
}

.cron-input {
  min-width: 0;
  flex: 1 1 auto;
}

.cron-input /deep/ .el-input__inner {
  font-family: Consolas, "Cascadia Code", monospace;
}

.cron-shortcuts {
  display: flex;
  flex: 0 0 auto;
  gap: 5px;
}

.cron-shortcuts /deep/ .el-button {
  height: 30px;
  margin-left: 0;
  padding: 7px 10px;
}

.field-with-help {
  display: flex;
  align-items: center;
  gap: 6px;
}

.field-with-help > .el-input,
.field-with-help > .el-select,
.field-with-help > .el-radio-group {
  min-width: 0;
  width: auto;
  flex: 1;
}

.field-help-button {
  width: 22px;
  height: 22px;
  padding: 0;
  flex: 0 0 22px;
  border: 0;
  background: transparent;
  color: #d5803b;
  cursor: help;
  font-size: 16px;
  line-height: 22px;
}

.field-help-button:focus {
  outline: 1px solid #d5803b;
  outline-offset: 1px;
}

.switch-form-item /deep/ .el-form-item__label {
  display: none;
}

.switch-field {
  display: flex;
  height: 30px;
  align-items: center;
  gap: 8px;
  color: #2c2c2b;
  font-size: 13px;
}

.command-panel-heading {
  display: flex;
  min-height: 32px;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 4px;
  flex: 0 0 auto;
}

.command-panel-heading /deep/ .el-radio-group {
  width: 152px;
  flex: 0 0 152px;
}

.command-panel-label {
  color: #7d7a75;
  font-size: 12px;
}

.command-panel-label b,
.required-note b {
  color: #e56458;
}

.command-form-item {
  display: flex;
  height: auto;
  max-height: none;
  min-height: 0;
  flex: 1 1 auto;
  flex-direction: column;
  overflow: hidden;
}

.command-form-item /deep/ .el-form-item__content {
  display: flex;
  min-height: 0;
  flex: 1;
}

.command-editor {
  display: flex;
  min-height: 0;
  flex: 1;
  flex-direction: column;
  margin-bottom: 0;
  overflow: hidden;
  border: 1px solid #343a46;
  border-radius: 7px;
  background: #1f232b;
}

.command-editor:focus-within {
  border-color: #2783de;
}

.command-editor__body {
  min-height: 0;
  flex: 1;
  background: #1f232b;
}

.command-editor__input {
  width: 100%;
  height: 100%;
  min-height: 0;
  box-sizing: border-box;
  resize: none;
  overflow-x: hidden;
  overflow-y: auto;
  padding: 12px;
  border: 0;
  outline: 0;
  background: transparent;
  color: #e6edf3;
  caret-color: #7dcfff;
  font: 13px/20px Consolas, "Cascadia Code", monospace;
  letter-spacing: 0;
  overflow-wrap: anywhere;
  word-break: break-all;
  white-space: pre-wrap;
  tab-size: 2;
}

.command-editor__input::placeholder {
  color: #7f8997;
}

.task-edit-dialog /deep/ .el-dialog {
  display: flex;
  max-width: 1080px;
  max-height: 94vh;
  flex-direction: column;
  overflow: hidden;
  border-radius: 8px;
  box-shadow: 0 8px 28px rgba(0, 0, 0, 0.1);
}

.task-edit-dialog /deep/ .el-dialog__header {
  flex: 0 0 auto;
  padding: 14px 22px;
  border-bottom: 1px solid #e6e5e3;
}

.task-edit-dialog /deep/ .el-dialog__body {
  min-height: 0;
  flex: 1 1 auto;
  overflow: hidden;
  padding: 0;
}

.task-edit-dialog /deep/ .el-dialog__footer {
  position: relative;
  z-index: 2;
  flex: 0 0 auto;
  padding: 11px 22px;
  border-top: 1px solid #e6e5e3;
  background: #ffffff;
}

.compact-dialog-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.required-note {
  color: #7d7a75;
  font-size: 12px;
}

.footer-actions {
  display: flex;
  gap: 8px;
}

.footer-actions .el-button + .el-button {
  margin-left: 0;
}

@media (max-width: 960px) {
  .task-edit-dialog /deep/ .el-dialog {
    width: 96% !important;
  }

  .task-editor-shell,
  .task-editor-layout {
    height: auto;
    min-height: auto;
    max-height: none;
  }

  .task-editor-layout {
    display: block;
    max-height: calc(96vh - 112px);
    overflow-y: auto;
  }

  .task-command-panel {
    min-height: 440px;
    border-top: 1px solid #e6e5e3;
  }

  .task-editor-layout::after {
    display: none;
  }

  .command-editor__input {
    min-height: 360px;
  }
}

@media (max-width: 620px) {
  .span-4,
  .span-5,
  .span-6,
  .span-7,
  .span-12 {
    grid-column: span 12;
  }

  .notification-settings-row,
  .notification-settings-row.is-webhook {
    grid-template-columns: minmax(0, 1fr);
  }

  .compact-dialog-footer {
    justify-content: flex-end;
  }

  .required-note {
    display: none;
  }
}
</style>
