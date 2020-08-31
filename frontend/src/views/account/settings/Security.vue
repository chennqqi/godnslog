<template>
  <div>
    <a-list itemLayout="horizontal" :dataSource="infos">
      <a-list-item slot="renderItem" slot-scope="item, index" :key="index">
        <a-list-item-meta>
          <a slot="title">{{ item.title }}</a>
          <span slot="description">
            <span class="security-list-value" :ref="item.ref">{{ item.value }}</span>
          </span>
        </a-list-item-meta>
        <template v-if="item.actions">
          <a slot="actions" @click="item.actions.callback(item.value)">{{ item.actions.title }}</a>
        </template>
      </a-list-item>
    </a-list>
    <hr>
      <div>
          <span>{{ $t('Change Password') }}</span>
          <p></p>
          <a-input-password placeholder="input password" v-model="value" />
          <p></p>
          <a-button type="primary" @click="changePassword">
            {{ $t('Modify') }}
          </a-button>
      </div>
  </div>
</template>

<script>
  import { getSettingSecurity, setSettingSecurity } from '@/api/manage'

  export default {
    data () {
      return {
        value: '123456',
        infos: [
          {
            title: this.$t('DNS Addr'),
            value: '-',
            key: 'dns_addr',
            actions: {
              title: this.$t('Copy'),
              callback: (val) => {
                this.copy(val)
              }
            }
          },
          {
            title: this.$t('HTTP Addr'),
            value: '-',
            key: 'http_addr',
            actions: {
              title: this.$t('Copy'),
              callback: (val) => {
                this.copy(val)
              }
            }
          },
          {
            title: this.$t('Secret'),
            value: '138****8293',
            key: 'token',
            actions: {
              title: this.$t('Copy'),
              callback: (val) => {
                this.copy(val)
              }
            }
          }
        ]
      }
    },
    created () {
      this.$nextTick(() => getSettingSecurity().then((res) => {
          console.log('getSettingSecurity:', res)
          const data = res.result
          this.infos.forEach((item, index) => {
            this.infos[index].value = data[item.key]
          })
          this.value = '654321'
      }))
    },
    methods: {
      changePassword () {
        // TODO: 验证密码强度
        setSettingSecurity({
          password: this.value
        }).then(res => {
          console.log('changePassword res:', res.code, res.message)
          if (res.code === 0) {
            this.$message.info(res.message)
            return this.$store.dispatch('Logout').then(() => {
              this.$router.push({ name: 'login' })
            })
          }
          this.$message.warn(res.message)
        })
        this.value = '123456'
      },
      copy (val) {
        const input = document.createElement('input')
        document.body.appendChild(input)
        input.setAttribute('value', val)
        input.select()
        if (document.execCommand('copy')) {
          document.execCommand('copy')
        }
        document.body.removeChild(input)
        this.$message.success('success copyted ' + val)
      }
    }
  }
</script>

<style scoped>
</style>
