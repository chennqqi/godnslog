'use client'

import { useState, useEffect, Suspense } from 'react'
import { useRouter, useSearchParams } from 'next/navigation'
import { useCases, useCreateCase } from '@/features/cases/hooks/use-cases'
import type { CaseCreateRequest } from '@/types'
import { KanbanBoard } from '@/components/kanban'
import { LayoutGrid, List } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import { Label } from '@/components/ui/label'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Badge } from '@/components/ui/badge'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { caseSchema, type CaseFormValues } from '@/features/cases/schemas/case-schema'
import { useI18n } from '@/lib/i18n-context'

/** Sentinel for Radix Select: empty string is reserved for clearing selection */
const STATUS_FILTER_ALL = 'all'

function CasesPageContent() {
  const router = useRouter()
  const searchParams = useSearchParams()
  const { t } = useI18n()
  const [showCreateModal, setShowCreateModal] = useState(false)
  const [searchTerm, setSearchTerm] = useState(searchParams.get('search') ?? '')
  const [statusFilter, setStatusFilter] = useState(searchParams.get('status') ?? STATUS_FILTER_ALL)
  const [viewMode, setViewMode] = useState<'table' | 'board'>((searchParams.get('view') as 'table' | 'board') ?? 'table')
  const [error, setError] = useState('')

  const { data, isLoading: loading } = useCases({
    page: 1,
    page_size: 50,
    search: searchTerm,
    ...(statusFilter !== STATUS_FILTER_ALL ? { status: statusFilter } : {}),
  })
  const createCase = useCreateCase()
  const cases = data?.data?.items ?? []

  // Sync filters to URL
  useEffect(() => {
    const params = new URLSearchParams()
    if (searchTerm) params.set('search', searchTerm)
    if (statusFilter !== STATUS_FILTER_ALL) params.set('status', statusFilter)
    if (viewMode !== 'table') params.set('view', viewMode)
    const qs = params.toString()
    router.replace(`/cases${qs ? `?${qs}` : ''}`, { scroll: false })
  }, [searchTerm, statusFilter, viewMode, router])

  const form = useForm<CaseFormValues>({
    resolver: zodResolver(caseSchema),
    defaultValues: {
      title: '',
      description: '',
      target: '',
      status: 'active',
      tags: [],
    },
  })

  const handleCreateCase = async (values: CaseFormValues) => {
    try {
      setError('')
      await createCase.mutateAsync(values as CaseCreateRequest)
      setShowCreateModal(false)
      form.reset()
    } catch (error) {
      console.error('Failed to create case:', error)
      setError(t('cases.create_failed'))
    }
  }

  if (loading) {
    return <div className="text-center py-12 text-gray-500">{t('cases.loading')}</div>
  }

  return (
    <div className="space-y-4">
      <div className="flex justify-between items-center">
        <div>
          <h2 className="text-xl font-bold text-gray-900 dark:text-gray-100">{t('cases.title')}</h2>
          <p className="text-sm text-gray-500 dark:text-gray-400">{t('cases.subtitle')}</p>
        </div>
        <Button onClick={() => setShowCreateModal(true)}>
          {t('cases.new')}
        </Button>
      </div>

      {error && (
        <div className="bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-md p-4 flex items-start gap-3">
          <svg className="w-5 h-5 text-red-500 mt-0.5 shrink-0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
          <p className="text-sm text-red-700 dark:text-red-300">{error}</p>
        </div>
      )}

      {/* Search and Filter */}
      <div className="bg-white dark:bg-gray-800 shadow rounded-lg p-4 border border-gray-200 dark:border-gray-700">
        <div className="flex flex-wrap gap-3">
          <Input
            placeholder={t('cases.search')}
            className="flex-1 min-w-[160px]"
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
          />
          <Select value={statusFilter} onValueChange={setStatusFilter}>
            <SelectTrigger className="w-[160px]">
              <SelectValue placeholder={t('cases.all_statuses')} />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value={STATUS_FILTER_ALL}>{t('cases.all_statuses')}</SelectItem>
              <SelectItem value="active">{t('cases.active')}</SelectItem>
              <SelectItem value="completed">{t('cases.completed')}</SelectItem>
              <SelectItem value="archived">{t('cases.archived')}</SelectItem>
            </SelectContent>
          </Select>
          <div className="flex items-center gap-1 border border-gray-200 dark:border-gray-600 rounded-md p-0.5">
            <button
              type="button"
              onClick={() => setViewMode('table')}
              className={`flex items-center gap-1 px-2 py-1 text-xs rounded ${
                viewMode === 'table'
                  ? 'bg-indigo-600 text-white'
                  : 'text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-300'
              }`}
            >
              <List className="w-3.5 h-3.5" />
              {t('cases.view_table')}
            </button>
            <button
              type="button"
              onClick={() => setViewMode('board')}
              className={`flex items-center gap-1 px-2 py-1 text-xs rounded ${
                viewMode === 'board'
                  ? 'bg-indigo-600 text-white'
                  : 'text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-300'
              }`}
            >
              <LayoutGrid className="w-3.5 h-3.5" />
              {t('cases.view_board')}
            </button>
          </div>
        </div>
      </div>

      {viewMode === 'board' ? (
        <KanbanBoard
          cases={cases}
          onCardClick={(id) => router.push(`/cases/${id}`)}
          labels={{
            active: t('cases.board_active'),
            completed: t('cases.board_completed'),
            archived: t('cases.board_archived'),
            empty: t('cases.board_empty'),
          }}
        />
      ) : (
      <div className="bg-white dark:bg-gray-800 shadow rounded-lg border border-gray-200 dark:border-gray-700">
        <div className="px-4 py-5 sm:p-6">
          {cases.length === 0 ? (
            <div className="text-center py-10">
              <div className="text-4xl mb-3">📂</div>
              <p className="text-sm font-medium text-gray-700 dark:text-gray-300">{t('cases.no_data')}</p>
              <p className="text-xs text-gray-400 mt-1">{t('cases.no_data_hint')}</p>
            </div>
          ) : (
            <ul className="divide-y divide-gray-200 dark:divide-gray-700">
              <li className="py-2 flex items-center text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wide">
                <span className="flex-1">{t('cases.col_title')}</span>
                <span className="w-24">{t('cases.col_status')}</span>
                <span className="w-32">{t('cases.col_created')}</span>
              </li>
              {cases.map((case_) => (
                <li
                  key={case_.id}
                  className="py-4 flex items-center hover:bg-gray-50 dark:hover:bg-gray-700/30 transition-colors cursor-pointer"
                  onClick={() => router.push(`/cases/${case_.id}`)}
                >
                  <div className="flex-1">
                    <p className="text-sm font-medium text-indigo-600 dark:text-indigo-400">{case_.title}</p>
                    <p className="text-sm text-gray-500 dark:text-gray-400">{case_.description}</p>
                    {case_.target && (
                      <p className="text-xs text-gray-400 mt-1">{t('cases.target')}: {case_.target}</p>
                    )}
                  </div>
                  <div className="w-24">
                    <Badge variant={
                      case_.status === 'active' ? 'default' :
                      case_.status === 'completed' ? 'secondary' :
                      'outline'
                    }>
                      {case_.status}
                    </Badge>
                  </div>
                  <div className="w-32 text-xs text-gray-400">
                    {new Date(case_.created_at).toLocaleDateString()}
                  </div>
                </li>
              ))}
            </ul>
          )}
        </div>
      </div>
      )}

      {/* Create Modal */}
      <Dialog open={showCreateModal} onOpenChange={(open) => {
        setShowCreateModal(open)
        if (!open) form.reset()
      }}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{t('cases.new')}</DialogTitle>
          </DialogHeader>
          <form onSubmit={form.handleSubmit(handleCreateCase)}>
            <div className="mb-4">
              <Label htmlFor="title">{t('cases.col_title')}</Label>
              <Input
                id="title"
                {...form.register('title')}
              />
              {form.formState.errors.title && (
                <p className="text-xs text-red-500 mt-1">{form.formState.errors.title.message}</p>
              )}
            </div>
            <div className="mb-4">
              <Label htmlFor="description">{t('cases.description')}</Label>
              <Textarea
                id="description"
                rows={3}
                {...form.register('description')}
              />
            </div>
            <div className="mb-4">
              <Label htmlFor="target">{t('cases.target')}</Label>
              <Input
                id="target"
                placeholder="e.g. internal-api.corp.com"
                {...form.register('target')}
              />
            </div>
            <DialogFooter>
              <Button type="button" variant="outline" onClick={() => setShowCreateModal(false)}>
                {t('cases.cancel')}
              </Button>
              <Button type="submit">
                {t('cases.create')}
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>
    </div>
  )
}

export default function CasesPage() {
  return (
    <Suspense fallback={<div className="text-center py-12 text-gray-500">Loading...</div>}>
      <CasesPageContent />
    </Suspense>
  )
}
