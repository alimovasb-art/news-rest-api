/* ==========================================================================
   NewsPulse - Modern Frontend Application Logic
   Connects to Go REST API running on http://localhost:8080
   ========================================================================== */

const API_BASE_URL = 'http://localhost:8080';

// App State
let state = {
  token: localStorage.getItem('jwt_token') || null,
  user: JSON.parse(localStorage.getItem('user_info')) || null,
  currentPage: 1,
  limit: 6,
  searchTitle: '',
  filterAuthorId: null,
  currentDetailArticle: null
};

// Initialize App on DOM Load
document.addEventListener('DOMContentLoaded', () => {
  renderNavActions();
  loadNewsFeed();
});

/* ==================== AUTHENTICATION & HEADERS ==================== */

function getAuthHeaders() {
  const headers = { 'Content-Type': 'application/json' };
  if (state.token) {
    headers['Authorization'] = `Bearer ${state.token}`;
  }
  return headers;
}

function parseJwt(token) {
  try {
    const base64Url = token.split('.')[1];
    const base64 = base64Url.replace(/-/g, '+').replace(/_/g, '/');
    const jsonPayload = decodeURIComponent(atob(base64).split('').map(c => {
      return '%' + ('00' + c.charCodeAt(0).toString(16)).slice(-2);
    }).join(''));
    return JSON.parse(jsonPayload);
  } catch (e) {
    return null;
  }
}

function renderNavActions() {
  const container = document.getElementById('navActions');
  const btnCreateNav = document.getElementById('btnCreateNewsNav');

  if (state.token && state.user) {
    if (btnCreateNav) btnCreateNav.style.display = 'inline-flex';
    container.innerHTML = `
      <div class="user-badge" onclick="openProfileModal()" title="View Profile">
        <div class="avatar-sm">${(state.user.first_name || 'U')[0].toUpperCase()}</div>
        <span class="user-name-text">${state.user.first_name} ${state.user.last_name || ''}</span>
      </div>
      <button class="btn-icon-secondary" onclick="handleLogout()" title="Logout">
        <i class="fa-solid fa-right-from-bracket"></i>
      </button>
    `;
  } else {
    if (btnCreateNav) btnCreateNav.style.display = 'none';
    container.innerHTML = `
      <button class="btn-secondary" onclick="openModal('loginModal')">Log In</button>
      <button class="btn-primary" onclick="openModal('signupModal')">Sign Up</button>
    `;
  }
}

async function handleLogin(e) {
  e.preventDefault();
  const email = document.getElementById('loginEmail').value.trim();
  const password = document.getElementById('loginPassword').value;

  try {
    const res = await fetch(`${API_BASE_URL}/auth/login`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ email, password })
    });
    const result = await res.json();

    if (result.success && result.data && result.data.token) {
      state.token = result.data.token;
      localStorage.setItem('jwt_token', state.token);

      const claims = parseJwt(state.token);
      if (claims) {
        state.user = {
          id: claims.author_id,
          first_name: claims.first_name,
          last_name: claims.last_name,
          email: claims.email
        };
        localStorage.setItem('user_info', JSON.stringify(state.user));
      }

      showToast('Logged in successfully!', 'success');
      closeModal('loginModal');
      renderNavActions();
      loadNewsFeed();
    } else {
      showToast(result.massage || result.message || 'Login failed', 'error');
    }
  } catch (err) {
    showToast('Failed to connect to backend server', 'error');
  }
}

async function handleSignup(e) {
  e.preventDefault();
  const first_name = document.getElementById('signupFirstName').value.trim();
  const last_name = document.getElementById('signupLastName').value.trim();
  const email = document.getElementById('signupEmail').value.trim();
  const password = document.getElementById('signupPassword').value;

  try {
    const res = await fetch(`${API_BASE_URL}/auth/signup`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ first_name, last_name, email, password })
    });
    const result = await res.json();

    if (result.success) {
      showToast('Account created! Please log in.', 'success');
      switchToModal('signupModal', 'loginModal');
      document.getElementById('loginEmail').value = email;
    } else {
      showToast(result.massage || result.message || 'Registration failed', 'error');
    }
  } catch (err) {
    showToast('Failed to connect to backend server', 'error');
  }
}

function handleLogout() {
  state.token = null;
  state.user = null;
  localStorage.removeItem('jwt_token');
  localStorage.removeItem('user_info');
  showToast('Logged out', 'success');
  renderNavActions();
  loadNewsFeed();
}

/* ==================== NEWS FEED & PAGINATION ==================== */

async function loadNewsFeed() {
  const grid = document.getElementById('newsGrid');
  grid.innerHTML = '<div class="loading-state" style="grid-column: 1/-1; text-align: center; padding: 3rem;"><i class="fa-solid fa-spinner fa-spin fa-2x" style="color: var(--accent-primary);"></i><p style="margin-top: 1rem; color: var(--text-muted);">Loading latest news...</p></div>';

  let url = `${API_BASE_URL}/news?page=${state.currentPage}&limit=${state.limit}`;
  if (state.searchTitle) {
    url += `&title=${encodeURIComponent(state.searchTitle)}`;
  }
  if (state.filterAuthorId) {
    url += `&author_id=${state.filterAuthorId}`;
  }

  try {
    const res = await fetch(url);
    const result = await res.json();

    if (result.success && Array.isArray(result.data)) {
      renderNewsGrid(result.data);
      updatePaginationControls(result.data.length);
    } else {
      grid.innerHTML = '<div style="grid-column: 1/-1; text-align: center; padding: 3rem; color: var(--text-muted);"><i class="fa-regular fa-folder-open fa-2x"></i><p style="margin-top: 1rem;">No news articles found.</p></div>';
    }
  } catch (err) {
    grid.innerHTML = '<div style="grid-column: 1/-1; text-align: center; padding: 3rem; color: #ef4444;"><i class="fa-solid fa-triangle-exclamation fa-2x"></i><p style="margin-top: 1rem;">Unable to load news feed. Is backend server running on port 8080?</p></div>';
  }
}

function renderNewsGrid(articles) {
  const grid = document.getElementById('newsGrid');

  if (articles.length === 0) {
    grid.innerHTML = '<div style="grid-column: 1/-1; text-align: center; padding: 3rem; color: var(--text-muted);"><i class="fa-regular fa-folder-open fa-2x"></i><p style="margin-top: 1rem;">No news articles found.</p></div>';
    return;
  }

  grid.innerHTML = articles.map(item => {
    const authorName = item.author ? `${item.author.first_name} ${item.author.last_name || ''}` : 'Unknown Author';
    const authorInitial = item.author && item.author.first_name ? item.author.first_name[0].toUpperCase() : 'A';
    const createdDate = item.created_at ? new Date(item.created_at).toLocaleDateString(undefined, { month: 'short', day: 'numeric', year: 'numeric' }) : 'Recently';

    return `
      <div class="news-card" onclick="openDetailModal(${item.id})">
        <div class="card-top">
          <div class="card-author-row">
            <div class="card-author">
              <div class="avatar-sm" style="width:26px; height:26px; font-size:0.75rem;">${authorInitial}</div>
              <span class="author-name-btn" onclick="event.stopPropagation(); filterByAuthor(${item.author_id || (item.author ? item.author.id : null)}, '${authorName}')">${authorName}</span>
            </div>
            <span style="font-size:0.78rem; color:var(--text-dim);"><i class="fa-regular fa-clock"></i> ${createdDate}</span>
          </div>

          <h3 class="card-title">${escapeHtml(item.title)}</h3>
          <p class="card-desc">${escapeHtml(item.short_description)}</p>
        </div>

        <div class="card-bottom">
          <div class="card-meta-item">
            <i class="fa-regular fa-eye"></i> ${item.views || 0} views
          </div>
          <div style="color: var(--accent-primary); font-weight:600; font-size:0.85rem;">
            Read Story <i class="fa-solid fa-arrow-right" style="font-size:0.75rem;"></i>
          </div>
        </div>
      </div>
    `;
  }).join('');
}

function updatePaginationControls(itemsReturned) {
  document.getElementById('pageIndicator').innerText = `Page ${state.currentPage}`;
  document.getElementById('btnPrevPage').disabled = (state.currentPage === 1);
  document.getElementById('btnNextPage').disabled = (itemsReturned < state.limit);
}

function changePage(delta) {
  state.currentPage += delta;
  if (state.currentPage < 1) state.currentPage = 1;
  loadNewsFeed();
}

/* ==================== SEARCH & FILTERS ==================== */

function handleSearchKey(e) {
  if (e.key === 'Enter') {
    triggerSearch();
  }
}

function triggerSearch() {
  const val = document.getElementById('searchInput').value.trim();
  state.searchTitle = val;
  state.currentPage = 1;

  updateFilterTag();
  loadNewsFeed();
}

function filterByAuthor(authorId, authorName) {
  if (!authorId) return;
  state.filterAuthorId = authorId;
  state.currentPage = 1;

  updateFilterTag(`Author: ${authorName}`);
  loadNewsFeed();
}

function updateFilterTag(customText) {
  const tag = document.getElementById('activeFilterTag');
  if (state.filterAuthorId || state.searchTitle) {
    const text = customText || (state.searchTitle ? `Search: "${state.searchTitle}"` : '');
    tag.innerHTML = `${text} <i class="fa-solid fa-xmark" onclick="clearFilters()" style="cursor:pointer; margin-left:4px;"></i>`;
    tag.style.display = 'inline-flex';
  } else {
    tag.style.display = 'none';
  }
}

function clearFilters() {
  state.searchTitle = '';
  state.filterAuthorId = null;
  state.currentPage = 1;
  document.getElementById('searchInput').value = '';
  updateFilterTag();
  loadNewsFeed();
}

function resetFeed() {
  clearFilters();
}

/* ==================== ARTICLE DETAIL VIEW ==================== */

async function openDetailModal(id) {
  openModal('detailModal');
  document.getElementById('detailTitle').innerText = 'Loading article...';
  document.getElementById('detailShortDesc').innerText = '';
  document.getElementById('detailContent').innerText = '';
  document.getElementById('detailAuthorActions').style.display = 'none';

  try {
    const res = await fetch(`${API_BASE_URL}/news/${id}`);
    const result = await res.json();

    if (result.success && result.data) {
      const news = result.data;
      state.currentDetailArticle = news;

      document.getElementById('detailTitle').innerText = news.title;
      document.getElementById('detailShortDesc').innerText = news.short_description;
      document.getElementById('detailContent').innerText = news.description;
      document.getElementById('detailViews').innerText = news.views || 0;

      const authorName = news.author ? `${news.author.first_name} ${news.author.last_name || ''}` : 'Unknown';
      const authorEmail = news.author ? news.author.email : '';
      const authorInitial = news.author && news.author.first_name ? news.author.first_name[0].toUpperCase() : 'A';

      document.getElementById('detailAuthorAvatar').innerText = authorInitial;
      document.getElementById('detailAuthorName').innerText = authorName;
      document.getElementById('detailAuthorEmail').innerText = authorEmail;
      document.getElementById('detailDate').innerText = news.created_at ? new Date(news.created_at).toLocaleDateString() : 'Recently';

      // Check if logged in user is the author
      if (state.user && (state.user.id === news.author_id || (news.author && state.user.id === news.author.id))) {
        document.getElementById('detailAuthorActions').style.display = 'flex';
      }
    } else {
      showToast(result.massage || result.message || 'Article not found', 'error');
      closeModal('detailModal');
    }
  } catch (err) {
    showToast('Failed to fetch article details', 'error');
    closeModal('detailModal');
  }
}

/* ==================== CREATE / EDIT ARTICLE ==================== */

function openCreateNewsModal() {
  if (!state.token) {
    showToast('Please log in to create articles', 'error');
    openModal('loginModal');
    return;
  }
  document.getElementById('newsEditId').value = '';
  document.getElementById('newsModalTitle').innerText = 'Create News Article';
  document.getElementById('newsForm').reset();
  openModal('newsModal');
}

function triggerEditCurrentArticle() {
  if (!state.currentDetailArticle) return;
  const news = state.currentDetailArticle;

  closeModal('detailModal');
  document.getElementById('newsEditId').value = news.id;
  document.getElementById('newsModalTitle').innerText = 'Edit Article';
  document.getElementById('newsTitle').value = news.title;
  document.getElementById('newsShortDesc').value = news.short_description;
  document.getElementById('newsFullDesc').value = news.description;
  openModal('newsModal');
}

async function handleSaveNews(e) {
  e.preventDefault();
  const editId = document.getElementById('newsEditId').value;
  const title = document.getElementById('newsTitle').value.trim();
  const short_description = document.getElementById('newsShortDesc').value.trim();
  const description = document.getElementById('newsFullDesc').value.trim();

  const isEdit = !!editId;
  const url = isEdit ? `${API_BASE_URL}/news/${editId}` : `${API_BASE_URL}/news`;
  const method = isEdit ? 'PUT' : 'POST';

  try {
    const res = await fetch(url, {
      method: method,
      headers: getAuthHeaders(),
      body: JSON.stringify({ title, short_description, description })
    });
    const result = await res.json();

    if (result.success) {
      showToast(isEdit ? 'Article updated successfully!' : 'Article published successfully!', 'success');
      closeModal('newsModal');
      loadNewsFeed();
    } else {
      showToast(result.massage || result.message || 'Failed to save article', 'error');
    }
  } catch (err) {
    showToast('Network error while saving article', 'error');
  }
}

async function triggerDeleteCurrentArticle() {
  if (!state.currentDetailArticle) return;
  const news = state.currentDetailArticle;

  if (!confirm(`Are you sure you want to delete "${news.title}"?`)) return;

  try {
    const res = await fetch(`${API_BASE_URL}/news/${news.id}`, {
      method: 'DELETE',
      headers: getAuthHeaders()
    });
    const result = await res.json();

    if (result.success) {
      showToast('Article deleted successfully!', 'success');
      closeModal('detailModal');
      loadNewsFeed();
    } else {
      showToast(result.massage || result.message || 'Failed to delete article', 'error');
    }
  } catch (err) {
    showToast('Network error while deleting article', 'error');
  }
}

/* ==================== USER PROFILE ==================== */

function openProfileModal() {
  if (!state.user) return;
  document.getElementById('profileFirstName').value = state.user.first_name || '';
  document.getElementById('profileLastName').value = state.user.last_name || '';
  document.getElementById('profileEmail').value = state.user.email || '';
  document.getElementById('profilePassword').value = '';
  openModal('profileModal');
}

async function handleUpdateProfile(e) {
  e.preventDefault();
  if (!state.user) return;

  const first_name = document.getElementById('profileFirstName').value.trim();
  const last_name = document.getElementById('profileLastName').value.trim();
  const email = document.getElementById('profileEmail').value.trim();
  const password = document.getElementById('profilePassword').value;

  const body = {};
  if (first_name) body.first_name = first_name;
  if (last_name) body.last_name = last_name;
  if (email) body.email = email;
  if (password) body.password = password;

  try {
    const res = await fetch(`${API_BASE_URL}/users/${state.user.id}`, {
      method: 'PATCH',
      headers: getAuthHeaders(),
      body: JSON.stringify(body)
    });
    const result = await res.json();

    if (result.success && result.data) {
      state.user = {
        ...state.user,
        first_name: result.data.first_name || state.user.first_name,
        last_name: result.data.last_name || state.user.last_name,
        email: result.data.email || state.user.email
      };
      localStorage.setItem('user_info', JSON.stringify(state.user));

      showToast('Profile updated successfully!', 'success');
      closeModal('profileModal');
      renderNavActions();
    } else {
      showToast(result.massage || result.message || 'Failed to update profile', 'error');
    }
  } catch (err) {
    showToast('Network error while updating profile', 'error');
  }
}

/* ==================== UTILS & MODALS ==================== */

function openModal(id) {
  const modal = document.getElementById(id);
  if (modal) modal.classList.add('active');
}

function closeModal(id) {
  const modal = document.getElementById(id);
  if (modal) modal.classList.remove('active');
}

function switchToModal(closeId, openId) {
  closeModal(closeId);
  openModal(openId);
}

function showToast(message, type = 'info') {
  const container = document.getElementById('toastContainer');
  const toast = document.createElement('div');
  toast.className = `toast toast-${type}`;
  toast.innerHTML = `
    <i class="fa-solid ${type === 'success' ? 'fa-circle-check' : 'fa-triangle-exclamation'}"></i>
    <span>${escapeHtml(message)}</span>
  `;
  container.appendChild(toast);

  setTimeout(() => {
    toast.remove();
  }, 4000);
}

function escapeHtml(str) {
  if (!str) return '';
  return str.replace(/&/g, '&amp;')
            .replace(/</g, '&lt;')
            .replace(/>/g, '&gt;')
            .replace(/"/g, '&quot;')
            .replace(/'/g, '&#039;');
}
