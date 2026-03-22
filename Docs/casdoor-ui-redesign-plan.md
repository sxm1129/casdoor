# Casdoor 前端 UI 深度重构实施图纸 (UI Redesign Plan)

## 目标 (Objective)
采用路线 B 方案，引入 Stitch AI 辅助设计平台对 Casdoor 这款 IAM 中台的核心前端界面（高频 C 端鉴权页、高价值 Admin 仪表盘）进行全局重构，替代原生且过于泛化的 Ant Design 纯表单展现模式。

## 风险评估与预设前提 (Risks & Assumptions)
1. **风险点一：响应式断点兼容风险 (Responsive Degradation)**。
   高保真的双分栏 (Split-Screen) 登录窗在移动端可能显示异常。需要在最终 React 代码中写入严谨的 Media Queries。
2. **风险点二：大盘看板的数据来源 (Dashboard Data Sources)**。
   本次 UI 焕新为纯前端形态。若采用各类数据大图，将暂以现有的 `/api/get-user-count` API 作为数据骨架，或者接入 mock 数据，不涉及繁重后端的 `Aggregation API` 新增开发。
3. **预设前提**：
   本次改版将利用 `mcp_stitch` 工具链全自动生成 `Casdoor UI Redesign` 工程，且直接生成无 AI 生成感的企业级 Azure Slate 风格的高质感应用界面。

---

## IMPLEMENTATION CHECKLIST

### 阶段 1：Stitch 项目创立与原型生成 (Stitch Prototype Generation)
- [x] 1. **创建工作区**: 调用 `mcp_stitch_create_project` 初始化名为 "Casdoor Auth & Admin Redesign" 的 Stitch 空间。
- [x] 2. **生成鉴权门面 (Auth Page Variant)**: 调用 `mcp_stitch_generate_screen_from_text` 或 `generate_variants` 以生成现代高端的大型登录/认证着陆页。
- [x] 3. **生成大盘界面 (Admin Dashboard Variant)**: 调用 `mcp_stitch_generate_screen_from_text` 生成带质感数据统计卡片（Metric Cards）和高密度图表的全新控制台首页。

### 阶段 2：前端代码映射与重写 (React Refactoring)
- [ ] 4. 修改主应用层布局：
      `web/src/App.js` :
      调整总体 Router 的背景设定与全局 AntD CSS Token (如 `ConfigProvider`)，改变其圆角规则与背景基色，为新 UI 铺路。
- [ ] 5. 替换登录流程界面 (Login/Signup Flow)：
      `web/src/auth/` 及其关联登录表单页 :
      引入 Stitch 导出的静态结构代码（HTML/JSX/CSS），并绑定对应的 Casdoor 接口 (如 `/api/login`)。
- [ ] 6. 编写全新大盘面板页面 (Dashboard View)：
      `web/src/pages/Dashboard.js` 或类似的首屏应用 :
      根据 Stitch 生成的复杂 ECharts/图表形态和卡片，手动替换现存极其空洞的 `/admin` 纯表格首页。从后端获取组织全量数据进行 Mock 或渲染。
- [ ] 7. **全局样式梳理 (Global CSS Tuning)**:
      针对 Stitch 生成脱漏出的零散样式进行提取，使用 TailwindCSS 或是原生 CSS 注入到项目中以保证所有卡片风格统一。

### 阶段 3：体验走查与交付 (Verification & Delivery)
- [ ] 8. 编译运行 `yarn start` 查看渲染表现，修复破板或交互不畅的情况。
- [ ] 9. 将变更统一提交至分支，并生成详细的产品焕新文档与 Walkthrough。
