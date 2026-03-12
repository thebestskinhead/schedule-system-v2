<template>
  <div class="main-container">
    <div class="card">
      <div class="card-header">
        <span class="card-title">使用说明</span>
      </div>
      
      <div class="readme-content">
        <h2>欢迎使用排班管理系统 V2</h2>
        
        <h3>📋 功能介绍</h3>
        <el-divider />
        
        <h4 id="availability">1. 无课表管理</h4>
        <p>无课表是排班的基础数据，用于记录您的空闲时间。</p>
        <ul>
          <li><strong>手动录入</strong>：在表格中点击任意时段，选择无课的周次</li>
          <li><strong>Cookie导入</strong>：从教务系统自动抓取课程表转换为无课表</li>
          <li><strong>XLS导入</strong>：通过Excel文件批量导入课程表</li>
        </ul>
        <p class="path-tip">访问路径：<router-link to="/availability">无课表</router-link></p>
        
        <h4 id="schedule">2. 排班管理</h4>
        <p>根据成员的无课表自动生成最优排班方案。</p>
        <ul>
          <li><strong>生成排班</strong>：配置周次、值班日期、每时段人数等参数</li>
          <li><strong>预览调整</strong>：预览自动生成的排班，支持手动更换人员</li>
          <li><strong>确认排班</strong>：确认后排班生效，通知相关人员</li>
          <li><strong>排班设置</strong>：配置每时段人数、每人每日/每周最多值班次数</li>
        </ul>
        <p class="path-tip">访问路径：<router-link to="/schedule">排班管理</router-link></p>
        
        <h4 id="duty">3. 值班管理</h4>
        <p>查看个人值班安排并确认值班状态。</p>
        <ul>
          <li><strong>我的值班</strong>：查看本周及未来的值班安排</li>
          <li><strong>值班确认</strong>：确认会按时值班</li>
          <li><strong>值班完成</strong>：值班完成后标记为已完成</li>
        </ul>
        <p class="path-tip">访问路径：<router-link to="/duty/my">我的值班</router-link></p>
        
        <h4 id="duty-assignment">4. 每周分工</h4>
        <p>设置各部门本周的值班日期安排。</p>
        <ul>
          <li><strong>发布分工</strong>：为各部门勾选值班日期</li>
          <li><strong>查看预览</strong>：查看各部门分工安排卡片</li>
        </ul>
        <p class="path-tip">访问路径：<router-link to="/admin/duty-assignments">每周分工</router-link>（需权限）</p>
        
        <h4 id="users">5. 用户管理</h4>
        <p>管理系统用户账号、部门归属和角色权限。</p>
        <ul>
          <li><strong>查看用户</strong>：查看所有用户信息列表</li>
          <li><strong>编辑用户</strong>：修改用户姓名、邮箱、部门、角色</li>
          <li><strong>删除用户</strong>：删除用户账号</li>
        </ul>
        <p class="path-tip">访问路径：<router-link to="/admin/users">用户管理</router-link>（需权限）</p>
        
        <h4 id="temp-permission">6. 临时权限管理</h4>
        <p>授予用户临时的额外管理权限。</p>
        <ul>
          <li><strong>授予权限</strong>：选择用户、权限类型、资源范围、有效期</li>
          <li><strong>撤销权限</strong>：手动撤销已授予的权限</li>
          <li><strong>权限查看</strong>：查看所有临时权限状态，过期提醒</li>
        </ul>
        <p class="path-tip">访问路径：<router-link to="/admin/temp-permissions">临时权限</router-link>（需系统管理员权限）</p>
        
        <h4 id="smtp">7. SMTP邮件配置</h4>
        <p>配置邮件服务器用于发送系统通知。</p>
        <ul>
          <li><strong>密码重置</strong>：发送密码重置邮件</li>
        </ul>
        <p class="path-tip">访问路径：<router-link to="/admin/smtp">邮件配置</router-link>（需系统管理员权限）</p>
        
        <h3>🔐 权限说明</h3>
        <el-divider />
        
        <el-descriptions :column="1" border>
          <el-descriptions-item label="普通用户">
            录入无课表、查看个人值班、查看排班结果
          </el-descriptions-item>
          <el-descriptions-item label="部门管理员">
            普通用户权限 + <router-link to="/schedule">排班管理</router-link>（仅限本部门）+ <router-link to="/admin/users">用户管理</router-link>（仅限本部门）
          </el-descriptions-item>
          <el-descriptions-item label="办公室管理员">
            部门管理员权限 + <router-link to="/admin/duty-assignments">每周分工</router-link> + <router-link to="/admin/users">用户管理</router-link>（全部）+ <router-link to="/admin/semester">学期设置</router-link>
          </el-descriptions-item>
          <el-descriptions-item label="系统管理员">
            全部权限 + <router-link to="/admin/temp-permissions">临时权限管理</router-link> + <router-link to="/admin/smtp">SMTP配置</router-link>
          </el-descriptions-item>
        </el-descriptions>
        
        <h3>📅 排班流程</h3>
        <el-divider />
        <el-steps direction="vertical" :space="80" :active="5">
          <el-step title="学期初" description="管理员设置学期起始日和当前周次" />
          <el-step title="成员录入" description="所有成员录入个人无课表" />
          <el-step title="设置分工" description="办公室管理员设置各部门本周值班日期" />
          <el-step title="生成排班" description="部门管理员根据无课表生成排班方案" />
          <el-step title="确认排班" description="排班确认后正式生效" />
          <el-step title="执行值班" description="成员按时值班并确认完成状态" />
        </el-steps>
        
        <h3>📖 使用指南</h3>
        <el-divider />
        
        <h4 id="register">一、注册与登录</h4>
        <p>新用户需要先注册账号，注册后即可登录系统。</p>
        <ul>
          <li><strong>学号</strong>：请填写真实学号，作为登录凭证</li>
          <li><strong>邮箱</strong>：用于接收密码重置邮件等通知</li>
          <li><strong>密码</strong>：6-20位字符</li>
        </ul>
        
        <h4 id="forgot-password">二、忘记密码</h4>
        <p>如果忘记密码，可以通过以下步骤重置：</p>
        <ol>
          <li>在登录页面点击"忘记密码？"</li>
          <li>输入您的学号</li>
          <li>系统向您的注册邮箱发送重置链接</li>
          <li>点击邮件中的链接，设置新密码</li>
        </ol>
        <p class="path-tip">访问路径：<router-link to="/forgot-password">找回密码</router-link></p>
        
        <h4 id="availability-input">三、录入无课表</h4>
        <p>在排班前，您需要先录入自己的无课时间：</p>
        <ol>
          <li>进入<router-link to="/availability">无课表</router-link>页面</li>
          <li>点击任意时段格子</li>
          <li>在弹窗中勾选该时段无课的周次</li>
          <li>点击保存</li>
        </ol>
        <p class="tip"><strong>提示</strong>：绿色表示无课（可排班），红色表示有课。点击格子可快速切换状态。</p>
        
        <h4 id="schedule-generate">四、生成排班</h4>
        <p>部门管理员可以生成排班：</p>
        <ol>
          <li>进入<router-link to="/schedule">排班管理</router-link>页面</li>
          <li>选择周次、值班日期、节次</li>
          <li>设置每时段人数和每人限制</li>
          <li>点击"预览排班"</li>
          <li>确认无误后点击"确认排班"</li>
        </ol>
        
        <h4 id="duty-assignment-publish">五、发布每周分工</h4>
        <p>办公室管理员可以设置各部门每周的值班日期：</p>
        <ol>
          <li>进入<router-link to="/admin/duty-assignments">每周分工</router-link>页面</li>
          <li>选择要设置的周次</li>
          <li>为各部门勾选值班日期</li>
          <li>点击保存设置</li>
        </ol>
        
        <h4 id="temp-permission-grant">六、授予临时权限</h4>
        <p>系统管理员可以临时授予用户额外权限：</p>
        <ol>
          <li>进入<router-link to="/admin/temp-permissions">临时权限</router-link>页面</li>
          <li>点击"授予权限"</li>
          <li>选择用户、权限类型、资源范围</li>
          <li>设置有效期（1小时至7天）</li>
          <li>点击确定</li>
        </ol>
        
        <h3>❓ 常见问题</h3>
        <el-divider />
        
        <p><strong>Q: 如何修改无课表？</strong><br>
        A: 进入<router-link to="/availability">无课表</router-link>页面，点击相应时段重新选择周次即可。</p>
        
        <p><strong>Q: 排班时提示"无人可排"怎么办？</strong><br>
        A: 检查部门成员是否都已录入无课表，或放宽排班约束条件（如增加每时段人数）。</p>
        
        <p><strong>Q: 排班冲突怎么办？</strong><br>
        A: 在排班预览页面点击"更换"按钮手动调整人员。</p>
        
        <p><strong>Q: 忘记密码怎么办？</strong><br>
        A: 在登录页面点击"忘记密码"，通过注册邮箱重置密码。注意需要先配置SMTP邮件服务。</p>
        
        <p><strong>Q: 为什么看不到管理菜单？</strong><br>
        A: 需要相应的管理权限。请联系管理员申请权限。</p>
        
        <p><strong>Q: 如何修改部门信息？</strong><br>
        A: 联系部门管理员或办公室管理员，在用户管理中修改。</p>
        
        <p><strong>Q: 临时权限过期了怎么办？</strong><br>
        A: 需要重新由管理员授予临时权限。</p>
        
        <h3>📞 帮助与支持</h3>
        <el-divider />
        <p>如遇问题，请联系：</p>
        <ul>
          <li><strong>系统管理员</strong> - 系统配置、权限问题</li>
          <li><strong>办公室管理员</strong> - 排班协调、部门调整</li>
          <li><strong>部门管理员</strong> - 本部门相关问题</li>
        </ul>
      </div>
    </div>
  </div>
</template>

<script setup>
// 使用说明页面
</script>

<style scoped>
.readme-content {
  padding: 20px;
  line-height: 1.8;
  color: #333;
}

.readme-content h2 {
  color: #409eff;
  margin-bottom: 20px;
}

.readme-content h3 {
  margin-top: 30px;
  color: #303133;
}

.readme-content h4 {
  margin-top: 20px;
  color: #606266;
}

.readme-content ul, .readme-content ol {
  margin: 10px 0 20px 20px;
}

.readme-content li {
  margin: 8px 0;
}

.readme-content p {
  margin: 12px 0;
}

.path-tip {
  color: #409eff;
  font-size: 14px;
  margin: 8px 0 20px 0;
}

.path-tip a {
  color: #409eff;
  text-decoration: none;
}

.path-tip a:hover {
  text-decoration: underline;
}

.tip {
  color: #e6a23c;
  background: #fdf6ec;
  padding: 10px 15px;
  border-radius: 4px;
  border-left: 3px solid #e6a23c;
}
</style>
