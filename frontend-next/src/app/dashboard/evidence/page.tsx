'use client'

import { useEffect, useState } from 'react'
import { useRouter } from 'next/navigation'

export default function EvidenceReportPage() {
  const router = useRouter()

  useEffect(() => {
    const token = localStorage.getItem('token')
    if (!token) {
      router.push('/login')
    }
  }, [router])
  const [selectedCase, setSelectedCase] = useState<string>('')
  const [format, setFormat] = useState('markdown')
  const [includeRaw, setIncludeRaw] = useState(false)

  const cases = [
    { id: '1', title: '项目A SSRF测试', status: 'active' },
    { id: '2', title: '项目B XXE验证', status: 'completed' },
    { id: '3', title: '项目C RCE检测', status: 'active' },
  ]

  const handleGenerate = () => {
    alert(`生成报告: Case ${selectedCase}, 格式: ${format}`)
  }

  return (
    <div>
      <h2 className="text-2xl font-bold text-gray-900 mb-6">证据报告</h2>

      <div className="grid grid-cols-2 gap-6">
        <div className="bg-white shadow rounded-lg">
          <div className="px-4 py-5 sm:p-6">
            <h3 className="text-lg font-medium text-gray-900 mb-4">生成报告</h3>

            <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  选择Case
                </label>
                <select
                  className="w-full px-3 py-2 border border-gray-300 rounded focus:outline-none focus:ring-2 focus:ring-indigo-500"
                  value={selectedCase}
                  onChange={(e) => setSelectedCase(e.target.value)}
                >
                  <option value="">选择Case</option>
                  {cases.map((c) => (
                    <option key={c.id} value={c.id}>
                      {c.title} ({c.status})
                    </option>
                  ))}
                </select>
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  报告格式
                </label>
                <select
                  className="w-full px-3 py-2 border border-gray-300 rounded focus:outline-none focus:ring-2 focus:ring-indigo-500"
                  value={format}
                  onChange={(e) => setFormat(e.target.value)}
                >
                  <option value="markdown">Markdown</option>
                  <option value="json">JSON</option>
                  <option value="html">HTML</option>
                  <option value="pdf">PDF</option>
                </select>
              </div>

              <div>
                <label className="flex items-center">
                  <input
                    type="checkbox"
                    checked={includeRaw}
                    onChange={(e) => setIncludeRaw(e.target.checked)}
                    className="mr-2"
                  />
                  <span className="text-sm text-gray-700">包含原始数据</span>
                </label>
              </div>

              <div>
                <label className="flex items-center">
                  <input type="checkbox" className="mr-2" defaultChecked />
                  <span className="text-sm text-gray-700">包含时间线</span>
                </label>
              </div>

              <div>
                <label className="flex items-center">
                  <input type="checkbox" className="mr-2" defaultChecked />
                  <span className="text-sm text-gray-700">包含分析总结</span>
                </label>
              </div>

              <button
                onClick={handleGenerate}
                disabled={!selectedCase}
                className="w-full px-4 py-2 bg-indigo-600 text-white rounded hover:bg-indigo-700 disabled:opacity-50"
              >
                生成报告
              </button>
            </div>
          </div>
        </div>

        <div className="bg-white shadow rounded-lg">
          <div className="px-4 py-5 sm:p-6">
            <h3 className="text-lg font-medium text-gray-900 mb-4">报告模板</h3>

            <div className="space-y-3">
              <div className="border border-gray-200 rounded p-3 hover:border-indigo-600 cursor-pointer">
                <p className="font-medium text-sm">标准渗透测试报告</p>
                <p className="text-xs text-gray-500">包含完整的漏洞详情和证据链</p>
              </div>
              <div className="border border-gray-200 rounded p-3 hover:border-indigo-600 cursor-pointer">
                <p className="font-medium text-sm">快速验证报告</p>
                <p className="text-xs text-gray-500">简洁的证据汇总和结论</p>
              </div>
              <div className="border border-gray-200 rounded p-3 hover:border-indigo-600 cursor-pointer">
                <p className="font-medium text-sm">合规审计报告</p>
                <p className="text-xs text-gray-500">符合行业标准的审计格式</p>
              </div>
              <div className="border border-gray-200 rounded p-3 hover:border-indigo-600 cursor-pointer">
                <p className="font-medium text-sm">自定义模板</p>
                <p className="text-xs text-gray-500">上传自定义报告模板</p>
              </div>
            </div>
          </div>
        </div>
      </div>

      <div className="mt-6 bg-white shadow rounded-lg">
        <div className="px-4 py-5 sm:p-6">
          <h3 className="text-lg font-medium text-gray-900 mb-4">最近生成的报告</h3>
          <p className="text-gray-500 text-center py-4">暂无生成的报告</p>
        </div>
      </div>
    </div>
  )
}
