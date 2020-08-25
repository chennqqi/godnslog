<template>
  <div>
    <a-input-number
      id="inputNumber"
      v-model="number"
      @change="handleNumberChange" />
    <a-radio-group
      v-model="unit"
      style="padding-left: 20px"
      :options="options"
      @change="handleUnitChange">
    </a-radio-group>
  </div>
</template>

<script>
export default {
  props: {
    value: {
      type: Object
    }
  },
  data () {
    const { unit, number } = this.value || {}
    console.log('init value:', this.value, unit, number)
    return {
      options: [
        { label: this.$t('setting.system.base.unit.day'), value: 1 },
        { label: this.$t('setting.system.base.unit.hour'), value: 2 }
      ],
      unit: unit || 1,
      number: number || 1
    }
  },
  watch: {
    value (val) {
      // Object.assign(this, val)
      this.unit = val.unit
      this.number = val.number
    }
  },
  methods: {
    handleUnitChange (val) {
      console.log('handleUnitChanged', val.target.value)
      this.$emit('change', { ...this.value, unit: val.taget.value })
    },
    handleNumberChange (val) {
      console.log('handleNumberChange', val)
      this.$emit('change', { ...this.value, number: val })
    }
  }
}
</script>

<style>
</style>
