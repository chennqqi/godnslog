import Mock from 'mockjs2'
import { builder } from '../util'
import responseOk from './auth'

const appSetting = (options) => {
  return builder({
    cleanHour: 24,
    rebind: [ '127.0.0.1', '10.10.10.10', '127.0.0.2' ],
    callback: 'http://127.0.0.1/callback'
  })
}

const securitySetting = (options) => {
  return builder({
    domain: 'xxxx.dommain.com',
    token: '12312312313123123'
  })
}

Mock.mock(/\/setting\/app/, 'get', appSetting)
Mock.mock(/\/setting\/app/, 'post', responseOk)
Mock.mock(/\/setting\/security/, 'get', securitySetting)
Mock.mock(/\/setting\/security/, 'post', responseOk)
