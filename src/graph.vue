<template>
  <div class="ct-double-octave" ref="container"></div>
</template>

<script>
import chartist from 'chartist'
import moment from 'moment'

export default {
  name: 'graph',
  props: ['metric'],

  computed: {
    data() {
      const vh = this.metric.value_history

      let labels = []
      let series = []

      for (const k of Object.keys(vh)) {
        labels.push(moment(k*1000).format('lll'))
        series.push(vh[k])
      }

      return {
        labels,
        series: [series],
      }
    },
  },

  data() {
    return {
      chart: null,
    }
  },

  methods: {
  },

  mounted() {
    this.chart = chartist.Line(this.$refs.container, this.data)
  },

  watch: {
    metric() {
      this.chart.update(this.data)
    },
  },
}
</script>
