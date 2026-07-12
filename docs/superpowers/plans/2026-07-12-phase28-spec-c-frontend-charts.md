# Phase 2.8 Spec C: Frontend Charts Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add reusable DonutChart and LineChart components using recharts, back them with a new daily-stats API endpoint, and integrate them into the dashboard (replacing the CSS-only protocol bar and adding a 7-day trend chart).

**Architecture:** New backend handler `GET /api/v2/interactions/stats/daily` groups interactions by UTC date. New recharts-based `DonutChart` and `LineChart` components live in `frontend-next/src/components/charts/`. Dashboard consumes `by_type` (existing) for the donut and the new daily endpoint for the line chart.

**Tech Stack:** recharts (React charting), Next.js 16, React 18, Tailwind, Go 1.25 backend (xorm), Gin.

## Global Constraints

- Frontend dir: `frontend-next/`
- Backend test command: `GOCACHE=/tmp/gocache go test ./...`
- Frontend has no unit test runner - verify via `npm run build` + manual browser check
- recharts version: latest stable (peer-compatible with React 18)
- Protocol color map must be consistent across DonutChart and any legend
- All new UI strings go through i18n (`useI18n`) with both `en-US` and `zh-CN` keys

---

## File Structure

| File | Responsibility |
|------|----------------|
| `server/v2_api.go` | New `v2InteractionDailyStats` handler |
| `server/webserver.go` | Register `/interactions/stats/daily` route |
| `server/v2_api_test.go` | New handler test |
| `frontend-next/package.json` | Add `recharts` dependency |
| `frontend-next/src/components/charts/donut-chart.tsx` | New: reusable donut/ring chart |
| `frontend-next/src/components/charts/line-chart.tsx` | New: reusable line chart |
| `frontend-next/src/components/charts/index.ts` | New: barrel export |
| `frontend-next/src/lib/api-client.ts` | Add `interactionApi.dailyStats` |
| `frontend-next/src/types/index.ts` | Add `DailyStat` type |
| `frontend-next/src/app/dashboard/page.tsx` | Replace ProtocolBar with DonutChart; add LineChart trend |
| `frontend-next/src/lib/i18n-context.tsx` | Add chart-related i18n keys |

---

### Task 1: Add backend daily-stats endpoint

**Files:**
- Modify: `server/v2_api.go` (add handler near line 2171, after `v2InteractionStats`)
- Modify: `server/webserver.go` (register route near line 89)
- Test: `server/v2_api_test.go`

**Interfaces:**
- Produces: `GET /api/v2/interactions/stats/daily` returning `{data: [{date, count}, ...]}`

- [ ] **Step 1: Write the failing test**

Add to `server/v2_api_test.go` (follow existing test setup patterns in that file - look for an existing `setupTestServer` or similar helper near the top):

```go
func TestV2InteractionDailyStats(t *testing.T) {
	srv := newTestWebServer(t)
	// Insert interactions on 3 distinct UTC dates.
	now := time.Now().UTC()
	dates := []time.Time{
		time.Date(now.Year(), now.Month(), now.Day(), 1, 0, 0, 0, time.UTC),
		time.Date(now.Year(), now.Month(), now.Day()-1, 2, 0, 0, 0, time.UTC),
		time.Date(now.Year(), now.Month(), now.Day()-2, 3, 0, 0, 0, time.UTC),
	}
	for _, ts := range dates {
		for i := 0; i < 2; i++ {
			inter := &v2models.Interaction{
				ID:        v2models.GenerateID(),
				Type:      "dns",
				Timestamp: ts,
				SourceIP:  "1.2.3.4",
			}
			if _, err := srv.orm.InsertOne(inter); err != nil {
				t.Fatalf("insert: %v", err)
			}
		}
	}

	req := httptest.NewRequest("GET", "/api/v2/interactions/stats/daily?days=7", nil)
	w := httptest.NewRecorder()
	srv.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body: %s", w.Code, w.Body.String())
	}
	var resp struct {
		Code int `json:"code"`
		Data []struct {
			Date  string `json:"date"`
			Count int64  `json:"count"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if resp.Code != 0 {
		t.Fatalf("code = %d, want 0", resp.Code)
	}
	// Expect at least 3 non-zero days in the last 7 days.
	if len(resp.Data) < 3 {
		t.Errorf("expected >= 3 date buckets, got %d: %+v", len(resp.Data), resp.Data)
	}
}
```

Note: check `newTestWebServer` / `srv.router` field names by reading the top of `server/v2_api_test.go` - adapt the helper name to whatever the file already uses. If no such helper exists, follow the pattern of an existing test that constructs a `WebServer` and inserts interactions.

- [ ] **Step 2: Run test to verify it fails**

Run: `GOCACHE=/tmp/gocache go test ./server/... -run TestV2InteractionDailyStats -v`
Expected: FAIL - route not registered (404).

- [ ] **Step 3: Implement the handler**

In `server/v2_api.go`, add after `v2InteractionStats` (after line 2171):

```go
// v2InteractionDailyStats returns daily interaction counts for the last N days.
// @Summary Get daily interaction stats
// @Tags interactions
// @Produce json
// @Param case_id query string false "filter by case"
// @Param payload_id query string false "filter by payload"
// @Param days query int false "number of days (default 7, max 90)"
// @Success 200 {object} map[string]interface{}
// @Router /api/v2/interactions/stats/daily [get]
func (self *WebServer) v2InteractionDailyStats(c *gin.Context) {
	caseId := c.Query("case_id")
	payloadId := c.Query("payload_id")
	days := 7
	if d, err := strconv.Atoi(c.Query("days")); err == nil && d > 0 && d <= 90 {
		days = d
	}

	now := time.Now().UTC()
	start := time.Date(now.Year(), now.Month(), now.Day()-days+1, 0, 0, 0, 0, time.UTC)

	session := self.orm.NewSession()
	defer session.Close()
	query := session.Table(new(v2models.Interaction)).Where("timestamp >= ?", start)
	if caseId != "" {
		query = query.Where("case_id = ?", caseId)
	}
	if payloadId != "" {
		query = query.Where("payload_id = ?", payloadId)
	}

	type dailyStat struct {
		Date  string `xorm:"date" json:"date"`
		Count int64  `xorm:"count" json:"count"`
	}
	var rows []dailyStat
	// SQLite strftime, MySQL DATE() - use xorm raw SQL via SQL() for portability.
	// Use DATE(timestamp) which works on both SQLite and MySQL.
	if err := query.Select("DATE(timestamp) as date, count(*) as count").
		GroupBy("DATE(timestamp)").OrderBy("date ASC").Find(&rows); err != nil {
		logrus.Errorf("[v2_api.go::v2InteractionDailyStats] query error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "server internal error"})
		return
	}

	// Fill missing days with zero counts for a continuous chart.
	countMap := make(map[string]int64, len(rows))
	for _, r := range rows {
		countMap[r.Date] = r.Count
	}
	result := make([]gin.H, 0, days)
	for i := 0; i < days; i++ {
		day := time.Date(now.Year(), now.Month(), now.Day()-days+1+i, 0, 0, 0, 0, time.UTC)
		dateStr := day.Format("2006-01-02")
		result = append(result, gin.H{"date": dateStr, "count": countMap[dateStr]})
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    result,
	})
}
```

Ensure `strconv` is imported in `v2_api.go` (it likely already is - check the import block).

- [ ] **Step 4: Register the route**

In `server/webserver.go`, find the interactions route group (around line 88-89 where `/stats` is registered) and add the daily route **before** `/:id` to avoid path capture:

```go
			interactions.GET("/stats/daily", self.v2InteractionDailyStats)
			interactions.GET("/stats", self.v2InteractionStats)
```

- [ ] **Step 5: Run test to verify it passes**

Run: `GOCACHE=/tmp/gocache go test ./server/... -run TestV2InteractionDailyStats -v`
Expected: PASS.

- [ ] **Step 6: Run full server test suite**

Run: `GOCACHE=/tmp/gocache go test ./server/...`
Expected: PASS.

- [ ] **Step 7: Commit**

```bash
git add server/v2_api.go server/webserver.go server/v2_api_test.go
git commit -m "feat(api): add GET /api/v2/interactions/stats/daily endpoint for trend charts"
```

---

### Task 2: Install recharts

**Files:**
- Modify: `frontend-next/package.json`

- [ ] **Step 1: Install recharts**

Run:
```bash
cd frontend-next && npm install recharts
```

- [ ] **Step 2: Verify it installs without peer dep errors**

Run: `cd frontend-next && npm run build 2>&1 | tail -20`
Expected: build succeeds (or only pre-existing warnings).

- [ ] **Step 3: Commit**

```bash
git add frontend-next/package.json frontend-next/package-lock.json
git commit -m "chore(frontend): add recharts charting dependency"
```

---

### Task 3: Create DonutChart component

**Files:**
- Create: `frontend-next/src/components/charts/donut-chart.tsx`
- Create: `frontend-next/src/components/charts/index.ts`

**Interfaces:**
- Produces: `DonutChart` component with props `{ data: {name, value, color?}[], size?, innerRadius?, outerRadius? }`

- [ ] **Step 1: Create the component**

Create `frontend-next/src/components/charts/donut-chart.tsx`:

```tsx
'use client'

import { PieChart, Pie, Cell, ResponsiveContainer, Tooltip } from 'recharts'

export interface DonutChartDataPoint {
  name: string
  value: number
  color?: string
}

export interface DonutChartProps {
  data: DonutChartDataPoint[]
  size?: number
  innerRadius?: number
  outerRadius?: number
  className?: string
}

const DEFAULT_COLORS = ['#3b82f6', '#10b981', '#f59e0b', '#8b5cf6', '#ec4899', '#f97316', '#6b7280']

export function DonutChart({
  data,
  size = 200,
  innerRadius = 60,
  outerRadius = 80,
  className,
}: DonutChartProps) {
  const hasData = data.some((d) => d.value > 0)
  if (!hasData) {
    return (
      <div
        className={`flex items-center justify-center text-sm text-gray-400 ${className ?? ''}`}
        style={{ height: size }}
      >
        No data
      </div>
    )
  }
  return (
    <div className={className} style={{ height: size }}>
      <ResponsiveContainer width="100%" height="100%">
        <PieChart>
          <Pie
            data={data}
            dataKey="value"
            nameKey="name"
            innerRadius={innerRadius}
            outerRadius={outerRadius}
            paddingAngle={2}
          >
            {data.map((entry, idx) => (
              <Cell
                key={entry.name}
                fill={entry.color ?? DEFAULT_COLORS[idx % DEFAULT_COLORS.length]}
              />
            ))}
          </Pie>
          <Tooltip
            formatter={(value: number, name: string) => [value, name]}
            contentStyle={{ backgroundColor: '#1f2937', border: 'none', borderRadius: 4, color: '#fff' }}
          />
        </PieChart>
      </ResponsiveContainer>
    </div>
  )
}
```

- [ ] **Step 2: Create barrel export**

Create `frontend-next/src/components/charts/index.ts`:

```ts
export { DonutChart } from './donut-chart'
export type { DonutChartDataPoint, DonutChartProps } from './donut-chart'
export { LineChart } from './line-chart'
export type { LineChartDataPoint, LineChartProps } from './line-chart'
```

- [ ] **Step 3: Verify typecheck**

Run: `cd frontend-next && npx tsc --noEmit 2>&1 | grep -i 'donut\|charts/index'`
Expected: no errors mentioning donut-chart (LineChart reference in index.ts will error until Task 4 - that's expected; proceed).

- [ ] **Step 4: Commit**

```bash
git add frontend-next/src/components/charts/donut-chart.tsx frontend-next/src/components/charts/index.ts
git commit -m "feat(frontend): add reusable DonutChart component with recharts"
```

---

### Task 4: Create LineChart component

**Files:**
- Create: `frontend-next/src/components/charts/line-chart.tsx`

**Interfaces:**
- Produces: `LineChart` component with props `{ data: {date, count}[], height?, color? }`

- [ ] **Step 1: Create the component**

Create `frontend-next/src/components/charts/line-chart.tsx`:

```tsx
'use client'

import {
  LineChart as RechartsLineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
} from 'recharts'

export interface LineChartDataPoint {
  date: string
  count: number
}

export interface LineChartProps {
  data: LineChartDataPoint[]
  height?: number
  color?: string
  className?: string
}

export function LineChart({
  data,
  height = 300,
  color = '#3b82f6',
  className,
}: LineChartProps) {
  const hasData = data.some((d) => d.count > 0)
  if (!hasData) {
    return (
      <div
        className={`flex items-center justify-center text-sm text-gray-400 ${className ?? ''}`}
        style={{ height }}
      >
        No data
      </div>
    )
  }
  return (
    <div className={className} style={{ height }}>
      <ResponsiveContainer width="100%" height="100%">
        <RechartsLineChart data={data} margin={{ top: 8, right: 16, left: 0, bottom: 8 }}>
          <CartesianGrid strokeDasharray="3 3" stroke="#374151" vertical={false} />
          <XAxis
            dataKey="date"
            stroke="#9ca3af"
            tick={{ fontSize: 11 }}
            tickFormatter={(v: string) => v.slice(5)}
          />
          <YAxis stroke="#9ca3af" tick={{ fontSize: 11 }} allowDecimals={false} />
          <Tooltip
            contentStyle={{ backgroundColor: '#1f2937', border: 'none', borderRadius: 4, color: '#fff' }}
          />
          <Line
            type="monotone"
            dataKey="count"
            stroke={color}
            strokeWidth={2}
            dot={{ r: 3, fill: color }}
            activeDot={{ r: 5 }}
          />
        </RechartsLineChart>
      </ResponsiveContainer>
    </div>
  )
}
```

- [ ] **Step 2: Verify typecheck**

Run: `cd frontend-next && npx tsc --noEmit 2>&1 | grep -i 'line-chart\|charts/index'`
Expected: no errors.

- [ ] **Step 3: Commit**

```bash
git add frontend-next/src/components/charts/line-chart.tsx
git commit -m "feat(frontend): add reusable LineChart component with recharts"
```

---

### Task 5: Add daily-stats API client + type

**Files:**
- Modify: `frontend-next/src/types/index.ts`
- Modify: `frontend-next/src/lib/api-client.ts`

**Interfaces:**
- Produces: `DailyStat` type, `interactionApi.dailyStats(params)` method

- [ ] **Step 1: Add the type**

In `frontend-next/src/types/index.ts`, add near `InteractionStats`:

```ts
export interface DailyStat {
  date: string
  count: number
}
```

- [ ] **Step 2: Add the API method**

In `frontend-next/src/lib/api-client.ts`, extend `interactionApi` (add after `stats` at line 127):

```ts
  dailyStats: (params?: { case_id?: string; payload_id?: string; days?: number }) =>
    api.get<DailyStat[]>('/interactions/stats/daily', params),
```

Ensure `DailyStat` is imported from `../types` (check the existing import block at the top of `api-client.ts`).

- [ ] **Step 3: Verify typecheck + build**

Run: `cd frontend-next && npx tsc --noEmit`
Expected: no errors.

- [ ] **Step 4: Commit**

```bash
git add frontend-next/src/types/index.ts frontend-next/src/lib/api-client.ts
git commit -m "feat(frontend): add interactionApi.dailyStats and DailyStat type"
```

---

### Task 6: Integrate charts into dashboard

**Files:**
- Modify: `frontend-next/src/app/dashboard/page.tsx`
- Modify: `frontend-next/src/lib/i18n-context.tsx`

**Interfaces:**
- Consumes: `DonutChart`, `LineChart` from Task 3/4, `interactionApi.dailyStats` from Task 5, existing `useInteractionStats`

- [ ] **Step 1: Add i18n keys**

In `frontend-next/src/lib/i18n-context.tsx`, in the `en-US` translations object (and the matching `zh-CN` object), add to the `dashboard.*` namespace:

```ts
// en-US
dashboard.trend_7d: '7-Day Interaction Trend'
dashboard.protocol_distribution: 'Protocol Distribution'  // likely exists; verify
// zh-CN
dashboard.trend_7d: '近 7 天交互趋势'
dashboard.protocol_distribution: '协议分布'
```

Check whether `dashboard.protocol_distribution` already exists (it may) - if so, do not duplicate. Add `dashboard.trend_7d` to both language blocks.

- [ ] **Step 2: Add a useQuery hook for daily stats in the dashboard**

In `frontend-next/src/app/dashboard/page.tsx`, add imports at the top:

```tsx
import { DonutChart, LineChart } from '@/components/charts'
import { useQuery } from '@tanstack/react-query'
import { interactionApi } from '@/lib/api-client'
```

(Verify `useQuery` and `interactionApi` are not already imported; if they are, skip.)

Inside `DashboardPage()` after the existing `useInteractionStats()` call (around line 116), add:

```tsx
const { data: dailyResp, isLoading: dailyLoading } = useQuery({
  queryKey: ['interactions', 'daily'],
  queryFn: () => interactionApi.dailyStats({ days: 7 }),
})
const dailyStats = dailyResp?.data ?? []
```

- [ ] **Step 3: Replace ProtocolBar with DonutChart**

Find the Protocol Distribution card (around lines 183-201) and replace the `<ProtocolBar .../>` usage with a `<DonutChart>`:

```tsx
<Card className="dark:bg-gray-800 dark:border-gray-700">
  <CardContent className="p-5">
    <h3 className="text-sm font-medium text-gray-700 dark:text-gray-300 mb-4">
      {t('dashboard.protocol_distribution')}
    </h3>
    <div className="flex items-center gap-6">
      <DonutChart
        size={180}
        data={[
          { name: 'DNS', value: stats.dns_count ?? 0, color: '#8b5cf6' },
          { name: 'HTTP', value: stats.http_count ?? 0, color: '#3b82f6' },
          { name: 'SMTP', value: stats.smtp_count ?? 0, color: '#10b981' },
          { name: 'Other', value: otherCount, color: '#6b7280' },
        ]}
      />
      <div className="space-y-2 text-sm">
        <LegendItem color="#8b5cf6" label="DNS" value={stats.dns_count ?? 0} />
        <LegendItem color="#3b82f6" label="HTTP" value={stats.http_count ?? 0} />
        <LegendItem color="#10b981" label="SMTP" value={stats.smtp_count ?? 0} />
        <LegendItem color="#6b7280" label="Other" value={otherCount} />
      </div>
    </div>
  </CardContent>
</Card>
```

Add `LegendItem` helper near `StatCard` (top of file):

```tsx
function LegendItem({ color, label, value }: { color: string; label: string; value: number }) {
  return (
    <div className="flex items-center gap-2">
      <span className="w-3 h-3 rounded-sm" style={{ backgroundColor: color }} />
      <span className="text-gray-600 dark:text-gray-400">{label}</span>
      <span className="ml-auto font-medium text-gray-900 dark:text-gray-100">{value}</span>
    </div>
  )
}
```

Compute `otherCount` near where `stats` is derived (around line 116-122):

```tsx
const stats = statsResp?.data ?? { total: 0, today: 0, dns_count: 0, http_count: 0, smtp_count: 0 }
const otherCount = Math.max(0, (stats.total ?? 0) - (stats.dns_count ?? 0) - (stats.http_count ?? 0) - (stats.smtp_count ?? 0))
```

Delete the now-unused `ProtocolBar` function (lines 64-99) since it's replaced.

- [ ] **Step 4: Add the LineChart trend card**

Insert a new card after the charts row (after line 229, before the live hit stream). Add inside the existing charts grid or as a new full-width card:

```tsx
<Card className="dark:bg-gray-800 dark:border-gray-700">
  <CardContent className="p-5">
    <h3 className="text-sm font-medium text-gray-700 dark:text-gray-300 mb-4">
      {t('dashboard.trend_7d')}
    </h3>
    <LineChart data={dailyStats} height={260} />
  </CardContent>
</Card>
```

- [ ] **Step 5: Build the frontend**

Run: `cd frontend-next && npm run build 2>&1 | tail -30`
Expected: build succeeds.

- [ ] **Step 6: Browser verification**

Start the backend and frontend dev servers:
```bash
cd frontend-next && npm run dev &
# in another shell, run the backend with test creds
./godnslog serve -domain localhost -4 127.0.0.1 -test
```

Open `http://localhost:3000/dashboard`, log in, and verify:
1. Donut chart renders with protocol segments (DNS/HTTP/SMTP/Other) and a legend with counts.
2. Line chart renders the 7-day trend (may be flat/empty if no interactions - generate one via a DNS query to your token domain to confirm it populates).
3. No console errors in the browser devtools.
4. Theme toggle still works (charts visible in both light and dark).

If you cannot run the browser, say so explicitly - do not claim success from build alone.

- [ ] **Step 7: Commit**

```bash
git add frontend-next/src/app/dashboard/page.tsx frontend-next/src/lib/i18n-context.tsx
git commit -m "feat(dashboard): integrate DonutChart for protocol distribution and LineChart for 7-day trend"
```

---

## Self-Review

**Spec coverage:**
- ✅ DonutChart 环形图组件 -> Task 3
- ✅ LineChart 折线图组件 -> Task 4
- ✅ 后端 daily stats 端点 -> Task 1
- ✅ Dashboard 集成（替换 CSS 协议分布条 + 新增趋势图） -> Task 6
- ✅ 协议颜色一致性 -> DonutChart uses the spec's color map (dns=#8b5cf6 matches existing ProtocolBar purple, http=#3b82f6 blue, smtp=#10b981 emerald, other=#6b7280 gray)
- ✅ i18n keys -> Task 6 Step 1

**Placeholder scan:** None - all code complete. LegendItem, otherCount, dailyStats all defined.

**Type consistency:** `DailyStat {date, count}`, `interactionApi.dailyStats` returns `DailyStat[]`, `LineChartDataPoint {date, count}` matches, `DonutChartDataPoint {name, value, color?}` matches dashboard usage.
