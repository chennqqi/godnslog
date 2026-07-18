'use client'

import { useEffect, useState } from 'react'
import { useRouter } from 'next/navigation'
import { rulesApi } from '@/lib/api-client'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import { useI18n } from '@/lib/i18n-context'

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
  config: Record<string, unknown>
}

export default function WorkflowBuilderPage() {
  const router = useRouter()
  const { t } = useI18n()

  useEffect(() => {
    const token = localStorage.getItem('token')
    if (!token) {
      window.location.href = '/login'
    }
  }, [router])

  const [rules, setRules] = useState<Rule[]>([])
  const [selectedRule, setSelectedRule] = useState<Rule | null>(null)
  const [loading, setLoading] = useState(true)

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

  useEffect(() => {
    const timer = setTimeout(() => loadRules(), 0)
    return () => clearTimeout(timer)
  }, [])

  const addRule = async () => {
    const newRule: Partial<Rule> = {
      name: `Rule ${rules.length + 1}`,
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

  const updateRule = async (ruleId: string, field: string, value: unknown) => {
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
        <p className="text-gray-500">{t('common.loading')}</p>
      </div>
    )
  }

  return (
    <div>
      <h2 className="text-2xl font-bold text-gray-900 mb-6">{t('workflow.title')}</h2>

      <div className="grid grid-cols-3 gap-6">
        {/* 规则列表 */}
        <div className="col-span-1">
          <Card>
            <CardHeader>
              <div className="flex justify-between items-center">
                <CardTitle>{t('workflow.rule_list')}</CardTitle>
                <Button onClick={addRule} size="sm">
                  {t('workflow.new_rule')}
                </Button>
              </div>
            </CardHeader>
            <CardContent>
              <div className="space-y-2">
                {rules.length === 0 ? (
                  <p className="text-gray-500 text-sm">{t('workflow.no_rules')}</p>
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
                        {rule.conditions.length} {t('workflow.condition_count')}, {rule.actions.length} {t('workflow.action_count')}
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
                    {t('workflow.rule_name')}
                  </label>
                  <Input
                    type="text"
                    value={selectedRule.name}
                    onChange={(e) => updateRule(selectedRule.id, 'name', e.target.value)}
                  />
                </div>

                <div className="mb-6">
                  <label className="block text-sm font-medium text-gray-700 mb-2">
                    {t('workflow.description')}
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
                    <h3 className="text-lg font-medium text-gray-900">{t('workflow.conditions')}</h3>
                    <button className="text-indigo-600 hover:text-indigo-800 text-sm">
                      {t('workflow.add_condition')}
                    </button>
                  </div>
                  {selectedRule.conditions.length === 0 ? (
                    <p className="text-gray-500 text-sm mb-4">{t('workflow.no_conditions')}</p>
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
                    <h3 className="text-lg font-medium text-gray-900">{t('workflow.actions')}</h3>
                    <button className="text-indigo-600 hover:text-indigo-800 text-sm">
                      {t('workflow.add_action')}
                    </button>
                  </div>
                  {selectedRule.actions.length === 0 ? (
                    <p className="text-gray-500 text-sm mb-4">{t('workflow.no_actions')}</p>
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
                    {selectedRule.enabled ? t('workflow.disable_rule') : t('workflow.enable_rule')}
                  </Button>
                  <Button 
                    onClick={() => deleteRule(selectedRule.id)}
                    variant="destructive"
                  >
                    {t('workflow.delete_rule')}
                  </Button>
                </div>
              </CardContent>
            </Card>
          ) : (
            <Card>
              <CardContent className="pt-6">
                <p className="text-gray-500 text-center">{t('workflow.select_rule')}</p>
              </CardContent>
            </Card>
          )}
        </div>
      </div>
    </div>
  )
}
