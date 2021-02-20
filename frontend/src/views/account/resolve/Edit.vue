<template>
  <div>
    <a-form :form="form" @submit="handleSubmit">
      <!--参考https://www.cnblogs.com/cirry/p/12483131.html，给form传递数据能少传，不能多传，否则会报一个警告，这里添加一个隐藏的ID-->
      <a-form-item
        :labelCol="labelCol"
        :wrapperCol="wrapperCol"
        label="编号"
        hasFeedback
        validateStatus="success"
        v-if="!newrecord">
        <a-input
          placeholder="编号"
          v-decorator="[
            'id',
          ]"
          :disabled="true"></a-input>
      </a-form-item>

      <a-form-item
        :labelCol="labelCol"
        :wrapperCol="wrapperCol"
        :label="$t('Host')"
        hasFeedback>
        <a-input
          :disabled="!newrecord"
          placeholder=""
          v-decorator="[
            'host', {
              rules: [
                { required: true, message: '请输入主机记录' }
              ]
            }
          ]"></a-input>
      </a-form-item>

      <a-form-item
        :labelCol="labelCol"
        :wrapperCol="wrapperCol"
        :label="$t('Type')"
        hasFeedback>
        <a-radio-group
          :disabled="!newrecord"
          style="width: 100%;"
          v-decorator="[
            'type', {
              rules: [
                { required: true, message: '请选择解析类型' }
              ],
              initialValue: 'A'
            }]">
          <a-radio value="A">
            A
          </a-radio>
          <a-radio value="CNAME">
            CNAME
          </a-radio>
          <a-radio value="TXT">
            TXT
          </a-radio>
          <a-radio value="MX">
            MX
          </a-radio>
        </a-radio-group>
      </a-form-item>

      <a-form-item
        :labelCol="labelCol"
        :wrapperCol="wrapperCol"
        :label="$t('Value')"
        hasFeedback>
        <a-input
          style="width: 100%"
          v-decorator="['value',
                        {rules: [
                          { required: true },
                          { type: 'string' },
                          { min: 1, max: 255, message: '记录值长度不合法' },
                          { validator: validateValue }
                        ]}]" />
      </a-form-item>

      <a-form-item :labelCol="labelCol" :wrapperCol="wrapperCol" label="TTL" hasFeedback>
        <a-input
          :min="2"
          style="width: 100%"
          v-decorator="['ttl', {rules: [
                                  {
                                    required: true,
                                    type: 'integer',
                                    transform: (value) => { return Number(value) }
                                  },
                                ],
                                initialValue: 600,
          }]" />
      </a-form-item>

      <a-form-item v-bind="buttonCol">
        <a-row>
          <a-col span="6">
            <a-button type="primary" @click="handleSubmit">{{ $t('Submit') }}</a-button>
          </a-col>
          <a-col span="10">
            <a-button @click="handleGoBack">{{ $t('Back') }}</a-button>
          </a-col>
          <a-col span="8"></a-col>
        </a-row>
      </a-form-item>
    </a-form>
  </div>
</template>

<script>
  import pick from 'lodash.pick'
  import {
    setResolve
  } from '@/api/manage'
  import {
    isIPv4,
    isDomain
  } from '@/utils/util'

  export default {
    name: 'TableEdit',
    props: {
      record: {
        type: [Object, String],
        default: ''
      }
    },
    data () {
      return {
        labelCol: {
          xs: {
            span: 24
          },
          sm: {
            span: 5
          }
        },
        wrapperCol: {
          xs: {
            span: 24
          },
          sm: {
            span: 12
          }
        },
        buttonCol: {
          wrapperCol: {
            xs: {
              span: 24
            },
            sm: {
              span: 12,
              offset: 5
            }
          }
        },
        form: this.$form.createForm(this),
        id: 0,
        newrecord: false
      }
    },
    // beforeCreate () {
    //   this.form = this.$form.createForm(this)
    // },
    mounted () {
      this.$nextTick(() => {
        this.loadEditInfo(this.record)
      })
    },
    methods: {
      validateValue (rule, value, callback) {
        if (!value) {
          callback()
          return
        }
        switch (this.form.getFieldValue('type')) {
          case 'A':
            if (!isIPv4(value)) {
              callback(this.$t('A记录必须指向IPv4地址'))
              return
            }
            break
          case 'MX':
            if (!isDomain(value) && !isIPv4(value)) {
              callback(this.$t('MX记录必须指向合法的域名或者IP'))
              return
            }
            break
          case 'TXT':
            break
          case 'CNAME':
            if (!isDomain(value)) {
              callback(this.$t('CNAME记录必须指向合法的域名'))
              return
            }
        }
        callback()
      },
      handleGoBack () {
        this.$emit('onGoBack')
      },
      handleSubmit (e) {
        e.preventDefault()
        const {
          form: {
            validateFields
          }
        } = this
        const { $message } = this
        validateFields((err, values) => {
          if (!err) {
            values.ttl = Number(values.ttl)
            // eslint-disable-next-line no-console
            console.log('Received val0ues of form: ', values, err)
            setResolve(values).then(res => {
              // $message.info(`${res.message}`)
              this.$emit('onGoBack')
            }).catch(err => {
              $message.error(`${err.response.data.message}`)
            })
          } else {
            // TODO: messaage
            console.log('else', err)
          }
        })
      },
      loadEditInfo (data) {
        const {
          form
        } = this
        // ajax
        this.newrecord = data.id === undefined
        console.log(`将加载 ${this.id} 信息到表单`)
        new Promise((resolve) => {
          setTimeout(resolve, 500)
        }).then(() => {
          const formData = pick(data, ['id', 'host', 'type', 'value', 'ttl'])
          console.log('formData', formData)
          console.log(this.newrecord, form.id)
          form.setFieldsValue(formData)
        })
      }
    }
  }
</script>
