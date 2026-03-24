<template>
  <div class="main-container">
    <div class="card">
      <div class="card-header">
        <span class="card-title">我的权限</span>
      </div>

      <el-descriptions :column="1" border>
        <el-descriptions-item label="姓名">{{ userStore.user?.name }}</el-descriptions-item>
        <el-descriptions-item label="学号">{{ userStore.user?.student_id }}</el-descriptions-item>
        <el-descriptions-item label="部门">
          <el-tag size="small">{{ userStore.user?.department || '未设置' }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="系统角色">
          <el-tag :type="userStore.isAdmin ? 'danger' : 'info'" size="small">
            {{ userStore.isAdmin ? '系统管理员' : userStore.isOfficeAdmin ? '办公室管理员' : '普通用户' }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="部门角色">
          <el-tag :type="userStore.isDeptAdmin ? 'warning' : 'info'" size="small">
            {{ userStore.isDeptAdmin ? '部门管理员' : '普通成员' }}
          </el-tag>
        </el-descriptions-item>
      </el-descriptions>
    </div>

    <div class="card" v-if="tempPermissions.length > 0">
      <div class="card-header">
        <span class="card-title">临时权限</span>
      </div>
      <el-table :data="tempPermissions" class="data-table">
        <el-table-column label="权限" width="200">
          <template #default="{ row }">
            <el-tag size="small">{{ getPermissionText(row) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="授权人" prop="granted_by_name" width="120" />
        <el-table-column label="有效期">
          <template #default="{ row }">
            <span :class="{ 'text-danger': isExpired(row.expires_at) }">
              {{ formatTime(row.expires_at) }}
            </span>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <div class="card">
      <div class="card-header">
        <span class="card-title">权限说明</span>
      </div>
      <div class="permission-list">
        <div class="permission-item">
          <el-tag type="danger">系统管理员</el-tag>
          <span>拥有所有功能的管理权限，包括用户管理、系统配置等</span>
        </div>
        <div class="permission-item">
          <el-tag type="warning">办公室管理员</el-tag>
          <span>可以管理所有部门的排班和用户</span>
        </div>
        <div class="permission-item">
          <el-tag type="warning">部门管理员</el-tag>
          <span>可以管理本部门的排班和用户</span>
        </div>
        <div class="permission-item">
          <el-tag>普通用户</el-tag>
          <span>可以录入无课表、查看自己的值班安排</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { useUserStore } from '../stores/user'

const userStore = useUserStore()

const tempPermissions = computed(() => userStore.tempPermissions || [])

const permissionMap = {
  'schedule:publish': '设置每周分工',
  'schedule:manage:all': '排班管理（全部）',
  'user:manage:all': '用户管理（全部）',
  'schedule:manage:dept': '排班管理（部门）',
  'user:manage:dept': '用户管理（部门）'
}

const getPermissionText = (row) => permissionMap[row.permission] || row.permission

const isExpired = (time) => new Date(time) < new Date()

const formatTime = (time) => {
  if (!time) return '-'
  return new Date(time).toLocaleString('zh-CN')
}
</script>

<style scoped>
.permission-list {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.permission-item {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 12px;
  background: #fafafa;
  border-radius: 8px;
}

.permission-item span {
  color: #666;
  font-size: 14px;
}

.text-danger {
  color: #f56c6c;
}
</style>
