<template>
  <el-container>
    <el-main class="user-form-main">
      <div class="user-form-shell">
        <div class="user-form-header">
          <h2 class="user-form-title">修改用户密码</h2>
        </div>
        <el-form ref="form" class="compact-user-form" :model="form" :rules="formRules" label-position="top">
        <el-form-item label="新密码" prop="new_password">
          <el-input v-model="form.new_password" type="password"></el-input>
        </el-form-item>
        <el-form-item label="确认新密码" prop="confirm_new_password">
          <el-input v-model="form.confirm_new_password" type="password"></el-input>
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
  name: 'user-edit-password',
  data: function () {
    return {
      form: {
        id: '',
        new_password: '',
        confirm_new_password: ''
      },
      formRules: {
        new_password: [
          {required: true, message: '请输入新密码', trigger: 'blur'}
        ],
        confirm_new_password: [
          {required: true, message: '请再次输入新密码', trigger: 'blur'}
        ]
      }
    }
  },
  created () {
    const id = this.$route.params.id
    if (!id) {
      return
    }
    this.form.id = id
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
      userService.editPassword(this.form, () => {
        this.$router.push('/user')
      })
    },
    cancel () {
      this.$router.push('/user')
    }
  }
}
</script>
