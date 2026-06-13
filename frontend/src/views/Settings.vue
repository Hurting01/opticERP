<script setup>
// Settings — управление должностями и сотрудниками.
// CRUD через Wails -> Go -> SQLite.
import { ref, onMounted } from 'vue';
import NavigationHeader from '../components/NavigationHeader.vue';
import api from '../api';

const activeTab = ref('positions');
const positions = ref([]);
const staff = ref([]);

const modalMode = ref(null); // null | 'position' | 'employee'
const newPositionName = ref('');
const newPositionData = ref({ norm_hours: null, hours_per_shift: null, salary: null, additional_payments: null });
const selectedPositionId = ref('');
const newFullName = ref('');

const editingPosition = ref(null);
const editingStaff = ref(null);

async function loadData() {
  try {
    positions.value = (await api.positions.list()) || [];
  } catch (err) {
    console.error('Ошибка загрузки должностей:', err);
  }
  try {
    staff.value = (await api.staff.list()) || [];
  } catch (err) {
    console.error('Ошибка загрузки персонала:', err);
  }
}

function openAddPosition() {
  modalMode.value = 'position';
  newPositionName.value = '';
  newPositionData.value = { norm_hours: null, hours_per_shift: null, salary: null, additional_payments: null };
}

function openAddEmployee() {
  modalMode.value = 'employee';
  selectedPositionId.value = positions.value.length > 0 ? positions.value[0].id : '';
  newFullName.value = '';
}

function closeModal() {
  modalMode.value = null;
}

async function handleSaveNewItem() {
  if (modalMode.value === 'position') {
    if (!newPositionName.value.trim()) return;
    try {
      const created = await api.positions.create(
        newPositionName.value,
        newPositionData.value.norm_hours,
        newPositionData.value.hours_per_shift,
        newPositionData.value.salary,
        newPositionData.value.additional_payments
      );
      if (created) positions.value = [...positions.value, created];
    } catch (err) {
      console.error('Ошибка создания должности:', err);
      alert('Ошибка создания должности: ' + err);
    }
  } else if (modalMode.value === 'employee') {
    if (!selectedPositionId.value || !newFullName.value.trim()) return;
    try {
      const created = await api.staff.create(newFullName.value, Number(selectedPositionId.value));
      if (created) {
        const pos = positions.value.find((p) => p.id === Number(selectedPositionId.value));
        staff.value = [...staff.value, { ...created, position_name: pos?.name || '' }];
      }
    } catch (err) {
      console.error('Ошибка создания сотрудника:', err);
      alert('Ошибка создания сотрудника: ' + err);
    }
  }
  closeModal();
}

async function removePosition(pos) {
  if (!confirm('Удалить должность?')) return;
  try {
    const ok = await api.positions.remove(pos.id);
    if (ok) positions.value = positions.value.filter((p) => p.id !== pos.id);
  } catch (err) {
    console.error('Ошибка удаления:', err);
    alert('Ошибка удаления: ' + err);
  }
}

async function removeStaff(emp) {
  if (!confirm('Удалить сотрудника?')) return;
  try {
    const ok = await api.staff.remove(emp.id);
    if (ok) staff.value = staff.value.filter((s) => s.id !== emp.id);
  } catch (err) {
    console.error('Ошибка удаления:', err);
    alert('Ошибка удаления: ' + err);
  }
}

function startEditPosition(pos) {
  editingPosition.value = { ...pos };
}

async function saveEditPosition() {
  if (!editingPosition.value || !editingPosition.value.name?.trim()) return;
  try {
    const updated = await api.positions.update(
      editingPosition.value.id,
      editingPosition.value.name,
      editingPosition.value.norm_hours,
      editingPosition.value.hours_per_shift,
      editingPosition.value.salary,
      editingPosition.value.additional_payments
    );
    positions.value = positions.value.map((p) => (p.id === updated.id ? updated : p));
    editingPosition.value = null;
  } catch (err) {
    console.error('Ошибка обновления должности:', err);
    alert('Ошибка обновления должности: ' + err);
  }
}

function cancelEditPosition() {
  editingPosition.value = null;
}

function startEditStaff(emp) {
  editingStaff.value = { ...emp };
  selectedPositionId.value = emp.position_id;
  newFullName.value = emp.full_name;
}

async function saveEditStaff() {
  if (!editingStaff.value || !selectedPositionId.value || !newFullName.value.trim()) return;
  try {
    const updated = await api.staff.update(editingStaff.value.id, newFullName.value, Number(selectedPositionId.value));
    const pos = positions.value.find((p) => p.id === Number(selectedPositionId.value));
    staff.value = staff.value.map((s) =>
      s.id === updated.id ? { ...updated, position_name: pos?.name || '' } : s
    );
    editingStaff.value = null;
    selectedPositionId.value = '';
    newFullName.value = '';
  } catch (err) {
    console.error('Ошибка обновления сотрудника:', err);
    alert('Ошибка обновления сотрудника: ' + err);
  }
}

function cancelEditStaff() {
  editingStaff.value = null;
  selectedPositionId.value = '';
  newFullName.value = '';
}

function getPositionName(id) {
  const p = positions.value.find((x) => x.id === id);
  return p ? p.name : '';
}

function isManager(pos) {
  return pos && pos.name && pos.name.toLowerCase().includes('управляющ');
}

function formatSalary(v) {
  return v != null ? Number(v).toLocaleString('ru-RU') : '—';
}

function onNumber(field, value) {
  if (!editingPosition.value) return;
  const val = value === '' || value === null ? null : Math.max(0, Number(value));
  editingPosition.value = { ...editingPosition.value, [field]: val };
}

function onNewNumber(field, value) {
  const val = value === '' || value === null ? null : Math.max(0, Number(value));
  newPositionData.value = { ...newPositionData.value, [field]: val };
}

onMounted(loadData);
</script>

<template>
  <div>
    <NavigationHeader title="Настройки" />

    <ul class="nav-tabs">
      <li class="nav-item">
        <button class="nav-link" :class="{ active: activeTab === 'positions' }" @click="activeTab = 'positions'">Должности</button>
      </li>
      <li class="nav-item">
        <button class="nav-link" :class="{ active: activeTab === 'staff' }" @click="activeTab = 'staff'">Персонал</button>
      </li>
    </ul>

    <!-- === Должности === -->
    <div v-if="activeTab === 'positions'" class="card">
      <div class="card-header card-header-row">
        <span>Таблица должностей</span>
        <button class="btn btn-outline-primary btn-sm" @click="openAddPosition">+ Добавить должность</button>
      </div>
      <div class="card-body">
        <div v-if="editingPosition" class="d-flex flex-col gap-2">
          <div class="form-group">
            <input
              v-model="editingPosition.name"
              type="text"
              class="form-control"
              placeholder="Название должности"
              style="max-width: 300px"
              autofocus
            />
          </div>
          <div class="form-group">
            <input
              type="number"
              min="0"
              class="form-control"
              placeholder="Норма часов"
              :value="editingPosition.norm_hours ?? ''"
              @input="onNumber('norm_hours', $event.target.value)"
            />
          </div>
          <div class="form-group">
            <input
              type="number"
              step="0.5"
              min="0"
              class="form-control"
              placeholder="Часов/смена"
              :value="editingPosition.hours_per_shift ?? ''"
              @input="onNumber('hours_per_shift', $event.target.value)"
            />
          </div>
          <div class="form-group">
            <input
              type="number"
              min="0"
              class="form-control"
              placeholder="Зарплата (₽)"
              :value="editingPosition.salary ?? ''"
              @input="onNumber('salary', $event.target.value)"
            />
          </div>
          <div class="form-group">
            <input
              type="number"
              min="0"
              class="form-control"
              placeholder="Дополнительные выплаты (₽)"
              :value="editingPosition.additional_payments ?? ''"
              @input="onNumber('additional_payments', $event.target.value)"
            />
          </div>
          <div class="d-flex gap-2">
            <button class="btn btn-success btn-sm" @click="saveEditPosition">Сохранить</button>
            <button class="btn btn-secondary btn-sm" @click="cancelEditPosition">Отмена</button>
          </div>
        </div>
        <table v-else class="table table-striped table-bordered table-hover table-sm">
          <thead>
            <tr>
              <th>Должность</th>
              <th>Норма ч.</th>
              <th>Часов/смена</th>
              <th>Зарплата (₽)</th>
              <th>Доп. выплаты (₽)</th>
              <th style="width: 120px">Действия</th>
            </tr>
          </thead>
          <tbody>
            <tr v-if="positions.length === 0">
              <td colspan="6" class="text-center text-muted">Нет должностей</td>
            </tr>
            <tr v-for="pos in positions" :key="`pos-${pos.id}`">
              <td><strong>{{ pos.name }}</strong></td>
              <td>{{ pos.norm_hours ?? '—' }}</td>
              <td>{{ pos.hours_per_shift ?? '—' }}</td>
              <td>{{ formatSalary((pos.salary || 0) + (isManager(pos) ? (pos.additional_payments || 0) : 0)) }}</td>
              <td>{{ pos.additional_payments != null && pos.additional_payments > 0 ? formatSalary(pos.additional_payments) : '—' }}</td>
              <td>
                <div class="d-flex gap-1">
                  <button class="btn btn-outline-primary btn-sm" @click="startEditPosition(pos)">✎</button>
                  <button class="btn btn-outline-danger btn-sm" @click="removePosition(pos)">✕</button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- === Персонал === -->
    <div v-if="activeTab === 'staff'" class="card">
      <div class="card-header card-header-row">
        <span>Персонал</span>
        <button class="btn btn-outline-success btn-sm" @click="openAddEmployee">+ Добавить сотрудника</button>
      </div>
      <div class="card-body">
        <table class="table table-striped table-bordered table-hover table-sm">
          <thead>
            <tr>
              <th>Должность</th>
              <th>ФИО</th>
              <th style="width: 120px">Действия</th>
            </tr>
          </thead>
          <tbody>
            <tr v-if="editingStaff">
              <td>
                <select class="form-select form-select-sm" v-model.number="selectedPositionId" style="max-width: 200px">
                  <option v-for="pos in positions" :key="pos.id" :value="pos.id">{{ pos.name }}</option>
                </select>
              </td>
              <td>
                <input
                  v-model="newFullName"
                  type="text"
                  class="form-control form-control-sm"
                  placeholder="ФИО"
                  style="max-width: 250px"
                />
              </td>
              <td>
                <div class="d-flex gap-1">
                  <button class="btn btn-success btn-sm" @click="saveEditStaff">Сохранить</button>
                  <button class="btn btn-secondary btn-sm" @click="cancelEditStaff">Отмена</button>
                </div>
              </td>
            </tr>
            <tr v-else-if="staff.length === 0">
              <td colspan="3" class="text-center text-muted">Нет сотрудников</td>
            </tr>
            <tr v-for="emp in staff" v-else :key="`staff-${emp.id}`">
              <td>{{ getPositionName(emp.position_id) }}</td>
              <td>{{ emp.full_name }}</td>
              <td>
                <div class="d-flex gap-1">
                  <button class="btn btn-outline-primary btn-sm" @click="startEditStaff(emp)">✎</button>
                  <button class="btn btn-outline-danger btn-sm" @click="removeStaff(emp)">✕</button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- Модалка -->
    <div v-if="modalMode" class="modal-backdrop" @click="closeModal"></div>
    <div v-if="modalMode" class="modal">
      <div class="modal-dialog">
        <div class="modal-header">
          <h3 class="modal-title">{{ modalMode === 'position' ? 'Добавить должность' : 'Добавить сотрудника' }}</h3>
          <button class="btn btn-light btn-sm" @click="closeModal">✕</button>
        </div>
        <div class="modal-body d-flex flex-col gap-3">
          <div class="form-group">
            <label class="form-label">{{ modalMode === 'employee' ? 'Должность' : 'Название должности' }}</label>
            <select
              v-if="modalMode === 'employee'"
              v-model.number="selectedPositionId"
              class="form-select"
              autofocus
            >
              <option value="">Выберите должность</option>
              <option v-for="pos in positions" :key="pos.id" :value="pos.id">{{ pos.name }}</option>
            </select>
            <input
              v-else
              v-model="newPositionName"
              type="text"
              class="form-control"
              placeholder="Введите должность"
              autofocus
            />
          </div>
          <template v-if="modalMode === 'position'">
            <div class="form-group">
              <label class="form-label">Норма часов</label>
              <input
                type="number"
                min="0"
                class="form-control"
                :value="newPositionData.norm_hours ?? ''"
                @input="onNewNumber('norm_hours', $event.target.value)"
                placeholder="Введите значение"
              />
            </div>
            <div class="form-group">
              <label class="form-label">Часов/смена</label>
              <input
                type="number"
                step="0.5"
                min="0"
                class="form-control"
                :value="newPositionData.hours_per_shift ?? ''"
                @input="onNewNumber('hours_per_shift', $event.target.value)"
                placeholder="Введите значение"
              />
            </div>
            <div class="form-group">
              <label class="form-label">Зарплата (₽)</label>
              <input
                type="number"
                min="0"
                class="form-control"
                :value="newPositionData.salary ?? ''"
                @input="onNewNumber('salary', $event.target.value)"
                placeholder="Введите значение"
              />
            </div>
            <div class="form-group">
              <label class="form-label">Дополнительные выплаты (₽)</label>
              <input
                type="number"
                min="0"
                class="form-control"
                :value="newPositionData.additional_payments ?? ''"
                @input="onNewNumber('additional_payments', $event.target.value)"
                placeholder="Введите значение"
              />
            </div>
          </template>
          <div v-if="modalMode === 'employee'" class="form-group">
            <label class="form-label">ФИО</label>
            <input
              v-model="newFullName"
              type="text"
              class="form-control"
              placeholder="Введите ФИО"
            />
          </div>
          <div class="d-flex gap-2">
            <button class="btn btn-primary" @click="handleSaveNewItem">Сохранить</button>
            <button class="btn btn-secondary" @click="closeModal">Отмена</button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.d-flex-col {
  display: flex;
  flex-direction: column;
}
.gap-1 { gap: 4px; }
.gap-2 { gap: 8px; }
.gap-3 { gap: 12px; }
.d-flex {
  display: flex;
}
.card-header-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
</style>
