<template>
  <el-dialog
    class="host-edit-dialog"
    :title="dialogTitle"
    :visible="visible"
    width="520px"
    top="8vh"
    :close-on-click-modal="false"
    @close="cancel">
    <div class="compact-dialog-body" v-loading="loading">
      <el-form ref="form" class="compact-edit-form" :model="form" :rules="formRules" label-width="90px">
        <el-input v-model="form.id" type="hidden"></el-input>
        <el-form-item label="节点名称" prop="alias">
          <el-input v-model="form.alias"></el-input>
        </el-form-item>
        <el-form-item label="主机名" prop="name">
          <el-input v-model="form.name"></el-input>
        </el-form-item>
        <el-form-item label="端口" prop="port">
          <el-input v-model.number="form.port"></el-input>
        </el-form-item>
        <el-form-item label="备注">
          <el-input
            v-model="form.remark"
            type="textarea"
            :rows="3">
          </el-input>
        </el-form-item>
      </el-form>
    </div>
    <span slot="footer" class="compact-dialog-footer">
      <el-button size="small" @click="cancel">取消</el-button>
      <el-button size="small" type="primary" icon="el-icon-check" @click="submit">保存</el-button>
    </span>
  </el-dialog>
</template>

<script>
import hostService from '../../api/host'

function emptyForm () {
  return {
    id: '',
    name: '',
    port: 5921,
    alias: '',
    remark: ''
  }
}

export default {
  name: 'host-edit',
  props: {
    visible: {
      type: Boolean,
      default: false
    },
    hostId: {
      type: [Number, String],
      default: null
    }
  },
  data () {
    return {
      form: emptyForm(),
      loading: false,
      formRules: {
        name: [
          {required: true, message: '请输入主机名', trigger: 'blur'}
        ],
        port: [
          {required: true, message: '请输入端口', trigger: 'blur'},
          {type: 'number', message: '端口无效'}
        ],
        alias: [
          {required: true, message: '请输入节点名称', trigger: 'blur'}
        ]
      }
    }
  },
  computed: {
    dialogTitle () {
      return this.hostId ? `编辑节点 #${this.hostId}` : '新增节点'
    }
  },
  watch: {
    visible (value) {
      if (value) {
        this.open()
      }
    }
  },
  methods: {
    open () {
      this.reset()
      if (!this.hostId) {
        return
      }

      this.loading = true
      hostService.detail(this.hostId, (data) => {
        this.loading = false
        if (!data) {
          this.$message.error('数据不存在')
          this.cancel()
          return
        }
        this.form.id = data.id
        this.form.name = data.name
        this.form.port = data.port
        this.form.alias = data.alias
        this.form.remark = data.remark
      })
    },
    reset () {
      this.form = emptyForm()
      this.$nextTick(() => {
        if (this.$refs.form) {
          this.$refs.form.clearValidate()
        }
      })
    },
    submit () {
      this.$refs.form.validate((valid) => {
        if (!valid) {
          return false
        }
        this.save()
      })
    },
    save () {
      hostService.update(this.form, () => {
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
.host-edit-dialog /deep/ .el-dialog {
  max-width: 94%;
}

.host-edit-dialog /deep/ .el-dialog__header {
  padding: 16px 20px 12px;
  border-bottom: 1px solid #ebeef5;
}

.host-edit-dialog /deep/ .el-dialog__body {
  padding: 14px 20px 2px;
}

.host-edit-dialog /deep/ .el-dialog__footer {
  padding: 10px 20px 16px;
  border-top: 1px solid #ebeef5;
}

.compact-edit-form /deep/ .el-form-item {
  margin-bottom: 12px;
}

.compact-dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
}

.compact-dialog-footer .el-button + .el-button {
  margin-left: 0;
}
</style>
