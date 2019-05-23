<template>
  <b-row class="mb-2">
    <b-col>
      <b-card :border-variant="variantFromStatus(metric.status)">

        <b-row slot="header">
          <b-col cols="7">
            <strong>{{ metric.title }}</strong>
          </b-col>
          <b-col class="d-flex align-items-center" cols="5">
            <historyBar :metric="metric"></historyBar>
          </b-col>
        </b-row>

        <b-card-text>
          {{ metric.description }}
        </b-card-text>

        <graph :metric="metric" v-if="!metric.config.hide_mad"></graph>

        <div slot="footer" class="d-flex justify-content-between">
          <span>
            <b-badge :variant="variantFromStatus(metric.status)" v-if="!metric.config.hide_value">Current value: {{ metric.value.toFixed(3) }}</b-badge>
            <span v-if="!metric.config.hide_mad">
              <abbr title="Median Absolute Deviation">MAD</abbr>: {{ metric.mad_multiplier.toFixed(3) }} above the median ({{ metric.median.toFixed(3) }})
            </span>
          </span>
          <small class="text-muted">
            <span v-b-tooltip.hover :title="moment(metric.last_update).format('lll')">Updated {{ moment(metric.last_update).fromNow() }}</span>
            <span v-if="metric.status !== 'OK'"> (Last OK {{ moment(metric.last_ok).fromNow() }})</span>
          </small>
        </div>

      </b-card>
    </b-col>
  </b-row>
</template>

<script>
import moment from 'moment'

import graph from './graph.vue'
import historyBar from './history-bar.vue'

export default {
  name: 'metric',
  props: ['metric'],

  components: { graph, historyBar },

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
