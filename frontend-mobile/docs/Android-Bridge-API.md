# Android WebView JS Bridge 接口文档

> 版本：v1.2
> 适用于：排班系统移动端 `frontend-mobile`

---

## 概述

安卓端通过 WebView 的 `addJavascriptInterface` 注入一个名为 `Android` 的 JavaScript 对象，H5 页面通过 `window.Android.xxx()` 调用原生方法。

```java
webView.addJavascriptInterface(new NativeBridge(), "Android");
```

安卓端需要实现以下 3 个方法：

| 方法 | 用途 | 调用时机 |
|------|------|----------|
| `fetchScheduleFromSchool()` | 获取课表 xlsx 文件（base64） | 用户点击"从教务系统直接导入" |
| `shareSchedule(json)` | 保存/分享排班表 xlsx 文件 | 用户点击"导出Excel" |
| `downloadAndShare(json)` | 下载文件并分享 | 用户在安卓环境下载文件时 |

---

## 核心原理

`<input type="file">` 选择文件后得到的本质是 `File` 对象（`Blob` 的子类），H5 端可以直接用 base64 数据构造出完全相同的 `File` 对象，从而无缝复用现有的文件上传接口。

**不需要返回文件路径**——JS 无法访问安卓本地文件系统。

安卓端直接返回文件内容的 base64，H5 端构造 `File` 对象后走正常的表单上传流程。

```
安卓端: 生成 xlsx → 读取字节 → Base64.encodeToString → JSON 返回
H5端:   解析 JSON → atob(base64) → new Uint8Array → new File([blob]) → FormData 上传
```

---

## 方法详细说明

### 1. fetchScheduleFromSchool()

从教务系统抓取课表，生成 xlsx 文件，将文件内容以 base64 形式直接返回。

**调用方式**（同步，在 JS 线程执行）：

```javascript
const jsonStr = window.Android.fetchScheduleFromSchool();
```

**返回值**：JSON 字符串

```json
// 成功
{
  "success": true,
  "base64": "UEsDBBQABgAIAAAAIQBi7K...",  // xlsx 文件内容，不含 data: 前缀
  "fileName": "schedule_2024001_2025-2026-2.xlsx",
  "mimeType": "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
}

// 失败
{
  "success": false,
  "message": "教务系统登录过期，请重新登录"
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `success` | boolean | 是 | 是否成功 |
| `base64` | string | 成功时必填 | xlsx 文件内容的 Base64 编码（不含 `data:` 前缀） |
| `fileName` | string | 成功时建议填写 | 文件名（用于构造 File 对象） |
| `mimeType` | string | 否 | MIME 类型，默认 `application/vnd.openxmlformats-officedocument.spreadsheetml.sheet` |
| `message` | string | 失败时建议填写 | 错误描述，会显示给用户 |

**安卓端实现要点**：

1. 调用教务系统 API 抓取课表数据
2. 使用 Apache POI 生成 xlsx 到内存（`ByteArrayOutputStream`）或缓存文件
3. `Base64.encodeToString(bytes, Base64.NO_WRAP)` 转为 base64
4. 返回 JSON 字符串
5. 此方法为**同步调用**，所有耗时操作必须在调用前完成

**H5 端处理流程**：

```javascript
// 1. 调用原生方法，获得 JSON
const result = JSON.parse(window.Android.fetchScheduleFromSchool())

// 2. base64 → Uint8Array → File（等同于用户手动选择文件得到的 File 对象）
const binary = atob(result.base64)
const bytes = new Uint8Array(binary.length)
for (let i = 0; i < binary.length; i++) bytes[i] = binary.charCodeAt(i)
const file = new File([bytes], result.fileName, { type: result.mimeType })

// 3. 直接复用现有 XLS 导入接口
const formData = new FormData()
formData.append('file', file)
axios.post('/api/v1/availability/import/xls', formData)
```

---

### 2. shareSchedule(json)

将排班表文件保存到本地或通过系统分享面板分享。

**调用方式**（同步）：

```javascript
const result = window.Android.shareSchedule(jsonString);
// 返回字符串 "true" 或 "false"
```

**参数**：JSON 字符串

```json
{
  "week": 5,
  "fileName": "排班表_第5周.xlsx",
  "fileData": "UEsDBBQABgAIAAAAIQBi7K..."  // xlsx base64
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| `week` | number | 排班所属周次 |
| `fileName` | string | 建议保存的文件名 |
| `fileData` | string | xlsx 文件内容的 Base64 编码（不含 `data:` 前缀） |

**返回值**：字符串

| 值 | 说明 |
|------|------|
| `"true"` | 成功 |
| `"false"` | 用户取消或失败 |

**安卓端实现要点**：

1. `Base64.decode(fileData, Base64.DEFAULT)` 解码为字节数组
2. 写入文件（建议 `Downloads` 目录）
3. 通过 `Intent.ACTION_SEND` 弹出系统分享面板
4. 需配置 `FileProvider` 以支持 `content://` URI

---

### 3. downloadAndShare(json) ⭐ 新增

下载文件并调用系统分享面板。支持 GET/POST 请求，由安卓端完成网络下载。

**调用方式**（同步）：

```javascript
const result = window.Android.downloadAndShare(jsonString);
// 返回 JSON 字符串
```

**参数**：JSON 字符串

```json
{
  "url": "https://api.example.com/admin/schedule/export",
  "fileName": "排班表_第5周.xlsx",
  "method": "POST",
  "headers": {
    "Authorization": "Bearer eyJhbGciOiJIUzI1NiIs...",
    "Content-Type": "application/json"
  },
  "body": "{\"week\":5,\"templateId\":1,\"department\":\"办公室\"}"
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `url` | string | 是 | 完整的下载 URL（需包含域名） |
| `fileName` | string | 是 | 保存的文件名 |
| `method` | string | 否 | 请求方法，默认 `GET` |
| `headers` | object | 否 | 请求头对象 |
| `body` | string | 否 | POST 请求体（JSON 字符串） |

**返回值**：JSON 字符串

```json
// 成功
{
  "success": true
}

// 失败
{
  "success": false,
  "message": "网络请求失败: 401 Unauthorized"
}
```

**安卓端实现要点**：

1. 使用 OkHttp 或 Retrofit 发起网络请求
2. 将 `headers` 对象中的所有键值对添加到请求头
3. 如果 `method` 为 `POST` 且有 `body`，发送 JSON 请求体
4. 下载完成后保存到 `Downloads` 目录
5. 通过 `Intent.ACTION_SEND` 弹出系统分享面板
6. 需配置 `FileProvider` 以支持 `content://` URI

**Kotlin 实现示例**：

```kotlin
@JavascriptInterface
fun downloadAndShare(json: String): String {
    return try {
        val data = JSONObject(json)
        val url = data.getString("url")
        val fileName = data.getString("fileName")
        val method = data.optString("method", "GET").uppercase()
        val headers = data.optJSONObject("headers")
        val body = data.optString("body", null)

        // 构建请求
        val client = OkHttpClient.Builder()
            .connectTimeout(30, TimeUnit.SECONDS)
            .readTimeout(60, TimeUnit.SECONDS)
            .build()

        val requestBuilder = Request.Builder().url(url)

        // 添加请求头
        headers?.let {
            val keys = it.keys()
            while (keys.hasNext()) {
                val key = keys.next()
                requestBuilder.addHeader(key, it.getString(key))
            }
        }

        // POST 请求体
        if (method == "POST" && body != null) {
            val mediaType = "application/json; charset=utf-8".toMediaType()
            requestBuilder.post(body.toRequestBody(mediaType))
        }

        val response = client.newCall(requestBuilder.build()).execute()
        if (!response.isSuccessful) {
            return JSONObject().apply {
                put("success", false)
                put("message", "下载失败: ${response.code}")
            }.toString()
        }

        // 保存文件
        val bytes = response.body?.bytes() ?: return JSONObject().apply {
            put("success", false)
            put("message", "响应体为空")
        }.toString()

        val file = File(
            Environment.getExternalStoragePublicDirectory(Environment.DIRECTORY_DOWNLOADS),
            fileName
        )
        file.writeBytes(bytes)

        // 分享
        val uri = FileProvider.getUriForFile(
            context,
            "${context.packageName}.fileprovider",
            file
        )
        val intent = Intent(Intent.ACTION_SEND).apply {
            type = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
            putExtra(Intent.EXTRA_STREAM, uri)
            addFlags(Intent.FLAG_GRANT_READ_URI_PERMISSION)
        }
        context.startActivity(Intent.createChooser(intent, "分享文件"))

        JSONObject().apply { put("success", true) }.toString()
    } catch (e: Exception) {
        JSONObject().apply {
            put("success", false)
            put("message", e.message ?: "下载失败")
        }.toString()
    }
}
```

---

## H5 端完整调用流程

### 无课表导入（Availability 页面）

```
用户点击"从教务系统直接导入"
  ↓
fetchScheduleFromSchool()          → 安卓返回 { success, base64, fileName }
  ↓
atob(base64) → Uint8Array → File   → 构造出和 <input type="file"> 相同的 File 对象
  ↓
POST /api/v1/availability/import/xls (multipart/form-data)
  ↓
后端解析 xlsx，写入无课表数据库
  ↓
刷新无课表列表
```

### 排班表导出（Schedule 页面）

```
用户点击"导出Excel"
  ↓
[安卓 WebView - 使用 downloadAndShare]
  获取 API 地址和 Token
  downloadAndShare({ url, fileName, method: 'POST', headers, body })
  安卓下载文件 + 弹出系统分享面板
  ↓
[浏览器 - 使用原生下载]
  blob 下载 → URL.createObjectURL → <a> 点击下载
```

---

## Android 端完整实现模板（Kotlin）

```kotlin
class NativeBridge(private val context: Context) {

    @JavascriptInterface
    fun fetchScheduleFromSchool(): String {
        return try {
            // 1. 从教务系统抓取课表
            val courses = fetchFromJWXT()

            // 2. 生成 xlsx 到内存
            val outputStream = ByteArrayOutputStream()
            val workbook = XSSFWorkbook()
            // ... 使用 POI 填充课表数据 ...
            workbook.write(outputStream)
            workbook.close()
            val bytes = outputStream.toByteArray()

            // 3. 转 base64
            val base64 = Base64.encodeToString(bytes, Base64.NO_WRAP)

            // 4. 返回 JSON
            JSONObject().apply {
                put("success", true)
                put("base64", base64)
                put("fileName", "schedule_${studentId}_${semester}.xlsx")
                put("mimeType", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
            }.toString()
        } catch (e: Exception) {
            JSONObject().apply {
                put("success", false)
                put("message", e.message ?: "获取课表失败")
            }.toString()
        }
    }

    @JavascriptInterface
    fun shareSchedule(json: String): String {
        return try {
            val data = JSONObject(json)
            val fileName = data.getString("fileName")
            val fileData = data.getString("fileData")

            // 1. base64 解码
            val bytes = Base64.decode(fileData, Base64.DEFAULT)

            // 2. 写入 Downloads 目录
            val file = File(
                Environment.getExternalStoragePublicDirectory(Environment.DIRECTORY_DOWNLOADS),
                fileName
            )
            file.writeBytes(bytes)

            // 3. 弹出系统分享面板
            val uri = FileProvider.getUriForFile(
                context,
                "${context.packageName}.fileprovider",
                file
            )
            val intent = Intent(Intent.ACTION_SEND).apply {
                type = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
                putExtra(Intent.EXTRA_STREAM, uri)
                addFlags(Intent.FLAG_GRANT_READ_URI_PERMISSION)
            }
            context.startActivity(Intent.createChooser(intent, "分享排班表"))

            "true"
        } catch (e: Exception) {
            "false"
        }
    }
}
```

---

## WebView 配置

```kotlin
webView.settings.apply {
    javaScriptEnabled = true
    domStorageEnabled = true
    allowContentAccess = true
}
```

---

## 注意事项

1. **同步返回**：`@JavascriptInterface` 方法在 JS 线程调用，必须同步返回。耗时操作（如网络请求）需提前完成
2. **Base64 编码**：使用 `Base64.NO_WRAP`，不要包含换行符
3. **错误处理**：失败时返回 `{ success: false, message: "错误描述" }`，message 会显示给用户
4. **ProGuard**：混淆时需保留 `@JavascriptInterface` 注解的方法
5. **内存**：大文件的 base64 会占用较多内存，xlsx 课表通常在几十 KB 到几百 KB，问题不大
