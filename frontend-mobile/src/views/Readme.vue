<template>
  <div class="main-container">
    <Navbar title="使用说明" show-back />
    <div class="card">
      <div class="card-header">
        <span class="card-title">使用说明</span>
      </div>

      <div class="readme-content">
        <h2>欢迎使用排班管理系统 V2</h2>

        <h3>📋 功能介绍</h3>
        <van-divider />

        <h4 id="availability">1. 无课表管理</h4>
        <p>无课表是排班的基础数据，用于记录您的空闲时间。表格覆盖周一至周五，共四个大节。</p>
        <ul>
          <li><strong>手动录入</strong>：在表格中点击任意时段，选择该时段无课的周次</li>
          <li><strong>教务系统导入</strong>：从教务系统自动抓取课程表转换为无课表</li>
          <li><strong>XLS导入</strong>：通过Excel文件批量导入课程表</li>
        </ul>
        <p class="path-tip">访问路径：<router-link to="/availability">无课表</router-link></p>

        <h4 id="schedule">2. 排班管理</h4>
        <p>根据成员的无课表自动生成最优排班方案。</p>
        <ul>
          <li><strong>生成排班</strong>：配置周次、值班日期、节次等参数</li>
          <li><strong>预览调整</strong>：预览自动生成的排班，支持手动添加或移除人员</li>
          <li><strong>确认排班</strong>：确认后排班生效</li>
          <li><strong>排班设置</strong>：配置每时段人数、每人每日/每周最多值班次数</li>
          <li><strong>导出模板</strong>：配置独立的导出模板，导出为列表或课表格式</li>
        </ul>
        <p class="path-tip">访问路径：<router-link to="/schedule">排班管理</router-link></p>

        <h4 id="duty">3. 值班管理</h4>
        <p>查看个人值班安排及状态。</p>
        <ul>
          <li><strong>我的值班</strong>：查看自己的值班安排与状态</li>
          <li><strong>值班状态</strong>：待确认 / 已确认 / 已完成 / 已取消</li>
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
          <li><strong>录入成员</strong>：录入正式成员，支持单个录入与批量录入，系统自动发送激活邮件</li>
          <li><strong>查看用户</strong>：查看用户信息列表，可按待入职/在职/离职筛选</li>
          <li><strong>编辑用户</strong>：修改用户姓名、邮箱、部门、角色</li>
          <li><strong>离职处理</strong>：标记成员离职，或批量离职</li>
        </ul>
        <p class="path-tip">访问路径：<router-link to="/admin/users">用户管理</router-link>（需权限）</p>

        <h4 id="temp-permission">6. 临时权限管理</h4>
        <p>支持权限申请审批流程，普通用户可申请临时管理权限。</p>
        <ul>
          <li><strong>申请权限</strong>：普通用户选择需要的权限类型，填写申请原因提交</li>
          <li><strong>审批申请</strong>：部门管理员在「待我审批」中审批权限申请</li>
          <li><strong>授予权限</strong>：管理员可直接授予用户临时权限</li>
          <li><strong>撤销权限</strong>：手动撤销已授予的权限</li>
          <li><strong>权限查看</strong>：查看所有临时权限状态</li>
        </ul>
        <p class="path-tip">访问路径：<router-link to="/admin/temp-permissions">权限申请</router-link>（所有用户可访问）</p>

        <h4 id="smtp">7. 邮件配置</h4>
        <p>配置邮件服务器用于发送系统通知，并设置网站域名。</p>
        <ul>
          <li><strong>SMTP配置</strong>：设置邮件服务器参数</li>
          <li><strong>网站域名</strong>：用于生成密码重置与激活邮件中的链接</li>
        </ul>
        <p class="path-tip">访问路径：<router-link to="/admin/smtp">邮件配置</router-link>（需系统管理员权限）</p>

        <h4 id="semester">8. 学期设置</h4>
        <p>设置学期起始日，系统据此自动计算当前周次。</p>
        <p class="path-tip">访问路径：<router-link to="/admin/semester">学期设置</router-link>（需权限）</p>

        <h3>🔐 权限说明</h3>
        <van-divider />

        <van-cell-group inset>
          <van-cell title="普通用户" label="录入无课表、查看个人值班、查看排班结果、申请临时权限" is-link />
          <van-cell title="部门管理员" is-link>
            <template #label>
              <span>普通用户权限 + 排班管理（仅限本部门）+ 用户管理（仅限本部门）+ 审批权限申请</span>
            </template>
          </van-cell>
          <van-cell title="办公室管理员" is-link>
            <template #label>
              <span>部门管理员权限 + 每周分工 + 用户管理（全部）+ 学期设置</span>
            </template>
          </van-cell>
          <van-cell title="系统管理员" is-link>
            <template #label>
              <span>全部权限 + 临时权限管理 + 邮件配置</span>
            </template>
          </van-cell>
        </van-cell-group>

        <h3>📅 排班流程</h3>
        <van-divider />
        <van-steps direction="vertical" :active="5">
          <van-step>
            <h4>学期初</h4>
            <p>管理员设置学期起始日和当前周次</p>
          </van-step>
          <van-step>
            <h4>成员录入</h4>
            <p>所有成员录入个人无课表</p>
          </van-step>
          <van-step>
            <h4>设置分工</h4>
            <p>办公室管理员设置各部门本周值班日期</p>
          </van-step>
          <van-step>
            <h4>生成排班</h4>
            <p>部门管理员根据无课表生成排班方案</p>
          </van-step>
          <van-step>
            <h4>确认排班</h4>
            <p>排班确认后正式生效</p>
          </van-step>
          <van-step>
            <h4>执行值班</h4>
            <p>成员按时值班并确认完成状态</p>
          </van-step>
        </van-steps>

        <h3>📖 使用指南</h3>
        <van-divider />

        <h4 id="register">一、激活与登录</h4>
        <p>账号由管理员统一录入，系统不开放自助注册。管理员录入正式成员后，系统会向该成员的邮箱发送激活邮件，成员点击邮件中的链接设置密码，账号方可使用。</p>
        <ul>
          <li><strong>学号</strong>：作为登录凭证</li>
          <li><strong>邮箱</strong>：用于接收激活邮件、密码重置邮件等通知</li>
          <li><strong>密码</strong>：6-20位字符</li>
        </ul>

        <h4 id="forgot-password">二、忘记密码</h4>
        <p>如果忘记密码，可以通过以下步骤重置：</p>
        <ol>
          <li>在登录页面点击"忘记密码？"</li>
          <li>输入您的学号</li>
          <li>系统向您的邮箱发送重置链接</li>
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
        <p class="tip"><strong>提示</strong>：绿色表示无课（可排班），红色表示有课。也可以通过教务系统导入或 XLS 导入快速生成。</p>

        <h4 id="schedule-generate">四、生成排班</h4>
        <p>部门管理员可以生成排班：</p>
        <ol>
          <li>进入<router-link to="/schedule">排班管理</router-link>页面</li>
          <li>选择周次、值班日期、节次</li>
          <li>设置每时段人数和每人限制</li>
          <li>点击"生成预览"</li>
          <li>在预览中调整人员，确认无误后点击"确认排班"</li>
        </ol>
        <p class="tip"><strong>提示</strong>：若本周已发布分工，可点击"应用分工日期"直接填入值班日期。</p>

        <h4 id="duty-assignment-publish">五、发布每周分工</h4>
        <p>办公室管理员可以设置各部门每周的值班日期：</p>
        <ol>
          <li>进入<router-link to="/admin/duty-assignments">每周分工</router-link>页面</li>
          <li>选择要设置的周次</li>
          <li>为各部门勾选值班日期</li>
          <li>点击保存设置</li>
        </ol>

        <h4 id="temp-permission-apply">六、申请临时权限</h4>
        <p>普通用户可以申请临时管理权限：</p>
        <ol>
          <li>进入<router-link to="/admin/temp-permissions">权限申请</router-link>页面</li>
          <li>点击"申请权限"按钮</li>
          <li>选择申请类型：临时权限 或 请假</li>
          <li>申请临时权限时选择权限类型并填写期望到期日；申请请假时选择请假类型和起止日期</li>
          <li>填写申请原因，提交申请，等待审批</li>
        </ol>
        <p class="tip"><strong>提示</strong>：申请全局权限时会发送给办公室管理员或系统管理员审批，其余申请发送给本部门管理员审批。</p>

        <h4 id="temp-permission-approve">七、审批权限申请</h4>
        <p>管理员可以审批成员的权限申请：</p>
        <ol>
          <li>进入<router-link to="/admin/temp-permissions">权限申请</router-link>页面</li>
          <li>切换到"待我审批"标签页</li>
          <li>查看待审批的申请列表</li>
          <li>点击"通过"或"拒绝"处理申请</li>
          <li>审批通过后，申请人自动获得临时权限</li>
        </ol>

        <h4 id="temp-permission-grant">八、授予临时权限（管理员）</h4>
        <p>管理员可以直接授予用户临时权限：</p>
        <ol>
          <li>进入<router-link to="/admin/temp-permissions">权限申请</router-link>页面，切换到"权限管理"标签页</li>
          <li>点击"授予权限"</li>
          <li>选择用户、权限类型</li>
          <li>设置到期日期（留空表示永久有效）</li>
          <li>点击确定</li>
        </ol>
        <p class="tip"><strong>提示</strong>：授予权限时，被授予的权限必须包含在您自身的权限范围内。</p>

        <h3>❓ 常见问题</h3>
        <van-divider />

        <p><strong>Q: 如何修改无课表？</strong><br>
        A: 进入<router-link to="/availability">无课表</router-link>页面，点击相应时段重新选择周次即可。</p>

        <p><strong>Q: 排班时提示"人数不足"怎么办？</strong><br>
        A: 检查部门成员是否都已录入无课表，或放宽排班约束条件（如增加每时段人数），也可在预览中手动添加人员。</p>

        <p><strong>Q: 排班冲突怎么办？</strong><br>
        A: 在排班预览页面点击人员标签的关闭按钮移除，再点击"+ 添加"补充其他人员。</p>

        <p><strong>Q: 忘记密码怎么办？</strong><br>
        A: 在登录页面点击"忘记密码"，通过邮箱重置密码。注意需要先配置邮件服务。</p>

        <p><strong>Q: 为什么看不到管理菜单？</strong><br>
        A: 需要相应的管理权限。请联系管理员申请权限。</p>

        <p><strong>Q: 如何修改部门信息？</strong><br>
        A: 联系部门管理员或办公室管理员，在用户管理中修改。</p>

        <p><strong>Q: 临时权限过期了怎么办？</strong><br>
        A: 可以重新提交权限申请，或联系管理员授予临时权限。</p>

        <p><strong>Q: 如何申请临时权限？</strong><br>
        A: 进入<router-link to="/admin/temp-permissions">权限申请</router-link>页面，点击"申请权限"按钮，填写申请信息后提交。</p>

        <p><strong>Q: 可以撤销自己提交的申请吗？</strong><br>
        A: 可以，在"我的申请"中找到待审批的申请，点击"撤销"按钮即可。</p>

        <h3>📞 帮助与支持</h3>
        <van-divider />
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
import Navbar from '../components/Navbar.vue'
// 使用说明页面
</script>

<style scoped>
.readme-content {
  padding: calc(20px + env(safe-area-inset-top)) 20px calc(20px + env(safe-area-inset-bottom));
  line-height: 1.8;
  color: #333;
  min-height: 100%;
  background: #f7f8fa;
}

.readme-content h2 {
  color: #1989fa;
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
  color: #1989fa;
  font-size: 14px;
  margin: 8px 0 20px 0;
}

.path-tip a {
  color: #1989fa;
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

:deep(.van-step__title) h4 {
  margin: 0;
  font-size: 14px;
  color: #323233;
}

:deep(.van-step__title) p {
  margin: 4px 0 0;
  font-size: 12px;
  color: #969799;
}

:deep(.van-cell-group--inset) {
  margin: 12px 0;
}
</style>
