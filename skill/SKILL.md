---
name: xhs
description: >
  搜索、阅读、互动小红书内容。使用 xhs CLI 通过浏览器自动化直接操作小红书。
  Use when: (1) 用户要搜小红书笔记, (2) 用户分享了小红书链接,
  (3) 用户要看小红书热门内容, (4) 用户要点赞/收藏/评论小红书笔记,
  (5) 用户要查看小红书用户主页。
  Triggers: "搜小红书", "小红书", "xiaohongshu", "xhs", "红书搜", "小红书热门",
  "小红书笔记", "看小红书", "小红书评论", "小红书点赞", "xhslink.com",
  "xiaohongshu.com".
---

# xhs CLI — 小红书操作指南

通过 `xhs` 命令直接操作小红书，所有命令加 `--json` 输出结构化数据。

## 前置检查

```bash
which xhs          # 确认已安装
xhs login status --json   # 检查登录状态
```

如果未登录，运行 `xhs login` 打开浏览器扫码。

## 搜索笔记

```bash
# 基础搜索
xhs search "关键词" --json

# 限制数量
xhs search "关键词" --limit 10 --json

# 带筛选
xhs search "关键词" --sort "最新" --json
xhs search "关键词" --sort "最多点赞" --type "视频" --time "一周内" --json
```

筛选参数：
- `--sort`: 综合 | 最新 | 最多点赞 | 最多评论 | 最多收藏
- `--type`: 不限 | 视频 | 图文
- `--time`: 不限 | 一天内 | 一周内 | 半年内
- `--scope`: 不限 | 已看过 | 未看过 | 已关注
- `--location`: 不限 | 同城 | 附近

## 搜索热门笔记的推荐流程

当用户要求搜热门笔记时，按以下步骤操作：

1. **搜索**：用 `xhs search` 获取列表，注意保存每条结果的 `id` 和 `xsecToken`
2. **筛选热门**：根据 `likedCount`（点赞数）排序，找出互动量最高的笔记
3. **获取详情**：对感兴趣的笔记调用 `xhs detail` 获取全文
4. **汇总呈现**：整理标题、作者、点赞/收藏数、内容摘要给用户

示例完整流程：
```bash
# Step 1: 搜索
xhs search "AI编程" --sort "最多点赞" --limit 10 --json

# Step 2: 从结果中取 id 和 xsecToken，获取详情
xhs detail <note_id> --xsec-token <token> --json

# Step 3: 如需评论
xhs detail <note_id> --xsec-token <token> --comments --comment-limit 20 --json
```

## 查看笔记详情

```bash
# 基本详情
xhs detail <note_id> --xsec-token <token> --json

# 加载评论
xhs detail <note_id> --xsec-token <token> --comments --json

# 控制评论数量和展开子回复
xhs detail <note_id> --xsec-token <token> --comments --comment-limit 50 --replies --json
```

> `note_id` 和 `xsec-token` 从 search 结果的 `id` 和 `xsecToken` 字段获取。

## 获取首页推荐

```bash
xhs feed --json
```

## 查看用户主页

```bash
xhs user <user_id> --xsec-token <token> --json
```

> `user_id` 和 `xsec-token` 从笔记详情的 `user.userId` 和搜索结果的 `xsecToken` 获取。

## 互动操作

```bash
# 点赞 / 取消点赞
xhs like <note_id> --xsec-token <token>
xhs like <note_id> --xsec-token <token> --undo

# 收藏 / 取消收藏
xhs favorite <note_id> --xsec-token <token>
xhs favorite <note_id> --xsec-token <token> --undo

# 评论
xhs comment <note_id> "评论内容" --xsec-token <token>

# 回复评论
xhs reply <note_id> "回复内容" --xsec-token <token> --comment-id <comment_id>
```

## 输出格式

搜索结果关键字段：
```json
{
  "id": "笔记ID，用于 detail/like/comment",
  "xsecToken": "访问令牌，传给 --xsec-token",
  "noteCard": {
    "displayTitle": "标题",
    "user": { "nickname": "作者", "userId": "用户ID" },
    "interactInfo": {
      "likedCount": "点赞数",
      "collectedCount": "收藏数",
      "commentCount": "评论数"
    }
  }
}
```

## 注意事项

- 所有需要 `xsec-token` 的命令，token 来自搜索结果，**有时效性**，过期需重新搜索获取
- Cookie 保存在 `~/.xhs-cli/cookies.json`，通常 1-2 周需重新登录
- 如果命令报错，先检查登录状态：`xhs login status --json`
