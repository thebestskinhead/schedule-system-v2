/**
 * Repository 统一导出
 */

import { UserRepository } from './UserRepository.js'
import { ScheduleRepository } from './ScheduleRepository.js'
import { AvailabilityRepository } from './AvailabilityRepository.js'
import { db } from '../core/Database.js'

// 创建并导出Repository实例
export const userRepo = new UserRepository(db)
export const scheduleRepo = new ScheduleRepository(db)
export const availabilityRepo = new AvailabilityRepository(db)

export {
  UserRepository,
  ScheduleRepository,
  AvailabilityRepository
}

export default {
  users: userRepo,
  schedules: scheduleRepo,
  availability: availabilityRepo,
  db
}
