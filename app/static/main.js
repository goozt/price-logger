let chart = null;
let chartType = "timeseries";
let chartUnit = "day";
let chartXRange = "timestamp";
let productNames = [];
let timeRangeDays = 1;
const refreshRate = 10; // minutes
const toggle = (handle, a, b) => (handle == a ? b : a);
const destroyChart = () => (chart !== null ? chart.destroy() : null);

function getMaxYValue(prices) {
  let value = Math.max(...prices);
  value = value + value * 0.5;
  const slabs = [1, 2, 3, 4, 5, 7.5, 10, 15, 20, 25, 50, 75, 100].map(
    (x) => x * 1000
  );
  for (idx in slabs) if (value < slabs[idx]) return slabs[idx];
  return Math.round(value / 1000) * 1000;
}

async function loadProductData(productName) {
  try {
    document.getElementById("chart-header").textContent = productName;
    const res = await fetch(
      `/api/prices?name=${encodeURIComponent(productName)}`
    );

    if (!res.ok) {
      throw new Error(`HTTP error! Status: ${res.status}`);
    }

    let data = await res.json();
    if (!data || data.length === 0) {
      console.warn(`No price data found for ${productName}`);
      return;
    }

    data.sort((a, b) => a.updated_at > b.updated_at);
    const reversedData = [...data].reverse();
    const labels = data.map((p) => new Date(p.updated_at));
    const prices = data.map((p) => p.price);
    const currency = new Intl.NumberFormat("en-IN", {
      minimumFractionDigits: 2,
    });
    document.getElementById("chart-price").textContent =
      "₹" + currency.format(reversedData[0].price.toFixed(2));

    let xMinRange = data[0].created_at;
    let xMaxRange = reversedData[0].updated_at;
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
    destroyChart();

    const ctx = document.getElementById("priceChart").getContext("2d");
    chart = new Chart(ctx, {
      type: "line",
      data: {
        labels: labels,
        datasets: [
          {
            label: productName + " Price",
            data: prices,
            borderColor: "red",
            fill: false,
            borderWidth: 1,
          },
        ],
      },
      options: {
        responsive: true,
        scales: {
          x: xScale,
          y: yScale,
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
  } catch (error) {
    console.error("Error loading product data:", error);
  }
}

function updateProductList(products) {
  const list = document.getElementById("product-list");
  list.innerHTML = "";
  if (products != null) {
    products.forEach((product) => {
      const li = document.createElement("li");
      li.textContent = product;
      li.onclick = () => {
        document
          .querySelectorAll("#product-list li")
          .forEach((item) => item.classList.remove("selected"));
        li.classList.add("selected");
        loadProductData(product);
      };
      list.appendChild(li);
    });
  } else {
    list.textContent = "No data found";
  }
}

async function fetchProducts() {
  const res = await fetch("/api/products");
  productNames = await res.json();
  updateProductList(productNames);
}

async function fetchNewData() {
  const res = await fetch("/api/new");
  if (!res.ok) {
    throw new Error(`HTTP error! Status: ${res.status}`);
  }
  fetchProducts();
}

async function resetData() {
  const res = await fetch("/api/reset");
  if (!res.ok) {
    throw new Error(`HTTP error! Status: ${res.status}`);
  }
  fetchNewData();
}

function reloadProduct() {
  const selected = document.querySelector("#product-list li.selected");
  if (selected) loadProductData(selected.textContent);
}

addEventListener("load", (event) => {
  setInterval(() => reloadProduct(), refreshRate * 60000);
  fetchProducts();

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
});
