# 排班系统移动端 - 页面风格文档

## 1. 设计原则

### 1.1 设计目标
- **触控优先**：所有交互元素适合手指操作（最小 44px 触控区域）
- **单手操作**：核心功能在屏幕下半部分可完成
- **即时反馈**：操作后立即显示结果，减少页面跳转
- **简洁高效**：常用功能 3 步内完成

### 1.2 视觉层级
- 重要信息优先展示
- 使用留白区分内容区块
- 色彩引导用户注意力

### 1.3 内容密度
- 移动端采用"卡片式"布局
- 每屏显示 4-6 个主要元素
- 避免信息过载

---

## 2. 色彩系统

### 2.1 主色调

```css
:root {
  /* 品牌色 */
  --color-primary: #1989FA;        /* 主要按钮、选中状态 */
  --color-primary-light: #E6F2FF;  /* 浅色背景 */
  --color-primary-dark: #0F7AE7;   /* 按下状态 */
  
  /* 功能色 */
  --color-success: #07C160;        /* 成功、无课状态 */
  --color-success-light: #E6F7EF;  /* 成功浅色背景 */
  
  --color-warning: #FF976A;        /* 警告、待处理 */
  --color-warning-light: #FFF3ED;  /* 警告浅色背景 */
  
  --color-danger: #EE0A24;         /* 危险、有课状态 */
  --color-danger-light: #FFEBED;   /* 危险浅色背景 */
  
  /* 中性色 */
  --color-text-primary: #323233;   /* 主要文字 */
  --color-text-secondary: #646566; /* 次要文字 */
  --color-text-tertiary: #969799;  /* 辅助文字、图标 */
  --color-text-placeholder: #C8C9CC; /* 占位符 */
  
  /* 背景色 */
  --color-bg: #F7F8FA;             /* 页面背景 */
  --color-surface: #FFFFFF;        /* 卡片背景 */
  --color-border: #EBEDF0;         /* 分割线、边框 */
}
```

### 2.2 色彩使用规范

| 颜色 | 变量 | 使用场景 |
|------|------|----------|
| 主色 | `--color-primary` | 主要按钮、选中状态、链接 |
| 成功绿 | `--color-success` | 无课状态、完成状态、成功提示 |
| 警告橙 | `--color-warning` | 待处理、提示信息 |
| 危险红 | `--color-danger` | 有课状态、删除操作、错误提示 |
| 主要文字 | `--color-text-primary` | 标题、正文 |
| 次要文字 | `--color-text-secondary` | 副标题、说明文字 |
| 辅助文字 | `--color-text-tertiary` | 时间戳、占位符、图标 |
| 页面背景 | `--color-bg` | 页面底色 |
| 卡片背景 | `--color-surface` | 卡片、弹窗、浮层 |
| 分割线 | `--color-border` | 分割线、边框 |

### 2.3 状态色映射

| 功能场景 | 正常状态 | 浅色背景 |
|----------|----------|----------|
| 无课/空闲 | `#07C160` | `#E6F7EF` |
| 有课/忙碌 | `#EE0A24` | `#FFEBED` |
| 待值班 | `#FF976A` | `#FFF3ED` |
| 已完成 | `#07C160` | `#E6F7EF` |
| 选中状态 | `#1989FA` | `#E6F2FF` |
| 默认状态 | `#969799` | `#F2F3F5` |

---

## 3. 字体系统

### 3.1 字体族

```css
--font-family-base: -apple-system, BlinkMacSystemFont, 'Helvetica Neue', 
                    Helvetica, Segoe UI, Arial, Roboto, 'PingFang SC', 
                    'miui', 'Hiragino Sans GB', 'Microsoft Yahei', sans-serif;

--font-family-mono: 'SF Mono', Monaco, 'Cascadia Code', 'Roboto Mono', 
                    Consolas, 'Courier New', monospace;
```

### 3.2 字号规范

| 级别 | 变量 | 大小 | 字重 | 行高 | 用途 |
|------|------|------|------|------|------|
| 大标题 | `--font-size-h1` | 20px | 600 | 1.4 | 页面大标题 |
| 标题 | `--font-size-h2` | 18px | 600 | 1.4 | 页面标题、卡片标题 |
| 小标题 | `--font-size-h3` | 16px | 500 | 1.4 | 区块标题 |
| 正文 | `--font-size-body` | 14px | 400 | 1.5 | 正文内容、按钮文字 |
| 辅助 | `--font-size-small` | 12px | 400 | 1.5 | 辅助说明、时间戳 |
| 标签 | `--font-size-mini` | 10px | 500 | 1.2 | 小标签、徽章 |
| 数据 | `--font-size-large` | 24px | 600 | 1.2 | 关键数据展示 |

### 3.3 字重规范

```css
--font-weight-normal: 400;   /* 正文 */
--font-weight-medium: 500;   /* 中等强调 */
--font-weight-semibold: 600; /* 标题、强调 */
--font-weight-bold: 700;     /* 特别强调 */
```

### 3.4 文字样式示例

```css
/* 页面标题 */
.page-title {
  font-size: var(--font-size-h1);
  font-weight: var(--font-weight-semibold);
  color: var(--color-text-primary);
  line-height: 1.4;
}

/* 卡片标题 */
.card-title {
  font-size: var(--font-size-h2);
  font-weight: var(--font-weight-semibold);
  color: var(--color-text-primary);
  line-height: 1.4;
}

/* 正文内容 */
.body-text {
  font-size: var(--font-size-body);
  font-weight: var(--font-weight-normal);
  color: var(--color-text-primary);
  line-height: 1.5;
}

/* 辅助说明 */
.caption-text {
  font-size: var(--font-size-small);
  font-weight: var(--font-weight-normal);
  color: var(--color-text-tertiary);
  line-height: 1.5;
}
```

---

## 4. 间距系统

### 4.1 基础单位

基础间距单位为 **4px**，所有间距基于此倍数。

```css
--space-unit: 4px;
```

### 4.2 间距变量

| 名称 | 变量 | 数值 | 用途 |
|------|------|------|------|
| 超小 | `--space-xs` | 4px | 图标与文字间距、紧凑内联间距 |
| 小 | `--space-sm` | 8px | 紧凑元素间距、小图标间距 |
| 中 | `--space-md` | 12px | 一般元素间距 |
| 默认 | `--space` | 16px | 卡片内边距、元素间距 |
| 大 | `--space-lg` | 20px | 区块内间距 |
| 超大 | `--space-xl` | 24px | 区块间距、大卡片内边距 |
| 两倍 | `--space-2xl` | 32px | 大区块分隔 |
| 三倍 | `--space-3xl` | 48px | 页面级间距 |

### 4.3 页面边距

```css
/* 水平边距 */
--page-padding-x: 16px;

/* 垂直边距 */
--page-padding-y: 12px;

/* 安全区域（适配刘海屏、手势条） */
--safe-area-top: env(safe-area-inset-top);
--safe-area-bottom: env(safe-area-inset-bottom);
--safe-area-left: env(safe-area-inset-left);
--safe-area-right: env(safe-area-inset-right);
```

### 4.4 组件间距

```css
/* 卡片 */
--card-padding: 16px;
--card-margin: 12px;
--card-border-radius: 12px;

/* 列表项 */
--list-item-padding: 16px;
--list-item-height: 48px;

/* 按钮 */
--button-padding-x: 16px;
--button-height: 44px;
--button-border-radius: 8px;

/* 输入框 */
--input-height: 48px;
--input-padding-x: 16px;
--input-border-radius: 8px;
```

---

## 5. 圆角系统

### 5.1 圆角变量

```css
--radius-none: 0;
--radius-xs: 4px;      /* 小标签、徽章 */
--radius-sm: 6px;      /* 小按钮、输入框 */
--radius-md: 8px;      /* 按钮、输入框 */
--radius-lg: 12px;     /* 卡片、弹窗 */
--radius-xl: 16px;     /* 大卡片、浮层 */
--radius-full: 9999px; /* 圆形、胶囊形状 */
```

### 5.2 圆角使用规范

| 组件 | 圆角 | 说明 |
|------|------|------|
| 页面卡片 | 12px | 主内容卡片 |
| 按钮 | 8px | 标准按钮 |
| 小按钮 | 6px | 次级操作 |
| 输入框 | 8px | 表单输入 |
| 标签/徽章 | 4px | 小标签 |
| 用户头像 | 9999px | 圆形 |
| 底部弹窗 | 16px（顶部）| 底部弹出层 |

---

## 6. 阴影系统

### 6.1 阴影变量

```css
--shadow-none: none;
--shadow-xs: 0 1px 2px rgba(0, 0, 0, 0.04);
--shadow-sm: 0 2px 4px rgba(0, 0, 0, 0.04);
--shadow-md: 0 4px 8px rgba(0, 0, 0, 0.06);
--shadow-lg: 0 8px 16px rgba(0, 0, 0, 0.08);
--shadow-xl: 0 12px 24px rgba(0, 0, 0, 0.1);
```

### 6.2 阴影使用规范

| 场景 | 阴影 | 说明 |
|------|------|------|
| 卡片默认 | `--shadow-sm` | 轻微阴影 |
| 卡片悬浮 | `--shadow-md` | 悬浮时加深 |
| 底部操作栏 | `0 -2px 8px rgba(0,0,0,0.04)` | 上阴影 |
| 顶部导航 | `0 2px 4px rgba(0,0,0,0.04)` | 下阴影 |
| 弹窗/浮层 | `--shadow-lg` | 突出显示 |
| 模态框 | `--shadow-xl` | 强阴影 |

---

## 7. 组件规范

### 7.1 卡片组件 (Card)

```
┌─────────────────────────────────────┐
│  ┌─────────────────────────────┐    │
│  │                             │    │
│  │     卡片内容区域              │    │
│  │                             │    │
│  └─────────────────────────────┘    │
└─────────────────────────────────────┘
```

**样式规范：**
```css
.card {
  background: var(--color-surface);
  border-radius: var(--radius-lg);
  padding: var(--card-padding);
  box-shadow: var(--shadow-sm);
  margin-bottom: var(--card-margin);
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: var(--space-md);
}

.card-title {
  font-size: var(--font-size-h2);
  font-weight: var(--font-weight-semibold);
  color: var(--color-text-primary);
}
```

### 7.2 节次卡片 (PeriodCard)

```
无课状态：                    有课状态：
┌─────────────────────┐     ┌─────────────────────┐
│ 🌅 1-2节      [✓]   │     │ 🌅 1-2节      [✗]   │
│ 08:00 - 09:40       │     │ 08:00 - 09:40       │
└─────────────────────┘     └─────────────────────┘

选中状态：
┌─────────────────────┐
│ 🌅 1-2节      [✓]   │  ← 蓝色边框 + 阴影
│ 08:00 - 09:40       │
└─────────────────────┘
```

**样式规范：**
```css
.period-card {
  border-radius: var(--radius-md);
  padding: var(--space-md);
  display: flex;
  justify-content: space-between;
  align-items: center;
  min-height: 64px;
}

/* 无课状态 */
.period-card--free {
  background: var(--color-success-light);
  border: 1px solid var(--color-success);
}

/* 有课状态 */
.period-card--busy {
  background: var(--color-danger-light);
  border: 1px solid var(--color-danger);
}

/* 选中状态 */
.period-card--active {
  border-color: var(--color-primary);
  box-shadow: 0 0 0 2px var(--color-primary-light);
}

.period-card__time {
  font-size: var(--font-size-body);
  color: var(--color-text-secondary);
}

.period-card__icon {
  width: 24px;
  height: 24px;
  border-radius: var(--radius-full);
  display: flex;
  align-items: center;
  justify-content: center;
}

.period-card--free .period-card__icon {
  background: var(--color-success);
  color: white;
}

.period-card--busy .period-card__icon {
  background: var(--color-danger);
  color: white;
}
```

### 7.3 按钮规范

#### 主要按钮
```css
.btn-primary {
  background: var(--color-primary);
  color: white;
  border: none;
  height: var(--button-height);
  padding: 0 var(--button-padding-x);
  border-radius: var(--button-border-radius);
  font-size: var(--font-size-body);
  font-weight: var(--font-weight-medium);
}

.btn-primary:active {
  background: var(--color-primary-dark);
}

.btn-primary:disabled {
  opacity: 0.5;
}
```

#### 次要按钮
```css
.btn-secondary {
  background: var(--color-bg);
  color: var(--color-text-primary);
  border: 1px solid var(--color-border);
}
```

#### 文字按钮
```css
.btn-text {
  background: transparent;
  color: var(--color-primary);
  padding: 0 var(--space-sm);
}
```

#### 按钮尺寸
| 尺寸 | 高度 | 内边距 | 字体大小 |
|------|------|--------|----------|
| 大 | 48px | 24px | 16px |
| 中（默认）| 44px | 16px | 14px |
| 小 | 36px | 12px | 12px |
| 迷你 | 28px | 8px | 10px |

### 7.4 输入框规范

```css
.input {
  height: var(--input-height);
  padding: 0 var(--input-padding-x);
  border: 1px solid var(--color-border);
  border-radius: var(--input-border-radius);
  font-size: var(--font-size-body);
  background: var(--color-surface);
}

.input:focus {
  border-color: var(--color-primary);
  outline: none;
}

.input::placeholder {
  color: var(--color-text-placeholder);
}

.input--error {
  border-color: var(--color-danger);
}
```

### 7.5 标签/徽章规范

```
┌────────┐  ┌────────┐  ┌────────┐  ┌────────┐
│  标签  │  │ 成功  │  │ 警告  │  │ 危险  │
└────────┘  └────────┘  └────────┘  └────────┘
```

```css
.tag {
  display: inline-flex;
  align-items: center;
  padding: 2px 8px;
  border-radius: var(--radius-xs);
  font-size: var(--font-size-mini);
  font-weight: var(--font-weight-medium);
}

.tag--default {
  background: var(--color-bg);
  color: var(--color-text-secondary);
}

.tag--success {
  background: var(--color-success-light);
  color: var(--color-success);
}

.tag--warning {
  background: var(--color-warning-light);
  color: var(--color-warning);
}

.tag--danger {
  background: var(--color-danger-light);
  color: var(--color-danger);
}

.tag--primary {
  background: var(--color-primary-light);
  color: var(--color-primary);
}
```

### 7.6 列表项规范

```css
.list-item {
  display: flex;
  align-items: center;
  padding: var(--list-item-padding);
  min-height: var(--list-item-height);
  background: var(--color-surface);
  border-bottom: 1px solid var(--color-border);
}

.list-item__icon {
  width: 24px;
  height: 24px;
  margin-right: var(--space-sm);
  color: var(--color-text-tertiary);
}

.list-item__content {
  flex: 1;
}

.list-item__title {
  font-size: var(--font-size-body);
  color: var(--color-text-primary);
}

.list-item__subtitle {
  font-size: var(--font-size-small);
  color: var(--color-text-tertiary);
  margin-top: 2px;
}

.list-item__arrow {
  color: var(--color-text-tertiary);
  margin-left: var(--space-sm);
}
```

### 7.7 底部操作栏

```css
.bottom-bar {
  position: fixed;
  bottom: 0;
  left: 0;
  right: 0;
  padding: var(--space-md) var(--page-padding-x);
  padding-bottom: calc(var(--space-md) + var(--safe-area-bottom));
  background: var(--color-surface);
  box-shadow: 0 -2px 8px rgba(0, 0, 0, 0.04);
}

.bottom-bar__button {
  width: 100%;
  height: 48px;
}
```

### 7.8 顶部导航栏

```css
.navbar {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  height: 44px;
  padding-top: var(--safe-area-top);
  background: var(--color-surface);
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 var(--page-padding-x);
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.04);
  z-index: 100;
}

.navbar__title {
  font-size: var(--font-size-h2);
  font-weight: var(--font-weight-semibold);
  color: var(--color-text-primary);
}

.navbar__left,
.navbar__right {
  width: 44px;
  height: 44px;
  display: flex;
  align-items: center;
  justify-content: center;
}
```

---

## 8. 动效规范

### 8.1 动画时长

```css
--duration-fast: 150ms;      /* 微交互 */
--duration-normal: 250ms;    /* 标准过渡 */
--duration-slow: 350ms;      /* 页面切换 */
--duration-slower: 450ms;    /* 复杂动画 */
```

### 8.2 缓动函数

```css
--ease-default: cubic-bezier(0.4, 0, 0.2, 1);
--ease-in: cubic-bezier(0.4, 0, 1, 1);
--ease-out: cubic-bezier(0, 0, 0.2, 1);
--ease-bounce: cubic-bezier(0.68, -0.55, 0.265, 1.55);
```

### 8.3 过渡动画

| 场景 | 动画 | 时长 | 缓动 |
|------|------|------|------|
| 页面进入 | 从右向左滑入 + 淡入 | 300ms | ease-out |
| 页面退出 | 从左向右滑出 + 淡出 | 250ms | ease-in |
| 弹窗出现 | 从底部滑入 + 淡入 | 250ms | ease-out |
| 弹窗消失 | 向底部滑出 + 淡出 | 200ms | ease-in |
| 状态切换 | 背景色渐变 | 200ms | ease-default |
| 列表加载 | 骨架屏渐显 | 300ms | ease-default |
| 按钮点击 | 缩放 0.95 | 100ms | ease-default |
| 悬浮效果 | 阴影加深 + 上移 2px | 200ms | ease-default |

### 8.4 动画代码示例

```css
/* 页面切换 - 进入 */
.page-enter {
  animation: pageEnter var(--duration-slow) var(--ease-out);
}

@keyframes pageEnter {
  from {
    opacity: 0;
    transform: translateX(100%);
  }
  to {
    opacity: 1;
    transform: translateX(0);
  }
}

/* 页面切换 - 退出 */
.page-leave {
  animation: pageLeave var(--duration-normal) var(--ease-in);
}

@keyframes pageLeave {
  from {
    opacity: 1;
    transform: translateX(0);
  }
  to {
    opacity: 0;
    transform: translateX(-30%);
  }
}

/* 按钮点击反馈 */
.btn:active {
  transform: scale(0.95);
  transition: transform var(--duration-fast) var(--ease-default);
}

/* 状态切换 */
.status-transition {
  transition: background-color var(--duration-normal) var(--ease-default),
              border-color var(--duration-normal) var(--ease-default);
}

/* 悬浮效果 */
.card-hover:hover {
  transform: translateY(-2px);
  box-shadow: var(--shadow-md);
  transition: all var(--duration-normal) var(--ease-default);
}
```

### 8.5 反馈动画

```css
/* 成功反馈 - 轻微缩放 */
@keyframes successPulse {
  0% { transform: scale(1); }
  50% { transform: scale(1.1); }
  100% { transform: scale(1); }
}

.success-feedback {
  animation: successPulse var(--duration-normal) var(--ease-bounce);
}

/* 错误反馈 - 抖动 */
@keyframes shake {
  0%, 100% { transform: translateX(0); }
  25% { transform: translateX(-5px); }
  75% { transform: translateX(5px); }
}

.error-feedback {
  animation: shake var(--duration-normal) var(--ease-default);
}

/* 加载旋转 */
@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

.loading {
  animation: spin 1s linear infinite;
}
```

---

## 9. 图标规范

### 9.1 图标尺寸

| 尺寸 | 数值 | 用途 |
|------|------|------|
| 小 | 16px | 内联图标、标签 |
| 中 | 20px | 按钮图标、列表项 |
| 大 | 24px | 导航图标、工具栏 |
| 超大 | 32px | 空状态、大按钮 |
| 巨型 | 48px | 首页入口、强调 |

### 9.2 图标颜色

```css
.icon--primary { color: var(--color-primary); }
.icon--success { color: var(--color-success); }
.icon--warning { color: var(--color-warning); }
.icon--danger { color: var(--color-danger); }
.icon--default { color: var(--color-text-tertiary); }
.icon--white { color: white; }
```

### 9.3 图标库

使用 **Vant Icons** 图标库，保持一致性。

常用图标：
- 导航：home-o, home, calendar-o, calendar, user-o, user
- 操作：arrow-left, cross, plus, check, edit, delete-o
- 状态：success, fail, warning-o, info-o
- 功能：search, setting-o, bell, share-o

---

## 10. 布局规范

### 10.1 页面结构

```
┌─────────────────────────┐  ← 顶部导航栏 (固定)
│ Navbar                  │
├─────────────────────────┤
│                         │
│                         │
│     Content Area        │  ← 内容区域 (可滚动)
│                         │
│                         │
├─────────────────────────┤
│ Bottom Bar (可选)        │  ← 底部操作栏 (固定)
└─────────────────────────┘
│ TabBar                  │  ← 底部导航 (固定)
└─────────────────────────┘
```

### 10.2 安全区域适配

```css
/* 顶部安全区域 */
.safe-top {
  padding-top: constant(safe-area-inset-top);
  padding-top: env(safe-area-inset-top);
}

/* 底部安全区域 */
.safe-bottom {
  padding-bottom: constant(safe-area-inset-bottom);
  padding-bottom: env(safe-area-inset-bottom);
}

/* 完整页面容器 */
.page-container {
  min-height: 100vh;
  padding-top: calc(44px + env(safe-area-inset-top));
  padding-bottom: calc(50px + env(safe-area-inset-bottom));
  background: var(--color-bg);
}
```

### 10.3 响应式断点

本项目为移动端专供，仅支持移动端尺寸，但考虑不同屏幕大小：

```css
/* 小屏手机 */
@media (max-width: 320px) {
  --page-padding-x: 12px;
  --font-size-body: 13px;
}

/* 标准手机 */
@media (min-width: 321px) and (max-width: 414px) {
  --page-padding-x: 16px;
}

/* 大屏手机 */
@media (min-width: 415px) {
  --page-padding-x: 20px;
  --card-max-width: 400px;
}
```

---

## 11. 无障碍规范

### 11.1 触控区域

- 所有可点击元素最小尺寸：**44px × 44px**
- 相邻可点击元素间距：**≥ 8px**

### 11.2 对比度

- 文字与背景对比度：**≥ 4.5:1**
- 大文字对比度：**≥ 3:1**
- 界面组件对比度：**≥ 3:1**

### 11.3 焦点状态

```css
.focusable:focus {
  outline: 2px solid var(--color-primary);
  outline-offset: 2px;
}

.focusable:focus-visible {
  outline: 2px solid var(--color-primary);
  outline-offset: 2px;
}
```

---

## 12. CSS 变量汇总

```css
:root {
  /* 色彩 */
  --color-primary: #1989FA;
  --color-primary-light: #E6F2FF;
  --color-primary-dark: #0F7AE7;
  --color-success: #07C160;
  --color-success-light: #E6F7EF;
  --color-warning: #FF976A;
  --color-warning-light: #FFF3ED;
  --color-danger: #EE0A24;
  --color-danger-light: #FFEBED;
  --color-text-primary: #323233;
  --color-text-secondary: #646566;
  --color-text-tertiary: #969799;
  --color-text-placeholder: #C8C9CC;
  --color-bg: #F7F8FA;
  --color-surface: #FFFFFF;
  --color-border: #EBEDF0;

  /* 字体 */
  --font-family-base: -apple-system, BlinkMacSystemFont, 'Helvetica Neue', 
                      Helvetica, Segoe UI, Arial, Roboto, 'PingFang SC', 
                      'miui', 'Hiragino Sans GB', 'Microsoft Yahei', sans-serif;
  --font-size-h1: 20px;
  --font-size-h2: 18px;
  --font-size-h3: 16px;
  --font-size-body: 14px;
  --font-size-small: 12px;
  --font-size-mini: 10px;
  --font-size-large: 24px;
  --font-weight-normal: 400;
  --font-weight-medium: 500;
  --font-weight-semibold: 600;
  --font-weight-bold: 700;

  /* 间距 */
  --space-unit: 4px;
  --space-xs: 4px;
  --space-sm: 8px;
  --space-md: 12px;
  --space: 16px;
  --space-lg: 20px;
  --space-xl: 24px;
  --space-2xl: 32px;
  --space-3xl: 48px;
  --page-padding-x: 16px;
  --page-padding-y: 12px;

  /* 圆角 */
  --radius-none: 0;
  --radius-xs: 4px;
  --radius-sm: 6px;
  --radius-md: 8px;
  --radius-lg: 12px;
  --radius-xl: 16px;
  --radius-full: 9999px;

  /* 阴影 */
  --shadow-xs: 0 1px 2px rgba(0, 0, 0, 0.04);
  --shadow-sm: 0 2px 4px rgba(0, 0, 0, 0.04);
  --shadow-md: 0 4px 8px rgba(0, 0, 0, 0.06);
  --shadow-lg: 0 8px 16px rgba(0, 0, 0, 0.08);
  --shadow-xl: 0 12px 24px rgba(0, 0, 0, 0.1);

  /* 组件 */
  --card-padding: 16px;
  --card-margin: 12px;
  --button-height: 44px;
  --button-padding-x: 16px;
  --input-height: 48px;
  --input-padding-x: 16px;
  --list-item-height: 48px;
  --list-item-padding: 16px;

  /* 动画 */
  --duration-fast: 150ms;
  --duration-normal: 250ms;
  --duration-slow: 350ms;
  --duration-slower: 450ms;
  --ease-default: cubic-bezier(0.4, 0, 0.2, 1);
  --ease-in: cubic-bezier(0.4, 0, 1, 1);
  --ease-out: cubic-bezier(0, 0, 0.2, 1);
  --ease-bounce: cubic-bezier(0.68, -0.55, 0.265, 1.55);

  /* 安全区域 */
  --safe-area-top: env(safe-area-inset-top);
  --safe-area-bottom: env(safe-area-inset-bottom);
  --safe-area-left: env(safe-area-inset-left);
  --safe-area-right: env(safe-area-inset-right);
}
```

---

**文档版本**：v1.0  
**更新日期**：2024-03-26  
**设计系统**：基于 Vant 4 Design System 定制
