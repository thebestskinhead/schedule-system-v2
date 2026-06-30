<template>
  <Teleport to="body">
    <div class="qr-overlay" v-if="visible" @click.self="handleClose">
      <div class="qr-modal">
        <div class="qr-header">
          <h3>微信扫码登录</h3>
          <button class="qr-close" @click="handleClose">&times;</button>
        </div>

        <!-- 等待扫码 / 已扫码 -->
        <div class="qr-body" v-if="status === 'pending' || status === 'scanned' || status === 'confirmed'">
          <div class="qr-img-wrap">
            <img v-if="qrCode" :src="qrCode" alt="二维码" class="qr-img" />
            <div class="qr-overlay-mask" v-if="status === 'scanned' || status === 'confirmed'">
              <div class="qr-status-icon" v-if="status === 'scanned'">&#10003;</div>
              <div class="qr-spinner" v-if="status === 'confirmed'"></div>
            </div>
          </div>
          <p class="qr-tip">{{ statusText }}</p>
          <p class="qr-sub" v-if="status === 'pending'">二维码 {{ countdown }} 秒后过期</p>
        </div>

        <!-- 过期 -->
        <div class="qr-body" v-else-if="status === 'expired'">
          <div class="qr-icon-wrap error">
            <span>&#9888;</span>
          </div>
          <p class="qr-tip">二维码已过期</p>
          <button class="qr-btn" @click="refresh">重新获取</button>
        </div>

        <!-- 错误 -->
        <div class="qr-body" v-else-if="status === 'error'">
          <div class="qr-icon-wrap error">
            <span>&#10007;</span>
          </div>
          <p class="qr-tip">登录失败</p>
          <p class="qr-sub">{{ message }}</p>
          <button class="qr-btn" @click="refresh">重试</button>
        </div>

        <!-- 需要注册 -->
        <div class="qr-body" v-else-if="status === 'need_reg'">
          <div class="qr-icon-wrap success">
            <span>&#10003;</span>
          </div>
          <p class="qr-tip">扫码成功，但账号未注册</p>
          <p class="qr-sub">即将跳转到注册页面...</p>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<script setup>
import { ref, computed, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { useUserStore } from '../stores/user'
import { qrLoginStart, qrLoginPoll } from '../api/user'

const props = defineProps({
  modelValue: { type: Boolean, default: false }
})
const emit = defineEmits(['update:modelValue'])

const router = useRouter()
const userStore = useUserStore()

const visible = computed({
  get: () => props.modelValue,
  set: (v) => emit('update:modelValue', v)
})

const sessionId = ref('')
const qrCode = ref('')
const status = ref('pending')
const message = ref('')
const countdown = ref(300)

let pollTimer = null
let countTimer = null

const statusText = computed(() => {
  const map = {
    pending: '请使用微信扫描二维码',
    scanned: '已扫码，请在手机上确认登录',
    confirmed: '正在处理登录...',
  }
  return map[status.value] || message.value
})

// 开始扫码登录
async function start() {
  try {
    const res = await qrLoginStart()
    sessionId.value = res.session_id
    qrCode.value = res.qrcode
    status.value = 'pending'
    message.value = ''
    countdown.value = res.expires_in || 300

    startPolling()
    startCountdown()
  } catch (error) {
    status.value = 'error'
    message.value = error.message || '获取二维码失败'
  }
}

// 轮询扫码状态
function startPolling() {
  const poll = async () => {
    if (!sessionId.value) return

    try {
      const res = await qrLoginPoll({ session_id: sessionId.value })

      if (res.status === 'success') {
        // 登录成功
        userStore.qrLogin({
          token: res.token,
          user: res.user
        })
        visible.value = false
        router.push('/')
        return
      } else if (res.status === 'need_reg') {
        // 未注册，跳转注册页
        status.value = 'need_reg'
        setTimeout(() => {
          visible.value = false
          router.push({
            name: 'Register',
            query: {
              student_id: res.student_id,
              name: res.name
            }
          })
        }, 1500)
        return
      } else if (res.status === 'expired' || res.status === 'error') {
        status.value = res.status
        message.value = res.message || ''
        return
      }

      status.value = res.status
      message.value = res.message || ''
      pollTimer = setTimeout(poll, 2000)
    } catch (error) {
      // 网络错误，继续轮询
      pollTimer = setTimeout(poll, 2000)
    }
  }

  pollTimer = setTimeout(poll, 2000)
}

// 倒计时
function startCountdown() {
  countTimer = setInterval(() => {
    countdown.value--
    if (countdown.value <= 0) {
      clearInterval(countTimer)
    }
  }, 1000)
}

// 刷新二维码
function refresh() {
  clearTimeout(pollTimer)
  clearInterval(countTimer)
  start()
}

// 关闭弹窗
function handleClose() {
  clearTimeout(pollTimer)
  clearInterval(countTimer)
  visible.value = false
}

onUnmounted(() => {
  clearTimeout(pollTimer)
  clearInterval(countTimer)
})

// 监听 visible 变化，打开时自动开始
import { watch } from 'vue'
watch(visible, (val) => {
  if (val) {
    start()
  }
})
</script>

<style scoped>
.qr-overlay {
  position: fixed;
  inset: 0;
  z-index: 1000;
  background: rgba(0, 0, 0, 0.45);
  display: flex;
  align-items: center;
  justify-content: center;
  backdrop-filter: blur(4px);
}

.qr-modal {
  width: 360px;
  background: rgba(255, 255, 255, 0.15);
  backdrop-filter: blur(24px);
  -webkit-backdrop-filter: blur(24px);
  border-radius: 20px;
  border: 1px solid rgba(255, 255, 255, 0.35);
  box-shadow: 0 12px 40px rgba(0, 0, 0, 0.25);
  overflow: hidden;
}

.qr-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 20px 24px 0;
}

.qr-header h3 {
  font-size: 18px;
  font-weight: 500;
  color: rgba(255, 255, 255, 0.95);
  margin: 0;
  letter-spacing: 2px;
}

.qr-close {
  width: 32px;
  height: 32px;
  border: none;
  background: rgba(255,255,255,0.1);
  color: rgba(255, 255, 255, 0.8);
  font-size: 20px;
  border-radius: 50%;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.3s;
}

.qr-close:hover {
  background: rgba(255,255,255,0.2);
  color: #fff;
}

.qr-body {
  padding: 24px 32px 32px;
  text-align: center;
}

.qr-img-wrap {
  position: relative;
  width: 200px;
  height: 200px;
  margin: 0 auto 20px;
  background: #fff;
  border-radius: 12px;
  padding: 8px;
  box-sizing: border-box;
}

.qr-img {
  width: 100%;
  height: 100%;
  object-fit: contain;
  border-radius: 8px;
}

.qr-overlay-mask {
  position: absolute;
  inset: 8px;
  background: rgba(255, 255, 255, 0.85);
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.qr-status-icon {
  font-size: 48px;
  color: #52c41a;
  font-weight: bold;
}

.qr-spinner {
  width: 40px;
  height: 40px;
  border: 3px solid rgba(0, 0, 0, 0.1);
  border-top-color: #1677ff;
  border-radius: 50%;
  animation: qrSpin 0.8s linear infinite;
}

@keyframes qrSpin {
  to { transform: rotate(360deg); }
}

.qr-tip {
  font-size: 14px;
  color: rgba(255, 255, 255, 0.9);
  margin: 0 0 8px;
  letter-spacing: 1px;
}

.qr-sub {
  font-size: 12px;
  color: rgba(255, 255, 255, 0.55);
  margin: 0;
}

.qr-icon-wrap {
  width: 64px;
  height: 64px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  margin: 0 auto 20px;
  font-size: 28px;
}

.qr-icon-wrap.error {
  background: rgba(255, 77, 79, 0.2);
  color: #ff4d4f;
}

.qr-icon-wrap.success {
  background: rgba(82, 196, 26, 0.2);
  color: #52c41a;
}

.qr-btn {
  margin-top: 16px;
  padding: 10px 32px;
  border-radius: 8px;
  border: 1px solid rgba(255, 255, 255, 0.3);
  background: rgba(255, 255, 255, 0.1);
  color: rgba(255, 255, 255, 0.9);
  font-size: 14px;
  cursor: pointer;
  transition: all 0.3s;
  font-family: inherit;
  letter-spacing: 1px;
}

.qr-btn:hover {
  background: rgba(255, 255, 255, 0.2);
  border-color: rgba(255, 255, 255, 0.5);
}
</style>
