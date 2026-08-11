<template>
  <div class="login-page">
    <main class="login-main">
      <section class="login-panel">
        <header class="login-header">
          <h1>账号登录</h1>
        </header>

        <el-form
          ref="form"
          class="login-form"
          :model="form"
          :rules="formRules"
          label-position="top"
          @submit.native.prevent>
          <el-form-item label="账号" prop="username">
            <el-input
              ref="username"
              v-model.trim="form.username"
              name="username"
              autocomplete="username"
              placeholder="用户名或邮箱"
              @keyup.enter.native="focusPassword">
            </el-input>
          </el-form-item>

          <el-form-item label="密码" prop="password">
            <el-input
              ref="password"
              v-model.trim="form.password"
              name="password"
              type="password"
              autocomplete="current-password"
              placeholder="请输入密码"
              @keyup.enter.native="submit">
            </el-input>
          </el-form-item>

          <el-button
            class="login-submit"
            type="primary"
            icon="el-icon-check"
            native-type="submit"
            @click="submit">
            登录
          </el-button>
        </el-form>
      </section>
    </main>
  </div>
</template>

<script>
import userService from '../../api/user'

export default {
  name: 'login',
  data () {
    return {
      form: {
        username: '',
        password: ''
      },
      formRules: {
        username: [
          {required: true, message: '请输入用户名', trigger: 'blur'}
        ],
        password: [
          {required: true, message: '请输入密码', trigger: 'blur'}
        ]
      }
    }
  },
  mounted () {
    this.$nextTick(() => {
      this.$refs.username.focus()
    })
  },
  methods: {
    focusPassword () {
      if (this.form.username) {
        this.$refs.password.focus()
      }
    },
    submit () {
      this.$refs.form.validate((valid) => {
        if (!valid) {
          return false
        }
        this.login()
      })
    },
    login () {
      userService.login(this.form.username, this.form.password, (data) => {
        this.$store.commit('setUser', {
          token: data.token,
          uid: data.uid,
          username: data.username,
          isAdmin: data.is_admin
        })
        this.$router.push(this.$route.query.redirect || '/')
      })
    }
  }
}
</script>

<style scoped>
.login-page {
  display: flex;
  width: 100%;
  min-height: 100%;
  overflow: hidden;
  background: #ffffff;
  color: #2c2c2b;
}

.login-main {
  display: flex;
  min-width: 0;
  min-height: 100vh;
  padding: 32px;
  align-items: center;
  justify-content: center;
  flex: 1;
  box-sizing: border-box;
  background: #f6f7f8;
}

.login-panel {
  width: 360px;
  max-width: 100%;
  padding: 30px 32px 32px;
  box-sizing: border-box;
  border: 1px solid #e2e3e5;
  border-radius: 8px;
  background: #ffffff;
  box-shadow: 0 8px 24px rgba(31, 32, 34, 0.06);
}

.login-header {
  margin-bottom: 22px;
}

.login-header h1 {
  margin: 0;
  color: #2c2c2b;
  font-size: 20px;
  font-weight: 600;
  letter-spacing: 0;
  line-height: 28px;
}

.login-header::after {
  display: block;
  width: 28px;
  height: 2px;
  margin-top: 10px;
  background: #2783de;
  content: '';
}

.login-form /deep/ .el-form-item {
  margin-bottom: 17px;
}

.login-form /deep/ .el-form-item__label {
  height: 22px;
  padding: 0 0 4px;
  color: #65686c;
  font-size: 12px;
  line-height: 18px;
}

.login-form /deep/ .el-form-item__content {
  line-height: 38px;
}

.login-form /deep/ .el-input__inner {
  height: 38px;
  padding: 0 12px;
  border-color: #dcdfe3;
  border-radius: 6px;
  color: #2c2c2b;
  font-size: 13px;
  line-height: 38px;
}

.login-form /deep/ .el-input__inner:focus {
  border-color: #2783de;
}

.login-form /deep/ .el-form-item__error {
  padding-top: 3px;
  line-height: 14px;
}

.login-submit {
  width: 100%;
  height: 38px;
  margin-top: 3px;
  border-color: #2783de;
  border-radius: 6px;
  background: #2783de;
  font-size: 13px;
}

.login-submit:hover,
.login-submit:focus {
  border-color: #1f73c5;
  background: #1f73c5;
}

@media (max-width: 680px) {
  .login-page {
    display: block;
    min-height: 100vh;
    background: #f6f7f8;
  }

  .login-main {
    min-height: 100vh;
    padding: 20px 16px;
    align-items: flex-start;
  }

  .login-panel {
    margin-top: 34px;
    padding: 26px 24px 28px;
  }
}
</style>
