'use client'

import { useState } from 'react'

export default function RebindingLabPage() {
  const [stages, setStages] = useState<any[]>([
    { id: 1, name: '首次解析', target_ip: '127.0.0.1', ttl: 10, condition: 'first_visit' },
    { id: 2, name: '后续解析', target_ip: '192.168.1.1', ttl: 60, condition: 'always' },
  ])
  const [selectedStage, setSelectedStage] = useState<any | null>(null)

  const scenarios = [
    { id: 'browser', name: '浏览器 DNS Rebinding', description: '利用浏览器DNS缓存绕过同源策略' },
    { id: 'cloud_metadata', name: '云元数据访问', description: '访问云服务元数据端点' },
    { id: 'internal_mgmt', name: '内网管理面探测', description: '探测内网管理接口' },
    { id: 'iot_router', name: 'IoT/路由器', description: '针对IoT设备和路由器的Rebinding' },
  ]

  const addStage = () => {
    const newStage = {
      id: Date.now(),
      name: `阶段 ${stages.length + 1}`,
      target_ip: '',
      ttl: 60,
      condition: 'always',
    }
    setStages([...stages, newStage])
    setSelectedStage(newStage)
  }

  return (
    <div>
      <h2 className="text-2xl font-bold text-gray-900 mb-6">Rebinding Lab</h2>

      <div className="grid grid-cols-3 gap-6">
        {/* 场景选择 */}
        <div className="col-span-1">
          <div className="bg-white shadow rounded-lg mb-6">
            <div className="px-4 py-5 sm:p-6">
              <h3 className="text-lg font-medium text-gray-900 mb-4">内置场景</h3>
              <div className="space-y-2">
                {scenarios.map((scenario) => (
                  <div
                    key={scenario.id}
                    className="p-3 border border-gray-200 rounded hover:border-indigo-600 cursor-pointer"
                  >
                    <p className="font-medium text-sm">{scenario.name}</p>
                    <p className="text-xs text-gray-500 mt-1">{scenario.description}</p>
                  </div>
                ))}
              </div>
            </div>
          </div>

          <div className="bg-white shadow rounded-lg">
            <div className="px-4 py-5 sm:p-6">
              <h3 className="text-lg font-medium text-gray-900 mb-4">配置说明</h3>
              <div className="text-sm text-gray-600 space-y-2">
                <p>• 首次解析：首次DNS查询时返回的IP</p>
                <p>• 后续解析：后续DNS查询时返回的IP</p>
                <p>• TTL：DNS记录的生存时间</p>
                <p>• 条件：触发解析的条件</p>
              </div>
            </div>
          </div>
        </div>

        {/* 阶段配置 */}
        <div className="col-span-2">
          <div className="bg-white shadow rounded-lg">
            <div className="px-4 py-5 sm:p-6">
              <div className="flex justify-between items-center mb-6">
                <h3 className="text-lg font-medium text-gray-900">DNS解析阶段配置</h3>
                <button
                  onClick={addStage}
                  className="px-3 py-1 bg-indigo-600 text-white rounded hover:bg-indigo-700 text-sm"
                >
                  + 添加阶段
                </button>
              </div>

              <div className="space-y-4">
                {stages.map((stage, idx) => (
                  <div
                    key={stage.id}
                    onClick={() => setSelectedStage(stage)}
                    className={`p-4 border rounded cursor-pointer ${
                      selectedStage?.id === stage.id ? 'border-indigo-600 bg-indigo-50' : 'border-gray-200'
                    }`}
                  >
                    <div className="flex justify-between items-center mb-2">
                      <span className="font-medium">阶段 {idx + 1}: {stage.name}</span>
                      <span className="text-xs text-gray-500">TTL: {stage.ttl}s</span>
                    </div>
                    <div className="text-sm text-gray-600">
                      目标IP: {stage.target_ip || '未设置'}
                    </div>
                    <div className="text-sm text-gray-600">
                      条件: {stage.condition}
                    </div>
                  </div>
                ))}
              </div>

              {selectedStage && (
                <div className="mt-6 pt-6 border-t border-gray-200">
                  <h4 className="text-md font-medium text-gray-900 mb-4">编辑阶段</h4>
                  <div className="space-y-4">
                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-2">
                        阶段名称
                      </label>
                      <input
                        type="text"
                        className="w-full px-3 py-2 border border-gray-300 rounded focus:outline-none focus:ring-2 focus:ring-indigo-500"
                        value={selectedStage.name}
                        onChange={(e) => {
                          setStages(stages.map(s => s.id === selectedStage.id ? { ...s, name: e.target.value } : s))
                          setSelectedStage({ ...selectedStage, name: e.target.value })
                        }}
                      />
                    </div>

                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-2">
                        目标IP
                      </label>
                      <input
                        type="text"
                        className="w-full px-3 py-2 border border-gray-300 rounded focus:outline-none focus:ring-2 focus:ring-indigo-500"
                        value={selectedStage.target_ip}
                        onChange={(e) => {
                          setStages(stages.map(s => s.id === selectedStage.id ? { ...s, target_ip: e.target.value } : s))
                          setSelectedStage({ ...selectedStage, target_ip: e.target.value })
                        }}
                        placeholder="例如: 127.0.0.1"
                      />
                    </div>

                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-2">
                        TTL（秒）
                      </label>
                      <input
                        type="number"
                        className="w-full px-3 py-2 border border-gray-300 rounded focus:outline-none focus:ring-2 focus:ring-indigo-500"
                        value={selectedStage.ttl}
                        onChange={(e) => {
                          setStages(stages.map(s => s.id === selectedStage.id ? { ...s, ttl: parseInt(e.target.value) } : s))
                          setSelectedStage({ ...selectedStage, ttl: parseInt(e.target.value) })
                        }}
                        min={1}
                        max={86400}
                      />
                    </div>

                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-2">
                        触发条件
                      </label>
                      <select
                        className="w-full px-3 py-2 border border-gray-300 rounded focus:outline-none focus:ring-2 focus:ring-indigo-500"
                        value={selectedStage.condition}
                        onChange={(e) => {
                          setStages(stages.map(s => s.id === selectedStage.id ? { ...s, condition: e.target.value } : s))
                          setSelectedStage({ ...selectedStage, condition: e.target.value })
                        }}
                      >
                        <option value="always">总是</option>
                        <option value="first_visit">首次访问</option>
                        <option value="delayed">延迟访问</option>
                        <option value="repeat">重复访问</option>
                      </select>
                    </div>

                    <div className="flex space-x-4">
                      <button className="px-4 py-2 bg-indigo-600 text-white rounded hover:bg-indigo-700">
                        保存配置
                      </button>
                      <button className="px-4 py-2 bg-gray-200 text-gray-700 rounded hover:bg-gray-300">
                        测试解析
                      </button>
                    </div>
                  </div>
                </div>
              )}

              <div className="mt-6 pt-6 border-t border-gray-200">
                <div className="flex space-x-4">
                  <button className="px-4 py-2 bg-green-600 text-white rounded hover:bg-green-700">
                    启用Rebinding
                  </button>
                  <button className="px-4 py-2 bg-gray-200 text-gray-700 rounded hover:bg-gray-300">
                    重置配置
                  </button>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}
