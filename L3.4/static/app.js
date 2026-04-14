const form = document.getElementById("uploadForm");
const imagesContainer = document.getElementById("images");

form.addEventListener("submit", async (e) => {
  e.preventDefault();

  const fileInput = document.getElementById("fileInput");
  const resize = document.getElementById("resize").checked;
  const thumb = document.getElementById("thumb").checked;
  const watermark = document.getElementById("watermark").checked;

  if (!fileInput.files.length) {
    alert("Выберите файл");
    return;
  }

  const formData = new FormData();
  formData.append("file", fileInput.files[0]);
  formData.append("resize", String(resize));
  formData.append("thumb", String(thumb));
  formData.append("watermark", String(watermark));

  try {
    const res = await fetch("/upload", {
      method: "POST",
      body: formData,
    });

    const data = await res.json();

    if (!res.ok) {
      alert(data.error || "Ошибка загрузки");
      return;
    }

    if (!data.id) {
      alert("Сервер не вернул id изображения");
      return;
    }

    addImageCard(data.id);
  } catch (err) {
    console.error(err);
    alert("Ошибка сети при загрузке файла");
  }
});

function addImageCard(id) {
  const card = document.createElement("div");
  card.className = "card";
  card.id = `card-${id}`;

  card.innerHTML = `
    <p><b>ID:</b> ${id}</p>
    <p class="status">Статус: pending</p>
    <div class="preview">В обработке...</div>
    <button class="delete-btn">Удалить</button>
  `;

  card.querySelector(".delete-btn").addEventListener("click", async () => {
    try {
      const res = await fetch(`/image/${id}`, { method: "DELETE" });
      if (!res.ok) {
        const data = await res.json();
        alert(data.error || "Ошибка удаления");
        return;
      }
      card.remove();
    } catch (err) {
      console.error(err);
      alert("Ошибка сети при удалении");
    }
  });

  imagesContainer.prepend(card);
  pollStatus(id, card);
}

async function pollStatus(id, card) {
  const interval = setInterval(async () => {
    try {
      const res = await fetch(`/image/${id}`);
      if (!res.ok) {
        const data = await res.json();
        card.querySelector(".preview").textContent = data.error || "Ошибка получения статуса";
        clearInterval(interval);
        return;
      }

      const data = await res.json();
      card.querySelector(".status").textContent = `Статус: ${data.status}`;

      const preview = card.querySelector(".preview");

      if (data.status === "done") {
        preview.innerHTML = `
          <div>
            ${data.image_url ? `<img src="${data.image_url}" class="main-image" />` : ""}
            ${data.thumb_url ? `<img src="${data.thumb_url}" class="thumb-image" />` : ""}
          </div>
        `;
        clearInterval(interval);
      }

      if (data.status === "failed") {
        preview.textContent = `Ошибка: ${data.error_text || "unknown error"}`;
        clearInterval(interval);
      }
    } catch (err) {
      console.error(err);
      card.querySelector(".preview").textContent = "Ошибка сети";
      clearInterval(interval);
    }
  }, 2000);
}