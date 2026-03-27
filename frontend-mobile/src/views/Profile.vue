<template>
  <div class="profile-page">
    <van-nav-bar title="我的" />

    <!-- 用户信息 -->
    <div class="profile-card">
      <van-cell-group inset>
        <van-cell
          :title="userStore.user?.name || '未知用户'"
          :label="userStore.user?.department || ''"
          is-link
          to="/my-permissions"
        >
          <template #icon>
            <van-icon name="user-o" size="24" style="margin-right: 10px; color: #1989FA;" />
          </template>
          <template #value>
            <van-tag v-if="userStore.isAdmin" type="danger" size="medium">管理员</van-tag>
            <van-tag v-else-if="userStore.isDeptAdmin" type="warning" size="medium">部门管理</van-tag>
          </template>
        </van-cell>
      </van-cell-group>
    </div>

    <!-- 功能列表 -->
    <van-cell-group inset title="功能">
      <van-cell title="我的值班" icon="clock-o" to="/duty/my" is-link />
      <van-cell title="我的权限" icon="shield-o" to="/my-permissions" is-link />
      <van-cell title="使用说明" icon="info-o" to="/readme" is-link />
    </van-cell-group>

    <!-- 管理功能 -->
    <van-cell-group v-if="userStore.canManageDept" inset title="管理">
      <van-cell title="排班管理" icon="chart-trending-o" to="/schedule" is-link />
      <van-cell title="值班安排" icon="orders-o" to="/admin/duty-assignments" is-link />
      <van-cell v-if="userStore.canManageDept" title="用户管理" icon="friends-o" to="/admin/users" is-link />
      <van-cell v-if="userStore.canManageAll" title="临时权限" icon="medal-o" to="/admin/temp-permissions" is-link />
      <van-cell v-if="userStore.canManageAll" title="学期设置" icon="setting-o" to="/admin/semester" is-link />
      <van-cell v-if="userStore.isAdmin" title="邮件设置" icon="envelop-o" to="/admin/smtp" is-link />
    </van-cell-group>

    <!-- 账号 -->
    <van-cell-group inset title="账号">
      <van-cell title="修改密码" icon="lock" is-link @click="showPasswordDialog = true" />
      <van-cell title="退出登录" icon="revoke" is-link class="logout-cell" @click="handleLogout" />
    </van-cell-group>

    <!-- 修改密码弹窗 -->
    <van-dialog
      v-model:show="showPasswordDialog"
      title="修改密码"
      show-cancel-button
      :before-close="handleChangePassword"
    >
      <div class="password-form">
        <van-field
          v-model="passwordForm.oldPassword"
          type="password"
          label="原密码"
          placeholder="请输入原密码"
        />
        <van-field
          v-model="passwordForm.newPassword"
          type="password"
          label="新密码"
          placeholder="请输入新密码"
        />
        <van-field
          v-model="passwordForm.confirmPassword"
          type="password"
          label="确认密码"
          placeholder="请再次输入新密码"
        />
      </div>
    </van-dialog>
  </div>
</template>

<script setup>
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { showDialog, showSuccessToast } from 'vant'
import { useUserStore } from '../stores/user'
import { changePassword } from '../api/user'

const router = useRouter()
const userStore = useUserStore()

const showPasswordDialog = ref(false)
const passwordForm = reactive({
  oldPassword: '',
  newPassword: '',
  confirmPassword: ''
})

function handleLogout() {
  showDialog({
    title: '退出登录',
    message: '确定要退出登录吗？',
    showCancelButton: true
  }).then(() => {
    userStore.logout()
    router.replace('/login')
  }).catch(() => {})
}

async function handleChangePassword(action) {
  if (action !== 'confirm') return true

  if (!passwordForm.oldPassword || !passwordForm.newPassword) {
    showDialog({ title: '提示', message: '请填写完整' })
    return false
  }

  if (passwordForm.newPassword !== passwordForm.confirmPassword) {
    showDialog({ title: '提示', message: '两次输入的密码不一致' })
    return false
  }

  try {
    await changePassword({
      old_password: passwordForm.oldPassword,
      new_password: passwordForm.newPassword
    })
    showSuccessToast('密码修改成功')
    passwordForm.oldPassword = ''
    passwordForm.newPassword = ''
    passwordForm.confirmPassword = ''
    return true
  } catch (error) {
    showDialog({ title: '修改失败', message: error.message || '请稍后重试' })
    return false
  }
}
</script>

<style scoped>
.profile-page {
  min-height: 100%;
  background: #f7f8fa;
  padding-top: env(safe-area-inset-top);
  padding-bottom: calc(20px + env(safe-area-inset-bottom));
}

.profile-card {
  margin-top: 12px;
}

.logout-cell :deep(.van-cell__title) {
  color: #ee0a24;
}

.password-form {
  padding: 12px 0;
}
</style>
