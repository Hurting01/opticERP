<script setup>
// Reports — вкладки: Бонусы, ЗП, Конверсия, Заказы.
// Все данные хранятся в localStorage по ключам "<report>:<year>-<month>".
import { ref, computed, onMounted, watch } from 'vue';
import NavigationHeader from '../components/NavigationHeader.vue';
import api from '../api';

const month = new Date().getMonth() + 1;
const year = new Date().getFullYear();

const monthNames = [
  'Январь', 'Февраль', 'Март', 'Апрель', 'Май', 'Июнь',
  'Июль', 'Август', 'Сентябрь', 'Октябрь', 'Ноябрь', 'Декабрь',
];

const activeTab = ref('bonus');
const employees = ref([]);
const bonusData = ref([]);
const salaryData = ref([]);
const conversionData = ref([]);
const ordersData = ref([]);
const saveStatus = ref(null);

function reportKey(prefix) {
  return `${prefix}:${year}-${month}`;
}

function getStored(key) {
  const v = localStorage.getItem(key);
  return v ? JSON.parse(v) : null;
}
function setStored(key, value) {
  localStorage.setItem(key, JSON.stringify(value));
}

async function loadEmployees() {
  try {
    const staff = await api.staff.list();
    const positions = await api.positions.list();
    employees.value = staff.map((e) => {
      // Бэкенд уже подтянул position_name через JOIN; используем его напрямую.
      const pos = positions.find((p) => p.id === e.position_id);
      const positionName = e.position_name || pos?.name || '';
      return {
        ...e,
        fullName: e.full_name,
        isActive: e.is_active !== 0,
        position: positionName,
        position_name: positionName,
      };
    });
  } catch (err) {
    console.error('Ошибка загрузки сотрудников:', err);
  }
}

function loadBonusData() {
  const saved = getStored(reportKey('bonusReport'));
  if (saved && saved.data && saved.data.length > 0) {
    bonusData.value = saved.data;
    return;
  }
  const daysInMonth = new Date(year, month, 0).getDate();
  const data = [];
  for (let day = 1; day <= daysInMonth; day++) {
    const dayData = { day };
    employees.value.forEach((emp) => {
      if (emp.fullName && emp.isActive !== false) {
        dayData[emp.fullName] = Math.random() > 0.3 ? (Math.random() * 1000).toFixed(2) : 0;
      }
    });
    data.push(dayData);
  }
  bonusData.value = data;
  setStored(reportKey('bonusReport'), { month, year, data });
}

function loadSalaryData() {
  const saved = getStored(reportKey('salaryReport'));
  if (saved && saved.data && saved.data.length > 0) {
    salaryData.value = saved.data;
    return;
  }
  salaryData.value = [
    { name: 'Письминская Ю.', base: 37500, bonus: 11795.47, extra: 8000, total: 54920.47 },
    { name: 'Каргина Е.', base: 37500, bonus: 11795.47, extra: 8000, total: 54920.47 },
    { name: 'Липенкова Т.', base: 40000, bonus: 12000, extra: 5000, total: 57000 },
    { name: 'Машалова Т.', base: 37500, bonus: 11795.47, extra: 8000, total: 54920.47 },
  ];
  setStored(reportKey('salaryReport'), { month, year, data: salaryData.value });
}

function loadConversionData() {
  const saved = getStored(reportKey('conversionReport'));
  if (saved && saved.data && saved.data.length > 0) {
    conversionData.value = saved.data;
    return;
  }
  conversionData.value = [
    { date: '01.03.2026', employees: 'Письминская Ю./Машалова Т.', visitors: 17, sales: 3, conversion: 17.6, orders: 0, diagnostics: 1 },
    { date: '02.03.2026', employees: 'Каргина Е./Липенкова Т.', visitors: 12, sales: 2, conversion: 16.7, orders: 1, diagnostics: 0 },
  ];
  setStored(reportKey('conversionReport'), { month, year, data: conversionData.value });
}

function loadOrdersData() {
  const saved = getStored(reportKey('ordersReport'));
  if (saved && saved.data && saved.data.length > 0) {
    ordersData.value = saved.data;
    return;
  }
  ordersData.value = [
    { id: 1, item: 'Очки солнцезащитные', status: '✓' },
    { id: 2, item: 'Линзы контактные', status: '✗' },
    { id: 3, item: 'Оправа пластиковая', status: '✓' },
  ];
  setStored(reportKey('ordersReport'), { month, year, data: ordersData.value });
}

function handleBonusChange(dayIdx, empName, value) {
  const copy = [...bonusData.value];
  copy[dayIdx] = { ...copy[dayIdx], [empName]: value };
  bonusData.value = copy;
}

function showSave(name) {
  saveStatus.value = name;
  setTimeout(() => { if (saveStatus.value === name) saveStatus.value = null; }, 2000);
}

const filteredEmployees = computed(() => employees.value.filter((e) => e.fullName));

onMounted(async () => {
  await loadEmployees();
  loadBonusData();
  loadSalaryData();
  loadConversionData();
  loadOrdersData();
});

let timers = {};
watch(bonusData, (val) => {
  if (!val || val.length === 0) return;
  if (timers.bonus) clearTimeout(timers.bonus);
  timers.bonus = setTimeout(() => {
    setStored(reportKey('bonusReport'), { month, year, data: val });
    showSave('bonus');
  }, 1000);
}, { deep: true });

watch(conversionData, (val) => {
  if (!val || val.length === 0) return;
  if (timers.conversion) clearTimeout(timers.conversion);
  timers.conversion = setTimeout(() => {
    setStored(reportKey('conversionReport'), { month, year, data: val });
    showSave('conversion');
  }, 1000);
}, { deep: true });

watch(ordersData, (val) => {
  if (!val || val.length === 0) return;
  if (timers.orders) clearTimeout(timers.orders);
  timers.orders = setTimeout(() => {
    setStored(reportKey('ordersReport'), { month, year, data: val });
    showSave('orders');
  }, 1000);
}, { deep: true });

function fmt(n) {
  return n.toLocaleString('ru-RU');
}
</script>

<template>
  <div>
    <NavigationHeader title="Отчеты" :hide-reports="true" />

    <ul class="nav nav-tabs mb-3">
      <li class="nav-item">
        <button type="button" class="nav-link" :class="{ active: activeTab === 'bonus' }" @click="activeTab = 'bonus'">Бонусы</button>
      </li>
      <li class="nav-item">
        <button type="button" class="nav-link" :class="{ active: activeTab === 'salary' }" @click="activeTab = 'salary'">ЗП</button>
      </li>
      <li class="nav-item">
        <button type="button" class="nav-link" :class="{ active: activeTab === 'conversion' }" @click="activeTab = 'conversion'">Конверсия</button>
      </li>
      <li class="nav-item">
        <button type="button" class="nav-link" :class="{ active: activeTab === 'orders' }" @click="activeTab = 'orders'">Заказы</button>
      </li>
    </ul>

    <!-- Бонусы -->
    <div v-if="activeTab === 'bonus'" class="card">
      <div class="card-header card-header-row">
        <span>📋 Итоги бонусов по продажам за {{ monthNames[month - 1] }} {{ year }}</span>
        <span v-if="saveStatus === 'bonus'" class="save-status">✓ Сохранено</span>
      </div>
      <div class="card-body">
        <div v-if="filteredEmployees.length === 0" class="empty-state">
          <p class="mb-0">Сотрудники не указаны</p>
          <small>Добавьте сотрудников в настройках</small>
        </div>
        <div v-else class="table-wrapper">
          <table class="table table-striped table-bordered table-sm">
            <thead>
              <tr>
                <th style="width:60px">День</th>
                <th v-for="emp in filteredEmployees" :key="emp.id">{{ emp.fullName }}</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="(dayData, idx) in bonusData" :key="idx">
                <td class="text-center">{{ dayData.day }}</td>
                <td v-for="emp in filteredEmployees" :key="emp.id" class="text-center">
                  <input
                    type="text"
                    class="form-control form-control-sm"
                    :value="dayData[emp.fullName] || ''"
                    @input="handleBonusChange(idx, emp.fullName, $event.target.value)"
                  />
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>

    <!-- ЗП -->
    <div v-if="activeTab === 'salary'" class="card">
      <div class="card-header">Заработная плата за {{ monthNames[month - 1] }} {{ year }}</div>
      <div class="card-body">
        <table class="table table-striped table-bordered table-hover">
          <thead>
            <tr>
              <th>ФИО</th>
              <th>Оклад (₽)</th>
              <th>Бонусы (₽)</th>
              <th>Доп. выплаты (₽)</th>
              <th>Итог (₽)</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="(emp, idx) in salaryData" :key="idx">
              <td>{{ emp.name }}</td>
              <td>{{ fmt(emp.base) }}</td>
              <td>{{ fmt(emp.bonus) }}</td>
              <td>{{ fmt(emp.extra) }}</td>
              <td><strong>{{ fmt(emp.total) }}</strong></td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- Конверсия -->
    <div v-if="activeTab === 'conversion'" class="card">
      <div class="card-header card-header-row">
        <span>Конверсия продаж</span>
        <span v-if="saveStatus === 'conversion'" class="save-status">✓ Сохранено</span>
      </div>
      <div class="card-body">
        <table class="table table-striped table-bordered table-hover table-sm">
          <thead>
            <tr>
              <th>Дата</th>
              <th>Сотрудники</th>
              <th>Посетители</th>
              <th>Продажи</th>
              <th>Конверсия %</th>
              <th>Заказы</th>
              <th>Диагностики</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="(row, idx) in conversionData" :key="idx">
              <td>{{ row.date }}</td>
              <td>{{ row.employees }}</td>
              <td>{{ row.visitors }}</td>
              <td>{{ row.sales }}</td>
              <td>{{ row.conversion }}%</td>
              <td>{{ row.orders }}</td>
              <td>{{ row.diagnostics }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- Заказы -->
    <div v-if="activeTab === 'orders'" class="card">
      <div class="card-header card-header-row">
        <span>Заказ позиций</span>
        <span v-if="saveStatus === 'orders'" class="save-status">✓ Сохранено</span>
      </div>
      <div class="card-body">
        <table class="table table-striped table-bordered table-hover">
          <thead>
            <tr>
              <th>№</th>
              <th>Наименование</th>
              <th>Статус</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="order in ordersData" :key="order.id">
              <td>{{ order.id }}</td>
              <td>{{ order.item }}</td>
              <td>
                <span :class="order.status === '✓' ? 'text-success' : 'text-danger'">{{ order.status }}</span>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>

<style scoped>
.card-header-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.save-status {
  font-size: 12px;
  color: var(--color-success);
  font-weight: 600;
  animation: fadeIn 0.3s ease;
}
.empty-state {
  text-align: center;
  padding: 24px;
  color: var(--color-muted);
}
.table-wrapper {
  overflow-x: auto;
}
</style>
