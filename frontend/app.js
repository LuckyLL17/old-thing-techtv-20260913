const $ = s => document.querySelector(s);
const $$ = s => [...document.querySelectorAll(s)];
const API = '/api/v1';
const state = { token: localStorage.getItem('token') || '', user: null, page: 1, size: 12 };

function headers(auth = true) {
  const h = { 'Content-Type': 'application/json' };
  if (auth && state.token) h['Authorization'] = 'Bearer ' + state.token;
  return h;
}
async function get(url, auth = true) {
  const r = await fetch(API + url, { headers: headers(auth) });
  return r.json();
}
async function post(url, body, auth = true, method = 'POST') {
  const r = await fetch(API + url, { method, headers: headers(auth), body: body ? JSON.stringify(body) : undefined });
  return r.json();
}
function toast(msg, type = '') {
  const t = $('#toast');
  t.textContent = msg;
  t.className = 'toast show ' + type;
  setTimeout(() => t.classList.remove('show'), 2200);
}
function el(tag, cls, html) {
  const e = document.createElement(tag);
  if (cls) e.className = cls;
  if (html !== undefined) e.innerHTML = html;
  return e;
}
function avatarFor(u) {
  if (u && u.avatar) return `<img src="${u.avatar}" class="avatar" onerror="this.outerHTML='<div class=avatar>${(u.nickname||u.username||'U').charAt(0).toUpperCase()}</div>'">`;
  const name = (u && (u.nickname || u.username)) || 'U';
  return `<div class="avatar">${name.charAt(0).toUpperCase()}</div>`;
}
function timeAgo(t) {
  if (!t) return '';
  const d = new Date(t).getTime();
  const s = (Date.now() - d) / 1000;
  if (s < 60) return '刚刚';
  if (s < 3600) return Math.floor(s / 60) + '分钟前';
  if (s < 86400) return Math.floor(s / 3600) + '小时前';
  if (s < 86400 * 30) return Math.floor(s / 86400) + '天前';
  return t.slice(0, 10);
}
function difficultyBadge(d) {
  const map = { easy: ['简单', 'easy'], medium: ['中等', 'medium'], hard: ['困难', 'hard'] };
  const v = map[d] || map.medium;
  return `<span class="difficulty difficulty-${v[1]}">${v[0]}</span>`;
}
function levelLabel(l) {
  return { novice: '新手', apprentice: '学徒', craftsman: '匠人', master: '大师' }[l] || '新手';
}
function stars(n) {
  let s = '';
  for (let i = 0; i < 5; i++) s += i < (n|0) ? '★' : '☆';
  return `<span class="stars">${s}</span>`;
}
async function ensureUser() {
  if (state.token && !state.user) {
    const r = await get('/auth/me');
    if (r.success) state.user = r.data;
  }
  renderUserArea();
  return state.user;
}
function renderUserArea() {
  const area = $('#user-area');
  if (!area) return;
  if (!state.user) {
    area.innerHTML = `<button class="btn btn-ghost" onclick="showLogin()">登录</button><button class="btn btn-ghost" onclick="showRegister()">注册</button>`;
    return;
  }
  const u = state.user;
  area.innerHTML = `<div class="user-menu">${avatarFor(u)}<span>${u.nickname||u.username}</span><div class="dropdown" id="ud" style="display:none"><a href="#/me">个人中心</a><a href="#/editor">发布教程</a><a href="#/me/projects">我的作品</a><a href="javascript:logout()">退出登录</a></div></div>`;
  area.querySelector('.user-menu').onclick = e => { e.stopPropagation(); const d = $('#ud'); d.style.display = d.style.display === 'none' ? 'block' : 'none'; };
  document.body.onclick = () => { const d = $('#ud'); if (d) d.style.display = 'none'; };
}
function logout() {
  state.token = ''; state.user = null;
  localStorage.removeItem('token');
  toast('已退出', 'success');
  renderUserArea(); route();
}
function showLogin() {
  const m = el('div', 'modal-bg');
  m.innerHTML = `<div class="modal"><h3>登录</h3><div><label class="label">账号（邮箱/用户名）</label><input class="input" id="la" placeholder="请输入邮箱或用户名"></div><div><label class="label">密码</label><input class="input" type="password" id="lp" placeholder="6-32位字母数字"></div><div class="modal-actions"><button class="btn btn-gray" onclick="this.closest('.modal-bg').remove()">取消</button><button class="btn btn-solid" id="lsb">登录</button></div></div>`;
  m.onclick = e => { if (e.target === m) m.remove(); };
  document.body.appendChild(m);
  $('#lsb').onclick = async () => {
    const r = await post('/auth/login', { account: $('#la').value.trim(), password: $('#lp').value });
    if (!r.success) return toast(r.message, 'error');
    state.token = r.data.token; state.user = r.data.user;
    localStorage.setItem('token', state.token);
    toast('登录成功', 'success');
    m.remove(); renderUserArea(); route();
  };
}
function showRegister() {
  const m = el('div', 'modal-bg');
  m.innerHTML = `<div class="modal"><h3>注册</h3><div><label class="label">用户名</label><input class="input" id="ru" placeholder="3-30字符"></div><div><label class="label">邮箱</label><input class="input" id="re" placeholder="you@example.com"></div><div><label class="label">密码</label><input class="input" type="password" id="rp" placeholder="6-32位字母数字"></div><div class="modal-actions"><button class="btn btn-gray" onclick="this.closest('.modal-bg').remove()">取消</button><button class="btn btn-solid" id="rsb">注册</button></div></div>`;
  m.onclick = e => { if (e.target === m) m.remove(); };
  document.body.appendChild(m);
  $('#rsb').onclick = async () => {
    const r = await post('/auth/register', { username: $('#ru').value.trim(), email: $('#re').value.trim(), password: $('#rp').value });
    if (!r.success) return toast(r.message, 'error');
    state.token = r.data.token; state.user = r.data.user;
    localStorage.setItem('token', state.token);
    toast('注册成功', 'success');
    m.remove(); renderUserArea(); route();
  };
}
function requireLogin() {
  if (!state.user) { toast('请先登录', 'error'); showLogin(); return false; }
  return true;
}

function renderHome(data) {
  let h = `<div class="hero"><h1>让旧物重新发光 ✨</h1><p>发现创意教程，分享改造灵感，把即将丢弃的物品变成独特的宝藏</p><div class="flex"><button class="btn btn-solid btn-lg" onclick="location.hash='#/tutorials'">浏览教程</button><button class="btn btn-outline btn-lg" onclick="location.hash='#/projects'">看看作品</button></div></div>`;
  h += `<div class="section-title">🔥 热门教程<small onclick="location.hash='#/tutorials?sort=popular'">查看全部 →</small></div><div class="grid grid-4">${renderTutorialCards(data.hot_tutorials||[])}</div>`;
  h += `<div class="section-title">🆕 最新发布<small onclick="location.hash='#/tutorials?sort=new'">查看全部 →</small></div><div class="grid grid-4">${renderTutorialCards(data.new_tutorials||[])}</div>`;
  h += `<div class="section-title">🎯 最多尝试<small onclick="location.hash='#/tutorials?sort=attempted'">查看全部 →</small></div><div class="grid grid-3">${renderTutorialCards(data.most_attempted||[])}</div>`;
  h += `<div class="section-title">🎲 随机灵感<small onclick="location.hash='#/random'">换一批 →</small></div><div class="grid grid-3">${renderTutorialCards(data.random_pick||[])}</div>`;
  if (data.categories && data.categories.length) {
    h += `<div class="section-title">📂 分类浏览</div><div class="grid grid-3 grid-4"><div class="cat" onclick="location.hash='#/tutorials'">${data.categories.map(c => `<div class="cat"><span class="cat-icon">${c.icon}</span><div><b>${c.name}</b><div style="font-size:12px;color:#888">${c.tutorial_count||0} 个教程</div></div></div>`).join('')}</div></div>`;
  }
  if (data.popular_tags && data.popular_tags.length) {
    h += `<div class="section-title">🏷️ 热门标签</div><div class="chip-row">${data.popular_tags.map(t => `<span class="tag tag-hot" onclick="location.hash='#/tutorials?keyword=${encodeURIComponent(t.name)}'">#${t.name} (${t.tutorial_count})</span>`).join('')}</div>`;
  }
  return h;
}
function renderTutorialCards(list) {
  if (!list.length) return `<div class="empty grid-4" style="grid-column:1/-1"><div class="empty-icon">📭</div>暂无内容</div>`;
  return list.map(t => {
    const user = t.user || {};
    const cat = t.category || {};
    return `<div class="card" onclick="location.hash='#/tutorials/${t.id}'"><div class="before-after"><img src="${t.cover_before||'https://picsum.photos/seed/b'+t.id+'/400'}" onerror="this.src='https://picsum.photos/seed/b'+${t.id}+'/400'" alt="改造前"><img src="${t.cover_after||'https://picsum.photos/seed/a'+t.id+'/400'}" onerror="this.src='https://picsum.photos/seed/a'+${t.id}+'/400'" alt="改造后"></div><div class="meta"><span>👁 ${t.view_count}</span><span>❤️ ${t.favorite_count}</span><span>🛠 ${t.attempt_count}</span><span>${(t.tags||[]).slice(0,2).map(x=>`<span class="tag">#${x.name}</span>`).join('')}</span></div><div class="card-body"><div class="card-title">${t.title}</div><div class="flex" style="justify-content:space-between;margin-top:8px"><div class="flex">${avatarFor(user)}<span style="font-size:13px;color:#666">${user.nickname||user.username||'匿名'}</span></div><div style="font-size:12px">${difficultyBadge(t.difficulty)} · ${t.estimated_hours}h</div></div></div></div>`;
  }).join('');
}
function renderProjectCards(list) {
  if (!list.length) return `<div class="empty grid-4" style="grid-column:1/-1"><div class="empty-icon">🎨</div>暂无作品，来发布第一个吧！</div>`;
  return list.map(p => {
    const user = p.user || {};
    const tut = p.tutorial || {};
    const img = (p.images || '').split(/[|,;]/)[0] || `https://picsum.photos/seed/p${p.id}/400`;
    return `<div class="card" onclick="location.hash='#/projects/${p.id}'"><img src="${img}" onerror="this.src='https://picsum.photos/seed/p'+${p.id}+'/400'" style="height:220px"><div class="meta"><span>👍 ${p.like_count}</span><span>💬 ${p.comment_count}</span><span>${stars(p.rating)}</span></div><div class="card-body"><div class="card-title">${p.title||'未命名作品'}</div><div style="font-size:12px;color:#888;margin-top:4px">来自教程：<b>${tut.title||'-'}</b></div><div class="flex" style="margin-top:10px"><div class="flex">${avatarFor(user)}<span style="font-size:13px;color:#666">${user.nickname||user.username||'匿名'}</span></div></div></div></div>`;
  }).join('');
}
async function viewTutorials() {
  const qs = parseHashQuery();
  const params = new URLSearchParams();
  ['page','size','category','difficulty','sort','keyword','user_id'].forEach(k => { if (qs[k]) params.set(k, qs[k]); });
  if (state.page > 1) params.set('page', state.page);
  const r = await get('/tutorials?' + params.toString(), false);
  const cats = (await get('/categories', false)).data || [];
  if (!r.success) { $('#app').innerHTML = `<div class="empty">${r.message}</div>`; return; }
  let h = `<div class="bread"><a href="#/">首页</a> / 全部教程</div>`;
  h += `<h1 style="font-size:26px;margin-bottom:14px">📚 教程库</h1>`;
  h += `<div class="flex-between" style="margin-bottom:18px"><div class="chip-row"><span class="pill ${!qs.category?'pill-active':''}" onclick="clearQ('category')">全部分类</span>${cats.map(c=>`<span class="pill ${qs.category==c.id?'pill-active':''}" onclick="setQ('category',${c.id})">${c.icon} ${c.name}</span>`).join('')}</div><div class="chip-row"><span class="pill ${!qs.difficulty?'pill-active':''}" onclick="clearQ('difficulty')">全部难度</span><span class="pill ${qs.difficulty=='easy'?'pill-active':''}" onclick="setQ('difficulty','easy')">简单</span><span class="pill ${qs.difficulty=='medium'?'pill-active':''}" onclick="setQ('difficulty','medium')">中等</span><span class="pill ${qs.difficulty=='hard'?'pill-active':''}" onclick="setQ('difficulty','hard')">困难</span></div></div>`;
  h += `<div class="flex-between" style="margin-bottom:14px"><div class="chip-row"><span class="pill ${!qs.sort||qs.sort=='new'?'pill-active':''}" onclick="setQ('sort','new')">最新</span><span class="pill ${qs.sort=='popular'?'pill-active':''}" onclick="setQ('sort','popular')">最热</span><span class="pill ${qs.sort=='attempted'?'pill-active':''}" onclick="setQ('sort','attempted')">最多尝试</span><span class="pill ${qs.sort=='views'?'pill-active':''}" onclick="setQ('sort','views')">最多浏览</span></div><div><input class="input" style="width:260px;display:inline-block" placeholder="搜索关键词..." value="${qs.keyword||''}" onkeydown="if(event.key==='Enter')setQ('keyword',this.value)"></div></div>`;
  h += `<div class="grid grid-4">${renderTutorialCards(r.data.list||[])}</div>`;
  h += pagination(r.data.total, r.data.page, r.data.size);
  $('#app').innerHTML = h;
}
function setQ(k, v) {
  const q = parseHashQuery();
  q[k] = v; state.page = 1;
  const base = location.hash.split('?')[0] || '#/tutorials';
  location.hash = base + '?' + new URLSearchParams(q).toString();
}
function clearQ(k) { setQ(k, ''); }
function parseHashQuery() {
  const h = location.hash.split('?')[1] || '';
  const p = new URLSearchParams(h);
  const o = {}; p.forEach((v, k) => o[k] = v);
  return o;
}
function pagination(total, page, size) {
  const pages = Math.ceil(total / size) || 1;
  if (pages <= 1) return '';
  let h = `<div class="pagination">`;
  if (page > 1) h += `<button class="page-btn" onclick="state.page=${page-1};route()">‹</button>`;
  for (let i = Math.max(1, page-2); i <= Math.min(pages, page+2); i++) {
    h += `<button class="page-btn ${i===page?'page-btn-active':''}" onclick="state.page=${i};route()">${i}</button>`;
  }
  if (page < pages) h += `<button class="page-btn" onclick="state.page=${page+1};route()">›</button>`;
  h += `</div>`;
  return h;
}
async function viewTutorial(id) {
  const r = await get('/tutorials/' + id, true);
  if (!r.success) { $('#app').innerHTML = `<div class="empty">${r.message}</div>`; return; }
  const t = r.data.tutorial;
  const fav = r.data.favorited;
  const user = t.user || {};
  const cat = t.category || {};
  let h = `<div class="bread"><a href="#/">首页</a> / <a href="#/tutorials">教程</a> / <span>${t.title}</span></div>`;
  h += `<div style="display:flex;gap:10px;margin-bottom:10px;align-items:center">${difficultyBadge(t.difficulty)}<span style="color:#888">⏱ ${t.estimated_hours} 小时</span><span>${(t.tags||[]).map(x=>`<span class="tag">#${x.name}</span>`).join('')}</span></div>`;
  h += `<h1 style="font-size:30px;margin-bottom:12px">${t.title}</h1>`;
  h += `<div class="flex-between" style="margin-bottom:18px"><div class="flex">${avatarFor(user)}<div><div style="font-weight:600">${user.nickname||user.username||'匿名'} <span style="color:#7c5cff;font-size:12px">[${levelLabel(user.level)}]</span></div><div style="font-size:12px;color:#888">${timeAgo(t.created_at)} · 👁${t.view_count} ❤️${t.favorite_count} 🛠${t.attempt_count}</div></div></div><div class="flex"><button class="btn btn-outline" onclick="toggleFav('tutorial',${t.id},this)">${fav?'❤️ 已收藏':'🤍 收藏'}</button><button class="btn btn-solid" onclick="attemptTut(${t.id})">🛠 我要尝试</button></div></div>`;
  h += `<div class="before-after" style="margin-bottom:24px"><div><div style="padding:6px 12px;background:#ffe3e3;color:#c92a2a;border-radius:8px 8px 0 0;font-size:12px;font-weight:600;display:inline-block">改造前</div><img src="${t.cover_before}" style="border-radius:0 14px 14px 14px;width:100%;height:300px;object-fit:cover"></div><div><div style="padding:6px 12px;background:#d3f9d8;color:#2b8a3e;border-radius:8px 8px 0 0;font-size:12px;font-weight:600;display:inline-block">改造后</div><img src="${t.cover_after}" style="border-radius:0 14px 14px 14px;width:100%;height:300px;object-fit:cover"></div></div>`;
  h += `<div class="card" style="padding:24px;margin-bottom:24px"><h2 style="font-size:18px;margin-bottom:10px">📝 简介</h2><p style="color:#555">${t.summary||'暂无简介'}</p></div>`;
  if ((t.materials||[]).length) {
    h += `<div class="card" style="padding:24px;margin-bottom:24px"><h2 style="font-size:18px;margin-bottom:14px">🧰 所需物品</h2>`;
    const mats = t.materials.filter(m=>!m.is_tool); const tools = t.materials.filter(m=>m.is_tool);
    if (mats.length) { h += `<h3 style="font-size:15px;margin-bottom:10px;color:#7c5cff">📦 材料</h3><ul style="list-style:none;display:grid;grid-template-columns:1fr 1fr;gap:8px;margin-bottom:14px">${mats.map(m=>`<li style="padding:8px 12px;background:#f7f8fa;border-radius:8px">· ${m.name}${m.quantity?` <span style="color:#888">${m.quantity}${m.unit||''}</span>`:''}</li>`).join('')}</ul>`; }
    if (tools.length) { h += `<h3 style="font-size:15px;margin-bottom:10px;color:#3bc9db">🔨 工具</h3><ul style="list-style:none;display:grid;grid-template-columns:1fr 1fr;gap:8px">${tools.map(m=>`<li style="padding:8px 12px;background:#f7f8fa;border-radius:8px">· ${m.name}${m.quantity?` <span style="color:#888">${m.quantity}${m.unit||''}</span>`:''}</li>`).join('')}</ul>`; }
    h += `</div>`;
  }
  if ((t.steps||[]).length) {
    h += `<div class="section-title">📖 步骤说明</div>`;
    t.steps.forEach((s, i) => {
      h += `<div class="step"><div class="step-index">${i+1}</div><h3 style="font-size:16px;margin-bottom:10px">${s.title||'步骤 '+(i+1)}</h3>${s.image?`<img src="${s.image}" style="width:100%;border-radius:10px;margin:10px 0;max-height:400px;object-fit:cover">`:''}<p style="color:#444;white-space:pre-wrap">${s.content}</p>${s.reminder?`<div style="margin-top:10px;padding:10px 14px;background:#fff9db;border-left:4px solid #fcc419;border-radius:4px;color:#7a5f00">⚠️ ${s.reminder}</div>`:''}${s.estimated_minutes?`<div style="margin-top:8px;color:#888;font-size:12px">⏱ 预计耗时：${s.estimated_minutes} 分钟</div>`:''}</div>`;
    });
  }
  h += `<div class="section-title">🎨 大家的作品<small onclick="location.hash='#/projects?tutorial_id=${t.id}'">查看全部 →</small></div>`;
  const pr = await get('/projects?tutorial_id=' + t.id + '&size=6', false);
  h += `<div class="grid grid-3">${renderProjectCards((pr.data&&pr.data.list)||[])}</div>`;
  h += `<div class="section-title">💬 评论</div>`;
  h += `<div class="card" style="padding:18px;margin-bottom:16px"><div class="flex"><input class="input" id="cinput" placeholder="写下你的感受或提问..."><button class="btn btn-solid" onclick="addComment('tutorial',${t.id})">发布</button></div></div>`;
  const cr = await get('/tutorials/' + id + '/comments', false);
  if (cr.data && cr.data.list && cr.data.list.length) {
    cr.data.list.forEach(c => {
      const cu = c.user || {};
      h += `<div class="card" style="padding:16px;margin-bottom:12px"><div class="flex" style="margin-bottom:8px">${avatarFor(cu)}<div><div style="font-weight:600">${cu.nickname||cu.username||'匿名'}</div><div style="font-size:12px;color:#888">${timeAgo(c.created_at)} · 👍 ${c.like_count}</div></div></div><p>${c.content}</p></div>`;
    });
  } else {
    h += `<div class="empty" style="padding:30px">还没有评论，抢个沙发吧~</div>`;
  }
  $('#app').innerHTML = h;
}
async function toggleFav(type, id, btn) {
  if (!requireLogin()) return;
  const r = await post('/me/favorites', { target_type: type, target_id: id });
  if (!r.success) return toast(r.message, 'error');
  toast(r.data.favorited ? '已收藏' : '已取消', 'success');
  if (btn) btn.innerHTML = r.data.favorited ? '❤️ 已收藏' : '🤍 收藏';
}
async function attemptTut(id) {
  if (!requireLogin()) return;
  const r = await post('/tutorials/' + id + '/attempt', {});
  if (!r.success) return toast(r.message, 'error');
  toast('已加入"我的尝试"！加油完成它 ✊', 'success');
}
async function addComment(type, id) {
  if (!requireLogin()) return;
  const v = $('#cinput').value.trim();
  if (!v) return toast('请输入内容', 'error');
  const r = await post(`/${type}s/${id}/comments`, { content: v });
  if (!r.success) return toast(r.message, 'error');
  toast('发布成功', 'success'); route();
}
async function viewProjects() {
  const qs = parseHashQuery();
  const params = new URLSearchParams();
  ['page','size','tutorial_id','user_id','sort'].forEach(k => { if (qs[k]) params.set(k, qs[k]); });
  if (state.page > 1) params.set('page', state.page);
  const r = await get('/projects?' + params.toString(), false);
  if (!r.success) { $('#app').innerHTML = `<div class="empty">${r.message}</div>`; return; }
  let h = `<div class="bread"><a href="#/">首页</a> / 作品展示</div>`;
  h += `<h1 style="font-size:26px;margin-bottom:14px">🎨 创意作品</h1>`;
  h += `<div class="flex-between" style="margin-bottom:18px"><div class="chip-row"><span class="pill ${!qs.sort?'pill-active':''}" onclick="clearQ('sort')">最新</span><span class="pill ${qs.sort=='rating'?'pill-active':''}" onclick="setQ('sort','rating')">高评分</span><span class="pill ${qs.sort=='likes'?'pill-active':''}" onclick="setQ('sort','likes')">最受欢迎</span></div><button class="btn btn-solid" onclick="location.hash='#/projects/new'">+ 发布作品</button></div>`;
  h += `<div class="grid grid-3">${renderProjectCards(r.data.list||[])}</div>`;
  h += pagination(r.data.total, r.data.page, r.data.size);
  $('#app').innerHTML = h;
}
async function viewProject(id) {
  const r = await get('/projects/' + id, false);
  if (!r.success) { $('#app').innerHTML = `<div class="empty">${r.message}</div>`; return; }
  const p = r.data;
  const user = p.user || {};
  const tut = p.tutorial || {};
  const imgs = (p.images||'').split(/[|,;]/).filter(Boolean);
  if (!imgs.length) imgs.push(`https://picsum.photos/seed/p${p.id}/800`);
  let h = `<div class="bread"><a href="#/">首页</a> / <a href="#/projects">作品</a> / <span>${p.title||'作品详情'}</span></div>`;
  h += `<h1 style="font-size:28px;margin-bottom:12px">${p.title||'未命名作品'}</h1>`;
  h += `<div class="flex-between" style="margin-bottom:20px"><div class="flex">${avatarFor(user)}<div><div style="font-weight:600">${user.nickname||user.username||'匿名'}</div><div style="font-size:12px;color:#888">${timeAgo(p.created_at)} · 👍 ${p.like_count} 💬 ${p.comment_count} ${stars(p.rating)}</div></div></div><div class="flex"><button class="btn btn-outline" onclick="likeProject(${p.id},this)">👍 点赞</button><a class="btn btn-solid" href="#/tutorials/${tut.id}">📖 查看原教程</a></div></div>`;
  if (imgs.length === 1) {
    h += `<img src="${imgs[0]}" style="width:100%;border-radius:14px;margin-bottom:20px;max-height:500px;object-fit:cover">`;
  } else {
    h += `<div class="grid grid-3" style="margin-bottom:20px">${imgs.map(i=>`<img src="${i}" style="border-radius:14px;aspect-ratio:1;object-fit:cover;width:100%">`).join('')}</div>`;
  }
  if (p.description) h += `<div class="card" style="padding:24px;margin-bottom:20px"><h2 style="font-size:18px;margin-bottom:10px">📝 作品描述</h2><p style="color:#444;white-space:pre-wrap">${p.description}</p></div>`;
  if (p.custom_notes) h += `<div class="card" style="padding:24px;margin-bottom:20px;background:#f8f5ff"><h2 style="font-size:18px;margin-bottom:10px">💡 我的改动与心得</h2><p style="color:#444;white-space:pre-wrap">${p.custom_notes}</p></div>`;
  h += `<div class="section-title">💬 评论</div>`;
  h += `<div class="card" style="padding:18px;margin-bottom:16px"><div class="flex"><input class="input" id="cinput" placeholder="为创作者点赞鼓励..."><button class="btn btn-solid" onclick="addComment('project',${p.id})">发布</button></div></div>`;
  const cr = await get('/projects/' + id + '/comments', false);
  if (cr.data && cr.data.list && cr.data.list.length) {
    cr.data.list.forEach(c => {
      const cu = c.user || {};
      h += `<div class="card" style="padding:16px;margin-bottom:12px"><div class="flex" style="margin-bottom:8px">${avatarFor(cu)}<div><div style="font-weight:600">${cu.nickname||cu.username||'匿名'}</div><div style="font-size:12px;color:#888">${timeAgo(c.created_at)} · 👍 ${c.like_count}</div></div></div><p>${c.content}</p></div>`;
    });
  }
  $('#app').innerHTML = h;
}
async function likeProject(id, btn) {
  const r = await post('/projects/' + id + '/like', {}, false);
  if (!r.success) return toast(r.message, 'error');
  toast('+1 👍', 'success');
}
async function viewStats() {
  const r = await get('/stats', false);
  if (!r.success) { $('#app').innerHTML = `<div class="empty">${r.message}</div>`; return; }
  const d = r.data;
  let h = `<div class="bread"><a href="#/">首页</a> / 数据看板</div>`;
  h += `<h1 style="font-size:26px;margin-bottom:20px">📊 平台统计看板</h1>`;
  h += `<div class="grid grid-4" style="margin-bottom:28px"><div class="stat-card"><div class="stat-num">${d.tutorial_count||0}</div><div class="stat-label">教程总数（已发布 ${d.published_count||0}）</div></div><div class="stat-card"><div class="stat-num">${d.project_count||0}</div><div class="stat-label">改造作品数</div></div><div class="stat-card"><div class="stat-num">${d.user_count||0}</div><div class="stat-label">注册用户</div></div><div class="stat-card"><div class="stat-num">${Object.keys(d.category_stats||{}).length}</div><div class="stat-label">分类数</div></div></div>`;
  h += `<div class="section-title">🔥 热门教程 TOP10</div>`;
  h += `<div class="card" style="padding:8px">${(d.top_tutorials||[]).map((t,i)=>`<div class="flex" style="padding:12px 16px;gap:16px;border-bottom:1px solid #f0f0f0;cursor:pointer" onclick="location.hash='#/tutorials/${t.id}'"><div style="font-weight:800;color:${i<3?'#ff5e6a':'#aaa'};font-size:18px;width:28px">${i+1}</div><img src="${t.cover_after}" style="width:60px;height:60px;border-radius:10px;object-fit:cover"><div style="flex:1"><div style="font-weight:600">${t.title}</div><div style="font-size:12px;color:#888;margin-top:4px">❤️ ${t.favorite_count} · 👁 ${t.view_count} · 🛠 ${t.attempt_count}</div></div></div>`).join('')}</div>`;
  h += `<div class="section-title">🏆 最活跃用户</div>`;
  h += `<div class="grid grid-4">${(d.top_users||[]).map(u=>`<div class="card" style="padding:20px;text-align:center" onclick="location.hash='#/tutorials?user_id=${u.id}'">${avatarFor(u)}<div style="font-weight:600;margin:10px 0 4px">${u.nickname||u.username}</div><div style="font-size:12px;color:#7c5cff">[${levelLabel(u.level)}]</div><div style="font-size:13px;color:#888;margin-top:6px">积分 ${u.score} · 教程 ${u.tutorial_count}</div></div>`).join('')}</div>`;
  if (d.category_stats) {
    h += `<div class="section-title">📈 分类占比</div>`;
    const total = Object.values(d.category_stats).reduce((a,b)=>a+b,1)||1;
    h += `<div class="card" style="padding:20px">${Object.entries(d.category_stats).map(([k,v])=>`<div style="margin-bottom:14px"><div class="flex-between" style="margin-bottom:4px"><span style="font-weight:500">${k}</span><span style="font-size:12px;color:#888">${v} · ${(v*100/total).toFixed(1)}%</span></div><div class="prog"><span style="width:${Math.max(3,v*100/total)}%"></span></div></div>`).join('')}</div>`;
  }
  $('#app').innerHTML = h;
}
async function viewMe() {
  if (!requireLogin()) return;
  const r = await get('/auth/center');
  if (!r.success) return toast(r.message, 'error');
  const s = r.data;
  const u = state.user;
  let h = `<div class="bread"><a href="#/">首页</a> / 个人中心</div>`;
  h += `<div class="card" style="padding:28px;margin-bottom:24px;display:flex;gap:24px;align-items:center"><div style="width:88px;height:88px;border-radius:50%;background:linear-gradient(135deg,#7c5cff,#3bc9db);color:#fff;display:flex;align-items:center;justify-content:center;font-weight:700;font-size:32px">${(u.nickname||u.username).charAt(0).toUpperCase()}</div><div style="flex:1"><h1 style="font-size:24px;margin-bottom:6px">${u.nickname||u.username} <span style="font-size:13px;color:#7c5cff;background:#f8f5ff;padding:3px 10px;border-radius:999px;">[${levelLabel(s.level)}]</span></h1><div style="color:#888">@${u.username} · ${u.email}</div>${u.specialty?`<div style="margin-top:8px;color:#555">🔧 ${u.specialty}</div>`:''}${u.bio?`<div style="margin-top:6px;color:#555">${u.bio}</div>`:''}</div><div style="text-align:right"><button class="btn btn-outline" onclick="editProfile()">编辑资料</button></div></div>`;
  h += `<div class="grid grid-4" style="margin-bottom:24px"><div class="stat-card"><div class="stat-num">${s.tutorial_count||0}</div><div class="stat-label">发布教程</div></div><div class="stat-card"><div class="stat-num">${s.project_count||0}</div><div class="stat-label">改造作品</div></div><div class="stat-card"><div class="stat-num">${s.favorite_count||0}</div><div class="stat-label">我的收藏</div></div><div class="stat-card"><div class="stat-num">${s.total_items||0}</div><div class="stat-label">累计改造 (件)</div></div><div class="stat-card"><div class="stat-num">${s.attempt_count||0}</div><div class="stat-label">尝试中 (${s.completed_count||0}完成)</div></div><div class="stat-card"><div class="stat-num">${s.score||0}</div><div class="stat-label">总积分</div></div></div>`;
  h += `<div class="tabs"><div class="tab tab-active">📚 我的教程</div><div class="tab" onclick="location.hash='#/me/projects'">🎨 我的作品</div><div class="tab" onclick="location.hash='#/me/favorites'">⭐ 我的收藏</div><div class="tab" onclick="location.hash='#/me/attempts'">🛠 我的尝试</div><div class="tab" onclick="location.hash='#/me/messages'">✉️ 消息</div></div>`;
  const tuts = await get('/tutorials?user_id=' + u.id + '&size=20');
  if (tuts.data && tuts.data.list && tuts.data.list.length) {
    h += `<div class="grid grid-4">${renderTutorialCards(tuts.data.list)}</div>`;
  } else {
    h += `<div class="empty">还没有发布教程 <a href="#/editor" class="btn btn-solid">现在发布</a></div>`;
  }
  $('#app').innerHTML = h;
}
function editProfile() {
  const u = state.user;
  const m = el('div', 'modal-bg');
  m.innerHTML = `<div class="modal" style="max-width:500px"><h3>编辑资料</h3><label class="label">昵称</label><input class="input" id="p_nick" value="${u.nickname||''}"><label class="label">头像URL</label><input class="input" id="p_av" value="${u.avatar||''}"><label class="label">擅长领域</label><input class="input" id="p_sp" value="${u.specialty||''}"><label class="label">个人简介</label><textarea id="p_bio" class="input">${u.bio||''}</textarea><div class="modal-actions"><button class="btn btn-gray" onclick="this.closest('.modal-bg').remove()">取消</button><button class="btn btn-solid" id="p_sb">保存</button></div></div>`;
  m.onclick = e => { if (e.target === m) m.remove(); };
  document.body.appendChild(m);
  $('#p_sb').onclick = async () => {
    const r = await post('/auth/me', { nickname: $('#p_nick').value, avatar: $('#p_av').value, specialty: $('#p_sp').value, bio: $('#p_bio').value }, true, 'PUT');
    if (!r.success) return toast(r.message, 'error');
    toast('已更新', 'success'); state.user = null; ensureUser(); m.remove(); route();
  };
}
async function editorView() {
  if (!requireLogin()) return;
  const cats = (await get('/categories', false)).data || [];
  let h = `<div class="bread"><a href="#/">首页</a> / 发布教程</div><h1 style="font-size:26px;margin-bottom:18px">✏️ 教程编辑器</h1>`;
  h += `<div class="card" style="padding:24px"><div class="form-grid"><div><label class="label">教程标题 *</label><input class="input" id="e_title" placeholder="比如：旧牛仔裤改造时尚背包"></div><div><label class="label">分类 *</label><select class="input" id="e_cat">${cats.map(c=>`<option value="${c.id}">${c.icon} ${c.name}</option>`).join('')}</select></div><div><label class="label">难度</label><select class="input" id="e_diff"><option value="easy">简单</option><option value="medium" selected>中等</option><option value="hard">困难</option></select></div><div><label class="label">预计耗时（小时）</label><input class="input" type="number" id="e_hours" value="2" step="0.5" min="0.1"></div><div style="grid-column:1/-1"><label class="label">标签（逗号分隔，如：复古,收纳,极简）</label><input class="input" id="e_tags" placeholder="复古,收纳,极简"></div></div>`;
  h += `<label class="label">简介</label><textarea class="input" id="e_sum" placeholder="简要描述改造思路和亮点..."></textarea>`;
  h += `<div class="form-grid"><div><label class="label">改造前图片URL *</label><input class="input" id="e_cb" placeholder="https://... 或 /uploads/xxx.jpg"></div><div><label class="label">改造后图片URL *</label><input class="input" id="e_ca" placeholder="改造完成后的效果"></div></div>`;
  h += `<div class="section-title">📦 材料清单</div><div id="mat_list"></div><button class="btn btn-gray" onclick="addMat()">+ 添加材料</button>`;
  h += `<div class="section-title">🔨 工具清单</div><div id="tool_list"></div><button class="btn btn-gray" onclick="addTool()">+ 添加工具</button>`;
  h += `<div class="section-title">📖 步骤说明 <small style="font-size:13px;color:#888">拖拽排序</small></div><div id="step_list"></div><button class="btn btn-gray" onclick="addStep()">+ 添加步骤</button>`;
  h += `<div style="margin-top:30px;display:flex;gap:10px;justify-content:flex-end"><button class="btn btn-gray" onclick="saveTutorial(false)">保存草稿</button><button class="btn btn-solid btn-lg" onclick="saveTutorial(true)">🚀 发布教程</button></div></div>`;
  $('#app').innerHTML = h;
  addMat(); addTool(); addStep();
}
function addMat() {
  const div = el('div', 'mat-item');
  div.innerHTML = `<div class="form-grid"><div><label class="label">名称</label><input class="input m-name" placeholder="比如：旧牛仔裤"></div><div><label class="label">数量</label><input class="input m-qty" placeholder="比如：1条"></div><div><label class="label">单位</label><input class="input m-unit" placeholder="条 / 个 / 米"></div><div style="display:flex;align-items:flex-end"><button class="btn btn-danger" onclick="this.closest('.mat-item').remove()">删除</button></div></div>`;
  $('#mat_list').appendChild(div);
}
function addTool() {
  const div = el('div', 'tool-item');
  div.innerHTML = `<div class="form-grid"><div><label class="label">名称</label><input class="input t-name" placeholder="比如：缝纫机"></div><div><label class="label">备注</label><input class="input t-notes" placeholder="可选"></div><div style="display:flex;align-items:flex-end"><button class="btn btn-danger" onclick="this.closest('.tool-item').remove()">删除</button></div></div>`;
  $('#tool_list').appendChild(div);
}
function addStep() {
  const list = $('#step_list');
  const n = list.children.length + 1;
  const div = el('div', 'step');
  div.innerHTML = `<div class="flex-between" style="margin-bottom:10px"><div class="step-index">${n}</div><button class="btn btn-danger" onclick="this.closest('.step').remove();renumSteps()">删除</button></div><label class="label">步骤标题</label><input class="input s-title" placeholder="比如：裁剪裤腿"><label class="label">内容说明</label><textarea class="input s-content" placeholder="详细描述这一步的操作..."></textarea><label class="label">步骤图片URL</label><input class="input s-image" placeholder="可选：图片地址"><div class="form-grid"><div><label class="label">提醒事项</label><input class="input s-remind" placeholder="比如：这一步比较费力 / 注意安全"></div><div><label class="label">预计分钟</label><input class="input s-time" type="number" min="0" value="15"></div></div>`;
  list.appendChild(div);
}
function renumSteps() {
  $$('#step_list .step-index').forEach((el, i) => el.textContent = i + 1);
}
async function saveTutorial(publish) {
  const title = $('#e_title').value.trim();
  const cb = $('#e_cb').value.trim(); const ca = $('#e_ca').value.trim();
  if (!title || !cb || !ca) return toast('请填写必填项：标题 + 改造前后图片', 'error');
  const materials = $$('#mat_list .mat-item').map(x => ({ name: x.querySelector('.m-name').value.trim(), quantity: x.querySelector('.m-qty').value.trim(), unit: x.querySelector('.m-unit').value.trim() })).filter(m => m.name);
  const tools = $$('#tool_list .tool-item').map(x => ({ name: x.querySelector('.t-name').value.trim(), notes: x.querySelector('.t-notes').value.trim(), quantity: '', unit: '' })).filter(m => m.name);
  const steps = $$('#step_list .step').map(x => ({ title: x.querySelector('.s-title').value.trim(), content: x.querySelector('.s-content').value.trim(), image: x.querySelector('.s-image').value.trim(), reminder: x.querySelector('.s-remind').value.trim(), estimated_minutes: parseInt(x.querySelector('.s-time').value)||0 })).filter(s => s.content);
  if (!steps.length) return toast('至少添加一个步骤', 'error');
  const body = {
    category_id: parseInt($('#e_cat').value), title, summary: $('#e_sum').value,
    cover_before: cb, cover_after: ca, difficulty: $('#e_diff').value,
    estimated_hours: parseFloat($('#e_hours').value)||1,
    status: publish ? 'published' : 'draft',
    tags: ($('#e_tags').value.split(/[,，]/).map(x=>x.trim()).filter(Boolean)),
    materials, tools, steps
  };
  const r = await post('/tutorials', body);
  if (!r.success) return toast(r.message, 'error');
  toast(publish ? '发布成功！🎉' : '草稿已保存', 'success');
  location.hash = '#/tutorials/' + r.data.id;
}
async function viewMessages() {
  if (!requireLogin()) return;
  let h = `<div class="bread"><a href="#/">首页</a> / <a href="#/me">个人中心</a> / 消息</div><h1 style="font-size:26px;margin-bottom:18px">✉️ 私信</h1>`;
  h += `<div class="empty">此功能测试中 · 敬请期待</div>`;
  $('#app').innerHTML = h;
}
async function viewFavorites() {
  if (!requireLogin()) return;
  const r = await get('/me/favorites');
  let h = `<div class="bread"><a href="#/">首页</a> / <a href="#/me">个人中心</a> / 收藏</div><h1 style="font-size:26px;margin-bottom:18px">⭐ 我的收藏</h1>`;
  const ids = (r.data?.list||[]).filter(f=>f.target_type==='tutorial').map(f=>f.target_id);
  if (ids.length) {
    const all = await Promise.all(ids.map(id => get('/tutorials/' + id, false)));
    const list = all.filter(r=>r.success).map(r=>r.data.tutorial);
    h += `<div class="grid grid-4">${renderTutorialCards(list)}</div>`;
  } else {
    h += `<div class="empty">还没有收藏内容，去发现一些好教程吧~</div>`;
  }
  $('#app').innerHTML = h;
}
async function viewAttempts() {
  if (!requireLogin()) return;
  const r = await get('/me/attempts');
  let h = `<div class="bread"><a href="#/">首页</a> / <a href="#/me">个人中心</a> / 尝试记录</div><h1 style="font-size:26px;margin-bottom:18px">🛠 我的尝试清单</h1>`;
  const list = r.data?.list || [];
  if (list.length) {
    const ids = [...new Set(list.map(x => x.tutorial_id))];
    const all = await Promise.all(ids.map(id => get('/tutorials/' + id, false)));
    const map = {};
    all.filter(r=>r.success).forEach(r => { map[r.data.tutorial.id] = r.data.tutorial; });
    h += `<div class="grid grid-3">${renderTutorialCards(ids.map(id=>map[id]).filter(Boolean))}</div>`;
  } else {
    h += `<div class="empty">还没有开始尝试任何教程，快行动起来吧！</div>`;
  }
  $('#app').innerHTML = h;
}
async function viewRandom() {
  const r = await get('/random?n=12', false);
  let h = `<div class="bread"><a href="#/">首页</a> / 随机灵感</div><h1 style="font-size:26px;margin-bottom:18px">🎲 随机灵感 · 换一批试试运气</h1>`;
  h += `<div style="margin-bottom:18px"><button class="btn btn-solid" onclick="route()">🔄 换一批</button> <button class="btn btn-outline" onclick="location.hash='#/tutorials?sort=random'">返回教程列表</button></div>`;
  h += `<div class="grid grid-4">${renderTutorialCards(r.data||[])}</div>`;
  $('#app').innerHTML = h;
}
async function route() {
  const hash = location.hash.slice(1) || '/';
  const [path, query] = hash.split('?');
  state.page = 1;
  if (query) {
    const p = new URLSearchParams(query);
    if (p.get('page')) state.page = parseInt(p.get('page'));
  }
  const parts = path.split('/').filter(Boolean);
  const name = parts[0] || 'home';
  $('#app').innerHTML = `<div class="loading">加载中...</div>`;
  try {
    switch (name) {
      case '': case 'home': {
        const r = await get('/home', false);
        $('#app').innerHTML = r.success ? renderHome(r.data) : `<div class="empty">${r.message}</div>`;
        break;
      }
      case 'tutorials': {
        if (parts[1] === undefined) viewTutorials();
        else viewTutorial(parts[1]);
        break;
      }
      case 'projects': {
        if (parts[1] === undefined || parts[1] === 'new') {
          if (parts[1] === 'new') {
            if (!requireLogin()) return;
            editorProjectView();
          } else viewProjects();
        } else viewProject(parts[1]);
        break;
      }
      case 'stats': viewStats(); break;
      case 'editor': editorView(); break;
      case 'random': viewRandom(); break;
      case 'me': {
        const sub = parts[1];
        if (sub === 'projects') viewProjects();
        else if (sub === 'favorites') viewFavorites();
        else if (sub === 'attempts') viewAttempts();
        else if (sub === 'messages') viewMessages();
        else viewMe();
        break;
      }
      default: {
        const r = await get('/home', false);
        $('#app').innerHTML = r.success ? renderHome(r.data) : `<div class="empty">页面不存在</div>`;
      }
    }
  } catch (e) {
    console.error(e);
    $('#app').innerHTML = `<div class="empty">出错了：${e.message || e}</div>`;
  }
  window.scrollTo(0, 0);
}
async function editorProjectView() {
  const t = parseHashQuery();
  let h = `<div class="bread"><a href="#/">首页</a> / 发布作品</div><h1 style="font-size:26px;margin-bottom:18px">🎨 发布改造作品</h1>`;
  h += `<div class="card" style="padding:24px"><label class="label">关联教程ID * ${t.tutorial_id?'(已预填)':''}</label><input class="input" id="pp_tid" placeholder="从教程详情页点击"上传作品"按钮会自动预填" value="${t.tutorial_id||''}"><label class="label">作品标题</label><input class="input" id="pp_title" placeholder="给作品起个好听的名字"><label class="label">图片（多张用 | 分隔URL）</label><input class="input" id="pp_imgs" placeholder="https://... 或 /uploads/a.jpg | /uploads/b.jpg"><label class="label">描述</label><textarea class="input" id="pp_desc" placeholder="分享这次改造的过程、遇到的问题..."></textarea><label class="label">个性化改动（可选）</label><textarea class="input" id="pp_notes" placeholder="做了哪些和教程不同的改动？效果如何？"></textarea><label class="label">评分</label><select class="input" id="pp_rate"><option value="0">未评分</option><option value="1">★</option><option value="2">★★</option><option value="3">★★★</option><option value="4">★★★★</option><option value="5">★★★★★</option></select><div style="margin-top:24px;display:flex;gap:10px;justify-content:flex-end"><button class="btn btn-gray" onclick="history.back()">取消</button><button class="btn btn-solid btn-lg" onclick="submitProject()">发布作品</button></div></div>`;
  $('#app').innerHTML = h;
}
async function submitProject() {
  const body = {
    tutorial_id: parseInt($('#pp_tid').value)||0,
    title: $('#pp_title').value, images: $('#pp_imgs').value,
    description: $('#pp_desc').value, custom_notes: $('#pp_notes').value,
    rating: parseInt($('#pp_rate').value)||0
  };
  if (!body.tutorial_id) return toast('请填写关联教程ID', 'error');
  const r = await post('/projects', body);
  if (!r.success) return toast(r.message, 'error');
  toast('发布成功 🎉', 'success');
  location.hash = '#/projects/' + r.data.id;
}
window.addEventListener('hashchange', route);
window.addEventListener('DOMContentLoaded', async () => {
  await ensureUser();
  if (!location.hash) location.hash = '#/';
  route();
});
