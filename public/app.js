document.addEventListener('DOMContentLoaded', function(){
  const tz = Intl.DateTimeFormat().resolvedOptions().timeZone;
  const label = document.getElementById('view-browser-tz');

  if (label && label.parentElement) {
    const span = document.createElement('span');
    span.style.marginLeft = '6px';
    span.style.fontSize = '0.9em';
    span.style.color = '#555';
    span.textContent = `(Your TZ: ${tz})`;
    label.parentElement.appendChild(span);
  }
});

function pad2(n){ 
  return n < 10 ? '0'+n : ''+n 
}

function formatLocal(dtString) {
  const d = new Date(dtString);
  const yyyy = d.getFullYear();
  const mm = pad2(d.getMonth()+1);
  const dd = pad2(d.getDate());
  const HH = pad2(d.getHours());
  const MM = pad2(d.getMinutes());
  const SS = pad2(d.getSeconds());

  return `${yyyy}-${mm}-${dd} ${HH}:${MM}:${SS}`;
}

function toggleBrowserTZ(cb) {
  const runCells = document.querySelectorAll('td.dt-run');
  const dueCells = document.querySelectorAll('td.dt-due');

  runCells.forEach(td => {
    const utc = td.getAttribute('data-utc');
    const local = td.getAttribute('data-local');
    td.textContent = cb.checked ? formatLocal(utc) : local;
  });
  dueCells.forEach(td => {
    const utc = td.getAttribute('data-utc');
    const local = td.getAttribute('data-local');
    td.textContent = cb.checked ? formatLocal(utc) : local;
  });
}

window.toggleBrowserTZ = toggleBrowserTZ;

function formatDueInFromNow(dtString){
  const target = new Date(dtString).getTime();
  const now = Date.now();
  let diff = target - now; // ms
  let suffix = 'in ';
  
  if (diff < 0) { diff = -diff; suffix = ''; }
  
  const sec = Math.floor(diff/1000);
  const d = Math.floor(sec/86400);
  const h = Math.floor((sec%86400)/3600);
  const m = Math.floor((sec%3600)/60);
  const s = Math.floor(sec%60);
  
  let parts = [];
  
  if (d>0) parts.push(d+'d');
  if (h>0 && parts.length<2) parts.push(h+'h');
  if (m>0 && parts.length<2) parts.push(m+'m');
  if (parts.length===0) parts.push(s+'s');
  
  return suffix + parts.join(' ');
}

function tickDueIn(){
  document.querySelectorAll('td.dt-duein').forEach(td => {
    const utc = td.getAttribute('data-utc');
    td.textContent = formatDueInFromNow(utc);
  });
}

function fillDueBrowser(){
  document.querySelectorAll('td.dt-due-browser').forEach(td => {
    const utc = td.getAttribute('data-utc');
    td.textContent = formatLocal(utc);
  });
}

document.addEventListener('DOMContentLoaded', function(){
  tickDueIn();
  fillDueBrowser();
  setInterval(tickDueIn, 30000);
});

