// eslint-disable-next-line
import { UserLayout, BasicLayout, BlankLayout } from '@/layouts'

const RouteView = {
  name: 'RouteView',
  render: (h) => h('router-view')
}

export const asyncRouterMap = [

  {
    path: '/',
    name: 'index',
    component: BasicLayout,
    meta: { title: 'menu.home' },
    redirect: '/document/introduce',
    children: [
      // document
      {
        path: '/document',
        name: 'Document',
        component: RouteView,
        meta: { title: 'menu.document', icon: 'copy', permission: [ 'document' ] },
        redirect: '/document/introduce',
        children: [
          // introduce
          {
            path: '/document/introduce',
            name: 'Introduce',
            component: () => import('@/views/doc/Introduce'),
            meta: {
              title: 'menu.document.introduce',
              keepAlive: true,
              permission: ['document']
            }
          },
          // payload
          {
            path: '/document/payload',
            name: 'Payload',
            component: () => import('@/views/doc/Payload'),
            meta: {
              title: 'menu.document.payload',
              keepAlive: true,
              permission: ['document']
            }
          },
          // payload
          {
            path: '/document/api',
            name: 'Api',
            component: () => import('@/views/doc/Api'),
            meta: {
              title: 'menu.document.api',
              keepAlive: true,
              permission: ['document']
            }
          },
          // rebinding
          {
            path: '/document/rebinding',
            name: 'Rebinding',
            component: () => import('@/views/doc/Rebinding'),
            meta: {
              title: 'menu.document.rebinding',
              keepAlive: true,
              permission: ['document']
            }
          },
          // rebinding
          {
            path: '/document/history',
            name: 'History',
            component: () => import('@/views/doc/History'),
            meta: {
              title: 'menu.document.history',
              keepAlive: true,
              permission: ['document']
            }
          },
          // rebinding
          {
              path: '/document/install',
              name: 'Install',
              component: () => import('@/views/doc/Install'),
              meta: {
                title: 'menu.document.install',
                keepAlive: true,
                permission: ['document']
              }
          }
        ]
      },

      // records
      {
        path: '/record',
        name: 'Record',
        component: RouteView,
        meta: { title: 'menu.record', icon: 'form', permission: [ 'record' ] },
        redirect: '/record/dns',
        children: [
          // dashboard
          {
            path: '/record/dns',
            name: 'DnsRecord',
            component: () => import('@/views/record/Dns'),
            meta: {
              title: 'menu.record.dns',
              keepAlive: true,
              permission: ['record']
            }
          },
          {
            path: '/record/http',
            name: 'HttpRecord',
            component: () => import('@/views/record/Http'),
            meta: { title: 'menu.record.http', keepAlive: true, permission: ['record'] }
          }
        ]
      },

      // account
      {
        path: '/setting',
        component: RouteView,
        redirect: '/setting/system/base',
        name: 'Setting',
        meta: { title: 'menu.setting', icon: 'setting', keepAlive: true, permission: [ 'setting' ] },
        children: [
          {
            path: '/setting/system',
            name: 'SettingSystem',
            component: () => import('@/views/account/settings/Index'),
            meta: { title: 'menu.setting.system', hideHeader: true, permission: [ 'setting' ] },
            redirect: '/setting/system/base',
            hideChildrenInMenu: true,
            children: [
              {
                path: '/setting/system/base',
                name: 'BaseSetting',
                component: () => import('@/views/account/settings/BaseSetting'),
                meta: { title: 'menu.setting.system.base', hidden: true, permission: [ 'setting' ] }
              },
              {
                path: '/setting/system/security',
                name: 'SecuritySetting',
                component: () => import('@/views/account/settings/Security'),
                meta: { title: 'menu.setting.system.security', hidden: true, keepAlive: true, permission: [ 'setting' ] }
              }
            ]
          },
          {
            path: '/setting/user',
            name: 'UserSetting',
            component: () => import('@/views/account/user/Index'),
            meta: { title: 'menu.setting.user', hideHeader: true, permission: [ 'manage' ] }
          }
        ]
      }
    ]
  },
  {
    path: '*', redirect: '/404', hidden: true
  }
]

/**
 * 基础路由
 * @type { *[] }
 */
export const constantRouterMap = [
  {
    path: '/user',
    component: UserLayout,
    redirect: '/user/login',
    hidden: true,
    children: [
      {
        path: 'login',
        name: 'login',
        component: () => import(/* webpackChunkName: "user" */ '@/views/user/Login'),
        meta: { title: 'Login' }
      },
      {
        path: 'recover',
        name: 'recover',
        component: undefined
      }
    ]
  },

  {
    path: '/404',
    component: () => import(/* webpackChunkName: "fail" */ '@/views/exception/404')
  }
]
