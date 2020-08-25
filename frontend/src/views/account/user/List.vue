<template>
  <div>
    <div class="table-page-search-wrapper">
      <a-form layout="inline">
        <a-row :gutter="48">
          <a-col :md="8" :sm="24">
            <a-form-item :label="$t('Username')">
              <a-input v-model="queryParam.username" placeholder=""/>
            </a-form-item>
          </a-col>
          <a-col :md="8" :sm="24">
            <a-form-item :label="$t('Email')">
              <a-input v-model="queryParam.email" placeholder=""/>
            </a-form-item>
          </a-col>
          <a-col :md="8" :sm="24">
            <a-form-item :label="$t('UpdateTime')">
              <a-date-picker v-model="queryParam.date" style="width: 100%" placeholder="请输入更新日期"/>
            </a-form-item>
          </a-col>
        </a-row>
      </a-form>
    </div>

    <div class="table-operator">
      <a-button type="primary" icon="plus" @click="handleEdit()">{{ $t('New User') }}</a-button>
      <a-button style="margin-left: 8px" @click="handleDeleteSelect" v-if="selectedRowKeys.length > 0">
          {{ $t('Delete Select') }}
      </a-button>
      <a-button type="dashed" @click="tableOption">{{ optionAlertShow && $t('Close') || $t('Open') }} {{ $t('Batch') }}</a-button>
    </div>

    <s-table
      ref="table"
      size="default"
      rowKey="id"
      :columns="columns"
      :data="loadData"
      :alert="options.alert"
      :rowSelection="options.rowSelection"
    >
      <span slot="serial" slot-scope="text, record, index">
        {{ index + 1 }}
      </span>
      <span slot="action" slot-scope="text, record">
        <template>
          <a @click="handleEdit(record)">{{ $t('Edit') }}</a>
          <a-divider type="vertical" />
          <a @click="handleDelete(record)">{{ $t('Delete') }}</a>
        </template>
      </span>
    </s-table>
  </div>
</template>

<script>
import moment from 'moment'
import { STable } from '@/components'
import { getUserList, delUser } from '@/api/manage'

export default {
  name: 'TableList',
  components: {
    STable
  },
  data () {
    return {
      mdl: {},
      // 高级搜索 展开/关闭
      advanced: false,
      // 查询参数
      queryParam: {},
      // 表头
      columns: [
        {
          title: '#',
          scopedSlots: { customRender: 'serial' },
          dataIndex: 'id'
        },
        {
          title: this.$t('Username'),
          dataIndex: 'username'
        },
        {
          title: this.$t('Email'),
          dataIndex: 'email'
        },
        {
          title: this.$t('UpdateTime'),
          dataIndex: 'utime',
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
        console.log('loadData.parameter', parameter)
        return getUserList(Object.assign(parameter, this.queryParam))
          .then(res => {
            console.log('loadData', res.result)
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
      optionAlertShow: true
    }
  },
  created () {
    this.tableOption()
    // getRoleList({ t: new Date() })
  },
  methods: {
    tableOption () {
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
    handleDelete (record) {
      console.log('handleDelete id:', record.id)
      delUser({
        ids: [record.id]
      }).then(res => {
        this.$message.info(res.message)
      })
      setTimeout(() => {
          this.$refs.table.refresh() // refresh() 不传参默认值 false 不刷新到分页第一页
      }, 1000)
    },
    handleDeleteSelect () {
      delUser({
        ids: this.selectedRowKeys
      }).then(res => {
        this.$message.info(res.message)
      })
      setTimeout(() => {
          this.$refs.table.refresh() // refresh() 不传参默认值 false 不刷新到分页第一页
      }, 1000)
    },
    activated () {
      console.log('LIST activated')
      this.$refs.table.refresh() // refresh() 不传参默认值 false 不刷新到分页第一页
    },
    handleEdit (record) {
      this.$emit('onEdit', record)
    },
    onSelectChange (selectedRowKeys, selectedRows) {
      this.selectedRowKeys = selectedRowKeys
      this.selectedRows = selectedRows
    },
    resetSearchForm () {
      this.queryParam = {
        date: moment(new Date())
      }
    }
  }
}
</script>
