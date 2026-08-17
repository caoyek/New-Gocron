<template>
  <el-container class="app-shell">
    <app-nav-menu v-if="showNavigation"></app-nav-menu>
    <el-main class="app-shell__content">
      <div id="main-container" v-cloak>
        <router-view/>
      </div>
    </el-main>
  </el-container>
</template>

<script>
import installService from './api/install'
import appNavMenu from './components/common/navMenu.vue'

export default {
  name: 'App',
  data () {
    return {}
  },
  computed: {
    showNavigation () {
      const path = this.$route.path
      return this.$store.getters.login && path !== '/install' && path !== '/user/login'
    }
  },
  created () {
    installService.status((data) => {
      if (!data) {
        this.$router.push('/install')
      }
    })
  },
  components: {
    appNavMenu
  }
}
</script>
<style>
  [v-cloak] {
    display: none !important;
  }

  html,
  body,
  #app {
    height: 100%;
  }

  body {
    min-width: 320px;
    margin: 0;
  }

  .app-shell {
    width: 100%;
    height: 100vh;
    margin: 0;
    overflow: hidden;
  }

  .app-shell__content.el-main {
    min-width: 0;
    margin: 0;
    padding: 0;
    overflow: hidden;
    background: #ffffff;
  }

  #main-container {
    height: 100%;
  }

  #main-container > .el-container {
    min-width: 0;
    height: 100%;
  }

  #main-container .el-main {
    height: calc(100vh - 40px);
    margin: 20px;
    padding: 0;
  }

  @media (max-width: 768px) {
    .app-shell__content.el-main {
      padding-top: 56px;
      box-sizing: border-box;
    }

    #main-container .el-main {
      height: calc(100vh - 80px);
      margin: 12px;
    }
  }
</style>
