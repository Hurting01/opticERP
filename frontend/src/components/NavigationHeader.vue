<script setup>
// Верхняя панель навигации. Аналог react-bootstrap-версии из
// opticTauri/src/components/NavigationHeader.jsx, но для Vue.
import { useRouter, useRoute } from 'vue-router';
import { computed } from 'vue';

const props = defineProps({
  title: { type: String, default: 'OpticERP' },
  showBackButton: { type: Boolean, default: false },
  backUrl: { type: String, default: '/' },
});

const router = useRouter();
const route = useRoute();
const isHome = computed(() => route.path === '/' || route.path === '');
</script>

<template>
  <div class="top-panel fade-in">
    <div class="top-panel-body">
      <div class="header-row">
        <h1 class="title">{{ title }}</h1>

        <div v-if="showBackButton" class="top-buttons">
          <button class="btn btn-light" @click="router.push(backUrl)">← Назад</button>
        </div>

        <div v-else class="top-buttons">
          <button class="btn btn-light" @click="router.push('/reports')">Отчеты</button>
          <button class="btn btn-light" @click="router.push('/schedule')">График</button>
          <button class="btn btn-light" disabled>Касса</button>
          <button class="btn btn-light" @click="router.push('/settings')">Настройки</button>
          <button v-if="!isHome" class="btn btn-light home-btn" @click="router.push('/')" title="Главная">
            <svg width="22" height="22" viewBox="0 0 24 24" fill="currentColor" xmlns="http://www.w3.org/2000/svg">
              <path d="M10 20v-6h4v6h5v-8h3L12 3 2 12h3v8z" />
            </svg>
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.top-panel {
  background: #fff;
  border: 1px solid var(--color-border);
  border-radius: var(--radius);
  box-shadow: var(--shadow-sm);
  margin-bottom: 16px;
}
.top-panel-body {
  padding: 16px 20px;
}
.header-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  flex-wrap: wrap;
}
.title {
  margin: 0;
  font-size: 22px;
  font-weight: 700;
  color: var(--color-text);
}
.top-buttons {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}
.home-btn {
  padding: 6px 10px;
}
</style>
