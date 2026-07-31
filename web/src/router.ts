import { createRouter, createWebHashHistory } from 'vue-router'

export default createRouter({
  history: createWebHashHistory(),
  routes: [
    { path: '/', component: () => import('./pages/OverviewPage.vue') },
    { path: '/tasks', component: () => import('./pages/TaskListPage.vue') },
    { path: '/logs', component: () => import('./pages/LogPage.vue') },
    { path: '/settings', component: () => import('./pages/SettingsPage.vue') },
  ],
})
