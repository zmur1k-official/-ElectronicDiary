const state = {
  token: localStorage.getItem("token") || "",
  user: null,
};

// Состояние табличного журнала учителя.
const teacherJournal = {
  subject: "",
  students: [],
  dates: [],
  gradesMap: new Map(),
};

const logBox = document.getElementById("log");
const sessionUser = document.getElementById("sessionUser");
const logoutBtn = document.getElementById("logoutBtn");
const authSection = document.getElementById("authSection");
const dashboard = document.getElementById("dashboard");

const registerRole = document.getElementById("registerRole");
const classLabel = document.getElementById("classLabel");

registerRole.addEventListener("change", () => {
  classLabel.classList.toggle("hidden", registerRole.value !== "student");
});

function log(msg, data) {
  const line = `[${new Date().toLocaleTimeString()}] ${msg}`;
  logBox.textContent = data ? `${line}\n${JSON.stringify(data, null, 2)}` : line;
}

async function api(path, options = {}) {
  const headers = { "Content-Type": "application/json", ...(options.headers || {}) };
  if (state.token) headers.Authorization = `Bearer ${state.token}`;
  const res = await fetch(path, { ...options, headers });
  const data = await res.json().catch(() => ({}));
  if (!res.ok) throw new Error(data.error || `HTTP ${res.status}`);
  return data;
}

function escapeHtml(value) {
  return String(value || "")
    .replaceAll("&", "&amp;")
    .replaceAll("<", "&lt;")
    .replaceAll(">", "&gt;")
    .replaceAll('"', "&quot;")
    .replaceAll("'", "&#039;");
}

function formField(name, type = "text", placeholder = "") {
  return `<label>${name}<input name="${name}" type="${type}" placeholder="${placeholder}" required /></label>`;
}

function card(title, body) {
  return `<div class="panel"><h3>${title}</h3>${body}</div>`;
}

function dateToISO(d) {
  const y = d.getFullYear();
  const m = String(d.getMonth() + 1).padStart(2, "0");
  const day = String(d.getDate()).padStart(2, "0");
  return `${y}-${m}-${day}`;
}

// Формирует окно дат: от -7 до +7 дней относительно текущей даты.
function buildDateWindow() {
  const base = new Date();
  base.setHours(0, 0, 0, 0);
  const res = [];
  for (let i = -7; i <= 7; i++) {
    const d = new Date(base);
    d.setDate(base.getDate() + i);
    res.push(dateToISO(d));
  }
  return res;
}

// Строит HTML-таблицу оценок ученика: строки=предметы, столбцы=даты.
function buildStudentGradesTable(rows) {
  const subjects = [...new Set((rows || []).map((r) => (r.subject || "").trim()).filter(Boolean))]
    .sort((a, b) => a.localeCompare(b, "ru"));
  const dates = [...new Set((rows || []).map((r) => (r.date || "").trim()).filter(Boolean))]
    .sort((a, b) => a.localeCompare(b));

  const map = new Map();
  for (const r of rows || []) {
    const subject = (r.subject || "").trim();
    const date = (r.date || "").trim();
    if (!subject || !date) continue;
    const key = `${subject}__${date}`;
    const prev = map.get(key);
    if (!prev || (r.id || 0) > (prev.id || 0)) map.set(key, r);
  }

  if (!subjects.length || !dates.length) return `<div class="item">Оценок пока нет.</div>`;

  const head = dates.map((d) => `<th>${escapeHtml(d)}</th>`).join("");
  const body = subjects
    .map((subject) => {
      const tds = dates
        .map((date) => {
          const row = map.get(`${subject}__${date}`);
          if (!row) return `<td class="empty-cell">-</td>`;
          const title = escapeHtml(row.comment || "Без комментария");
          return `<td><span class="grade-badge" title="${title}">${escapeHtml(row.value)}</span></td>`;
        })
        .join("");
      return `<tr><th>${escapeHtml(subject)}</th>${tds}</tr>`;
    })
    .join("");

  return `<div class="table-wrap"><table class="grade-table"><thead><tr><th>Предмет \\ Дата</th>${head}</tr></thead><tbody>${body}</tbody></table></div>`;
}

// Рисует журнал учителя: строки=ученики, столбцы=даты, ячейка кликабельна.
function renderTeacherJournalTable() {
  const box = document.getElementById("teacherJournalWrap");
  if (!box) return;

  if (!teacherJournal.students.length) {
    box.innerHTML = `<div class="item">Список учеников пуст.</div>`;
    return;
  }

  const head = teacherJournal.dates.map((d) => `<th>${escapeHtml(d)}</th>`).join("");
  const body = teacherJournal.students
    .map((student) => {
      const firstCol = `<th>${escapeHtml(student.className || "-")} | ${escapeHtml(student.fullName)}</th>`;
      const tds = teacherJournal.dates
        .map((date) => {
          const key = `${student.id}__${date}`;
          const row = teacherJournal.gradesMap.get(key);
          const value = row ? escapeHtml(row.value) : "+";
          const title = escapeHtml((row && row.comment) || "Кликните, чтобы поставить оценку");
          const cls = row ? "grade-cell has-grade" : "grade-cell empty-grade";
          return `<td><button type="button" class="${cls}" data-student-id="${student.id}" data-date="${date}" title="${title}">${value}</button></td>`;
        })
        .join("");
      return `<tr>${firstCol}${tds}</tr>`;
    })
    .join("");

  box.innerHTML = `<div class="table-wrap"><table class="grade-table"><thead><tr><th>Ученик</th>${head}</tr></thead><tbody>${body}</tbody></table></div>`;
}

// Загружает оценки учителя за диапазон дат и обновляет карту оценок.
async function loadTeacherJournalData() {
  const from = teacherJournal.dates[0];
  const to = teacherJournal.dates[teacherJournal.dates.length - 1];
  const payload = await api(`/api/teacher/grades/journal?from=${encodeURIComponent(from)}&to=${encodeURIComponent(to)}`);
  teacherJournal.subject = payload.subject || "";
  document.getElementById("teacherSubjectInput").value = teacherJournal.subject;
  document.getElementById("teacherSubjectCurrent").textContent = teacherJournal.subject || "не задан";

  teacherJournal.gradesMap = new Map();
  for (const g of payload.grades || []) {
    const key = `${g.studentId}__${g.date}`;
    const prev = teacherJournal.gradesMap.get(key);
    if (!prev || (g.id || 0) > (prev.id || 0)) teacherJournal.gradesMap.set(key, g);
  }
  renderTeacherJournalTable();
}

// Навешивает действия для предмета и кликов по ячейкам журнала.
function setupTeacherJournalActions() {
  const saveSubjectBtn = document.getElementById("saveTeacherSubjectBtn");
  const loadJournalBtn = document.getElementById("loadTeacherJournalBtn");
  const tableWrap = document.getElementById("teacherJournalWrap");
  const subjectInput = document.getElementById("teacherSubjectInput");

  saveSubjectBtn.onclick = async () => {
    const subject = (subjectInput.value || "").trim();
    if (!subject) {
      log("Предмет", { error: "Введите предмет" });
      return;
    }
    try {
      const data = await api("/api/teacher/subject", {
        method: "POST",
        body: JSON.stringify({ subject }),
      });
      teacherJournal.subject = data.subject || "";
      document.getElementById("teacherSubjectCurrent").textContent = teacherJournal.subject || "не задан";
      log("Предмет учителя сохранен", data);
    } catch (e) {
      log("Ошибка сохранения предмета", { error: e.message });
    }
  };

  loadJournalBtn.onclick = async () => {
    try {
      await loadTeacherJournalData();
    } catch (e) {
      log("Ошибка загрузки журнала", { error: e.message });
    }
  };

  tableWrap.onclick = async (e) => {
    const btn = e.target.closest("button.grade-cell");
    if (!btn) return;
    if (!teacherJournal.subject) {
      log("Журнал", { error: "Сначала задайте предмет учителя" });
      return;
    }

    const studentId = Number(btn.dataset.studentId);
    const date = btn.dataset.date || "";
    if (!studentId || !date) return;

    const existing = teacherJournal.gradesMap.get(`${studentId}__${date}`);
    const valueRaw = prompt(`Оценка на ${date} (1-5):`, existing ? String(existing.value) : "5");
    if (valueRaw === null) return;
    const value = Number(valueRaw);
    if (!Number.isInteger(value) || value < 1 || value > 5) {
      log("Ошибка", { error: "Оценка должна быть от 1 до 5" });
      return;
    }
    const comment = prompt("Комментарий:", existing ? (existing.comment || "") : "");
    if (comment === null) return;

    try {
      await api("/api/teacher/grades", {
        method: "POST",
        body: JSON.stringify({
          studentId,
          value,
          comment,
          date,
        }),
      });
      await loadTeacherJournalData();
      log("Оценка сохранена", { studentId, date, value, subject: teacherJournal.subject });
    } catch (err) {
      log("Ошибка сохранения оценки", { error: err.message });
    }
  };
}

// Загружает учеников (в порядке сортировки с бэкенда по классам).
async function loadTeacherStudents(selectId = "") {
  try {
    const students = await api("/api/teacher/students");
    teacherJournal.students = students || [];
    const select = selectId ? document.getElementById(selectId) : null;
    if (select) {
      const options = ['<option value="">Выберите ученика</option>'];
      for (const student of teacherJournal.students) {
        const className = student.className || "-";
        options.push(`<option value="${student.id}">${className} | ${student.fullName} (#${student.id})</option>`);
      }
      select.innerHTML = options.join("");
    }
    renderTeacherJournalTable();
  } catch (err) {
    log("Ошибка загрузки учеников", { error: err.message });
  }
}

function submitForm(path, mapper = (x) => x) {
  return async (e) => {
    e.preventDefault();
    const form = e.target;
    const fd = new FormData(form);
    const body = Object.fromEntries(fd.entries());
    try {
      const result = await api(path, { method: "POST", body: JSON.stringify(mapper(body)) });
      form.reset();
      log("Успешно сохранено", result);
    } catch (err) {
      log("Ошибка", { error: err.message });
    }
  };
}

function renderDashboard() {
  if (state.user.role === "admin") {
    dashboard.innerHTML = [
      card("Пользователи", `
        <button id="loadUsers">Обновить список</button>
        <div id="usersList" class="list"></div>
      `),
      card("Расписание (фото по классу)", `
        <form id="scheduleImportForm" class="grid">
          ${formField("className", "text", "например, 7A")}
          <label>Фото расписания<input type="file" name="file" accept="image/*" required /></label>
          <button type="submit">Сохранить фото расписания</button>
        </form>
        <button id="clearScheduleBtn" type="button">Удалить все фото расписания</button>
      `),
    ].join("");

    document.getElementById("loadUsers").onclick = async () => {
      try {
        const users = await api("/api/admin/users");
        document.getElementById("usersList").innerHTML = users
          .map((u) => `<div class="item">#${u.id} ${u.fullName} | ${u.email} | ${u.role} ${u.className ? `| ${u.className}` : ""}</div>`)
          .join("");
      } catch (e) {
        log("Ошибка загрузки пользователей", { error: e.message });
      }
    };

    document.getElementById("scheduleImportForm").onsubmit = async (e) => {
      e.preventDefault();
      const form = e.target;
      const fileInput = form.querySelector("input[name='file']");
      const file = fileInput.files[0];
      const className = (form.querySelector("input[name='className']").value || "").trim();
      if (!file) return log("Ошибка импорта", { error: "Выберите фото файла" });
      if (!className) return log("Ошибка импорта", { error: "Укажите класс" });
      const fd = new FormData();
      fd.append("className", className);
      fd.append("file", file);
      try {
        const res = await fetch("/api/admin/schedule/import", {
          method: "POST",
          headers: { Authorization: `Bearer ${state.token}` },
          body: fd,
        });
        const data = await res.json().catch(() => ({}));
        if (!res.ok) throw new Error(data.error || `HTTP ${res.status}`);
        form.reset();
        log("Фото расписания сохранено", data);
      } catch (err) {
        log("Ошибка импорта", { error: err.message });
      }
    };

    document.getElementById("clearScheduleBtn").onclick = async () => {
      try {
        const data = await api("/api/admin/schedule", { method: "DELETE" });
        log("Фото расписания удалено", data);
      } catch (err) {
        log("Ошибка удаления расписания", { error: err.message });
      }
    };
    return;
  }

  if (state.user.role === "teacher") {
    teacherJournal.dates = buildDateWindow();
    teacherJournal.gradesMap = new Map();
    teacherJournal.students = [];
    teacherJournal.subject = "";

    dashboard.innerHTML = [
      card("Добавить урок в расписание", `
        <form id="scheduleForm" class="grid">
          ${formField("className")}
          ${formField("subject")}
          ${formField("weekday")}
          ${formField("startTime", "time")}
          ${formField("endTime", "time")}
          ${formField("room")}
          <button type="submit">Сохранить урок</button>
        </form>
      `),
      card("Журнал оценок учителя", `
        <div class="grid">
          <label>Предмет учителя<input id="teacherSubjectInput" placeholder="например, Математика" /></label>
          <button id="saveTeacherSubjectBtn" type="button">Сохранить предмет</button>
          <div class="item">Текущий предмет: <b id="teacherSubjectCurrent">не задан</b></div>
          <button id="loadTeacherJournalBtn" type="button">Загрузить журнал</button>
        </div>
        <div id="teacherJournalWrap"></div>
        <div class="hint">Столбцы дат: неделя назад + сегодня + неделя вперед. Клик по ячейке выставляет оценку и комментарий.</div>
      `),
      card("Выдать домашнее задание", `
        <form id="homeworkForm" class="grid">
          ${formField("className")}
          ${formField("subject")}
          <label>description<textarea name="description" required></textarea></label>
          ${formField("dueDate", "date")}
          <button type="submit">Сохранить ДЗ</button>
        </form>
      `),
    ].join("");

    document.getElementById("scheduleForm").onsubmit = submitForm("/api/teacher/schedule");
    document.getElementById("homeworkForm").onsubmit = submitForm("/api/teacher/homework");
    loadTeacherStudents();
    setupTeacherJournalActions();

    api("/api/teacher/subject")
      .then((d) => {
        teacherJournal.subject = d.subject || "";
        document.getElementById("teacherSubjectInput").value = teacherJournal.subject;
        document.getElementById("teacherSubjectCurrent").textContent = teacherJournal.subject || "не задан";
      })
      .catch(() => {});

    return;
  }

  dashboard.innerHTML = [
    card("Моё расписание", `<button id="loadSchedule">Загрузить</button><div id="scheduleList" class="list"></div>`),
    card("Мои оценки", `<button id="loadGrades">Загрузить</button><div id="gradesList"></div>`),
    card("Моя домашка", `<button id="loadHomework">Загрузить</button><div id="homeworkList" class="list"></div>`),
  ].join("");

  document.getElementById("loadSchedule").onclick = async () => {
    try {
      const data = await api("/api/student/schedule");
      if (!data.imageData) {
        document.getElementById("scheduleList").innerHTML = `<div class="item">Расписание пока не найдено для вашего класса.</div>`;
        return;
      }
      document.getElementById("scheduleList").innerHTML = `
        <div class="item">Класс: ${data.className || "-"}</div>
        <div class="item"><img src="${data.imageData}" alt="Расписание класса" style="max-width:100%;height:auto;border-radius:8px;" /></div>
      `;
    } catch (e) {
      log("Ошибка расписания", { error: e.message });
    }
  };

  document.getElementById("loadGrades").onclick = async () => {
    try {
      const rows = await api("/api/student/grades");
      document.getElementById("gradesList").innerHTML = buildStudentGradesTable(rows);
    } catch (e) {
      log("Ошибка оценок", { error: e.message });
    }
  };

  document.getElementById("loadHomework").onclick = async () => {
    try {
      const rows = await api("/api/student/homework");
      document.getElementById("homeworkList").innerHTML = rows
        .map((r) => `<div class="item">${r.subject}: ${r.description} (до ${r.dueDate})</div>`)
        .join("");
    } catch (e) {
      log("Ошибка домашки", { error: e.message });
    }
  };
}

function renderSession() {
  if (!state.user) {
    sessionUser.textContent = "Не авторизован";
    logoutBtn.classList.add("hidden");
    authSection.classList.remove("hidden");
    dashboard.classList.add("hidden");
    return;
  }
  sessionUser.textContent = `${state.user.fullName} (${state.user.role})`;
  logoutBtn.classList.remove("hidden");
  authSection.classList.add("hidden");
  dashboard.classList.remove("hidden");
  renderDashboard();
}

document.getElementById("loginForm").onsubmit = async (e) => {
  e.preventDefault();
  const form = e.target;
  const fd = new FormData(form);
  const body = Object.fromEntries(fd.entries());
  try {
    const result = await api("/api/login", { method: "POST", body: JSON.stringify(body) });
    state.token = result.token;
    state.user = result.user;
    localStorage.setItem("token", state.token);
    renderSession();
    log("Вход выполнен", state.user);
  } catch (err) {
    log("Ошибка входа", { error: err.message });
  }
};

document.getElementById("registerForm").onsubmit = async (e) => {
  e.preventDefault();
  const form = e.target;
  const fd = new FormData(form);
  const body = Object.fromEntries(fd.entries());
  if (body.role !== "student") body.className = "";
  try {
    const result = await api("/api/register", { method: "POST", body: JSON.stringify(body) });
    form.reset();
    registerRole.dispatchEvent(new Event("change"));
    log("Регистрация успешна", result);
  } catch (err) {
    log("Ошибка регистрации", { error: err.message });
  }
};

logoutBtn.onclick = () => {
  state.token = "";
  state.user = null;
  localStorage.removeItem("token");
  renderSession();
};

async function bootstrap() {
  if (!state.token) {
    renderSession();
    return;
  }
  try {
    state.user = await api("/api/me");
  } catch {
    state.token = "";
    localStorage.removeItem("token");
  }
  renderSession();
}

bootstrap();
registerRole.dispatchEvent(new Event("change"));
