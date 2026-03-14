/**
 * 爬虫导入 Mock Handler
 */

import { repositories } from '../database/index.js'
import { success, error } from '../utils/response.js'
import { delay } from '../utils/delay.js'

// 导入任务队列
let crawlerTasks = []
let taskIdCounter = 1

// 模拟课表数据模板
const courseTemplates = [
  { name: '高等数学', teacher: '张教授', location: '教学楼A101' },
  { name: '大学英语', teacher: '李老师', location: '教学楼B203' },
  { name: '计算机基础', teacher: '王讲师', location: '机房C301' },
  { name: '数据结构', teacher: '赵教授', location: '教学楼A205' },
  { name: '线性代数', teacher: '陈副教授', location: '教学楼B102' },
  { name: '物理实验', teacher: '刘老师', location: '实验楼D401' }
]

export const crawlerHandlers = {
  // 爬虫导入（带Cookie）
  async importCrawler(data, user) {
    await delay(1500)
    
    try {
      const { cookie, student_id, password, week_range } = data
      
      if (!cookie && !password) {
        return error('请提供Cookie或密码')
      }
      
      // 创建导入任务
      const taskId = taskIdCounter++
      const task = {
        id: taskId,
        task_id: `crawler_${taskId}`,
        type: 'crawler_import',
        status: 'processing',
        user_id: user.id,
        progress: 0,
        stage: '正在登录教务系统...',
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString()
      }
      crawlerTasks.push(task)
      
      // 模拟异步处理过程
      simulateCrawlerProcess(task, user.id, week_range)
      
      return success({
        task_id: task.task_id,
        message: '导入任务已创建'
      })
    } catch (err) {
      return error(err.message)
    }
  },

  // 预览课表
  async previewCrawler(data, user) {
    await delay(1200)
    
    try {
      const { cookie, student_id, password, week } = data
      
      if (!cookie && !password) {
        return error('请提供Cookie或密码')
      }
      
      // 生成模拟预览数据
      const previewData = generateMockSchedule(week, user.id)
      
      return success({
        week: week || 1,
        preview: previewData,
        total_courses: previewData.length,
        message: '预览数据生成成功（模拟数据）'
      })
    } catch (err) {
      return error(err.message)
    }
  },

  // 获取导入任务状态
  async getCrawlerStatus(query) {
    await delay(100)
    
    const { task_id } = query
    const task = crawlerTasks.find(t => t.task_id === task_id)
    
    if (!task) {
      return error('任务不存在')
    }
    
    return success({
      task_id: task.task_id,
      status: task.status,
      progress: task.progress,
      stage: task.stage,
      result: task.result
    })
  },

  // 获取导入任务列表
  async getCrawlerTaskList(user) {
    await delay(150)
    
    const tasks = crawlerTasks
      .filter(t => t.user_id === user.id)
      .sort((a, b) => new Date(b.created_at) - new Date(a.created_at))
      .slice(0, 10)
    
    return success(tasks)
  }
}

// 模拟爬虫处理过程
function simulateCrawlerProcess(task, userId, weekRange) {
  const stages = [
    { progress: 10, stage: '正在登录教务系统...', delay: 1000 },
    { progress: 30, stage: '登录成功，正在获取课表页面...', delay: 1500 },
    { progress: 50, stage: '正在解析课表数据...', delay: 2000 },
    { progress: 70, stage: '正在处理无课表信息...', delay: 1500 },
    { progress: 90, stage: '正在保存数据...', delay: 1000 },
    { progress: 100, stage: '导入完成', delay: 500 }
  ]
  
  let currentStage = 0
  
  function nextStage() {
    if (currentStage >= stages.length) {
      // 完成
      task.status = 'completed'
      task.progress = 100
      task.stage = '导入完成'
      task.result = {
        success: true,
        imported_weeks: weekRange ? (weekRange[1] - weekRange[0] + 1) : 4,
        imported_courses: Math.floor(Math.random() * 10) + 15,
        failed_count: 0
      }
      task.updated_at = new Date().toISOString()
      
      // 生成无课表数据
      generateMockAvailability(userId, weekRange)
      return
    }
    
    const stage = stages[currentStage]
    task.progress = stage.progress
    task.stage = stage.stage
    task.updated_at = new Date().toISOString()
    
    currentStage++
    setTimeout(nextStage, stage.delay)
  }
  
  // 开始处理
  setTimeout(nextStage, 500)
}

// 生成模拟课表预览
function generateMockSchedule(week, userId) {
  const courses = []
  const weekdays = [1, 2, 3, 4, 5]
  const periods = [1, 2, 3, 4]
  
  // 随机生成5-8节课
  const courseCount = Math.floor(Math.random() * 4) + 5
  
  for (let i = 0; i < courseCount; i++) {
    const template = courseTemplates[Math.floor(Math.random() * courseTemplates.length)]
    const weekday = weekdays[Math.floor(Math.random() * weekdays.length)]
    const period = periods[Math.floor(Math.random() * periods.length)]
    
    courses.push({
      id: i + 1,
      name: template.name,
      teacher: template.teacher,
      location: template.location,
      weekday,
      period,
      week_start: 1,
      week_end: 16
    })
  }
  
  return courses
}

// 生成模拟无课表数据
async function generateMockAvailability(userId, weekRange) {
  const [startWeek, endWeek] = weekRange || [1, 4]
  const items = []
  
  for (let week = startWeek; week <= endWeek; week++) {
    for (let weekday = 1; weekday <= 5; weekday++) {
      for (let period = 1; period <= 4; period++) {
        // 模拟课程占用（约40%概率有课）
        const hasCourse = Math.random() < 0.4
        
        items.push({
          user_id: userId,
          week,
          weekday,
          period,
          is_available: !hasCourse,
          source: 'crawler_import',
          created_at: new Date().toISOString()
        })
      }
    }
  }
  
  if (items.length > 0) {
    await repositories.availability.batchCreate(items)
  }
  
  return items.length
}

export default crawlerHandlers
