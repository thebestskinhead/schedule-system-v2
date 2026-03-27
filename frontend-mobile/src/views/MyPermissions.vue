<template>
  <div class="main-container">
    <Navbar title="我的权限" show-back />
    <!-- 用户信息 -->
    <div class="section-card">
      <div class="section-title">
        <span>我的权限</span>
      </div>
      <van-cell-group inset>
        <van-cell title="姓名" :value="userStore.user?.name || '-'" />
        <van-cell title="学号" :value="userStore.user?.student_id || '-'" />
        <van-cell title="部门">
          <template #value>
            <van-tag type="primary" size="medium">
              {{ userStore.user?.department || '未设置' }}
            </van-tag>
          </template>
        </van-cell>
        <van-cell title="系统角色">
          <template #value>
            <van-tag
              :type="userStore.isAdmin ? 'danger' : 'default'"
              size="medium"
            >
              {{
                userStore.isAdmin
                  ? '系统管理员'
                  : userStore.isOfficeAdmin
                    ? '办公室管理员'
                    : '普通用户'
              }}
            </van-tag>
          </template>
        </van-cell>
        <van-cell title="部门角色">
          <template #value>
            <van-tag
              :type="userStore.isDeptAdmin ? 'warning' : 'default'"
              size="medium"
            >
              {{ userStore.isDeptAdmin ? '部门管理员' : '普通成员' }}
            </van-tag>
          </template>
        </van-cell>
      </van-cell-group>
    </div>

    <!-- 临时权限 -->
    <div class="section-card mt-2" v-if="tempPermissions.length > 0">
      <div class="section-title">
        <span>临时权限</span>
        <van-tag type="primary" size="small" round>
          {{ tempPermissions.length }} 项
        </van-tag>
      </div>
      <div class="temp-perm-list">
        <div
          v-for="(item, index) in tempPermissions"
          :key="index"
          class="temp-perm-card"
        >
          <div class="perm-card-header">
            <van-tag size="medium">{{ getPermissionText(item) }}</van-tag>
          </div>
          <van-cell-group>
            <van-cell title="授权人" :value="item.granted_by_name || '-'" />
            <van-cell title="有效期">
              <template #value>
                <span :class="{ 'text-danger': isExpired(item.expires_at) }">
                  {{ formatTime(item.expires_at) }}
                </span>
              </template>
            </van-cell>
          </van-cell-group>
        </div>
      </div>
    </div>

    <!-- 权限说明 -->
    <div class="section-card mt-2">
      <div class="section-title">
        <span>权限说明</span>
      </div>
      <div class="permission-list">
        <div class="permission-item">
          <van-tag type="danger" size="medium">系统管理员</van-tag>
          <span class="perm-desc">拥有所有功能的管理权限，包括用户管理、系统配置等</span>
        </div>
        <van-divider :style="{ margin: '4px 0' }" />
        <div class="permission-item">
          <van-tag type="warning" size="medium">办公室管理员</van-tag>
          <span class="perm-desc">可以管理所有部门的排班和用户</span>
        </div>
        <van-divider :style="{ margin: '4px 0' }" />
        <div class="permission-item">
          <van-tag type="warning" size="medium">部门管理员</van-tag>
          <span class="perm-desc">可以管理本部门的排班和用户</span>
        </div>
        <van-divider :style="{ margin: '4px 0' }" />
        <div class="permission-item">
          <van-tag type="primary" size="medium">普通用户</van-tag>
          <span class="perm-desc">可以录入无课表、查看自己的值班安排</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { useUserStore } from '../stores/user'
import Navbar from '../components/Navbar.vue'

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
.main-container {
  min-height: 100%;
  background: var(--color-bg, #F7F8FA);
  padding: calc(12px + env(safe-area-inset-top)) 0 calc(var(--safe-area-bottom, env(safe-area-inset-bottom)) + 60px);
}

.section-card {
  margin: 0 12px;
  background: #fff;
  border-radius: var(--card-border-radius, 12px);
  overflow: hidden;
}

.section-title {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px;
  font-size: var(--font-size-h3, 16px);
  font-weight: var(--font-weight-semibold, 600);
  color: var(--color-text-primary, #323233);
  border-bottom: 1px solid var(--color-border, #EBEDF0);
}

.mt-2 {
  margin-top: 8px;
}

/* 临时权限 */
.temp-perm-list {
  padding: 12px;
}

.temp-perm-card {
  background: var(--color-bg, #F7F8FA);
  border-radius: var(--radius-md, 8px);
  margin-bottom: 12px;
  overflow: hidden;
}

.temp-perm-card:last-child {
  margin-bottom: 0;
}

.perm-card-header {
  padding: 12px 16px 4px;
}

/* 权限说明 */
.permission-list {
  padding: 16px;
}

.permission-item {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  padding: 12px 0;
}

.perm-desc {
  color: var(--color-text-secondary, #646566);
  font-size: var(--font-size-body, 14px);
  line-height: 1.6;
  flex: 1;
}

.text-danger {
  color: var(--color-danger, #EE0A24);
  font-weight: var(--font-weight-medium, 500);
}
</style>
