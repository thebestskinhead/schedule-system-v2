package update

// ProgressPageHTML 安装进度监控页面（单文件、零外部依赖，内嵌于 update.exe）。
// 页面通过轮询 /api/state 展示安装进度；安装阶段不可终止，失败后可重试或自动恢复。
const ProgressPageHTML = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>正在安装更新 - 排班管理系统</title>
<style>
  * { margin: 0; padding: 0; box-sizing: border-box; }
  body {
    font-family: -apple-system, "Segoe UI", "Microsoft YaHei", sans-serif;
    background: #f0f2f5; color: #303133; min-height: 100vh;
    display: flex; align-items: center; justify-content: center; padding: 24px;
  }
  .card {
    background: #fff; border-radius: 12px; width: 100%; max-width: 640px;
    box-shadow: 0 4px 24px rgba(0,0,0,.08); padding: 32px;
  }
  h1 { font-size: 20px; margin-bottom: 4px; display: flex; align-items: center; gap: 8px; }
  .sub { color: #909399; font-size: 13px; margin-bottom: 24px; }
  .versions {
    display: flex; align-items: center; justify-content: center; gap: 16px;
    background: #fafafa; border-radius: 8px; padding: 16px; margin-bottom: 24px;
  }
  .ver { text-align: center; }
  .ver .label { font-size: 12px; color: #909399; margin-bottom: 4px; }
  .ver .num { font-size: 22px; font-weight: 600; font-variant-numeric: tabular-nums; }
  .ver.target .num { color: #409eff; }
  .arrow { color: #c0c4cc; font-size: 20px; }
  .steps { margin-bottom: 20px; }
  .step { display: flex; align-items: center; gap: 10px; padding: 7px 0; color: #909399; font-size: 14px; }
  .step .dot {
    width: 20px; height: 20px; border-radius: 50%; border: 2px solid #dcdfe6;
    display: inline-flex; align-items: center; justify-content: center;
    font-size: 11px; flex-shrink: 0; background: #fff;
  }
  .step.active { color: #409eff; font-weight: 500; }
  .step.active .dot { border-color: #409eff; }
  .step.done { color: #67c23a; }
  .step.done .dot { border-color: #67c23a; background: #67c23a; color: #fff; }
  .step.failed .dot { border-color: #f56c6c; background: #f56c6c; color: #fff; }
  .step.failed { color: #f56c6c; }
  .progress-wrap { margin-bottom: 8px; }
  .progress-bar { height: 10px; background: #ebeef5; border-radius: 5px; overflow: hidden; }
  .progress-bar .fill {
    height: 100%; background: linear-gradient(90deg, #409eff, #66b1ff);
    border-radius: 5px; transition: width .3s; width: 0%;
  }
  .progress-info { display: flex; justify-content: space-between; font-size: 12px; color: #909399; margin-top: 6px; }
  .msg { font-size: 14px; margin: 12px 0; min-height: 20px; }
  .msg.error { color: #f56c6c; }
  .msg.success { color: #67c23a; font-weight: 600; }
  .logs {
    background: #1e1e2e; color: #a9b1d6; border-radius: 8px; padding: 12px 14px;
    font-family: Consolas, "Courier New", monospace; font-size: 12px; line-height: 1.7;
    height: 168px; overflow-y: auto; margin-bottom: 20px; white-space: pre-wrap; word-break: break-all;
  }
  .logs .t { color: #565f89; }
  .actions { display: flex; gap: 12px; justify-content: flex-end; }
  button {
    border: none; border-radius: 6px; padding: 9px 18px; font-size: 14px; cursor: pointer;
    transition: opacity .2s;
  }
  button:disabled { opacity: .5; cursor: not-allowed; }
  .btn-primary { background: #409eff; color: #fff; }
  .btn-default { background: #fff; color: #606266; border: 1px solid #dcdfe6; }
  .banner {
    display: none; background: #fdf6ec; color: #e6a23c; border: 1px solid #faecd8;
    border-radius: 6px; padding: 10px 14px; font-size: 13px; margin-bottom: 16px;
  }
  .spinner {
    display: inline-block; width: 14px; height: 14px; border: 2px solid #409eff;
    border-top-color: transparent; border-radius: 50%; animation: spin .8s linear infinite;
  }
  @keyframes spin { to { transform: rotate(360deg); } }
</style>
</head>
<body>
<div class="card">
  <h1><span class="spinner" id="spin"></span> 正在安装更新</h1>
  <div class="sub">更新程序正在后台工作，请保持本页面打开。安装期间服务将短暂中断。</div>

  <div class="banner" id="banner"></div>

  <div class="versions">
    <div class="ver"><div class="label">当前版本</div><div class="num" id="cur">-</div></div>
    <div class="arrow">&#10132;</div>
    <div class="ver target"><div class="label">目标版本</div><div class="num" id="target">-</div></div>
  </div>

  <div class="steps" id="steps">
    <div class="step" data-step="verifying-pkg"><span class="dot">1</span>校验安装包</div>
    <div class="step" data-step="stopping"><span class="dot">2</span>停止服务</div>
    <div class="step" data-step="backing-up"><span class="dot">3</span>备份当前版本</div>
    <div class="step" data-step="installing"><span class="dot">4</span>应用新版本</div>
    <div class="step" data-step="restarting"><span class="dot">5</span>重启服务</div>
    <div class="step" data-step="verifying"><span class="dot">6</span>验证服务</div>
  </div>

  <div class="progress-wrap">
    <div class="progress-bar"><div class="fill" id="fill"></div></div>
    <div class="progress-info">
      <span id="pct">0%</span>
      <span id="bytes"></span>
    </div>
  </div>

  <div class="msg" id="msg">正在准备…</div>
  <div class="logs" id="logs"></div>

  <div class="actions">
    <button class="btn-primary" id="btnDone" style="display:none" onclick="location.href='/'">进入系统</button>
    <button class="btn-default" id="btnRetry" style="display:none" onclick="retry()">重试更新</button>
  </div>
</div>

<script>
var STEP_ORDER = ['verifying-pkg','stopping','backing-up','installing','restarting','verifying'];
// 进入这些阶段后页面将自动刷新（服务端口已交还给 server，刷新后即进入系统）
var RELOAD_PHASES = ['done','restarting','verifying','rolled-back'];
var logCount = 0;
var reloadTimer = null;

function fmtBytes(n) {
  if (!n && n !== 0) return '';
  if (n < 1024) return n + ' B';
  if (n < 1048576) return (n/1024).toFixed(1) + ' KB';
  if (n < 1073741824) return (n/1048576).toFixed(2) + ' MB';
  return (n/1073741824).toFixed(2) + ' GB';
}

function setStep(phase) {
  var idx = STEP_ORDER.indexOf(phase);
  document.querySelectorAll('.step').forEach(function(el) {
    var s = el.getAttribute('data-step');
    el.classList.remove('active','done','failed');
    if (phase === 'failed' || phase === 'rolled-back') {
      if (idx >= 0 && STEP_ORDER.indexOf(s) < idx) el.classList.add('done');
    } else if (idx >= 0) {
      if (STEP_ORDER.indexOf(s) < idx) el.classList.add('done');
      else if (s === phase) el.classList.add('active');
    }
  });
}

function render(st) {
  document.getElementById('cur').textContent = st.currentVersion || '-';
  document.getElementById('target').textContent = st.targetVersion || '-';
  document.getElementById('fill').style.width = (st.progress || 0) + '%';
  document.getElementById('pct').textContent = (st.progress || 0) + '%';
  document.getElementById('bytes').textContent =
    (st.total > 0) ? (fmtBytes(st.received) + ' / ' + fmtBytes(st.total)) : '';
  document.getElementById('msg').textContent = st.message || '';
  document.getElementById('msg').className = 'msg' +
    (st.phase === 'failed' ? ' error' : '') + (st.phase === 'done' ? ' success' : '');
  setStep(st.phase);

  var spin = document.getElementById('spin');
  spin.style.display = (st.phase === 'done' || st.phase === 'failed' || st.phase === 'rolled-back') ? 'none' : 'inline-block';

  var logsEl = document.getElementById('logs');
  if (st.logs && st.logs.length !== logCount) {
    logsEl.innerHTML = st.logs.map(function(l) {
      var t = l.substring(0, 12);
      return '<span class="t">[' + t + ']</span> ' + l.substring(13);
    }).join('\n');
    logsEl.scrollTop = logsEl.scrollHeight;
    logCount = st.logs.length;
  }

  document.getElementById('btnRetry').style.display = (st.phase === 'failed') ? 'inline-block' : 'none';
  document.getElementById('btnDone').style.display = (st.phase === 'done' || st.phase === 'rolled-back') ? 'inline-block' : 'none';

  // 端口已交还给 server：自动刷新，服务恢复后即进入系统
  if (RELOAD_PHASES.indexOf(st.phase) >= 0 && !reloadTimer) {
    reloadTimer = setInterval(function() { location.reload(); }, 1500);
  } else if (RELOAD_PHASES.indexOf(st.phase) < 0 && reloadTimer) {
    clearInterval(reloadTimer); reloadTimer = null;
  }

  if (st.phase === 'done') {
    document.getElementById('banner').style.display = 'block';
    document.getElementById('banner').textContent = '更新完成，服务正在恢复，页面即将自动刷新进入系统…';
  } else if (st.phase === 'failed') {
    document.getElementById('banner').style.display = 'block';
    document.getElementById('banner').textContent = '安装失败：' + (st.error || '未知错误') + '。可点击「重试更新」，或等待 ' + (st.autoAbortIn || 60) + ' 秒后自动恢复更新前的版本。';
  } else if (st.phase === 'rolled-back') {
    document.getElementById('banner').style.display = 'block';
    document.getElementById('banner').textContent = '安装未成功，服务正在恢复为更新前的版本，页面即将自动刷新…';
  }
}

function poll() {
  fetch('/api/state').then(function(r) { return r.json(); }).then(render).catch(function() {});
}

function retry() {
  fetch('/api/retry', { method: 'POST' }).then(function() {
    logCount = 0;
  }).catch(function() {});
}

poll();
setInterval(poll, 800);
</script>
</body>
</html>`
