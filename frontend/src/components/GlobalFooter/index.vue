<template>
  <global-footer class="footer custom-render">
    <template v-slot:links>
      <a href="https://github.com/chennqqi/godnslog" target="_blank">Wechat</a>
      <a href="https://github.com/chennqqi/godnslog" target="_blank">Github</a>
      <a href="https://github.com/chennqqi/" target="_blank">@chennqqi</a>
    </template>
    <template v-slot:copyright>
      <a href="https://www.shellpub.com" target="_blank">SHELLPUB.COM</a>
    </template>
  </global-footer>
</template>

<script>
import { GlobalFooter } from '@ant-design-vue/pro-layout'

export default {
  name: 'ProGlobalFooter',
  components: {
    GlobalFooter
  }
}
</script>
