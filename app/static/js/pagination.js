(function () {
  var PAGE_SIZE = 15;

  function init() {
    var table = document.getElementById('results-table');
    if (!table) return;

    var tbody = table.querySelector('tbody');
    if (!tbody) return;

    var rows = Array.from(tbody.querySelectorAll('tr'));
    if (rows.length <= PAGE_SIZE) return;

    var total = rows.length;
    var totalPages = Math.ceil(total / PAGE_SIZE);
    var current = 1;

    function showPage(n) {
      current = n;
      var start = (n - 1) * PAGE_SIZE;
      var end = start + PAGE_SIZE;
      rows.forEach(function (row, i) {
        row.style.display = (i >= start && i < end) ? '' : 'none';
      });
      updateControls();
    }

    function updateControls() {
      var start = (current - 1) * PAGE_SIZE + 1;
      var end = Math.min(current * PAGE_SIZE, total);
      infoEl.textContent = 'Showing ' + start + '–' + end + ' of ' + total;
      prevBtn.disabled = current === 1;
      nextBtn.disabled = current === totalPages;
    }

    var bar = document.createElement('div');
    bar.className = 'pagination';

    var prevBtn = document.createElement('button');
    prevBtn.type = 'button';
    prevBtn.className = 'page-btn';
    prevBtn.textContent = '← Prev';
    prevBtn.addEventListener('click', function () {
      if (current > 1) showPage(current - 1);
    });

    var infoEl = document.createElement('span');
    infoEl.className = 'page-info';

    var nextBtn = document.createElement('button');
    nextBtn.type = 'button';
    nextBtn.className = 'page-btn';
    nextBtn.textContent = 'Next →';
    nextBtn.addEventListener('click', function () {
      if (current < totalPages) showPage(current + 1);
    });

    bar.appendChild(prevBtn);
    bar.appendChild(infoEl);
    bar.appendChild(nextBtn);

    var wrapper = table.closest('.table-wrapper');
    wrapper.parentNode.insertBefore(bar, wrapper.nextSibling);

    showPage(1);
  }

  function initFileDisplay() {
    var input = document.getElementById('csv_file');
    var display = document.getElementById('file-name-display');
    if (!input || !display) return;
    input.addEventListener('change', function () {
      display.textContent = input.files.length > 0 ? input.files[0].name : 'No file chosen';
    });
  }

  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', function () { init(); initFileDisplay(); });
  } else {
    init();
    initFileDisplay();
  }
}());
