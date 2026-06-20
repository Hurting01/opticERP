<script setup>
// Schedule — таблица графика по дням месяца.
//
// Данные теперь хранятся в БД (таблица schedule). При монтировании
// компонент:
//   1) подтягивает список сотрудников из staff (через api.staff.list),
//   2) запрашивает все записи schedule за текущий месяц (api.schedule.list),
//   3) склеивает сотрудников и смены в локальный массив scheduleData.
//
// Запись смены:
//   - клик по ячейке открывает мини-инпуты (1/к/Я/о или своё значение),
//   - при подтверждении дёргается api.schedule.saveShift, после чего
//     состояние в scheduleData обновляется,
//   - пустая смена (выходной) удаляет запись из БД.
import { ref, computed, onMounted } from 'vue';
import NavigationHeader from '../components/NavigationHeader.vue';
import api from '../api';
import { Modal } from 'bootstrap';

// Локальная директива: авто-фокус на input при появлении ячейки редактирования.
const vFocus = {
  mounted(el) {
    el.focus();
    el.select?.();
  },
};

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
const isLoadingSchedule = ref(false);
const lastError = ref('');

// scheduleData — единый источник правды для UI.
// Каждый элемент: { id, name, position, position_name, schedule: { [day]: shift }, hours, days, serviceType }
const scheduleData = ref([]);

// Редактируемая в данный момент ячейка: { employeeId, day, value, dirty }
const editingCell = ref(null);
const isSaving = ref(false);

// === работа с сотрудниками ===

async function loadEmployees() {
  isLoadingEmployees.value = true;
  try {
    const [staff, positions] = await Promise.all([
      api.staff.list(),
      api.positions.list(),
    ]);
    // На случай, если бэкенд вернул null/неопределено (например, в
    // старых версиях Go nil-слайс маршалится в null), нормализуем
    // значения к пустым массивам, чтобы .filter/.find не падали.
    const safeStaff = Array.isArray(staff) ? staff : [];
    const safePositions = Array.isArray(positions) ? positions : [];
    employees.value = safeStaff
      .filter((e) => e.is_active !== 0)
      .map((e) => {
        const pos = safePositions.find((p) => p.id === e.position_id);
        return {
          ...e,
          fullName: e.full_name,
          isActive: e.is_active !== 0,
          position: pos?.name || '',
          position_name: pos?.name || '',
        };
      });
  } catch (e) {
    lastError.value = `Ошибка загрузки сотрудников: ${e?.message || e}`;
    console.error(lastError.value);
  } finally {
    isLoadingEmployees.value = false;
  }
}

// === работа с графиком ===

function pad2(n) {
  return String(n).padStart(2, '0');
}

function dateKey(day) {
  return `${year}-${pad2(month)}-${pad2(day)}`;
}

function monthRange() {
  // Границы текущего месяца в формате YYYY-MM-DD.
  return {
    from: `${year}-${pad2(month)}-01`,
    to: `${year}-${pad2(month)}-${pad2(daysInMonth)}`,
  };
}

async function loadSchedule() {
  isLoadingSchedule.value = true;
  try {
    const { from, to } = monthRange();
    const [rows, members] = await Promise.all([
      api.schedule.list(from, to),
      api.schedule.members(year, month),
    ]);
    const safeRows = Array.isArray(rows) ? rows : [];
    const memberIds = new Set(Array.isArray(members) ? members : []);

    // Сгруппируем по user_id, чтобы быстро отрисовать.
    // map: user_id -> { [day(int)]: shift }
    const byUser = new Map();
    for (const r of safeRows) {
      const m = /^(\d{4})-(\d{2})-(\d{2})$/.exec(r.date || '');
      if (!m) continue;
      const d = parseInt(m[3], 10);
      if (!byUser.has(r.user_id)) byUser.set(r.user_id, {});
      byUser.get(r.user_id)[d] = r.shift;
    }

    // Показываем только тех сотрудников, у которых есть хотя бы одна
    // запись в графике за текущий месяц. Новые сотрудники (созданные в
    // настройках) НЕ появляются здесь автоматически — только после
    // добавления через модалку «Добавить в график».
    const result = [];
    for (const emp of employees.value) {
      const days = byUser.get(emp.id);
      if (!days && !memberIds.has(emp.id)) continue;
      result.push(buildEmployeeRow(emp, days || {}));
    }
    scheduleData.value = result;
  } catch (e) {
    lastError.value = `Ошибка загрузки графика: ${e?.message || e}`;
    console.error(lastError.value);
  } finally {
    isLoadingSchedule.value = false;
  }
}

function buildEmployeeRow(emp, daysMap) {
  let hours = 0;
  let days = 0;
  for (const day of Object.keys(daysMap)) {
    const shift = daysMap[day];
    if (shift && shift !== '') {
      days += 1;
      // 1 смена = 1 условный час. Точную конвертацию можно вынести в
      // настройки должности (Position.hours_per_shift) позже.
      hours += 1;
    }
  }
  return {
    id: emp.id,
    name: emp.fullName,
    schedule: { ...daysMap },
    hours,
    days,
    serviceType: emp.position_name || 'Без должности',
  };
}

// === модалки добавления/удаления сотрудника из таблицы ===

function openModal() {
  loadEmployees();
  modalInstance.value?.show();
}

function closeModal() {
  modalInstance.value?.hide();
}

async function addRecord() {
  if (!selectedEmployee.value) return;
  const employee = employees.value.find((e) => e.id === selectedEmployee.value);
  if (!employee) return;

  if (scheduleData.value.some((item) => item.id === employee.id)) {
    // Уже в таблице графика — это не ошибка, просто молча закрываем.
    closeModal();
    return;
  }

  try {
    await api.schedule.addMember(employee.id, year, month);
    scheduleData.value.push({
      id: employee.id,
      name: employee.fullName,
      schedule: {},
      hours: 0,
      days: 0,
      serviceType: employee.position_name || 'Без должности',
    });
    closeModal();
  } catch (e) {
    lastError.value = `Ошибка добавления сотрудника в график: ${e?.message || e}`;
    console.error(lastError.value);
  }
}

function removeRecord(employeeId) {
  const employee = scheduleData.value.find((item) => item.id === employeeId);
  if (employee) {
    employeeToDelete.value = employee;
    deleteModalInstance.value?.show();
  }
}

async function confirmDelete() {
  if (!employeeToDelete.value) return;
  const id = employeeToDelete.value.id;
  try {
    // Сносим все записи сотрудника из таблицы schedule.
    await api.schedule.clearForUser(id);
  } catch (e) {
    lastError.value = `Ошибка очистки графика: ${e?.message || e}`;
    console.error(lastError.value);
  }
  const index = scheduleData.value.findIndex((item) => item.id === id);
  if (index !== -1) {
    scheduleData.value.splice(index, 1);
  }
  closeDeleteModal();
}

function closeDeleteModal() {
  deleteModalInstance.value?.hide();
  employeeToDelete.value = null;
}

// === редактирование ячейки ===

function getDayOfWeek(day) {
  const date = new Date(year, month - 1, day);
  return weekDays[date.getDay()];
}

function getScheduleValue(employee, day) {
  return employee.schedule[day] || false;
}

function toggleShift(employee, day) {
  const currentValue = getScheduleValue(employee, day);
  const newValue = !currentValue;
  
  saveCell(employee, day, newValue);
}

function cancelEditCell() {
  editingCell.value = null;
}

function presetCycle(current) {
  // Цикл по 'легенде' графика: false → true → true → true → false → false
  switch (current) {
    case false: return true;
    case true: return true;
    default: return false;
  }
}

async function saveCell(employee, day, value) {
  if (isSaving.value) return;
  isSaving.value = true;
  try {
    const date = dateKey(day);
    const v = value ? '1' : '';
    console.log('[Schedule.saveCell]', { id: employee.id, date, shift: v });
    await api.schedule.saveShift(employee.id, date, v);
    console.log('[Schedule.saveCell] OK');

    employee.schedule[day] = value;
    recountEmployee(employee);
    editingCell.value = null;
  } catch (e) {
    lastError.value = `Не удалось сохранить смену: ${e?.message || e}`;
    console.error(lastError.value);
  } finally {
    isSaving.value = false;
  }
}

function recountEmployee(employee) {
  let hours = 0;
  let days = 0;
  for (const day of Object.keys(employee.schedule)) {
    const shift = employee.schedule[day];
    if (shift) {
      days += 1;
      hours += 1;
    }
  }
  employee.hours = hours;
  employee.days = days;
}

const days = computed(() => Array.from({ length: daysInMonth }, (_, i) => i + 1));

onMounted(async () => {
  await loadEmployees();
  await loadSchedule();

  if (modalEl.value) {
    modalInstance.value = new Modal(modalEl.value);
    modalEl.value.addEventListener('hidden.bs.modal', () => {
      selectedEmployee.value = '';
    });
  }
  if (deleteModalEl.value) {
    deleteModalInstance.value = new Modal(deleteModalEl.value);
    deleteModalEl.value.addEventListener('hidden.bs.modal', () => {
      employeeToDelete.value = null;
    });
  }
});
</script>

<template>
  <div>
    <NavigationHeader title="График" :hide-schedule="true" />

    <div v-if="lastError" class="alert alert-danger py-2 px-3" role="alert">
      {{ lastError }}
      <button type="button" class="btn-close" aria-label="Закрыть" @click="lastError = ''"></button>
    </div>

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
            <tr v-if="isLoadingSchedule">
              <td :colspan="daysInMonth + 4" class="text-center text-muted" style="padding: 24px">
                Загрузка графика из БД…
              </td>
            </tr>
            <template v-else>
              <tr v-if="scheduleData.length > 0" class="service-type-row">
                <td :colspan="daysInMonth + 4" class="service-type-cell">
                  {{ scheduleData[0].serviceType }}
                </td>
              </tr>
              <tr v-for="employee in scheduleData" :key="employee.id" class="employee-row">
                <td class="employee-name">{{ employee.name }}</td>
                <td
                  v-for="day in days"
                  :key="`${employee.id}-${day}`"
                  class="schedule-cell"
                  :class="{
                    'schedule-cell--working': getScheduleValue(employee, day),
                    'schedule-cell--editing': editingCell && editingCell.employeeId === employee.id && editingCell.day === day
                  }"
                  @click="!editingCell && toggleShift(employee, day)"
                >
                  <template v-if="editingCell && editingCell.employeeId === employee.id && editingCell.day === day">
                    <input
                      v-model="editingCell.value"
                      type="checkbox"
                      class="cell-checkbox"
                      :disabled="isSaving"
                      @change="saveCell(employee, day, editingCell.value)"
                      v-focus
                    />
                  </template>
                  <template v-else>
                    <span class="cell-indicator" :class="{ 'cell-indicator--active': getScheduleValue(employee, day) }"></span>
                  </template>
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
              <tr v-if="scheduleData.length === 0 && !isLoadingSchedule">
                <td :colspan="daysInMonth + 4" class="text-center text-muted" style="padding: 24px">
                  Добавьте сотрудников, чтобы увидеть график
                </td>
              </tr>
            </template>
          </tbody>
        </table>
      </div>
    </div>

    <div class="card">
      <div class="card-body legend-body">
        <div class="legend-item"><span class="legend-label">☑</span><span>— рабочий день</span></div>
        <div class="legend-item text-muted">
          <span>Клик по ячейке — отметить/снять рабочий день.</span>
        </div>
      </div>
    </div>

    <!-- Bootstrap Modal: добавление записи -->
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
            <p class="text-muted mb-0">Все сохранённые смены этого сотрудника будут удалены из БД.</p>
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
.schedule-cell { min-width: 32px; cursor: pointer; }
.schedule-cell:not(.schedule-cell--working):hover { background: #f1f5f9; }
.schedule-cell--working:hover { background: rgba(59, 130, 246, 0.7); }
.schedule-cell--working { background: var(--color-primary); }
.schedule-cell--editing { padding: 0 !important; background: #fffbe6; }
.cell-checkbox {
  width: 16px;
  height: 16px;
  cursor: pointer;
  accent-color: var(--color-primary);
  margin: 0;
}
.cell-indicator {
  width: 16px;
  height: 16px;
  border-radius: 4px;
  background: transparent;
  transition: all var(--transition);
}
.cell-indicator--active {
  background: var(--color-primary);
}
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
