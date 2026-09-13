package service

// exportTemplate 是 A4 打印版教程的自包含 HTML 模板。
// 排版要点：
//   - @page size:A4，内容区按纸张宽度布局
//   - 图片 max-width:100% + height:auto，按纸宽等比缩放且不跨页撕裂
//   - 材料/工具用表格，thead 跨页自动重复，行级 break-inside:avoid，
//     材料再多也自然分页、不截断任何一行
//   - 步骤编号、提醒框完整保留；步骤头部与首段不孤行
const exportTemplate = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>{{.Title}} · A4 打印版</title>
<style>
  * { margin: 0; padding: 0; box-sizing: border-box; }
  html { -webkit-print-color-adjust: exact; print-color-adjust: exact; }
  body {
    font-family: "PingFang SC", "Hiragino Sans GB", "Microsoft YaHei", "Noto Sans CJK SC", sans-serif;
    color: #1f2430; font-size: 11pt; line-height: 1.75;
    max-width: 180mm; margin: 0 auto; padding: 24px 16px 60px;
  }
  /* ---- 屏幕端工具条，打印时隐藏 ---- */
  .toolbar {
    display: flex; align-items: center; gap: 14px; flex-wrap: wrap;
    padding: 12px 16px; margin-bottom: 22px;
    background: #f4f1ff; border: 1px solid #e2d9ff; border-radius: 10px;
  }
  .toolbar button {
    font-size: 14px; font-weight: 600; cursor: pointer;
    padding: 9px 22px; border: none; border-radius: 8px;
    background: #7c5cff; color: #fff;
  }
  .toolbar .tip { font-size: 12px; color: #7768ae; }
  /* ---- 头部 ---- */
  header.doc { border-bottom: 3px solid #7c5cff; padding-bottom: 14px; margin-bottom: 18px; }
  header.doc h1 { font-size: 22pt; line-height: 1.35; margin-bottom: 10px; }
  .meta { display: flex; flex-wrap: wrap; gap: 8px 16px; align-items: center; font-size: 10pt; color: #555; }
  .badge {
    display: inline-block; padding: 2px 12px; border-radius: 999px;
    background: #7c5cff; color: #fff; font-size: 9.5pt; font-weight: 600;
  }
  .tags span { color: #7c5cff; margin-right: 8px; }
  /* ---- 章节 ---- */
  section { margin-bottom: 20px; }
  h2 {
    font-size: 13.5pt; margin-bottom: 10px; padding-left: 10px;
    border-left: 5px solid #7c5cff; break-after: avoid;
  }
  .summary { color: #333; white-space: pre-wrap; }
  /* ---- 对比图 ---- */
  .covers { display: flex; gap: 4%; }
  .covers figure { width: 48%; break-inside: avoid; }
  .covers figcaption {
    display: inline-block; font-size: 9pt; font-weight: 600;
    padding: 2px 10px; border-radius: 6px 6px 0 0; margin-bottom: -1px;
  }
  .covers .before figcaption { background: #ffe3e3; color: #c92a2a; }
  .covers .after figcaption { background: #d3f9d8; color: #2b8a3e; }
  img { max-width: 100%; height: auto; break-inside: avoid; }
  .covers img { width: 100%; border-radius: 0 8px 8px 8px; border: 1px solid #eee; }
  /* ---- 材料 / 工具表格：行不截断，长表自然分页，表头跨页重复 ---- */
  table { width: 100%; border-collapse: collapse; font-size: 10.5pt; }
  thead { display: table-header-group; }
  tr { break-inside: avoid; }
  th, td { border: 1px solid #d8dce6; padding: 6px 10px; text-align: left; vertical-align: top; }
  th { background: #f2f0fb; font-weight: 600; }
  td.qty { white-space: nowrap; width: 18%; }
  td.notes { color: #666; width: 32%; }
  /* ---- 步骤 ---- */
  .step { margin-bottom: 18px; }
  .step-head {
    display: flex; align-items: center; gap: 10px;
    break-inside: avoid; break-after: avoid; margin-bottom: 6px;
  }
  .step-num {
    flex: none; width: 26px; height: 26px; border-radius: 50%;
    background: #7c5cff; color: #fff; font-weight: 700; font-size: 11pt;
    display: flex; align-items: center; justify-content: center;
  }
  .step-head h3 { font-size: 12pt; }
  .step-mins { margin-left: auto; font-size: 9pt; color: #888; white-space: nowrap; }
  .step img { display: block; margin: 8px 0; border-radius: 8px; border: 1px solid #eee; }
  .step p { white-space: pre-wrap; color: #2b2f38; }
  .reminder {
    margin-top: 8px; padding: 8px 12px; break-inside: avoid;
    background: #fff9db; border-left: 4px solid #fcc419; border-radius: 4px;
    color: #7a5f00; font-size: 10pt;
  }
  /* ---- 页脚 ---- */
  footer.doc {
    margin-top: 30px; padding-top: 10px; border-top: 1px solid #ddd;
    font-size: 9pt; color: #999; display: flex; justify-content: space-between; flex-wrap: wrap; gap: 6px;
  }
  /* ---- A4 打印 ---- */
  @page { size: A4; margin: 16mm 15mm; }
  @media print {
    body { max-width: none; padding: 0; }
    .toolbar { display: none; }
    a { color: inherit; text-decoration: none; }
  }
</style>
</head>
<body>
<div class="toolbar">
  <button onclick="window.print()">🖨️ 打印 / 另存为 PDF</button>
  <span class="tip">本页为教程快照（版本 v{{.Version}} · 导出于 {{.ExportedAt}}），内容与线上后续修改无关。打印时请选 A4 纸张。</span>
</div>

<header class="doc">
  <h1>{{.Title}}</h1>
  <div class="meta">
    <span class="badge">难度：{{.Difficulty}}</span>
    <span>⏱ 预计耗时：{{.Hours}} 小时</span>
    {{if .Category}}<span>📂 {{.Category}}</span>{{end}}
    {{if .Author}}<span>✍️ {{.Author}}</span>{{end}}
    {{if .Tags}}<span class="tags">{{range .Tags}}<span>#{{.}}</span>{{end}}</span>{{end}}
  </div>
</header>

{{if or .CoverBefore .CoverAfter}}
<section class="covers">
  {{if .CoverBefore}}<figure class="before"><figcaption>改造前</figcaption><img src="{{.CoverBefore}}" alt="改造前"></figure>{{end}}
  {{if .CoverAfter}}<figure class="after"><figcaption>改造后</figcaption><img src="{{.CoverAfter}}" alt="改造后"></figure>{{end}}
</section>
{{end}}

{{if .Summary}}
<section>
  <h2>📝 简介</h2>
  <p class="summary">{{.Summary}}</p>
</section>
{{end}}

{{if .Materials}}
<section>
  <h2>📦 材料清单（共 {{len .Materials}} 项）</h2>
  <table>
    <thead><tr><th>#</th><th>名称</th><th>数量</th><th>备注</th></tr></thead>
    <tbody>
      {{range $i, $m := .Materials}}
      <tr><td>{{inc $i}}</td><td>{{$m.Name}}</td><td class="qty">{{$m.Quantity}}{{if and $m.Quantity $m.Unit}} {{end}}{{$m.Unit}}</td><td class="notes">{{$m.Notes}}</td></tr>
      {{end}}
    </tbody>
  </table>
</section>
{{end}}

{{if .Tools}}
<section>
  <h2>🔨 工具清单（共 {{len .Tools}} 项）</h2>
  <table>
    <thead><tr><th>#</th><th>名称</th><th>数量</th><th>备注</th></tr></thead>
    <tbody>
      {{range $i, $m := .Tools}}
      <tr><td>{{inc $i}}</td><td>{{$m.Name}}</td><td class="qty">{{$m.Quantity}}{{if and $m.Quantity $m.Unit}} {{end}}{{$m.Unit}}</td><td class="notes">{{$m.Notes}}</td></tr>
      {{end}}
    </tbody>
  </table>
</section>
{{end}}

{{if .Steps}}
<section>
  <h2>📖 步骤说明（共 {{len .Steps}} 步）</h2>
  {{range .Steps}}
  <div class="step">
    <div class="step-head">
      <span class="step-num">{{.Index}}</span>
      <h3>{{if .Title}}{{.Title}}{{else}}步骤 {{.Index}}{{end}}</h3>
      {{if .EstimatedMinutes}}<span class="step-mins">⏱ 约 {{.EstimatedMinutes}} 分钟</span>{{end}}
    </div>
    {{if .Image}}<img src="{{.Image}}" alt="步骤 {{.Index}}">{{end}}
    <p>{{.Content}}</p>
    {{if .Reminder}}<div class="reminder">⚠️ {{.Reminder}}</div>{{end}}
  </div>
  {{end}}
</section>
{{end}}

<footer class="doc">
  <span>旧物改造灵感与教程平台 · A4 打印快照</span>
  <span>教程版本 v{{.Version}} · 导出于 {{.ExportedAt}}</span>
</footer>
</body>
</html>
`
