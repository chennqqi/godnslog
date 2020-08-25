<template>
  <page-header-wrapper>
    <a-timeline mode="alternate">
      <a-timeline-item>Create a services site 2015-09-01</a-timeline-item>
      <a-timeline-item color="green">
        Solve initial network problems 2015-09-01
      </a-timeline-item>
      <a-timeline-item>
        <a-icon slot="dot" type="clock-circle-o" style="font-size: 16px;" />
        Sed ut perspiciatis unde omnis iste natus error sit voluptatem accusantium doloremque
        laudantium, totam rem aperiam, eaque ipsa quae ab illo inventore veritatis et quasi architecto
        beatae vitae dicta sunt explicabo.
      </a-timeline-item>
      <a-timeline-item color="red">
        Network problems being solved 2015-09-01
      </a-timeline-item>
      <a-timeline-item>Create a services site 2015-09-01</a-timeline-item>
      <a-timeline-item>
        <a-icon slot="dot" type="clock-circle-o" style="font-size: 16px;" />
        Technical testing 2015-09-01
      </a-timeline-item>
    </a-timeline>
  </page-header-wrapper>
</template>

<script>
  import { Timeline } from 'ant-design-vue'
  export default {
    components: {
      Timeline
    },
    data () {
      return {}
    }
  }
</script>

<style scoped>
  .ant-timeline-item-head{
  top: 4px !important;
  }
  .ant-timeline-item-content{
  top: 0px !important;
  }
</style>
