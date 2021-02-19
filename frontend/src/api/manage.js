import request from '@/utils/request'

const manageApi = {
  UserList: '/admin/user/list',
  User: '/admin/user',

  SettingApp: '/setting/app',
  SettingSecurity: '/setting/security',
  SettingResolve: '/setting/resolve',

  Permission: '/permission',
  PermissionNoPager: '/permission/no-pager',
  OrgTree: '/org/tree'
}

export function getUserList (parameter) {
  return request({
    url: manageApi.UserList,
    method: 'get',
    params: parameter
  })
}

export function delUser (parameter) {
  return request({
    url: manageApi.User,
    method: 'delete',
    data: parameter,
    headers: {
      'Content-Type': 'application/json;charset=UTF-8'
    }
  })
}

export function saveUser (user) {
  return request({
    url: manageApi.User,
    method: user.id === 0 ? 'put' : 'post',
    data: user,
    headers: {
      'Content-Type': 'application/json;charset=UTF-8'
    }
  })
}

export function getSettingSecurity (parameter) {
  return request({
    url: manageApi.SettingSecurity,
    method: 'get'
  })
}

export function setSettingSecurity (parameter) {
  return request({
    url: manageApi.SettingSecurity,
    method: 'post',
    data: parameter,
    headers: {
      'Content-Type': 'application/json;charset=UTF-8'
    }
  })
}

export function getSettingApp (parameter) {
  return request({
    url: manageApi.SettingApp,
    method: 'get'
  })
}

export function setSettingApp (parameter) {
  return request({
    url: manageApi.SettingApp,
    method: 'post',
    data: parameter,
    headers: {
      'Content-Type': 'application/json;charset=UTF-8'
    }
  })
}

export function getPermissions (parameter) {
  return request({
    url: manageApi.PermissionNoPager,
    method: 'get',
    params: parameter
  })
}

export function getOrgTree (parameter) {
  return request({
    url: manageApi.OrgTree,
    method: 'get',
    params: parameter
  })
}

export function getResolves (parameter) {
  return request({
    url: manageApi.SettingResolve,
    method: 'get',
    params: parameter
  })
}

export function setResolve (parameter) {
  return request({
    url: manageApi.SettingResolve,
    method: 'post',
    data: parameter
  })
}

export function delResolve (parameter) {
  return request({
    url: manageApi.SettingResolve,
    method: 'delete',
    params: parameter
  })
}
