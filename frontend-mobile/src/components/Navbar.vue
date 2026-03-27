<template>
  <div class="navbar">
    <div class="navbar__left" v-if="showBack">
      <van-icon name="arrow-left" size="20" @click="handleBack" />
    </div>
    <div class="navbar__placeholder" v-else></div>

    <div class="navbar__title">{{ title }}</div>

    <div class="navbar__right">
      <slot name="right"></slot>
    </div>
  </div>
</template>

<script setup>
import { useRouter } from 'vue-router'

defineProps({
  title: { type: String, default: '' },
  showBack: { type: Boolean, default: false }
})

const router = useRouter()

const handleBack = () => {
  router.back()
}
</script>

<style scoped>
.navbar {
  display: flex;
  align-items: center;
  height: 44px;
  padding: 0 var(--space, 16px);
  background: var(--color-surface, #FFFFFF);
  box-shadow: var(--shadow-sm, 0 2px 4px rgba(0, 0, 0, 0.04));
  position: sticky;
  top: 0;
  z-index: 100;
}

.navbar__left {
  width: 32px;
  height: 32px;
  display: flex;
  align-items: center;
  cursor: pointer;
}

.navbar__placeholder {
  width: 32px;
}

.navbar__title {
  flex: 1;
  text-align: center;
  font-size: var(--font-size-h2, 18px);
  font-weight: var(--font-weight-semibold, 600);
  color: var(--color-text-primary, #323233);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.navbar__right {
  width: 32px;
  display: flex;
  align-items: center;
  justify-content: flex-end;
}
</style>
