import antd from 'ant-design-vue/es/locale-provider/zh_CN'
import momentCN from 'moment/locale/zh-cn'

const components = {
  antLocale: antd,
  momentName: 'zh-cn',
  momentLocale: momentCN
}

const locale = {
  'message': '-',
  'menu.home': '主页',
  'menu.dashboard': '仪表盘',
  'menu.dashboard.analysis': '分析页',
  'menu.dashboard.monitor': '监控页',
  'menu.dashboard.workplace': '工作台',

  'menu.document': '文档',
  'menu.document.introduce': '简介',
  'menu.document.payload': '用法',
  'menu.document.api': '接口',

  'menu.setting': '配置',
  'menu.setting.system.security': '安全设置',
  'menu.setting.system.base': '基础设置',
  'menu.setting.system': '系统',
  'menu.setting.user': '用户',

  'menu.record': '记录',
  'menu.record.dns': 'dns',
  'menu.record.http': 'http',
  'menu.document.rebinding': '重绑定',
  'menu.document.install': '安装',
  'menu.document.history': '历史',

  // setting.system
  'setting.system.base.callback': '回调地址',
  'setting.system.base.cleanInterval(hour)': '清理周期(小时)',
  'setting.system.base.unit.day': '天',
  'setting.system.base.unit.hour': '小时',
  'setting.system.base.rebind': '域名重绑定',

  'DNS Addr': 'DNS地址',
  'HTTP Addr': 'HTTP地址',
  'Secret': '密钥',

  'auto clean in hours': '清理数小时前的记录',
  'Domain': '域名',
  'New User': '新建用户',
  'Open': '打开',
  'Close': '关闭',
  'Batch': '批量操作',
  'Username': '用户名',
  'UpdateTime': '更新时间',
  'Edit': '编辑',
  'Back': '返回',
  'Delete': '删除',
  'Change Password': '修改密码',
  'Copy': '复制',
  'Modify': '修改',
  'Error': '错误',
  'Welcome': '欢迎',
  'Delete Select': '删除选中',
  'Query': '查询',
  'Reset': '重置',
  'Clear': '清空',
  'Expand': '展开',
  'Collapse': '收起',
  'domain': '域名',
  'data': '数据',
  'date': '日期',
  'delete': '删除',
  'Delete All': '删除全部',
  'open': '开启',
  'close': '关闭',
  'Action': '操作',
  'logout': '登出',
  'account': '账户',
  'Submit': '提交',
  'password': '密码',
  'Please input a valid account or email': '请输入帐户名或邮箱地址',
  'Please input a valid password': '请输入密码',
  'Auto login': '自动登录',
  'I forget password': '忘记密码',
  'Login': '登录',
  'An exquisite dns&http log server for verify SSRF/XXE/RFI/RCE vulnerability': '一个精致的SSRF/XXE/RFI/RCE漏洞dns&http LOG验证服务器'
}

export default {
  ...components,
  ...locale
}
