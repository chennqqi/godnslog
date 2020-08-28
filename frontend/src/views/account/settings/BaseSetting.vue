<template>
  <div class="account-settings-info-view">
        <a-form layout="vertical" :form="form">
          <a-form-item :label="$t('setting.system.base.callback')" >
            <a-textarea
              rows="4"
              v-decorator="[
                'callback',
                { rules: [
                    { validator: (rule, value, callback) => {
                        if (typeof value !== 'string') {
                          callback($t('should be a valid url'))
                        }
                        if (!isURL(value)) {
                          callback($t('should be a valid url'))
                        }
                        callback()
                      }
                    }
                  ]
                }
              ]"
            />
          </a-form-item>
          <a-form-item
            :label="$t('setting.system.base.cleanInterval(hour)')"
            required>
            <a-input-number
              v-decorator="[
                'cleanHour',
                {
                  initialValue: cleanHour,
                  rules: [
                    { validator: (rule, value, callback) => {
                        if (typeof value !== 'number') {
                          callback($t('should be a valid number'))
                        }
                        if (value < 1) {
                           callback($t('can\'t less than 1 hour'))
                        }
                        if (value > 48) {
                           callback($t('can\'t less than 48 hours'))
                        }
                        callback()
                      }
                    }
                  ]
                }
              ]"
            />
          </a-form-item>
          <a-form-item
            :label="$t('setting.system.base.rebind')"
          >
            <InputTag
              :values.sync="rebind"
            ></InputTag>
          </a-form-item>
          <a-form-item>
            <a-button type="primary" @click="handleSubmit">{{ $t('Submit') }}</a-button>
          </a-form-item>
        </a-form>
  </div>
</template>

<script>
  import pick from 'lodash.pick'
  import { getSettingApp, setSettingApp } from '@/api/manage'
  import { isURL } from '@/utils/util'
  import InputTag from './InputTag'
  import InputUnit from './InputUnit'
  export default {
    components: {
      InputTag,
      InputUnit
    },
    mounted () {
      this.$nextTick(this.loadData())
    },
    data () {
      return {
        callback: '',
        cleanHour: 0,
        rebind: [],
        isURL: isURL,
        form: this.$form.createForm(this)
      }
    },
    watch: {
      rebind: (newval, oldval) => {
        console.log('bindItems.old', oldval, 'bindItem.new', newval)
      }
    },
    methods: {
      loadData () {
        getSettingApp()
          .then(res => {
              const data = res.result
              const formData = pick(data, [ 'callback', 'cleanHour' ])
              this.rebind = data.rebind || []
              console.log('rebind:', this.rebind)
              this.form.setFieldsValue(formData)
          })
      },
      handleSubmit () {
        this.form.validateFields((err, values) => {
          if (!err) {
            const data = { ...values, rebind: this.rebind }
            setSettingApp(data).then(res => {
              this.$message.info(res.message)
            })
          } else {
            console.log('verify error:', err)
            this.$message.warn(err)
          }
        })
        this.loadData()
      }
    }
  }
</script>

<style lang="less" scoped>
  .ant-upload-preview {
    position: relative;
    margin: 0 auto;
    width: 100%;
    max-width: 180px;
    border-radius: 50%;
    box-shadow: 0 0 4px #ccc;

    .mask {
      opacity: 0;
      position: absolute;
      background: rgba(0, 0, 0, 0.4);
      cursor: pointer;
      transition: opacity 0.4s;

      &:hover {
        opacity: 1;
      }

      i {
        font-size: 2rem;
        position: absolute;
        top: 50%;
        left: 50%;
        margin-left: -1rem;
        margin-top: -1rem;
        color: #d6d6d6;
      }
    }
  }
</style>
