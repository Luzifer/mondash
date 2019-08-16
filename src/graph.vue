<template>
  <div
    ref="container"
    class="ct-double-octave"
  />
</template>

<script>
import chartist from 'chartist'
import moment from 'moment'

export default {
  name: 'Graph',
  props: {
    metric: {
      required: true,
      type: Object,
    },
  },

  data() {
    return {
      chart: null,
    }
  },

  computed: {
    data() {
      const vh = this.metric.value_history

      const labels = []
      const series = []

      for (const k of Object.keys(vh)) {
        labels.push(moment(k * 1000).format('lll'))
        series.push(vh[k])
      }

      return {
        labels,
        series: [series],
      }
    },
  },

  watch: {
    metric() {
      this.chart.update(this.data)
    },
  },

  mounted() {
    this.chart = chartist.Line(this.$refs.container, this.data)
  },
}
</script>
