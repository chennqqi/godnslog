'use client'

import { useEffect, useState } from 'react'
import { useRouter } from 'next/navigation'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'

export default function DocsPage() {
  const router = useRouter()

  useEffect(() => {
    const token = localStorage.getItem('token')
    if (!token) {
      router.push('/login')
    }
  }, [router])

  const docs = [
    {
      title: '快速开始',
      description: '了解如何快速开始使用GODNSLOG 2.0',
      link: '/docs/quick-start',
    },
    {
      title: 'API文档',
      description: '查看完整的API参考文档',
      link: '/docs/api',
    },
    {
      title: '用户指南',
      description: '详细的使用指南和最佳实践',
      link: '/docs/user-guide',
    },
    {
      title: '配置参考',
      description: '系统配置选项的详细说明',
      link: '/docs/config',
    },
    {
      title: '常见问题',
      description: '常见问题的解答和故障排除',
      link: '/docs/faq',
    },
    {
      title: '安全指南',
      description: '安全最佳实践和注意事项',
      link: '/docs/security',
    },
  ]

  return (
    <div>
      <h2 className="text-2xl font-bold text-gray-900 mb-6">文档中心</h2>

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        {docs.map((doc, index) => (
          <Card key={index} className="hover:shadow-lg transition-shadow">
            <CardHeader>
              <CardTitle>{doc.title}</CardTitle>
            </CardHeader>
            <CardContent>
              <p className="text-sm text-gray-500 mb-4">{doc.description}</p>
              <Button variant="outline" className="w-full">
                查看文档
              </Button>
            </CardContent>
          </Card>
        ))}
      </div>

      <Card className="mt-8">
        <CardHeader>
          <CardTitle>获取帮助</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            <div>
              <h3 className="font-medium text-gray-900 mb-2">GitHub Issues</h3>
              <p className="text-sm text-gray-500 mb-2">
                在GitHub上提交问题或功能请求
              </p>
              <Button variant="outline" size="sm">
                访问GitHub
              </Button>
            </div>
            <div>
              <h3 className="font-medium text-gray-900 mb-2">社区支持</h3>
              <p className="text-sm text-gray-500 mb-2">
                加入社区讨论和获取帮助
              </p>
              <Button variant="outline" size="sm">
                加入社区
              </Button>
            </div>
          </div>
        </CardContent>
      </Card>
    </div>
  )
}
