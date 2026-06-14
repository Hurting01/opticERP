<script setup>
// Schedule — таблица графика по дням месяца. Исторически
// данные графика — мок (см. оригинал); сейчас читаем сотрудников
// из БД, график пока статический-заглушка, как в Tauri-версии.
import { ref, computed, onMounted } from 'vue';
import NavigationHeader from '../components/NavigationHeader.vue';
import api from '../api';
import { Modal } from 'bootstrap';

const month = new Date().getMonth() + 1;
const year = new Date().getFullYear();

const monthNames = [
  'Январь', 'Февраль', 'Март', 'Апрель', 'Май', 'Июнь',
  'Июль', 'Август', 'Сентябрь', 'Октябрь', 'Ноябрь', 'Декабрь',
];
const weekDays = ['Вс', 'Пн', 'Вт', 'Ср', 'Чт', 'Пт', 'Сб'];

const daysInMonth = new Date(year, month, 0).getDate();
const modalInstance = ref(null);
const modalEl = ref(null);
const employees = ref([]);
const selectedEmployee = ref('');
const isLoadingEmployees = ref(false);

// В оригинале данные статические. Подменим под реальный список сотрудников,
// чтобы шаблон был непустой, но логика графика — заглушка.
const scheduleData = ref([]);

async function loadEmployees() {
  isLoadingEmployees.value = true;
  try {
    const staff = await api.staff.list();
    const positions = await api.positions.list();
    employees.value = staff
      .filter((e) => e.is_active !== 0)
      .map((e) => {
        const pos = positions.find((p) => p.id === e.position_id);
        return {
          ...e,
          fullName: e.full_name,
          isActive: e.is_active !== 0,
          position: pos?.name || '',
          position_name: pos?.name || '',
        };
      });
    // сгенерируем пустые записи графика
    scheduleData.value = employees.value.map((emp, idx) => ({
      id: emp.id,
      name: emp.fullName,
      schedule: {},
      hours: 0,
      days: 0,
      serviceType: idx === 0 ? 'Оптик-консультанты с 10.00 до 22.00' : 'Оптометристы с 10.00 до 20.00',
    }));
  } finally {
    isLoadingEmployees.value = false;
  }
}

function openModal() {
  loadEmployees();
  modalInstance.value?.show();
}

function closeModal() {
  modalInstance.value?.hide();
}

function addRecord() {
  if (!selectedEmployee.value) return;
  console.log('Добавить запись для сотрудника:', selectedEmployee.value);
  closeModal();
}

function getDayOfWeek(day) {
  const date = new Date(year, month - 1, day);
  return weekDays[date.getDay()];
}

function getScheduleValue(employee, day) {
  return employee.schedule[day] || '';
}

const days = computed(() => Array.from({ length: daysInMonth }, (_, i) => i + 1));

onMounted(() => {
  loadEmployees();
  
  // Инициализируем Bootstrap Modal после монтирования компонента
  if (modalEl.value) {
    modalInstance.value = new Modal(modalEl.value);
    
    // Очищаем форму при закрытии модального окна
    modalEl.value.addEventListener('hidden.bs.modal', () => {
      selectedEmployee.value = '';
    });
  }
});
</script>

<template>
  <div>
    <NavigationHeader title="График" />

    <div class="card">
      <div class="card-body py-3">
        <div class="d-flex justify-content-between align-items-center">
          <h1 class="page-title">График — {{ monthNames[month - 1] }} {{ year }}</h1>
          <button class="btn btn-primary" @click="openModal">+ Добавить запись</button>
        </div>
      </div>
    </div>

    <div class="card">
      <div class="table-wrapper">
        <table class="schedule-table">
          <thead>
            <tr class="header-row">
              <th class="name-column"><span class="month-label">{{ monthNames[month - 1] }}</span></th>
              <th v-for="day in days" :key="day" class="day-header">
                <div class="day-number">{{ day }}</div>
                <div class="day-of-week">{{ getDayOfWeek(day) }}</div>
              </th>
              <th class="hours-column">Часы</th>
              <th class="days-column">Дни</th>
            </tr>
          </thead>
          <tbody>
            <tr v-if="scheduleData.length > 0" class="service-type-row">
              <td :colspan="daysInMonth + 3" class="service-type-cell">
                {{ scheduleData[0].serviceType }}
              </td>
            </tr>
            <tr v-for="employee in scheduleData" :key="employee.id" class="employee-row">
              <td class="employee-name">{{ employee.name }}</td>
              <td v-for="day in days" :key="`${employee.id}-${day}`" class="schedule-cell">
                {{ getScheduleValue(employee, day) }}
              </td>
              <td class="hours-cell">{{ employee.hours.toFixed(1) }}</td>
              <td class="days-cell">{{ employee.days }}</td>
            </tr>
            <tr v-if="scheduleData.length === 0">
              <td :colspan="daysInMonth + 3" class="text-center text-muted" style="padding: 24px">
                Добавьте сотрудников в настройках, чтобы увидеть график
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <div class="card">
      <div class="card-body legend-body">
        <div class="legend-item"><span class="legend-label">1</span><span>— рабочий день</span></div>
        <div class="legend-item"><span class="legend-label">к</span><span>— командировка</span></div>
        <div class="legend-item"><span class="legend-label">Я</span><span>— отпуск</span></div>
        <div class="legend-item"><span class="legend-label">о</span><span>— выходной</span></div>
      </div>
    </div>

    <!-- Bootstrap Modal -->
    <div ref="modalEl" class="modal fade" tabindex="-1" aria-labelledby="addRecordModalLabel" aria-hidden="true">
      <div class="modal-dialog">
        <div class="modal-content">
          <div class="modal-header">
            <h5 id="addRecordModalLabel" class="modal-title">Добавить запись</h5>
            <button type="button" class="btn-close" data-bs-dismiss="modal" aria-label="Close"></button>
          </div>
          <div class="modal-body">
            <div class="mb-3">
              <label for="employeeSelect" class="form-label">Сотрудник</label>
              <select
                id="employeeSelect"
                v-model="selectedEmployee"
                class="form-select"
                :disabled="isLoadingEmployees"
              >
                <option value="">Выберите сотрудника</option>
                <option v-for="emp in employees" :key="emp.id" :value="emp.id">{{ emp.fullName }}</option>
              </select>
            </div>
          </div>
          <div class="modal-footer">
            <button type="button" class="btn btn-secondary" data-bs-dismiss="modal">Отмена</button>
            <button type="button" class="btn btn-primary" :disabled="!selectedEmployee" @click="addRecord">Добавить</button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.page-title {
  font-size: 20px;
  margin: 0;
  font-weight: 600;
}
.table-wrapper {
  overflow-x: auto;
}
.schedule-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 12px;
  min-width: 1100px;
}
.schedule-table th, .schedule-table td {
  border: 1px solid var(--color-border);
  padding: 4px 6px;
  text-align: center;
}
.name-column {
  position: sticky;
  left: 0;
  background: #fff;
  z-index: 1;
  min-width: 200px;
  text-align: left;
}
.month-label {
  font-weight: 600;
  padding-left: 8px;
}
.day-header {
  min-width: 32px;
  background: #f8fafc;
}
.day-number { font-weight: 600; }
.day-of-week { font-size: 10px; color: var(--color-muted); }
.hours-column, .days-column { min-width: 60px; background: #f8fafc; }
.service-type-cell {
  text-align: left !important;
  font-style: italic;
  padding: 6px 12px !important;
  color: var(--color-muted);
  background: #fafbfc;
}
.employee-name {
  position: sticky;
  left: 0;
  background: #fff;
  text-align: left;
  padding-left: 12px !important;
  font-weight: 500;
  border-right: 2px solid var(--color-border);
}
.schedule-cell { min-width: 32px; }
.legend-body {
  display: flex;
  gap: 24px;
  flex-wrap: wrap;
}
.legend-item {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 12px;
}
.legend-label {
  background: #f1f5f9;
  border: 1px solid var(--color-border);
  border-radius: 4px;
  padding: 2px 8px;
  font-weight: 600;
  min-width: 24px;
  text-align: center;
}
</style>
