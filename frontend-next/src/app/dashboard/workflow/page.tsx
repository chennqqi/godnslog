'use client'

import { useEffect, useState } from 'react'
import { useRouter } from 'next/navigation'

export default function WorkflowBuilderPage() {
  const router = useRouter()

  useEffect(() => {
    const token = localStorage.getItem('token')
    if (!token) {
      router.push('/login')
    }
  }, [router])
  const [rules, setRules] = useState<any[]>([])
  const [selectedRule, setSelectedRule] = useState<any | null>(null)

  const conditions = [
    { id: 'protocol', label: '协议', type: 'select', options: ['DNS', 'HTTP', 'SMTP', 'LDAP', 'SMB', 'FTP'] },
    { id: 'source_ip', label: '来源IP', type: 'text' },
    { id: 'token', label: 'Token', type: 'text' },
    { id: 'path', label: '路径', type: 'text' },
    { id: 'header', label: 'Header', type: 'text' },
    { id: 'body', label: 'Body', type: 'text' },
    { id: 'keyword', label: '关键词', type: 'text' },
    { id: 'case', label: 'Case', type: 'text' },
    { id: 'risk_level', label: '风险等级', type: 'select', options: ['low', 'medium', 'high', 'critical'] },
  ]

  const actions = [
    { id: 'notify', label: '发送通知' },
    { id: 'tag', label: '打标签' },
    { id: 'webhook', label: '转发Webhook' },
    { id: 'modify_response', label: '修改响应' },
    { id: 'save_attachment', label: '保存附件' },
    { id: 'discard', label: '丢弃噪声' },
    { id: 'create_report', label: '创建报告' },
    { id: 'call_api', label: '调用外部API' },
  ]

  const addRule = () => {
    const newRule = {
      id: Date.now(),
      name: `规则 ${rules.length + 1}`,
      conditions: [],
      actions: [],
      enabled: true,
    }
    setRules([...rules, newRule])
    setSelectedRule(newRule)
  }

  const updateRule = (ruleId: number, field: string, value: any) => {
    setRules(rules.map(r => r.id === ruleId ? { ...r, [field]: value } : r))
  }

  return (
    <div>
      <h2 className="text-2xl font-bold text-gray-900 mb-6">Workflow Builder</h2>

      <div className="grid grid-cols-3 gap-6">
        {/* 规则列表 */}
        <div className="col-span-1">
          <div className="bg-white shadow rounded-lg">
            <div className="px-4 py-5 sm:p-6">
              <div className="flex justify-between items-center mb-4">
                <h3 className="text-lg font-medium text-gray-900">规则列表</h3>
                <button
                  onClick={addRule}
                  className="px-3 py-1 bg-indigo-600 text-white rounded hover:bg-indigo-700 text-sm"
                >
                  + 新建
                </button>
              </div>

              <div className="space-y-2">
                {rules.length === 0 ? (
                  <p className="text-gray-500 text-sm">暂无规则</p>
                ) : (
                  rules.map((rule) => (
                    <div
                      key={rule.id}
                      onClick={() => setSelectedRule(rule)}
                      className={`p-3 border rounded cursor-pointer ${
                        selectedRule?.id === rule.id ? 'border-indigo-600 bg-indigo-50' : 'border-gray-200'
                      }`}
                    >
                      <div className="flex justify-between items-center">
                        <span className="font-medium text-sm">{rule.name}</span>
                        <label className="flex items-center">
                          <input
                            type="checkbox"
                            checked={rule.enabled}
                            onChange={(e) => updateRule(rule.id, 'enabled', e.target.checked)}
                            className="mr-1"
                          />
                        </label>
                      </div>
                      <p className="text-xs text-gray-500 mt-1">
                        {rule.conditions.length} 条件, {rule.actions.length} 动作
                      </p>
                    </div>
                  ))
                )}
              </div>
            </div>
          </div>
        </div>

        {/* 规则编辑器 */}
        <div className="col-span-2">
          {selectedRule ? (
            <div className="bg-white shadow rounded-lg">
              <div className="px-4 py-5 sm:p-6">
                <div className="mb-6">
                  <label className="block text-sm font-medium text-gray-700 mb-2">
                    规则名称
                  </label>
                  <input
                    type="text"
                    className="w-full px-3 py-2 border border-gray-300 rounded focus:outline-none focus:ring-2 focus:ring-indigo-500"
                    value={selectedRule.name}
                    onChange={(e) => updateRule(selectedRule.id, 'name', e.target.value)}
                  />
                </div>

                <div className="mb-6">
                  <div className="flex justify-between items-center mb-3">
                    <h3 className="text-lg font-medium text-gray-900">条件</h3>
                    <button className="text-indigo-600 hover:text-indigo-800 text-sm">
                      + 添加条件
                    </button>
                  </div>
                  {selectedRule.conditions.length === 0 ? (
                    <p className="text-gray-500 text-sm mb-4">暂无条件</p>
                  ) : (
                    <div className="space-y-2 mb-4">
                      {selectedRule.conditions.map((condition: any, idx: number) => (
                        <div key={idx} className="border border-gray-200 rounded p-3">
                          <span className="text-sm">{condition}</span>
                        </div>
                      ))}
                    </div>
                  )}
                </div>

                <div className="mb-6">
                  <div className="flex justify-between items-center mb-3">
                    <h3 className="text-lg font-medium text-gray-900">动作</h3>
                    <button className="text-indigo-600 hover:text-indigo-800 text-sm">
                      + 添加动作
                    </button>
                  </div>
                  {selectedRule.actions.length === 0 ? (
                    <p className="text-gray-500 text-sm mb-4">暂无动作</p>
                  ) : (
                    <div className="space-y-2 mb-4">
                      {selectedRule.actions.map((action: any, idx: number) => (
                        <div key={idx} className="border border-gray-200 rounded p-3">
                          <span className="text-sm">{action}</span>
                        </div>
                      ))}
                    </div>
                  )}
                </div>

                <div className="flex space-x-4">
                  <button className="px-4 py-2 bg-indigo-600 text-white rounded hover:bg-indigo-700">
                    保存规则
                  </button>
                  <button className="px-4 py-2 bg-gray-200 text-gray-700 rounded hover:bg-gray-300">
                    测试规则
                  </button>
                </div>
              </div>
            </div>
          ) : (
            <div className="bg-white shadow rounded-lg">
              <div className="px-4 py-5 sm:p-6">
                <p className="text-gray-500 text-center">选择一个规则进行编辑</p>
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  )
}
