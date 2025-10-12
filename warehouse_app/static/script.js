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
  if (pageId === 'dashboard' || pageId === 'sales' || pageId === 'aggregated') loadOres();
  if (pageId === 'sales') loadSales();
  if (pageId === 'reports') loadReports();
  if (pageId === 'logs') loadLogs();
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

// Загрузка данных
function loadOres() {
  fetch('/api/ores')
    .then(response => response.json())
    .then(ores => {
      const tbody = document.getElementById('ores-table-body');
      const salesTbody = document.getElementById('sales-ores-table-body');
      const aggregatedTbody = document.getElementById('aggregated-table-body');
      tbody.innerHTML = '';
      salesTbody.innerHTML = '';
      aggregatedTbody.innerHTML = '';
      ores.forEach(ore => {
        const status = ore.quantity < 100 ? 'Критический уровень' : 'В наличии';
        const rowClass = ore.quantity < 100 ? 'critical' : '';
        const row = `<tr class="${rowClass}">
          <td>${ore.type}</td>
          <td>${ore.quantity}</td>
          <td>${status}</td>
        </tr>`;
        tbody.innerHTML += row;
        salesTbody.innerHTML += row;
        aggregatedTbody.innerHTML += `<tr class="${rowClass}">
          <td>Завод 1</td>
          <td>${ore.quantity}</td>
          <td>${status}</td>
        </tr>`;
      });
    })
    .catch(error => console.error('Ошибка загрузки руды:', error));
}

function loadSales() {
  fetch('/api/sales')
    .then(response => response.json())
    .then(sales => {
      const tbody = document.getElementById('sales-table-body');
      tbody.innerHTML = '';
      sales.forEach(sale => {
        tbody.innerHTML += `<tr>
          <td>${new Date(sale.created_at).toLocaleDateString()}</td>
          <td>${sale.buyer || 'Не указан'}</td>
          <td>${sale.ore_type}</td>
          <td>${sale.quantity}</td>
          <td>${sale.status}</td>
        </tr>`;
      });
    })
    .catch(error => console.error('Ошибка загрузки продаж:', error));
}

function loadReports() {
  fetch('/api/sales')
    .then(response => response.json())
    .then(sales => {
      const tbody = document.getElementById('reports-table-body');
      tbody.innerHTML = '';
      const totalSold = sales.reduce((sum, sale) => sum + sale.quantity, 0);
      fetch('/api/ores')
        .then(response => response.json())
        .then(ores => {
          const totalAvailable = ores.reduce((sum, ore) => sum + ore.quantity, 0);
          fetch('/api/tools')
            .then(response => response.json())
            .then(tools => {
              const inventory = tools.map(t => `${t.type}: ${t.quantity}`).join(', ');
              tbody.innerHTML += `<tr>
                <td>Октябрь 2025</td>
                <td>${totalSold}</td>
                <td>${totalAvailable}</td>
                <td>${inventory}</td>
              </tr>`;
            });
        });
    })
    .catch(error => console.error('Ошибка загрузки отчётов:', error));
}

function loadLogs() {
  fetch('/api/logs')
    .then(response => response.json())
    .then(logs => {
      const tbody = document.getElementById('logs-table-body');
      tbody.innerHTML = '';
      logs.forEach(log => {
        tbody.innerHTML += `<tr>
          <td>${log.date}</td>
          <td>${log.user}</td>
          <td>${log.action}</td>
        </tr>`;
      });
    })
    .catch(error => console.error('Ошибка загрузки логов:', error));
}

// Сохранение данных
function saveOre() {
  const form = document.getElementById('ore-form');
  const data = {
    type: form.querySelector('[name="type"]').value,
    quantity: parseFloat(form.querySelector('[name="quantity"]').value),
    location: form.querySelector('[name="location"]').value,
    quality: parseFloat(form.querySelector('[name="quality"]').value) || null,
    priority: form.querySelector('[name="priority"]').value
  };
  if (!data.type || !data.quantity) {
    alert('Заполните обязательные поля!');
    return;
  }
  fetch('/api/ores', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(data)
  })
    .then(response => response.json())
    .then(result => {
      alert(result.message);
      form.reset();
      loadOres();
    })
    .catch(error => alert('Ошибка: ' + error));
}

function saveTool() {
  const form = document.getElementById('tool-form');
  const data = {
    type: form.querySelector('[name="type"]').value,
    quantity: parseInt(form.querySelector('[name="quantity"]').value),
    serial_number: form.querySelector('[name="serial_number"]').value,
    service_life: parseInt(form.querySelector('[name="service_life"]').value) || null
  };
  if (!data.type || !data.quantity) {
    alert('Заполните обязательные поля!');
    return;
  }
  fetch('/api/tools', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(data)
  })
    .then(response => response.json())
    .then(result => {
      alert(result.message);
      form.reset();
      loadOres();
    })
    .catch(error => alert('Ошибка: ' + error));
}

function updateSaleStatus() {
  const form = document.getElementById('sale-form');
  const saleId = form.dataset.saleId || '0'; // Для упрощения, предполагаем выбор последней продажи
  const data = {
    id: saleId,
    status: form.querySelector('[name="status"]').value
  };
  fetch(`/api/sales/${saleId}`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(data)
  })
    .then(response => response.json())
    .then(result => {
      alert(result.message);
      document.getElementById('sale-status').innerText = `Статус: ${data.status}`;
      loadSales();
    })
    .catch(error => alert('Ошибка: ' + error));
}

function confirmSale() {
  if (confirm('Подтвердить списание руды?')) {
    const form = document.getElementById('sale-form');
    const data = {
      ore_type: form.querySelector('[name="ore_type"]').value,
      buyer: form.querySelector('[name="buyer"]').value,
      quantity: parseFloat(form.querySelector('[name="quantity"]').value),
      status: form.querySelector('Списано')
    };
    if (!data.ore_type || !data.quantity) {
      alert('Заполните обязательные поля!');
      return;
    }
    fetch('/api/sales', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(data)
    })
      .then(response => response.json())
      .then(result => {
        alert(result.message);
        form.reset();
        document.getElementById('sale-status').innerText = 'Статус: Списано';
        loadSales();
      })
      .catch(error => alert('Ошибка: ' + error));
  }
}

// Загрузка данных при старте
document.addEventListener('DOMContentLoaded', () => {
  loadOres();
  loadSales();
  loadReports();
  loadLogs();
});