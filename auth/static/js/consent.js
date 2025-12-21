// accordion toggle
document.querySelectorAll('button[data-target]').forEach((btn) => {
  btn.addEventListener('click', () => {
    const id = btn.getAttribute('data-target');
    const panel = document.getElementById(id);
    const chev = document.querySelector(`[data-chevron="${id}"]`);
    panel?.classList.toggle('hidden');
    chev?.classList.toggle('rotate-180');
  });
});

function toggleGroup(selector) {
  const boxes = Array.from(document.querySelectorAll(selector));
  const allChecked = boxes.every((b) => b.checked);
  boxes.forEach((b) => (b.checked = !allChecked));
}

document.getElementById('accountsSelectAll')?.addEventListener('click', () => {
  toggleGroup('.account-checkbox');
});

document.getElementById('loansSelectAll')?.addEventListener('click', () => {
  toggleGroup('.loan-checkbox');
});