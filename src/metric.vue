<template>
  <b-row class="mb-3">
    <b-col>
      <b-card :border-variant="variantFromStatus(metric.status)">
        <b-row slot="header">
          <b-col cols="7">
            <strong>{{ metric.title }}</strong>
            <a
              v-if="metric.detail_url"
              class="ml-1"
              :href="metric.detail_url"
            >
              <i class="fas fa-link fa-xs" />
            </a>
          </b-col>
          <b-col
            class="d-flex align-items-center"
            cols="5"
          >
            <historyBar :metric="metric" />
          </b-col>
        </b-row>

        <b-card-text v-html="description" />

        <graph
          v-if="!metric.config.hide_mad"
          :metric="metric"
        />

        <div
          slot="footer"
          class="d-flex justify-content-between"
        >
          <span>
            <b-badge
              v-if="!metric.config.hide_value"
              :variant="variantFromStatus(metric.status)"
            >Current value: {{ metric.value.toFixed(3) }}</b-badge>
            <span v-if="!metric.config.hide_mad">
              <abbr title="Median Absolute Deviation">MAD</abbr>: {{ metric.mad_multiplier.toFixed(3) }} above the median ({{ metric.median.toFixed(3) }})
            </span>
          </span>
          <small class="text-muted">
            <span
              v-b-tooltip.hover
              :title="moment(metric.last_update).format('lll')"
            >Updated {{ moment(metric.last_update).fromNow() }}</span>
            <span
              v-if="metric.status !== 'OK'"
              v-b-tooltip.hover
              :title="moment(metric.last_ok).format('lll')"
            > (Last OK {{ moment(metric.last_ok).fromNow() }})</span>
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
  name: 'Metric',

  components: { graph, historyBar },
  props: {
    metric: {
      required: true,
      type: Object,
    },
  },

  computed: {
    description() {
      return this.metric.description.replace(new RegExp(/\n/, 'g'), '<br>')
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
