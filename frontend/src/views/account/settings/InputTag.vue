<template>
  <div class="page-header-index-wide">
    <div>
      <a-tag v-for="item in items" :key="item" closable @close="remove(item)">{{ item }}</a-tag>
    </div>
    <a-input @pressEnter="add(input)" v-model="input"></a-input>
  </div>
</template>

<script>
  import { isIPv4 } from '../../../utils/util'
  export default {
    props: {
      values: {
        type: Array,
        required: true,
        default: () => { return [] }
      }
    },
    data () {
      return {
        input: '',
        items: this.values
      }
    },
    watch: {
      values (val) {
        this.items = val
      }
    },
    methods: {
      add (item) {
        console.log(this.items, 'on add', item)
        if (!isIPv4(item)) {
          this.$message.info('无效IP地址' + item)
          return
        }
        if (this.items.some((iter) => iter === item)) {
          this.$message.info('已存在: ' + item)
        } else {
          this.items.push(item)
        }
        this.input = ''
        console.log(this.items, 'after add', this.items)
        this.$emit('update:values', this.items)
      },
      remove (item) {
        console.log(this.items, 'on remove', item)
        this.items = this.items.filter((iter) => iter !== item)
        console.log(this.items, 'after remove', this.items)
        this.$emit('update:values', this.items)
      }
    }
  }
</script>

<style scoped>
  .append {
    padding-left: 5px;
  }
</style>
