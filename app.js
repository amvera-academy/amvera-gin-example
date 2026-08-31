const statusBox = document.querySelector("#response-status");
const bodyBox = document.querySelector("#response-body");
const itemsBox = document.querySelector("#items");

async function request(path, options = {}) {
  const response = await fetch(path, options);
  const body = await response.json();
  statusBox.textContent = `${options.method || "GET"} ${path} — HTTP ${response.status}`;
  bodyBox.textContent = JSON.stringify(body, null, 2);
  return { response, body };
}

async function loadItems() {
  const result = await request("/api/items");
  itemsBox.replaceChildren();
  for (const item of result.body.items) {
    const row = document.createElement("div");
    row.className = "item";
    const name = document.createElement("span");
    name.textContent = `#${item.id} — ${item.name}`;
    const button = document.createElement("button");
    button.textContent = `DELETE /api/items/${item.id}`;
    button.addEventListener("click", async () => {
      await request(`/api/items/${item.id}`, { method: "DELETE" });
      await loadItems();
    });
    row.append(name, button);
    itemsBox.append(row);
  }
}

document.querySelector('[data-action="health"]').addEventListener("click", () => request("/api/health"));
document.querySelector('[data-action="items"]').addEventListener("click", loadItems);

document.querySelector("#item-form").addEventListener("submit", async event => {
  event.preventDefault();
  const input = document.querySelector("#item-name");
  await request("/api/items", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ name: input.value })
  });
  input.value = "";
  await loadItems();
});

loadItems();
