// Rows keep the grid's arrays so additions to empty slots persist into confirmation/export.
export function draftRows(preview, days, periods) {
  if (!preview?.grid) return []
  const rows = []
  for (const weekday of days) {
    for (let period = 1; period <= periods; period++) {
      const cell = preview.grid[weekday - 1]?.[period - 1]
      if (!cell) continue
      cell.weekday = weekday
      cell.period = period
      if (!Array.isArray(cell.users)) cell.users = []
      rows.push(cell)
    }
  }
  return rows
}

export function draftCells(preview) {
  return (preview?.grid || []).flatMap((day, d) => day.flatMap((cell, p) =>
    cell?.users?.length ? [{ weekday: d + 1, period: p + 1, user_ids: cell.users.map(u => u.id) }] : []
  ))
}
