import request from './request'

export const getMyAvailability = () => request.get('/availability')
export const addAvailability = (data) => request.post('/availability', data)
export const deleteAvailability = (data) => request.delete('/availability', { data })
export const importFromCookie = (data) => request.post('/availability/import/cookie', data)
export const importFromXLS = (file) => {
  const formData = new FormData()
  formData.append('file', file)
  return request.post('/availability/import/xls', formData, {
    headers: { 'Content-Type': 'multipart/form-data' }
  })
}
