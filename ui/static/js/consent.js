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

document.getElementById('auto-policies-select-all')?.addEventListener('click', () => {
  toggleGroup('.auto-policy-checkbox');
});

document.getElementById('capitalization-title-plans-select-all')?.addEventListener('click', () => {
  toggleGroup('.capitalization-title-plan-checkbox');
});

document.getElementById('financial-assistance-contracts-select-all')?.addEventListener('click', () => {
  toggleGroup('.financial-assistance-contract-checkbox');
});

document.getElementById('acceptance-and-branches-abroad-policies-select-all')?.addEventListener('click', () => {
  toggleGroup('.acceptance-and-branches-abroad-policy-checkbox');
});

document.getElementById('financial-risk-policies-select-all')?.addEventListener('click', () => {
  toggleGroup('.financial-risk-policy-checkbox');
});

document.getElementById('housing-policies-select-all')?.addEventListener('click', () => {
  toggleGroup('.housing-policy-checkbox');
});

document.getElementById('life-pension-contracts-select-all')?.addEventListener('click', () => {
  toggleGroup('.life-pension-contract-checkbox');
});

document.getElementById('patrimonial-policies-select-all')?.addEventListener('click', () => {
  toggleGroup('.patrimonial-policy-checkbox');
});