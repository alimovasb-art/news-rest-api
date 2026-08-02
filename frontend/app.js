/**
 * NewsPulse — Clean Modern SPA Controller
 * Connects directly to Go REST API & PostgreSQL
 * Supports Dark/Light Theme, Scrollable Reader, and Drag & Drop Media.
 */

const API_BASE_URL = 'http://localhost:8080';

const AppState = {
  token: localStorage.getItem('np_token') || null,
  currentUser: null,
  theme: localStorage.getItem('np_theme') || 'dark',
  page: 1,
  limit: 6,
  totalLoaded: 0,
  searchQuery: '',
  filterAuthorId: null,
  activeTab: 'all', // 'all' | 'my'
  currentReadingArticle: null,
  selectedImageFile: null,
  searchDebounceTimer: null
};

// ==================== INITIALIZATION ====================
document.addEventListener('DOMContentLoaded', async () => {
  initTheme();
  setupEventListeners();
  setupDragAndDrop();

  if (AppState.token) {
    await verifyAuthSession();
  } else {
    renderAuthNavigation();
  }

  loadNewsFeed();
});

// ==================== THEME MANAGEMENT ====================
function initTheme() {
  const savedTheme = localStorage.getItem('np_theme');
  if (savedTheme) {
    AppState.theme = savedTheme;
  } else if (window.matchMedia && window.matchMedia('(prefers-color-scheme: light)').matches) {
    AppState.theme = 'light';
  } else {
    AppState.theme = 'dark';
  }

  applyTheme(AppState.theme);
}

function applyTheme(theme) {
  document.documentElement.setAttribute('data-theme', theme);
  const icon = document.getElementById('themeIcon');
  if (icon) {
    if (theme === 'dark') {
      icon.className = 'fa-solid fa-sun';
      icon.parentElement.title = 'Switch to Light Theme';
    } else {
      icon.className = 'fa-solid fa-moon';
      icon.parentElement.title = 'Switch to Dark Theme';
    }
  }
}

function toggleTheme() {
  AppState.theme = AppState.theme === 'dark' ? 'light' : 'dark';
  localStorage.setItem('np_theme', AppState.theme);
  applyTheme(AppState.theme);
  showToast(`Switched to ${AppState.theme} theme`, 'info');
}

// ==================== EVENT LISTENERS ====================
function setupEventListeners() {
  document.querySelectorAll('.modal-overlay').forEach(modal => {
    modal.addEventListener('click', (e) => {
      if (e.target === modal) {
        closeModal(modal.id);
      }
    });
  });

  document.addEventListener('keydown', (e) => {
    if (e.key === 'Escape') {
      const activeModal = document.querySelector('.modal-overlay.active');
      if (activeModal) closeModal(activeModal.id);
    }
  });
}

// ==================== DRAG & DROP FOR COVER ====================
function setupDragAndDrop() {
  const dropzone = document.getElementById('imageDropzone');
  if (!dropzone) return;

  ['dragenter', 'dragover'].forEach(name => {
    dropzone.addEventListener(name, (e) => {
      e.preventDefault();
      e.stopPropagation();
      dropzone.classList.add('dragover');
    });
  });

  ['dragleave', 'drop'].forEach(name => {
    dropzone.addEventListener(name, (e) => {
      e.preventDefault();
      e.stopPropagation();
      dropzone.classList.remove('dragover');
    });
  });

  dropzone.addEventListener('drop', (e) => {
    const files = e.dataTransfer.files;
    if (files && files.length > 0) {
      processSelectedImage(files[0]);
    }
  });
}

// ==================== API HELPER ====================
async function apiRequest(endpoint, options = {}) {
  const url = `${API_BASE_URL}${endpoint}`;
  const headers = options.headers || {};

  if (AppState.token) {
    headers['Authorization'] = `Bearer ${AppState.token}`;
  }

  if (options.body && !(options.body instanceof FormData)) {
    headers['Content-Type'] = 'application/json';
  }

  try {
    const response = await fetch(url, { ...options, headers });
    const data = await response.json().catch(() => null);

    if (!response.ok) {
      if (response.status === 401) {
        handleLogout(false);
      }
      const errorMsg = (data && (data.error || data.message)) || `Server returned ${response.status}`;
      throw new Error(errorMsg);
    }

    return data;
  } catch (err) {
    console.error(`[API] ${endpoint}:`, err);
    throw err;
  }
}

// ==================== AUTH LOGIC ====================
async function verifyAuthSession() {
  try {
    const res = await apiRequest('/auth/me');
    if (res && res.data) {
      AppState.currentUser = res.data;
      renderAuthNavigation();
    } else {
      handleLogout(false);
    }
  } catch (err) {
    handleLogout(false);
  }
}

function renderAuthNavigation() {
  const container = document.getElementById('navAuthContainer');
  const myArticlesTab = document.getElementById('tabMyArticles');

  if (AppState.currentUser) {
    const initials = (AppState.currentUser.first_name[0] || 'U') + (AppState.currentUser.last_name[0] || '');
    
    container.innerHTML = `
      <button class="user-profile-btn" onclick="openProfileModal()" title="Account Settings">
        <div class="user-avatar-dot">${escapeHtml(initials.toUpperCase())}</div>
        <span>${escapeHtml(AppState.currentUser.first_name)}</span>
      </button>
      <button class="btn-auth-logout" onclick="handleLogout(true)" title="Sign Out">
        <i class="fa-solid fa-arrow-right-from-bracket"></i>
      </button>
    `;

    if (myArticlesTab) myArticlesTab.style.display = 'inline-flex';
  } else {
    container.innerHTML = `
      <button class="btn-auth-login" onclick="openModal('loginModal')">Sign In</button>
      <button class="btn-auth-signup" onclick="openModal('signupModal')">Sign Up</button>
    `;

    if (myArticlesTab) myArticlesTab.style.display = 'none';
  }
}

async function handleLoginSubmit(event) {
  event.preventDefault();
  const btn = document.getElementById('btnLoginSubmit');
  const email = document.getElementById('loginEmail').value.trim();
  const password = document.getElementById('loginPassword').value;

  btn.disabled = true;
  btn.textContent = 'Signing in...';

  try {
    const res = await apiRequest('/auth/login', {
      method: 'POST',
      body: JSON.stringify({ email, password })
    });

    if (res && res.data && res.data.token) {
      AppState.token = res.data.token;
      localStorage.setItem('np_token', AppState.token);
      await verifyAuthSession();
      closeModal('loginModal');
      document.getElementById('loginForm').reset();
      showToast('Signed in successfully!', 'success');
      loadNewsFeed();
    }
  } catch (err) {
    showToast(err.message, 'error');
  } finally {
    btn.disabled = false;
    btn.textContent = 'Sign In';
  }
}

async function handleSignupSubmit(event) {
  event.preventDefault();
  const btn = document.getElementById('btnSignupSubmit');
  const firstName = document.getElementById('signupFirstName').value.trim();
  const lastName = document.getElementById('signupLastName').value.trim();
  const email = document.getElementById('signupEmail').value.trim();
  const password = document.getElementById('signupPassword').value;

  btn.disabled = true;
  btn.textContent = 'Creating account...';

  try {
    await apiRequest('/auth/signup', {
      method: 'POST',
      body: JSON.stringify({
        first_name: firstName,
        last_name: lastName,
        email: email,
        password: password
      })
    });

    showToast('Account created! Logging in...', 'success');
    closeModal('signupModal');
    document.getElementById('signupForm').reset();

    const loginRes = await apiRequest('/auth/login', {
      method: 'POST',
      body: JSON.stringify({ email, password })
    });

    if (loginRes && loginRes.data && loginRes.data.token) {
      AppState.token = loginRes.data.token;
      localStorage.setItem('np_token', AppState.token);
      await verifyAuthSession();
      loadNewsFeed();
    }
  } catch (err) {
    showToast(err.message, 'error');
  } finally {
    btn.disabled = false;
    btn.textContent = 'Create Account';
  }
}

function handleLogout(notify = true) {
  AppState.token = null;
  AppState.currentUser = null;
  localStorage.removeItem('np_token');
  renderAuthNavigation();
  if (AppState.activeTab === 'my') switchFeedTab('all');
  else loadNewsFeed();
  if (notify) showToast('Signed out.', 'info');
}

// ==================== PROFILE SETTINGS ====================
function openProfileModal() {
  if (!AppState.currentUser) return;
  document.getElementById('profileFirstName').value = AppState.currentUser.first_name || '';
  document.getElementById('profileLastName').value = AppState.currentUser.last_name || '';
  document.getElementById('profileEmail').value = AppState.currentUser.email || '';
  document.getElementById('profilePassword').value = '';
  openModal('profileModal');
}

async function handleProfileUpdate(event) {
  event.preventDefault();
  const btn = document.getElementById('btnSaveProfile');
  const firstName = document.getElementById('profileFirstName').value.trim();
  const lastName = document.getElementById('profileLastName').value.trim();
  const email = document.getElementById('profileEmail').value.trim();
  const password = document.getElementById('profilePassword').value;

  const payload = {};
  if (firstName) payload.first_name = firstName;
  if (lastName) payload.last_name = lastName;
  if (email) payload.email = email;
  if (password) payload.password = password;

  btn.disabled = true;
  btn.textContent = 'Saving...';

  try {
    const res = await apiRequest('/users', {
      method: 'PATCH',
      body: JSON.stringify(payload)
    });

    if (res && res.data) {
      AppState.currentUser = res.data;
      renderAuthNavigation();
      closeModal('profileModal');
      showToast('Profile updated!', 'success');
      loadNewsFeed();
    }
  } catch (err) {
    showToast(err.message, 'error');
  } finally {
    btn.disabled = false;
    btn.textContent = 'Save Changes';
  }
}

async function handleDeleteMyAccount() {
  if (!AppState.currentUser) return;
  if (!confirm('Are you sure you want to delete your account?')) return;

  try {
    await apiRequest(`/users/${AppState.currentUser.id}`, { method: 'DELETE' });
    closeModal('profileModal');
    handleLogout(false);
    showToast('Account deleted.', 'info');
  } catch (err) {
    showToast(err.message, 'error');
  }
}

// ==================== NEWS FEED ====================
async function loadNewsFeed() {
  const grid = document.getElementById('newsGrid');
  const refreshIcon = document.getElementById('refreshIcon');
  if (refreshIcon) refreshIcon.classList.add('fa-spin');

  grid.innerHTML = Array(AppState.limit).fill(0).map(() => `
    <div class="skeleton-card">
      <div class="skeleton-shimmer"></div>
    </div>
  `).join('');

  let queryParams = `?page=${AppState.page}&limit=${AppState.limit}`;
  if (AppState.searchQuery) queryParams += `&title=${encodeURIComponent(AppState.searchQuery)}`;
  if (AppState.activeTab === 'my' && AppState.currentUser) queryParams += `&author_id=${AppState.currentUser.id}`;
  else if (AppState.filterAuthorId) queryParams += `&author_id=${AppState.filterAuthorId}`;

  try {
    const res = await apiRequest(`/news${queryParams}`);
    const newsList = (res && res.data) || [];
    AppState.totalLoaded = newsList.length;

    renderNewsCards(newsList);
    updatePaginationUI();
  } catch (err) {
    grid.innerHTML = `
      <div class="empty-state">
        <div class="empty-icon"><i class="fa-solid fa-triangle-exclamation"></i></div>
        <h3 class="empty-title">Cannot connect to Go REST API</h3>
        <p class="empty-desc">${escapeHtml(err.message || 'Make sure the Go server is running on http://localhost:8080')}</p>
        <button class="btn-primary" onclick="loadNewsFeed()">Retry</button>
      </div>
    `;
  } finally {
    if (refreshIcon) refreshIcon.classList.remove('fa-spin');
  }
}

function renderNewsCards(newsList) {
  const grid = document.getElementById('newsGrid');

  if (!newsList || newsList.length === 0) {
    grid.innerHTML = `
      <div class="empty-state">
        <div class="empty-icon"><i class="fa-regular fa-newspaper"></i></div>
        <h3 class="empty-title">No Stories Found</h3>
        <p class="empty-desc">
          ${AppState.searchQuery ? `No articles matching "${escapeHtml(AppState.searchQuery)}"` : 'There are currently no articles in this feed.'}
        </p>
        <button class="btn-primary" onclick="handleNewArticleClick()">Write First Story</button>
      </div>
    `;
    return;
  }

  const gradientCovers = [
    'linear-gradient(135deg, #1e1b4b, #312e81)',
    'linear-gradient(135deg, #022c22, #064e3b)',
    'linear-gradient(135deg, #450a0a, #7f1d1d)',
    'linear-gradient(135deg, #172554, #1e3a8a)',
    'linear-gradient(135deg, #3b0764, #581c87)'
  ];

  grid.innerHTML = newsList.map((item, idx) => {
    const author = item.author || {};
    const authorName = (author.first_name ? `${author.first_name} ${author.last_name || ''}` : `Author #${item.author_id}`).trim();
    const initials = (author.first_name ? author.first_name[0] : 'A') + (author.last_name ? author.last_name[0] : '');
    const timeAgo = formatRelativeTime(item.created_at);

    let mediaHtml = '';
    if (item.image_url && item.image_url.trim() !== '') {
      const fullImg = item.image_url.startsWith('http') ? item.image_url : `${API_BASE_URL}${item.image_url}`;
      mediaHtml = `<img class="card-img" src="${escapeHtml(fullImg)}" alt="${escapeHtml(item.title)}" loading="lazy" onerror="this.parentElement.innerHTML='<div class=\\'card-fallback-cover\\' style=\\'background:${gradientCovers[idx % gradientCovers.length]};\\'>📰</div>'">`;
    } else {
      mediaHtml = `<div class="card-fallback-cover" style="background:${gradientCovers[idx % gradientCovers.length]};"><i class="fa-solid fa-newspaper"></i></div>`;
    }

    return `
      <article class="article-card" onclick="openArticleReader(${item.id})">
        <div class="card-media-wrap">
          ${mediaHtml}
          <div class="card-views-badge"><i class="fa-regular fa-eye"></i> <span>${item.views || 0}</span></div>
        </div>

        <div class="card-body">
          <div>
            <div class="card-author-bar">
              <div class="author-pill" onclick="event.stopPropagation(); filterByAuthor(${item.author_id}, '${escapeHtml(authorName)}')">
                <div class="author-dot">${escapeHtml(initials.toUpperCase())}</div>
                <span>${escapeHtml(authorName)}</span>
              </div>
              <span class="card-date">${timeAgo}</span>
            </div>

            <h3 class="card-title">${escapeHtml(item.title)}</h3>
            <p class="card-summary">${escapeHtml(item.short_description || item.description)}</p>
          </div>

          <div class="card-footer">
            <span class="card-tag">Tech &bull; #${item.id}</span>
            <span class="card-read-link"><span>Read</span> <i class="fa-solid fa-arrow-right"></i></span>
          </div>
        </div>
      </article>
    `;
  }).join('');
}

// ==================== READER MODAL ====================
async function openArticleReader(newsId) {
  try {
    const res = await apiRequest(`/news/${newsId}`);
    if (!res || !res.data) return;

    const article = res.data;
    AppState.currentReadingArticle = article;

    const author = article.author || {};
    const authorName = (author.first_name ? `${author.first_name} ${author.last_name || ''}` : `Author #${article.author_id}`).trim();
    const initials = (author.first_name ? author.first_name[0] : 'A') + (author.last_name ? author.last_name[0] : '');

    document.getElementById('detailTitle').textContent = article.title;
    document.getElementById('detailViews').textContent = article.views;
    document.getElementById('detailDate').textContent = formatFullDate(article.created_at);
    document.getElementById('detailAuthorName').textContent = authorName;
    document.getElementById('detailAuthorEmail').textContent = author.email || `author_${article.author_id}@restapi.local`;
    document.getElementById('detailAuthorAvatar').textContent = initials.toUpperCase();
    document.getElementById('detailShortDesc').textContent = article.short_description;
    document.getElementById('detailContent').textContent = article.description;

    const imgWrap = document.getElementById('detailImageWrap');
    const imgEl = document.getElementById('detailImage');
    if (article.image_url && article.image_url.trim() !== '') {
      const fullImg = article.image_url.startsWith('http') ? article.image_url : `${API_BASE_URL}${article.image_url}`;
      imgEl.src = fullImg;
      imgWrap.style.display = 'block';
    } else {
      imgWrap.style.display = 'none';
    }

    const authorActions = document.getElementById('detailAuthorActions');
    if (AppState.currentUser && AppState.currentUser.id === article.author_id) {
      authorActions.style.display = 'flex';
    } else {
      authorActions.style.display = 'none';
    }

    openModal('detailModal');
  } catch (err) {
    showToast(err.message, 'error');
  }
}

function triggerFilterCurrentAuthor() {
  if (AppState.currentReadingArticle) {
    const author = AppState.currentReadingArticle.author || {};
    const name = author.first_name ? `${author.first_name} ${author.last_name}` : `Author #${AppState.currentReadingArticle.author_id}`;
    closeModal('detailModal');
    filterByAuthor(AppState.currentReadingArticle.author_id, name);
  }
}

function handleShareArticle() {
  if (navigator.clipboard) {
    navigator.clipboard.writeText(window.location.href);
    showToast('Article link copied to clipboard!', 'success');
  }
}

// ==================== CREATE & EDIT STORY ====================
function handleNewArticleClick() {
  if (!AppState.currentUser) {
    openModal('loginModal');
    showToast('Please sign in to publish a story.', 'info');
    return;
  }

  document.getElementById('newsForm').reset();
  document.getElementById('newsEditId').value = '';
  document.getElementById('newsModalTitle').textContent = 'Write a New Story';
  document.getElementById('newsModalSubtitle').textContent = 'Publish your thoughts and insights to the platform.';
  document.getElementById('btnSaveNewsText').textContent = 'Publish Story';
  document.getElementById('imageUploadGroup').style.display = 'block';
  
  removeSelectedImage();
  updateCharCounters();
  openModal('newsModal');
}

function triggerEditCurrentArticle() {
  if (!AppState.currentReadingArticle) return;
  const article = AppState.currentReadingArticle;

  closeModal('detailModal');

  document.getElementById('newsEditId').value = article.id;
  document.getElementById('newsTitle').value = article.title;
  document.getElementById('newsShortDesc').value = article.short_description;
  document.getElementById('newsFullDesc').value = article.description;

  document.getElementById('newsModalTitle').textContent = 'Edit Story';
  document.getElementById('newsModalSubtitle').textContent = `Editing article #${article.id}`;
  document.getElementById('btnSaveNewsText').textContent = 'Save Changes';

  updateCharCounters();
  removeSelectedImage();
  openModal('newsModal');
}

async function handleNewsSubmit(event) {
  event.preventDefault();
  const btn = document.getElementById('btnSaveNews');
  const editId = document.getElementById('newsEditId').value;
  const isEdit = Boolean(editId);

  const title = document.getElementById('newsTitle').value.trim();
  const shortDesc = document.getElementById('newsShortDesc').value.trim();
  const description = document.getElementById('newsFullDesc').value.trim();

  if (title.length < 3 || title.length > 50) {
    showToast('Title must be between 3 and 50 characters', 'error');
    return;
  }
  if (shortDesc.length < 10 || shortDesc.length > 150) {
    showToast('Short summary must be between 10 and 150 characters', 'error');
    return;
  }
  if (description.length < 15) {
    showToast('Article content must be at least 15 characters', 'error');
    return;
  }

  btn.disabled = true;
  btn.textContent = isEdit ? 'Saving...' : 'Publishing...';

  try {
    if (isEdit) {
      await apiRequest(`/news/${editId}`, {
        method: 'PUT',
        body: JSON.stringify({ title, short_description: shortDesc, description })
      });
      showToast('Article updated!', 'success');
    } else {
      const formData = new FormData();
      formData.append('title', title);
      formData.append('short_description', shortDesc);
      formData.append('description', description);

      if (AppState.selectedImageFile) {
        formData.append('image', AppState.selectedImageFile);
      }

      await apiRequest('/news', {
        method: 'POST',
        body: formData
      });
      showToast('Story published successfully!', 'success');
    }

    closeModal('newsModal');
    loadNewsFeed();
  } catch (err) {
    showToast(err.message, 'error');
  } finally {
    btn.disabled = false;
    btn.innerHTML = `<span id="btnSaveNewsText">${isEdit ? 'Save Changes' : 'Publish Story'}</span>`;
  }
}

async function triggerDeleteCurrentArticle() {
  if (!AppState.currentReadingArticle) return;
  if (!confirm('Are you sure you want to delete this story?')) return;

  try {
    await apiRequest(`/news/${AppState.currentReadingArticle.id}`, { method: 'DELETE' });
    closeModal('detailModal');
    showToast('Story deleted.', 'info');
    loadNewsFeed();
  } catch (err) {
    showToast(err.message, 'error');
  }
}

// ==================== IMAGE PICKER & PREVIEW ====================
function handleImageSelected(event) {
  const file = event.target.files[0];
  if (file) processSelectedImage(file);
}

function processSelectedImage(file) {
  if (file.size > 10 * 1024 * 1024) {
    showToast('Image size exceeds 10MB', 'error');
    return;
  }
  if (!file.type.startsWith('image/')) {
    showToast('Invalid file format. Please select PNG or JPG.', 'error');
    return;
  }

  AppState.selectedImageFile = file;

  const reader = new FileReader();
  reader.onload = (e) => {
    document.getElementById('imagePreviewImg').src = e.target.result;
    document.getElementById('previewFileName').textContent = file.name;
    document.getElementById('dropzoneEmpty').style.display = 'none';
    document.getElementById('dropzonePreview').style.display = 'block';
  };
  reader.readAsDataURL(file);
}

function removeSelectedImage() {
  AppState.selectedImageFile = null;
  const fileInput = document.getElementById('newsImageFile');
  if (fileInput) fileInput.value = '';
  document.getElementById('dropzoneEmpty').style.display = 'flex';
  document.getElementById('dropzonePreview').style.display = 'none';
}

// ==================== AUTHORS DIRECTORY ====================
async function openAuthorsModal() {
  openModal('authorsModal');
  const container = document.getElementById('authorsListContainer');
  container.innerHTML = `<div style="grid-column: 1/-1; text-align: center; color: var(--text-muted); padding: 2rem;">Loading authors...</div>`;

  try {
    const res = await apiRequest('/users');
    const users = (res && res.data) || [];

    if (users.length === 0) {
      container.innerHTML = `<div style="grid-column: 1/-1; text-align: center; color: var(--text-muted); padding: 2rem;">No authors found.</div>`;
      return;
    }

    container.innerHTML = users.map(u => {
      const initials = (u.first_name ? u.first_name[0] : 'U') + (u.last_name ? u.last_name[0] : '');
      const fullName = `${u.first_name || ''} ${u.last_name || ''}`.trim() || `User #${u.id}`;

      return `
        <div class="author-card-item">
          <div style="display: flex; align-items: center; gap: 0.75rem;">
            <div class="author-item-avatar">${escapeHtml(initials.toUpperCase())}</div>
            <div>
              <div style="font-weight: 600; font-size: 0.92rem; color: var(--text-white);">${escapeHtml(fullName)}</div>
              <div style="font-size: 0.78rem; color: var(--text-dim);">${escapeHtml(u.email || '')}</div>
            </div>
          </div>
          <button class="btn-author-filter" onclick="closeModal('authorsModal'); filterByAuthor(${u.id}, '${escapeHtml(fullName)}')">
            <span>Articles</span> <i class="fa-solid fa-arrow-right"></i>
          </button>
        </div>
      `;
    }).join('');
  } catch (err) {
    container.innerHTML = `<div style="grid-column: 1/-1; color: #ef4444; text-align: center; padding: 2rem;">Failed to load authors.</div>`;
  }
}

// ==================== SEARCH & FILTERS ====================
function handleSearchInput(event) {
  const val = event.target.value;
  const clearBtn = document.getElementById('searchClearBtn');
  if (clearBtn) clearBtn.style.display = val ? 'block' : 'none';

  clearTimeout(AppState.searchDebounceTimer);
  AppState.searchDebounceTimer = setTimeout(() => {
    AppState.searchQuery = val.trim();
    AppState.page = 1;
    loadNewsFeed();
  }, 300);
}

function clearSearchInput() {
  document.getElementById('searchInput').value = '';
  document.getElementById('searchClearBtn').style.display = 'none';
  AppState.searchQuery = '';
  AppState.page = 1;
  loadNewsFeed();
}

function toggleMobileSearch() {
  const drawer = document.getElementById('mobileSearchDrawer');
  drawer.classList.toggle('active');
  if (drawer.classList.contains('active')) {
    document.getElementById('mobileSearchInput').focus();
  }
}

function handleMobileSearch(event) {
  clearTimeout(AppState.searchDebounceTimer);
  AppState.searchDebounceTimer = setTimeout(() => {
    AppState.searchQuery = event.target.value.trim();
    AppState.page = 1;
    loadNewsFeed();
  }, 300);
}

function filterByAuthor(authorId, authorName) {
  AppState.filterAuthorId = authorId;
  AppState.page = 1;

  document.getElementById('activeFilterText').textContent = `Filtered by Author: ${authorName}`;
  document.getElementById('activeFilterBadge').style.display = 'flex';

  loadNewsFeed();
}

function clearAllFilters() {
  AppState.filterAuthorId = null;
  document.getElementById('activeFilterBadge').style.display = 'none';
  loadNewsFeed();
}

function switchFeedTab(tab) {
  AppState.activeTab = tab;
  AppState.page = 1;

  document.getElementById('tabAllFeed')?.classList.toggle('active', tab === 'all');
  document.getElementById('tabMyArticles')?.classList.toggle('active', tab === 'my');

  const heading = document.getElementById('feedHeading');
  const subheading = document.getElementById('feedSubheading');

  if (tab === 'my') {
    heading.textContent = 'My Articles';
    subheading.textContent = 'Stories written by you';
  } else {
    heading.textContent = 'Latest Stories';
    subheading.textContent = 'Discover articles on engineering, architecture, and design.';
  }

  loadNewsFeed();
}

function handleLimitChange(val) {
  AppState.limit = parseInt(val, 10) || 6;
  AppState.page = 1;
  loadNewsFeed();
}

function navigatePage(direction) {
  AppState.page = Math.max(1, AppState.page + direction);
  loadNewsFeed();
  window.scrollTo({ top: 0, behavior: 'smooth' });
}

function updatePaginationUI() {
  document.getElementById('pageIndicator').textContent = `Page ${AppState.page}`;
  document.getElementById('btnPrevPage').disabled = AppState.page <= 1;
  document.getElementById('btnNextPage').disabled = AppState.totalLoaded < AppState.limit;
}

function resetFeed() {
  AppState.page = 1;
  AppState.searchQuery = '';
  AppState.filterAuthorId = null;
  document.getElementById('searchInput').value = '';
  document.getElementById('searchClearBtn').style.display = 'none';
  document.getElementById('activeFilterBadge').style.display = 'none';
  switchFeedTab('all');
  window.scrollTo({ top: 0, behavior: 'smooth' });
}

// ==================== MODAL HELPERS ====================
function openModal(id) {
  const el = document.getElementById(id);
  if (el) {
    el.classList.add('active');
    document.body.style.overflow = 'hidden';
  }
}

function closeModal(id) {
  const el = document.getElementById(id);
  if (el) {
    el.classList.remove('active');
    document.body.style.overflow = '';
  }
}

function switchModals(from, to) {
  closeModal(from);
  setTimeout(() => openModal(to), 100);
}

// ==================== UTILS ====================
function showToast(msg, type = 'info') {
  const stack = document.getElementById('toastStack');
  if (!stack) return;

  const toast = document.createElement('div');
  toast.className = `toast-item toast-${type}`;
  toast.textContent = msg;

  stack.appendChild(toast);
  setTimeout(() => {
    toast.style.opacity = '0';
    toast.style.transform = 'translateX(100%)';
    toast.style.transition = 'all 0.3s ease';
    setTimeout(() => toast.remove(), 300);
  }, 3500);
}

function updateCharCount(inputId, countId, min, max) {
  const input = document.getElementById(inputId);
  const countEl = document.getElementById(countId);
  if (!input || !countEl) return;
  const len = input.value.length;
  countEl.textContent = `${len}/${max}`;
  countEl.style.color = len < min ? '#ef4444' : 'var(--text-dim)';
}

function updateCharCounters() {
  updateCharCount('newsTitle', 'titleCharCount', 3, 50);
  updateCharCount('newsShortDesc', 'shortDescCharCount', 10, 150);
  updateCharCount('newsFullDesc', 'fullDescCharCount', 15, 5000);
}

function formatRelativeTime(dateStr) {
  if (!dateStr) return 'Recently';
  const diffSec = Math.floor((new Date() - new Date(dateStr)) / 1000);
  if (diffSec < 60) return 'Just now';
  if (diffSec < 3600) return `${Math.floor(diffSec / 60)}m ago`;
  if (diffSec < 86400) return `${Math.floor(diffSec / 3600)}h ago`;
  return `${Math.floor(diffSec / 86400)}d ago`;
}

function formatFullDate(dateStr) {
  if (!dateStr) return 'N/A';
  return new Date(dateStr).toLocaleDateString(undefined, { year: 'numeric', month: 'long', day: 'numeric' });
}

function escapeHtml(str) {
  if (!str) return '';
  return String(str).replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;').replace(/"/g, '&quot;');
}
