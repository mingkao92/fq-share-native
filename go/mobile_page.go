package main

import (
	"encoding/json"
	"fmt"
	"html"
	"strings"
)

type mobileMessages struct {
	Lang            string `json:"lang"`
	Subtitle        string `json:"subtitle"`
	ComposerTitle   string `json:"composer_title"`
	TextPlaceholder string `json:"text_placeholder"`
	AddFileButton   string `json:"add_file_button"`
	SendButton      string `json:"send_button"`
	TimelineTitle   string `json:"timeline_title"`
	FromComputer    string `json:"from_computer"`
	FromPhone       string `json:"from_phone"`
	NoMessages      string `json:"no_messages"`
	SelectFile      string `json:"select_file"`
	EmptyText       string `json:"empty_text"`
	RemoveFile      string `json:"remove_file"`
	FailedFmt       string `json:"failed_fmt"`
	DownloadLabel   string `json:"download_label"`
	JustNow         string `json:"just_now"`
}

func normalizeLang(raw string) string {
	lower := strings.ToLower(strings.TrimSpace(raw))
	if strings.HasPrefix(lower, "zh") {
		return "zh-CN"
	}
	return "en"
}

func mobileMessagesFor(lang string) mobileMessages {
	if normalizeLang(lang) == "zh-CN" {
		return mobileMessages{
			Lang:            "zh-CN",
			Subtitle:        "像聊天一样发送文件或文本，内容最多保留 24 小时。",
			ComposerTitle:   "发送到电脑",
			TextPlaceholder: "输入要发送的文字...",
			AddFileButton:   "添加文件",
			SendButton:      "发送",
			TimelineTitle:   "传送记录",
			FromComputer:    "电脑",
			FromPhone:       "手机",
			NoMessages:      "暂时还没有内容",
			SelectFile:      "请先选择文件",
			EmptyText:       "请输入要发送的文字",
			RemoveFile:      "移除文件",
			FailedFmt:       "发送失败：%s",
			DownloadLabel:   "下载",
			JustNow:         "刚刚",
		}
	}
	return mobileMessages{
		Lang:            "en",
		Subtitle:        "Send files or text like a chat. Content is kept for up to 24 hours.",
		ComposerTitle:   "Send to Computer",
		TextPlaceholder: "Type something to send...",
		AddFileButton:   "Add File",
		SendButton:      "Send",
		TimelineTitle:   "Transfer Timeline",
		FromComputer:    "Computer",
		FromPhone:       "Phone",
		NoMessages:      "Nothing here yet",
		SelectFile:      "Please select file(s) first",
		EmptyText:       "Please enter some text",
		RemoveFile:      "Remove File",
		FailedFmt:       "Send failed: %s",
		DownloadLabel:   "Download",
		JustNow:         "just now",
	}
}

func renderMobilePage(token string, lang string) string {
	msg := mobileMessagesFor(lang)
	tokenJSON, _ := json.Marshal(token)
	msgJSON, _ := json.Marshal(msg)

	return fmt.Sprintf(`<!doctype html>
<html lang="%s">
<head>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width,initial-scale=1" />
  <title>Free Quick Share</title>
  <style>
    :root {
      --bg-a: #f6f4ee;
      --bg-b: #dff2e9;
      --panel: rgba(255,255,255,.82);
      --ink: #152019;
      --muted: #5a6d63;
      --line: #dbe6df;
      --brand: #13795b;
      --brand-deep: #0c5b44;
      --brand-line: rgba(19,121,91,.18);
    }
    * { box-sizing: border-box; }
    body {
      margin: 0;
      font-family: "PingFang SC", "Noto Sans SC", "Microsoft YaHei", "Segoe UI", sans-serif;
      color: var(--ink);
      background:
        radial-gradient(circle at 12%% 0%%, #fefee8 0, transparent 40%%),
        radial-gradient(circle at 90%% 6%%, #c9ecd8 0, transparent 45%%),
        linear-gradient(155deg, var(--bg-a), var(--bg-b));
      padding: 16px;
    }
    .shell {
      max-width: 780px;
      margin: 0 auto;
      display: grid;
      gap: 12px;
    }
    .card {
      border: 1px solid rgba(255,255,255,.7);
      background: linear-gradient(180deg, var(--panel), rgba(248,252,250,.96));
      backdrop-filter: blur(10px);
      border-radius: 18px;
      box-shadow: 0 18px 30px rgba(9,49,35,.07);
      padding: 14px;
    }
    h1 {
      margin: 0 0 6px;
      color: #0f6d55;
      font-size: 24px;
    }
    h2 {
      margin: 0 0 8px;
      font-size: 15px;
      color: #3d584c;
    }
    .subtitle, .hint, .status, .meta {
      color: var(--muted);
      font-size: 13px;
    }
    button {
      border: 1px solid var(--brand-line);
      background: rgba(247,251,249,.98);
      color: var(--brand-deep);
      border-radius: 10px;
      padding: 9px 12px;
      font-size: 13px;
      font-weight: 600;
      cursor: pointer;
      box-shadow: none;
      transition: transform .08s ease, background .12s ease, border-color .12s ease, color .12s ease, opacity .12s ease;
    }
    button:hover {
      border-color: rgba(19,121,91,.3);
      background: rgba(240,248,244,.98);
    }
    button:active {
      transform: scale(.985);
      background: rgba(232,243,237,.98);
    }
    button:focus-visible {
      outline: none;
      box-shadow: 0 0 0 3px rgba(19,121,91,.14);
    }
    button:disabled {
      cursor: not-allowed;
      transform: none;
      color: #98aaa1;
      border-color: rgba(152,170,161,.28);
      background: rgba(243,246,245,.96);
      box-shadow: none;
      opacity: 1;
    }
    button:disabled:hover,
    button:disabled:active {
      transform: none;
      border-color: rgba(152,170,161,.28);
      background: rgba(243,246,245,.96);
    }
    .composer {
      display: grid;
      gap: 10px;
    }
    .composerBar {
      display: grid;
      grid-template-columns: 1fr auto;
      gap: 10px;
      align-items: end;
    }
    .inputWrap {
      min-height: 54px;
      border: 1px solid var(--line);
      border-radius: 14px;
      padding: 8px 10px;
      background: #fff;
      transition: border-color .12s ease, background .12s ease, box-shadow .12s ease;
    }
    .inputWrap.hasFiles {
      border-color: rgba(19,121,91,.38);
      background: linear-gradient(180deg, #fbfffd, #f2fbf6);
    }
    .inputRow {
      display: flex;
      align-items: center;
      gap: 8px;
      min-height: 38px;
      min-width: 0;
      flex-wrap: wrap;
    }
    .pickBtn {
      min-width: 34px;
      width: 34px;
      height: 34px;
      padding: 0;
      border-radius: 10px;
      background: #eef6f2;
      color: var(--brand-deep);
      box-shadow: none;
      font-size: 20px;
      line-height: 1;
    }
    .pickBtn:disabled {
      opacity: .45;
    }
    .chips {
      display: inline-flex;
      flex-wrap: wrap;
      gap: 6px;
      align-items: center;
      min-width: 0;
      flex: 0 1 auto;
    }
    .chips.hidden { display: none; }
    .chip {
      display: inline-flex;
      align-items: center;
      gap: 6px;
      max-width: 200px;
      padding: 5px 9px;
      border-radius: 10px;
      background: #eef8f3;
      color: var(--brand-deep);
      white-space: nowrap;
      font-size: 13px;
    }
    .chipText {
      overflow: hidden;
      text-overflow: ellipsis;
    }
    .chip button {
      padding: 0;
      margin: 0;
      background: transparent;
      color: inherit;
      box-shadow: none;
      border-radius: 0;
      font-size: 14px;
      min-width: auto;
    }
    .inputWrap textarea {
      flex: 1 1 140px;
      min-width: 0;
      min-height: 28px;
      max-height: 120px;
      padding: 5px 2px 5px 4px;
      border: 0;
      resize: none;
      background: transparent;
      font: inherit;
      color: var(--ink);
      line-height: 1.4;
      font-size: 15px;
      overflow-y: auto;
    }
    .inputWrap textarea:focus {
      outline: none;
    }
    .inputWrap textarea.isHidden {
      flex-basis: 0;
      width: 0;
      min-width: 0;
      padding: 0;
      opacity: 0;
      overflow: hidden;
    }
    .sendBtn {
      align-self: stretch;
      margin: 0;
    }
    .composerStatus {
      min-height: 18px;
      margin: 6px 4px 0;
      color: #a2373a;
    }
    .composerStatus:empty {
      display: none;
    }
    .thread {
      list-style: none;
      margin: 0;
      padding: 4px 2px 10px;
      display: grid;
      gap: 10px;
      min-height: 240px;
      max-height: 42vh;
      overflow: auto;
    }
    .row {
      display: flex;
      flex-direction: column;
      gap: 4px;
    }
    .row.shared { align-items: flex-end; }
    .row.uploads { align-items: flex-start; }
    .bubble {
      max-width: min(82%%, 480px);
      padding: 11px 13px;
      border-radius: 12px;
      line-height: 1.5;
      white-space: pre-wrap;
      word-break: break-word;
      font-size: 15px;
      box-shadow: 0 6px 14px rgba(0,0,0,.04);
    }
    .row.shared .bubble {
      background: linear-gradient(145deg, #0b8a6a, var(--brand-deep));
      color: #fff;
    }
    .row.uploads .bubble {
      background: #fff;
      color: var(--ink);
      border: 1px solid var(--line);
    }
    .bubble .title {
      font-weight: 700;
      margin-bottom: 2px;
    }
    .bubble .action {
      display: inline-flex;
      margin-top: 8px;
      padding: 6px 10px;
      border-radius: 999px;
      text-decoration: none;
      font-size: 12px;
      font-weight: 700;
      background: rgba(255,255,255,.14);
      color: inherit;
      border: 1px solid rgba(255,255,255,.18);
      border-radius: 999px;
    }
    .row.uploads .bubble .action {
      background: #eef8f3;
      color: var(--brand-deep);
      border-color: #d6ebe2;
    }
    .meta {
      padding: 0 4px;
    }
    .deviceTag {
      margin-left: 6px;
      padding: 2px 7px;
      border-radius: 999px;
      background: rgba(10,127,99,.1);
      color: var(--brand-deep);
      font-size: 12px;
    }
    .empty {
      text-align: center;
      padding: 18px 10px;
      border: 1px dashed var(--line);
      border-radius: 14px;
      background: rgba(255,255,255,.55);
    }
    @media (max-width: 520px) {
      .composerBar {
        grid-template-columns: 1fr auto;
      }
    }
  </style>
</head>
<body>
  <main class="shell">
    <section class="card">
      <h1>Free Quick Share</h1>
      <div class="subtitle">%s</div>
    </section>

    <section class="card">
      <h2>%s</h2>
      <ul id="timelineList" class="thread"></ul>
    </section>

    <section class="card">
      <h2>%s</h2>
      <form id="composerForm" class="composer">
        <input type="file" id="uploadInput" name="file" multiple hidden />
        <div class="composerBar">
          <div id="inputWrap" class="inputWrap">
            <div class="inputRow">
              <button id="pickFileBtn" class="pickBtn" type="button" aria-label="%s">+</button>
              <div id="selectedFiles" class="chips hidden"></div>
              <textarea id="textInput" rows="1" placeholder="%s"></textarea>
            </div>
          </div>
          <button id="sendBtn" class="sendBtn" type="submit">%s</button>
        </div>
        <p id="composerStatus" class="composerStatus"></p>
      </form>
    </section>
  </main>

  <script>
    const token = %s;
    const i18n = %s;
    const composerForm = document.getElementById('composerForm');
    const uploadInput = document.getElementById('uploadInput');
    const pickFileBtn = document.getElementById('pickFileBtn');
    const selectedFiles = document.getElementById('selectedFiles');
    const inputWrap = document.getElementById('inputWrap');
    const textInput = document.getElementById('textInput');
    const sendBtn = document.getElementById('sendBtn');
    const composerStatus = document.getElementById('composerStatus');
    const timelineList = document.getElementById('timelineList');
    let pendingFiles = [];

    function prettySize(bytes) {
      if (bytes < 1024) return bytes + ' B';
      if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB';
      if (bytes < 1024 * 1024 * 1024) return (bytes / (1024 * 1024)).toFixed(1) + ' MB';
      return (bytes / (1024 * 1024 * 1024)).toFixed(1) + ' GB';
    }

    function fmt(template, value) {
      return String(template || '').replace('%%d', String(value)).replace('%%s', String(value));
    }

    function escapeHtml(text) {
      return String(text || '')
        .replace(/&/g, '&amp;')
        .replace(/</g, '&lt;')
        .replace(/>/g, '&gt;')
        .replace(/"/g, '&quot;')
        .replace(/'/g, '&#39;');
    }

    function formatTime(unixSeconds) {
      if (!unixSeconds) return i18n.just_now;
      const date = new Date(unixSeconds * 1000);
      return date.toLocaleString(i18n.lang === 'zh-CN' ? 'zh-CN' : 'en-US', {
        month: 'numeric',
        day: 'numeric',
        hour: '2-digit',
        minute: '2-digit',
      });
    }

    const deviceAliases = new Map();

    function deviceFingerprint(row) {
      return [
        row.device_id,
        row.device_name,
        row.remote_addr,
        row.user_agent,
      ].filter(Boolean).join('|');
    }

    function deviceAlias(row) {
      const fingerprint = deviceFingerprint(row);
      if (!fingerprint) return i18n.from_phone;
      if (!deviceAliases.has(fingerprint)) {
        const letter = String.fromCharCode(65 + deviceAliases.size);
        deviceAliases.set(fingerprint, i18n.from_phone + ' ' + letter);
      }
      return deviceAliases.get(fingerprint);
    }

    function resizeComposer() {
      if (pendingFiles.length > 0) {
        textInput.style.height = '28px';
        return;
      }
      textInput.style.height = 'auto';
      textInput.style.height = Math.min(textInput.scrollHeight, 120) + 'px';
    }

    function updateComposerMode() {
      const hasFiles = pendingFiles.length > 0;
      const hasText = Boolean(String(textInput.value || '').trim());
      pickFileBtn.disabled = hasText;
      textInput.readOnly = hasFiles;
      inputWrap.classList.toggle('hasFiles', hasFiles);
      if (hasFiles) {
        textInput.value = '';
        textInput.classList.add('isHidden');
      } else {
        textInput.classList.remove('isHidden');
      }
    }

    function renderSelectedFiles() {
      if (!pendingFiles.length) {
        selectedFiles.innerHTML = '';
        selectedFiles.classList.add('hidden');
        return;
      }
      selectedFiles.classList.remove('hidden');
      selectedFiles.innerHTML = pendingFiles.map((file, index) => {
        return '<span class="chip">' +
          '<span class="chipText">' + escapeHtml(file.name) + '</span>' +
          '<button type="button" class="removeFileBtn" data-index="' + index + '" aria-label="' + escapeHtml(i18n.remove_file) + '">×</button>' +
          '</span>';
      }).join('');
    }

    function resetComposer() {
      pendingFiles = [];
      uploadInput.value = '';
      composerStatus.textContent = '';
      updateComposerMode();
      renderSelectedFiles();
      resizeComposer();
    }

    function renderTimeline(rows) {
      if (!rows.length) {
        timelineList.innerHTML = '<li class="empty hint">' + i18n.no_messages + '</li>';
        return;
      }
      timelineList.innerHTML = rows.map((row) => {
        const isShared = row.bucket === 'shared';
        const peer = isShared ? i18n.from_computer : i18n.from_phone;
        const klass = isShared ? 'shared' : 'uploads';
        const device = isShared ? '' : '<span class="deviceTag">' + escapeHtml(deviceAlias(row)) + '</span>';
        let body = '';
        if (row.kind === 'text') {
          body = '<div>' + escapeHtml(row.text).replace(/\n/g, '<br>') + '</div>';
        } else {
          const href = '/api/download/' + row.bucket + '/' + encodeURIComponent(row.name) + '?token=' + encodeURIComponent(token);
          body = '<div class="title">' + escapeHtml(row.name) + '</div>' +
            '<div>' + prettySize(row.size || 0) + '</div>' +
            '<a class="action" href="' + href + '" download>' + i18n.download_label + '</a>';
        }
        return '<li class="row ' + klass + '">' +
          '<div class="meta">' + peer + device + ' · ' + formatTime(row.mtime) + '</div>' +
          '<div class="bubble">' + body + '</div>' +
          '</li>';
      }).join('');
      timelineList.scrollTop = timelineList.scrollHeight;
    }

    async function loadList() {
      try {
        const res = await fetch('/api/list?token=' + encodeURIComponent(token), { cache: 'no-store' });
        if (!res.ok) return;
        const payload = await res.json();
        renderTimeline(payload.timeline || []);
      } catch (_) {}
    }

    pickFileBtn.addEventListener('click', () => {
      uploadInput.click();
    });

    uploadInput.addEventListener('change', () => {
      const files = Array.from(uploadInput.files || []);
      if (!files.length) return;
      if (String(textInput.value || '').trim()) {
        uploadInput.value = '';
        return;
      }
      pendingFiles = pendingFiles.concat(files);
      composerStatus.textContent = '';
      updateComposerMode();
      renderSelectedFiles();
      resizeComposer();
      uploadInput.value = '';
    });

    selectedFiles.addEventListener('click', (event) => {
      const button = event.target.closest('.removeFileBtn');
      if (!button) return;
      const index = Number(button.dataset.index);
      if (Number.isNaN(index)) return;
      pendingFiles = pendingFiles.filter((_, currentIndex) => currentIndex !== index);
      if (!pendingFiles.length) {
        composerStatus.textContent = '';
      }
      updateComposerMode();
      renderSelectedFiles();
      resizeComposer();
    });

    textInput.addEventListener('input', () => {
      updateComposerMode();
      resizeComposer();
    });

    composerForm.addEventListener('submit', async (event) => {
      event.preventDefault();
      const text = String(textInput.value || '').trim();
      const hasFiles = pendingFiles.length > 0;
      const hasText = Boolean(text);
      if (!hasFiles && !hasText) {
        composerStatus.textContent = i18n.empty_text;
        return;
      }

      sendBtn.disabled = true;
      pickFileBtn.disabled = true;
      try {
        if (hasFiles) {
          const formData = new FormData();
          for (const file of pendingFiles) {
            formData.append('file', file, file.name);
          }
          const res = await fetch('/api/upload?token=' + encodeURIComponent(token), {
            method: 'POST',
            body: formData,
            headers: { Accept: 'application/json' },
          });
          if (!res.ok) {
            throw new Error('HTTP ' + res.status);
          }
        } else {
          const res = await fetch('/api/text?token=' + encodeURIComponent(token), {
            method: 'POST',
            headers: { 'Content-Type': 'application/json', Accept: 'application/json' },
            body: JSON.stringify({ text }),
          });
          if (!res.ok) {
            throw new Error('HTTP ' + res.status);
          }
        }
        textInput.value = '';
        resetComposer();
        await loadList();
      } catch (error) {
        composerStatus.textContent = fmt(i18n.failed_fmt, error.message);
      } finally {
        sendBtn.disabled = false;
        updateComposerMode();
      }
    });

    updateComposerMode();
    renderSelectedFiles();
    resizeComposer();
    loadList();
    setInterval(loadList, 5000);
  </script>
</body>
</html>
`, msg.Lang, html.EscapeString(msg.Subtitle), html.EscapeString(msg.TimelineTitle), html.EscapeString(msg.ComposerTitle), html.EscapeString(msg.AddFileButton), html.EscapeString(msg.TextPlaceholder), html.EscapeString(msg.SendButton), string(tokenJSON), string(msgJSON))
}
