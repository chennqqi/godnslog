import Mock from 'mockjs2'
import { builder, getQueryParameters } from '../util'
import responseOk from './auth'

const totalCount = 5701

function getRandIP () {
  var ip = []
  for (var i = 0; i < 4; i++) {
    ip = ip + Math.floor(Math.random() * 256) + '.'
  }
  return ip
}

const httpRecordList = (options) => {
  const parameters = getQueryParameters(options)

  const result = []
  const pageNo = parseInt(parameters.pageNo)
  const pageSize = parseInt(parameters.pageSize)
  const totalPage = Math.ceil(totalCount / pageSize)
  const key = (pageNo - 1) * pageSize
  const next = (pageNo >= totalPage ? (totalCount % pageSize) : pageSize) + 1

  for (let i = 1; i < next; i++) {
    const tmpKey = key + i
    result.push({
      key: tmpKey,
      id: tmpKey,
      domain: String(tmpKey) + '.baidu.com',
      method: Math.random() > 0.5 ? 'GET' : 'POST',
      addr: getRandIP(),
      data: '123',
      ctype: 'application/json',
      ctime: Mock.mock('@datetime')
    })
  }

  return builder({
    pageSize: pageSize,
    pageNo: pageNo,
    totalCount: totalCount,
    totalPage: totalPage,
    data: result
  })
}

const dnsRecordList = (options) => {
  const parameters = getQueryParameters(options)

  const result = []
  const pageNo = parseInt(parameters.pageNo)
  const pageSize = parseInt(parameters.pageSize)
  const totalPage = Math.ceil(totalCount / pageSize)
  const key = (pageNo - 1) * pageSize
  const next = (pageNo >= totalPage ? (totalCount % pageSize) : pageSize) + 1

  for (let i = 1; i < next; i++) {
    const tmpKey = key + i
    result.push({
      key: tmpKey,
      id: tmpKey,
      domain: String(tmpKey) + '.baidu.com',
      addr: getRandIP(),
      ctime: Mock.mock('@datetime')
    })
  }

  return builder({
    pageSize: pageSize,
    pageNo: pageNo,
    totalCount: totalCount,
    totalPage: totalPage,
    data: result
  })
}

Mock.mock(/\/record\/dns/, 'get', dnsRecordList)
Mock.mock(/\/record\/http/, 'get', httpRecordList)
Mock.mock(/\/record\/http/, 'delete', responseOk)
Mock.mock(/\/record\/http/, 'delete', responseOk)
