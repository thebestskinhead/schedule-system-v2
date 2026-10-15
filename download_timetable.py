#!/usr/bin/env python
# -*- coding: utf-8 -*-
"""从 kdjw.hnust.edu.cn 下载学期课表。

流程分两步，都走纯 HTTP：
  1. GET  /jsxsd/xskb/xskb_list.do          —— 读页面里学年学期下拉框的选中项，即当前学期
  2. POST /jsxsd/xskb/xskb_print.do?xnxq01id=<学期>&zc=&kbjcmsid=<模板>&wkbkc=1
                                            —— 返回 application/msexcel 附件（旧版 .xls）

依赖登录态 Cookie。kdjw.hnust.edu.cn 域下需要三枚：
  bzb_jsxsd   登录会话令牌（HttpOnly）
  bzb_njw     登录流程中种下
  SERVERID    后端节点会话保持

Cookie 可通过 --cookie 手动传入，或 --from-bridge 从本地浏览器扩展自动读取。

用法：
  python download_timetable.py --cookie "bzb_jsxsd=...; bzb_njw=...; SERVERID=124" -o 课表.xls
  python download_timetable.py --from-bridge -o 课表.xls
  python download_timetable.py --from-bridge --term 2025-2026-2 -o 上学期课表.xls
"""

import argparse
import sys
import os

try:
    import requests
except ImportError:
    sys.exit("需要 requests：pip install requests")


BASE = "https://kdjw.hnust.edu.cn"
LIST_URL = f"{BASE}/jsxsd/xskb/xskb_list.do"
PRINT_URL = f"{BASE}/jsxsd/xskb/xskb_print.do"
DEFAULT_KBJCMSID = "0B841F8A531A4C05BE8DB7DB4B40AEF1"  # 默认节次模式
BRIDGE = "http://127.0.0.1:10086"

HEADERS = {
    "User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 "
                  "(KHTML, like Gecko) Chrome/126.0.0.0 Safari/537.36",
    "Accept": "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
    "Accept-Language": "zh-CN,zh;q=0.9,en;q=0.8",
    "Referer": LIST_URL,
}

XLS_MAGIC = b"\xd0\xcf\x11\xe0\xa1\xb1\x1a\xe1"   # OLE2 复合文档
ZIP_MAGIC = b"PK\x03\x04"                          # xlsx


def fetch_cookies_from_bridge(session_name="timetable-dl"):
    """经本地浏览器扩展读取 kdjw 域 Cookie；需要该站点已在浏览器中登录。"""
    import requests as rq

    url = f"{BASE}/jsxsd/xskb/xskb_list.do"
    body = {"action": "navigate", "args": {"url": url}, "session": session_name}
    rq.post(f"{BRIDGE}/command", json=body, timeout=20)

    r = rq.post(f"{BRIDGE}/command",
                json={"action": "cdp", "args": {"method": "Network.getAllCookies", "params": {}},
                      "session": session_name},
                timeout=20)
    data = r.json()
    cookies = data.get("data", {}).get("cookies", [])
    jar = {}
    for c in cookies:
        if "kdjw.hnust.edu.cn" in c.get("domain", ""):
            jar[c["name"]] = c["value"]
    if not jar:
        raise RuntimeError("未读取到 kdjw.hnust.edu.cn 的 Cookie，请先在浏览器登录教务系统")
    return jar


def detect_current_term(sess):
    """GET 课表列表页，取学年学期下拉框中 selected 的选项值。"""
    import re

    r = sess.get(LIST_URL, timeout=20)
    r.encoding = r.apparent_encoding or "utf-8"
    html = r.text

    m = re.search(r'<select[^>]*name=["\']xnxq01id["\'][^>]*>(.*?)</select>', html, re.S | re.I)
    if not m:
        raise RuntimeError("页面中未找到学年学期下拉框，Cookie 可能已失效")

    block = m.group(1)
    sel = re.search(r'<option[^>]*value=["\']([^"\']+)["\'][^>]*selected', block, re.I)
    if not sel:
        first = re.search(r'<option[^>]*value=["\']([^"\']+)["\']', block, re.I)
        if not first:
            raise RuntimeError("学年学期下拉框为空")
        return first.group(1)

    return sel.group(1)


def download(sess, term, kbjcmsid=DEFAULT_KBJCMSID, wkbkc=1):
    """下载指定学期课表，返回 (二进制内容, 服务端文件名)。"""
    params = {"xnxq01id": term, "zc": "", "kbjcmsid": kbjcmsid}
    if wkbkc:
        params["wkbkc"] = "1"

    r = sess.post(PRINT_URL, params=params, timeout=60)
    ctype = r.headers.get("Content-Type", "")
    body = r.content

    if "msexcel" not in ctype.lower() and not (body[:8] == XLS_MAGIC or body[:4] == ZIP_MAGIC):
        raise RuntimeError(
            f"未取得课表文件（Content-Type={ctype}，{len(body)} 字节）。Cookie 可能已失效，"
            f"请重新登录后重试。")

    filename = "timetable.xls"
    disp = r.headers.get("Content-disposition", "")
    if "filename=" in disp:
        raw = disp.split("filename=")[-1].strip().strip('"')
        try:
            filename = raw.encode("latin-1").decode("gbk")
        except (UnicodeDecodeError, UnicodeEncodeError):
            filename = raw

    return body, filename


def main():
    ap = argparse.ArgumentParser(description="下载湖南科技大学教务系统学期课表")
    ap.add_argument("--cookie", help='Cookie 字符串，如 "bzb_jsxsd=...; bzb_njw=...; SERVERID=124"')
    ap.add_argument("--from-bridge", action="store_true",
                    help="从本地浏览器自动读取 Cookie（需浏览器已登录教务系统）")
    ap.add_argument("--term", help="学年学期，如 2026-2027-1；省略则自动识别当前学期")
    ap.add_argument("--kbjcmsid", default=DEFAULT_KBJCMSID, help="节次模板 ID")
    ap.add_argument("--no-wkbkc", action="store_true", help="不包含无课表课程")
    ap.add_argument("-o", "--output", help="输出文件路径；省略则使用服务端文件名")
    args = ap.parse_args()

    if not args.cookie and not args.from_bridge:
        ap.error("需指定 --cookie 或 --from-bridge")

    sess = requests.Session()
    sess.headers.update(HEADERS)

    if args.cookie:
        for pair in args.cookie.split(";"):
            pair = pair.strip()
            if "=" in pair:
                k, v = pair.split("=", 1)
                sess.cookies.set(k.strip(), v.strip(), domain="kdjw.hnust.edu.cn")
    else:
        jar = fetch_cookies_from_bridge()
        for k, v in jar.items():
            sess.cookies.set(k, v, domain="kdjw.hnust.edu.cn")
        print(f"已读取 Cookie: {', '.join(jar.keys())}")

    term = args.term or detect_current_term(sess)
    print(f"学年学期: {term}")

    body, filename = download(sess, term, args.kbjcmsid, 0 if args.no_wkbkc else 1)

    out = args.output or filename
    with open(out, "wb") as f:
        f.write(body)

    print(f"已保存: {os.path.abspath(out)}  ({len(body)} 字节)")


if __name__ == "__main__":
    main()
