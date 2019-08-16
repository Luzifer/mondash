<template>
  <b-progress
    class="w-100"
    :max="10000"
  >
    <b-progress-bar
      v-for="seg in metric.history_bar"
      :key="seg.start"
      v-b-tooltip.hover
      :value="seg.percentage * 10000"
      :variant="variantFromStatus(seg.status)"
      :title="`${moment(seg.start).format('lll')}\n${moment(seg.end).format('lll')}`"
    />
  </b-progress>
</template>

<script>
import moment from 'moment'

export default {
  name: 'HistoryBar',
  props: {
    metric: {
      required: true,
      type: Object,
    },
  },

  methods: {
    moment,

    variantFromStatus(status) {
      return {
        OK: 'success',
        Warning: 'warning',
        Critical: 'danger',
        Unknown: 'info',
      }[status]
    },
  },
}
</script>
