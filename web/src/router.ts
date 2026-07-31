import { createRouter, createWebHashHistory } from 'vue-router'

const router = createRouter({
  history: createWebHashHistory(),
  routes: [
    { path: '/', name: 'overview', component: () => import('./pages/OverviewPage.vue') },
    { path: '/tasks', name: 'tasks', component: () => import('./pages/TaskListPage.vue') },
    { path: '/logs', name: 'logs', component: () => import('./pages/LogPage.vue') },
    { path: '/settings', name: 'settings', component: () => import('./pages/SettingsPage.vue') },
  ],
})

export default router
