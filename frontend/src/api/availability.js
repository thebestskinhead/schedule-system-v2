import request from './request'

export const addAvailability = (data) => {
  return request.post('/availability', data)
}

export const getMyAvailability = () => {
  return request.get('/availability')
}

export const getAllAvailability = () => {
  return request.get('/admin/availability/all')
}

export const deleteAvailability = (data) => {
  return request.delete('/availability', { data })
}

// 从Cookie导入
export const importFromCookie = (data) => {
  return request.post('/availability/import/cookie', data)
}

// 从XLS文件导入
export const importFromXLS = (file) => {
  const formData = new FormData()
  formData.append('file', file)
  return request.post('/availability/import/xls', formData, {
    headers: {
      'Content-Type': 'multipart/form-data'
    }
  })
}
