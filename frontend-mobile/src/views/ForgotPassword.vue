<template>
  <div class="login-page">
    <div class="bg-layer"></div>
    <div class="vignette"></div>
    <canvas class="rain-canvas" ref="rainCanvas"></canvas>

    <div class="container">
      <div class="auth-card initial-load" ref="authCard">
        <div class="view-section active">
          <div class="brand">
            <h1>找回密码</h1>
            <p>输入学号，我们将发送重置链接到您的邮箱</p>
          </div>

          <form @submit.prevent="handleSubmit">
            <div class="form-group">
              <label class="form-label">学号</label>
              <input
                type="text"
                class="form-input"
                placeholder="请输入学号"
                v-model="form.student_id"
                autocomplete="username"
              />
            </div>
            <button type="submit" class="btn-primary" :disabled="loading">
              {{ loading ? '发送中...' : '发送重置链接' }}
            </button>
          </form>

          <div class="footer">
            <router-link to="/login" class="text-link">返回登录</router-link>
          </div>
        </div>
      </div>
    </div>

    <div class="toast" :class="{ show: toast.visible }">{{ toast.message }}</div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { showToast as vantShowToast } from 'vant'
import { forgotPassword } from '../api/user'

const router = useRouter()
const loading = ref(false)
const rainCanvas = ref(null)
const authCard = ref(null)

const form = reactive({
  student_id: ''
})

const toast = reactive({
  visible: false,
  message: ''
})

let toastTimer = null
function showToast(msg) {
  toast.message = msg
  toast.visible = true
  clearTimeout(toastTimer)
  toastTimer = setTimeout(() => {
    toast.visible = false
  }, 2500)
}

async function handleSubmit() {
  if (!form.student_id) {
    showToast('请输入学号')
    return
  }

  loading.value = true
  try {
    await forgotPassword(form)
    vantShowToast({ message: '重置链接已发送到您的邮箱', type: 'success' })
  } catch {
    // 错误已在拦截器处理
  } finally {
    loading.value = false
  }
}

let animationId = null
let resizeHandler = null

onMounted(() => {
  setTimeout(() => {
    if (authCard.value) {
      authCard.value.classList.remove('initial-load')
    }
  }, 1000)

  const canvas = rainCanvas.value
  if (!canvas) return
  const ctx = canvas.getContext('2d')
  let width, height
  const dpr = window.devicePixelRatio || 1

  function resize() {
    width = window.innerWidth
    height = window.innerHeight
    canvas.width = width * dpr
    canvas.height = height * dpr
    canvas.style.width = width + 'px'
    canvas.style.height = height + 'px'
    ctx.setTransform(dpr, 0, 0, dpr, 0, 0)
  }
  resize()
  resizeHandler = resize
  window.addEventListener('resize', resize)

  const rainCount = Math.min(window.innerWidth * 0.1, 80)
  const rains = []
  class RainDrop {
    constructor() {
      this.reset()
      this.y = Math.random() * height
    }
    reset() {
      this.x = Math.random() * width
      this.y = -50
      this.length = Math.random() * 16 + 10
      this.speed = Math.random() * 6 + 4
      this.opacity = Math.random() * 0.3 + 0.1
      this.angle = Math.random() * 0.2 + 0.1
      this.width = Math.random() * 1.2 + 0.5
    }
    update() {
      this.y += this.speed
      this.x += this.angle
      if (this.y > height + this.length) this.reset()
    }
    draw() {
      ctx.beginPath()
      ctx.moveTo(this.x, this.y)
      ctx.lineTo(this.x + this.angle * 2, this.y + this.length)
      ctx.strokeStyle = `rgba(200, 220, 255, ${this.opacity})`
      ctx.lineWidth = this.width
      ctx.lineCap = 'round'
      ctx.stroke()
    }
  }
  for (let i = 0; i < rainCount; i++) rains.push(new RainDrop())

  const petals = []
  class Petal {
    constructor() {
      this.reset()
      this.y = Math.random() * height
    }
    reset() {
      this.x = Math.random() * width
      this.y = -20
      this.size = Math.random() * 5 + 2
      this.speedY = Math.random() * 1.2 + 0.4
      this.speedX = Math.random() * 1 - 0.5
      this.rotation = Math.random() * 360
      this.rotationSpeed = Math.random() * 2 - 1
      this.opacity = Math.random() * 0.5 + 0.3
      this.sway = Math.random() * 0.02 + 0.01
      this.swayOffset = Math.random() * Math.PI * 2
    }
    update(time) {
      this.y += this.speedY
      this.x += this.speedX + Math.sin(time * this.sway + this.swayOffset) * 0.5
      this.rotation += this.rotationSpeed
      if (this.y > height + 20 || this.x < -20 || this.x > width + 20) this.reset()
    }
    draw() {
      ctx.save()
      ctx.translate(this.x, this.y)
      ctx.rotate(this.rotation * Math.PI / 180)
      ctx.fillStyle = `rgba(255, 220, 230, ${this.opacity})`
      ctx.beginPath()
      ctx.ellipse(0, 0, this.size, this.size * 0.6, 0, 0, Math.PI * 2)
      ctx.fill()
      ctx.restore()
    }
  }
  for (let i = 0; i < 15; i++) petals.push(new Petal())

  let time = 0
  function animate() {
    ctx.clearRect(0, 0, width, height)
    time += 0.01
    rains.forEach(r => { r.update(); r.draw() })
    petals.forEach(p => { p.update(time); p.draw() })
    animationId = requestAnimationFrame(animate)
  }
  animate()
})

onUnmounted(() => {
  if (animationId) cancelAnimationFrame(animationId)
  if (resizeHandler) window.removeEventListener('resize', resizeHandler)
  clearTimeout(toastTimer)
})
</script>

<style scoped>
.login-page {
  position: fixed;
  inset: 0;
  overflow: hidden;
  font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", "PingFang SC", "Hiragino Sans GB", "Microsoft YaHei", sans-serif;
  z-index: 0;
}

.bg-layer {
  position: fixed;
  inset: 0;
  z-index: 0;
  background-image: url('../assets/bg-mobile.png');
  background-size: cover;
  background-position: center;
  transition: transform 0.3s ease-out;
}

.vignette {
  position: fixed;
  inset: 0;
  z-index: 1;
  background: radial-gradient(ellipse at center, transparent 0%, rgba(0,0,0,0.15) 100%);
  pointer-events: none;
}

.rain-canvas {
  position: fixed;
  inset: 0;
  z-index: 2;
  pointer-events: none;
}

.container {
  position: relative;
  z-index: 10;
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 20px;
  padding-top: env(safe-area-inset-top);
  padding-bottom: env(safe-area-inset-bottom);
}

.auth-card {
  width: 90%;
  max-width: 400px;
  padding: 40px 32px;
  border-radius: 24px;
  background: rgba(255, 255, 255, 0.12);
  backdrop-filter: blur(20px);
  -webkit-backdrop-filter: blur(20px);
  border: 1px solid rgba(255, 255, 255, 0.35);
  box-shadow:
    0 8px 32px 0 rgba(0, 0, 0, 0.2),
    inset 0 0 0 1px rgba(255,255,255,0.1);
  position: relative;
  overflow: hidden;
}

@media (max-width: 480px) {
  .auth-card {
    padding: 36px 28px;
  }
}

.view-section {
  display: none;
  animation: viewEnter 0.5s cubic-bezier(0.22, 1, 0.36, 1) forwards;
}

.view-section.active {
  display: block;
}

@keyframes viewEnter {
  from {
    opacity: 0;
    transform: translateY(20px) scale(0.98);
  }
  to {
    opacity: 1;
    transform: translateY(0) scale(1);
  }
}

.auth-card.initial-load {
  animation: cardEnter 1s cubic-bezier(0.22, 1, 0.36, 1) forwards;
}

@keyframes cardEnter {
  from {
    opacity: 0;
    transform: translateY(30px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.brand {
  text-align: center;
  margin-bottom: 32px;
}

.brand h1 {
  font-size: 26px;
  font-weight: 300;
  color: rgba(255, 255, 255, 0.95);
  letter-spacing: 5px;
  text-shadow: 0 2px 8px rgba(0,0,0,0.15);
  margin-bottom: 8px;
}

.brand p {
  font-size: 13px;
  color: rgba(255, 255, 255, 0.7);
  letter-spacing: 1px;
  text-shadow: 0 1px 4px rgba(0,0,0,0.1);
}

@media (max-width: 480px) {
  .brand h1 {
    font-size: 22px;
    letter-spacing: 4px;
  }
}

.form-group {
  margin-bottom: 22px;
  opacity: 0;
  animation: fadeUp 0.6s ease-out forwards;
}

.form-group:nth-child(1) { animation-delay: 0.1s; }

@keyframes fadeUp {
  from {
    opacity: 0;
    transform: translateY(12px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.form-label {
  display: block;
  font-size: 13px;
  color: rgba(255, 255, 255, 0.7);
  margin-bottom: 8px;
  letter-spacing: 1px;
  transition: color 0.3s;
}

.form-input {
  width: 100%;
  padding: 14px 18px;
  border-radius: 12px;
  border: 1px solid rgba(255,255,255,0.25);
  background: rgba(0, 0, 0, 0.15);
  color: rgba(255, 255, 255, 0.95);
  font-size: 15px;
  outline: none;
  transition: all 0.3s ease;
  backdrop-filter: blur(4px);
  font-family: inherit;
}

.form-input::placeholder {
  color: rgba(255,255,255,0.4);
}

.form-input:focus {
  border-color: rgba(255,255,255,0.7);
  background: rgba(0,0,0,0.25);
  box-shadow: 0 0 20px rgba(255,255,255,0.15), inset 0 0 0 1px rgba(255,255,255,0.1);
}

.form-group:focus-within .form-label {
  color: rgba(255, 255, 255, 0.95);
}

.text-link {
  color: rgba(255, 255, 255, 0.7);
  text-decoration: none;
  transition: all 0.3s;
  position: relative;
  background: none;
  border: none;
  font-size: 13px;
  cursor: pointer;
  font-family: inherit;
  padding: 0;
}

.text-link::after {
  content: '';
  position: absolute;
  bottom: -2px;
  left: 0;
  width: 0;
  height: 1px;
  background: rgba(255,255,255,0.6);
  transition: width 0.3s;
}

.text-link:hover {
  color: rgba(255, 255, 255, 0.95);
}

.text-link:hover::after {
  width: 100%;
}

.btn-primary {
  width: 100%;
  padding: 14px;
  border-radius: 12px;
  border: 1px solid rgba(255,255,255,0.4);
  background: rgba(255,255,255,0.15);
  color: rgba(255, 255, 255, 0.95);
  font-size: 16px;
  letter-spacing: 4px;
  cursor: pointer;
  transition: all 0.3s ease;
  position: relative;
  overflow: hidden;
  font-family: inherit;
  font-weight: 500;
  opacity: 0;
  animation: fadeUp 0.6s ease-out 0.4s forwards;
}

.btn-primary::before {
  content: '';
  position: absolute;
  inset: 0;
  background: rgba(255,255,255,0.2);
  transform: translateX(-100%);
  transition: transform 0.5s ease;
}

.btn-primary:hover:not(:disabled) {
  transform: translateY(-2px);
  box-shadow: 0 8px 24px rgba(0,0,0,0.2);
  border-color: rgba(255,255,255,0.7);
  background: rgba(255,255,255,0.25);
}

.btn-primary:hover:not(:disabled)::before {
  transform: translateX(100%);
}

.btn-primary:active:not(:disabled) {
  transform: translateY(0);
}

.btn-primary:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.footer {
  text-align: center;
  margin-top: 24px;
  font-size: 13px;
  color: rgba(255, 255, 255, 0.7);
  opacity: 0;
  animation: fadeUp 0.6s ease-out 0.5s forwards;
}

.footer :deep(a) {
  color: rgba(255, 255, 255, 0.95);
  text-decoration: none;
  font-weight: 500;
  transition: opacity 0.3s;
}

.footer :deep(a:hover) {
  opacity: 0.8;
}

.form-input:-webkit-autofill,
.form-input:-webkit-autofill:hover,
.form-input:-webkit-autofill:focus {
  -webkit-text-fill-color: rgba(255, 255, 255, 0.95);
  -webkit-box-shadow: 0 0 0px 1000px rgba(0,0,0,0.3) inset;
  transition: background-color 5000s ease-in-out 0s;
}

.toast {
  position: fixed;
  top: 40px;
  left: 50%;
  transform: translateX(-50%) translateY(-20px);
  background: rgba(255,255,255,0.2);
  backdrop-filter: blur(20px);
  -webkit-backdrop-filter: blur(20px);
  border: 1px solid rgba(255,255,255,0.4);
  color: rgba(255, 255, 255, 0.95);
  padding: 12px 28px;
  border-radius: 50px;
  font-size: 14px;
  opacity: 0;
  pointer-events: none;
  transition: all 0.4s ease;
  z-index: 100;
  letter-spacing: 1px;
  box-shadow: 0 8px 32px rgba(0,0,0,0.15);
}

.toast.show {
  opacity: 1;
  transform: translateX(-50%) translateY(0);
}
</style>
