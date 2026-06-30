import requests
import uuid
import time
import base64
from urllib.parse import urlparse, parse_qs

class HnustJwxtAuth:
    def __init__(self):
        self.session = requests.Session()
        # 必须设置真实浏览器请求头，否则可能被拦截
        self.session.headers.update({
            "User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 "
                          "(KHTML, like Gecko) Chrome/126.0.0.0 Safari/537.36",
            "Accept": "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8",
            "Accept-Language": "zh-CN,zh;q=0.9,en;q=0.8",
            "Accept-Encoding": "gzip, deflate, br",
            "Connection": "keep-alive",
            "Upgrade-Insecure-Requests": "1",
        })
        self.base_url = "https://kdjw.hnust.edu.cn"
        self.uuid = str(uuid.uuid4())
        
    def create_qrcode(self):
        """
        第一步：获取二维码，同时服务器种下 bzb_njw 和 SERVERID
        返回: (qrcode_base64, 是否成功)
        """
        url = f"{self.base_url}/Logon.do?method=QrCodeCreate&uuid={self.uuid}"
        
        # 关键：这次请求会返回 Set-Cookie: bzb_njw 和 SERVERID
        # Session 会自动保存它们
        resp = self.session.get(url, timeout=10)
        
        if resp.status_code != 200:
            return None, False
            
        # 将二维码图片转为 base64，供前端展示
        img_b64 = base64.b64encode(resp.content).decode()
        return f"data:image/jpeg;base64,{img_b64}", True
    
    def poll_status(self, max_retry=60, interval=2):
        """
        第二步：轮询扫码状态
        返回: (是否成功, 状态信息)
        """
        for _ in range(max_retry):
            timestamp = int(time.time() * 1000)
            url = f"{self.base_url}/Logon.do?method=checksfhd&sid={self.uuid}&_={timestamp}"
            
            resp = self.session.get(url, timeout=10)
            # 根据实际响应调整解析逻辑，常见返回格式：
            # "0"=等待扫码, "1"=已扫码待确认, "2"=已确认登录成功
            # 也可能是 JSON: {"status":"1"} 等
            status_text = resp.text.strip()
            
            print(f"[轮询] 状态: {status_text}")
            
            # 根据实际响应调整判断条件
            # 常见状态: "0"=等待扫码, "1"=已扫码待确认, "2"=已确认登录成功
            # 也可能是: "yes"=已确认登录成功
            if status_text == "2" or status_text.lower() == "yes" or "success" in status_text.lower():
                return True, "登录已确认"
            elif status_text == "1":
                print("[轮询] 已扫码，等待确认...")
            else:
                # 继续等待，状态可能是 "0" 或其他
                pass
                
            time.sleep(interval)
            
        return False, "轮询超时"
    
    def do_login(self):
        """
        第三步：执行登录，获取 ticket 并跟随重定向
        返回: (是否成功, cookies字典)
        """
        # 关闭自动重定向，我们需要手动提取 Location 中的 ticket
        login_url = f"{self.base_url}/Logon.do?method=logon_kd&type=wx&sid={self.uuid}"
        resp = self.session.get(login_url, allow_redirects=False, timeout=10)
        
        if resp.status_code != 302:
            print(f"登录请求未返回302，状态码: {resp.status_code}")
            return False, {}
            
        # 提取 Location 头中的跳转 URL
        location = resp.headers.get("Location", "")
        if not location:
            return False, {}
            
        print(f"[登录] 302 跳转至: {location}")
        
        # 跟随跳转到 LoginToXk 接口，获取第三枚 Cookie
        # 这里 location 可能是相对路径或绝对路径
        if location.startswith("http"):
            next_url = location
        else:
            next_url = self.base_url + location
            
        resp2 = self.session.get(next_url, allow_redirects=False, timeout=10)
        
        # 此时应该收到第三枚 Set-Cookie
        print(f"[Ticket] 状态码: {resp2.status_code}")
        print(f"[Ticket] Set-Cookie: {resp2.headers.get('Set-Cookie', '无')}")
        
        # 继续跟随跳转到主页
        if resp2.status_code == 302:
            home_location = resp2.headers.get("Location", "")
            if home_location.startswith("http"):
                home_url = home_location
            else:
                home_url = self.base_url + home_location
                
            # 最终请求主页，验证 Cookie 有效性
            resp3 = self.session.get(home_url, timeout=10)
            print(f"[主页] 状态码: {resp3.status_code}, 长度: {len(resp3.text)}")
            
        # 返回所有 Cookie
        cookies = self.session.cookies.get_dict()
        return True, cookies
    
    def get_cookies(self):
        """获取当前 Session 中的所有 Cookie"""
        return self.session.cookies.get_dict()
    
    def access_page(self, url):
        """使用已获取的 Cookie 访问教务系统内页"""
        resp = self.session.get(url, timeout=10)
        return resp.text
    
    def get_user_info(self):
        """获取用户个人信息（学号和姓名）"""
        url = "https://kdjw.hnust.edu.cn/jsxsd/grsz/grsz_xggrxx.do"
        resp = self.session.get(url, timeout=10)
        
        if resp.status_code != 200:
            print(f"❌ 获取用户信息失败，状态码: {resp.status_code}")
            return None
        
        html = resp.text
        
        # 使用正则表达式提取学号和姓名
        import re
        
        # 提取学号（登录帐号）
        student_id_match = re.search(r'name="account"[^>]*value="([^"]*)"', html)
        student_id = student_id_match.group(1) if student_id_match else "未找到"
        
        # 提取姓名（真实姓名）
        name_match = re.search(r'name="realName"[^>]*value="([^"]*)"', html)
        name = name_match.group(1) if name_match else "未找到"
        
        return {
            "student_id": student_id,
            "name": name
        }


# ==================== 使用示例 ====================

if __name__ == "__main__":
    auth = HnustJwxtAuth()
    
    # 1. 获取二维码
    qrcode_b64, ok = auth.create_qrcode()
    if not ok:
        print("获取二维码失败")
        exit(1)
        
    print(f"UUID: {auth.uuid}")
    print(f"二维码 Base64 (前100字符): {qrcode_b64[:100]}...")
    print(f"当前 Cookie: {auth.get_cookies()}")
    print("请使用微信/企业微信扫描此二维码...")
    
    # 2. 轮询等待扫码（实际部署时应在异步任务中执行）
    success, msg = auth.poll_status()
    if not success:
        print(f"登录失败: {msg}")
        exit(1)
        
    # 3. 执行登录流程
    ok, cookies = auth.do_login()
    if not ok:
        print("登录流程失败")
        exit(1)
        
    print(f"\n✅ 登录成功！最终 Cookie: {cookies}")
    
    # 4. 获取用户信息
    print("\n正在获取用户信息...")
    user_info = auth.get_user_info()
    if user_info:
        print(f"\n📋 用户信息:")
        print(f"   学号: {user_info['student_id']}")
        print(f"   姓名: {user_info['name']}")
    else:
        print("❌ 获取用户信息失败")
