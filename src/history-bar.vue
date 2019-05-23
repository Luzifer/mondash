<template>
  <b-progress class="w-100" :max="10000">
    <b-progress-bar
      :key="seg.start"
      :value="seg.percentage * 10000"
      :variant="variantFromStatus(seg.status)"
      :title="`${moment(seg.start).format('lll')}\n${moment(seg.end).format('lll')}`"
      v-b-tooltip.hover
      v-for="seg in metric.history_bar"
      >
    </b-progress-bar>
  </b-progress>
</template>

<script>
import moment from 'moment'

export default {
  name: 'history-bar',
  props: ['metric'],

  methods: {
    moment,

    variantFromStatus(status) {
      return {
        "OK": "success",
        "Warning": "warning",
        "Critical": "danger",
        "Unknown": "info",
      }[status]
    },
  },
}
</script>
