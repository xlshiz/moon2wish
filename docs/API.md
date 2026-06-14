# API 文档

所有接口需要在 Header 中携带认证信息：

```
Authorization: Token <your-token>
```

## Highlights

### 创建高亮

```
POST /api/moon/v2/highlights/
```

**请求体：**

```json
{
  "highlights": [
    {
      "text": "高亮内容",
      "title": "书名",
      "author": "作者",
      "note": "笔记",
      "chapter": "章节",
      "color": "yellow"
    }
  ]
}
```

也支持单个高亮：

```json
{
  "text": "高亮内容",
  "title": "书名",
  "author": "作者"
}
```

**字段说明：**

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| text | string | 是 | 高亮文本 |
| title | string | 是 | 书名 |
| author | string | 否 | 作者 |
| note | string | 否 | 笔记 |
| chapter | string | 否 | 章节 |
| color | string | 否 | 颜色 |
| source | string | 否 | 来源 |
| book_type | string | 否 | 书籍类型 |

**响应：**

```json
{
  "highlights": [
    {
      "id": 1,
      "readwise_id": "5fe9aeb543f4c80d0935276946bf9326",
      "text": "高亮内容",
      "title": "书名",
      "author": "作者",
      "note": "笔记",
      "source": "来源",
      "color": "yellow",
      "book_type": "",
      "created_at": "2026-06-13T20:56:55+08:00",
      "updated_at": "2026-06-13T20:56:55+08:00"
    }
  ]
}
```

### 获取所有高亮

```
GET /api/moon/v2/highlights/
```

**响应：**

```json
{
  "highlights": [
    {
      "id": 1,
      "readwise_id": "...",
      "text": "...",
      "title": "...",
      "author": "...",
      "note": "...",
      "source": "...",
      "color": "...",
      "book_type": "...",
      "created_at": "...",
      "updated_at": "..."
    }
  ]
}
```

### 获取单个高亮

```
GET /api/moon/v2/highlights/:id
```

**参数：**

| 参数 | 类型 | 说明 |
|------|------|------|
| id | int | 高亮 ID |

**响应：**

```json
{
  "id": 1,
  "readwise_id": "...",
  "text": "...",
  "title": "...",
  "author": "...",
  "note": "...",
  "source": "...",
  "color": "...",
  "book_type": "...",
  "created_at": "...",
  "updated_at": "..."
}
```

### 更新高亮

```
PATCH /api/moon/v2/highlights/:id
```

**参数：**

| 参数 | 类型 | 说明 |
|------|------|------|
| id | int | 高亮 ID |

**请求体：**

```json
{
  "text": "新内容",
  "note": "新备注",
  "title": "新书名",
  "author": "新作者",
  "source": "新来源",
  "source_url": "新来源URL",
  "color": "red",
  "book_type": "new"
}
```

**字段说明：**

| 字段 | 类型 | 说明 |
|------|------|------|
| text | string | 高亮文本 |
| note | string | 笔记 |
| title | string | 书名 |
| author | string | 作者 |
| source | string | 来源 |
| source_url | string | 来源URL (映射到 source 字段) |
| color | string | 颜色 |
| book_type | string | 书籍类型 |

所有字段都是可选的，只更新提供的字段。

**响应：**

```json
{
  "id": 1,
  "readwise_id": "...",
  "text": "新内容",
  "title": "新书名",
  "author": "新作者",
  "note": "新备注",
  "source": "新来源",
  "color": "red",
  "book_type": "new",
  "created_at": "2026-06-13T20:56:55+08:00",
  "updated_at": "2026-06-14T10:30:00+08:00"
}
```

### 删除单个高亮

```
DELETE /api/moon/v2/highlights/:id
```

**响应：**

```json
{
  "message": "deleted"
}
```

### 删除所有高亮

```
DELETE /api/moon/v2/highlights/
```

**响应：**

```json
{
  "message": "all highlights deleted"
}
```

## Books

### 获取所有书籍

```
GET /api/moon/v2/books/
```

**响应：**

```json
[
  {
    "id": 1,
    "title": "书名",
    "author": "作者",
    "num_highlights": 5
  }
]
```

### 获取单个书籍

```
GET /api/moon/v2/books/:id
```

**响应：**

```json
{
  "id": 1,
  "title": "书名",
  "author": "作者",
  "num_highlights": 5
}
```

## Export

### JSON 导出

```
GET /api/moon/export/json
```

**响应：**

```json
{
  "highlights": [...],
  "total": 10,
  "exported_at": "2026-06-13T12:57:44Z"
}
```

### CSV 导出

```
GET /api/moon/export/csv
```

**响应：**

```csv
id,text,note,title,author,source,color,created_at
1,高亮内容,,书名,作者,,,2026-06-13T20:56:55+08:00
```

## 错误响应

| 状态码 | 说明 |
|--------|------|
| 400 | 请求参数错误 |
| 401 | 未认证或 Token 无效 |
| 404 | 资源不存在 |
| 415 | Content-Type 不正确 |
| 429 | 请求过于频繁 |
| 500 | 服务器内部错误 |

**错误格式：**

```json
{
  "error": "错误信息"
}
```

## 安全特性

### 安全 HTTP 头

调试模式下启用以下安全头：

- `X-Content-Type-Options: nosniff`
- `X-Frame-Options: DENY`
- `X-XSS-Protection: 1; mode=block`
- `Content-Security-Policy: default-src 'none'`
- `Cache-Control: no-store`

### 内容类型验证

POST、PUT、PATCH 请求会验证 Content-Type 是否为 `application/json`。

### 请求体限制

请求体大小限制为 1MB。

### Token 认证

支持时序攻击防护的 Token 比较。

## curl 示例

```bash
# 创建高亮
curl -X POST http://localhost:8080/api/moon/v2/highlights/ \
  -H "Authorization: Token your-token" \
  -H "Content-Type: application/json" \
  -d '{"highlights":[{"text":"test","title":"book"}]}'

# 获取所有高亮
curl http://localhost:8080/api/moon/v2/highlights/ \
  -H "Authorization: Token your-token"

# 更新高亮
curl -X PATCH http://localhost:8080/api/moon/v2/highlights/1 \
  -H "Authorization: Token your-token" \
  -H "Content-Type: application/json" \
  -d '{"text":"updated text","note":"new note"}'

# 删除高亮
curl -X DELETE http://localhost:8080/api/moon/v2/highlights/1 \
  -H "Authorization: Token your-token"

# 导出 JSON
curl http://localhost:8080/api/moon/export/json \
  -H "Authorization: Token your-token"
```
