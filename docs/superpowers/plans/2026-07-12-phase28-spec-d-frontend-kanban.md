# Phase 2.8 Spec D: Frontend Kanban Board Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a Kanban board view to the Cases page with drag-and-drop between status columns (active / completed / archived), backed by the existing `caseApi.update` partial-update endpoint.

**Architecture:** Three new components (`KanbanBoard`, `KanbanColumn`, `KanbanCard`) using `@dnd-kit/core` + `@dnd-kit/sortable`. The Cases page gets a Table/Board view toggle. Drag-end calls `useUpdateCase` with the new status; optimistic local state update with rollback on error.

**Tech Stack:** @dnd-kit/core + @dnd-kit/sortable + @dnd-kit/utilities, Next.js 16, React 18, Tailwind, React Query (TanStack), react-hook-form (existing create modal).

## Global Constraints

- Frontend dir: `frontend-next/`
- No frontend unit test runner - verify via `npm run build` + manual browser drag-and-drop
- Case statuses are exactly: `active`, `completed`, `archived` (3 columns, fixed order)
- Backend `CaseUpdateRequest.Status` accepts `oneof=active archived completed` - partial `{status}` updates are valid
- All new UI strings via i18n (`useI18n`) with both `en-US` and `zh-CN` keys
- Kanban is an *additional* view - the existing table view must remain via a toggle

---

## File Structure

| File | Responsibility |
|------|----------------|
| `frontend-next/package.json` | Add @dnd-kit dependencies |
| `frontend-next/src/components/kanban/kanban-card.tsx` | New: draggable case card |
| `frontend-next/src/components/kanban/kanban-column.tsx` | New: droppable status column |
| `frontend-next/src/components/kanban/kanban-board.tsx` | New: DnD context + 3 columns + drag-end logic |
| `frontend-next/src/components/kanban/index.ts` | New: barrel export |
| `frontend-next/src/app/dashboard/cases/page.tsx` | Add view toggle + board render |
| `frontend-next/src/lib/i18n-context.tsx` | Add kanban i18n keys |

---

### Task 1: Install @dnd-kit

**Files:**
- Modify: `frontend-next/package.json`

- [ ] **Step 1: Install the packages**

Run:
```bash
cd frontend-next && npm install @dnd-kit/core @dnd-kit/sortable @dnd-kit/utilities
```

- [ ] **Step 2: Verify build still passes**

Run: `cd frontend-next && npm run build 2>&1 | tail -20`
Expected: build succeeds.

- [ ] **Step 3: Commit**

```bash
git add frontend-next/package.json frontend-next/package-lock.json
git commit -m "chore(frontend): add @dnd-kit dependencies for drag-and-drop"
```

---

### Task 2: Create KanbanCard component

**Files:**
- Create: `frontend-next/src/components/kanban/kanban-card.tsx`

**Interfaces:**
- Produces: `KanbanCard` with props `{ case: Case, onClick?: (id) => void }`

- [ ] **Step 1: Create the component**

Create `frontend-next/src/components/kanban/kanban-card.tsx`:

```tsx
'use client'

import { useSortable } from '@dnd-kit/sortable'
import { CSS } from '@dnd-kit/utilities'
import { Calendar } from 'lucide-react'
import type { Case } from '@/types'

export interface KanbanCardProps {
  case: Case
  onClick?: (id: string) => void
}

export function KanbanCard({ case: c, onClick }: KanbanCardProps) {
  const { attributes, listeners, setNodeRef, transform, transition, isDragging } = useSortable({
    id: c.id,
  })

  const style = {
    transform: CSS.Transform.toString(transform),
    transition,
    opacity: isDragging ? 0.5 : 1,
  }

  return (
    <div
      ref={setNodeRef}
      style={style}
      {...attributes}
      {...listeners}
      onClick={() => onClick?.(c.id)}
      className="bg-white dark:bg-gray-700 border border-gray-200 dark:border-gray-600 rounded-md p-3 cursor-grab active:cursor-grabbing hover:shadow-md transition-shadow"
    >
      <p className="font-medium text-sm text-gray-900 dark:text-gray-100 line-clamp-2 mb-1">
        {c.title}
      </p>
      {c.target && (
        <p className="text-xs text-gray-500 dark:text-gray-400 truncate mb-2">{c.target}</p>
      )}
      <div className="flex items-center justify-between">
        <div className="flex flex-wrap gap-1">
          {c.tags?.slice(0, 2).map((tag) => (
            <span
              key={tag}
              className="text-xs px-1.5 py-0.5 rounded bg-gray-100 dark:bg-gray-600 text-gray-600 dark:text-gray-300"
            >
              {tag}
            </span>
          ))}
        </div>
        <span className="flex items-center gap-1 text-xs text-gray-400">
          <Calendar className="w-3 h-3" />
          {new Date(c.created_at).toLocaleDateString()}
        </span>
      </div>
    </div>
  )
}
```

- [ ] **Step 2: Verify typecheck**

Run: `cd frontend-next && npx tsc --noEmit 2>&1 | grep -i kanban`
Expected: no errors (the import of `Case` from `@/types` resolves since the type exists).

- [ ] **Step 3: Commit**

```bash
git add frontend-next/src/components/kanban/kanban-card.tsx
git commit -m "feat(frontend): add KanbanCard draggable component"
```

---

### Task 3: Create KanbanColumn component

**Files:**
- Create: `frontend-next/src/components/kanban/kanban-column.tsx`

**Interfaces:**
- Consumes: `Case` type, `KanbanCard` from Task 2
- Produces: `KanbanColumn` with props `{ id: string, title: string, accentColor: string, cases: Case[], onCardClick? }`

- [ ] **Step 1: Create the component**

Create `frontend-next/src/components/kanban/kanban-column.tsx`:

```tsx
'use client'

import { useDroppable } from '@dnd-kit/core'
import type { Case } from '@/types'
import { KanbanCard } from './kanban-card'

export interface KanbanColumnProps {
  id: string
  title: string
  accentColor: string
  cases: Case[]
  onCardClick?: (id: string) => void
  emptyLabel?: string
}

export function KanbanColumn({
  id,
  title,
  accentColor,
  cases,
  onCardClick,
  emptyLabel = 'No cases',
}: KanbanColumnProps) {
  const { setNodeRef, isOver } = useDroppable({ id })

  return (
    <div className="flex flex-col min-w-[260px] w-72 bg-gray-50 dark:bg-gray-800/50 rounded-lg p-3">
      <div className="flex items-center gap-2 mb-3 px-1">
        <span className="w-2.5 h-2.5 rounded-full" style={{ backgroundColor: accentColor }} />
        <h3 className="text-sm font-semibold text-gray-700 dark:text-gray-300">{title}</h3>
        <span className="ml-auto text-xs text-gray-400">{cases.length}</span>
      </div>
      <div
        ref={setNodeRef}
        className={`flex-1 space-y-2 min-h-[120px] rounded-md p-1 transition-colors ${
          isOver ? 'bg-indigo-50 dark:bg-indigo-900/30' : ''
        }`}
      >
        {cases.length === 0 ? (
          <p className="text-xs text-gray-400 text-center py-8">{emptyLabel}</p>
        ) : (
          cases.map((c) => <KanbanCard key={c.id} case={c} onClick={onCardClick} />)
        )}
      </div>
    </div>
  )
}
```

- [ ] **Step 2: Verify typecheck**

Run: `cd frontend-next && npx tsc --noEmit 2>&1 | grep -i kanban`
Expected: no errors.

- [ ] **Step 3: Commit**

```bash
git add frontend-next/src/components/kanban/kanban-column.tsx
git commit -m "feat(frontend): add KanbanColumn droppable component"
```

---

### Task 4: Create KanbanBoard component

**Files:**
- Create: `frontend-next/src/components/kanban/kanban-board.tsx`
- Create: `frontend-next/src/components/kanban/index.ts`

**Interfaces:**
- Consumes: `Case` type, `KanbanColumn` from Task 3, `useUpdateCase` hook
- Produces: `KanbanBoard` with props `{ cases: Case[], onCardClick?: (id) => void }`

- [ ] **Step 1: Create the component**

Create `frontend-next/src/components/kanban/kanban-board.tsx`:

```tsx
'use client'

import { useMemo, useState } from 'react'
import {
  DndContext,
  DragOverlay,
  PointerSensor,
  useSensor,
  useSensors,
  closestCorners,
  type DragEndEvent,
  type DragStartEvent,
} from '@dnd-kit/core'
import { useUpdateCase } from '@/features/cases/hooks/use-cases'
import type { Case } from '@/types'
import { KanbanColumn } from './kanban-column'
import { KanbanCard } from './kanban-card'

export interface KanbanBoardProps {
  cases: Case[]
  onCardClick?: (id: string) => void
  labels?: {
    active?: string
    completed?: string
    archived?: string
    empty?: string
  }
}

const COLUMNS = [
  { id: 'active', accentColor: '#3b82f6' },
  { id: 'completed', accentColor: '#10b981' },
  { id: 'archived', accentColor: '#6b7280' },
] as const

export function KanbanBoard({ cases, onCardClick, labels }: KanbanBoardProps) {
  const updateCase = useUpdateCase()
  const [activeId, setActiveId] = useState<string | null>(null)
  const [localCases, setLocalCases] = useState<Case[]>(cases)

  // Sync external cases prop into local state for optimistic updates.
  useMemo(() => {
    setLocalCases(cases)
  }, [cases])

  const sensors = useSensors(
    useSensor(PointerSensor, { activationConstraint: { distance: 5 } })
  )

  const activeCase = activeId ? localCases.find((c) => c.id === activeId) : null

  const grouped = useMemo(() => {
    const g: Record<string, Case[]> = { active: [], completed: [], archived: [] }
    for (const c of localCases) {
      if (g[c.status]) g[c.status].push(c)
    }
    return g
  }, [localCases])

  function handleDragStart(event: DragStartEvent) {
    setActiveId(String(event.active.id))
  }

  function handleDragEnd(event: DragEndEvent) {
    setActiveId(null)
    const { active, over } = event
    if (!over) return

    const newStatus = String(over.id) // column id == status
    const movedCase = localCases.find((c) => c.id === String(active.id))
    if (!movedCase || movedCase.status === newStatus) return

    // Optimistic update.
    setLocalCases((prev) =>
      prev.map((c) => (c.id === movedCase.id ? { ...c, status: newStatus as Case['status'] } : c))
    )

    // Persist; rollback on error.
    updateCase.mutate(
      { id: movedCase.id, data: { status: newStatus } },
      {
        onError: () => {
          setLocalCases((prev) =>
            prev.map((c) => (c.id === movedCase.id ? { ...c, status: movedCase.status } : c))
          )
        },
      }
    )
  }

  return (
    <DndContext
      sensors={sensors}
      collisionDetection={closestCorners}
      onDragStart={handleDragStart}
      onDragEnd={handleDragEnd}
    >
      <div className="flex gap-4 overflow-x-auto pb-2">
        {COLUMNS.map((col) => (
          <KanbanColumn
            key={col.id}
            id={col.id}
            title={labels?.[col.id] ?? col.id}
            accentColor={col.accentColor}
            cases={grouped[col.id]}
            onCardClick={onCardClick}
            emptyLabel={labels?.empty}
          />
        ))}
      </div>
      <DragOverlay>
        {activeCase ? <KanbanCard case={activeCase} /> : null}
      </DragOverlay>
    </DndContext>
  )
}
```

- [ ] **Step 2: Create barrel export**

Create `frontend-next/src/components/kanban/index.ts`:

```ts
export { KanbanBoard } from './kanban-board'
export type { KanbanBoardProps } from './kanban-board'
export { KanbanColumn } from './kanban-column'
export { KanbanCard } from './kanban-card'
```

- [ ] **Step 3: Verify typecheck**

Run: `cd frontend-next && npx tsc --noEmit 2>&1 | grep -i kanban`
Expected: no errors.

- [ ] **Step 4: Commit**

```bash
git add frontend-next/src/components/kanban/kanban-board.tsx frontend-next/src/components/kanban/index.ts
git commit -m "feat(frontend): add KanbanBoard with optimistic drag-and-drop status updates"
```

---

### Task 5: Integrate Kanban into Cases page with view toggle

**Files:**
- Modify: `frontend-next/src/app/dashboard/cases/page.tsx`
- Modify: `frontend-next/src/lib/i18n-context.tsx`

**Interfaces:**
- Consumes: `KanbanBoard` from Task 4, `useCases`/`useUpdateCase` (existing), i18n keys

- [ ] **Step 1: Add i18n keys**

In `frontend-next/src/lib/i18n-context.tsx`, add to the `cases.*` namespace in BOTH the `en-US` and `zh-CN` translation objects:

```ts
// en-US
cases.view_table: 'Table'
cases.view_board: 'Board'
cases.board_active: 'Active'
cases.board_completed: 'Completed'
cases.board_archived: 'Archived'
cases.board_empty: 'No cases'
// zh-CN
cases.view_table: '列表'
cases.view_board: '看板'
cases.board_active: '进行中'
cases.board_completed: '已完成'
cases.board_archived: '已归档'
cases.board_empty: '暂无用例'
```

- [ ] **Step 2: Add view-mode state and toggle UI**

In `frontend-next/src/app/dashboard/cases/page.tsx`:

Add imports at the top:
```tsx
import { KanbanBoard } from '@/components/kanban'
import { LayoutGrid, List } from 'lucide-react'
```

Inside the component (after the existing `statusFilter` state around line 38), add:
```tsx
const [viewMode, setViewMode] = useState<'table' | 'board'>('table')
```

In the search/filter card (around lines 88-108), add a view toggle next to the filter controls. After the status `<Select>`, add:

```tsx
<div className="flex items-center gap-1 border border-gray-200 dark:border-gray-600 rounded-md p-0.5">
  <button
    type="button"
    onClick={() => setViewMode('table')}
    className={`flex items-center gap-1 px-2 py-1 text-xs rounded ${
      viewMode === 'table'
        ? 'bg-indigo-600 text-white'
        : 'text-gray-500 dark:text-gray-400'
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
        : 'text-gray-500 dark:text-gray-400'
    }`}
  >
    <LayoutGrid className="w-3.5 h-3.5" />
    {t('cases.view_board')}
  </button>
</div>
```

- [ ] **Step 3: Conditionally render board vs table**

Find the cases list `<ul>` block (around lines 110-155). Wrap it in a conditional so the board shows when `viewMode === 'board'`:

```tsx
{viewMode === 'board' ? (
  <KanbanBoard
    cases={cases}
    onCardClick={(id) => router.push(`/dashboard/cases/${id}`)}
    labels={{
      active: t('cases.board_active'),
      completed: t('cases.board_completed'),
      archived: t('cases.board_archived'),
      empty: t('cases.board_empty'),
    }}
  />
) : (
  <ul className="divide-y divide-gray-100 dark:divide-gray-700">
    {/* existing list items stay here unchanged */}
  </ul>
)}
```

Ensure `router` from `useRouter()` is available - check the top of the file; if not present, add `const router = useRouter()` and import `useRouter` from `next/navigation`.

- [ ] **Step 4: Build the frontend**

Run: `cd frontend-next && npm run build 2>&1 | tail -30`
Expected: build succeeds.

- [ ] **Step 5: Browser verification**

Start backend + frontend dev servers (same as Plan C Task 6 Step 6). On `/dashboard/cases`:

1. Default view is the table (existing behavior unchanged).
2. Click the "Board" toggle - three columns render (Active / Completed / Archived) with case cards grouped by status.
3. Drag a case card from one column to another:
   - Card visually moves to the new column.
   - After drop, the case's status updates (refresh the page to confirm persistence - it should stay in the new column).
   - If you can simulate a failure, the card rolls back to the original column.
4. Click a card in board view - navigates to `/dashboard/cases/[id]`.
5. Empty columns show the "No cases" placeholder.
6. Switch back to "Table" - existing list view returns.
7. No console errors.

Create a couple of cases with different statuses first (via the New Case modal) if the board is empty.

If you cannot run the browser, say so explicitly - do not claim success from build alone.

- [ ] **Step 6: Commit**

```bash
git add frontend-next/src/app/dashboard/cases/page.tsx frontend-next/src/lib/i18n-context.tsx
git commit -m "feat(cases): add Kanban board view with table/board toggle"
```

---

## Self-Review

**Spec coverage:**
- ✅ KanbanBoard 看板组件 -> Task 4 (+ Task 2 card, Task 3 column)
- ✅ Case 管理 3 列 active/completed/archived -> Task 4 COLUMNS constant
- ✅ 拖拽切换状态自动调用 API -> Task 4 handleDragEnd calls useUpdateCase
- ✅ Table/Board 视图切换 -> Task 5
- ✅ 乐观更新 + 回滚 -> Task 4 onError rollback
- ✅ i18n keys -> Task 5 Step 1

**Placeholder scan:** None - all code complete. router, viewMode, labels all wired.

**Type consistency:** `Case['status']` is `'active' | 'completed' | 'archived'` (matches types/index.ts); `KanbanBoardProps.labels` keys match COLUMNS ids; `useUpdateCase` mutation signature `{id, data: CaseUpdateRequest}` matches Task 4 usage; `CaseUpdateRequest.status` accepted by backend `oneof=active archived completed`.
