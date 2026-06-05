---
trigger: model_decision
---

### E2E 测试自动化规则（最高优先级）

1. **Playwright 测试执行**：
   - 运行测试时必须使用非阻塞模式。
   - 命令必须包含：`CI=true` 环境变量，或在 playwright.config 中确保 html reporter 的 `open: 'never'`。
   - 严禁使用会启动服务器并等待 Ctrl+C 的命令（如默认 `npx playwright show-report`）。
   - 执行完成后，自动检查 `playwright-report/index.html` 是否生成，并总结测试结果（通过/失败/截图路径）。

2. **自动交互协议**：
   - 任何需要人工输入（Ctrl+C、按键、确认）的步骤，必须通过配置或 flag 绕过。
   - 如果检测到 "Serving HTML report" 或 "Press Ctrl+C" 提示，立即终止进程并记录为配置错误，切换到非阻塞模式重试。

3. **命令模板**（必须遵守）：
   - 测试运行：`CI=true npx playwright test --reporter=html`
   - 查看报告：仅读取静态 HTML 文件，不启动服务器。