import './assets/main.css'

import { createApp } from 'vue'
import { createPinia } from 'pinia'
import { VueResponsiveness, Presets } from '@/lib/vue-responsiveness'

import App from './App.vue'
import router from './router'

const app = createApp(App)

app.use(createPinia())
app.use(VueResponsiveness, Presets.Tailwind_CSS)
app.use(router)

app.mount('#app')
