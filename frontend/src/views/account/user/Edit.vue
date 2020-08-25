<template>
  <div>
    <h3>{{ createMode ? $t('New User') : $t('Edit User') }}</h3>
    <a-form :form="form" @submit="handleSubmit">
      <a-form-item
        :labelCol="labelCol"
        :wrapperCol="wrapperCol"
        label="UID"
        :disabled="true"
        v-if="!createMode"
      >
      <a-input
        disabled
        v-decorator="[
        'id',
          {
            initialValue: '-',
            rules: [
            ]
          }
        ]"
      >ID: {{}}</a-input>
      </a-form-item>
      <a-form-item
        :labelCol="labelCol"
        :wrapperCol="wrapperCol"
        :label="$t('username')"
        hasFeedback
      >
        <a-input
          placeholder="请输入用户名"
          v-decorator="[
            'username',
            {
              rules: [
                { required: true, message: $t('请输入规则编号') },
                { required: true, min: 4, message: $t('最小长度4') },
                { whitespace: true, message: $t('不能为空') }
              ]
            }
          ]"
        ></a-input>
      </a-form-item>
      <a-form-item
        :labelCol="labelCol"
        :wrapperCol="wrapperCol"
        :label="$t('email')"
        hasFeedback
      >
        <a-input
          placeholder="邮箱"
          v-decorator="[
            'email',
            {
              rules: [
                { required: true },
                { type: 'email', message: $t('请输入正确的邮件地址') }
              ]
            }
          ]"
        ></a-input>
      </a-form-item>
      <a-form-item
        :labelCol="labelCol"
        :wrapperCol="wrapperCol"
        :label="$t('password')"
        hasFeedback>
        <a-input-password
          :min="8"
          style="width: 100%"
          v-decorator="[
            'password',
            {
              rules: [
                { required: true },
                { validator: passwordValidator }
              ]
            }
          ]" />
      </a-form-item>
      <!-- <a-form-item
        :labelCol="labelCol"
        :wrapperCol="wrapperCol"
        :label="$t('role')"
        hasFeedback
      >
        <a-radio-group
         name="role"
         style="width: 100%"
         v-decorator="[
           'role',
           {
              rules: [
                { required: true },
                { type: 'integer' }
              ],
              initialValue: 1
           }
         ]">
          <a-radio :value="1">
            {{ $t('Normal User') }}
          </a-radio>
          <a-radio :value="2">
            {{ $t('Admin') }}
          </a-radio>
        </a-radio-group>
      </a-form-item> -->

      <a-form-item
        v-bind="buttonCol"
      >
        <a-row type="flex" justify="space-between">
          <a-col :span="3">
            <a-button type="primary" @click="handleSubmit">{{ $t('Submit') }}</a-button>
          </a-col>
          <a-col :span="3">
            <a-button @click="handleGoBack" type="">{{ $t('Back') }}</a-button>
          </a-col>
          <a-col :span="17">
          </a-col>
        </a-row>
      </a-form-item>
    </a-form>
  </div>
</template>

<script>
import pick from 'lodash.pick'
import { saveUser } from '@/api/manage'

const passwordValidator = { validator: validatePassword }
function validatePassword (rule, val, callback) {
  if (val) {
    val = val.trim()
    const isMob = /^[\d-]+$/
    const phoneReg = /^1[3|4|5|7|8][0-9]{9}$/
    if (!isMob.test(val) && !phoneReg.test(val)) {
      const message = '电话/手机格式不对'
      callback(message)
    }
  }
  callback()
};

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
        xs: { span: 24 },
        sm: { span: 5 }
      },
      wrapperCol: {
        xs: { span: 24 },
        sm: { span: 12 }
      },
      buttonCol: {
        wrapperCol: {
          xs: { span: 24 },
          sm: { span: 12, offset: 5 }
        }
      },
      form: this.$form.createForm(this),
      id: 0,
      passwordValidator: passwordValidator,
      createMode: false
    }
  },
  // beforeCreate () {
  //   this.form = this.$form.createForm(this)
  // },
  created () {
    this.$nextTick(() => {
      this.loadEditInfo(this.record)
    })
  },
  methods: {
    handleGoBack () {
      this.$emit('onGoBack', false)
    },
    handleSubmit () {
      const { form: { validateFields } } = this
      validateFields((err, values) => {
        console.log('Received values of form: ', values)
        if (err) {
          console.log('Received values of form error: ', err)
          this.$message.warn(err)
          return
        }
        const req = {}
        Object.assign(req, values)

        console.log('saveUser req:', req)
        if (this.createMode) {
          // eslint-disable-next-line no-console
          if (req.id === undefined || req.id === '-') {
            req.id = 0
          }
        }
        saveUser(req).then(res => {
          console.log('saveUser:', res)
        })
        this.$emit('onGoBack', true)
      })
    },
    loadEditInfo (data) {
      const { form } = this
      // ajax
      console.log(`将加载 ${this.id} 信息到表单`)
        if (!data.hasOwnProperty('id')) {
          this.createMode = true
        } else {
          this.createMode = false
          const formData = pick(data, [ 'id', 'username', 'email', 'password' ])
          form.setFieldsValue(formData)
        }
    }
  }
}
</script>
