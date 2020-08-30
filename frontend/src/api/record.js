import request from '@/utils/request'

const api = {
  dnsRecord: '/record/dns',
  httpRecord: '/record/http'
}

export default api

export function getDnsList (parameter) {
  return request({
    url: api.dnsRecord,
    method: 'get',
    params: parameter
  })
}

export function deleteDnsList (parameter) {
  return request({
    url: api.dnsRecord,
    method: 'delete',
    data: parameter,
    headers: {
      'Content-Type': 'application/json;charset=UTF-8'
    }
  })
}

export function getHttpList (parameter) {
  return request({
    url: api.httpRecord,
    method: 'get',
    params: parameter
  })
}

export function deleteHttpList (parameter) {
  return request({
    url: api.httpRecord,
    method: 'delete',
    data: parameter,
    headers: {
      'Content-Type': 'application/json;charset=UTF-8'
    }
  })
}
