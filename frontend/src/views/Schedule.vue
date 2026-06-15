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
const deleteModalInstance = ref(null);
const deleteModalEl = ref(null);
const employees = ref([]);
const selectedEmployee = ref('');
const employeeToDelete = ref(null);
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
  
  const employee = employees.value.find((e) => e.id === selectedEmployee.value);
  if (!employee) return;
  
  // Проверяем, не добавлен ли уже этот сотрудник
  if (scheduleData.value.some((item) => item.id === employee.id)) {
    console.warn('Сотрудник уже добавлен в график');
    closeModal();
    return;
  }
  
  // Добавляем нового сотрудника в график
  scheduleData.value.push({
    id: employee.id,
    name: employee.fullName,
    schedule: {},
    hours: 0,
    days: 0,
    serviceType: employee.position_name || 'Без должности',
  });
  
  closeModal();
}

function removeRecord(employeeId) {
  console.log('removeRecord вызвана для ID:', employeeId);
  const employee = scheduleData.value.find((item) => item.id === employeeId);
  console.log('Найден сотрудник:', employee);
  if (employee) {
    employeeToDelete.value = employee;
    console.log('Открываем модальное окно удаления');
    deleteModalInstance.value?.show();
  }
}

function confirmDelete() {
  if (!employeeToDelete.value) return;
  
  const index = scheduleData.value.findIndex((item) => item.id === employeeToDelete.value.id);
  if (index !== -1) {
    scheduleData.value.splice(index, 1);
  }
  
  closeDeleteModal();
}

function closeDeleteModal() {
  deleteModalInstance.value?.hide();
  employeeToDelete.value = null;
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
  
  // Инициализируем модальное окно подтверждения удаления
  if (deleteModalEl.value) {
    deleteModalInstance.value = new Modal(deleteModalEl.value);
    
    // Очищаем данные при закрытии модального окна
    deleteModalEl.value.addEventListener('hidden.bs.modal', () => {
      employeeToDelete.value = null;
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
              <th class="action-column"></th>
            </tr>
          </thead>
          <tbody>
            <tr v-if="scheduleData.length > 0" class="service-type-row">
              <td :colspan="daysInMonth + 4" class="service-type-cell">
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
              <td class="action-cell">
                <button 
                  type="button"
                  class="btn-delete" 
                  @click.stop="removeRecord(employee.id)"
                  title="Удалить из графика"
                >
                  ✕
                </button>
              </td>
            </tr>
            <tr v-if="scheduleData.length === 0">
              <td :colspan="daysInMonth + 4" class="text-center text-muted" style="padding: 24px">
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

    <!-- Модальное окно подтверждения удаления -->
    <div ref="deleteModalEl" class="modal fade" tabindex="-1" aria-labelledby="deleteModalLabel" aria-hidden="true">
      <div class="modal-dialog">
        <div class="modal-content">
          <div class="modal-header">
            <h5 id="deleteModalLabel" class="modal-title">Подтверждение удаления</h5>
            <button type="button" class="btn-close" data-bs-dismiss="modal" aria-label="Close"></button>
          </div>
          <div class="modal-body">
            <p>Вы уверены, что хотите удалить сотрудника <strong>{{ employeeToDelete?.name }}</strong> из графика?</p>
            <p class="text-muted mb-0">Все данные графика для этого сотрудника будут потеряны.</p>
          </div>
          <div class="modal-footer">
            <button type="button" class="btn btn-secondary" data-bs-dismiss="modal">Отмена</button>
            <button type="button" class="btn btn-danger" @click="confirmDelete">Удалить</button>
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
.action-column {
  min-width: 50px;
  background: #f8fafc;
}
.action-cell {
  text-align: center;
  padding: 4px !important;
  position: relative;
}
.btn-delete {
  position: relative;
  z-index: 10;
  background: transparent;
  border: none;
  color: #dc3545;
  font-size: 18px;
  line-height: 1;
  padding: 2px 8px;
  cursor: pointer;
  border-radius: 4px;
  transition: background-color 0.2s;
}
.btn-delete:hover {
  background: #fee;
  color: #c82333;
}
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
