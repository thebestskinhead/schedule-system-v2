<template>
  <div class="my-permissions-page">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>我的权限</span>
        </div>
      </template>

      <!-- 角色概览 -->
      <div class="role-overview">
        <h3>角色概览</h3>
        <div class="role-badges">
          <el-tag v-if="userStore.isAdmin" size="large" type="danger" effect="dark">
            <el-icon><Star /></el-icon>
            系统管理员
          </el-tag>
          <el-tag v-else-if="userStore.isOfficeAdmin" size="large" type="warning" effect="dark">
            <el-icon><OfficeBuilding /></el-icon>
            办公室管理员
          </el-tag>
          <el-tag v-else-if="userStore.isDeptAdmin" size="large" type="primary" effect="dark">
            <el-icon><UserFilled /></el-icon>
            部门管理员
          </el-tag>
          <el-tag v-else size="large" type="info">
            <el-icon><User /></el-icon>
            部门成员
          </el-tag>
          <el-tag size="large" style="margin-left: 10px;">
            {{ userStore.user?.department || '未分配部门' }}
          </el-tag>
        </div>
        <p class="role-desc">{{ currentRoleDesc }}</p>
      </div>

      <!-- 权限矩阵 -->
      <div class="permissions-section">
        <h3>我的权限矩阵</h3>

        <!-- 个人权限 - 所有人都有 -->
        <div class="perm-category">
          <div class="category-header">
            <el-icon :size="20" color="#67c23a"><User /></el-icon>
            <span>个人权限</span>
            <el-tag size="small" type="success">自带</el-tag>
          </div>
          <div class="perm-list">
            <div class="perm-item" v-for="perm in personalPerms" :key="perm.code">
              <el-icon color="#67c23a"><CircleCheck /></el-icon>
              <span>{{ perm.name }}</span>
              <el-tooltip :content="perm.desc" placement="top">
                <el-icon class="info-icon"><InfoFilled /></el-icon>
              </el-tooltip>
            </div>
          </div>
        </div>

        <!-- 部门权限 - 部门管理员及以上 -->
        <div class="perm-category" v-if="hasDeptPerms">
          <div class="category-header">
            <el-icon :size="20" color="#409eff"><OfficeBuilding /></el-icon>
            <span>部门管理权限</span>
            <el-tag size="small" type="primary">角色/临时</el-tag>
          </div>
          <div class="perm-list">
            <div class="perm-item" v-for="perm in deptPerms" :key="perm.code" :class="{ 'has-perm': perm.has }">
              <el-icon :color="perm.has ? '#409eff' : '#c0c4cc'">
                <CircleCheck v-if="perm.has" />
                <CircleClose v-else />
              </el-icon>
              <span :class="{ 'disabled': !perm.has }">{{ perm.name }}</span>
              <el-tooltip :content="perm.desc" placement="top">
                <el-icon class="info-icon"><InfoFilled /></el-icon>
              </el-tooltip>
            </div>
          </div>
        </div>

        <!-- 全局权限 - 办公室管理员及以上 -->
        <div class="perm-category" v-if="hasGlobalPerms">
          <div class="category-header">
            <el-icon :size="20" color="#e6a23c"><Star /></el-icon>
            <span>全局管理权限</span>
            <el-tag size="small" type="warning">角色/临时</el-tag>
          </div>
          <div class="perm-list">
            <div class="perm-item" v-for="perm in globalPerms" :key="perm.code" :class="{ 'has-perm': perm.has }">
              <el-icon :color="perm.has ? '#e6a23c' : '#c0c4cc'">
                <CircleCheck v-if="perm.has" />
                <CircleClose v-else />
              </el-icon>
              <span :class="{ 'disabled': !perm.has }">{{ perm.name }}</span>
              <el-tooltip :content="perm.desc" placement="top">
                <el-icon class="info-icon"><InfoFilled /></el-icon>
              </el-tooltip>
            </div>
          </div>
        </div>

        <!-- 系统权限 - 仅系统管理员 -->
        <div class="perm-category" v-if="userStore.isAdmin">
          <div class="category-header">
            <el-icon :size="20" color="#f56c6c"><Setting /></el-icon>
            <span>系统权限</span>
            <el-tag size="small" type="danger">系统管理员</el-tag>
          </div>
          <div class="perm-list">
            <div class="perm-item has-perm" v-for="perm in systemPerms" :key="perm.code">
              <el-icon color="#f56c6c"><CircleCheck /></el-icon>
              <span>{{ perm.name }}</span>
              <el-tooltip :content="perm.desc" placement="top">
                <el-icon class="info-icon"><InfoFilled /></el-icon>
              </el-tooltip>
            </div>
          </div>
        </div>
      </div>

      <!-- 临时权限详情 -->
      <div class="permissions-section" v-if="tempPermissions.length > 0">
        <h3>
          临时权限详情
          <el-tag type="primary" size="small">{{ tempPermissions.length }} 个有效</el-tag>
        </h3>
        <el-row :gutter="15">
          <el-col :span="12" v-for="perm in tempPermissions" :key="perm.id">
            <div class="temp-perm-card">
              <div class="temp-perm-header">
                <span class="perm-title">{{ perm.permission_name }}</span>
                <el-tag
                  :type="perm.resource_type === 'all' ? 'danger' : (perm.resource_type === 'dept' ? 'primary' : 'info')"
                  size="small"
                >
                  {{ perm.resource_type === 'all' ? '全局' : (perm.resource_type === 'dept' ? '部门' : '特定') }}
                </el-tag>
              </div>
              <p class="perm-desc" v-if="permissionGroupDesc[perm.permission]">
                {{ permissionGroupDesc[perm.permission] }}
              </p>
              <div class="perm-meta">
                <span class="meta-item">
                  <el-icon><Timer /></el-icon>
                  过期: {{ formatDate(perm.expires_at) }}
                </span>
                <el-tag :type="getDaysLeftType(perm.days_left)" size="small">
                  剩余 {{ perm.days_left }} 天
                </el-tag>
              </div>
              <p class="perm-reason" v-if="perm.reason">
                <el-icon><Document /></el-icon>
                {{ perm.reason }}
              </p>
            </div>
          </el-col>
        </el-row>
      </div>

      <!-- 无临时权限提示 -->
      <div class="permissions-section" v-else>
        <h3>临时权限</h3>
        <el-empty description="暂无临时权限">
          <template #description>
            <p>暂无临时权限</p>
            <p class="hint">如需临时授权，请联系系统管理员或办公室管理员</p>
          </template>
        </el-empty>
      </div>

      <!-- 权限说明 -->
      <div class="permissions-section">
        <h3>权限说明</h3>
        <el-collapse>
          <el-collapse-item title="权限层级说明">
            <ul class="desc-list">
              <li><strong>系统管理员</strong>：拥有所有权限，可设置用户系统角色</li>
              <li><strong>办公室管理员</strong>：可管理所有部门排班和用户（除系统角色设置）</li>
              <li><strong>部门管理员</strong>：可管理部门内排班和成员</li>
              <li><strong>部门成员</strong>：可查看排班、编辑自己的无课表</li>
            </ul>
          </el-collapse-item>
          <el-collapse-item title="临时权限说明">
            <ul class="desc-list">
              <li>临时权限由管理员授予，有过期时间</li>
              <li>权限组会自动包含该组的所有子权限</li>
              <li>临时权限不能替代系统管理员权限</li>
            </ul>
          </el-collapse-item>
        </el-collapse>
      </div>
    </el-card>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useUserStore } from '../stores/user'
import { tempPermissionAPI } from '../api/tempPermission'
import { CircleCheck, CircleClose, User, UserFilled, OfficeBuilding, Star, Setting, InfoFilled, Timer, Document } from '@element-plus/icons-vue'

const userStore = useUserStore()
const tempPermissions = ref([])

// 权限组说明
const permissionGroupDesc = {
  'schedule:publish': '可设置各部门本周值班哪几天（每周分工）',
  'schedule:manage:all': '可管理所有部门的排班（预览、确认、编辑、导出等）',
  'user:manage:all': '可管理所有用户（查看、编辑、设置部门）',
  'schedule:manage:dept': '可管理部门内排班（预览、确认、编辑）',
  'user:manage:dept': '可管理部门内用户（查看、编辑）'
}

// 当前角色描述
const currentRoleDesc = computed(() => {
  if (userStore.isAdmin) return '您是系统管理员，拥有系统的最高权限，可以管理所有功能和用户。'
  if (userStore.isOfficeAdmin) return '您是办公室管理员，可以管理所有部门的排班和用户（但不能设置系统角色）。'
  if (userStore.isDeptAdmin) return '您是部门管理员，可以管理本部门的排班和成员信息。'
  return '您是部门成员，可以查看排班、编辑自己的无课表和查看值班信息。'
})

// 是否有部门权限区域
const hasDeptPerms = computed(() => userStore.isDeptAdmin || userStore.isOfficeAdmin || userStore.isAdmin || tempPermissions.value.length > 0)

// 是否有全局权限区域
const hasGlobalPerms = computed(() => userStore.isOfficeAdmin || userStore.isAdmin || tempPermissions.value.some(p => p.permission.includes(':all') || p.permission === 'schedule:publish'))

// 个人权限 - 所有人都有
const personalPerms = [
  { code: 'view_schedule', name: '查看排班', desc: '查看排班表和自己的值班安排' },
  { code: 'edit_availability', name: '编辑无课表', desc: '编辑和导入自己的无课表' },
  { code: 'view_duty', name: '查看值班', desc: '查看自己的值班信息和状态' },
]

// 部门权限
const deptPerms = computed(() => [
  { code: 'manage_dept_schedule', name: '管理排班', desc: '预览、确认、编辑部门排班', has: userStore.canManageDept },
  { code: 'manage_dept_users', name: '管理成员', desc: '查看、编辑部门成员信息', has: userStore.canManageDept },
])

// 全局权限
const globalPerms = computed(() => [
  { code: 'publish_assignment', name: '设置每周分工', desc: '设置各部门本周值班哪几天', has: userStore.canManageAll },
  { code: 'manage_all_schedule', name: '管理全部排班', desc: '管理所有部门的排班', has: userStore.canManageAll },
  { code: 'manage_all_users', name: '管理全部用户', desc: '管理所有用户信息（不含系统角色）', has: userStore.canManageAll },
])

// 系统权限 - 仅系统管理员
const systemPerms = [
  { code: 'system_admin', name: '系统管理', desc: '设置用户系统角色、系统配置等' },
  { code: 'temp_permission', name: '临时授权', desc: '授予和撤销用户的临时权限' },
  { code: 'smtp_config', name: 'SMTP配置', desc: '配置邮件发送服务' },
]

const loadTempPermissions = async () => {
  try {
    const res = await tempPermissionAPI.getMy()
    tempPermissions.value = res || []
  } catch (err) {
    console.error('加载临时权限失败:', err)
  }
}

const getDaysLeftType = (days) => {
  if (days <= 1) return 'danger'
  if (days <= 3) return 'warning'
  return 'success'
}

const formatDate = (dateStr) => {
  if (!dateStr) return '-'
  return new Date(dateStr).toLocaleString('zh-CN')
}

onMounted(() => {
  loadTempPermissions()
})
</script>

<style scoped>
.my-permissions-page {
  padding: 20px;
  max-width: 1000px;
  margin: 0 auto;
}

.card-header {
  font-size: 18px;
  font-weight: bold;
}

.role-overview {
  margin-bottom: 30px;
  padding: 20px;
  background: linear-gradient(135deg, #f5f7fa 0%, #e4e7ed 100%);
  border-radius: 8px;
}

.role-badges {
  margin: 15px 0;
}

.role-badges .el-tag {
  font-size: 14px;
  padding: 8px 15px;
}

.role-desc {
  color: #606266;
  font-size: 14px;
  margin-top: 10px;
  line-height: 1.6;
}

.permissions-section {
  margin-bottom: 30px;
}

h3 {
  margin-bottom: 16px;
  color: #303133;
  font-size: 16px;
  border-left: 4px solid #409eff;
  padding-left: 10px;
  display: flex;
  align-items: center;
  gap: 10px;
}

.perm-category {
  margin-bottom: 20px;
  padding: 15px;
  background: #f5f7fa;
  border-radius: 8px;
}

.category-header {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 12px;
  font-weight: bold;
  font-size: 15px;
}

.perm-list {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}

.perm-item {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 12px;
  background: #fff;
  border-radius: 4px;
  border: 1px solid #e4e7ed;
}

.perm-item.has-perm {
  border-color: #409eff;
  background: #ecf5ff;
}

.perm-item .disabled {
  color: #c0c4cc;
  text-decoration: line-through;
}

.info-icon {
  color: #909399;
  cursor: help;
  font-size: 14px;
}

.temp-perm-card {
  padding: 15px;
  margin-bottom: 15px;
  background: #fff;
  border: 1px solid #e4e7ed;
  border-radius: 8px;
  transition: all 0.3s;
}

.temp-perm-card:hover {
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.1);
}

.temp-perm-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
}

.perm-title {
  font-weight: bold;
  font-size: 15px;
  color: #303133;
}

.perm-desc {
  color: #606266;
  font-size: 13px;
  margin: 8px 0;
  padding: 8px 12px;
  background: #ecf5ff;
  border-radius: 4px;
  border-left: 3px solid #409eff;
}

.perm-meta {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: 10px;
}

.meta-item {
  display: flex;
  align-items: center;
  gap: 4px;
  color: #909399;
  font-size: 13px;
}

.perm-reason {
  margin-top: 8px;
  padding-top: 8px;
  border-top: 1px dashed #e4e7ed;
  color: #909399;
  font-size: 12px;
  display: flex;
  align-items: center;
  gap: 4px;
}

.desc-list {
  padding-left: 20px;
  color: #606266;
  line-height: 2;
}

.desc-list li {
  margin-bottom: 5px;
}

.hint {
  color: #909399;
  font-size: 12px;
  margin-top: 8px;
}
</style>
