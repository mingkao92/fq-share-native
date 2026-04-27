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
	UploadTitle     string `json:"upload_title"`
	UploadButton    string `json:"upload_button"`
	TextTitle       string `json:"text_title"`
	TextPlaceholder string `json:"text_placeholder"`
	TextButton      string `json:"text_button"`
	TimelineTitle   string `json:"timeline_title"`
	FromComputer    string `json:"from_computer"`
	FromPhone       string `json:"from_phone"`
	NoMessages      string `json:"no_messages"`
	SelectFile      string `json:"select_file"`
	EmptyText       string `json:"empty_text"`
	UploadingFmt    string `json:"uploading_fmt"`
	TextSending     string `json:"text_sending"`
	SuccessFmt      string `json:"success_fmt"`
	TextSaved       string `json:"text_saved"`
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
			UploadTitle:     "发送文件到电脑",
			UploadButton:    "发送文件",
			TextTitle:       "发送文本到电脑",
			TextPlaceholder: "输入要发送的文字...",
			TextButton:      "发送文本",
			TimelineTitle:   "传送记录",
			FromComputer:    "电脑",
			FromPhone:       "手机",
			NoMessages:      "暂时还没有内容",
			SelectFile:      "请先选择文件",
			EmptyText:       "请输入要发送的文字",
			UploadingFmt:    "正在发送 %d 个文件...",
			TextSending:     "正在发送文本...",
			SuccessFmt:      "已发送 %d 个文件",
			TextSaved:       "文本已发送",
			FailedFmt:       "发送失败：%s",
			DownloadLabel:   "下载",
			JustNow:         "刚刚",
		}
	}
	return mobileMessages{
		Lang:            "en",
		Subtitle:        "Send files or text like a chat. Content is kept for up to 24 hours.",
		UploadTitle:     "Send Files to Computer",
		UploadButton:    "Send Files",
		TextTitle:       "Send Text to Computer",
		TextPlaceholder: "Type something to send...",
		TextButton:      "Send Text",
		TimelineTitle:   "Transfer Timeline",
		FromComputer:    "Computer",
		FromPhone:       "Phone",
		NoMessages:      "Nothing here yet",
		SelectFile:      "Please select file(s) first",
		EmptyText:       "Please enter some text",
		UploadingFmt:    "Sending %d file(s)...",
		TextSending:     "Sending text...",
		SuccessFmt:      "Sent %d file(s)",
		TextSaved:       "Text sent",
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
      --bg: #f4f0e6;
      --ink: #1d2b23;
      --muted: #617267;
      --line: #d8e4dc;
      --card: rgba(255,255,255,.88);
      --phone: #fefefe;
      --computer: #0c7c60;
      --computer-ink: #ffffff;
      --surface: #eef6f1;
    }
    * { box-sizing: border-box; }
    body {
      margin: 0;
      font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
      color: var(--ink);
      background:
        radial-gradient(circle at top left, #fff7d6 0, transparent 36%%),
        radial-gradient(circle at top right, #d6efe3 0, transparent 34%%),
        linear-gradient(180deg, #f8f4eb, #eef4ef);
      padding: 14px;
    }
    .shell {
      max-width: 780px;
      margin: 0 auto;
      display: grid;
      gap: 12px;
    }
    .card {
      background: var(--card);
      backdrop-filter: blur(8px);
      border: 1px solid rgba(255,255,255,.7);
      border-radius: 18px;
      box-shadow: 0 16px 40px rgba(0,0,0,.08);
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
    .composeGrid {
      display: grid;
      gap: 12px;
    }
    textarea, input[type=file] {
      width: 100%%;
    }
    textarea {
      min-height: 96px;
      border: 1px solid var(--line);
      border-radius: 14px;
      padding: 12px;
      resize: vertical;
      font: inherit;
      color: var(--ink);
      background: #fff;
    }
    .composer {
      padding: 12px;
      border-radius: 16px;
      background: linear-gradient(180deg, #fdfefe, var(--surface));
      border: 1px solid var(--line);
    }
    button {
      border: none;
      border-radius: 12px;
      padding: 10px 14px;
      background: linear-gradient(145deg, #149271, #0c6a53);
      color: white;
      font-weight: 700;
      cursor: pointer;
      margin-top: 10px;
    }
    .thread {
      list-style: none;
      margin: 0;
      padding: 4px 0 0;
      display: grid;
      gap: 10px;
      max-height: 52vh;
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
      max-width: min(82%%, 520px);
      border-radius: 18px;
      padding: 10px 12px;
      line-height: 1.45;
      white-space: pre-wrap;
      word-break: break-word;
      box-shadow: 0 8px 16px rgba(0,0,0,.05);
    }
    .row.shared .bubble {
      background: linear-gradient(180deg, #11906f, var(--computer));
      color: var(--computer-ink);
      border-bottom-right-radius: 8px;
    }
    .row.uploads .bubble {
      background: #fff;
      border: 1px solid var(--line);
      border-bottom-left-radius: 8px;
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
      background: rgba(255,255,255,.2);
      color: inherit;
      border: 1px solid rgba(255,255,255,.25);
    }
    .row.uploads .bubble .action {
      background: #edf7f2;
      color: #0f7158;
      border-color: #d6ebe2;
    }
    .meta {
      padding: 0 4px;
    }
    .empty {
      text-align: center;
      padding: 18px 10px;
      border: 1px dashed var(--line);
      border-radius: 14px;
      background: rgba(255,255,255,.55);
    }
    @media (min-width: 760px) {
      .composeGrid { grid-template-columns: 1fr 1fr; }
    }
  </style>
</head>
<body>
  <main class="shell">
    <section class="card">
      <h1>Free Quick Share</h1>
      <div class="subtitle">%s</div>
    </section>

    <section class="card composeGrid">
      <div class="composer">
        <h2>%s</h2>
        <form id="textForm">
          <textarea id="textInput" placeholder="%s"></textarea>
          <button type="submit">%s</button>
        </form>
        <p id="textResult" class="status"></p>
      </div>
      <div class="composer">
        <h2>%s</h2>
        <form method="post" action="/api/upload?token=%s" enctype="multipart/form-data" id="uploadForm">
          <input type="file" id="uploadInput" name="file" multiple required />
          <button type="submit">%s</button>
        </form>
        <p id="uploadResult" class="status"></p>
      </div>
    </section>

    <section class="card">
      <h2>%s</h2>
      <ul id="timelineList" class="thread"></ul>
    </section>
  </main>

  <script>
    const token = %s;
    const i18n = %s;
    const uploadForm = document.getElementById('uploadForm');
    const uploadInput = document.getElementById('uploadInput');
    const uploadResult = document.getElementById('uploadResult');
    const textForm = document.getElementById('textForm');
    const textInput = document.getElementById('textInput');
    const textResult = document.getElementById('textResult');
    const timelineList = document.getElementById('timelineList');

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

    function renderTimeline(rows) {
      if (!rows.length) {
        timelineList.innerHTML = '<li class="empty hint">' + i18n.no_messages + '</li>';
        return;
      }
      timelineList.innerHTML = rows.map((row) => {
        const isShared = row.bucket === 'shared';
        const peer = isShared ? i18n.from_computer : i18n.from_phone;
        const klass = isShared ? 'shared' : 'uploads';
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
          '<div class="meta">' + peer + ' · ' + formatTime(row.mtime) + '</div>' +
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

    uploadForm.addEventListener('submit', async (event) => {
      event.preventDefault();
      const files = Array.from(uploadInput.files || []);
      if (!files.length) {
        uploadResult.textContent = i18n.select_file;
        return;
      }

      const formData = new FormData();
      for (const file of files) {
        formData.append('file', file, file.name);
      }

      uploadResult.textContent = fmt(i18n.uploading_fmt, files.length);
      try {
        const res = await fetch('/api/upload?token=' + encodeURIComponent(token), {
          method: 'POST',
          body: formData,
          headers: { Accept: 'application/json' },
        });
        if (!res.ok) {
          throw new Error('HTTP ' + res.status);
        }
        const payload = await res.json();
        const savedCount = (payload.saved || []).length;
        uploadResult.textContent = fmt(i18n.success_fmt, savedCount);
        uploadInput.value = '';
        await loadList();
      } catch (error) {
        uploadResult.textContent = fmt(i18n.failed_fmt, error.message);
      }
    });

    textForm.addEventListener('submit', async (event) => {
      event.preventDefault();
      const text = String(textInput.value || '').trim();
      if (!text) {
        textResult.textContent = i18n.empty_text;
        return;
      }
      textResult.textContent = i18n.text_sending;
      try {
        const res = await fetch('/api/text?token=' + encodeURIComponent(token), {
          method: 'POST',
          headers: { 'Content-Type': 'application/json', Accept: 'application/json' },
          body: JSON.stringify({ text }),
        });
        if (!res.ok) {
          throw new Error('HTTP ' + res.status);
        }
        textInput.value = '';
        textResult.textContent = i18n.text_saved;
        await loadList();
      } catch (error) {
        textResult.textContent = fmt(i18n.failed_fmt, error.message);
      }
    });

    loadList();
    setInterval(loadList, 5000);
  </script>
</body>
</html>
`, msg.Lang, html.EscapeString(msg.Subtitle), html.EscapeString(msg.TextTitle), html.EscapeString(msg.TextPlaceholder), html.EscapeString(msg.TextButton), html.EscapeString(msg.UploadTitle), html.EscapeString(token), html.EscapeString(msg.UploadButton), html.EscapeString(msg.TimelineTitle), string(tokenJSON), string(msgJSON))
}
