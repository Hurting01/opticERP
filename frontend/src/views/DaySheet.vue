<script setup>
// DaySheet — детализация конкретного дня: сотрудники, продажи, касса, задачи.
// Все данные хранятся в localStorage (как в Tauri-версии).
import { ref, computed, watch, onMounted, onBeforeUnmount } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import NavigationHeader from '../components/NavigationHeader.vue';
import api from '../api';

const route = useRoute();
const router = useRouter();

const dayNum = computed(() => parseInt(route.params.day, 10));
const month = computed(() => new Date().getMonth() + 1);
const year = computed(() => new Date().getFullYear());

const allEmployees = ref([]);
const dayEmployees = ref([
  { id: null, position: '', fullName: '' },
  { id: null, position: '', fullName: '' },
]);
const cashMorning = ref('');
const cashEvening = ref('');
const sales = ref([]);
const tasks = ref([]);
const openAccordion = ref(null);
const showDeleteConfirm = ref(false);
const saleToDelete = ref(null);
const newTaskText = ref('');
const taskPage = ref(0);
const salePage = ref(0);
const tasksPerPage = 4;
const salesPerPage = 10;

const saveTimeout = ref(null);

const monthNamesGenitive = [
  'января', 'февраля', 'марта', 'апреля', 'мая', 'июня',
  'июля', 'августа', 'сентября', 'октября', 'ноября', 'декабря',
];
const weekDays = ['Вс', 'Пн', 'Вт', 'Ср', 'Чт', 'Пт', 'Сб'];

const firstDayOfMonth = computed(() => new Date(year.value, month.value - 1, 1).getDay());
const daysInMonth = computed(() => new Date(year.value, month.value, 0).getDate());

const calendarDays = computed(() => {
  const cells = [];
  for (let i = 0; i < firstDayOfMonth.value; i++) {
    cells.push({ type: 'empty', key: `empty-${i}` });
  }
  for (let d = 1; d <= daysInMonth.value; d++) {
    const date = new Date(year.value, month.value - 1, d);
    cells.push({
      type: 'day',
      day: d,
      weekDay: weekDays[date.getDay()],
      isToday: d === new Date().getDate() && month.value === new Date().getMonth() + 1,
      key: `day-${d}`,
    });
  }
  return cells;
});

const sphValues = (() => {
  const v = ['0'];
  for (let x = -0.25; x >= -20.0; x -= 0.25) v.push(x.toFixed(2));
  for (let x = 0.25; x <= 20.0; x += 0.25) v.push('+' + x.toFixed(2));
  return v;
})();
const cylValues = (() => {
  const v = ['0'];
  for (let x = -0.25; x >= -10.0; x -= 0.25) v.push(x.toFixed(2));
  for (let x = 0.25; x <= 10.0; x += 0.25) v.push('+' + x.toFixed(2));
  return v;
})();
const axValues = (() => {
  const v = [];
  for (let x = 0; x < 180; x += 5) v.push(x.toString());
  return v;
})();

const totalPages = computed(() => Math.ceil(tasks.value.length / tasksPerPage));
const currentTasks = computed(() => tasks.value.slice(taskPage.value * tasksPerPage, (taskPage.value + 1) * tasksPerPage));
const totalSalesPages = computed(() => Math.ceil(sales.value.length / salesPerPage));
const currentSales = computed(() => sales.value.slice(salePage.value * salesPerPage, (salePage.value + 1) * salesPerPage));

const totalCash = computed(() => (parseFloat(cashMorning.value) || 0) + (parseFloat(cashEvening.value) || 0));
const totalCashless = computed(() => sales.value.reduce((s, sale) => s + (parseFloat(sale.price) || 0), 0));
const totalCard = computed(() => totalCash.value + totalCashless.value);

const title = computed(() => `${dayNum.value} ${monthNamesGenitive[month.value - 1]} ${year.value}`);

function storageKey() {
  return `day:${year.value}-${month.value}-${dayNum.value}`;
}

function loadData() {
  const value = localStorage.getItem(storageKey());
  if (value) {
    const data = JSON.parse(value);
    dayEmployees.value = data.employees || [{ id: null, position: '', fullName: '' }, { id: null, position: '', fullName: '' }];
    cashMorning.value = data.cashMorning || '';
    cashEvening.value = data.cashEvening || '';
    sales.value = data.sales || [];
    tasks.value = data.tasks || [];
  } else {
    dayEmployees.value = [{ id: null, position: '', fullName: '' }, { id: null, position: '', fullName: '' }];
    cashMorning.value = '';
    cashEvening.value = '';
    sales.value = [];
    tasks.value = [];
  }
}

function saveDayData() {
  localStorage.setItem(
    storageKey(),
    JSON.stringify({
      employees: dayEmployees.value,
      cashMorning: cashMorning.value,
      cashEvening: cashEvening.value,
      sales: sales.value,
      tasks: tasks.value,
    })
  );
}

async function loadEmployees() {
  try {
    const staff = await api.staff.list();
    const positions = await api.positions.list();
    allEmployees.value = staff.map((e) => {
      const pos = positions.find((p) => p.id === e.position_id);
      return {
        ...e,
        fullName: e.full_name,
        isActive: e.is_active !== 0,
        position: pos?.name || '',
        position_name: pos?.name || '',
      };
    });
  } catch (err) {
    console.error('Ошибка загрузки сотрудников:', err);
  }
}

function addSale() {
  sales.value.push({
    id: Date.now() + Math.floor(Math.random() * 1000),
    name: '',
    price: '',
    sph: '',
    cyl: '',
    ax: '',
    add: '',
    pd: '',
    material: '',
    coating: '',
  });
}

function updateSale(id, field, value) {
  sales.value = sales.value.map((s) => (s.id === id ? { ...s, [field]: value } : s));
}

function deleteSale(id) {
  sales.value = sales.value.filter((s) => s.id !== id);
  showDeleteConfirm.value = false;
}

function confirmDelete() {
  if (saleToDelete.value != null) deleteSale(saleToDelete.value);
}

function showDeleteModal(id) {
  saleToDelete.value = id;
  showDeleteConfirm.value = true;
}

function cancelDelete() {
  showDeleteConfirm.value = false;
  saleToDelete.value = null;
}

function addTask() {
  const text = newTaskText.value.trim();
  if (text) {
    tasks.value.push({
      id: Date.now() + Math.floor(Math.random() * 1000),
      text,
      completed: false,
      notes: '',
    });
    newTaskText.value = '';
  }
}

function onTaskKeyPress(e) {
  if (e.key === 'Enter') addTask();
}

function toggleTask(id) {
  tasks.value = tasks.value.map((t) => (t.id === id ? { ...t, completed: !t.completed } : t));
}

function removeTask(id) {
  tasks.value = tasks.value.filter((t) => t.id !== id);
}

function updateTaskNotes(id, notes) {
  tasks.value = tasks.value.map((t) => (t.id === id ? { ...t, notes } : t));
}

function nextPage() {
  if (taskPage.value < totalPages.value - 1) taskPage.value += 1;
}
function prevPage() {
  if (taskPage.value > 0) taskPage.value -= 1;
}
function nextSalePage() {
  if (salePage.value < totalSalesPages.value - 1) salePage.value += 1;
}
function prevSalePage() {
  if (salePage.value > 0) salePage.value -= 1;
}

function handleEmployeeChange(index, field, value) {
  const copy = [...dayEmployees.value];
  copy[index] = { ...copy[index], [field]: value };
  dayEmployees.value = copy;
}
function addEmployeeRow() {
  dayEmployees.value = [...dayEmployees.value, { id: null, position: '', fullName: '' }];
}
function removeEmployeeRow(index) {
  dayEmployees.value = dayEmployees.value.filter((_, i) => i !== index);
}

function goToDay(d) {
  router.push(`/day/${d}`);
}

function fmtMoney(n) {
  return n.toLocaleString('ru-RU') + ' ₽';
}

watch(
  [sales, dayEmployees, cashMorning, cashEvening, tasks],
  () => {
    if (saveTimeout.value) clearTimeout(saveTimeout.value);
    saveTimeout.value = setTimeout(saveDayData, 1000);
  },
  { deep: true }
);

onMounted(() => {
  loadData();
  loadEmployees();
});

onBeforeUnmount(() => {
  if (saveTimeout.value) clearTimeout(saveTimeout.value);
});

const uniquePositions = computed(() => [...new Set(allEmployees.value.map((e) => e.position).filter(Boolean))]);
</script>

<template>
  <div>
    <NavigationHeader :title="title" :show-back-button="true" back-url="/" />

    <div class="day-wrapper">
      <div class="main-section">
        <!-- Сотрудники -->
        <div class="card">
          <div class="card-header">Сотрудники</div>
          <div class="card-body">
            <div v-for="(emp, idx) in dayEmployees" :key="idx" class="employee-row">
              <input
                type="text"
                class="form-control"
                placeholder="Должность"
                :value="emp.position"
                @input="handleEmployeeChange(idx, 'position', $event.target.value)"
                list="positions-list"
              />
              <input
                type="text"
                class="form-control"
                placeholder="ФИО"
                :value="emp.fullName"
                @input="handleEmployeeChange(idx, 'fullName', $event.target.value)"
                list="employees-list"
              />
              <button class="btn btn-outline-danger btn-sm" @click="removeEmployeeRow(idx)">✕</button>
            </div>
            <button class="btn btn-outline-primary btn-sm mt-2" @click="addEmployeeRow">+ Добавить сотрудника</button>
            <datalist id="positions-list">
              <option v-for="pos in uniquePositions" :key="pos" :value="pos" />
            </datalist>
            <datalist id="employees-list">
              <option v-for="emp in allEmployees" :key="emp.id" :value="emp.fullName" />
            </datalist>
          </div>
        </div>

        <!-- Продажи -->
        <div class="card">
          <div class="card-header">🛍️ Продажи</div>
          <div class="card-body">
            <div class="sales-list">
              <div
                v-for="sale in currentSales"
                :key="sale.id"
                class="sale-item"
                :class="{ open: openAccordion === sale.id }"
              >
                <div class="sale-header" @click="openAccordion = openAccordion === sale.id ? null : sale.id">
                  <span class="sale-header-title">{{ sale.name || 'Новая продажа' }}</span>
                  <span class="sale-header-price">{{ sale.price ? fmtMoney(parseFloat(sale.price)) : '0 ₽' }}</span>
                </div>
                <div v-if="openAccordion === sale.id" class="sale-body">
                  <div class="form-group">
                    <label class="form-label">Наименование</label>
                    <input
                      type="text"
                      class="form-control"
                      :value="sale.name"
                      @input="updateSale(sale.id, 'name', $event.target.value)"
                    />
                  </div>
                  <div class="form-group">
                    <label class="form-label">Цена (₽)</label>
                    <input
                      type="number"
                      class="form-control"
                      :value="sale.price"
                      @input="updateSale(sale.id, 'price', $event.target.value)"
                    />
                  </div>
                  <div class="grid-3">
                    <div class="form-group">
                      <label class="form-label">SPH</label>
                      <input
                        type="text"
                        class="form-control"
                        :value="sale.sph"
                        @input="updateSale(sale.id, 'sph', $event.target.value)"
                        list="sph-options"
                      />
                    </div>
                    <div class="form-group">
                      <label class="form-label">CYL</label>
                      <input
                        type="text"
                        class="form-control"
                        :value="sale.cyl"
                        @input="updateSale(sale.id, 'cyl', $event.target.value)"
                        list="cyl-options"
                      />
                    </div>
                    <div class="form-group">
                      <label class="form-label">AX</label>
                      <input
                        type="text"
                        class="form-control"
                        :value="sale.ax"
                        @input="updateSale(sale.id, 'ax', $event.target.value)"
                        list="ax-options"
                      />
                    </div>
                  </div>
                  <div class="form-group">
                    <label class="form-label">ADD</label>
                    <input
                      type="text"
                      class="form-control"
                      :value="sale.add"
                      @input="updateSale(sale.id, 'add', $event.target.value)"
                    />
                  </div>
                  <div class="form-group">
                    <label class="form-label">PD</label>
                    <input
                      type="text"
                      class="form-control"
                      :value="sale.pd"
                      @input="updateSale(sale.id, 'pd', $event.target.value)"
                    />
                  </div>
                  <div class="form-group">
                    <label class="form-label">Материал</label>
                    <input
                      type="text"
                      class="form-control"
                      :value="sale.material"
                      @input="updateSale(sale.id, 'material', $event.target.value)"
                    />
                  </div>
                  <div class="form-group">
                    <label class="form-label">Покрытие</label>
                    <input
                      type="text"
                      class="form-control"
                      :value="sale.coating"
                      @input="updateSale(sale.id, 'coating', $event.target.value)"
                    />
                  </div>
                  <button class="btn btn-outline-danger btn-sm" @click="showDeleteModal(sale.id)">Удалить</button>
                </div>
              </div>
            </div>
            <button class="btn btn-primary btn-sm mt-2" @click="addSale">+ Добавить продажу</button>
            <div v-if="totalSalesPages > 1" class="pagination-row">
              <button class="btn btn-outline-secondary btn-sm" :disabled="salePage === 0" @click="prevSalePage">←</button>
              <span class="page-info">{{ salePage + 1 }} / {{ totalSalesPages }}</span>
              <button class="btn btn-outline-secondary btn-sm" :disabled="salePage === totalSalesPages - 1" @click="nextSalePage">→</button>
            </div>
          </div>
        </div>

        <!-- Итоги -->
        <div class="summary-row">
          <div class="summary-card">
            <h3>Наличные</h3>
            <div class="summary-value cash">{{ fmtMoney(totalCash) }}</div>
          </div>
          <div class="summary-card">
            <h3>Безналичные</h3>
            <div class="summary-value cashless">{{ fmtMoney(totalCashless) }}</div>
          </div>
          <div class="summary-card">
            <h3>Всего</h3>
            <div class="summary-value card-total">{{ fmtMoney(totalCard) }}</div>
          </div>
        </div>
      </div>

      <!-- Сайдбар -->
      <div class="sidebar-section">
        <div class="card">
          <div class="card-header">Касса</div>
          <div class="card-body">
            <div class="form-group">
              <label class="form-label">Касса утро:</label>
              <input
                type="number"
                class="form-control"
                :value="cashMorning"
                @input="cashMorning = $event.target.value"
                placeholder="0"
              />
            </div>
            <div class="form-group">
              <label class="form-label">Касса вечер:</label>
              <input
                type="number"
                class="form-control"
                :value="cashEvening"
                @input="cashEvening = $event.target.value"
                placeholder="0"
              />
            </div>
          </div>
        </div>

        <div class="card">
          <div class="card-header calendar-header">Календарь</div>
          <div class="card-body calendar-body">
            <div class="calendar-grid-header">
              <div v-for="d in weekDays" :key="d" class="calendar-weekday">{{ d }}</div>
            </div>
            <div class="calendar-grid">
              <template v-for="cell in calendarDays" :key="cell.key">
                <div v-if="cell.type === 'empty'" class="calendar-cell calendar-cell-empty"></div>
                <button
                  v-else
                  class="calendar-cell"
                  :class="{ 'calendar-cell-today': cell.isToday, 'calendar-cell-selected': cell.day === dayNum }"
                  @click="goToDay(cell.day)"
                >
                  {{ cell.day }}
                </button>
              </template>
            </div>
          </div>
        </div>

        <div class="card">
          <div class="card-header">Задачи</div>
          <div class="card-body tasks-body">
            <div class="task-add-row">
              <input
                type="text"
                class="form-control"
                placeholder="Новая задача"
                :value="newTaskText"
                @input="newTaskText = $event.target.value"
                @keypress="onTaskKeyPress"
              />
              <button class="btn btn-primary btn-sm" @click="addTask">+</button>
            </div>
            <div v-if="tasks.length === 0" class="text-muted small">Нет задач</div>
            <div v-else>
              <div
                v-for="task in currentTasks"
                :key="task.id"
                class="task-item"
                :class="{ 'task-completed': task.completed }"
              >
                <div class="task-item-header">
                  <label class="form-check">
                    <input
                      type="checkbox"
                      :checked="task.completed"
                      @change="toggleTask(task.id)"
                    />
                    <span>{{ task.text }}</span>
                  </label>
                  <button class="btn btn-outline-danger btn-sm" @click.stop="removeTask(task.id)">✕</button>
                </div>
                <div class="task-notes">
                  <textarea
                    class="form-control"
                    rows="2"
                    placeholder="Заметки (необязательно)"
                    :value="task.notes || ''"
                    @input="updateTaskNotes(task.id, $event.target.value)"
                  ></textarea>
                </div>
              </div>
              <div v-if="totalPages > 1" class="pagination-row">
                <button class="btn btn-outline-secondary btn-sm" :disabled="taskPage === 0" @click="prevPage">←</button>
                <span class="page-info">{{ taskPage + 1 }} / {{ totalPages }}</span>
                <button class="btn btn-outline-secondary btn-sm" :disabled="taskPage === totalPages - 1" @click="nextPage">→</button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <datalist id="sph-options">
      <option v-for="v in sphValues" :key="v" :value="v" />
    </datalist>
    <datalist id="cyl-options">
      <option v-for="v in cylValues" :key="v" :value="v" />
    </datalist>
    <datalist id="ax-options">
      <option v-for="v in axValues" :key="v" :value="v" />
    </datalist>

    <!-- Модальное окно подтверждения удаления -->
    <div v-if="showDeleteConfirm" class="modal-backdrop" @click="cancelDelete"></div>
    <div v-if="showDeleteConfirm" class="modal">
      <div class="modal-dialog">
        <div class="modal-header">
          <h3 class="modal-title">Подтвердите действие</h3>
          <button class="btn btn-light btn-sm" @click="cancelDelete">✕</button>
        </div>
        <div class="modal-body">Вы уверены что хотите удалить данный товар?</div>
        <div class="modal-footer">
          <button class="btn btn-secondary" @click="cancelDelete">Отмена</button>
          <button class="btn btn-danger" @click="confirmDelete">Удалить</button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.day-wrapper {
  display: grid;
  grid-template-columns: 1fr 360px;
  gap: 16px;
}
.main-section {
  display: flex;
  flex-direction: column;
  gap: 16px;
}
.sidebar-section {
  display: flex;
  flex-direction: column;
  gap: 16px;
}
.employee-row {
  display: grid;
  grid-template-columns: 1fr 1fr auto;
  gap: 8px;
  margin-bottom: 8px;
}
.sales-list {
  display: flex;
  flex-direction: column;
  gap: 6px;
}
.sale-item {
  border: 1px solid var(--color-border);
  border-radius: var(--radius-sm);
  background: #fafbfc;
  overflow: hidden;
}
.sale-item.open {
  background: #fff;
  box-shadow: var(--shadow-sm);
}
.sale-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 10px 14px;
  cursor: pointer;
  font-weight: 500;
}
.sale-header:hover {
  background: #f1f5f9;
}
.sale-header-title {
  color: var(--color-text);
}
.sale-header-price {
  color: var(--color-primary);
  font-weight: 600;
}
.sale-body {
  padding: 12px 14px;
  border-top: 1px solid var(--color-border);
}
.grid-3 {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 8px;
}
.pagination-row {
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 12px;
  margin-top: 12px;
}
.page-info {
  font-size: 12px;
  color: var(--color-muted);
  font-weight: 600;
}
.summary-row {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 12px;
}
.summary-card {
  background: #fff;
  border: 1px solid var(--color-border);
  border-radius: var(--radius);
  padding: 16px;
  text-align: center;
}
.summary-card h3 {
  font-size: 13px;
  color: var(--color-muted);
  margin: 0 0 8px;
  font-weight: 500;
}
.summary-value {
  font-size: 22px;
  font-weight: 700;
}
.cash { color: var(--color-cash); }
.cashless { color: var(--color-cashless); }
.card-total { color: var(--color-card); }

.calendar-header {
  background: var(--color-primary);
  color: #fff;
}
.calendar-body { padding: 12px; }
.calendar-grid-header {
  display: grid;
  grid-template-columns: repeat(7, 1fr);
  gap: 4px;
  margin-bottom: 8px;
}
.calendar-weekday {
  text-align: center;
  font-size: 11px;
  font-weight: 600;
  color: var(--color-muted);
  text-transform: uppercase;
}
.calendar-grid {
  display: grid;
  grid-template-columns: repeat(7, 1fr);
  gap: 4px;
}
.calendar-cell {
  aspect-ratio: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-sm);
  background: #fff;
  cursor: pointer;
  font-size: 13px;
  font-weight: 500;
  transition: all var(--transition);
  font-family: inherit;
}
.calendar-cell:hover:not(.calendar-cell-empty) {
  background: var(--color-primary);
  color: #fff;
  border-color: var(--color-primary);
}
.calendar-cell-empty {
  cursor: default;
  background: transparent;
  border-color: transparent;
}
.calendar-cell-today {
  background: var(--color-primary);
  color: #fff;
  border-color: var(--color-primary);
}
.calendar-cell-selected {
  background: var(--color-warning);
  color: #fff;
  border-color: var(--color-warning);
}

.tasks-body { padding: 12px; }
.task-add-row {
  display: grid;
  grid-template-columns: 1fr auto;
  gap: 6px;
  margin-bottom: 10px;
}
.task-item {
  border: 1px solid var(--color-border);
  border-radius: var(--radius-sm);
  margin-bottom: 6px;
  background: #fff;
}
.task-item-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 6px 10px;
  gap: 6px;
}
.task-completed span { text-decoration: line-through; color: var(--color-muted); }
.task-notes { padding: 0 10px 8px; }

@media (max-width: 1100px) {
  .day-wrapper { grid-template-columns: 1fr; }
  .summary-row { grid-template-columns: 1fr; }
}
</style>
