<template>
  <div id="app">
    <b-navbar toggleable="lg" type="light" variant="light">
      <b-navbar-brand href="welcome">MonDash</b-navbar-brand>

      <b-navbar-toggle target="nav-collapse"></b-navbar-toggle>

      <b-collapse id="nav-collapse" is-nav>
        <b-navbar-nav>
          <b-nav-item href="http://docs.mondash.apiary.io/" target="_blank">API Docs</b-nav-item>
        </b-navbar-nav>

        <!-- Right aligned nav items -->
        <b-navbar-nav class="ml-auto">
          <b-nav-item href="/create">Get your own dashboard</b-nav-item>
        </b-navbar-nav>
      </b-collapse>
    </b-navbar>

    <b-container class="mt-4">

      <b-row v-if="dash_id == 'welcome'">
        <b-col>
          <b-jumbotron header="Welcome to MonDash!" class="text-center">
            <p>You're currently seeing a demo dashboard updated with random numbers below. To get started read the <a href="http://docs.mondash.apiary.io/" target="_blank">API documentation</a> and create your own dashboard by clicking the button in the upper right hand corner&hellip;
            <p>If you have any questions about this project don't hesitate to ask <a href="https://ahlers.me/" target="_blank">Knut</a>.</p>
          </b-jumbotron>
        </b-col>
      </b-row>

      <b-row v-else-if="api_key && !metrics" class="justify-content-md-center">
        <b-col cols="12" md="6" class="text-center">
          <p>Welcome to your new dashboard. Your API-key is:</p>

          <code>{{ api_key }}</code>

          <p>
          After you sent your first metric you can reach your dashboard here:<br>
          <a :href="location">{{ location }}</a>
          </p>
        </b-col>
      </b-row>

      <metric v-for="metric in metrics" :metric="metric" :key="metric.id"></metric>

    </b-container>
  </div>
</template>

<script>
import axios from 'axios'

import metric from './metric.vue'

export default {
  name: 'app',

  components: {
    metric,
  },
  computed: {
    dash_id() {
      return window.location.pathname.substr(1)
    },

    location() {
      return window.location.href
    },
  },

  data() {
    return {
      api_key: null,
      metrics: [],
    }
  },

  methods: {
    updateDashboardData() {
      const path = window.location.pathname
      axios.get(`${path}.json?history_bar=true&value_history=true`)
        .then(resp => {
          this.api_key = resp.data.api_key
          this.metrics = resp.data.metrics
        })
        .catch(err => console.error(err))
    },
  },

  mounted() {
    this.updateDashboardData()
    window.setInterval(() => this.updateDashboardData(), 10000)
  },

  watch: {
  },
}
</script>

<style>
</style>
