// src/router/index.js
// Маршрутизация OpticERP. Используем hash-историю —
// идентично Tauri-версии, маршруты не зависят от сервера.
import { createRouter, createWebHashHistory } from 'vue-router';

import Dashboard from '../views/Dashboard.vue';
import DaySheet from '../views/DaySheet.vue';
import Reports from '../views/Reports.vue';
import Settings from '../views/Settings.vue';
import Schedule from '../views/Schedule.vue';

const routes = [
  { path: '/', name: 'dashboard', component: Dashboard, meta: { title: 'OpticERP' } },
  { path: '/day/:day', name: 'day', component: DaySheet, meta: { title: 'День' } },
  { path: '/reports', name: 'reports', component: Reports, meta: { title: 'Отчёты' } },
  { path: '/settings', name: 'settings', component: Settings, meta: { title: 'Настройки' } },
  { path: '/schedule', name: 'schedule', component: Schedule, meta: { title: 'График' } },
  { path: '/:pathMatch(.*)*', redirect: '/' },
];

const router = createRouter({
  history: createWebHashHistory(),
  routes,
});

router.afterEach((to) => {
  if (to.meta && to.meta.title) {
    document.title = `${to.meta.title} — OpticERP`;
  }
});

export default router;
