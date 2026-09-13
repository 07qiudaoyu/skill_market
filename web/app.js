/* ============================================================
   技能市 · 前端逻辑
   - 数据优先来自 FastAPI（同源 /api/v1/*），后端不可用时
     自动降级到仓库内真实样例数据
   - 搜索 / 分类 / 标签 / 排序 / 版本抽屉 / 下载 / 登录 / 上传
   - 上传表单携带 token 输入（后端改为请求头鉴权后移除）
   ============================================================ */
(function () {
  'use strict';

  var API = {
    search:   '/api/v1/skills/search',
    versions: '/api/v1/skills/find_versions',
    download: '/api/v1/skills/request_file',
    login:    '/api/v1/auth/login',
    register: '/api/v1/auth/register',
    upload:   '/api/v1/skills/upload_zip'
  };

  var LS_TOKEN = 'skm_token';
  var LS_NAME = 'skm_name';

  /* ---------- 样例数据（与仓库 data/skills、try/*_insert_try.py 对齐） ---------- */
  var SEED = {
    categories: [
      { slug: 'web-dev', name: 'Web开发' },
      { slug: 'data-ai', name: '数据与AI' },
      { slug: 'automation', name: '自动化工具' },
      { slug: 'productivity', name: '效率工具' },
      { slug: 'devops', name: '运维与部署' }
    ],
    skills: [
      {
        public_id: 'uuid-001',
        slug: 'auto-email-sender',
        display_name: '邮件自动发送工具',
        summary: '根据模板自动发送邮件，支持 SMTP 配置、HTML 模板渲染和附件发送。',
        category_name: 'Web开发',
        category_slug: 'web-dev',
        tags: ['email', 'automation', 'smtp', 'attachment'],
        download_count: 12,
        rating_avg: 4.5,
        created_at: '2026-09-01',
        excerpt: '1. 读取邮件模板和收件人列表。\n2. 通过 SMTP 连接发送邮件。\n3. 支持附件发送。\n4. 记录发送日志。',
        versions: [
          { id: 2, version: '1.1.0', size_bytes: 153600, file_name: 'auto-email-sender_v1.1.0.zip', created_at: '2026-09-01', founder_name: 'admin' },
          { id: 1, version: '1.0.0', size_bytes: 102400, file_name: 'auto-email-sender_v1.0.0.zip', created_at: '2026-08-20', founder_name: 'admin' }
        ]
      },
      {
        public_id: 'uuid-002',
        slug: 'daily-report-generator',
        display_name: '日报自动生成器',
        summary: '自动读取数据源并生成 Excel/PDF 格式的日报。',
        category_name: '自动化工具',
        category_slug: 'automation',
        tags: ['report', 'excel', 'pdf', 'automation'],
        download_count: 8,
        rating_avg: 4.0,
        created_at: '2026-08-28',
        excerpt: '1. 连接数据源（数据库/CSV/Excel）。\n2. 按日汇总数据。\n3. 生成 Excel 或 PDF 报告。',
        versions: [
          { id: 3, version: '1.0.0', size_bytes: 204800, file_name: 'daily-report-generator_v1.0.0.zip', created_at: '2026-08-28', founder_name: 'admin' }
        ]
      },
      {
        public_id: 'uuid-003',
        slug: 'web-data-crawler',
        display_name: '网页数据采集器',
        summary: '输入 URL 自动抓取网页内容并导出为结构化数据。',
        category_name: 'Web开发',
        category_slug: 'web-dev',
        tags: ['crawler', 'scraping', 'web', 'data'],
        download_count: 25,
        rating_avg: 4.8,
        created_at: '2026-08-25',
        excerpt: '1. 接收目标 URL。\n2. 解析 HTML 内容。\n3. 提取结构化数据。\n4. 导出为 JSON/CSV。',
        versions: [
          { id: 4, version: '1.0.0', size_bytes: 307200, file_name: 'web-data-crawler_v1.0.0.zip', created_at: '2026-08-25', founder_name: 'admin' }
        ]
      }
    ]
  };

  /* ---------- hero 指令卡：真实的 auto-email-sender SKILL.md ---------- */
  var HERO_LINES = [
    { text: '---', meta: true },
    { text: 'name: auto-email-sender', meta: true },
    { text: 'description: 根据模板自动发送邮件', meta: true },
    { text: 'license: MIT', meta: true },
    { text: '---', meta: true },
    { text: '# Auto Email Sender' },
    { text: '1. 读取邮件模板和收件人列表。' },
    { text: '2. 通过 SMTP 连接发送邮件。' },
    { text: '3. 支持附件发送。' },
    { text: '4. 记录发送日志。' }
  ];

  var state = {
    q: '',
    category: 'all',
    tag: '',
    sort: 'newest',
    source: 'seed',      // 'seed' | 'live'
    liveLocked: false,   // 后端确认不可用后锁定，避免反复失败
    openDrawers: {}      // public_id -> true
  };

  var versionsCache = {};  // public_id -> versions[]（live 模式懒加载）
  var lastItems = [];

  var $ = function (id) { return document.getElementById(id); };

  /* ---------- 工具函数 ---------- */
  function escapeHtml(s) {
    return String(s == null ? '' : s)
      .replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;')
      .replace(/"/g, '&quot;').replace(/'/g, '&#39;');
  }

  function formatSize(bytes) {
    if (bytes == null) return '—';
    if (bytes < 1024) return bytes + ' B';
    if (bytes < 1024 * 1024) return Math.round(bytes / 1024) + ' KB';
    return (bytes / 1024 / 1024).toFixed(1) + ' MB';
  }

  function formatDate(iso) {
    return String(iso || '').slice(0, 10) || '—';
  }

  function semverDesc(a, b) {
    var pa = String(a.version).split('.').map(Number);
    var pb = String(b.version).split('.').map(Number);
    for (var i = 0; i < 3; i++) {
      if ((pa[i] || 0) !== (pb[i] || 0)) return (pb[i] || 0) - (pa[i] || 0);
    }
    return 0;
  }

  function apiError(data, fallback) {
    if (data && typeof data.detail === 'string') return data.detail;
    if (data && Array.isArray(data.detail)) {
      return data.detail.map(function (d) { return d.msg || ''; }).join('；');
    }
    return fallback;
  }

  var toastTimer = null;
  function toast(html) {
    var el = $('toast');
    el.innerHTML = html;
    el.hidden = false;
    clearTimeout(toastTimer);
    toastTimer = setTimeout(function () { el.hidden = true; }, 4200);
  }

  function getToken() { return localStorage.getItem(LS_TOKEN) || ''; }
  function getName() { return localStorage.getItem(LS_NAME) || ''; }

  /* ---------- 数据获取 ---------- */
  function fetchLiveItems() {
    var params = new URLSearchParams();
    if (state.q) params.set('q', state.q);
    if (state.category !== 'all') params.set('category', state.category); // 后端按 slug 匹配
    if (state.tag) params.set('tags', state.tag);
    params.set('sort', state.sort);
    params.set('page', '1');
    params.set('size', '50');

    return fetch(API.search + '?' + params.toString())
      .then(function (res) {
        if (!res.ok) throw new Error('HTTP ' + res.status);
        return res.json();
      })
      .then(function (data) {
        if (!data || !Array.isArray(data.items)) throw new Error('bad payload');
        return data.items;
      });
  }

  function fetchVersions(publicId) {
    if (versionsCache[publicId]) return Promise.resolve(versionsCache[publicId]);
    return fetch(API.versions + '?public_id=' + encodeURIComponent(publicId))
      .then(function (res) {
        if (!res.ok) throw new Error('HTTP ' + res.status);
        return res.json();
      })
      .then(function (data) {
        var versions = (data && data.versions) || [];
        versionsCache[publicId] = versions;
        return versions;
      });
  }

  function seedItems() {
    var q = state.q.trim().toLowerCase();
    var list = SEED.skills.filter(function (s) {
      if (state.category !== 'all' && s.category_slug !== state.category) return false;
      if (state.tag && (s.tags || []).indexOf(state.tag) === -1) return false;
      if (q) {
        var hay = [s.display_name, s.slug, s.summary, (s.tags || []).join(' ')]
          .join(' ').toLowerCase();
        if (hay.indexOf(q) === -1) return false;
      }
      return true;
    });

    list.sort(function (a, b) {
      if (state.sort === 'downloads') return b.download_count - a.download_count;
      if (state.sort === 'rating') return b.rating_avg - a.rating_avg;
      return String(b.created_at).localeCompare(String(a.created_at));
    });
    return list;
  }

  /* ---------- 渲染 ---------- */
  function renderCats() {
    var row = $('category-row');
    var html = ['<button type="button" class="cat-link' +
      (state.category === 'all' ? ' is-active' : '') + '" data-cat="all">全部</button>'];
    SEED.categories.forEach(function (c) {
      var count = SEED.skills.filter(function (s) { return s.category_slug === c.slug; }).length;
      var active = state.category === c.slug ? ' is-active' : '';
      html.push('<button type="button" class="cat-link' + active + '" data-cat="' + c.slug + '">' +
        escapeHtml(c.name) + '<span class="cat-count">' + count + '</span></button>');
    });
    row.innerHTML = html.join('');
  }

  function renderActiveFilter() {
    var el = $('active-filter');
    if (!state.q && !state.tag) { el.hidden = true; el.innerHTML = ''; return; }
    var label = [];
    if (state.q) label.push('“' + state.q + '”');
    if (state.tag) label.push('#' + state.tag);
    el.innerHTML = '筛选：' + escapeHtml(label.join(' ')) +
      ' <button type="button" id="filter-clear" aria-label="清除筛选">×</button>';
    el.hidden = false;
  }

  function renderStats(items) {
    var el = $('hero-stats');
    if (state.source === 'live') {
      el.innerHTML = '<span>' + items.length + ' 个技能 · 后端在线</span>';
    } else {
      var versionCount = SEED.skills.reduce(function (n, s) {
        return n + (s.versions ? s.versions.length : 1);
      }, 0);
      el.innerHTML =
        '<span>' + SEED.skills.length + ' 个技能</span><span class="sep" aria-hidden="true">·</span>' +
        '<span>' + versionCount + ' 个版本</span><span class="sep" aria-hidden="true">·</span>' +
        '<span>' + SEED.categories.length + ' 个分类</span>';
    }
    $('source-badge').textContent = state.source === 'live' ? '后端在线' : '示例数据';
    $('source-badge').classList.toggle('is-live', state.source === 'live');
    $('footer-source').textContent = state.source === 'live' ? '后端 API' : '示例数据';
  }

  function cardHtml(s) {
    var versions = s.versions || versionsCache[s.public_id] || [];
    var latest = versions.slice().sort(semverDesc)[0];
    var verLabel = latest ? latest.version : '';
    var excerpt = s.excerpt
      ? '<div class="card-excerpt">' + escapeHtml(s.excerpt) + '</div>'
      : '';
    var tags = (s.tags || []).map(function (t) {
      return '<button type="button" class="card-tag" data-tag="' + escapeHtml(t) + '">' +
        escapeHtml(t) + '</button>';
    }).join('');
    var rating = s.rating_avg != null
      ? '<span class="rating">' + Number(s.rating_avg).toFixed(1) + '</span>'
      : '';
    var drawer = state.openDrawers[s.public_id]
      ? '<div class="version-drawer" data-drawer="' + escapeHtml(s.public_id) + '">' +
        '<p class="drawer-loading">正在取版本列表…</p></div>'
      : '';

    return (
      '<article class="skill-card" data-public-id="' + escapeHtml(s.public_id) + '">' +
        '<div class="card-top">' +
          '<span class="card-slug">' + escapeHtml(s.slug) + '</span>' +
          (verLabel ? '<span class="card-version">v' + escapeHtml(verLabel) + '</span>' : '') +
        '</div>' +
        '<h3 class="card-name">' + escapeHtml(s.display_name) + '</h3>' +
        '<p class="card-summary">' + escapeHtml(s.summary) + '</p>' +
        excerpt +
        '<div class="card-tags">' + tags + '</div>' +
        '<div class="card-meta">' +
          (s.category_name ? '<span>' + escapeHtml(s.category_name) + '</span>' : '') +
          rating +
          '<span>↓ ' + (s.download_count != null ? s.download_count : '—') + '</span>' +
          '<span>' + escapeHtml(formatDate(s.created_at)) + '</span>' +
        '</div>' +
        '<div class="card-actions">' +
          '<button type="button" class="btn btn-primary" data-dl="' + escapeHtml(s.public_id) + '">' +
            (verLabel ? '下载 v' + escapeHtml(verLabel) : '下载最新版') + '</button>' +
          '<button type="button" class="versions-toggle" data-ver="' + escapeHtml(s.public_id) + '">' +
            '全部版本' + (versions.length ? '（' + versions.length + '）' : '') + '</button>' +
        '</div>' +
        drawer +
      '</article>'
    );
  }

  function renderItems(items) {
    lastItems = items;
    var grid = $('skill-grid');
    $('empty-state').hidden = items.length > 0;
    grid.innerHTML = items.map(cardHtml).join('');
    $('grid-count').textContent = items.length
      ? '共 ' + items.length + ' 件 · 更新于 ' + new Date().toISOString().slice(0, 10)
      : '';
    renderStats(items);
    renderActiveFilter();

    // 重开已展开的抽屉并填充内容
    Object.keys(state.openDrawers).forEach(function (pid) {
      var s = items.filter(function (x) { return x.public_id === pid; })[0];
      if (s) fillDrawer(s);
    });

    // live 模式：为每张卡懒加载版本信息（补版本号与版本数）
    if (state.source === 'live') hydrateVersions(items);
  }

  function hydrateVersions(items) {
    items.forEach(function (s) {
      fetchVersions(s.public_id)
        .then(function (versions) {
          if (!versions.length) return;
          var card = document.querySelector(
            '.skill-card[data-public-id="' + CSS.escape(s.public_id) + '"]');
          if (!card) return;
          var top = card.querySelector('.card-top');
          if (top && !top.querySelector('.card-version')) {
            var latest = versions.slice().sort(semverDesc)[0];
            var pill = document.createElement('span');
            pill.className = 'card-version';
            pill.textContent = 'v' + latest.version;
            top.appendChild(pill);
          }
          var btn = card.querySelector('[data-dl]');
          if (btn) btn.textContent = '下载 v' + versions.slice().sort(semverDesc)[0].version;
          var toggle = card.querySelector('.versions-toggle');
          if (toggle && !toggle.textContent.match(/（/)) {
            toggle.textContent = '全部版本（' + versions.length + '）';
          }
        })
        .catch(function () { /* 版本信息拿不到时保持静默，点击时再提示 */ });
    });
  }

  function refresh() {
    if (state.liveLocked) {
      state.source = 'seed';
      renderItems(seedItems());
      return;
    }
    fetchLiveItems()
      .then(function (items) {
        state.source = 'live';
        renderItems(items);
      })
      .catch(function () {
        state.liveLocked = true;
        state.source = 'seed';
        renderItems(seedItems());
      });
  }

  /* ---------- 版本抽屉 ---------- */
  function drawerTable(versions) {
    var rows = versions.slice().sort(semverDesc).map(function (v, i) {
      return (
        '<tr>' +
          '<td class="' + (i === 0 ? 'ver-latest' : '') + '">v' + escapeHtml(v.version) + '</td>' +
          '<td>' + escapeHtml(formatSize(v.size_bytes)) + '</td>' +
          '<td>' + escapeHtml(v.founder_name || '—') + '</td>' +
          '<td>' + escapeHtml(formatDate(v.created_at)) + '</td>' +
          '<td><button type="button" class="dl-link" data-dlid="' + escapeHtml(v.id) +
            '" data-fname="' + escapeHtml(v.file_name || '') + '">下载 zip</button></td>' +
        '</tr>'
      );
    });
    return (
      '<table>' +
        '<thead><tr><th>版本</th><th>大小</th><th>上传者</th><th>发布日期</th><th></th></tr></thead>' +
        '<tbody>' + rows.join('') + '</tbody>' +
      '</table>'
    );
  }

  function fillDrawer(skill) {
    var drawer = document.querySelector('[data-drawer="' + CSS.escape(skill.public_id) + '"]');
    if (!drawer) return;
    var seedVersions = skill.versions || versionsCache[skill.public_id];
    if (state.source === 'seed' && seedVersions) {
      drawer.innerHTML = drawerTable(seedVersions);
      return;
    }
    fetchVersions(skill.public_id)
      .then(function (versions) {
        if (!versions.length) throw new Error('empty');
        drawer.innerHTML = drawerTable(versions);
      })
      .catch(function () {
        drawer.innerHTML = '<p class="drawer-loading">版本列表拿不到 —— 后端未启动或接口不可用。</p>';
      });
  }

  function toggleDrawer(pid) {
    var skill = lastItems.filter(function (x) { return x.public_id === pid; })[0] ||
      SEED.skills.filter(function (x) { return x.public_id === pid; })[0];
    if (state.openDrawers[pid]) delete state.openDrawers[pid];
    else state.openDrawers[pid] = true;
    renderItems(state.source === 'live' ? lastItems : seedItems());
    if (state.openDrawers[pid] && skill) fillDrawer(skill);
  }

  /* ---------- 下载（规范端点：request_file → FileResponse） ---------- */
  function downloadZip(id, fileName) {
    var url = API.download + '?skill_versions_id=' + encodeURIComponent(id);
    fetch(url)
      .then(function (res) {
        if (!res.ok) throw new Error('HTTP ' + res.status);
        return res.blob();
      })
      .then(function (blob) {
        var a = document.createElement('a');
        a.href = URL.createObjectURL(blob);
        a.download = fileName || 'skill.zip';
        document.body.appendChild(a);
        a.click();
        a.remove();
        setTimeout(function () { URL.revokeObjectURL(a.href); }, 4000);
        toast('已开始下载 <code>' + escapeHtml(fileName || 'skill.zip') + '</code>');
      })
      .catch(function () {
        toast('后端未响应。启动 FastAPI 后即可下载：<code>' + url + '</code>');
      });
  }

  function downloadLatest(pid) {
    var skill = lastItems.filter(function (x) { return x.public_id === pid; })[0];
    var latest = skill && (skill.versions || versionsCache[pid] || []).slice().sort(semverDesc)[0];
    if (latest && latest.id != null) {
      downloadZip(latest.id, latest.file_name);
      return;
    }
    fetchVersions(pid)
      .then(function (versions) {
        var v = versions.slice().sort(semverDesc)[0];
        if (v && v.id != null) downloadZip(v.id, v.file_name);
        else toast('该技能暂时没有可下载的版本。');
      })
      .catch(function () {
        toast('后端未响应，暂时无法下载。');
      });
  }

  /* ---------- hero 指令卡动画 ---------- */
  function playMdCard() {
    var body = $('md-body');
    var status = $('md-status');
    var footState = $('md-foot-state');
    var reduce = window.matchMedia('(prefers-reduced-motion: reduce)').matches;

    body.innerHTML = HERO_LINES.map(function (l) {
      return '<span class="md-line' + (l.meta ? ' is-meta' : '') + '">' +
        escapeHtml(l.text) + '</span>';
    }).join('');

    var lines = body.querySelectorAll('.md-line');

    function finish() {
      for (var i = 0; i < lines.length; i++) {
        lines[i].classList.remove('is-reading');
        lines[i].classList.add(i === lines.length - 1 ? 'is-read' : 'is-read');
      }
      status.textContent = '已习得 ✓';
      footState.classList.add('is-done');
      footState.textContent = '✓ 已习得 — agent 现在会发邮件了';
    }

    if (reduce) { finish(); return; }

    var i = 0;
    var timer = setInterval(function () {
      if (i > 0) lines[i - 1].classList.replace('is-reading', 'is-read');
      if (i >= lines.length) {
        clearInterval(timer);
        finish();
        return;
      }
      lines[i].classList.add('is-reading');
      footState.textContent = '▍正在读第 ' + (i + 1) + ' 行…';
      i++;
    }, 420);
  }

  /* ---------- 登录区 ---------- */
  function renderAuthArea() {
    var area = $('auth-area');
    var name = getName();
    if (name) {
      area.innerHTML = '<span class="auth-name" title="已登录">' + escapeHtml(name) + '</span>' +
        '<button type="button" class="auth-logout" id="logout-btn">退出</button>';
    } else {
      area.innerHTML = '<button type="button" class="btn btn-ghost" id="login-btn">登录</button>';
    }
  }

  function openModal(id) { $(id).hidden = false; }
  function closeModal(id) { $(id).hidden = true; }

  function switchAuthTab(which) {
    var isLogin = which === 'login';
    $('tab-login').classList.toggle('is-active', isLogin);
    $('tab-register').classList.toggle('is-active', !isLogin);
    $('tab-login').setAttribute('aria-selected', String(isLogin));
    $('tab-register').setAttribute('aria-selected', String(!isLogin));
    $('login-form').hidden = !isLogin;
    $('register-form').hidden = isLogin;
    $('login-error').hidden = true;
    $('register-error').hidden = true;
  }

  function showError(id, message) {
    var el = $(id);
    el.textContent = message;
    el.hidden = false;
  }

  function handleLogin(e) {
    e.preventDefault();
    var errEl = $('login-error');
    errEl.hidden = true;
    fetch(API.login, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        email: $('login-email').value.trim(),
        password: $('login-password').value
      })
    })
      .then(function (res) { return res.json().then(function (d) { return { ok: res.ok, data: d }; }); })
      .then(function (r) {
        if (!r.ok) throw new Error(apiError(r.data, '登录失败，请检查邮箱和密码。'));
        localStorage.setItem(LS_TOKEN, r.data.token || '');
        localStorage.setItem(LS_NAME, r.data.name || '');
        renderAuthArea();
        closeModal('auth-modal');
        $('login-password').value = '';
        toast('登录成功。上传技能时 token 会自动填入。');
      })
      .catch(function (err) {
        showError('login-error', err.message || '登录失败，后端未响应。');
      });
  }

  function handleRegister(e) {
    e.preventDefault();
    var errEl = $('register-error');
    errEl.hidden = true;
    fetch(API.register, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        email: $('reg-email').value.trim(),
        name: $('reg-name').value.trim(),
        password: $('reg-password').value
      })
    })
      .then(function (res) { return res.json().then(function (d) { return { ok: res.ok, data: d }; }); })
      .then(function (r) {
        if (!r.ok) throw new Error(apiError(r.data, '注册失败。'));
        toast('注册成功，请登录。');
        switchAuthTab('login');
        $('login-email').value = $('reg-email').value;
        $('login-password').focus();
      })
      .catch(function (err) {
        showError('register-error', err.message || '注册失败，后端未响应。');
      });
  }

  /* ---------- 上传 ---------- */
  function openUpload() {
    $('up-token').value = getToken();
    $('upload-error').hidden = true;
    openModal('upload-modal');
  }

  function handleUpload(e) {
    e.preventDefault();
    var errEl = $('upload-error');
    errEl.hidden = true;

    var form = $('upload-form');
    var fd = new FormData();
    fd.append('name', $('up-name').value.trim());
    fd.append('slug', $('up-slug').value.trim());
    fd.append('version', $('up-version').value.trim());
    fd.append('category', $('up-category').value);
    fd.append('tags', $('up-tags').value.trim());
    fd.append('summary', $('up-summary').value.trim());
    fd.append('readme_html', $('up-readme').value);
    fd.append('token', $('up-token').value.trim());
    fd.append('upload_zip', $('up-zip').files[0]);

    var submitBtn = $('upload-submit');
    submitBtn.disabled = true;
    submitBtn.textContent = '上传中…';

    fetch(API.upload, { method: 'POST', body: fd })
      .then(function (res) { return res.json().then(function (d) { return { ok: res.ok, data: d }; }); })
      .then(function (r) {
        if (!r.ok) throw new Error(apiError(r.data, '上传失败。'));
        closeModal('upload-modal');
        form.reset();
        $('up-token').value = getToken();
        toast('上传成功：<code>' + escapeHtml(r.data.file_name || '') + '</code>');
        versionsCache = {};
        state.liveLocked = false;   // 上传成功说明后端在线，重新拉取
        refresh();
      })
      .catch(function (err) {
        showError('upload-error', err.message || '上传失败，后端未响应。');
      })
      .finally(function () {
        submitBtn.disabled = false;
        submitBtn.textContent = '上传';
      });
  }

  /* ---------- 事件 ---------- */
  function bindEvents() {
    $('search-form').addEventListener('submit', function (e) {
      e.preventDefault();
      state.q = $('search-input').value;
      refresh();
    });

    $('category-row').addEventListener('click', function (e) {
      var btn = e.target.closest('.cat-link');
      if (!btn) return;
      state.category = btn.dataset.cat;
      renderCats();
      refresh();
    });

    $('sort-select').addEventListener('change', function (e) {
      state.sort = e.target.value;
      refresh();
    });

    document.addEventListener('click', function (e) {
      // 标签筛选（工具条芯片 + 卡片标签）
      var chip = e.target.closest('.card-tag');
      if (chip) {
        state.tag = state.tag === chip.dataset.tag ? '' : chip.dataset.tag;
        refresh();
        return;
      }

      // 清除筛选
      if (e.target.closest('#filter-clear') || e.target.closest('#clear-filters')) {
        state.q = '';
        state.tag = '';
        state.category = 'all';
        state.openDrawers = {};
        $('search-input').value = '';
        renderCats();
        refresh();
        return;
      }

      // 版本抽屉
      var ver = e.target.closest('.versions-toggle');
      if (ver) { toggleDrawer(ver.dataset.ver); return; }

      // 下载最新版
      var dl = e.target.closest('[data-dl]');
      if (dl) { downloadLatest(dl.dataset.dl); return; }

      // 抽屉里指定版本下载
      var dlid = e.target.closest('[data-dlid]');
      if (dlid) { downloadZip(dlid.dataset.dlid, dlid.dataset.fname); return; }

      // 登录 / 退出
      if (e.target.closest('#login-btn')) {
        switchAuthTab('login');
        openModal('auth-modal');
        return;
      }
      if (e.target.closest('#logout-btn')) {
        localStorage.removeItem(LS_TOKEN);
        localStorage.removeItem(LS_NAME);
        renderAuthArea();
        toast('已退出登录。');
        return;
      }
      if (e.target.closest('#upload-btn')) { openUpload(); return; }

      // 弹窗关闭
      if (e.target.closest('[data-close-auth]')) { closeModal('auth-modal'); return; }
      if (e.target.closest('[data-close-upload]')) { closeModal('upload-modal'); return; }
    });

    // 弹窗内 tab 切换
    $('tab-login').addEventListener('click', function () { switchAuthTab('login'); });
    $('tab-register').addEventListener('click', function () { switchAuthTab('register'); });

    $('login-form').addEventListener('submit', handleLogin);
    $('register-form').addEventListener('submit', handleRegister);
    $('upload-form').addEventListener('submit', handleUpload);

    // ESC 关闭弹窗
    document.addEventListener('keydown', function (e) {
      if (e.key === 'Escape') {
        closeModal('auth-modal');
        closeModal('upload-modal');
      }
    });
  }

  /* ---------- 启动 ---------- */
  function init() {
    renderCats();
    renderAuthArea();
    playMdCard();
    bindEvents();
    refresh();
  }

  init();
})();
