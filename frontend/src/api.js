// src/api.js
// Тонкая обёртка над сгенерированным wailsjs-биндингом. Все вызовы
// в одном месте — проще рефакторить, проще мокать в тестах.
import {
  GetPositions,
  CreatePosition,
  UpdatePosition,
  DeletePosition,
  GetStaff,
  CreateStaff,
  UpdateStaff,
  DeleteStaff,
  GetSchedule,
  GetScheduleMembers,
  AddScheduleMember,
  SaveScheduleShift,
  DeleteScheduleRecord,
  DeleteScheduleForUser,
  GetScheduleForDate,
  GetSalesByDate,
  CreateSale,
  UpdateSale,
  DeleteSale,
  GetSalesByDateRange,
} from '../wailsjs/go/main/App';

export const api = {
  positions: {
    list: () => GetPositions(),
    create: (name, normHours, hoursPerShift, salary, additionalPayments) =>
      CreatePosition(name, normHours ?? null, salary ?? null, hoursPerShift ?? null, additionalPayments ?? null),
    update: (positionId, positionName, normHours, hoursPerShift, salary, additionalPayments) =>
      UpdatePosition(positionId, positionName, normHours ?? null, salary ?? null, hoursPerShift ?? null, additionalPayments ?? null),
    remove: (positionId) => DeletePosition(positionId),
  },
  staff: {
    list: () => GetStaff(),
    create: (fullName, positionId) => CreateStaff(fullName, positionId),
    update: (staffId, newFullName, newPositionId) => UpdateStaff(staffId, newFullName, newPositionId),
    remove: (staffId) => DeleteStaff(staffId),
  },
  // Работа с графиком сотрудников. Дата — строка YYYY-MM-DD.
  // shift — код смены ('1' / 'к' / 'Я' / 'о' / ''). Пустая строка = выходной
  // (запись в БД удаляется).
  schedule: {
    list: (from, to) => GetSchedule(from ?? '', to ?? ''),
    members: (year, month) => GetScheduleMembers(year, month),
    addMember: (userId, year, month) => AddScheduleMember(userId, year, month),
    saveShift: (userId, date, shift, hours, isWorkingDay) => SaveScheduleShift(userId, date, shift, hours ?? 0, isWorkingDay ?? true),
    remove: (userId, date) => DeleteScheduleRecord(userId, date),
    clearForUser: (userId) => DeleteScheduleForUser(userId),
    getForDate: (date) => GetScheduleForDate(date),
  },
  // Работа с продажами.
  sales: {
    getByDate: (date) => GetSalesByDate(date),
    create: (dateTime, productName, recipe, totalAmount, advanceAmount, cashAmount, cardAmount, sbpAmount, comment) =>
      CreateSale(dateTime, productName, recipe ?? null, totalAmount, advanceAmount, cashAmount, cardAmount, sbpAmount, comment ?? null),
    update: (id, dateTime, productName, recipe, totalAmount, advanceAmount, cashAmount, cardAmount, sbpAmount, comment) =>
      UpdateSale(id, dateTime, productName, recipe ?? null, totalAmount, advanceAmount, cashAmount, cardAmount, sbpAmount, comment ?? null),
    remove: (id) => DeleteSale(id),
    getByDateRange: (from, to) => GetSalesByDateRange(from, to),
  },
};

export default api;
