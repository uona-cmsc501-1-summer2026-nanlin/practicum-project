(() => {
  const root = document.getElementById("app");
  const modalRoot = document.getElementById("modal-root");
  const confirmRoot = document.getElementById("confirm-root");
  const flashEl = document.getElementById("flash");
  const LAST_GROUP_KEY = "billsplitter.lastGroupId";
  const THEME_KEY = "billsplitter.theme";
  const THEMES = [
    { id: "forest", label: "Forest" },
    { id: "ocean", label: "Ocean" },
    { id: "dusk", label: "Dusk" },
    { id: "dark", label: "Dark" },
  ];

  let flashTimer = null;
  let currenciesCache = null;
  let activeGroupId = null;

  const icons = {
    switch: `<i class="fa-solid fa-right-left" aria-hidden="true"></i>`,
    members: `<i class="fa-solid fa-users" aria-hidden="true"></i>`,
    plus: `<i class="fa-solid fa-plus" aria-hidden="true"></i>`,
    settle: `<i class="fa-solid fa-scale-balanced" aria-hidden="true"></i>`,
    edit: `<i class="fa-solid fa-pen" aria-hidden="true"></i>`,
    trash: `<i class="fa-solid fa-trash" aria-hidden="true"></i>`,
    close: `<i class="fa-solid fa-xmark" aria-hidden="true"></i>`,
    people: `<i class="fa-solid fa-user-group" aria-hidden="true"></i>`,
    tags: `<i class="fa-solid fa-tags" aria-hidden="true"></i>`,
    chevron: `<i class="fa-solid fa-chevron-down" aria-hidden="true"></i>`,
  };

  const CHARGE_SORT_KEY = "billsplitter.chargeSort";
  const SIDE_TAB_KEY = "billsplitter.sideTab";

  // Popular defaults shown before search; full Free solid set loads from fa-solid-icons.json.
  const SUGGESTED_ICONS = [
    { icon: "fa-solid fa-utensils", labels: "food restaurant meal utensils dining" },
    { icon: "fa-solid fa-cart-shopping", labels: "groceries cart shopping market" },
    { icon: "fa-solid fa-car", labels: "transport car drive taxi" },
    { icon: "fa-solid fa-bus", labels: "bus transit transport" },
    { icon: "fa-solid fa-train", labels: "train transit rail" },
    { icon: "fa-solid fa-plane", labels: "travel plane flight airport" },
    { icon: "fa-solid fa-bed", labels: "lodging hotel bed sleep" },
    { icon: "fa-solid fa-house", labels: "home house rent" },
    { icon: "fa-solid fa-film", labels: "entertainment film movie cinema" },
    { icon: "fa-solid fa-music", labels: "music entertainment concert" },
    { icon: "fa-solid fa-gamepad", labels: "games entertainment" },
    { icon: "fa-solid fa-bag-shopping", labels: "shopping bag retail" },
    { icon: "fa-solid fa-shirt", labels: "clothes shirt shopping" },
    { icon: "fa-solid fa-bolt", labels: "utilities electricity power" },
    { icon: "fa-solid fa-wifi", labels: "wifi internet utilities" },
    { icon: "fa-solid fa-droplet", labels: "water utilities" },
    { icon: "fa-solid fa-heart-pulse", labels: "health medical doctor" },
    { icon: "fa-solid fa-pills", labels: "pharmacy medicine health" },
    { icon: "fa-solid fa-ellipsis", labels: "other misc more" },
    { icon: "fa-solid fa-gift", labels: "gift present" },
    { icon: "fa-solid fa-cake-candles", labels: "cake party birthday" },
    { icon: "fa-solid fa-mug-hot", labels: "coffee cafe drink mug" },
    { icon: "fa-solid fa-beer-mug-empty", labels: "beer drink bar" },
    { icon: "fa-solid fa-wine-glass", labels: "wine drink" },
    { icon: "fa-solid fa-ice-cream", labels: "dessert ice cream" },
    { icon: "fa-solid fa-pizza-slice", labels: "pizza food" },
    { icon: "fa-solid fa-gas-pump", labels: "gas fuel transport" },
    { icon: "fa-solid fa-parking", labels: "parking transport" },
    { icon: "fa-solid fa-ticket", labels: "ticket event entertainment" },
    { icon: "fa-solid fa-dumbbell", labels: "gym fitness health" },
    { icon: "fa-solid fa-paw", labels: "pet dog cat animal" },
    { icon: "fa-solid fa-book", labels: "book education study" },
    { icon: "fa-solid fa-laptop", labels: "laptop work tech" },
    { icon: "fa-solid fa-phone", labels: "phone mobile" },
    { icon: "fa-solid fa-taxi", labels: "taxi uber transport" },
    { icon: "fa-solid fa-bicycle", labels: "bike bicycle" },
    { icon: "fa-solid fa-umbrella-beach", labels: "beach vacation travel" },
    { icon: "fa-solid fa-campground", labels: "camp outdoor travel" },
    { icon: "fa-solid fa-mountain", labels: "hike mountain outdoor" },
    { icon: "fa-solid fa-baby", labels: "baby kids child" },
    { icon: "fa-solid fa-briefcase", labels: "work office business" },
    { icon: "fa-solid fa-receipt", labels: "receipt bill" },
    { icon: "fa-solid fa-wallet", labels: "wallet money" },
    { icon: "fa-solid fa-credit-card", labels: "card payment" },
    { icon: "fa-solid fa-hand-holding-dollar", labels: "money tip cash" },
    { icon: "fa-solid fa-screwdriver-wrench", labels: "repair tools maintenance" },
    { icon: "fa-solid fa-couch", labels: "furniture home" },
    { icon: "fa-solid fa-soap", labels: "cleaning laundry" },
  ];

  let categoriesCache = null;
  let faSolidIconsCache = null;

  function getTheme() {
    const stored = localStorage.getItem(THEME_KEY);
    if (THEMES.some((t) => t.id === stored)) return stored;
    return "forest";
  }

  function applyTheme(id) {
    const theme = THEMES.find((t) => t.id === id) || THEMES[0];
    document.documentElement.setAttribute("data-theme", theme.id);
    localStorage.setItem(THEME_KEY, theme.id);
    const btn = document.getElementById("btn-theme");
    if (btn) {
      btn.title = `Theme: ${theme.label} (click to switch)`;
      btn.setAttribute("aria-label", `Color theme: ${theme.label}. Click to switch.`);
    }
  }

  function cycleTheme() {
    const current = getTheme();
    const idx = THEMES.findIndex((t) => t.id === current);
    const next = THEMES[(idx + 1) % THEMES.length];
    applyTheme(next.id);
    flash(`Theme: ${next.label}`, "ok");
  }

  function flash(message, kind = "error") {
    flashEl.hidden = false;
    flashEl.className = "flash " + kind;
    flashEl.textContent = message;
    clearTimeout(flashTimer);
    flashTimer = setTimeout(() => {
      flashEl.hidden = true;
    }, 4000);
  }

  function esc(s) {
    return String(s ?? "")
      .replaceAll("&", "&amp;")
      .replaceAll("<", "&lt;")
      .replaceAll(">", "&gt;")
      .replaceAll('"', "&quot;");
  }

  function money(n) {
    return Number(n).toFixed(2);
  }

  function todayISO() {
    const d = new Date();
    const y = d.getFullYear();
    const m = String(d.getMonth() + 1).padStart(2, "0");
    const day = String(d.getDate()).padStart(2, "0");
    return `${y}-${m}-${day}`;
  }

  function formatChargeDate(iso) {
    const s = String(iso || "").trim();
    if (!/^\d{4}-\d{2}-\d{2}$/.test(s)) return s || "—";
    const [y, m, d] = s.split("-");
    return `${m}/${d}/${y}`;
  }

  function getChargeSort() {
    const v = localStorage.getItem(CHARGE_SORT_KEY);
    const allowed = [
      "date-desc",
      "date-asc",
      "amount-desc",
      "amount-asc",
      "name-asc",
      "name-desc",
    ];
    return allowed.includes(v) ? v : "date-desc";
  }

  function sortCharges(charges, mode) {
    const list = [...(charges || [])];
    const cmpStr = (a, b) => String(a).localeCompare(String(b));
    list.sort((a, b) => {
      switch (mode) {
        case "date-asc":
          return cmpStr(a.date || "", b.date || "") || a.id - b.id;
        case "amount-desc":
          return Number(b.amount) - Number(a.amount) || b.id - a.id;
        case "amount-asc":
          return Number(a.amount) - Number(b.amount) || a.id - b.id;
        case "name-asc":
          return cmpStr(a.description || "", b.description || "") || a.id - b.id;
        case "name-desc":
          return cmpStr(b.description || "", a.description || "") || b.id - a.id;
        case "date-desc":
        default:
          return cmpStr(b.date || "", a.date || "") || b.id - a.id;
      }
    });
    return list;
  }

  function getSideTab() {
    const v = localStorage.getItem(SIDE_TAB_KEY);
    return v === "settlement" ? "settlement" : "overview";
  }

  function setSideTab(id) {
    localStorage.setItem(SIDE_TAB_KEY, id === "settlement" ? "settlement" : "overview");
  }

  function sideTabsHTML(settle, currency) {
    const tab = getSideTab();
    const overviewActive = tab === "overview";
    return `
      <section class="pane side-tabs-pane">
        <div class="side-tabs" role="tablist" aria-label="Overview and settlement">
          <button type="button" class="side-tab${
            overviewActive ? " is-active" : ""
          }" role="tab" id="tab-overview" data-side-tab="overview" aria-selected="${
            overviewActive ? "true" : "false"
          }" aria-controls="panel-overview">Overview</button>
          <button type="button" class="side-tab${
            overviewActive ? "" : " is-active"
          }" role="tab" id="tab-settlement" data-side-tab="settlement" aria-selected="${
            overviewActive ? "false" : "true"
          }" aria-controls="panel-settlement">Settlement</button>
        </div>
        <div class="pane-body side-tab-panels">
          <div class="side-tab-panel${
            overviewActive ? " is-active" : ""
          }" role="tabpanel" id="panel-overview" aria-labelledby="tab-overview" ${
            overviewActive ? "" : "hidden"
          }>
            ${overviewPanelBodyHTML(settle, currency)}
          </div>
          <div class="side-tab-panel${
            overviewActive ? "" : " is-active"
          }" role="tabpanel" id="panel-settlement" aria-labelledby="tab-settlement" ${
            overviewActive ? "hidden" : ""
          }>
            ${settlementPanelBodyHTML(settle, currency)}
          </div>
        </div>
      </section>`;
  }

  function wireSideTabs(scope) {
    const pane = scope.querySelector(".side-tabs-pane");
    if (!pane) return;
    pane.querySelectorAll("[data-side-tab]").forEach((btn) => {
      btn.addEventListener("click", () => {
        const id = btn.dataset.sideTab === "settlement" ? "settlement" : "overview";
        setSideTab(id);
        pane.querySelectorAll("[data-side-tab]").forEach((b) => {
          const on = b.dataset.sideTab === id;
          b.classList.toggle("is-active", on);
          b.setAttribute("aria-selected", on ? "true" : "false");
        });
        pane.querySelectorAll(".side-tab-panel").forEach((panel) => {
          const on = panel.id === "panel-" + id;
          panel.classList.toggle("is-active", on);
          panel.hidden = !on;
        });
      });
    });
  }

  function signedMoney(n) {
    const v = Number(n) || 0;
    if (v > 0.005) return "+" + money(v);
    if (v < -0.005) return "-" + money(Math.abs(v));
    return money(0);
  }

  /** Per-charge net for each group member: paid − share (equal split). */
  function chargePersonNetsHTML(charge, members, nameById) {
    const amount = Number(charge.amount) || 0;
    const share = Number(charge.sharePerPerson) || 0;
    const payerId = Number(charge.paidByUserId);
    const partSet = new Set((charge.participantIds || []).map(Number));
    const rows = (members || [])
      .map((m) => {
        const id = Number(m.id);
        const paid = id === payerId ? amount : 0;
        const owed = partSet.has(id) ? share : 0;
        return {
          id,
          name: nameById[id] || m.name || "#" + id,
          net: paid - owed,
        };
      })
      .filter((r) => Math.abs(r.net) > 0.005)
      .sort((a, b) => b.net - a.net || a.name.localeCompare(b.name));

    if (rows.length === 0) return `<span class="tx-person-empty">—</span>`;
    return rows
      .map((r) => {
        const tone = r.net > 0 ? "credit" : "debt";
        return `<span class="tx-person ${tone}">${esc(r.name)} ${signedMoney(
          r.net
        )}</span>`;
      })
      .join('<span class="tx-person-sep"> · </span>');
  }

  function getStoredGroupId() {
    const v = localStorage.getItem(LAST_GROUP_KEY);
    return v ? Number(v) : null;
  }

  function setStoredGroupId(id) {
    if (id == null) localStorage.removeItem(LAST_GROUP_KEY);
    else localStorage.setItem(LAST_GROUP_KEY, String(id));
  }

  function parseRoute() {
    const hash = location.hash.replace(/^#\/?/, "");
    if (!hash || hash === "home") return { page: "home" };
    const [page, id] = hash.split("/");
    if (page === "group" && id) return { page: "group", id };
    return { page: "home" };
  }

  function goGroup(id) {
    setStoredGroupId(id);
    location.hash = "#/group/" + id;
  }

  async function getCurrencies() {
    if (!currenciesCache) {
      const { data } = await api("GET", "/currencies");
      currenciesCache = data;
    }
    return currenciesCache;
  }

  async function getCategories(force = false) {
    if (force || !categoriesCache) {
      const { data } = await api("GET", "/categories");
      categoriesCache = data || [];
    }
    return categoriesCache;
  }

  function categoryIconHTML(iconClass, extra = "") {
    const cls = String(iconClass || "fa-solid fa-ellipsis").replace(/[^a-z0-9\-\s]/gi, "");
    return `<i class="${esc(cls)}${extra ? " " + extra : ""}" aria-hidden="true"></i>`;
  }

  function categorySelectHTML(categories, selectedId) {
    const sel =
      selectedId != null ? Number(selectedId) : defaultCategoryId(categories);
    const current =
      categories.find((c) => Number(c.id) === sel) || categories[0] || null;
    const disabled = !categories.length;
    return `
      <div class="category-select-field">
        <span class="field-label">Category</span>
        <div class="category-picker${disabled ? " is-disabled" : ""}" data-category-picker>
          <input type="hidden" name="categoryId" value="${
            current ? esc(current.id) : ""
          }" ${disabled ? "disabled" : "required"} />
          <button type="button" class="category-trigger" aria-haspopup="listbox" aria-expanded="false" ${
            disabled ? "disabled" : ""
          }>
            <span class="category-trigger-icon">${categoryIconHTML(
              current?.icon
            )}</span>
            <span class="category-trigger-text">${esc(
              current?.name || "No categories"
            )}</span>
            <span class="category-caret" aria-hidden="true">▾</span>
          </button>
          <ul class="category-menu" role="listbox" hidden>
            ${categories
              .map((c) => {
                const id = Number(c.id);
                return `
              <li role="option" data-id="${id}" data-name="${esc(
                  c.name
                )}" data-icon="${esc(c.icon)}" ${
                  id === Number(current?.id) ? 'aria-selected="true"' : ""
                }>
                <span class="category-option-icon">${categoryIconHTML(
                  c.icon
                )}</span>
                <span>${esc(c.name)}</span>
              </li>`;
              })
              .join("")}
          </ul>
        </div>
      </div>`;
  }

  function wireCategorySelect(scope) {
    const picker = scope.querySelector("[data-category-picker]");
    if (!picker || picker.classList.contains("is-disabled")) return;
    const input = picker.querySelector('input[name="categoryId"]');
    const trigger = picker.querySelector(".category-trigger");
    const menu = picker.querySelector(".category-menu");
    const iconEl = picker.querySelector(".category-trigger-icon");
    const label = picker.querySelector(".category-trigger-text");

    const close = () => {
      menu.hidden = true;
      trigger.setAttribute("aria-expanded", "false");
      picker.classList.remove("is-open");
    };

    const open = () => {
      menu.hidden = false;
      trigger.setAttribute("aria-expanded", "true");
      picker.classList.add("is-open");
    };

    trigger.addEventListener("click", (e) => {
      e.preventDefault();
      e.stopPropagation();
      if (menu.hidden) open();
      else close();
    });

    menu.querySelectorAll("[data-id]").forEach((item) => {
      item.addEventListener("click", (e) => {
        e.preventDefault();
        e.stopPropagation();
        input.value = item.dataset.id;
        label.textContent = item.dataset.name || "";
        iconEl.innerHTML = categoryIconHTML(item.dataset.icon);
        menu.querySelectorAll("[aria-selected]").forEach((el) =>
          el.removeAttribute("aria-selected")
        );
        item.setAttribute("aria-selected", "true");
        close();
      });
    });

    document.addEventListener("click", (e) => {
      if (!menu.hidden && !picker.contains(e.target)) close();
    });
  }

  function defaultCategoryId(categories) {
    const other = categories.find((c) => c.name === "Other" && c.builtin);
    if (other) return Number(other.id);
    return categories[0] ? Number(categories[0].id) : null;
  }

  function flagImg(country, size = 20) {
    const code = String(country || "").toLowerCase();
    if (!code) return "";
    const h = Math.round((size * 3) / 4);
    return `<img class="flag" src="https://flagcdn.com/w40/${esc(
      code
    )}.png" width="${size}" height="${h}" alt="" loading="lazy" />`;
  }

  function currencyPickerHTML(currencies, selected = "USD") {
    const current = currencies.find((c) => c.code === selected) || currencies[0];
    return `
      <div class="currency-picker" data-currency-picker>
        <input type="hidden" name="currency" value="${esc(current.code)}" required />
        <button type="button" class="currency-trigger" aria-haspopup="listbox" aria-expanded="false">
          ${flagImg(current.country)}
          <span class="currency-trigger-text">${esc(current.code)} — ${esc(current.name)}</span>
          <span class="currency-caret" aria-hidden="true">▾</span>
        </button>
        <ul class="currency-menu" role="listbox" hidden>
          ${currencies
            .map(
              (c) => `
            <li role="option" data-code="${esc(c.code)}" data-country="${esc(
                c.country
              )}" data-name="${esc(c.name)}" ${
                c.code === current.code ? 'aria-selected="true"' : ""
              }>
              ${flagImg(c.country)}
              <span>${esc(c.code)} — ${esc(c.name)}</span>
            </li>`
            )
            .join("")}
        </ul>
      </div>
    `;
  }

  function wireCurrencyPicker(scope) {
    const picker = scope.querySelector("[data-currency-picker]");
    if (!picker) return;
    const input = picker.querySelector('input[name="currency"]');
    const trigger = picker.querySelector(".currency-trigger");
    const menu = picker.querySelector(".currency-menu");
    const label = picker.querySelector(".currency-trigger-text");

    const close = () => {
      menu.hidden = true;
      trigger.setAttribute("aria-expanded", "false");
      picker.classList.remove("is-open");
    };

    const open = () => {
      menu.hidden = false;
      trigger.setAttribute("aria-expanded", "true");
      picker.classList.add("is-open");
    };

    trigger.addEventListener("click", (e) => {
      e.preventDefault();
      e.stopPropagation();
      if (menu.hidden) open();
      else close();
    });

    menu.querySelectorAll("[data-code]").forEach((item) => {
      item.addEventListener("click", (e) => {
        e.preventDefault();
        e.stopPropagation();
        input.value = item.dataset.code;
        label.textContent = `${item.dataset.code} — ${item.dataset.name}`;
        const wrap = document.createElement("div");
        wrap.innerHTML = flagImg(item.dataset.country);
        const nextFlag = wrap.firstElementChild;
        const oldFlag = trigger.querySelector(".flag");
        if (oldFlag && nextFlag) oldFlag.replaceWith(nextFlag);
        menu.querySelectorAll("[aria-selected]").forEach((el) =>
          el.removeAttribute("aria-selected")
        );
        item.setAttribute("aria-selected", "true");
        close();
      });
    });

    document.addEventListener("click", (e) => {
      if (!menu.hidden && !picker.contains(e.target)) close();
    });
  }

  function participantPickerHTML(members) {
    if (!members.length) {
      return `<div class="participant-picker" data-participant-picker>
        <p class="empty">Add members first.</p>
      </div>`;
    }
    return `
      <div class="participant-picker" data-participant-picker>
        <ul class="participant-list" data-participant-list></ul>
      </div>
    `;
  }

  function wireParticipantPicker(scope, members, selectedIds, payerId) {
    const picker = scope.querySelector("[data-participant-picker]");
    if (!picker || !picker.querySelector("[data-participant-list]")) return null;

    const listEl = picker.querySelector("[data-participant-list]");
    const byId = Object.fromEntries(members.map((m) => [Number(m.id), m]));
    const selected = new Set((selectedIds || []).map(Number).filter((id) => byId[id]));
    let payer =
      payerId != null && byId[Number(payerId)] ? Number(payerId) : null;

    const render = () => {
      listEl.innerHTML = members
        .map((m) => {
          const id = Number(m.id);
          const sharing = selected.has(id);
          const isPayer = payer === id;
          const paidTitle = isPayer
            ? "Paid this charge"
            : "Mark as who paid (can Skip the split)";
          return `
            <li class="participant-row${sharing ? " is-sharing" : " is-skipped"}${
              isPayer ? " is-payer" : ""
            }">
              <button type="button" class="participant-paid${
                isPayer ? " is-on" : ""
              }" data-set-payer="${id}" role="radio" aria-checked="${
                isPayer ? "true" : "false"
              }" title="${esc(paidTitle)}" aria-label="${esc(paidTitle)}">
                <span class="participant-paid-dot" aria-hidden="true"></span>
                <span class="participant-paid-label">${isPayer ? "Paid" : "Paid?"}</span>
              </button>
              <div class="participant-toggle-track" data-member-id="${id}">
                <button type="button" class="participant-toggle-side on participant-side-hit" data-set-share="${id}" aria-label="Share: ${esc(
                  m.name
                )}">
                  <span class="participant-toggle-caption">Share</span>
                  <span class="participant-toggle-name">${esc(m.name)}</span>
                </button>
                <button type="button" class="participant-toggle-side off participant-side-hit" data-set-skip="${id}" aria-label="Skip: ${esc(
                  m.name
                )}">
                  <span class="participant-toggle-caption">Skip</span>
                  <span class="participant-toggle-name">${esc(m.name)}</span>
                </button>
                <span class="participant-toggle-knob" aria-hidden="true"></span>
              </div>
            </li>`;
        })
        .join("");

      listEl.querySelectorAll("[data-set-share]").forEach((btn) => {
        btn.addEventListener("click", () => {
          selected.add(Number(btn.dataset.setShare));
          render();
        });
      });
      listEl.querySelectorAll("[data-set-skip]").forEach((btn) => {
        btn.addEventListener("click", () => {
          selected.delete(Number(btn.dataset.setSkip));
          render();
        });
      });
      listEl.querySelectorAll("[data-set-payer]").forEach((btn) => {
        btn.addEventListener("click", () => {
          payer = Number(btn.dataset.setPayer);
          render();
        });
      });
    };

    render();
    return {
      getIds: () => members.map((m) => Number(m.id)).filter((id) => selected.has(id)),
      getPayerId: () => payer,
    };
  }

  function closeModal() {
    modalRoot.innerHTML = "";
  }

  function askConfirm(title, message, confirmLabel = "Delete") {
    return askDialog({
      title,
      bodyHtml: `<p>${esc(message)}</p>`,
      confirmLabel,
      showConfirm: true,
    });
  }

  function askNotice(title, bodyHtml) {
    return askDialog({
      title,
      bodyHtml,
      confirmLabel: "Close",
      showConfirm: false,
    }).then(() => undefined);
  }

  function askDialog({ title, bodyHtml, confirmLabel = "OK", showConfirm = true }) {
    return new Promise((resolve) => {
      const actions = showConfirm
        ? `<button type="button" class="ghost" data-confirm-cancel>Cancel</button>
           <button type="button" class="danger-confirm" data-confirm-ok>${esc(confirmLabel)}</button>`
        : `<button type="button" data-confirm-ok>${esc(confirmLabel)}</button>`;
      confirmRoot.innerHTML = `
        <div class="confirm-backdrop">
          <div class="confirm-card" role="dialog" aria-modal="true" aria-labelledby="confirm-title">
            <h3 id="confirm-title">${esc(title)}</h3>
            <div class="confirm-body">${bodyHtml}</div>
            <div class="confirm-actions">${actions}</div>
          </div>
        </div>
      `;
      const finish = (ok) => {
        confirmRoot.innerHTML = "";
        resolve(ok);
      };
      confirmRoot.querySelector("[data-confirm-cancel]")?.addEventListener("click", () => finish(false));
      confirmRoot.querySelector("[data-confirm-ok]").addEventListener("click", () => finish(showConfirm));
    });
  }

  async function noticeError(title, err) {
    const reason = (err && err.message) || "Something went wrong";
    await askNotice(title, `<p>${esc(reason)}</p>`);
  }

  function chargeInvolvesUser(charge, userId) {
    const uid = Number(userId);
    if (Number(charge.paidByUserId) === uid) return true;
    return (charge.participantIds || []).some((id) => Number(id) === uid);
  }

  function chargesForUser(charges, userId) {
    return (charges || []).filter((c) => chargeInvolvesUser(c, userId));
  }

  function chargeCountLabel(n) {
    if (n === 0) return "0 charges";
    if (n === 1) return "1 charge";
    return `${n} charges`;
  }

  function memberChargesBody(name, blocking) {
    if (!blocking.length) {
      return `<p>Remove “${esc(name)}” from this group?</p>`;
    }
    const list = blocking
      .map(
        (c) =>
          `<li><strong>${esc(c.description)}</strong> <span class="meta">${money(
            c.amount
          )}</span></li>`
      )
      .join("");
    return `<p>“${esc(
      name
    )}” cannot be removed while linked to these group charges:</p>
      <ul class="confirm-list">${list}</ul>
      <p class="confirm-hint">Until all of these charges are cleared (deleted or edited so they are no longer the payer or a participant), this person cannot be removed from the group.</p>`;
  }

  async function showMemberCharges(name, blocking) {
    await askNotice(
      blocking.length ? "Associated charges" : "No associated charges",
      blocking.length
        ? memberChargesBody(name, blocking)
        : `<p>“${esc(name)}” is not linked to any group charges.</p>`
    );
  }

  async function confirmRemoveMember(name, blocking) {
    const blocked = blocking.length > 0;
    return askDialog({
      title: "Remove member",
      bodyHtml: memberChargesBody(name, blocking),
      confirmLabel: blocked ? "Close" : "Remove",
      showConfirm: !blocked,
    });
  }

  function groupDeleteBody(name, memberCount, chargeCount) {
    const blocked = memberCount > 0 || chargeCount > 0;
    if (!blocked) {
      return `<p>Delete “${esc(name)}”? This cannot be undone.</p>`;
    }
    const parts = [];
    if (memberCount > 0) {
      parts.push(`${memberCount} member${memberCount === 1 ? "" : "s"}`);
    }
    if (chargeCount > 0) {
      parts.push(`${chargeCount} charge${chargeCount === 1 ? "" : "s"}`);
    }
    return `<p>“${esc(name)}” cannot be deleted while it still has:</p>
      <ul class="confirm-list">${parts.map((p) => `<li><strong>${esc(p)}</strong></li>`).join("")}</ul>
      <p class="confirm-hint">Remove all members and clear all charges first. Then you can delete the group.</p>`;
  }

  async function confirmDeleteGroup(name, memberCount, chargeCount) {
    const blocked = memberCount > 0 || chargeCount > 0;
    return askDialog({
      title: "Delete group",
      bodyHtml: groupDeleteBody(name, memberCount, chargeCount),
      confirmLabel: blocked ? "Close" : "Delete",
      showConfirm: !blocked,
    });
  }

  function openModal(html, wide = false) {
    modalRoot.innerHTML = `
      <div class="modal-backdrop" data-backdrop>
        <div class="modal${wide ? " wide" : ""}" role="dialog" aria-modal="true">${html}</div>
      </div>
    `;
    // Form / content modals only close via explicit controls (not backdrop click).
    modalRoot.querySelector("[data-close]")?.addEventListener("click", closeModal);
    return modalRoot.querySelector(".modal");
  }

  async function openGroupModal(currentId) {
    const id = currentId != null ? Number(currentId) : null;
    const [{ data: groups }, currencies] = await Promise.all([
      api("GET", "/groups"),
      getCurrencies(),
    ]);
    const byCode = Object.fromEntries(currencies.map((c) => [c.code, c]));
    const currentGroup = id ? groups.find((g) => g.id === id) || null : null;

    const statsEntries = await Promise.all(
      groups.map(async (g) => {
        const [{ data: members }, { data: charges }] = await Promise.all([
          api("GET", "/groups/" + g.id + "/members"),
          api("GET", "/groups/" + g.id + "/charges"),
        ]);
        return [g.id, { members: members.length, charges: charges.length }];
      })
    );
    const statsById = Object.fromEntries(statsEntries);
    const memberCount = currentGroup ? statsById[currentGroup.id]?.members || 0 : 0;
    const chargeCount = currentGroup ? statsById[currentGroup.id]?.charges || 0 : 0;
    const deleteBlocked = memberCount > 0 || chargeCount > 0;
    const deleteTitle = deleteBlocked
      ? `Delete group — blocked by ${memberCount} member${
          memberCount === 1 ? "" : "s"
        }, ${chargeCount} charge${chargeCount === 1 ? "" : "s"}`
      : "Delete group";

    const sorted = [...groups].sort((a, b) => {
      if (id && a.id === id) return -1;
      if (id && b.id === id) return 1;
      return b.id - a.id;
    });

    const brief = currentGroup
      ? `
        <div class="modal-head">
          <div>
            <h2>${esc(currentGroup.name)}</h2>
            <div class="group-brief">
              <button type="button" class="text-link" id="btn-modal-members">
                ${memberCount} member${memberCount === 1 ? "" : "s"}
              </button>
              <span class="group-brief-sep">·</span>
              <span class="group-stat">
                <span>${chargeCount} charge${chargeCount === 1 ? "" : "s"}</span>
              </span>
            </div>
          </div>
          <div class="modal-head-actions">
            <button type="button" class="icon-btn" id="btn-modal-edit" title="Edit group" aria-label="Edit group">${icons.edit}</button>
            <button type="button" class="icon-btn danger" id="btn-modal-delete" title="${esc(
              deleteTitle
            )}" aria-label="${esc(deleteTitle)}">${icons.trash}</button>
            <button type="button" class="icon-btn" data-close title="Close" aria-label="Close">${icons.close}</button>
          </div>
        </div>`
      : `
        <div class="modal-head">
          <div>
            <h2>Switch group</h2>
            <p>Pick a workspace.</p>
          </div>
          <button type="button" class="icon-btn" data-close title="Close" aria-label="Close">${icons.close}</button>
        </div>`;

    const modal = openModal(`
      ${brief}
      ${
        sorted.length === 0
          ? `<p class="empty">No groups yet.</p>`
          : `<p class="modal-section-label">Groups</p>
            <ul class="modal-list">
              ${sorted
                .map((g) => {
                  const cur = byCode[g.currency];
                  const isCurrent = id === g.id;
                  const stats = statsById[g.id] || { members: 0, charges: 0 };
                  const membersLabel = `${stats.members} member${
                    stats.members === 1 ? "" : "s"
                  }`;
                  const chargesLabel = `${stats.charges} charge${
                    stats.charges === 1 ? "" : "s"
                  }`;
                  return `
                  <li class="${isCurrent ? "is-current" : ""}">
                    <button type="button" class="pick${isCurrent ? " is-current" : ""}" data-pick-group="${g.id}">
                      ${cur ? flagImg(cur.country, 22) : ""}
                      <span class="pick-main">
                        <strong>${esc(g.name)}</strong>
                        <span class="meta">${esc(g.currency)}${
                          isCurrent
                            ? ' · <span class="current-tag">current</span>'
                            : ""
                        }</span>
                      </span>
                      <span class="group-list-stats">${membersLabel} · ${chargesLabel}</span>
                    </button>
                  </li>`;
                })
                .join("")}
            </ul>`
      }
      <div class="modal-actions">
        <button type="button" id="btn-open-create-group">${icons.plus} New group</button>
      </div>
    `);

    modal.querySelectorAll("[data-pick-group]").forEach((btn) => {
      btn.addEventListener("click", () => {
        const pickId = Number(btn.dataset.pickGroup);
        closeModal();
        if (id && pickId === id) return;
        goGroup(pickId);
      });
    });

    modal.querySelector("#btn-open-create-group").addEventListener("click", () => {
      closeModal();
      openCreateGroupModal();
    });

    modal.querySelector("#btn-modal-members")?.addEventListener("click", () => {
      openMembersModal(id).catch((err) => flash(err.message));
    });

    modal.querySelector("#btn-modal-edit")?.addEventListener("click", () => {
      openEditGroupModal(currentGroup).catch((err) => flash(err.message));
    });

    modal.querySelector("#btn-modal-delete")?.addEventListener("click", async () => {
      if (!(await confirmDeleteGroup(currentGroup.name, memberCount, chargeCount))) {
        return;
      }
      try {
        await api("DELETE", "/groups/" + currentGroup.id);
        flash("Group deleted", "ok");
        closeModal();
        setStoredGroupId(null);
        location.hash = "#/";
        await render();
      } catch (err) {
        await noticeError("Cannot delete group", err);
      }
    });
  }

  async function openCreateGroupModal() {
    const currencies = await getCurrencies();
    const modal = openModal(`
      <div class="modal-head">
        <div>
          <h2>New group</h2>
          <p>Start a new split session.</p>
        </div>
        <button type="button" class="icon-btn" data-close title="Close" aria-label="Close">${icons.close}</button>
      </div>
      <form id="modal-create-group" class="row stack">
        <label>Name<input name="name" required placeholder="July trip" /></label>
        <label>Currency${currencyPickerHTML(currencies, "USD")}</label>
        <button type="submit">Create & open</button>
      </form>
    `);
    wireCurrencyPicker(modal);
    modal.querySelector("#modal-create-group").addEventListener("submit", async (e) => {
      e.preventDefault();
      const fd = new FormData(e.target);
      try {
        const { data } = await api("POST", "/groups", {
          name: fd.get("name"),
          currency: fd.get("currency") || "USD",
        });
        flash("Group created", "ok");
        closeModal();
        goGroup(data.id);
      } catch (err) {
        flash(err.message);
      }
    });
  }

  async function openCategoriesModal() {
    const categories = await getCategories(true);
    const modal = openModal(
      `
      <div class="modal-head">
        <div>
          <h2>Categories</h2>
          <p>Labels for charges. Built-ins are fixed; add your own anytime.</p>
        </div>
        <button type="button" class="icon-btn" data-close title="Close" aria-label="Close">${icons.close}</button>
      </div>
      <form id="modal-add-category" class="category-form">
        <label>Name<input name="name" required placeholder="Pets" maxlength="40" /></label>
        <div class="field">
          <span class="field-label">Icon</span>
          <div class="icon-picker" data-icon-picker>
            <input type="hidden" name="icon" value="fa-solid fa-paw" required />
            <div class="icon-picker-current">
              <span class="icon-picker-preview">${categoryIconHTML("fa-solid fa-paw")}</span>
              <input type="search" class="icon-picker-filter" placeholder="Search all Free solid icons…" autocomplete="off" />
            </div>
            <div class="icon-picker-grid" data-icon-grid></div>
          </div>
        </div>
        <button type="submit">${icons.plus} Add category</button>
      </form>
      ${
        categories.length === 0
          ? `<p class="empty">No categories yet.</p>`
          : `<ul class="modal-list category-list">
              ${categories
                .map(
                  (c) => `
                <li class="category-row">
                  <div class="category-info">
                    <span class="category-badge">${categoryIconHTML(c.icon)}</span>
                    <div>
                      <strong>${esc(c.name)}</strong>
                      <div class="meta">${c.builtin ? "Built-in" : "Custom"}</div>
                    </div>
                  </div>
                  <div class="icon-bar">
                    <button type="button" class="icon-btn" data-edit-category="${c.id}" ${
                      c.builtin ? 'aria-disabled="true"' : ""
                    } title="${
                      c.builtin
                        ? "Built-in categories cannot be edited"
                        : "Edit category"
                    }" aria-label="${
                      c.builtin
                        ? "Built-in categories cannot be edited"
                        : "Edit " + esc(c.name)
                    }">${icons.edit}</button>
                    <button type="button" class="icon-btn danger" data-del-category="${c.id}" ${
                      c.builtin ? 'aria-disabled="true"' : ""
                    } title="${
                      c.builtin
                        ? "Built-in categories cannot be deleted"
                        : "Delete category"
                    }" aria-label="${
                      c.builtin
                        ? "Built-in categories cannot be deleted"
                        : "Delete " + esc(c.name)
                    }">${icons.trash}</button>
                  </div>
                </li>`
                )
                .join("")}
            </ul>`
      }
    `,
      true
    );

    const byId = Object.fromEntries(categories.map((c) => [String(c.id), c]));
    await wireIconPicker(modal, "fa-solid fa-paw");

    modal.querySelector("#modal-add-category").addEventListener("submit", async (e) => {
      e.preventDefault();
      const fd = new FormData(e.target);
      try {
        await api("POST", "/categories", {
          name: fd.get("name"),
          icon: fd.get("icon"),
        });
        categoriesCache = null;
        flash("Category added", "ok");
        closeModal();
        await openCategoriesModal();
      } catch (err) {
        flash(err.message);
      }
    });

    modal.querySelectorAll("[data-edit-category]").forEach((btn) => {
      btn.addEventListener("click", async () => {
        if (btn.getAttribute("aria-disabled") === "true") {
          await askNotice(
            "Cannot edit category",
            `<p>Built-in categories cannot be edited.</p>`
          );
          return;
        }
        const cat = byId[btn.dataset.editCategory];
        if (!cat) return;
        openEditCategoryModal(cat).catch((err) => flash(err.message));
      });
    });

    modal.querySelectorAll("[data-del-category]").forEach((btn) => {
      btn.addEventListener("click", async () => {
        if (btn.getAttribute("aria-disabled") === "true") {
          await askNotice(
            "Cannot delete category",
            `<p>Built-in categories cannot be deleted.</p>`
          );
          return;
        }
        const cat = byId[btn.dataset.delCategory];
        const name = cat?.name || "this category";
        if (!(await askConfirm("Delete category", `Delete “${name}”?`))) return;
        try {
          await api("DELETE", "/categories/" + btn.dataset.delCategory);
          categoriesCache = null;
          flash("Category deleted", "ok");
          closeModal();
          await openCategoriesModal();
        } catch (err) {
          await noticeError("Cannot delete category", err);
        }
      });
    });
  }

  async function getFaSolidIcons() {
    if (faSolidIconsCache) return faSolidIconsCache;
    const res = await fetch("/app/js/fa-solid-icons.json");
    if (!res.ok) throw new Error("Could not load icon library");
    faSolidIconsCache = await res.json();
    return faSolidIconsCache;
  }

  async function wireIconPicker(scope, initialIcon) {
    const picker = scope.querySelector("[data-icon-picker]");
    if (!picker) return;
    const hidden = picker.querySelector('input[name="icon"]');
    const preview = picker.querySelector(".icon-picker-preview");
    const filter = picker.querySelector(".icon-picker-filter");
    const grid = picker.querySelector("[data-icon-grid]");
    let current = initialIcon || "fa-solid fa-ellipsis";
    let catalog = SUGGESTED_ICONS;

    const setIcon = (icon) => {
      current = icon;
      hidden.value = icon;
      preview.innerHTML = categoryIconHTML(icon);
      grid.querySelectorAll("[data-icon]").forEach((el) => {
        el.classList.toggle("is-selected", el.dataset.icon === icon);
      });
    };

    const renderGrid = (q) => {
      const query = String(q || "")
        .trim()
        .toLowerCase();
      let items;
      if (!query) {
        items = SUGGESTED_ICONS.slice();
        if (current && !items.some((i) => i.icon === current)) {
          items.unshift({ icon: current, labels: current });
        }
      } else {
        items = catalog
          .filter(
            (item) =>
              item.labels.includes(query) ||
              item.icon.toLowerCase().includes(query)
          )
          .slice(0, 96);
      }
      if (items.length === 0) {
        grid.innerHTML = `<p class="icon-picker-empty">No icons match “${esc(query)}”.</p>`;
        return;
      }
      grid.innerHTML = items
        .map(
          (item) => `
          <button type="button" class="icon-picker-item${
            item.icon === current ? " is-selected" : ""
          }" data-icon="${esc(item.icon)}" title="${esc(item.labels)}">
            ${categoryIconHTML(item.icon)}
          </button>`
        )
        .join("");
      grid.querySelectorAll("[data-icon]").forEach((btn) => {
        btn.addEventListener("click", () => setIcon(btn.dataset.icon));
      });
    };

    filter.addEventListener("input", () => renderGrid(filter.value));
    renderGrid("");
    setIcon(current);

    try {
      catalog = await getFaSolidIcons();
      if (filter.value.trim()) renderGrid(filter.value);
    } catch (_) {
      // Fall back to suggested icons only.
    }
  }

  async function openEditCategoryModal(cat) {
    if (cat.builtin) {
      await askNotice(
        "Cannot edit category",
        `<p>Built-in categories cannot be edited.</p>`
      );
      return;
    }
    const modal = openModal(`
      <div class="modal-head">
        <div>
          <h2>Edit category</h2>
          <p>Update name or icon.</p>
        </div>
        <button type="button" class="icon-btn" data-close title="Close" aria-label="Close">${icons.close}</button>
      </div>
      <form id="modal-edit-category" class="category-form">
        <label>Name<input name="name" required value="${esc(cat.name)}" maxlength="40" /></label>
        <div class="field">
          <span class="field-label">Icon</span>
          <div class="icon-picker" data-icon-picker>
            <input type="hidden" name="icon" value="${esc(cat.icon)}" required />
            <div class="icon-picker-current">
              <span class="icon-picker-preview">${categoryIconHTML(cat.icon)}</span>
              <input type="search" class="icon-picker-filter" placeholder="Search all Free solid icons…" autocomplete="off" />
            </div>
            <div class="icon-picker-grid" data-icon-grid></div>
          </div>
        </div>
        <button type="submit">Save changes</button>
      </form>
    `);
    await wireIconPicker(modal, cat.icon);
    modal.querySelector("#modal-edit-category").addEventListener("submit", async (e) => {
      e.preventDefault();
      const fd = new FormData(e.target);
      try {
        await api("PUT", "/categories/" + cat.id, {
          name: fd.get("name"),
          icon: fd.get("icon"),
        });
        categoriesCache = null;
        flash("Category updated", "ok");
        closeModal();
        await openCategoriesModal();
      } catch (err) {
        flash(err.message);
      }
    });
  }

  async function openPeopleModal() {
    const { data: people } = await api("GET", "/people");
    const modal = openModal(`
      <div class="modal-head">
        <div>
          <h2>People</h2>
          <p>Global contacts for any group.</p>
        </div>
        <button type="button" class="icon-btn" data-close title="Close" aria-label="Close">${icons.close}</button>
      </div>
      <form id="modal-add-person" class="row stack">
        <label>Name<input name="name" required placeholder="Alex" /></label>
        <label>Email<input name="email" type="email" placeholder="alex@example.com" /></label>
        <button type="submit">Add person</button>
      </form>
      ${
        people.length === 0
          ? `<p class="empty">No people yet.</p>`
          : `<ul class="modal-list" style="margin-top:0.85rem">
              ${people
                .map(
                  (p) => `
                <li class="person-row">
                  <div class="member-info">
                    <strong>${esc(p.name)}</strong>
                    <div class="meta">${esc(p.email || "—")}</div>
                  </div>
                  <div class="icon-bar">
                    <button type="button" class="icon-btn" data-edit-person="${p.id}" title="Edit person" aria-label="Edit ${esc(p.name)}">${icons.edit}</button>
                    <button type="button" class="icon-btn danger" data-del-person="${p.id}" title="Delete person" aria-label="Delete ${esc(p.name)}">${icons.trash}</button>
                  </div>
                </li>`
                )
                .join("")}
            </ul>`
      }
    `);

    const byId = Object.fromEntries(people.map((p) => [String(p.id), p]));

    modal.querySelector("#modal-add-person").addEventListener("submit", async (e) => {
      e.preventDefault();
      const fd = new FormData(e.target);
      try {
        await api("POST", "/people", {
          name: fd.get("name"),
          email: fd.get("email") || "",
        });
        flash("Person added", "ok");
        closeModal();
        await openPeopleModal();
      } catch (err) {
        flash(err.message);
      }
    });

    modal.querySelectorAll("[data-edit-person]").forEach((btn) => {
      btn.addEventListener("click", () => {
        const person = byId[btn.dataset.editPerson];
        if (!person) return;
        openEditPersonModal(person).catch((err) => flash(err.message));
      });
    });

    modal.querySelectorAll("[data-del-person]").forEach((btn) => {
      btn.addEventListener("click", async () => {
        const person = byId[btn.dataset.delPerson];
        const name = person?.name || "this person";
        if (!(await askConfirm("Delete person", `Delete “${name}”? This cannot be undone.`))) {
          return;
        }
        try {
          await api("DELETE", "/people/" + btn.dataset.delPerson);
          flash("Person deleted", "ok");
          closeModal();
          await openPeopleModal();
          if (activeGroupId) await renderWorkspace(activeGroupId);
        } catch (err) {
          await noticeError("Cannot delete person", err);
        }
      });
    });
  }

  async function openEditPersonModal(person) {
    const modal = openModal(`
      <div class="modal-head">
        <div>
          <h2>Edit person</h2>
          <p>Update name or email.</p>
        </div>
        <button type="button" class="icon-btn" data-close title="Close" aria-label="Close">${icons.close}</button>
      </div>
      <form id="modal-edit-person" class="row stack">
        <label>Name<input name="name" required value="${esc(person.name)}" /></label>
        <label>Email<input name="email" type="email" value="${esc(person.email || "")}" placeholder="alex@example.com" /></label>
        <button type="submit">Save changes</button>
      </form>
    `);

    modal.querySelector("#modal-edit-person").addEventListener("submit", async (e) => {
      e.preventDefault();
      const fd = new FormData(e.target);
      try {
        await api("PUT", "/people/" + person.id, {
          name: fd.get("name"),
          email: fd.get("email") || "",
        });
        flash("Person updated", "ok");
        closeModal();
        await openPeopleModal();
        if (activeGroupId) await renderWorkspace(activeGroupId);
      } catch (err) {
        flash(err.message);
      }
    });
  }

  async function openMembersModal(groupId) {
    const [{ data: members }, { data: people }, { data: charges }] = await Promise.all([
      api("GET", "/groups/" + groupId + "/members"),
      api("GET", "/people"),
      api("GET", "/groups/" + groupId + "/charges"),
    ]);
    const memberIds = new Set(members.map((m) => m.id));
    const nonMembers = people.filter((p) => !memberIds.has(p.id));
    const groupCharges = charges || [];

    const modal = openModal(`
      <div class="modal-head">
        <div>
          <h2>Members</h2>
          <p>People in this group.</p>
        </div>
        <button type="button" class="icon-btn" data-close title="Close" aria-label="Close">${icons.close}</button>
      </div>
      ${
        members.length === 0
          ? `<p class="empty">No members yet.</p>`
          : `<ul class="modal-list">
              ${members
                .map((m) => {
                  const linked = chargesForUser(groupCharges, m.id);
                  const n = linked.length;
                  const blocked = n > 0;
                  const rmTitle = blocked
                    ? `Remove member — linked to ${chargeCountLabel(n)}`
                    : "Remove member";
                  return `
                <li class="member-row">
                  <div class="member-info">
                    <strong>${esc(m.name)}</strong>
                    <div class="meta">${esc(m.email || "—")}</div>
                  </div>
                  <button type="button" class="member-charges${
                    blocked ? " has-charges" : ""
                  }" data-member-charges="${m.id}" title="${
                    blocked
                      ? "View associated charges"
                      : "No associated charges"
                  }">${esc(chargeCountLabel(n))}</button>
                  <button type="button" class="icon-btn danger" data-rm-member="${m.id}" title="${esc(
                    rmTitle
                  )}" aria-label="${esc(rmTitle)}">${icons.trash}</button>
                </li>`;
                })
                .join("")}
            </ul>`
      }
      <form id="modal-add-member" class="row stack" style="margin-top:0.9rem">
        <label>Add person
          <select name="userId" required ${nonMembers.length ? "" : "disabled"}>
            <option value="">${
              nonMembers.length ? "Select…" : "Create people first"
            }</option>
            ${nonMembers
              .map((p) => `<option value="${p.id}">${esc(p.name)}</option>`)
              .join("")}
          </select>
        </label>
        <button type="submit" ${nonMembers.length ? "" : "disabled"}>Add to group</button>
      </form>
    `);

    modal.querySelector("#modal-add-member").addEventListener("submit", async (e) => {
      e.preventDefault();
      const fd = new FormData(e.target);
      try {
        await api("POST", "/groups/" + groupId + "/members", {
          userId: Number(fd.get("userId")),
        });
        flash("Member added", "ok");
        closeModal();
        await openMembersModal(groupId);
        await renderWorkspace(groupId);
      } catch (err) {
        flash(err.message);
      }
    });

    modal.querySelectorAll("[data-member-charges]").forEach((btn) => {
      btn.addEventListener("click", async () => {
        const memberId = Number(btn.dataset.memberCharges);
        const name =
          btn.closest("li")?.querySelector("strong")?.textContent || "this member";
        await showMemberCharges(name, chargesForUser(groupCharges, memberId));
      });
    });

    modal.querySelectorAll("[data-rm-member]").forEach((btn) => {
      btn.addEventListener("click", async () => {
        const memberId = Number(btn.dataset.rmMember);
        const name =
          btn.closest("li")?.querySelector("strong")?.textContent || "this member";
        const blocking = chargesForUser(groupCharges, memberId);
        if (!(await confirmRemoveMember(name, blocking))) return;
        try {
          await api("DELETE", "/groups/" + groupId + "/members/" + memberId);
          flash("Member removed", "ok");
          closeModal();
          await openMembersModal(groupId);
          await renderWorkspace(groupId);
        } catch (err) {
          await noticeError("Cannot remove member", err);
        }
      });
    });
  }

  async function openChargeModal(groupId) {
    const [{ data: members }, categories] = await Promise.all([
      api("GET", "/groups/" + groupId + "/members"),
      getCategories(),
    ]);
    const defaultCat = defaultCategoryId(categories);
    const modal = openModal(
      `
      <div class="modal-head">
        <div>
          <h2>New charge</h2>
          <p>Record a shared expense.</p>
        </div>
        <button type="button" class="icon-btn" data-close title="Close" aria-label="Close">${icons.close}</button>
      </div>
      <form id="modal-add-charge" class="charge-form-split">
        <div class="charge-form-side charge-form-participants">
          <span class="field-label">Who shares</span>
          ${participantPickerHTML(members)}
        </div>
        <div class="charge-form-side charge-form-details">
          <label>Description<input name="description" required placeholder="Dinner" /></label>
          <label>Amount<input name="amount" type="number" min="0.01" step="0.01" required /></label>
          <label>Date<input name="date" type="date" required value="${todayISO()}" /></label>
          ${categorySelectHTML(categories, defaultCat)}
          <p class="field-hint">Use Paid? on any member (even if they Skip). Use Share/Skip for who splits the cost.</p>
          <button type="submit" ${members.length ? "" : "disabled"}>Save charge</button>
        </div>
      </form>
    `,
      true
    );

    const participants = wireParticipantPicker(
      modal,
      members,
      members.map((m) => m.id),
      members[0] ? members[0].id : null
    );
    wireCategorySelect(modal);

    modal.querySelector("#modal-add-charge").addEventListener("submit", async (e) => {
      e.preventDefault();
      const fd = new FormData(e.target);
      const participantIds = participants ? participants.getIds() : [];
      const paidByUserId = participants ? participants.getPayerId() : null;
      const categoryId = Number(fd.get("categoryId")) || null;
      if (!participantIds.length) {
        flash("Add at least one participant");
        return;
      }
      if (!paidByUserId) {
        flash("Choose who paid with Paid?");
        return;
      }
      try {
        await api("POST", "/groups/" + groupId + "/charges", {
          description: fd.get("description"),
          amount: Number(fd.get("amount")),
          date: fd.get("date") || todayISO(),
          paidByUserId,
          categoryId,
          participantIds,
          splitRule: "equal",
        });
        flash("Charge added", "ok");
        closeModal();
        await renderWorkspace(groupId);
      } catch (err) {
        flash(err.message);
      }
    });
  }

  async function openEditChargeModal(groupId, charge, currency) {
    const [{ data: members }, categories] = await Promise.all([
      api("GET", "/groups/" + groupId + "/members"),
      getCategories(),
    ]);
    const selected = (charge.participantIds || []).map(Number);
    const selectedCat =
      charge.categoryId != null
        ? Number(charge.categoryId)
        : defaultCategoryId(categories);
    const modal = openModal(
      `
      <div class="modal-head">
        <div>
          <h2>Edit charge</h2>
          <p>Update who shares and the charge details.</p>
        </div>
        <div class="modal-head-actions">
          <button type="button" class="icon-btn danger" id="btn-delete-charge" title="Delete charge" aria-label="Delete charge">${icons.trash}</button>
          <button type="button" class="icon-btn" data-close title="Close" aria-label="Close">${icons.close}</button>
        </div>
      </div>
      <form id="modal-edit-charge" class="charge-form-split">
        <div class="charge-form-side charge-form-participants">
          <span class="field-label">Who shares</span>
          ${participantPickerHTML(members)}
        </div>
        <div class="charge-form-side charge-form-details">
          <label>Description<input name="description" required value="${esc(charge.description)}" /></label>
          <label>Amount (${esc(currency)})
            <input name="amount" type="number" min="0.01" step="0.01" required value="${esc(charge.amount)}" />
          </label>
          <label>Date
            <input name="date" type="date" required value="${esc(charge.date || todayISO())}" />
          </label>
          ${categorySelectHTML(categories, selectedCat)}
          <p class="field-hint">Use Paid? on any member (even if they Skip). Use Share/Skip for who splits the cost.</p>
          <button type="submit">Save changes</button>
        </div>
      </form>
    `,
      true
    );

    const participants = wireParticipantPicker(
      modal,
      members,
      selected,
      charge.paidByUserId
    );
    wireCategorySelect(modal);

    modal.querySelector("#modal-edit-charge").addEventListener("submit", async (e) => {
      e.preventDefault();
      const fd = new FormData(e.target);
      const participantIds = participants ? participants.getIds() : [];
      const paidByUserId = participants ? participants.getPayerId() : null;
      const categoryId = Number(fd.get("categoryId")) || null;
      if (!participantIds.length) {
        flash("Add at least one participant");
        return;
      }
      if (!paidByUserId) {
        flash("Choose who paid with Paid?");
        return;
      }
      try {
        await api("PUT", "/groups/" + groupId + "/charges/" + charge.id, {
          description: fd.get("description"),
          amount: Number(fd.get("amount")),
          date: fd.get("date") || todayISO(),
          paidByUserId,
          categoryId,
          participantIds,
          splitRule: "equal",
        });
        flash("Charge updated", "ok");
        closeModal();
        await renderWorkspace(groupId);
      } catch (err) {
        flash(err.message);
      }
    });

    modal.querySelector("#btn-delete-charge").addEventListener("click", async () => {
      if (
        !(await askConfirm(
          "Delete charge",
          `Delete “${charge.description}”? This cannot be undone.`
        ))
      ) {
        return;
      }
      try {
        await api("DELETE", "/groups/" + groupId + "/charges/" + charge.id);
        flash("Charge deleted", "ok");
        closeModal();
        await renderWorkspace(groupId);
      } catch (err) {
        await noticeError("Cannot delete charge", err);
      }
    });
  }

  async function openEditGroupModal(group) {
    const currencies = await getCurrencies();
    const modal = openModal(`
      <div class="modal-head">
        <div>
          <h2>Edit group</h2>
          <p>Update name or currency.</p>
        </div>
        <button type="button" class="icon-btn" data-close title="Close" aria-label="Close">${icons.close}</button>
      </div>
      <form id="modal-edit-group" class="row stack">
        <label>Name<input name="name" required value="${esc(group.name)}" /></label>
        <label>Currency${currencyPickerHTML(currencies, group.currency)}</label>
        <button type="submit">Save changes</button>
      </form>
    `);
    wireCurrencyPicker(modal);

    modal.querySelector("#modal-edit-group").addEventListener("submit", async (e) => {
      e.preventDefault();
      const fd = new FormData(e.target);
      try {
        await api("PUT", "/groups/" + group.id, {
          name: fd.get("name"),
          currency: fd.get("currency") || group.currency,
        });
        flash("Group updated", "ok");
        closeModal();
        await renderWorkspace(group.id);
      } catch (err) {
        flash(err.message);
      }
    });
  }

  function renderEmptyHome() {
    activeGroupId = null;
    root.innerHTML = `
      <div class="hero-empty">
        <div class="hero-empty-card">
          <h1>Split shared bills</h1>
          <p>Create a group to track charges and settle up.</p>
          <button type="button" id="empty-open-groups">Choose or create a group</button>
        </div>
      </div>
    `;
    document.getElementById("empty-open-groups").addEventListener("click", () => {
      openGroupModal(null);
    });
  }

  function pieSliceColor(tone, index, total) {
    if (tone === "category") {
      const hue = (index * 37 + 18) % 360;
      const t = total <= 1 ? 0.4 : index / Math.max(total - 1, 1);
      return `hsl(${hue} 52% ${38 + t * 22}%)`;
    }
    const hue = tone === "credit" ? 158 : 4;
    const sat = tone === "credit" ? 68 : 58;
    const t = total <= 1 ? 0.45 : index / Math.max(total - 1, 1);
    const light = 36 + t * 28;
    return `hsl(${hue} ${sat}% ${light}%)`;
  }

  function pieChartSVG(slices, tone, currency) {
    const usable = slices.filter((s) => s.amount > 0.005);
    if (usable.length === 0) {
      return `<div class="pie-empty">None</div>`;
    }

    const total = usable.reduce((sum, s) => sum + s.amount, 0);
    const cx = 50;
    const cy = 50;
    const r = 42;
    let angle = 0;
    const chartLabel =
      tone === "credit" ? "Credits" : tone === "debt" ? "Debts" : "Category spend";

    const parts = usable.map((s, i) => {
      const portion = (s.amount / total) * 360;
      const start = angle;
      const end = angle + portion;
      angle = end;
      const tip = `${s.name}: ${currency} ${money(s.amount)}`;
      const color = pieSliceColor(tone, i, usable.length);

      if (usable.length === 1 || portion >= 359.9) {
        return `<circle class="pie-slice" cx="${cx}" cy="${cy}" r="${r}" fill="${color}" data-tip="${esc(
          tip
        )}" tabindex="0"></circle>`;
      }

      const startRad = ((start - 90) * Math.PI) / 180;
      const endRad = ((end - 90) * Math.PI) / 180;
      const x1 = cx + r * Math.cos(startRad);
      const y1 = cy + r * Math.sin(startRad);
      const x2 = cx + r * Math.cos(endRad);
      const y2 = cy + r * Math.sin(endRad);
      const large = portion > 180 ? 1 : 0;
      const d = `M ${cx} ${cy} L ${x1} ${y1} A ${r} ${r} 0 ${large} 1 ${x2} ${y2} Z`;
      return `<path class="pie-slice" d="${d}" fill="${color}" data-tip="${esc(
        tip
      )}" tabindex="0"></path>`;
    });

    return `
      <svg class="pie-svg" viewBox="0 0 100 100" role="img" aria-label="${esc(
        chartLabel
      )} pie chart">
        ${parts.join("")}
      </svg>`;
  }

  function settlePiesHTML(balances, currency) {
    const credits = balances
      .filter((b) => Number(b.net) > 0.005)
      .map((b) => ({ name: b.name, amount: Number(b.net) }))
      .sort((a, b) => b.amount - a.amount);
    const debts = balances
      .filter((b) => Number(b.net) < -0.005)
      .map((b) => ({ name: b.name, amount: Math.abs(Number(b.net)) }))
      .sort((a, b) => b.amount - a.amount);

    return `
      <div class="settle-pies">
        <div class="pie-block credit">
          <div class="pie-label">To receive</div>
          ${pieChartSVG(credits, "credit", currency)}
        </div>
        <div class="pie-block debt">
          <div class="pie-label">To pay</div>
          ${pieChartSVG(debts, "debt", currency)}
        </div>
        <div class="pie-tip" hidden></div>
      </div>`;
  }

  function wireSettlePies(scope) {
    scope.querySelectorAll(".settle-pies, .category-spend").forEach((container) => {
      const tip = container.querySelector(".pie-tip");
      if (!tip) return;
      const show = (el, e) => {
        tip.hidden = false;
        tip.textContent = el.dataset.tip || "";
        const rect = container.getBoundingClientRect();
        const x = e.clientX - rect.left;
        const y = e.clientY - rect.top;
        tip.style.left = `${Math.min(Math.max(x + 12, 8), rect.width - 8)}px`;
        tip.style.top = `${Math.max(y - 28, 8)}px`;
      };
      const hide = () => {
        tip.hidden = true;
      };
      container.querySelectorAll(".pie-slice").forEach((slice) => {
        slice.addEventListener("mousemove", (e) => show(slice, e));
        slice.addEventListener("mouseenter", (e) => show(slice, e));
        slice.addEventListener("mouseleave", hide);
        slice.addEventListener("focus", () => {
          tip.hidden = false;
          tip.textContent = slice.dataset.tip || "";
          tip.style.left = "50%";
          tip.style.top = "8px";
        });
        slice.addEventListener("blur", hide);
      });
    });
  }

  function categorySpendHTML(breakdown, currency) {
    const rows = (breakdown || [])
      .map((r) => ({
        name: r.name || "Uncategorized",
        icon: r.icon || "fa-solid fa-ellipsis",
        amount: Number(r.amount) || 0,
        chargeCount: Number(r.chargeCount) || 0,
      }))
      .filter((r) => r.amount > 0.005);
    if (rows.length === 0) return "";

    const total = rows.reduce((sum, r) => sum + r.amount, 0);
    const slices = rows.map((r) => ({ name: r.name, amount: r.amount }));

    return `
      <div class="overview-card category-spend">
        <h3 class="settle-section-title">Category spend</h3>
        <div class="category-spend-layout">
          <div class="pie-block category">
            <div class="pie-label">By category</div>
            ${pieChartSVG(slices, "category", currency)}
          </div>
          <ul class="category-bars">
            ${rows
              .map((r, i) => {
                const pct = total > 0 ? (r.amount / total) * 100 : 0;
                const color = pieSliceColor("category", i, rows.length);
                return `
                  <li class="category-bar-row">
                    <div class="category-bar-label">
                      <span class="category-badge">${categoryIconHTML(r.icon)}</span>
                      <span class="category-bar-name">${esc(r.name)}</span>
                      <span class="category-bar-amt">${esc(currency)} ${money(r.amount)}</span>
                    </div>
                    <div class="category-bar-track" aria-hidden="true">
                      <div class="category-bar-fill" style="width:${pct.toFixed(
                        1
                      )}%;background:${color}"></div>
                    </div>
                    <div class="category-bar-meta">${pct.toFixed(0)}% · ${
                      r.chargeCount
                    } charge${r.chargeCount === 1 ? "" : "s"}</div>
                  </li>`;
              })
              .join("")}
          </ul>
        </div>
        <div class="pie-tip" hidden></div>
      </div>`;
  }

  const SPEND_VIEW_KEY = "billsplitter.spendView";

  function getSpendView() {
    const v = localStorage.getItem(SPEND_VIEW_KEY);
    return v === "month" || v === "heat" ? v : "week";
  }

  function setSpendView(v) {
    localStorage.setItem(
      SPEND_VIEW_KEY,
      v === "month" || v === "heat" ? v : "week"
    );
  }

  function parseISODate(iso) {
    const m = String(iso || "").match(/^(\d{4})-(\d{2})-(\d{2})$/);
    if (!m) return null;
    return new Date(Number(m[1]), Number(m[2]) - 1, Number(m[3]));
  }

  function toISODate(d) {
    const y = d.getFullYear();
    const m = String(d.getMonth() + 1).padStart(2, "0");
    const day = String(d.getDate()).padStart(2, "0");
    return `${y}-${m}-${day}`;
  }

  function addDays(d, n) {
    const x = new Date(d.getFullYear(), d.getMonth(), d.getDate());
    x.setDate(x.getDate() + n);
    return x;
  }

  function startOfWeekMonday(d) {
    const x = new Date(d.getFullYear(), d.getMonth(), d.getDate());
    const dow = (x.getDay() + 6) % 7; // Mon=0
    x.setDate(x.getDate() - dow);
    return x;
  }

  function dayAmountMap(timeline) {
    const map = Object.create(null);
    (timeline || []).forEach((r) => {
      const d = String(r.date || "").trim();
      if (!/^\d{4}-\d{2}-\d{2}$/.test(d)) return;
      map[d] = (map[d] || 0) + (Number(r.amount) || 0);
    });
    return map;
  }

  function float64round(n) {
    return Math.round((Number(n) || 0) * 100) / 100;
  }

  function heatLevel(amount, max) {
    if (max <= 0 || amount <= 0) return 0;
    const t = amount / max;
    if (t > 0.75) return 4;
    if (t > 0.5) return 3;
    if (t > 0.25) return 2;
    return 1;
  }

  function spendBarsHTML(rows, currency) {
    if (!rows.length) return `<p class="empty">None</p>`;
    const max = Math.max(...rows.map((r) => r.amount), 0.01);
    return `
      <ul class="spend-bars">
        ${rows
          .map((r) => {
            const pct = Math.max(r.amount > 0 ? 4 : 0, (r.amount / max) * 100);
            const tip = `${r.label}: ${currency} ${money(r.amount)}`;
            return `
            <li class="spend-bar-row" title="${esc(tip)}">
              <div class="spend-bar-head">
                <span class="spend-bar-label">${esc(r.short || r.label)}</span>
                <span class="spend-bar-amt">${esc(currency)} ${money(r.amount)}</span>
              </div>
              <div class="spend-bar-track" aria-hidden="true">
                <div class="spend-bar-fill" style="width:${pct.toFixed(1)}%"></div>
              </div>
            </li>`;
          })
          .join("")}
      </ul>`;
  }

  function weekSpendRows(byDay) {
    const today = parseISODate(todayISO());
    const endWeek = startOfWeekMonday(today);
    const weeks = [];
    for (let i = 7; i >= 0; i--) {
      const start = addDays(endWeek, -7 * i);
      const end = addDays(start, 6);
      let amount = 0;
      for (let d = 0; d < 7; d++) {
        amount += byDay[toISODate(addDays(start, d))] || 0;
      }
      const startKey = toISODate(start);
      const endKey = toISODate(end);
      const range = `${formatChargeDate(startKey)} – ${formatChargeDate(endKey)}`;
      weeks.push({
        label: range,
        short: range,
        amount: float64round(amount),
      });
    }
    return weeks;
  }

  function monthSpendRows(byDay) {
    const today = parseISODate(todayISO());
    const names = [
      "Jan",
      "Feb",
      "Mar",
      "Apr",
      "May",
      "Jun",
      "Jul",
      "Aug",
      "Sep",
      "Oct",
      "Nov",
      "Dec",
    ];
    const months = [];
    for (let i = 5; i >= 0; i--) {
      const first = new Date(today.getFullYear(), today.getMonth() - i, 1);
      const next = new Date(first.getFullYear(), first.getMonth() + 1, 1);
      let amount = 0;
      for (let d = new Date(first); d < next; d = addDays(d, 1)) {
        amount += byDay[toISODate(d)] || 0;
      }
      months.push({
        label: `${names[first.getMonth()]} ${first.getFullYear()}`,
        short: `${names[first.getMonth()]} ${String(first.getFullYear()).slice(2)}`,
        amount: float64round(amount),
      });
    }
    return months;
  }

  /** GitHub-style heat map for the last 180 days (fixed cell size). */
  function heat180DaysHTML(byDay, currency) {
    const today = parseISODate(todayISO());
    const windowStart = addDays(today, -179);
    const gridStart = startOfWeekMonday(windowStart);
    const gridEnd = addDays(startOfWeekMonday(today), 6);
    const weeks = [];
    for (let w = new Date(gridStart); w <= gridEnd; w = addDays(w, 7)) {
      weeks.push(new Date(w));
    }

    const amounts = [];
    for (let i = 179; i >= 0; i--) {
      amounts.push(byDay[toISODate(addDays(today, -i))] || 0);
    }
    const max = Math.max(...amounts, 0);
    const dowLabels = ["Mon", "", "Wed", "", "Fri", "", ""];

    const cells = [];
    weeks.forEach((weekStart) => {
      for (let dow = 0; dow < 7; dow++) {
        const day = addDays(weekStart, dow);
        const key = toISODate(day);
        const inWindow = day >= windowStart && day <= today;
        const amount = inWindow ? float64round(byDay[key] || 0) : 0;
        const level = inWindow ? heatLevel(amount, max) : -1;
        const tip = !inWindow
          ? ""
          : amount > 0
            ? `${formatChargeDate(key)}: ${currency} ${money(amount)}`
            : `${formatChargeDate(key)}: no spend`;
        cells.push(
          `<button type="button" class="heat-cell${
            level < 0 ? " is-out" : ` level-${level}`
          }" ${
            tip
              ? `data-tip="${esc(tip)}" aria-label="${esc(tip)}"`
              : 'tabindex="-1" aria-hidden="true"'
          }></button>`
        );
      }
    });

    const monthMarks = weeks
      .map((weekStart, i) => {
        const prev = i === 0 ? null : addDays(weekStart, -7);
        const show = !prev || weekStart.getMonth() !== prev.getMonth();
        if (!show) return `<span class="gh-heat-month"></span>`;
        const names = [
          "Jan",
          "Feb",
          "Mar",
          "Apr",
          "May",
          "Jun",
          "Jul",
          "Aug",
          "Sep",
          "Oct",
          "Nov",
          "Dec",
        ];
        return `<span class="gh-heat-month">${names[weekStart.getMonth()]}</span>`;
      })
      .join("");

    return `
      <div class="gh-heat">
        <div class="gh-heat-months" style="--weeks:${weeks.length}">${monthMarks}</div>
        <div class="gh-heat-body">
          <div class="gh-heat-dows">${dowLabels
            .map((l) => `<span>${esc(l)}</span>`)
            .join("")}</div>
          <div class="gh-heat-cells" style="--weeks:${weeks.length}">${cells.join(
            ""
          )}</div>
        </div>
        <div class="heat-legend">
          <span>Less</span>
          <span class="heat-swatch level-0" aria-hidden="true"></span>
          <span class="heat-swatch level-1" aria-hidden="true"></span>
          <span class="heat-swatch level-2" aria-hidden="true"></span>
          <span class="heat-swatch level-3" aria-hidden="true"></span>
          <span class="heat-swatch level-4" aria-hidden="true"></span>
          <span>More</span>
        </div>
      </div>`;
  }

  function spendOverTimeHTML(timeline, currency) {
    const byDay = dayAmountMap(timeline);
    const hasAny = Object.keys(byDay).some((k) => byDay[k] > 0.005);
    if (!hasAny) {
      return `<div class="overview-card spend-timeline"><h3 class="settle-section-title">Spend over time</h3><p class="empty">No dated spend yet.</p></div>`;
    }

    const view = getSpendView();
    const weeks = weekSpendRows(byDay);
    const months = monthSpendRows(byDay);
    const weekTotal = weeks.reduce((s, r) => s + r.amount, 0);
    const monthTotal = months.reduce((s, r) => s + r.amount, 0);
    let heatTotal = 0;
    {
      const today = parseISODate(todayISO());
      for (let i = 179; i >= 0; i--) {
        heatTotal += byDay[toISODate(addDays(today, -i))] || 0;
      }
    }
    heatTotal = float64round(heatTotal);

    return `
      <div class="overview-card spend-timeline" data-spend-timeline>
        <div class="spend-timeline-head">
          <div>
            <h3 class="settle-section-title">Spend over time</h3>
            <p class="pane-sub spend-view-sub" data-spend-sub></p>
          </div>
          <div class="spend-view-tabs" role="tablist" aria-label="Spend time range">
            <button type="button" class="spend-view-tab${
              view === "week" ? " is-active" : ""
            }" data-spend-view="week" role="tab" aria-selected="${
              view === "week"
            }">Week</button>
            <button type="button" class="spend-view-tab${
              view === "month" ? " is-active" : ""
            }" data-spend-view="month" role="tab" aria-selected="${
              view === "month"
            }">Month</button>
            <button type="button" class="spend-view-tab${
              view === "heat" ? " is-active" : ""
            }" data-spend-view="heat" role="tab" aria-selected="${
              view === "heat"
            }">180-day</button>
          </div>
        </div>
        <div class="spend-view${view === "week" ? " is-active" : ""}" data-spend-panel="week" data-sub="Last 8 weeks · ${esc(
          currency
        )} ${money(weekTotal)}" ${view === "week" ? "" : "hidden"}>
          <div class="timeline-chart">${spendBarsHTML(weeks, currency)}</div>
        </div>
        <div class="spend-view${view === "month" ? " is-active" : ""}" data-spend-panel="month" data-sub="Last 6 months · ${esc(
          currency
        )} ${money(monthTotal)}" ${view === "month" ? "" : "hidden"}>
          <div class="timeline-chart">${spendBarsHTML(months, currency)}</div>
        </div>
        <div class="spend-view${view === "heat" ? " is-active" : ""}" data-spend-panel="heat" data-sub="Last 180 days · ${esc(
          currency
        )} ${money(heatTotal)}" ${view === "heat" ? "" : "hidden"}>
          <div class="timeline-chart heat-chart">
            ${heat180DaysHTML(byDay, currency)}
            <div class="pie-tip" hidden></div>
          </div>
        </div>
      </div>`;
  }

  function wireSpendTimeline(scope) {
    const root = scope.querySelector("[data-spend-timeline]");
    if (!root) return;
    const sub = root.querySelector("[data-spend-sub]");

    const syncSub = () => {
      const active = root.querySelector(".spend-view.is-active, .spend-view:not([hidden])");
      if (sub && active) sub.textContent = active.dataset.sub || "";
    };
    syncSub();

    root.querySelectorAll("[data-spend-view]").forEach((btn) => {
      btn.addEventListener("click", () => {
        const id = btn.dataset.spendView;
        setSpendView(id);
        root.querySelectorAll("[data-spend-view]").forEach((b) => {
          const on = b.dataset.spendView === id;
          b.classList.toggle("is-active", on);
          b.setAttribute("aria-selected", on ? "true" : "false");
        });
        root.querySelectorAll("[data-spend-panel]").forEach((panel) => {
          const on = panel.dataset.spendPanel === id;
          panel.classList.toggle("is-active", on);
          panel.hidden = !on;
        });
        syncSub();
      });
    });

    const heatChart = root.querySelector(".heat-chart");
    if (heatChart) {
      const tip = heatChart.querySelector(".pie-tip");
      if (tip) {
        const show = (el, e) => {
          if (!el.dataset.tip) return;
          tip.hidden = false;
          tip.textContent = el.dataset.tip;
          const rect = heatChart.getBoundingClientRect();
          tip.style.left = `${Math.min(
            Math.max(e.clientX - rect.left + 10, 8),
            rect.width - 8
          )}px`;
          tip.style.top = `${Math.max(e.clientY - rect.top - 28, 8)}px`;
        };
        const hide = () => {
          tip.hidden = true;
        };
        heatChart.querySelectorAll(".heat-cell[data-tip]").forEach((el) => {
          el.addEventListener("mousemove", (e) => show(el, e));
          el.addEventListener("mouseenter", (e) => show(el, e));
          el.addEventListener("mouseleave", hide);
        });
      }
    }
  }

  function balanceRankingHTML(balances, currency) {
    const rows = [...(balances || [])]
      .map((b) => ({
        name: b.name || "#" + b.userId,
        net: Number(b.net) || 0,
      }))
      .sort((a, b) => b.net - a.net || a.name.localeCompare(b.name));
    if (rows.length === 0) return "";

    const maxAbs = Math.max(...rows.map((r) => Math.abs(r.net)), 0.01);

    return `
      <div class="overview-card balance-rank">
        <h3 class="settle-section-title">Balance summary</h3>
        <p class="pane-sub" style="margin:0 0 0.55rem">Current net for each member</p>
        <div class="rank-axis-legend" aria-hidden="true">
          <span>To pay</span>
          <span>0</span>
          <span>To receive</span>
        </div>
        <ul class="rank-list">
          ${rows
            .map((r) => {
              const tone =
                r.net > 0.005 ? "credit" : r.net < -0.005 ? "debt" : "even";
              const pct = Math.min(
                100,
                Math.max(r.net === 0 ? 0 : 4, (Math.abs(r.net) / maxAbs) * 100)
              );
              return `
              <li class="rank-row ${tone}">
                <div class="rank-line">
                  <strong class="rank-name">${esc(r.name)}</strong>
                  <span class="rank-net">${esc(currency)} ${signedMoney(
                    r.net
                  )}</span>
                </div>
                <div class="rank-axis" aria-hidden="true">
                  <div class="rank-half left">
                    ${
                      tone === "debt"
                        ? `<div class="rank-fill" style="width:${pct.toFixed(
                            1
                          )}%"></div>`
                        : ""
                    }
                  </div>
                  <div class="rank-zero"></div>
                  <div class="rank-half right">
                    ${
                      tone === "credit"
                        ? `<div class="rank-fill" style="width:${pct.toFixed(
                            1
                          )}%"></div>`
                        : ""
                    }
                  </div>
                </div>
              </li>`;
            })
            .join("")}
        </ul>
      </div>`;
  }

  function overviewPanelBodyHTML(settle, currency) {
    if (!settle) {
      return `<p class="empty">Could not load overview.</p>`;
    }
    const cats = categorySpendHTML(settle.categoryBreakdown, currency);
    const time = spendOverTimeHTML(settle.spendOverTime, currency);
    if (!cats && !(settle.spendOverTime || []).length) {
      return `<p class="empty">Add charges to see spend overview.</p>`;
    }
    return `
      ${time}
      ${cats || ""}
    `;
  }

  function settlementPanelBodyHTML(settle, currency) {
    if (!settle) {
      return `<p class="empty">Could not calculate settlement.</p>`;
    }
    if (!settle.balances || settle.balances.length === 0) {
      return `<p class="empty">Add members to see balances.</p>`;
    }

    const transfers = settle.transfers || [];
    const debtors = settle.balances.filter((b) => Number(b.net) < -0.005);
    const summary = balanceRankingHTML(settle.balances, currency);

    return `
      ${summary}
      ${settlePiesHTML(settle.balances, currency)}
      ${
        debtors.length === 0
          ? `<p class="empty">Everyone is settled up.</p>`
          : `<ul class="settle-people">
        ${debtors
          .map((b) => {
            const net = Number(b.net) || 0;
            const related = transfers.filter(
              (t) => Number(t.fromUserId) === Number(b.userId)
            );
            return `
              <li class="person-card debt">
                <div class="person-top">
                  <strong>${esc(b.name)}</strong>
                  <span class="person-net">${money(net)}</span>
                </div>
                ${
                  related.length === 0
                    ? `<p class="person-xfer-empty">No transfers</p>`
                    : `<ul class="person-xfers">
                        ${related
                          .map(
                            (t) => `
                              <li class="xfer-row out">
                                <span class="xfer-from">${esc(t.fromName)}</span>
                                <i class="fa-solid fa-arrow-right xfer-arrow" aria-hidden="true"></i>
                                <span class="xfer-to">${esc(t.toName)}</span>
                                <span class="xfer-amt">${esc(currency)} ${money(t.amount)}</span>
                              </li>`
                          )
                          .join("")}
                      </ul>`
                }
              </li>`;
          })
          .join("")}
      </ul>`
      }
    `;
  }

  async function renderWorkspace(groupId) {
    const [
      { data: group },
      { data: members },
      { data: people },
      { data: charges },
      currencies,
      settleResult,
    ] = await Promise.all([
      api("GET", "/groups/" + groupId),
      api("GET", "/groups/" + groupId + "/members"),
      api("GET", "/people"),
      api("GET", "/groups/" + groupId + "/charges"),
      getCurrencies(),
      api("POST", "/groups/" + groupId + "/settle", {}).catch(() => null),
    ]);

    activeGroupId = group.id;
    setStoredGroupId(group.id);
    const currencyMeta = currencies.find((c) => c.code === group.currency);
    const nameById = Object.fromEntries(people.map((p) => [p.id, p.name]));
    const sortMode = getChargeSort();
    const sortedCharges = sortCharges(charges, sortMode);
    const settle = settleResult ? settleResult.data : null;

    const sortSelect = `
      <label class="charge-sort">
        <span class="sr-only">Sort charges</span>
        <select id="charge-sort" aria-label="Sort charges">
          <option value="date-desc" ${sortMode === "date-desc" ? "selected" : ""}>Date ↓</option>
          <option value="date-asc" ${sortMode === "date-asc" ? "selected" : ""}>Date ↑</option>
          <option value="amount-desc" ${sortMode === "amount-desc" ? "selected" : ""}>Amount ↓</option>
          <option value="amount-asc" ${sortMode === "amount-asc" ? "selected" : ""}>Amount ↑</option>
          <option value="name-asc" ${sortMode === "name-asc" ? "selected" : ""}>Name A–Z</option>
          <option value="name-desc" ${sortMode === "name-desc" ? "selected" : ""}>Name Z–A</option>
        </select>
      </label>`;

    root.innerHTML = `
      <div class="page-head">
        <div class="page-head-main">
          <div class="title-row">
            <h1 class="page-title" id="btn-switch-title" tabindex="0" role="button" title="Switch group">${esc(group.name)}</h1>
            <div class="icon-bar title-actions">
              <button type="button" class="icon-btn" id="btn-switch" title="Switch group" aria-label="Switch group">${icons.switch}</button>
            </div>
          </div>
          <p class="page-meta">
            ${
              currencyMeta
                ? `${flagImg(currencyMeta.country, 16)} <span>${esc(currencyMeta.code)}</span>`
                : esc(group.currency)
            }
            <span>·</span>
            <button type="button" class="text-link" id="btn-open-members">
              ${members.length} member${members.length === 1 ? "" : "s"}
            </button>
          </p>
        </div>
      </div>

      <div class="layout">
        <section class="pane">
          <div class="pane-head">
            <div>
              <h2>Charges</h2>
              <p class="pane-sub">${sortedCharges.length} transaction${
                sortedCharges.length === 1 ? "" : "s"
              }</p>
            </div>
            <div class="pane-head-tools">
              ${sortedCharges.length ? sortSelect : ""}
              <div class="icon-bar">
                <button type="button" class="icon-btn primary" id="btn-new-charge" title="New charge" aria-label="New charge" ${
                  members.length ? "" : "disabled"
                }>${icons.plus}</button>
              </div>
            </div>
          </div>
          <div class="pane-body">
            ${
              sortedCharges.length === 0
                ? `<p class="empty">No charges yet. Tap + to add one.</p>`
                : `<ul class="tx-list">
                    ${sortedCharges
                      .map((c) => {
                        const nets = chargePersonNetsHTML(c, members, nameById);
                        return `
                        <li class="tx-item">
                          <div class="tx-main">
                            <button type="button" class="tx-title" data-edit-charge="${c.id}" title="Edit charge">
                              <span class="tx-cat" title="${esc(
                                c.categoryName || "Uncategorized"
                              )}">${categoryIconHTML(
                                c.categoryIcon || "fa-solid fa-ellipsis"
                              )}</span>
                              <span class="tx-title-text">${esc(c.description)}</span>
                            </button>
                            <div class="tx-amount">${esc(group.currency)} ${money(c.amount)}</div>
                            <div class="tx-meta-left">
                              <span class="tx-date">${esc(formatChargeDate(c.date))}</span>
                              · Paid by ${esc(
                                nameById[c.paidByUserId] || "#" + c.paidByUserId
                              )} · share ${money(c.sharePerPerson)}
                            </div>
                            <div class="tx-meta-nets">${nets}</div>
                          </div>
                        </li>`;
                      })
                      .join("")}
                  </ul>`
            }
          </div>
        </section>

        ${sideTabsHTML(settle, group.currency)}
      </div>
    `;

    const chargeById = Object.fromEntries(sortedCharges.map((c) => [String(c.id), c]));
    wireSettlePies(root);
    wireSpendTimeline(root);
    wireSideTabs(root);

    const sortEl = document.getElementById("charge-sort");
    if (sortEl) {
      sortEl.addEventListener("change", () => {
        localStorage.setItem(CHARGE_SORT_KEY, sortEl.value);
        renderWorkspace(groupId).catch((err) => flash(err.message));
      });
    }

    const openSwitch = () => {
      openGroupModal(groupId).catch((err) => flash(err.message));
    };
    document.getElementById("btn-switch").addEventListener("click", openSwitch);
    const titleEl = document.getElementById("btn-switch-title");
    titleEl.addEventListener("click", openSwitch);
    titleEl.addEventListener("keydown", (e) => {
      if (e.key === "Enter" || e.key === " ") {
        e.preventDefault();
        openSwitch();
      }
    });
    document.getElementById("btn-open-members").addEventListener("click", () => {
      openMembersModal(groupId).catch((err) => flash(err.message));
    });
    document.getElementById("btn-new-charge").addEventListener("click", () => {
      openChargeModal(groupId).catch((err) => flash(err.message));
    });
    root.querySelectorAll("[data-edit-charge]").forEach((btn) => {
      btn.addEventListener("click", () => {
        const charge = chargeById[btn.dataset.editCharge];
        if (!charge) return;
        openEditChargeModal(groupId, charge, group.currency).catch((err) => flash(err.message));
      });
    });
  }

  async function resolveLandingGroupId() {
    const { data: groups } = await api("GET", "/groups");
    if (!groups.length) return null;
    const stored = getStoredGroupId();
    if (stored && groups.some((g) => g.id === stored)) return stored;
    return [...groups].sort((a, b) => b.id - a.id)[0].id;
  }

  async function render() {
    try {
      const route = parseRoute();
      if (route.page === "group" && route.id) {
        await renderWorkspace(route.id);
        return;
      }
      const latestId = await resolveLandingGroupId();
      if (latestId) {
        const target = "#/group/" + latestId;
        if (location.hash !== target) goGroup(latestId);
        else await renderWorkspace(latestId);
        return;
      }
      renderEmptyHome();
    } catch (err) {
      root.innerHTML = `<p class="empty">${esc(err.message)}</p>`;
      flash(err.message);
    }
  }

  document.getElementById("btn-categories").addEventListener("click", () => {
    openCategoriesModal().catch((err) => flash(err.message));
  });
  document.getElementById("btn-people").addEventListener("click", () => {
    openPeopleModal().catch((err) => flash(err.message));
  });

  document.getElementById("btn-theme").addEventListener("click", cycleTheme);
  applyTheme(getTheme());

  window.addEventListener("hashchange", render);
  render();
})();
