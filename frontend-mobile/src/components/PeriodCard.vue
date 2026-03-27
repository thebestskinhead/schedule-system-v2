<template>
  <div
    class="period-card"
    :class="cardClass"
    @click="handleClick"
  >
    <div class="period-card__left">
      <span class="period-card__emoji">{{ emoji }}</span>
      <div class="period-card__info">
        <div class="period-card__name">{{ periodName }}</div>
        <div class="period-card__time">{{ timeRange }}</div>
      </div>
    </div>
    <div class="period-card__status" :class="statusClass">
      <van-icon :name="status === 'free' ? 'success' : 'cross'" size="18" color="#fff" />
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  period: { type: Number, required: true },
  status: { type: String, default: 'free' }, // 'free' | 'busy'
  active: { type: Boolean, default: false }
})

const emit = defineEmits(['change'])

const config = {
  1: { name: '1-2节', emoji: '🌅', start: '08:00', end: '09:40' },
  2: { name: '3-4节', emoji: '🌞', start: '10:00', end: '11:40' },
  3: { name: '5-6节', emoji: '🌆', start: '14:00', end: '15:40' },
  4: { name: '7-8节', emoji: '🌙', start: '16:00', end: '17:40' }
}

const periodName = computed(() => config[props.period]?.name || '')
const emoji = computed(() => config[props.period]?.emoji || '')
const timeRange = computed(() => {
  const c = config[props.period]
  return c ? `${c.start} - ${c.end}` : ''
})

const cardClass = computed(() => ({
  'period-card--free': props.status === 'free',
  'period-card--busy': props.status === 'busy',
  'period-card--active': props.active
}))

const statusClass = computed(() => ({
  'period-card__status--free': props.status === 'free',
  'period-card__status--busy': props.status === 'busy'
}))

const handleClick = () => {
  emit('change', {
    period: props.period,
    status: props.status === 'free' ? 'busy' : 'free'
  })
}
</script>

<style scoped>
.period-card {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: var(--space-md, 12px) var(--space, 16px);
  border-radius: var(--radius-lg, 12px);
  border: 1px solid var(--color-border, #EBEDF0);
  background: var(--color-surface, #FFFFFF);
  margin-bottom: var(--space-sm, 8px);
  transition: all 0.2s;
  user-select: none;
}

.period-card:active {
  transform: scale(0.98);
}

.period-card--free {
  background: var(--color-success-light, #E6F7EF);
  border-color: var(--color-success, #07C160);
}

.period-card--busy {
  background: var(--color-danger-light, #FFEBED);
  border-color: var(--color-danger, #EE0A24);
}

.period-card--active {
  border-color: var(--color-primary, #1989FA);
  box-shadow: 0 0 0 2px var(--color-primary-light, #E6F2FF);
}

.period-card__left {
  display: flex;
  align-items: center;
  gap: var(--space-md, 12px);
}

.period-card__emoji {
  font-size: 24px;
}

.period-card__name {
  font-size: var(--font-size-h3, 16px);
  font-weight: var(--font-weight-semibold, 600);
  color: var(--color-text-primary, #323233);
}

.period-card__time {
  font-size: var(--font-size-small, 12px);
  color: var(--color-text-tertiary, #969799);
  margin-top: 2px;
}

.period-card__status {
  width: 28px;
  height: 28px;
  border-radius: var(--radius-full, 9999px);
  display: flex;
  align-items: center;
  justify-content: center;
}

.period-card__status--free {
  background: var(--color-success, #07C160);
}

.period-card__status--busy {
  background: var(--color-danger, #EE0A24);
}
</style>
