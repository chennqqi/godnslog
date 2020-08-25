<template>
  <page-header-wrapper>
      <mavon-editor
          defaultOpen="preview"
          :toolbarsFlag="false"
          :subfield="false"
          :preview="true"
          :editable="false"
          :boxShadow="false"
          :navigation="false"
          v-model="value"
      ></mavon-editor>
  </page-header-wrapper>
</template>

<script>
import md from '!!raw-loader!../../md/introduce.md'
import mdcn from '!!raw-loader!../../md/introduce_cn.md'

export default {
  components: {
  },
  data () {
    return {
    }
  },
  computed: {
    value: {
      get () {
        if (this.$i18n.locale === 'zh-CN') {
          return mdcn
        } else {
            return md
        }
      },
      set (val) {
      }
    }
  }
}
</script>
<style scoped>
  .mavonEditor {
    width: 100%;
    height: 100%;
  }
</style>
