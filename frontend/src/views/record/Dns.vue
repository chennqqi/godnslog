<template>
  <page-header-wrapper>
    <a-card :bordered="false">
      <div class="table-page-search-wrapper">
        <a-form layout="inline">
          <a-row :gutter="48">
            <a-col :md="8" :sm="24">
              <a-form-item :label="$t('Domain')">
                <a-input v-model="queryParam.domain" placeholder="" />
              </a-form-item>
            </a-col>
            <a-col :md="8" :sm="24">
              <a-form-item label="IP">
                <a-input v-model="queryParam.ip" placeholder="" />
              </a-form-item>
            </a-col>
            <template v-if="advanced">
              <a-col :md="8" :sm="24">
                <a-form-item :label="$t('UpdateTime')">
                  <a-date-picker v-model="queryParam.date" show-time style="width: 100%" :placeholder="$t('date')" />
                </a-form-item>
              </a-col>
            </template>
            <a-col :md="!advanced && 8 || 24" :sm="24">
              <span class="table-page-search-submitButtons" :style="advanced && { float: 'right', overflow: 'hidden' } || {} ">
                <a-button type="primary" @click="$refs.table.refresh(true)">{{ $t('Query') }}</a-button>
                <a-button style="margin-left: 8px" @click="() => this.queryParam = {}">{{ $t('Reset') }}</a-button>
                <a @click="toggleAdvanced" style="margin-left: 8px">
                  {{ advanced ? $t('Collapse') : $t('Expand') }}
                  <a-icon :type="advanced ? 'up' : 'down'" />
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
        :rowSelection="options.rowSelection"
        showPagination="auto">
        <span slot="serial" slot-scope="text, record, index">
          {{ index + 1 }}
        </span>
        <span slot="domain" slot-scope="text">
          <ellipsis :length="128" tooltip>{{ text }}</ellipsis>
        </span>
        <span slot="action" slot-scope="text, record">
          <template>
            <a @click="handleDelete(record)">{{ $t('delete') }}</a>
          </template>
        </span>
      </s-table>
    </a-card>
  </page-header-wrapper>
</template>

<script>
  import moment from 'moment'
  import { STable, Ellipsis } from '@/components'
  import { getDnsList, deleteDnsList } from '@/api/record'
    export default {
      name: 'TableList',
      components: {
        STable,
        Ellipsis
      },
      data () {
        return {
          // create model
          visible: false,
          confirmLoading: false,
          mdl: {},
          // 高级搜索 展开/关闭
          advanced: false,
          // 查询参数
          queryParam: {},
          // 加载数据方法 必须为 Promise 对象
          loadData: parameter => {
            const requestParameters = Object.assign({}, parameter, this.queryParam)
            console.log('loadData request parameters:', requestParameters)
            return getDnsList(requestParameters)
              .then(res => {
                return res.result
              })
          },
          selectedRowKeys: [],
          selectedRows: [],
          // custom table alert & rowSelection
          options: {
            alert: {
              show: true,
              clear: () => {
                this.selectedRowKeys = []
              }
            },
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
      },
      mounted () {
        console.log('mounted dns')
        this.tableOption()
      },
      computed: {
        rowSelection () {
          return {
            selectedRowKeys: this.selectedRowKeys,
            onChange: this.onSelectChange
          }
        },
        columns () {
          return [{
              title: '#',
              scopedSlots: {
                customRender: 'serial'
              }
            },
            {
              title: this.$t('Domain'),
              dataIndex: 'domain',
              scopedSlots: {
                customRender: 'domain'
              },
              sorter: true
            },
            {
              title: 'IP',
              dataIndex: 'addr',
              sorter: true
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
              scopedSlots: {
                customRender: 'action'
              }
            }
          ]
        }
      },
      methods: {
        tableOption () {
          console.log('tableOption run old value', this.optionAlertShow)
          if (!this.optionAlertShow) {
            this.options = {
              alert: {
                show: true,
                clear: () => {
                  this.selectedRowKeys = []
                }
              },
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
          deleteDnsList({
            ids: null
          }).then(res => {
            this.$message.info(res.result.message)
          })
          setTimeout(() => {
            this.$refs.table.refresh() // refresh() 不传参默认值 false 不刷新到分页第一页
          }, 1000)
        },
        handleDeleteSelect () {
          var ids = this.selectedRows.map(n => n.id)
          console.log('handleDeleteSelect', ids)
          deleteDnsList({
            ids: ids
          }).then(res => {
            this.$message.info(res.result.message)
          })
          // TODO: force update
          setTimeout(() => {
            this.$refs.table.refresh() // refresh() 不传参默认值 false 不刷新到分页第一页
          }, 1000)
        },
        handleDelete (record) {
          console.log('handleDelete', record)
          deleteDnsList({
            ids: [record.id]
          }).then(res => {
            this.$message.info(res.message)
          })
          setTimeout(() => {
            this.$refs.table.refresh() // refresh() 不传参默认值 false 不刷新到分页第一页
          }, 1000)
        },
        onSelectChange (selectedRowKeys, selectedRows) {
          console.log('onSelectChange', selectedRowKeys, selectedRows)
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
