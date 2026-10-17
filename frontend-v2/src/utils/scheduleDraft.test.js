import { describe, it, expect } from 'vitest'
import { draftRows, draftCells } from './scheduleDraft'
import { draftRows as mobileRows, draftCells as mobileCells } from '../../../frontend-mobile/src/utils/scheduleDraft'

for (const [name, rows, cells] of [['desktop', draftRows, draftCells], ['mobile', mobileRows, mobileCells]]) {
  describe(name + ' draft edits', () => {
    it('persists adding a user to an empty null slot through recomputation and confirmation', () => {
      const preview = { week: 9, grid: [[{ users: null }, { users: [] }]] }
      const row = rows(preview, [1], 2)[0]
      row.users.push({ id: 7, name: '测试成员' })
      expect(rows(preview, [1], 2)[0].users).toEqual([{ id: 7, name: '测试成员' }])
      expect(cells(preview)).toEqual([{ weekday: 1, period: 1, user_ids: [7] }])
      row.users.splice(0, 1)
      expect(cells(preview)).toEqual([])
    })
    it('limits editable rows to selected days and periods', () => {
      const preview = { grid: [[{ users: [] }, { users: [] }], [{ users: [] }]] }
      expect(rows(preview, [2], 1).map(row => [row.weekday, row.period])).toEqual([[2, 1]])
    })
  })
}
