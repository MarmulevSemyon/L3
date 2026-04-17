async function api(url, options = {}) {
  const res = await fetch(url, options);
  const text = await res.text();

  let data = {};
  try {
    data = text ? JSON.parse(text) : {};
  } catch (_) {
    data = { raw: text };
  }

  if (!res.ok) {
    const msg = data.error || data.message || text || `HTTP ${res.status}`;
    throw new Error(msg);
  }

  return data;
}

function normEvent(e) {
  return {
    id: e.id ?? e.ID,
    title: e.title ?? e.Title,
    description: e.description ?? e.Description,
    event_date: e.event_date ?? e.EventDate,
    total_spots: e.total_spots ?? e.TotalSpots,
    requires_payment: e.requires_payment ?? e.RequiresPayment,
    booking_ttl: e.booking_ttl ?? e.BookingTTL,
    created_at: e.created_at ?? e.CreatedAt,
    updated_at: e.updated_at ?? e.UpdatedAt,
  };
}

function normBooking(b) {
  return {
    id: b.id ?? b.ID,
    event_id: b.event_id ?? b.EventID,
    user_id: b.user_id ?? b.UserID,
    status: b.status ?? b.Status,
    created_at: b.created_at ?? b.CreatedAt,
    updated_at: b.updated_at ?? b.UpdatedAt,
  };
}

function normUser(u) {
  return {
    id: u.id ?? u.ID,
    username: u.username ?? u.Username,
  };
}

function normDetails(payload) {
  const details = payload.event ?? payload.details ?? payload;
  const event = normEvent(details.event ?? details.Event ?? {});
  const bookingsRaw = details.bookings ?? details.Bookings ?? [];
  const availableSpots = details.available_spots ?? details.AvailableSpots ?? 0;

  return {
    event,
    available_spots: availableSpots,
    bookings: bookingsRaw.map(normBooking),
  };
}

function currentUserId() {
  const select = document.getElementById("currentUser");
  return select ? select.value : "";
}

function formatDate(s) {
  if (!s) return "—";
  const d = new Date(s);
  if (Number.isNaN(d.getTime())) return s;
  return d.toLocaleString();
}

function renderBookingsList(bookings) {
  if (!bookings.length) {
    return "<p>Активных броней нет</p>";
  }

  return `
    <ul>
      ${bookings.map(b => `
        <li>
          user_id: ${b.user_id}, статус: <strong>${b.status}</strong>, создана: ${formatDate(b.created_at)}
        </li>
      `).join("")}
    </ul>
  `;
}

async function loadUsers() {
  try {
    const data = await api("/api/users");
    const users = (data.users || []).map(normUser);

    const select = document.getElementById("currentUser");
    if (!select) return;

    const prev = select.value;
    select.innerHTML = "";

    users.forEach(u => {
      const option = document.createElement("option");
      option.value = u.id;
      option.textContent = `${u.username} (${u.id})`;
      select.appendChild(option);
    });

    if (prev) {
      select.value = prev;
    }
  } catch (err) {
    alert("Ошибка загрузки пользователей: " + err.message);
  }
}

async function createUser() {
  const input = document.getElementById("username");
  const username = input?.value?.trim();

  if (!username) {
    alert("Введите имя пользователя");
    return;
  }

  try {
    const data = await api("/api/users", {
      method: "POST",
      headers: {"Content-Type": "application/json"},
      body: JSON.stringify({ username })
    });

    alert("Пользователь создан:\n" + JSON.stringify(data, null, 2));
    input.value = "";
    await loadUsers();
  } catch (err) {
    alert("Ошибка создания пользователя: " + err.message);
  }
}

async function submitEventForm() {
  const title = document.getElementById("title")?.value?.trim();
  const description = document.getElementById("description")?.value?.trim();
  const rawDate = document.getElementById("event_date")?.value;
  const totalSpots = Number(document.getElementById("total_spots")?.value);
  const bookingTTLMin = Number(document.getElementById("booking_ttl_min")?.value);
  const requiresPayment = document.getElementById("requires_payment")?.checked ?? true;

  if (!title) {
    alert("Введите название мероприятия");
    return;
  }

  if (!description) {
    alert("Введите описание мероприятия");
    return;
  }

  if (!rawDate) {
    alert("Выберите дату мероприятия");
    return;
  }

  const date = new Date(rawDate);
  if (Number.isNaN(date.getTime())) {
    alert("Некорректная дата мероприятия");
    return;
  }

  if (!Number.isInteger(totalSpots) || totalSpots <= 0) {
    alert("Количество мест должно быть больше 0");
    return;
  }

  if (!Number.isInteger(bookingTTLMin) || bookingTTLMin <= 0) {
    alert("TTL брони должен быть больше 0 минут");
    return;
  }

  const payload = {
    title,
    description,
    event_date: date.toISOString(),
    total_spots: totalSpots,
    requires_payment: requiresPayment,
    booking_ttl_min: bookingTTLMin
  };

  try {
    const data = await api("/api/events", {
      method: "POST",
      headers: {"Content-Type": "application/json"},
      body: JSON.stringify(payload)
    });

    alert("Мероприятие создано:\n" + JSON.stringify(data, null, 2));

    document.getElementById("title").value = "";
    document.getElementById("description").value = "";
    document.getElementById("event_date").value = "";
    document.getElementById("total_spots").value = "";
    document.getElementById("booking_ttl_min").value = "";
    document.getElementById("requires_payment").checked = true;

    await loadAdminEvents();
  } catch (err) {
    alert("Ошибка создания мероприятия: " + err.message);
  }
}

async function fetchEventDetails(eventId) {
  const data = await api(`/api/events/${eventId}`);
  return normDetails(data);
}

async function loadAdminEvents() {
  const root = document.getElementById("admin-events");
  if (!root) return;

  root.innerHTML = "Загрузка...";

  try {
    const data = await api("/api/events");
    const events = (data.events || []).map(normEvent);

    if (!events.length) {
      root.innerHTML = "<p>Мероприятий пока нет</p>";
      return;
    }

    const detailsList = await Promise.all(events.map(e => fetchEventDetails(e.id)));

    root.innerHTML = detailsList.map(d => `
      <div class="card">
        <h3>${d.event.title}</h3>
        <p>${d.event.description}</p>
        <p><strong>ID:</strong> ${d.event.id}</p>
        <p><strong>Дата:</strong> ${formatDate(d.event.event_date)}</p>
        <p><strong>Всего мест:</strong> ${d.event.total_spots}</p>
        <p><strong>Свободные места:</strong> ${d.available_spots}</p>
        <p><strong>Требует подтверждение:</strong> ${d.event.requires_payment ? "да" : "нет"}</p>
        <p><strong>Текущие брони:</strong> ${d.bookings.length}</p>
        ${renderBookingsList(d.bookings)}
      </div>
    `).join("");
  } catch (err) {
    root.innerHTML = `<p>Ошибка загрузки мероприятий: ${err.message}</p>`;
  }
}

async function bookEvent(eventId) {
  const userId = currentUserId();
  if (!userId) {
    alert("Сначала выбери пользователя");
    return;
  }

  try {
    const data = await api(`/api/events/${eventId}/book`, {
      method: "POST",
      headers: {"Content-Type": "application/json"},
      body: JSON.stringify({ user_id: userId })
    });

    alert(JSON.stringify(data, null, 2));
    await loadUserEvents();
  } catch (err) {
    alert("Ошибка бронирования: " + err.message);
  }
}

async function confirmBooking(eventId) {
  const userId = currentUserId();
  if (!userId) {
    alert("Сначала выбери пользователя");
    return;
  }

  try {
    const data = await api(`/api/events/${eventId}/confirm`, {
      method: "POST",
      headers: {"Content-Type": "application/json"},
      body: JSON.stringify({ user_id: userId })
    });

    alert(JSON.stringify(data, null, 2));
    await loadUserEvents();
  } catch (err) {
    alert("Ошибка подтверждения: " + err.message);
  }
}

async function loadUserEvents() {
  const root = document.getElementById("user-events");
  if (!root) return;

  root.innerHTML = "Загрузка...";

  try {
    const data = await api("/api/events");
    const events = (data.events || []).map(normEvent);

    if (!events.length) {
      root.innerHTML = "<p>Мероприятий пока нет</p>";
      return;
    }

    const detailsList = await Promise.all(events.map(e => fetchEventDetails(e.id)));

    root.innerHTML = detailsList.map(d => `
      <div class="card">
        <h3>${d.event.title}</h3>
        <p>${d.event.description}</p>
        <p><strong>Дата:</strong> ${formatDate(d.event.event_date)}</p>
        <p><strong>ID:</strong> ${d.event.id}</p>
        <p><strong>Свободные места:</strong> ${d.available_spots}</p>
        <p><strong>Текущие брони:</strong> ${d.bookings.length}</p>
        <p><strong>Статусы броней:</strong></p>
        ${renderBookingsList(d.bookings)}

        <button type="button" onclick="bookEvent('${d.event.id}')">Забронировать место</button>
        <button type="button" onclick="confirmBooking('${d.event.id}')">Подтвердить бронь</button>
      </div>
    `).join("");
  } catch (err) {
    root.innerHTML = `<p>Ошибка загрузки мероприятий: ${err.message}</p>`;
  }
}

document.addEventListener("DOMContentLoaded", async () => {
  await loadUsers();
  await loadUserEvents();
  await loadAdminEvents();
});