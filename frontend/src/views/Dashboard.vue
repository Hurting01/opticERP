<script setup>
// Dashboard — главная страница с задачами на сегодня,
// невыполненными задачами из прошлых дней и календарём.
// Данные читаются из localStorage (как в Tauri-версии).
import { ref, computed, onMounted } from 'vue';
import { useRouter } from 'vue-router';
import NavigationHeader from '../components/NavigationHeader.vue';

const router = useRouter();

const currentMonth = new Date().getMonth() + 1;
const currentYear = new Date().getFullYear();
const today = new Date().getDate();

const todayTasks = ref([]);
const pastTasks = ref([]);

const monthNames = [
  'Январь', 'Февраль', 'Март', 'Апрель', 'Май', 'Июнь',
  'Июль', 'Август', 'Сентябрь', 'Октябрь', 'Ноябрь', 'Декабрь',
];
const monthNamesGenitive = [
  'января', 'февраля', 'марта', 'апреля', 'мая', 'июня',
  'июля', 'августа', 'сентября', 'октября', 'ноября', 'декабря',
];
const weekDays = ['Вс', 'Пн', 'Вт', 'Ср', 'Чт', 'Пт', 'Сб'];

const firstDayOfMonth = new Date(currentYear, currentMonth - 1, 1).getDay();
const daysInMonth = new Date(currentYear, currentMonth, 0).getDate();

const calendarDays = computed(() => {
  const cells = [];
  for (let i = 0; i < firstDayOfMonth; i++) {
    cells.push({ type: 'empty', key: `empty-${i}` });
  }
  for (let day = 1; day <= daysInMonth; day++) {
    const date = new Date(currentYear, currentMonth - 1, day);
    cells.push({
      type: 'day',
      day,
      weekDay: weekDays[date.getDay()],
      isToday: day === today && currentMonth === new Date().getMonth() + 1,
      key: `day-${day}`,
    });
  }
  return cells;
});

function getDayData(day) {
  const value = localStorage.getItem(`day:${currentYear}-${currentMonth}-${day}`);
  return value ? JSON.parse(value) : null;
}

function saveDayData(day, data) {
  localStorage.setItem(`day:${currentYear}-${currentMonth}-${day}`, JSON.stringify(data));
}

function loadAllTasks() {
  const todayList = [];
  const pastList = [];
  for (let day = 1; day <= daysInMonth; day++) {
    const data = getDayData(day);
    if (data && data.tasks && Array.isArray(data.tasks)) {
      const dayTasks = data.tasks.map((t) => ({ ...t, day }));
      if (day === today) {
        todayList.push(...dayTasks);
      } else if (day < today) {
        pastList.push(...dayTasks);
      }
    }
  }
  todayTasks.value = todayList;
  pastTasks.value = pastList.filter((t) => !t.completed);
}

function toggleTask(taskId, day) {
  if (day === today) {
    const newTasks = todayTasks.value.map((t) =>
      t.id === taskId ? { ...t, completed: !t.completed } : t
    );
    todayTasks.value = newTasks;
    saveDayData(day, { ...(getDayData(day) || {}), tasks: newTasks });
  } else {
    const dayData = getDayData(day);
    if (dayData && dayData.tasks) {
      const updatedTasks = dayData.tasks.map((t) =>
        t.id === taskId ? { ...t, completed: true } : t
      );
      saveDayData(day, { ...dayData, tasks: updatedTasks });
    }
    pastTasks.value = pastTasks.value.filter((t) => t.id !== taskId);
  }
}

function goToDay(day) {
  router.push(`/day/${day}`);
}

onMounted(loadAllTasks);
</script>

<template>
  <div>
    <NavigationHeader :title="`${monthNames[currentMonth - 1]} ${currentYear}`" />

    <div class="main-wrapper">
      <!-- Левая часть: задачи -->
      <div class="left-section">
        <div class="card">
          <div class="card-header">Задачи на сегодня</div>
          <div class="card-body">
            <div v-if="todayTasks.length === 0" class="text-muted text-center">
              Нет задач на сегодня
            </div>
            <div v-else>
              <div
                v-for="task in todayTasks"
                :key="`today-${task.id}`"
                class="task-item"
                :class="{ 'task-completed': task.completed }"
              >
                <label class="form-check">
                  <input
                    type="checkbox"
                    :checked="task.completed"
                    @change="toggleTask(task.id, today)"
                  />
                  <span>{{ task.text }}</span>
                </label>
              </div>
            </div>
          </div>
        </div>

        <div v-if="pastTasks.length > 0" class="card past-tasks-card">
          <div class="card-header">Невыполненные задачи из прошлых дней</div>
          <div class="card-body">
            <div
              v-for="task in pastTasks"
              :key="`past-${task.id}`"
              class="task-item"
              :class="{ 'task-completed': task.completed }"
            >
              <div class="task-with-date">
                <label class="form-check">
                  <input
                    type="checkbox"
                    :checked="task.completed"
                    @change="toggleTask(task.id, task.day)"
                  />
                  <span>{{ task.text }}</span>
                </label>
                <span class="task-date-badge">{{ task.day }} {{ monthNamesGenitive[currentMonth - 1] }}</span>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Правая часть: календарь -->
      <div class="right-section">
        <div class="card calendar-card">
          <div class="card-header calendar-header">Календарь</div>
          <div class="card-body calendar-body">
            <div class="calendar-grid-header">
              <div v-for="d in weekDays" :key="d" class="calendar-weekday">{{ d }}</div>
            </div>
            <div class="calendar-grid">
              <template v-for="cell in calendarDays" :key="cell.key">
                <div
                  v-if="cell.type === 'empty'"
                  class="calendar-cell calendar-cell-empty"
                ></div>
                <button
                  v-else
                  class="calendar-cell"
                  :class="{ 'calendar-cell-today': cell.isToday }"
                  @click="goToDay(cell.day)"
                >
                  {{ cell.day }}
                </button>
              </template>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.main-wrapper {
  display: grid;
  grid-template-columns: 1fr 360px;
  gap: 16px;
}
.left-section {
  display: flex;
  flex-direction: column;
  gap: 16px;
}
.right-section {
  display: flex;
  flex-direction: column;
}
.task-item {
  padding: 8px 12px;
  border-radius: var(--radius-sm);
  margin-bottom: 6px;
  transition: background var(--transition);
}
.task-item:hover {
  background: #f8fafc;
}
.task-completed span {
  text-decoration: line-through;
  color: var(--color-muted);
}
.task-with-date {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}
.task-date-badge {
  background: var(--color-warning);
  color: #fff;
  padding: 2px 8px;
  border-radius: 12px;
  font-size: 11px;
  font-weight: 600;
}
.past-tasks-card .card-header {
  background: #fef3c7;
  color: #92400e;
}
.calendar-header {
  background: var(--color-primary);
  color: #fff;
}
.calendar-body {
  padding: 12px;
}
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
@media (max-width: 1024px) {
  .main-wrapper { grid-template-columns: 1fr; }
}
</style>
