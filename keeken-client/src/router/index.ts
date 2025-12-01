import { createRouter, createWebHistory } from 'vue-router'
import HomeView from '../views/HomeView.vue'
import PlayerView from '@/views/PlayerView.vue'
import HostView from '@/views/HostView.vue'
import SpectateView from '@/views/SpectateView.vue'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      name: 'home',
      component: HomeView,
    },
    {
      path: '/player',
      name: 'Player',
      component: PlayerView,
    },
    {
      path: '/host',
      name: 'Host',
      component: HostView,
    },
    {
      path: '/editor',
      name: 'Editor',
      component: () => import('../views/EditorView.vue'),
    },
    {
      path: '/spectate',
      name: 'Spectate',
      component: SpectateView,
    },
  ],
})

export default router
