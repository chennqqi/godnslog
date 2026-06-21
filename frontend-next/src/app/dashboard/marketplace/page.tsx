'use client'

import { useEffect, useState, useCallback } from 'react'
import { useRouter } from 'next/navigation'
import { marketplaceApi } from '@/lib/api-client'
import { useI18n } from '@/lib/i18n-context'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog'

interface Plugin {
  id: string
  name: string
  description: string
  version: string
  author: string
  downloads: number
  rating: number
  installed: boolean
}

interface Template {
  id: string
  name: string
  description: string
  category: string
  downloads: number
  installed: boolean
}

export default function MarketplacePage() {
  const router = useRouter()
  const { t } = useI18n()

  useEffect(() => {
    const token = localStorage.getItem('token')
    if (!token) {
      router.push('/login')
    }
  }, [router])

  const [activeTab, setActiveTab] = useState<'plugins' | 'templates' | 'installed'>('plugins')
  const [plugins, setPlugins] = useState<Plugin[]>([])
  const [templates, setTemplates] = useState<Template[]>([])
  const [installedPlugins, setInstalledPlugins] = useState<Plugin[]>([])
  const [loading, setLoading] = useState(true)
  const [searchQuery, setSearchQuery] = useState('')
  const [createDialogOpen, setCreateDialogOpen] = useState(false)
  const [createMode, setCreateMode] = useState<'plugin' | 'template'>('plugin')
  const [createName, setCreateName] = useState('')
  const [createDescription, setCreateDescription] = useState('')
  const [createCategory, setCreateCategory] = useState('')
  const [createType, setCreateType] = useState('')
  const [createContent, setCreateContent] = useState('')
  const [createVersion, setCreateVersion] = useState('1.0.0')
  const [createAuthor, setCreateAuthor] = useState('')
  const [creating, setCreating] = useState(false)

  const resetCreateForm = () => {
    setCreateName('')
    setCreateDescription('')
    setCreateCategory('')
    setCreateType('')
    setCreateContent('')
    setCreateVersion('1.0.0')
    setCreateAuthor('')
  }

  const loadData = useCallback(async () => {
    setLoading(true)
    try {
      if (activeTab === 'plugins') {
        const response = await marketplaceApi.listPlugins()
        if (response.data && response.data.items) {
          setPlugins(response.data.items)
        }
      } else if (activeTab === 'templates') {
        const response = await marketplaceApi.listTemplates()
        if (response.data && response.data.items) {
          setTemplates(response.data.items)
        }
      } else if (activeTab === 'installed') {
        const response = await marketplaceApi.listInstalled()
        if (response.data && response.data.items) {
          setInstalledPlugins(response.data.items || [])
        }
      }
    } catch (error) {
      console.error('Failed to load data:', error)
    } finally {
      setLoading(false)
    }
  }, [activeTab])

  useEffect(() => {
    const token = localStorage.getItem('token')
    if (!token) {
      router.push('/login')
    }
  }, [router])

  useEffect(() => {
    const timer = setTimeout(() => loadData(), 0)
    return () => clearTimeout(timer)
  }, [loadData])

  const installPlugin = async (pluginId: string) => {
    try {
      await marketplaceApi.installPlugin(pluginId)
      setPlugins(plugins.map(p => p.id === pluginId ? { ...p, installed: true } : p))
    } catch (error) {
      console.error('Failed to install plugin:', error)
    }
  }

  const installTemplate = async (templateId: string) => {
    try {
      await marketplaceApi.getTemplate(templateId)
      setTemplates(templates.map(t => t.id === templateId ? { ...t, installed: true } : t))
    } catch (error) {
      console.error('Failed to install template:', error)
    }
  }

  const handleCreate = async () => {
    if (!createName.trim()) return
    setCreating(true)
    try {
      if (createMode === 'plugin') {
        await marketplaceApi.createPlugin({
          name: createName,
          description: createDescription,
          version: createVersion,
          author: createAuthor,
          type: createType || 'processor',
          category: createCategory || 'general',
          code: createContent,
          language: 'javascript',
          is_official: false,
        })
      } else {
        await marketplaceApi.createTemplate({
          name: createName,
          description: createDescription,
          type: createType || 'payload',
          category: createCategory || 'general',
          content: createContent,
          format: 'yaml',
          is_official: false,
        })
      }
      resetCreateForm()
      setCreateDialogOpen(false)
      loadData()
    } catch (error) {
      console.error('Failed to create marketplace item:', error)
    } finally {
      setCreating(false)
    }
  }

  const openCreateDialog = () => {
    setCreateMode(activeTab === 'templates' ? 'template' : 'plugin')
    resetCreateForm()
    setCreateDialogOpen(true)
  }

  return (
    <div>
      <h2 className="text-2xl font-bold text-gray-900 dark:text-gray-100 mb-6">{t('marketplace.title')}</h2>

      {/* Tab Navigation */}
      <div className="mb-6">
        <div className="flex items-center justify-between">
          <div className="flex space-x-4">
            <button
            onClick={() => setActiveTab('plugins')}
            className={`px-4 py-2 rounded-lg ${
              activeTab === 'plugins'
                ? 'bg-indigo-600 text-white'
                : 'bg-gray-200 text-gray-700 hover:bg-gray-300'
            }`}
          >
            {t('marketplace.plugins')}
          </button>
          <button
            onClick={() => setActiveTab('templates')}
            className={`px-4 py-2 rounded-lg ${
              activeTab === 'templates'
                ? 'bg-indigo-600 text-white'
                : 'bg-gray-200 text-gray-700 hover:bg-gray-300'
            }`}
          >
            {t('marketplace.templates')}
          </button>
          <button
            onClick={() => setActiveTab('installed')}
            className={`px-4 py-2 rounded-lg ${
              activeTab === 'installed'
                ? 'bg-indigo-600 text-white'
                : 'bg-gray-200 text-gray-700 hover:bg-gray-300'
            }`}
          >
            {t('marketplace.installed')}
          </button>
          </div>
          {activeTab !== 'installed' && (
            <Button onClick={openCreateDialog}>
              {activeTab === 'plugins' ? t('marketplace.create_plugin') : t('marketplace.create_template')}
            </Button>
          )}
        </div>
      </div>

      {/* Search */}
      <div className="mb-6">
        <input
          type="text"
          placeholder={t('marketplace.search')}
          value={searchQuery}
          onChange={(e) => setSearchQuery(e.target.value)}
          className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-indigo-500 dark:bg-gray-800 dark:border-gray-600 dark:text-gray-100"
        />
      </div>

      {/* Create Dialog */}
      <Dialog open={createDialogOpen} onOpenChange={setCreateDialogOpen}>
        <DialogContent className="max-w-lg">
          <DialogHeader>
            <DialogTitle>
              {createMode === 'plugin' ? t('marketplace.create_plugin') : t('marketplace.create_template')}
            </DialogTitle>
            <DialogDescription>{t('marketplace.create_description')}</DialogDescription>
          </DialogHeader>
          <div className="space-y-4 py-4">
            <div>
              <Label htmlFor="create-name">{t('marketplace.create_name')}</Label>
              <Input
                id="create-name"
                value={createName}
                onChange={(e) => setCreateName(e.target.value)}
                placeholder={t('marketplace.create_name')}
              />
            </div>
            <div>
              <Label htmlFor="create-description">{t('marketplace.create_description')}</Label>
              <Textarea
                id="create-description"
                value={createDescription}
                onChange={(e) => setCreateDescription(e.target.value)}
                placeholder={t('marketplace.create_description')}
              />
            </div>
            <div>
              <Label htmlFor="create-category">{t('marketplace.create_category')}</Label>
              <Input
                id="create-category"
                value={createCategory}
                onChange={(e) => setCreateCategory(e.target.value)}
                placeholder={t('marketplace.create_category')}
              />
            </div>
            <div>
              <Label htmlFor="create-type">{t('marketplace.create_type')}</Label>
              <Input
                id="create-type"
                value={createType}
                onChange={(e) => setCreateType(e.target.value)}
                placeholder={createMode === 'plugin' ? 'processor / notifier / exporter' : 'payload / workflow / rule'}
              />
            </div>
            {createMode === 'plugin' && (
              <>
                <div>
                  <Label htmlFor="create-version">{t('marketplace.create_version')}</Label>
                  <Input
                    id="create-version"
                    value={createVersion}
                    onChange={(e) => setCreateVersion(e.target.value)}
                    placeholder="1.0.0"
                  />
                </div>
                <div>
                  <Label htmlFor="create-author">{t('marketplace.create_author')}</Label>
                  <Input
                    id="create-author"
                    value={createAuthor}
                    onChange={(e) => setCreateAuthor(e.target.value)}
                    placeholder={t('marketplace.create_author')}
                  />
                </div>
              </>
            )}
            <div>
              <Label htmlFor="create-content">{t('marketplace.create_content')}</Label>
              <Textarea
                id="create-content"
                value={createContent}
                onChange={(e) => setCreateContent(e.target.value)}
                placeholder={createMode === 'plugin' ? 'plugin code / implementation' : 'template content (YAML/JSON)'}
                rows={6}
              />
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setCreateDialogOpen(false)} disabled={creating}>
              {t('common.cancel')}
            </Button>
            <Button onClick={handleCreate} disabled={creating || !createName.trim()}>
              {creating ? t('common.saving') : t('common.create')}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Content */}
      {loading ? (
        <div className="flex items-center justify-center h-64">
          <p className="text-gray-500">Loading...</p>
        </div>
      ) : (
        <div>
          {activeTab === 'plugins' && (
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
              {plugins.length === 0 ? (
                <div className="col-span-full text-center py-12">
                  <p className="text-gray-500">{t('marketplace.no_plugins')}</p>
                </div>
              ) : (
                plugins
                  .filter((p) => !searchQuery || p.name.toLowerCase().includes(searchQuery.toLowerCase()) || p.description.toLowerCase().includes(searchQuery.toLowerCase()))
                  .map((plugin) => (
                  <div key={plugin.id} className="bg-white shadow rounded-lg p-6">
                    <h3 className="text-lg font-semibold text-gray-900 mb-2">{plugin.name}</h3>
                    <p className="text-sm text-gray-600 mb-4">{plugin.description}</p>
                    <div className="flex items-center justify-between text-sm text-gray-500 mb-4">
                      <span>v{plugin.version}</span>
                      <span>by {plugin.author}</span>
                    </div>
                    <div className="flex items-center justify-between mb-4">
                      <div className="flex items-center space-x-2">
                        <span className="text-sm">⭐ {plugin.rating}</span>
                        <span className="text-sm">↓ {plugin.downloads}</span>
                      </div>
                    </div>
                    <button
                      onClick={() => installPlugin(plugin.id)}
                      disabled={plugin.installed}
                      className={`w-full py-2 rounded-lg ${
                        plugin.installed
                          ? 'bg-gray-300 text-gray-600 cursor-not-allowed'
                          : 'bg-indigo-600 text-white hover:bg-indigo-700'
                      }`}
                    >
                      {plugin.installed ? t('marketplace.installed_badge') : t('marketplace.install')}
                    </button>
                  </div>
                ))
              )}
            </div>
          )}

          {activeTab === 'templates' && (
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
              {templates.length === 0 ? (
                <div className="col-span-full text-center py-12">
                  <p className="text-gray-500">{t('marketplace.no_templates')}</p>
                </div>
              ) : (
                templates
                  .filter((t) => !searchQuery || t.name.toLowerCase().includes(searchQuery.toLowerCase()) || t.description.toLowerCase().includes(searchQuery.toLowerCase()))
                  .map((template) => (
                  <div key={template.id} className="bg-white shadow rounded-lg p-6">
                    <h3 className="text-lg font-semibold text-gray-900 mb-2">{template.name}</h3>
                    <p className="text-sm text-gray-600 mb-4">{template.description}</p>
                    <div className="flex items-center justify-between text-sm text-gray-500 mb-4">
                      <span className="px-2 py-1 bg-gray-100 rounded">{template.category}</span>
                      <span>↓ {template.downloads}</span>
                    </div>
                    <button
                      onClick={() => installTemplate(template.id)}
                      disabled={template.installed}
                      className={`w-full py-2 rounded-lg ${
                        template.installed
                          ? 'bg-gray-300 text-gray-600 cursor-not-allowed'
                          : 'bg-indigo-600 text-white hover:bg-indigo-700'
                      }`}
                    >
                      {template.installed ? t('marketplace.installed_badge') : t('marketplace.install')}
                    </button>
                  </div>
                ))
              )}
            </div>
          )}

          {activeTab === 'installed' && (
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
              {installedPlugins.length === 0 ? (
                <div className="col-span-full text-center py-12">
                  <p className="text-gray-500">{t('marketplace.no_installed')}</p>
                </div>
              ) : (
                installedPlugins.map((plugin) => (
                  <div key={plugin.id} className="bg-white dark:bg-gray-800 shadow rounded-lg p-6">
                    <h3 className="text-lg font-semibold text-gray-900 dark:text-gray-100 mb-2">{plugin.name}</h3>
                    <p className="text-sm text-gray-600 dark:text-gray-400 mb-4">{plugin.description}</p>
                    <div className="flex items-center justify-between text-sm text-gray-500 mb-4">
                      <span>v{plugin.version}</span>
                      <span>by {plugin.author}</span>
                    </div>
                    <div className="flex items-center justify-between mb-4">
                      <div className="flex items-center space-x-2">
                        <span className="text-sm">★ {plugin.rating}</span>
                        <span className="text-sm">↓ {plugin.downloads}</span>
                      </div>
                    </div>
                    <button
                      className="w-full py-2 rounded-lg bg-gray-300 text-gray-600 cursor-not-allowed"
                      disabled
                    >
                      {t('marketplace.installed_badge')}
                    </button>
                  </div>
                ))
              )}
            </div>
          )}
        </div>
      )}
    </div>
  )
}
