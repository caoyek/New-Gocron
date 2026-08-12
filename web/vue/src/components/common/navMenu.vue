<template>
  <aside class="app-sidebar" v-cloak>
    <div class="sidebar-brand">
      <img class="sidebar-brand__icon" :src="brandIcon" alt="New-Gocron">
      <span class="sidebar-brand__text">New-Gocron</span>
    </div>

    <nav class="sidebar-nav" aria-label="主导航">
      <section class="sidebar-section">
        <div class="sidebar-section__label">任务</div>
        <el-menu
          class="sidebar-menu"
          :default-active="currentRoute"
          background-color="transparent"
          text-color="#aeb1b5"
          active-text-color="#ffffff"
          router>
          <el-menu-item index="/dashboard">
            <i class="el-icon-menu"></i>
            <span slot="title">数据看板</span>
          </el-menu-item>
          <el-menu-item index="/task">
            <i class="el-icon-date"></i>
            <span slot="title">定时任务</span>
          </el-menu-item>
          <el-menu-item index="/task/log">
            <i class="el-icon-document"></i>
            <span slot="title">执行日志</span>
          </el-menu-item>
          <el-menu-item index="/host">
            <i class="el-icon-location"></i>
            <span slot="title">任务节点</span>
          </el-menu-item>
        </el-menu>
      </section>

      <section v-if="isAdmin" class="sidebar-section">
        <div class="sidebar-section__label">系统</div>
        <el-menu
          class="sidebar-menu"
          :default-active="currentRoute"
          background-color="transparent"
          text-color="#aeb1b5"
          active-text-color="#ffffff"
          router>
          <el-menu-item index="/user">
            <i class="el-icon-service"></i>
            <span slot="title">用户管理</span>
          </el-menu-item>
          <el-menu-item index="/system">
            <i class="el-icon-setting"></i>
            <span slot="title">推送设置</span>
          </el-menu-item>
          <el-menu-item index="/system/login-log">
            <i class="el-icon-tickets"></i>
            <span slot="title">登录日志</span>
          </el-menu-item>
          <el-menu-item index="/system/login-security">
            <i class="el-icon-warning"></i>
            <span slot="title">登录安全</span>
          </el-menu-item>
        </el-menu>
      </section>
    </nav>

    <div class="sidebar-version">v2.0</div>
    <div class="sidebar-account">
      <el-dropdown trigger="click" placement="top-start" @command="handleUserCommand">
        <button class="sidebar-account__button" type="button" aria-label="账号菜单">
          <span class="sidebar-account__avatar">{{accountInitial}}</span>
          <span class="sidebar-account__name">{{username}}</span>
          <i class="sidebar-account__arrow el-icon-arrow-up"></i>
        </button>
        <el-dropdown-menu slot="dropdown">
          <el-dropdown-item command="password" icon="el-icon-edit">修改密码</el-dropdown-item>
          <el-dropdown-item command="logout" icon="el-icon-close">退出</el-dropdown-item>
        </el-dropdown-menu>
      </el-dropdown>
    </div>
  </aside>
</template>

<script>
import brandIcon from '../../assets/brand/new-gocron/new-gocron-icon.svg'

export default {
  name: 'app-nav-menu',
  data () {
    return {
      brandIcon
    }
  },
  computed: {
    isAdmin () {
      return this.$store.getters.user.isAdmin
    },
    username () {
      return this.$store.getters.user.username || '用户'
    },
    accountInitial () {
      return this.username.charAt(0).toUpperCase()
    },
    currentRoute () {
      const path = this.$route.path
      if (path.indexOf('/dashboard') === 0) {
        return '/dashboard'
      }
      if (path === '/' || path === '/task') {
        return '/task'
      }
      if (path.indexOf('/task/log') === 0) {
        return '/task/log'
      }
      if (path.indexOf('/host') === 0) {
        return '/host'
      }
      if (path.indexOf('/system/login-log') === 0) {
        return '/system/login-log'
      }
      if (path.indexOf('/system/login-security') === 0) {
        return '/system/login-security'
      }
      if (path.indexOf('/system') === 0) {
        return '/system'
      }
      if (path.indexOf('/user/edit-my-password') === 0) {
        return ''
      }
      if (path.indexOf('/user') === 0) {
        return '/user'
      }
      return ''
    }
  },
  methods: {
    handleUserCommand (command) {
      if (command === 'password') {
        this.$router.push('/user/edit-my-password')
        return
      }
      if (command === 'logout') {
        this.logout()
      }
    },
    logout () {
      this.$store.commit('logout')
      this.$router.push('/user/login')
    }
  }
}
</script>

<style scoped>
.app-sidebar {
  display: flex;
  width: 188px;
  height: 100vh;
  box-sizing: border-box;
  flex: 0 0 188px;
  flex-direction: column;
  overflow: hidden;
  background: #1f2022;
  color: #ffffff;
}

.sidebar-brand {
  display: flex;
  height: 82px;
  padding: 0;
  align-items: center;
  justify-content: center;
  gap: 10px;
  flex: 0 0 82px;
}

.sidebar-brand__icon {
  width: 38px;
  height: 38px;
  flex: 0 0 38px;
  object-fit: contain;
}

.sidebar-brand__text {
  overflow: hidden;
  color: #f2f3f4;
  font-size: 15px;
  font-weight: 600;
  line-height: 38px;
  white-space: nowrap;
}

.sidebar-nav {
  min-height: 0;
  padding: 12px 0 18px;
  flex: 1;
  overflow-y: auto;
}

.sidebar-section + .sidebar-section {
  margin-top: 24px;
}

.sidebar-section__label {
  height: 28px;
  padding: 0 22px;
  color: #777b80;
  font-size: 12px;
  font-weight: 600;
  line-height: 28px;
}

.sidebar-menu {
  border-right: 0;
}

.sidebar-menu /deep/ .el-menu-item {
  height: 44px;
  margin: 3px 12px;
  padding: 0 14px !important;
  border-radius: 6px;
  font-size: 14px;
  line-height: 44px;
}

.sidebar-menu /deep/ .el-menu-item i {
  width: 22px;
  margin-right: 8px;
  color: inherit;
  text-align: center;
}

.sidebar-menu /deep/ .el-menu-item:hover {
  background: #2b2d30 !important;
  color: #ffffff !important;
}

.sidebar-menu /deep/ .el-menu-item.is-active {
  background: #37393c !important;
  color: #ffffff !important;
  font-weight: 500;
}

.sidebar-account {
  padding: 14px 12px;
  flex: 0 0 auto;
  border-top: 1px solid #35373a;
}

.sidebar-version {
  padding: 0 12px 8px;
  flex: 0 0 auto;
  color: #6f7378;
  font-size: 11px;
  line-height: 18px;
  text-align: center;
}

.sidebar-account /deep/ .el-dropdown {
  display: block;
  width: 100%;
}

.sidebar-account__button {
  display: flex;
  width: 100%;
  height: 48px;
  padding: 0 10px;
  align-items: center;
  gap: 11px;
  border: 0;
  border-radius: 6px;
  outline: 0;
  background: transparent;
  color: #d8dadd;
  cursor: pointer;
  font: inherit;
  text-align: left;
}

.sidebar-account__button:hover,
.sidebar-account__button:focus {
  background: #2b2d30;
}

.sidebar-account__avatar {
  display: inline-flex;
  width: 34px;
  height: 34px;
  align-items: center;
  justify-content: center;
  flex: 0 0 34px;
  border-radius: 50%;
  background: #53657a;
  color: #ffffff;
  font-size: 14px;
}

.sidebar-account__name {
  min-width: 0;
  flex: 1;
  overflow: hidden;
  font-size: 14px;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.sidebar-account__arrow {
  color: #8f9499;
  font-size: 11px;
}

@media (max-width: 840px) {
  .app-sidebar {
    width: 72px;
    flex-basis: 72px;
  }

  .sidebar-brand {
    padding: 0;
    justify-content: center;
  }

  .sidebar-brand__text,
  .sidebar-section__label,
  .sidebar-account__name,
  .sidebar-account__arrow {
    display: none;
  }

  .sidebar-section + .sidebar-section {
    margin-top: 12px;
  }

  .sidebar-menu /deep/ .el-menu-item {
    margin-right: 10px;
    margin-left: 10px;
    padding: 0 !important;
    text-align: center;
  }

  .sidebar-menu /deep/ .el-menu-item i {
    width: 100%;
    margin: 0;
    font-size: 18px;
  }

  .sidebar-account {
    padding-right: 9px;
    padding-left: 9px;
  }

  .sidebar-account__button {
    padding: 0;
    justify-content: center;
  }
}
</style>
