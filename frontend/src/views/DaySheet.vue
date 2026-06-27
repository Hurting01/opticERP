<script setup>
// DaySheet — детализация конкретного дня: сотрудники, продажи, касса, задачи.
// Все данные хранятся в localStorage (как в Tauri-версии).
import { ref, computed, watch, onMounted, onBeforeUnmount } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { Modal } from 'bootstrap';
import NavigationHeader from '../components/NavigationHeader.vue';
import api from '../api';

const route = useRoute();
const router = useRouter();

const dayNum = computed(() => parseInt(route.params.day, 10));
const month = computed(() => new Date().getMonth() + 1);
const year = computed(() => new Date().getFullYear());

const dayEmployees = ref([]);
const cashMorning = ref('');
const cashEvening = ref('');
const sales = ref([]);
const tasks = ref([]);
const openAccordion = ref(null);
const deleteModalRef = ref(null);
let deleteModalInstance = null;
const saleModalRef = ref(null);
let saleModalInstance = null;
const saleToDelete = ref(null);

// Форма для новой продажи
const saleForm = ref({
  productName: '',
  sph: '',
  cyl: '',
  ax: '',
  totalAmount: 0,
  advanceAmount: 0,
  cashAmount: 0,
  cardAmount: 0,
  comment: '',
});
const isLoadingSales = ref(false);
const isSavingSale = ref(false);
const isEditMode = ref(false);
const editingSaleId = ref(null);
const newTaskText = ref('');
const taskPage = ref(0);
const salePage = ref(0);
const tasksPerPage = 4;
const salesPerPage = 9;

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
    cashMorning.value = data.cashMorning || '';
    cashEvening.value = data.cashEvening || '';
    // Продажи теперь загружаются из БД через loadSales()
    tasks.value = data.tasks || [];
  } else {
    cashMorning.value = '';
    cashEvening.value = '';
    tasks.value = [];
  }
}

// Загрузка продаж из БД
async function loadSales() {
  isLoadingSales.value = true;
  try {
    const dateStr = `${year.value}-${String(month.value).padStart(2, '0')}-${String(dayNum.value).padStart(2, '0')}`;
    const salesData = await api.sales.getByDate(dateStr);
    sales.value = Array.isArray(salesData) ? salesData.map(s => ({
      id: s.id,
      name: s.product_name,
      price: s.total_amount,
      recipe: s.recipe || '',
      advanceAmount: s.advance_amount,
      cashAmount: s.cash_amount,
      cardAmount: s.card_amount,
      sbpAmount: s.sbp_amount,
      comment: s.comment || '',
      datetime: s.datetime,
    })) : [];
  } catch (err) {
    console.error('Ошибка загрузки продаж:', err);
    sales.value = [];
  } finally {
    isLoadingSales.value = false;
  }
}

function saveDayData() {
  localStorage.setItem(
    storageKey(),
    JSON.stringify({
      cashMorning: cashMorning.value,
      cashEvening: cashEvening.value,
      // Не сохраняем продажи — они теперь в БД
      // Не сохраняем сотрудников — они теперь из графика
      tasks: tasks.value,
    })
  );
}

async function loadEmployees() {
  try {
    const dateStr = `${year.value}-${String(month.value).padStart(2, '0')}-${String(dayNum.value).padStart(2, '0')}`;
    const scheduledStaff = await api.schedule.getForDate(dateStr);
    
    if (Array.isArray(scheduledStaff) && scheduledStaff.length > 0) {
      dayEmployees.value = scheduledStaff.map(emp => ({
        id: emp.id,
        position: emp.position_name || '',
        fullName: emp.full_name || '',
      }));
    } else {
      dayEmployees.value = [];
    }
  } catch (err) {
    console.error('Ошибка загрузки сотрудников из графика:', err);
    dayEmployees.value = [];
  }
}

function openSaleModal() {
  isEditMode.value = false;
  editingSaleId.value = null;
  resetSaleForm();
  saleModalInstance?.show();
}

function closeSaleModal() {
  saleModalInstance?.hide();
  isEditMode.value = false;
  editingSaleId.value = null;
}

function resetSaleForm() {
  saleForm.value = {
    productName: '',
    sph: '',
    cyl: '',
    ax: '',
    totalAmount: 0,
    advanceAmount: 0,
    cashAmount: 0,
    cardAmount: 0,
    comment: '',
  };
}

async function saveSale() {
  if (!saleForm.value.productName.trim()) {
    alert('Укажите название товара');
    return;
  }

  isSavingSale.value = true;
  try {
    const dateStr = `${year.value}-${String(month.value).padStart(2, '0')}-${String(dayNum.value).padStart(2, '0')}`;
    const dateTime = `${dateStr} ${new Date().toTimeString().slice(0, 8)}`;
    
    // Формируем recipe из sph, cyl, ax
    const recipeParts = [];
    if (saleForm.value.sph) recipeParts.push(`Sph: ${saleForm.value.sph}`);
    if (saleForm.value.cyl) recipeParts.push(`Cyl: ${saleForm.value.cyl}`);
    if (saleForm.value.ax) recipeParts.push(`Ax: ${saleForm.value.ax}`);
    const recipe = recipeParts.length > 0 ? recipeParts.join(', ') : null;
    
    // Явно приводим все числовые поля к float64 для Go backend
    const totalAmount = parseFloat(saleForm.value.totalAmount) || 0;
    const advanceAmount = parseFloat(saleForm.value.advanceAmount) || 0;
    const cashAmount = parseFloat(saleForm.value.cashAmount) || 0;
    const cardAmount = parseFloat(saleForm.value.cardAmount) || 0;
    
    // Комментарий передаём как null если пустой
    const comment = saleForm.value.comment.trim() ? saleForm.value.comment.trim() : null;
    
    if (isEditMode.value && editingSaleId.value) {
      // Режим редактирования
      await api.sales.update(
        editingSaleId.value,
        dateTime,
        saleForm.value.productName,
        recipe,
        totalAmount,
        advanceAmount,
        cashAmount,
        cardAmount,
        0, // sbpAmount
        comment,
      );
    } else {
      // Режим создания
      await api.sales.create(
        dateTime,
        saleForm.value.productName,
        recipe,
        totalAmount,
        advanceAmount,
        cashAmount,
        cardAmount,
        0, // sbpAmount
        comment,
      );
    }
    
    closeSaleModal();
    await loadSales();
  } catch (e) {
    console.error('Ошибка сохранения продажи:', e);
    alert(`Не удалось сохранить продажу: ${e?.message || e}`);
  } finally {
    isSavingSale.value = false;
  }
}

function editSale(sale) {
  isEditMode.value = true;
  editingSaleId.value = sale.id;
  
  // Заполняем форму данными продажи
  saleForm.value.productName = sale.name || '';
  
  // Парсим recipe обратно в поля sph, cyl, ax
  if (sale.recipe) {
    const recipeStr = sale.recipe;
    const sphMatch = recipeStr.match(/Sph:\s*([^\,]+)/);
    const cylMatch = recipeStr.match(/Cyl:\s*([^\,]+)/);
    const axMatch = recipeStr.match(/Ax:\s*(.+)/);
    
    saleForm.value.sph = sphMatch ? sphMatch[1].trim() : '';
    saleForm.value.cyl = cylMatch ? cylMatch[1].trim() : '';
    saleForm.value.ax = axMatch ? axMatch[1].trim() : '';
  } else {
    saleForm.value.sph = '';
    saleForm.value.cyl = '';
    saleForm.value.ax = '';
  }
  
  saleForm.value.totalAmount = parseFloat(sale.price) || 0;
  saleForm.value.advanceAmount = parseFloat(sale.advanceAmount) || 0;
  saleForm.value.cashAmount = parseFloat(sale.cashAmount) || 0;
  saleForm.value.cardAmount = parseFloat(sale.cardAmount) || 0;
  saleForm.value.comment = sale.comment || '';
  
  saleModalInstance?.show();
}

function updateSale(id, field, value) {
  sales.value = sales.value.map((s) => (s.id === id ? { ...s, [field]: value } : s));
}

async function deleteSale(id) {
  try {
    await api.sales.remove(id);
    await loadSales();
    deleteModalInstance?.hide();
  } catch (e) {
    console.error('Ошибка удаления продажи:', e);
    alert(`Не удалось удалить продажу: ${e?.message || e}`);
  }
}

function confirmDelete() {
  if (saleToDelete.value != null) deleteSale(saleToDelete.value);
}

function showDeleteModal(id) {
  saleToDelete.value = id;
  deleteModalInstance?.show();
}

function cancelDelete() {
  deleteModalInstance?.hide();
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

function goToDay(d) {
  router.push(`/day/${d}`);
}

function fmtMoney(n) {
  return n.toLocaleString('ru-RU') + ' ₽';
}

watch(
  [cashMorning, cashEvening, tasks],
  () => {
    if (saveTimeout.value) clearTimeout(saveTimeout.value);
    saveTimeout.value = setTimeout(saveDayData, 1000);
  },
  { deep: true }
);

watch(dayNum, () => {
  loadData();
  loadEmployees();
  loadSales();
});

onMounted(() => {
  loadData();
  loadEmployees();
  loadSales();
  
  // Инициализируем Bootstrap Modal для подтверждения удаления
  if (deleteModalRef.value) {
    deleteModalInstance = new Modal(deleteModalRef.value);
    deleteModalRef.value.addEventListener('hidden.bs.modal', () => {
      saleToDelete.value = null;
    });
  }

  // Инициализируем Bootstrap Modal для добавления продажи
  if (saleModalRef.value) {
    saleModalInstance = new Modal(saleModalRef.value);
    saleModalRef.value.addEventListener('hidden.bs.modal', () => {
      resetSaleForm();
    });
  }
});

onBeforeUnmount(() => {
  if (saveTimeout.value) clearTimeout(saveTimeout.value);
  if (deleteModalInstance) {
    deleteModalInstance.dispose();
    deleteModalInstance = null;
  }
  if (saleModalInstance) {
    saleModalInstance.dispose();
    saleModalInstance = null;
  }
});
</script>

<template>
  <div>
    <NavigationHeader :title="title" />

    <div class="day-wrapper">
      <div class="main-section">
        <!-- Сотрудники -->
        <div class="card">
          <div class="card-header">Сотрудники в смене</div>
          <div class="card-body">
            <div v-if="dayEmployees.length === 0" class="text-muted">
              Нет сотрудников по графику на этот день
            </div>
            <div v-else class="employee-list">
              <div v-for="emp in dayEmployees" :key="emp.id" class="employee-item">
                <span class="employee-position">{{ emp.position || 'Без должности' }}</span>
                <span class="employee-name">{{ emp.fullName }}</span>
              </div>
            </div>
          </div>
        </div>

        <!-- Продажи -->
        <div class="card">
          <div class="card-header card-header-with-action">
            <span>🛍️ Продажи</span>
            <button class="btn btn-primary btn-sm" @click="openSaleModal">+ Добавить продажу</button>
          </div>
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
                  <div class="sale-header-right">
                    <span class="sale-header-price">{{ sale.price ? fmtMoney(parseFloat(sale.price)) : '0 ₽' }}</span>
                    <div class="sale-actions">
                      <button class="sale-action-btn" @click.stop="editSale(sale)" title="Редактировать">
                        <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                          <path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"></path>
                          <path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"></path>
                        </svg>
                      </button>
                      <button class="sale-action-btn sale-action-delete" @click.stop="showDeleteModal(sale.id)" title="Удалить">
                        <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                          <line x1="18" y1="6" x2="6" y2="18"></line>
                          <line x1="6" y1="6" x2="18" y2="18"></line>
                        </svg>
                      </button>
                    </div>
                  </div>
                </div>
                <div v-if="openAccordion === sale.id" class="sale-body">
                  <div class="sale-main-grid">
                    <div v-if="sale.recipe" class="sale-info-block sale-recipe-block">
                      <div class="sale-info-header">Рецепт</div>
                      <div class="sale-info-content sale-recipe-content">{{ sale.recipe }}</div>
                    </div>
                    <div v-else class="sale-info-block sale-recipe-block sale-recipe-empty">
                      <div class="sale-info-header">Рецепт</div>
                      <div class="sale-info-content sale-recipe-content">—</div>
                    </div>
                    <div class="sale-info-block sale-amount-block">
                      <div class="sale-info-header">Общая сумма</div>
                      <div class="sale-info-content sale-amount-value">{{ fmtMoney(sale.price || 0) }}</div>
                    </div>
                    <div class="sale-info-block sale-amount-block">
                      <div class="sale-info-header">Аванс</div>
                      <div class="sale-info-content sale-amount-value">{{ fmtMoney(sale.advanceAmount || 0) }}</div>
                    </div>
                    <div class="sale-info-block sale-amount-block">
                      <div class="sale-info-header">Наличные</div>
                      <div class="sale-info-content sale-amount-value">{{ fmtMoney(sale.cashAmount || 0) }}</div>
                    </div>
                    <div class="sale-info-block sale-amount-block">
                      <div class="sale-info-header">Безнал</div>
                      <div class="sale-info-content sale-amount-value">{{ fmtMoney((sale.cardAmount || 0) + (sale.sbpAmount || 0)) }}</div>
                    </div>
                  </div>
                  <div v-if="sale.comment" class="sale-info-block">
                    <div class="sale-info-header">Комментарий</div>
                    <div class="sale-info-content sale-comment-content">{{ sale.comment }}</div>
                  </div>
                </div>
              </div>
            </div>
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

    <!-- Bootstrap Modal: подтверждение удаления -->
    <div ref="deleteModalRef" class="modal fade" tabindex="-1" aria-labelledby="deleteModalLabel" aria-hidden="true">
      <div class="modal-dialog modal-dialog-centered">
        <div class="modal-content">
          <div class="modal-header">
            <h5 id="deleteModalLabel" class="modal-title">Подтвердите действие</h5>
            <button type="button" class="btn-close" data-bs-dismiss="modal" aria-label="Close"></button>
          </div>
          <div class="modal-body">Вы уверены что хотите удалить данный товар?</div>
          <div class="modal-footer">
            <button type="button" class="btn btn-secondary" data-bs-dismiss="modal">Отмена</button>
            <button type="button" class="btn btn-danger" @click="confirmDelete">Удалить</button>
          </div>
        </div>
      </div>
    </div>

    <!-- Bootstrap Modal: создание продажи -->
    <div ref="saleModalRef" class="modal fade" tabindex="-1" aria-labelledby="saleModalLabel" aria-hidden="true">
      <div class="modal-dialog modal-dialog-centered">
        <div class="modal-content">
          <div class="modal-header">
            <h5 id="saleModalLabel" class="modal-title">{{ isEditMode ? 'Редактировать продажу' : 'Добавить продажу' }}</h5>
            <button type="button" class="btn-close" data-bs-dismiss="modal" aria-label="Close"></button>
          </div>
          <div class="modal-body">
            <div class="form-group mb-3">
              <label class="form-label">Наименование</label>
              <input
                type="text"
                class="form-control"
                v-model="saleForm.productName"
                placeholder="Очки, линзы и т.д."
              />
            </div>
            <div class="row mb-3">
              <div class="col-md-4">
                <label class="form-label">Sph</label>
                <input
                  type="text"
                  class="form-control"
                  v-model="saleForm.sph"
                  placeholder="-"
                />
              </div>
              <div class="col-md-4">
                <label class="form-label">Cyl</label>
                <input
                  type="text"
                  class="form-control"
                  v-model="saleForm.cyl"
                  placeholder="-"
                />
              </div>
              <div class="col-md-4">
                <label class="form-label">Ax</label>
                <input
                  type="text"
                  class="form-control"
                  v-model="saleForm.ax"
                  placeholder="-"
                />
              </div>
            </div>
            <div class="form-group mb-3">
              <label class="form-label">Общая сумма (₽)</label>
              <input
                type="number"
                class="form-control"
                v-model.number="saleForm.totalAmount"
                placeholder="0"
              />
            </div>
            <div class="form-group mb-3">
              <label class="form-label">Аванс (₽)</label>
              <input
                type="number"
                class="form-control"
                v-model.number="saleForm.advanceAmount"
                placeholder="0"
              />
            </div>
            <div class="form-group mb-3">
              <label class="form-label">Наличный расчет (₽)</label>
              <input
                type="number"
                class="form-control"
                v-model.number="saleForm.cashAmount"
                placeholder="0"
              />
            </div>
            <div class="form-group mb-3">
              <label class="form-label">Оплата картой/СБП (₽)</label>
              <input
                type="number"
                class="form-control"
                v-model.number="saleForm.cardAmount"
                placeholder="0"
              />
            </div>
            <div class="form-group mb-3">
              <label class="form-label">Комментарий</label>
              <textarea
                class="form-control"
                v-model="saleForm.comment"
                placeholder="Дополнительная информация"
                rows="3"
              ></textarea>
            </div>
          </div>
          <div class="modal-footer">
            <button type="button" class="btn btn-secondary" data-bs-dismiss="modal">Отмена</button>
            <button type="button" class="btn btn-primary" @click="saveSale" :disabled="isSavingSale">
              {{ isSavingSale ? 'Сохранение...' : (isEditMode ? 'Обновить' : 'Сохранить') }}
            </button>
          </div>
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
.employee-list {
  display: flex;
  gap: 8px;
}
.employee-item {
  display: flex;
  align-items: stretch;
  gap: 0;
  flex: 1;
  border: 1px solid var(--color-border);
  border-radius: var(--radius-sm);
  overflow: hidden;
  background: #fff;
}
.employee-position {
  display: flex;
  align-items: center;
  padding: 8px 12px;
  background: #e2e8f0;
  font-size: 13px;
  font-weight: 600;
  color: #475569;
  white-space: nowrap;
  min-width: 120px;
}
.employee-name {
  display: flex;
  align-items: center;
  padding: 8px 12px;
  font-size: 14px;
  color: var(--color-text);
  font-weight: 500;
  flex: 1;
  background: #fff;
}
.sales-list {
  display: flex;
  flex-direction: column;
  gap: 6px;
  max-height: 60vh;
  overflow-y: auto;
  padding-right: 4px;
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
  padding: 8px 12px;
  cursor: pointer;
  font-weight: 500;
  font-size: 14px;
}
.sale-header:hover {
  background: #f1f5f9;
}
.sale-header-title {
  color: var(--color-text);
  font-size: 13px;
}
.sale-header-right {
  display: flex;
  align-items: center;
  gap: 6px;
}
.sale-header-price {
  color: var(--color-primary);
  font-weight: 600;
  font-size: 13px;
}
.sale-actions {
  display: flex;
  gap: 3px;
}
.sale-action-btn {
  background: none;
  border: none;
  padding: 3px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 3px;
  color: #64748b;
  transition: all 0.15s ease;
}
.sale-action-btn:hover {
  background: #e2e8f0;
  color: #334155;
}
.sale-action-btn.sale-action-delete:hover {
  background: #fee2e2;
  color: #dc2626;
}
.sale-body {
  padding: 10px 12px;
  border-top: 1px solid var(--color-border);
}
.sale-main-grid {
  display: grid;
  grid-template-columns: 4fr repeat(4, 1fr);
  grid-template-rows: 1fr;
  gap: 3px;
  margin-bottom: 4px;
  min-height: 34px;
}
.sale-recipe-block {
  margin-bottom: 0;
}
.sale-recipe-block.sale-recipe-empty {
  /* Заглушка для одинаковой высоты строки */
}
.sale-info-block {
  margin-bottom: 0;
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 3px;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  min-height: 34px;
  height: 100%;
}
.sale-main-grid > .sale-info-block {
  height: 100%;
}
.sale-info-header {
  background: #e2e8f0;
  padding: 2px 6px;
  text-align: left;
  font-size: 9px;
  font-weight: 600;
  color: var(--color-muted);
  text-transform: uppercase;
  letter-spacing: 0.3px;
  flex-shrink: 0;
  height: 12px;
  box-sizing: border-box;
  display: flex;
  align-items: center;
}
.sale-info-content {
  padding: 3px 6px;
  font-size: 11px;
  color: var(--color-text);
  line-height: 1.3;
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 22px;
  box-sizing: border-box;
}
.sale-comment-content {
  max-height: 60px;
  overflow-y: auto;
  justify-content: flex-start;
  align-items: flex-start;
}
.sale-comment-content::-webkit-scrollbar {
  width: 4px;
}
.sale-comment-content::-webkit-scrollbar-track {
  background: #f1f5f9;
  border-radius: 2px;
}
.sale-comment-content::-webkit-scrollbar-thumb {
  background: #cbd5e1;
  border-radius: 2px;
}
.sale-comment-content::-webkit-scrollbar-thumb:hover {
  background: #94a3b8;
}
.sale-amount-block {
  margin-bottom: 0;
  min-height: 34px;
}
.sale-recipe-content {
  text-align: center;
  font-size: 10px;
}
.sale-amount-value {
  text-align: center;
  font-weight: 600;
  font-size: 11px;
  padding: 3px 2px;
}

.grid-2 {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 8px;
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

.card-header-with-action {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

/* Стилизация скроллбара для списка продаж */
.sales-list::-webkit-scrollbar {
  width: 6px;
}
.sales-list::-webkit-scrollbar-track {
  background: #f1f5f9;
  border-radius: 3px;
}
.sales-list::-webkit-scrollbar-thumb {
  background: #cbd5e1;
  border-radius: 3px;
}
.sales-list::-webkit-scrollbar-thumb:hover {
  background: #94a3b8;
}

@media (max-width: 1100px) {
  .day-wrapper { grid-template-columns: 1fr; }
  .summary-row { grid-template-columns: 1fr; }
}
</style>
