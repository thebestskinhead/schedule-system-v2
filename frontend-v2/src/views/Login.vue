<template>
  <div class="login-page">
    <div class="bg-layer"></div>
    <div class="vignette"></div>
    <canvas class="rain-canvas" ref="rainCanvas"></canvas>

    <div class="container">
      <div class="auth-card initial-load" ref="authCard">
        <div class="view-section active">
          <div class="brand">
            <h1>排班系统</h1>
            <p>高效协作 · 智能排班</p>
          </div>
          <form @submit.prevent="handleLogin">
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
            <div class="form-group">
              <label class="form-label">密码</label>
              <input
                type="password"
                class="form-input"
                placeholder="请输入密码"
                v-model="form.password"
                autocomplete="current-password"
              />
            </div>
            <div class="options">
              <label class="remember">
                <input type="checkbox" v-model="form.remember" />
                <span>记住我</span>
              </label>
              <router-link to="/forgot-password" class="text-link">忘记密码？</router-link>
            </div>
            <button type="submit" class="btn-primary" :disabled="loading">
              {{ loading ? '登录中...' : '登 录' }}
            </button>
          </form>
          <div class="footer">
            还没有账号？<router-link to="/register" class="text-link" style="font-size:13px;">立即注册</router-link>
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
import { ElMessage } from 'element-plus'
import { useUserStore } from '../stores/user'
import { login } from '../api/user'

const router = useRouter()
const userStore = useUserStore()
const loading = ref(false)
const rainCanvas = ref(null)
const authCard = ref(null)

const form = reactive({
  student_id: '',
  password: '',
  remember: true
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

async function handleLogin() {
  if (!form.student_id || !form.password) {
    showToast('请填写完整信息')
    return
  }

  loading.value = true
  try {
    const res = await login(form)
    userStore.setToken(res.token)
    await userStore.loadUserInfo()
    ElMessage.success('登录成功')
    router.push('/')
  } catch (error) {
    // 错误已在请求拦截器处理
  } finally {
    loading.value = false
  }
}

let animationId = null
let resizeHandler = null
let mouseHandler = null

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

  const rainCount = Math.min(window.innerWidth * 0.15, 180)
  const rains = []
  class RainDrop {
    constructor() {
      this.reset()
      this.y = Math.random() * height
    }
    reset() {
      this.x = Math.random() * width
      this.y = -50
      this.length = Math.random() * 20 + 15
      this.speed = Math.random() * 8 + 6
      this.opacity = Math.random() * 0.3 + 0.1
      this.angle = Math.random() * 0.2 + 0.1
      this.width = Math.random() * 1.5 + 0.5
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
      this.size = Math.random() * 6 + 3
      this.speedY = Math.random() * 1.5 + 0.5
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
  for (let i = 0; i < 25; i++) petals.push(new Petal())

  let time = 0
  function animate() {
    ctx.clearRect(0, 0, width, height)
    time += 0.01
    rains.forEach(r => { r.update(); r.draw() })
    petals.forEach(p => { p.update(time); p.draw() })
    animationId = requestAnimationFrame(animate)
  }
  animate()

  const bg = document.querySelector('.login-page .bg-layer')
  mouseHandler = (e) => {
    const x = (e.clientX / width - 0.5) * 10
    const y = (e.clientY / height - 0.5) * 10
    if (bg) {
      bg.style.transform = `scale(1.05) translate(${-x}px, ${-y}px)`
    }
  }
  document.addEventListener('mousemove', mouseHandler)
})

onUnmounted(() => {
  if (animationId) cancelAnimationFrame(animationId)
  if (resizeHandler) window.removeEventListener('resize', resizeHandler)
  if (mouseHandler) document.removeEventListener('mousemove', mouseHandler)
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
  background-size: cover;
  background-position: center;
  transition: transform 0.3s ease-out;
}

@media (min-width: 769px) {
  .bg-layer {
    background-image: url('../assets/bg-pc.png');
  }
}

@media (max-width: 768px) {
  .bg-layer {
    background-image: url('../assets/bg-mobile.png');
  }
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
  padding: 20px;
}

@media (min-width: 769px) {
  .container {
    justify-content: flex-start;
    padding-left: 12%;
  }
}

@media (max-width: 768px) {
  .container {
    justify-content: center;
  }
}

.auth-card {
  width: 420px;
  padding: 48px 40px;
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

@media (max-width: 768px) {
  .auth-card {
    width: 90%;
    max-width: 400px;
    padding: 40px 32px;
  }
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
  margin-bottom: 40px;
}

.brand h1 {
  font-size: 32px;
  font-weight: 300;
  color: rgba(255, 255, 255, 0.95);
  letter-spacing: 8px;
  text-shadow: 0 2px 8px rgba(0,0,0,0.15);
  margin-bottom: 8px;
}

.brand p {
  font-size: 14px;
  color: rgba(255, 255, 255, 0.7);
  letter-spacing: 2px;
  text-shadow: 0 1px 4px rgba(0,0,0,0.1);
}

@media (max-width: 480px) {
  .brand h1 {
    font-size: 26px;
    letter-spacing: 6px;
  }
}

.form-group {
  margin-bottom: 22px;
  opacity: 0;
  animation: fadeUp 0.6s ease-out forwards;
}

.form-group:nth-child(1) { animation-delay: 0.1s; }
.form-group:nth-child(2) { animation-delay: 0.2s; }
.form-group:nth-child(3) { animation-delay: 0.3s; }
.form-group:nth-child(4) { animation-delay: 0.4s; }

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

.options {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 28px;
  font-size: 13px;
  opacity: 0;
  animation: fadeUp 0.6s ease-out 0.3s forwards;
}

.remember {
  display: flex;
  align-items: center;
  gap: 8px;
  color: rgba(255, 255, 255, 0.7);
  cursor: pointer;
  transition: color 0.3s;
}

.remember:hover {
  color: rgba(255, 255, 255, 0.95);
}

.remember input[type="checkbox"] {
  width: 16px;
  height: 16px;
  accent-color: rgba(255,255,255,0.8);
  cursor: pointer;
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
  margin-top: 28px;
  font-size: 13px;
  color: rgba(255, 255, 255, 0.7);
  opacity: 0;
  animation: fadeUp 0.6s ease-out 0.5s forwards;
}

.footer :deep(a) {
  color: rgba(255, 255, 255, 0.95);
  text-decoration: none;
  font-weight: 500;
  margin-left: 4px;
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
