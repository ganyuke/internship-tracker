document.addEventListener('DOMContentLoaded', () => {
  document.querySelectorAll('[data-confirm]').forEach((form) => {
    form.addEventListener('submit', (event) => {
      const message = form.getAttribute('data-confirm') || 'Are you sure?';
      if (!window.confirm(message)) event.preventDefault();
    });
  });

  document.querySelectorAll('.clickable-row[data-href]').forEach((row) => {
    row.addEventListener('click', (event) => {
      if (event.target.closest('a, button, input, select, textarea')) return;
      window.location.href = row.dataset.href;
    });
  });

  const form = document.querySelector('[data-autofill-form]');
  if (!form) return;

  const link = form.querySelector('#link');
  const company = form.querySelector('#company');
  const role = form.querySelector('#role');
  if (!link || !company || !role) return;

  link.addEventListener('blur', () => {
    const value = link.value.trim();
    if (!value) return;

    if (!company.value.trim()) {
      const companyMatch = value.match(/https?:\/\/(?:www\.)?([^\.\/]+)\./i);
      if (companyMatch) {
        company.value = companyMatch[1].charAt(0).toUpperCase() + companyMatch[1].slice(1);
      }
    }

    if (!role.value.trim()) {
      const roleMatch = value.match(/([^\/]+)(?:-intern|software|engineer|designer|developer)[^\/]*$/i);
      if (roleMatch) {
        role.value = roleMatch[0]
          .split('-')
          .filter(Boolean)
          .map((word) => word.charAt(0).toUpperCase() + word.slice(1))
          .join(' ');
      }
    }
  });
});
