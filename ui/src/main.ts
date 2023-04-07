import { createApp } from 'vue'
import { i18n } from './i18n'
import App from './App.vue'
import { router } from './router'
import { pinia } from './store'
import 'uno.css'

createApp(App).use(router).use(pinia).use(i18n).mount('#app')
