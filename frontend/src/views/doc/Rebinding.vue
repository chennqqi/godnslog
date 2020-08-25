<template>
  <page-header-wrapper>
    <div
      class="post"
    >
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
    </div>
  </page-header-wrapper>
</template>

<script>
import md from '!!raw-loader!../../md/rebinding.md'
import mdcn from '!!raw-loader!../../md/rebinding_cn.md'

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
