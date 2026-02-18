const state = {
  token: localStorage.getItem("token") || "",
  user: null,
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

function card(title, body) {
  return `<div class="panel"><h3>${title}</h3>${body}</div>`;
}

function formField(name, type = "text", placeholder = "") {
  return `<label>${name}<input name="${name}" type="${type}" placeholder="${placeholder}" required /></label>`;
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
      if (!file) {
        log("Ошибка импорта", { error: "Выберите фото файла" });
        return;
      }
      if (!className) {
        log("Ошибка импорта", { error: "Укажите класс" });
        return;
      }
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
      card("Поставить оценку", `
        <form id="gradeForm" class="grid">
          <label>studentId
            <select name="studentId" id="studentSelect" required>
              <option value="">Выберите ученика</option>
            </select>
          </label>
          ${formField("subject")}
          ${formField("value", "number")}
          <label>comment<input name="comment" /></label>
          <button type="submit">Сохранить оценку</button>
        </form>
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
    document.getElementById("gradeForm").onsubmit = submitForm("/api/teacher/grades", normalizeGrade);
    document.getElementById("homeworkForm").onsubmit = submitForm("/api/teacher/homework");
    loadTeacherStudents();
    return;
  }

  dashboard.innerHTML = [
    card("Моё расписание", `<button id="loadSchedule">Загрузить</button><div id="scheduleList" class="list"></div>`),
    card("Мои оценки", `<button id="loadGrades">Загрузить</button><div id="gradesList" class="list"></div>`),
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
      document.getElementById("gradesList").innerHTML = rows
        .map((r) => `<div class="item">${r.date}: ${r.subject} - ${r.value} (${r.comment || "без комментария"})</div>`)
        .join("");
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

function normalizeGrade(body) {
  return {
    studentId: Number(body.studentId),
    subject: body.subject,
    value: Number(body.value),
    comment: body.comment,
  };
}

async function loadTeacherStudents() {
  const select = document.getElementById("studentSelect");
  if (!select) return;

  try {
    const students = await api("/api/teacher/students");
    const options = ['<option value="">Выберите ученика</option>'];
    for (const student of students) {
      const className = student.className || "-";
      options.push(
        `<option value="${student.id}">${className} | ${student.fullName} (#${student.id})</option>`
      );
    }
    select.innerHTML = options.join("");
  } catch (err) {
    log("Ошибка загрузки учеников", { error: err.message });
  }
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
