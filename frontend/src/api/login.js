import request from '@/utils/request'

const loginApi = {
  Login: '/auth/login',
  Logout: '/auth/logout',
  Role: '/auth/role',
  UserInfo: '/auth/info',
  UserMenu: '/auth/nav'
}

/**
 * login func
 * parameter: {
 *     username: '',
 *     password: '',
 *     remember_me: true,
 *     captcha: '12345'
 * }
 * @param parameter
 * @returns {*}
 */
export function login (parameter) {
  return request({
    url: loginApi.Login,
    method: 'post',
    data: parameter
  })
}

export function getInfo () {
  return request({
    url: loginApi.UserInfo,
    method: 'get'
  })
}

export function logout () {
  return request({
    url: loginApi.Logout,
    method: 'post',
    headers: {
      'Content-Type': 'application/json;charset=UTF-8'
    }
  })
}

export function getCurrentUserNav () {
  return request({
    url: loginApi.UserMenu,
    method: 'get'
  })
}

export function getRoleList (parameter) {
  return request({
    url: loginApi.Role,
    method: 'get',
    params: parameter
  })
}
