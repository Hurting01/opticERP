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
} from '../wailsjs/go/main/App';

export const api = {
  positions: {
    list: () => GetPositions(),
    create: (name, normHours, hoursPerShift, salary, additionalPayments) =>
      CreatePosition(name, normHours ?? null, hoursPerShift ?? null, salary ?? null, additionalPayments ?? null),
    update: (positionId, positionName, normHours, hoursPerShift, salary, additionalPayments) =>
      UpdatePosition(positionId, positionName, normHours ?? null, hoursPerShift ?? null, salary ?? null, additionalPayments ?? null),
    remove: (positionId) => DeletePosition(positionId),
  },
  staff: {
    list: () => GetStaff(),
    create: (fullName, positionId) => CreateStaff(fullName, positionId),
    update: (staffId, newFullName, newPositionId) => UpdateStaff(staffId, newFullName, newPositionId),
    remove: (staffId) => DeleteStaff(staffId),
  },
};

export default api;
