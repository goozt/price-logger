let chart = null;
let chartUnit = "day";
let chartType = "timeseries";
let chartXRange = "timestamp";
let timeRangeDays = 1;
let selectedProduct = "";
let urlCheckboxSelected = [];
const refreshRate = 60; // minutes
const toggle = (handle, a, b) => (handle == a ? b : a);
const destroyChart = () => (chart !== null ? chart.destroy() : null);
const newItem = (item, time) => ({ price: item.price, time });
pb.autoCancellation(false);

const loginBtn = document.getElementById("login-btn");
const loginModal = new bootstrap.Modal("#loginModal");
const deleteUrlButton = document.getElementById("delete-urls-btn");

// Calculate Maximum Y Value for the chart
function getMaxYValue(prices) {
  let value = Math.max(...prices);
  value = value + value * 0.5;
  const slabs = [1, 2, 3, 4, 5, 7.5, 10, 15, 20, 25, 50, 75, 100].map(
    (x) => x * 1000
  );
  for (const idx in slabs) if (value < slabs[idx]) return slabs[idx];
  return Math.round(value / 1000) * 1000;
}

// Draw Chart
async function drawChart() {
  const data = await loadProductData();
  destroyChart();
  const ctx = document.getElementById("priceChart").getContext("2d");
  chart = new Chart(ctx, {
    type: "line",
    data: {
      labels: data.labels,
      datasets: [
        {
          label: "Price",
          data: data.prices,
          borderColor: "red",
          fill: false,
          borderWidth: 1,
        },
      ],
    },
    options: {
      responsive: true,
      scales: {
        x: data.scales.xScale,
        y: data.scales.yScale,
      },
      animation: false,
      spanGaps: true,
      plugins: {
        legend: {
          display: false,
        },
      },
    },
  });
}

// Prepare data to draw chart
async function loadProductData() {
  const predata = await fetchPrices();
  let data = [];
  predata.forEach((item) => {
    const created = item.created;
    const updated = item.updated;
    if (created == updated) {
      data.push(newItem(item, created));
    } else {
      data.push(newItem(item, created));
      data.push(newItem(item, updated));
    }
  });
  const reversedData = [...data].reverse();
  const labels = data.map((p) => new Date(p.time));
  const prices = data.map((p) => p.price);
  const currency = new Intl.NumberFormat("en-IN", {
    minimumFractionDigits: 2,
  });
  document.getElementById("chart-price").textContent =
    "₹" + currency.format(reversedData[0].price.toFixed(2));

  let xMinRange = data[0].time;
  let xMaxRange = reversedData[0].time;
  if (chartXRange != "timestamp") {
    const now = new Date();
    xMaxRange = now.valueOf() + 0.5;
    now.setDate(now.getDate() - timeRangeDays);
    xMinRange = now.getTime();
  }

  const xScale = {
    type: chartType,
    time: {
      unit: chartUnit,
      minUnit: "hour",
      displayFormats: { hour: "MMM DD hh:mm a", day: "MMM DD hh:mm a" },
    },
    min: xMinRange,
    max: xMaxRange,
    ticks: { major: { enabled: true } },
  };

  const yScale = { min: 0, max: getMaxYValue(prices) };

  return { labels, prices, scales: { xScale, yScale } };
}

// Populate fetched product data in to a list group
function updateProductList(products) {
  const list = document.getElementById("product-list");
  list.innerHTML = "";
  if (products != null) {
    products.forEach((product) => {
      const li = document.createElement("li");
      li.setAttribute("data-id", product.id);
      li.textContent = product.name;
      li.onclick = () => {
        document
          .querySelectorAll("#product-list li")
          .forEach((item) => item.classList.remove("selected"));
        li.classList.add("selected");
        document.getElementById("chart-header").textContent = product.name;
        selectedProduct = product.id;
        drawChart();
      };
      list.appendChild(li);
    });
  } else {
    list.textContent = "No data found";
  }
}

// Reload product data for the selected product
function reloadProduct() {
  const selected = document.querySelector("#product-list li.selected");
  if (selected) {
    selectedProduct = selected.getAttribute("data-id");
    drawChart();
  }
}

// Fetch prices from database
async function fetchPrices() {
  try {
    return (
      await pb.collection("prices").getList(1, 50, {
        sort: "updated",
        filter: 'product~"' + selectedProduct + '"',
        fields: "price,created,updated",
      })
    )["items"];
  } catch (error) {
    console.warn("Error loading product data:", error);
  }
}

// Fetch products from database
async function fetchProducts() {
  try {
    const records = await pb.collection("products").getFullList({
      sort: "-created",
      fields: "id,name",
    });
    updateProductList(records);
  } catch (error) {
    console.warn(error);
  }
}

// Loads all urls to settings
async function loadUrls() {
  try {
    const urls = await pb.collection("urls").getFullList();
    const urlList = document.getElementById("url-list");
    urlList.innerHTML = "";
    urls.forEach((url) => {
      const li = document.createElement("li");
      li.className = "list-group-item";
      const checkbox = document.createElement("input");
      checkbox.className = "form-check-input me-3";
      checkbox.type = "checkbox";
      checkbox.id = "url-" + url.id;
      checkbox.onchange = deleteUrl;
      li.appendChild(checkbox);
      const label = document.createElement("label");
      label.className = "form-check-label";
      label.type = "checkbox";
      label.setAttribute("for", "url-" + url.id);
      let a = document.createElement("a");
      let link = document.createTextNode(url.url);
      a.appendChild(link);
      a.setAttribute("target", "_blank");
      a.title = a.href = url.url;
      label.appendChild(a);
      li.appendChild(label);
      urlList.appendChild(li);
    });
  } catch (error) {
    console.warn(error);
  }
}

// Delete selected URL in settings
function deleteUrl(e) {
  const checkbox = e.target;
  const id = checkbox.id.split("-")[1];
  if (checkbox.checked) {
    urlCheckboxSelected.push(id);
  } else {
    const index = urlCheckboxSelected.indexOf(id);
    if (index >= 0) {
      urlCheckboxSelected.splice(index, 1);
    }
  }
  if (urlCheckboxSelected.length > 0) {
    deleteUrlButton.classList.remove("hidden");
  } else {
    deleteUrlButton.classList.add("hidden");
  }
}

// Check for User Authentication when page loads
function checkAuth() {
  if (pb.authStore.isValid && pb.authStore.record) {
    document.getElementById("username-tag").textContent =
      pb.authStore.record.name != ""
        ? pb.authStore.record.name
        : pb.authStore.record.email;
    document.querySelector(".welcome-user").classList.remove("hidden");
    loginBtn.innerHTML = `<svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" fill="currentColor" class="bi bi-box-arrow-left" viewBox="0 0 16 16">
        <path fill-rule="evenodd" d="M6 12.5a.5.5 0 0 0 .5.5h8a.5.5 0 0 0 .5-.5v-9a.5.5 0 0 0-.5-.5h-8a.5.5 0 0 0-.5.5v2a.5.5 0 0 1-1 0v-2A1.5 1.5 0 0 1 6.5 2h8A1.5 1.5 0 0 1 16 3.5v9a1.5 1.5 0 0 1-1.5 1.5h-8A1.5 1.5 0 0 1 5 12.5v-2a.5.5 0 0 1 1 0z"/>
        <path fill-rule="evenodd" d="M.146 8.354a.5.5 0 0 1 0-.708l3-3a.5.5 0 1 1 .708.708L1.707 7.5H10.5a.5.5 0 0 1 0 1H1.707l2.147 2.146a.5.5 0 0 1-.708.708z"/>
        </svg>`;
  }
}

addEventListener("load", (event) => {
  // Reload product data for chart periodically
  setInterval(() => reloadProduct(), refreshRate * 60000);
  checkAuth();
  fetchProducts();

  // Toggle between Time Series and Time Cartesian chart
  document.getElementById("toggleChart").onchange = async () => {
    destroyChart();
    let label = document.getElementById("toggleChartLabel");
    chartType = toggle(chartType, "time", "timeseries");
    chartUnit = toggle(chartUnit, "hour", "day");
    label.textContent = toggle(
      label.textContent,
      "Time Cartesian Chart",
      "Time Series Chart"
    );
    reloadProduct();
  };

  // Toggle Time Axis type
  document.getElementById("toggleXRange").onchange = async (e) => {
    destroyChart();
    const slider = document.getElementById("time-slider-container");
    if (e.target.checked) {
      chartXRange = "timerange";
      slider.classList.remove("hidden");
    } else {
      chartXRange = "timestamp";
      slider.classList.add("hidden");
    }
    document.getElementById("toggleXRangeLabel").textContent =
      "Use " + chartXRange;
    reloadProduct();
  };

  // Time Range Slider on new Input
  document.getElementById("time-slider").oninput = async (e) => {
    timeRangeDays = parseFloat(e.target.value);
    document.getElementById(
      "time-slider-value"
    ).textContent = `${timeRangeDays} days`;
    reloadProduct();
  };

  // Search Name on KeyUp
  document.getElementById("search-box").onkeyup = async (e) => {
    const searchValue = e.target.value.toLowerCase();
    document.querySelectorAll("#product-list li").forEach((li) => {
      if (li.textContent.toLowerCase().includes(searchValue)) {
        li.classList.remove("hidden");
      } else {
        li.classList.add("hidden");
      }
    });
  };

  // Add URL on Button Click
  document.getElementById("add-url-btn").addEventListener("click", async () => {
    const newUrl = document.getElementById("new-url").value;
    if (newUrl == "") return;
    try {
      const record = await pb.collection("urls").create({
        url: newUrl,
        type: "wishlist",
      });
      if (newUrl == record.url) {
        document.getElementById("new-url").value = "";
      }
    } catch (error) {
      console.warn(error);
    }
    await loadUrls();
  });

  // Reload URLs on Settings button click
  document
    .getElementById("settings-btn")
    .addEventListener("click", async () => {
      await loadUrls();
    });

  // Setting menu side panel
  document.querySelectorAll("#settings-menu button").forEach((button) => {
    button.addEventListener("click", () => {
      document
        .querySelectorAll(".settings-section")
        .forEach((section) => section.classList.add("hidden"));
      document
        .getElementById(button.dataset.section)
        .classList.remove("hidden");
    });
  });

  // Login Button on Click
  loginBtn.addEventListener("click", () => {
    if (pb.authStore.isValid) {
      pb.authStore.clear();
      document.getElementById("username-tag").textContent = "";
      document.querySelector(".welcome-user").classList.add("hidden");
      loginBtn.innerHTML = `<svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" fill="currentColor" class="bi bi-person-fill" viewBox="0 0 16 16">
  <path d="M3 14s-1 0-1-1 1-4 6-4 6 3 6 4-1 1-1 1zm5-6a3 3 0 1 0 0-6 3 3 0 0 0 0 6"/>
</svg>`;
    } else {
      loginModal.show();
    }
  });

  // User Login on Submit click
  document.getElementById("loginSubmit").addEventListener("click", async () => {
    const email = document.getElementById("email").value;
    const password = document.getElementById("password").value;
    try {
      const authData = await pb
        .collection("users")
        .authWithPassword(email, password);
      if (pb.authStore.isValid) {
        document.getElementById("username-tag").textContent =
          authData.record.email;
        document.querySelector(".welcome-user").classList.remove("hidden");
        loginModal.hide();
        loginBtn.innerHTML = `<svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" fill="currentColor" class="bi bi-box-arrow-left" viewBox="0 0 16 16">
          <path fill-rule="evenodd" d="M6 12.5a.5.5 0 0 0 .5.5h8a.5.5 0 0 0 .5-.5v-9a.5.5 0 0 0-.5-.5h-8a.5.5 0 0 0-.5.5v2a.5.5 0 0 1-1 0v-2A1.5 1.5 0 0 1 6.5 2h8A1.5 1.5 0 0 1 16 3.5v9a1.5 1.5 0 0 1-1.5 1.5h-8A1.5 1.5 0 0 1 5 12.5v-2a.5.5 0 0 1 1 0z"/>
          <path fill-rule="evenodd" d="M.146 8.354a.5.5 0 0 1 0-.708l3-3a.5.5 0 1 1 .708.708L1.707 7.5H10.5a.5.5 0 0 1 0 1H1.707l2.147 2.146a.5.5 0 0 1-.708.708z"/>
          </svg>`;
      }
    } catch (err) {
      alert("Login failed: " + err.message);
    }
  });

  // Delete url on Delete button click
  deleteUrlButton.addEventListener("click", async (e) => {
    try {
      urlCheckboxSelected.forEach(async (id) => {
        await pb.collection("urls").delete(id);
      });
    } catch (error) {
      console.warn(error);
    }
    urlCheckboxSelected = [];
    e.target.classList.add("hidden");
    await loadUrls();
  });
});
