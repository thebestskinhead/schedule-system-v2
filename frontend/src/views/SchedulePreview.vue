<template>
  <div class="schedule-preview-page">
    <el-container>
      <el-header class="header">
        <div class="header-content">
          <h1 class="logo">排班系统 - 排班预览</h1>
          <div class="nav-menu">
            <el-button @click="router.push('/schedule')">返回排班</el-button>
            <el-button @click="router.push('/')">返回首页</el-button>
          </div>
        </div>
      </el-header>

      <el-main class="main-content">
        <div class="page-container">
          <el-card>
            <template #header>
              <div class="card-header">
                <span>排班预览结果</span>
              </div>
            </template>
            <p>请前往排班管理页面生成预览</p>
          </el-card>
        </div>
      </el-main>
    </el-container>
  </div>
</template>

<script setup>
import { useRouter } from 'vue-router'

const router = useRouter()
</script>

<style scoped>
.header {
  background: #fff;
  box-shadow: 0 2px 4px rgba(0,0,0,0.1);
  padding: 0;
}

.header-content {
  max-width: 1200px;
  margin: 0 auto;
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 60px;
  padding: 0 20px;
}

.logo {
  margin: 0;
  font-size: 20px;
  color: #409eff;
}

.main-content {
  min-height: calc(100vh - 60px);
  padding: 20px;
}
</style>
