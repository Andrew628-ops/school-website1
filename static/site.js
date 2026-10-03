'use strict';
const menuButton = document.querySelector('.menu-toggle');
const mainNav = document.querySelector('.main-nav');
function closeMenu() {
  mainNav?.classList.remove('is-open');
  menuButton?.setAttribute('aria-expanded', 'false');
  menuButton?.setAttribute('aria-label', 'Open navigation');
}
menuButton?.addEventListener('click', () => {
  const open = menuButton.getAttribute('aria-expanded') !== 'true';
  mainNav.classList.toggle('is-open', open);
  menuButton.setAttribute('aria-expanded', String(open));
  menuButton.setAttribute('aria-label', open ? 'Close navigation' : 'Open navigation');
});
document.addEventListener('keydown', event => { if (event.key === 'Escape') closeMenu(); });
mainNav?.addEventListener('click', event => { if (event.target.closest('a')) closeMenu(); });
const searchDialog = document.querySelector('#search-dialog');
document.querySelector('.search-toggle')?.addEventListener('click', event => {
  if (typeof searchDialog?.showModal !== 'function') return;
  event.preventDefault(); searchDialog.showModal();
  document.querySelector('#dialog-query')?.focus();
});
document.querySelector('.dialog-close')?.addEventListener('click', () => searchDialog.close());
searchDialog?.addEventListener('click', event => { if (event.target === searchDialog) searchDialog.close(); });
document.querySelectorAll('form[data-confirm]').forEach(form => {
  form.addEventListener('submit', event => { if (!window.confirm(form.dataset.confirm)) event.preventDefault(); });
});
document.querySelector('.print-button')?.addEventListener('click', () => window.print());
document.querySelectorAll('.faculty-photo img').forEach(img => {
  const replace = () => {
    const fallback = document.createElement('span');
    fallback.className = 'faculty-initials';
    const name = img.closest('.faculty-card').querySelector('h3').textContent;
    fallback.textContent = name.split(' ').filter(part => !['Dr.', 'Prof.'].includes(part)).slice(0, 2).map(part => part[0]).join('');
    img.replaceWith(fallback);
  };
  img.addEventListener('error', replace, {once: true});
  if (img.complete && img.naturalWidth === 0) replace();
});
