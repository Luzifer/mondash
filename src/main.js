import Vue from 'vue'
import BootstrapVue from 'bootstrap-vue'

import app from './app.vue'

import 'bootstrap/dist/css/bootstrap.css'
import 'bootstrap-vue/dist/bootstrap-vue.css'
import 'bootswatch/dist/cosmo/bootstrap.min.css'
import 'chartist/dist/chartist.min.css'

Vue.use(BootstrapVue)

new Vue({
  el: '#app',
  components: { app },
  render: c => c('app'),
})
