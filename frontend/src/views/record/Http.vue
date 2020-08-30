<template>
  <page-header-wrapper>
    <div class="table-page-search-wrapper">
      <a-form layout="inline">
        <a-row :gutter="48">
          <a-col :md="8" :sm="24">
            <a-form-item :label="$t('Path')">
              <a-input v-model="queryParam.id" placeholder=""/>
            </a-form-item>
          </a-col>
          <a-col :md="8" :sm="24">
            <a-form-item label="IP">
              <a-input v-model="queryParam.addr" placeholder=""/>
            </a-form-item>
          </a-col>
          <template v-if="advanced">
            <a-col :md="8" :sm="24">
              <a-form-item label="Method">
                <a-input v-model="queryParam.method" style="width: 100%"/>
              </a-form-item>
            </a-col>
            <a-col :md="8" :sm="24">
              <a-form-item label="User-Agent">
                <a-input v-model="queryParam.ua" style="width: 100%"/>
              </a-form-item>
            </a-col>
            <a-col :md="8" :sm="24">
              <a-form-item label="Content-Type">
                <a-input v-model="queryParam.ctype" style="width: 100%"/>
              </a-form-item>
            </a-col>
            <a-col :md="8" :sm="24">
              <a-form-item label="Data">
                <a-input v-model="queryParam.data" style="width: 100%"/>
              </a-form-item>
            </a-col>
            <a-col :md="8" :sm="24">
              <a-form-item :label="$t('UpdateTime')">
                <a-date-picker v-model="queryParam.date" show-time style="width: 100%" :placeholder="$t('date')"/>
              </a-form-item>
            </a-col>
          </template>
          <a-col :md="!advanced && 8 || 24" :sm="24">
            <span class="table-page-search-submitButtons" :style="advanced && { float: 'right', overflow: 'hidden' } || {} ">
              <a-button type="primary" @click="$refs.table.refresh(true)">{{ $t('Query') }}</a-button>
              <a-button style="margin-left: 8px" @click="() => queryParam = {}">{{ $t('Reset') }}</a-button>
              <a @click="toggleAdvanced" style="margin-left: 8px">
                {{ advanced ? $t('Collapse') : $t('Expand') }}
                <a-icon :type="advanced ? 'up' : 'down'"/>
              </a>
            </span>
          </a-col>
        </a-row>
      </a-form>
    </div>

    <div class="table-operator">
      <a-button type="primary" @click="handleDeleteAll">{{ $t('Delete All') }}</a-button>
      <a-button type="primary" @click="handleDeleteSelect" :disabled="selectedRowKeys.length === 0">{{ $t('Delete Select') }}</a-button>
      <a-button type="dashed" @click="tableOption">{{ optionAlertShow && $t('Close') || $t('Open') }} {{ $t('Batch') }}</a-button>
    </div>

    <s-table
      ref="table"
      size="default"
      rowKey="key"
      :columns="columns"
      :data="loadData"
      :alert="options.alert"
      :pagination="pagination"
      :rowSelection="options.rowSelection"
    >
      <span slot="serial" slot-scope="text, record, index">
        {{ index + 1 }}
      </span>
      <span slot="domain" slot-scope="text">
        <ellipsis :length="128" tooltip>{{ text }}</ellipsis>
      </span>
      <span slot="action" slot-scope="text, record">
        <template>
          <a @click="handleDelete(record)">删除</a>
        </template>
      </span>
    </s-table>
  </page-header-wrapper>
</template>

<script>
import moment from 'moment'
import { STable, Ellipsis } from '@/components'
import { getHttpList, deleteHttpList } from '@/api/record'

export default {
  name: 'TableList',
  components: {
    STable,
    Ellipsis
  },
  data () {
    return {
      mdl: {},

      loading: false,
      pagination: {},
      // 分页

      // 高级搜索 展开/关闭
      advanced: false,
      // 查询参数
      queryParam: {},
      // 表头

      columns: [
        {
          title: '#',
          scopedSlots: { customRender: 'serial' }
        },
        {
          title: this.$t('Path'),
          dataIndex: 'path',
          scopedSlots: { customRender: 'path' }
        },
        {
          title: 'IP',
          dataIndex: 'addr',
          sorter: true
        },
        {
          title: 'Method',
          dataIndex: 'method'
        },
        {
          title: 'Data',
          dataIndex: 'data',
          scopedSlots: { customRender: 'data' }
        },
        {
          title: 'User-Agent',
          dataIndex: 'ua'
        },
        {
          title: 'Content-Type',
          dataIndex: 'ctype'
        },
        {
          title: this.$t('UpdateTime'),
          dataIndex: 'ctime',
          sorter: true
        },
        {
          title: this.$t('Action'),
          dataIndex: 'action',
          width: '150px',
          scopedSlots: { customRender: 'action' }
        }
      ],
      // 加载数据方法 必须为 Promise 对象
      loadData: parameter => {
        console.log('loadData.parameter:', parameter)
        return getHttpList(Object.assign(parameter, this.queryParam))
          .then(res => {
            console.log('res:', res)
            console.log('loadData:', res.results)
            return res.result
          })
      },
      selectedRowKeys: [],
      selectedRows: [],

      // custom table alert & rowSelection
      options: {
        alert: { show: true, clear: () => { this.selectedRowKeys = [] } },
        rowSelection: {
          selectedRowKeys: this.selectedRowKeys,
          onChange: this.onSelectChange
        }
      },
      optionAlertShow: false
    }
  },
  created () {
    this.tableOption()
    // getRoleList({ t: new Date() })
  },
  mounted () {
    console.log('mounted dns')
    this.tableOption()
  },
  methods: {
    handleTableChange (pagination, filters, sorter) {
          const pager = { ...this.pagination }
          pager.current = pagination.current
          this.pagination = pager
          this.fetch({
            results: pagination.pageSize,
            page: pagination.current,
            sortField: sorter.field,
            sortOrder: sorter.order,
            ...filters
       })
    },
    fetch (params = {}) {
        console.log('params:', params)
        this.loading = true
        var parameter = {}
        getHttpList(Object.assign(parameter, params)).then(res => {
          const result = res.result
          const pagination = { ...this.pagination }
          //   // Read total count from server
          pagination.total = result.totalCount
          this.loading = false
          // this.loadData = result.data
          this.pagination = pagination
          return res.result
        })
    },
    tableOption () {
      console.log('tableOption run old value', this.optionAlertShow)
      if (!this.optionAlertShow) {
        this.options = {
          alert: { show: true, clear: () => { this.selectedRowKeys = [] } },
          rowSelection: {
            selectedRowKeys: this.selectedRowKeys,
            onChange: this.onSelectChange
          }
        }
        this.optionAlertShow = true
      } else {
        this.options = {
          alert: false,
          rowSelection: null
        }
        this.optionAlertShow = false
      }
    },
    handleDeleteAll () {
      console.log('handleDeleteAll')
      deleteHttpList({
        ids: null
      }).then(res => {
        this.$message.info(res.message)
      })
      setTimeout(() => {
        this.$refs.table.refresh() // refresh() 不传参默认值 false 不刷新到分页第一页
      }, 1000)
    },
    handleDeleteSelect () {
      var ids = this.selectedRows.map(n => n.id)
      console.log('handleDeleteSelect', ids)
      deleteHttpList({
        ids: ids
      }).then(res => {
        this.$message.info(res.message)
      })
      setTimeout(() => {
        this.$refs.table.refresh() // refresh() 不传参默认值 false 不刷新到分页第一页
      }, 1000)
    },
    handleDelete (record) {
      console.log('handleDelete', record)
      deleteHttpList({
        ids: [record.id]
      }).then(res => {
         this.$message.info(res.message)
      })
      setTimeout(() => {
        this.$refs.table.refresh() // refresh() 不传参默认值 false 不刷新到分页第一页
      }, 1000)
    },
    onSelectChange (selectedRowKeys, selectedRows) {
      console.log('onSelectChanged rowKeys:', selectedRowKeys.length)
      this.selectedRowKeys = selectedRowKeys
      this.selectedRows = selectedRows
    },
    toggleAdvanced () {
      this.advanced = !this.advanced
    },
    resetSearchForm () {
      this.queryParam = {
        date: moment(new Date())
      }
    }
  }
}
</script>
