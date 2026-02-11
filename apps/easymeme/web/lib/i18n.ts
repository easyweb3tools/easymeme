export type Lang = "zh" | "en";

export function resolveLang(
  input?: string | string[] | null,
  acceptLanguage?: string | null,
): Lang {
  const value = Array.isArray(input) ? input[0] : input;
  if (value === "en" || value === "zh") {
    return value;
  }
  const langHeader = (acceptLanguage ?? "").toLowerCase();
  if (langHeader.includes("zh")) return "zh";
  if (langHeader.includes("en")) return "en";
  return "zh";
}

export function withLang(path: string, lang: Lang): string {
  const separator = path.includes("?") ? "&" : "?";
  return `${path}${separator}lang=${lang}`;
}

type Dict = Record<string, string>;

const ZH: Dict = {
  nav_golden: "金狗列表",
  nav_trades: "AI 交易历史",
  nav_github: "GitHub",
  home_badge: "BNB Chain • Autonomous Agent",
  home_title: "你的专属 AI Meme 币猎手",
  home_desc:
    "EasyMeme 持续发现、分析并追踪金狗机会。基于 OpenClaw 的学习型 Agent，支持个人自部署与长期运行。",
  home_view_golden: "查看金狗列表",
  home_view_trades: "AI 交易历史",
  home_view_github: "查看 GitHub",
  home_deploy_title: "一键自部署",
  home_deploy_1: "1. 拉取仓库并配置 .env",
  home_deploy_2: "2. docker compose up --build 启动服务",
  home_deploy_3: "3. OpenClaw 连接 Server 自动分析",
  home_deploy_4: "4. Web 查看金狗与 AI 决策",
  home_cards_title: "今日金狗机会",
  home_cards_sub: "已按有效分数排序，展示最新优先机会。",
  home_cards_view_all: "查看全部 →",
  home_cards_empty: "暂无金狗数据，请稍后再试。",
  home_feature_1_title: "动态金狗时效",
  home_feature_1_desc: "通过时间衰减模型把握黄金窗口期。",
  home_feature_2_title: "可学习策略",
  home_feature_2_desc: "OpenClaw Memory 让规则随反馈进化。",
  home_feature_3_title: "个人部署",
  home_feature_3_desc: "每个人都能拥有自己的 AI 交易系统。",
  gd_back_home: "← 返回首页",
  gd_title: "金狗列表",
  gd_sub: "按有效分数排序，已过滤 EXPIRED。",
  gd_total: "共 {count} 个机会",
  gd_top: "最高有效分数 {score}",
  gd_filters_keyword: "关键词",
  gd_filters_placeholder: "Symbol / Name / Address",
  gd_filters_min_score: "最小有效分数",
  gd_filters_risk: "风险等级",
  gd_filters_sort: "排序",
  gd_filters_order: "排序方向",
  gd_filters_page_size: "每页数量",
  gd_filters_all: "全部",
  gd_filters_submit: "筛选",
  gd_sort_effective: "有效分数",
  gd_sort_golden: "金狗分数",
  gd_sort_risk: "风险分数",
  gd_order_desc: "降序",
  gd_order_asc: "升序",
  gd_empty: "暂无金狗数据，请稍后再试。",
  gd_effective: "Effective Score",
  gd_golden: "Golden Dog Score",
  gd_risk: "Risk Score",
  gd_decay: "Time Decay",
  page: "Page {page} / {total}",
  prev: "上一页",
  next: "下一页",
  trades_title: "AI 交易历史",
  trades_sub: "仅展示 Agent 自动交易记录",
  trades_filters_user: "用户 ID",
  trades_filters_status: "状态筛选",
  trades_pl_label: "盈亏",
  trades_status_label: "状态",
  trades_amount_in: "成交金额",
  trades_amount_out: "实际成交",
  trades_strategy: "策略",
  trades_decision: "决策理由",
  trades_info: "交易信息",
  trades_status_all: "全部",
  trades_status_success: "SUCCESS",
  trades_status_pending: "PENDING",
  trades_status_failed: "FAILED",
  trades_filters_submit: "筛选",
  trades_stat_total: "总交易数",
  trades_stat_win: "胜率",
  trades_stat_avg: "平均盈亏",
  trades_stat_strategy: "分策略胜率",
  trades_stat_period: "周期收益",
  trades_empty: "暂无 AI 交易记录。",
  trades_back_home: "← 返回首页",
  trades_view_golden: "查看金狗列表",
  token_back: "← 返回金狗列表",
  token_summary: "决策摘要",
  token_reasoning: "Reasoning",
  token_recommendation: "Recommendation",
  token_basic: "基础信息",
  token_dex: "DEX",
  token_liquidity: "Liquidity",
  token_creator: "Creator",
  token_created: "Created At",
  token_analyzed: "Analyzed At",
  token_tools: "外部工具",
  token_risk_factors: "风险因子",
  token_analysis: "AI 分析详情",
  token_effective: "有效分数",
  token_golden: "金狗分数",
  token_risk_score: "风险评分",
  token_phase: "Phase",
  token_time_decay: "Time Decay"
  ,copy: "复制"
  ,copied: "已复制"
  ,token_verdict_title: "🐕 金狗判断"
  ,token_verdict_score: "分数: {score} / 100"
  ,token_indicator_pass: "✅ 通过"
  ,token_indicator_medium: "⚠️ 中等"
  ,token_indicator_fail: "❌ 高风险"
  ,token_indicator_safety: "Safety"
  ,token_indicator_tax: "Tax"
  ,token_indicator_ownership: "Ownership"
  ,token_indicator_momentum: "Momentum"
  ,token_detail_honeypot_yes: "检测到蜜罐"
  ,token_detail_honeypot_no: "未检测到蜜罐"
  ,token_detail_tax: "买入 {buy}% / 卖出 {sell}%"
  ,token_detail_mintable_yes: "可增发 ⚠️"
  ,token_detail_mintable_no: "不可增发"
  ,token_verdict_positive: "该代币通过了关键安全检查。"
  ,token_verdict_negative: "该代币当前不满足金狗条件。"
  ,token_verdict_momentum: "1小时买卖比：{buys} / {sells}。"
  ,token_verdict_wait: "建议先观望，等待更强动量或更低风险。"
  ,token_contract_safety_title: "🔒 合约安全（GoPlus）"
  ,token_contract_safety_empty: "合约安全数据暂不可用。"
  ,token_safety_honeypot: "蜜罐"
  ,token_safety_open_source: "开源"
  ,token_safety_mintable: "可增发"
  ,token_safety_proxy: "代理合约"
  ,token_safety_take_back_ownership: "可回收所有权"
  ,token_safety_holders: "持有人数"
  ,token_safety_lp_holders: "LP 持有人数"
  ,token_yes: "是"
  ,token_no: "否"
  ,token_market_title: "📊 市场数据（DEXScreener）"
  ,token_market_empty: "市场数据暂不可用。"
  ,token_market_price: "价格"
  ,token_market_volume_h1: "1小时成交量"
  ,token_market_liquidity: "流动性"
  ,token_market_txns_h1: "1小时交易"
  ,token_market_buys: "买入"
  ,token_market_sells: "卖出"
  ,token_holder_title: "👥 持仓分布"
  ,token_holder_empty: "持仓分布数据暂不可用。"
  ,token_holder_top10: "Top 10 占比"
  ,token_holder_total: "追踪持有人总数"
  ,token_alerts_title: "⚠️ 市场预警"
  ,token_alert_change: "变化"
};

const EN: Dict = {
  nav_golden: "Golden Dogs",
  nav_trades: "AI Trades",
  nav_github: "GitHub",
  home_badge: "BNB Chain • Autonomous Agent",
  home_title: "Your Personal AI Meme Hunter",
  home_desc:
    "EasyMeme continuously discovers, analyzes, and tracks golden dog opportunities. Powered by OpenClaw, built for long-running personal deployment.",
  home_view_golden: "View Golden Dogs",
  home_view_trades: "AI Trades",
  home_view_github: "View GitHub",
  home_deploy_title: "One-Click Deploy",
  home_deploy_1: "1. Clone repo and configure .env",
  home_deploy_2: "2. docker compose up --build",
  home_deploy_3: "3. OpenClaw connects to Server",
  home_deploy_4: "4. Web shows golden dogs & AI decisions",
  home_cards_title: "Today's Golden Dogs",
  home_cards_sub: "Sorted by effective score with latest opportunities.",
  home_cards_view_all: "View all →",
  home_cards_empty: "No golden dogs yet. Try again soon.",
  home_feature_1_title: "Time-Sensitive Golden Dogs",
  home_feature_1_desc: "Catch the window with time-decay scoring.",
  home_feature_2_title: "Learning Strategy",
  home_feature_2_desc: "OpenClaw Memory evolves with feedback.",
  home_feature_3_title: "Personal Deployment",
  home_feature_3_desc: "Everyone can run their own AI trading system.",
  gd_back_home: "← Back to Home",
  gd_title: "Golden Dogs",
  gd_sub: "Sorted by effective score, EXPIRED filtered out.",
  gd_total: "{count} opportunities",
  gd_top: "Top effective score {score}",
  gd_filters_keyword: "Keyword",
  gd_filters_placeholder: "Symbol / Name / Address",
  gd_filters_min_score: "Min Effective Score",
  gd_filters_risk: "Risk Level",
  gd_filters_sort: "Sort By",
  gd_filters_order: "Order",
  gd_filters_page_size: "Page Size",
  gd_filters_all: "All",
  gd_filters_submit: "Filter",
  gd_sort_effective: "Effective Score",
  gd_sort_golden: "Golden Dog Score",
  gd_sort_risk: "Risk Score",
  gd_order_desc: "Desc",
  gd_order_asc: "Asc",
  gd_empty: "No golden dogs yet. Try again soon.",
  gd_effective: "Effective Score",
  gd_golden: "Golden Dog Score",
  gd_risk: "Risk Score",
  gd_decay: "Time Decay",
  page: "Page {page} / {total}",
  prev: "Prev",
  next: "Next",
  trades_title: "AI Trades",
  trades_sub: "Only agent auto-trade records are shown.",
  trades_filters_user: "User ID",
  trades_filters_status: "Status",
  trades_pl_label: "P/L",
  trades_status_label: "Status",
  trades_amount_in: "Amount In",
  trades_amount_out: "Amount Out",
  trades_strategy: "Strategy",
  trades_decision: "Decision Reason",
  trades_info: "Trade Info",
  trades_status_all: "All",
  trades_status_success: "SUCCESS",
  trades_status_pending: "PENDING",
  trades_status_failed: "FAILED",
  trades_filters_submit: "Filter",
  trades_stat_total: "Total Trades",
  trades_stat_win: "Win Rate",
  trades_stat_avg: "Avg P/L",
  trades_stat_strategy: "Win Rate by Strategy",
  trades_stat_period: "Performance by Period",
  trades_empty: "No AI trades yet.",
  trades_back_home: "← Back to Home",
  trades_view_golden: "View Golden Dogs",
  token_back: "← Back to Golden Dogs",
  token_summary: "Decision Summary",
  token_reasoning: "Reasoning",
  token_recommendation: "Recommendation",
  token_basic: "Basics",
  token_dex: "DEX",
  token_liquidity: "Liquidity",
  token_creator: "Creator",
  token_created: "Created At",
  token_analyzed: "Analyzed At",
  token_tools: "External Tools",
  token_risk_factors: "Risk Factors",
  token_analysis: "AI Analysis",
  token_effective: "Effective Score",
  token_golden: "Golden Dog Score",
  token_risk_score: "Risk Score",
  token_phase: "Phase",
  token_time_decay: "Time Decay"
  ,copy: "Copy"
  ,copied: "Copied"
  ,token_verdict_title: "🐕 Golden Dog Verdict"
  ,token_verdict_score: "Score: {score} / 100"
  ,token_indicator_pass: "✅ PASS"
  ,token_indicator_medium: "⚠️ MEDIUM"
  ,token_indicator_fail: "❌ HIGH"
  ,token_indicator_safety: "Safety"
  ,token_indicator_tax: "Tax"
  ,token_indicator_ownership: "Ownership"
  ,token_indicator_momentum: "Momentum"
  ,token_detail_honeypot_yes: "Honeypot detected"
  ,token_detail_honeypot_no: "No honeypot"
  ,token_detail_tax: "Buy {buy}% / Sell {sell}%"
  ,token_detail_mintable_yes: "Mintable ⚠️"
  ,token_detail_mintable_no: "Not mintable"
  ,token_verdict_positive: "This token passes key safety checks."
  ,token_verdict_negative: "This token does not currently meet golden dog criteria."
  ,token_verdict_momentum: "1h buys/sells: {buys} / {sells}."
  ,token_verdict_wait: "Wait for stronger momentum or lower risk signals."
  ,token_contract_safety_title: "🔒 Contract Safety (GoPlus)"
  ,token_contract_safety_empty: "Contract safety data is not available yet."
  ,token_safety_honeypot: "Honeypot"
  ,token_safety_open_source: "Open Source"
  ,token_safety_mintable: "Mintable"
  ,token_safety_proxy: "Proxy Contract"
  ,token_safety_take_back_ownership: "Can Take Back Ownership"
  ,token_safety_holders: "Holders"
  ,token_safety_lp_holders: "LP Holders"
  ,token_yes: "Yes"
  ,token_no: "No"
  ,token_market_title: "📊 Market Data (DEXScreener)"
  ,token_market_empty: "Market data is not available yet."
  ,token_market_price: "Price"
  ,token_market_volume_h1: "Volume (1h)"
  ,token_market_liquidity: "Liquidity"
  ,token_market_txns_h1: "Transactions (1h)"
  ,token_market_buys: "buys"
  ,token_market_sells: "sells"
  ,token_holder_title: "👥 Holder Distribution"
  ,token_holder_empty: "Holder distribution data is not available yet."
  ,token_holder_top10: "Top 10 holders"
  ,token_holder_total: "Total tracked holders"
  ,token_alerts_title: "⚠️ Market Alerts"
  ,token_alert_change: "Change"
};

export function t(lang: Lang, key: keyof typeof ZH, vars?: Record<string, string | number>) {
  const dict = lang === "en" ? EN : ZH;
  let text = dict[key] ?? ZH[key] ?? key;
  if (vars) {
    for (const [k, v] of Object.entries(vars)) {
      text = text.replace(`{${k}}`, String(v));
    }
  }
  return text;
}
