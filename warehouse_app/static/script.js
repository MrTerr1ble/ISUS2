let referenceData = {
  units: [],
  warehouses: [],
  ore_types: [],
  equipment_categories: [],
  contractors: [],
  transport: []
};
let oreBatches = [];
let equipmentList = [];
let orders = [];
let shipments = [];

// Навигация
function showPage(pageId) {
  document.querySelectorAll('.page').forEach(page => page.classList.remove('active'));
  document.getElementById(pageId).classList.add('active');
  document.querySelectorAll('.nav-button').forEach(btn => btn.classList.remove('active'));
  document.querySelector(`[onclick="showPage('${pageId}')"]`).classList.add('active');
  if (window.innerWidth <= 768) {
    document.getElementById('sidebar').classList.remove('active');
    document.getElementById('content').classList.add('full');
  }

  switch (pageId) {
    case 'dashboard':
      renderDashboard();
      break;
    case 'input-ore':
      renderOreBatchTable();
      break;
    case 'input-equipment':
      renderEquipmentTable();
      break;
    case 'orders':
      renderOrdersTable();
      break;
    case 'shipments':
      renderShipmentsTable();
      break;
    case 'reports':
      loadReports();
      break;
    case 'logs':
      loadLogs();
      break;
  }
}

function toggleSidebar() {
  const sidebar = document.getElementById('sidebar');
  const content = document.getElementById('content');
  sidebar.classList.toggle('active');
  content.classList.toggle('full');
}

function toggleTheme() {
  document.body.classList.toggle('dark');
}

// Поиск в таблицах
function filterTable(input, tableId) {
  const filter = input.value.toLowerCase();
  const table = document.getElementById(tableId);
  if (!table) return;
  const rows = table.getElementsByTagName('tr');
  for (let i = 1; i < rows.length; i++) {
    const cells = rows[i].getElementsByTagName('td');
    let match = false;
    for (let j = 0; j < cells.length; j++) {
      if (cells[j].innerText.toLowerCase().includes(filter)) {
        match = true;
        break;
      }
    }
    rows[i].style.display = match ? '' : 'none';
  }
}

// Справочники
function loadReferenceData() {
  return fetch('/api/reference-data')
    .then(response => response.json())
    .then(data => {
      referenceData = data;
      populateSelect('ore-type-select', referenceData.ore_types, item => item.id, item => item.name);
      populateSelect('ore-warehouse-select', referenceData.warehouses, item => item.id, item => `${item.name} (${item.location || '—'})`);
      populateSelect('ore-unit-select', referenceData.units, item => item.id, item => `${item.name} (${item.symbol})`);

      populateSelect('equipment-category-select', referenceData.equipment_categories, item => item.id, item => item.name);
      populateSelect('equipment-warehouse-select', referenceData.warehouses, item => item.id, item => item.name);
      populateSelect('equipment-unit-select', referenceData.units, item => item.id, item => `${item.name} (${item.symbol})`);

      populateSelect('contractor-select', referenceData.contractors.filter(c => c.type !== 'Поставщик'), item => item.id, item => item.name);
      populateSelect('order-warehouse-select', referenceData.warehouses, item => item.id, item => item.name);

      populateSelect('shipment-transport-select', referenceData.transport, item => item.id, item => `${item.name} (${item.type || '—'})`);

      populateSelect('reports-warehouse-filter', [{ id: '', name: 'Все склады' }, ...referenceData.warehouses], item => item.id, item => item.name || item);
      populateSelect('reports-oretype-filter', [{ id: '', name: 'Все типы руды' }, ...referenceData.ore_types], item => item.id, item => item.name || item);
    })
    .catch(error => console.error('Ошибка загрузки справочников:', error));
}

function populateSelect(elementId, items, valueFn, labelFn) {
  const select = document.getElementById(elementId);
  if (!select) return;
  const previousValue = select.value;
  select.innerHTML = '';
  const defaultOption = document.createElement('option');
  defaultOption.value = '';
  defaultOption.textContent = 'Выберите значение';
  select.appendChild(defaultOption);
  items.forEach(item => {
    const option = document.createElement('option');
    option.value = valueFn(item);
    option.textContent = labelFn(item);
    select.appendChild(option);
  });
  if (previousValue) {
    select.value = previousValue;
  }
}

// Руды
function loadOreBatches() {
  return fetch('/api/ore-batches')
    .then(response => response.json())
    .then(data => {
      oreBatches = data;
      renderDashboard();
      renderOreBatchTable();
      refreshOrderItemRows();
      loadReports();
    })
    .catch(error => console.error('Ошибка загрузки партий руды:', error));
}

function renderDashboard() {
  const tbody = document.getElementById('dashboard-batches-body');
  if (!tbody) return;
  tbody.innerHTML = '';
  oreBatches.forEach(batch => {
    const quality = batch.quality ? `${batch.quality.toFixed(2)}%` : '—';
    const statusClass = batch.status === 'Критический' || batch.priority === 'Критический' ? 'critical' : '';
    const row = document.createElement('tr');
    row.className = statusClass;
    row.innerHTML = `
      <td>${batch.batch_code || 'Партия ' + batch.id}</td>
      <td>${batch.ore_type_name}</td>
      <td>${batch.warehouse_name}</td>
      <td>${batch.quantity.toFixed(2)}</td>
      <td>${batch.unit_symbol || batch.unit_name}</td>
      <td>${quality}</td>
      <td>${batch.status || '—'}</td>
    `;
    tbody.appendChild(row);
  });
}

function renderOreBatchTable() {
  const tbody = document.getElementById('ore-batches-table-body');
  if (!tbody) return;
  tbody.innerHTML = '';
  oreBatches.forEach(batch => {
    const row = document.createElement('tr');
    const statusClass = batch.priority === 'Критический' || batch.status === 'Критический' ? 'critical' : '';
    row.className = statusClass;
    row.innerHTML = `
      <td>${batch.batch_code || 'Партия ' + batch.id}</td>
      <td>${batch.ore_type_name}</td>
      <td>${batch.warehouse_name}</td>
      <td>${batch.quantity.toFixed(2)}</td>
      <td>${batch.unit_symbol || batch.unit_name}</td>
      <td>${batch.quality ? batch.quality.toFixed(2) + '%' : '—'}</td>
      <td>${batch.priority || '—'}</td>
      <td>${batch.status || '—'}</td>
    `;
    tbody.appendChild(row);
  });
}

function saveOreBatch() {
  const form = document.getElementById('ore-form');
  const data = {
    ore_type_id: parseInt(form.querySelector('[name="ore_type_id"]').value, 10),
    warehouse_id: parseInt(form.querySelector('[name="warehouse_id"]').value, 10),
    unit_id: parseInt(form.querySelector('[name="unit_id"]').value, 10),
    batch_code: form.querySelector('[name="batch_code"]').value.trim(),
    quantity: parseFloat(form.querySelector('[name="quantity"]').value),
    quality: form.querySelector('[name="quality"]').value ? parseFloat(form.querySelector('[name="quality"]').value) : null,
    priority: form.querySelector('[name="priority"]').value,
    extraction_date: form.querySelector('[name="extraction_date"]').value,
    status: form.querySelector('[name="status"]').value
  };
  if (!data.ore_type_id || !data.warehouse_id || !data.unit_id || !data.quantity) {
    alert('Заполните обязательные поля!');
    return;
  }
  fetch('/api/ore-batches', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(data)
  })
    .then(response => response.json())
    .then(result => {
      alert(result.message);
      form.reset();
      loadOreBatches();
    })
    .catch(error => alert('Ошибка: ' + error));
}

// Оборудование
function loadEquipment() {
  return fetch('/api/equipment')
    .then(response => response.json())
    .then(data => {
      equipmentList = data;
      renderEquipmentTable();
    })
    .catch(error => console.error('Ошибка загрузки оборудования:', error));
}

function renderEquipmentTable() {
  const tbody = document.getElementById('equipment-table-body');
  if (!tbody) return;
  tbody.innerHTML = '';
  equipmentList.forEach(item => {
    const row = document.createElement('tr');
    row.innerHTML = `
      <td>${item.name}</td>
      <td>${item.category_name}</td>
      <td>${item.warehouse_name}</td>
      <td>${item.quantity}</td>
      <td>${item.unit_symbol || item.unit_name}</td>
      <td>${item.status || '—'}</td>
      <td>${item.serial_number || '—'}</td>
    `;
    tbody.appendChild(row);
  });
}

function saveEquipment() {
  const form = document.getElementById('equipment-form');
  const data = {
    name: form.querySelector('[name="name"]').value.trim(),
    category_id: parseInt(form.querySelector('[name="category_id"]').value, 10),
    warehouse_id: parseInt(form.querySelector('[name="warehouse_id"]').value, 10),
    quantity: parseFloat(form.querySelector('[name="quantity"]').value),
    unit_id: parseInt(form.querySelector('[name="unit_id"]').value, 10),
    serial_number: form.querySelector('[name="serial_number"]').value.trim(),
    service_life_months: form.querySelector('[name="service_life_months"]').value ? parseInt(form.querySelector('[name="service_life_months"]').value, 10) : null,
    status: form.querySelector('[name="status"]').value,
    purchase_date: form.querySelector('[name="purchase_date"]').value
  };
  if (!data.name || !data.category_id || !data.warehouse_id || !data.unit_id || !data.quantity) {
    alert('Заполните обязательные поля!');
    return;
  }
  fetch('/api/equipment', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(data)
  })
    .then(response => response.json())
    .then(result => {
      alert(result.message);
      form.reset();
      loadEquipment();
    })
    .catch(error => alert('Ошибка: ' + error));
}

// Заказы
function loadOrders() {
  return fetch('/api/orders')
    .then(response => response.json())
    .then(data => {
      orders = data;
      renderOrdersTable();
      populateSelect('shipment-order-select', orders, order => order.id, order => `${order.order_number} (${order.contractor_name})`);
      renderShipmentsTable();
      loadReports();
    })
    .catch(error => console.error('Ошибка загрузки заказов:', error));
}

function renderOrdersTable() {
  const tbody = document.getElementById('orders-table-body');
  if (!tbody) return;
  tbody.innerHTML = '';
  orders.forEach(order => {
    const row = document.createElement('tr');
    row.innerHTML = `
      <td>${order.order_number}</td>
      <td>${order.contractor_name}</td>
      <td>${order.warehouse_name}</td>
      <td>${order.status || '—'}</td>
      <td>${order.order_date ? new Date(order.order_date).toLocaleDateString() : '—'}</td>
      <td>${order.total_quantity ? order.total_quantity.toFixed(2) : '0.00'}</td>
    `;
    tbody.appendChild(row);
  });
}

function addOrderItemRow() {
  const container = document.getElementById('order-items');
  const row = document.createElement('div');
  row.className = 'order-item-row';
  row.style.display = 'grid';
  row.style.gridTemplateColumns = '2fr 1fr 1fr 1fr auto';
  row.style.gap = '10px';
  row.style.marginBottom = '10px';
  row.innerHTML = `
    <select class="select order-item-ore"></select>
    <select class="select order-item-unit"></select>
    <input class="input order-item-qty" type="number" step="0.01" min="0" placeholder="Кол-во" />
    <input class="input order-item-price" type="number" step="0.01" min="0" placeholder="Цена за ед." />
    <div class="button danger" style="padding: 8px;" onclick="removeOrderItemRow(this)"><i class="fas fa-trash"></i></div>
  `;
  container.appendChild(row);
  populateOreBatchSelect(row.querySelector('.order-item-ore'));
  populateSelectElement(row.querySelector('.order-item-unit'), referenceData.units, item => item.id, item => `${item.name} (${item.symbol})`);
}

function populateOreBatchSelect(select, selectedValue = '') {
  if (!select) return;
  select.innerHTML = '';
  const defaultOption = document.createElement('option');
  defaultOption.value = '';
  defaultOption.textContent = 'Выберите партию';
  select.appendChild(defaultOption);
  oreBatches.forEach(batch => {
    const option = document.createElement('option');
    option.value = batch.id;
    option.textContent = `${batch.batch_code || 'Партия ' + batch.id} — ${batch.ore_type_name} (${batch.quantity.toFixed(2)} ${batch.unit_symbol || batch.unit_name})`;
    if (String(batch.id) === String(selectedValue)) {
      option.selected = true;
    }
    select.appendChild(option);
  });
}

function populateSelectElement(select, items, valueFn, labelFn) {
  if (!select) return;
  const prev = select.value;
  select.innerHTML = '';
  const option = document.createElement('option');
  option.value = '';
  option.textContent = 'Выберите значение';
  select.appendChild(option);
  items.forEach(item => {
    const opt = document.createElement('option');
    opt.value = valueFn(item);
    opt.textContent = labelFn(item);
    select.appendChild(opt);
  });
  if (prev) select.value = prev;
}

function refreshOrderItemRows() {
  document.querySelectorAll('.order-item-row').forEach(row => {
    const oreSelect = row.querySelector('.order-item-ore');
    const currentOre = oreSelect.value;
    populateOreBatchSelect(oreSelect, currentOre);
  });
}

function removeOrderItemRow(button) {
  const row = button.closest('.order-item-row');
  if (row) {
    row.remove();
  }
  if (document.querySelectorAll('.order-item-row').length === 0) {
    addOrderItemRow();
  }
}

function saveOrder() {
  const form = document.getElementById('order-form');
  const items = [];
  document.querySelectorAll('.order-item-row').forEach(row => {
    const ore = row.querySelector('.order-item-ore').value;
    const unit = row.querySelector('.order-item-unit').value;
    const qty = row.querySelector('.order-item-qty').value;
    if (ore && unit && qty) {
      items.push({
        ore_batch_id: parseInt(ore, 10),
        unit_id: parseInt(unit, 10),
        quantity: parseFloat(qty),
        price_per_unit: row.querySelector('.order-item-price').value ? parseFloat(row.querySelector('.order-item-price').value) : 0
      });
    }
  });
  const data = {
    order_number: form.querySelector('[name="order_number"]').value.trim(),
    contractor_id: parseInt(form.querySelector('[name="contractor_id"]').value, 10),
    warehouse_id: parseInt(form.querySelector('[name="warehouse_id"]').value, 10),
    order_date: form.querySelector('[name="order_date"]').value,
    status: form.querySelector('[name="status"]').value,
    items
  };
  if (!data.order_number || !data.contractor_id || !data.warehouse_id || data.items.length === 0) {
    alert('Заполните обязательные поля и добавьте хотя бы одну позицию!');
    return;
  }
  fetch('/api/orders', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(data)
  })
    .then(response => response.json())
    .then(result => {
      alert(result.message);
      form.reset();
      document.getElementById('order-items').innerHTML = '';
      addOrderItemRow();
      loadOrders();
    })
    .catch(error => alert('Ошибка: ' + error));
}

// Отгрузки
function loadShipments() {
  return fetch('/api/shipments')
    .then(response => response.json())
    .then(data => {
      shipments = data;
      renderShipmentsTable();
      loadReports();
    })
    .catch(error => console.error('Ошибка загрузки отгрузок:', error));
}

function renderShipmentsTable() {
  const tbody = document.getElementById('shipments-table-body');
  if (!tbody) return;
  tbody.innerHTML = '';
  shipments.forEach(shipment => {
    const row = document.createElement('tr');
    row.innerHTML = `
      <td>${shipment.order_number}</td>
      <td>${shipment.transport_name || '—'}</td>
      <td>${shipment.planned_date ? new Date(shipment.planned_date).toLocaleDateString() : '—'}</td>
      <td>${shipment.actual_date ? new Date(shipment.actual_date).toLocaleDateString() : '—'}</td>
      <td>${shipment.status || '—'}</td>
    `;
    tbody.appendChild(row);
  });
}

function saveShipment() {
  const form = document.getElementById('shipment-form');
  const data = {
    order_id: parseInt(form.querySelector('[name="order_id"]').value, 10),
    transport_id: form.querySelector('[name="transport_id"]').value ? parseInt(form.querySelector('[name="transport_id"]').value, 10) : null,
    planned_date: form.querySelector('[name="planned_date"]').value,
    actual_date: form.querySelector('[name="actual_date"]').value,
    status: form.querySelector('[name="status"]').value
  };
  if (!data.order_id) {
    alert('Выберите заказ для отгрузки!');
    return;
  }
  fetch('/api/shipments', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(data)
  })
    .then(response => response.json())
    .then(result => {
      alert(result.message);
      form.reset();
      loadShipments();
    })
    .catch(error => alert('Ошибка: ' + error));
}

// Отчеты
function loadReports() {
  const tbody = document.getElementById('reports-table-body');
  if (!tbody) return;
  const totalOre = oreBatches.reduce((sum, batch) => sum + (batch.quantity || 0), 0);
  const critical = oreBatches.filter(batch => batch.priority === 'Критический' || batch.status === 'Критический');
  const totalOrdersQuantity = orders.reduce((sum, order) => sum + (order.total_quantity || 0), 0);
  const completedShipments = shipments.filter(s => s.status === 'Завершена').length;

  tbody.innerHTML = '';
  const rows = [
    { label: 'Остатки руды на складах (в выбранных единицах)', value: `${totalOre.toFixed(2)}` },
    { label: 'Количество критических партий', value: critical.length },
    { label: 'Заказано к отгрузке (т)', value: totalOrdersQuantity.toFixed(2) },
    { label: 'Количество отгрузок', value: shipments.length },
    { label: 'Завершено отгрузок', value: completedShipments }
  ];
  rows.forEach(row => {
    const tr = document.createElement('tr');
    tr.innerHTML = `<td>${row.label}</td><td>${row.value}</td>`;
    tbody.appendChild(tr);
  });
}

// Логи
function loadLogs() {
  return fetch('/api/logs')
    .then(response => response.json())
    .then(logs => {
      const tbody = document.getElementById('logs-table-body');
      if (!tbody) return;
      tbody.innerHTML = '';
      logs.forEach(log => {
        const row = document.createElement('tr');
        row.innerHTML = `
          <td>${log.event_time ? new Date(log.event_time).toLocaleString() : '—'}</td>
          <td>${log.user || '—'}</td>
          <td>${log.action || '—'}</td>
          <td>${log.entity || '—'}</td>
          <td>${log.details || '—'}</td>
        `;
        tbody.appendChild(row);
      });
    })
    .catch(error => console.error('Ошибка загрузки логов:', error));
}

// Инициализация
document.addEventListener('DOMContentLoaded', () => {
  loadReferenceData()
    .then(() => Promise.all([loadOreBatches(), loadEquipment(), loadOrders(), loadShipments(), loadLogs()]))
    .then(() => {
      if (document.querySelectorAll('.order-item-row').length === 0) {
        addOrderItemRow();
      }
    });
});
