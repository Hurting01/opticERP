// src/main.js
import { createApp } from 'vue';
import App from './App.vue';
import router from './router';

// Подкл��чаем стили. Bootstrap 5 даёт формы/таблицы/модалки, наш
// index.css переопределяет оттенки и добавляет утилиты под Tauri-версию.
import 'bootstrap/dist/css/bootstrap.min.css';
import './styles/index.css';

const app = createApp(App);
app.use(router);
app.mount('#app');
