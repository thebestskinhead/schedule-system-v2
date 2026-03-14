/**
 * 示例排班数据生成脚本
 * 在浏览器控制台运行: seedDemoData()
 */

// 生成示例排班数据
async function seedDemoData(options = {}) {
  const {
    startWeek = 1,
    endWeek = 16,
    departments = ['办公室', '竞赛部', '项目部', '科普部']
  } = options

  console.log('🚀 开始生成示例排班数据...')
  console.log(`📅 周次范围: 第${startWeek}周 - 第${endWeek}周`)
  console.log(`🏢 部门: ${departments.join(', ')}`)

  const results = {
    totalSchedules: 0,
    byDepartment: {}
  }

  for (const dept of departments) {
    console.log(`\n🏢 正在生成 ${dept} 的排班...`)
    const count = await generateForDepartment(dept, startWeek, endWeek)
    results.totalSchedules += count
    results.byDepartment[dept] = count
    console.log(`✅ ${dept}: ${count} 条排班记录`)
  }

  console.log('\n📊 生成完成!')
  console.log(`📈 总计: ${results.totalSchedules} 条排班记录`)
  
  return results
}

// 为单个部门生成排班
async function generateForDepartment(department, startWeek, endWeek) {
  // 从 localStorage 获取用户数据
  const users = getDepartmentUsers(department)
  if (users.length === 0) {
    console.warn(`⚠️ ${department} 没有找到用户`)
    return 0
  }

  let totalCount = 0

  for (let week = startWeek; week <= endWeek; week++) {
    // 获取无课表数据
    const availability = getAvailabilityForUsers(users.map(u => u.id), week)
    
    // 生成排班
    const schedules = generateScheduleForWeek(week, users, availability)
    
    // 保存到 localStorage
    if (schedules.length > 0) {
      saveSchedules(schedules, week)
      totalCount += schedules.length
    }
  }

  return totalCount
}

// 获取部门用户
function getDepartmentUsers(department) {
  const usersData = localStorage.getItem('mock_db_users')
  if (!usersData) return []
  
  try {
    const allUsers = JSON.parse(usersData)
    return allUsers.filter(u => u.department === department)
  } catch (e) {
    return []
  }
}

// 获取用户无课表
function getAvailabilityForUsers(userIds, week) {
  const availData = localStorage.getItem('mock_db_availability')
  if (!availData) return []
  
  try {
    const allAvail = JSON.parse(availData)
    return allAvail.filter(a => 
      userIds.includes(a.user_id) && 
      a.week === week && 
      a.is_available
    )
  } catch (e) {
    return []
  }
}

// 生成周排班
function generateScheduleForWeek(week, users, availability) {
  const schedules = []
  const weekdays = [1, 2, 3, 4, 5]
  const periods = [1, 2, 3, 4]
  const needPerCell = 2

  // 构建可用性映射
  const availMap = new Map()
  for (const av of availability) {
    const key = `${av.user_id}_${av.weekday}_${av.period}`
    availMap.set(key, true)
  }

  for (const weekday of weekdays) {
    for (const period of periods) {
      // 找到该时段有空的用户
      const availableUsers = users.filter(u => {
        const key = `${u.id}_${weekday}_${period}`
        return availMap.has(key)
      })

      // 随机选择
      const shuffled = shuffleArray([...availableUsers])
      const selected = shuffled.slice(0, needPerCell)

      for (const user of selected) {
        schedules.push({
          id: Date.now() + Math.random(),
          week,
          weekday,
          period,
          user_id: user.id,
          user_name: user.name,
          department: user.department,
          status: 'confirmed',
          created_at: new Date().toISOString()
        })
      }
    }
  }

  return schedules
}

// 保存排班到 localStorage
function saveSchedules(schedules, week) {
  const key = `mock_db_schedules_${week}`
  const existing = localStorage.getItem(key)
  
  let allSchedules = []
  if (existing) {
    try {
      allSchedules = JSON.parse(existing)
    } catch (e) {}
  }
  
  allSchedules.push(...schedules)
  localStorage.setItem(key, JSON.stringify(allSchedules))
}

// Fisher-Yates 洗牌算法
function shuffleArray(array) {
  const arr = [...array]
  for (let i = arr.length - 1; i > 0; i--) {
    const j = Math.floor(Math.random() * (i + 1))
    ;[arr[i], arr[j]] = [arr[j], arr[i]]
  }
  return arr
}

// 生成个人值班数据
async function seedMyDuties(userId, userName, weeks = [8, 9, 10, 11, 12]) {
  console.log(`👤 为 ${userName} 生成值班数据...`)
  
  const duties = []
  
  for (const week of weeks) {
    const count = 1 + Math.floor(Math.random() * 2)
    const used = new Set()
    
    for (let i = 0; i < count; i++) {
      let weekday, period, key
      do {
        weekday = 1 + Math.floor(Math.random() * 5)
        period = 1 + Math.floor(Math.random() * 4)
        key = `${weekday}_${period}`
      } while (used.has(key))
      used.add(key)
      
      // 根据周次决定状态
      let status = 'pending'
      if (week < 10) status = 'completed'
      else if (week === 10) status = Math.random() > 0.5 ? 'completed' : 'confirmed'
      else status = 'confirmed'
      
      duties.push({
        id: `duty_${userId}_${week}_${weekday}_${period}`,
        week,
        weekday,
        period,
        user_id: userId,
        user_name: userName,
        status: 'confirmed',
        duty_status: status,
        check_in_time: status === 'completed' ? new Date().toISOString() : null
      })
    }
  }
  
  // 保存到 localStorage
  const key = `mock_db_duties_${userId}`
  localStorage.setItem(key, JSON.stringify(duties))
  
  console.log(`✅ 生成了 ${duties.length} 条值班记录`)
  return duties
}

// 清空排班数据
async function clearSchedules() {
  console.log('🗑️ 清空排班数据...')
  
  let count = 0
  for (let i = 0; i < localStorage.length; i++) {
    const key = localStorage.key(i)
    if (key && key.startsWith('mock_db_schedules_')) {
      localStorage.removeItem(key)
      count++
    }
  }
  
  console.log(`✅ 已清空 ${count} 周的排班数据`)
  return count
}

// 查看排班统计
async function showScheduleStats() {
  console.log('📊 排班数据统计')
  console.log('==================')
  
  const stats = {
    total: 0,
    byWeek: {},
    byDepartment: {}
  }
  
  for (let i = 0; i < localStorage.length; i++) {
    const key = localStorage.key(i)
    if (key && key.startsWith('mock_db_schedules_')) {
      try {
        const data = JSON.parse(localStorage.getItem(key))
        const week = key.replace('mock_db_schedules_', '')
        stats.byWeek[week] = data.length
        stats.total += data.length
        
        // 按部门统计
        for (const s of data) {
          if (!stats.byDepartment[s.department]) {
            stats.byDepartment[s.department] = 0
          }
          stats.byDepartment[s.department]++
        }
      } catch (e) {}
    }
  }
  
  console.log(`📈 总排班记录: ${stats.total}`)
  console.log('\n📅 按周次分布:')
  Object.entries(stats.byWeek)
    .sort((a, b) => parseInt(a[0]) - parseInt(b[0]))
    .forEach(([week, count]) => {
      console.log(`  第${week.padStart(2, '0')}周: ${count} 条`)
    })
  
  console.log('\n🏢 按部门分布:')
  Object.entries(stats.byDepartment)
    .sort((a, b) => b[1] - a[1])
    .forEach(([dept, count]) => {
      console.log(`  ${dept}: ${count} 条`)
    })
  
  return stats
}

// 导出到全局
window.seedDemoData = seedDemoData
window.seedMyDuties = seedMyDuties
window.clearSchedules = clearSchedules
window.showScheduleStats = showScheduleStats

console.log('✅ 排班数据生成脚本已加载!')
console.log('可用命令:')
console.log('  seedDemoData() - 生成示例排班数据')
console.log('  seedDemoData({ startWeek: 1, endWeek: 20 }) - 生成指定周次')
console.log('  seedMyDuties(1, "张三") - 为指定用户生成值班数据')
console.log('  clearSchedules() - 清空排班数据')
console.log('  showScheduleStats() - 查看统计信息')
