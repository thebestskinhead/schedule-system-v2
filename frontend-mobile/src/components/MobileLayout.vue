<template>
  <div class="mobile-layout">
    <!-- 主内容区 -->
    <main class="mobile-layout__content">
      <router-view />
    </main>

    <!-- 底部 TabBar -->
    <van-tabbar
      v-model="activeTab"
      :safe-area-inset-bottom="true"
      active-color="#1989FA"
      inactive-color="#969799"
      @change="handleTabChange"
    >
      <van-tabbar-item
        v-for="tab in tabs"
        :key="tab.name"
        :name="tab.name"
        :icon="tab.icon"
        :to="tab.path"
      >
        {{ tab.label }}
      </van-tabbar-item>
    </van-tabbar>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'

const route = useRoute()
const router = useRouter()

const tabs = [
  { name: 'home', label: '首页', path: '/', icon: 'home-o' },
  { name: 'availability', label: '无课表', path: '/availability', icon: 'calendar-o' },
  { name: 'schedule', label: '我的排班', path: '/duty/my', icon: 'chart-trending-o' },
  { name: 'profile', label: '我的', path: '/profile', icon: 'user-o' }
]

const activeTab = computed(() => {
  if (route.path === '/') return 'home'
  const match = tabs.find(tab => tab.path !== '/' && route.path.startsWith(tab.path))
  return match?.name || 'home'
})

const handleTabChange = (name) => {
  const tab = tabs.find(t => t.name === name)
  if (tab) {
    router.push(tab.path)
  }
}
</script>

<style scoped>
.mobile-layout {
  height: 100%;
  display: flex;
  flex-direction: column;
  background: var(--color-bg, #F7F8FA);
}

.mobile-layout__content {
  flex: 1;
  overflow-y: auto;
  -webkit-overflow-scrolling: touch;
}
</style>
