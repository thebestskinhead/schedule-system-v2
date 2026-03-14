<template>
  <router-view />
</template>

<style>
* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

body {
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
  background-color: #f0f2f5;
  color: #333;
  line-height: 1.6;
}

/* 全局通用样式 */
.page-wrapper {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
}

.main-container {
  flex: 1;
  max-width: 1200px;
  width: 100%;
  margin: 0 auto;
  padding: 24px;
}

/* 卡片样式统一 */
.card {
  background: #fff;
  border-radius: 12px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05);
  padding: 24px;
  margin-bottom: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
  padding-bottom: 16px;
  border-bottom: 1px solid #f0f0f0;
}

.card-title {
  font-size: 18px;
  font-weight: 600;
  color: #1a1a1a;
}

/* 按钮组 */
.action-bar {
  display: flex;
  gap: 12px;
  flex-wrap: wrap;
}

/* 表格样式优化 */
.data-table {
  width: 100%;
}

.data-table :deep(.el-table__header) {
  background-color: #fafafa;
}

.data-table :deep(th.el-table__cell) {
  font-weight: 600;
  color: #333;
  background-color: #fafafa;
}

/* 空状态 */
.empty-state {
  padding: 60px 0;
  text-align: center;
}

/* 响应式 */
@media (max-width: 768px) {
  .main-container {
    padding: 16px;
  }
  
  .card {
    padding: 16px;
  }
}
</style>
