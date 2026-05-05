'use client'

import { useEffect, useState } from 'react'
import { useRouter } from 'next/navigation'
import { rulesApi } from '@/lib/api-client'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'

interface Rule {
  id: string
  name: string
  description: string
  enabled: boolean
  priority: number
  conditions: Condition[]
  actions: Action[]
}

interface Condition {
  id: string
  field: string
  operator: string
  value: string
}

interface Action {
  id: string
  type: string
  config: Record<string, any>
}

export default function WorkflowBuilderPage() {
  const router = useRouter()

  useEffect(() => {
    const token = localStorage.getItem('token')
    if (!token) {
      router.push('/login')
    }
  }, [router])

  const [rules, setRules] = useState<Rule[]>([])
  const [selectedRule, setSelectedRule] = useState<Rule | null>(null)
  const [loading, setLoading] = useState(true)

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

  useEffect(() => {
    loadRules()
  }, [])

  const loadRules = async () => {
    try {
      const response = await rulesApi.list()
      if (response.data && response.data.items) {
        setRules(response.data.items)
      }
    } catch (error) {
      console.error('Failed to load rules:', error)
    } finally {
      setLoading(false)
    }
  }

  const addRule = async () => {
    const newRule: Partial<Rule> = {
      name: `规则 ${rules.length + 1}`,
      description: '',
      enabled: true,
      priority: rules.length + 1,
      conditions: [],
      actions: [],
    }
    try {
      const response = await rulesApi.create(newRule)
      if (response.data) {
        setRules([...rules, response.data])
        setSelectedRule(response.data)
      }
    } catch (error) {
      console.error('Failed to create rule:', error)
    }
  }

  const updateRule = async (ruleId: string, field: string, value: any) => {
    const ruleIndex = rules.findIndex(r => r.id === ruleId)
    if (ruleIndex === -1) return

    const updatedRules = [...rules]
    updatedRules[ruleIndex] = { ...updatedRules[ruleIndex], [field]: value }
    setRules(updatedRules)

    if (selectedRule?.id === ruleId) {
      setSelectedRule({ ...selectedRule, [field]: value })
    }

    try {
      await rulesApi.update(ruleId, { [field]: value })
    } catch (error) {
      console.error('Failed to update rule:', error)
    }
  }

  const deleteRule = async (ruleId: string) => {
    try {
      await rulesApi.delete(ruleId)
      setRules(rules.filter(r => r.id !== ruleId))
      if (selectedRule?.id === ruleId) {
        setSelectedRule(null)
      }
    } catch (error) {
      console.error('Failed to delete rule:', error)
    }
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <p className="text-gray-500">加载中...</p>
      </div>
    )
  }

  return (
    <div>
      <h2 className="text-2xl font-bold text-gray-900 mb-6">Workflow Builder</h2>

      <div className="grid grid-cols-3 gap-6">
        {/* 规则列表 */}
        <div className="col-span-1">
          <Card>
            <CardHeader>
              <div className="flex justify-between items-center">
                <CardTitle>规则列表</CardTitle>
                <Button onClick={addRule} size="sm">
                  + 新建
                </Button>
              </div>
            </CardHeader>
            <CardContent>
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
            </CardContent>
          </Card>
        </div>

        {/* 规则编辑器 */}
        <div className="col-span-2">
          {selectedRule ? (
            <Card>
              <CardContent className="pt-6">
                <div className="mb-6">
                  <label className="block text-sm font-medium text-gray-700 mb-2">
                    规则名称
                  </label>
                  <Input
                    type="text"
                    value={selectedRule.name}
                    onChange={(e) => updateRule(selectedRule.id, 'name', e.target.value)}
                  />
                </div>

                <div className="mb-6">
                  <label className="block text-sm font-medium text-gray-700 mb-2">
                    描述
                  </label>
                  <textarea
                    className="w-full px-3 py-2 border border-gray-300 rounded focus:outline-none focus:ring-2 focus:ring-indigo-500"
                    rows={2}
                    value={selectedRule.description}
                    onChange={(e) => updateRule(selectedRule.id, 'description', e.target.value)}
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
                      {selectedRule.conditions.map((condition, idx) => (
                        <div key={idx} className="border border-gray-200 rounded p-3">
                          <span className="text-sm">{condition.field} {condition.operator} {condition.value}</span>
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
                      {selectedRule.actions.map((action, idx) => (
                        <div key={idx} className="border border-gray-200 rounded p-3">
                          <span className="text-sm">{action.type}</span>
                        </div>
                      ))}
                    </div>
                  )}
                </div>

                <div className="flex space-x-4">
                  <Button 
                    onClick={() => updateRule(selectedRule.id, 'enabled', !selectedRule.enabled)}
                  >
                    {selectedRule.enabled ? '禁用规则' : '启用规则'}
                  </Button>
                  <Button 
                    onClick={() => deleteRule(selectedRule.id)}
                    variant="destructive"
                  >
                    删除规则
                  </Button>
                </div>
              </CardContent>
            </Card>
          ) : (
            <Card>
              <CardContent className="pt-6">
                <p className="text-gray-500 text-center">选择一个规则进行编辑</p>
              </CardContent>
            </Card>
          )}
        </div>
      </div>
    </div>
  )
}
