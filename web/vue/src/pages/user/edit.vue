<template>
  <el-container>
    <el-main class="user-form-main">
      <div class="user-form-shell">
        <div class="user-form-header">
          <h2 class="user-form-title">{{form.id ? '编辑用户' : '新增用户'}}</h2>
        </div>
        <el-form ref="form" class="compact-user-form" :model="form" :rules="formRules" label-position="top">
          <el-input v-model="form.id" type="hidden"></el-input>
        <el-form-item label="用户名" prop="name">
          <el-input v-model="form.name"></el-input>
        </el-form-item>
        <el-form-item label="邮箱" prop="email">
          <el-input v-model="form.email"></el-input>
        </el-form-item>
        <template v-if="!form.id">
          <el-form-item label="密码" prop="password">
            <el-input v-model="form.password" type="password"></el-input>
          </el-form-item>
          <el-form-item label="确认密码" prop="confirm_password">
            <el-input v-model="form.confirm_password" type="password"></el-input>
          </el-form-item>
        </template>
        <el-form-item label="角色" prop="is_admin">
          <el-radio-group v-model="form.is_admin" size="small">
            <el-radio-button :label="0">普通用户</el-radio-button>
            <el-radio-button :label="1">管理员</el-radio-button>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="状态" prop="status">
          <el-radio-group v-model="form.status" size="small">
            <el-radio-button :label="1">启用</el-radio-button>
            <el-radio-button :label="0">禁用</el-radio-button>
          </el-radio-group>
        </el-form-item>
          <div class="user-form-actions">
            <el-button size="small" @click="cancel">取消</el-button>
            <el-button size="small" type="primary" icon="el-icon-check" @click="submit()">保存</el-button>
          </div>
        </el-form>
      </div>
    </el-main>
  </el-container>
</template>

<script>
import userService from '../../api/user'
import './form.css'
export default {
  name: 'user-edit',
  data: function () {
    return {
      form: {
        id: '',
        name: '',
        email: '',
        is_admin: 0,
        password: '',
        confirm_password: '',
        status: 1
      },
      formRules: {
        name: [
          {required: true, message: '请输入用户名', trigger: 'blur'}
        ],
        email: [
          {type: 'email', required: true, message: '请输入有效邮箱地址', trigger: 'blur'}
        ],
        password: [
          {required: true, message: '请输入密码', trigger: 'blur'}
        ],
        confirm_password: [
          {required: true, message: '请再次输入密码', trigger: 'blur'}
        ]
      }
    }
  },
  created () {
    const id = this.$route.params.id
    if (!id) {
      return
    }
    userService.detail(id, (data) => {
      if (!data) {
        this.$message.error('数据不存在')
        return
      }
      this.form.id = data.id
      this.form.name = data.name
      this.form.email = data.email
      this.form.is_admin = data.is_admin
      this.form.status = data.status
    })
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
      userService.update(this.form, () => {
        this.$router.push('/user')
      })
    },
    cancel () {
      this.$router.push('/user')
    }
  }
}
</script>
